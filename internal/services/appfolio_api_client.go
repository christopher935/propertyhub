package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type AppFolioAPIClient struct {
	db           *gorm.DB
	client       *http.Client
	apiKey       string
	clientID     string
	clientSecret string
	baseURL      string
	errorHandler *AppFolioErrorHandler
}

type AppFolioClientConfig struct {
	APIKey       string
	ClientID     string
	ClientSecret string
	BaseURL      string
	Timeout      time.Duration
}

func NewAppFolioAPIClient(db *gorm.DB, config AppFolioClientConfig) *AppFolioAPIClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.appfolio.com/v1"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &AppFolioAPIClient{
		db:           db,
		client:       &http.Client{Timeout: config.Timeout},
		apiKey:       config.APIKey,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		baseURL:      config.BaseURL,
		errorHandler: NewAppFolioErrorHandler(),
	}
}

func NewAppFolioAPIClientSimple(db *gorm.DB, apiKey string) *AppFolioAPIClient {
	return NewAppFolioAPIClient(db, AppFolioClientConfig{
		APIKey: apiKey,
	})
}

func (c *AppFolioAPIClient) makeRequest(method, endpoint string, body interface{}, queryParams map[string]string) (*http.Response, error) {
	operation := fmt.Sprintf("%s %s", method, endpoint)

	return c.errorHandler.ExecuteWithRetry(operation, func() (*http.Response, error) {
		var reqBody io.Reader
		if body != nil {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %v", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		}

		fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
		if len(queryParams) > 0 {
			params := url.Values{}
			for key, value := range queryParams {
				params.Add(key, value)
			}
			fullURL = fmt.Sprintf("%s?%s", fullURL, params.Encode())
		}

		req, err := http.NewRequest(method, fullURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "PropertyHub-AppFolio-Client/1.0")

		if c.apiKey != "" {
			req.Header.Set("X-AppFolio-API-Key", c.apiKey)
		}

		if c.clientID != "" && c.clientSecret != "" {
			req.Header.Set("X-Client-ID", c.clientID)
			req.Header.Set("X-Client-Secret", c.clientSecret)
		}

		return c.client.Do(req)
	})
}

func (c *AppFolioAPIClient) Get(endpoint string, queryParams map[string]string) (*http.Response, error) {
	return c.makeRequest("GET", endpoint, nil, queryParams)
}

func (c *AppFolioAPIClient) Post(endpoint string, body interface{}) (*http.Response, error) {
	return c.makeRequest("POST", endpoint, body, nil)
}

func (c *AppFolioAPIClient) Put(endpoint string, body interface{}) (*http.Response, error) {
	return c.makeRequest("PUT", endpoint, body, nil)
}

func (c *AppFolioAPIClient) Patch(endpoint string, body interface{}) (*http.Response, error) {
	return c.makeRequest("PATCH", endpoint, body, nil)
}

func (c *AppFolioAPIClient) Delete(endpoint string) (*http.Response, error) {
	return c.makeRequest("DELETE", endpoint, nil, nil)
}

type AppFolioPropertyResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	PropertyType string                 `json:"property_type"`
	Address      map[string]interface{} `json:"address"`
	UnitCount    int                    `json:"unit_count"`
	Status       string                 `json:"status"`
	RentAmount   float64                `json:"rent_amount"`
	OwnerID      string                 `json:"owner_id"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type AppFolioTenantResponse struct {
	ID             string     `json:"id"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	UnitID         string     `json:"unit_id"`
	PropertyID     string     `json:"property_id"`
	LeaseStartDate *time.Time `json:"lease_start_date"`
	LeaseEndDate   *time.Time `json:"lease_end_date"`
	RentAmount     float64    `json:"rent_amount"`
	Balance        float64    `json:"balance"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type AppFolioMaintenanceResponse struct {
	ID            string     `json:"id"`
	PropertyID    string     `json:"property_id"`
	UnitID        string     `json:"unit_id"`
	TenantID      string     `json:"tenant_id"`
	Description   string     `json:"description"`
	Priority      string     `json:"priority"`
	Status        string     `json:"status"`
	AssignedTo    string     `json:"assigned_to"`
	ScheduledAt   *time.Time `json:"scheduled_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	EstimatedCost float64    `json:"estimated_cost"`
	ActualCost    float64    `json:"actual_cost"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AppFolioPaymentResponse struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	PropertyID    string     `json:"property_id"`
	Amount        float64    `json:"amount"`
	DueDate       time.Time  `json:"due_date"`
	PaidDate      *time.Time `json:"paid_date"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	TransactionID string     `json:"transaction_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AppFolioOwnerResponse struct {
	ID            string    `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	CompanyName   string    `json:"company_name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	PropertyIDs   []string  `json:"property_ids"`
	PropertyCount int       `json:"property_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AppFolioListParams struct {
	Page     int
	PageSize int
	Status   string
	Search   string
	SortBy   string
	SortDir  string
}

