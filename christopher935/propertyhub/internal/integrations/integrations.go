package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/utils"
)

// IntegrationManager manages all external integrations
type IntegrationManager struct {
	harClient *HARClient
	fubClient *FUBClient
	ctx       context.Context
	cancel    context.CancelFunc
}

// HARClient handles Houston Association of Realtors MLS integration
type HARClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// FUBClient handles Follow Up Boss CRM integration
type FUBClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Integration types
const (
	IntegrationHAR = "har"
	IntegrationFUB = "fub"
)

// HAR API endpoints
const (
	HAREndpointListings     = "/listings"
	HAREndpointComparables  = "/comparables"
	HAREndpointMarketStats  = "/market-stats"
	HAREndpointProperties   = "/properties"
	HAREndpointAgents       = "/agents"
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

// HARListing represents a HAR MLS listing
type HARListing struct {
	MLSID           string    `json:"mls_id"`
	Address         string    `json:"address"`
	City            string    `json:"city"`
	State           string    `json:"state"`
	ZipCode         string    `json:"zip_code"`
	County          string    `json:"county"`
	ListPrice       float64   `json:"list_price"`
	SoldPrice       float64   `json:"sold_price,omitempty"`
	Status          string    `json:"status"`
	PropertyType    string    `json:"property_type"`
	Bedrooms        int       `json:"bedrooms"`
	Bathrooms       float64   `json:"bathrooms"`
	SquareFeet      int       `json:"square_feet"`
	LotSize         float64   `json:"lot_size"`
	YearBuilt       int       `json:"year_built"`
	ListingDate     time.Time `json:"listing_date"`
	SoldDate        *time.Time `json:"sold_date,omitempty"`
	DaysOnMarket    int       `json:"days_on_market"`
	Photos          []string  `json:"photos"`
	Description     string    `json:"description"`
	Features        []string  `json:"features"`
	ListingAgent    string    `json:"listing_agent"`
	ListingOffice   string    `json:"listing_office"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	LastUpdated     time.Time `json:"last_updated"`
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
func NewIntegrationManager(harAPIKey, fubAPIKey string) *IntegrationManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &IntegrationManager{
		harClient: &HARClient{
			apiKey:  harAPIKey,
			baseURL: "https://api.har.com/v1",
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		},
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

// GetHARListings fetches listings from HAR MLS
func (im *IntegrationManager) GetHARListings(filters map[string]interface{}) ([]HARListing, error) {
	url := fmt.Sprintf("%s%s", im.harClient.baseURL, HAREndpointListings)
	
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
	
	req.Header.Set("Authorization", "Bearer "+im.harClient.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := im.harClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HAR API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Listings []HARListing `json:"listings"`
		Total    int          `json:"total"`
		Page     int          `json:"page"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	return response.Listings, nil
}

// GetMarketComparables gets comparable properties from HAR
func (im *IntegrationManager) GetMarketComparables(address string, filters map[string]interface{}) ([]HARListing, error) {
	url := fmt.Sprintf("%s%s", im.harClient.baseURL, HAREndpointComparables)
	
	requestData := map[string]interface{}{
		"address": address,
		"filters": filters,
	}
	
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(im.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+im.harClient.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := im.harClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HAR API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Comparables []HARListing `json:"comparables"`
		Total       int          `json:"total"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	return response.Comparables, nil
}

// SyncHARListings synchronizes HAR listings
func (im *IntegrationManager) SyncHARListings(filters map[string]interface{}) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		Integration: IntegrationHAR,
		StartTime:   startTime,
		Success:     true,
	}
	
	listings, err := im.GetHARListings(filters)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		return result, err
	}
	
	result.RecordsTotal = len(listings)
	
	// Process each listing (this would typically save to database)
	for _, listing := range listings {
		// Convert to internal model and save
		if err := im.processHARListing(listing); err != nil {
			result.RecordsErrors++
			log.Printf("Error processing HAR listing %s: %v", listing.MLSID, err)
		} else {
			result.RecordsNew++ // This would be determined by checking if record exists
		}
	}
	
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()
	
	log.Printf("HAR sync completed: %d total, %d new, %d errors", 
		result.RecordsTotal, result.RecordsNew, result.RecordsErrors)
	
	return result, nil
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

// processHARListing processes a HAR listing and converts to internal model
func (im *IntegrationManager) processHARListing(listing HARListing) error {
	// Convert HAR listing to internal Property model
	property := &models.Property{
		ID:              utils.GenerateID(),
		ExternalID:      listing.MLSID,
		Source:          "HAR",
		Title:           fmt.Sprintf("%s, %s", listing.Address, listing.City),
		Description:     listing.Description,
		Address:         listing.Address,
		City:            listing.City,
		State:           listing.State,
		ZipCode:         listing.ZipCode,
		Country:         "USA",
		Latitude:        listing.Latitude,
		Longitude:       listing.Longitude,
		PropertyType:    listing.PropertyType,
		Bedrooms:        listing.Bedrooms,
		Bathrooms:       listing.Bathrooms,
		MaxGuests:       listing.Bedrooms * 2, // Estimate
		BasePrice:       listing.ListPrice * 0.003, // Estimate daily rate
		CleaningFee:     75.0,
		SquareFeet:      listing.SquareFeet,
		LotSize:         listing.LotSize,
		YearBuilt:       listing.YearBuilt,
		Features:        strings.Join(listing.Features, ","),
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// This would typically save to database via repository
	log.Printf("Processed HAR listing: %s", property.Title)
	
	return nil
}

// processFUBContact processes a FUB contact and converts to internal model
func (im *IntegrationManager) processFUBContact(contact FUBContact) error {
	// Convert FUB contact to internal User model
	user := &models.User{
		ID:              utils.GenerateID(),
		ExternalID:      contact.ID,
		FirstName:       contact.FirstName,
		LastName:        contact.LastName,
		Email:           contact.Email,
		Phone:           contact.Phone,
		Source:          "FUB",
		Role:            "user",
		Status:          "active",
		EmailVerified:   true,
		CreatedAt:       contact.CreatedDate,
		UpdatedAt:       time.Now(),
		LastLoginAt:     contact.LastActivity,
	}
	
	// This would typically save to database via repository
	log.Printf("Processed FUB contact: %s %s", user.FirstName, user.LastName)
	
	return nil
}

// HealthCheck checks the health of all integrations
func (im *IntegrationManager) HealthCheck() map[string]IntegrationStatus {
	status := make(map[string]IntegrationStatus)
	
	// Check HAR integration
	if err := im.testHARConnection(); err != nil {
		status[IntegrationHAR] = StatusError
		log.Printf("HAR integration health check failed: %v", err)
	} else {
		status[IntegrationHAR] = StatusActive
	}
	
	// Check FUB integration
	if err := im.testFUBConnection(); err != nil {
		status[IntegrationFUB] = StatusError
		log.Printf("FUB integration health check failed: %v", err)
	} else {
		status[IntegrationFUB] = StatusActive
	}
	
	return status
}

// testHARConnection tests HAR API connection
func (im *IntegrationManager) testHARConnection() error {
	url := fmt.Sprintf("%s/health", im.harClient.baseURL)
	
	req, err := http.NewRequestWithContext(im.ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+im.harClient.apiKey)
	
	resp, err := im.harClient.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HAR API returned status %d", resp.StatusCode)
	}
	
	return nil
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
