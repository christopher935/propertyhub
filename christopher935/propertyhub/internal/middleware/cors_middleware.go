package middleware

import (
	"log"
	"net/http"
	"strings"
)

// CORSMiddleware provides secure CORS handling
type CORSMiddleware struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	maxAge           int
	allowCredentials bool
	logger           *log.Logger
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	MaxAge           int
	AllowCredentials bool
}

// NewCORSMiddleware creates a new CORS middleware with secure defaults
func NewCORSMiddleware(logger *log.Logger) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://landlords-texas.com",
			"https://www.landlords-texas.com",
			"https://propertyhub.com",
			"https://www.propertyhub.com",
		},
		allowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"HEAD",
		},
		allowedHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Type",
			"Content-Language",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"Cache-Control",
		},
		exposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		maxAge:           86400, // 24 hours
		allowCredentials: false, // More secure default
		logger:           logger,
	}
}

// NewCORSMiddlewareWithConfig creates CORS middleware with custom configuration
func NewCORSMiddlewareWithConfig(config CORSConfig, logger *log.Logger) *CORSMiddleware {
	cors := NewCORSMiddleware(logger)

	if len(config.AllowedOrigins) > 0 {
		cors.allowedOrigins = config.AllowedOrigins
	}
	if len(config.AllowedMethods) > 0 {
		cors.allowedMethods = config.AllowedMethods
	}
	if len(config.AllowedHeaders) > 0 {
		cors.allowedHeaders = config.AllowedHeaders
	}
	if len(config.ExposedHeaders) > 0 {
		cors.exposedHeaders = config.ExposedHeaders
	}
	if config.MaxAge > 0 {
		cors.maxAge = config.MaxAge
	}
	cors.allowCredentials = config.AllowCredentials

	return cors
}

// Middleware returns the CORS middleware handler
func (c *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			c.handlePreflight(w, r, origin)
			return
		}

		// Handle actual requests
		c.handleActualRequest(w, r, origin)

		next.ServeHTTP(w, r)
	})
}

// handlePreflight handles CORS preflight requests
func (c *CORSMiddleware) handlePreflight(w http.ResponseWriter, r *http.Request, origin string) {
	// Check if origin is allowed
	if !c.isOriginAllowed(origin) {
		c.logger.Printf("🚨 CORS: Blocked preflight request from unauthorized origin: %s", origin)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check requested method
	requestedMethod := r.Header.Get("Access-Control-Request-Method")
	if !c.isMethodAllowed(requestedMethod) {
		c.logger.Printf("🚨 CORS: Blocked preflight request with unauthorized method: %s from origin: %s", requestedMethod, origin)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check requested headers
	requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
	if !c.areHeadersAllowed(requestedHeaders) {
		c.logger.Printf("🚨 CORS: Blocked preflight request with unauthorized headers: %s from origin: %s", requestedHeaders, origin)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Set CORS headers for preflight
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.allowedMethods, ", "))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.allowedHeaders, ", "))
	w.Header().Set("Access-Control-Max-Age", string(rune(c.maxAge)))

	if c.allowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Additional security headers for preflight
	w.Header().Set("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")

	w.WriteHeader(http.StatusNoContent)
}

// handleActualRequest handles actual CORS requests
func (c *CORSMiddleware) handleActualRequest(w http.ResponseWriter, r *http.Request, origin string) {
	// Check if origin is allowed
	if origin != "" {
		if c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.logger.Printf("🚨 CORS: Blocked request from unauthorized origin: %s", origin)
			// Don't set CORS headers for unauthorized origins
			return
		}
	}

	// Set exposed headers
	if len(c.exposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(c.exposedHeaders, ", "))
	}

	// Set credentials header if enabled
	if c.allowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Add Vary header for caching
	w.Header().Set("Vary", "Origin")
}

