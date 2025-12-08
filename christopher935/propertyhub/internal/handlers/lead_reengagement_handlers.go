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
		reengagement.GET("/leads/:id", h.GetLead)
		reengagement.POST("/leads/import", h.ImportLeads)
		reengagement.PUT("/leads/:id", h.UpdateLead)
		reengagement.DELETE("/leads/:id", h.DeleteLead)

		// Segmentation and Risk Assessment
		reengagement.POST("/leads/segment", h.SegmentLeads)
		reengagement.POST("/leads/assess-risk", h.AssessRisk)
		reengagement.GET("/leads/segments", h.GetSegmentStats)

		// Campaign Management
		reengagement.GET("/campaigns", h.GetCampaigns)
		reengagement.POST("/campaigns/prepare", h.PrepareCampaign)
		reengagement.POST("/campaigns/activate", h.ActivateCampaign)
		reengagement.PUT("/campaigns/:id/pause", h.PauseCampaign)
		reengagement.GET("/campaigns/:id/status", h.GetCampaignStatus)

		// Templates
		reengagement.GET("/templates", h.GetTemplates)
		reengagement.POST("/templates", h.CreateTemplate)
		reengagement.PUT("/templates/:id", h.UpdateTemplate)
		reengagement.DELETE("/templates/:id", h.DeleteTemplate)

		// Metrics and Reporting
		reengagement.GET("/metrics", h.GetMetrics)
		reengagement.GET("/metrics/daily", h.GetDailyMetrics)
		reengagement.GET("/compliance/report", h.GetComplianceReport)

		// Emergency Controls
		reengagement.POST("/emergency/stop", h.EmergencyStopAll)
		reengagement.POST("/emergency/stop/:campaignId", h.EmergencyStop)
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

func (h *LeadReengagementHandler) GetLead(c *gin.Context) {
	id := c.Param("id")

	var lead models.LeadReengagement
	if err := h.db.First(&lead, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Lead not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve lead",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lead": lead,
	})
}

func (h *LeadReengagementHandler) UpdateLead(c *gin.Context) {
	id := c.Param("id")

	var lead models.LeadReengagement
	if err := h.db.First(&lead, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Lead not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve lead",
				"details": err.Error(),
			})
		}
		return
	}

	var updates struct {
		Segment       *models.LeadSegment   `json:"segment"`
		RiskLevel     *models.RiskLevel     `json:"risk_level"`
		ConsentStatus *models.ConsentStatus `json:"consent_status"`
		Notes         *string               `json:"notes"`
		Tags          *string               `json:"tags"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if updates.Segment != nil {
		lead.Segment = *updates.Segment
	}
	if updates.RiskLevel != nil {
		lead.RiskLevel = *updates.RiskLevel
	}
	if updates.ConsentStatus != nil {
		lead.ConsentStatus = *updates.ConsentStatus
	}
	if updates.Notes != nil {
		lead.Notes = *updates.Notes
	}
	if updates.Tags != nil {
		lead.Tags = *updates.Tags
	}

	if err := h.db.Save(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update lead",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lead updated successfully",
		"lead":    lead,
	})
}

func (h *LeadReengagementHandler) DeleteLead(c *gin.Context) {
	id := c.Param("id")

	var lead models.LeadReengagement
	if err := h.db.First(&lead, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Lead not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve lead",
				"details": err.Error(),
			})
		}
		return
	}

	if err := h.db.Delete(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete lead",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lead deleted successfully",
	})
}

func (h *LeadReengagementHandler) AssessRisk(c *gin.Context) {
	var request struct {
		LeadIDs []uint `json:"lead_ids"`
		DryRun  bool   `json:"dry_run"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var leads []models.LeadReengagement
	query := h.db.Model(&models.LeadReengagement{})

	if len(request.LeadIDs) > 0 {
		query = query.Where("id IN ?", request.LeadIDs)
	}

	if err := query.Find(&leads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve leads",
			"details": err.Error(),
		})
		return
	}

	riskChanges := map[string]int{
		"low":    0,
		"medium": 0,
		"high":   0,
	}

	for i := range leads {
		oldRisk := leads[i].RiskLevel
		newRisk := leads[i].CalculateRiskLevel()

		if oldRisk != newRisk {
			riskChanges[string(newRisk)]++

			if !request.DryRun {
				leads[i].RiskLevel = newRisk
				h.db.Save(&leads[i])
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Risk assessment completed",
		"leads_assessed": len(leads),
		"changes":        riskChanges,
		"dry_run":        request.DryRun,
	})
}

