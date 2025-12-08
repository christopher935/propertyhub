package models

import (
	"gorm.io/gorm"
	"time"
)

// PropertyPhoto represents uploaded photos for properties with lifecycle management
type PropertyPhoto struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PropertyID uint   `json:"property_id" gorm:"not null;index"`
	MLSId      string `json:"mls_id" gorm:"not null;index"` // For cross-reference

	// Photo details
	FileName     string `json:"file_name" gorm:"not null"`
	OriginalName string `json:"original_name"`
	FilePath     string `json:"file_path" gorm:"not null"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	FileHash     string `json:"file_hash" gorm:"index"` // MD5 hash for duplicate detection

	// Photo metadata
	IsPrimary    bool   `json:"is_primary" gorm:"default:false;index"`
	DisplayOrder int    `json:"display_order" gorm:"default:0"`
	Caption      string `json:"caption"`
	AltText      string `json:"alt_text"`

	// Lifecycle management
	IsActive   bool   `json:"is_active" gorm:"default:true;index"`
	UploadedBy string `json:"uploaded_by"` // Agent ID or username

	// Status tracking
	PropertyStatus string    `json:"property_status"` // Track property status when photo was uploaded
	LastUsed       time.Time `json:"last_used"`       // When property was last active with this photo

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Property Property `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
}

