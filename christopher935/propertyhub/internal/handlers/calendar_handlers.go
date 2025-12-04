package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CalendarHandlers provides HTTP handlers for calendar operations
type CalendarHandlers struct {
	db               *gorm.DB
	calendarService  *services.CalendarIntegrationService
	automationService *services.SMSEmailAutomationService
}

// NewCalendarHandlers creates new calendar handlers
func NewCalendarHandlers(db *gorm.DB) *CalendarHandlers {
	return &CalendarHandlers{
		db:               db,
		calendarService:  services.NewCalendarIntegrationService(db),
		automationService: services.NewSMSEmailAutomationService(db),
	}
}

// CreateShowingEvent creates a new showing event
// POST /api/v1/calendar/showing
func (h *CalendarHandlers) CreateShowingEvent(c *gin.Context) {
	var bookingData services.BookingCalendarData
	if err := c.ShouldBindJSON(&bookingData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking data", err)
		return
	}

	event, err := h.calendarService.CreateShowingEvent(bookingData)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create showing event", err)
		return
	}

	// Trigger automation for booking created
	automationData := map[string]interface{}{
		"contact_id":       event.ContactID,
		"property_address": bookingData.PropertyAddress,
		"showing_time":     bookingData.ShowingTime.Format("Jan 2 at 3:04 PM"),
		"event_type":       "booking_created",
	}
	
	go func() {
		if err := h.automationService.TriggerAutomation("booking_created", automationData); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to trigger automation: %v", err)
		}
	}()

	utils.SuccessResponse(c, gin.H{
		"event":   event,
		"message": "Showing event created successfully",
	})
}

// CreateFollowUpEvent creates a follow-up event
// POST /api/v1/calendar/followup
func (h *CalendarHandlers) CreateFollowUpEvent(c *gin.Context) {
	var request struct {
		ContactID   string    `json:"contact_id" binding:"required"`
		Title       string    `json:"title" binding:"required"`
		Time        time.Time `json:"time" binding:"required"`
		Description string    `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	event, err := h.calendarService.CreateFollowUpEvent(
		request.ContactID, 
		request.Title, 
		request.Time, 
		request.Description,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create follow-up event", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"event":   event,
		"message": "Follow-up event created successfully",
	})
}

// GetUpcomingEvents retrieves upcoming events
// GET /api/v1/calendar/upcoming
func (h *CalendarHandlers) GetUpcomingEvents(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 90 {
		days = 7
	}

	events, err := h.calendarService.GetUpcomingEvents(days)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve events", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"events": events,
		"days":   days,
		"total":  len(events),
	})
}

// GetTodayEvents retrieves today's events
// GET /api/v1/calendar/today
func (h *CalendarHandlers) GetTodayEvents(c *gin.Context) {
	events, err := h.calendarService.GetTodayEvents()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve today's events", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"events": events,
		"date":   time.Now().Format("2006-01-02"),
		"total":  len(events),
	})
}

// UpdateEventStatus updates an event's status
// PATCH /api/v1/calendar/events/:id/status
func (h *CalendarHandlers) UpdateEventStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid event ID", err)
		return
	}

	var request struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	// Validate status
	validStatuses := []string{"scheduled", "confirmed", "completed", "cancelled"}
	isValid := false
	for _, status := range validStatuses {
		if request.Status == status {
			isValid = true
			break
		}
	}
	
	if !isValid {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	if err := h.calendarService.UpdateEventStatus(uint(id), request.Status); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update event status", err)
		return
	}

	// Trigger automation if event is completed
	if request.Status == "completed" {
		// Get the event to trigger appropriate automation
		var event services.CalendarEvent
		if err := h.db.First(&event, uint(id)).Error; err == nil {
			if event.EventType == "showing" {
				automationData := map[string]interface{}{
					"contact_id": event.ContactID,
					"property_address": event.Location,
					"event_type": "showing_completed",
				}
				
				go func() {
					h.automationService.TriggerAutomation("showing_completed", automationData)
				}()
			}
		}
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Event status updated successfully",
		"status":  request.Status,
	})
}

// SyncCalendar syncs calendar with Follow Up Boss
// POST /api/v1/calendar/sync
func (h *CalendarHandlers) SyncCalendar(c *gin.Context) {
	go func() {
		if err := h.calendarService.SyncWithFUB(); err != nil {
			log.Printf("Calendar sync failed: %v", err)
		}
	}()

	utils.SuccessResponse(c, gin.H{
		"message": "Calendar sync initiated",
		"status":  "started",
	})
}

// GetCalendarStats retrieves calendar statistics
// GET /api/v1/calendar/stats
func (h *CalendarHandlers) GetCalendarStats(c *gin.Context) {
	stats, err := h.calendarService.GetCalendarStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve calendar stats", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"stats": stats,
		"generated_at": time.Now().UTC(),
	})
}

// ScheduleFollowUp schedules automatic follow-up after showing
// POST /api/v1/calendar/events/:id/schedule-followup
func (h *CalendarHandlers) ScheduleFollowUp(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid event ID", err)
		return
	}

	var request struct {
		HoursAfter int `json:"hours_after"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if request.HoursAfter < 1 || request.HoursAfter > 168 { // Max 1 week
		request.HoursAfter = 2 // Default to 2 hours
	}

	if err := h.calendarService.ScheduleAutomaticFollowUp(uint(id), request.HoursAfter); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to schedule follow-up", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Follow-up scheduled successfully",
		"hours_after": request.HoursAfter,
	})
}

