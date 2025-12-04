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
			"timestamp":           time.Now(),
		})
	})

	// Error handlers
	r.NoRoute(func(c *gin.Context) {
		c.HTML(404, "404.html", gin.H{"Title": "Page Not Found"})
	})

	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.HTML(500, "500.html", gin.H{
			"Title": "Server Error",
			"Error": recovered,
		})
	}))
}
