package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BehavioralSessionsHandler struct {
	db *gorm.DB
}

func NewBehavioralSessionsHandler(db *gorm.DB) *BehavioralSessionsHandler {
	return &BehavioralSessionsHandler{db: db}
}

// ActiveSessionResponse represents the response for active sessions
type ActiveSessionResponse struct {
	Sessions   []ActiveSession `json:"sessions"`
	TotalCount int             `json:"total_count"`
	HotCount   int             `json:"hot_count"`
	WarmCount  int             `json:"warm_count"`
	ColdCount  int             `json:"cold_count"`
}

// ActiveSession represents a single active session with enriched data
type ActiveSession struct {
	SessionID       string           `json:"session_id"`
	LeadID          int64            `json:"lead_id"`
	LeadEmail       string           `json:"lead_email,omitempty"`
	LeadName        string           `json:"lead_name,omitempty"`
	IsAnonymous     bool             `json:"is_anonymous"`
	Location        SessionLocation  `json:"location"`
	BehavioralScore int              `json:"behavioral_score"`
	ScoreCategory   string           `json:"score_category"`
	CurrentPage     string           `json:"current_page"`
	CurrentProperty *PropertySummary `json:"current_property,omitempty"`
	PageViews       int              `json:"page_views"`
	PropertyViews   int              `json:"property_views"`
	PropertySaves   int              `json:"property_saves"`
	SessionDuration int              `json:"session_duration_seconds"`
	LastActivity    time.Time        `json:"last_activity"`
	StartTime       time.Time        `json:"start_time"`
	DeviceType      string           `json:"device_type"`
	Browser         string           `json:"browser"`
	ReferrerSource  string           `json:"referrer_source"`
}

// SessionLocation represents location information
type SessionLocation struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

// PropertySummary represents basic property information
type PropertySummary struct {
	ID      uint    `json:"id"`
	Address string  `json:"address"`
	City    string  `json:"city"`
	State   string  `json:"state"`
	Price   float64 `json:"price"`
}

// GetActiveSessions retrieves currently active user sessions with behavioral data
func (h *BehavioralSessionsHandler) GetActiveSessions(c *gin.Context) {
	// Active sessions are those without end_time and recent activity (within 15 minutes)
	cutoffTime := time.Now().Add(-15 * time.Minute)

	var sessions []models.BehavioralSession
	err := h.db.Where("end_time IS NULL AND start_time >= ?", cutoffTime).
		Order("start_time DESC").
		Find(&sessions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch active sessions",
		})
		return
	}

	// Build active session responses
	activeSessions := make([]ActiveSession, 0, len(sessions))
	hotCount := 0
	warmCount := 0
	coldCount := 0

	for _, session := range sessions {
		activeSession := h.buildActiveSession(session)
		activeSessions = append(activeSessions, activeSession)

		// Count by temperature
		switch activeSession.ScoreCategory {
		case "hot":
			hotCount++
		case "warm":
			warmCount++
		case "cold":
			coldCount++
		}
	}

	response := ActiveSessionResponse{
		Sessions:   activeSessions,
		TotalCount: len(activeSessions),
		HotCount:   hotCount,
		WarmCount:  warmCount,
		ColdCount:  coldCount,
	}

	c.JSON(http.StatusOK, response)
}

