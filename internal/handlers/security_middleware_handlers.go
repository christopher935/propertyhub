package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SecurityMiddlewareHandlers handles advanced security middleware operations
type SecurityMiddlewareHandlers struct {
	db          *gorm.DB
	authManager auth.AuthenticationManager
	redis       *redis.Client
}

// NewSecurityMiddlewareHandlers creates new security middleware handlers
func NewSecurityMiddlewareHandlers(db *gorm.DB, authManager auth.AuthenticationManager) *SecurityMiddlewareHandlers {
	return &SecurityMiddlewareHandlers{
		db:          db,
		authManager: authManager,
	}
}

// SecurityMiddleware provides comprehensive security monitoring
func (h *SecurityMiddlewareHandlers) SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Pre-request security checks
		if blocked := h.checkSecurityBlocks(c); blocked {
			return
		}

		// Rate limiting check
		if limited := h.checkRateLimit(c); limited {
			return
		}

		// Process request
		c.Next()

		// Post-request security logging
		h.logSecurityEvent(c, startTime)
	}
}

// checkSecurityBlocks checks for blocked IPs and suspicious activity
func (h *SecurityMiddlewareHandlers) checkSecurityBlocks(c *gin.Context) bool {
	clientIP := c.ClientIP()

	// Check IP blacklist
	if h.isIPBlacklisted(clientIP) {
		h.createSecurityEvent("IP_BLOCKED", "HIGH", clientIP, c)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		c.Abort()
		return true
	}

	// Check for suspicious patterns
	if h.detectSuspiciousActivity(c) {
		h.createSecurityEvent("SUSPICIOUS_ACTIVITY", "MEDIUM", clientIP, c)
	}

	return false
}

// checkRateLimit implements rate limiting
func (h *SecurityMiddlewareHandlers) checkRateLimit(c *gin.Context) bool {
	if h.redis == nil {
		return false
	}

	clientIP := c.ClientIP()
	key := "rate_limit:" + clientIP

	ctx := context.Background()
	count, err := h.redis.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("Rate limit check failed: %v", err)
		return false
	}

	if count == 1 {
		h.redis.Expire(ctx, key, time.Minute)
	}

	// 100 requests per minute limit
	if count > 100 {
		h.createSecurityEvent("RATE_LIMIT_EXCEEDED", "MEDIUM", clientIP, c)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		c.Abort()
		return true
	}

	return false
}

// logSecurityEvent logs security events
func (h *SecurityMiddlewareHandlers) logSecurityEvent(c *gin.Context, startTime time.Time) {
	statusCode := c.Writer.Status()

	// Log suspicious status codes
	if statusCode >= 400 {
		severity := "LOW"
		if statusCode >= 500 {
			severity = "HIGH"
		} else if statusCode == 401 || statusCode == 403 {
			severity = "MEDIUM"
		}

		h.createSecurityEvent("HTTP_ERROR", severity, c.ClientIP(), c)
	}
}

// createSecurityEvent creates and stores a security event
// Note: Uses SecurityEvent type from advanced_security_api_handlers.go
func (h *SecurityMiddlewareHandlers) createSecurityEvent(eventType, severity, ip string, c *gin.Context) {
	event := SecurityEvent{
		Type:       eventType,
		Severity:   severity,
		IPAddress:  ip,
		UserAgent:  c.GetHeader("User-Agent"),
		Endpoint:   c.Request.URL.Path,
		Method:     c.Request.Method,
		StatusCode: c.Writer.Status(),
		Metadata: map[string]interface{}{
			"headers": h.sanitizeHeaders(c.Request.Header),
			"query":   c.Request.URL.RawQuery,
		},
		Timestamp: time.Now(),
		Resolved:  false,
	}

	// Store in database
	if err := h.db.Create(&event).Error; err != nil {
		log.Printf("Failed to create security event: %v", err)
	}
}

// isIPBlacklisted checks if IP is blacklisted
func (h *SecurityMiddlewareHandlers) isIPBlacklisted(ip string) bool {
	// Implementation would check against blacklist in database/Redis
	// For now, return false
	return false
}

// detectSuspiciousActivity detects suspicious request patterns
func (h *SecurityMiddlewareHandlers) detectSuspiciousActivity(c *gin.Context) bool {
	userAgent := c.GetHeader("User-Agent")

	// Check for suspicious user agents
	suspiciousAgents := []string{"sqlmap", "nikto", "nmap", "masscan"}
	for _, agent := range suspiciousAgents {
		if len(userAgent) > 0 && userAgent == agent {
			return true
		}
	}

	// Check for SQL injection patterns in query parameters
	for _, values := range c.Request.URL.Query() {
		for _, value := range values {
			if h.containsSQLInjection(value) {
				return true
			}
		}
	}

	return false
}

// containsSQLInjection checks for SQL injection patterns
func (h *SecurityMiddlewareHandlers) containsSQLInjection(input string) bool {
	patterns := []string{"' OR '1'='1", "UNION SELECT", "DROP TABLE", "INSERT INTO"}
	for _, pattern := range patterns {
		if len(input) > len(pattern) && input[:len(pattern)] == pattern {
			return true
		}
	}
	return false
}

// sanitizeHeaders removes sensitive headers for logging
func (h *SecurityMiddlewareHandlers) sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)

	for key, values := range headers {
		if key != "Authorization" && key != "Cookie" {
			if len(values) > 0 {
				sanitized[key] = values[0]
			}
		}
	}

	return sanitized
}

// RegisterSecurityMiddlewareRoutes registers security middleware routes
func RegisterSecurityMiddlewareRoutes(mux *http.ServeMux, db *gorm.DB, authManager auth.AuthenticationManager) {
	handlers := NewSecurityMiddlewareHandlers(db, authManager)

	mux.HandleFunc("/security/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.handleGetSecurityEvents(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/security/events/resolve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.handleResolveSecurityEvent(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// handleGetSecurityEvents handles GET /security/events
func (h *SecurityMiddlewareHandlers) handleGetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	var events []SecurityEvent

	query := h.db.Order("timestamp DESC")

	// Filter by severity if provided
	if severity := r.URL.Query().Get("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// Filter by resolved status
	if resolved := r.URL.Query().Get("resolved"); resolved != "" {
		if resolvedBool, err := strconv.ParseBool(resolved); err == nil {
			query = query.Where("resolved = ?", resolvedBool)
		}
	}

	// Pagination
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	query = query.Limit(limit)

	if err := query.Find(&events).Error; err != nil {
		http.Error(w, "Failed to fetch events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gin.H{"events": events})
}

// handleResolveSecurityEvent handles POST /security/events/resolve
func (h *SecurityMiddlewareHandlers) handleResolveSecurityEvent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EventID    uint `json:"event_id" binding:"required"`
		ResolvedBy uint `json:"resolved_by" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	now := time.Now()
	err := h.db.Model(&SecurityEvent{}).Where("id = ?", request.EventID).Updates(map[string]interface{}{
		"resolved":    true,
		"resolved_by": request.ResolvedBy,
		"resolved_at": &now,
	}).Error

	if err != nil {
		http.Error(w, "Failed to resolve event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gin.H{"success": true})
}
