package models

import (
	"gorm.io/gorm"
	"time"
)

// PreListingItem tracks the operational workflow from Terry's email to MLS listing
type PreListingItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Property Information
	Address     string `json:"address" gorm:"not null"`
	FullAddress string `json:"full_address"`
	City        string `json:"city"`
	State       string `json:"state"`
	ZipCode     string `json:"zip_code"`

	// Status and Progress
	Status        string `json:"status" gorm:"default:'email_received'"` // email_received, lockbox_pending, lockbox_placed, photos_scheduled, photos_complete, pricing_set, listed, confirmed
	IsOverdue     bool   `json:"is_overdue" gorm:"default:false"`
	OverdueReason string `json:"overdue_reason"`

	// Terry's Email Data
	TerryEmailDate    *time.Time `json:"terry_email_date"`
	TerryEmailSubject string     `json:"terry_email_subject"`
	TerryEmailContent string     `json:"terry_email_content" gorm:"type:text"`
	TargetListingDate *time.Time `json:"target_listing_date"`

	// Lockbox Information
	LockboxDeadline   *time.Time `json:"lockbox_deadline"`
	LockboxPlacedDate *time.Time `json:"lockbox_placed_date"`
	LockboxType       string     `json:"lockbox_type"` // lockbox, door_panel, already_placed
	LockboxNotes      string     `json:"lockbox_notes"`

	// Photo Information
	PhotoScheduledDate *time.Time `json:"photo_scheduled_date"`
	PhotoExpectedDate  *time.Time `json:"photo_expected_date"`
	PhotoCompletedDate *time.Time `json:"photo_completed_date"`
	PhotoStatus        string     `json:"photo_status"` // needed, scheduled, complete, provided, reused
	PhotoURL           string     `json:"photo_url"`
	PhotoNotes         string     `json:"photo_notes"`

	// Sign Information (rarely used)
	SignPlacedDate *time.Time `json:"sign_placed_date"`
	SignStatus     string     `json:"sign_status"` // needed, placed, not_needed
	SignNotes      string     `json:"sign_notes"`

	// Pricing and Listing
	PricingSetDate *time.Time `json:"pricing_set_date"`
	ListingPrice   *float64   `json:"listing_price"`
	MLSListedDate  *time.Time `json:"mls_listed_date"`
	MLSNumber      string     `json:"mls_number"`
	ConfirmedDate  *time.Time `json:"confirmed_date"`

	// Administrative
	AdminNotes     string     `json:"admin_notes" gorm:"type:text"`
	LastAlertSent  *time.Time `json:"last_alert_sent"`
	ManualOverride bool       `json:"manual_override" gorm:"default:false"`
	OverrideReason string     `json:"override_reason"`
}

// EmailAlert tracks alerts sent for overdue items
type EmailAlert struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Alert Information
	PreListingItemID uint   `json:"pre_listing_item_id" gorm:"not null"`
	AlertType        string `json:"alert_type"`  // lockbox_overdue, photo_overdue, pricing_overdue, listing_overdue
	AlertLevel       string `json:"alert_level"` // warning, urgent, critical
	Message          string `json:"message" gorm:"type:text"`

	// Status
	IsSent     bool       `json:"is_sent" gorm:"default:false"`
	SentAt     *time.Time `json:"sent_at"`
	IsResolved bool       `json:"is_resolved" gorm:"default:false"`
	ResolvedAt *time.Time `json:"resolved_at"`

	// Relationships
	PreListingItem PreListingItem `json:"pre_listing_item" gorm:"foreignKey:PreListingItemID"`
}

