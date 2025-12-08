package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// JSONProtectionConfig holds configuration for JSON protection middleware
type JSONProtectionConfig struct {
	MaxRequestSize      int64    // Maximum request size in bytes (default: 1MB)
	MaxDepth            int      // Maximum JSON nesting depth (default: 10)
	MaxArrayLength      int      // Maximum array length (default: 1000)
	MaxObjectKeys       int      // Maximum number of object keys (default: 100)
	MaxStringLength     int      // Maximum string length (default: 10000)
	AllowedContentTypes []string // Allowed content types
	BlockedPatterns     []string // Patterns to block in JSON values
	DangerousJSPatterns []string // Dangerous JavaScript patterns to block
}

// DefaultJSONProtectionConfig returns default configuration
func DefaultJSONProtectionConfig() *JSONProtectionConfig {
	return &JSONProtectionConfig{
		MaxRequestSize:  1024 * 1024, // 1MB
		MaxDepth:        10,
		MaxArrayLength:  1000,
		MaxObjectKeys:   100,
		MaxStringLength: 10000,
		AllowedContentTypes: []string{
			"application/json",
			"application/json; charset=utf-8",
			"text/json",
		},
		BlockedPatterns: []string{
			// Script injection patterns
			"<script[^>]*>",
			"</script>",
			"javascript:",
			"vbscript:",
			"data:text/html",
			"data:text/javascript",

			// Event handler patterns
			"on(load|error|click|focus|blur|submit|change)",

			// SQL injection patterns
			"(union|select|insert|update|delete|drop|create|alter)\\s",
			"(or|and)\\s+\\d+\\s*=\\s*\\d+",
			"'\\s*(or|and)\\s+'",

			// Command injection patterns
			"(\\||&|;|`|\\$\\(|\\$\\{)",
			"(exec|eval|system|shell|cmd)",
		},
		DangerousJSPatterns: []string{
			"eval\\s*\\(",
			"Function\\s*\\(",
			"setTimeout\\s*\\(",
			"setInterval\\s*\\(",
			"execScript\\s*\\(",
			"new\\s+Function\\s*\\(",
			"document\\.write",
			"document\\.writeln",
			"innerHTML\\s*=",
			"outerHTML\\s*=",
		},
	}
}

// JSONProtectionMiddleware provides JSON validation and protection
type JSONProtectionMiddleware struct {
	config           *JSONProtectionConfig
	blockedRegexes   []*regexp.Regexp
	jsPatternRegexes []*regexp.Regexp
}

// NewJSONProtectionMiddleware creates a new JSON protection middleware
func NewJSONProtectionMiddleware(config *JSONProtectionConfig) *JSONProtectionMiddleware {
	if config == nil {
		config = DefaultJSONProtectionConfig()
	}

	// Compile blocked patterns
	var blockedRegexes []*regexp.Regexp
	for _, pattern := range config.BlockedPatterns {
		if regex, err := regexp.Compile("(?i)" + pattern); err == nil {
			blockedRegexes = append(blockedRegexes, regex)
		}
	}

	// Compile JavaScript patterns
	var jsPatternRegexes []*regexp.Regexp
	for _, pattern := range config.DangerousJSPatterns {
		if regex, err := regexp.Compile("(?i)" + pattern); err == nil {
			jsPatternRegexes = append(jsPatternRegexes, regex)
		}
	}

	return &JSONProtectionMiddleware{
		config:           config,
		blockedRegexes:   blockedRegexes,
		jsPatternRegexes: jsPatternRegexes,
	}
}

// JSONProtection returns the Gin middleware function
func (jpm *JSONProtectionMiddleware) JSONProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only process JSON requests
		contentType := strings.ToLower(c.GetHeader("Content-Type"))
		if !jpm.isAllowedContentType(contentType) {
			c.Next()
			return
		}

		// Check request size
		if c.Request.ContentLength > jpm.config.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request too large",
				"message": fmt.Sprintf("Maximum allowed size is %d bytes", jpm.config.MaxRequestSize),
			})
			c.Abort()
			return
		}

		// Read and validate JSON body
		body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, jpm.config.MaxRequestSize))
		if err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request too large",
				"message": "Failed to read request body",
			})
			c.Abort()
			return
		}

		// Validate JSON structure and content
		if err := jpm.validateJSON(body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON request",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		c.Next()
	}
}

// isAllowedContentType checks if content type is allowed
func (jpm *JSONProtectionMiddleware) isAllowedContentType(contentType string) bool {
	for _, allowed := range jpm.config.AllowedContentTypes {
		if strings.Contains(contentType, allowed) {
			return true
		}
	}
	return false
}

// validateJSON validates JSON structure and content
func (jpm *JSONProtectionMiddleware) validateJSON(data []byte) error {
	// Parse JSON to validate structure
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("invalid JSON syntax: %v", err)
	}

	// Validate structure constraints
	if err := jpm.validateStructure(parsed, 0); err != nil {
		return err
	}

	// Validate content for malicious patterns
	jsonString := string(data)
	if err := jpm.validateContent(jsonString); err != nil {
		return err
	}

	return nil
}

