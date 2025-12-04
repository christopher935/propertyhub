package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
)

// CreatePropertyRequest represents the request body for creating a property
type CreatePropertyRequest struct {
	MLSId             string   `json:"mls_id"`
	Address           string   `json:"address"`
	City              string   `json:"city"`
	State             string   `json:"state"`
	ZipCode           string   `json:"zip_code"`
	Price             float64  `json:"price"`
	Bedrooms          *int     `json:"bedrooms"`
	Bathrooms         *float32 `json:"bathrooms"`
	SquareFeet        *int     `json:"square_feet"`
	PropertyType      string   `json:"property_type"`
	Description       string   `json:"description"`
	Images            []string `json:"images"`
	Status            string   `json:"status"`
	ListingType       string   `json:"listing_type"`
	ListingAgent      string   `json:"listing_agent"`
	ListingOffice     string   `json:"listing_office"`
	YearBuilt         int      `json:"year_built"`
	ManagementCompany string   `json:"management_company"`
	Source            string   `json:"source"`
}

// UpdatePropertyRequest represents the request body for updating a property
type UpdatePropertyRequest struct {
	Address           *string  `json:"address,omitempty"`
	City              *string  `json:"city,omitempty"`
	State             *string  `json:"state,omitempty"`
	ZipCode           *string  `json:"zip_code,omitempty"`
	Price             *float64 `json:"price,omitempty"`
	Bedrooms          *int     `json:"bedrooms,omitempty"`
	Bathrooms         *float32 `json:"bathrooms,omitempty"`
	SquareFeet        *int     `json:"square_feet,omitempty"`
	PropertyType      *string  `json:"property_type,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Images            []string `json:"images,omitempty"`
	Status            *string  `json:"status,omitempty"`
	ListingType       *string  `json:"listing_type,omitempty"`
	ListingAgent      *string  `json:"listing_agent,omitempty"`
	ListingOffice     *string  `json:"listing_office,omitempty"`
	YearBuilt         *int     `json:"year_built,omitempty"`
	ManagementCompany *string  `json:"management_company,omitempty"`
	Source            *string  `json:"source,omitempty"`
}

// PropertyCRUDHandler handles property CRUD operations
type PropertyCRUDHandler struct {
	db *gorm.DB
}

func NewPropertyCRUDHandler(db *gorm.DB) *PropertyCRUDHandler {
	return &PropertyCRUDHandler{
		db: db,
	}
}

// CreateProperty handles POST /api/v1/admin/properties
func (h *PropertyCRUDHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	var req CreatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Address == "" || req.City == "" || req.State == "" || req.Price <= 0 {
		http.Error(w, "Missing required fields: address, city, state, price", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Status == "" {
		req.Status = "available"
	}
	if req.Source == "" {
		req.Source = "admin"
	}
	if req.ListingType == "" {
		req.ListingType = "for_rent"
	}

	// Create property (fix Address type conversion)
	property := models.Property{
		MLSId:             req.MLSId,
		Address:           security.EncryptedString(req.Address), // Convert string to EncryptedString
		City:              req.City,
		State:             req.State,
		ZipCode:           req.ZipCode,
		Price:             req.Price,
		Bedrooms:          req.Bedrooms,
		Bathrooms:         req.Bathrooms,
		SquareFeet:        req.SquareFeet,
		PropertyType:      req.PropertyType,
		Description:       req.Description,
		Status:            req.Status,
		ListingAgent:      req.ListingAgent,
		ListingOffice:     req.ListingOffice,
		YearBuilt:         req.YearBuilt,
		ManagementCompany: req.ManagementCompany,
		Source:            req.Source,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.db.Create(&property).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Property with this MLS ID already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create property", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property created successfully",
		"data": map[string]interface{}{
			"property": property,
		},
	})
}

// UpdateProperty handles PUT /api/v1/admin/properties/:id
func (h *PropertyCRUDHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	// Extract property ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	propertyID, err := strconv.ParseUint(pathParts[4], 10, 32)
	if err != nil {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	var req UpdatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find existing property
	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Property not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to find property", http.StatusInternalServerError)
		return
	}

	// Update fields if provided (fix Address type conversion)
	if req.Address != nil {
		property.Address = security.EncryptedString(*req.Address) // Convert string to EncryptedString
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
	if req.Price != nil {
		property.Price = *req.Price
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
	if req.Description != nil {
		property.Description = *req.Description
	}
	if req.Status != nil {
		property.Status = *req.Status
	}
	if req.ListingAgent != nil {
		property.ListingAgent = *req.ListingAgent
	}
	if req.ListingOffice != nil {
		property.ListingOffice = *req.ListingOffice
	}
	if req.YearBuilt != nil {
		property.YearBuilt = *req.YearBuilt
	}
	if req.ManagementCompany != nil {
		property.ManagementCompany = *req.ManagementCompany
	}

	property.UpdatedAt = time.Now()

	if err := h.db.Save(&property).Error; err != nil {
		http.Error(w, "Failed to update property", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property updated successfully",
		"data": map[string]interface{}{
			"property": property,
		},
	})
}

// DeleteProperty handles DELETE /api/v1/admin/properties/:id
func (h *PropertyCRUDHandler) DeleteProperty(w http.ResponseWriter, r *http.Request) {
	// Extract property ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	propertyID, err := strconv.ParseUint(pathParts[4], 10, 32)
	if err != nil {
		http.Error(w, "Invalid property ID", http.StatusBadRequest)
		return
	}

	// Check if property exists
	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Property not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to find property", http.StatusInternalServerError)
		return
	}

	// Check for related bookings
	var bookingCount int64
	h.db.Model(&models.Booking{}).Where("property_id = ?", propertyID).Count(&bookingCount)

	if bookingCount > 0 {
		// Soft delete - change status instead of removing
		property.Status = "deleted"
		property.UpdatedAt = time.Now()

		if err := h.db.Save(&property).Error; err != nil {
			http.Error(w, "Failed to delete property", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Property marked as deleted (has existing bookings)",
			"data": map[string]interface{}{
				"property_id": propertyID,
				"soft_delete": true,
			},
		})
		return
	}

	// Hard delete if no bookings
	if err := h.db.Delete(&property).Error; err != nil {
		http.Error(w, "Failed to delete property", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property deleted successfully",
		"data": map[string]interface{}{
			"property_id": propertyID,
			"soft_delete": false,
		},
	})
}

// GetPropertyByID handles GET /api/v1/admin/properties/:id
func (h *PropertyCRUDHandler) GetPropertyByID(w http.ResponseWriter, r *http.Request) {
	// Extract property ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	propertyID, err := strconv.ParseUint(pathParts[4], 10, 32)
	if err != nil {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"property": property,
		},
	})
}

// RegisterPropertyCRUDRoutes registers property CRUD routes
func RegisterPropertyCRUDRoutes(mux *http.ServeMux, db *gorm.DB) {
	handler := NewPropertyCRUDHandler(db)

	// Property CRUD routes
	mux.HandleFunc("/api/v1/admin/properties", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handler.CreateProperty(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/admin/properties/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handler.GetPropertyByID(w, r)
		case "PUT":
			handler.UpdateProperty(w, r)
		case "DELETE":
			handler.DeleteProperty(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("🏠 Property CRUD routes registered")
}
