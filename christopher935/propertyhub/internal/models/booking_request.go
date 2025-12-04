package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// BookingRequest represents the incoming booking request from the frontend
type BookingRequest struct {
	// Contact Information
	FirstName        string `json:"first_name" binding:"required"`
	LastName         string `json:"last_name" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Phone            string `json:"phone" binding:"required"`
	PreferredContact string `json:"preferred_contact" binding:"required"`

	// Service Type Information (TREC Compliance)
	HasAgent           string `json:"has_agent" binding:"required"` // "yes", "no", "unsure"
	InterestedInBuying bool   `json:"interested_in_buying"`
	ServiceType        string `json:"service_type"` // Determined by frontend logic

	// Property Information
	PropertyID string `json:"property_id" binding:"required"`

	// Scheduling Details
	ShowingDate     string `json:"showing_date" binding:"required"`
	ShowingTime     string `json:"showing_time" binding:"required"`
	DurationMinutes int    `json:"duration_minutes" binding:"required,min=15,max=180"`
	AttendeeCount   int    `json:"attendee_count" binding:"required,min=1,max=10"`
	ShowingType     string `json:"showing_type" binding:"required"`
	Notes           string `json:"notes"`
	Message         string `json:"message"` // Additional message field
	PreQualified    bool   `json:"pre_qualified"`

	// Agent Information (when client has existing agent)
	AgentName  string `json:"agent_name,omitempty"`
	AgentEmail string `json:"agent_email,omitempty"`
	AgentPhone string `json:"agent_phone,omitempty"`

	// Lead information for FUB integration
	Lead *Lead `json:"lead,omitempty"`
}

// BookingResponse represents the response sent back to the frontend
type BookingResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	BookingID   uint   `json:"booking_id,omitempty"`
	FUBLeadID   string `json:"fub_lead_id,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

// Validate performs comprehensive validation on the booking request
func (br *BookingRequest) Validate() error {
	// Validate contact information
	if err := br.validateContactInfo(); err != nil {
		return err
	}

	// Validate service type information (TREC compliance)
	if err := br.validateServiceType(); err != nil {
		return err
	}

	// Validate scheduling details
	if err := br.validateSchedulingDetails(); err != nil {
		return err
	}

	// Validate property information
	if err := br.validatePropertyInfo(); err != nil {
		return err
	}

	return nil
}

func (br *BookingRequest) validateContactInfo() error {
	// Validate first name
	if len(strings.TrimSpace(br.FirstName)) == 0 {
		return errors.New("first name is required")
	}
	if len(br.FirstName) > 50 {
		return errors.New("first name must be less than 50 characters")
	}

	// Validate last name
	if len(strings.TrimSpace(br.LastName)) == 0 {
		return errors.New("last name is required")
	}
	if len(br.LastName) > 50 {
		return errors.New("last name must be less than 50 characters")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(br.Email) {
		return errors.New("invalid email format")
	}
	if len(br.Email) > 100 {
		return errors.New("email must be less than 100 characters")
	}

	// Validate phone number
	cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(br.Phone, "")
	if len(cleanPhone) < 10 || len(cleanPhone) > 15 {
		return errors.New("phone number must be between 10-15 digits")
	}

	// Validate preferred contact method
	validContactMethods := map[string]bool{
		"email": true,
		"phone": true,
		"text":  true,
		"sms":   true,
	}
	if !validContactMethods[strings.ToLower(br.PreferredContact)] {
		return errors.New("preferred contact method must be email, phone, text, or sms")
	}

	return nil
}

func (br *BookingRequest) validateSchedulingDetails() error {
	// Validate preferred date format (YYYY-MM-DD or MM/DD/YYYY)
	dateFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006-1-2",
	}

	var parsedDate time.Time
	var err error
	for _, format := range dateFormats {
		parsedDate, err = time.Parse(format, br.ShowingDate)
		if err == nil {
			break
		}
	}
	if err != nil {
		return errors.New("invalid date format, use YYYY-MM-DD or MM/DD/YYYY")
	}

	// Check if date is in the future
	if parsedDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return errors.New("booking date must be in the future")
	}

	// Check if date is not too far in the future (6 months)
	sixMonthsFromNow := time.Now().AddDate(0, 6, 0)
	if parsedDate.After(sixMonthsFromNow) {
		return errors.New("booking date cannot be more than 6 months in the future")
	}

	// Validate preferred time format (HH:MM AM/PM or HH:MM)
	timeFormats := []string{
		"3:04 PM",
		"15:04",
		"3:04PM",
		"15:04:00",
	}

	for _, format := range timeFormats {
		_, err = time.Parse(format, br.ShowingTime)
		if err == nil {
			break
		}
	}
	if err != nil {
		return errors.New("invalid time format, use HH:MM AM/PM or HH:MM")
	}

	// Validate duration
	if br.DurationMinutes < 15 || br.DurationMinutes > 180 {
		return errors.New("duration must be between 15 and 180 minutes")
	}

	// Validate attendee count
	if br.AttendeeCount < 1 || br.AttendeeCount > 10 {
		return errors.New("attendee count must be between 1 and 10")
	}

	// Validate showing type
	validShowingTypes := map[string]bool{
		"in-person":   true,
		"virtual":     true,
		"self-guided": true,
	}
	if !validShowingTypes[strings.ToLower(br.ShowingType)] {
		return errors.New("showing type must be in-person, virtual, or self-guided")
	}

	// Validate notes length
	if len(br.Notes) > 1000 {
		return errors.New("notes must be less than 1000 characters")
	}

	return nil
}

