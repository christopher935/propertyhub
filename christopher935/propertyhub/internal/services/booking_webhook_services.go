package services

import (
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
)

// BookingService handles booking operations
type BookingService struct {
	db                *gorm.DB
	fubService        *EnhancedFUBIntegrationService
	encryptionManager *security.EncryptionManager
}

// CreateBookingRequest represents a booking creation request
type CreateBookingRequest struct {
	PropertyID      uint      `json:"property_id"`
	ShowingDate     time.Time `json:"showing_date"`
	AttendeeCount   int       `json:"attendee_count"`
	ShowingType     string    `json:"showing_type"`
	SpecialRequests string    `json:"special_requests"`
	DurationMinutes int       `json:"duration_minutes"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Name            string    `json:"name"`
}

// NewBookingService creates a new booking service
func NewBookingService(db *gorm.DB, fubService *EnhancedFUBIntegrationService, encryptionManager *security.EncryptionManager) *BookingService {
	return &BookingService{
		db:                db,
		fubService:        fubService,
		encryptionManager: encryptionManager,
	}
}

// GetBookings gets bookings with filters
func (bs *BookingService) GetBookings(filters map[string]interface{}) ([]models.Booking, error) {
	var bookings []models.Booking
	query := bs.db.Model(&models.Booking{})

	for key, value := range filters {
		switch key {
		case "property_id":
			query = query.Where("property_id = ?", value)
		case "status":
			query = query.Where("status = ?", value)
		}
	}

	err := query.Find(&bookings).Error
	return bookings, err
}

// GetBookingByID gets a booking by ID
func (bs *BookingService) GetBookingByID(id uint) (*models.Booking, error) {
	var booking models.Booking
	err := bs.db.First(&booking, id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// GetBookingWithLeadData gets booking with associated lead data
func (bs *BookingService) GetBookingWithLeadData(id uint) (interface{}, error) {
	booking, err := bs.GetBookingByID(id)
	if err != nil {
		return nil, err
	}

	// Return booking with any associated FUB lead data
	return map[string]interface{}{
		"booking":     booking,
		"fub_lead_id": booking.FUBLeadID,
		"fub_synced":  booking.FUBSynced,
	}, nil
}

// CreateBooking creates a new booking and optionally converts to FUB lead
func (bs *BookingService) CreateBooking(req *CreateBookingRequest) (*models.Booking, error) {
	// Encrypt PII fields before storing
	encryptedEmail, err := bs.encryptionManager.EncryptEmail(req.Email)
	if err != nil {
		return nil, err
	}

	encryptedPhone, err := bs.encryptionManager.EncryptPhone(req.Phone)
	if err != nil {
		return nil, err
	}

	encryptedName, err := bs.encryptionManager.Encrypt(req.Name)
	if err != nil {
		return nil, err
	}

	booking := &models.Booking{
		PropertyID:      req.PropertyID,
		ShowingDate:     req.ShowingDate,
		AttendeeCount:   req.AttendeeCount,
		ShowingType:     req.ShowingType,
		SpecialRequests: req.SpecialRequests,
		DurationMinutes: req.DurationMinutes,
		Email:           encryptedEmail,
		Phone:           encryptedPhone,
		Name:            encryptedName,
		Status:          "confirmed",
	}

	err = bs.db.Create(booking).Error
	if err != nil {
		return nil, err
	}

	// Auto-convert to FUB lead if FUB service is available
	if bs.fubService != nil {
		go func() {
			conversionRules := []string{"urgent_showing", "high_value_property"}
			_, err := bs.fubService.AutoConvertBookingToLead(booking, conversionRules)
			if err != nil {
				// Log error but don't fail the booking
				// TODO: Add proper logging
			}
		}()
	}

	return booking, nil
}

// UpdateBookingStatus updates booking status
func (bs *BookingService) UpdateBookingStatus(id uint, status string) error {
	return bs.db.Model(&models.Booking{}).Where("id = ?", id).Update("status", status).Error
}

// CancelBooking cancels a booking
func (bs *BookingService) CancelBooking(id uint, reason string) error {
	return bs.db.Model(&models.Booking{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           "cancelled",
		"special_requests": reason,
	}).Error
}

// RescheduleBooking reschedules a booking
func (bs *BookingService) RescheduleBooking(id uint, newDate time.Time) error {
	return bs.db.Model(&models.Booking{}).Where("id = ?", id).Update("showing_date", newDate).Error
}

// WebhookService handles webhook operations
type WebhookService struct {
	db *gorm.DB
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db}
}

// ProcessWebhook processes incoming webhooks
func (ws *WebhookService) ProcessWebhook(eventType string, payload map[string]interface{}) error {
	// TODO: Implement webhook processing logic
	return nil
}

// GetWebhookEvents gets webhook events with filters
func (ws *WebhookService) GetWebhookEvents(filters map[string]interface{}) ([]interface{}, error) {
	// TODO: Implement webhook event retrieval
	return []interface{}{}, nil
}

// ReprocessWebhookEvent reprocesses a webhook event
func (ws *WebhookService) ReprocessWebhookEvent(id uint) error {
	// TODO: Implement webhook event reprocessing
	return nil
}
