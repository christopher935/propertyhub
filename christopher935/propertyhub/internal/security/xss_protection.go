package security

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// XSSProtector provides comprehensive XSS protection
type XSSProtector struct {
	htmlSanitizer      *HTMLSanitizer
	jsSanitizer        *JavaScriptSanitizer
	urlSanitizer       *URLSanitizer
	attributeSanitizer *AttributeSanitizer
}

// NewXSSProtector creates a new XSS protector
func NewXSSProtector() *XSSProtector {
	return &XSSProtector{
		htmlSanitizer:      NewHTMLSanitizer(),
		jsSanitizer:        NewJavaScriptSanitizer(),
		urlSanitizer:       NewURLSanitizer(),
		attributeSanitizer: NewAttributeSanitizer(),
	}
}

// SanitizeHTML sanitizes HTML content to prevent XSS
func (xp *XSSProtector) SanitizeHTML(input string) string {
	return xp.htmlSanitizer.Sanitize(input)
}

// SanitizeHTMLAttribute sanitizes HTML attributes
func (xp *XSSProtector) SanitizeHTMLAttribute(input string) string {
	return xp.attributeSanitizer.SanitizeAttribute(input)
}

// SanitizeJavaScript sanitizes JavaScript content
func (xp *XSSProtector) SanitizeJavaScript(input string) string {
	return xp.jsSanitizer.Sanitize(input)
}

// SanitizeURL sanitizes URLs to prevent XSS
func (xp *XSSProtector) SanitizeURL(input string) string {
	return xp.urlSanitizer.Sanitize(input)
}

// SanitizeForJSON sanitizes content for JSON output
func (xp *XSSProtector) SanitizeForJSON(input string) string {
	// Escape HTML entities
	sanitized := html.EscapeString(input)

	// Remove dangerous JavaScript patterns
	sanitized = xp.jsSanitizer.RemoveDangerousPatterns(sanitized)

	return sanitized
}

// SanitizeForHTML sanitizes content for HTML template output
func (xp *XSSProtector) SanitizeForHTML(input string) template.HTML {
	sanitized := xp.htmlSanitizer.Sanitize(input)
	return template.HTML(sanitized)
}

// HTMLSanitizer handles HTML sanitization
type HTMLSanitizer struct {
	dangerousTags  *regexp.Regexp
	dangerousAttrs *regexp.Regexp
	scriptPatterns *regexp.Regexp
	eventHandlers  *regexp.Regexp
}

