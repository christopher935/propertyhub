package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type AppFolioAPIError struct {
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

func (e *AppFolioAPIError) Error() string {
	return fmt.Sprintf("AppFolio API Error [%s]: %s (HTTP %d)", e.Operation, e.Message, e.HTTPStatus)
}

func (e *AppFolioAPIError) IsRetryable() bool {
	return e.Retryable
}

func (e *AppFolioAPIError) IsRateLimited() bool {
	return e.HTTPStatus == 429
}

func (e *AppFolioAPIError) IsAuthError() bool {
	return e.HTTPStatus == 401 || e.HTTPStatus == 403
}

func (e *AppFolioAPIError) IsNotFound() bool {
	return e.HTTPStatus == 404
}

func (e *AppFolioAPIError) IsValidationError() bool {
	return e.HTTPStatus == 400 || e.HTTPStatus == 422
}

type AppFolioErrorHandler struct {
	maxRetries    int
	baseDelay     time.Duration
	maxDelay      time.Duration
	backoffFactor float64
}

func NewAppFolioErrorHandler() *AppFolioErrorHandler {
	return &AppFolioErrorHandler{
		maxRetries:    3,
		baseDelay:     1 * time.Second,
		maxDelay:      60 * time.Second,
		backoffFactor: 2.0,
	}
}

func NewAppFolioErrorHandlerWithConfig(maxRetries int, baseDelay, maxDelay time.Duration, backoffFactor float64) *AppFolioErrorHandler {
	return &AppFolioErrorHandler{
		maxRetries:    maxRetries,
		baseDelay:     baseDelay,
		maxDelay:      maxDelay,
		backoffFactor: backoffFactor,
	}
}

func (eh *AppFolioErrorHandler) HandleHTTPResponse(resp *http.Response, operation string) *AppFolioAPIError {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	var errorDetails map[string]interface{}
	if resp.Body != nil {
		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&errorDetails)
	}

	appFolioError := &AppFolioAPIError{
		Operation:  operation,
		HTTPStatus: resp.StatusCode,
		Timestamp:  time.Now(),
		Details:    errorDetails,
	}

	if requestID := resp.Header.Get("X-Request-ID"); requestID != "" {
		appFolioError.RequestID = requestID
	}

	switch resp.StatusCode {
	case 400:
		appFolioError.ErrorCode = "BAD_REQUEST"
		appFolioError.Message = "Invalid request data or parameters"
		appFolioError.Retryable = false
	case 401:
		appFolioError.ErrorCode = "UNAUTHORIZED"
		appFolioError.Message = "AppFolio API authentication failed - check API key"
		appFolioError.Retryable = false
	case 403:
		appFolioError.ErrorCode = "FORBIDDEN"
		appFolioError.Message = "AppFolio API access denied - insufficient permissions"
		appFolioError.Retryable = false
	case 404:
		appFolioError.ErrorCode = "NOT_FOUND"
		appFolioError.Message = "Requested resource not found in AppFolio"
		appFolioError.Retryable = false
	case 409:
		appFolioError.ErrorCode = "CONFLICT"
		appFolioError.Message = "Resource conflict - the resource may already exist or has been modified"
		appFolioError.Retryable = false
	case 422:
		appFolioError.ErrorCode = "UNPROCESSABLE_ENTITY"
		appFolioError.Message = "Request data validation failed"
		appFolioError.Retryable = false
	case 429:
		appFolioError.ErrorCode = "RATE_LIMITED"
		appFolioError.Message = "AppFolio API rate limit exceeded"
		appFolioError.Retryable = true
		appFolioError.RetryAfter = eh.parseRetryAfterHeader(resp)
	case 500:
		appFolioError.ErrorCode = "INTERNAL_SERVER_ERROR"
		appFolioError.Message = "AppFolio API internal server error"
		appFolioError.Retryable = true
	case 502:
		appFolioError.ErrorCode = "BAD_GATEWAY"
		appFolioError.Message = "AppFolio API bad gateway error"
		appFolioError.Retryable = true
	case 503:
		appFolioError.ErrorCode = "SERVICE_UNAVAILABLE"
		appFolioError.Message = "AppFolio API service temporarily unavailable"
		appFolioError.Retryable = true
		appFolioError.RetryAfter = eh.parseRetryAfterHeader(resp)
	case 504:
		appFolioError.ErrorCode = "GATEWAY_TIMEOUT"
		appFolioError.Message = "AppFolio API gateway timeout"
		appFolioError.Retryable = true
	default:
		appFolioError.ErrorCode = "UNKNOWN_ERROR"
		appFolioError.Message = fmt.Sprintf("Unexpected HTTP status: %d", resp.StatusCode)
		appFolioError.Retryable = resp.StatusCode >= 500
	}

	if errorDetails != nil {
		if msg, exists := errorDetails["message"]; exists {
			if msgStr, ok := msg.(string); ok && msgStr != "" {
				appFolioError.Message = msgStr
			}
		}
		if code, exists := errorDetails["error"]; exists {
			if codeStr, ok := code.(string); ok && codeStr != "" {
				appFolioError.ErrorCode = codeStr
			}
		}
		if code, exists := errorDetails["code"]; exists {
			if codeStr, ok := code.(string); ok && codeStr != "" {
				appFolioError.ErrorCode = codeStr
			}
		}
	}

	return appFolioError
}

