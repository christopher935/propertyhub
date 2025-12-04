package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/repositories"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"
)

type PropertiesHandler struct {
	db                *gorm.DB
	repos             *repositories.Repositories
	encryptionManager *security.EncryptionManager
	behavioralService *services.BehavioralEventService  // ADDED: Behavioral tracking
}

func NewPropertiesHandler(db *gorm.DB, repos *repositories.Repositories, encryptionManager *security.EncryptionManager) *PropertiesHandler {
	return &PropertiesHandler{
		db:                db,
		repos:             repos,
		encryptionManager: encryptionManager,
		behavioralService: services.NewBehavioralEventService(db),  // ADDED: Initialize tracking service
	}
}

type PropertyStatsResponse struct {
	TotalProperties    int64   `json:"total_properties"`
	ActiveProperties   int64   `json:"active_properties"`
	PendingProperties  int64   `json:"pending_properties"`
	SoldProperties     int64   `json:"sold_properties"`
	AveragePrice      float64 `json:"average_price"`
	TotalValue        float64 `json:"total_value"`
	ViewsLast30Days   int     `json:"views_last_30_days"`
}

func (h *PropertiesHandler) GetProperties(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	status := r.URL.Query().Get("status")
	city := r.URL.Query().Get("city")
	search := r.URL.Query().Get("search")

	query := h.db.Model(&models.Property{})

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	} else {
		adminStatuses := []string{"active", "pending_images", "available"}
		query = query.Where("status IN (?)", adminStatuses)
	}

	if city != "" && city != "all" {
		query = query.Where("city = ?", city)
	}

	if search != "" {
		query = query.Where("address ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var properties []models.Property
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&properties).Error

	if err != nil {
		log.Printf("Error fetching admin properties: %v", err)
		http.Error(w, "Failed to fetch properties", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"properties":  properties,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *PropertiesHandler) GetConsumerProperties(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	city := r.URL.Query().Get("city")
	search := r.URL.Query().Get("search")

	query := h.db.Model(&models.Property{}).Where("status = ?", "active")

	if city != "" && city != "all" {
		query = query.Where("city = ?", city)
	}

	if search != "" {
		query = query.Where("address ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var totalCount int64
	query.Count(&totalCount)

	var properties []models.Property
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&properties).Error

	if err != nil {
		log.Printf("Error fetching consumer properties: %v", err)
		http.Error(w, "Failed to fetch properties", http.StatusInternalServerError)
		return
	}

	// ============ ADDED: BEHAVIORAL TRACKING ============
	// Track property browsing behavior
	if leadID := extractLeadID(r); leadID > 0 {
		sessionID := extractSessionID(r)
		ipAddress := extractIPAddress(r)
		userAgent := r.UserAgent()
		
		eventData := map[string]interface{}{
			"action":       "browse_properties",
			"page":         page,
			"city_filter":  city,
			"search_term":  search,
			"result_count": len(properties),
		}
		
		// Track the browsing event (non-blocking)
		go h.behavioralService.TrackEvent(leadID, "browsed", eventData, nil, sessionID, ipAddress, userAgent)
	}
	// ============ END TRACKING ============

	// Convert to response format with decrypted addresses
	propertyResponses := models.ToResponseList(properties, h.encryptionManager)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": propertyResponses,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  int((totalCount + int64(limit) - 1) / int64(limit)),
			"total_count":  totalCount,
			"per_page":     limit,
		},
	})
}

func (h *PropertiesHandler) GetPropertyStats(w http.ResponseWriter, r *http.Request) {
	var stats PropertyStatsResponse

	availableStatuses := []string{"active", "pending_images", "available"}

	h.db.Model(&models.Property{}).Count(&stats.TotalProperties)
	h.db.Model(&models.Property{}).Where("status IN (?)", availableStatuses).Count(&stats.ActiveProperties)
	h.db.Model(&models.Property{}).Where("status = ?", "pending").Count(&stats.PendingProperties)
	h.db.Model(&models.Property{}).Where("status = ?", "sold").Count(&stats.SoldProperties)

	h.db.Model(&models.Property{}).Where("status IN (?)", availableStatuses).Select("COALESCE(AVG(price), 0)").Scan(&stats.AveragePrice)
	h.db.Model(&models.Property{}).Where("status IN (?)", availableStatuses).Select("COALESCE(SUM(price), 0)").Scan(&stats.TotalValue)

	stats.ViewsLast30Days = 1247

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

func (h *PropertiesHandler) GetFeaturedProperties(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 3
	}

	var properties []models.Property
	availableStatuses := []string{"active", "pending_images", "available"}
	err := h.db.Where("status IN (?)", availableStatuses).Limit(limit).Find(&properties).Error

	if err != nil {
		log.Printf("Error fetching featured properties: %v", err)
		http.Error(w, "Failed to fetch featured properties", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"properties": properties,
			"count":      len(properties),
		},
	})
}