// validateStructure validates JSON structure constraints
func (jpm *JSONProtectionMiddleware) validateStructure(obj interface{}, depth int) error {
	// Check depth
	if depth > jpm.config.MaxDepth {
		return fmt.Errorf("JSON nesting too deep (max: %d)", jpm.config.MaxDepth)
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		// Check object key count
		if len(v) > jpm.config.MaxObjectKeys {
			return fmt.Errorf("too many object keys (max: %d)", jpm.config.MaxObjectKeys)
		}

		// Recursively validate nested objects
		for key, value := range v {
			// Validate key
			if len(key) > jpm.config.MaxStringLength {
				return fmt.Errorf("object key too long (max: %d)", jpm.config.MaxStringLength)
			}
			if err := jpm.validateStringContent(key); err != nil {
				return fmt.Errorf("invalid object key: %v", err)
			}

			// Recursively validate value
			if err := jpm.validateStructure(value, depth+1); err != nil {
				return err
			}
		}

	case []interface{}:
		// Check array length
		if len(v) > jpm.config.MaxArrayLength {
			return fmt.Errorf("array too long (max: %d)", jpm.config.MaxArrayLength)
		}

		// Recursively validate array elements
		for _, item := range v {
			if err := jpm.validateStructure(item, depth+1); err != nil {
				return err
			}
		}

	case string:
		// Check string length
		if len(v) > jpm.config.MaxStringLength {
			return fmt.Errorf("string too long (max: %d)", jpm.config.MaxStringLength)
		}

		// Validate string content
		if err := jpm.validateStringContent(v); err != nil {
			return err
		}
	}

	return nil
}

// validateContent validates JSON content for malicious patterns
func (jpm *JSONProtectionMiddleware) validateContent(content string) error {
	// Check for blocked patterns
	for _, regex := range jpm.blockedRegexes {
		if regex.MatchString(content) {
			return fmt.Errorf("blocked pattern detected in JSON content")
		}
	}

	// Check for dangerous JavaScript patterns
	for _, regex := range jpm.jsPatternRegexes {
		if regex.MatchString(content) {
			return fmt.Errorf("dangerous JavaScript pattern detected in JSON content")
		}
	}

	return nil
}

// validateStringContent validates individual string content
func (jpm *JSONProtectionMiddleware) validateStringContent(content string) error {
	// Check for blocked patterns
	for _, regex := range jpm.blockedRegexes {
		if regex.MatchString(content) {
			return fmt.Errorf("blocked pattern detected")
		}
	}

	// Check for dangerous JavaScript patterns
	for _, regex := range jpm.jsPatternRegexes {
		if regex.MatchString(content) {
			return fmt.Errorf("dangerous JavaScript pattern detected")
		}
	}

	// Additional specific checks
	if strings.Contains(strings.ToLower(content), "<script") {
		return fmt.Errorf("script tag detected")
	}

	if strings.Contains(strings.ToLower(content), "javascript:") {
		return fmt.Errorf("javascript protocol detected")
	}

	return nil
}

// SanitizeJSON sanitizes JSON content by removing dangerous patterns
func (jpm *JSONProtectionMiddleware) SanitizeJSON(data []byte) ([]byte, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	sanitized := jpm.sanitizeObject(obj)

	return json.Marshal(sanitized)
}

// sanitizeObject recursively sanitizes JSON object
func (jpm *JSONProtectionMiddleware) sanitizeObject(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			sanitizedKey := jpm.sanitizeString(key)
			sanitized[sanitizedKey] = jpm.sanitizeObject(value)
		}
		return sanitized

	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = jpm.sanitizeObject(item)
		}
		return sanitized

	case string:
		return jpm.sanitizeString(v)

	default:
		return v
	}
}

// sanitizeString sanitizes string content
func (jpm *JSONProtectionMiddleware) sanitizeString(content string) string {
	sanitized := content

	// Remove blocked patterns
	for _, regex := range jpm.blockedRegexes {
		sanitized = regex.ReplaceAllString(sanitized, "")
	}

	// Remove dangerous JavaScript patterns
	for _, regex := range jpm.jsPatternRegexes {
		sanitized = regex.ReplaceAllString(sanitized, "")
	}

	return sanitized
}

// LogJSONValidationError logs JSON validation errors for monitoring
func (jpm *JSONProtectionMiddleware) LogJSONValidationError(c *gin.Context, err error) {
	// Log validation error with context
	c.Header("X-Validation-Error", "true")

	// You can integrate with your logging system here
	// For now, we'll use Gin's logger
	c.Error(fmt.Errorf("JSON validation failed: %v", err))
}

// GetValidationStats returns validation statistics
type ValidationStats struct {
	TotalRequests     int64
	BlockedRequests   int64
	SanitizedRequests int64
	LastBlockedAt     string
}

// You can implement statistics collection if needed
func (jpm *JSONProtectionMiddleware) GetValidationStats() *ValidationStats {
	// Implementation would depend on your metrics collection system
	return &ValidationStats{}
}
