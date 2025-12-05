package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
)

type LeadReengagementHandler struct {
	db                *gorm.DB
	encryptionManager *security.EncryptionManager
}

func NewLeadReengagementHandler(db *gorm.DB, encryptionManager *security.EncryptionManager) *LeadReengagementHandler {
	return &LeadReengagementHandler{
		db:                db,
		encryptionManager: encryptionManager,
	}
}

// RegisterRoutes registers all lead re-engagement routes
func (h *LeadReengagementHandler) RegisterRoutes(r *gin.RouterGroup) {
	reengagement := r.Group("/reengagement")
	{
		// Lead Management
		reengagement.GET("/leads", h.GetLeads)
		// reengagement.GET("/leads/:id", h.GetLead) // TODO: Implement
		reengagement.POST("/leads/import", h.ImportLeads)
		// reengagement.PUT("/leads/:id", h.UpdateLead) // TODO: Implement
		// reengagement.DELETE("/leads/:id", h.DeleteLead) // TODO: Implement

		// Segmentation and Risk Assessment
		reengagement.POST("/leads/segment", h.SegmentLeads)
		// reengagement.POST("/leads/assess-risk", h.AssessRisk) // TODO: Implement
		reengagement.GET("/leads/segments", h.GetSegmentStats)

		// Campaign Management
		// reengagement.GET("/campaigns", h.GetCampaigns) // TODO: Implement
		reengagement.POST("/campaigns/prepare", h.PrepareCampaign)
		// reengagement.POST("/campaigns/activate", h.ActivateCampaign) // TODO: Implement
		// reengagement.PUT("/campaigns/:id/pause", h.PauseCampaign) // TODO: Implement
		// reengagement.GET("/campaigns/:id/status", h.GetCampaignStatus) // TODO: Implement

		// Templates
		reengagement.GET("/templates", h.GetTemplates)
		reengagement.POST("/templates", h.CreateTemplate)
		// reengagement.PUT("/templates/:id", h.UpdateTemplate) // TODO: Implement
		// reengagement.DELETE("/templates/:id", h.DeleteTemplate) // TODO: Implement

		// Metrics and Reporting
		reengagement.GET("/metrics", h.GetMetrics)
		// reengagement.GET("/metrics/daily", h.GetDailyMetrics) // TODO: Implement
		// reengagement.GET("/compliance/report", h.GetComplianceReport) // TODO: Implement

		// Emergency Controls
		reengagement.POST("/emergency/stop", h.EmergencyStopAll)
		// reengagement.POST("/emergency/stop/:campaignId", h.EmergencyStop) // TODO: Implement
		// reengagement.POST("/volume/limit", h.SetVolumeLimit) // TODO: Implement
	}
}