// NewHTMLSanitizer creates a new HTML sanitizer
func NewHTMLSanitizer() *HTMLSanitizer {
	return &HTMLSanitizer{
		dangerousTags:  regexp.MustCompile(`(?i)<\s*(script|iframe|object|embed|form|input|textarea|select|button|link|meta|style|base|applet|bgsound|blink|body|frame|frameset|head|html|ilayer|layer|plaintext|title|xml)[^>]*>`),
		dangerousAttrs: regexp.MustCompile(`(?i)(on\w+|javascript:|vbscript:|data:|about:|mocha:|livescript:)`),
		scriptPatterns: regexp.MustCompile(`(?i)(<script[^>]*>.*?</script>|<style[^>]*>.*?</style>)`),
		eventHandlers:  regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["\']?[^"\'>\s]*["\']?`),
	}
}

// Sanitize sanitizes HTML content
func (hs *HTMLSanitizer) Sanitize(input string) string {
	if input == "" {
		return ""
	}

	// Remove script and style tags completely
	sanitized := hs.scriptPatterns.ReplaceAllString(input, "")

	// Remove dangerous tags
	sanitized = hs.dangerousTags.ReplaceAllStringFunc(sanitized, func(match string) string {
		return html.EscapeString(match)
	})

	// Remove event handlers
	sanitized = hs.eventHandlers.ReplaceAllString(sanitized, "")

	// Remove dangerous attributes
	sanitized = hs.dangerousAttrs.ReplaceAllStringFunc(sanitized, func(match string) string {
		return html.EscapeString(match)
	})

	// Final HTML escape for any remaining dangerous content
	sanitized = html.EscapeString(sanitized)

	return sanitized
}

// StripAllHTML removes all HTML tags
func (hs *HTMLSanitizer) StripAllHTML(input string) string {
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	stripped := htmlTagRegex.ReplaceAllString(input, "")
	return html.UnescapeString(stripped)
}

// AllowBasicHTML allows only basic safe HTML tags
func (hs *HTMLSanitizer) AllowBasicHTML(input string) string {
	// Allow only specific safe tags
	allowedTags := []string{"p", "br", "strong", "b", "em", "i", "u", "span", "div"}

	// First escape everything
	sanitized := html.EscapeString(input)

	// Then unescape allowed tags
	for _, tag := range allowedTags {
		openTag := "&lt;" + tag + "&gt;"
		closeTag := "&lt;/" + tag + "&gt;"
		sanitized = strings.ReplaceAll(sanitized, openTag, "<"+tag+">")
		sanitized = strings.ReplaceAll(sanitized, closeTag, "</"+tag+">")
	}

	return sanitized
}

// JavaScriptSanitizer handles JavaScript sanitization
type JavaScriptSanitizer struct {
	dangerousPatterns *regexp.Regexp
	functionCalls     *regexp.Regexp
	evalPatterns      *regexp.Regexp
}

// NewJavaScriptSanitizer creates a new JavaScript sanitizer
func NewJavaScriptSanitizer() *JavaScriptSanitizer {
	return &JavaScriptSanitizer{
		dangerousPatterns: regexp.MustCompile(`(?i)(javascript:|vbscript:|data:|about:|mocha:|livescript:)`),
		functionCalls:     regexp.MustCompile(`(?i)(eval|setTimeout|setInterval|Function|execScript|msWriteProfilerMark)\s*\(`),
		evalPatterns:      regexp.MustCompile(`(?i)(eval|new\s+Function|setTimeout|setInterval)\s*\(\s*["\']`),
	}
}

// Sanitize sanitizes JavaScript content
func (js *JavaScriptSanitizer) Sanitize(input string) string {
	if input == "" {
		return ""
	}

	// Remove dangerous JavaScript patterns
	sanitized := js.RemoveDangerousPatterns(input)

	// HTML escape the result
	sanitized = html.EscapeString(sanitized)

	return sanitized
}

// RemoveDangerousPatterns removes dangerous JavaScript patterns
func (js *JavaScriptSanitizer) RemoveDangerousPatterns(input string) string {
	// Remove dangerous protocols
	sanitized := js.dangerousPatterns.ReplaceAllString(input, "")

	// Remove dangerous function calls
	sanitized = js.functionCalls.ReplaceAllString(sanitized, "")

	// Remove eval patterns
	sanitized = js.evalPatterns.ReplaceAllString(sanitized, "")

	return sanitized
}

// IsJavaScriptInjection checks if input contains JavaScript injection
func (js *JavaScriptSanitizer) IsJavaScriptInjection(input string) bool {
	return js.dangerousPatterns.MatchString(input) ||
		js.functionCalls.MatchString(input) ||
		js.evalPatterns.MatchString(input)
}

// URLSanitizer handles URL sanitization
type URLSanitizer struct {
	dangerousSchemes *regexp.Regexp
	validURL         *regexp.Regexp
}

// NewURLSanitizer creates a new URL sanitizer
func NewURLSanitizer() *URLSanitizer {
	return &URLSanitizer{
		dangerousSchemes: regexp.MustCompile(`(?i)^(javascript|vbscript|data|about|mocha|livescript|file|ftp):`),
		validURL:         regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`),
	}
}

// Sanitize sanitizes URLs
func (us *URLSanitizer) Sanitize(input string) string {
	if input == "" {
		return ""
	}

	// Remove dangerous schemes
	if us.dangerousSchemes.MatchString(input) {
		return ""
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return ""
	}

	// Only allow HTTP and HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" && parsedURL.Scheme != "" {
		return ""
	}

	// If no scheme, assume https
	if parsedURL.Scheme == "" {
		input = "https://" + input
	}

	// Re-parse with scheme
	parsedURL, err = url.Parse(input)
	if err != nil {
		return ""
	}

	// Validate hostname
	if parsedURL.Host == "" {
		return ""
	}

	return parsedURL.String()
}

// IsValidURL checks if URL is valid and safe
func (us *URLSanitizer) IsValidURL(input string) bool {
	if us.dangerousSchemes.MatchString(input) {
		return false
	}

	_, err := url.Parse(input)
	return err == nil
}

// AttributeSanitizer handles HTML attribute sanitization
type AttributeSanitizer struct {
	dangerousAttrs *regexp.Regexp
	eventHandlers  *regexp.Regexp
}

// NewAttributeSanitizer creates a new attribute sanitizer
func NewAttributeSanitizer() *AttributeSanitizer {
	return &AttributeSanitizer{
		dangerousAttrs: regexp.MustCompile(`(?i)(javascript:|vbscript:|data:|about:|mocha:|livescript:)`),
		eventHandlers:  regexp.MustCompile(`(?i)^on\w+`),
	}
}

// SanitizeAttribute sanitizes HTML attributes
func (as *AttributeSanitizer) SanitizeAttribute(input string) string {
	if input == "" {
		return ""
	}

	// Remove dangerous protocols
	sanitized := as.dangerousAttrs.ReplaceAllString(input, "")

	// HTML escape the result
	sanitized = html.EscapeString(sanitized)

	return sanitized
}

// IsEventHandler checks if attribute is an event handler
func (as *AttributeSanitizer) IsEventHandler(attrName string) bool {
	return as.eventHandlers.MatchString(attrName)
}

// IsDangerousAttribute checks if attribute contains dangerous content
func (as *AttributeSanitizer) IsDangerousAttribute(attrValue string) bool {
	return as.dangerousAttrs.MatchString(attrValue)
}

// ContentSecurityPolicy generates CSP headers
type ContentSecurityPolicy struct {
	directives map[string][]string
}

// NewContentSecurityPolicy creates a new CSP generator
func NewContentSecurityPolicy() *ContentSecurityPolicy {
	return &ContentSecurityPolicy{
		directives: make(map[string][]string),
	}
}

// AddDirective adds a CSP directive
func (csp *ContentSecurityPolicy) AddDirective(directive string, sources ...string) {
	csp.directives[directive] = append(csp.directives[directive], sources...)
}

// SetDefaultDirectives sets secure default CSP directives
func (csp *ContentSecurityPolicy) SetDefaultDirectives() {
	csp.directives = map[string][]string{
		"default-src":     {"'self'"},
		"script-src":      {"'self'", "'unsafe-inline'"},
		"style-src":       {"'self'", "'unsafe-inline'"},
		"img-src":         {"'self'", "data:", "https:"},
		"font-src":        {"'self'", "https:"},
		"connect-src":     {"'self'", "https:"},
		"frame-ancestors": {"'none'"},
		"base-uri":        {"'self'"},
		"form-action":     {"'self'"},
	}
}

// GenerateHeader generates the CSP header value
func (csp *ContentSecurityPolicy) GenerateHeader() string {
	var parts []string

	for directive, sources := range csp.directives {
		part := directive + " " + strings.Join(sources, " ")
		parts = append(parts, part)
	}

	return strings.Join(parts, "; ")
}

// XSSProtectionMiddleware provides XSS protection middleware
type XSSProtectionMiddleware struct {
	protector *XSSProtector
	csp       *ContentSecurityPolicy
}

// NewXSSProtectionMiddleware creates XSS protection middleware
func NewXSSProtectionMiddleware() *XSSProtectionMiddleware {
	csp := NewContentSecurityPolicy()
	csp.SetDefaultDirectives()

	return &XSSProtectionMiddleware{
		protector: NewXSSProtector(),
		csp:       csp,
	}
}

// Middleware returns the XSS protection middleware
func (xpm *XSSProtectionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set XSS protection headers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", xpm.csp.GenerateHeader())

		next.ServeHTTP(w, r)
	})
}

// SanitizeResponseData sanitizes data before sending in response
func (xp *XSSProtector) SanitizeResponseData(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		return xp.SanitizeForJSON(v)
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			sanitizedKey := xp.SanitizeForJSON(key)
			sanitized[sanitizedKey] = xp.SanitizeResponseData(value)
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = xp.SanitizeResponseData(item)
		}
		return sanitized
	default:
		return v
	}
}

// ValidateAndSanitizeInput validates and sanitizes user input
func (xp *XSSProtector) ValidateAndSanitizeInput(input map[string]interface{}) (map[string]interface{}, []string) {
	sanitized := make(map[string]interface{})
	var warnings []string

	for key, value := range input {
		// Sanitize key
		sanitizedKey := xp.SanitizeForJSON(key)
		if sanitizedKey != key {
			warnings = append(warnings, fmt.Sprintf("Key '%s' was sanitized", key))
		}

		// Sanitize value based on type
		switch v := value.(type) {
		case string:
			sanitizedValue := xp.SanitizeForJSON(v)
			if sanitizedValue != v {
				warnings = append(warnings, fmt.Sprintf("Value for key '%s' was sanitized", key))
			}
			sanitized[sanitizedKey] = sanitizedValue
		case map[string]interface{}:
			sanitizedNested, nestedWarnings := xp.ValidateAndSanitizeInput(v)
			sanitized[sanitizedKey] = sanitizedNested
			warnings = append(warnings, nestedWarnings...)
		default:
			sanitized[sanitizedKey] = v
		}
	}

	return sanitized, warnings
}

// DetectXSSAttempt detects potential XSS attempts
func (xp *XSSProtector) DetectXSSAttempt(input string) (bool, string) {
	// Check for script tags
	if strings.Contains(strings.ToLower(input), "<script") {
		return true, "Script tag detected"
	}

	// Check for JavaScript protocols
	if xp.jsSanitizer.IsJavaScriptInjection(input) {
		return true, "JavaScript injection detected"
	}

	// Check for event handlers
	eventHandlerRegex := regexp.MustCompile(`(?i)on\w+\s*=`)
	if eventHandlerRegex.MatchString(input) {
		return true, "Event handler detected"
	}

	// Check for dangerous HTML tags
	dangerousTagRegex := regexp.MustCompile(`(?i)<\s*(iframe|object|embed|form|input)`)
	if dangerousTagRegex.MatchString(input) {
		return true, "Dangerous HTML tag detected"
	}

	return false, ""
}
