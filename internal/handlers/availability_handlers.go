package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/services"
)

type AvailabilityHandler struct {
	db                  *gorm.DB
	availabilityService *services.AvailabilityService
}

func NewAvailabilityHandler(db *gorm.DB) *AvailabilityHandler {
	return &AvailabilityHandler{
		db:                  db,
		availabilityService: services.NewAvailabilityService(db),
	}
}

// CheckAvailabilityGin checks if a property is available for booking (Gin handler)
func (h *AvailabilityHandler) CheckAvailabilityGin(c *gin.Context) {
	propertyID := c.Query("property_id")
	mlsID := c.Query("mls_id")
	dateStr := c.Query("date")
	
	if propertyID != "" && mlsID == "" {
		var property struct {
			MLSId string
		}
		if id, err := strconv.ParseUint(propertyID, 10, 32); err == nil {
			if err := h.db.Table("properties").Select("mls_id").Where("id = ?", id).First(&property).Error; err == nil {
				mlsID = property.MLSId
			}
		}
	}
	
	if mlsID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "property_id or mls_id required"})
		return
	}
	
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	
	requestedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}
	
	check, err := h.availabilityService.CheckAvailability(mlsID, requestedDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "data": check})
}

func (h *AvailabilityHandler) GetBlackoutDatesGin(c *gin.Context) {
	mlsID := c.Query("mls_id")
	propertyID := c.Query("property_id")
	
	if propertyID != "" && mlsID == "" {
		var property struct { MLSId string }
		if id, err := strconv.ParseUint(propertyID, 10, 32); err == nil {
			h.db.Table("properties").Select("mls_id").Where("id = ?", id).First(&property)
			mlsID = property.MLSId
		}
	}
	
	blackouts, err := h.availabilityService.GetBlackoutDates(mlsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": blackouts, "count": len(blackouts)})
}

func (h *AvailabilityHandler) CreateBlackoutDateGin(c *gin.Context) {
	var req struct {
		MLSID       string `json:"mls_id"`
		PropertyID  string `json:"property_id"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Reason      string `json:"reason"`
		IsGlobal    bool   `json:"is_global"`
		CreatedBy   string `json:"created_by"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	mlsID := req.MLSID
	if req.PropertyID != "" && mlsID == "" {
		var property struct { MLSId string }
		if id, err := strconv.ParseUint(req.PropertyID, 10, 32); err == nil {
			h.db.Table("properties").Select("mls_id").Where("id = ?", id).First(&property)
			mlsID = property.MLSId
		}
	}
	
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}
	
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}
	
	if req.CreatedBy == "" {
		req.CreatedBy = "admin"
	}
	
	blackout, err := h.availabilityService.CreateBlackoutDate(mlsID, startDate, endDate, req.Reason, req.CreatedBy, req.IsGlobal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Blackout date created", "data": blackout})
}

func (h *AvailabilityHandler) RemoveBlackoutDateGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}
	
	if err := h.availabilityService.RemoveBlackoutDate(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Blackout date removed"})
}

func (h *AvailabilityHandler) GetUpcomingBlackoutsGin(c *gin.Context) {
	blackouts, err := h.availabilityService.GetUpcomingBlackouts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": blackouts, "count": len(blackouts)})
}

func (h *AvailabilityHandler) GetAvailabilityStatsGin(c *gin.Context) {
	stats, err := h.availabilityService.GetAvailabilityStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *AvailabilityHandler) ValidateBookingGin(c *gin.Context) {
	var req struct {
		MLSID      string `json:"mls_id"`
		PropertyID string `json:"property_id"`
		Date       string `json:"date"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	mlsID := req.MLSID
	if req.PropertyID != "" && mlsID == "" {
		var property struct { MLSId string }
		if id, err := strconv.ParseUint(req.PropertyID, 10, 32); err == nil {
			h.db.Table("properties").Select("mls_id").Where("id = ?", id).First(&property)
			mlsID = property.MLSId
		}
	}
	
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}
	
	err = h.availabilityService.ValidateBookingDate(mlsID, date)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "valid": false, "message": err.Error(), "can_book": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "valid": true, "message": "Booking date is available", "can_book": true})
}

func (h *AvailabilityHandler) CleanupExpiredBlackoutsGin(c *gin.Context) {
	if err := h.availabilityService.CleanupExpiredBlackouts(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Expired blackouts cleaned up"})
}
