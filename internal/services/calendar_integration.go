package services

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// CalendarIntegrationService handles calendar operations with Follow Up Boss
type CalendarIntegrationService struct {
	db        *gorm.DB
	fubAPIKey string
}

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FUBEventID  string    `gorm:"unique" json:"fub_event_id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	StartTime   time.Time `gorm:"not null" json:"start_time"`
	EndTime     time.Time `gorm:"not null" json:"end_time"`
	Location    string    `json:"location"`
	ContactID   string    `json:"contact_id"`
	PropertyID  string    `json:"property_id,omitempty"`
	EventType   string    `gorm:"not null" json:"event_type"`
	Status      string    `gorm:"default:'scheduled'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CalendarStats represents calendar statistics
type CalendarStats struct {
	TotalEvents     int64   `json:"total_events"`
	UpcomingEvents  int64   `json:"upcoming_events"`
	CompletedEvents int64   `json:"completed_events"`
	TodayEvents     int64   `json:"today_events"`
	CompletionRate  float64 `json:"completion_rate"`
}

// BookingCalendarData represents booking information for calendar events
type BookingCalendarData struct {
	PropertyAddress string    `json:"property_address"`
	ShowingTime     time.Time `json:"showing_time"`
	ContactName     string    `json:"contact_name"`
	ContactPhone    string    `json:"contact_phone"`
	ContactEmail    string    `json:"contact_email"`
	SpecialRequests string    `json:"special_requests"`
}

// NewCalendarIntegrationService creates a new calendar service
func NewCalendarIntegrationService(db *gorm.DB) *CalendarIntegrationService {
	return &CalendarIntegrationService{
		db: db,
	}
}

// CreateShowingEvent creates a showing event - EXACT SIGNATURE handlers expect
func (c *CalendarIntegrationService) CreateShowingEvent(bookingData BookingCalendarData) (*CalendarEvent, error) {
	event := CalendarEvent{
		FUBEventID:  fmt.Sprintf("showing_%d", time.Now().Unix()),
		Title:       fmt.Sprintf("Property Showing - %s", bookingData.PropertyAddress),
		Description: c.generateShowingDescription(bookingData),
		StartTime:   bookingData.ShowingTime,
		EndTime:     bookingData.ShowingTime.Add(1 * time.Hour),
		Location:    bookingData.PropertyAddress,
		ContactID:   "contact_placeholder",
		EventType:   "showing",
		Status:      "scheduled",
	}

	if err := c.db.Create(&event).Error; err != nil {
		return nil, fmt.Errorf("failed to create local event record: %v", err)
	}

	log.Printf("📅 Created showing event: %s at %s", bookingData.PropertyAddress, bookingData.ShowingTime.Format("Jan 2 at 3:04 PM"))
	return &event, nil
}

// CreateFollowUpEvent creates a follow-up event - EXACT SIGNATURE handlers expect
func (c *CalendarIntegrationService) CreateFollowUpEvent(contactID, title string, scheduledTime time.Time, description string) (*CalendarEvent, error) {
	event := CalendarEvent{
		FUBEventID:  fmt.Sprintf("followup_%d", time.Now().Unix()),
		Title:       title,
		Description: description,
		StartTime:   scheduledTime,
		EndTime:     scheduledTime.Add(30 * time.Minute),
		ContactID:   contactID,
		EventType:   "follow_up",
		Status:      "scheduled",
	}

	if err := c.db.Create(&event).Error; err != nil {
		return nil, fmt.Errorf("failed to create local follow-up event: %v", err)
	}

	log.Printf("📞 Created follow-up event: %s at %s", title, scheduledTime.Format("Jan 2 at 3:04 PM"))
	return &event, nil
}

// UpdateEventStatus updates the status of a calendar event
func (c *CalendarIntegrationService) UpdateEventStatus(eventID uint, status string) error {
	var event CalendarEvent
	if err := c.db.First(&event, eventID).Error; err != nil {
		return fmt.Errorf("event not found: %v", err)
	}

	event.Status = status
	if status == "completed" {
		event.UpdatedAt = time.Now()
	}

	if err := c.db.Save(&event).Error; err != nil {
		return fmt.Errorf("failed to update local event: %v", err)
	}

	log.Printf("📅 Updated event %d status to: %s", eventID, status)
	return nil
}

