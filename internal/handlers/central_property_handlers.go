package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"
)

// CentralPropertyHandler handles API requests for the Central Property State Manager
type CentralPropertyHandler struct {
	stateManager *services.CentralPropertyStateManager
}

// NewCentralPropertyHandler creates a new central property handler
func NewCentralPropertyHandler(db *gorm.DB, encryptionManager *security.EncryptionManager) *CentralPropertyHandler {
	return &CentralPropertyHandler{
		stateManager: services.NewCentralPropertyStateManager(db, encryptionManager),
	}
}

// CreateOrUpdateProperty handles POST/PUT /api/v1/central-properties
func (cph *CentralPropertyHandler) CreateOrUpdateProperty(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST and PUT methods
	if r.Method != "POST" && r.Method != "PUT" {
		cph.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST and PUT methods are allowed")
		return
	}

	// Parse request body
	var updateReq models.PropertyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		log.Printf("Error parsing property update request: %v", err)
		cph.sendErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if updateReq.Address == "" {
		cph.sendErrorResponse(w, http.StatusBadRequest, "MISSING_ADDRESS", "Property address is required")
		return
	}

	if updateReq.Source == "" {
		cph.sendErrorResponse(w, http.StatusBadRequest, "MISSING_SOURCE", "Source system is required")
		return
	}

	// Create or update property
	property, err := cph.stateManager.CreateOrUpdateProperty(updateReq)
	if err != nil {
		log.Printf("Error creating/updating property: %v", err)
		cph.sendErrorResponse(w, http.StatusInternalServerError, "PROPERTY_ERROR", fmt.Sprintf("Failed to process property: %v", err))
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Property processed successfully",
		"property": property,
	})

	log.Printf("✅ Central Property API: Property %s processed from %s", property.Address, updateReq.Source)
}

// GetProperty handles GET /api/v1/central-properties/{id}
func (cph *CentralPropertyHandler) GetProperty(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET method
	if r.Method != "GET" {
		cph.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Get property identifier from URL path
	identifier := r.URL.Path[len("/api/v1/central-properties/"):]
	if identifier == "" {
		cph.sendErrorResponse(w, http.StatusBadRequest, "MISSING_IDENTIFIER", "Property identifier is required")
		return
	}

	// Get property from central state
	property, err := cph.stateManager.GetPropertyState(identifier)
	if err != nil {
		log.Printf("Error retrieving property %s: %v", identifier, err)
		cph.sendErrorResponse(w, http.StatusNotFound, "PROPERTY_NOT_FOUND", fmt.Sprintf("Property not found: %s", identifier))
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"property": property,
	})
}

// GetPropertiesByStatus handles GET /api/v1/central-properties?status={status}
func (cph *CentralPropertyHandler) GetPropertiesByStatus(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET method
	if r.Method != "GET" {
		cph.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Get status from query parameters
	status := r.URL.Query().Get("status")
	if status == "" {
		cph.sendErrorResponse(w, http.StatusBadRequest, "MISSING_STATUS", "Status parameter is required")
		return
	}

	// Get all properties and filter by status
	allProperties, err := cph.stateManager.GetPublicProperties()
	if err != nil {
		log.Printf("Error retrieving properties: %v", err)
		cph.sendErrorResponse(w, http.StatusInternalServerError, "RETRIEVAL_ERROR", fmt.Sprintf("Failed to retrieve properties: %v", err))
		return
	}

	// Filter by status
	var properties []models.PropertyState
	for _, prop := range allProperties {
		if prop.Status == status {
			properties = append(properties, prop)
		}
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"status":     status,
		"count":      len(properties),
		"properties": properties,
	})
}