func (br *BookingRequest) validatePropertyInfo() error {
	// Validate property ID
	if len(strings.TrimSpace(br.PropertyID)) == 0 {
		return errors.New("property ID is required")
	}
	if len(br.PropertyID) > 100 {
		return errors.New("property ID must be less than 100 characters")
	}

	return nil
}

func (br *BookingRequest) validateServiceType() error {
	// Validate has_agent field
	validAgentStatuses := map[string]bool{
		"yes":    true,
		"no":     true,
		"unsure": true,
	}
	if !validAgentStatuses[strings.ToLower(br.HasAgent)] {
		return errors.New("has_agent must be 'yes', 'no', or 'unsure'")
	}

	// Validate service_type if provided (optional, can be determined by backend)
	if br.ServiceType != "" {
		validServiceTypes := map[string]bool{
			"agent_referred":                      true,
			"managed_rental":                      true,
			"external_rental":                     true,
			"managed_rental_with_buyer_interest":  true,
			"external_rental_with_buyer_interest": true,
		}
		if !validServiceTypes[br.ServiceType] {
			return errors.New("invalid service_type")
		}
	}

	return nil
}

// ToBooking converts a BookingRequest to a Booking model for database storage
func (br *BookingRequest) ToBooking(propertyID uint, fubLeadID string) (*Booking, error) {
	// Parse the date and time
	var parsedDate time.Time
	var err error

	dateFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006-1-2",
	}

	for _, format := range dateFormats {
		parsedDate, err = time.Parse(format, br.ShowingDate)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.New("invalid date format")
	}

	// Parse time
	timeFormats := []string{
		"3:04 PM",
		"15:04",
		"3:04PM",
		"15:04:00",
	}

	var parsedTime time.Time
	for _, format := range timeFormats {
		parsedTime, err = time.Parse(format, br.ShowingTime)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.New("invalid time format")
	}

	// Combine date and time
	showingDateTime := time.Date(
		parsedDate.Year(),
		parsedDate.Month(),
		parsedDate.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		0, 0,
		time.Local,
	)

	// Create booking
	booking := &Booking{
		PropertyID:      propertyID,
		FUBLeadID:       fubLeadID,
		ShowingDate:     showingDateTime,
		DurationMinutes: br.DurationMinutes,
		Status:          "scheduled",
		Notes:           br.Notes,
		ShowingType:     strings.ToLower(br.ShowingType),
		AttendeeCount:   br.AttendeeCount,
	}

	return booking, nil
}

// ToFUBLead converts a BookingRequest to FUB lead format
func (br *BookingRequest) ToFUBLead() map[string]interface{} {
	lead := map[string]interface{}{
		"firstName": br.FirstName,
		"lastName":  br.LastName,
		"emails": []map[string]interface{}{
			{
				"value": br.Email,
				"type":  "work",
			},
		},
		"phones": []map[string]interface{}{
			{
				"value": br.Phone,
				"type":  "mobile",
			},
		},
		"source": "Website Booking",
		"tags":   []string{"Property Showing", "Website Lead"},
	}

	// Add hot lead tag if pre-qualified
	if br.PreQualified {
		tags := lead["tags"].([]string)
		tags = append(tags, "Hot Lead", "Pre-Qualified")
		lead["tags"] = tags
	}

	return lead
}
