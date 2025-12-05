package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SessionManager handles advanced session management
type SessionManager struct {
	db          *gorm.DB
	auditLogger *AuditLogger
}

// NewSessionManager creates a new session manager
func NewSessionManager(db *gorm.DB) *SessionManager {
	return &SessionManager{
		db:          db,
		auditLogger: NewAuditLogger(db),
	}
}

// Session represents an active user session
type Session struct {
	ID                uint                   `json:"id" gorm:"primaryKey"`
	SessionID         string                 `json:"session_id" gorm:"uniqueIndex;not null"`
	UserID            uint                   `json:"user_id" gorm:"index;not null"`
	DeviceFingerprint string                 `json:"device_fingerprint" gorm:"index"`
	IPAddress         string                 `json:"ip_address" gorm:"index"`
	UserAgent         string                 `json:"user_agent"`
	Location          *GeoLocation           `json:"location" gorm:"embedded"`
	IsActive          bool                   `json:"is_active" gorm:"default:true;index"`
	IsTrusted         bool                   `json:"is_trusted" gorm:"default:false"`
	RiskScore         int                    `json:"risk_score" gorm:"default:0"` // 0-100
	LastActivity      time.Time              `json:"last_activity" gorm:"index"`
	ExpiresAt         time.Time              `json:"expires_at" gorm:"index"`
	LoginMethod       string                 `json:"login_method"` // "password", "mfa", "sso"
	MFAVerified       bool                   `json:"mfa_verified" gorm:"default:false"`
	Metadata          map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// GeoLocation represents geographical location data
type GeoLocation struct {
	Country      string  `json:"country"`
	Region       string  `json:"region"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Timezone     string  `json:"timezone"`
	ISP          string  `json:"isp"`
	Organization string  `json:"organization"`
}

// DeviceInfo represents device information
type DeviceInfo struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"user_id" gorm:"index;not null"`
	DeviceFingerprint string    `json:"device_fingerprint" gorm:"uniqueIndex;not null"`
	DeviceName        string    `json:"device_name"`
	DeviceType        string    `json:"device_type"` // "desktop", "mobile", "tablet"
	OS                string    `json:"os"`
	Browser           string    `json:"browser"`
	BrowserVersion    string    `json:"browser_version"`
	ScreenResolution  string    `json:"screen_resolution"`
	Timezone          string    `json:"timezone"`
	Language          string    `json:"language"`
	IsTrusted         bool      `json:"is_trusted" gorm:"default:false"`
	FirstSeen         time.Time `json:"first_seen"`
	LastSeen          time.Time `json:"last_seen"`
	SessionCount      int       `json:"session_count" gorm:"default:0"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SessionActivity tracks session activities
type SessionActivity struct {
	ID        uint                   `json:"id" gorm:"primaryKey"`
	SessionID string                 `json:"session_id" gorm:"index;not null"`
	UserID    uint                   `json:"user_id" gorm:"index;not null"`
	Activity  string                 `json:"activity" gorm:"index"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Endpoint  string                 `json:"endpoint"`
	Method    string                 `json:"method"`
	Details   map[string]interface{} `json:"details" gorm:"type:json"`
	RiskScore int                    `json:"risk_score"`
	CreatedAt time.Time              `json:"created_at"`
}

// SuspiciousActivity represents detected suspicious activities
type SuspiciousActivity struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	UserID       *uint                  `json:"user_id" gorm:"index"`
	SessionID    string                 `json:"session_id" gorm:"index"`
	ActivityType string                 `json:"activity_type" gorm:"index"` // "multiple_locations", "unusual_hours", "rapid_requests", etc.
	Description  string                 `json:"description"`
	RiskScore    int                    `json:"risk_score" gorm:"index"`
	IPAddress    string                 `json:"ip_address" gorm:"index"`
	UserAgent    string                 `json:"user_agent"`
	Details      map[string]interface{} `json:"details" gorm:"type:json"`
	IsResolved   bool                   `json:"is_resolved" gorm:"default:false"`
	ResolvedAt   *time.Time             `json:"resolved_at"`
	Actions      []string               `json:"actions" gorm:"type:json"`
	CreatedAt    time.Time              `json:"created_at"`
}

