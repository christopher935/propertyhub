package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinCORSWrapper provides an optimized Gin wrapper for CORS middleware
func GinCORSWrapper(corsMiddleware *CORSMiddleware) gin.HandlerFunc {
	return gin.WrapH(corsMiddleware.Middleware(nil))
}

// GinSecurityWrapper provides an optimized Gin wrapper for security headers
func GinSecurityWrapper(securityMiddleware *SecurityMiddleware) gin.HandlerFunc {
	return gin.WrapH(securityMiddleware.SecurityHeaders(nil))
}

// GinRateLimitWrapper provides an optimized Gin wrapper for rate limiting
func GinRateLimitWrapper(securityMiddleware *SecurityMiddleware, config RateLimitConfig) gin.HandlerFunc {
	return gin.WrapH(securityMiddleware.RateLimit(config)(nil))
}

// GinSQLProtectionWrapper provides an optimized Gin wrapper for SQL injection protection
func GinSQLProtectionWrapper(securityMiddleware *SecurityMiddleware) gin.HandlerFunc {
	return gin.WrapH(securityMiddleware.SQLInjectionProtection(nil))
}

// GinXSSProtectionWrapper provides an optimized Gin wrapper for XSS protection
func GinXSSProtectionWrapper(securityMiddleware *SecurityMiddleware) gin.HandlerFunc {
	return gin.WrapH(securityMiddleware.XSSProtection(nil))
}

// GinBruteForceWrapper provides an optimized Gin wrapper for brute force protection
func GinBruteForceWrapper(securityMiddleware *SecurityMiddleware) gin.HandlerFunc {
	return gin.WrapH(securityMiddleware.BruteForceProtection(nil))
}

// GinAssetOptimizationWrapper provides an optimized conditional wrapper for asset optimization
func GinAssetOptimizationWrapper(assetMiddleware *AssetOptimizationMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to static assets
		if c.Request.URL.Path[:8] == "/static/" {
			handler := assetMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Next()
			}))
			handler.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Next()
	}
}
