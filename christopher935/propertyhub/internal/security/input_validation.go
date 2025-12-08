package security

import (
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	emailRegex        *regexp.Regexp
	phoneRegex        *regexp.Regexp
	addressRegex      *regexp.Regexp
	nameRegex         *regexp.Regexp
	sqlInjectionRegex *regexp.Regexp
	xssRegex          *regexp.Regexp
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{
		emailRegex:        regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		phoneRegex:        regexp.MustCompile(`^\+?1?[-.\s]?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}$`),
		addressRegex:      regexp.MustCompile(`^[a-zA-Z0-9\s,.-]+$`),
		nameRegex:         regexp.MustCompile(`^[a-zA-Z\s'-]+$`),
		sqlInjectionRegex: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick|alert|confirm|prompt|eval|expression|<script|</script>|<iframe|</iframe>)`),
		xssRegex:          regexp.MustCompile(`(?i)(<script|</script>|<iframe|</iframe>|javascript:|vbscript:|onload=|onerror=|onclick=|onmouseover=|onfocus=|onblur=|onchange=|onsubmit=|alert\(|confirm\(|prompt\(|eval\(|expression\()`),
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationResult contains validation results
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors"`
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(field, message, value string) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// FirstError returns the first error message, or empty string if no errors
func (vr *ValidationResult) FirstError() string {
	if len(vr.Errors) > 0 {
		return vr.Errors[0].Message
	}
	return ""
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return !vr.IsValid
}

// ValidateEmail validates and sanitizes email addresses
func (iv *InputValidator) ValidateEmail(email string) (string, error) {
	if email == "" {
		return "", ValidationError{Field: "email", Message: "email is required"}
	}

	// Sanitize
	email = strings.TrimSpace(strings.ToLower(email))
	email = html.EscapeString(email)

	// Length check
	if len(email) > 254 {
		return "", ValidationError{Field: "email", Message: "email too long (max 254 characters)"}
	}

	// Format validation using Go's mail package (more robust than regex)
	if _, err := mail.ParseAddress(email); err != nil {
		return "", ValidationError{Field: "email", Message: "invalid email format"}
	}

	// Additional regex check for common patterns
	if !iv.emailRegex.MatchString(email) {
		return "", ValidationError{Field: "email", Message: "invalid email format"}
	}

	// Check for SQL injection patterns
	if iv.sqlInjectionRegex.MatchString(email) {
		return "", ValidationError{Field: "email", Message: "invalid email format"}
	}

	return email, nil
}

// ValidatePhone validates and sanitizes phone numbers
func (iv *InputValidator) ValidatePhone(phone string) (string, error) {
	if phone == "" {
		return "", nil // Phone is optional in most cases
	}

	// Sanitize
	phone = strings.TrimSpace(phone)
	phone = html.EscapeString(phone)

	// Length check
	if len(phone) > 20 {
		return "", ValidationError{Field: "phone", Message: "phone number too long"}
	}

	// Format validation
	if !iv.phoneRegex.MatchString(phone) {
		return "", ValidationError{Field: "phone", Message: "invalid phone number format"}
	}

	// Check for SQL injection patterns
	if iv.sqlInjectionRegex.MatchString(phone) {
		return "", ValidationError{Field: "phone", Message: "invalid phone number format"}
	}

	return phone, nil
}

// ValidateName validates and sanitizes names
func (iv *InputValidator) ValidateName(name string) (string, error) {
	if name == "" {
		return "", ValidationError{Field: "name", Message: "name is required"}
	}

	// Sanitize
	name = strings.TrimSpace(name)
	name = html.EscapeString(name)

	// Length check
	if len(name) < 2 {
		return "", ValidationError{Field: "name", Message: "name too short (minimum 2 characters)"}
	}
	if len(name) > 100 {
		return "", ValidationError{Field: "name", Message: "name too long (maximum 100 characters)"}
	}

	// Format validation
	if !iv.nameRegex.MatchString(name) {
		return "", ValidationError{Field: "name", Message: "name contains invalid characters"}
	}

	// Check for SQL injection patterns
	if iv.sqlInjectionRegex.MatchString(name) {
		return "", ValidationError{Field: "name", Message: "name contains invalid characters"}
	}

	// Check for XSS patterns
	if iv.xssRegex.MatchString(name) {
		return "", ValidationError{Field: "name", Message: "name contains invalid characters"}
	}

	return name, nil
}

// ValidateAddress validates and sanitizes addresses
func (iv *InputValidator) ValidateAddress(address string) (string, error) {
	if address == "" {
		return "", ValidationError{Field: "address", Message: "address is required"}
	}

	// Sanitize
	address = strings.TrimSpace(address)
	address = html.EscapeString(address)

	// Length check
	if len(address) < 5 {
		return "", ValidationError{Field: "address", Message: "address too short"}
	}
	if len(address) > 200 {
		return "", ValidationError{Field: "address", Message: "address too long"}
	}

	// Format validation
	if !iv.addressRegex.MatchString(address) {
		return "", ValidationError{Field: "address", Message: "address contains invalid characters"}
	}

	// Check for SQL injection patterns
	if iv.sqlInjectionRegex.MatchString(address) {
		return "", ValidationError{Field: "address", Message: "address contains invalid characters"}
	}

	// Check for XSS patterns
	if iv.xssRegex.MatchString(address) {
		return "", ValidationError{Field: "address", Message: "address contains invalid characters"}
	}

	return address, nil
}

// ValidateText validates and sanitizes general text input
func (iv *InputValidator) ValidateText(text string, fieldName string, minLength, maxLength int) (string, error) {
	if text == "" && minLength > 0 {
		return "", ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is required", fieldName)}
	}

	// Sanitize
	text = strings.TrimSpace(text)
	text = html.EscapeString(text)

	// Length check
	if len(text) < minLength {
		return "", ValidationError{Field: fieldName, Message: fmt.Sprintf("%s too short (minimum %d characters)", fieldName, minLength)}
	}
	if len(text) > maxLength {
		return "", ValidationError{Field: fieldName, Message: fmt.Sprintf("%s too long (maximum %d characters)", fieldName, maxLength)}
	}

	// Check for SQL injection patterns
	if iv.sqlInjectionRegex.MatchString(text) {
		return "", ValidationError{Field: fieldName, Message: fmt.Sprintf("%s contains invalid characters", fieldName)}
	}

	// Check for XSS patterns
	if iv.xssRegex.MatchString(text) {
		return "", ValidationError{Field: fieldName, Message: fmt.Sprintf("%s contains invalid characters", fieldName)}
	}

	return text, nil
}

// ValidateID validates numeric IDs
func (iv *InputValidator) ValidateID(id uint, fieldName string) error {
	if id == 0 {
		return ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is required", fieldName)}
	}

	// Check for reasonable ID range (prevent overflow attacks)
	if id > 4294967295 { // Max uint32
		return ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is invalid", fieldName)}
	}

	return nil
}

// ValidatePrice validates price values
func (iv *InputValidator) ValidatePrice(price float64, fieldName string) error {
	if price < 0 {
		return ValidationError{Field: fieldName, Message: fmt.Sprintf("%s cannot be negative", fieldName)}
	}

	// Reasonable price range for real estate (prevent overflow/underflow attacks)
	if price > 100000000 { // $100M max
		return ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is too high", fieldName)}
	}

	return nil
}

// SanitizeHTML removes potentially dangerous HTML tags and attributes
func (iv *InputValidator) SanitizeHTML(input string) string {
	// First escape all HTML
	sanitized := html.EscapeString(input)

	// Remove any remaining script-like patterns
	sanitized = iv.xssRegex.ReplaceAllString(sanitized, "")

	return sanitized
}

// IsSQLInjectionAttempt checks if input contains SQL injection patterns
func (iv *InputValidator) IsSQLInjectionAttempt(input string) bool {
	return iv.sqlInjectionRegex.MatchString(input)
}

// IsXSSAttempt checks if input contains XSS patterns
func (iv *InputValidator) IsXSSAttempt(input string) bool {
	return iv.xssRegex.MatchString(input)
}

// ValidateBookingRequest validates a complete booking request
func (iv *InputValidator) ValidateBookingRequest(req map[string]interface{}) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Validate customer name
	if name, ok := req["customer_name"].(string); ok {
		if _, err := iv.ValidateName(name); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, name)
			}
		}
	} else {
		result.AddError("customer_name", "customer name is required", "")
	}

	// Validate customer email
	if email, ok := req["customer_email"].(string); ok {
		if _, err := iv.ValidateEmail(email); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, email)
			}
		}
	} else {
		result.AddError("customer_email", "customer email is required", "")
	}

	// Validate customer phone (optional)
	if phone, ok := req["customer_phone"].(string); ok && phone != "" {
		if _, err := iv.ValidatePhone(phone); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, phone)
			}
		}
	}

	// Validate property address
	if address, ok := req["property_address"].(string); ok {
		if _, err := iv.ValidateAddress(address); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, address)
			}
		}
	} else {
		result.AddError("property_address", "property address is required", "")
	}

	// Validate property ID
	if propertyID, ok := req["property_id"].(float64); ok {
		if err := iv.ValidateID(uint(propertyID), "property_id"); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, fmt.Sprintf("%.0f", propertyID))
			}
		}
	} else {
		result.AddError("property_id", "property ID is required", "")
	}

	return result
}

