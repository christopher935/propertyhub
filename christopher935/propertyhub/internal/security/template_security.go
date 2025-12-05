package security

import (
	"encoding/json"
	"fmt"
	"html/template"
)

// TemplateSecurity provides secure template functions
type TemplateSecurity struct {
	xssProtector *XSSProtector
}

// NewTemplateSecurity creates a new template security instance
func NewTemplateSecurity() *TemplateSecurity {
	return &TemplateSecurity{
		xssProtector: NewXSSProtector(),
	}
}

// GetSecureFuncMap returns a map of secure template functions
func (ts *TemplateSecurity) GetSecureFuncMap() template.FuncMap {
	return template.FuncMap{
		// HTML context escaping
		"safeHTML": ts.SafeHTML,
		"safeText": ts.SafeText,
		"safeAttr": ts.SafeAttribute,
		"safeURL":  ts.SafeURL,

		// JavaScript context escaping
		"safeJS":       ts.SafeJavaScript,
		"safeJSONAttr": ts.SafeJSONAttribute,

		// Format helpers with escaping
		"formatPrice":   ts.FormatPrice,
		"formatAddress": ts.FormatAddress,
		"formatText":    ts.FormatText,
		"truncate":      ts.TruncateText,

		// Utility functions
		"jsonEncode":   ts.JSONEncode,
		"escapeQuotes": ts.EscapeQuotes,
		"stripHTML":    ts.StripHTML,

		// Validation helpers
		"isValidURL": ts.IsValidURL,
		"detectXSS":  ts.DetectXSS,
	}
}

// SafeHTML sanitizes content for HTML context
func (ts *TemplateSecurity) SafeHTML(input interface{}) template.HTML {
	str := ts.toString(input)
	if str == "" {
		return template.HTML("")
	}

	sanitized := ts.xssProtector.SanitizeHTML(str)
	return template.HTML(sanitized)
}

// SafeText escapes text for safe HTML output
func (ts *TemplateSecurity) SafeText(input interface{}) string {
	str := ts.toString(input)
	if str == "" {
		return ""
	}

	// Use Go's built-in HTML escaping for text content
	return template.HTMLEscapeString(str)
}

// SafeAttribute sanitizes content for HTML attributes
func (ts *TemplateSecurity) SafeAttribute(input interface{}) string {
	str := ts.toString(input)
	if str == "" {
		return ""
	}

	sanitized := ts.xssProtector.SanitizeHTMLAttribute(str)
	return template.HTMLEscapeString(sanitized)
}

// SafeURL sanitizes URLs
func (ts *TemplateSecurity) SafeURL(input interface{}) string {
	str := ts.toString(input)
	if str == "" {
		return ""
	}

	return ts.xssProtector.SanitizeURL(str)
}

// SafeJavaScript sanitizes content for JavaScript context
func (ts *TemplateSecurity) SafeJavaScript(input interface{}) template.JS {
	str := ts.toString(input)
	if str == "" {
		return template.JS("")
	}

	sanitized := ts.xssProtector.SanitizeJavaScript(str)
	return template.JS(sanitized)
}

// SafeJSONAttribute creates safe JSON for HTML attributes
func (ts *TemplateSecurity) SafeJSONAttribute(input interface{}) string {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "{}"
	}

	jsonStr := string(jsonBytes)
	sanitized := ts.xssProtector.SanitizeForJSON(jsonStr)
	return template.HTMLEscapeString(sanitized)
}

// FormatPrice safely formats price values
func (ts *TemplateSecurity) FormatPrice(price interface{}) string {
	switch v := price.(type) {
	case int:
		return fmt.Sprintf("$%d", v)
	case int64:
		return fmt.Sprintf("$%d", v)
	case float64:
		return fmt.Sprintf("$%.0f", v)
	case string:
		// Sanitize string input first
		sanitized := ts.SafeText(v)
		return sanitized
	default:
		return "$0"
	}
}

// FormatAddress safely formats addresses
func (ts *TemplateSecurity) FormatAddress(address interface{}) string {
	str := ts.toString(address)
	if str == "" {
		return ""
	}

	// Remove any potentially dangerous content from addresses
	sanitized := ts.xssProtector.SanitizeHTML(str)
	return template.HTMLEscapeString(sanitized)
}

// FormatText safely formats text content with line breaks
func (ts *TemplateSecurity) FormatText(input interface{}) template.HTML {
	str := ts.toString(input)
	if str == "" {
		return template.HTML("")
	}

	// First sanitize the content
	sanitized := ts.xssProtector.SanitizeHTML(str)

	// Then allow only safe line breaks
	formatted := template.HTMLEscapeString(sanitized)

	// Convert \n to <br> tags
	formatted = template.HTMLEscapeString(formatted)

	return template.HTML(formatted)
}