func (eh *AppFolioErrorHandler) parseRetryAfterHeader(resp *http.Response) *time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return nil
	}

	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		duration := time.Duration(seconds) * time.Second
		return &duration
	}

	if retryTime, err := http.ParseTime(retryAfter); err == nil {
		duration := time.Until(retryTime)
		if duration > 0 {
			return &duration
		}
	}

	return nil
}

func (eh *AppFolioErrorHandler) ExecuteWithRetry(operation string, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastError error

	for attempt := 0; attempt <= eh.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("🔄 AppFolio: Executing %s (attempt %d/%d)", operation, attempt+1, eh.maxRetries+1)
		}

		resp, err := fn()
		if err != nil {
			lastError = err
			if attempt < eh.maxRetries {
				delay := eh.calculateBackoffDelay(attempt)
				log.Printf("⏰ AppFolio: Retrying %s in %v due to error: %v", operation, delay, err)
				time.Sleep(delay)
			}
			continue
		}

		if appFolioError := eh.HandleHTTPResponse(resp, operation); appFolioError != nil {
			lastError = appFolioError
			if appFolioError.IsRetryable() && attempt < eh.maxRetries {
				resp.Body.Close()

				delay := eh.calculateBackoffDelay(attempt)
				if appFolioError.RetryAfter != nil && *appFolioError.RetryAfter > 0 {
					delay = *appFolioError.RetryAfter
					if delay > eh.maxDelay {
						delay = eh.maxDelay
					}
				}

				log.Printf("⏰ AppFolio: Retrying %s in %v due to retryable error: %s (Code: %s)",
					operation, delay, appFolioError.Message, appFolioError.ErrorCode)
				time.Sleep(delay)
				continue
			}
			return resp, appFolioError
		}

		if attempt > 0 {
			log.Printf("✅ AppFolio: Successfully executed %s on attempt %d", operation, attempt+1)
		}
		return resp, nil
	}

	log.Printf("❌ AppFolio: Failed to execute %s after %d attempts. Last error: %v", operation, eh.maxRetries+1, lastError)
	return nil, lastError
}

func (eh *AppFolioErrorHandler) calculateBackoffDelay(attempt int) time.Duration {
	if attempt == 0 {
		return eh.baseDelay
	}

	multiplier := 1.0
	for i := 0; i < attempt; i++ {
		multiplier *= eh.backoffFactor
	}

	delay := time.Duration(float64(eh.baseDelay) * multiplier)
	if delay > eh.maxDelay {
		delay = eh.maxDelay
	}
	return delay
}

type AppFolioValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e *AppFolioValidationError) Error() string {
	return fmt.Sprintf("AppFolio validation error [%s]: %s", e.Field, e.Message)
}

type AppFolioOperationResult struct {
	Success          bool                      `json:"success"`
	Operation        string                    `json:"operation"`
	ResourceType     string                    `json:"resource_type,omitempty"`
	ResourceID       string                    `json:"resource_id,omitempty"`
	Data             map[string]interface{}    `json:"data,omitempty"`
	Error            *AppFolioAPIError         `json:"error,omitempty"`
	ValidationErrors []AppFolioValidationError `json:"validation_errors,omitempty"`
	ExecutionTime    time.Duration             `json:"execution_time"`
	AttemptCount     int                       `json:"attempt_count"`
	Timestamp        time.Time                 `json:"timestamp"`
}

func CreateAppFolioSuccessResult(operation, resourceType, resourceID string, data map[string]interface{}, executionTime time.Duration, attemptCount int) *AppFolioOperationResult {
	return &AppFolioOperationResult{
		Success:       true,
		Operation:     operation,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Data:          data,
		ExecutionTime: executionTime,
		AttemptCount:  attemptCount,
		Timestamp:     time.Now(),
	}
}

func CreateAppFolioErrorResult(operation string, err *AppFolioAPIError, executionTime time.Duration, attemptCount int) *AppFolioOperationResult {
	return &AppFolioOperationResult{
		Success:       false,
		Operation:     operation,
		Error:         err,
		ExecutionTime: executionTime,
		AttemptCount:  attemptCount,
		Timestamp:     time.Now(),
	}
}

func CreateAppFolioValidationErrorResult(operation string, validationErrors []AppFolioValidationError) *AppFolioOperationResult {
	return &AppFolioOperationResult{
		Success:          false,
		Operation:        operation,
		ValidationErrors: validationErrors,
		ExecutionTime:    0,
		AttemptCount:     1,
		Timestamp:        time.Now(),
	}
}

func LogAppFolioOperation(result *AppFolioOperationResult) {
	if result.Success {
		log.Printf("✅ AppFolio Operation Success: %s completed in %v (attempts: %d)",
			result.Operation, result.ExecutionTime, result.AttemptCount)
		if result.ResourceID != "" {
			log.Printf("   Resource: %s/%s", result.ResourceType, result.ResourceID)
		}
	} else {
		if result.Error != nil {
			log.Printf("❌ AppFolio Operation Error: %s failed after %v (attempts: %d) - %s [%s]",
				result.Operation, result.ExecutionTime, result.AttemptCount,
				result.Error.Message, result.Error.ErrorCode)
			if result.Error.RequestID != "" {
				log.Printf("   Request ID: %s", result.Error.RequestID)
			}
		}
		if len(result.ValidationErrors) > 0 {
			log.Printf("⚠️ AppFolio Operation Validation Errors: %s - %d validation issues",
				result.Operation, len(result.ValidationErrors))
			for _, validationError := range result.ValidationErrors {
				log.Printf("   - %s: %s", validationError.Field, validationError.Message)
			}
		}
	}
}

func ValidateAppFolioPropertyRequest(property map[string]interface{}) []AppFolioValidationError {
	var errors []AppFolioValidationError

	if address, ok := property["address"].(map[string]interface{}); ok {
		if street, _ := address["street_1"].(string); street == "" {
			errors = append(errors, AppFolioValidationError{
				Field:   "address.street_1",
				Message: "Street address is required",
			})
		}
		if city, _ := address["city"].(string); city == "" {
			errors = append(errors, AppFolioValidationError{
				Field:   "address.city",
				Message: "City is required",
			})
		}
		if state, _ := address["state"].(string); state == "" {
			errors = append(errors, AppFolioValidationError{
				Field:   "address.state",
				Message: "State is required",
			})
		}
		if postalCode, _ := address["postal_code"].(string); postalCode == "" {
			errors = append(errors, AppFolioValidationError{
				Field:   "address.postal_code",
				Message: "Postal code is required",
			})
		}
	} else {
		errors = append(errors, AppFolioValidationError{
			Field:   "address",
			Message: "Address is required",
		})
	}

	if rentAmount, ok := property["rent_amount"].(float64); ok {
		if rentAmount < 0 {
			errors = append(errors, AppFolioValidationError{
				Field:   "rent_amount",
				Message: "Rent amount cannot be negative",
				Value:   rentAmount,
			})
		}
	}

	if unitCount, ok := property["unit_count"].(int); ok {
		if unitCount < 1 {
			errors = append(errors, AppFolioValidationError{
				Field:   "unit_count",
				Message: "Unit count must be at least 1",
				Value:   unitCount,
			})
		}
	}

	return errors
}

