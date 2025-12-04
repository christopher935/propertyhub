package models

import (
	"time"

	"chrisgross-ctrl-project/internal/security"
)

// BookingDataResponse represents a booking with decrypted sensitive fields
// This is returned to the API/frontend instead of the raw Booking model
type BookingDataResponse struct {
	ID                 uint       `json:"id"`
	ReferenceNumber    string     `json:"reference_number"`
	PropertyID         uint       `json:"property_id"`
	PropertyAddress    string     `json:"property_address"`
	FUBLeadID          string     `json:"fub_lead_id"`
	Email              string     `json:"email"`               // Decrypted
	Name               string     `json:"name"`                // Decrypted
	Phone              string     `json:"phone"`               // Decrypted
	FUBSynced          bool       `json:"fub_synced"`
	InterestLevel      string     `json:"interest_level"`
	ShowingDate        time.Time  `json:"showing_date"`
	DurationMinutes    int        `json:"duration_minutes"`
	Status             string     `json:"status"`
	Notes              string     `json:"notes"`
	ShowingType        string     `json:"showing_type"`
	AttendeeCount      int        `json:"attendee_count"`
	SpecialRequests    string     `json:"special_requests"`
	FUBActionPlanID    string     `json:"fub_action_plan_id"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CancellationReason string     `json:"cancellation_reason,omitempty"`
	RescheduledFrom    *uint      `json:"rescheduled_from,omitempty"`
	ConsentGiven       bool       `json:"consent_given"`
	ConsentSource      string     `json:"consent_source"`
	ConsentTimestamp   *time.Time `json:"consent_timestamp,omitempty"`
	MarketingConsent   bool       `json:"marketing_consent"`
	TermsAccepted      bool       `json:"terms_accepted"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ToBookingDataResponse converts a Booking to BookingDataResponse with decrypted fields
func ToBookingDataResponse(booking Booking, encryptionManager *security.EncryptionManager) BookingDataResponse {
	// Decrypt sensitive fields
	name := string(booking.Name)
	email := string(booking.Email)
	phone := string(booking.Phone)

	// If encryption manager is available, decrypt the fields
	if encryptionManager != nil {
		if decrypted, err := encryptionManager.Decrypt(booking.Name); err == nil {
			name = decrypted
		}
		if decrypted, err := encryptionManager.Decrypt(booking.Email); err == nil {
			email = decrypted
		}
		if decrypted, err := encryptionManager.Decrypt(booking.Phone); err == nil {
			phone = decrypted
		}
	}

	return BookingDataResponse{
		ID:                 booking.ID,
		ReferenceNumber:    booking.ReferenceNumber,
		PropertyID:         booking.PropertyID,
		PropertyAddress:    booking.PropertyAddress,
		FUBLeadID:          booking.FUBLeadID,
		Email:              email,
		Name:               name,
		Phone:              phone,
		FUBSynced:          booking.FUBSynced,
		InterestLevel:      booking.InterestLevel,
		ShowingDate:        booking.ShowingDate,
		DurationMinutes:    booking.DurationMinutes,
		Status:             booking.Status,
		Notes:              booking.Notes,
		ShowingType:        booking.ShowingType,
		AttendeeCount:      booking.AttendeeCount,
		SpecialRequests:    booking.SpecialRequests,
		FUBActionPlanID:    booking.FUBActionPlanID,
		CompletedAt:        booking.CompletedAt,
		CancellationReason: booking.CancellationReason,
		RescheduledFrom:    booking.RescheduledFrom,
		ConsentGiven:       booking.ConsentGiven,
		ConsentSource:      booking.ConsentSource,
		ConsentTimestamp:   booking.ConsentTimestamp,
		MarketingConsent:   booking.MarketingConsent,
		TermsAccepted:      booking.TermsAccepted,
		CreatedAt:          booking.CreatedAt,
		UpdatedAt:          booking.UpdatedAt,
	}
}

// ToBookingDataResponseList converts a slice of Bookings to BookingDataResponses
func ToBookingDataResponseList(bookings []Booking, encryptionManager *security.EncryptionManager) []BookingDataResponse {
	responses := make([]BookingDataResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = ToBookingDataResponse(booking, encryptionManager)
	}
	return responses
}
