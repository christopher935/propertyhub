package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

type AnalyticsAPIHandlers struct {
	db *gorm.DB
}

func NewAnalyticsAPIHandlers(db *gorm.DB) *AnalyticsAPIHandlers {
	return &AnalyticsAPIHandlers{
		db: db,
	}
}

// GetPropertyAnalytics returns property analytics data
func (h *AnalyticsAPIHandlers) GetPropertyAnalytics(c *gin.Context) {
	var properties []models.Property
	var totalCount int64

	h.db.Model(&models.Property{}).Count(&totalCount)
	h.db.Where("status = ?", "active").Find(&properties)

	activeCount := len(properties)
	
	analytics := gin.H{
		"total_properties": totalCount,
		"active_properties": activeCount,
		"inactive_properties": totalCount - int64(activeCount),
		"last_updated": time.Now(),
	}

	c.JSON(http.StatusOK, analytics)
}

// GetBookingAnalytics returns booking analytics data  
func (h *AnalyticsAPIHandlers) GetBookingAnalytics(c *gin.Context) {
	var bookings []models.Booking
	var totalCount int64

	h.db.Model(&models.Booking{}).Count(&totalCount)
	h.db.Where("status = ?", "scheduled").Find(&bookings)

	scheduledCount := len(bookings)

	analytics := gin.H{
		"total_bookings": totalCount,
		"scheduled_bookings": scheduledCount,
		"completed_bookings": totalCount - int64(scheduledCount),
		"last_updated": time.Now(),
	}

	c.JSON(http.StatusOK, analytics)
}

// RegisterAnalyticsRoutes registers all analytics routes
func RegisterAnalyticsRoutes(r *gin.Engine, db *gorm.DB) {
	handler := NewAnalyticsAPIHandlers(db)
	
	analyticsGroup := r.Group("/api/v1/analytics")
	{
		analyticsGroup.GET("/properties", handler.GetPropertyAnalytics)
		analyticsGroup.GET("/bookings", handler.GetBookingAnalytics)
	}
}

// Helper function to use strconv
func parseLimit(limitStr string) int {
	if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
		return limit
	}
	return 50
}