// GetUpcomingEvents retrieves upcoming events from the calendar
func (c *CalendarIntegrationService) GetUpcomingEvents(days int) ([]CalendarEvent, error) {
	endDate := time.Now().AddDate(0, 0, days)

	var events []CalendarEvent
	err := c.db.Where("start_time >= ? AND start_time <= ? AND status != 'cancelled'",
		time.Now(), endDate).
		Order("start_time ASC").
		Find(&events).Error

	return events, err
}

// GetTodayEvents retrieves today's events
func (c *CalendarIntegrationService) GetTodayEvents() ([]CalendarEvent, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var events []CalendarEvent
	err := c.db.Where("start_time >= ? AND start_time < ? AND status != 'cancelled'",
		startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&events).Error

	return events, err
}

// SyncWithFUB syncs local calendar events with Follow Up Boss
func (c *CalendarIntegrationService) SyncWithFUB() error {
	log.Println("🔄 Calendar sync with Follow Up Boss - mock implementation")
	return nil
}

// GetCalendarStats returns calendar statistics - EXACT TYPE handlers expect
func (c *CalendarIntegrationService) GetCalendarStats() (*CalendarStats, error) {
	var totalEvents int64
	var upcomingEvents int64
	var completedEvents int64
	var todayEvents int64

	c.db.Model(&CalendarEvent{}).Count(&totalEvents)
	c.db.Model(&CalendarEvent{}).Where("start_time > ? AND status != 'cancelled'", time.Now()).Count(&upcomingEvents)
	c.db.Model(&CalendarEvent{}).Where("status = 'completed'").Count(&completedEvents)

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	c.db.Model(&CalendarEvent{}).Where("start_time >= ? AND start_time < ? AND status != 'cancelled'", startOfDay, endOfDay).Count(&todayEvents)

	completionRate := float64(0)
	if totalEvents > 0 {
		completionRate = (float64(completedEvents) / float64(totalEvents)) * 100
	}

	stats := &CalendarStats{
		TotalEvents:     totalEvents,
		UpcomingEvents:  upcomingEvents,
		CompletedEvents: completedEvents,
		TodayEvents:     todayEvents,
		CompletionRate:  completionRate,
	}

	return stats, nil
}

// ScheduleAutomaticFollowUp schedules a follow-up event - EXACT SIGNATURE handlers expect
func (c *CalendarIntegrationService) ScheduleAutomaticFollowUp(showingEventID uint, hoursAfter int) error {
	var showingEvent CalendarEvent
	if err := c.db.First(&showingEvent, showingEventID).Error; err != nil {
		return fmt.Errorf("showing event not found: %v", err)
	}

	// Convert int hours to time.Duration
	followUpTime := showingEvent.EndTime.Add(time.Duration(hoursAfter) * time.Hour)

	title := fmt.Sprintf("Follow up: %s showing", showingEvent.Location)
	description := fmt.Sprintf("Follow up about the showing at %s", showingEvent.Location)

	_, err := c.CreateFollowUpEvent(showingEvent.ContactID, title, followUpTime, description)
	if err != nil {
		return fmt.Errorf("failed to create follow-up event: %v", err)
	}

	log.Printf("📞 Scheduled automatic follow-up for showing at %s", showingEvent.Location)
	return nil
}

func (c *CalendarIntegrationService) generateShowingDescription(data BookingCalendarData) string {
	description := fmt.Sprintf("Property Showing\n\n")
	description += fmt.Sprintf("Property: %s\n", data.PropertyAddress)
	description += fmt.Sprintf("Contact: %s\n", data.ContactName)
	description += fmt.Sprintf("Phone: %s\n", data.ContactPhone)
	description += fmt.Sprintf("Email: %s\n", data.ContactEmail)

	if data.SpecialRequests != "" {
		description += fmt.Sprintf("\nSpecial Requests: %s\n", data.SpecialRequests)
	}

	return description
}