// PropertyBookingEligibility represents computed booking eligibility status
type PropertyBookingEligibility struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PropertyID uint   `json:"property_id" gorm:"uniqueIndex;not null"`
	MLSID      string `json:"mls_id" gorm:"not null;index"`

	// Eligibility criteria
	HasActiveStatus bool `json:"has_active_status"`
	HasPrimaryPhoto bool `json:"has_primary_photo"`
	IsBookable      bool `json:"is_bookable"`

	// Status details
	PropertyStatus string `json:"property_status"`
	PhotoCount     int    `json:"photo_count"`      // Number of active photos
	PrimaryPhotoID *uint  `json:"primary_photo_id"` // Reference to primary photo

	// Lifecycle tracking
	LastStatusChange time.Time `json:"last_status_change"` // When bookable status changed
	LastPhotoUpdate  time.Time `json:"last_photo_update"`  // When photos were last modified
	BookableHistory  JSONB     `json:"bookable_history"`   // History of status changes

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Property     Property       `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
	PrimaryPhoto *PropertyPhoto `json:"primary_photo,omitempty" gorm:"foreignKey:PrimaryPhotoID"`
}

// PhotoUploadRequest represents a photo upload request
type PhotoUploadRequest struct {
	MLSID        string `json:"mls_id" binding:"required"`
	IsPrimary    bool   `json:"is_primary"`
	Caption      string `json:"caption"`
	AltText      string `json:"alt_text"`
	DisplayOrder int    `json:"display_order"`
}

// PhotoUploadResponse represents the response after photo upload
type PhotoUploadResponse struct {
	Success           bool                       `json:"success"`
	Message           string                     `json:"message"`
	Photo             PropertyPhoto              `json:"photo"`
	BookingEligible   bool                       `json:"booking_eligible"`
	EligibilityStatus PropertyBookingEligibility `json:"eligibility_status"`
}

// BookingEligibilityResponse represents booking eligibility check response
type BookingEligibilityResponse struct {
	Success           bool                       `json:"success"`
	IsBookable        bool                       `json:"is_bookable"`
	Reason            string                     `json:"reason"`
	EligibilityStatus PropertyBookingEligibility `json:"eligibility_status"`
	RequiredActions   []string                   `json:"required_actions"`
}

// PropertyPhotoSummary represents a summary of property photos
type PropertyPhotoSummary struct {
	PropertyID      uint      `json:"property_id"`
	MLSId           string    `json:"mls_id"`
	TotalPhotos     int       `json:"total_photos"`
	ActivePhotos    int       `json:"active_photos"`
	HasPrimaryPhoto bool      `json:"has_primary_photo"`
	PrimaryPhotoURL string    `json:"primary_photo_url"`
	LastUpdated     time.Time `json:"last_updated"`
}

// Methods for PropertyPhoto

// GetPublicURL returns the public URL for accessing the photo
func (pp *PropertyPhoto) GetPublicURL(baseURL string) string {
	return baseURL + "/api/v1/photos/" + pp.FileName
}

// SetAsPrimary sets this photo as the primary photo for the property
func (pp *PropertyPhoto) SetAsPrimary(db *gorm.DB) error {
	// First, unset any existing primary photos for this property
	if err := db.Model(&PropertyPhoto{}).
		Where("property_id = ? AND id != ?", pp.PropertyID, pp.ID).
		Update("is_primary", false).Error; err != nil {
		return err
	}

	// Set this photo as primary
	pp.IsPrimary = true
	return db.Save(pp).Error
}

// Deactivate marks the photo as inactive without deleting it
func (pp *PropertyPhoto) Deactivate(db *gorm.DB) error {
	pp.IsActive = false
	return db.Save(pp).Error
}

// Reactivate marks the photo as active again
func (pp *PropertyPhoto) Reactivate(db *gorm.DB) error {
	pp.IsActive = true
	pp.LastUsed = time.Now()
	return db.Save(pp).Error
}

// Methods for PropertyBookingEligibility

// UpdateEligibility recalculates and updates booking eligibility
func (pbe *PropertyBookingEligibility) UpdateEligibility(db *gorm.DB) error {
	// Get current property status
	var property Property
	if err := db.First(&property, pbe.PropertyID).Error; err != nil {
		return err
	}

	// Check if property has active status
	pbe.HasActiveStatus = property.Status == "active"
	pbe.PropertyStatus = property.Status

	// Count active photos and check for primary photo
	var photoCount int64
	var primaryPhoto PropertyPhoto

	db.Model(&PropertyPhoto{}).
		Where("property_id = ? AND is_active = ?", pbe.PropertyID, true).
		Count(&photoCount)

	pbe.PhotoCount = int(photoCount)

	// Check for primary photo
	if err := db.Where("property_id = ? AND is_primary = ? AND is_active = ?",
		pbe.PropertyID, true, true).First(&primaryPhoto).Error; err == nil {
		pbe.HasPrimaryPhoto = true
		pbe.PrimaryPhotoID = &primaryPhoto.ID
	} else {
		pbe.HasPrimaryPhoto = false
		pbe.PrimaryPhotoID = nil
	}

	// Calculate booking eligibility
	wasBookable := pbe.IsBookable
	pbe.IsBookable = pbe.HasActiveStatus && pbe.HasPrimaryPhoto

	// Track status changes
	if wasBookable != pbe.IsBookable {
		pbe.LastStatusChange = time.Now()

		// Add to history
		if pbe.BookableHistory == nil {
			pbe.BookableHistory = make(JSONB)
		}

		historyEntry := map[string]interface{}{
			"timestamp":         time.Now(),
			"bookable":          pbe.IsBookable,
			"has_active_status": pbe.HasActiveStatus,
			"has_primary_photo": pbe.HasPrimaryPhoto,
			"property_status":   pbe.PropertyStatus,
			"photo_count":       pbe.PhotoCount,
		}

		// Add to history array
		if history, ok := pbe.BookableHistory["changes"].([]interface{}); ok {
			pbe.BookableHistory["changes"] = append(history, historyEntry)
		} else {
			pbe.BookableHistory["changes"] = []interface{}{historyEntry}
		}
	}

	return db.Save(pbe).Error
}

// GetEligibilityReason returns a human-readable reason for booking eligibility status
func (pbe *PropertyBookingEligibility) GetEligibilityReason() string {
	if pbe.IsBookable {
		return "Property is active and has primary photo - ready for bookings"
	}

	if !pbe.HasActiveStatus {
		return "Property is not in active status (current: " + pbe.PropertyStatus + ")"
	}

	if !pbe.HasPrimaryPhoto {
		return "Property needs a primary photo to accept bookings"
	}

	return "Property does not meet booking requirements"
}

// GetRequiredActions returns a list of actions needed to make property bookable
func (pbe *PropertyBookingEligibility) GetRequiredActions() []string {
	var actions []string

	if !pbe.HasActiveStatus {
		actions = append(actions, "Wait for property to return to active status")
	}

	if !pbe.HasPrimaryPhoto {
		actions = append(actions, "Upload and set a primary photo for the property")
	}

	return actions
}

// Helper functions

// FindOrCreateBookingEligibility finds existing or creates new booking eligibility record
func FindOrCreateBookingEligibility(db *gorm.DB, propertyID uint, mlsID string) (*PropertyBookingEligibility, error) {
	var eligibility PropertyBookingEligibility

	err := db.Where("property_id = ?", propertyID).First(&eligibility).Error
	if err == gorm.ErrRecordNotFound {
		// Create new eligibility record
		eligibility = PropertyBookingEligibility{
			PropertyID:      propertyID,
			MLSID:           mlsID,
			BookableHistory: make(JSONB),
		}

		if err := eligibility.UpdateEligibility(db); err != nil {
			return nil, err
		}

		return &eligibility, nil
	} else if err != nil {
		return nil, err
	}

	return &eligibility, nil
}

// GetBookableProperties returns all properties that are currently bookable
func GetBookableProperties(db *gorm.DB) ([]PropertyBookingEligibility, error) {
	var bookableProperties []PropertyBookingEligibility

	err := db.Where("is_bookable = ?", true).
		Preload("Property").
		Preload("PrimaryPhoto").
		Find(&bookableProperties).Error

	return bookableProperties, err
}

// GetPropertyPhotoSummary returns a summary of photos for a property
func GetPropertyPhotoSummary(db *gorm.DB, propertyID uint, mlsID string) (*PropertyPhotoSummary, error) {
	var summary PropertyPhotoSummary
	summary.PropertyID = propertyID
	summary.MLSId = mlsID

	// Count total and active photos
	var totalCount, activeCount int64

	db.Model(&PropertyPhoto{}).Where("property_id = ?", propertyID).Count(&totalCount)
	db.Model(&PropertyPhoto{}).Where("property_id = ? AND is_active = ?", propertyID, true).Count(&activeCount)

	summary.TotalPhotos = int(totalCount)
	summary.ActivePhotos = int(activeCount)

	// Check for primary photo
	var primaryPhoto PropertyPhoto
	if err := db.Where("property_id = ? AND is_primary = ? AND is_active = ?",
		propertyID, true, true).First(&primaryPhoto).Error; err == nil {
		summary.HasPrimaryPhoto = true
		summary.PrimaryPhotoURL = primaryPhoto.FilePath
		summary.LastUpdated = primaryPhoto.UpdatedAt
	}

	return &summary, nil
}
