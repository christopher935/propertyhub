package analytics

import (
	"context"
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/utils"
)

// AnalyticsEngine handles all analytics tracking and processing
type AnalyticsEngine struct {
	eventQueue chan *models.AnalyticsEvent
	ctx        context.Context
	cancel     context.CancelFunc
}

// EventType defines types of analytics events
type EventType string

const (
	EventPageView        EventType = "page_view"
	EventPropertyView    EventType = "property_view"
	EventSearchQuery     EventType = "search_query"
	EventBookingStart    EventType = "booking_start"
	EventBookingComplete EventType = "booking_complete"
	EventBookingCancel   EventType = "booking_cancel"
	EventUserRegister    EventType = "user_register"
	EventUserLogin       EventType = "user_login"
	EventPropertyCreate  EventType = "property_create"
	EventPropertyUpdate  EventType = "property_update"
	EventContactSubmit   EventType = "contact_submit"
	EventCartAbandonment EventType = "cart_abandonment"
	EventFormSubmit      EventType = "form_submit"
	EventButtonClick     EventType = "button_click"
	EventFileDownload    EventType = "file_download"
	EventError           EventType = "error"
)

// EventProperties holds custom properties for events
type EventProperties map[string]interface{}

// UserSegment defines user segments for analytics
type UserSegment string

const (
	SegmentNewUser       UserSegment = "new_user"
	SegmentReturningUser UserSegment = "returning_user"
	SegmentPowerUser     UserSegment = "power_user"
	SegmentPremiumUser   UserSegment = "premium_user"
	SegmentInactiveUser  UserSegment = "inactive_user"
)

// ConversionFunnelStep represents a step in the conversion funnel
type ConversionFunnelStep struct {
	Name        string    `json:"name"`
	Count       int       `json:"count"`
	Percentage  float64   `json:"percentage"`
	DropoffRate float64   `json:"dropoff_rate"`
	Timestamp   time.Time `json:"timestamp"`
}

// ConversionFunnel represents the complete conversion funnel
type ConversionFunnel struct {
	Steps          []ConversionFunnelStep `json:"steps"`
	TotalVisitors  int                    `json:"total_visitors"`
	Conversions    int                    `json:"conversions"`
	ConversionRate float64                `json:"conversion_rate"`
	Period         string                 `json:"period"`
}

// UserBehavior represents aggregated user behavior data
type UserBehavior struct {
	UserID            string        `json:"user_id"`
	SessionID         string        `json:"session_id"`
	TotalPageViews    int           `json:"total_page_views"`
	UniquePages       int           `json:"unique_pages"`
	SessionDuration   time.Duration `json:"session_duration"`
	BounceRate        float64       `json:"bounce_rate"`
	PropertiesViewed  int           `json:"properties_viewed"`
	SearchQueries     int           `json:"search_queries"`
	BookingAttempts   int           `json:"booking_attempts"`
	CompletedBookings int           `json:"completed_bookings"`
	ConversionRate    float64       `json:"conversion_rate"`
	LastActivity      time.Time     `json:"last_activity"`
	UserSegment       UserSegment   `json:"user_segment"`
	DeviceType        string        `json:"device_type"`
	TrafficSource     string        `json:"traffic_source"`
	ReferrerDomain    string        `json:"referrer_domain"`
}

// PropertyPerformance represents property analytics data
type PropertyPerformance struct {
	PropertyID         string    `json:"property_id"`
	Views              int       `json:"views"`
	UniqueViews        int       `json:"unique_views"`
	BookingInquiries   int       `json:"booking_inquiries"`
	CompletedBookings  int       `json:"completed_bookings"`
	ConversionRate     float64   `json:"conversion_rate"`
	AverageViewTime    int       `json:"average_view_time"`
	BounceRate         float64   `json:"bounce_rate"`
	Revenue            float64   `json:"revenue"`
	OccupancyRate      float64   `json:"occupancy_rate"`
	AverageDailyRate   float64   `json:"average_daily_rate"`
	ReviewScore        float64   `json:"review_score"`
	PopularSearchTerms []string  `json:"popular_search_terms"`
	PeakBookingDays    []string  `json:"peak_booking_days"`
	LastUpdated        time.Time `json:"last_updated"`
}

