package booking

import (
	"errors"
	"fmt"
	"time"

	"chrisgross-ctrl-project/internal/models"
)

// ShowingEngine handles all showing/appointment operations and TREC compliance
type ShowingEngine struct {
	trecComplianceEnabled bool
	maxAdvanceBookingDays int
	cancellationGraceDays int
}

// ShowingStatus represents showing appointment states
type ShowingStatus string

const (
	StatusScheduled   ShowingStatus = "scheduled"
	StatusConfirmed   ShowingStatus = "confirmed"
	StatusCompleted   ShowingStatus = "completed"
	StatusCancelled   ShowingStatus = "cancelled"
	StatusNoShow      ShowingStatus = "no_show"
	StatusRescheduled ShowingStatus = "rescheduled"
)

// TRECDisclosure represents required TREC disclosures for real estate showings
type TRECDisclosure struct {
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Required     bool      `json:"required"`
	Acknowledged bool      `json:"acknowledged"`
	SignedAt     time.Time `json:"signed_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}

// ShowingValidationError represents showing validation errors
type ShowingValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ShowingValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Common showing errors
var (
	ErrPropertyNotAvailable    = errors.New("property not available for selected time slot")
	ErrInvalidDateTime         = errors.New("invalid date or time")
	ErrPastDateBooking         = errors.New("cannot schedule showings in the past")
	ErrAdvanceBookingLimit     = errors.New("booking too far in advance")
	ErrAttendeeLimit           = errors.New("attendee limit exceeded")
	ErrTRECComplianceRequired  = errors.New("TREC compliance acknowledgment required")
	ErrCancellationNotAllowed  = errors.New("cancellation not allowed")
	ErrInvalidDuration         = errors.New("invalid showing duration")
	ErrConflictingAppointment  = errors.New("time slot conflicts with existing appointment")
)

// NewShowingEngine creates a new showing engine
func NewShowingEngine(trecEnabled bool, maxAdvanceDays, cancellationGraceDays int) *ShowingEngine {
	return &ShowingEngine{
		trecComplianceEnabled: trecEnabled,
		maxAdvanceBookingDays: maxAdvanceDays,
		cancellationGraceDays: cancellationGraceDays,
	}
}

// ValidateShowingRequest validates a showing request
func (s *ShowingEngine) ValidateShowingRequest(request *ShowingRequest, property *models.Property) ([]ShowingValidationError, error) {
	var errors []ShowingValidationError
	
	// Validate date/time
	if err := s.validateDateTime(request.ShowingDate, request.ShowingTime); err != nil {
		errors = append(errors, ShowingValidationError{
			Field:   "datetime",
			Code:    "INVALID_DATETIME",
			Message: err.Error(),
		})
	}
	
	// Validate availability
	showingDateTime := s.combineDateTime(request.ShowingDate, request.ShowingTime)
	if !s.isPropertyAvailable(property, showingDateTime, request.DurationMinutes) {
		errors = append(errors, ShowingValidationError{
			Field:   "availability",
			Code:    "NOT_AVAILABLE",
			Message: "Property not available for selected time slot",
		})
	}
	
	// Validate duration
	if err := s.validateDuration(request.DurationMinutes); err != nil {
		errors = append(errors, ShowingValidationError{
			Field:   "duration",
			Code:    "INVALID_DURATION",
			Message: err.Error(),
		})
	}
	
	// Validate attendee count (1-10 people for showings)
	if request.AttendeeCount < 1 || request.AttendeeCount > 10 {
		errors = append(errors, ShowingValidationError{
			Field:   "attendee_count",
			Code:    "ATTENDEE_LIMIT_EXCEEDED",
			Message: "Attendee count must be between 1 and 10",
		})
	}
	
	// TREC compliance validation
	if s.trecComplianceEnabled && !request.TRECCompliant {
		errors = append(errors, ShowingValidationError{
			Field:   "trec_compliance",
			Code:    "TREC_REQUIRED",
			Message: "TREC disclosure acknowledgment required",
		})
	}
	
	if len(errors) > 0 {
		return errors, fmt.Errorf("showing validation failed")
	}
	
	return nil, nil
}

// CreateShowing creates a new showing appointment
func (s *ShowingEngine) CreateShowing(request *ShowingRequest, property *models.Property, leadID string) (*models.Booking, error) {
	// Validate the request
	if validationErrors, err := s.ValidateShowingRequest(request, property); err != nil {
		return nil, fmt.Errorf("validation failed: %v", validationErrors)
	}
	
	// Combine date and time
	showingDateTime := s.combineDateTime(request.ShowingDate, request.ShowingTime)
	
	// Create showing appointment
	showing := &models.Booking{
		PropertyID:      property.ID,
		FUBLeadID:       leadID,
		ShowingDate:     showingDateTime,
		DurationMinutes: request.DurationMinutes,
		Status:          string(StatusScheduled),
		ShowingType:     request.ShowingType,
		AttendeeCount:   request.AttendeeCount,
		Notes:           request.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	return showing, nil
}

// CheckAvailability checks property availability for a specific time slot
func (s *ShowingEngine) CheckAvailability(property *models.Property, showingDate time.Time, showingTime string, durationMinutes int) (*AvailabilityResponse, error) {
	if err := s.validateDateTime(showingDate, showingTime); err != nil {
		return nil, err
	}
	
	showingDateTime := s.combineDateTime(showingDate, showingTime)
	
	// Check blackout dates (agent unavailability)
	blackoutDates := s.getBlackoutDates(property.ID, showingDateTime)
	if len(blackoutDates) > 0 {
		return &AvailabilityResponse{
			Available:     false,
			Reason:        "Agent unavailable",
			BlackoutDates: blackoutDates,
		}, nil
	}
	
	// Check existing appointments
	conflictingShowings := s.getConflictingShowings(property.ID, showingDateTime, durationMinutes)
	if len(conflictingShowings) > 0 {
		return &AvailabilityResponse{
			Available:           false,
			Reason:              "Time slot already booked",
			ConflictingShowings: conflictingShowings,
		}, nil
	}
	
	return &AvailabilityResponse{
		Available: true,
	}, nil
}

// ConfirmShowing confirms a scheduled showing
func (s *ShowingEngine) ConfirmShowing(showingID, leadID string) error {
	// This would typically fetch showing from repository
	// Send confirmation emails
	// Update showing status to confirmed
	
	return nil
}

// CancelShowing cancels a showing appointment
func (s *ShowingEngine) CancelShowing(showingID, leadID, reason string) (*CancellationResult, error) {
	// This would fetch showing from repository
	showing := &models.Booking{} // placeholder
	
	// Check if cancellation is allowed (e.g., must be more than 2 hours before showing)
	hoursUntilShowing := showing.ShowingDate.Sub(time.Now()).Hours()
	if hoursUntilShowing < 2 {
		return nil, ErrCancellationNotAllowed
	}
	
	result := &CancellationResult{
		ShowingID:       showingID,
		CancellationFee: 0, // No fees for showing cancellations
		ProcessTime:     "immediate",
	}
	
	return result, nil
}

// RescheduleShowing reschedules an existing showing to a new time
func (s *ShowingEngine) RescheduleShowing(showingID, leadID string, newDate time.Time, newTime string) error {
	// Validate new date/time
	if err := s.validateDateTime(newDate, newTime); err != nil {
		return err
	}
	
	// Check availability for new time slot
	// Update showing record
	// Send reschedule notifications
	
	return nil
}

// validateDateTime validates showing date and time
func (s *ShowingEngine) validateDateTime(showingDate time.Time, showingTime string) error {
	now := time.Now()
	
	// Check if date is in the past
	if showingDate.Before(now.Truncate(24 * time.Hour)) {
		return ErrPastDateBooking
	}
	
	// Check advance booking limit (e.g., 180 days)
	maxAdvanceDate := now.AddDate(0, 0, s.maxAdvanceBookingDays)
	if showingDate.After(maxAdvanceDate) {
		return ErrAdvanceBookingLimit
	}
	
	// Validate time format
	if showingTime == "" {
		return ErrInvalidDateTime
	}
	
	return nil
}

// validateDuration validates showing duration (15-180 minutes)
func (s *ShowingEngine) validateDuration(durationMinutes int) error {
	if durationMinutes < 15 || durationMinutes > 180 {
		return ErrInvalidDuration
	}
	
	return nil
}

// isPropertyAvailable checks if property is available for showing at specified time
func (s *ShowingEngine) isPropertyAvailable(property *models.Property, showingDateTime time.Time, durationMinutes int) bool {
	// Check blackout dates
	blackoutDates := s.getBlackoutDates(property.ID, showingDateTime)
	if len(blackoutDates) > 0 {
		return false
	}
	
	// Check existing showings
	conflictingShowings := s.getConflictingShowings(property.ID, showingDateTime, durationMinutes)
	return len(conflictingShowings) == 0
}

// combineDateTime combines date and time into a single datetime
func (s *ShowingEngine) combineDateTime(date time.Time, timeStr string) time.Time {
	// Parse time string and combine with date
	// This is a simplified implementation
	return date
}

// generateTRECDisclosures generates required TREC disclosures for showings
func (s *ShowingEngine) generateTRECDisclosures(request *ShowingRequest) string {
	disclosures := []TRECDisclosure{
		{
			Type:     "property_access",
			Title:    "Property Access Disclosure",
			Content:  "The property showing is for informational purposes only...",
			Required: true,
		},
		{
			Type:     "representation",
			Title:    "Agency Representation Disclosure",
			Content:  "Landlords of Texas represents the property owner/landlord...",
			Required: true,
		},
	}
	
	// Convert to JSON string
	return fmt.Sprintf("%v", disclosures)
}

// getBlackoutDates gets blackout dates (agent unavailability) for property
func (s *ShowingEngine) getBlackoutDates(propertyID uint, showingDateTime time.Time) []time.Time {
	// This would query database for agent blackout dates
	// Placeholder implementation
	return []time.Time{}
}

// getConflictingShowings gets existing showings that conflict with time slot
func (s *ShowingEngine) getConflictingShowings(propertyID uint, showingDateTime time.Time, durationMinutes int) []string {
	// This would query database for conflicting showings
	// Placeholder implementation
	return []string{}
}

// Supporting types for showing operations
type ShowingRequest struct {
	PropertyID      string    `json:"property_id"`
	ShowingDate     time.Time `json:"showing_date"`
	ShowingTime     string    `json:"showing_time"`
	DurationMinutes int       `json:"duration_minutes"` // 15, 30, 45, 60, 90, 120, 180
	ShowingType     string    `json:"showing_type"`     // in-person, virtual, self-guided
	AttendeeCount   int       `json:"attendee_count"`   // 1-10 people viewing
	Notes           string    `json:"notes"`
	TRECCompliant   bool      `json:"trec_compliant"`
}

type AvailabilityResponse struct {
	Available           bool              `json:"available"`
	Reason              string            `json:"reason,omitempty"`
	BlackoutDates       []time.Time       `json:"blackout_dates,omitempty"`
	ConflictingShowings []string          `json:"conflicting_showings,omitempty"`
}

type CancellationResult struct {
	ShowingID       string  `json:"showing_id"`
	CancellationFee float64 `json:"cancellation_fee"`
	ProcessTime     string  `json:"process_time"`
}
