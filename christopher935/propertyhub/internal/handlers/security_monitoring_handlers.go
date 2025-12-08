package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SecurityMonitoringHandlers handles security event monitoring and analysis
type SecurityMonitoringHandlers struct {
	db *gorm.DB
}

// NewSecurityMonitoringHandlers creates new security monitoring handlers
func NewSecurityMonitoringHandlers(db *gorm.DB) *SecurityMonitoringHandlers {
	return &SecurityMonitoringHandlers{db: db}
}

// SecurityEventData represents incoming security event data
type SecurityEventData struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	UserID      *uint  `json:"user_id,omitempty"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	StatusCode  int    `json:"status_code"`
	Metadata    string `json:"metadata"` // Raw JSON string
	Description string `json:"description"`
}

// GetSecurityEvents handles GET /api/v1/security/events
func (h *SecurityMonitoringHandlers) GetSecurityEvents(c *gin.Context) {
	var events []SecurityEvent

	query := h.db.Order("timestamp DESC")

	// Apply filters
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}

	if eventType := c.Query("type"); eventType != "" {
		query = query.Where("type = ?", eventType)
	}

	if resolved := c.Query("resolved"); resolved != "" {
		if resolvedBool, err := strconv.ParseBool(resolved); err == nil {
			query = query.Where("resolved = ?", resolvedBool)
		}
	}

	// Date range filtering
	if since := c.Query("since"); since != "" {
		if sinceTime, err := time.Parse(time.RFC3339, since); err == nil {
			query = query.Where("timestamp >= ?", sinceTime)
		}
	}

	// Pagination
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	query = query.Limit(limit).Offset(offset)

	result := query.Find(&events)

	// Handle empty results gracefully
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch security events"})
		return
	}

	// Return empty array if no events
	if len(events) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"events": []SecurityEvent{},
			"total":  0,
		})
		return
	}

	// Get total count for pagination
	var totalCount int64
	countQuery := h.db.Model(&SecurityEvent{})
	if severity := c.Query("severity"); severity != "" {
		countQuery = countQuery.Where("severity = ?", severity)
	}
	if eventType := c.Query("type"); eventType != "" {
		countQuery = countQuery.Where("type = ?", eventType)
	}
	countQuery.Count(&totalCount)

	c.JSON(http.StatusOK, gin.H{
		"events":   events,
		"total":    totalCount,
		"limit":    limit,
		"offset":   offset,
		"has_more": int64(offset+limit) < totalCount,
	})
}

// CreateSecurityEvent handles POST /api/v1/security/events
func (h *SecurityMonitoringHandlers) CreateSecurityEvent(c *gin.Context) {
	var eventData SecurityEventData

	if err := c.ShouldBindJSON(&eventData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if eventData.Type == "" || eventData.Severity == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type and severity are required"})
		return
	}

	// Parse metadata JSON string to map
	var metadata map[string]interface{}
	if eventData.Metadata != "" {
		if err := json.Unmarshal([]byte(eventData.Metadata), &metadata); err != nil {
			// If JSON parsing fails, store as simple string in map
			metadata = map[string]interface{}{
				"raw_data": eventData.Metadata,
			}
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Create security event
	event := SecurityEvent{
		Type:       eventData.Type,
		Severity:   eventData.Severity,
		IPAddress:  eventData.IPAddress,
		UserAgent:  eventData.UserAgent,
		Endpoint:   eventData.Endpoint,
		Method:     eventData.Method,
		StatusCode: eventData.StatusCode,
		Metadata:   metadata,
		Timestamp:  time.Now(),
		Resolved:   false,
	}

	// Handle UserID pointer properly
	if eventData.UserID != nil {
		event.UserID = eventData.UserID
	}

	if err := h.db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create security event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"event": event})
}

// ResolveSecurityEvent handles POST /api/v1/security/events/:id/resolve
func (h *SecurityMonitoringHandlers) ResolveSecurityEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	var request struct {
		ResolvedBy uint   `json:"resolved_by" binding:"required"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	now := time.Now()
	result := h.db.Model(&SecurityEvent{}).Where("id = ? AND resolved = false", eventID).Updates(map[string]interface{}{
		"resolved":    true,
		"resolved_by": request.ResolvedBy,
		"resolved_at": &now,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve security event"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Security event not found or already resolved"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Security event resolved"})
}

// GetSecurityMetrics handles GET /api/v1/security/metrics
// Returns comprehensive security metrics including:
// - severity_breakdown: Event counts by severity level
// - type_breakdown: Event counts by event type
// - resolution_stats: Total, resolved, and unresolved event counts
// - top_ips: Top IP addresses by event count
// - threat_analysis: Threat-specific metrics (threat levels, blocked requests, top threat sources)
// Query parameters:
// - since: Start time for metrics (RFC3339 format, default: 24 hours ago)
// - filter: Optional filter ("threat_analysis" to filter by threat events only)
func (h *SecurityMonitoringHandlers) GetSecurityMetrics(c *gin.Context) {
	since := time.Now().Add(-24 * time.Hour)
	if sinceStr := c.Query("since"); sinceStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsedTime
		}
	}

	filter := c.Query("filter")

	baseQuery := h.db.Model(&SecurityEvent{}).Where("timestamp >= ?", since)
	if filter == "threat_analysis" {
		baseQuery = baseQuery.Where("type = ?", "THREAT_ANALYSIS")
	}

	var severityStats []struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}

	baseQuery.Session(&gorm.Session{}).
		Select("severity, COUNT(*) as count").
		Group("severity").
		Scan(&severityStats)

	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	h.db.Model(&SecurityEvent{}).
		Select("type, COUNT(*) as count").
		Where("timestamp >= ?", since).
		Group("type").
		Order("count DESC").
		Limit(10).
		Scan(&typeStats)

	var resolutionStats struct {
		Total      int64 `json:"total"`
		Resolved   int64 `json:"resolved"`
		Unresolved int64 `json:"unresolved"`
	}

	h.db.Model(&SecurityEvent{}).Where("timestamp >= ?", since).Count(&resolutionStats.Total)
	h.db.Model(&SecurityEvent{}).Where("timestamp >= ? AND resolved = true", since).Count(&resolutionStats.Resolved)
	resolutionStats.Unresolved = resolutionStats.Total - resolutionStats.Resolved

	var ipStats []struct {
		IPAddress string `json:"ip_address"`
		Count     int64  `json:"count"`
	}

	h.db.Model(&SecurityEvent{}).
		Select("ip_address, COUNT(*) as count").
		Where("timestamp >= ? AND ip_address != ''", since).
		Group("ip_address").
		Order("count DESC").
		Limit(10).
		Scan(&ipStats)

	var blockedCount int64
	h.db.Model(&SecurityEvent{}).
		Where("type = ? AND timestamp >= ? AND metadata->>'action_taken' = ?", "THREAT_ANALYSIS", since, "BLOCK_REQUEST").
		Count(&blockedCount)

	var threatStats []struct {
		RiskLevel string `json:"risk_level"`
		Count     int64  `json:"count"`
	}

	h.db.Model(&SecurityEvent{}).
		Select("severity as risk_level, COUNT(*) as count").
		Where("type = ? AND timestamp >= ?", "THREAT_ANALYSIS", since).
		Group("severity").
		Scan(&threatStats)

	var topThreats []struct {
		IPAddress string `json:"ip_address"`
		Count     int64  `json:"count"`
	}

	h.db.Model(&SecurityEvent{}).
		Select("ip_address, COUNT(*) as count").
		Where("type = ? AND timestamp >= ?", "THREAT_ANALYSIS", since).
		Group("ip_address").
		Order("count DESC").
		Limit(10).
		Scan(&topThreats)

	response := gin.H{
		"time_range": gin.H{
			"since": since,
			"until": time.Now(),
		},
		"severity_breakdown": severityStats,
		"type_breakdown":     typeStats,
		"resolution_stats":   resolutionStats,
		"top_ips":            ipStats,
		"threat_analysis": gin.H{
			"threat_levels":      threatStats,
			"blocked_requests":   blockedCount,
			"top_threat_sources": topThreats,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetSecurityEventDetails handles GET /api/v1/security/events/:id
func (h *SecurityMonitoringHandlers) GetSecurityEventDetails(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	var event SecurityEvent
	if err := h.db.First(&event, eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Security event not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch security event"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": event})
}

// CreateSecuritySession handles POST /api/v1/security/sessions
func (h *SecurityMonitoringHandlers) CreateSecuritySession(c *gin.Context) {
	var request struct {
		UserID      uint   `json:"user_id" binding:"required"`
		IPAddress   string `json:"ip_address" binding:"required"`
		UserAgent   string `json:"user_agent"`
		SessionType string `json:"session_type"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Generate session token (simplified)
	sessionToken := generateSessionToken()

	// Create session metadata as proper map
	sessionMetadata := map[string]interface{}{
		"session_token": sessionToken,
		"created_at":    time.Now().Format(time.RFC3339),
		"session_type":  request.SessionType,
	}

	// Create security session record
	session := SecurityEvent{
		Type:       "SESSION_CREATED",
		Severity:   "INFO",
		UserID:     &request.UserID,
		IPAddress:  request.IPAddress,
		UserAgent:  request.UserAgent,
		Endpoint:   "/security/sessions",
		Method:     "POST",
		StatusCode: 201,
		Metadata:   sessionMetadata,
		Timestamp:  time.Now(),
		Resolved:   true, // Sessions are automatically "resolved"
	}

	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create security session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_token": sessionToken,
		"session_id":    session.ID,
		"expires_at":    time.Now().Add(24 * time.Hour),
	})
}

// generateSessionToken generates a simple session token
func generateSessionToken() string {
	// Simplified token generation - in production, use crypto/rand
	return "sess_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// RegisterSecurityMonitoringRoutes registers security monitoring routes
func RegisterSecurityMonitoringRoutes(r *gin.Engine, db *gorm.DB) {
	handlers := NewSecurityMonitoringHandlers(db)

	securityGroup := r.Group("/api/v1/security")
	{
		// Security events
		securityGroup.GET("/events", handlers.GetSecurityEvents)
		securityGroup.POST("/events", handlers.CreateSecurityEvent)
		securityGroup.GET("/events/:id", handlers.GetSecurityEventDetails)
		securityGroup.POST("/events/:id/resolve", handlers.ResolveSecurityEvent)

		// Security metrics and monitoring
		securityGroup.GET("/metrics", handlers.GetSecurityMetrics)
		securityGroup.POST("/sessions", handlers.CreateSecuritySession)
	}
}
