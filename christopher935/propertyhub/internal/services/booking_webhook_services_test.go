package services

import (
	"encoding/json"
	"testing"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupWebhookTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.WebhookEvent{}, &models.Contact{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestProcessWebhook_FUB(t *testing.T) {
	db := setupWebhookTestDB(t)
	ws := NewWebhookService(db)

	payload := map[string]interface{}{
		"event": "person.created",
		"eventId": "test-event-123",
		"data": map[string]interface{}{
			"id":        "fub-lead-456",
			"email":     "test@example.com",
			"firstName": "John",
			"lastName":  "Doe",
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	err := ws.ProcessWebhook(payloadBytes, "fub")

	if err != nil {
		t.Errorf("ProcessWebhook failed: %v", err)
	}

	var event models.WebhookEvent
	db.Where("source = ? AND event_type = ?", "fub", "person.created").First(&event)

	if event.ID == 0 {
		t.Error("Webhook event was not stored")
	}

	if event.Status != "processed" {
		t.Errorf("Expected status 'processed', got '%s'", event.Status)
	}

	var contact models.Contact
	db.Where("fub_lead_id = ?", "fub-lead-456").First(&contact)

	if contact.ID == 0 {
		t.Error("Contact was not created from webhook")
	}
}

func TestProcessWebhook_InvalidJSON(t *testing.T) {
	db := setupWebhookTestDB(t)
	ws := NewWebhookService(db)

	invalidPayload := []byte(`{invalid json}`)
	err := ws.ProcessWebhook(invalidPayload, "fub")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetWebhookEvents(t *testing.T) {
	db := setupWebhookTestDB(t)
	ws := NewWebhookService(db)

	db.Create(&models.WebhookEvent{
		Source:    "fub",
		EventType: "person.created",
		Status:    "processed",
		Payload:   models.JSONB{"test": "data"},
		CreatedAt: time.Now(),
	})

	db.Create(&models.WebhookEvent{
		Source:    "stripe",
		EventType: "payment.succeeded",
		Status:    "processed",
		Payload:   models.JSONB{"test": "data"},
		CreatedAt: time.Now(),
	})

	events, err := ws.GetWebhookEvents("", 10)
	if err != nil {
		t.Errorf("GetWebhookEvents failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	fubEvents, _ := ws.GetWebhookEvents("fub", 10)
	if len(fubEvents) != 1 {
		t.Errorf("Expected 1 FUB event, got %d", len(fubEvents))
	}
}

func TestReprocessWebhookEvent(t *testing.T) {
	db := setupWebhookTestDB(t)
	ws := NewWebhookService(db)

	event := &models.WebhookEvent{
		Source:     "fub",
		EventType:  "person.created",
		Status:     "failed",
		Error:      "test error",
		RetryCount: 0,
		Payload: models.JSONB{
			"event": "person.created",
			"data": map[string]interface{}{
				"id":        "fub-lead-789",
				"email":     "retry@example.com",
				"firstName": "Retry",
				"lastName":  "Test",
			},
		},
		CreatedAt: time.Now(),
	}
	db.Create(event)

	err := ws.ReprocessWebhookEvent("1")
	if err != nil {
		t.Errorf("ReprocessWebhookEvent failed: %v", err)
	}

	var updatedEvent models.WebhookEvent
	db.First(&updatedEvent, event.ID)

	if updatedEvent.RetryCount == 0 {
		t.Error("Expected retry count to be incremented")
	}
}

func TestReprocessWebhookEvent_MaxRetries(t *testing.T) {
	db := setupWebhookTestDB(t)
	ws := NewWebhookService(db)

	event := &models.WebhookEvent{
		Source:     "fub",
		EventType:  "person.created",
		Status:     "failed",
		Error:      "test error",
		RetryCount: 5,
		Payload:    models.JSONB{"event": "test"},
		CreatedAt:  time.Now(),
	}
	db.Create(event)

	err := ws.ReprocessWebhookEvent("1")
	if err == nil {
		t.Error("Expected error for max retries exceeded")
	}
}

func TestHandleFUBPersonCreated(t *testing.T) {
	db := setupWebhookTestDB(t)
	ws := NewWebhookService(db)

	data := map[string]interface{}{
		"id":        "fub-123",
		"email":     "john@example.com",
		"firstName": "John",
		"lastName":  "Doe",
	}

	err := ws.handleFUBPersonCreated(data)
	if err != nil {
		t.Errorf("handleFUBPersonCreated failed: %v", err)
	}

	var contact models.Contact
	db.Where("fub_lead_id = ?", "fub-123").First(&contact)

	if contact.ID == 0 {
		t.Error("Contact was not created")
	}

	if !contact.FUBSynced {
		t.Error("Expected FUBSynced to be true")
	}
}

func TestHandleFUBPersonUpdated(t *testing.T) {
	db := setupWebhookTestDB(t)
	ws := NewWebhookService(db)

	db.Create(&models.Contact{
		Email:     "existing@example.com",
		Name:      "Old Name",
		FUBLeadID: "fub-456",
		FUBSynced: true,
		Status:    "new",
	})

	data := map[string]interface{}{
		"id":        "fub-456",
		"email":     "updated@example.com",
		"firstName": "Updated",
		"lastName":  "Name",
	}

	err := ws.handleFUBPersonUpdated(data)
	if err != nil {
		t.Errorf("handleFUBPersonUpdated failed: %v", err)
	}

	var contact models.Contact
	db.Where("fub_lead_id = ?", "fub-456").First(&contact)

	if string(contact.Email) != "updated@example.com" {
		t.Errorf("Expected email to be updated to 'updated@example.com', got '%s'", string(contact.Email))
	}
}
