package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// APISecurityManager handles advanced API security features
type APISecurityManager struct {
	db           *gorm.DB
	logger       *log.Logger
	hmacSecret   []byte
	apiKeys      map[string]*APIKey
	rateLimiters map[string]*APIRateLimiter
	mutex        sync.RWMutex
	auditLogger  *AuditLogger
	monitor      *RealtimeMonitor
}

// APIKey represents an API key for external integrations
type APIKey struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	KeyID       string     `json:"key_id" gorm:"uniqueIndex;not null"`
	HashedKey   string     `json:"hashed_key" gorm:"not null"`
	Name        string     `json:"name" gorm:"not null"`
	Description string     `json:"description"`
	Scope       []string   `json:"scope" gorm:"type:json"`         // permissions/endpoints
	RateLimit   int        `json:"rate_limit" gorm:"default:1000"` // requests per hour
	IPWhitelist []string   `json:"ip_whitelist" gorm:"type:json"`
	Active      bool       `json:"active" gorm:"default:true"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	UsageCount  int64      `json:"usage_count" gorm:"default:0"`

	// Webhook-specific fields
	WebhookURL    string `json:"webhook_url"`
	WebhookSecret string `json:"webhook_secret"`

	// Integration metadata
	IntegrationType string                 `json:"integration_type"` // fub, mls, crm, analytics
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:json"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIRequest represents an API request log
type APIRequest struct {
	ID             uint                   `json:"id" gorm:"primaryKey"`
	KeyID          string                 `json:"key_id"`
	IPAddress      string                 `json:"ip_address" gorm:"not null"`
	Method         string                 `json:"method" gorm:"not null"`
	Endpoint       string                 `json:"endpoint" gorm:"not null"`
	UserAgent      string                 `json:"user_agent"`
	RequestSize    int64                  `json:"request_size"`
	ResponseSize   int64                  `json:"response_size"`
	StatusCode     int                    `json:"status_code"`
	Duration       int64                  `json:"duration"` // in milliseconds
	Authorized     bool                   `json:"authorized"`
	SignatureValid bool                   `json:"signature_valid"`
	RateLimited    bool                   `json:"rate_limited"`
	RequestData    map[string]interface{} `json:"request_data,omitempty" gorm:"type:json"`
	ErrorMessage   string                 `json:"error_message"`
	CreatedAt      time.Time              `json:"created_at"`

	// Relationships
	APIKey *APIKey `json:"api_key,omitempty" gorm:"foreignKey:KeyID;references:KeyID"`
}

// APIRateLimiter implements rate limiting for API keys
type APIRateLimiter struct {
	Requests    map[time.Time]int
	MaxRequests int
	TimeWindow  time.Duration
	mutex       sync.RWMutex
}

// SignatureValidation represents webhook signature validation
type SignatureValidation struct {
	Timestamp   int64  `json:"timestamp"`
	Body        string `json:"body"`
	Signature   string `json:"signature"`
	ExpectedSig string `json:"expected_sig"`
	Valid       bool   `json:"valid"`
}

// API scopes for different operations
const (
	ScopePropertiesRead  = "properties:read"
	ScopePropertiesWrite = "properties:write"
	ScopeBookingsRead    = "bookings:read"
	ScopeBookingsWrite   = "bookings:write"
	ScopeContactsRead    = "contacts:read"
	ScopeContactsWrite   = "contacts:write"
	ScopeAnalyticsRead   = "analytics:read"
	ScopeWebhookReceive  = "webhook:receive"
	ScopeAdminRead       = "admin:read"
	ScopeAdminWrite      = "admin:write"
)

// Integration types
const (
	IntegrationFUB       = "followup_boss"
	IntegrationMLS       = "mls"
	IntegrationCRM       = "crm"
	IntegrationAnalytics = "analytics"
	IntegrationWebsite   = "website"
	IntegrationMobile    = "mobile_app"
)

// NewAPISecurityManager creates a new API security manager
func NewAPISecurityManager(db *gorm.DB, logger *log.Logger, auditLogger *AuditLogger, monitor *RealtimeMonitor) (*APISecurityManager, error) {
	// Auto-migrate tables
	err := db.AutoMigrate(&APIKey{}, &APIRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate API security tables: %v", err)
	}

	// Get or generate HMAC secret
	hmacSecret, err := getHMACSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to get HMAC secret: %v", err)
	}

	manager := &APISecurityManager{
		db:           db,
		logger:       logger,
		hmacSecret:   hmacSecret,
		apiKeys:      make(map[string]*APIKey),
		rateLimiters: make(map[string]*APIRateLimiter),
		auditLogger:  auditLogger,
		monitor:      monitor,
	}

	// Load existing API keys
	manager.loadAPIKeys()

	return manager, nil
}

// CreateAPIKey creates a new API key
func (asm *APISecurityManager) CreateAPIKey(name, description, integrationType string, scope []string, rateLimit int, ipWhitelist []string, expiresAt *time.Time) (*APIKey, string, error) {
	// Generate key ID and secret
	keyID, err := generateKeyID()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate key ID: %v", err)
	}

	secret, err := generateAPISecret()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API secret: %v", err)
	}

	// Hash the secret for storage
	hashedSecret := hashAPISecret(secret)

	// Generate webhook secret if needed
	webhookSecret, _ := generateAPISecret()

	apiKey := &APIKey{
		KeyID:           keyID,
		HashedKey:       hashedSecret,
		Name:            name,
		Description:     description,
		Scope:           scope,
		RateLimit:       rateLimit,
		IPWhitelist:     ipWhitelist,
		Active:          true,
		ExpiresAt:       expiresAt,
		WebhookSecret:   webhookSecret,
		IntegrationType: integrationType,
		Metadata:        make(map[string]interface{}),
		CreatedAt:       time.Now(),
	}

	if err := asm.db.Create(apiKey).Error; err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %v", err)
	}

	// Cache the API key
	asm.mutex.Lock()
	asm.apiKeys[keyID] = apiKey
	asm.rateLimiters[keyID] = &APIRateLimiter{
		Requests:    make(map[time.Time]int),
		MaxRequests: rateLimit,
		TimeWindow:  time.Hour,
	}
	asm.mutex.Unlock()

	// Log creation
	asm.auditLogger.LogSecurityEvent("api_key_created", nil, "", "", fmt.Sprintf("API key created: %s", name), map[string]interface{}{
		"key_id":           keyID,
		"name":             name,
		"integration_type": integrationType,
		"scope":            scope,
		"rate_limit":       rateLimit,
	}, 20)

	asm.logger.Printf("✅ API key created: %s (ID: %s)", name, keyID)

	// Return the plain secret only during creation
	return apiKey, fmt.Sprintf("%s.%s", keyID, secret), nil
}

// ValidateAPIRequest validates an API request with signature verification
func (asm *APISecurityManager) ValidateAPIRequest(r *http.Request) (*APIKey, *APIRequest, error) {
	startTime := time.Now()

	// Extract API key from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, asm.logAPIRequest("", r, 401, false, false, false, "missing authorization header", time.Since(startTime)),
			fmt.Errorf("missing Authorization header")
	}

	// Parse Bearer token format: "Bearer keyid.secret"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, asm.logAPIRequest("", r, 401, false, false, false, "invalid authorization format", time.Since(startTime)),
			fmt.Errorf("invalid Authorization format")
	}

	keyParts := strings.Split(parts[1], ".")
	if len(keyParts) != 2 {
		return nil, asm.logAPIRequest("", r, 401, false, false, false, "invalid api key format", time.Since(startTime)),
			fmt.Errorf("invalid API key format")
	}

	keyID := keyParts[0]
	secret := keyParts[1]

	// Get API key from cache or database
	apiKey, err := asm.getAPIKey(keyID)
	if err != nil {
		return nil, asm.logAPIRequest(keyID, r, 401, false, false, false, "api key not found", time.Since(startTime)),
			fmt.Errorf("invalid API key")
	}

	// Validate API key
	if !apiKey.Active {
		return nil, asm.logAPIRequest(keyID, r, 401, false, false, false, "api key inactive", time.Since(startTime)),
			fmt.Errorf("API key is inactive")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, asm.logAPIRequest(keyID, r, 401, false, false, false, "api key expired", time.Since(startTime)),
			fmt.Errorf("API key has expired")
	}

	// Validate secret
	if !validateAPISecret(apiKey.HashedKey, secret) {
		return nil, asm.logAPIRequest(keyID, r, 401, false, false, false, "invalid secret", time.Since(startTime)),
			fmt.Errorf("invalid API key secret")
	}

	// Check IP whitelist
	if len(apiKey.IPWhitelist) > 0 && !asm.isIPWhitelisted(r.RemoteAddr, apiKey.IPWhitelist) {
		return nil, asm.logAPIRequest(keyID, r, 403, true, false, false, "ip not whitelisted", time.Since(startTime)),
			fmt.Errorf("IP address not whitelisted")
	}

	// Check rate limiting
	if asm.isRateLimited(keyID) {
		return nil, asm.logAPIRequest(keyID, r, 429, true, false, true, "rate limit exceeded", time.Since(startTime)),
			fmt.Errorf("rate limit exceeded")
	}

	// Verify HMAC signature if present
	signatureValid := true
	if signature := r.Header.Get("X-Signature-SHA256"); signature != "" {
		signatureValid = asm.verifyHMACSignature(r, apiKey.WebhookSecret, signature)
		if !signatureValid {
			return nil, asm.logAPIRequest(keyID, r, 401, true, false, false, "invalid signature", time.Since(startTime)),
				fmt.Errorf("invalid HMAC signature")
		}
	}

	// Update usage statistics
	asm.updateAPIKeyUsage(keyID)

	// Log successful request
	apiRequest := asm.logAPIRequest(keyID, r, 200, true, signatureValid, false, "", time.Since(startTime))

	return apiKey, apiRequest, nil
}

// ValidateScope checks if an API key has the required scope for an endpoint
func (asm *APISecurityManager) ValidateScope(apiKey *APIKey, requiredScope string) bool {
	for _, scope := range apiKey.Scope {
		if scope == requiredScope || scope == "*" {
			return true
		}
	}
	return false
}

// GenerateWebhookSignature generates an HMAC signature for webhook payloads
func (asm *APISecurityManager) GenerateWebhookSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature verifies an HMAC signature for incoming webhooks
func (asm *APISecurityManager) VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	expectedSig := asm.GenerateWebhookSignature(payload, secret)
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSig)) == 1
}

// GetAPIUsageStats returns usage statistics for an API key
func (asm *APISecurityManager) GetAPIUsageStats(keyID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var requests []APIRequest
	err := asm.db.Where("key_id = ? AND created_at BETWEEN ? AND ?", keyID, startDate, endDate).
		Find(&requests).Error
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_requests":      len(requests),
		"authorized_requests": 0,
		"failed_requests":     0,
		"rate_limited":        0,
		"average_duration":    float64(0),
		"endpoints":           make(map[string]int),
		"status_codes":        make(map[int]int),
		"daily_usage":         make(map[string]int),
	}

	var totalDuration int64
	for _, req := range requests {
		if req.Authorized {
			stats["authorized_requests"] = stats["authorized_requests"].(int) + 1
		} else {
			stats["failed_requests"] = stats["failed_requests"].(int) + 1
		}

		if req.RateLimited {
			stats["rate_limited"] = stats["rate_limited"].(int) + 1
		}

		totalDuration += req.Duration

		// Count by endpoint
		endpoints := stats["endpoints"].(map[string]int)
		endpoints[req.Endpoint]++

		// Count by status code
		statusCodes := stats["status_codes"].(map[int]int)
		statusCodes[req.StatusCode]++

		// Count by day
		dailyUsage := stats["daily_usage"].(map[string]int)
		day := req.CreatedAt.Format("2006-01-02")
		dailyUsage[day]++
	}

	if len(requests) > 0 {
		stats["average_duration"] = float64(totalDuration) / float64(len(requests))
	}

	return stats, nil
}

// RevokeAPIKey revokes an API key
func (asm *APISecurityManager) RevokeAPIKey(keyID, revokedBy string) error {
	err := asm.db.Model(&APIKey{}).
		Where("key_id = ?", keyID).
		Update("active", false).Error
	if err != nil {
		return err
	}

	// Remove from cache
	asm.mutex.Lock()
	delete(asm.apiKeys, keyID)
	delete(asm.rateLimiters, keyID)
	asm.mutex.Unlock()

	// Log revocation
	asm.auditLogger.LogSecurityEvent("api_key_revoked", nil, "", "", fmt.Sprintf("API key revoked: %s", keyID), map[string]interface{}{
		"key_id":     keyID,
		"revoked_by": revokedBy,
	}, 30)

	asm.logger.Printf("✅ API key revoked: %s by %s", keyID, revokedBy)
	return nil
}

// GetActiveAPIKeys returns all active API keys
func (asm *APISecurityManager) GetActiveAPIKeys() ([]APIKey, error) {
	var keys []APIKey
	err := asm.db.Where("active = ?", true).Find(&keys).Error
	return keys, err
}

// Private methods

func (asm *APISecurityManager) loadAPIKeys() {
	var keys []APIKey
	if err := asm.db.Where("active = ?", true).Find(&keys).Error; err != nil {
		asm.logger.Printf("❌ Failed to load API keys: %v", err)
		return
	}

	asm.mutex.Lock()
	defer asm.mutex.Unlock()

	for _, key := range keys {
		asm.apiKeys[key.KeyID] = &key
		asm.rateLimiters[key.KeyID] = &APIRateLimiter{
			Requests:    make(map[time.Time]int),
			MaxRequests: key.RateLimit,
			TimeWindow:  time.Hour,
		}
	}

	asm.logger.Printf("✅ Loaded %d active API keys", len(keys))
}

func (asm *APISecurityManager) getAPIKey(keyID string) (*APIKey, error) {
	asm.mutex.RLock()
	if key, exists := asm.apiKeys[keyID]; exists {
		asm.mutex.RUnlock()
		return key, nil
	}
	asm.mutex.RUnlock()

	// Not in cache, try database
	var key APIKey
	err := asm.db.Where("key_id = ? AND active = ?", keyID, true).First(&key).Error
	if err != nil {
		return nil, err
	}

	// Add to cache
	asm.mutex.Lock()
	asm.apiKeys[keyID] = &key
	if _, exists := asm.rateLimiters[keyID]; !exists {
		asm.rateLimiters[keyID] = &APIRateLimiter{
			Requests:    make(map[time.Time]int),
			MaxRequests: key.RateLimit,
			TimeWindow:  time.Hour,
		}
	}
	asm.mutex.Unlock()

	return &key, nil
}

func (asm *APISecurityManager) isIPWhitelisted(remoteAddr string, whitelist []string) bool {
	// Extract IP from remoteAddr (may include port)
	ip := strings.Split(remoteAddr, ":")[0]

	for _, whitelistedIP := range whitelist {
		if ip == whitelistedIP {
			return true
		}
	}
	return false
}

func (asm *APISecurityManager) isRateLimited(keyID string) bool {
	asm.mutex.RLock()
	limiter, exists := asm.rateLimiters[keyID]
	asm.mutex.RUnlock()

	if !exists {
		return false
	}

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-limiter.TimeWindow)

	// Clean old entries
	for timestamp := range limiter.Requests {
		if timestamp.Before(cutoff) {
			delete(limiter.Requests, timestamp)
		}
	}

	// Count current requests
	totalRequests := 0
	for _, count := range limiter.Requests {
		totalRequests += count
	}

	if totalRequests >= limiter.MaxRequests {
		return true
	}

	// Record this request
	minute := now.Truncate(time.Minute)
	limiter.Requests[minute]++

	return false
}

func (asm *APISecurityManager) verifyHMACSignature(r *http.Request, secret, signature string) bool {
	body := ""
	// In a real implementation, you'd read the request body
	// For now, we'll use a placeholder

	timestamp := r.Header.Get("X-Timestamp")
	payload := timestamp + body

	expectedSig := asm.GenerateWebhookSignature([]byte(payload), secret)
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSig)) == 1
}

func (asm *APISecurityManager) updateAPIKeyUsage(keyID string) {
	asm.db.Model(&APIKey{}).
		Where("key_id = ?", keyID).
		Updates(map[string]interface{}{
			"last_used_at": time.Now(),
			"usage_count":  gorm.Expr("usage_count + 1"),
		})
}

func (asm *APISecurityManager) logAPIRequest(keyID string, r *http.Request, statusCode int, authorized, signatureValid, rateLimited bool, errorMessage string, duration time.Duration) *APIRequest {
	apiRequest := &APIRequest{
		KeyID:          keyID,
		IPAddress:      r.RemoteAddr,
		Method:         r.Method,
		Endpoint:       r.URL.Path,
		UserAgent:      r.UserAgent(),
		StatusCode:     statusCode,
		Duration:       duration.Milliseconds(),
		Authorized:     authorized,
		SignatureValid: signatureValid,
		RateLimited:    rateLimited,
		ErrorMessage:   errorMessage,
		CreatedAt:      time.Now(),
	}

	// Log to database asynchronously
	go func() {
		asm.db.Create(apiRequest)
	}()

	// Report to monitor if there's an issue
	if !authorized || rateLimited {
		asm.monitor.ReportUnauthorizedAccess(r.URL.Path, keyID, r.RemoteAddr, r.UserAgent())
	}

	return apiRequest
}

// Helper functions

func generateKeyID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "ak_" + base64.URLEncoding.EncodeToString(bytes)[:22], nil
}

func generateAPISecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func hashAPISecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

func validateAPISecret(hashedSecret, secret string) bool {
	expectedHash := hashAPISecret(secret)
	return subtle.ConstantTimeCompare([]byte(hashedSecret), []byte(expectedHash)) == 1
}

func getHMACSecret() ([]byte, error) {
	// In production, this would come from environment variables or a secure store
	// For now, generate a random secret
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	return secret, nil
}
