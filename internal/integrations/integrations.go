package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// IntegrationManager manages all external integrations
type IntegrationManager struct {
	fubClient *FUBClient
	ctx       context.Context
	cancel    context.CancelFunc
}

// FUBClient handles Follow Up Boss CRM integration
type FUBClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Integration types
const (
	IntegrationFUB = "fub"
)

// FUB API endpoints
const (
	FUBEndpointPeople      = "/people"
	FUBEndpointEvents      = "/events"
	FUBEndpointTags        = "/tags"
	FUBEndpointNotes       = "/notes"
	FUBEndpointTasks       = "/tasks"
	FUBEndpointPipelines   = "/pipelines"
	FUBEndpointWebhooks    = "/webhooks"
)

// Integration status
type IntegrationStatus string

const (
	StatusActive     IntegrationStatus = "active"
	StatusInactive   IntegrationStatus = "inactive"
	StatusError      IntegrationStatus = "error"
	StatusSyncing    IntegrationStatus = "syncing"
)

// SyncResult represents integration sync results
type SyncResult struct {
	Integration   string    `json:"integration"`
	RecordsTotal  int       `json:"records_total"`
	RecordsNew    int       `json:"records_new"`
	RecordsUpdated int      `json:"records_updated"`
	RecordsErrors int       `json:"records_errors"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Duration      string    `json:"duration"`
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// FUBContact represents a Follow Up Boss contact
type FUBContact struct {
	ID              string            `json:"id"`
	FirstName       string            `json:"firstName"`
	LastName        string            `json:"lastName"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	Source          string            `json:"source"`
	Status          string            `json:"status"`
	Stage           string            `json:"stage"`
	Tags            []string          `json:"tags"`
	AssignedTo      string            `json:"assignedTo"`
	CreatedDate     time.Time         `json:"createdDate"`
	LastActivity    time.Time         `json:"lastActivity"`
	NextFollowUp    *time.Time        `json:"nextFollowUp,omitempty"`
	LeadScore       int               `json:"leadScore"`
	CustomFields    map[string]string `json:"customFields"`
	PropertyInterest []PropertyInterest `json:"propertyInterest"`
	Notes           []ContactNote     `json:"notes"`
	Events          []ContactEvent    `json:"events"`
}

// PropertyInterest represents property interest data
type PropertyInterest struct {
	PropertyType string   `json:"property_type"`
	MinPrice     float64  `json:"min_price"`
	MaxPrice     float64  `json:"max_price"`
	Bedrooms     int      `json:"bedrooms"`
	Bathrooms    float64  `json:"bathrooms"`
	Areas        []string `json:"areas"`
	Features     []string `json:"features"`
}

// ContactNote represents a contact note
type ContactNote struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"`
}

