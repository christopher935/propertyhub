package models

import (
	"time"

	"chrisgross-ctrl-project/internal/security"
)

// PropertyResponse is the API response format with decrypted fields
type PropertyResponse struct {
	ID                uint     `json:"id"`
	MLSId             string   `json:"mls_id"`
	Address           string   `json:"address"` // Decrypted
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
	ViewCount         int        `json:"view_count"`
	DaysOnMarket      *int       `json:"days_on_market"`
	YearBuilt         int        `json:"year_built"`
	ManagementCompany string     `json:"management_company"`
	PetFriendly       bool       `json:"pet_friendly"`
	Stories           int        `json:"stories"`
	AvailableDate     *time.Time `json:"available_date"`
}

// ToResponse converts a Property to PropertyResponse with decrypted address
func ToResponse(property Property, encryptionManager *security.EncryptionManager) PropertyResponse {
	// Decrypt the address
	decryptedAddress, err := encryptionManager.Decrypt(property.Address)
	if err != nil {
		decryptedAddress = "[DECRYPTION ERROR]"
	}

	return PropertyResponse{
		ID:                property.ID,
		MLSId:             property.MLSId,
		Address:           decryptedAddress,
		City:              property.City,
		State:             property.State,
		ZipCode:           property.ZipCode,
		Bedrooms:          property.Bedrooms,
		Bathrooms:         property.Bathrooms,
		SquareFeet:        property.SquareFeet,
		PropertyType:      property.PropertyType,
		Price:             property.Price,
		ListingType:       property.ListingType,
		Status:            property.Status,
		Description:       property.Description,
		Images:            property.Images,
		FeaturedImage:     property.FeaturedImage,
		ListingAgent:      property.ListingAgent,
		ListingAgentID:    property.ListingAgentID,
		ListingOffice:     property.ListingOffice,
		PropertyFeatures:  property.PropertyFeatures,
		Source:            property.Source,
		HarUrl:            property.HarUrl,
		ViewCount:         property.ViewCount,
		DaysOnMarket:      property.DaysOnMarket,
		YearBuilt:         property.YearBuilt,
		ManagementCompany: property.ManagementCompany,
		PetFriendly:       property.PetFriendly,
		Stories:           property.Stories,
		AvailableDate:     property.AvailableDate,
	}
}

// ToResponseList converts a slice of Properties to PropertyResponses with decrypted addresses
func ToResponseList(properties []Property, encryptionManager *security.EncryptionManager) []PropertyResponse {
	responses := make([]PropertyResponse, len(properties))
	for i, property := range properties {
		responses[i] = ToResponse(property, encryptionManager)
	}
	return responses
}
