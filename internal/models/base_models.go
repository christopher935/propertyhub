package models

import (
	"chrisgross-ctrl-project/internal/security"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Property - Core property model for real estate listings
type Property struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Basic property information
	MLSId   string                   `json:"mls_id" gorm:"uniqueIndex"`
	Address security.EncryptedString `json:"address" gorm:"not null"`
	City    string                   `json:"city" gorm:"default:'Houston'"`
	State   string                   `json:"state" gorm:"default:'TX'"`
	ZipCode string                   `json:"zip_code"`

	// Property details
	Bedrooms     *int     `json:"bedrooms"`
	Bathrooms    *float32 `json:"bathrooms"`
	SquareFeet   *int     `json:"square_feet"`
	PropertyType string   `json:"property_type"`

	// Listing information
	Price       float64 `json:"price"`
	ListingType string  `json:"listing_type" gorm:"default:'For Sale'"`
	Status      string  `json:"status" gorm:"default:'active'"`
	Description string  `json:"description"`

	// Images and media
	Images        pq.StringArray `json:"images" gorm:"type:text[]"`
	FeaturedImage string         `json:"featured_image"`

	// Agent and office information
	ListingAgent     string `json:"listing_agent" gorm:"default:'Christopher Gross'"`
	ListingAgentID   string `json:"listing_agent_id"`
	ListingOffice    string `json:"listing_office" gorm:"default:'Landlords of Texas, LLC'"`
	PropertyFeatures string `json:"property_features"`

	// Source and URLs
	Source string `json:"source" gorm:"default:'PropertyHub'"`
	HarUrl string `json:"har_url"`

	// Basic analytics
	ViewCount int `json:"view_count" gorm:"default:0"`

	// Additional fields for advanced functionality
	DaysOnMarket      *int       `json:"days_on_market"`
	ScrapedAt         *time.Time `json:"scraped_at"`
	YearBuilt         int        `json:"year_built"`
	ManagementCompany string     `json:"management_company"`

	// Filter-specific fields
	PetFriendly   bool       `json:"pet_friendly" gorm:"default:false"`
	Stories       int        `json:"stories" gorm:"default:1"`
	AvailableDate *time.Time `json:"available_date"`

	// Timestamps
	DateAdded *time.Time     `json:"date_added"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Bookings []Booking `json:"bookings,omitempty" gorm:"foreignKey:PropertyID"`
}

// AdminUser - Admin authentication model (canonical definition)
type AdminUser struct {
	ID           string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string     `json:"username" gorm:"uniqueIndex;not null"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"column:password_hash;not null"`
	Role         string     `json:"role" gorm:"default:'admin'"`
	Active       bool       `json:"active" gorm:"default:true"`
	LastLogin    *time.Time `json:"last_login"`
	LoginCount   int        `json:"login_count" gorm:"default:0"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

// GetFormattedPrice returns a formatted price string
func (p *Property) GetFormattedPrice() string {
	if p.Price == 0 {
		return "Contact for Price"
	}
	return fmt.Sprintf("$%.0f", p.Price)
}

// GetBedroomBathroom returns formatted bedroom/bathroom string
func (p *Property) GetBedroomBathroom() string {
	bedrooms := "—"
	bathrooms := "—"

	if p.Bedrooms != nil {
		bedrooms = fmt.Sprintf("%d", *p.Bedrooms)
	}
	if p.Bathrooms != nil {
		bathrooms = fmt.Sprintf("%.1f", *p.Bathrooms)
	}

	return fmt.Sprintf("%s bed, %s bath", bedrooms, bathrooms)
}

// GetFormattedSquareFeet returns formatted square footage
func (p *Property) GetFormattedSquareFeet() string {
	if p.SquareFeet == nil {
		return "—"
	}
	return fmt.Sprintf("%d sq ft", *p.SquareFeet)
}

// IncrementViewCount increments the view count for analytics
func (p *Property) IncrementViewCount(db *gorm.DB) error {
	return db.Model(p).Update("view_count", p.ViewCount+1).Error
}

// IsActive returns true if property is active/available
func (p *Property) IsActive() bool {
	return p.Status == "active"
}

// BeforeUpdate hook to track price changes
func (p *Property) BeforeUpdate(tx *gorm.DB) error {
	if p.ID == 0 {
		return nil
	}

	var original Property
	if err := tx.First(&original, p.ID).Error; err != nil {
		return nil
	}

	if original.Price != p.Price && original.Price > 0 && p.Price > 0 {
		percentChange := ((p.Price - original.Price) / original.Price) * 100

		priceChangeEvent := PriceChangeEvent{
			PropertyID:      p.ID,
			PropertyAddress: string(p.Address),
			OldPrice:        original.Price,
			NewPrice:        p.Price,
			ChangeAmount:    p.Price - original.Price,
			ChangePercent:   percentChange,
			ChangedAt:       time.Now(),
		}

		go func() {
			if err := tx.Create(&priceChangeEvent).Error; err != nil {
			}
		}()
	}

	return nil
}

type PriceChangeEvent struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	PropertyID      uint       `gorm:"not null;index" json:"property_id"`
	PropertyAddress string     `json:"property_address"`
	OldPrice        float64    `gorm:"not null" json:"old_price"`
	NewPrice        float64    `gorm:"not null" json:"new_price"`
	ChangeAmount    float64    `json:"change_amount"`
	ChangePercent   float64    `json:"change_percent"`
	ChangedAt       time.Time  `gorm:"not null;index" json:"changed_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	CampaignSent    bool       `gorm:"default:false" json:"campaign_sent"`
	CreatedAt       time.Time  `json:"created_at"`
}

type DataImport struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Type           string    `json:"type"`
	FileName       string    `json:"file_name"`
	RecordsTotal   int       `json:"records_total"`
	RecordsSuccess int       `json:"records_success"`
	RecordsFailed  int       `json:"records_failed"`
	RecordsSkipped int       `json:"records_skipped"`
	Status         string    `json:"status"`
	ErrorLog       string    `json:"error_log" gorm:"type:text"`
	ImportedBy     string    `json:"imported_by"`
	DurationMs     int64     `json:"duration_ms"`
	CreatedAt      time.Time `json:"created_at"`
}