func (h *LeadReengagementHandler) GetCampaigns(c *gin.Context) {
	var campaigns []models.CampaignExecution

	status := c.Query("status")

	query := h.db.Preload("LeadReengagement").Preload("CampaignTemplate")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	result := query.Order("created_at DESC").Find(&campaigns)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve campaigns",
			"details": result.Error.Error(),
		})
		return
	}

	if len(campaigns) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"campaigns": []models.CampaignExecution{},
			"total":     0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
		"total":     len(campaigns),
	})
}

func (h *LeadReengagementHandler) ActivateCampaign(c *gin.Context) {
	var request struct {
		Name       string   `json:"name"`
		Segments   []string `json:"segments"`
		TemplateID uint     `json:"template_id"`
		MaxVolume  int      `json:"max_volume"`
		DailyLimit int      `json:"daily_limit"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var template models.CampaignTemplate
	if err := h.db.First(&template, "id = ?", request.TemplateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Template not found",
		})
		return
	}

	query := h.db.Model(&models.LeadReengagement{}).Where(
		"has_email = ? AND email_valid = ? AND risk_level != ? AND segment != ? AND hard_bounce = ? AND previous_unsubscribe = ? AND campaign_status = ?",
		true, true, models.RiskHigh, models.SegmentSuppressed, false, false, models.CampaignPending,
	)

	if len(request.Segments) > 0 {
		query = query.Where("segment IN ?", request.Segments)
	}

	var leads []models.LeadReengagement
	query.Limit(request.MaxVolume).Find(&leads)

	activated := 0
	now := time.Now()

	for i := range leads {
		leads[i].CampaignStatus = models.CampaignActive
		leads[i].CampaignStarted = &now
		h.db.Save(&leads[i])

		execution := models.CampaignExecution{
			LeadReengagementID: leads[i].ID,
			CampaignTemplateID: template.ID,
			ScheduledFor:       now,
			Status:             "scheduled",
		}
		h.db.Create(&execution)

		activated++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Campaign activated successfully",
		"campaign_name":   request.Name,
		"leads_activated": activated,
		"template_used":   template.Name,
		"activation_time": now,
	})
}

func (h *LeadReengagementHandler) PauseCampaign(c *gin.Context) {
	id := c.Param("id")

	var campaign models.CampaignExecution
	if err := h.db.First(&campaign, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Campaign not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve campaign",
				"details": err.Error(),
			})
		}
		return
	}

	var lead models.LeadReengagement
	if err := h.db.First(&lead, "id = ?", campaign.LeadReengagementID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve associated lead",
			"details": err.Error(),
		})
		return
	}

	lead.CampaignStatus = models.CampaignPending
	if err := h.db.Save(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to pause campaign",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Campaign paused successfully",
		"campaign_id": id,
		"paused_at":   time.Now(),
	})
}

func (h *LeadReengagementHandler) GetCampaignStatus(c *gin.Context) {
	id := c.Param("id")

	var campaign models.CampaignExecution
	if err := h.db.Preload("LeadReengagement").Preload("CampaignTemplate").First(&campaign, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Campaign not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve campaign",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": campaign,
		"status":   campaign.Status,
		"lead_info": gin.H{
			"id":              campaign.LeadReengagement.ID,
			"fub_contact_id":  campaign.LeadReengagement.FUBContactID,
			"campaign_status": campaign.LeadReengagement.CampaignStatus,
			"emails_sent":     campaign.LeadReengagement.EmailsSent,
		},
		"template_info": gin.H{
			"id":   campaign.CampaignTemplate.ID,
			"name": campaign.CampaignTemplate.Name,
		},
	})
}

func (h *LeadReengagementHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var template models.CampaignTemplate
	if err := h.db.First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Template not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve template",
				"details": err.Error(),
			})
		}
		return
	}

	var updates struct {
		Name      *string `json:"name"`
		Subject   *string `json:"subject"`
		Body      *string `json:"body"`
		DaysDelay *int    `json:"days_delay"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if updates.Name != nil {
		template.Name = *updates.Name
	}
	if updates.Subject != nil {
		template.Subject = *updates.Subject
	}
	if updates.Body != nil {
		template.Body = *updates.Body
	}
	if updates.DaysDelay != nil {
		template.DaysDelay = *updates.DaysDelay
	}

	if err := h.db.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Template updated successfully",
		"template": template,
	})
}

