package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// BehavioralFUBAPIClient provides sophisticated FUB API integration with behavioral intelligence
type BehavioralFUBAPIClient struct {
	db           *gorm.DB
	client       *http.Client
	apiKey       string
	baseURL      string
	errorHandler *FUBErrorHandler
}

// NewBehavioralFUBAPIClient creates a new behavioral intelligence-driven FUB API client
func NewBehavioralFUBAPIClient(db *gorm.DB, apiKey string) *BehavioralFUBAPIClient {
	return &BehavioralFUBAPIClient{
		db:           db,
		client:       &http.Client{Timeout: 30 * time.Second},
		apiKey:       apiKey,
		baseURL:      "https://api.followupboss.com/v1",
		errorHandler: NewFUBErrorHandler(),
	}
}

// FUBContact represents a FUB contact/lead structure
type FUBContact struct {
	ID           string                 `json:"id,omitempty"`
	Name         string                 `json:"name"`
	FirstName    string                 `json:"firstName,omitempty"`
	LastName     string                 `json:"lastName,omitempty"`
	Email        string                 `json:"email,omitempty"`
	Phone        string                 `json:"phone,omitempty"`
	Source       string                 `json:"source"`
	Status       string                 `json:"status,omitempty"`
	Stage        string                 `json:"stage,omitempty"`
	AssignedTo   string                 `json:"assignedTo,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
	Created      *time.Time             `json:"created,omitempty"`
	Updated      *time.Time             `json:"updated,omitempty"`
}

// FUBDeal represents a FUB deal structure
type FUBDeal struct {
	ID              string                 `json:"id,omitempty"`
	Name            string                 `json:"name"`
	ContactID       string                 `json:"contactId"`
	PropertyAddress string                 `json:"propertyAddress,omitempty"`
	Value           float64                `json:"value,omitempty"`
	Stage           string                 `json:"stage,omitempty"`
	Probability     float64                `json:"probability,omitempty"`
	ExpectedClose   *time.Time             `json:"expectedClose,omitempty"`
	AssignedTo      string                 `json:"assignedTo,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	CustomFields    map[string]interface{} `json:"customFields,omitempty"`
	Created         *time.Time             `json:"created,omitempty"`
	Updated         *time.Time             `json:"updated,omitempty"`
}

// FUBActionPlan represents a FUB action plan structure
type FUBActionPlan struct {
	ID          string              `json:"id,omitempty"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	IsActive    bool                `json:"isActive"`
	Steps       []FUBActionPlanStep `json:"steps,omitempty"`
	Created     *time.Time          `json:"created,omitempty"`
	Updated     *time.Time          `json:"updated,omitempty"`
}

// FUBActionPlanStep represents a step in a FUB action plan
type FUBActionPlanStep struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Type      string `json:"type"`  // email, call, task, etc.
	Delay     int    `json:"delay"` // delay in hours
	Content   string `json:"content,omitempty"`
	IsActive  bool   `json:"isActive"`
	SortOrder int    `json:"sortOrder"`
}

// FUBAutomation represents a FUB automation structure
type FUBAutomation struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	TriggerType string                 `json:"triggerType"`
	TriggerData map[string]interface{} `json:"triggerData,omitempty"`
	ActionType  string                 `json:"actionType"`
	ActionData  map[string]interface{} `json:"actionData,omitempty"`
	IsActive    bool                   `json:"isActive"`
	Priority    int                    `json:"priority,omitempty"`
	Created     *time.Time             `json:"created,omitempty"`
	Updated     *time.Time             `json:"updated,omitempty"`
}

// FUBPond represents a FUB pond/list structure
type FUBPond struct {
	ID          string     `json:"id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Type        string     `json:"type,omitempty"` // smart, static
	Criteria    string     `json:"criteria,omitempty"`
	IsActive    bool       `json:"isActive"`
	Created     *time.Time `json:"created,omitempty"`
	Updated     *time.Time `json:"updated,omitempty"`
}

