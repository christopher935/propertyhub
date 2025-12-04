package handlers

import (
	"fmt"
	"chrisgross-ctrl-project/internal/models"
	"time"
)

// PublicPropertyResponse represents a property in public API responses
type PublicPropertyResponse struct {
	ID          string                `json:"id"`
	Address     string                `json:"address"`
	City        string                `json:"city"`
	State       string                `json:"state"`
	ZipCode     string                `json:"zip_code"`
	Price       float64               `json:"price,omitempty"`
	Bedrooms    int                   `json:"bedrooms,omitempty"`
	Bathrooms   int                   `json:"bathrooms,omitempty"`
	SquareFeet  int                   `json:"square_feet,omitempty"`
	Description string                `json:"description,omitempty"`
	Features    []string              `json:"features,omitempty"`
	Images      []PublicImageResponse `json:"images,omitempty"`
	Status      string                `json:"status"`
	// Excluded: CreatedAt, UpdatedAt, DeletedAt, InternalNotes, etc.
}

// PublicImageResponse represents a property image in public responses
type PublicImageResponse struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	AltText      string `json:"alt_text,omitempty"`
	DisplayOrder int    `json:"display_order"`
	// Excluded: InternalPath, UploadedBy, CreatedAt, etc.
}

// PublicBookingResponse represents a booking in public API responses
type PublicBookingResponse struct {
	ID              string    `json:"id"`
	ReferenceNumber string    `json:"reference_number"`
	PropertyID      string    `json:"property_id"`
	Status          string    `json:"status"`
	ScheduledDate   time.Time `json:"scheduled_date,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	// Excluded: CreatedAt, UpdatedAt, InternalNotes, AgentInfo, etc.
}

// PublicContactResponse represents a contact in public API responses
type PublicContactResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	// Excluded: Name, Phone, Email, Message, InternalNotes, etc.
}

// AdminPropertyResponse represents a property in admin API responses
type AdminPropertyResponse struct {
	ID          string                `json:"id"`
	Address     string                `json:"address"`
	City        string                `json:"city"`
	State       string                `json:"state"`
	ZipCode     string                `json:"zip_code"`
	Price       float64               `json:"price,omitempty"`
	Bedrooms    int                   `json:"bedrooms,omitempty"`
	Bathrooms   int                   `json:"bathrooms,omitempty"`
	SquareFeet  int                   `json:"square_feet,omitempty"`
	Description string                `json:"description,omitempty"`
	Features    []string              `json:"features,omitempty"`
	Images      []PublicImageResponse `json:"images,omitempty"`
	Status      string                `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	// Still excluded: DeletedAt, some sensitive internal fields
}

// ResponseFilter provides methods to filter sensitive data from responses
type ResponseFilter struct{}

// NewResponseFilter creates a new response filter
func NewResponseFilter() *ResponseFilter {
	return &ResponseFilter{}
}

// FilterProperty converts a property model to public response
func (rf *ResponseFilter) FilterProperty(property *models.Property) PublicPropertyResponse {
	// Convert types safely
	var bedrooms, bathrooms, squareFeet int
	var features []string

	if property.Bedrooms != nil {
		bedrooms = *property.Bedrooms
	}
	if property.Bathrooms != nil {
		bathrooms = int(*property.Bathrooms)
	}
	if property.SquareFeet != nil {
		squareFeet = *property.SquareFeet
	}

	// Convert PropertyFeatures string to slice if not empty
	if property.PropertyFeatures != "" {
		features = []string{property.PropertyFeatures}
	} else {
		features = []string{}
	}

	return PublicPropertyResponse{
		ID:          fmt.Sprintf("%d", property.ID),     // Convert uint to string
		Address:     "[Address Available Upon Request]", // Don't expose encrypted address
		City:        property.City,
		State:       property.State,
		ZipCode:     property.ZipCode,
		Price:       property.Price,
		Bedrooms:    bedrooms,
		Bathrooms:   bathrooms,
		SquareFeet:  squareFeet,
		Description: property.Description,
		Features:    features,
		Status:      property.Status,
	}
}

