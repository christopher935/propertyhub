package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAppFolioAPIClient(t *testing.T) {
	config := AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.test.com",
		Timeout: 15 * time.Second,
	}

	client := NewAppFolioAPIClient(nil, config)

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey 'test-api-key', got '%s'", client.apiKey)
	}
	if client.baseURL != "https://api.test.com" {
		t.Errorf("Expected baseURL 'https://api.test.com', got '%s'", client.baseURL)
	}
	if client.client.Timeout != 15*time.Second {
		t.Errorf("Expected timeout 15s, got %v", client.client.Timeout)
	}
}

func TestNewAppFolioAPIClientDefaults(t *testing.T) {
	config := AppFolioClientConfig{
		APIKey: "test-api-key",
	}

	client := NewAppFolioAPIClient(nil, config)

	if client.baseURL != "https://api.appfolio.com/v1" {
		t.Errorf("Expected default baseURL 'https://api.appfolio.com/v1', got '%s'", client.baseURL)
	}
	if client.client.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.client.Timeout)
	}
}

func TestNewAppFolioAPIClientSimple(t *testing.T) {
	client := NewAppFolioAPIClientSimple(nil, "simple-api-key")

	if client.apiKey != "simple-api-key" {
		t.Errorf("Expected apiKey 'simple-api-key', got '%s'", client.apiKey)
	}
	if client.baseURL != "https://api.appfolio.com/v1" {
		t.Errorf("Expected default baseURL, got '%s'", client.baseURL)
	}
}

func TestAppFolioAPIClient_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		config   AppFolioClientConfig
		expected bool
	}{
		{
			name:     "configured with API key",
			config:   AppFolioClientConfig{APIKey: "test-key"},
			expected: true,
		},
		{
			name:     "configured with client credentials",
			config:   AppFolioClientConfig{ClientID: "client-id", ClientSecret: "client-secret"},
			expected: true,
		},
		{
			name:     "not configured",
			config:   AppFolioClientConfig{},
			expected: false,
		},
		{
			name:     "partial client credentials",
			config:   AppFolioClientConfig{ClientID: "client-id"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAppFolioAPIClient(nil, tt.config)
			if got := client.IsConfigured(); got != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppFolioListParams_ToQueryParams(t *testing.T) {
	params := AppFolioListParams{
		Page:     2,
		PageSize: 50,
		Status:   "active",
		Search:   "test search",
		SortBy:   "created_at",
		SortDir:  "desc",
	}

	queryParams := params.ToQueryParams()

	if queryParams["page"] != "2" {
		t.Errorf("Expected page '2', got '%s'", queryParams["page"])
	}
	if queryParams["page_size"] != "50" {
		t.Errorf("Expected page_size '50', got '%s'", queryParams["page_size"])
	}
	if queryParams["status"] != "active" {
		t.Errorf("Expected status 'active', got '%s'", queryParams["status"])
	}
	if queryParams["search"] != "test search" {
		t.Errorf("Expected search 'test search', got '%s'", queryParams["search"])
	}
	if queryParams["sort_by"] != "created_at" {
		t.Errorf("Expected sort_by 'created_at', got '%s'", queryParams["sort_by"])
	}
	if queryParams["sort_dir"] != "desc" {
		t.Errorf("Expected sort_dir 'desc', got '%s'", queryParams["sort_dir"])
	}
}

func TestAppFolioListParams_ToQueryParams_Empty(t *testing.T) {
	params := AppFolioListParams{}

	queryParams := params.ToQueryParams()

	if len(queryParams) != 0 {
		t.Errorf("Expected empty query params, got %d params", len(queryParams))
	}
}

func TestAppFolioAPIClient_ListProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/properties" {
			t.Errorf("Expected path '/properties', got '%s'", r.URL.Path)
		}
		if r.Header.Get("X-AppFolio-API-Key") != "test-api-key" {
			t.Errorf("Expected API key header")
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":            "prop-1",
					"name":          "Test Property",
					"property_type": "residential",
					"status":        "active",
					"rent_amount":   1500.00,
				},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	properties, total, err := client.ListProperties(AppFolioListParams{})
	if err != nil {
		t.Fatalf("ListProperties failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if len(properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(properties))
	}
	if properties[0].ID != "prop-1" {
		t.Errorf("Expected property ID 'prop-1', got '%s'", properties[0].ID)
	}
}

func TestAppFolioAPIClient_GetProperty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/properties/prop-123" {
			t.Errorf("Expected path '/properties/prop-123', got '%s'", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":            "prop-123",
				"name":          "Test Property",
				"property_type": "residential",
				"status":        "active",
				"rent_amount":   2000.00,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	property, err := client.GetProperty("prop-123")
	if err != nil {
		t.Fatalf("GetProperty failed: %v", err)
	}

	if property.ID != "prop-123" {
		t.Errorf("Expected property ID 'prop-123', got '%s'", property.ID)
	}
	if property.RentAmount != 2000.00 {
		t.Errorf("Expected rent amount 2000.00, got %f", property.RentAmount)
	}
}

func TestAppFolioAPIClient_CreateProperty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/properties" {
			t.Errorf("Expected path '/properties', got '%s'", r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":            "new-prop-1",
				"name":          body["name"],
				"property_type": body["property_type"],
				"status":        "active",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	propertyData := map[string]interface{}{
		"name":          "New Property",
		"property_type": "residential",
		"address": map[string]interface{}{
			"street_1":    "123 Main St",
			"city":        "Houston",
			"state":       "TX",
			"postal_code": "77001",
		},
		"rent_amount": 1800.00,
	}

	property, err := client.CreateProperty(propertyData)
	if err != nil {
		t.Fatalf("CreateProperty failed: %v", err)
	}

	if property.ID != "new-prop-1" {
		t.Errorf("Expected property ID 'new-prop-1', got '%s'", property.ID)
	}
}

func TestAppFolioAPIClient_CreateProperty_ValidationError(t *testing.T) {
	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey: "test-api-key",
	})

	propertyData := map[string]interface{}{
		"name": "Property without address",
	}

	_, err := client.CreateProperty(propertyData)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}
}

