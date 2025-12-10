package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type AppFolioAPIClient struct {
	clientID     string
	clientSecret string
	baseURL      string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
}

type AppFolioTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type AppFolioTenant struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PropertyID      string    `json:"property_id"`
	UnitID          string    `json:"unit_id"`
	Status          string    `json:"status"`
	LeaseStart      time.Time `json:"lease_start"`
	LeaseEnd        time.Time `json:"lease_end"`
	MonthlyRent     float64   `json:"monthly_rent"`
	SecurityDeposit float64   `json:"security_deposit"`
	MoveInDate      time.Time `json:"move_in_date"`
	MoveOutDate     time.Time `json:"move_out_date,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AppFolioProperty struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	State         string  `json:"state"`
	ZipCode       string  `json:"zip_code"`
	PropertyType  string  `json:"property_type"`
	Units         int     `json:"units"`
	Status        string  `json:"status"`
	MonthlyRent   float64 `json:"monthly_rent"`
}

type AppFolioUnit struct {
	ID           string  `json:"id"`
	PropertyID   string  `json:"property_id"`
	UnitNumber   string  `json:"unit_number"`
	Bedrooms     int     `json:"bedrooms"`
	Bathrooms    float32 `json:"bathrooms"`
	SquareFeet   int     `json:"square_feet"`
	MonthlyRent  float64 `json:"monthly_rent"`
	Status       string  `json:"status"`
	CurrentLease string  `json:"current_lease_id,omitempty"`
}

type AppFolioAPIResponse struct {
	Data    json.RawMessage `json:"data"`
	Meta    json.RawMessage `json:"meta,omitempty"`
	Error   string          `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
}

type AppFolioTenantsResponse struct {
	Tenants    []AppFolioTenant `json:"tenants"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
}

type CreateTenantRequest struct {
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PropertyID      string    `json:"property_id"`
	UnitID          string    `json:"unit_id"`
	LeaseStart      time.Time `json:"lease_start"`
	LeaseEnd        time.Time `json:"lease_end"`
	MonthlyRent     float64   `json:"monthly_rent"`
	SecurityDeposit float64   `json:"security_deposit"`
}

type UpdateTenantRequest struct {
	Name       string    `json:"name,omitempty"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Status     string    `json:"status,omitempty"`
	LeaseStart time.Time `json:"lease_start,omitempty"`
	LeaseEnd   time.Time `json:"lease_end,omitempty"`
}

func NewAppFolioAPIClient(clientID, clientSecret, baseURL string) *AppFolioAPIClient {
	if baseURL == "" {
		baseURL = "https://api.appfolio.com/v1"
	}

	return &AppFolioAPIClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AppFolioAPIClient) authenticate() error {
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequest("POST", c.baseURL+"/oauth/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp AppFolioTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	log.Printf("✅ AppFolio authentication successful, token expires in %d seconds", tokenResp.ExpiresIn)
	return nil
}

func (c *AppFolioAPIClient) doRequest(method, endpoint string, body interface{}) (*AppFolioAPIResponse, error) {
	if err := c.authenticate(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp AppFolioAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		apiResp.Data = respBody
	}

	return &apiResp, nil
}

func (c *AppFolioAPIClient) GetTenants(page, perPage int) (*AppFolioTenantsResponse, error) {
	endpoint := fmt.Sprintf("/tenants?page=%d&per_page=%d", page, perPage)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var tenantsResp AppFolioTenantsResponse
	if err := json.Unmarshal(resp.Data, &tenantsResp); err != nil {
		var tenants []AppFolioTenant
		if err := json.Unmarshal(resp.Data, &tenants); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tenants response: %w", err)
		}
		tenantsResp.Tenants = tenants
		tenantsResp.TotalCount = len(tenants)
	}

	return &tenantsResp, nil
}

func (c *AppFolioAPIClient) GetTenant(tenantID string) (*AppFolioTenant, error) {
	resp, err := c.doRequest("GET", "/tenants/"+tenantID, nil)
	if err != nil {
		return nil, err
	}

	var tenant AppFolioTenant
	if err := json.Unmarshal(resp.Data, &tenant); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant: %w", err)
	}

	return &tenant, nil
}