// GetAutomationStats retrieves automation statistics
// GET /api/v1/calendar/automation-stats
func (h *CalendarHandlers) GetAutomationStats(c *gin.Context) {
	stats, err := h.automationService.GetAutomationStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve automation stats", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"stats": stats,
		"generated_at": time.Now().UTC(),
	})
}

// TriggerAutomation manually triggers an automation
// POST /api/v1/calendar/trigger-automation
func (h *CalendarHandlers) TriggerAutomation(c *gin.Context) {
	var request struct {
		TriggerType string                 `json:"trigger_type" binding:"required"`
		Data        map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if err := h.automationService.TriggerAutomation(request.TriggerType, request.Data); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to trigger automation", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Automation triggered successfully",
		"trigger_type": request.TriggerType,
	})
}

// CalendarManagementDashboard renders the calendar management template with real data
func (h *CalendarHandlers) CalendarManagementDashboard(c *gin.Context) {
	// Get today's events
	todayEvents, err := h.calendarService.GetTodayEvents()
	if err != nil {
		log.Printf("Error getting today's events: %v", err)
		todayEvents = []services.CalendarEvent{} // Fallback to empty slice
	}

	// Get upcoming events
	upcomingEvents, err := h.calendarService.GetUpcomingEvents(7) // Next 7 days
	if err != nil {
		log.Printf("Error getting upcoming events: %v", err)
		upcomingEvents = []services.CalendarEvent{} // Fallback to empty slice
	}

	// Get calendar statistics
	stats, err := h.calendarService.GetCalendarStats()
	if err != nil {
		log.Printf("Error getting calendar stats: %v", err)
		stats = &services.CalendarStats{} // Fallback to empty stats
	}

	// Get automation statistics
	automationStats, err := h.automationService.GetAutomationStats()
	if err != nil {
		log.Printf("Error getting automation stats: %v", err)
		automationStats = &services.AutomationStats{} // Fallback to empty stats
	}

	// Prepare template data
	data := gin.H{
		"Title":           "Calendar Management - PropertyHub",
		"TodayEvents":     todayEvents,
		"UpcomingEvents":  upcomingEvents,
		"Stats":           stats,
		"AutomationStats": automationStats,
		"CreateEndpoint":  "/api/v1/calendar/showing",
		"SyncEndpoint":    "/api/v1/calendar/sync",
		"StatsEndpoint":   "/api/v1/calendar/stats",
	}

	c.HTML(http.StatusOK, "calendar-management.html", data)
}

// RegisterCalendarRoutes registers all calendar-related routes
func RegisterCalendarRoutes(r *gin.Engine, db *gorm.DB) {
	handlers := NewCalendarHandlers(db)
	
	// Template dashboard route (NEW)
	r.GET("/admin/calendar", handlers.CalendarManagementDashboard)
	
	api := r.Group("/api/v1")
	{
		calendar := api.Group("/calendar")
		{
			// Event creation (existing JSON APIs)
			calendar.POST("/showing", handlers.CreateShowingEvent)
			calendar.POST("/followup", handlers.CreateFollowUpEvent)
			
			// Event management (existing JSON APIs)
			calendar.GET("/upcoming", handlers.GetUpcomingEvents)
			calendar.GET("/today", handlers.GetTodayEvents)
			calendar.PATCH("/events/:id/status", handlers.UpdateEventStatus)
			calendar.POST("/events/:id/schedule-followup", handlers.ScheduleFollowUp)
			
			// Sync and stats (existing JSON APIs)
			calendar.POST("/sync", handlers.SyncCalendar)
			calendar.GET("/stats", handlers.GetCalendarStats)
			
			// Automation (existing JSON APIs)
			calendar.GET("/automation-stats", handlers.GetAutomationStats)
			calendar.POST("/trigger-automation", handlers.TriggerAutomation)
		}
	}
}