// CreateSession creates a new session with device tracking
func (sm *SessionManager) CreateSession(userID uint, r *http.Request, loginMethod string) (*Session, error) {
	// Generate session ID
	sessionID, err := sm.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Extract device information
	deviceFingerprint := sm.generateDeviceFingerprint(r)
	ipAddress := sm.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Get or create device info
	deviceInfo, err := sm.getOrCreateDeviceInfo(userID, deviceFingerprint, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	// Get geolocation
	location, err := sm.getGeoLocation(ipAddress)
	if err != nil {
		// Log error but continue without location
		fmt.Printf("Warning: Failed to get geolocation for IP %s: %v\n", ipAddress, err)
	}

	// Calculate risk score
	riskScore := sm.calculateSessionRiskScore(userID, deviceInfo, location, ipAddress)

	// Create session
	session := &Session{
		SessionID:         sessionID,
		UserID:            userID,
		DeviceFingerprint: deviceFingerprint,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		Location:          location,
		IsActive:          true,
		IsTrusted:         deviceInfo.IsTrusted,
		RiskScore:         riskScore,
		LastActivity:      time.Now(),
		ExpiresAt:         time.Now().Add(24 * time.Hour), // 24 hour default
		LoginMethod:       loginMethod,
		MFAVerified:       false,
		Metadata:          make(map[string]interface{}),
	}

	if err := sm.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update device info
	deviceInfo.LastSeen = time.Now()
	deviceInfo.SessionCount++
	sm.db.Save(deviceInfo)

	// Log session creation
	sm.auditLogger.LogSecurityEvent(
		"session_created",
		&userID,
		ipAddress,
		userAgent,
		"New session created",
		map[string]interface{}{
			"session_id":         sessionID,
			"device_fingerprint": deviceFingerprint,
			"login_method":       loginMethod,
			"risk_score":         riskScore,
			"location":           location,
		},
		riskScore,
	)

	// Check for suspicious activity
	sm.checkSuspiciousActivity(session)

	return session, nil
}

// ValidateSession validates and updates a session
func (sm *SessionManager) ValidateSession(sessionID string, r *http.Request) (*Session, error) {
	var session Session
	if err := sm.db.Where("session_id = ? AND is_active = true AND expires_at > ?", sessionID, time.Now()).First(&session).Error; err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Update last activity
	session.LastActivity = time.Now()

	// Check if IP or device changed
	currentIP := sm.getClientIP(r)
	currentFingerprint := sm.generateDeviceFingerprint(r)

	if currentIP != session.IPAddress || currentFingerprint != session.DeviceFingerprint {
		// Potential session hijacking
		riskIncrease := 30
		if currentIP != session.IPAddress {
			riskIncrease += 20 // IP change is more suspicious
		}

		session.RiskScore += riskIncrease
		if session.RiskScore > 100 {
			session.RiskScore = 100
		}

		// Log suspicious activity
		sm.auditLogger.LogSecurityEvent(
			"session_anomaly",
			&session.UserID,
			currentIP,
			r.Header.Get("User-Agent"),
			"Session IP or device fingerprint changed",
			map[string]interface{}{
				"session_id":           sessionID,
				"original_ip":          session.IPAddress,
				"current_ip":           currentIP,
				"original_fingerprint": session.DeviceFingerprint,
				"current_fingerprint":  currentFingerprint,
				"new_risk_score":       session.RiskScore,
			},
			session.RiskScore,
		)

		// Update session with new info
		session.IPAddress = currentIP
		session.DeviceFingerprint = currentFingerprint
	}

	// Save updated session
	if err := sm.db.Save(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Log activity
	sm.logSessionActivity(sessionID, session.UserID, "session_validated", r, nil, 0)

	return &session, nil
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(sessionID string, reason string) error {
	var session Session
	if err := sm.db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return fmt.Errorf("session not found")
	}

	// Mark as inactive
	session.IsActive = false
	if err := sm.db.Save(&session).Error; err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	// Log session invalidation
	sm.auditLogger.LogSecurityEvent(
		"session_invalidated",
		&session.UserID,
		session.IPAddress,
		session.UserAgent,
		"Session invalidated: "+reason,
		map[string]interface{}{
			"session_id": sessionID,
			"reason":     reason,
		},
		10, // Low risk for normal logout
	)

	return nil
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (sm *SessionManager) InvalidateAllUserSessions(userID uint, reason string) error {
	if err := sm.db.Model(&Session{}).Where("user_id = ? AND is_active = true", userID).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	// Log mass session invalidation
	sm.auditLogger.LogSecurityEvent(
		"all_sessions_invalidated",
		&userID,
		"",
		"",
		"All user sessions invalidated: "+reason,
		map[string]interface{}{
			"user_id": userID,
			"reason":  reason,
		},
		30, // Medium risk event
	)

	return nil
}

// GetUserSessions returns active sessions for a user
func (sm *SessionManager) GetUserSessions(userID uint) ([]Session, error) {
	var sessions []Session
	err := sm.db.Where("user_id = ? AND is_active = true", userID).Order("last_activity DESC").Find(&sessions).Error
	return sessions, err
}

// GetSessionStatistics returns session statistics
func (sm *SessionManager) GetSessionStatistics() map[string]interface{} {
	var activeSessions int64
	sm.db.Model(&Session{}).Where("is_active = true AND expires_at > ?", time.Now()).Count(&activeSessions)

	var totalSessions int64
	sm.db.Model(&Session{}).Count(&totalSessions)

	var trustedSessions int64
	sm.db.Model(&Session{}).Where("is_active = true AND is_trusted = true").Count(&trustedSessions)

	var highRiskSessions int64
	sm.db.Model(&Session{}).Where("is_active = true AND risk_score >= 60").Count(&highRiskSessions)

	var suspiciousActivities int64
	sm.db.Model(&SuspiciousActivity{}).Where("is_resolved = false").Count(&suspiciousActivities)

	var uniqueDevices int64
	sm.db.Model(&DeviceInfo{}).Count(&uniqueDevices)

	var recentSessions int64
	sm.db.Model(&Session{}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).Count(&recentSessions)

	trustRate := float64(0)
	if activeSessions > 0 {
		trustRate = float64(trustedSessions) / float64(activeSessions) * 100
	}

	return map[string]interface{}{
		"active_sessions":       activeSessions,
		"total_sessions":        totalSessions,
		"trusted_sessions":      trustedSessions,
		"high_risk_sessions":    highRiskSessions,
		"trust_rate":            trustRate,
		"suspicious_activities": suspiciousActivities,
		"unique_devices":        uniqueDevices,
		"recent_sessions_24h":   recentSessions,
	}
}

// Helper methods

func (sm *SessionManager) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (sm *SessionManager) generateDeviceFingerprint(r *http.Request) string {
	// Combine various headers and properties to create a device fingerprint
	fingerprint := fmt.Sprintf("%s|%s|%s|%s|%s",
		r.Header.Get("User-Agent"),
		r.Header.Get("Accept-Language"),
		r.Header.Get("Accept-Encoding"),
		r.Header.Get("Accept"),
		r.Header.Get("DNT"),
	)

	// Hash the fingerprint
	hash := sha256.Sum256([]byte(fingerprint))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (sm *SessionManager) getClientIP(r *http.Request) string {
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
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func (sm *SessionManager) getOrCreateDeviceInfo(userID uint, fingerprint string, r *http.Request) (*DeviceInfo, error) {
	var deviceInfo DeviceInfo

	// Try to find existing device
	if err := sm.db.Where("user_id = ? AND device_fingerprint = ?", userID, fingerprint).First(&deviceInfo).Error; err == nil {
		return &deviceInfo, nil
	}

	// Create new device info
	deviceInfo = DeviceInfo{
		UserID:            userID,
		DeviceFingerprint: fingerprint,
		DeviceName:        sm.extractDeviceName(r.Header.Get("User-Agent")),
		DeviceType:        sm.extractDeviceType(r.Header.Get("User-Agent")),
		OS:                sm.extractOS(r.Header.Get("User-Agent")),
		Browser:           sm.extractBrowser(r.Header.Get("User-Agent")),
		BrowserVersion:    sm.extractBrowserVersion(r.Header.Get("User-Agent")),
		Timezone:          r.Header.Get("Timezone"),
		Language:          r.Header.Get("Accept-Language"),
		IsTrusted:         false,
		FirstSeen:         time.Now(),
		LastSeen:          time.Now(),
		SessionCount:      0,
	}

	if err := sm.db.Create(&deviceInfo).Error; err != nil {
		return nil, err
	}

	return &deviceInfo, nil
}

func (sm *SessionManager) getGeoLocation(ipAddress string) (*GeoLocation, error) {
	// In a real implementation, you would use a geolocation service like MaxMind GeoIP2
	// For now, return a mock location based on IP patterns

	if strings.HasPrefix(ipAddress, "192.168.") || strings.HasPrefix(ipAddress, "10.") || strings.HasPrefix(ipAddress, "172.") {
		// Private IP - assume local
		return &GeoLocation{
			Country: "US",
			Region:  "Local",
			City:    "Local Network",
		}, nil
	}

	// Mock geolocation for demo
	return &GeoLocation{
		Country:  "US",
		Region:   "Texas",
		City:     "Houston",
		Timezone: "America/Chicago",
		ISP:      "Unknown ISP",
	}, nil
}

func (sm *SessionManager) calculateSessionRiskScore(userID uint, device *DeviceInfo, location *GeoLocation, ipAddress string) int {
	riskScore := 0

	// New device increases risk
	if !device.IsTrusted {
		riskScore += 20
	}

	// Check for unusual location (simplified)
	if location != nil && location.Country != "US" {
		riskScore += 30
	}

	// Check for unusual hours (simplified - would need user's typical hours)
	hour := time.Now().Hour()
	if hour < 6 || hour > 22 {
		riskScore += 10
	}

	// Check for multiple recent sessions from different IPs
	var recentSessions int64
	sm.db.Model(&Session{}).Where("user_id = ? AND created_at > ? AND ip_address != ?",
		userID, time.Now().Add(-1*time.Hour), ipAddress).Count(&recentSessions)

	if recentSessions > 0 {
		riskScore += int(recentSessions) * 15
	}

	if riskScore > 100 {
		riskScore = 100
	}

	return riskScore
}

func (sm *SessionManager) checkSuspiciousActivity(session *Session) {
	// Check for multiple locations
	var recentSessions []Session
	sm.db.Where("user_id = ? AND created_at > ? AND id != ?",
		session.UserID, time.Now().Add(-1*time.Hour), session.ID).Find(&recentSessions)

	for _, recentSession := range recentSessions {
		if recentSession.IPAddress != session.IPAddress {
			// Different IP within an hour - suspicious
			sm.createSuspiciousActivity(
				&session.UserID,
				session.SessionID,
				"multiple_locations",
				"Multiple login locations within short timeframe",
				60,
				session.IPAddress,
				session.UserAgent,
				map[string]interface{}{
					"current_ip":  session.IPAddress,
					"previous_ip": recentSession.IPAddress,
					"time_diff":   time.Since(recentSession.CreatedAt).Minutes(),
				},
			)
		}
	}

	// Check for unusual hours
	hour := time.Now().Hour()
	if hour < 6 || hour > 22 {
		sm.createSuspiciousActivity(
			&session.UserID,
			session.SessionID,
			"unusual_hours",
			"Login during unusual hours",
			20,
			session.IPAddress,
			session.UserAgent,
			map[string]interface{}{
				"login_hour": hour,
			},
		)
	}
}

func (sm *SessionManager) createSuspiciousActivity(userID *uint, sessionID, activityType, description string, riskScore int, ipAddress, userAgent string, details map[string]interface{}) {
	suspiciousActivity := &SuspiciousActivity{
		UserID:       userID,
		SessionID:    sessionID,
		ActivityType: activityType,
		Description:  description,
		RiskScore:    riskScore,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      details,
		IsResolved:   false,
	}

	sm.db.Create(suspiciousActivity)
}

func (sm *SessionManager) logSessionActivity(sessionID string, userID uint, activity string, r *http.Request, details map[string]interface{}, riskScore int) {
	sessionActivity := &SessionActivity{
		SessionID: sessionID,
		UserID:    userID,
		Activity:  activity,
		IPAddress: sm.getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Endpoint:  r.URL.Path,
		Method:    r.Method,
		Details:   details,
		RiskScore: riskScore,
	}

	sm.db.Create(sessionActivity)
}

// Simple user agent parsing (in production, use a proper library)
func (sm *SessionManager) extractDeviceName(userAgent string) string {
	if strings.Contains(userAgent, "iPhone") {
		return "iPhone"
	} else if strings.Contains(userAgent, "iPad") {
		return "iPad"
	} else if strings.Contains(userAgent, "Android") {
		return "Android Device"
	} else if strings.Contains(userAgent, "Windows") {
		return "Windows PC"
	} else if strings.Contains(userAgent, "Macintosh") {
		return "Mac"
	}
	return "Unknown Device"
}

func (sm *SessionManager) extractDeviceType(userAgent string) string {
	if strings.Contains(userAgent, "Mobile") || strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "Android") {
		return "mobile"
	} else if strings.Contains(userAgent, "Tablet") || strings.Contains(userAgent, "iPad") {
		return "tablet"
	}
	return "desktop"
}

func (sm *SessionManager) extractOS(userAgent string) string {
	if strings.Contains(userAgent, "Windows NT") {
		return "Windows"
	} else if strings.Contains(userAgent, "Macintosh") {
		return "macOS"
	} else if strings.Contains(userAgent, "Linux") {
		return "Linux"
	} else if strings.Contains(userAgent, "iPhone OS") || strings.Contains(userAgent, "iOS") {
		return "iOS"
	} else if strings.Contains(userAgent, "Android") {
		return "Android"
	}
	return "Unknown"
}

func (sm *SessionManager) extractBrowser(userAgent string) string {
	if strings.Contains(userAgent, "Chrome") {
		return "Chrome"
	} else if strings.Contains(userAgent, "Firefox") {
		return "Firefox"
	} else if strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome") {
		return "Safari"
	} else if strings.Contains(userAgent, "Edge") {
		return "Edge"
	}
	return "Unknown"
}

func (sm *SessionManager) extractBrowserVersion(userAgent string) string {
	// Simplified version extraction
	if strings.Contains(userAgent, "Chrome/") {
		parts := strings.Split(userAgent, "Chrome/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			return version
		}
	}
	return "Unknown"
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions() error {
	return sm.db.Where("expires_at < ? OR (is_active = false AND updated_at < ?)",
		time.Now(), time.Now().Add(-7*24*time.Hour)).Delete(&Session{}).Error
}