// ValidatePropertyNotificationRequest validates a property notification request
func (iv *InputValidator) ValidatePropertyNotificationRequest(req map[string]interface{}) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Validate customer name
	if name, ok := req["customer_name"].(string); ok {
		if _, err := iv.ValidateName(name); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, name)
			}
		}
	} else {
		result.AddError("customer_name", "customer name is required", "")
	}

	// Validate customer email
	if email, ok := req["customer_email"].(string); ok {
		if _, err := iv.ValidateEmail(email); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, email)
			}
		}
	} else {
		result.AddError("customer_email", "customer email is required", "")
	}

	// Validate notification type
	if notificationType, ok := req["notification_type"].(string); ok {
		validTypes := []string{"property_specific", "city_based", "area_based"}
		isValid := false
		for _, validType := range validTypes {
			if notificationType == validType {
				isValid = true
				break
			}
		}
		if !isValid {
			result.AddError("notification_type", "invalid notification type", notificationType)
		}
	} else {
		result.AddError("notification_type", "notification type is required", "")
	}

	// Validate city if provided
	if city, ok := req["city"].(string); ok && city != "" {
		if _, err := iv.ValidateText(city, "city", 2, 50); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(ve.Field, ve.Message, city)
			}
		}
	}

	return result
}

// ContainsOnlyPrintableChars checks if string contains only printable characters
func (iv *InputValidator) ContainsOnlyPrintableChars(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// TruncateString safely truncates a string to maximum length
func (iv *InputValidator) TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	// Truncate at word boundary if possible
	if maxLength > 10 {
		truncated := s[:maxLength-3]
		if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLength/2 {
			return truncated[:lastSpace] + "..."
		}
	}

	return s[:maxLength-3] + "..."
}