// BusinessMetrics represents high-level business metrics
type BusinessMetrics struct {
	Date                time.Time      `json:"date"`
	TotalRevenue        float64        `json:"total_revenue"`
	BookingRevenue      float64        `json:"booking_revenue"`
	SubscriptionRevenue float64        `json:"subscription_revenue"`
	TotalBookings       int            `json:"total_bookings"`
	NewUsers            int            `json:"new_users"`
	ActiveUsers         int            `json:"active_users"`
	ChurnRate           float64        `json:"churn_rate"`
	CustomerAcquisition float64        `json:"customer_acquisition"`
	AverageOrderValue   float64        `json:"average_order_value"`
	LifetimeValue       float64        `json:"lifetime_value"`
	ConversionRate      float64        `json:"conversion_rate"`
	TrafficSources      map[string]int `json:"traffic_sources"`
	TopPerformingPages  []string       `json:"top_performing_pages"`
	ErrorRate           float64        `json:"error_rate"`
	AverageLoadTime     float64        `json:"average_load_time"`
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine() *AnalyticsEngine {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &AnalyticsEngine{
		eventQueue: make(chan *models.AnalyticsEvent, 10000),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start event processing goroutine
	go engine.processEvents()

	return engine
}

// TrackEvent tracks an analytics event
func (a *AnalyticsEngine) TrackEvent(userID, sessionID string, eventType EventType, properties EventProperties) {
	event := &models.AnalyticsEvent{
		ID:         utils.GenerateID(),
		UserID:     userID,
		SessionID:  sessionID,
		EventType:  string(eventType),
		Properties: models.JSONB(properties),
		Timestamp:  time.Now(),
		IPAddress:  getStringProperty(properties, "ip_address"),
		UserAgent:  getStringProperty(properties, "user_agent"),
		Referrer:   getStringProperty(properties, "referrer"),
		Path:       getStringProperty(properties, "path"),
		CreatedAt:  time.Now(),
	}

	select {
	case a.eventQueue <- event:
	default:
		log.Printf("Analytics event queue full, dropping event: %s", eventType)
	}
}

// TrackPageView tracks a page view event
func (a *AnalyticsEngine) TrackPageView(userID, sessionID, path, referrer, ipAddress, userAgent string) {
	properties := EventProperties{
		"path":       path,
		"referrer":   referrer,
		"ip_address": ipAddress,
		"user_agent": userAgent,
		"timestamp":  time.Now(),
	}

	a.TrackEvent(userID, sessionID, EventPageView, properties)
}

// TrackPropertyView tracks property viewing
func (a *AnalyticsEngine) TrackPropertyView(userID, sessionID, propertyID string, viewDuration int, properties EventProperties) {
	if properties == nil {
		properties = make(EventProperties)
	}

	properties["property_id"] = propertyID
	properties["view_duration"] = viewDuration
	properties["timestamp"] = time.Now()

	a.TrackEvent(userID, sessionID, EventPropertyView, properties)
}

// TrackSearchQuery tracks search queries
func (a *AnalyticsEngine) TrackSearchQuery(userID, sessionID, query string, resultsCount int, filters map[string]interface{}) {
	properties := EventProperties{
		"query":         query,
		"results_count": resultsCount,
		"filters":       filters,
		"timestamp":     time.Now(),
	}

	a.TrackEvent(userID, sessionID, EventSearchQuery, properties)
}

// TrackBookingEvent tracks booking-related events
func (a *AnalyticsEngine) TrackBookingEvent(userID, sessionID, bookingID, propertyID string, eventType EventType, amount float64) {
	properties := EventProperties{
		"booking_id":  bookingID,
		"property_id": propertyID,
		"amount":      amount,
		"timestamp":   time.Now(),
	}

	a.TrackEvent(userID, sessionID, eventType, properties)
}

// TrackConversion tracks conversion events
func (a *AnalyticsEngine) TrackConversion(userID, sessionID, conversionType string, value float64, properties EventProperties) {
	if properties == nil {
		properties = make(EventProperties)
	}

	properties["conversion_type"] = conversionType
	properties["value"] = value
	properties["timestamp"] = time.Now()

	a.TrackEvent(userID, sessionID, EventType("conversion"), properties)
}

// TrackError tracks error events
func (a *AnalyticsEngine) TrackError(userID, sessionID, errorType, errorMessage, stackTrace string) {
	properties := EventProperties{
		"error_type":    errorType,
		"error_message": errorMessage,
		"stack_trace":   stackTrace,
		"timestamp":     time.Now(),
	}

	a.TrackEvent(userID, sessionID, EventError, properties)
}

// processEvents processes events from the queue
func (a *AnalyticsEngine) processEvents() {
	for {
		select {
		case event := <-a.eventQueue:
			// Process the event (save to database, send to external analytics, etc.)
			a.processEvent(event)
		case <-a.ctx.Done():
			return
		}
	}
}

// processEvent processes a single analytics event
func (a *AnalyticsEngine) processEvent(event *models.AnalyticsEvent) {
	// This would typically save to database via repository
	// For now, we'll log it
	log.Printf("Processing analytics event: %s for user %s", event.EventType, event.UserID)

	// Additional processing based on event type
	switch EventType(event.EventType) {
	case EventBookingComplete:
		a.updateConversionMetrics(event)
	case EventPropertyView:
		a.updatePropertyMetrics(event)
	case EventUserRegister:
		a.updateUserAcquisitionMetrics(event)
	case EventError:
		a.processErrorEvent(event)
	}
}

// updateConversionMetrics updates conversion-related metrics
func (a *AnalyticsEngine) updateConversionMetrics(event *models.AnalyticsEvent) {
	// Update conversion funnel data
	log.Printf("Updating conversion metrics for booking completion")
}

// updatePropertyMetrics updates property-related metrics
func (a *AnalyticsEngine) updatePropertyMetrics(event *models.AnalyticsEvent) {
	// Update property view counts, average view time, etc.
	log.Printf("Updating property metrics for property view")
}

// updateUserAcquisitionMetrics updates user acquisition metrics
func (a *AnalyticsEngine) updateUserAcquisitionMetrics(event *models.AnalyticsEvent) {
	// Update user registration and acquisition data
	log.Printf("Updating user acquisition metrics for new registration")
}

// processErrorEvent processes error events for monitoring
func (a *AnalyticsEngine) processErrorEvent(event *models.AnalyticsEvent) {
	// Could send alerts, update error rate metrics, etc.
	log.Printf("Processing error event: %s", event.Properties)
}

// GetUserSegment determines the user segment based on behavior
func (a *AnalyticsEngine) GetUserSegment(userID string, behavior *UserBehavior) UserSegment {
	if behavior == nil {
		return SegmentNewUser
	}

	// Logic to determine user segment
	if behavior.TotalPageViews > 100 && behavior.CompletedBookings > 5 {
		return SegmentPowerUser
	} else if behavior.CompletedBookings > 0 {
		return SegmentReturningUser
	} else if time.Since(behavior.LastActivity) > 30*24*time.Hour {
		return SegmentInactiveUser
	}

	return SegmentNewUser
}

// CalculateConversionRate calculates conversion rate between two metrics
func (a *AnalyticsEngine) CalculateConversionRate(conversions, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(conversions) / float64(total)) * 100
}

// GenerateFunnelAnalysis generates conversion funnel analysis
func (a *AnalyticsEngine) GenerateFunnelAnalysis(startDate, endDate time.Time) *ConversionFunnel {
	// This would query the database for funnel data
	// Placeholder implementation
	steps := []ConversionFunnelStep{
		{Name: "Landing Page", Count: 10000, Percentage: 100.0, DropoffRate: 0.0},
		{Name: "Property Search", Count: 7500, Percentage: 75.0, DropoffRate: 25.0},
		{Name: "Property View", Count: 5000, Percentage: 50.0, DropoffRate: 33.3},
		{Name: "Booking Start", Count: 2500, Percentage: 25.0, DropoffRate: 50.0},
		{Name: "Booking Complete", Count: 1250, Percentage: 12.5, DropoffRate: 50.0},
	}

	return &ConversionFunnel{
		Steps:          steps,
		TotalVisitors:  10000,
		Conversions:    1250,
		ConversionRate: 12.5,
		Period:         fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
	}
}

// GenerateBusinessMetrics generates business metrics for reporting
func (a *AnalyticsEngine) GenerateBusinessMetrics(date time.Time) *BusinessMetrics {
	// This would query the database for actual metrics
	// Placeholder implementation
	return &BusinessMetrics{
		Date:                date,
		TotalRevenue:        125000.00,
		BookingRevenue:      100000.00,
		SubscriptionRevenue: 25000.00,
		TotalBookings:       850,
		NewUsers:            150,
		ActiveUsers:         1200,
		ChurnRate:           2.5,
		CustomerAcquisition: 45.50,
		AverageOrderValue:   147.06,
		LifetimeValue:       892.35,
		ConversionRate:      12.5,
		TrafficSources: map[string]int{
			"organic":  4500,
			"direct":   3200,
			"social":   1800,
			"paid":     2100,
			"referral": 1200,
		},
		TopPerformingPages: []string{
			"/properties/search",
			"/dashboard",
			"/bookings",
		},
		ErrorRate:       0.05,
		AverageLoadTime: 1.2,
	}
}

// Close stops the analytics engine
func (a *AnalyticsEngine) Close() {
	a.cancel()
	close(a.eventQueue)
}

// Helper function to safely get string property
func getStringProperty(properties EventProperties, key string) string {
	if val, ok := properties[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// AbandonmentTracker tracks cart/booking abandonment
type AbandonmentTracker struct {
	engine *AnalyticsEngine
}

// NewAbandonmentTracker creates a new abandonment tracker
func NewAbandonmentTracker(engine *AnalyticsEngine) *AbandonmentTracker {
	return &AbandonmentTracker{engine: engine}
}

// TrackAbandonmentRecovery tracks when users complete after abandonment
func (at *AbandonmentTracker) TrackAbandonmentRecovery(userID, sessionID, bookingID string, recoveryMethod string) {
	properties := EventProperties{
		"booking_id":      bookingID,
		"recovery_method": recoveryMethod,
		"timestamp":       time.Now(),
	}

	at.engine.TrackEvent(userID, sessionID, EventType("abandonment_recovery"), properties)
}