func (h *PropertiesHandler) UploadPropertyPhoto(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	propertyIDStr := pathParts[5]
	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Photo upload endpoint working",
		"property_id": propertyID,
	})
}

func (h *PropertiesHandler) PromotePropertyToActive(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	propertyIDStr := pathParts[len(pathParts)-2]
	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	if property.Status != "pending_images" {
		http.Error(w, "Property is not in pending_images status", http.StatusBadRequest)
		return
	}

	property.Status = "active"
	if err := h.db.Save(&property).Error; err != nil {
		log.Printf("Error updating property status: %v", err)
		http.Error(w, "Failed to update property status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property activated successfully",
		"property": map[string]interface{}{
			"id":     property.ID,
			"status": property.Status,
		},
	})
}

func (h *PropertiesHandler) UpdatePropertyStatus(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	var propertyID uint64
	var err error
	
	for i, part := range pathParts {
		if part == "properties" && i+1 < len(pathParts) {
			propertyID, err = strconv.ParseUint(pathParts[i+1], 10, 32)
			break
		}
	}
	
	if err != nil || propertyID == 0 {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validStatuses := []string{"active", "pending_images", "available", "pending", "sold", "withdrawn", "deleted"}
	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Property not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to find property", http.StatusInternalServerError)
		return
	}

	property.Status = req.Status
	property.UpdatedAt = time.Now()

	if err := h.db.Save(&property).Error; err != nil {
		http.Error(w, "Failed to update property status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property status updated successfully",
		"data": map[string]interface{}{
			"property_id": propertyID,
			"status":      property.Status,
		},
	})
}

// ============================================================================
// BEHAVIORAL TRACKING HELPER FUNCTIONS
// ============================================================================

// extractLeadID gets lead_id from request context or cookie
func extractLeadID(r *http.Request) int64 {
	// Try context first (authenticated user)
	if leadIDVal := r.Context().Value("lead_id"); leadIDVal != nil {
		if id, ok := leadIDVal.(int64); ok {
			return id
		}
		if id, ok := leadIDVal.(uint); ok {
			return int64(id)
		}
	}

	// Try cookie (visitor tracking)
	if cookie, err := r.Cookie("lead_id"); err == nil {
		if id, err := strconv.ParseInt(cookie.Value, 10, 64); err == nil {
			return id
		}
	}

	return 0 // Anonymous visitor
}

// extractSessionID gets session_id from cookie
func extractSessionID(r *http.Request) string {
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}
	return ""
}

// extractIPAddress gets real IP address from request
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func RegisterPropertiesRoutes(mux *http.ServeMux, db *gorm.DB, repos *repositories.Repositories, encryptionManager *security.EncryptionManager) {
	handler := NewPropertiesHandler(db, repos, encryptionManager)

	mux.HandleFunc("/api/v1/admin/properties", handler.GetProperties)
	mux.HandleFunc("/api/v1/admin/properties/stats", handler.GetPropertyStats)
	mux.HandleFunc("/api/v1/admin/properties/featured", handler.GetFeaturedProperties)
	
	mux.HandleFunc("/api/v1/admin/properties/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/status") {
			handler.UpdatePropertyStatus(w, r)
		} else if strings.Contains(r.URL.Path, "/photos") && r.Method == "POST" {
			handler.UploadPropertyPhoto(w, r)
		} else if strings.Contains(r.URL.Path, "/activate") && r.Method == "PUT" {
			handler.PromotePropertyToActive(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/api/v1/properties", handler.GetConsumerProperties)

	log.Println("🏠 Properties management routes registered (WITH BEHAVIORAL TRACKING)")
}

// GetPropertyByID returns a single property by ID with decrypted address
func (h *PropertiesHandler) GetPropertyByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	var propertyIDStr string
	for i, part := range pathParts {
		if part == "properties" && i+1 < len(pathParts) {
			propertyIDStr = pathParts[i+1]
			break
		}
	}

	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 32)
	if err != nil || propertyID == 0 {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Property not found", http.StatusNotFound)
			return
		}
		log.Printf("Error fetching property %d: %v", propertyID, err)
		http.Error(w, "Failed to fetch property", http.StatusInternalServerError)
		return
	}

	// Convert to response format with decrypted address
	propertyResponse := models.ToResponse(property, h.encryptionManager)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"property": propertyResponse,
		},
	})
}

