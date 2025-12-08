package middleware

import (
	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// InjectDB creates middleware that injects database into Gin context
func InjectDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// InjectRedis creates middleware that injects Redis client into Gin context
func InjectRedis(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("redis", redis)
		c.Next()
	}
}

// InjectAWSService creates middleware that injects AWS communication service into Gin context
func InjectAWSService(awsService *services.AWSCommunicationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("awsService", awsService)
		c.Next()
	}
}
