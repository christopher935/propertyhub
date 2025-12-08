package handlers

import (
	"net/http"
	"strconv"

	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BehavioralEventHandler struct {
	db                  *gorm.DB
	eventService        *services.BehavioralEventService
	activityBroadcaster *services.ActivityBroadcastService
}

func NewBehavioralEventHandler(db *gorm.DB, eventService *services.BehavioralEventService, activityBroadcaster *services.ActivityBroadcastService) *BehavioralEventHandler {
	return &BehavioralEventHandler{
		db:                  db,
		eventService:        eventService,
		activityBroadcaster: activityBroadcaster,
	}
}

func (h *BehavioralEventHandler) TrackPropertyView(c *gin.Context) {
	var req struct {
		LeadID     int64  `json:"lead_id" binding:"required"`
		PropertyID int64  `json:"property_id" binding:"required"`
		SessionID  string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.eventService.TrackPropertyView(req.LeadID, req.PropertyID, req.SessionID, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track event"})
		return
	}

	eventData := map[string]interface{}{
		"property_id": req.PropertyID,
		"action":      "view",
	}
	h.activityBroadcaster.BroadcastPropertyView(req.LeadID, req.PropertyID, req.SessionID, eventData)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BehavioralEventHandler) TrackPropertySave(c *gin.Context) {
	var req struct {
		LeadID     int64  `json:"lead_id" binding:"required"`
		PropertyID int64  `json:"property_id" binding:"required"`
		SessionID  string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.eventService.TrackPropertySave(req.LeadID, req.PropertyID, req.SessionID, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track event"})
		return
	}

	eventData := map[string]interface{}{
		"property_id": req.PropertyID,
		"action":      "save",
	}
	h.activityBroadcaster.BroadcastPropertySave(req.LeadID, req.PropertyID, req.SessionID, eventData)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BehavioralEventHandler) TrackInquiry(c *gin.Context) {
	var req struct {
		LeadID      int64  `json:"lead_id" binding:"required"`
		PropertyID  *int64 `json:"property_id"`
		InquiryType string `json:"inquiry_type" binding:"required"`
		SessionID   string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.eventService.TrackInquiry(req.LeadID, req.PropertyID, req.InquiryType, req.SessionID, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track event"})
		return
	}

	eventData := map[string]interface{}{
		"inquiry_type": req.InquiryType,
		"action":       "inquiry",
	}
	if req.PropertyID != nil {
		eventData["property_id"] = *req.PropertyID
	}
	h.activityBroadcaster.BroadcastInquiry(req.LeadID, req.PropertyID, req.InquiryType, req.SessionID, eventData)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BehavioralEventHandler) TrackSearch(c *gin.Context) {
	var req struct {
		LeadID         int64                  `json:"lead_id" binding:"required"`
		SessionID      string                 `json:"session_id" binding:"required"`
		SearchCriteria map[string]interface{} `json:"search_criteria"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.activityBroadcaster.BroadcastSearch(req.LeadID, req.SessionID, req.SearchCriteria)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BehavioralEventHandler) GetActiveSessionsCount(c *gin.Context) {
	var count int64
	cutoff := c.DefaultQuery("minutes", "15")
	minutesAgo, _ := strconv.Atoi(cutoff)

	h.db.Raw("SELECT COUNT(DISTINCT session_id) FROM behavioral_sessions WHERE end_time IS NULL AND start_time >= NOW() - INTERVAL '? minutes'", minutesAgo).Scan(&count)

	c.JSON(http.StatusOK, gin.H{
		"active_count": count,
		"timestamp":    nil,
	})
}