// buildActiveSession enriches a session with additional data
func (h *BehavioralSessionsHandler) buildActiveSession(session models.BehavioralSession) ActiveSession {
	activeSession := ActiveSession{
		SessionID:       session.ID,
		LeadID:          session.LeadID,
		StartTime:       session.StartTime,
		PageViews:       session.PageViews,
		SessionDuration: int(time.Since(session.StartTime).Seconds()),
		DeviceType:      session.DeviceType,
		Browser:         session.Browser,
		ReferrerSource:  session.Referrer,
		IsAnonymous:     true,
		BehavioralScore: 0,
		ScoreCategory:   "cold",
	}

	// Get lead information
	var lead models.Lead
	if err := h.db.Where("id = ?", session.LeadID).First(&lead).Error; err == nil {
		activeSession.LeadEmail = lead.Email
		activeSession.LeadName = fmt.Sprintf("%s %s", lead.FirstName, lead.LastName)
		activeSession.IsAnonymous = lead.Email == ""

		// Get location from lead
		if lead.City != "" || lead.State != "" {
			activeSession.Location = SessionLocation{
				City:    lead.City,
				State:   lead.State,
				Country: "USA",
			}
		}
	}

	// Get behavioral score
	var score models.BehavioralScore
	if err := h.db.Where("lead_id = ?", session.LeadID).
		Order("last_calculated DESC").
		First(&score).Error; err == nil {
		activeSession.BehavioralScore = score.CompositeScore
		activeSession.ScoreCategory = h.getScoreCategory(score.CompositeScore)
	}

	// Get session statistics
	var eventCount int64
	h.db.Model(&models.BehavioralEvent{}).
		Where("session_id = ?", session.ID).
		Count(&eventCount)

	// Get property views count
	var propertyViews int64
	h.db.Model(&models.BehavioralEvent{}).
		Where("session_id = ? AND event_type IN (?)", session.ID, []string{"viewed", "property_viewed"}).
		Count(&propertyViews)
	activeSession.PropertyViews = int(propertyViews)

	// Get property saves count
	var propertySaves int64
	h.db.Model(&models.BehavioralEvent{}).
		Where("session_id = ? AND event_type = ?", session.ID, "saved").
		Count(&propertySaves)
	activeSession.PropertySaves = int(propertySaves)

	// Get last activity timestamp and current page/property
	var lastEvent models.BehavioralEvent
	if err := h.db.Where("session_id = ?", session.ID).
		Order("created_at DESC").
		First(&lastEvent).Error; err == nil {
		activeSession.LastActivity = lastEvent.CreatedAt

		// Set current page based on event type
		activeSession.CurrentPage = h.getPageFromEvent(lastEvent.EventType)

		// Get current property if viewing a property
		if lastEvent.PropertyID != nil {
			var property models.Property
			if err := h.db.Where("id = ?", *lastEvent.PropertyID).First(&property).Error; err == nil {
				activeSession.CurrentProperty = &PropertySummary{
					ID:      property.ID,
					Address: string(property.Address),
					City:    property.City,
					State:   property.State,
					Price:   property.Price,
				}
			}
		}
	} else {
		// Default to session start time if no events found
		activeSession.LastActivity = session.StartTime
		activeSession.CurrentPage = "Homepage"
	}

	// Parse device type from user agent if not set
	if activeSession.DeviceType == "" {
		activeSession.DeviceType = h.parseDeviceType(session.UserAgent)
	}

	// Parse browser from user agent if not set
	if activeSession.Browser == "" {
		activeSession.Browser = h.parseBrowser(session.UserAgent)
	}

	return activeSession
}

// getScoreCategory returns the category based on composite score
func (h *BehavioralSessionsHandler) getScoreCategory(score int) string {
	if score >= 70 {
		return "hot"
	} else if score >= 40 {
		return "warm"
	}
	return "cold"
}

// getPageFromEvent returns a human-readable page name from event type
func (h *BehavioralSessionsHandler) getPageFromEvent(eventType string) string {
	switch eventType {
	case "viewed", "property_viewed":
		return "Property Details"
	case "saved":
		return "Saved Properties"
	case "applied":
		return "Application Form"
	case "inquired":
		return "Contact Form"
	case "searched":
		return "Search Results"
	default:
		return "Browsing"
	}
}

// parseDeviceType extracts device type from user agent
func (h *BehavioralSessionsHandler) parseDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	return "desktop"
}

// parseBrowser extracts browser name from user agent
func (h *BehavioralSessionsHandler) parseBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "edg/") || strings.Contains(ua, "edge") {
		return "Edge"
	} else if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		return "Chrome"
	} else if strings.Contains(ua, "firefox") {
		return "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Safari"
	}
	return "Other"
}

// GetSessionJourney retrieves the full journey/timeline for a specific session
func (h *BehavioralSessionsHandler) GetSessionJourney(c *gin.Context) {
	sessionID := c.Param("id")

	// Get session details
	var session models.BehavioralSession
	if err := h.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Session not found",
		})
		return
	}

	// Get all events for this session
	var events []models.BehavioralEvent
	h.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&events)

	// Build journey timeline
	timeline := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		timelineItem := map[string]interface{}{
			"event_type": event.EventType,
			"timestamp":  event.CreatedAt,
			"event_data": event.EventData,
		}

		// Add property details if available
		if event.PropertyID != nil {
			var property models.Property
			if err := h.db.Where("id = ?", *event.PropertyID).First(&property).Error; err == nil {
				timelineItem["property"] = map[string]interface{}{
					"id":      property.ID,
					"address": property.Address,
					"city":    property.City,
					"state":   property.State,
					"price":   property.Price,
				}
			}
		}

		timeline = append(timeline, timelineItem)
	}

	// Get behavioral score changes during session
	var scoreHistory []models.BehavioralScoreHistory
	h.db.Where("lead_id = ? AND calculated_at BETWEEN ? AND ?",
		session.LeadID,
		session.StartTime,
		time.Now()).
		Order("calculated_at ASC").
		Find(&scoreHistory)

	// Build response
	response := gin.H{
		"session_id":    session.ID,
		"lead_id":       session.LeadID,
		"start_time":    session.StartTime,
		"end_time":      session.EndTime,
		"duration":      int(time.Since(session.StartTime).Seconds()),
		"page_views":    session.PageViews,
		"interactions":  session.Interactions,
		"timeline":      timeline,
		"score_changes": scoreHistory,
		"total_events":  len(events),
	}

	c.JSON(http.StatusOK, response)
}