// TruncateText safely truncates text
func (ts *TemplateSecurity) TruncateText(input interface{}, length int) string {
	str := ts.toString(input)
	if str == "" {
		return ""
	}

	// First sanitize
	sanitized := ts.xssProtector.SanitizeHTML(str)
	sanitized = template.HTMLEscapeString(sanitized)

	// Then truncate
	if len(sanitized) <= length {
		return sanitized
	}

	return sanitized[:length] + "..."
}

// JSONEncode safely encodes data as JSON
func (ts *TemplateSecurity) JSONEncode(input interface{}) template.JS {
	// First sanitize the input if it contains strings
	sanitized := ts.xssProtector.SanitizeResponseData(input)

	jsonBytes, err := json.Marshal(sanitized)
	if err != nil {
		return template.JS("{}")
	}

	return template.JS(string(jsonBytes))
}

// EscapeQuotes escapes quotes for JavaScript strings
func (ts *TemplateSecurity) EscapeQuotes(input interface{}) string {
	str := ts.toString(input)
	if str == "" {
		return ""
	}

	// First sanitize
	sanitized := ts.xssProtector.SanitizeJavaScript(str)

	// Then escape quotes
	escaped := template.JSEscapeString(sanitized)
	return escaped
}

// StripHTML removes all HTML tags
func (ts *TemplateSecurity) StripHTML(input interface{}) string {
	str := ts.toString(input)
	if str == "" {
		return ""
	}

	return ts.xssProtector.htmlSanitizer.StripAllHTML(str)
}

// IsValidURL validates if URL is safe
func (ts *TemplateSecurity) IsValidURL(input interface{}) bool {
	str := ts.toString(input)
	if str == "" {
		return false
	}

	return ts.xssProtector.urlSanitizer.IsValidURL(str)
}

// DetectXSS detects potential XSS attempts
func (ts *TemplateSecurity) DetectXSS(input interface{}) bool {
	str := ts.toString(input)
	if str == "" {
		return false
	}

	detected, _ := ts.xssProtector.DetectXSSAttempt(str)
	return detected
}

// toString safely converts interface{} to string
func (ts *TemplateSecurity) toString(input interface{}) string {
	if input == nil {
		return ""
	}

	switch v := input.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// SanitizeTemplateData sanitizes data before passing to template
func (ts *TemplateSecurity) SanitizeTemplateData(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		// Sanitize the key
		sanitizedKey := ts.xssProtector.SanitizeForJSON(key)

		// Sanitize the value based on context
		switch v := value.(type) {
		case string:
			// For string values, provide both raw (for further processing) and safe versions
			sanitized[sanitizedKey] = v
			sanitized[sanitizedKey+"_safe"] = ts.SafeText(v)
		case map[string]interface{}:
			// Recursively sanitize nested maps
			sanitized[sanitizedKey] = ts.SanitizeTemplateData(v)
		default:
			// For non-string values, pass through
			sanitized[sanitizedKey] = value
		}
	}

	return sanitized
}

// TemplateDataWrapper wraps data with security methods
type TemplateDataWrapper struct {
	Data map[string]interface{}
	ts   *TemplateSecurity
}

// NewTemplateDataWrapper creates a new template data wrapper
func (ts *TemplateSecurity) NewTemplateDataWrapper(data map[string]interface{}) *TemplateDataWrapper {
	return &TemplateDataWrapper{
		Data: data,
		ts:   ts,
	}
}

// SafeGet gets a value and returns it safely escaped for HTML
func (tdw *TemplateDataWrapper) SafeGet(key string) string {
	if value, exists := tdw.Data[key]; exists {
		return tdw.ts.SafeText(value)
	}
	return ""
}

// SafeGetURL gets a URL value and returns it safely
func (tdw *TemplateDataWrapper) SafeGetURL(key string) string {
	if value, exists := tdw.Data[key]; exists {
		return tdw.ts.SafeURL(value)
	}
	return ""
}

// SafeGetAttr gets a value for use in HTML attributes
func (tdw *TemplateDataWrapper) SafeGetAttr(key string) string {
	if value, exists := tdw.Data[key]; exists {
		return tdw.ts.SafeAttribute(value)
	}
	return ""
}

// RawGet gets a raw value (use with caution)
func (tdw *TemplateDataWrapper) RawGet(key string) interface{} {
	if value, exists := tdw.Data[key]; exists {
		return value
	}
	return nil
}
