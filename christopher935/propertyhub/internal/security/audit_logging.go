package security

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AuditLogger handles comprehensive audit logging
type AuditLogger struct {
	db *gorm.DB
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *gorm.DB) *AuditLogger {
	return &AuditLogger{
		db: db,
	}
}

// AuditLog represents a comprehensive audit log entry
type AuditLog struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	Timestamp    time.Time              `json:"timestamp" gorm:"index"`
	UserID       *uint                  `json:"user_id" gorm:"index"`
	SessionID    string                 `json:"session_id" gorm:"index"`
	IPAddress    string                 `json:"ip_address" gorm:"index"`
	UserAgent    string                 `json:"user_agent"`
	Action       string                 `json:"action" gorm:"index"`
	Resource     string                 `json:"resource" gorm:"index"`
	ResourceID   *uint                  `json:"resource_id" gorm:"index"`
	Method       string                 `json:"method"`
	Endpoint     string                 `json:"endpoint"`
	StatusCode   int                    `json:"status_code"`
	Success      bool                   `json:"success" gorm:"index"`
	ErrorMessage string                 `json:"error_message"`
	RequestData  map[string]interface{} `json:"request_data" gorm:"type:jsonb"`
	ResponseData map[string]interface{} `json:"response_data" gorm:"type:jsonb"`
	Duration     int64                  `json:"duration"`              // milliseconds
	Severity     string                 `json:"severity" gorm:"index"` // "low", "medium", "high", "critical"
	Category     string                 `json:"category" gorm:"index"` // "auth", "data", "admin", "security", "system"
	Tags         []string               `json:"tags" gorm:"type:jsonb"`
	Metadata     map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt    time.Time              `json:"created_at"`
}

// SecurityEvent represents a security-specific audit event
type SecurityEvent struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	Timestamp   time.Time              `json:"timestamp" gorm:"index"`
	EventType   string                 `json:"event_type" gorm:"index"` // "login_attempt", "mfa_failure", "suspicious_activity", etc.
	UserID      *uint                  `json:"user_id" gorm:"index"`
	IPAddress   string                 `json:"ip_address" gorm:"index"`
	UserAgent   string                 `json:"user_agent"`
	Severity    string                 `json:"severity" gorm:"index"` // "low", "medium", "high", "critical"
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details" gorm:"type:jsonb"`
	RiskScore   int                    `json:"risk_score"` // 0-100
	Resolved    bool                   `json:"resolved" gorm:"default:false"`
	ResolvedAt  *time.Time             `json:"resolved_at"`
	ResolvedBy  *uint                  `json:"resolved_by"`
	Actions     []string               `json:"actions" gorm:"type:jsonb"` // Actions taken in response
	CreatedAt   time.Time              `json:"created_at"`
}

// DataAccessLog tracks data access for compliance
type DataAccessLog struct {
	ID             uint                   `json:"id" gorm:"primaryKey"`
	Timestamp      time.Time              `json:"timestamp" gorm:"index"`
	UserID         uint                   `json:"user_id" gorm:"index"`
	TableName      string                 `json:"table_name" gorm:"index"`
	RecordID       uint                   `json:"record_id" gorm:"index"`
	Operation      string                 `json:"operation" gorm:"index"` // "read", "create", "update", "delete"
	FieldsAccessed []string               `json:"fields_accessed" gorm:"type:jsonb"`
	OldValues      map[string]interface{} `json:"old_values" gorm:"type:jsonb"`
	NewValues      map[string]interface{} `json:"new_values" gorm:"type:jsonb"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	Purpose        string                 `json:"purpose"` // Business purpose for access
	CreatedAt      time.Time              `json:"created_at"`
}

// AdminAction tracks administrative actions
type AdminAction struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	Timestamp   time.Time              `json:"timestamp" gorm:"index"`
	AdminID     uint                   `json:"admin_id" gorm:"index"`
	Action      string                 `json:"action" gorm:"index"`
	TargetType  string                 `json:"target_type"` // "user", "property", "booking", "system"
	TargetID    *uint                  `json:"target_id"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details" gorm:"type:jsonb"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Success     bool                   `json:"success"`
	ErrorMsg    string                 `json:"error_msg"`
	CreatedAt   time.Time              `json:"created_at"`
}

// LogAction logs a general action
func (al *AuditLogger) LogAction(userID *uint, sessionID, ipAddress, userAgent, action, resource string, resourceID *uint, success bool, errorMsg string, metadata map[string]interface{}) {
	auditLog := &AuditLog{
		Timestamp:    time.Now(),
		UserID:       userID,
		SessionID:    sessionID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
		Success:      success,
		ErrorMessage: errorMsg,
		Severity:     al.determineSeverity(action, success),
		Category:     al.determineCategory(action, resource),
		Metadata:     metadata,
	}

	al.db.Create(auditLog)
}

