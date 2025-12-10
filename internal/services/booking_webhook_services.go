package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
			defer func() {
				if r := recover(); r != nil {
					log.Printf("❌ Panic in FUB lead conversion: %v", r)
				}
			}()

			conversionRules := []string{"urgent_showing", "high_value_property"}
			_, err := bs.fubService.AutoConvertBookingToLead(booking, conversionRules)
			if err != nil {
				log.Printf("❌ Failed to convert booking to FUB lead: %v", err)
				bs.queueFailedConversion(booking.ID, err)
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

// queueFailedConversion queues a failed FUB conversion for retry
func (bs *BookingService) queueFailedConversion(bookingID uint, err error) {
	failedConversion := &models.FailedConversion{
		BookingID:  bookingID,
		Error:      err.Error(),
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}

	if dbErr := bs.db.Create(failedConversion).Error; dbErr != nil {
		log.Printf("❌ Failed to queue conversion for retry: %v", dbErr)
	}
}

// WebhookService handles webhook operations
type WebhookService struct {
	db         *gorm.DB
	fubService *EnhancedFUBIntegrationService
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{
		db:         db,
		fubService: nil,
	}
}

// SetFUBService sets the FUB service for webhook processing
func (ws *WebhookService) SetFUBService(fubService *EnhancedFUBIntegrationService) {
	ws.fubService = fubService
}

// ProcessWebhook processes incoming webhooks from various sources
func (ws *WebhookService) ProcessWebhook(payload []byte, source string) error {
	log.Printf("📥 Processing webhook from source: %s", source)

	var eventType string
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payload, &payloadMap); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %v", err)
	}

	if et, ok := payloadMap["event"].(string); ok {
		eventType = et
	} else if et, ok := payloadMap["eventType"].(string); ok {
		eventType = et
	} else if et, ok := payloadMap["type"].(string); ok {
		eventType = et
	}

	event := &models.WebhookEvent{
		Source:    source,
		EventType: eventType,
		Payload:   payloadMap,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if eventID, ok := payloadMap["eventId"].(string); ok {
		event.EventID = eventID
	} else if eventID, ok := payloadMap["id"].(string); ok {
		event.EventID = eventID
	}

	if err := ws.db.Create(event).Error; err != nil {
		return fmt.Errorf("failed to store webhook event: %v", err)
	}

	var processErr error
	switch source {
	case "fub":
		processErr = ws.processFUBWebhook(payloadMap)
	case "buildium":
		processErr = ws.processBuildiumWebhook(payloadMap)
	case "stripe":
		processErr = ws.processStripeWebhook(payloadMap)
	default:
		log.Printf("⚠️ Unknown webhook source: %s", source)
	}

	if processErr != nil {
		event.Status = "failed"
		event.Error = processErr.Error()
		event.RetryCount++
		log.Printf("❌ Webhook processing failed: %v", processErr)
	} else {
		event.Status = "processed"
		now := time.Now()
		event.ProcessedAt = &now
		log.Printf("✅ Webhook processed successfully")
	}

	ws.db.Save(event)
	return processErr
}

// GetWebhookEvents retrieves webhook events with filters
func (ws *WebhookService) GetWebhookEvents(source string, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent

	query := ws.db.Model(&models.WebhookEvent{})
	if source != "" {
		query = query.Where("source = ?", source)
	}

	if limit <= 0 {
		limit = 50
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&events).Error
	return events, err
}

// ReprocessWebhookEvent reprocesses a failed webhook event
func (ws *WebhookService) ReprocessWebhookEvent(eventID string) error {
	var event models.WebhookEvent
	if err := ws.db.First(&event, eventID).Error; err != nil {
		return fmt.Errorf("event not found: %v", err)
	}

	if event.RetryCount >= 5 {
		return fmt.Errorf("max retry count reached")
	}

	log.Printf("🔄 Reprocessing webhook event %s (retry %d)", eventID, event.RetryCount+1)

	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	return ws.ProcessWebhook(payload, event.Source)
}

// processFUBWebhook processes Follow Up Boss webhook events
func (ws *WebhookService) processFUBWebhook(payload map[string]interface{}) error {
	eventType, ok := payload["event"].(string)
	if !ok {
		return fmt.Errorf("missing event type in FUB webhook")
	}

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing data in FUB webhook")
	}

	log.Printf("📥 FUB Webhook: %s", eventType)

	switch eventType {
	case "person.created":
		return ws.handleFUBPersonCreated(data)
	case "person.updated":
		return ws.handleFUBPersonUpdated(data)
	case "deal.created":
		return ws.handleFUBDealCreated(data)
	case "deal.updated":
		return ws.handleFUBDealUpdated(data)
	case "task.completed":
		return ws.handleFUBTaskCompleted(data)
	case "event.created":
		return ws.handleFUBEventCreated(data)
	default:
		log.Printf("⚠️ Unhandled FUB event: %s", eventType)
	}

	return nil
}

// handleFUBPersonCreated syncs new FUB contact to local database
func (ws *WebhookService) handleFUBPersonCreated(data map[string]interface{}) error {
	email, _ := data["email"].(string)
	firstName, _ := data["firstName"].(string)
	lastName, _ := data["lastName"].(string)
	fubID, _ := data["id"].(string)

	if email == "" || fubID == "" {
		return fmt.Errorf("missing required fields in person.created")
	}

	log.Printf("👤 Creating contact: %s %s (%s)", firstName, lastName, email)

	contact := &models.Contact{
		Email:     security.EncryptedString(email),
		Name:      security.EncryptedString(fmt.Sprintf("%s %s", firstName, lastName)),
		FUBLeadID: fubID,
		Source:    "fub_webhook",
		Status:    "new",
		FUBSynced: true,
	}

	return ws.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "fub_lead_id", "fub_synced", "updated_at"}),
	}).Create(contact).Error
}