func (c *AppFolioAPIClient) GetTenantByEmail(email string) (*AppFolioTenant, error) {
	endpoint := fmt.Sprintf("/tenants?email=%s", url.QueryEscape(email))
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var tenants []AppFolioTenant
	if err := json.Unmarshal(resp.Data, &tenants); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenants: %w", err)
	}

	if len(tenants) == 0 {
		return nil, nil
	}

	return &tenants[0], nil
}

func (c *AppFolioAPIClient) CreateTenant(req *CreateTenantRequest) (*AppFolioTenant, error) {
	resp, err := c.doRequest("POST", "/tenants", req)
	if err != nil {
		return nil, err
	}

	var tenant AppFolioTenant
	if err := json.Unmarshal(resp.Data, &tenant); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created tenant: %w", err)
	}

	log.Printf("✅ Created tenant in AppFolio: %s (ID: %s)", tenant.Name, tenant.ID)
	return &tenant, nil
}

func (c *AppFolioAPIClient) UpdateTenant(tenantID string, req *UpdateTenantRequest) (*AppFolioTenant, error) {
	resp, err := c.doRequest("PUT", "/tenants/"+tenantID, req)
	if err != nil {
		return nil, err
	}

	var tenant AppFolioTenant
	if err := json.Unmarshal(resp.Data, &tenant); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated tenant: %w", err)
	}

	log.Printf("✅ Updated tenant in AppFolio: %s (ID: %s)", tenant.Name, tenant.ID)
	return &tenant, nil
}

func (c *AppFolioAPIClient) UpdateTenantStatus(tenantID, status string) error {
	req := &UpdateTenantRequest{Status: status}
	_, err := c.doRequest("PUT", "/tenants/"+tenantID, req)
	if err != nil {
		return err
	}

	log.Printf("✅ Updated tenant status in AppFolio: %s -> %s", tenantID, status)
	return nil
}

func (c *AppFolioAPIClient) DeleteTenant(tenantID string) error {
	_, err := c.doRequest("DELETE", "/tenants/"+tenantID, nil)
	if err != nil {
		return err
	}

	log.Printf("✅ Deleted tenant from AppFolio: %s", tenantID)
	return nil
}

func (c *AppFolioAPIClient) GetProperties(page, perPage int) ([]AppFolioProperty, error) {
	endpoint := fmt.Sprintf("/properties?page=%d&per_page=%d", page, perPage)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var properties []AppFolioProperty
	if err := json.Unmarshal(resp.Data, &properties); err != nil {
		return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
	}

	return properties, nil
}

func (c *AppFolioAPIClient) GetProperty(propertyID string) (*AppFolioProperty, error) {
	resp, err := c.doRequest("GET", "/properties/"+propertyID, nil)
	if err != nil {
		return nil, err
	}

	var property AppFolioProperty
	if err := json.Unmarshal(resp.Data, &property); err != nil {
		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
	}

	return &property, nil
}

func (c *AppFolioAPIClient) GetPropertyByAddress(address string) (*AppFolioProperty, error) {
	endpoint := fmt.Sprintf("/properties?address=%s", url.QueryEscape(address))
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var properties []AppFolioProperty
	if err := json.Unmarshal(resp.Data, &properties); err != nil {
		return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
	}

	if len(properties) == 0 {
		return nil, nil
	}

	return &properties[0], nil
}

func (c *AppFolioAPIClient) GetUnits(propertyID string) ([]AppFolioUnit, error) {
	endpoint := fmt.Sprintf("/properties/%s/units", propertyID)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var units []AppFolioUnit
	if err := json.Unmarshal(resp.Data, &units); err != nil {
		return nil, fmt.Errorf("failed to unmarshal units: %w", err)
	}

	return units, nil
}

func (c *AppFolioAPIClient) GetUnit(unitID string) (*AppFolioUnit, error) {
	resp, err := c.doRequest("GET", "/units/"+unitID, nil)
	if err != nil {
		return nil, err
	}

	var unit AppFolioUnit
	if err := json.Unmarshal(resp.Data, &unit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal unit: %w", err)
	}

	return &unit, nil
}

func (c *AppFolioAPIClient) HealthCheck() error {
	_, err := c.doRequest("GET", "/health", nil)
	return err
}
