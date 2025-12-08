package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockFUBAPIServer creates a mock FUB API server for testing
type MockFUBAPIServer struct {
	server          *httptest.Server
	contactsCreated []map[string]interface{}
	dealsCreated    []map[string]interface{}
	tasksCreated    []map[string]interface{}
	notesCreated    []map[string]interface{}
	eventsCreated   []map[string]interface{}
}

func NewMockFUBAPIServer() *MockFUBAPIServer {
	mock := &MockFUBAPIServer{
		contactsCreated: []map[string]interface{}{},
		dealsCreated:    []map[string]interface{}{},
		tasksCreated:    []map[string]interface{}{},
		notesCreated:    []map[string]interface{}{},
		eventsCreated:   []map[string]interface{}{},
	}

	mux := http.NewServeMux()

	// POST /people - Create contact
	mux.HandleFunc("/people", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var contact map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		contact["id"] = "fub_contact_123"
		now := time.Now()
		contact["created"] = now
		contact["updated"] = now

		mock.contactsCreated = append(mock.contactsCreated, contact)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(contact)
	})

	// PUT /people/:id - Update contact
	mux.HandleFunc("/people/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		updates["id"] = "fub_contact_123"
		now := time.Now()
		updates["updated"] = now

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updates)
	})

	// POST /deals - Create deal
	mux.HandleFunc("/deals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var deal map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&deal); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		deal["id"] = "fub_deal_456"
		now := time.Now()
		deal["created"] = now
		deal["updated"] = now

		mock.dealsCreated = append(mock.dealsCreated, deal)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(deal)
	})

	// POST /tasks - Create task
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var task map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		task["id"] = "fub_task_789"
		now := time.Now()
		task["created"] = now

		mock.tasksCreated = append(mock.tasksCreated, task)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	})

	// POST /notes - Create note
	mux.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var note map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		note["id"] = "fub_note_101"
		now := time.Now()
		note["created"] = now

		mock.notesCreated = append(mock.notesCreated, note)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(note)
	})

	// POST /events - Create event
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var event map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		event["id"] = "fub_event_202"
		now := time.Now()
		event["created"] = now

		mock.eventsCreated = append(mock.eventsCreated, event)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(event)
	})

	// POST /people/:id/actionplans - Assign action plan
	mux.HandleFunc("/people/fub_contact_123/actionplans", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	// POST /lists/:id/people - Add to pond
	mux.HandleFunc("/lists/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	// GET /users - Get agents
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		agents := []map[string]interface{}{
			{"id": "agent_1", "name": "John Doe", "email": "john@example.com"},
			{"id": "agent_2", "name": "Jane Smith", "email": "jane@example.com"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(agents)
	})

	// GET /actionplans - Get action plans
	mux.HandleFunc("/actionplans", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		plans := []map[string]interface{}{
			{"id": "plan_1", "name": "Luxury Buyer Plan", "isActive": true},
			{"id": "plan_2", "name": "Rental Prospect Plan", "isActive": true},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(plans)
	})

	// GET /lists - Get ponds
	mux.HandleFunc("/lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ponds := []map[string]interface{}{
			{"id": "pond_1", "name": "Hot Leads", "type": "smart"},
			{"id": "pond_2", "name": "Luxury Buyers", "type": "static"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ponds)
	})

	mock.server = httptest.NewServer(mux)
	return mock
}

func (m *MockFUBAPIServer) Close() {
	m.server.Close()
}

