package services

import (
	"chrisgross-ctrl-project/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// FUBAPIError represents a structured error from FUB API operations
type FUBAPIError struct {
	Operation  string                 `json:"operation"`
	ErrorCode  string                 `json:"error_code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"http_status"`
	RequestID  string                 `json:"request_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Retryable  bool                   `json:"retryable"`
	RetryAfter *time.Duration         `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e *FUBAPIError) Error() string {
	return fmt.Sprintf("FUB API Error [%s]: %s (HTTP %d)", e.Operation, e.Message, e.HTTPStatus)
}

// IsRetryable indicates whether the error condition is retryable
func (e *FUBAPIError) IsRetryable() bool {
	return e.Retryable
}

// FUBErrorHandler provides centralized error handling for FUB API operations
type FUBErrorHandler struct {
	maxRetries    int
	baseDelay     time.Duration
	maxDelay      time.Duration
	backoffFactor float64
}

// NewFUBErrorHandler creates a new error handler with default settings
func NewFUBErrorHandler() *FUBErrorHandler {
	return &FUBErrorHandler{
		maxRetries:    3,
		baseDelay:     1 * time.Second,
		maxDelay:      30 * time.Second,
		backoffFactor: 2.0,
	}
}

// HandleHTTPResponse processes HTTP responses and creates appropriate errors
func (eh *FUBErrorHandler) HandleHTTPResponse(resp *http.Response, operation string) *FUBAPIError {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil // Success
	}

	// Parse error response body
	var errorDetails map[string]interface{}
	if resp.Body != nil {
		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&errorDetails)
	}

	// Create structured error
	fubError := &FUBAPIError{
		Operation:  operation,
		HTTPStatus: resp.StatusCode,
		Timestamp:  time.Now(),
		Details:    errorDetails,
	}

	// Map HTTP status codes to error details
	switch resp.StatusCode {
	case 400:
		fubError.ErrorCode = "BAD_REQUEST"
		fubError.Message = "Invalid request data or parameters"
		fubError.Retryable = false
	case 401:
		fubError.ErrorCode = "UNAUTHORIZED"
		fubError.Message = "API authentication failed - check API key"
		fubError.Retryable = false
	case 403:
		fubError.ErrorCode = "FORBIDDEN"
		fubError.Message = "API access denied - insufficient permissions"
		fubError.Retryable = false
	case 404:
		fubError.ErrorCode = "NOT_FOUND"
		fubError.Message = "Requested resource not found"
		fubError.Retryable = false
	case 422:
		fubError.ErrorCode = "UNPROCESSABLE_ENTITY"
		fubError.Message = "Request data validation failed"
		fubError.Retryable = false
	case 429:
		fubError.ErrorCode = "RATE_LIMITED"
		fubError.Message = "API rate limit exceeded"
		fubError.Retryable = true
		// Extract retry-after header if present
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if duration, err := time.ParseDuration(retryAfter + "s"); err == nil {
				fubError.RetryAfter = &duration
			}
		}
	case 500:
		fubError.ErrorCode = "INTERNAL_SERVER_ERROR"
		fubError.Message = "FUB API internal server error"
		fubError.Retryable = true
	case 502, 503, 504:
		fubError.ErrorCode = "SERVICE_UNAVAILABLE"
		fubError.Message = "FUB API service temporarily unavailable"
		fubError.Retryable = true
	default:
		fubError.ErrorCode = "UNKNOWN_ERROR"
		fubError.Message = fmt.Sprintf("Unexpected HTTP status: %d", resp.StatusCode)
		fubError.Retryable = resp.StatusCode >= 500
	}

	// Extract detailed error message from response if available
	if errorDetails != nil {
		if msg, exists := errorDetails["message"]; exists {
			if msgStr, ok := msg.(string); ok {
				fubError.Message = msgStr
			}
		}
		if code, exists := errorDetails["error"]; exists {
			if codeStr, ok := code.(string); ok {
				fubError.ErrorCode = codeStr
			}
		}
	}

	return fubError
}

// ExecuteWithRetry executes a function with retry logic for retryable errors
func (eh *FUBErrorHandler) ExecuteWithRetry(operation string, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastError error

	for attempt := 0; attempt <= eh.maxRetries; attempt++ {
		log.Printf("🔄 Executing %s (attempt %d/%d)", operation, attempt+1, eh.maxRetries+1)

		resp, err := fn()
		if err != nil {
			lastError = err
			if attempt < eh.maxRetries {
				delay := eh.calculateBackoffDelay(attempt)
				log.Printf("⏰ Retrying %s in %v due to error: %v", operation, delay, err)
				time.Sleep(delay)
			}
			continue
		}

		// Check if response indicates an error
		if fubError := eh.HandleHTTPResponse(resp, operation); fubError != nil {
			lastError = fubError
			if fubError.IsRetryable() && attempt < eh.maxRetries {
				resp.Body.Close() // Clean up response body

				// Use custom retry delay if provided
				delay := eh.calculateBackoffDelay(attempt)
				if fubError.RetryAfter != nil && *fubError.RetryAfter > 0 {
					delay = *fubError.RetryAfter
				}

				log.Printf("⏰ Retrying %s in %v due to retryable error: %v", operation, delay, fubError.Message)
				time.Sleep(delay)
				continue
			}
			return resp, fubError
		}

		// Success
		log.Printf("✅ Successfully executed %s on attempt %d", operation, attempt+1)
		return resp, nil
	}

	log.Printf("❌ Failed to execute %s after %d attempts. Last error: %v", operation, eh.maxRetries+1, lastError)
	return nil, lastError
}

