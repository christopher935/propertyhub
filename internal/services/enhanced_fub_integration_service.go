package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// EnhancedFUBIntegrationService provides additional FUB integration capabilities
// that complement the existing FUBIntegrationHandler
type EnhancedFUBIntegrationService struct {
	db      *gorm.DB
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewEnhancedFUBIntegrationService creates a new enhanced FUB integration service
func NewEnhancedFUBIntegrationService(db *gorm.DB, apiKey string) *EnhancedFUBIntegrationService {
	return &EnhancedFUBIntegrationService{
		db:      db,
		client:  &http.Client{Timeout: 30 * time.Second},
		apiKey:  apiKey,
		baseURL: "https://api.followupboss.com/v1",
	}
}

// AutomatedLeadConversion represents automated booking to lead conversion
type AutomatedLeadConversion struct {
	BookingID       uint                   `json:"booking_id"`
	PropertyID      uint                   `json:"property_id"`
	FUBLeadID       string                 `json:"fub_lead_id"`
	ConversionType  string                 `json:"conversion_type"` // instant, scheduled, conditional
	TriggerEvent    string                 `json:"trigger_event"`
	CustomerData    map[string]interface{} `json:"customer_data"`
	PropertyContext map[string]interface{} `json:"property_context"`
	AutomationRules []string               `json:"automation_rules"`
	Success         bool                   `json:"success"`
	ProcessedAt     time.Time              `json:"processed_at"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
}

// PropertyLifecycleSync represents property lifecycle synchronization with FUB
type PropertyLifecycleSync struct {
	PropertyID      uint      `json:"property_id"`
	MLSID           string    `json:"mls_id"`
	LifecycleStage  string    `json:"lifecycle_stage"`
	PreviousStage   string    `json:"previous_stage"`
	FUBPropertyID   string    `json:"fub_property_id"`
	AffectedLeads   []string  `json:"affected_leads"`
	AutomationTasks []string  `json:"automation_tasks"`
	SyncedAt        time.Time `json:"synced_at"`
	Success         bool      `json:"success"`
}

// CustomerJourneyMapping represents customer journey stages mapped to FUB pipelines
type CustomerJourneyMapping struct {
	CustomerID       string                   `json:"customer_id"`
	FUBLeadID        string                   `json:"fub_lead_id"`
	JourneyStage     string                   `json:"journey_stage"`
	PropertyInterest []uint                   `json:"property_interest"`
	Touchpoints      []map[string]interface{} `json:"touchpoints"`
	Preferences      map[string]interface{}   `json:"preferences"`
	LastActivity     time.Time                `json:"last_activity"`
	NextAction       string                   `json:"next_action"`
	ActionPlanID     string                   `json:"action_plan_id"`
}

// AutoConvertBookingToLead automatically converts bookings to FUB leads based on rules
func (efub *EnhancedFUBIntegrationService) AutoConvertBookingToLead(booking *models.Booking, conversionRules []string) (*AutomatedLeadConversion, error) {
	log.Printf("🤖 Auto-converting booking %d to FUB lead with rules: %v", booking.ID, conversionRules)

	conversion := &AutomatedLeadConversion{
		BookingID:       booking.ID,
		PropertyID:      booking.PropertyID,
		ConversionType:  "instant",
		TriggerEvent:    "booking_created",
		AutomationRules: conversionRules,
		ProcessedAt:     time.Now(),
	}

	// Check if lead already exists
	if booking.FUBLeadID != "" {
		conversion.FUBLeadID = booking.FUBLeadID
		conversion.Success = true
		return conversion, nil
	}

	// Get property details for context
	var property models.Property
	if err := efub.db.First(&property, booking.PropertyID).Error; err != nil {
		conversion.Success = false
		conversion.ErrorMessage = fmt.Sprintf("Failed to get property: %v", err)
		return conversion, err
	}

	// Extract customer data using enhanced extraction
	customerData := efub.extractEnhancedCustomerData(booking, &property)
	conversion.CustomerData = customerData

	// Extract property context
	propertyContext := efub.extractPropertyContext(&property, booking)
	conversion.PropertyContext = propertyContext

	// Apply conversion rules
	leadData := efub.applyConversionRules(customerData, propertyContext, conversionRules)

	// Create FUB lead
	fubLead, err := efub.createEnhancedFUBLead(leadData)
	if err != nil {
		conversion.Success = false
		conversion.ErrorMessage = err.Error()
		return conversion, err
	}

	conversion.FUBLeadID = fubLead.ID
	conversion.Success = true

	// Update booking with FUB lead ID
	booking.FUBLeadID = fubLead.ID
	efub.db.Save(booking)

	// Trigger automated workflows
	go efub.triggerAutomatedWorkflows(fubLead.ID, booking, &property, conversionRules)

	log.Printf("✅ Auto-converted booking %d to FUB lead %s", booking.ID, fubLead.ID)
	return conversion, nil
}

// extractEnhancedCustomerData extracts comprehensive customer data from booking
func (efub *EnhancedFUBIntegrationService) extractEnhancedCustomerData(booking *models.Booking, property *models.Property) map[string]interface{} {
	customerData := map[string]interface{}{
		"source":           "PropertyHub Auto-Conversion",
		"lead_type":        "property_inquiry",
		"inquiry_type":     "showing_request",
		"property_address": property.Address,
		"property_price":   property.Price,
		"showing_date":     booking.ShowingDate,
		"attendee_count":   booking.AttendeeCount,
		"showing_type":     booking.ShowingType,
		"urgency_level":    efub.calculateUrgencyLevel(booking),
		"interest_score":   efub.calculateInterestScore(booking, property),
	}

	// Extract customer preferences from booking data
	if booking.SpecialRequests != "" {
		customerData["special_requests"] = booking.SpecialRequests
		customerData["has_special_needs"] = true
	}

	// Analyze showing timing for customer behavior insights
	customerData["preferred_time_slot"] = efub.analyzePreferredTimeSlot(booking.ShowingDate)
	customerData["advance_booking_days"] = efub.calculateAdvanceBookingDays(booking)

	// Property type preferences
	customerData["property_type_interest"] = property.PropertyType
	if property.Bedrooms != nil {
		customerData["bedroom_preference"] = *property.Bedrooms
	}
	if property.Bathrooms != nil {
		customerData["bathroom_preference"] = *property.Bathrooms
	}

	// Price range analysis
	customerData["price_range_interest"] = efub.categorizePriceRange(property.Price)

	return customerData
}

// extractPropertyContext extracts comprehensive property context
func (efub *EnhancedFUBIntegrationService) extractPropertyContext(property *models.Property, booking *models.Booking) map[string]interface{} {
	context := map[string]interface{}{
		"property_id":    property.ID,
		"mls_id":         property.MLSId,
		"address":        property.Address,
		"city":           property.City,
		"state":          property.State,
		"zip_code":       property.ZipCode,
		"price":          property.Price,
		"property_type":  property.PropertyType,
		"listing_type":   property.ListingType,
		"status":         property.Status,
		"listing_agent":  property.ListingAgent,
		"listing_office": property.ListingOffice,
		"days_on_market": property.DaysOnMarket,
		"year_built":     property.YearBuilt,
	}

	if property.Bedrooms != nil {
		context["bedrooms"] = *property.Bedrooms
	}
	if property.Bathrooms != nil {
		context["bathrooms"] = *property.Bathrooms
	}
	if property.SquareFeet != nil {
		context["square_feet"] = *property.SquareFeet
	}

	// Market analysis
	context["market_position"] = efub.analyzeMarketPosition(property)
	context["price_per_sqft"] = efub.calculatePricePerSqft(property)
	context["property_features"] = efub.extractPropertyFeatures(property)

	// Booking context
	context["booking_urgency"] = efub.calculateBookingUrgency(booking)
	context["showing_preferences"] = efub.extractShowingPreferences(booking)

	return context
}

// applyConversionRules applies conversion rules to customize lead creation
func (efub *EnhancedFUBIntegrationService) applyConversionRules(customerData, propertyContext map[string]interface{}, rules []string) map[string]interface{} {
	leadData := map[string]interface{}{
		"firstName":    "PropertyHub",
		"lastName":     "Lead",
		"email":        fmt.Sprintf("lead_%d@propertyhub.com", time.Now().Unix()),
		"source":       "PropertyHub Auto-Conversion",
		"status":       "new",
		"customFields": make(map[string]interface{}),
		"tags":         []string{"PropertyHub", "Auto-Converted"},
	}

	// Apply each rule
	for _, rule := range rules {
		switch rule {
		case "high_value_property":
			if price, ok := propertyContext["price"].(float64); ok && price > 500000 {
				leadData["tags"] = append(leadData["tags"].([]string), "High Value Property")
				leadData["priority"] = "high"
			}

		case "urgent_showing":
			if urgency, ok := customerData["urgency_level"].(string); ok && urgency == "high" {
				leadData["tags"] = append(leadData["tags"].([]string), "Urgent Showing")
				leadData["priority"] = "high"
			}

		case "luxury_segment":
			if price, ok := propertyContext["price"].(float64); ok && price > 1000000 {
				leadData["tags"] = append(leadData["tags"].([]string), "Luxury Segment")
				leadData["customFields"].(map[string]interface{})["segment"] = "luxury"
			}

		case "first_time_buyer":
			if priceRange, ok := customerData["price_range_interest"].(string); ok && priceRange == "entry_level" {
				leadData["tags"] = append(leadData["tags"].([]string), "First Time Buyer")
				leadData["customFields"].(map[string]interface{})["buyer_type"] = "first_time"
			}

		case "investor_profile":
			if propertyType, ok := propertyContext["property_type"].(string); ok && (propertyType == "investment" || propertyType == "multi_family") {
				leadData["tags"] = append(leadData["tags"].([]string), "Investor")
				leadData["customFields"].(map[string]interface{})["buyer_type"] = "investor"
			}
		}
	}

	// Merge customer and property data into custom fields
	for key, value := range customerData {
		leadData["customFields"].(map[string]interface{})[key] = value
	}

	for key, value := range propertyContext {
		leadData["customFields"].(map[string]interface{})["property_"+key] = value
	}

	return leadData
}

// createEnhancedFUBLead creates a lead in FUB with enhanced data
func (efub *EnhancedFUBIntegrationService) createEnhancedFUBLead(leadData map[string]interface{}) (*FUBLeadResponse, error) {
	jsonData, err := json.Marshal(leadData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lead data: %v", err)
	}

	req, err := http.NewRequest("POST", efub.baseURL+"/people", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(efub.apiKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := efub.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FUB API returned status %d", resp.StatusCode)
	}

	var fubLead FUBLeadResponse
	if err := json.NewDecoder(resp.Body).Decode(&fubLead); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &fubLead, nil
}

// triggerAutomatedWorkflows triggers automated workflows based on conversion rules
func (efub *EnhancedFUBIntegrationService) triggerAutomatedWorkflows(leadID string, booking *models.Booking, property *models.Property, rules []string) {
	log.Printf("🔄 Triggering automated workflows for lead %s", leadID)

	// Create automated tasks based on rules
	for _, rule := range rules {
		switch rule {
		case "high_value_property":
			efub.createHighValuePropertyTask(leadID, booking, property)
		case "urgent_showing":
			efub.createUrgentShowingTask(leadID, booking, property)
		case "luxury_segment":
			efub.createLuxurySegmentTask(leadID, booking, property)
		case "first_time_buyer":
			efub.createFirstTimeBuyerTask(leadID, booking, property)
		case "investor_profile":
			efub.createInvestorProfileTask(leadID, booking, property)
		}
	}

	// Create standard follow-up task
	efub.createStandardFollowUpTask(leadID, booking, property)
}

// SyncPropertyLifecycle syncs property lifecycle changes with FUB
func (efub *EnhancedFUBIntegrationService) SyncPropertyLifecycle(propertyID uint, newStage string, previousStage string) (*PropertyLifecycleSync, error) {
	log.Printf("🔄 Syncing property lifecycle: %d %s -> %s", propertyID, previousStage, newStage)

	var property models.Property
	if err := efub.db.First(&property, propertyID).Error; err != nil {
		return nil, fmt.Errorf("property not found: %v", err)
	}

	sync := &PropertyLifecycleSync{
		PropertyID:     propertyID,
		MLSID:          property.MLSId,
		LifecycleStage: newStage,
		PreviousStage:  previousStage,
		SyncedAt:       time.Now(),
	}

	// Find all leads associated with this property
	var bookings []models.Booking
	efub.db.Where("property_id = ? AND fub_lead_id != ''", propertyID).Find(&bookings)

	affectedLeads := []string{}
	for _, booking := range bookings {
		affectedLeads = append(affectedLeads, booking.FUBLeadID)
	}
	sync.AffectedLeads = affectedLeads

	// Update each affected lead
	automationTasks := []string{}
	for _, leadID := range affectedLeads {
		tasks, err := efub.updateLeadForLifecycleChange(leadID, sync)
		if err != nil {
			log.Printf("⚠️ Failed to update lead %s: %v", leadID, err)
		} else {
			automationTasks = append(automationTasks, tasks...)
		}
	}

	sync.AutomationTasks = automationTasks
	sync.Success = true

	log.Printf("✅ Synced property lifecycle for %d leads", len(affectedLeads))
	return sync, nil
}

// updateLeadForLifecycleChange updates a lead based on property lifecycle change
func (efub *EnhancedFUBIntegrationService) updateLeadForLifecycleChange(leadID string, sync *PropertyLifecycleSync) ([]string, error) {
	// Update lead custom fields
	updateData := map[string]interface{}{
		"customFields": map[string]interface{}{
			"property_lifecycle_stage": sync.LifecycleStage,
			"property_previous_stage":  sync.PreviousStage,
			"lifecycle_updated_at":     sync.SyncedAt.Format(time.RFC3339),
		},
	}

	// Send update to FUB
	jsonData, _ := json.Marshal(updateData)
	req, err := http.NewRequest("PUT", efub.baseURL+"/people/"+leadID, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(efub.apiKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := efub.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Create lifecycle-specific tasks
	tasks := []string{}
	switch sync.LifecycleStage {
	case "active":
		tasks = append(tasks, "property_went_live")
	case "pending":
		tasks = append(tasks, "property_under_contract")
	case "sold":
		tasks = append(tasks, "property_sold_followup")
	case "withdrawn":
		tasks = append(tasks, "property_withdrawn_recovery")
	}

	return tasks, nil
}

// Helper functions for data analysis
func (efub *EnhancedFUBIntegrationService) calculateUrgencyLevel(booking *models.Booking) string {
	// Calculate urgency based on showing date proximity
	daysUntilShowing := int(booking.ShowingDate.Sub(time.Now()).Hours() / 24)

	if daysUntilShowing <= 1 {
		return "high"
	} else if daysUntilShowing <= 3 {
		return "medium"
	}
	return "low"
}

func (efub *EnhancedFUBIntegrationService) calculateInterestScore(booking *models.Booking, property *models.Property) int {
	score := 50 // Base score

	// Adjust based on attendee count
	score += booking.AttendeeCount * 10

	// Adjust based on special requests
	if booking.SpecialRequests != "" {
		score += 15
	}

	// Adjust based on property price
	if property.Price > 500000 {
		score += 20
	}

	// Adjust based on advance booking
	daysAdvance := int(booking.ShowingDate.Sub(booking.CreatedAt).Hours() / 24)
	if daysAdvance > 7 {
		score += 10 // Planned ahead
	} else if daysAdvance < 2 {
		score += 20 // Urgent need
	}

	if score > 100 {
		score = 100
	}

	return score
}

func (efub *EnhancedFUBIntegrationService) analyzePreferredTimeSlot(showingDate time.Time) string {
	hour := showingDate.Hour()

	if hour < 12 {
		return "morning"
	} else if hour < 17 {
		return "afternoon"
	}
	return "evening"
}

func (efub *EnhancedFUBIntegrationService) calculateAdvanceBookingDays(booking *models.Booking) int {
	return int(booking.ShowingDate.Sub(booking.CreatedAt).Hours() / 24)
}

func (efub *EnhancedFUBIntegrationService) categorizePriceRange(price float64) string {
	if price < 200000 {
		return "entry_level"
	} else if price < 500000 {
		return "mid_market"
	} else if price < 1000000 {
		return "upper_market"
	}
	return "luxury"
}

func (efub *EnhancedFUBIntegrationService) analyzeMarketPosition(property *models.Property) string {
	if property.DaysOnMarket == nil {
		return "new_listing"
	}

	days := *property.DaysOnMarket
	if days < 7 {
		return "hot_property"
	} else if days < 30 {
		return "active_market"
	} else if days < 90 {
		return "slow_market"
	}
	return "stale_listing"
}

func (efub *EnhancedFUBIntegrationService) calculatePricePerSqft(property *models.Property) float64 {
	if property.SquareFeet == nil || *property.SquareFeet == 0 {
		return 0
	}
	return property.Price / float64(*property.SquareFeet)
}

func (efub *EnhancedFUBIntegrationService) extractPropertyFeatures(property *models.Property) []string {
	features := []string{}

	if property.Bedrooms != nil && *property.Bedrooms >= 4 {
		features = append(features, "large_family_home")
	}

	if property.Bathrooms != nil && *property.Bathrooms >= 3 {
		features = append(features, "multiple_bathrooms")
	}

	if property.SquareFeet != nil && *property.SquareFeet > 3000 {
		features = append(features, "spacious")
	}

	if property.YearBuilt > 2010 {
		features = append(features, "modern_construction")
	} else if property.YearBuilt < 1980 {
		features = append(features, "vintage_character")
	}

	return features
}

func (efub *EnhancedFUBIntegrationService) calculateBookingUrgency(booking *models.Booking) string {
	return efub.calculateUrgencyLevel(booking)
}

func (efub *EnhancedFUBIntegrationService) extractShowingPreferences(booking *models.Booking) map[string]interface{} {
	return map[string]interface{}{
		"showing_type":         booking.ShowingType,
		"attendee_count":       booking.AttendeeCount,
		"duration_minutes":     booking.DurationMinutes,
		"has_special_requests": booking.SpecialRequests != "",
	}
}

// Task creation helper functions
func (efub *EnhancedFUBIntegrationService) createHighValuePropertyTask(leadID string, booking *models.Booking, property *models.Property) {
	taskData := map[string]interface{}{
		"leadId":      leadID,
		"title":       fmt.Sprintf("High Value Property Follow-up - %s", property.Address),
		"description": fmt.Sprintf("High value property ($%.0f) showing scheduled. Ensure premium service delivery.", property.Price),
		"dueDate":     booking.ShowingDate.Add(-2 * time.Hour).Format(time.RFC3339),
		"priority":    "high",
		"type":        "call",
	}

	efub.createFUBTask(taskData)
}

func (efub *EnhancedFUBIntegrationService) createUrgentShowingTask(leadID string, booking *models.Booking, property *models.Property) {
	taskData := map[string]interface{}{
		"leadId":      leadID,
		"title":       fmt.Sprintf("URGENT: Showing Today - %s", property.Address),
		"description": "Urgent showing request. Contact immediately to confirm details.",
		"dueDate":     time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		"priority":    "high",
		"type":        "call",
	}

	efub.createFUBTask(taskData)
}

func (efub *EnhancedFUBIntegrationService) createLuxurySegmentTask(leadID string, booking *models.Booking, property *models.Property) {
	taskData := map[string]interface{}{
		"leadId":      leadID,
		"title":       fmt.Sprintf("Luxury Client Service - %s", property.Address),
		"description": "Luxury segment client. Ensure white-glove service and premium experience.",
		"dueDate":     booking.ShowingDate.Add(-4 * time.Hour).Format(time.RFC3339),
		"priority":    "high",
		"type":        "call",
	}

	efub.createFUBTask(taskData)
}

func (efub *EnhancedFUBIntegrationService) createFirstTimeBuyerTask(leadID string, booking *models.Booking, property *models.Property) {
	taskData := map[string]interface{}{
		"leadId":      leadID,
		"title":       fmt.Sprintf("First-Time Buyer Education - %s", property.Address),
		"description": "Potential first-time buyer. Provide educational materials and process guidance.",
		"dueDate":     booking.ShowingDate.Add(-24 * time.Hour).Format(time.RFC3339),
		"priority":    "medium",
		"type":        "email",
	}

	efub.createFUBTask(taskData)
}

func (efub *EnhancedFUBIntegrationService) createInvestorProfileTask(leadID string, booking *models.Booking, property *models.Property) {
	taskData := map[string]interface{}{
		"leadId":      leadID,
		"title":       fmt.Sprintf("Investor Analysis - %s", property.Address),
		"description": "Potential investor. Prepare ROI analysis and investment metrics.",
		"dueDate":     booking.ShowingDate.Add(-12 * time.Hour).Format(time.RFC3339),
		"priority":    "medium",
		"type":        "email",
	}

	efub.createFUBTask(taskData)
}

func (efub *EnhancedFUBIntegrationService) createStandardFollowUpTask(leadID string, booking *models.Booking, property *models.Property) {
	taskData := map[string]interface{}{
		"leadId":      leadID,
		"title":       fmt.Sprintf("Standard Follow-up - %s", property.Address),
		"description": fmt.Sprintf("Follow up on showing request for %s scheduled %s", property.Address, booking.ShowingDate.Format("Jan 2, 3:04 PM")),
		"dueDate":     booking.ShowingDate.Add(-6 * time.Hour).Format(time.RFC3339),
		"priority":    "medium",
		"type":        "call",
	}

	efub.createFUBTask(taskData)
}

func (efub *EnhancedFUBIntegrationService) createFUBTask(taskData map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	operation := func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", efub.baseURL+"/tasks", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(efub.apiKey+":")))
		return efub.client.Do(req)
	}

	resp, err := WithRetry(ctx, DefaultRetryConfig, operation)
	if err != nil {
		return fmt.Errorf("failed to create FUB task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FUB task creation failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// FUBLeadResponse represents the response from FUB lead creation
type FUBLeadResponse struct {
	ID           string                 `json:"id"`
	FirstName    string                 `json:"firstName"`
	LastName     string                 `json:"lastName"`
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone"`
	Source       string                 `json:"source"`
	Tags         []string               `json:"tags"`
	CustomFields map[string]interface{} `json:"customFields"`
	CreatedDate  time.Time              `json:"createdDate"`
	UpdatedDate  time.Time              `json:"updatedDate"`
}
