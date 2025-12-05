package security

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RealtimeMonitor provides real-time security monitoring and alerting
type RealtimeMonitor struct {
	db             *gorm.DB
	logger         *log.Logger
	auditLogger    *AuditLogger
	alertChannels  map[string]chan SecurityAlert
	threatPatterns []ThreatPattern
	mutex          sync.RWMutex
	isRunning      bool
	ctx            context.Context
	cancel         context.CancelFunc
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID             uint                   `json:"id" gorm:"primaryKey"`
	AlertType      string                 `json:"alert_type" gorm:"not null"`
	Severity       string                 `json:"severity" gorm:"not null"` // low, medium, high, critical
	Title          string                 `json:"title" gorm:"not null"`
	Description    string                 `json:"description" gorm:"type:text"`
	Source         string                 `json:"source" gorm:"not null"` // system, user, external
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	UserID         string                 `json:"user_id"`
	RawData        map[string]interface{} `json:"raw_data" gorm:"type:json"`
	Acknowledged   bool                   `json:"acknowledged" gorm:"default:false"`
	AcknowledgedBy string                 `json:"acknowledged_by"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at"`
	Resolved       bool                   `json:"resolved" gorm:"default:false"`
	ResolvedBy     string                 `json:"resolved_by"`
	ResolvedAt     *time.Time             `json:"resolved_at"`
	CreatedAt      time.Time              `json:"created_at"`
}