func TestAppFolioAPIClient_ListTenants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":         "tenant-1",
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john@example.com",
				},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	tenants, total, err := client.ListTenants(AppFolioListParams{})
	if err != nil {
		t.Fatalf("ListTenants failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if len(tenants) != 1 {
		t.Errorf("Expected 1 tenant, got %d", len(tenants))
	}
	if tenants[0].FirstName != "John" {
		t.Errorf("Expected first name 'John', got '%s'", tenants[0].FirstName)
	}
}

func TestAppFolioAPIClient_ListMaintenanceRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":          "maint-1",
					"property_id": "prop-1",
					"description": "Fix plumbing",
					"priority":    "high",
					"status":      "open",
				},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	requests, total, err := client.ListMaintenanceRequests(AppFolioListParams{})
	if err != nil {
		t.Fatalf("ListMaintenanceRequests failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if requests[0].Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", requests[0].Priority)
	}
}

func TestAppFolioAPIClient_ListPayments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":        "pay-1",
					"tenant_id": "tenant-1",
					"amount":    1500.00,
					"status":    "paid",
				},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	payments, total, err := client.ListPayments(AppFolioListParams{})
	if err != nil {
		t.Fatalf("ListPayments failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if payments[0].Amount != 1500.00 {
		t.Errorf("Expected amount 1500.00, got %f", payments[0].Amount)
	}
}

func TestAppFolioAPIClient_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "unauthorized",
			"message": "Invalid API key",
		})
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
	})

	_, _, err := client.ListProperties(AppFolioListParams{})
	if err == nil {
		t.Error("Expected error for unauthorized request, got nil")
	}
}

func TestAppFolioAPIClient_RateLimitRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate_limited",
			})
			return
		}

		response := map[string]interface{}{
			"data":        []map[string]interface{}{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	_, _, err := client.ListProperties(AppFolioListParams{})
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls (1 retry), got %d", callCount)
	}
}

func TestAppFolioAPIClient_SetTimeout(t *testing.T) {
	client := NewAppFolioAPIClientSimple(nil, "test-key")

	client.SetTimeout(60 * time.Second)

	if client.client.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", client.client.Timeout)
	}
}

func TestAppFolioAPIClient_GetBaseURL(t *testing.T) {
	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-key",
		BaseURL: "https://custom.api.com",
	})

	if client.GetBaseURL() != "https://custom.api.com" {
		t.Errorf("Expected base URL 'https://custom.api.com', got '%s'", client.GetBaseURL())
	}
}

func TestAppFolioAPIClient_HealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("Expected path '/health', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	err := client.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}
}

func TestAppFolioAPIClient_UpdateProperty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/properties/prop-123" {
			t.Errorf("Expected path '/properties/prop-123', got '%s'", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "prop-123",
				"name":        "Updated Property",
				"rent_amount": 2500.00,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	updates := map[string]interface{}{
		"name":        "Updated Property",
		"rent_amount": 2500.00,
	}

	property, err := client.UpdateProperty("prop-123", updates)
	if err != nil {
		t.Fatalf("UpdateProperty failed: %v", err)
	}

	if property.RentAmount != 2500.00 {
		t.Errorf("Expected rent amount 2500.00, got %f", property.RentAmount)
	}
}

func TestAppFolioAPIClient_DeleteProperty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/properties/prop-123" {
			t.Errorf("Expected path '/properties/prop-123', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	err := client.DeleteProperty("prop-123")
	if err != nil {
		t.Errorf("DeleteProperty failed: %v", err)
	}
}

func TestAppFolioAPIClient_GetTenantsByProperty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		propertyID := r.URL.Query().Get("property_id")
		if propertyID != "prop-123" {
			t.Errorf("Expected property_id 'prop-123', got '%s'", propertyID)
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":          "tenant-1",
					"property_id": "prop-123",
					"first_name":  "Jane",
				},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewAppFolioAPIClient(nil, AppFolioClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	tenants, total, err := client.GetTenantsByProperty("prop-123", AppFolioListParams{})
	if err != nil {
		t.Fatalf("GetTenantsByProperty failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if tenants[0].PropertyID != "prop-123" {
		t.Errorf("Expected property_id 'prop-123', got '%s'", tenants[0].PropertyID)
	}
}