func (p AppFolioListParams) ToQueryParams() map[string]string {
	params := make(map[string]string)
	if p.Page > 0 {
		params["page"] = strconv.Itoa(p.Page)
	}
	if p.PageSize > 0 {
		params["page_size"] = strconv.Itoa(p.PageSize)
	}
	if p.Status != "" {
		params["status"] = p.Status
	}
	if p.Search != "" {
		params["search"] = p.Search
	}
	if p.SortBy != "" {
		params["sort_by"] = p.SortBy
	}
	if p.SortDir != "" {
		params["sort_dir"] = p.SortDir
	}
	return params
}

func (c *AppFolioAPIClient) ListProperties(params AppFolioListParams) ([]AppFolioPropertyResponse, int, error) {
	resp, err := c.Get("/properties", params.ToQueryParams())
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioPropertyResponse `json:"data"`
		TotalCount int                        `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d properties (total: %d)", len(result.Data), result.TotalCount)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetProperty(propertyID string) (*AppFolioPropertyResponse, error) {
	endpoint := fmt.Sprintf("/properties/%s", propertyID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioPropertyResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved property %s", propertyID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) CreateProperty(property map[string]interface{}) (*AppFolioPropertyResponse, error) {
	validationErrors := ValidateAppFolioPropertyRequest(property)
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", validationErrors)
	}

	resp, err := c.Post("/properties", property)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioPropertyResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Created property %s", result.Data.ID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) UpdateProperty(propertyID string, updates map[string]interface{}) (*AppFolioPropertyResponse, error) {
	endpoint := fmt.Sprintf("/properties/%s", propertyID)
	resp, err := c.Put(endpoint, updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioPropertyResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Updated property %s", propertyID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) DeleteProperty(propertyID string) error {
	endpoint := fmt.Sprintf("/properties/%s", propertyID)
	resp, err := c.Delete(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("✅ AppFolio: Deleted property %s", propertyID)
	return nil
}

func (c *AppFolioAPIClient) ListTenants(params AppFolioListParams) ([]AppFolioTenantResponse, int, error) {
	resp, err := c.Get("/tenants", params.ToQueryParams())
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioTenantResponse `json:"data"`
		TotalCount int                      `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d tenants (total: %d)", len(result.Data), result.TotalCount)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetTenant(tenantID string) (*AppFolioTenantResponse, error) {
	endpoint := fmt.Sprintf("/tenants/%s", tenantID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioTenantResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved tenant %s", tenantID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) CreateTenant(tenant map[string]interface{}) (*AppFolioTenantResponse, error) {
	validationErrors := ValidateAppFolioTenantRequest(tenant)
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", validationErrors)
	}

	resp, err := c.Post("/tenants", tenant)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioTenantResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Created tenant %s", result.Data.ID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) UpdateTenant(tenantID string, updates map[string]interface{}) (*AppFolioTenantResponse, error) {
	endpoint := fmt.Sprintf("/tenants/%s", tenantID)
	resp, err := c.Put(endpoint, updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioTenantResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Updated tenant %s", tenantID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) DeleteTenant(tenantID string) error {
	endpoint := fmt.Sprintf("/tenants/%s", tenantID)
	resp, err := c.Delete(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("✅ AppFolio: Deleted tenant %s", tenantID)
	return nil
}

func (c *AppFolioAPIClient) ListMaintenanceRequests(params AppFolioListParams) ([]AppFolioMaintenanceResponse, int, error) {
	resp, err := c.Get("/maintenance_requests", params.ToQueryParams())
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioMaintenanceResponse `json:"data"`
		TotalCount int                           `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d maintenance requests (total: %d)", len(result.Data), result.TotalCount)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetMaintenanceRequest(requestID string) (*AppFolioMaintenanceResponse, error) {
	endpoint := fmt.Sprintf("/maintenance_requests/%s", requestID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioMaintenanceResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved maintenance request %s", requestID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) CreateMaintenanceRequest(request map[string]interface{}) (*AppFolioMaintenanceResponse, error) {
	validationErrors := ValidateAppFolioMaintenanceRequest(request)
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", validationErrors)
	}

	resp, err := c.Post("/maintenance_requests", request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioMaintenanceResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Created maintenance request %s", result.Data.ID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) UpdateMaintenanceRequest(requestID string, updates map[string]interface{}) (*AppFolioMaintenanceResponse, error) {
	endpoint := fmt.Sprintf("/maintenance_requests/%s", requestID)
	resp, err := c.Put(endpoint, updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioMaintenanceResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Updated maintenance request %s", requestID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) ListPayments(params AppFolioListParams) ([]AppFolioPaymentResponse, int, error) {
	resp, err := c.Get("/payments", params.ToQueryParams())
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioPaymentResponse `json:"data"`
		TotalCount int                       `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d payments (total: %d)", len(result.Data), result.TotalCount)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetPayment(paymentID string) (*AppFolioPaymentResponse, error) {
	endpoint := fmt.Sprintf("/payments/%s", paymentID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioPaymentResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved payment %s", paymentID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) RecordPayment(payment map[string]interface{}) (*AppFolioPaymentResponse, error) {
	validationErrors := ValidateAppFolioPaymentRequest(payment)
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", validationErrors)
	}

	resp, err := c.Post("/payments", payment)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioPaymentResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Recorded payment %s", result.Data.ID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) ListOwners(params AppFolioListParams) ([]AppFolioOwnerResponse, int, error) {
	resp, err := c.Get("/owners", params.ToQueryParams())
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioOwnerResponse `json:"data"`
		TotalCount int                     `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d owners (total: %d)", len(result.Data), result.TotalCount)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetOwner(ownerID string) (*AppFolioOwnerResponse, error) {
	endpoint := fmt.Sprintf("/owners/%s", ownerID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data AppFolioOwnerResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved owner %s", ownerID)
	return &result.Data, nil
}

func (c *AppFolioAPIClient) GetTenantsByProperty(propertyID string, params AppFolioListParams) ([]AppFolioTenantResponse, int, error) {
	queryParams := params.ToQueryParams()
	queryParams["property_id"] = propertyID

	resp, err := c.Get("/tenants", queryParams)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioTenantResponse `json:"data"`
		TotalCount int                      `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d tenants for property %s", len(result.Data), propertyID)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetMaintenanceByProperty(propertyID string, params AppFolioListParams) ([]AppFolioMaintenanceResponse, int, error) {
	queryParams := params.ToQueryParams()
	queryParams["property_id"] = propertyID

	resp, err := c.Get("/maintenance_requests", queryParams)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioMaintenanceResponse `json:"data"`
		TotalCount int                           `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d maintenance requests for property %s", len(result.Data), propertyID)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetPaymentsByTenant(tenantID string, params AppFolioListParams) ([]AppFolioPaymentResponse, int, error) {
	queryParams := params.ToQueryParams()
	queryParams["tenant_id"] = tenantID

	resp, err := c.Get("/payments", queryParams)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioPaymentResponse `json:"data"`
		TotalCount int                       `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d payments for tenant %s", len(result.Data), tenantID)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) GetPropertiesByOwner(ownerID string, params AppFolioListParams) ([]AppFolioPropertyResponse, int, error) {
	queryParams := params.ToQueryParams()
	queryParams["owner_id"] = ownerID

	resp, err := c.Get("/properties", queryParams)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data       []AppFolioPropertyResponse `json:"data"`
		TotalCount int                        `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ AppFolio: Retrieved %d properties for owner %s", len(result.Data), ownerID)
	return result.Data, result.TotalCount, nil
}

func (c *AppFolioAPIClient) HealthCheck() error {
	resp, err := c.Get("/health", nil)
	if err != nil {
		return fmt.Errorf("AppFolio API health check failed: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("✅ AppFolio: API health check passed")
	return nil
}

func (c *AppFolioAPIClient) GetAPIStatus() (map[string]interface{}, error) {
	resp, err := c.Get("/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result, nil
}

func (c *AppFolioAPIClient) IsConfigured() bool {
	return c.apiKey != "" || (c.clientID != "" && c.clientSecret != "")
}

func (c *AppFolioAPIClient) GetBaseURL() string {
	return c.baseURL
}

func (c *AppFolioAPIClient) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}
