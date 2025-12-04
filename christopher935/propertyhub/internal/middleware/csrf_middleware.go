package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// CSRFProtection middleware prevents cross-site request forgery attacks
// Uses constant-time comparison to prevent timing attacks
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for internal API endpoints (sync, jobs, etc.)
		path := c.Request.URL.Path
		if isCSRFExemptPath(path) {
			c.Next()
			return
		}

		// Skip CSRF for safe methods - but still set token for forms
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			// Generate and set CSRF token for forms
			token := generateCSRFToken()
			c.SetCookie("csrf_token", token, 3600, "/", "", false, false)
			c.Header("X-CSRF-Token", token)
			// ISSUE #4 FIX: Set token in context so handlers can retrieve it
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// For unsafe methods (POST, PUT, DELETE, etc.), verify CSRF token
		expectedToken, err := c.Cookie("csrf_token")
		if err != nil || expectedToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token required",
				"message": "No CSRF token found in cookies",
			})
			c.Abort()
			return
		}

		// Get token from request (header or form)
		providedToken := c.GetHeader("X-CSRF-Token")
		if providedToken == "" {
			providedToken = c.PostForm("csrf_token")
		}

		if providedToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token required",
				"message": "No CSRF token provided in request",
			})
			c.Abort()
			return
		}

		// SECURITY FIX: Use constant-time comparison to prevent timing attacks
		if !constantTimeEquals(providedToken, expectedToken) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token invalid",
				"message": "CSRF token validation failed",
			})
			c.Abort()
			return
		}

		// Token is valid, continue processing
		c.Next()
	}
}

// constantTimeEquals performs constant-time string comparison
func constantTimeEquals(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// generateCSRFToken creates a cryptographically secure random token
func generateCSRFToken() string {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		// Fallback: use timestamp-based token (less secure)
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.URLEncoding.EncodeToString(token)
}

// isCSRFExemptPath checks if a path should be exempt from CSRF protection
// CSRF is designed to prevent malicious websites from making unauthorized requests
// on behalf of authenticated users. It's appropriate for browser-based form submissions.
//
// Exempt paths fall into these categories:
// 1. Internal automation/sync endpoints (should use API auth instead)
// 2. Read-only analytics/reporting endpoints (no state mutation)
// 3. Webhook receivers (external systems can't provide CSRF tokens)
//
// User-facing form submissions (bookings, contact, lead capture) KEEP CSRF protection.
func isCSRFExemptPath(path string) bool {
	// Category 1: Internal automation/sync endpoints
	// These should be protected by API authentication, IP whitelisting, or admin sessions
	internalAutomation := []string{
		"/api/v1/properties/sync/",     // HAR property synchronization
		"/api/v1/jobs/",                 // Background job management
		"/api/v1/har/trigger-scraping",  // Manual HAR scraping trigger
		"/api/har/properties/scrape",      // HAR property scraper (no v1 prefix)
		"/api/v1/har/schedule-weekly",   // HAR scraping scheduler
		"/api/v1/fub/sync",              // Follow Up Boss sync
		"/api/v1/email/process",         // Email processing jobs
	}

	// Category 2: Read-only endpoints (GET requests already exempt, but some POST analytics)
	readOnly := []string{
		"/api/v1/analytics/",            // Analytics data (read-only)
		"/api/v1/reports/",              // Report generation
		"/api/v1/stats/",                // Statistics endpoints
	}

	// Category 3: Webhook receivers (external systems)
	webhooks := []string{
		"/api/v1/webhooks/",             // Generic webhook receiver
		"/api/v1/fub/webhook",           // Follow Up Boss webhooks
	}

	// Combine all exempt paths
	exemptPrefixes := append(internalAutomation, readOnly...)
	exemptPrefixes = append(exemptPrefixes, webhooks...)

	for _, prefix := range exemptPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
