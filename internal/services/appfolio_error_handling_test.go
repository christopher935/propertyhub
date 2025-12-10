package services

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestNewAppFolioErrorHandler(t *testing.T) {
	handler := NewAppFolioErrorHandler()

	if handler.maxRetries != 3 {
		t.Errorf("Expected maxRetries 3, got %d", handler.maxRetries)
	}
	if handler.baseDelay != 1*time.Second {
		t.Errorf("Expected baseDelay 1s, got %v", handler.baseDelay)
	}
	if handler.maxDelay != 60*time.Second {
		t.Errorf("Expected maxDelay 60s, got %v", handler.maxDelay)
	}
	if handler.backoffFactor != 2.0 {
		t.Errorf("Expected backoffFactor 2.0, got %f", handler.backoffFactor)
	}
}

func TestNewAppFolioErrorHandlerWithConfig(t *testing.T) {
	handler := NewAppFolioErrorHandlerWithConfig(5, 2*time.Second, 120*time.Second, 3.0)

	if handler.maxRetries != 5 {
		t.Errorf("Expected maxRetries 5, got %d", handler.maxRetries)
	}
	if handler.baseDelay != 2*time.Second {
		t.Errorf("Expected baseDelay 2s, got %v", handler.baseDelay)
	}
	if handler.maxDelay != 120*time.Second {
		t.Errorf("Expected maxDelay 120s, got %v", handler.maxDelay)
	}
	if handler.backoffFactor != 3.0 {
		t.Errorf("Expected backoffFactor 3.0, got %f", handler.backoffFactor)
	}
}

func TestAppFolioAPIError_Error(t *testing.T) {
	err := &AppFolioAPIError{
		Operation:  "GET /properties",
		ErrorCode:  "NOT_FOUND",
		Message:    "Property not found",
		HTTPStatus: 404,
	}

	expected := "AppFolio API Error [GET /properties]: Property not found (HTTP 404)"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestAppFolioAPIError_ErrorMethods(t *testing.T) {
	tests := []struct {
		name          string
		err           *AppFolioAPIError
		isRetryable   bool
		isRateLimited bool
		isAuthError   bool
		isNotFound    bool
		isValidation  bool
	}{
		{
			name: "rate limited error",
			err: &AppFolioAPIError{
				HTTPStatus: 429,
				Retryable:  true,
			},
			isRetryable:   true,
			isRateLimited: true,
		},
		{
			name: "auth error 401",
			err: &AppFolioAPIError{
				HTTPStatus: 401,
				Retryable:  false,
			},
			isAuthError: true,
		},
		{
			name: "auth error 403",
			err: &AppFolioAPIError{
				HTTPStatus: 403,
				Retryable:  false,
			},
			isAuthError: true,
		},
		{
			name: "not found error",
			err: &AppFolioAPIError{
				HTTPStatus: 404,
				Retryable:  false,
			},
			isNotFound: true,
		},
		{
			name: "validation error 400",
			err: &AppFolioAPIError{
				HTTPStatus: 400,
				Retryable:  false,
			},
			isValidation: true,
		},
		{
			name: "validation error 422",
			err: &AppFolioAPIError{
				HTTPStatus: 422,
				Retryable:  false,
			},
			isValidation: true,
		},
		{
			name: "server error",
			err: &AppFolioAPIError{
				HTTPStatus: 500,
				Retryable:  true,
			},
			isRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsRetryable(); got != tt.isRetryable {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.isRetryable)
			}
			if got := tt.err.IsRateLimited(); got != tt.isRateLimited {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.isRateLimited)
			}
			if got := tt.err.IsAuthError(); got != tt.isAuthError {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.isAuthError)
			}
			if got := tt.err.IsNotFound(); got != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.isNotFound)
			}
			if got := tt.err.IsValidationError(); got != tt.isValidation {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.isValidation)
			}
		})
	}
}

func TestAppFolioErrorHandler_HandleHTTPResponse(t *testing.T) {
	handler := NewAppFolioErrorHandler()

	tests := []struct {
		name       string
		statusCode int
		body       string
		errorCode  string
		retryable  bool
	}{
		{"success 200", 200, "", "", false},
		{"success 201", 201, "", "", false},
		{"bad request", 400, `{"message":"Invalid input"}`, "BAD_REQUEST", false},
		{"unauthorized", 401, "", "UNAUTHORIZED", false},
		{"forbidden", 403, "", "FORBIDDEN", false},
		{"not found", 404, "", "NOT_FOUND", false},
		{"conflict", 409, "", "CONFLICT", false},
		{"unprocessable", 422, "", "UNPROCESSABLE_ENTITY", false},
		{"rate limited", 429, "", "RATE_LIMITED", true},
		{"server error", 500, "", "INTERNAL_SERVER_ERROR", true},
		{"bad gateway", 502, "", "BAD_GATEWAY", true},
		{"service unavailable", 503, "", "SERVICE_UNAVAILABLE", true},
		{"gateway timeout", 504, "", "GATEWAY_TIMEOUT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.body)),
				Header:     make(http.Header),
			}

			err := handler.HandleHTTPResponse(resp, "test operation")

			if tt.statusCode >= 200 && tt.statusCode < 300 {
				if err != nil {
					t.Errorf("Expected nil error for status %d, got %v", tt.statusCode, err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error for status %d, got nil", tt.statusCode)
				return
			}

			if err.ErrorCode != tt.errorCode {
				t.Errorf("Expected error code '%s', got '%s'", tt.errorCode, err.ErrorCode)
			}
			if err.Retryable != tt.retryable {
				t.Errorf("Expected retryable=%v, got %v", tt.retryable, err.Retryable)
			}
		})
	}
}

