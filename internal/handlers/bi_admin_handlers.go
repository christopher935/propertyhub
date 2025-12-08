package handlers

import (
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BIAdminHandlers provides BI-enhanced handlers for admin pages
type BIAdminHandlers struct {
	db            *gorm.DB
	scoringEngine *services.BehavioralScoringEngine
	fubSync       *services.FUBBidirectionalSync
}

// NewBIAdminHandlers creates new BI admin handlers
func NewBIAdminHandlers(db *gorm.DB, fubAPIKey string) *BIAdminHandlers {
	return &BIAdminHandlers{
		db:            db,
		scoringEngine: services.NewBehavioralScoringEngine(db),
		fubSync:       services.NewFUBBidirectionalSync(db, fubAPIKey),
	}
}

// ============================================================================
// ADMIN DASHBOARD (Enhanced with BI)
// ============================================================================

func (h *BIAdminHandlers) GetAdminDashboard(c *gin.Context) {
	// Get BI metrics
	var hotLeadsCount int64
	h.db.Model(&models.BehavioralScore{}).Where("segment = ?", "hot").Count(&hotLeadsCount)

	var warmLeadsCount int64
	h.db.Model(&models.BehavioralScore{}).Where("segment = ?", "warm").Count(&warmLeadsCount)

	var stalledLeadsCount int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	h.db.Model(&models.BehavioralScore{}).
		Where("last_activity_at < ? AND segment != ?", sevenDaysAgo, "dormant").
		Count(&stalledLeadsCount)

	// Get trending up count (score increased in last 7 days)
	var trendingUpCount int64
	h.db.Raw(`
		SELECT COUNT(DISTINCT lead_id) 
		FROM behavioral_scores bs1
		WHERE EXISTS (
			SELECT 1 FROM behavioral_events be
			WHERE be.lead_id = bs1.lead_id
			AND be.created_at > NOW() - INTERVAL '7 days'
		)
		AND bs1.overall_score > 50
	`).Scan(&trendingUpCount)

	// Get average score
	var avgScore float64
	h.db.Model(&models.BehavioralScore{}).Select("AVG(overall_score)").Scan(&avgScore)

	// Get recent activity
	var recentEvents []models.BehavioralEvent
	h.db.Order("created_at DESC").Limit(10).
		Preload("Lead").
		Find(&recentEvents)

	c.HTML(http.StatusOK, "admin-dashboard.html", gin.H{
		"hotLeadsCount":     hotLeadsCount,
		"warmLeadsCount":    warmLeadsCount,
		"stalledLeadsCount": stalledLeadsCount,
		"trendingUpCount":   trendingUpCount,
		"avgScore":          int(avgScore),
		"recentEvents":      recentEvents,
	})
}

// ============================================================================
// COMMUNICATION CENTER (Enhanced with BI Priority)
// ============================================================================

func (h *BIAdminHandlers) GetCommunicationCenter(c *gin.Context) {
	// Get all leads with their behavioral scores
	var leadsWithScores []struct {
		models.Lead
		Score   *models.BehavioralScore
		Segment string
	}

	h.db.Table("leads").
		Select("leads.*, behavioral_scores.overall_score, behavioral_scores.segment, behavioral_scores.last_activity_at").
		Joins("LEFT JOIN behavioral_scores ON leads.id = behavioral_scores.lead_id").
		Order("CASE behavioral_scores.segment WHEN 'hot' THEN 1 WHEN 'warm' THEN 2 WHEN 'cold' THEN 3 ELSE 4 END, behavioral_scores.overall_score DESC").
		Scan(&leadsWithScores)

	c.HTML(http.StatusOK, "communication-center.html", gin.H{
		"leads": leadsWithScores,
	})
}

// ============================================================================
// PROPERTY PERFORMANCE (Enhanced with BI Engagement)
// ============================================================================

func (h *BIAdminHandlers) GetPropertyPerformance(c *gin.Context) {
	// Get all properties with engagement metrics
	type PropertyMetrics struct {
		PropertyID     int64
		Address        string
		TotalViews     int
		HotLeadViews   int
		WarmLeadViews  int
		Saves          int
		Applications   int
		AvgLeadScore   float64
		ConversionRate float64
	}

	var metrics []PropertyMetrics
	h.db.Raw(`
		SELECT 
			p.id as property_id,
			p.address,
			COUNT(DISTINCT CASE WHEN be.event_type = 'viewed' THEN be.id END) as total_views,
			COUNT(DISTINCT CASE WHEN be.event_type = 'viewed' AND bs.segment = 'hot' THEN be.id END) as hot_lead_views,
			COUNT(DISTINCT CASE WHEN be.event_type = 'viewed' AND bs.segment = 'warm' THEN be.id END) as warm_lead_views,
			COUNT(DISTINCT CASE WHEN be.event_type = 'saved' THEN be.id END) as saves,
			COUNT(DISTINCT CASE WHEN be.event_type = 'applied' THEN be.id END) as applications,
			AVG(bs.overall_score) as avg_lead_score,
			CASE 
				WHEN COUNT(DISTINCT CASE WHEN be.event_type = 'viewed' THEN be.lead_id END) > 0
				THEN (COUNT(DISTINCT CASE WHEN be.event_type = 'applied' THEN be.lead_id END)::float / 
					  COUNT(DISTINCT CASE WHEN be.event_type = 'viewed' THEN be.lead_id END)::float * 100)
				ELSE 0
			END as conversion_rate
		FROM properties p
		LEFT JOIN behavioral_events be ON p.id = be.property_id
		LEFT JOIN behavioral_scores bs ON be.lead_id = bs.lead_id
		WHERE p.status = 'active'
		GROUP BY p.id, p.address
		ORDER BY total_views DESC
	`).Scan(&metrics)

	c.HTML(http.StatusOK, "property-performance.html", gin.H{
		"properties": metrics,
	})
}

// ============================================================================
// APPLICATION WORKFLOW (Enhanced with BI Context)
// ============================================================================

func (h *BIAdminHandlers) GetApplicationWorkflow(c *gin.Context) {
	// Get all applications with behavioral context
	type ApplicationWithBI struct {
		ApplicationID   int64
		LeadID          int64
		LeadName        string
		PropertyAddress string
		Status          string
		Score           int
		Segment         string
		TotalEvents     int
		PropertyViews   int
		LastActivity    time.Time
		SubmittedAt     time.Time
	}

	var applications []ApplicationWithBI
	h.db.Raw(`
		SELECT 
			a.id as application_id,
			a.lead_id,
			l.first_name || ' ' || l.last_name as lead_name,
			p.address as property_address,
			a.status,
			COALESCE(bs.overall_score, 0) as score,
			COALESCE(bs.segment, 'unknown') as segment,
			COALESCE(bs.total_events, 0) as total_events,
			COUNT(DISTINCT CASE WHEN be.event_type = 'viewed' THEN be.id END) as property_views,
			bs.last_activity_at as last_activity,
			a.created_at as submitted_at
		FROM applications a
		JOIN leads l ON a.lead_id = l.id
		JOIN properties p ON a.property_id = p.id
		LEFT JOIN behavioral_scores bs ON a.lead_id = bs.lead_id
		LEFT JOIN behavioral_events be ON a.lead_id = be.lead_id
		GROUP BY a.id, a.lead_id, l.first_name, l.last_name, p.address, a.status, 
				 bs.overall_score, bs.segment, bs.total_events, bs.last_activity_at, a.created_at
		ORDER BY a.created_at DESC
	`).Scan(&applications)

	c.HTML(http.StatusOK, "application-workflow.html", gin.H{
		"applications": applications,
	})
}

// ============================================================================
// BEHAVIORAL INTELLIGENCE DASHBOARD (Real Data)
// ============================================================================

func (h *BIAdminHandlers) GetBehavioralIntelligenceDashboard(c *gin.Context) {
	// Get segment breakdown
	type SegmentCount struct {
		Segment string
		Count   int64
	}
	var segments []SegmentCount
	h.db.Model(&models.BehavioralScore{}).
		Select("segment, COUNT(*) as count").
		Group("segment").
		Scan(&segments)

	// Get top performers
	var topLeads []struct {
		models.Lead
		Score models.BehavioralScore
	}
	h.db.Table("leads").
		Select("leads.*, behavioral_scores.*").
		Joins("JOIN behavioral_scores ON leads.id = behavioral_scores.lead_id").
		Order("behavioral_scores.overall_score DESC").
		Limit(20).
		Scan(&topLeads)

	// Get stalled leads
	var stalledLeads []struct {
		models.Lead
		Score     models.BehavioralScore
		DaysSince int
	}
	h.db.Raw(`
		SELECT 
			l.*,
			bs.*,
			EXTRACT(DAY FROM NOW() - bs.last_activity_at)::int as days_since
		FROM leads l
		JOIN behavioral_scores bs ON l.id = bs.lead_id
		WHERE bs.last_activity_at < NOW() - INTERVAL '7 days'
		AND bs.segment != 'dormant'
		ORDER BY bs.last_activity_at ASC
		LIMIT 20
	`).Scan(&stalledLeads)

	// Get trending up
	var trendingLeads []struct {
		models.Lead
		Score        models.BehavioralScore
		RecentEvents int
	}
	h.db.Raw(`
		SELECT 
			l.*,
			bs.*,
			COUNT(be.id)::int as recent_events
		FROM leads l
		JOIN behavioral_scores bs ON l.id = bs.lead_id
		JOIN behavioral_events be ON l.id = be.lead_id
		WHERE be.created_at > NOW() - INTERVAL '7 days'
		GROUP BY l.id, bs.id
		HAVING COUNT(be.id) >= 3
		ORDER BY recent_events DESC
		LIMIT 20
	`).Scan(&trendingLeads)

	// Get average score
	var avgScore float64
	h.db.Model(&models.BehavioralScore{}).Select("AVG(overall_score)").Scan(&avgScore)

	// Get conversion rate (applications / total leads with scores)
	var totalWithScores int64
	var totalApplications int64
	h.db.Model(&models.BehavioralScore{}).Count(&totalWithScores)
	h.db.Model(&models.BehavioralEvent{}).Where("event_type = ?", "applied").
		Distinct("lead_id").Count(&totalApplications)

	conversionRate := 0.0
	if totalWithScores > 0 {
		conversionRate = float64(totalApplications) / float64(totalWithScores) * 100
	}

	c.HTML(http.StatusOK, "behavioral-intelligence.html", gin.H{
		"segments":       segments,
		"topLeads":       topLeads,
		"stalledLeads":   stalledLeads,
		"trendingLeads":  trendingLeads,
		"avgScore":       int(avgScore),
		"conversionRate": int(conversionRate),
	})
}

// ============================================================================
// FUB SYNC API ENDPOINTS
// ============================================================================

func (h *BIAdminHandlers) LogCallToFUB(c *gin.Context) {
	var req struct {
		LeadID   int64  `json:"lead_id"`
		Duration int    `json:"duration"`
		Notes    string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := c.GetString("user_id") // Assuming auth middleware sets this

	if err := h.fubSync.LogCallToFUB(req.LeadID, req.Duration, req.Notes, agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BIAdminHandlers) LogEmailToFUB(c *gin.Context) {
	var req struct {
		LeadID  int64  `json:"lead_id"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := c.GetString("user_id")

	if err := h.fubSync.LogEmailToFUB(req.LeadID, req.Subject, req.Body, agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BIAdminHandlers) LogSMSToFUB(c *gin.Context) {
	var req struct {
		LeadID  int64  `json:"lead_id"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := c.GetString("user_id")

	if err := h.fubSync.LogSMSToFUB(req.LeadID, req.Message, agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BIAdminHandlers) HandleFUBWebhook(c *gin.Context) {
	var webhookData map[string]interface{}
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.fubSync.HandleFUBWebhook(webhookData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================================================
// ROUTE REGISTRATION
// ============================================================================

func RegisterBIAdminRoutes(r *gin.Engine, db *gorm.DB, fubAPIKey string) {
	h := NewBIAdminHandlers(db, fubAPIKey)

	// Admin pages
	admin := r.Group("/admin")
	{
		admin.GET("/dashboard", h.GetAdminDashboard)
		admin.GET("/communication-center", h.GetCommunicationCenter)
		admin.GET("/property-performance", h.GetPropertyPerformance)
		admin.GET("/application-workflow", h.GetApplicationWorkflow)
		admin.GET("/behavioral-intelligence", h.GetBehavioralIntelligenceDashboard)
	}

	// FUB sync API
	api := r.Group("/api/v1/fub")
	{
		api.POST("/log-call", h.LogCallToFUB)
		api.POST("/log-email", h.LogEmailToFUB)
		api.POST("/log-sms", h.LogSMSToFUB)
		api.POST("/webhook", h.HandleFUBWebhook)
	}
}