func ValidateAppFolioTenantRequest(tenant map[string]interface{}) []AppFolioValidationError {
	var errors []AppFolioValidationError

	firstName, _ := tenant["first_name"].(string)
	lastName, _ := tenant["last_name"].(string)

	if firstName == "" && lastName == "" {
		errors = append(errors, AppFolioValidationError{
			Field:   "name",
			Message: "At least first name or last name is required",
		})
	}

	email, _ := tenant["email"].(string)
	phone, _ := tenant["phone"].(string)

	if email == "" && phone == "" {
		errors = append(errors, AppFolioValidationError{
			Field:   "contact",
			Message: "At least email or phone is required",
		})
	}

	if email != "" && !isValidEmailFormat(email) {
		errors = append(errors, AppFolioValidationError{
			Field:   "email",
			Message: "Invalid email format",
			Value:   email,
		})
	}

	if rentAmount, ok := tenant["rent_amount"].(float64); ok {
		if rentAmount < 0 {
			errors = append(errors, AppFolioValidationError{
				Field:   "rent_amount",
				Message: "Rent amount cannot be negative",
				Value:   rentAmount,
			})
		}
	}

	return errors
}

func ValidateAppFolioMaintenanceRequest(request map[string]interface{}) []AppFolioValidationError {
	var errors []AppFolioValidationError

	if propertyID, _ := request["property_id"].(string); propertyID == "" {
		errors = append(errors, AppFolioValidationError{
			Field:   "property_id",
			Message: "Property ID is required",
		})
	}

	if description, _ := request["description"].(string); description == "" {
		errors = append(errors, AppFolioValidationError{
			Field:   "description",
			Message: "Description is required",
		})
	}

	if priority, ok := request["priority"].(string); ok {
		validPriorities := []string{"low", "medium", "high", "urgent", "emergency"}
		isValid := false
		for _, p := range validPriorities {
			if priority == p {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, AppFolioValidationError{
				Field:   "priority",
				Message: fmt.Sprintf("Invalid priority. Must be one of: %v", validPriorities),
				Value:   priority,
			})
		}
	}

	return errors
}

func ValidateAppFolioPaymentRequest(payment map[string]interface{}) []AppFolioValidationError {
	var errors []AppFolioValidationError

	if tenantID, _ := payment["tenant_id"].(string); tenantID == "" {
		errors = append(errors, AppFolioValidationError{
			Field:   "tenant_id",
			Message: "Tenant ID is required",
		})
	}

	if amount, ok := payment["amount"].(float64); ok {
		if amount <= 0 {
			errors = append(errors, AppFolioValidationError{
				Field:   "amount",
				Message: "Amount must be greater than 0",
				Value:   amount,
			})
		}
	} else {
		errors = append(errors, AppFolioValidationError{
			Field:   "amount",
			Message: "Amount is required",
		})
	}

	return errors
}

func isValidEmailFormat(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false
			}
			atIndex = i
		}
	}
	if atIndex < 1 || atIndex >= len(email)-1 {
		return false
	}
	domainPart := email[atIndex+1:]
	dotIndex := -1
	for i, c := range domainPart {
		if c == '.' {
			dotIndex = i
		}
	}
	return dotIndex > 0 && dotIndex < len(domainPart)-1
}