// isOriginAllowed checks if an origin is in the allowed list
func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true // Allow requests without Origin header (same-origin)
	}

	// Check exact matches
	for _, allowedOrigin := range c.allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
	}

	// Check wildcard patterns (basic implementation)
	for _, allowedOrigin := range c.allowedOrigins {
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// isMethodAllowed checks if a method is in the allowed list
func (c *CORSMiddleware) isMethodAllowed(method string) bool {
	if method == "" {
		return false
	}

	for _, allowedMethod := range c.allowedMethods {
		if allowedMethod == method {
			return true
		}
	}

	return false
}

// areHeadersAllowed checks if all requested headers are allowed
func (c *CORSMiddleware) areHeadersAllowed(requestedHeaders string) bool {
	if requestedHeaders == "" {
		return true
	}

	headers := strings.Split(requestedHeaders, ",")
	for _, header := range headers {
		header = strings.TrimSpace(header)
		if !c.isHeaderAllowed(header) {
			return false
		}
	}

	return true
}

// isHeaderAllowed checks if a header is in the allowed list
func (c *CORSMiddleware) isHeaderAllowed(header string) bool {
	header = strings.ToLower(strings.TrimSpace(header))

	// Always allow simple headers
	simpleHeaders := []string{
		"accept",
		"accept-language",
		"content-language",
		"content-type",
	}

	for _, simpleHeader := range simpleHeaders {
		if header == simpleHeader {
			return true
		}
	}

	// Check against allowed headers list
	for _, allowedHeader := range c.allowedHeaders {
		if strings.ToLower(allowedHeader) == header {
			return true
		}
	}

	return false
}

// AddAllowedOrigin adds an origin to the allowed list
func (c *CORSMiddleware) AddAllowedOrigin(origin string) {
	c.allowedOrigins = append(c.allowedOrigins, origin)
}

// RemoveAllowedOrigin removes an origin from the allowed list
func (c *CORSMiddleware) RemoveAllowedOrigin(origin string) {
	for i, allowedOrigin := range c.allowedOrigins {
		if allowedOrigin == origin {
			c.allowedOrigins = append(c.allowedOrigins[:i], c.allowedOrigins[i+1:]...)
			break
		}
	}
}

// SetAllowedOrigins sets the entire allowed origins list
func (c *CORSMiddleware) SetAllowedOrigins(origins []string) {
	c.allowedOrigins = origins
}

// GetAllowedOrigins returns the current allowed origins list
func (c *CORSMiddleware) GetAllowedOrigins() []string {
	return c.allowedOrigins
}

// EnableCredentials enables credentials support
func (c *CORSMiddleware) EnableCredentials() {
	c.allowCredentials = true
}

// DisableCredentials disables credentials support
func (c *CORSMiddleware) DisableCredentials() {
	c.allowCredentials = false
}

// SetMaxAge sets the preflight cache max age
func (c *CORSMiddleware) SetMaxAge(maxAge int) {
	c.maxAge = maxAge
}

// GetCORSStats returns CORS middleware statistics
func (c *CORSMiddleware) GetCORSStats() map[string]interface{} {
	return map[string]interface{}{
		"allowed_origins_count": len(c.allowedOrigins),
		"allowed_methods_count": len(c.allowedMethods),
		"allowed_headers_count": len(c.allowedHeaders),
		"max_age":               c.maxAge,
		"allow_credentials":     c.allowCredentials,
		"allowed_origins":       c.allowedOrigins,
		"allowed_methods":       c.allowedMethods,
		"allowed_headers":       c.allowedHeaders,
	}
}

// ValidateOrigin validates an origin URL format
func (c *CORSMiddleware) ValidateOrigin(origin string) bool {
	if origin == "" || origin == "*" {
		return true
	}

	// Basic URL validation
	if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
		return false
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"data:",
		"vbscript:",
		"file:",
		"ftp:",
	}

	originLower := strings.ToLower(origin)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(originLower, pattern) {
			return false
		}
	}

	return true
}

// SecureCORSMiddleware returns a CORS middleware with security-focused defaults
func SecureCORSMiddleware(logger *log.Logger) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: []string{
			// Only specific allowed origins, no wildcards
			"https://landlords-texas.com",
			"https://www.landlords-texas.com",
		},
		allowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		allowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
		},
		exposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
		},
		maxAge:           3600, // 1 hour (shorter for security)
		allowCredentials: false,
		logger:           logger,
	}
}

// DevelopmentCORSMiddleware returns a CORS middleware for development
func DevelopmentCORSMiddleware(logger *log.Logger) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		},
		allowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"HEAD",
			"PATCH",
		},
		allowedHeaders: []string{
			"*", // Allow all headers in development
		},
		exposedHeaders: []string{
			"*", // Expose all headers in development
		},
		maxAge:           86400, // 24 hours
		allowCredentials: true,  // Allow credentials in development
		logger:           logger,
	}
}