// LogHTTPRequest logs an HTTP request with full context
func (al *AuditLogger) LogHTTPRequest(r *http.Request, userID *uint, sessionID string, statusCode int, duration time.Duration, responseData map[string]interface{}) {
	// Extract request data (sanitized)
	requestData := al.sanitizeRequestData(r)

	auditLog := &AuditLog{
		Timestamp:    time.Now(),
		UserID:       userID,
		SessionID:    sessionID,
		IPAddress:    al.getClientIP(r),
		UserAgent:    r.Header.Get("User-Agent"),
		Action:       fmt.Sprintf("%s %s", r.Method, r.URL.Path),
		Resource:     al.extractResourceFromPath(r.URL.Path),
		Method:       r.Method,
		Endpoint:     r.URL.Path,
		StatusCode:   statusCode,
		Success:      statusCode < 400,
		RequestData:  requestData,
		ResponseData: responseData,
		Duration:     duration.Milliseconds(),
		Severity:     al.determineSeverityFromStatus(statusCode),
		Category:     al.determineCategoryFromPath(r.URL.Path),
	}

	if statusCode >= 400 {
		auditLog.ErrorMessage = fmt.Sprintf("HTTP %d", statusCode)
	}

	al.db.Create(auditLog)
}

// LogSecurityEvent logs a security-specific event
func (al *AuditLogger) LogSecurityEvent(eventType string, userID *uint, ipAddress, userAgent, description string, details map[string]interface{}, riskScore int) {
	severity := "low"
	if riskScore >= 80 {
		severity = "critical"
	} else if riskScore >= 60 {
		severity = "high"
	} else if riskScore >= 40 {
		severity = "medium"
	}

	securityEvent := &SecurityEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		UserID:      userID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    severity,
		Description: description,
		Details:     details,
		RiskScore:   riskScore,
	}

	al.db.Create(securityEvent)
}

// LogDataAccess logs data access for compliance
func (al *AuditLogger) LogDataAccess(userID uint, tableName string, recordID uint, operation string, fieldsAccessed []string, oldValues, newValues map[string]interface{}, ipAddress, userAgent, purpose string) {
	dataAccessLog := &DataAccessLog{
		Timestamp:      time.Now(),
		UserID:         userID,
		TableName:      tableName,
		RecordID:       recordID,
		Operation:      operation,
		FieldsAccessed: fieldsAccessed,
		OldValues:      oldValues,
		NewValues:      newValues,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		Purpose:        purpose,
	}

	al.db.Create(dataAccessLog)
}

// LogAdminAction logs administrative actions
func (al *AuditLogger) LogAdminAction(adminID uint, action, targetType string, targetID *uint, description string, details map[string]interface{}, ipAddress, userAgent string, success bool, errorMsg string) {
	adminAction := &AdminAction{
		Timestamp:   time.Now(),
		AdminID:     adminID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Description: description,
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Success:     success,
		ErrorMsg:    errorMsg,
	}

	al.db.Create(adminAction)
}

// GetAuditLogs retrieves audit logs with filtering
func (al *AuditLogger) GetAuditLogs(filters map[string]interface{}, limit, offset int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := al.db.Model(&AuditLog{})

	// Apply filters
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"]; ok {
		// Validate and sanitize action input to prevent SQL injection
		actionStr, isString := action.(string)
		if !isString || len(actionStr) > 100 { // Reasonable length limit
			return nil, 0, fmt.Errorf("invalid action filter")
		}
		// Sanitize input by removing SQL control characters
		safeAction := strings.ReplaceAll(strings.ReplaceAll(actionStr, "%", ""), "_", "")
		if len(safeAction) > 0 {
			query = query.Where("action ILIKE ?", "%"+safeAction+"%")
		}
	}
	if resource, ok := filters["resource"]; ok {
		query = query.Where("resource = ?", resource)
	}
	if severity, ok := filters["severity"]; ok {
		query = query.Where("severity = ?", severity)
	}
	if category, ok := filters["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if success, ok := filters["success"]; ok {
		query = query.Where("success = ?", success)
	}
	if startDate, ok := filters["start_date"]; ok {
		query = query.Where("timestamp >= ?", startDate)
	}
	if endDate, ok := filters["end_date"]; ok {
		query = query.Where("timestamp <= ?", endDate)
	}
	if ipAddress, ok := filters["ip_address"]; ok {
		query = query.Where("ip_address = ?", ipAddress)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&logs).Error

	return logs, total, err
}

// GetSecurityEvents retrieves security events with filtering
func (al *AuditLogger) GetSecurityEvents(filters map[string]interface{}, limit, offset int) ([]SecurityEvent, int64, error) {
	var events []SecurityEvent
	var total int64

	query := al.db.Model(&SecurityEvent{})

	// Apply filters
	if eventType, ok := filters["event_type"]; ok {
		query = query.Where("event_type = ?", eventType)
	}
	if severity, ok := filters["severity"]; ok {
		query = query.Where("severity = ?", severity)
	}
	if resolved, ok := filters["resolved"]; ok {
		query = query.Where("resolved = ?", resolved)
	}
	if minRiskScore, ok := filters["min_risk_score"]; ok {
		query = query.Where("risk_score >= ?", minRiskScore)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&events).Error

	return events, total, err
}