// calculateBackoffDelay calculates the delay for exponential backoff
func (eh *FUBErrorHandler) calculateBackoffDelay(attempt int) time.Duration {
	delay := float64(eh.baseDelay) * (eh.backoffFactor * float64(attempt))
	if time.Duration(delay) > eh.maxDelay {
		delay = float64(eh.maxDelay)
	}
	return time.Duration(delay)
}

// BehavioralTriggerValidationError represents validation errors for behavioral triggers
type BehavioralTriggerValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (e *BehavioralTriggerValidationError) Error() string {
	return fmt.Sprintf("Behavioral trigger validation error [%s]: %s", e.Field, e.Message)
}

// ValidateBehavioralTriggerRequest validates behavioral trigger request data
func ValidateBehavioralTriggerRequest(request *BehavioralTriggerRequest) []BehavioralTriggerValidationError {
	var errors []BehavioralTriggerValidationError

	// Required fields
	if request.SessionID == "" {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "session_id",
			Message: "Session ID is required",
		})
	}

	if request.TriggerType == "" {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "trigger_type",
			Message: "Trigger type is required",
		})
	}

	if request.Name == "" && request.Email == "" && request.Phone == "" {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "contact_info",
			Message: "At least one contact method (name, email, or phone) is required",
		})
	}

	// Validate email format if provided
	if request.Email != "" && !isValidEmail(request.Email) {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "email",
			Message: "Invalid email format",
			Value:   request.Email,
		})
	}

	// Validate behavioral scores (should be between 0 and 100)
	if request.UrgencyScore < 0 || request.UrgencyScore > 100 {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "urgency_score",
			Message: "Urgency score must be between 0 and 100",
			Value:   request.UrgencyScore,
		})
	}

	if request.FinancialQualScore < 0 || request.FinancialQualScore > 100 {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "financial_qualification_score",
			Message: "Financial qualification score must be between 0 and 100",
			Value:   request.FinancialQualScore,
		})
	}

	if request.EngagementScore < 0 || request.EngagementScore > 100 {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "engagement_score",
			Message: "Engagement score must be between 0 and 100",
			Value:   request.EngagementScore,
		})
	}

	// Validate lead type if provided
	validLeadTypes := []string{"tenant", "landlord", "buyer", "seller", "investor"}
	if request.LeadType != "" && !utils.Contains(validLeadTypes, request.LeadType) {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "lead_type",
			Message: fmt.Sprintf("Invalid lead type. Must be one of: %v", validLeadTypes),
			Value:   request.LeadType,
		})
	}

	// Validate property type if provided
	validPropertyTypes := []string{"rental", "sales", "mixed", "unknown"}
	if request.PropertyType != "" && !utils.Contains(validPropertyTypes, request.PropertyType) {
		errors = append(errors, BehavioralTriggerValidationError{
			Field:   "property_type",
			Message: fmt.Sprintf("Invalid property type. Must be one of: %v", validPropertyTypes),
			Value:   request.PropertyType,
		})
	}

	return errors
}

// FUBOperationResult represents the result of a FUB operation with comprehensive error handling
type FUBOperationResult struct {
	Success          bool                               `json:"success"`
	Operation        string                             `json:"operation"`
	Data             map[string]interface{}             `json:"data,omitempty"`
	Error            *FUBAPIError                       `json:"error,omitempty"`
	ValidationErrors []BehavioralTriggerValidationError `json:"validation_errors,omitempty"`
	ExecutionTime    time.Duration                      `json:"execution_time"`
	AttemptCount     int                                `json:"attempt_count"`
	Timestamp        time.Time                          `json:"timestamp"`
}

// CreateSuccessResult creates a success result
func CreateSuccessResult(operation string, data map[string]interface{}, executionTime time.Duration, attemptCount int) *FUBOperationResult {
	return &FUBOperationResult{
		Success:       true,
		Operation:     operation,
		Data:          data,
		ExecutionTime: executionTime,
		AttemptCount:  attemptCount,
		Timestamp:     time.Now(),
	}
}

// CreateErrorResult creates an error result
func CreateErrorResult(operation string, err *FUBAPIError, executionTime time.Duration, attemptCount int) *FUBOperationResult {
	return &FUBOperationResult{
		Success:       false,
		Operation:     operation,
		Error:         err,
		ExecutionTime: executionTime,
		AttemptCount:  attemptCount,
		Timestamp:     time.Now(),
	}
}

// CreateValidationErrorResult creates a validation error result
func CreateValidationErrorResult(operation string, validationErrors []BehavioralTriggerValidationError) *FUBOperationResult {
	return &FUBOperationResult{
		Success:          false,
		Operation:        operation,
		ValidationErrors: validationErrors,
		ExecutionTime:    0,
		AttemptCount:     1,
		Timestamp:        time.Now(),
	}
}

// Helper functions

func isValidEmail(email string) bool {
	// Simple email validation - in production you might want more sophisticated validation
	return len(email) > 3 &&
		len(email) <= 254 &&
		fmt.Sprintf("%s", email) != "" // Basic non-empty check
}

// LogFUBOperation logs FUB operations for monitoring and debugging
func LogFUBOperation(result *FUBOperationResult) {
	if result.Success {
		log.Printf("✅ FUB Operation Success: %s completed in %v (attempts: %d)",
			result.Operation, result.ExecutionTime, result.AttemptCount)
	} else {
		if result.Error != nil {
			log.Printf("❌ FUB Operation Error: %s failed after %v (attempts: %d) - %s",
				result.Operation, result.ExecutionTime, result.AttemptCount, result.Error.Message)
		}
		if len(result.ValidationErrors) > 0 {
			log.Printf("⚠️ FUB Operation Validation Errors: %s - %d validation issues",
				result.Operation, len(result.ValidationErrors))
			for _, validationError := range result.ValidationErrors {
				log.Printf("   - %s: %s", validationError.Field, validationError.Message)
			}
		}
	}
}