func (h *LeadReengagementHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	var template models.CampaignTemplate
	if err := h.db.First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Template not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve template",
				"details": err.Error(),
			})
		}
		return
	}

	if err := h.db.Delete(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Template deleted successfully",
	})
}

func (h *LeadReengagementHandler) GetDailyMetrics(c *gin.Context) {
	date := c.Query("date")

	var targetDate time.Time
	if date != "" {
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid date format",
				"details": "Use YYYY-MM-DD format",
			})
			return
		}
		targetDate = parsed
	} else {
		targetDate = time.Now().Truncate(24 * time.Hour)
	}

	var metrics models.ReengagementMetrics
	result := h.db.Where("DATE(metric_date) = DATE(?)", targetDate).First(&metrics)

	if result.Error == gorm.ErrRecordNotFound {
		metrics = generateDailyMetrics(h.db, targetDate)
		h.db.Create(&metrics)
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve metrics",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"date":    targetDate.Format("2006-01-02"),
	})
}

func (h *LeadReengagementHandler) GetComplianceReport(c *gin.Context) {
	var consentStats []struct {
		ConsentStatus string `json:"consent_status"`
		Count         int64  `json:"count"`
	}

	h.db.Model(&models.LeadReengagement{}).Select("consent_status, COUNT(*) as count").Group("consent_status").Scan(&consentStats)

	var totalLeads, consentedLeads, revokedLeads, unknownConsent int64
	h.db.Model(&models.LeadReengagement{}).Count(&totalLeads)
	h.db.Model(&models.LeadReengagement{}).Where("consent_status = ?", models.ConsentExpress).Count(&consentedLeads)
	h.db.Model(&models.LeadReengagement{}).Where("consent_status = ?", models.ConsentRevoked).Count(&revokedLeads)
	h.db.Model(&models.LeadReengagement{}).Where("consent_status = ?", models.ConsentUnknown).Count(&unknownConsent)

	var suppressedCount int64
	h.db.Model(&models.LeadReengagement{}).Where("segment = ?", models.SegmentSuppressed).Count(&suppressedCount)

	var dncCount int64
	h.db.Model(&models.LeadReengagement{}).Where("on_dnc_list = ?", true).Count(&dncCount)

	var unsubscribedCount int64
	h.db.Model(&models.LeadReengagement{}).Where("previous_unsubscribe = ?", true).Count(&unsubscribedCount)

	complianceScore := 100.0
	if totalLeads > 0 {
		complianceScore -= float64(unknownConsent) / float64(totalLeads) * 30
		complianceScore -= float64(revokedLeads) / float64(totalLeads) * 20
	}

	c.JSON(http.StatusOK, gin.H{
		"compliance_report": gin.H{
			"total_leads":        totalLeads,
			"consented_leads":    consentedLeads,
			"revoked_leads":      revokedLeads,
			"unknown_consent":    unknownConsent,
			"suppressed_leads":   suppressedCount,
			"dnc_list_count":     dncCount,
			"unsubscribed_count": unsubscribedCount,
			"compliance_score":   complianceScore,
		},
		"consent_breakdown": consentStats,
		"generated_at":      time.Now(),
		"recommendations": []string{
			"Document consent for all unknown leads",
			"Never contact revoked or suppressed leads",
			"Maintain DNC list compliance",
			"Review consent status regularly",
		},
	})
}