func (m *MockFUBAPIServer) URL() string {
	return m.server.URL
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate test models
	err = db.AutoMigrate(
		&models.Lead{},
		&models.Property{},
		&models.BehavioralEvent{},
		&models.BehavioralScore{},
		&models.FUBLead{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// TestFUBAPIClient_CreateContact tests contact creation
func TestFUBAPIClient_CreateContact(t *testing.T) {
	mockServer := NewMockFUBAPIServer()
	defer mockServer.Close()

	db, err := setupTestDB()
	assert.NoError(t, err)

	client := NewBehavioralFUBAPIClient(db, "test_api_key")
	client.baseURL = mockServer.URL()

	contact := &FUBContact{
		Name:      "John Doe",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "+1234567890",
		Source:    "Website",
		Tags:      []string{"buyer", "qualified"},
		CustomFields: map[string]interface{}{
			"property_type": "luxury",
			"budget":        500000,
		},
	}

	result, err := client.CreateContact(contact)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "fub_contact_123", result.ID)
	assert.Equal(t, "John Doe", result.Name)

	assert.Equal(t, 1, len(mockServer.contactsCreated))
}

// TestFUBAPIClient_CreateDeal tests deal creation
func TestFUBAPIClient_CreateDeal(t *testing.T) {
	mockServer := NewMockFUBAPIServer()
	defer mockServer.Close()

	db, err := setupTestDB()
	assert.NoError(t, err)

	client := NewBehavioralFUBAPIClient(db, "test_api_key")
	client.baseURL = mockServer.URL()

	deal := &FUBDeal{
		Name:            "Luxury Property Deal",
		ContactID:       "fub_contact_123",
		PropertyAddress: "123 Main St, Houston, TX",
		Value:           500000,
		Stage:           "qualified",
		Probability:     0.75,
	}

	result, err := client.CreateDeal(deal)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "fub_deal_456", result.ID)
	assert.Equal(t, "Luxury Property Deal", result.Name)

	assert.Equal(t, 1, len(mockServer.dealsCreated))
}

// TestBehavioralFUBIntegration_ProcessTrigger tests behavioral trigger processing
func TestBehavioralFUBIntegration_ProcessTrigger(t *testing.T) {
	mockServer := NewMockFUBAPIServer()
	defer mockServer.Close()

	db, err := setupTestDB()
	assert.NoError(t, err)

	service := NewBehavioralFUBIntegrationService(db, "test_api_key")
	service.apiClient.baseURL = mockServer.URL()

	triggerData := &PropertyCategoryTriggerData{
		PropertyCategory: "luxury_rental",
		PropertyTier:     "luxury",
		TargetDemo:       "high_net_worth",
		PriceRange:       "5000_plus",
		Location:         "river_oaks",
		BehavioralData: map[string]interface{}{
			"urgency_score":       85.0,
			"financial_readiness": 90.0,
			"engagement_depth":    75.0,
			"rent_budget":         8000.0,
		},
		MarketContext: map[string]interface{}{
			"seasonal_factor": "high_demand",
			"competition":     "moderate",
		},
	}

	contactInfo := map[string]string{
		"name":  "Sarah Johnson",
		"email": "sarah@example.com",
		"phone": "+15551234567",
	}

	result, err := service.ProcessBehavioralTrigger(triggerData, contactInfo)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.ContactID)
	assert.Equal(t, "URGENT", result.Priority)

	assert.Equal(t, 1, len(mockServer.contactsCreated))
	assert.Equal(t, 1, len(mockServer.dealsCreated))
}

// TestFUBBidirectionalSync_LogCallToFUB tests call logging
func TestFUBBidirectionalSync_LogCallToFUB(t *testing.T) {
	mockServer := NewMockFUBAPIServer()
	defer mockServer.Close()

	db, err := setupTestDB()
	assert.NoError(t, err)

	lead := &models.Lead{
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		FUBLeadID: "fub_contact_123",
	}
	db.Create(lead)

	sync := NewFUBBidirectionalSync(db, "test_api_key")
	sync.fubBaseURL = mockServer.URL()

	err = sync.LogCallToFUB(int64(lead.ID), 300, "Discussed property options", "agent_1")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(mockServer.eventsCreated))
	event := mockServer.eventsCreated[0]
	assert.Equal(t, "fub_contact_123", event["personId"])
	assert.Equal(t, "call", event["type"])
	assert.Equal(t, 300.0, event["duration"])
}