// ValidateInput validates individual input fields based on field name and value
func (iv *InputValidator) ValidateInput(fieldName string, value string) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Basic sanitization first
	value = strings.TrimSpace(value)
	value = html.EscapeString(value)

	// Check for SQL injection and XSS patterns first
	if iv.IsSQLInjectionAttempt(value) {
		result.AddError(fieldName, "Input contains invalid characters", value)
		return result
	}

	if iv.IsXSSAttempt(value) {
		result.AddError(fieldName, "Input contains invalid characters", value)
		return result
	}

	// Validate based on field name patterns
	switch {
	case strings.Contains(strings.ToLower(fieldName), "email"):
		if _, err := iv.ValidateEmail(value); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(fieldName, ve.Message, value)
			}
		}
	case strings.Contains(strings.ToLower(fieldName), "phone"):
		if _, err := iv.ValidatePhone(value); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(fieldName, ve.Message, value)
			}
		}
	case strings.Contains(strings.ToLower(fieldName), "name"):
		if _, err := iv.ValidateName(value); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(fieldName, ve.Message, value)
			}
		}
	case strings.Contains(strings.ToLower(fieldName), "address"):
		if _, err := iv.ValidateAddress(value); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(fieldName, ve.Message, value)
			}
		}
	case strings.Contains(strings.ToLower(fieldName), "sort"), strings.Contains(strings.ToLower(fieldName), "order"):
		// Special handling for sort/order parameters to prevent SQL injection
		allowedValues := []string{"asc", "desc", "created_at", "updated_at", "price", "address", "status", "name", "email", "date"}
		isValid := false
		for _, allowed := range allowedValues {
			if strings.ToLower(value) == allowed {
				isValid = true
				break
			}
		}
		if !isValid {
			result.AddError(fieldName, "Invalid sort parameter", value)
		}
	default:
		// General text validation
		if _, err := iv.ValidateText(value, fieldName, 0, 1000); err != nil {
			if ve, ok := err.(ValidationError); ok {
				result.AddError(fieldName, ve.Message, value)
			}
		}
	}

	return result
}

// ValidateAll validates a map of input fields
func (iv *InputValidator) ValidateAll(inputMap map[string]interface{}) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	for fieldName, value := range inputMap {
		// Convert value to string for validation
		var stringValue string
		switch v := value.(type) {
		case string:
			stringValue = v
		case int, int32, int64:
			stringValue = fmt.Sprintf("%d", v)
		case float32, float64:
			stringValue = fmt.Sprintf("%.2f", v)
		case bool:
			stringValue = fmt.Sprintf("%t", v)
		default:
			stringValue = fmt.Sprintf("%v", v)
		}

		// Validate the field
		fieldResult := iv.ValidateInput(fieldName, stringValue)
		if !fieldResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	}

	return result
}
