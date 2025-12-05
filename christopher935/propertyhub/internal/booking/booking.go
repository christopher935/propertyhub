package booking

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/utils"
)

// BookingEngine handles all booking operations and TREC compliance
type BookingEngine struct {
	trecComplianceEnabled bool
	maxAdvanceBookingDays int
	cancellationGraceDays int
}

// BookingStatus represents booking states
type BookingStatus string

const (
	StatusPending    BookingStatus = "pending"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusCheckedIn  BookingStatus = "checked_in"
	StatusCheckedOut BookingStatus = "checked_out"
	StatusCancelled  BookingStatus = "cancelled"
	StatusNoShow     BookingStatus = "no_show"
	StatusRefunded   BookingStatus = "refunded"
)

// PaymentStatus represents payment states
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentPaid      PaymentStatus = "paid"
	PaymentPartial   PaymentStatus = "partial"
	PaymentRefunded  PaymentStatus = "refunded"
	PaymentFailed    PaymentStatus = "failed"
)

// TRECDisclosure represents required TREC disclosures
type TRECDisclosure struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Required    bool      `json:"required"`
	Acknowledged bool     `json:"acknowledged"`
	SignedAt    time.Time `json:"signed_at"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
}

// BookingValidationError represents booking validation errors
type BookingValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BookingValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Common booking errors
var (
	ErrPropertyNotAvailable = errors.New("property not available for selected dates")
	ErrInvalidDateRange     = errors.New("invalid date range")
	ErrMinimumStayRequired  = errors.New("minimum stay requirement not met")
	ErrMaximumStayExceeded  = errors.New("maximum stay exceeded")
	ErrAdvanceBookingLimit  = errors.New("booking too far in advance")
	ErrPastDateBooking      = errors.New("cannot book dates in the past")
	ErrGuestLimitExceeded   = errors.New("guest limit exceeded")
	ErrTRECComplianceRequired = errors.New("TREC compliance acknowledgment required")
	ErrInsufficientPayment  = errors.New("insufficient payment amount")
	ErrCancellationNotAllowed = errors.New("cancellation not allowed")
)

// NewBookingEngine creates a new booking engine
func NewBookingEngine(trecEnabled bool, maxAdvanceDays, cancellationGraceDays int) *BookingEngine {
	return &BookingEngine{
		trecComplianceEnabled: trecEnabled,
		maxAdvanceBookingDays: maxAdvanceDays,
		cancellationGraceDays: cancellationGraceDays,
	}
}

// ValidateBookingRequest validates a booking request
func (b *BookingEngine) ValidateBookingRequest(request *BookingRequest, property *models.Property) ([]BookingValidationError, error) {
	var errors []BookingValidationError
	
	// Validate dates
	if err := b.validateDates(request.CheckIn, request.CheckOut); err != nil {
		errors = append(errors, BookingValidationError{
			Field:   "dates",
			Code:    "INVALID_DATES",
			Message: err.Error(),
		})
	}
	
	// Validate availability
	if !b.isPropertyAvailable(property, request.CheckIn, request.CheckOut) {
		errors = append(errors, BookingValidationError{
			Field:   "availability",
			Code:    "NOT_AVAILABLE",
			Message: "Property not available for selected dates",
		})
	}
	
	// Validate stay length
	if err := b.validateStayLength(request.CheckIn, request.CheckOut, property); err != nil {
		errors = append(errors, BookingValidationError{
			Field:   "stay_length",
			Code:    "INVALID_STAY_LENGTH",
			Message: err.Error(),
		})
	}
	
	// Validate guest count
	if request.Guests > property.MaxGuests {
		errors = append(errors, BookingValidationError{
			Field:   "guests",
			Code:    "GUEST_LIMIT_EXCEEDED",
			Message: fmt.Sprintf("Maximum %d guests allowed", property.MaxGuests),
		})
	}
	
	// TREC compliance validation
	if b.trecComplianceEnabled && !request.TRECCompliant {
		errors = append(errors, BookingValidationError{
			Field:   "trec_compliance",
			Code:    "TREC_REQUIRED",
			Message: "TREC disclosure acknowledgment required",
		})
	}
	
	// Validate payment
	expectedAmount := b.calculateTotalAmount(request.CheckIn, request.CheckOut, property)
	if request.PaymentAmount < expectedAmount {
		errors = append(errors, BookingValidationError{
			Field:   "payment",
			Code:    "INSUFFICIENT_PAYMENT",
			Message: fmt.Sprintf("Payment of $%.2f required", expectedAmount),
		})
	}
	
	if len(errors) > 0 {
		return errors, fmt.Errorf("booking validation failed")
	}
	
	return nil, nil
}

// CreateBooking creates a new booking
func (b *BookingEngine) CreateBooking(request *BookingRequest, property *models.Property, user *models.User) (*models.Booking, error) {
	// Validate the request
	if validationErrors, err := b.ValidateBookingRequest(request, property); err != nil {
		return nil, fmt.Errorf("validation failed: %v", validationErrors)
	}
	
	// Calculate pricing
	nights := int(request.CheckOut.Sub(request.CheckIn).Hours() / 24)
	baseAmount := property.BasePrice * float64(nights)
	cleaningFee := property.CleaningFee
	serviceFee := baseAmount * 0.15 // 15% service fee
	totalAmount := baseAmount + cleaningFee + serviceFee
	
	// Create booking
	booking := &models.Booking{
		ID:            utils.GenerateID(),
		PropertyID:    property.ID,
		UserID:        user.ID,
		CheckIn:       request.CheckIn,
		CheckOut:      request.CheckOut,
		Guests:        request.Guests,
		Status:        string(StatusPending),
		PaymentStatus: string(PaymentPending),
		BaseAmount:    baseAmount,
		CleaningFee:   cleaningFee,
		ServiceFee:    serviceFee,
		TotalAmount:   totalAmount,
		PaidAmount:    request.PaymentAmount,
		Currency:      "USD",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// Generate booking reference
	booking.BookingReference = utils.GenerateBookingReference()
	
	// Add TREC compliance data if enabled
	if b.trecComplianceEnabled {
		booking.TRECCompliant = true
		booking.TRECDisclosures = b.generateTRECDisclosures(request)
	}
	
	return booking, nil
}

// CheckAvailability checks property availability for date range
func (b *BookingEngine) CheckAvailability(property *models.Property, checkIn, checkOut time.Time) (*AvailabilityResponse, error) {
	if err := b.validateDates(checkIn, checkOut); err != nil {
		return nil, err
	}
	
	// Check blackout dates
	blackoutDates := b.getBlackoutDates(property.ID, checkIn, checkOut)
	if len(blackoutDates) > 0 {
		return &AvailabilityResponse{
			Available:     false,
			Reason:        "Blackout dates",
			BlackoutDates: blackoutDates,
		}, nil
	}
	
	// Check existing bookings
	conflictingBookings := b.getConflictingBookings(property.ID, checkIn, checkOut)
	if len(conflictingBookings) > 0 {
		return &AvailabilityResponse{
			Available:          false,
			Reason:            "Already booked",
			ConflictingBookings: conflictingBookings,
		}, nil
	}
	
	// Calculate pricing
	pricing := b.calculatePricing(property, checkIn, checkOut)
	
	return &AvailabilityResponse{
		Available: true,
		Pricing:   pricing,
	}, nil
}

// ConfirmBooking confirms a pending booking
func (b *BookingEngine) ConfirmBooking(bookingID, userID string) error {
	// This would typically fetch booking from repository
	// For now, placeholder logic
	
	// Verify payment
	// Send confirmation emails
	// Update booking status
	
	return nil
}

// CancelBooking cancels a booking based on cancellation policy
func (b *BookingEngine) CancelBooking(bookingID, userID, reason string) (*CancellationResult, error) {
	// This would fetch booking from repository
	booking := &models.Booking{} // placeholder
	
	// Check cancellation policy
	policy := b.getCancellationPolicy(booking.CheckIn)
	if !policy.AllowsCancellation {
		return nil, ErrCancellationNotAllowed
	}
	
	// Calculate refund amount
	refundAmount := b.calculateRefund(booking, policy)
	
	result := &CancellationResult{
		BookingID:    bookingID,
		RefundAmount: refundAmount,
		RefundMethod: "original_payment",
		ProcessTime:  "3-5 business days",
		Policy:       policy,
	}
	
	return result, nil
}

// ProcessCheckIn processes guest check-in
func (b *BookingEngine) ProcessCheckIn(bookingID, userID string, checkInData *CheckInData) error {
	// Validate check-in time
	// Verify identity if required
	// Generate access codes
	// Send check-in instructions
	// Update booking status
	
	return nil
}

// ProcessCheckOut processes guest check-out
func (b *BookingEngine) ProcessCheckOut(bookingID, userID string, checkOutData *CheckOutData) error {
	// Validate check-out time
	// Process damages assessment
	// Calculate additional charges
	// Process security deposit return
	// Send checkout confirmation
	// Update booking status
	
	return nil
}

// validateDates validates check-in and check-out dates
func (b *BookingEngine) validateDates(checkIn, checkOut time.Time) error {
	now := time.Now()
	
	// Check if dates are in the past
	if checkIn.Before(now) {
		return ErrPastDateBooking
	}
	
	// Check if check-out is after check-in
	if !checkOut.After(checkIn) {
		return ErrInvalidDateRange
	}
	
	// Check advance booking limit
	maxAdvanceDate := now.AddDate(0, 0, b.maxAdvanceBookingDays)
	if checkIn.After(maxAdvanceDate) {
		return ErrAdvanceBookingLimit
	}
	
	return nil
}

// validateStayLength validates minimum and maximum stay requirements
func (b *BookingEngine) validateStayLength(checkIn, checkOut time.Time, property *models.Property) error {
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	
	if property.MinStayNights > 0 && nights < property.MinStayNights {
		return ErrMinimumStayRequired
	}
	
	if property.MaxStayNights > 0 && nights > property.MaxStayNights {
		return ErrMaximumStayExceeded
	}
	
	return nil
}

// isPropertyAvailable checks if property is available for dates
func (b *BookingEngine) isPropertyAvailable(property *models.Property, checkIn, checkOut time.Time) bool {
	// Check blackout dates
	blackoutDates := b.getBlackoutDates(property.ID, checkIn, checkOut)
	if len(blackoutDates) > 0 {
		return false
	}
	
	// Check existing bookings
	conflictingBookings := b.getConflictingBookings(property.ID, checkIn, checkOut)
	return len(conflictingBookings) == 0
}

// calculateTotalAmount calculates total booking amount
func (b *BookingEngine) calculateTotalAmount(checkIn, checkOut time.Time, property *models.Property) float64 {
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	baseAmount := property.BasePrice * float64(nights)
	cleaningFee := property.CleaningFee
	serviceFee := baseAmount * 0.15 // 15% service fee
	
	return baseAmount + cleaningFee + serviceFee
}

// generateTRECDisclosures generates required TREC disclosures
func (b *BookingEngine) generateTRECDisclosures(request *BookingRequest) string {
	disclosures := []TRECDisclosure{
		{
			Type:     "property_condition",
			Title:    "Property Condition Disclosure",
			Content:  "The property is provided 'as-is' and guest acknowledges inspection of the property condition...",
			Required: true,
		},
		{
			Type:     "liability_waiver",
			Title:    "Liability Waiver",
			Content:  "Guest acknowledges and assumes all risks associated with the use of the property...",
			Required: true,
		},
		{
			Type:     "cancellation_policy",
			Title:    "Cancellation Policy Disclosure",
			Content:  "Guest acknowledges understanding of the cancellation policy and associated fees...",
			Required: true,
		},
	}
	
	return utils.ConvertToJSON(disclosures)
}

// getBlackoutDates gets blackout dates for property in date range
func (b *BookingEngine) getBlackoutDates(propertyID string, checkIn, checkOut time.Time) []time.Time {
	// This would query database for blackout dates
	// Placeholder implementation
	return []time.Time{}
}

// getConflictingBookings gets existing bookings that conflict with date range
func (b *BookingEngine) getConflictingBookings(propertyID string, checkIn, checkOut time.Time) []string {
	// This would query database for conflicting bookings
	// Placeholder implementation
	return []string{}
}

// calculatePricing calculates detailed pricing for date range
func (b *BookingEngine) calculatePricing(property *models.Property, checkIn, checkOut time.Time) *PricingBreakdown {
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	baseAmount := property.BasePrice * float64(nights)
	cleaningFee := property.CleaningFee
	serviceFee := baseAmount * 0.15
	taxes := (baseAmount + serviceFee) * 0.08 // 8% tax
	totalAmount := baseAmount + cleaningFee + serviceFee + taxes
	
	return &PricingBreakdown{
		BaseAmount:   baseAmount,
		CleaningFee:  cleaningFee,
		ServiceFee:   serviceFee,
		Taxes:        taxes,
		TotalAmount:  totalAmount,
		Currency:     "USD",
		Nights:       nights,
		AverageNightly: baseAmount / float64(nights),
	}
}

// getCancellationPolicy gets cancellation policy based on check-in date
func (b *BookingEngine) getCancellationPolicy(checkIn time.Time) *CancellationPolicy {
	daysUntilCheckIn := int(checkIn.Sub(time.Now()).Hours() / 24)
	
	if daysUntilCheckIn >= 14 {
		return &CancellationPolicy{
			AllowsCancellation: true,
			RefundPercentage:  100,
			Description:       "Full refund for cancellations 14+ days before check-in",
		}
	} else if daysUntilCheckIn >= 7 {
		return &CancellationPolicy{
			AllowsCancellation: true,
			RefundPercentage:  50,
			Description:       "50% refund for cancellations 7-13 days before check-in",
		}
	}
	
	return &CancellationPolicy{
		AllowsCancellation: false,
		RefundPercentage:  0,
		Description:       "No refund for cancellations less than 7 days before check-in",
	}
}

// calculateRefund calculates refund amount based on policy
func (b *BookingEngine) calculateRefund(booking *models.Booking, policy *CancellationPolicy) float64 {
	if !policy.AllowsCancellation {
		return 0
	}
	
	return booking.PaidAmount * (policy.RefundPercentage / 100)
}

// Supporting types for booking operations
type BookingRequest struct {
	PropertyID    string    `json:"property_id"`
	CheckIn       time.Time `json:"check_in"`
	CheckOut      time.Time `json:"check_out"`
	Guests        int       `json:"guests"`
	PaymentAmount float64   `json:"payment_amount"`
	TRECCompliant bool      `json:"trec_compliant"`
	SpecialRequests string  `json:"special_requests"`
}

type AvailabilityResponse struct {
	Available           bool              `json:"available"`
	Reason             string            `json:"reason,omitempty"`
	Pricing            *PricingBreakdown `json:"pricing,omitempty"`
	BlackoutDates      []time.Time       `json:"blackout_dates,omitempty"`
	ConflictingBookings []string         `json:"conflicting_bookings,omitempty"`
}

type PricingBreakdown struct {
	BaseAmount     float64 `json:"base_amount"`
	CleaningFee    float64 `json:"cleaning_fee"`
	ServiceFee     float64 `json:"service_fee"`
	Taxes          float64 `json:"taxes"`
	TotalAmount    float64 `json:"total_amount"`
	Currency       string  `json:"currency"`
	Nights         int     `json:"nights"`
	AverageNightly float64 `json:"average_nightly"`
}

type CancellationPolicy struct {
	AllowsCancellation bool    `json:"allows_cancellation"`
	RefundPercentage  float64 `json:"refund_percentage"`
	Description       string  `json:"description"`
}

type CancellationResult struct {
	BookingID    string              `json:"booking_id"`
	RefundAmount float64            `json:"refund_amount"`
	RefundMethod string             `json:"refund_method"`
	ProcessTime  string             `json:"process_time"`
	Policy       *CancellationPolicy `json:"policy"`
}

type CheckInData struct {
	ActualCheckInTime time.Time `json:"actual_check_in_time"`
	GuestCount       int       `json:"guest_count"`
	IDVerified       bool      `json:"id_verified"`
	Notes            string    `json:"notes"`
}

type CheckOutData struct {
	ActualCheckOutTime time.Time `json:"actual_check_out_time"`
	PropertyCondition  string    `json:"property_condition"`
	DamagesReported    bool      `json:"damages_reported"`
	AdditionalCharges  float64   `json:"additional_charges"`
	Notes              string    `json:"notes"`
}
