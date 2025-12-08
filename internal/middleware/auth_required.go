package middleware

import (
	"net/http"

	"chrisgross-ctrl-project/internal/auth"
	"chrisgross-ctrl-project/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthRequired creates middleware that requires authentication for admin routes
func AuthRequired(authManager interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try token-based authentication first
		sessionToken, err := c.Cookie("admin_session_token")
		if err != nil || sessionToken == "" {
			// Try fallback cookie
			sessionToken, err = c.Cookie("admin_session")
			if err != nil || sessionToken == "" {
				// No valid session cookie found
				c.Redirect(http.StatusFound, "/admin")
				c.Abort()
				return
			}
		}
		
		var user *models.AdminUser
		if simpleAuth, ok := authManager.(*auth.SimpleAuthManager); ok {
			user, err = simpleAuth.ValidateSessionToken(sessionToken)
		} else if cachedAuth, ok := authManager.(*auth.CachedSessionManager); ok {
			user, err = cachedAuth.ValidateSessionToken(sessionToken)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication system error"})
			c.Abort()
			return
		}
		
		if err != nil || user == nil {
			// Invalid session token
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		
		// Store user in context for use in handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Next()
	}
}
