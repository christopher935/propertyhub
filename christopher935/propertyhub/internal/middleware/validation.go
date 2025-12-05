package middleware

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

// ValidateIDParam validates that :id parameter exists and is a valid integer
func ValidateIDParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ID parameter is required",
			})
			c.Abort()
			return
		}
		
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ID must be a valid positive integer",
			})
			c.Abort()
			return
		}
		
		// Store parsed ID in context
		c.Set("id", id)
		c.Next()
	}
}

// ValidateAreaParam validates that :area parameter exists
func ValidateAreaParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		area := c.Param("area")
		if area == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Area parameter is required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateCityParam validates that :city parameter exists
func ValidateCityParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Param("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "City parameter is required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateTypeParam validates that :type parameter exists
func ValidateTypeParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		typeParam := c.Param("type")
		if typeParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Type parameter is required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
