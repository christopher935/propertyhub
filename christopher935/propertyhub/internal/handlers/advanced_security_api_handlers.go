package handlers

import (
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/security"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdvancedSecurityAPIHandlers handles advanced security API operations
type AdvancedSecurityAPIHandlers struct {
	db              *gorm.DB
	securityManager *security.EncryptionManager
}

// NewAdvancedSecurityAPIHandlers creates new advanced security API handlers
func NewAdvancedSecurityAPIHandlers(db *gorm.DB, securityManager *security.EncryptionManager) *AdvancedSecurityAPIHandlers {
	return &AdvancedSecurityAPIHandlers{
		db:              db,
		securityManager: securityManager,
	}
}

// SecurityThreatRequest represents a security threat analysis request
type SecurityThreatRequest struct {
	IPAddress    string                 `json:"ip_address" binding:"required"`
	UserAgent    string                 `json:"user_agent"`
	Endpoint     string                 `json:"endpoint" binding:"required"`
	Method       string                 `json:"method" binding:"required"`
	Headers      map[string]string      `json:"headers"`
	QueryParams  map[string]string      `json:"query_params"`
	RequestBody  string                 `json:"request_body"`
	ThreatLevel  string                 `json:"threat_level"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SecurityAnalysisResponse represents threat analysis results
type SecurityAnalysisResponse struct {
	ThreatScore     float64                `json:"threat_score"`
	RiskLevel       string                 `json:"risk_level"`
	Recommendations []string               `json:"recommendations"`
	BlockedReasons  []string               `json:"blocked_reasons,omitempty"`
	AllowedReasons  []string               `json:"allowed_reasons,omitempty"`
	Analysis        map[string]interface{} `json:"analysis"`
	ActionTaken     string                 `json:"action_taken"`
	Timestamp       time.Time              `json:"timestamp"`
}

// SecuritySession represents an active security session
type SecuritySession struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	UserID      uint                   `json:"user_id"`
	SessionID   string                 `json:"session_id" gorm:"uniqueIndex"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	StartTime   time.Time              `json:"start_time"`
	LastActivity time.Time             `json:"last_activity"`
	IsActive    bool                   `json:"is_active"`
	ThreatLevel string                 `json:"threat_level"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SecurityEvent represents a security event (centralized definition)
type SecurityEvent struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	UserID      *uint                  `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Endpoint    string                 `json:"endpoint"`
	Method      string                 `json:"method"`
	StatusCode  int                    `json:"status_code"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	ResolvedBy  *uint                  `json:"resolved_by,omitempty"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// AnalyzeThreat handles POST /api/v1/security/analyze-threat
func (h *AdvancedSecurityAPIHandlers) AnalyzeThreat(c *gin.Context) {
	var request SecurityThreatRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	analysis := h.performThreatAnalysis(request)
	h.logSecurityAnalysis(request, analysis)
	
	c.JSON(http.StatusOK, analysis)
}

// performThreatAnalysis conducts comprehensive threat analysis
func (h *AdvancedSecurityAPIHandlers) performThreatAnalysis(request SecurityThreatRequest) SecurityAnalysisResponse {
	analysis := SecurityAnalysisResponse{
		Analysis:    make(map[string]interface{}),
		Timestamp:   time.Now(),
		ThreatScore: 0.0,
	}
	
	ipScore := h.analyzeIPReputation(request.IPAddress)
	analysis.Analysis["ip_reputation"] = ipScore
	
	uaScore := h.analyzeUserAgent(request.UserAgent)
	analysis.Analysis["user_agent_analysis"] = uaScore
	
	patternScore := h.analyzeRequestPatterns(request)
	analysis.Analysis["request_patterns"] = patternScore
	
	headerScore := h.analyzeHeaders(request.Headers)
	analysis.Analysis["header_analysis"] = headerScore
	
	analysis.ThreatScore = (ipScore + uaScore + patternScore + headerScore) / 4.0
	analysis.RiskLevel = h.calculateRiskLevel(analysis.ThreatScore)
	analysis.Recommendations = h.generateRecommendations(analysis.ThreatScore, analysis.Analysis)
	analysis.ActionTaken = h.determineAction(analysis.RiskLevel)
	
	return analysis
}

// analyzeIPReputation analyzes IP reputation
func (h *AdvancedSecurityAPIHandlers) analyzeIPReputation(ipAddress string) float64 {
	score := 0.0
	
	if len(ipAddress) > 0 {
		if ipAddress == "127.0.0.1" || ipAddress == "::1" {
			score = 0.1
		} else {
			score = 0.3
		}
	}
	
	return score
}

// analyzeUserAgent analyzes user agent strings
func (h *AdvancedSecurityAPIHandlers) analyzeUserAgent(userAgent string) float64 {
	if userAgent == "" {
		return 0.7
	}
	
	suspiciousPatterns := []string{"sqlmap", "nikto", "nmap", "masscan", "curl", "wget"}
	for _, pattern := range suspiciousPatterns {
		if len(userAgent) >= len(pattern) && userAgent[:len(pattern)] == pattern {
			return 0.9
		}
	}
	
	return 0.2
}

// analyzeRequestPatterns analyzes request patterns for anomalies
func (h *AdvancedSecurityAPIHandlers) analyzeRequestPatterns(request SecurityThreatRequest) float64 {
	score := 0.0
	
	if request.Method == "DELETE" || request.Method == "PUT" {
		score += 0.2
	}
	
	sensitivePaths := []string{"/admin", "/api", "/config", "/backup"}
	for _, path := range sensitivePaths {
		if len(request.Endpoint) >= len(path) && request.Endpoint[:len(path)] == path {
			score += 0.3
		}
	}
	
	for _, value := range request.QueryParams {
		if h.containsInjectionPatterns(value) {
			score += 0.4
		}
	}
	
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// analyzeHeaders analyzes HTTP headers for suspicious patterns
func (h *AdvancedSecurityAPIHandlers) analyzeHeaders(headers map[string]string) float64 {
	score := 0.0
	
	if _, exists := headers["User-Agent"]; !exists {
		score += 0.2
	}
	
	for key, value := range headers {
		if h.containsInjectionPatterns(value) {
			score += 0.3
		}
		
		if key == "X-Forwarded-For" && len(value) > 100 {
			score += 0.2
		}
	}
	
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// containsInjectionPatterns checks for common injection patterns
func (h *AdvancedSecurityAPIHandlers) containsInjectionPatterns(input string) bool {
	patterns := []string{
		"' OR '1'='1",
		"UNION SELECT",
		"DROP TABLE",
		"<script>",
		"javascript:",
		"../../../",
		"../../../../",
	}
	
	for _, pattern := range patterns {
		if len(input) >= len(pattern) && input[:len(pattern)] == pattern {
			return true
		}
	}
	
	return false
}

// calculateRiskLevel determines risk level from threat score
func (h *AdvancedSecurityAPIHandlers) calculateRiskLevel(score float64) string {
	if score >= 0.8 {
		return "CRITICAL"
	} else if score >= 0.6 {
		return "HIGH"
	} else if score >= 0.4 {
		return "MEDIUM"
	} else if score >= 0.2 {
		return "LOW"
	}
	return "MINIMAL"
}

// generateRecommendations generates security recommendations
func (h *AdvancedSecurityAPIHandlers) generateRecommendations(score float64, analysis map[string]interface{}) []string {
	var recommendations []string
	
	if score >= 0.7 {
		recommendations = append(recommendations, "Consider blocking this request")
		recommendations = append(recommendations, "Increase monitoring for this IP address")
	}
	
	if score >= 0.5 {
		recommendations = append(recommendations, "Apply additional authentication requirements")
		recommendations = append(recommendations, "Log all activities from this source")
	}
	
	if score >= 0.3 {
		recommendations = append(recommendations, "Monitor user behavior patterns")
		recommendations = append(recommendations, "Consider rate limiting")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Continue normal monitoring")
	}
	
	return recommendations
}

// determineAction determines the action to take based on risk level
func (h *AdvancedSecurityAPIHandlers) determineAction(riskLevel string) string {
	switch riskLevel {
	case "CRITICAL":
		return "BLOCK_REQUEST"
	case "HIGH":
		return "REQUIRE_ADDITIONAL_AUTH"
	case "MEDIUM":
		return "ENHANCED_MONITORING"
	case "LOW":
		return "STANDARD_MONITORING"
	default:
		return "ALLOW"
	}
}

// logSecurityAnalysis logs the security analysis
func (h *AdvancedSecurityAPIHandlers) logSecurityAnalysis(request SecurityThreatRequest, analysis SecurityAnalysisResponse) {
	event := SecurityEvent{
		Type:       "THREAT_ANALYSIS",
		Severity:   analysis.RiskLevel,
		IPAddress:  request.IPAddress,
		UserAgent:  request.UserAgent,
		Endpoint:   request.Endpoint,
		Method:     request.Method,
		StatusCode: 200,
		Metadata: map[string]interface{}{
			"threat_score":    analysis.ThreatScore,
			"action_taken":    analysis.ActionTaken,
			"recommendations": analysis.Recommendations,
			"analysis_data":   analysis.Analysis,
		},
		Timestamp: time.Now(),
		Resolved:  true,
	}
	
	h.db.Create(&event)
}

// GetSecurityMetrics handles GET /api/v1/security/metrics/advanced
func (h *AdvancedSecurityAPIHandlers) GetSecurityMetrics(c *gin.Context) {
	since := time.Now().Add(-24 * time.Hour)
	if sinceStr := c.Query("since"); sinceStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsedTime
		}
	}
	
	var threatStats []struct {
		RiskLevel string `json:"risk_level"`
		Count     int64  `json:"count"`
	}
	
	h.db.Model(&SecurityEvent{}).
		Select("severity as risk_level, COUNT(*) as count").
		Where("type = ? AND timestamp >= ?", "THREAT_ANALYSIS", since).
		Group("severity").
		Scan(&threatStats)
	
	var blockedCount int64
	h.db.Model(&SecurityEvent{}).
		Where("type = ? AND timestamp >= ? AND metadata->>'action_taken' = ?", "THREAT_ANALYSIS", since, "BLOCK_REQUEST").
		Count(&blockedCount)
	
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
	
	c.JSON(http.StatusOK, gin.H{
		"time_range": gin.H{
			"since": since,
			"until": time.Now(),
		},
		"threat_levels":     threatStats,
		"blocked_requests":  blockedCount,
		"top_threat_sources": topThreats,
	})
}

// CreateSecuritySession handles POST /api/v1/security/sessions/advanced
func (h *AdvancedSecurityAPIHandlers) CreateSecuritySession(c *gin.Context) {
	var request struct {
		UserID      uint                   `json:"user_id" binding:"required"`
		IPAddress   string                 `json:"ip_address" binding:"required"`
		UserAgent   string                 `json:"user_agent"`
		SessionType string                 `json:"session_type"`
		Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	sessionID := "adv_sess_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	
	session := SecuritySession{
		UserID:       request.UserID,
		SessionID:    sessionID,
		IPAddress:    request.IPAddress,
		UserAgent:    request.UserAgent,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
		ThreatLevel:  "LOW",
		Metadata:     request.Metadata,
	}
	
	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create security session"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"session_id": sessionID,
		"expires_at": time.Now().Add(24 * time.Hour),
		"session":    session,
	})
}

// RegisterAdvancedSecurityRoutes registers advanced security API routes
func RegisterAdvancedSecurityRoutes(r *gin.Engine, db *gorm.DB, securityManager *security.EncryptionManager) {
	handlers := NewAdvancedSecurityAPIHandlers(db, securityManager)
	
	advancedGroup := r.Group("/api/v1/security")
	{
		advancedGroup.POST("/analyze-threat", handlers.AnalyzeThreat)
		advancedGroup.GET("/metrics/advanced", handlers.GetSecurityMetrics)
		advancedGroup.POST("/sessions/advanced", handlers.CreateSecuritySession)
	}
}
