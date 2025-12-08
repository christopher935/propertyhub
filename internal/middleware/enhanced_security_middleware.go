package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"chrisgross-ctrl-project/internal/security"
	"gorm.io/gorm"
)

// SecurityMiddleware provides comprehensive security features
type SecurityMiddleware struct {
	db          *gorm.DB
	auditLogger *security.AuditLogger
	rateLimiter *RateLimiter
	ipWhitelist map[string]bool
	ipBlacklist map[string]bool
	mutex       sync.RWMutex
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(db *gorm.DB) *SecurityMiddleware {
	return &SecurityMiddleware{
		db:          db,
		auditLogger: security.NewAuditLogger(db),
		rateLimiter: NewRateLimiter(),
		ipWhitelist: make(map[string]bool),
		ipBlacklist: make(map[string]bool),
	}
}

// RateLimiter handles rate limiting
type RateLimiter struct {
	clients map[string]*ClientLimiter
	mutex   sync.RWMutex
}

// ClientLimiter tracks rate limiting for a specific client
type ClientLimiter struct {
	requests    []time.Time
	lastRequest time.Time
	blocked     bool
	blockUntil  time.Time
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	BurstSize         int
	BlockDuration     time.Duration
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EnableRateLimit      bool
	EnableIPFiltering    bool
	EnableCSRFProtection bool
	EnableXSSProtection  bool
	EnableClickjacking   bool
	MaxRequestSize       int64
	AllowedOrigins       []string
	BlockedUserAgents    []string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientLimiter),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// SecurityHeaders applies security headers to responses
func (sm *SecurityMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// OWASP recommended security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self'; frame-ancestors 'none';")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Remove server information
		w.Header().Set("Server", "PropertyHub")

		next.ServeHTTP(w, r)
	})
}