// TestFUBBidirectionalSync_HandleWebhook tests webhook processing
func TestFUBBidirectionalSync_HandleWebhook(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	lead := &models.Lead{
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		FUBLeadID: "fub_contact_123",
	}
	db.Create(lead)

	sync := NewFUBBidirectionalSync(db, "test_api_key")

	webhookData := map[string]interface{}{
		"type":     "email.opened",
		"personId": "fub_contact_123",
		"eventId":  "event_123",
	}

	err = sync.HandleFUBWebhook(webhookData)
	assert.NoError(t, err)

	var event models.BehavioralEvent
	err = db.Where("lead_id = ? AND event_type = ?", lead.ID, "email_opened").First(&event).Error
	assert.NoError(t, err)
	assert.Equal(t, "email_opened", event.EventType)
}

// TestFUBErrorHandler_RetryLogic tests retry logic with rate limiting
func TestFUBErrorHandler_RetryLogic(t *testing.T) {
	handler := NewFUBErrorHandler()

	attemptCount := 0
	maxAttempts := 3

	operation := "test_operation"
	fn := func() (*http.Response, error) {
		attemptCount++

		if attemptCount < maxAttempts {
			rec := httptest.NewRecorder()
			rec.WriteHeader(http.StatusTooManyRequests)
			rec.Header().Set("Retry-After", "1")
			return rec.Result(), nil
		}

		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		return rec.Result(), nil
	}

	resp, err := handler.ExecuteWithRetry(operation, fn)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, maxAttempts, attemptCount)
}

// TestFieldMapping tests Lead to FUB Contact field mapping
func TestFieldMapping_LeadToFUBContact(t *testing.T) {
	mockServer := NewMockFUBAPIServer()
	defer mockServer.Close()

	db, err := setupTestDB()
	assert.NoError(t, err)

	service := NewBehavioralFUBIntegrationService(db, "test_api_key")
	service.apiClient.baseURL = mockServer.URL()

	triggerData := &PropertyCategoryTriggerData{
		PropertyCategory: "rental",
		PropertyTier:     "standard",
		TargetDemo:       "young_professional",
		PriceRange:       "1500_2500",
		Location:         "downtown",
		BehavioralData: map[string]interface{}{
			"urgency_score":       70.0,
			"financial_readiness": 80.0,
			"engagement_depth":    65.0,
		},
		MarketContext: map[string]interface{}{},
	}

	contactInfo := map[string]string{
		"name":  "Alex Martinez",
		"email": "alex@example.com",
		"phone": "+15559876543",
		"city":  "Houston",
		"state": "Texas",
	}

	result, err := service.ProcessBehavioralTrigger(triggerData, contactInfo)
	assert.NoError(t, err)
	assert.True(t, result.Success)

	contact := mockServer.contactsCreated[0]
	assert.Equal(t, "Alex Martinez", contact["name"])
	assert.Equal(t, "Alex", contact["firstName"])
	assert.Equal(t, "Martinez", contact["lastName"])
	assert.Equal(t, "alex@example.com", contact["email"])
	assert.Equal(t, "+15559876543", contact["phone"])

	customFields := contact["customFields"].(map[string]interface{})
	assert.Equal(t, "rental", customFields["property_category"])
	assert.Equal(t, "downtown", customFields["houston_location"])
	assert.Equal(t, 70.0, customFields["behavioral_urgency"])
}