// UpdatePropertyStatus handles PATCH /api/v1/central-properties/{mlsId}/status
func (cph *CentralPropertyHandler) UpdatePropertyStatus(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow PATCH method
	if r.Method != "PATCH" {
		cph.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PATCH method is allowed")
		return
	}

	// Get MLS ID from URL path
	mlsID := r.URL.Path[len("/api/v1/central-properties/"):]
	if statusIndex := len(mlsID) - len("/status"); statusIndex > 0 && mlsID[statusIndex:] == "/status" {
		mlsID = mlsID[:statusIndex]
	}

	if mlsID == "" {
		cph.sendErrorResponse(w, http.StatusBadRequest, "MISSING_MLS_ID", "MLS ID is required")
		return
	}

	// Parse request body
	var statusReq struct {
		Status string `json:"status"`
		Source string `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&statusReq); err != nil {
		log.Printf("Error parsing status update request: %v", err)
		cph.sendErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	// Validate required fields
	if statusReq.Status == "" {
		cph.sendErrorResponse(w, http.StatusBadRequest, "MISSING_STATUS", "Status is required")
		return
	}

	if statusReq.Source == "" {
		cph.sendErrorResponse(w, http.StatusBadRequest, "MISSING_SOURCE", "Source system is required")
		return
	}

	// Update property status
	if err := cph.stateManager.UpdatePropertyStatus(mlsID, statusReq.Status, statusReq.Source); err != nil {
		log.Printf("Error updating property status: %v", err)
		cph.sendErrorResponse(w, http.StatusInternalServerError, "STATUS_UPDATE_ERROR", fmt.Sprintf("Failed to update status: %v", err))
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Status updated to %s for property %s", statusReq.Status, mlsID),
	})

	log.Printf("✅ Central Property API: Status updated for %s to %s (Source: %s)", mlsID, statusReq.Status, statusReq.Source)
}

// GetSystemStats handles GET /api/v1/central-properties/stats
func (cph *CentralPropertyHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET method
	if r.Method != "GET" {
		cph.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Get system statistics
	stats, err := cph.stateManager.GetSystemStats()
	if err != nil {
		log.Printf("Error retrieving system stats: %v", err)
		cph.sendErrorResponse(w, http.StatusInternalServerError, "STATS_ERROR", fmt.Sprintf("Failed to retrieve stats: %v", err))
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// TestCentralState handles GET /api/v1/central-properties/test
func (cph *CentralPropertyHandler) TestCentralState(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET method
	if r.Method != "GET" {
		cph.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Create test property data
	testProperty := models.PropertyUpdateRequest{
		MLSId:        "TEST_MLS_12345",
		Address:      "123 Test Street, Houston, TX 77001",
		Price:        &[]float64{350000}[0],
		Bedrooms:     &[]int{3}[0],
		Bathrooms:    &[]float32{2.5}[0],
		SquareFeet:   &[]int{1800}[0],
		PropertyType: "Single Family",
		Status:       "active",
		Source:       "test",
		Data: models.JSONB{
			"test_mode":  true,
			"created_by": "central_property_test",
		},
	}

	// Create test property
	property, err := cph.stateManager.CreateOrUpdateProperty(testProperty)
	if err != nil {
		log.Printf("Error creating test property: %v", err)
		cph.sendErrorResponse(w, http.StatusInternalServerError, "TEST_ERROR", fmt.Sprintf("Failed to create test property: %v", err))
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Central Property State Manager test successful",
		"test_property": property,
	})

	log.Printf("✅ Central Property API: Test property created successfully")
}

// sendErrorResponse sends a standardized error response

// GetAllPublicProperties handles GET /api/v1/central-properties (returns only public properties)
func (cph *CentralPropertyHandler) GetAllPublicProperties(w http.ResponseWriter, r *http.Request ) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080"
	}
	w.Header( ).Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK )
		return
	}

	// Only allow GET method
	if r.Method != "GET" {
		cph.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed" )
		return
	}

	// Get public properties only
	properties, err := cph.stateManager.GetPublicProperties()
	if err != nil {
		log.Printf("Error retrieving public properties: %v", err)
		cph.sendErrorResponse(w, http.StatusInternalServerError, "RETRIEVAL_ERROR", fmt.Sprintf("Failed to retrieve properties: %v", err ))
		return
	}

	log.Printf("✅ Returning %d public properties", len(properties))

	// Send success response
	w.WriteHeader(http.StatusOK )
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"count":      len(properties),
		"properties": properties,
	})
}
func (cph *CentralPropertyHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    false,
		"error_code": errorCode,
		"message":    message,
	})
}

// RegisterCentralPropertyRoutes registers all central property routes
// Deprecated: Use Gin routes instead
func RegisterCentralPropertyRoutes(mux *http.ServeMux, db *gorm.DB, encryptionManager *security.EncryptionManager) {
	handler := NewCentralPropertyHandler(db, encryptionManager)
	
	mux.HandleFunc("/api/v1/central-properties", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler.GetAllPublicProperties(w, r)
		} else {
			handler.CreateOrUpdateProperty(w, r)
		}
	})
	mux.HandleFunc("/api/v1/central-properties/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > len("/api/v1/central-properties/") {
			handler.GetProperty(w, r)
		} else {
			handler.GetPropertiesByStatus(w, r)
		}
	})
	mux.HandleFunc("/api/v1/central-properties/stats", handler.GetSystemStats)
	mux.HandleFunc("/api/v1/central-properties/test", handler.TestCentralState)
	mux.HandleFunc("/api/v1/central-properties/status/", handler.UpdatePropertyStatus)
	
	log.Println("✅ Central Property routes registered successfully")
}

// ============================================================================
// GIN-COMPATIBLE METHODS
// ============================================================================

func (cph *CentralPropertyHandler) GetAllPublicPropertiesGin(c *gin.Context) {
	properties, err := cph.stateManager.GetPublicProperties()
	if err != nil {
		log.Printf("Error retrieving public properties: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error_code": "RETRIEVAL_ERROR", "message": fmt.Sprintf("Failed to retrieve properties: %v", err)})
		return
	}
	log.Printf("✅ Returning %d public properties", len(properties))
	c.JSON(http.StatusOK, gin.H{"success": true, "count": len(properties), "properties": properties})
}

func (cph *CentralPropertyHandler) CreateOrUpdatePropertyGin(c *gin.Context) {
	var updateReq models.PropertyUpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		log.Printf("Error parsing property update request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_JSON", "message": "Invalid JSON in request body"})
		return
	}
	if updateReq.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "MISSING_ADDRESS", "message": "Property address is required"})
		return
	}
	if updateReq.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "MISSING_SOURCE", "message": "Source system is required"})
		return
	}
	property, err := cph.stateManager.CreateOrUpdateProperty(updateReq)
	if err != nil {
		log.Printf("Error creating/updating property: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error_code": "PROPERTY_ERROR", "message": fmt.Sprintf("Failed to process property: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Property processed successfully", "property": property})
	log.Printf("✅ Central Property API: Property %s processed from %s", property.Address, updateReq.Source)
}

func (cph *CentralPropertyHandler) GetPropertyGin(c *gin.Context) {
	identifier := c.Param("id")
	if identifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "MISSING_IDENTIFIER", "message": "Property identifier is required"})
		return
	}
	property, err := cph.stateManager.GetPropertyState(identifier)
	if err != nil {
		log.Printf("Error retrieving property %s: %v", identifier, err)
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error_code": "PROPERTY_NOT_FOUND", "message": fmt.Sprintf("Property not found: %s", identifier)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "property": property})
}

func (cph *CentralPropertyHandler) UpdatePropertyStatusGin(c *gin.Context) {
	mlsID := c.Param("id")
	if mlsID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "MISSING_MLS_ID", "message": "MLS ID is required"})
		return
	}
	var statusReq struct {
		Status string `json:"status"`
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&statusReq); err != nil {
		log.Printf("Error parsing status update request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_JSON", "message": "Invalid JSON in request body"})
		return
	}
	if statusReq.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "MISSING_STATUS", "message": "Status is required"})
		return
	}
	if statusReq.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "MISSING_SOURCE", "message": "Source system is required"})
		return
	}
	if err := cph.stateManager.UpdatePropertyStatus(mlsID, statusReq.Status, statusReq.Source); err != nil {
		log.Printf("Error updating property status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error_code": "STATUS_UPDATE_ERROR", "message": fmt.Sprintf("Failed to update status: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Status updated to %s for property %s", statusReq.Status, mlsID)})
	log.Printf("✅ Central Property API: Status updated for %s to %s (Source: %s)", mlsID, statusReq.Status, statusReq.Source)
}

func (cph *CentralPropertyHandler) GetSystemStatsGin(c *gin.Context) {
	stats, err := cph.stateManager.GetSystemStats()
	if err != nil {
		log.Printf("Error retrieving system stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error_code": "STATS_ERROR", "message": fmt.Sprintf("Failed to retrieve stats: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "stats": stats})
}

func (cph *CentralPropertyHandler) TestCentralStateGin(c *gin.Context) {
	testProperty := models.PropertyUpdateRequest{
		MLSId: "TEST_MLS_12345", Address: "123 Test Street, Houston, TX 77001",
		Price: &[]float64{350000}[0], Bedrooms: &[]int{3}[0], Bathrooms: &[]float32{2.5}[0],
		SquareFeet: &[]int{1800}[0], PropertyType: "Single Family", Status: "active", Source: "test",
		Data: models.JSONB{"test_mode": true, "created_by": "central_property_test"},
	}
	property, err := cph.stateManager.CreateOrUpdateProperty(testProperty)
	if err != nil {
		log.Printf("Error creating test property: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error_code": "TEST_ERROR", "message": fmt.Sprintf("Failed to create test property: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Central Property State Manager test successful", "test_property": property})
	log.Printf("✅ Central Property API: Test property created successfully")
}