// CreateProperty creates a new property with encrypted address
func (h *PropertiesHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MLSId             string   `json:"mls_id"`
		Address           string   `json:"address"`
		City              string   `json:"city"`
		State             string   `json:"state"`
		ZipCode           string   `json:"zip_code"`
		Bedrooms          *int     `json:"bedrooms"`
		Bathrooms         *float32 `json:"bathrooms"`
		SquareFeet        *int     `json:"square_feet"`
		PropertyType      string   `json:"property_type"`
		Price             float64  `json:"price"`
		ListingType       string   `json:"listing_type"`
		Status            string   `json:"status"`
		Description       string   `json:"description"`
		Images            []string `json:"images"`
		FeaturedImage     string   `json:"featured_image"`
		ListingAgent      string   `json:"listing_agent"`
		ListingAgentID    string   `json:"listing_agent_id"`
		ListingOffice     string   `json:"listing_office"`
		PropertyFeatures  string   `json:"property_features"`
		Source            string   `json:"source"`
		HarUrl            string   `json:"har_url"`
		YearBuilt         int      `json:"year_built"`
		ManagementCompany string   `json:"management_company"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Encrypt the address
	encryptedAddress, err := h.encryptionManager.Encrypt(req.Address)
	if err != nil {
		log.Printf("Error encrypting address: %v", err)
		http.Error(w, "Failed to encrypt address", http.StatusInternalServerError)
		return
	}

	property := models.Property{
		MLSId:             req.MLSId,
		Address:           encryptedAddress,
		City:              req.City,
		State:             req.State,
		ZipCode:           req.ZipCode,
		Bedrooms:          req.Bedrooms,
		Bathrooms:         req.Bathrooms,
		SquareFeet:        req.SquareFeet,
		PropertyType:      req.PropertyType,
		Price:             req.Price,
		ListingType:       req.ListingType,
		Status:            req.Status,
		Description:       req.Description,
		Images:            req.Images,
		FeaturedImage:     req.FeaturedImage,
		ListingAgent:      req.ListingAgent,
		ListingAgentID:    req.ListingAgentID,
		ListingOffice:     req.ListingOffice,
		PropertyFeatures:  req.PropertyFeatures,
		Source:            req.Source,
		HarUrl:            req.HarUrl,
		YearBuilt:         req.YearBuilt,
		ManagementCompany: req.ManagementCompany,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.db.Create(&property).Error; err != nil {
		log.Printf("Error creating property: %v", err)
		http.Error(w, "Failed to create property", http.StatusInternalServerError)
		return
	}

	// Return decrypted response
	propertyResponse := models.ToResponse(property, h.encryptionManager)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property created successfully",
		"data": map[string]interface{}{
			"property": propertyResponse,
		},
	})
}

// UpdateProperty updates an existing property with encrypted address
func (h *PropertiesHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	var propertyIDStr string
	for i, part := range pathParts {
		if part == "properties" && i+1 < len(pathParts) {
			propertyIDStr = pathParts[i+1]
			break
		}
	}

	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 32)
	if err != nil || propertyID == 0 {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Property not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to find property", http.StatusInternalServerError)
		return
	}

	var req struct {
		MLSId             *string   `json:"mls_id"`
		Address           *string   `json:"address"`
		City              *string   `json:"city"`
		State             *string   `json:"state"`
		ZipCode           *string   `json:"zip_code"`
		Bedrooms          *int      `json:"bedrooms"`
		Bathrooms         *float32  `json:"bathrooms"`
		SquareFeet        *int      `json:"square_feet"`
		PropertyType      *string   `json:"property_type"`
		Price             *float64  `json:"price"`
		ListingType       *string   `json:"listing_type"`
		Status            *string   `json:"status"`
		Description       *string   `json:"description"`
		Images            *[]string `json:"images"`
		FeaturedImage     *string   `json:"featured_image"`
		ListingAgent      *string   `json:"listing_agent"`
		ListingAgentID    *string   `json:"listing_agent_id"`
		ListingOffice     *string   `json:"listing_office"`
		PropertyFeatures  *string   `json:"property_features"`
		Source            *string   `json:"source"`
		HarUrl            *string   `json:"har_url"`
		YearBuilt         *int      `json:"year_built"`
		ManagementCompany *string   `json:"management_company"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if req.MLSId != nil {
		property.MLSId = *req.MLSId
	}
	if req.Address != nil {
		// Encrypt the new address
		encryptedAddress, err := h.encryptionManager.Encrypt(*req.Address)
		if err != nil {
			log.Printf("Error encrypting address: %v", err)
			http.Error(w, "Failed to encrypt address", http.StatusInternalServerError)
			return
		}
		property.Address = encryptedAddress
	}
	if req.City != nil {
		property.City = *req.City
	}
	if req.State != nil {
		property.State = *req.State
	}
	if req.ZipCode != nil {
		property.ZipCode = *req.ZipCode
	}
	if req.Bedrooms != nil {
		property.Bedrooms = req.Bedrooms
	}
	if req.Bathrooms != nil {
		property.Bathrooms = req.Bathrooms
	}
	if req.SquareFeet != nil {
		property.SquareFeet = req.SquareFeet
	}
	if req.PropertyType != nil {
		property.PropertyType = *req.PropertyType
	}
	if req.Price != nil {
		property.Price = *req.Price
	}
	if req.ListingType != nil {
		property.ListingType = *req.ListingType
	}
	if req.Status != nil {
		property.Status = *req.Status
	}
	if req.Description != nil {
		property.Description = *req.Description
	}
	if req.Images != nil {
		property.Images = *req.Images
	}
	if req.FeaturedImage != nil {
		property.FeaturedImage = *req.FeaturedImage
	}
	if req.ListingAgent != nil {
		property.ListingAgent = *req.ListingAgent
	}
	if req.ListingAgentID != nil {
		property.ListingAgentID = *req.ListingAgentID
	}
	if req.ListingOffice != nil {
		property.ListingOffice = *req.ListingOffice
	}
	if req.PropertyFeatures != nil {
		property.PropertyFeatures = *req.PropertyFeatures
	}
	if req.Source != nil {
		property.Source = *req.Source
	}
	if req.HarUrl != nil {
		property.HarUrl = *req.HarUrl
	}
	if req.YearBuilt != nil {
		property.YearBuilt = *req.YearBuilt
	}
	if req.ManagementCompany != nil {
		property.ManagementCompany = *req.ManagementCompany
	}

	property.UpdatedAt = time.Now()

	if err := h.db.Save(&property).Error; err != nil {
		log.Printf("Error updating property: %v", err)
		http.Error(w, "Failed to update property", http.StatusInternalServerError)
		return
	}

	// Return decrypted response
	propertyResponse := models.ToResponse(property, h.encryptionManager)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property updated successfully",
		"data": map[string]interface{}{
			"property": propertyResponse,
		},
	})
}