// RateLimit applies rate limiting
func (sm *SecurityMiddleware) RateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := sm.getClientIP(r)

			if !sm.rateLimiter.Allow(clientIP, config) {
				// Log rate limit violation
				sm.auditLogger.LogSecurityEvent(
					"rate_limit_exceeded",
					nil,
					clientIP,
					r.Header.Get("User-Agent"),
					"Rate limit exceeded",
					map[string]interface{}{
						"endpoint": r.URL.Path,
						"method":   r.Method,
					},
					60, // Medium-high risk
				)

				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPFiltering filters requests based on IP whitelist/blacklist
func (sm *SecurityMiddleware) IPFiltering(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := sm.getClientIP(r)

		sm.mutex.RLock()
		isBlacklisted := sm.ipBlacklist[clientIP]
		isWhitelisted := sm.ipWhitelist[clientIP]
		hasWhitelist := len(sm.ipWhitelist) > 0
		sm.mutex.RUnlock()

		// Block if blacklisted
		if isBlacklisted {
			sm.auditLogger.LogSecurityEvent(
				"blacklisted_ip_access",
				nil,
				clientIP,
				r.Header.Get("User-Agent"),
				"Access attempt from blacklisted IP",
				map[string]interface{}{
					"endpoint": r.URL.Path,
					"method":   r.Method,
				},
				80, // High risk
			)

			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// If whitelist exists, only allow whitelisted IPs
		if hasWhitelist && !isWhitelisted {
			sm.auditLogger.LogSecurityEvent(
				"non_whitelisted_ip_access",
				nil,
				clientIP,
				r.Header.Get("User-Agent"),
				"Access attempt from non-whitelisted IP",
				map[string]interface{}{
					"endpoint": r.URL.Path,
					"method":   r.Method,
				},
				50, // Medium risk
			)

			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestValidation validates incoming requests
func (sm *SecurityMiddleware) RequestValidation(config SecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check request size
			if config.MaxRequestSize > 0 && r.ContentLength > config.MaxRequestSize {
				sm.auditLogger.LogSecurityEvent(
					"oversized_request",
					nil,
					sm.getClientIP(r),
					r.Header.Get("User-Agent"),
					"Request size exceeds limit",
					map[string]interface{}{
						"content_length": r.ContentLength,
						"max_size":       config.MaxRequestSize,
						"endpoint":       r.URL.Path,
					},
					40, // Medium risk
				)

				http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Check User-Agent blacklist
			userAgent := r.Header.Get("User-Agent")
			for _, blockedUA := range config.BlockedUserAgents {
				if strings.Contains(strings.ToLower(userAgent), strings.ToLower(blockedUA)) {
					sm.auditLogger.LogSecurityEvent(
						"blocked_user_agent",
						nil,
						sm.getClientIP(r),
						userAgent,
						"Blocked user agent detected",
						map[string]interface{}{
							"blocked_pattern": blockedUA,
							"endpoint":        r.URL.Path,
						},
						60, // Medium-high risk
					)

					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
			}

			// Validate HTTP method
			allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
			methodAllowed := false
			for _, method := range allowedMethods {
				if r.Method == method {
					methodAllowed = true
					break
				}
			}

			if !methodAllowed {
				sm.auditLogger.LogSecurityEvent(
					"invalid_http_method",
					nil,
					sm.getClientIP(r),
					userAgent,
					"Invalid HTTP method used",
					map[string]interface{}{
						"method":   r.Method,
						"endpoint": r.URL.Path,
					},
					30, // Low-medium risk
				)

				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Check for suspicious patterns in URL
			if sm.containsSuspiciousPatterns(r.URL.Path) {
				sm.auditLogger.LogSecurityEvent(
					"suspicious_url_pattern",
					nil,
					sm.getClientIP(r),
					userAgent,
					"Suspicious URL pattern detected",
					map[string]interface{}{
						"url": r.URL.Path,
					},
					70, // High risk
				)

				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BruteForceProtection protects against brute force attacks
func (sm *SecurityMiddleware) BruteForceProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to authentication endpoints
		if !strings.Contains(r.URL.Path, "/auth/") && !strings.Contains(r.URL.Path, "/login") {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := sm.getClientIP(r)

		// Check for recent failed attempts
		var failedAttempts int64
		sm.db.Model(&security.SecurityEvent{}).
			Where("event_type = ? AND ip_address = ? AND created_at > ?",
				"login_failure", clientIP, time.Now().Add(-15*time.Minute)).
			Count(&failedAttempts)

		if failedAttempts >= 5 {
			sm.auditLogger.LogSecurityEvent(
				"brute_force_detected",
				nil,
				clientIP,
				r.Header.Get("User-Agent"),
				"Brute force attack detected",
				map[string]interface{}{
					"failed_attempts": failedAttempts,
					"endpoint":        r.URL.Path,
				},
				90, // Critical risk
			)

			// Add to temporary blacklist
			sm.mutex.Lock()
			sm.ipBlacklist[clientIP] = true
			sm.mutex.Unlock()

			// Remove from blacklist after 1 hour
			go func() {
				time.Sleep(1 * time.Hour)
				sm.mutex.Lock()
				delete(sm.ipBlacklist, clientIP)
				sm.mutex.Unlock()
			}()

			http.Error(w, "Too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SQLInjectionProtection detects and blocks SQL injection attempts
func (sm *SecurityMiddleware) SQLInjectionProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		for key, values := range r.URL.Query() {
			for _, value := range values {
				if sm.containsSQLInjection(value) {
					sm.auditLogger.LogSecurityEvent(
						"sql_injection_attempt",
						nil,
						sm.getClientIP(r),
						r.Header.Get("User-Agent"),
						"SQL injection attempt detected",
						map[string]interface{}{
							"parameter": key,
							"value":     value,
							"endpoint":  r.URL.Path,
						},
						95, // Critical risk
					)

					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// XSSProtection detects and blocks XSS attempts
func (sm *SecurityMiddleware) XSSProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters and headers
		for key, values := range r.URL.Query() {
			for _, value := range values {
				if sm.containsXSS(value) {
					sm.auditLogger.LogSecurityEvent(
						"xss_attempt",
						nil,
						sm.getClientIP(r),
						r.Header.Get("User-Agent"),
						"XSS attempt detected",
						map[string]interface{}{
							"parameter": key,
							"value":     value,
							"endpoint":  r.URL.Path,
						},
						85, // High risk
					)

					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Allow checks if a request is allowed by rate limiter
func (rl *RateLimiter) Allow(clientIP string, config RateLimitConfig) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	client, exists := rl.clients[clientIP]
	if !exists {
		client = &ClientLimiter{
			requests:    make([]time.Time, 0),
			lastRequest: now,
		}
		rl.clients[clientIP] = client
	}

	// Check if client is currently blocked
	if client.blocked && now.Before(client.blockUntil) {
		return false
	}

	// Reset block status if block period has passed
	if client.blocked && now.After(client.blockUntil) {
		client.blocked = false
		client.requests = make([]time.Time, 0)
	}

	// Remove old requests (older than 1 hour)
	cutoff := now.Add(-1 * time.Hour)
	newRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			newRequests = append(newRequests, reqTime)
		}
	}
	client.requests = newRequests

	// Check hourly limit
	if len(client.requests) >= config.RequestsPerHour {
		client.blocked = true
		client.blockUntil = now.Add(config.BlockDuration)
		return false
	}

	// Check per-minute limit
	minuteCutoff := now.Add(-1 * time.Minute)
	recentRequests := 0
	for _, reqTime := range client.requests {
		if reqTime.After(minuteCutoff) {
			recentRequests++
		}
	}

	if recentRequests >= config.RequestsPerMinute {
		return false
	}

	// Allow request and record it
	client.requests = append(client.requests, now)
	client.lastRequest = now

	return true
}

// cleanup removes old client data
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		cutoff := time.Now().Add(-2 * time.Hour)

		for clientIP, client := range rl.clients {
			if client.lastRequest.Before(cutoff) && !client.blocked {
				delete(rl.clients, clientIP)
			}
		}
		rl.mutex.Unlock()
	}
}

// Helper methods

func (sm *SecurityMiddleware) getClientIP(r *http.Request) string {
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

func (sm *SecurityMiddleware) containsSuspiciousPatterns(path string) bool {
	suspiciousPatterns := []string{
		"../", "..\\", "..",
		"/etc/passwd", "/etc/shadow",
		"cmd.exe", "powershell",
		"<script", "</script>",
		"javascript:", "vbscript:",
		"onload=", "onerror=",
		"eval(", "alert(",
		"document.cookie",
		"union select", "drop table",
		"insert into", "delete from",
		"update set", "create table",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

func (sm *SecurityMiddleware) containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/",
		"union", "select", "insert", "update", "delete", "drop", "create", "alter",
		"exec", "execute", "sp_", "xp_",
		"or 1=1", "and 1=1", "' or '1'='1", "\" or \"1\"=\"1",
		"'; drop table", "\"; drop table",
		"benchmark(", "sleep(", "waitfor delay",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

func (sm *SecurityMiddleware) containsXSS(input string) bool {
	xssPatterns := []string{
		"<script", "</script>", "<iframe", "</iframe>",
		"<object", "</object>", "<embed", "</embed>",
		"javascript:", "vbscript:", "data:text/html",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"onfocus=", "onblur=", "onchange=", "onsubmit=",
		"eval(", "alert(", "confirm(", "prompt(",
		"document.cookie", "document.write", "window.location",
		"fromcharcode", "string.fromcharcode",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// Management methods

// AddToWhitelist adds an IP to the whitelist
func (sm *SecurityMiddleware) AddToWhitelist(ip string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ipWhitelist[ip] = true
}

// RemoveFromWhitelist removes an IP from the whitelist
func (sm *SecurityMiddleware) RemoveFromWhitelist(ip string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.ipWhitelist, ip)
}

// AddToBlacklist adds an IP to the blacklist
func (sm *SecurityMiddleware) AddToBlacklist(ip string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ipBlacklist[ip] = true
}

// RemoveFromBlacklist removes an IP from the blacklist
func (sm *SecurityMiddleware) RemoveFromBlacklist(ip string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.ipBlacklist, ip)
}

// GetWhitelist returns the current whitelist
func (sm *SecurityMiddleware) GetWhitelist() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	ips := make([]string, 0, len(sm.ipWhitelist))
	for ip := range sm.ipWhitelist {
		ips = append(ips, ip)
	}
	return ips
}

// GetBlacklist returns the current blacklist
func (sm *SecurityMiddleware) GetBlacklist() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	ips := make([]string, 0, len(sm.ipBlacklist))
	for ip := range sm.ipBlacklist {
		ips = append(ips, ip)
	}
	return ips
}

// GetRateLimitStats returns rate limiting statistics
func (sm *SecurityMiddleware) GetRateLimitStats() map[string]interface{} {
	sm.rateLimiter.mutex.RLock()
	defer sm.rateLimiter.mutex.RUnlock()

	totalClients := len(sm.rateLimiter.clients)
	blockedClients := 0

	for _, client := range sm.rateLimiter.clients {
		if client.blocked {
			blockedClients++
		}
	}

	return map[string]interface{}{
		"total_clients":   totalClients,
		"blocked_clients": blockedClients,
		"active_clients":  totalClients - blockedClients,
	}
}

// ClearRateLimitData clears all rate limiting data
func (sm *SecurityMiddleware) ClearRateLimitData() {
	sm.rateLimiter.mutex.Lock()
	defer sm.rateLimiter.mutex.Unlock()

	sm.rateLimiter.clients = make(map[string]*ClientLimiter)
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		EnableRateLimit:      true,
		EnableIPFiltering:    true,
		EnableCSRFProtection: true,
		EnableXSSProtection:  true,
		EnableClickjacking:   true,
		MaxRequestSize:       10 * 1024 * 1024, // 10MB
		AllowedOrigins:       []string{"*"},
		BlockedUserAgents: []string{
			"bot", "crawler", "spider", "scraper",
			"curl", "wget", "python-requests",
		},
	}
}

// DefaultRateLimitConfig returns a default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60,
		RequestsPerHour:   1000,
		BurstSize:         10,
		BlockDuration:     15 * time.Minute,
	}
}