func TestAppFolioErrorHandler_HandleHTTPResponse_WithRequestID(t *testing.T) {
	handler := NewAppFolioErrorHandler()

	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Request-ID", "req-123456")

	err := handler.HandleHTTPResponse(resp, "test operation")

	if err.RequestID != "req-123456" {
		t.Errorf("Expected request ID 'req-123456', got '%s'", err.RequestID)
	}
}

func TestAppFolioErrorHandler_ParseRetryAfterHeader(t *testing.T) {
	handler := NewAppFolioErrorHandler()

	tests := []struct {
		name         string
		retryAfter   string
		expectNil    bool
		expectedSecs int
	}{
		{"no header", "", true, 0},
		{"seconds value", "30", false, 30},
		{"invalid value", "invalid", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: make(http.Header),
			}
			if tt.retryAfter != "" {
				resp.Header.Set("Retry-After", tt.retryAfter)
			}

			duration := handler.parseRetryAfterHeader(resp)

			if tt.expectNil {
				if duration != nil {
					t.Errorf("Expected nil duration, got %v", *duration)
				}
			} else {
				if duration == nil {
					t.Error("Expected non-nil duration, got nil")
				} else if int(duration.Seconds()) != tt.expectedSecs {
					t.Errorf("Expected %d seconds, got %v", tt.expectedSecs, *duration)
				}
			}
		})
	}
}

func TestAppFolioErrorHandler_CalculateBackoffDelay(t *testing.T) {
	handler := NewAppFolioErrorHandler()

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
	}

	for _, tt := range tests {
		delay := handler.calculateBackoffDelay(tt.attempt)
		if delay != tt.expected {
			t.Errorf("For attempt %d, expected delay %v, got %v", tt.attempt, tt.expected, delay)
		}
	}
}

func TestAppFolioErrorHandler_CalculateBackoffDelay_MaxCap(t *testing.T) {
	handler := NewAppFolioErrorHandlerWithConfig(10, 1*time.Second, 30*time.Second, 2.0)

	delay := handler.calculateBackoffDelay(10)

	if delay > 30*time.Second {
		t.Errorf("Expected delay capped at 30s, got %v", delay)
	}
}

func TestAppFolioValidationError_Error(t *testing.T) {
	err := &AppFolioValidationError{
		Field:   "email",
		Message: "Invalid email format",
		Value:   "not-an-email",
	}

	expected := "AppFolio validation error [email]: Invalid email format"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestCreateAppFolioSuccessResult(t *testing.T) {
	data := map[string]interface{}{"id": "123"}
	result := CreateAppFolioSuccessResult("CreateProperty", "property", "prop-123", data, 100*time.Millisecond, 1)

	if !result.Success {
		t.Error("Expected success=true")
	}
	if result.Operation != "CreateProperty" {
		t.Errorf("Expected operation 'CreateProperty', got '%s'", result.Operation)
	}
	if result.ResourceType != "property" {
		t.Errorf("Expected resource type 'property', got '%s'", result.ResourceType)
	}
	if result.ResourceID != "prop-123" {
		t.Errorf("Expected resource ID 'prop-123', got '%s'", result.ResourceID)
	}
	if result.AttemptCount != 1 {
		t.Errorf("Expected attempt count 1, got %d", result.AttemptCount)
	}
}

func TestCreateAppFolioErrorResult(t *testing.T) {
	apiErr := &AppFolioAPIError{
		Operation:  "GetProperty",
		ErrorCode:  "NOT_FOUND",
		Message:    "Property not found",
		HTTPStatus: 404,
	}

	result := CreateAppFolioErrorResult("GetProperty", apiErr, 50*time.Millisecond, 1)

	if result.Success {
		t.Error("Expected success=false")
	}
	if result.Error != apiErr {
		t.Error("Expected error to be set")
	}
}

func TestCreateAppFolioValidationErrorResult(t *testing.T) {
	validationErrors := []AppFolioValidationError{
		{Field: "email", Message: "Invalid format"},
		{Field: "phone", Message: "Required"},
	}

	result := CreateAppFolioValidationErrorResult("CreateTenant", validationErrors)

	if result.Success {
		t.Error("Expected success=false")
	}
	if len(result.ValidationErrors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(result.ValidationErrors))
	}
}

