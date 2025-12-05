package handlers

import (
	"chrisgross-ctrl-project/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TieredStatsHandlers handles tiered statistics API endpoints
type TieredStatsHandlers struct {
	DB                    *gorm.DB
	dashboardStatsService *services.DashboardStatsService
}

// NewTieredStatsHandlers creates a new tiered stats handler
func NewTieredStatsHandlers(db *gorm.DB, dashboardStatsService *services.DashboardStatsService) *TieredStatsHandlers {
	return &TieredStatsHandlers{
		DB:                    db,
		dashboardStatsService: dashboardStatsService,
	}
}

// GetLiveStats returns live data (updates every 15 seconds)
// GET /api/stats/live
func (h *TieredStatsHandlers) GetLiveStats(c *gin.Context) {
	stats, err := h.dashboardStatsService.GetLiveStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetHotStats returns hot data (updates every 5 minutes)
// GET /api/stats/hot
func (h *TieredStatsHandlers) GetHotStats(c *gin.Context) {
	stats, err := h.dashboardStatsService.GetHotStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetWarmStats returns warm data (updates every 1 hour)
// GET /api/stats/warm
func (h *TieredStatsHandlers) GetWarmStats(c *gin.Context) {
	stats, err := h.dashboardStatsService.GetWarmStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetDailyStats returns daily data (updates once per day at midnight)
// GET /api/stats/daily
func (h *TieredStatsHandlers) GetDailyStats(c *gin.Context) {
	stats, err := h.dashboardStatsService.GetDailyStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Helper: Get top AI opportunities
func (h *TieredStatsHandlers) getTopOpportunities() []map[string]interface{} {
	// This should call the AI orchestrator service
	// For now, return sample data structure
	
	opportunities := []map[string]interface{}{}
	
	// Query recent high-scoring leads
	type OpportunityLead struct {
		ID                   uint
		Name                 string
		Email                string
		BehavioralScore      float64
		ConversionProbability float64
		LastContactAt        *time.Time
		ViewCount            int
	}
	
	var leads []OpportunityLead
	h.DB.Table("leads").
		Select("id, name, email, behavioral_score, conversion_probability, last_contact_at").
		Where("behavioral_score > ?", 70).
		Order("behavioral_score DESC").
		Limit(10).
		Find(&leads)
	
	for _, lead := range leads {
		oppType := "hot_lead"
		priority := int(lead.BehavioralScore)
		
		// Determine opportunity type
		if lead.LastContactAt != nil && time.Since(*lead.LastContactAt) > 7*24*time.Hour {
			oppType = "re_engagement"
			priority = 80
		}
		
		opportunities = append(opportunities, map[string]interface{}{
			"id":                     lead.ID,
			"type":                   oppType,
			"priority":               priority,
			"lead_name":              lead.Name,
			"lead_email":             lead.Email,
			"conversion_probability": lead.ConversionProbability,
			"behavioral_score":       lead.BehavioralScore,
			"action_sequence": []map[string]interface{}{
				{
					"action":      "send_email",
					"description": "Send Follow-Up Email",
				},
				{
					"action":      "schedule_call",
					"description": "Schedule Call",
				},
			},
		})
	}
	
	return opportunities
}

// Helper: Get recent behavioral scores
func (h *TieredStatsHandlers) getRecentBehavioralScores() map[string]interface{} {
	var avgScore float64
	var highScoreCount int64
	
	h.DB.Table("leads").
		Select("COALESCE(AVG(behavioral_score), 0)").
		Scan(&avgScore)
	
	h.DB.Table("leads").
		Where("behavioral_score > ?", 70).
		Count(&highScoreCount)
	
	return map[string]interface{}{
		"average_score":     avgScore,
		"high_score_count":  highScoreCount,
	}
}

// Helper: Get lead source breakdown
func (h *TieredStatsHandlers) getLeadSourceBreakdown() []map[string]interface{} {
	type SourceCount struct {
		Source string
		Count  int64
	}
	
	var sources []SourceCount
	h.DB.Table("leads").
		Select("source, COUNT(*) as count").
		Group("source").
		Order("count DESC").
		Limit(5).
		Find(&sources)
	
	result := []map[string]interface{}{}
	for _, s := range sources {
		result = append(result, map[string]interface{}{
			"source": s.Source,
			"count":  s.Count,
		})
	}
	
	return result
}

// Helper: Get funnel metrics
func (h *TieredStatsHandlers) getFunnelMetrics() map[string]interface{} {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	var totalVisitors int64
	var totalLeads int64
	var totalApplications int64
	var totalConversions int64
	
	h.DB.Table("behavioral_events").
		Where("created_at > ?", thirtyDaysAgo).
		Distinct("session_id").
		Count(&totalVisitors)
	
	h.DB.Table("leads").
		Where("created_at > ?", thirtyDaysAgo).
		Count(&totalLeads)
	
	h.DB.Table("applications").
		Where("created_at > ?", thirtyDaysAgo).
		Count(&totalApplications)
	
	h.DB.Table("applications").
		Where("created_at > ? AND status = ?", thirtyDaysAgo, "approved").
		Count(&totalConversions)
	
	return map[string]interface{}{
		"visitors":     totalVisitors,
		"leads":        totalLeads,
		"applications": totalApplications,
		"conversions":  totalConversions,
	}
}
