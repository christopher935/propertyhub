package handlers

import (
	"fmt"
	"net/http"
	"time"
	
	"chrisgross-ctrl-project/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LiveActivityHandler struct {
	db *gorm.DB
}

func NewLiveActivityHandler(db *gorm.DB) *LiveActivityHandler {
	return &LiveActivityHandler{db: db}
}

type LiveActivity struct {
	SessionID      string                 `json:"session_id"`
	EventType      string                 `json:"event_type"`
	PropertyID     *int64                 `json:"property_id,omitempty"`
	PropertyAddress string                `json:"property_address,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	EventData      map[string]interface{} `json:"event_data,omitempty"`
}

func (h *LiveActivityHandler) GetLiveActivity(c *gin.Context) {
	minutesAgo := 15
	if mins := c.Query("minutes"); mins != "" {
		fmt.Sscanf(mins, "%d", &minutesAgo)
	}
	
	cutoff := time.Now().Add(-time.Duration(minutesAgo) * time.Minute)
	
	var events []models.BehavioralEvent
	h.db.Where("created_at >= ?", cutoff).
		Order("created_at DESC").
		Limit(100).
		Find(&events)
	
	activities := make([]LiveActivity, 0, len(events))
	for _, event := range events {
		activity := LiveActivity{
			SessionID: event.SessionID,
			EventType: event.EventType,
			PropertyID: event.PropertyID,
			Timestamp: event.CreatedAt,
			UserAgent: event.UserAgent,
			EventData: event.EventData,
		}
		
		if event.PropertyID != nil {
			var property models.Property
			if err := h.db.Select("address").First(&property, *event.PropertyID).Error; err == nil {
				activity.PropertyAddress = string(property.Address)
			}
		}
		
		activities = append(activities, activity)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"count": len(activities),
		"timestamp": time.Now(),
	})
}

func (h *LiveActivityHandler) GetActiveSessions(c *gin.Context) {
	cutoff := time.Now().Add(-30 * time.Minute)
	
	var sessions []models.BehavioralSession
	h.db.Where("start_time >= ? AND (end_time IS NULL OR end_time >= ?)", cutoff, cutoff).
		Order("start_time DESC").
		Find(&sessions)
	
	activeSessions := make([]map[string]interface{}, 0, len(sessions))
	for _, session := range sessions {
		var eventsCount int64
		h.db.Model(&models.BehavioralEvent{}).Where("session_id = ?", session.ID).Count(&eventsCount)
		
		var propertiesViewed int64
		h.db.Model(&models.BehavioralEvent{}).
			Where("session_id = ? AND event_type IN (?)", session.ID, []string{"viewed", "property_viewed"}).
			Count(&propertiesViewed)
		
		activeSessions = append(activeSessions, map[string]interface{}{
			"session_id": session.ID,
			"start_time": session.StartTime,
			"duration_seconds": session.DurationSeconds,
			"page_views": session.PageViews,
			"interactions": session.Interactions,
			"properties_viewed": propertiesViewed,
			"events_count": eventsCount,
			"device_type": session.DeviceType,
			"is_active": session.EndTime == nil,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"active_sessions": activeSessions,
		"count": len(activeSessions),
		"timestamp": time.Now(),
	})
}

func (h *LiveActivityHandler) GetSessionDetails(c *gin.Context) {
	sessionID := c.Param("id")
	
	var session models.BehavioralSession
	if err := h.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}
	
	var events []models.BehavioralEvent
	h.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&events)
	
	var propertiesViewed []models.Property
	propertyIDs := make([]int64, 0)
	for _, event := range events {
		if event.PropertyID != nil {
			propertyIDs = append(propertyIDs, *event.PropertyID)
		}
	}
	
	if len(propertyIDs) > 0 {
		h.db.Where("id IN ?", propertyIDs).Find(&propertiesViewed)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"session": session,
		"events": events,
		"properties_viewed": propertiesViewed,
		"summary": gin.H{
			"total_events": len(events),
			"properties_viewed": len(propertiesViewed),
			"duration_minutes": session.DurationSeconds / 60,
		},
	})
}