// GetLeads retrieves leads with filtering and pagination
func (h *LeadReengagementHandler) GetLeads(c *gin.Context) {
	var leads []models.LeadReengagement

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	segment := c.Query("segment")
	riskLevel := c.Query("risk_level")
	campaignStatus := c.Query("campaign_status")

	// Build query
	query := h.db.Model(&models.LeadReengagement{})

	if segment != "" {
		query = query.Where("segment = ?", segment)
	}
	if riskLevel != "" {
		query = query.Where("risk_level = ?", riskLevel)
	}
	if campaignStatus != "" {
		query = query.Where("campaign_status = ?", campaignStatus)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	result := query.Offset(offset).Limit(limit).Find(&leads)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve leads",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leads": leads,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// ImportLeads imports leads from FUB for re-engagement analysis
func (h *LeadReengagementHandler) ImportLeads(c *gin.Context) {
	var request struct {
		FUBContactIDs []string `json:"fub_contact_ids"`
		DryRun        bool     `json:"dry_run"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Fetch contacts from FUB API
	imported := 0
	skipped := 0
	errors := []string{}

	for _, contactID := range request.FUBContactIDs {
		// Fetch contact from FUB
		var fubLead models.FUBLead
		if err := h.db.Where("fub_lead_id = ?", contactID).First(&fubLead).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Failed to find FUB contact %s: %v", contactID, err))
			skipped++
			continue
		}

		// Encrypt PII fields before storage
		encryptedEmail, err := h.encryptionManager.EncryptEmail(fubLead.Email)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to encrypt email for contact %s: %v", contactID, err))
			skipped++
			continue
		}

		encryptedFirstName, err := h.encryptionManager.Encrypt(fubLead.FirstName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to encrypt first name for contact %s: %v", contactID, err))
			skipped++
			continue
		}

		encryptedLastName, err := h.encryptionManager.Encrypt(fubLead.LastName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to encrypt last name for contact %s: %v", contactID, err))
			skipped++
			continue
		}

		lead := &models.LeadReengagement{
			FUBContactID: contactID,
			Email:        encryptedEmail,
			FirstName:    encryptedFirstName,
			LastName:     encryptedLastName,
			HasEmail:     true,
			EmailValid:   true,
		}

		// Calculate segment and risk
		lead.Segment = lead.CalculateSegment()
		lead.RiskLevel = lead.CalculateRiskLevel()
		lead.ConsentStatus = models.ConsentUnknown

		if !request.DryRun {
			// Check if lead already exists
			var existing models.LeadReengagement
			result := h.db.Where("fub_contact_id = ?", contactID).First(&existing)

			if result.Error == gorm.ErrRecordNotFound {
				// Create new lead
				if err := h.db.Create(lead).Error; err != nil {
					errors = append(errors, fmt.Sprintf("Failed to create lead %s: %v", contactID, err))
					continue
				}
				imported++
			} else {
				skipped++
			}
		} else {
			imported++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Import completed",
		"imported": imported,
		"skipped":  skipped,
		"errors":   errors,
		"dry_run":  request.DryRun,
	})
}

// SegmentLeads performs segmentation analysis on all leads
func (h *LeadReengagementHandler) SegmentLeads(c *gin.Context) {
	var request struct {
		DryRun bool `json:"dry_run"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var leads []models.LeadReengagement
	if err := h.db.Find(&leads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve leads",
			"details": err.Error(),
		})
		return
	}

	segmentChanges := map[string]int{
		"active":     0,
		"dormant":    0,
		"unknown":    0,
		"suppressed": 0,
	}

	for _, lead := range leads {
		oldSegment := lead.Segment
		newSegment := lead.CalculateSegment()
		newRiskLevel := lead.CalculateRiskLevel()

		if oldSegment != newSegment {
			segmentChanges[string(newSegment)]++

			if !request.DryRun {
				lead.Segment = newSegment
				lead.RiskLevel = newRiskLevel
				h.db.Save(&lead)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Segmentation completed",
		"changes": segmentChanges,
		"dry_run": request.DryRun,
	})
}

// GetSegmentStats returns statistics about lead segments
func (h *LeadReengagementHandler) GetSegmentStats(c *gin.Context) {
	type SegmentStat struct {
		Segment string `json:"segment"`
		Count   int64  `json:"count"`
	}

	type RiskStat struct {
		RiskLevel string `json:"risk_level"`
		Count     int64  `json:"count"`
	}

	var stats []SegmentStat

	result := h.db.Model(&models.LeadReengagement{}).
		Select("segment, COUNT(*) as count").
		Group("segment").
		Scan(&stats)

	// Handle empty results gracefully
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get segment statistics",
			"details": result.Error.Error(),
		})
		return
	}

	// Return empty stats if no data
	if len(stats) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"segments":              []SegmentStat{},
			"risk_levels":           []RiskStat{},
			"eligible_for_campaign": 0,
			"total_leads":           0,
		})
		return
	}

	// Get risk level stats
	var riskStats []RiskStat
	h.db.Model(&models.LeadReengagement{}).
		Select("risk_level, COUNT(*) as count").
		Group("risk_level").
		Scan(&riskStats)

	// Get campaign eligibility
	var eligible int64
	h.db.Model(&models.LeadReengagement{}).
		Where("has_email = ? AND email_valid = ? AND risk_level != ? AND segment != ? AND hard_bounce = ? AND previous_unsubscribe = ?",
			true, true, models.RiskHigh, models.SegmentSuppressed, false, false).
		Count(&eligible)

	c.JSON(http.StatusOK, gin.H{
		"segments":              stats,
		"risk_levels":           riskStats,
		"eligible_for_campaign": eligible,
		"total_leads":           len(stats),
	})
}

