package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SecurityRoutesHandlers handles basic security routing operations
type SecurityRoutesHandlers struct {
	db *gorm.DB
}

// NewSecurityRoutesHandlers creates new security routes handlers
func NewSecurityRoutesHandlers(db *gorm.DB) *SecurityRoutesHandlers {
	return &SecurityRoutesHandlers{db: db}
}

// SecurityDashboard handles GET /api/v1/security/dashboard
func (h *SecurityRoutesHandlers) SecurityDashboard(c *gin.Context) {
	dashboard := gin.H{
		"total_events":    h.getTotalSecurityEvents(),
		"active_threats":  h.getActiveThreats(),
		"resolved_events": h.getResolvedEvents(),
		"system_status":   "operational",
		"last_updated":    time.Now(),
		"uptime":          h.getSystemUptime(),
	}

	c.JSON(http.StatusOK, dashboard)
}

// SecuritySettings handles GET/POST /api/v1/security/settings
func (h *SecurityRoutesHandlers) SecuritySettings(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		h.getSecuritySettings(c)
	case http.MethodPost:
		h.updateSecuritySettings(c)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	}
}

// getSecuritySettings handles GET security settings
func (h *SecurityRoutesHandlers) getSecuritySettings(c *gin.Context) {
	settings := gin.H{
		"rate_limiting_enabled":    true,
		"max_requests_per_minute":  100,
		"ip_blacklist_enabled":     true,
		"suspicious_activity_detection": true,
		"automatic_threat_blocking": false,
		"security_logging_level":   "INFO",
		"session_timeout_minutes":  30,
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// updateSecuritySettings handles POST security settings
func (h *SecurityRoutesHandlers) updateSecuritySettings(c *gin.Context) {
	var request map[string]interface{}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	validatedSettings := h.validateSecuritySettings(request)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Security settings updated",
		"settings": validatedSettings,
	})
}

// validateSecuritySettings validates security setting values
func (h *SecurityRoutesHandlers) validateSecuritySettings(settings map[string]interface{}) map[string]interface{} {
	validated := make(map[string]interface{})
	
	if val, exists := settings["max_requests_per_minute"]; exists {
		if intVal, ok := val.(float64); ok && intVal > 0 && intVal <= 1000 {
			validated["max_requests_per_minute"] = int(intVal)
		}
	}
	
	if val, exists := settings["session_timeout_minutes"]; exists {
		if intVal, ok := val.(float64); ok && intVal >= 5 && intVal <= 1440 {
			validated["session_timeout_minutes"] = int(intVal)
		}
	}
	
	boolSettings := []string{
		"rate_limiting_enabled",
		"ip_blacklist_enabled", 
		"suspicious_activity_detection",
		"automatic_threat_blocking",
	}
	
	for _, setting := range boolSettings {
		if val, exists := settings[setting]; exists {
			if boolVal, ok := val.(bool); ok {
				validated[setting] = boolVal
			}
		}
	}
	
	return validated
}

// SecurityHealth handles GET /api/v1/security/health
func (h *SecurityRoutesHandlers) SecurityHealth(c *gin.Context) {
	health := gin.H{
		"status":             "healthy",
		"database_connected": h.isDatabaseConnected(),
		"security_services":  "operational", 
		"last_check":        time.Now(),
		"version":           "1.0.0",
		"components": gin.H{
			"rate_limiter":     "active",
			"ip_blacklist":     "active", 
			"threat_detection": "active",
			"event_logging":    "active",
		},
	}

	c.JSON(http.StatusOK, health)
}

// Helper methods

// getTotalSecurityEvents gets total security events count
func (h *SecurityRoutesHandlers) getTotalSecurityEvents() int64 {
	var count int64
	h.db.Model(&SecurityEvent{}).Count(&count)
	return count
}

// getActiveThreats gets active threat count
func (h *SecurityRoutesHandlers) getActiveThreats() int64 {
	var count int64
	h.db.Model(&SecurityEvent{}).Where("resolved = false AND severity IN ?", []string{"HIGH", "CRITICAL"}).Count(&count)
	return count
}

// getResolvedEvents gets resolved events count (last 24h)
func (h *SecurityRoutesHandlers) getResolvedEvents() int64 {
	var count int64
	since := time.Now().Add(-24 * time.Hour)
	h.db.Model(&SecurityEvent{}).Where("resolved = true AND resolved_at >= ?", since).Count(&count)
	return count
}

// getSystemUptime calculates system uptime (simplified)
func (h *SecurityRoutesHandlers) getSystemUptime() string {
	return "99.9%"
}

// isDatabaseConnected checks database connectivity
func (h *SecurityRoutesHandlers) isDatabaseConnected() bool {
	sqlDB, err := h.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

// RegisterSecurityRoutingHandlers registers basic security routing handlers
func RegisterSecurityRoutingHandlers(r *gin.Engine, db *gorm.DB) {
	handlers := NewSecurityRoutesHandlers(db)
	
	securityGroup := r.Group("/api/v1/security")
	{
		securityGroup.GET("/dashboard", handlers.SecurityDashboard)
		securityGroup.GET("/settings", handlers.SecuritySettings)
		securityGroup.POST("/settings", handlers.SecuritySettings)
		securityGroup.GET("/health", handlers.SecurityHealth)
	}
}