// GetAuditStatistics returns audit statistics
func (al *AuditLogger) GetAuditStatistics() map[string]interface{} {
	var totalLogs int64
	al.db.Model(&AuditLog{}).Count(&totalLogs)

	var successfulActions int64
	al.db.Model(&AuditLog{}).Where("success = true").Count(&successfulActions)

	var failedActions int64
	al.db.Model(&AuditLog{}).Where("success = false").Count(&failedActions)

	var securityEvents int64
	al.db.Model(&SecurityEvent{}).Count(&securityEvents)

	var unresolvedSecurityEvents int64
	al.db.Model(&SecurityEvent{}).Where("resolved = false").Count(&unresolvedSecurityEvents)

	var highRiskEvents int64
	al.db.Model(&SecurityEvent{}).Where("risk_score >= 60").Count(&highRiskEvents)

	var recentLogs int64
	al.db.Model(&AuditLog{}).Where("timestamp > ?", time.Now().Add(-24*time.Hour)).Count(&recentLogs)

	var adminActions int64
	al.db.Model(&AdminAction{}).Count(&adminActions)

	successRate := float64(0)
	if totalLogs > 0 {
		successRate = float64(successfulActions) / float64(totalLogs) * 100
	}

	return map[string]interface{}{
		"total_logs":                 totalLogs,
		"successful_actions":         successfulActions,
		"failed_actions":             failedActions,
		"success_rate":               successRate,
		"security_events":            securityEvents,
		"unresolved_security_events": unresolvedSecurityEvents,
		"high_risk_events":           highRiskEvents,
		"recent_logs_24h":            recentLogs,
		"admin_actions":              adminActions,
	}
}

// CleanupOldLogs removes old audit logs based on retention policy
func (al *AuditLogger) CleanupOldLogs(retentionDays int) error {
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	// Clean up audit logs
	if err := al.db.Where("timestamp < ?", cutoff).Delete(&AuditLog{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup audit logs: %w", err)
	}

	// Clean up data access logs
	if err := al.db.Where("timestamp < ?", cutoff).Delete(&DataAccessLog{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup data access logs: %w", err)
	}

	// Keep security events longer (2x retention)
	securityCutoff := time.Now().Add(-time.Duration(retentionDays*2) * 24 * time.Hour)
	if err := al.db.Where("timestamp < ? AND resolved = true", securityCutoff).Delete(&SecurityEvent{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup security events: %w", err)
	}

	return nil
}

// Helper methods

func (al *AuditLogger) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (al *AuditLogger) sanitizeRequestData(r *http.Request) map[string]interface{} {
	data := make(map[string]interface{})

	// Add query parameters
	if len(r.URL.RawQuery) > 0 {
		data["query"] = r.URL.RawQuery
	}

	// Add headers (sanitized)
	headers := make(map[string]string)
	for key, values := range r.Header {
		// Skip sensitive headers
		if strings.ToLower(key) == "authorization" || strings.ToLower(key) == "cookie" {
			headers[key] = "[REDACTED]"
		} else {
			headers[key] = strings.Join(values, ", ")
		}
	}
	data["headers"] = headers

	// Add content length
	data["content_length"] = r.ContentLength

	return data
}

func (al *AuditLogger) extractResourceFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" {
		return parts[2] // e.g., /api/v1/bookings -> bookings
	}
	return "unknown"
}

func (al *AuditLogger) determineSeverity(action string, success bool) string {
	if !success {
		return "medium"
	}

	action = strings.ToLower(action)
	if strings.Contains(action, "delete") || strings.Contains(action, "disable") {
		return "high"
	}
	if strings.Contains(action, "create") || strings.Contains(action, "update") {
		return "medium"
	}
	return "low"
}

func (al *AuditLogger) determineSeverityFromStatus(statusCode int) string {
	if statusCode >= 500 {
		return "high"
	}
	if statusCode >= 400 {
		return "medium"
	}
	return "low"
}

func (al *AuditLogger) determineCategory(action, resource string) string {
	action = strings.ToLower(action)
	resource = strings.ToLower(resource)

	if strings.Contains(action, "login") || strings.Contains(action, "auth") || strings.Contains(action, "mfa") {
		return "auth"
	}
	if strings.Contains(action, "admin") || strings.Contains(resource, "admin") {
		return "admin"
	}
	if strings.Contains(action, "security") || strings.Contains(action, "encrypt") {
		return "security"
	}
	if strings.Contains(resource, "booking") || strings.Contains(resource, "property") {
		return "data"
	}
	return "system"
}

func (al *AuditLogger) determineCategoryFromPath(path string) string {
	path = strings.ToLower(path)
	if strings.Contains(path, "/admin/") {
		return "admin"
	}
	if strings.Contains(path, "/auth/") || strings.Contains(path, "/mfa/") {
		return "auth"
	}
	if strings.Contains(path, "/security/") || strings.Contains(path, "/encryption/") {
		return "security"
	}
	if strings.Contains(path, "/booking") || strings.Contains(path, "/property") {
		return "data"
	}
	return "system"
}

// GetCallerInfo returns information about the calling function
func (al *AuditLogger) GetCallerInfo() string {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}