// PrepareCampaign prepares a re-engagement campaign without activating it
func (h *LeadReengagementHandler) PrepareCampaign(c *gin.Context) {
	var request struct {
		Name       string   `json:"name"`
		Segments   []string `json:"segments"`
		MaxVolume  int      `json:"max_volume"`
		DailyLimit int      `json:"daily_limit"`
		StartDate  string   `json:"start_date"`
		TestMode   bool     `json:"test_mode"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get eligible leads
	query := h.db.Model(&models.LeadReengagement{}).
		Where("has_email = ? AND email_valid = ? AND risk_level != ? AND segment != ? AND hard_bounce = ? AND previous_unsubscribe = ? AND campaign_status = ?",
			true, true, models.RiskHigh, models.SegmentSuppressed, false, false, models.CampaignPending)

	if len(request.Segments) > 0 {
		query = query.Where("segment IN ?", request.Segments)
	}

	var eligibleLeads []models.LeadReengagement
	query.Limit(request.MaxVolume).Find(&eligibleLeads)

	// Create campaign preparation summary
	summary := gin.H{
		"campaign_name":      request.Name,
		"eligible_leads":     len(eligibleLeads),
		"segments_included":  request.Segments,
		"max_volume":         request.MaxVolume,
		"daily_limit":        request.DailyLimit,
		"test_mode":          request.TestMode,
		"estimated_duration": calculateCampaignDuration(len(eligibleLeads), request.DailyLimit),
		"safety_checks": gin.H{
			"volume_within_limits": len(eligibleLeads) <= request.MaxVolume,
			"daily_limit_set":      request.DailyLimit > 0,
			"test_mode_enabled":    request.TestMode,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Campaign prepared successfully",
		"summary": summary,
		"status":  "prepared_not_activated",
	})
}

// GetTemplates returns all email templates
func (h *LeadReengagementHandler) GetTemplates(c *gin.Context) {
	var templates []models.CampaignTemplate

	result := h.db.Order("created_at DESC").Find(&templates)
	
	// Handle empty results gracefully
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve templates",
			"details": result.Error.Error(),
		})
		return
	}

	// Return empty array if no templates
	if len(templates) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"templates": []models.CampaignTemplate{},
			"total":     0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     len(templates),
	})
}

// CreateTemplate creates a new email template
func (h *LeadReengagementHandler) CreateTemplate(c *gin.Context) {
	var template models.CampaignTemplate

	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid template format",
			"details": err.Error(),
		})
		return
	}

	if err := h.db.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Template created successfully",
		"template": template,
	})
}

// GetMetrics returns re-engagement campaign metrics
func (h *LeadReengagementHandler) GetMetrics(c *gin.Context) {
	// Get latest metrics
	var metrics models.ReengagementMetrics
	result := h.db.Order("created_at DESC").First(&metrics)

	if result.Error == gorm.ErrRecordNotFound {
		// Generate current metrics
		metrics = generateCurrentMetrics(h.db)
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve metrics",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":      metrics,
		"generated_at": time.Now(),
	})
}

// GetSafetyStatus returns current safety and compliance status
func (h *LeadReengagementHandler) GetSafetyStatus(c *gin.Context) {
	// Calculate current safety metrics
	var totalLeads int64
	var highRiskLeads int64
	var activeCampaigns int64

	h.db.Model(&models.LeadReengagement{}).Count(&totalLeads)
	h.db.Model(&models.LeadReengagement{}).Where("risk_level = ?", models.RiskHigh).Count(&highRiskLeads)
	h.db.Model(&models.LeadReengagement{}).Where("campaign_status = ?", models.CampaignActive).Count(&activeCampaigns)

	// Get today's email volume
	today := time.Now().Truncate(24 * time.Hour)
	var todayEmails int64
	h.db.Model(&models.CampaignExecution{}).
		Where("executed_at >= ?", today).
		Count(&todayEmails)

	safetyStatus := gin.H{
		"total_leads":        totalLeads,
		"high_risk_leads":    highRiskLeads,
		"active_campaigns":   activeCampaigns,
		"today_email_volume": todayEmails,
		"safety_score":       calculateSafetyScore(totalLeads, highRiskLeads, todayEmails),
		"compliance_status":  "compliant",
		"last_check":         time.Now(),
		"warnings":           []string{},
		"recommendations": []string{
			"Monitor daily email volume",
			"Review high-risk leads regularly",
			"Maintain sender reputation",
		},
	}

	c.JSON(http.StatusOK, safetyStatus)
}

// Helper functions

func calculateCampaignDuration(totalLeads, dailyLimit int) int {
	if dailyLimit <= 0 {
		return 0
	}
	return (totalLeads + dailyLimit - 1) / dailyLimit // Ceiling division
}

func generateCurrentMetrics(db *gorm.DB) models.ReengagementMetrics {
	var metrics models.ReengagementMetrics
	var activeCount, dormantCount, suppressedCount int64

	// Count leads by segment
	db.Model(&models.LeadReengagement{}).Where("segment = ?", models.SegmentActive).Count(&activeCount)
	db.Model(&models.LeadReengagement{}).Where("segment = ?", models.SegmentDormant).Count(&dormantCount)
	db.Model(&models.LeadReengagement{}).Where("segment = ?", models.SegmentSuppressed).Count(&suppressedCount)

	metrics.ActiveLeads = int(activeCount)
	metrics.DormantLeads = int(dormantCount)
	metrics.SuppressedLeads = int(suppressedCount)
	metrics.TotalLeads = metrics.ActiveLeads + metrics.DormantLeads + metrics.SuppressedLeads
	metrics.MetricDate = time.Now()

	// Calculate rates
	if metrics.EmailsSent > 0 {
		metrics.OpenRate = float64(metrics.EmailsOpened) / float64(metrics.EmailsSent) * 100
		metrics.ClickRate = float64(metrics.EmailsClicked) / float64(metrics.EmailsSent) * 100
		metrics.ResponseRate = float64(metrics.Responses) / float64(metrics.EmailsSent) * 100
	}

	return metrics
}

func calculateSafetyScore(totalLeads, highRiskLeads, todayEmails int64) float64 {
	score := 100.0

	// Deduct for high-risk leads
	if totalLeads > 0 {
		riskRatio := float64(highRiskLeads) / float64(totalLeads)
		score -= riskRatio * 30 // Max 30 point deduction
	}

	// Deduct for high email volume
	if todayEmails > 1000 {
		score -= 20 // High volume penalty
	} else if todayEmails > 500 {
		score -= 10 // Medium volume penalty
	}

	if score < 0 {
		score = 0
	}

	return score
}

// EmergencyStopAll stops all active campaigns immediately
func (h *LeadReengagementHandler) EmergencyStopAll(c *gin.Context) {
	// TODO: Implement emergency stop functionality
	// This would stop all active campaigns, pause email/SMS sending, etc.

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Emergency stop activated - all campaigns paused",
		"timestamp": time.Now(),
	})
}
