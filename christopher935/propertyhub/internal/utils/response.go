package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	
	"github.com/gin-gonic/gin"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// ResponseHelper provides centralized response handling utilities
type ResponseHelper struct{}

// NewResponseHelper creates a new response helper instance
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// SetCORSHeaders sets standard CORS headers with configurable methods
func (rh *ResponseHelper) SetCORSHeaders(w http.ResponseWriter, methods ...string) {
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080" // Safe fallback for development
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if len(methods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
	} else {
		// Default to common methods if none specified
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	}
}

// SetJSONContentType sets the content type for JSON responses
func (rh *ResponseHelper) SetJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// SetStandardHeaders sets both CORS and JSON content type headers
func (rh *ResponseHelper) SetStandardHeaders(w http.ResponseWriter, methods ...string) {
	rh.SetCORSHeaders(w, methods...)
	rh.SetJSONContentType(w)
}

// WriteJSONResponse writes a standardized JSON response
func (rh *ResponseHelper) WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	rh.SetJSONContentType(w)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, write a simple error response
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"error":"Internal server error"}`))
	}
}

// WriteSuccessResponse writes a standardized success response
func (rh *ResponseHelper) WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	rh.WriteJSONResponse(w, http.StatusOK, response)
}

// WriteSuccessResponseWithCount writes a success response that includes a count field
func (rh *ResponseHelper) WriteSuccessResponseWithCount(w http.ResponseWriter, message string, data interface{}, count int) {
	response := StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
		Count:   count,
	}
	rh.WriteJSONResponse(w, http.StatusOK, response)
}

// WriteErrorResponse writes a standardized error response
func (rh *ResponseHelper) WriteErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string) {
	response := StandardResponse{
		Success: false,
		Error:   errorMessage,
		Message: errorMessage, // Include both for backward compatibility
	}
	rh.WriteJSONResponse(w, statusCode, response)
}

// WriteBadRequestError writes a 400 Bad Request error response
func (rh *ResponseHelper) WriteBadRequestError(w http.ResponseWriter, message string) {
	rh.WriteErrorResponse(w, http.StatusBadRequest, message)
}

// WriteUnauthorizedError writes a 401 Unauthorized error response
func (rh *ResponseHelper) WriteUnauthorizedError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	rh.WriteErrorResponse(w, http.StatusUnauthorized, message)
}

// WriteNotFoundError writes a 404 Not Found error response
func (rh *ResponseHelper) WriteNotFoundError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	rh.WriteErrorResponse(w, http.StatusNotFound, message)
}

// WriteInternalServerError writes a 500 Internal Server Error response
func (rh *ResponseHelper) WriteInternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	rh.WriteErrorResponse(w, http.StatusInternalServerError, message)
}

// WriteMethodNotAllowedError writes a 405 Method Not Allowed error response
func (rh *ResponseHelper) WriteMethodNotAllowedError(w http.ResponseWriter, allowedMethods ...string) {
	message := "Method not allowed"
	if len(allowedMethods) > 0 {
		message = "Method not allowed. Allowed methods: " + strings.Join(allowedMethods, ", ")
	}
	rh.WriteErrorResponse(w, http.StatusMethodNotAllowed, message)
}

// HandleOptionsRequest handles preflight OPTIONS requests
// Returns true if it was an OPTIONS request and was handled
func (rh *ResponseHelper) HandleOptionsRequest(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	if r.Method == "OPTIONS" {
		rh.SetCORSHeaders(w, methods...)
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

// ValidateHTTPMethod validates that the request method is allowed
// Returns true if method is valid, false if invalid (and writes error response)
func (rh *ResponseHelper) ValidateHTTPMethod(w http.ResponseWriter, r *http.Request, allowedMethods ...string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return true
		}
	}

	rh.WriteMethodNotAllowedError(w, allowedMethods...)
	return false
}

// ValidateJSONContentType validates that the request has JSON content type
// Returns true if valid, false if invalid (and writes error response)
func (rh *ResponseHelper) ValidateJSONContentType(w http.ResponseWriter, r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		rh.WriteBadRequestError(w, "Content-Type must be application/json")
		return false
	}
	return true
}

// ParseJSONBody parses JSON request body into the provided interface
// Returns true if successful, false if parsing failed (and writes error response)
func (rh *ResponseHelper) ParseJSONBody(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		rh.WriteBadRequestError(w, "Invalid JSON in request body")
		return false
	}
	return true
}

// HandleStandardHTTPFlow handles the common HTTP flow:
// 1. Handle OPTIONS requests
// 2. Validate HTTP method
// 3. Set standard headers
// Returns true if the request should continue, false if it was handled/rejected
func (rh *ResponseHelper) HandleStandardHTTPFlow(w http.ResponseWriter, r *http.Request, allowedMethods ...string) bool {
	// Handle preflight OPTIONS request
	if rh.HandleOptionsRequest(w, r, allowedMethods...) {
		return false // Request was handled, don't continue
	}

	// Validate HTTP method
	if !rh.ValidateHTTPMethod(w, r, allowedMethods...) {
		return false // Method not allowed, don't continue
	}

	// Set standard headers
	rh.SetStandardHeaders(w, allowedMethods...)

	return true // Continue processing
}

// Global response helper instance for convenience
var DefaultResponseHelper = NewResponseHelper()

// Convenience functions that use the global instance

// SetCORSHeaders sets standard CORS headers with configurable methods
func SetCORSHeaders(w http.ResponseWriter, methods ...string) {
	DefaultResponseHelper.SetCORSHeaders(w, methods...)
}

// WriteSuccessResponse writes a standardized success response
func WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	DefaultResponseHelper.WriteSuccessResponse(w, message, data)
}

// WriteErrorResponse writes a standardized error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string) {
	DefaultResponseHelper.WriteErrorResponse(w, statusCode, errorMessage)
}

// WriteBadRequestError writes a 400 Bad Request error response
func WriteBadRequestError(w http.ResponseWriter, message string) {
	DefaultResponseHelper.WriteBadRequestError(w, message)
}

// WriteUnauthorizedError writes a 401 Unauthorized error response
func WriteUnauthorizedError(w http.ResponseWriter, message string) {
	DefaultResponseHelper.WriteUnauthorizedError(w, message)
}

// WriteInternalServerError writes a 500 Internal Server Error response
func WriteInternalServerError(w http.ResponseWriter, message string) {
	DefaultResponseHelper.WriteInternalServerError(w, message)
}

// Gin helper functions for consistent API responses

// SuccessResponse sends a success response using Gin
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// ErrorResponse sends an error response using Gin
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	response := gin.H{
		"success": false,
		"error":   message,
	}
	
	if err != nil {
		response["details"] = err.Error()
	}
	
	c.JSON(statusCode, response)
}

// HandleStandardHTTPFlow handles the common HTTP flow
func HandleStandardHTTPFlow(w http.ResponseWriter, r *http.Request, allowedMethods ...string) bool {
	return DefaultResponseHelper.HandleStandardHTTPFlow(w, r, allowedMethods...)
}