// FilterPropertyForAdmin converts a property model to admin response
func (rf *ResponseFilter) FilterPropertyForAdmin(property *models.Property) AdminPropertyResponse {
	// Convert types safely for admin view
	var bedrooms, bathrooms, squareFeet int
	var features []string

	if property.Bedrooms != nil {
		bedrooms = *property.Bedrooms
	}
	if property.Bathrooms != nil {
		bathrooms = int(*property.Bathrooms)
	}
	if property.SquareFeet != nil {
		squareFeet = *property.SquareFeet
	}

	// Convert PropertyFeatures string to slice
	if property.PropertyFeatures != "" {
		features = []string{property.PropertyFeatures}
	} else {
		features = []string{}
	}

	return AdminPropertyResponse{
		ID:          fmt.Sprintf("%d", property.ID), // Convert uint to string
		Address:     "[Admin Access - Encrypted Address]",
		City:        property.City,
		State:       property.State,
		ZipCode:     property.ZipCode,
		Price:       property.Price,
		Bedrooms:    bedrooms,
		Bathrooms:   bathrooms,
		SquareFeet:  squareFeet,
		Description: property.Description,
		Features:    features,
		Status:      property.Status,
		CreatedAt:   property.CreatedAt,
		UpdatedAt:   property.UpdatedAt,
	}
}

// FilterBooking converts a booking model to public response
func (rf *ResponseFilter) FilterBooking(booking *models.Booking) PublicBookingResponse {
	return PublicBookingResponse{
		ID:              fmt.Sprintf("%d", booking.ID), // Convert uint to string
		ReferenceNumber: booking.ReferenceNumber,
		PropertyID:      fmt.Sprintf("%d", booking.PropertyID), // Convert uint to string
		Status:          booking.Status,
		ScheduledDate:   booking.ShowingDate, // Map ShowingDate to ScheduledDate
		Notes:           booking.Notes,       // Use regular notes field
	}
}

// FilterContact converts a contact model to public response (minimal info)
func (rf *ResponseFilter) FilterContact(contact *models.Contact) PublicContactResponse {
	return PublicContactResponse{
		ID:        fmt.Sprintf("%d", contact.ID), // Convert uint to string
		Status:    contact.Status,
		CreatedAt: contact.CreatedAt,
	}
}

// FilterProperties converts a slice of properties to public responses
func (rf *ResponseFilter) FilterProperties(properties []models.Property) []PublicPropertyResponse {
	filtered := make([]PublicPropertyResponse, len(properties))
	for i, property := range properties {
		filtered[i] = rf.FilterProperty(&property)
	}
	return filtered
}

// FilterPropertiesForAdmin converts a slice of properties to admin responses
func (rf *ResponseFilter) FilterPropertiesForAdmin(properties []models.Property) []AdminPropertyResponse {
	filtered := make([]AdminPropertyResponse, len(properties))
	for i, property := range properties {
		filtered[i] = rf.FilterPropertyForAdmin(&property)
	}
	return filtered
}

// RemoveSensitiveFields removes sensitive fields from generic map responses
func (rf *ResponseFilter) RemoveSensitiveFields(data map[string]interface{}) map[string]interface{} {
	// List of sensitive fields to always remove
	sensitiveFields := []string{
		"password", "password_hash", "secret", "token", "key",
		"internal_notes", "private_notes", "system_notes",
		"deleted_at", "admin_notes", "ip_address", "user_agent",
		"session_id", "auth_token", "api_key", "encryption_key",
	}

	filtered := make(map[string]interface{})
	for key, value := range data {
		// Skip sensitive fields
		sensitive := false
		for _, field := range sensitiveFields {
			if key == field {
				sensitive = true
				break
			}
		}

		if !sensitive {
			filtered[key] = value
		}
	}

	return filtered
}