// BehavioralTriggerResult represents the result of behavioral intelligence processing
type BehavioralTriggerResult struct {
	Success           bool                   `json:"success"`
	ContactID         string                 `json:"contact_id,omitempty"`
	DealID            string                 `json:"deal_id,omitempty"`
	ActionPlanID      string                 `json:"action_plan_id,omitempty"`
	AutomationID      string                 `json:"automation_id,omitempty"`
	PondID            string                 `json:"pond_id,omitempty"`
	WorkflowType      string                 `json:"workflow_type"`
	RecommendedAction string                 `json:"recommended_action"`
	Priority          string                 `json:"priority"`
	ScheduledAt       time.Time              `json:"scheduled_at"`
	PropertyCategory  map[string]interface{} `json:"property_category"`
	BehavioralData    map[string]interface{} `json:"behavioral_data"`
	MarketContext     map[string]interface{} `json:"market_context"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	ProcessedAt       time.Time              `json:"processed_at"`
}

// makeRequest makes an authenticated HTTP request to FUB API with comprehensive error handling
func (client *BehavioralFUBAPIClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	operation := fmt.Sprintf("%s %s", method, endpoint)

	// Use error handler with retry logic
	return client.errorHandler.ExecuteWithRetry(operation, func() (*http.Response, error) {
		var reqBody *bytes.Buffer
		if body != nil {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %v", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		} else {
			reqBody = bytes.NewBuffer(nil)
		}

		url := fmt.Sprintf("%s%s", client.baseURL, endpoint)
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// FUB API uses Basic Auth with API key as username and empty password
		auth := base64.StdEncoding.EncodeToString([]byte(client.apiKey + ":"))
		req.Header.Set("Authorization", "Basic "+auth)

		return client.client.Do(req)
	})
}

// CreateContact creates a new contact in FUB with behavioral intelligence data
func (client *BehavioralFUBAPIClient) CreateContact(contact *FUBContact) (*FUBContact, error) {
	resp, err := client.makeRequest("POST", "/people", contact)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to create contact: HTTP %d", resp.StatusCode)
	}

	var result FUBContact
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Created FUB contact: %s (ID: %s)", result.Name, result.ID)
	return &result, nil
}

// UpdateContact updates an existing contact in FUB
func (client *BehavioralFUBAPIClient) UpdateContact(contactID string, updates *FUBContact) (*FUBContact, error) {
	endpoint := fmt.Sprintf("/people/%s", contactID)
	resp, err := client.makeRequest("PUT", endpoint, updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to update contact: HTTP %d", resp.StatusCode)
	}

	var result FUBContact
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Updated FUB contact: %s (ID: %s)", result.Name, result.ID)
	return &result, nil
}

// CreateDeal creates a new deal in FUB with behavioral intelligence data
func (client *BehavioralFUBAPIClient) CreateDeal(deal *FUBDeal) (*FUBDeal, error) {
	resp, err := client.makeRequest("POST", "/deals", deal)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to create deal: HTTP %d", resp.StatusCode)
	}

	var result FUBDeal
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Created FUB deal: %s (ID: %s, Value: $%.0f)", result.Name, result.ID, result.Value)
	return &result, nil
}

// UpdateDeal updates an existing deal in FUB
func (client *BehavioralFUBAPIClient) UpdateDeal(dealID string, updates *FUBDeal) (*FUBDeal, error) {
	endpoint := fmt.Sprintf("/deals/%s", dealID)
	resp, err := client.makeRequest("PUT", endpoint, updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to update deal: HTTP %d", resp.StatusCode)
	}

	var result FUBDeal
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Updated FUB deal: %s (ID: %s)", result.Name, result.ID)
	return &result, nil
}

// AssignActionPlan assigns an action plan to a contact
func (client *BehavioralFUBAPIClient) AssignActionPlan(contactID, actionPlanID string) error {
	payload := map[string]interface{}{
		"actionPlanId": actionPlanID,
	}

	endpoint := fmt.Sprintf("/people/%s/actionplans", contactID)
	resp, err := client.makeRequest("POST", endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return fmt.Errorf("failed to assign action plan: HTTP %d", resp.StatusCode)
	}

	log.Printf("✅ Assigned action plan %s to contact %s", actionPlanID, contactID)
	return nil
}

// CreateActionPlan creates a new action plan in FUB
func (client *BehavioralFUBAPIClient) CreateActionPlan(actionPlan *FUBActionPlan) (*FUBActionPlan, error) {
	resp, err := client.makeRequest("POST", "/actionplans", actionPlan)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to create action plan: HTTP %d", resp.StatusCode)
	}

	var result FUBActionPlan
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Created FUB action plan: %s (ID: %s)", result.Name, result.ID)
	return &result, nil
}

// TriggerAutomation triggers a FUB automation with behavioral data
func (client *BehavioralFUBAPIClient) TriggerAutomation(automationID, contactID string, triggerData map[string]interface{}) error {
	payload := map[string]interface{}{
		"contactId":   contactID,
		"triggerData": triggerData,
	}

	endpoint := fmt.Sprintf("/automations/%s/trigger", automationID)
	resp, err := client.makeRequest("POST", endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("failed to trigger automation: HTTP %d", resp.StatusCode)
	}

	log.Printf("✅ Triggered automation %s for contact %s", automationID, contactID)
	return nil
}

// AddToPond adds a contact to a specific pond/list
func (client *BehavioralFUBAPIClient) AddToPond(contactID, pondID string) error {
	payload := map[string]interface{}{
		"contactId": contactID,
	}

	endpoint := fmt.Sprintf("/lists/%s/people", pondID)
	resp, err := client.makeRequest("POST", endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("failed to add to pond: HTTP %d", resp.StatusCode)
	}

	log.Printf("✅ Added contact %s to pond %s", contactID, pondID)
	return nil
}

// CreatePond creates a new pond/list in FUB
func (client *BehavioralFUBAPIClient) CreatePond(pond *FUBPond) (*FUBPond, error) {
	resp, err := client.makeRequest("POST", "/lists", pond)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to create pond: HTTP %d", resp.StatusCode)
	}

	var result FUBPond
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Created FUB pond: %s (ID: %s)", result.Name, result.ID)
	return &result, nil
}

// CreateTask creates a task in FUB for immediate agent action
func (client *BehavioralFUBAPIClient) CreateTask(contactID, title, description string, dueDate time.Time, priority string) error {
	payload := map[string]interface{}{
		"contactId":   contactID,
		"title":       title,
		"description": description,
		"dueDate":     dueDate.Format(time.RFC3339),
		"priority":    priority,
	}

	resp, err := client.makeRequest("POST", "/tasks", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create task: HTTP %d", resp.StatusCode)
	}

	log.Printf("✅ Created FUB task: %s for contact %s", title, contactID)
	return nil
}

// GetAgents retrieves available agents from FUB
func (client *BehavioralFUBAPIClient) GetAgents() ([]map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/users", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get agents: HTTP %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Retrieved %d FUB agents", len(result))
	return result, nil
}

// GetActionPlans retrieves available action plans from FUB
func (client *BehavioralFUBAPIClient) GetActionPlans() ([]FUBActionPlan, error) {
	resp, err := client.makeRequest("GET", "/actionplans", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get action plans: HTTP %d", resp.StatusCode)
	}

	var result []FUBActionPlan
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Retrieved %d FUB action plans", len(result))
	return result, nil
}

// GetPonds retrieves available ponds/lists from FUB
func (client *BehavioralFUBAPIClient) GetPonds() ([]FUBPond, error) {
	resp, err := client.makeRequest("GET", "/lists", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get ponds: HTTP %d", resp.StatusCode)
	}

	var result []FUBPond
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Retrieved %d FUB ponds", len(result))
	return result, nil
}

// SearchContactsByEmail searches for contacts by email address
func (client *BehavioralFUBAPIClient) SearchContactsByEmail(email string) ([]FUBContact, error) {
	endpoint := fmt.Sprintf("/people?email=%s", email)
	resp, err := client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to search contacts: HTTP %d", resp.StatusCode)
	}

	var result []FUBContact
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Found %d FUB contacts for email %s", len(result), email)
	return result, nil
}

// SearchContactsByPhone searches for contacts by phone number
func (client *BehavioralFUBAPIClient) SearchContactsByPhone(phone string) ([]FUBContact, error) {
	endpoint := fmt.Sprintf("/people?phone=%s", phone)
	resp, err := client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to search contacts: HTTP %d", resp.StatusCode)
	}

	var result []FUBContact
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Found %d FUB contacts for phone %s", len(result), phone)
	return result, nil
}

// SearchContactsByName searches for contacts by name
func (client *BehavioralFUBAPIClient) SearchContactsByName(name string) ([]FUBContact, error) {
	endpoint := fmt.Sprintf("/people?name=%s", name)
	resp, err := client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to search contacts: HTTP %d", resp.StatusCode)
	}

	var result []FUBContact
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Found %d FUB contacts for name %s", len(result), name)
	return result, nil
}