// IncomingEmail tracks all emails received for processing
type IncomingEmail struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Email Information
	FromEmail  string    `json:"from_email" gorm:"not null"`
	ToEmail    string    `json:"to_email" gorm:"not null"`
	Subject    string    `json:"subject" gorm:"not null"`
	Content    string    `json:"content" gorm:"type:text"`
	ReceivedAt time.Time `json:"received_at"`

	// Processing Information
	EmailType        string  `json:"email_type"`                                 // terry_alert, terry_listing, prs_lockbox, prs_photo_schedule, prs_photo_complete, prs_sign, other
	ProcessingStatus string  `json:"processing_status" gorm:"default:'pending'"` // pending, processed, failed, requires_review
	Confidence       float64 `json:"confidence"`
	ExtractedAddress string  `json:"extracted_address"`
	ExtractedData    string  `json:"extracted_data" gorm:"type:text"` // JSON

	// Relationships
	PreListingItemID *uint          `json:"pre_listing_item_id"`
	PreListingItem   PreListingItem `json:"pre_listing_item" gorm:"foreignKey:PreListingItemID"`
}

// Status constants
const (
	StatusEmailReceived   = "email_received"
	StatusLockboxPending  = "lockbox_pending"
	StatusLockboxPlaced   = "lockbox_placed"
	StatusPhotosScheduled = "photos_scheduled"
	StatusPhotosComplete  = "photos_complete"
	StatusPricingSet      = "pricing_set"
	StatusListed          = "listed"
	StatusConfirmed       = "confirmed"
)

// Email type constants
const (
	EmailTypeTerryAlert         = "terry_alert"
	EmailTypeTerryListing       = "terry_listing"
	EmailTypePRSLockbox         = "prs_lockbox"
	EmailTypePRSPhotoSchedule   = "prs_photo_schedule"
	EmailTypePRSPhotoComplete   = "prs_photo_complete"
	EmailTypePRSSign            = "prs_sign"
	EmailTypeLeaseSent          = "lease_sent"
	EmailTypeLeaseComplete      = "lease_complete"
	EmailTypeLeaseAmendment     = "lease_amendment"
	EmailTypeDepositReceived    = "deposit_received"
	EmailTypeFirstMonthReceived = "first_month_received"
	EmailTypeMondayPayment      = "monday_payment"
	EmailTypeOther              = "other"
)

// Processing status constants
const (
	ProcessingStatusPending        = "pending"
	ProcessingStatusProcessed      = "processed"
	ProcessingStatusFailed         = "failed"
	ProcessingStatusRequiresReview = "requires_review"
)

// Alert type constants
const (
	AlertTypeLockboxOverdue = "lockbox_overdue"
	AlertTypePhotoOverdue   = "photo_overdue"
	AlertTypePricingOverdue = "pricing_overdue"
	AlertTypeListingOverdue = "listing_overdue"
)

// Alert level constants
const (
	AlertLevelWarning  = "warning"
	AlertLevelUrgent   = "urgent"
	AlertLevelCritical = "critical"
)

// EmailProcessingResult represents the result of email processing
type EmailProcessingResult struct {
	Success          bool                   `json:"success"`
	EmailType        string                 `json:"email_type"`
	Confidence       float64                `json:"confidence"`
	ExtractedData    map[string]interface{} `json:"extracted_data"`
	PreListingItemID *uint                  `json:"pre_listing_item_id,omitempty"`
	PreListingID     *uint                  `json:"pre_listing_id,omitempty"` // Alias for PreListingItemID
	CreatedNewItem   bool                   `json:"created_new_item"`
	Alerts           []string               `json:"alerts,omitempty"`
	Errors           []string               `json:"errors,omitempty"`
	ProcessingNotes  string                 `json:"processing_notes,omitempty"`
	StatusUpdate     string                 `json:"status_update,omitempty"`
	Message          string                 `json:"message,omitempty"` // COMPLIANCE: For FUB boundary messages
}

// PreListingStats represents dashboard statistics
type PreListingStats struct {
	TotalActive       int `json:"total_active"`
	PendingLockbox    int `json:"pending_lockbox"`
	PendingPhotos     int `json:"pending_photos"`
	ReadyToList       int `json:"ready_to_list"`
	OverdueItems      int `json:"overdue_items"`
	CompletedToday    int `json:"completed_today"`
	CompletedThisWeek int `json:"completed_this_week"`
}