// ContactEvent represents a contact event
type ContactEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(fubAPIKey string) *IntegrationManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &IntegrationManager{
		fubClient: &FUBClient{
			apiKey:  fubAPIKey,
			baseURL: "https://api.followupboss.com/v1",
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// GetFUBContacts fetches contacts from Follow Up Boss
func (im *IntegrationManager) GetFUBContacts(filters map[string]interface{}) ([]FUBContact, error) {
	url := fmt.Sprintf("%s%s", im.fubClient.baseURL, FUBEndpointPeople)
	
	// Build query parameters
	queryParams := make([]string, 0)
	for key, value := range filters {
		queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
	}
	
	if len(queryParams) > 0 {
		url += "?" + strings.Join(queryParams, "&")
	}
	
	req, err := http.NewRequestWithContext(im.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Basic "+im.fubClient.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := im.fubClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FUB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		People []FUBContact `json:"people"`
		Total  int          `json:"total"`
		Page   int          `json:"page"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	return response.People, nil
}

// CreateFUBContact creates a new contact in Follow Up Boss
func (im *IntegrationManager) CreateFUBContact(contact *FUBContact) (*FUBContact, error) {
	url := fmt.Sprintf("%s%s", im.fubClient.baseURL, FUBEndpointPeople)
	
	jsonData, err := json.Marshal(contact)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(im.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Basic "+im.fubClient.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := im.fubClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FUB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var createdContact FUBContact
	if err := json.NewDecoder(resp.Body).Decode(&createdContact); err != nil {
		return nil, err
	}
	
	return &createdContact, nil
}

// UpdateFUBContact updates an existing Follow Up Boss contact
func (im *IntegrationManager) UpdateFUBContact(contactID string, updates map[string]interface{}) error {
	url := fmt.Sprintf("%s%s/%s", im.fubClient.baseURL, FUBEndpointPeople, contactID)
	
	jsonData, err := json.Marshal(updates)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(im.ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Basic "+im.fubClient.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := im.fubClient.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FUB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// SyncFUBContacts synchronizes Follow Up Boss contacts
func (im *IntegrationManager) SyncFUBContacts(filters map[string]interface{}) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		Integration: IntegrationFUB,
		StartTime:   startTime,
		Success:     true,
	}
	
	contacts, err := im.GetFUBContacts(filters)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		return result, err
	}
	
	result.RecordsTotal = len(contacts)
	
	// Process each contact
	for _, contact := range contacts {
		if err := im.processFUBContact(contact); err != nil {
			result.RecordsErrors++
			log.Printf("Error processing FUB contact %s: %v", contact.ID, err)
		} else {
			result.RecordsNew++
		}
	}
	
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()
	
	log.Printf("FUB sync completed: %d total, %d new, %d errors", 
		result.RecordsTotal, result.RecordsNew, result.RecordsErrors)
	
	return result, nil
}

// AddFUBContactNote adds a note to a Follow Up Boss contact
func (im *IntegrationManager) AddFUBContactNote(contactID, note, author string) error {
	url := fmt.Sprintf("%s%s/%s%s", im.fubClient.baseURL, FUBEndpointPeople, contactID, FUBEndpointNotes)
	
	noteData := map[string]interface{}{
		"content": note,
		"author":  author,
		"type":    "note",
	}
	
	jsonData, err := json.Marshal(noteData)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(im.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Basic "+im.fubClient.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := im.fubClient.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FUB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// CreateFUBTask creates a task in Follow Up Boss
func (im *IntegrationManager) CreateFUBTask(contactID, title, description string, dueDate time.Time) error {
	url := fmt.Sprintf("%s%s", im.fubClient.baseURL, FUBEndpointTasks)
	
	taskData := map[string]interface{}{
		"contactId":   contactID,
		"title":       title,
		"description": description,
		"dueDate":     dueDate.Format(time.RFC3339),
		"type":        "task",
	}
	
	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(im.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Basic "+im.fubClient.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := im.fubClient.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FUB API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// processFUBContact processes a FUB contact and converts to internal model
func (im *IntegrationManager) processFUBContact(contact FUBContact) error {
	// This would typically convert FUB contact to internal model and save to database
	// For now, just log the contact information
	log.Printf("Processed FUB contact: %s %s (%s)", contact.FirstName, contact.LastName, contact.Email)
	
	return nil
}

// HealthCheck checks the health of all integrations
func (im *IntegrationManager) HealthCheck() map[string]IntegrationStatus {
	status := make(map[string]IntegrationStatus)
	
	// Check FUB integration
	if err := im.testFUBConnection(); err != nil {
		status[IntegrationFUB] = StatusError
		log.Printf("FUB integration health check failed: %v", err)
	} else {
		status[IntegrationFUB] = StatusActive
	}
	
	return status
}

// testFUBConnection tests FUB API connection
func (im *IntegrationManager) testFUBConnection() error {
	url := fmt.Sprintf("%s/account", im.fubClient.baseURL)
	
	req, err := http.NewRequestWithContext(im.ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Basic "+im.fubClient.apiKey)
	
	resp, err := im.fubClient.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FUB API returned status %d", resp.StatusCode)
	}
	
	return nil
}

// Close stops the integration manager
func (im *IntegrationManager) Close() {
	im.cancel()
}