// TestActionPlanTriggering tests action plan assignment based on behavioral scores
func TestActionPlanTriggering(t *testing.T) {
	mockServer := NewMockFUBAPIServer()
	defer mockServer.Close()

	db, err := setupTestDB()
	assert.NoError(t, err)

	service := NewBehavioralFUBIntegrationService(db, "test_api_key")
	service.apiClient.baseURL = mockServer.URL()

	testCases := []struct {
		name                string
		urgencyScore        float64
		financialScore      float64
		propertyCategory    string
		expectedPriority    string
		expectedWorkflowKey string
	}{
		{
			name:                "Luxury High Urgency",
			urgencyScore:        90,
			financialScore:      85,
			propertyCategory:    "luxury_rental",
			expectedPriority:    "URGENT",
			expectedWorkflowKey: "luxury",
		},
		{
			name:                "Student Housing Medium",
			urgencyScore:        75,
			financialScore:      60,
			propertyCategory:    "student_housing",
			expectedPriority:    "HIGH",
			expectedWorkflowKey: "student",
		},
		{
			name:                "Investment Low Urgency",
			urgencyScore:        40,
			financialScore:      70,
			propertyCategory:    "investment_property",
			expectedPriority:    "MEDIUM",
			expectedWorkflowKey: "investment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			triggerData := &PropertyCategoryTriggerData{
				PropertyCategory: tc.propertyCategory,
				PropertyTier:     "standard",
				TargetDemo:       "general",
				PriceRange:       "market_rate",
				Location:         "houston",
				BehavioralData: map[string]interface{}{
					"urgency_score":       tc.urgencyScore,
					"financial_readiness": tc.financialScore,
					"engagement_depth":    60.0,
				},
				MarketContext: map[string]interface{}{},
			}

			contactInfo := map[string]string{
				"name":  "Test User",
				"email": "test@example.com",
				"phone": "+15551234567",
			}

			result, err := service.ProcessBehavioralTrigger(triggerData, contactInfo)
			assert.NoError(t, err)
			assert.True(t, result.Success)
			assert.Equal(t, tc.expectedPriority, result.Priority)
			assert.Contains(t, result.WorkflowType, tc.expectedWorkflowKey)
		})
	}
}

// TestRentalCommissionCalculation tests rental commission calculation
func TestRentalCommissionCalculation(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	service := NewBehavioralFUBIntegrationService(db, "test_api_key")

	testCases := []struct {
		name               string
		monthlyRent        float64
		expectedCommission float64
	}{
		{"Luxury Rental $5000", 5000, 4500},
		{"Standard Rental $2000", 2000, 1800},
		{"Budget Rental $1200", 1200, 1080},
		{"High-End Rental $8000", 8000, 7200},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			triggerData := &PropertyCategoryTriggerData{
				PropertyCategory: "rental",
				BehavioralData: map[string]interface{}{
					"rent_budget": tc.monthlyRent,
				},
				MarketContext: map[string]interface{}{},
			}

			commission := service.calculateRentalCommissionValue(triggerData)
			assert.Equal(t, tc.expectedCommission, commission)
		})
	}
}

// TestSalesCommissionCalculation tests sales commission calculation
func TestSalesCommissionCalculation(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	service := NewBehavioralFUBIntegrationService(db, "test_api_key")

	testCases := []struct {
		name               string
		propertyValue      float64
		expectedCommission float64
	}{
		{"Luxury $3M", 3000000, 54000},
		{"Mid-Range $750K", 750000, 13500},
		{"Starter $350K", 350000, 5250},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commission := service.calculateSalesCommissionValue(tc.propertyValue)
			assert.InDelta(t, tc.expectedCommission, commission, 100)
		})
	}
}

// TestSyncStatusTracking tests sync status tracking
func TestSyncStatusTracking(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	db.AutoMigrate(&models.FUBLead{})

	lead := &models.FUBLead{
		FUBLeadID:    "fub_123",
		FirstName:    "Test",
		LastName:     "User",
		Email:        "test@example.com",
		Status:       "active",
		LastSyncedAt: time.Now().Add(-5 * time.Minute),
		SyncErrors:   []string{},
	}
	db.Create(lead)

	lead.LastSyncedAt = time.Now()
	db.Save(lead)

	var updatedLead models.FUBLead
	err = db.First(&updatedLead, lead.ID).Error
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), updatedLead.LastSyncedAt, 5*time.Second)
}