// DeleteProperty soft deletes a property
func (h *PropertiesHandler) DeleteProperty(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	var propertyIDStr string
	for i, part := range pathParts {
		if part == "properties" && i+1 < len(pathParts) {
			propertyIDStr = pathParts[i+1]
			break
		}
	}

	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 32)
	if err != nil || propertyID == 0 {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Property not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to find property", http.StatusInternalServerError)
		return
	}

	// Soft delete by setting status to deleted
	property.Status = "deleted"
	property.UpdatedAt = time.Now()

	if err := h.db.Save(&property).Error; err != nil {
		log.Printf("Error deleting property: %v", err)
		http.Error(w, "Failed to delete property", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property deleted successfully",
		"data": map[string]interface{}{
			"property_id": propertyID,
		},
	})
}

// GetPropertiesGin returns properties list for admin dashboard
func (h *PropertiesHandler) GetPropertiesGin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	status := c.Query("status")
	propertyType := c.Query("property_type")
	bedrooms := c.Query("bedrooms")
	bathrooms := c.Query("bathrooms")
	search := c.Query("search")

	query := h.db.Model(&models.Property{})

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	} else {
		adminStatuses := []string{"active", "pending_images", "available"}
		query = query.Where("status IN (?)", adminStatuses)
	}

	if propertyType != "" && propertyType != "all" {
		query = query.Where("property_type = ?", propertyType)
	}

	if bedrooms != "" && bedrooms != "any" {
		if beds, err := strconv.Atoi(strings.TrimSuffix(bedrooms, "+")); err == nil {
			query = query.Where("bedrooms >= ?", beds)
		}
	}

	if bathrooms != "" && bathrooms != "any" {
		if baths, err := strconv.Atoi(strings.TrimSuffix(bathrooms, "+")); err == nil {
			query = query.Where("bathrooms >= ?", baths)
		}
	}

	if search != "" {
		query = query.Where("address ILIKE ? OR description ILIKE ? OR city ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var properties []models.Property
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&properties).Error

	if err != nil {
		log.Printf("Error fetching properties: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch properties",
		})
		return
	}

	stats := map[string]interface{}{
		"total":     total,
		"active":    0,
		"pending":   0,
		"available": 0,
	}

	for _, prop := range properties {
		switch prop.Status {
		case "active":
			stats["active"] = stats["active"].(int) + 1
		case "pending_images":
			stats["pending"] = stats["pending"].(int) + 1
		case "available":
			stats["available"] = stats["available"].(int) + 1
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"properties":  properties,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"stats": stats,
	})
}
