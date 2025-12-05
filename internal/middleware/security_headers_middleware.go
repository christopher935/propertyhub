package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		
		c.Header("X-Frame-Options", "DENY")
		
		c.Header("X-XSS-Protection", "1; mode=block")
		
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://cdn.socket.io; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com; " +
			"font-src 'self' https://fonts.gstatic.com https://cdnjs.cloudflare.com; " +
			"img-src 'self' data: https: blob:; " +
			"connect-src 'self' wss: https://api.followupboss.com; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self';"
		c.Header("Content-Security-Policy", csp)
		
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")
		
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		
		c.Header("Server", "PropertyHub")
		
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	}
}

func RelaxedSecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Server", "PropertyHub")

		c.Next()
	}
}

func StaticFileSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Header("X-Content-Type-Options", "nosniff")

		c.Next()
	}
}
