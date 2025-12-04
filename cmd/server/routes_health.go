package main

import (
	"time"

	"chrisgross-ctrl-project/internal/auth"
	"chrisgross-ctrl-project/internal/security"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterHealthRoutes registers health check and error handling routes
func RegisterHealthRoutes(r *gin.Engine, gormDB *gorm.DB, authManager *auth.SimpleAuthManager, encryptionManager *security.EncryptionManager) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":              "ok",
			"mode":                "enterprise",
			"database":            gormDB != nil,
			"enterprise_auth":     authManager != nil,
			"enterprise_security": encryptionManager != nil,
			"templates_loaded":    true,
			"timestamp":           time.Now(),
		})
	})

	// NoRoute handler with fallback
	r.NoRoute(func(c *gin.Context) {
		c.HTML(404, "404.html", gin.H{"Title": "Page Not Found"})
	})

	// Recovery handler with JSON fallback if template fails
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(500, gin.H{
					"status":  "error",
					"message": "Internal server error",
				})
			}
		}()
		c.HTML(500, "500.html", gin.H{
			"Title": "Server Error",
			"Error": recovered,
		})
	}))
}