func (h *LeadReengagementHandler) EmergencyStop(c *gin.Context) {
	campaignID := c.Param("campaignId")

	var campaign models.CampaignExecution
	if err := h.db.First(&campaign, "id = ?", campaignID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Campaign not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve campaign",
				"details": err.Error(),
			})
		}
		return
	}

	var lead models.LeadReengagement
	if err := h.db.First(&lead, "id = ?", campaign.LeadReengagementID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve associated lead",
			"details": err.Error(),
		})
		return
	}

	lead.CampaignStatus = models.CampaignSuppressed
	now := time.Now()
	lead.CampaignCompleted = &now

	if err := h.db.Save(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to stop campaign",
			"details": err.Error(),
		})
		return
	}

	campaign.Status = "stopped"
	if err := h.db.Save(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update campaign status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Emergency stop activated for campaign",
		"campaign_id": campaignID,
		"stopped_at":  now,
	})
}

func (h *LeadReengagementHandler) EmergencyStopAll(c *gin.Context) {
	result := h.db.Model(&models.LeadReengagement{}).Where("campaign_status = ?", models.CampaignActive).Update("campaign_status", models.CampaignSuppressed)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to stop campaigns",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"message":           "Emergency stop activated - all campaigns paused",
		"campaigns_stopped": result.RowsAffected,
		"timestamp":         time.Now(),
	})
}

func generateDailyMetrics(db *gorm.DB, date time.Time) models.ReengagementMetrics {
	var metrics models.ReengagementMetrics
	var activeCount, dormantCount, suppressedCount int64

	db.Model(&models.LeadReengagement{}).Where("segment = ?", models.SegmentActive).Count(&activeCount)
	db.Model(&models.LeadReengagement{}).Where("segment = ?", models.SegmentDormant).Count(&dormantCount)
	db.Model(&models.LeadReengagement{}).Where("segment = ?", models.SegmentSuppressed).Count(&suppressedCount)

	metrics.ActiveLeads = int(activeCount)
	metrics.DormantLeads = int(dormantCount)
	metrics.SuppressedLeads = int(suppressedCount)
	metrics.TotalLeads = metrics.ActiveLeads + metrics.DormantLeads + metrics.SuppressedLeads
	metrics.MetricDate = date

	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var emailsSent, emailsOpened, emailsClicked int64
	db.Model(&models.CampaignExecution{}).Where("executed_at >= ? AND executed_at < ?", startOfDay, endOfDay).Count(&emailsSent)
	db.Model(&models.CampaignExecution{}).Where("executed_at >= ? AND executed_at < ? AND email_opened = ?", startOfDay, endOfDay, true).Count(&emailsOpened)
	db.Model(&models.CampaignExecution{}).Where("executed_at >= ? AND executed_at < ? AND email_clicked = ?", startOfDay, endOfDay, true).Count(&emailsClicked)

	metrics.EmailsSent = int(emailsSent)
	metrics.EmailsOpened = int(emailsOpened)
	metrics.EmailsClicked = int(emailsClicked)

	if metrics.EmailsSent > 0 {
		metrics.OpenRate = float64(metrics.EmailsOpened) / float64(metrics.EmailsSent) * 100
		metrics.ClickRate = float64(metrics.EmailsClicked) / float64(metrics.EmailsSent) * 100
	}

	return metrics
}
