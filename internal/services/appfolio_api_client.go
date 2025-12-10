package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type AppFolioAPIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type AppFolioProperty struct {
	PropertyID    string   `json:"property_id"`
	Address       string   `json:"address"`
	Address2      string   `json:"address2"`
	City          string   `json:"city"`
	State         string   `json:"state"`
	Zip           string   `json:"zip"`
	Bedrooms      int      `json:"bedrooms"`
	Bathrooms     float32  `json:"bathrooms"`
	SquareFeet    int      `json:"square_feet"`
	RentAmount    float64  `json:"rent_amount"`
	MarketRent    float64  `json:"market_rent"`
	Status        string   `json:"status"`
	PropertyType  string   `json:"property_type"`
	Description   string   `json:"description"`
	YearBuilt     int      `json:"year_built"`
	AvailableDate string   `json:"available_date"`
	OwnerID       string   `json:"owner_id"`
	OwnerName     string   `json:"owner_name"`
	UnitCount     int      `json:"unit_count"`
	Features      string   `json:"features"`
	Images        []string `json:"images"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type AppFolioPropertiesResponse struct {
	Properties []AppFolioProperty `json:"properties"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	HasMore    bool               `json:"has_more"`
}

type AppFolioVacancy struct {
	PropertyID    string    `json:"property_id"`
	UnitID        string    `json:"unit_id"`
	Address       string    `json:"address"`
	UnitNumber    string    `json:"unit_number"`
	Bedrooms      int       `json:"bedrooms"`
	Bathrooms     float32   `json:"bathrooms"`
	SquareFeet    int       `json:"square_feet"`
	MarketRent    float64   `json:"market_rent"`
	AvailableDate time.Time `json:"available_date"`
	Status        string    `json:"status"`
}

type AppFolioVacanciesResponse struct {
	Vacancies  []AppFolioVacancy `json:"vacancies"`
	TotalCount int               `json:"total_count"`
}

type AppFolioError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewAppFolioAPIClient(baseURL, apiKey string) *AppFolioAPIClient {
	return &AppFolioAPIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AppFolioAPIClient) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr AppFolioError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("AppFolio API error (%d): %s - %s", resp.StatusCode, apiErr.Code, apiErr.Message)
		}
		return nil, fmt.Errorf("AppFolio API error: status %d", resp.StatusCode)
	}

	return body, nil
}

func (c *AppFolioAPIClient) GetProperties(page, pageSize int) (*AppFolioPropertiesResponse, error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))

	body, err := c.doRequest("GET", "/api/v1/properties", params)
	if err != nil {
		return nil, err
	}

	var response AppFolioPropertiesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *AppFolioAPIClient) GetAllProperties() ([]AppFolioProperty, error) {
	var allProperties []AppFolioProperty
	page := 1
	pageSize := 100

	for {
		response, err := c.GetProperties(page, pageSize)
		if err != nil {
			return nil, err
		}

		allProperties = append(allProperties, response.Properties...)

		if !response.HasMore || len(response.Properties) < pageSize {
			break
		}

		page++
	}

	return allProperties, nil
}

func (c *AppFolioAPIClient) GetProperty(propertyID string) (*AppFolioProperty, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/api/v1/properties/%s", propertyID), nil)
	if err != nil {
		return nil, err
	}

	var property AppFolioProperty
	if err := json.Unmarshal(body, &property); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &property, nil
}

func (c *AppFolioAPIClient) GetVacancies() (*AppFolioVacanciesResponse, error) {
	body, err := c.doRequest("GET", "/api/v1/vacancies", nil)
	if err != nil {
		return nil, err
	}

	var response AppFolioVacanciesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *AppFolioAPIClient) GetVacantProperties() ([]AppFolioProperty, error) {
	params := url.Values{}
	params.Set("status", "vacant")

	body, err := c.doRequest("GET", "/api/v1/properties", params)
	if err != nil {
		return nil, err
	}

	var response AppFolioPropertiesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Properties, nil
}

func (c *AppFolioAPIClient) SearchProperties(query string, filters map[string]string) ([]AppFolioProperty, error) {
	params := url.Values{}
	if query != "" {
		params.Set("q", query)
	}
	for key, value := range filters {
		params.Set(key, value)
	}

	body, err := c.doRequest("GET", "/api/v1/properties/search", params)
	if err != nil {
		return nil, err
	}

	var response AppFolioPropertiesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Properties, nil
}

func (c *AppFolioAPIClient) TestConnection() error {
	_, err := c.doRequest("GET", "/api/v1/ping", nil)
	return err
}