// ThreatPattern defines patterns to detect security threats
type ThreatPattern struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // sql_injection, xss, brute_force, etc.
	Patterns   []string               `json:"patterns"`
	Severity   string                 `json:"severity"`
	Threshold  int                    `json:"threshold"`
	TimeWindow time.Duration          `json:"time_window"`
	Enabled    bool                   `json:"enabled"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// SecurityMetrics tracks security statistics
type SecurityMetrics struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	MetricDate           time.Time `json:"metric_date" gorm:"not null"`
	TotalRequests        int64     `json:"total_requests"`
	BlockedRequests      int64     `json:"blocked_requests"`
	SQLInjectionAttempts int64     `json:"sql_injection_attempts"`
	XSSAttempts          int64     `json:"xss_attempts"`
	BruteForceAttempts   int64     `json:"brute_force_attempts"`
	SuspiciousIPs        int64     `json:"suspicious_ips"`
	AlertsGenerated      int64     `json:"alerts_generated"`
	CriticalAlerts       int64     `json:"critical_alerts"`
	HighAlerts           int64     `json:"high_alerts"`
	MediumAlerts         int64     `json:"medium_alerts"`
	LowAlerts            int64     `json:"low_alerts"`
	CreatedAt            time.Time `json:"created_at"`
}

// Alert types
const (
	AlertTypeSQLInjection        = "sql_injection"
	AlertTypeXSS                 = "xss_attempt"
	AlertTypeBruteForce          = "brute_force"
	AlertTypeUnauthorizedAccess  = "unauthorized_access"
	AlertTypeDataBreach          = "data_breach"
	AlertTypeSuspiciousIP        = "suspicious_ip"
	AlertTypeRateLimit           = "rate_limit_exceeded"
	AlertTypeFileUpload          = "suspicious_file_upload"
	AlertTypePrivilegeEscalation = "privilege_escalation"
	AlertTypeDataExfiltration    = "data_exfiltration"
	AlertTypeComplianceViolation = "compliance_violation"
)

// Severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// NewRealtimeMonitor creates a new real-time security monitor
func NewRealtimeMonitor(db *gorm.DB, logger *log.Logger, auditLogger *AuditLogger) *RealtimeMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	// Auto-migrate tables
	db.AutoMigrate(&SecurityAlert{}, &SecurityMetrics{})

	monitor := &RealtimeMonitor{
		db:            db,
		logger:        logger,
		auditLogger:   auditLogger,
		alertChannels: make(map[string]chan SecurityAlert),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Initialize default threat patterns
	monitor.initializeDefaultPatterns()

	return monitor
}

// Start begins real-time monitoring
func (rm *RealtimeMonitor) Start() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if rm.isRunning {
		return fmt.Errorf("monitor is already running")
	}

	rm.isRunning = true

	// Start monitoring goroutines
	go rm.alertProcessor()
	go rm.metricsCollector()
	go rm.threatDetector()
	go rm.cleanupOldData()

	rm.logger.Println("✅ Real-time security monitor started")
	return nil
}

// Stop stops the real-time monitoring
func (rm *RealtimeMonitor) Stop() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if !rm.isRunning {
		return
	}

	rm.cancel()
	rm.isRunning = false

	// Close alert channels
	for _, ch := range rm.alertChannels {
		close(ch)
	}
	rm.alertChannels = make(map[string]chan SecurityAlert)

	rm.logger.Println("🛑 Real-time security monitor stopped")
}

// ProcessHTTPRequest analyzes an HTTP request for security threats
func (rm *RealtimeMonitor) ProcessHTTPRequest(r *http.Request) []SecurityAlert {
	alerts := []SecurityAlert{}

	// Check for SQL injection patterns
	if sqlAlert := rm.checkSQLInjection(r); sqlAlert != nil {
		alerts = append(alerts, *sqlAlert)
		rm.sendAlert(*sqlAlert)
	}

	// Check for XSS patterns
	if xssAlert := rm.checkXSS(r); xssAlert != nil {
		alerts = append(alerts, *xssAlert)
		rm.sendAlert(*xssAlert)
	}

	// Check for suspicious file uploads
	if uploadAlert := rm.checkSuspiciousFileUpload(r); uploadAlert != nil {
		alerts = append(alerts, *uploadAlert)
		rm.sendAlert(*uploadAlert)
	}

	// Check for suspicious user agents
	if uaAlert := rm.checkSuspiciousUserAgent(r); uaAlert != nil {
		alerts = append(alerts, *uaAlert)
		rm.sendAlert(*uaAlert)
	}

	return alerts
}

// ReportBruteForceAttempt reports a brute force login attempt
func (rm *RealtimeMonitor) ReportBruteForceAttempt(username, ipAddress, userAgent string, attemptCount int) {
	alert := SecurityAlert{
		AlertType:   AlertTypeBruteForce,
		Severity:    rm.getSeverityForBruteForce(attemptCount),
		Title:       fmt.Sprintf("Brute Force Attack Detected - %d attempts", attemptCount),
		Description: fmt.Sprintf("Multiple failed login attempts detected for username '%s' from IP %s", username, ipAddress),
		Source:      "authentication_system",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		RawData: map[string]interface{}{
			"username":      username,
			"attempt_count": attemptCount,
		},
		CreatedAt: time.Now(),
	}

	rm.sendAlert(alert)
}

// ReportUnauthorizedAccess reports unauthorized access attempt
func (rm *RealtimeMonitor) ReportUnauthorizedAccess(resource, userID, ipAddress, userAgent string) {
	alert := SecurityAlert{
		AlertType:   AlertTypeUnauthorizedAccess,
		Severity:    SeverityHigh,
		Title:       "Unauthorized Access Attempt",
		Description: fmt.Sprintf("Attempt to access restricted resource '%s' by user %s", resource, userID),
		Source:      "authorization_system",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		UserID:      userID,
		RawData: map[string]interface{}{
			"resource":   resource,
			"user_id":    userID,
			"ip_address": ipAddress,
		},
		CreatedAt: time.Now(),
	}

	rm.sendAlert(alert)
}

// ReportDataAccess reports access to sensitive data
func (rm *RealtimeMonitor) ReportDataAccess(dataType, userID, ipAddress, userAgent string, recordCount int, suspicious bool) {
	severity := SeverityLow
	if suspicious {
		severity = SeverityHigh
	}

	alert := SecurityAlert{
		AlertType:   AlertTypeDataExfiltration,
		Severity:    severity,
		Title:       fmt.Sprintf("Sensitive Data Access - %s", dataType),
		Description: fmt.Sprintf("Access to %d %s records by user %s", recordCount, dataType, userID),
		Source:      "data_access_monitor",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		UserID:      userID,
		RawData: map[string]interface{}{
			"data_type":    dataType,
			"record_count": recordCount,
			"suspicious":   suspicious,
		},
		CreatedAt: time.Now(),
	}

	if suspicious || recordCount > 100 {
		rm.sendAlert(alert)
	}
}

// ReportComplianceViolation reports TREC compliance violations
func (rm *RealtimeMonitor) ReportComplianceViolation(violationType, description, clientName, propertyAddress string) {
	alert := SecurityAlert{
		AlertType:   AlertTypeComplianceViolation,
		Severity:    SeverityHigh,
		Title:       fmt.Sprintf("TREC Compliance Violation - %s", violationType),
		Description: description,
		Source:      "compliance_monitor",
		RawData: map[string]interface{}{
			"violation_type":   violationType,
			"client_name":      clientName,
			"property_address": propertyAddress,
		},
		CreatedAt: time.Now(),
	}

	rm.sendAlert(alert)
}

// SubscribeToAlerts subscribes to security alerts
func (rm *RealtimeMonitor) SubscribeToAlerts(subscriberID string) <-chan SecurityAlert {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	ch := make(chan SecurityAlert, 100) // Buffered channel
	rm.alertChannels[subscriberID] = ch

	return ch
}

// UnsubscribeFromAlerts unsubscribes from security alerts
func (rm *RealtimeMonitor) UnsubscribeFromAlerts(subscriberID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if ch, exists := rm.alertChannels[subscriberID]; exists {
		close(ch)
		delete(rm.alertChannels, subscriberID)
	}
}

// GetActiveAlerts returns all unresolved security alerts
func (rm *RealtimeMonitor) GetActiveAlerts() ([]SecurityAlert, error) {
	var alerts []SecurityAlert
	err := rm.db.Where("resolved = ?", false).
		Order("severity DESC, created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// GetSecurityMetrics returns security metrics for a date range
func (rm *RealtimeMonitor) GetSecurityMetrics(startDate, endDate time.Time) ([]SecurityMetrics, error) {
	var metrics []SecurityMetrics
	err := rm.db.Where("metric_date BETWEEN ? AND ?", startDate, endDate).
		Order("metric_date DESC").
		Find(&metrics).Error
	return metrics, err
}

// AcknowledgeAlert acknowledges a security alert
func (rm *RealtimeMonitor) AcknowledgeAlert(alertID uint, acknowledgedBy string) error {
	now := time.Now()
	return rm.db.Model(&SecurityAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_by": acknowledgedBy,
			"acknowledged_at": now,
		}).Error
}

// ResolveAlert resolves a security alert
func (rm *RealtimeMonitor) ResolveAlert(alertID uint, resolvedBy string) error {
	now := time.Now()
	return rm.db.Model(&SecurityAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"resolved":    true,
			"resolved_by": resolvedBy,
			"resolved_at": now,
		}).Error
}

// Private methods

func (rm *RealtimeMonitor) initializeDefaultPatterns() {
	rm.threatPatterns = []ThreatPattern{
		{
			Name:     "SQL Injection - UNION",
			Type:     "sql_injection",
			Patterns: []string{"UNION SELECT", "union select", "UNION ALL SELECT"},
			Severity: SeverityCritical,
			Enabled:  true,
		},
		{
			Name:     "SQL Injection - DROP TABLE",
			Type:     "sql_injection",
			Patterns: []string{"DROP TABLE", "drop table", "DELETE FROM", "delete from"},
			Severity: SeverityCritical,
			Enabled:  true,
		},
		{
			Name:     "XSS - Script Tags",
			Type:     "xss",
			Patterns: []string{"<script", "</script>", "javascript:", "eval("},
			Severity: SeverityHigh,
			Enabled:  true,
		},
		{
			Name:     "Path Traversal",
			Type:     "path_traversal",
			Patterns: []string{"../", "..\\", "%2e%2e%2f", "%2e%2e%5c"},
			Severity: SeverityHigh,
			Enabled:  true,
		},
	}
}

func (rm *RealtimeMonitor) checkSQLInjection(r *http.Request) *SecurityAlert {
	// Check URL parameters and form data
	content := r.URL.RawQuery + r.FormValue("q") + r.FormValue("search")

	for _, pattern := range rm.threatPatterns {
		if pattern.Type == "sql_injection" && pattern.Enabled {
			for _, p := range pattern.Patterns {
				if strings.Contains(strings.ToLower(content), strings.ToLower(p)) {
					return &SecurityAlert{
						AlertType:   AlertTypeSQLInjection,
						Severity:    pattern.Severity,
						Title:       "SQL Injection Attempt Detected",
						Description: fmt.Sprintf("Potential SQL injection pattern '%s' detected in request", p),
						Source:      "http_monitor",
						IPAddress:   r.RemoteAddr,
						UserAgent:   r.UserAgent(),
						RawData: map[string]interface{}{
							"pattern":    p,
							"url":        r.URL.String(),
							"method":     r.Method,
							"user_agent": r.UserAgent(),
						},
						CreatedAt: time.Now(),
					}
				}
			}
		}
	}

	return nil
}

func (rm *RealtimeMonitor) checkXSS(r *http.Request) *SecurityAlert {
	// Check form values
	for key, values := range r.Form {
		for _, value := range values {
			for _, pattern := range rm.threatPatterns {
				if pattern.Type == "xss" && pattern.Enabled {
					for _, p := range pattern.Patterns {
						if strings.Contains(strings.ToLower(value), strings.ToLower(p)) {
							return &SecurityAlert{
								AlertType:   AlertTypeXSS,
								Severity:    pattern.Severity,
								Title:       "XSS Attempt Detected",
								Description: fmt.Sprintf("Potential XSS pattern '%s' detected in form field '%s'", p, key),
								Source:      "http_monitor",
								IPAddress:   r.RemoteAddr,
								UserAgent:   r.UserAgent(),
								RawData: map[string]interface{}{
									"pattern": p,
									"field":   key,
									"value":   value[:min(len(value), 100)], // Truncate for storage
									"url":     r.URL.String(),
								},
								CreatedAt: time.Now(),
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (rm *RealtimeMonitor) checkSuspiciousFileUpload(r *http.Request) *SecurityAlert {
	if r.Method == "POST" && strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(1024); err == nil {
			for _, files := range r.MultipartForm.File {
				for _, file := range files {
					if rm.isSuspiciousFile(file.Filename) {
						return &SecurityAlert{
							AlertType:   AlertTypeFileUpload,
							Severity:    SeverityMedium,
							Title:       "Suspicious File Upload Detected",
							Description: fmt.Sprintf("Potentially dangerous file upload: %s", file.Filename),
							Source:      "file_upload_monitor",
							IPAddress:   r.RemoteAddr,
							UserAgent:   r.UserAgent(),
							RawData: map[string]interface{}{
								"filename": file.Filename,
								"size":     file.Size,
							},
							CreatedAt: time.Now(),
						}
					}
				}
			}
		}
	}

	return nil
}

func (rm *RealtimeMonitor) checkSuspiciousUserAgent(r *http.Request) *SecurityAlert {
	userAgent := r.UserAgent()
	suspiciousAgents := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"dirb",
		"gobuster",
		"masscan",
		"zap",
		"w3af",
	}

	for _, agent := range suspiciousAgents {
		if strings.Contains(strings.ToLower(userAgent), agent) {
			return &SecurityAlert{
				AlertType:   AlertTypeSuspiciousIP,
				Severity:    SeverityHigh,
				Title:       "Suspicious User Agent Detected",
				Description: fmt.Sprintf("Request from suspicious user agent: %s", userAgent),
				Source:      "user_agent_monitor",
				IPAddress:   r.RemoteAddr,
				UserAgent:   userAgent,
				RawData: map[string]interface{}{
					"user_agent": userAgent,
					"url":        r.URL.String(),
				},
				CreatedAt: time.Now(),
			}
		}
	}

	return nil
}

func (rm *RealtimeMonitor) isSuspiciousFile(filename string) bool {
	suspiciousExtensions := []string{
		".exe", ".bat", ".cmd", ".com", ".scr", ".pif",
		".php", ".jsp", ".asp", ".aspx",
		".sh", ".bash", ".ps1",
	}

	lowerFilename := strings.ToLower(filename)
	for _, ext := range suspiciousExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}

	return false
}

func (rm *RealtimeMonitor) getSeverityForBruteForce(attemptCount int) string {
	if attemptCount >= 10 {
		return SeverityCritical
	} else if attemptCount >= 5 {
		return SeverityHigh
	} else {
		return SeverityMedium
	}
}

func (rm *RealtimeMonitor) sendAlert(alert SecurityAlert) {
	// Store alert in database
	if err := rm.db.Create(&alert).Error; err != nil {
		rm.logger.Printf("❌ Failed to store security alert: %v", err)
		return
	}

	// Send to audit logger
	rm.auditLogger.LogSecurityEvent("security_alert_generated", nil, alert.IPAddress, alert.UserAgent, fmt.Sprintf("Security alert: %s", alert.Title), map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_type": alert.AlertType,
		"severity":   alert.Severity,
		"title":      alert.Title,
		"ip_address": alert.IPAddress,
	}, 70)

	// Send to subscribers
	rm.mutex.RLock()
	for subscriberID, ch := range rm.alertChannels {
		select {
		case ch <- alert:
		default:
			// Channel is full, log warning
			rm.logger.Printf("⚠️ Alert channel full for subscriber %s", subscriberID)
		}
	}
	rm.mutex.RUnlock()

	rm.logger.Printf("🚨 Security Alert: %s - %s (Severity: %s)", alert.AlertType, alert.Title, alert.Severity)
}

func (rm *RealtimeMonitor) alertProcessor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			// Process pending alerts (escalation, notifications, etc.)
			rm.processAlertEscalation()
		}
	}
}

func (rm *RealtimeMonitor) metricsCollector() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.collectHourlyMetrics()
		}
	}
}

func (rm *RealtimeMonitor) threatDetector() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			// Analyze patterns and detect emerging threats
			rm.analyzeThreats()
		}
	}
}

func (rm *RealtimeMonitor) cleanupOldData() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			// Clean up old resolved alerts and metrics
			cutoffDate := time.Now().AddDate(0, -3, 0) // 3 months

			// Delete old resolved alerts
			rm.db.Where("resolved = ? AND resolved_at < ?", true, cutoffDate).Delete(&SecurityAlert{})

			// Keep metrics for 1 year
			metricCutoff := time.Now().AddDate(-1, 0, 0)
			rm.db.Where("metric_date < ?", metricCutoff).Delete(&SecurityMetrics{})
		}
	}
}

func (rm *RealtimeMonitor) processAlertEscalation() {
	// Find high/critical alerts that haven't been acknowledged within 30 minutes
	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)

	var unacknowledgedAlerts []SecurityAlert
	rm.db.Where("acknowledged = ? AND severity IN (?, ?) AND created_at < ?",
		false, SeverityHigh, SeverityCritical, thirtyMinutesAgo).
		Find(&unacknowledgedAlerts)

	for _, alert := range unacknowledgedAlerts {
		rm.logger.Printf("🚨 ESCALATION: Unacknowledged %s alert: %s", alert.Severity, alert.Title)
		// Here you could send notifications to administrators
	}
}

func (rm *RealtimeMonitor) collectHourlyMetrics() {
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)

	metrics := SecurityMetrics{
		MetricDate: now.Truncate(time.Hour),
		CreatedAt:  now,
	}

	// Count alerts by severity
	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ?", hourAgo, now).
		Count(&metrics.AlertsGenerated)

	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ? AND severity = ?", hourAgo, now, SeverityCritical).
		Count(&metrics.CriticalAlerts)

	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ? AND severity = ?", hourAgo, now, SeverityHigh).
		Count(&metrics.HighAlerts)

	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ? AND severity = ?", hourAgo, now, SeverityMedium).
		Count(&metrics.MediumAlerts)

	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ? AND severity = ?", hourAgo, now, SeverityLow).
		Count(&metrics.LowAlerts)

	// Count specific attack types
	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ? AND alert_type = ?", hourAgo, now, AlertTypeSQLInjection).
		Count(&metrics.SQLInjectionAttempts)

	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ? AND alert_type = ?", hourAgo, now, AlertTypeXSS).
		Count(&metrics.XSSAttempts)

	rm.db.Model(&SecurityAlert{}).
		Where("created_at BETWEEN ? AND ? AND alert_type = ?", hourAgo, now, AlertTypeBruteForce).
		Count(&metrics.BruteForceAttempts)

	// Store metrics
	rm.db.Create(&metrics)
}

func (rm *RealtimeMonitor) analyzeThreats() {
	// Analyze recent alert patterns to detect coordinated attacks
	// This is a simplified implementation - in production you'd use ML/AI

	pastHour := time.Now().Add(-1 * time.Hour)

	// Check for distributed attacks (same attack type from multiple IPs)
	var results []struct {
		AlertType string
		IPCount   int64
	}

	rm.db.Model(&SecurityAlert{}).
		Select("alert_type, COUNT(DISTINCT ip_address) as ip_count").
		Where("created_at >= ?", pastHour).
		Group("alert_type").
		Having("COUNT(DISTINCT ip_address) >= ?", 5).
		Find(&results)

	for _, result := range results {
		alert := SecurityAlert{
			AlertType:   "coordinated_attack",
			Severity:    SeverityCritical,
			Title:       fmt.Sprintf("Coordinated %s Attack Detected", result.AlertType),
			Description: fmt.Sprintf("Detected %s attacks from %d different IP addresses in the past hour", result.AlertType, result.IPCount),
			Source:      "threat_analyzer",
			CreatedAt:   time.Now(),
		}
		rm.sendAlert(alert)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