func TestValidateAppFolioPropertyRequest(t *testing.T) {
	tests := []struct {
		name         string
		property     map[string]interface{}
		expectedErrs int
	}{
		{
			name: "valid property",
			property: map[string]interface{}{
				"address": map[string]interface{}{
					"street_1":    "123 Main St",
					"city":        "Houston",
					"state":       "TX",
					"postal_code": "77001",
				},
				"rent_amount": 1500.00,
				"unit_count":  1,
			},
			expectedErrs: 0,
		},
		{
			name:         "missing address",
			property:     map[string]interface{}{},
			expectedErrs: 1,
		},
		{
			name: "missing address fields",
			property: map[string]interface{}{
				"address": map[string]interface{}{},
			},
			expectedErrs: 4,
		},
		{
			name: "negative rent",
			property: map[string]interface{}{
				"address": map[string]interface{}{
					"street_1":    "123 Main St",
					"city":        "Houston",
					"state":       "TX",
					"postal_code": "77001",
				},
				"rent_amount": -100.00,
			},
			expectedErrs: 1,
		},
		{
			name: "invalid unit count",
			property: map[string]interface{}{
				"address": map[string]interface{}{
					"street_1":    "123 Main St",
					"city":        "Houston",
					"state":       "TX",
					"postal_code": "77001",
				},
				"unit_count": 0,
			},
			expectedErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAppFolioPropertyRequest(tt.property)
			if len(errors) != tt.expectedErrs {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrs, len(errors), errors)
			}
		})
	}
}

func TestValidateAppFolioTenantRequest(t *testing.T) {
	tests := []struct {
		name         string
		tenant       map[string]interface{}
		expectedErrs int
	}{
		{
			name: "valid tenant",
			tenant: map[string]interface{}{
				"first_name": "John",
				"last_name":  "Doe",
				"email":      "john@example.com",
			},
			expectedErrs: 0,
		},
		{
			name: "valid tenant with phone only",
			tenant: map[string]interface{}{
				"first_name": "John",
				"phone":      "555-1234",
			},
			expectedErrs: 0,
		},
		{
			name:         "missing name",
			tenant:       map[string]interface{}{"email": "test@example.com"},
			expectedErrs: 1,
		},
		{
			name:         "missing contact",
			tenant:       map[string]interface{}{"first_name": "John"},
			expectedErrs: 1,
		},
		{
			name: "invalid email",
			tenant: map[string]interface{}{
				"first_name": "John",
				"email":      "not-an-email",
			},
			expectedErrs: 1,
		},
		{
			name: "negative rent",
			tenant: map[string]interface{}{
				"first_name":  "John",
				"email":       "john@example.com",
				"rent_amount": -500.00,
			},
			expectedErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAppFolioTenantRequest(tt.tenant)
			if len(errors) != tt.expectedErrs {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrs, len(errors), errors)
			}
		})
	}
}

func TestValidateAppFolioMaintenanceRequest(t *testing.T) {
	tests := []struct {
		name         string
		request      map[string]interface{}
		expectedErrs int
	}{
		{
			name: "valid request",
			request: map[string]interface{}{
				"property_id": "prop-123",
				"description": "Fix the leak",
				"priority":    "high",
			},
			expectedErrs: 0,
		},
		{
			name: "missing property_id",
			request: map[string]interface{}{
				"description": "Fix the leak",
			},
			expectedErrs: 1,
		},
		{
			name: "missing description",
			request: map[string]interface{}{
				"property_id": "prop-123",
			},
			expectedErrs: 1,
		},
		{
			name: "invalid priority",
			request: map[string]interface{}{
				"property_id": "prop-123",
				"description": "Fix the leak",
				"priority":    "super-urgent",
			},
			expectedErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAppFolioMaintenanceRequest(tt.request)
			if len(errors) != tt.expectedErrs {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrs, len(errors), errors)
			}
		})
	}
}

func TestValidateAppFolioPaymentRequest(t *testing.T) {
	tests := []struct {
		name         string
		payment      map[string]interface{}
		expectedErrs int
	}{
		{
			name: "valid payment",
			payment: map[string]interface{}{
				"tenant_id": "tenant-123",
				"amount":    1500.00,
			},
			expectedErrs: 0,
		},
		{
			name: "missing tenant_id",
			payment: map[string]interface{}{
				"amount": 1500.00,
			},
			expectedErrs: 1,
		},
		{
			name: "missing amount",
			payment: map[string]interface{}{
				"tenant_id": "tenant-123",
			},
			expectedErrs: 1,
		},
		{
			name: "zero amount",
			payment: map[string]interface{}{
				"tenant_id": "tenant-123",
				"amount":    0.0,
			},
			expectedErrs: 1,
		},
		{
			name: "negative amount",
			payment: map[string]interface{}{
				"tenant_id": "tenant-123",
				"amount":    -100.0,
			},
			expectedErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAppFolioPaymentRequest(tt.payment)
			if len(errors) != tt.expectedErrs {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrs, len(errors), errors)
			}
		})
	}
}

func TestIsValidEmailFormat(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{"a@b.co", true},
		{"", false},
		{"notanemail", false},
		{"@example.com", false},
		{"test@", false},
		{"test@domain", false},
		{"test@@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmailFormat(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmailFormat(%s) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}
