package middleware

import (
	"github.com/gin-gonic/gin"
	"chrisgross-ctrl-project/internal/security"
	"net/http"
)

// EnhancedSecurityMiddleware provides a wrapper for the comprehensive security middleware
func EnhancedSecurityMiddleware(validator *security.InputValidator) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Apply security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self'; frame-ancestors 'none';")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Remove server information
		c.Header("Server", "PropertyHub")

		// Input validation for query parameters and form data
		if validator != nil {
			// Validate query parameters
			for key, values := range c.Request.URL.Query() {
				for _, value := range values {
					if result := validator.ValidateInput(key, value); !result.IsValid {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + result.FirstError()})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	})
}