// handleFUBPersonUpdated syncs updated FUB contact
func (ws *WebhookService) handleFUBPersonUpdated(data map[string]interface{}) error {
	fubID, _ := data["id"].(string)
	email, _ := data["email"].(string)
	firstName, _ := data["firstName"].(string)
	lastName, _ := data["lastName"].(string)

	if fubID == "" {
		return fmt.Errorf("missing FUB ID in person.updated")
	}

	log.Printf("👤 Updating contact: %s", fubID)

	var contact models.Contact
	if err := ws.db.Where("fub_lead_id = ?", fubID).First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ws.handleFUBPersonCreated(data)
		}
		return err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if email != "" {
		updates["email"] = security.EncryptedString(email)
	}
	if firstName != "" || lastName != "" {
		updates["name"] = security.EncryptedString(fmt.Sprintf("%s %s", firstName, lastName))
	}

	return ws.db.Model(&contact).Updates(updates).Error
}

// handleFUBDealCreated processes new deal creation
func (ws *WebhookService) handleFUBDealCreated(data map[string]interface{}) error {
	dealID, _ := data["id"].(string)
	stage, _ := data["stage"].(string)
	value, _ := data["value"].(float64)

	log.Printf("💼 Deal created: %s (stage: %s, value: $%.2f)", dealID, stage, value)

	return nil
}

// handleFUBDealUpdated processes deal updates
func (ws *WebhookService) handleFUBDealUpdated(data map[string]interface{}) error {
	dealID, _ := data["id"].(string)
	stage, _ := data["stage"].(string)

	log.Printf("💼 Deal updated: %s (stage: %s)", dealID, stage)

	return nil
}

// handleFUBTaskCompleted processes completed task
func (ws *WebhookService) handleFUBTaskCompleted(data map[string]interface{}) error {
	taskID, _ := data["id"].(string)
	title, _ := data["title"].(string)

	log.Printf("✅ Task completed: %s (%s)", taskID, title)

	return nil
}

// handleFUBEventCreated processes event creation (showings, appointments)
func (ws *WebhookService) handleFUBEventCreated(data map[string]interface{}) error {
	eventID, _ := data["id"].(string)
	eventType, _ := data["type"].(string)
	startTime, _ := data["startTime"].(string)

	log.Printf("📅 Event created: %s (type: %s, start: %s)", eventID, eventType, startTime)

	return nil
}

// processBuildiumWebhook processes Buildium property management webhooks
func (ws *WebhookService) processBuildiumWebhook(payload map[string]interface{}) error {
	eventType, _ := payload["eventType"].(string)
	log.Printf("🏢 Buildium Webhook: %s", eventType)

	switch eventType {
	case "lease.signed":
		return ws.handleBuildiumLeaseSigned(payload)
	case "tenant.created":
		return ws.handleBuildiumTenantCreated(payload)
	default:
		log.Printf("⚠️ Unhandled Buildium event: %s", eventType)
	}

	return nil
}

// handleBuildiumLeaseSigned processes lease signing events
func (ws *WebhookService) handleBuildiumLeaseSigned(payload map[string]interface{}) error {
	leaseID, _ := payload["leaseId"].(string)
	propertyID, _ := payload["propertyId"].(string)

	log.Printf("📝 Lease signed: %s for property %s", leaseID, propertyID)

	return nil
}

// handleBuildiumTenantCreated processes new tenant creation
func (ws *WebhookService) handleBuildiumTenantCreated(payload map[string]interface{}) error {
	tenantID, _ := payload["tenantId"].(string)
	tenantName, _ := payload["name"].(string)

	log.Printf("👤 Tenant created: %s (%s)", tenantID, tenantName)

	return nil
}

// processStripeWebhook processes Stripe payment webhooks
func (ws *WebhookService) processStripeWebhook(payload map[string]interface{}) error {
	eventType, _ := payload["type"].(string)
	log.Printf("💳 Stripe Webhook: %s", eventType)

	switch eventType {
	case "payment_intent.succeeded":
		return ws.handleStripePaymentSucceeded(payload)
	case "payment_intent.failed":
		return ws.handleStripePaymentFailed(payload)
	default:
		log.Printf("⚠️ Unhandled Stripe event: %s", eventType)
	}

	return nil
}

// handleStripePaymentSucceeded processes successful payments
func (ws *WebhookService) handleStripePaymentSucceeded(payload map[string]interface{}) error {
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing data in Stripe webhook")
	}

	object, _ := data["object"].(map[string]interface{})
	amount, _ := object["amount"].(float64)
	currency, _ := object["currency"].(string)

	log.Printf("✅ Payment succeeded: %.2f %s", amount/100, currency)

	return nil
}

// handleStripePaymentFailed processes failed payments
func (ws *WebhookService) handleStripePaymentFailed(payload map[string]interface{}) error {
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing data in Stripe webhook")
	}

	object, _ := data["object"].(map[string]interface{})
	amount, _ := object["amount"].(float64)
	currency, _ := object["currency"].(string)

	log.Printf("❌ Payment failed: %.2f %s", amount/100, currency)

	return nil
}
