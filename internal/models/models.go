package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/security"
	"log"
)

// DealType constants for commission tracking
const (
	DealTypeDoubleEnded = "double_ended"
	DealTypeListingSide = "listing_side"
	DealTypeTenantSide  = "tenant_side"
)



// StringArray type for storing arrays of strings
type StringArray []string

func (sa StringArray) Value() (driver.Value, error) {
	return json.Marshal(sa)
}

func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = make(StringArray, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, sa)
}

// Property ToDict method - converts Property to map for JSON responses

func (p Property) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                p.ID,
		"mls_id":            p.MLSId,
		"address":           p.Address,
		"city":              p.City,
		"state":             p.State,
		"zip_code":          p.ZipCode,
		"price":             p.Price,
		"bedrooms":          p.Bedrooms,
		"bathrooms":         p.Bathrooms,
		"square_feet":       p.SquareFeet,
		"property_type":     p.PropertyType,
		"description":       p.Description,
		"images":            p.Images,
		"status":            p.Status,
		"listing_agent_id":  p.ListingAgentID,
		"featured_image":    p.FeaturedImage,
		"property_features": p.PropertyFeatures,
		"created_at":        p.CreatedAt,
		"updated_at":        p.UpdatedAt,
	}
}

// Booking represents a property showing appointment (local coordination only)
// All lead data, communications, and workflows are handled by FUB
type Booking struct {
	ID                 uint                     `json:"id" gorm:"primaryKey"`
	ReferenceNumber    string                   `json:"reference_number" gorm:"unique;not null"` // Unique booking reference
	PropertyID         uint                     `json:"property_id"`                             // Optional for external bookings
	PropertyAddress    string                   `json:"property_address"`                        // For external bookings without PropertyID
	FUBLeadID          string                   `json:"fub_lead_id" gorm:"not null"`             // Reference to FUB lead
	Email              security.EncryptedString `json:"email"`
	Name               security.EncryptedString `json:"name"`
	FUBSynced          bool                     `json:"fub_synced" gorm:"default:false"`
	Phone              security.EncryptedString `json:"phone"`          // Synced from FUB
	InterestLevel      string                   `json:"interest_level"` // Synced from FUB
	ShowingDate        time.Time                `json:"showing_date" gorm:"not null"`
	DurationMinutes    int                      `json:"duration_minutes" gorm:"default:30"`
	Status             string                   `json:"status" gorm:"default:'scheduled'"`
	Notes              string                   `json:"notes" gorm:"type:text"`
	ShowingType        string                   `json:"showing_type" gorm:"default:'in-person'"`
	AttendeeCount      int                      `json:"attendee_count" gorm:"default:1"`
	SpecialRequests    string                   `json:"special_requests" gorm:"type:text"`
	FUBActionPlanID    string                   `json:"fub_action_plan_id"` // FUB handles all automation
	CompletedAt        *time.Time               `json:"completed_at"`
	CancellationReason string                   `json:"cancellation_reason"`
	RescheduledFrom    *uint                    `json:"rescheduled_from"`

	// COMPLIANCE: Consent tracking for PropertyHub services
	ConsentGiven     bool       `json:"consent_given" gorm:"default:false"`
	ConsentSource    string     `json:"consent_source" gorm:"default:'direct'"` // "direct", "fub", "inherited"
	ConsentTimestamp *time.Time `json:"consent_timestamp"`
	MarketingConsent bool       `json:"marketing_consent" gorm:"default:false"`
	TermsAccepted    bool       `json:"terms_accepted" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Property Property `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
}

func (b Booking) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                  b.ID,
		"reference_number":    b.ReferenceNumber,
		"property_id":         b.PropertyID,
		"fub_lead_id":         b.FUBLeadID,
		"phone":               b.Phone,
		"interest_level":      b.InterestLevel,
		"showing_date":        b.ShowingDate,
		"duration_minutes":    b.DurationMinutes,
		"status":              b.Status,
		"notes":               b.Notes,
		"showing_type":        b.ShowingType,
		"attendee_count":      b.AttendeeCount,
		"special_requests":    b.SpecialRequests,
		"fub_action_plan_id":  b.FUBActionPlanID,
		"completed_at":        b.CompletedAt,
		"cancellation_reason": b.CancellationReason,
		"rescheduled_from":    b.RescheduledFrom,
		"consent_given":       b.ConsentGiven,
		"consent_source":      b.ConsentSource,
		"consent_timestamp":   b.ConsentTimestamp,
		"marketing_consent":   b.MarketingConsent,
		"terms_accepted":      b.TermsAccepted,
		"created_at":          b.CreatedAt,
		"updated_at":          b.UpdatedAt,
		"property":            b.Property.ToDict(),
	}
}

// WebhookEvent represents incoming webhook events from various sources
type WebhookEvent struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Source      string         `json:"source" gorm:"not null;index"`          // fub, buildium, stripe, twilio, etc.
	EventType   string         `json:"event_type" gorm:"not null;index"`
	EventID     string         `json:"event_id" gorm:"uniqueIndex"`
	Payload     JSONB          `json:"payload" gorm:"type:json"`
	Status      string         `json:"status" gorm:"default:'pending';index"` // pending, processed, failed
	ProcessedAt *time.Time     `json:"processed_at"`
	Error       string         `json:"error" gorm:"type:text"`
	RetryCount  int            `json:"retry_count" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// BookingStatusLog tracks status changes for audit trail
type BookingStatusLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BookingID uint      `json:"booking_id" gorm:"not null"`
	FUBLeadID string    `json:"fub_lead_id"`
	OldStatus string    `json:"old_status" gorm:"not null"`
	NewStatus string    `json:"new_status" gorm:"not null"`
	Source    string    `json:"source" gorm:"not null"` // "FUB webhook", "manual", "system"
	ChangedAt time.Time `json:"changed_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemSettings represents application-wide settings
type SystemSettings struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	Type      string    `json:"type" gorm:"default:'string'"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationState tracks the state of notifications for iPhone-like behavior
type NotificationState struct {
	ID               uint   `json:"id" gorm:"primaryKey"`
	UserID           string `json:"user_id" gorm:"not null"`           // Agent identifier
	NotificationKey  string `json:"notification_key" gorm:"not null"`  // Unique key for the notification
	NotificationType string `json:"notification_type" gorm:"not null"` // "alert", "todo", "reminder", etc.
	SourceID         string `json:"source_id"`                         // ID of the source record (booking, property, etc.)
	SourceType       string `json:"source_type"`                       // "booking", "property", "pre_listing", etc.

	// iPhone-like state management
	IsViewed           bool `json:"is_viewed" gorm:"default:false"`
	IsDismissed        bool `json:"is_dismissed" gorm:"default:false"`
	IsCompleted        bool `json:"is_completed" gorm:"default:false"`
	AutoDismiss        bool `json:"auto_dismiss" gorm:"default:false"`        // Should auto-dismiss when viewed
	PersistUntilAction bool `json:"persist_until_action" gorm:"default:true"` // Persist until action taken

	// Timing
	ViewedAt    *time.Time `json:"viewed_at"`
	DismissedAt *time.Time `json:"dismissed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	ExpiresAt   *time.Time `json:"expires_at"` // Auto-expire after this time

	// Metadata
	Priority       string `json:"priority" gorm:"default:'medium'"` // "low", "medium", "high", "urgent"
	Title          string `json:"title" gorm:"not null"`
	Message        string `json:"message" gorm:"type:text"`
	ActionRequired bool   `json:"action_required" gorm:"default:false"`
	ActionURL      string `json:"action_url"`   // URL for action button
	ActionLabel    string `json:"action_label"` // Label for action button

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsActive returns true if the notification should be shown
func (ns NotificationState) IsActive() bool {
	// Don't show if dismissed or completed
	if ns.IsDismissed || ns.IsCompleted {
		return false
	}

	// Don't show if expired
	if ns.ExpiresAt != nil && time.Now().After(*ns.ExpiresAt) {
		return false
	}

	// Auto-dismiss notifications are hidden after being viewed
	if ns.AutoDismiss && ns.IsViewed {
		return false
	}

	return true
}

// NotificationBatch represents a batch of notifications for efficient updates
type NotificationBatch struct {
	UserID          string `json:"user_id"`
	Action          string `json:"action"` // "view", "dismiss", "complete"
	NotificationIDs []uint `json:"notification_ids"`
}

// ClosingPipeline tracks the lease workflow and closing process-in
type ClosingPipeline struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	PropertyID      *uint     `json:"property_id"` // Optional link to Property
	PropertyAddress string    `json:"property_address" gorm:"not null"`
	MLSID           string    `json:"mls_id"`
	SoldDate        time.Time `json:"sold_date" gorm:"not null"`

	// Lease Workflow Status
	LeaseSentOut           bool       `json:"lease_sent_out" gorm:"default:false"`
	LeaseSentDate          *time.Time `json:"lease_sent_date"`
	LeaseComplete          bool       `json:"lease_complete" gorm:"default:false"`
	LeaseCompleteDate      *time.Time `json:"lease_complete_date"`
	DepositReceived        bool       `json:"deposit_received" gorm:"default:false"`
	DepositReceivedDate    *time.Time `json:"deposit_received_date"`
	FirstMonthReceived     bool       `json:"first_month_received" gorm:"default:false"`
	FirstMonthReceivedDate *time.Time `json:"first_month_received_date"`

	// Move-in Information
	MoveInDate       *time.Time `json:"move_in_date"`
	MoveInDateSource string     `json:"move_in_date_source"` // "lease_pdf", "email", "manual"

	// Financial Information
	DepositAmount *float64 `json:"deposit_amount"`
	MonthlyRent   *float64 `json:"monthly_rent"`

	// Commission Tracking (NEW - for analytics)
	CommissionEarned    *float64   `json:"commission_earned" gorm:"type:decimal(10,2)"`
	CommissionRate      *float64   `json:"commission_rate" gorm:"type:decimal(5,2)"`
	DealType            string     `json:"deal_type" gorm:"type:varchar(20)"`
	ListingAgentID      *uint      `json:"listing_agent_id"`
	TenantAgentID       *uint      `json:"tenant_agent_id"`
	LeaseSignedDate     *time.Time `json:"lease_signed_date"`
	ApplicationDate     *time.Time `json:"application_date"`
	ApprovalDate        *time.Time `json:"approval_date"`

	// Tenant Information (minimal for tracking)
	TenantName  security.EncryptedString `json:"tenant_name"`
	TenantEmail security.EncryptedString `json:"tenant_email"`
	TenantPhone security.EncryptedString `json:"tenant_phone"`

	// Status and Alerts
	Status     string      `json:"status" gorm:"default:'pending'"` // pending, in_progress, ready, completed
	AlertFlags StringArray `json:"alert_flags" gorm:"type:json"`    // overdue_lease, missing_deposit, etc.

	// Audit Trail
	ProcessedEmails      StringArray `json:"processed_emails" gorm:"type:json"` // Track which emails updated this record
	LastEmailProcessedAt *time.Time  `json:"last_email_processed_at"`
	AmendmentNotes       string      `json:"amendment_notes" gorm:"type:text"` // Track lease amendments

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Property *Property `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
}

func (cp ClosingPipeline) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                        cp.ID,
		"property_id":               cp.PropertyID,
		"property_address":          cp.PropertyAddress,
		"mls_id":                    cp.MLSID,
		"sold_date":                 cp.SoldDate,
		"lease_sent_out":            cp.LeaseSentOut,
		"lease_sent_date":           cp.LeaseSentDate,
		"lease_complete":            cp.LeaseComplete,
		"lease_complete_date":       cp.LeaseCompleteDate,
		"deposit_received":          cp.DepositReceived,
		"deposit_received_date":     cp.DepositReceivedDate,
		"first_month_received":      cp.FirstMonthReceived,
		"first_month_received_date": cp.FirstMonthReceivedDate,
		"move_in_date":              cp.MoveInDate,
		"move_in_date_source":       cp.MoveInDateSource,
		"deposit_amount":            cp.DepositAmount,
		"monthly_rent":              cp.MonthlyRent,
		"tenant_name":               cp.TenantName,
		"tenant_email":              cp.TenantEmail,
		"tenant_phone":              cp.TenantPhone,
		"status":                    cp.Status,
		"alert_flags":               cp.AlertFlags,
		"processed_emails":          cp.ProcessedEmails,
		"last_email_processed_at":   cp.LastEmailProcessedAt,
		"created_at":                cp.CreatedAt,
		"updated_at":                cp.UpdatedAt,
	}
}

// GetDaysToMoveIn calculates days remaining until move-in
func (cp ClosingPipeline) GetDaysToMoveIn() *int {
	if cp.MoveInDate == nil {
		return nil
	}
	days := int(cp.MoveInDate.Sub(time.Now()).Hours() / 24)
	return &days
}

// GetCompletionPercentage calculates workflow completion percentage
func (cp ClosingPipeline) GetCompletionPercentage() int {
	completed := 0
	total := 4

	if cp.LeaseSentOut {
		completed++
	}
	if cp.LeaseComplete {
		completed++
	}
	if cp.DepositReceived {
		completed++
	}
	if cp.FirstMonthReceived {
		completed++
	}

	return (completed * 100) / total
}

// GetStatusSummary returns a human-readable status summary
func (cp ClosingPipeline) GetStatusSummary() string {
	if cp.FirstMonthReceived && cp.DepositReceived && cp.LeaseComplete {
		return "Ready for move-in"
	}
	if cp.LeaseComplete && cp.DepositReceived {
		return "Awaiting first month rent"
	}
	if cp.LeaseComplete {
		return "Awaiting payments"
	}
	if cp.LeaseSentOut {
		return "Awaiting lease signature"
	}
	return "Lease preparation"
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	// Migrate models one by one to identify issues
	if err := db.AutoMigrate(&Property{}); err != nil {
		log.Printf("❌ Property migration failed: %v", err)
		return err
	}
	log.Println("✅ Property migration completed")

	if err := db.AutoMigrate(&Booking{}); err != nil {
		log.Printf("❌ Booking migration failed: %v", err)
		return err
	}
	log.Println("✅ Booking migration completed")

	if err := db.AutoMigrate(&NotificationState{}); err != nil {
		log.Printf("❌ NotificationState migration failed: %v", err)
		return err
	}
	log.Println("✅ NotificationState migration completed")

	log.Println("✅ Database migrations completed successfully")
	log.Println("🎉 Database initialization complete - proceeding to HTTP server startup")
	return nil
}

// ScheduledAction stores automation actions to be executed later
type ScheduledAction struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	ActionType   string          `json:"action_type" gorm:"not null"`
	TargetID     string          `json:"target_id" gorm:"not null"`
	ActionData   json.RawMessage `json:"action_data" gorm:"type:jsonb"`
	ScheduledAt  time.Time       `json:"scheduled_at" gorm:"not null"`
	ExecutedAt   *time.Time      `json:"executed_at"`
	Status       string          `json:"status" gorm:"default:'pending'"`
	ErrorMessage string          `json:"error_message"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Lead represents a potential client for FUB integration
type Lead struct {
	ID               uint        `json:"id" gorm:"primaryKey"`
	FirstName        string      `json:"first_name" gorm:"not null"`
	LastName         string      `json:"last_name" gorm:"not null"`
	Email            string      `json:"email" gorm:"not null"`
	Phone            string      `json:"phone"`
	City             string      `json:"city"`
	State            string      `json:"state"`
	FUBLeadID        string      `json:"fub_lead_id" gorm:"uniqueIndex"`
	Source           string      `json:"source" gorm:"default:'Website'"`
	Status           string      `json:"status" gorm:"default:'new'"`
	AssignedAgentID  string      `json:"assigned_agent_id" gorm:"index"`
	Tags             StringArray `json:"tags" gorm:"type:json"`
	CustomFields     JSONB       `json:"custom_fields" gorm:"type:json"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FUB-specific models for integration (FUBLead is defined in fub_models.go)

type FUBCommunication struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	FUBLeadID   string     `json:"fub_lead_id" gorm:"not null"`
	Type        string     `json:"type"` // email, sms, call
	Subject     string     `json:"subject"`
	Body        string     `json:"body" gorm:"type:text"`
	Status      string     `json:"status"`
	SentAt      time.Time  `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	OpenedAt    *time.Time `json:"opened_at"`
	ClickedAt   *time.Time `json:"clicked_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type FUBTask struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	FUBTaskID   string     `json:"fub_task_id" gorm:"uniqueIndex"`
	FUBLeadID   string     `json:"fub_lead_id" gorm:"not null"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	Type        string     `json:"type"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status" gorm:"default:'pending'"`
	DueAt       time.Time  `json:"due_at"`
	CompletedAt *time.Time `json:"completed_at"`
	AgentID     string     `json:"agent_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type FUBCampaign struct {
	ID            uint        `json:"id" gorm:"primaryKey"`
	FUBCampaignID string      `json:"fub_campaign_id" gorm:"uniqueIndex"`
	Name          string      `json:"name" gorm:"not null"`
	Type          string      `json:"type"`
	Status        string      `json:"status"`
	LeadIDs       StringArray `json:"lead_ids" gorm:"type:json"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type FUBNote struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FUBNoteID string    `json:"fub_note_id" gorm:"uniqueIndex"`
	FUBLeadID string    `json:"fub_lead_id" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text"`
	AgentID   string    `json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type FUBProperty struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	FUBPropertyID string    `json:"fub_property_id" gorm:"uniqueIndex"`
	Address       string    `json:"address" gorm:"not null"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	ZipCode       string    `json:"zip_code"`
	Price         float64   `json:"price"`
	PropertyType  string    `json:"property_type"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FUBSmartList struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FUBSmartListID string    `json:"fub_smart_list_id" gorm:"uniqueIndex"`
	Name           string    `json:"name" gorm:"not null"`
	Criteria       JSONB     `json:"criteria" gorm:"type:json"`
	LeadCount      int       `json:"lead_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Role constants for admin users
const (
	RoleMainAdmin  = "main_admin"  // Full access
	RoleAdmin      = "admin"       // Standard admin access
	RoleSuperAdmin = "super_admin" // Legacy role, treated as main_admin
	RoleUser       = "user"        // Limited access
	RoleReadOnly   = "read_only"   // View-only access
)

// PropertyState represents the unified property state across all systems
// Central Property State Manager - Single source of truth
type PropertyState struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Property Data (from HAR scraper)
	MLSId        string   `json:"mls_id" gorm:"uniqueIndex"`
	Address      string   `json:"address"`
	Price        *float64 `json:"price"`
	Bedrooms     *int     `json:"bedrooms"`
	Bathrooms    *float32 `json:"bathrooms"`
	SquareFeet   *int     `json:"square_feet"`
	PropertyType string   `json:"property_type"`

	// Status Management (unified across all systems)
	Status          string    `json:"status"`        // active, pending, sold, off_market
	StatusSource    string    `json:"status_source"` // har, fub, email, manual
	StatusUpdatedAt time.Time `json:"status_updated_at"`

	// System Integration Tracking
	HARScraperData  JSONB `json:"har_scraper_data" gorm:"type:json"`
	PropertyHubData JSONB `json:"property_hub_data" gorm:"type:json"`
	FUBData         JSONB `json:"fub_data" gorm:"type:json"`
	EmailData       JSONB `json:"email_data" gorm:"type:json"`

	// Availability and Booking State
	IsBookable      bool       `json:"is_bookable"`
	BookingCount    int        `json:"booking_count"`
	LastBookingDate *time.Time `json:"last_booking_date"`

	// Lead and Customer Tracking
	LeadCount        int         `json:"lead_count"`
	FUBLeadIDs       StringArray `json:"fub_lead_ids" gorm:"type:json"`
	CustomerInterest int         `json:"customer_interest"` // 0-100 score

	// Conflict Resolution
	ConflictFlags  StringArray `json:"conflict_flags" gorm:"type:json"`
	LastConflictAt *time.Time  `json:"last_conflict_at"`

	// Metadata
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	LastSyncedAt time.Time   `json:"last_synced_at"`
	SyncErrors   StringArray `json:"sync_errors" gorm:"type:json"`
}

func (ps PropertyState) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                ps.ID,
		"mls_id":            ps.MLSId,
		"address":           ps.Address,
		"price":             ps.Price,
		"bedrooms":          ps.Bedrooms,
		"bathrooms":         ps.Bathrooms,
		"square_feet":       ps.SquareFeet,
		"property_type":     ps.PropertyType,
		"status":            ps.Status,
		"status_source":     ps.StatusSource,
		"status_updated_at": ps.StatusUpdatedAt,
		"is_bookable":       ps.IsBookable,
		"booking_count":     ps.BookingCount,
		"last_booking_date": ps.LastBookingDate,
		"lead_count":        ps.LeadCount,
		"fub_lead_ids":      ps.FUBLeadIDs,
		"customer_interest": ps.CustomerInterest,
		"conflict_flags":    ps.ConflictFlags,
		"last_conflict_at":  ps.LastConflictAt,
		"created_at":        ps.CreatedAt,
		"updated_at":        ps.UpdatedAt,
		"last_synced_at":    ps.LastSyncedAt,
		"sync_errors":       ps.SyncErrors,
	}
}

// PropertyUpdateRequest represents a request to update property state
type PropertyUpdateRequest struct {
	Source       string   `json:"source"` // har, propertyhub, fub, email
	MLSId        string   `json:"mls_id"`
	Address      string   `json:"address"`
	Price        *float64 `json:"price"`
	Bedrooms     *int     `json:"bedrooms"`
	Bathrooms    *float32 `json:"bathrooms"`
	SquareFeet   *int     `json:"square_feet"`
	PropertyType string   `json:"property_type"`
	Status       string   `json:"status"`
	Data         JSONB    `json:"data"` // Source-specific data
}

// PropertyConflict represents a data conflict between systems
type PropertyConflict struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	PropertyID uint       `json:"property_id"`
	Field      string     `json:"field"`
	OldValue   string     `json:"old_value"`
	NewValue   string     `json:"new_value"`
	Source     string     `json:"source"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Approval represents an approval request in the system
type Approval struct {
	ID              uint        `json:"id" gorm:"primaryKey"`
	ApprovalType    string      `json:"approval_type" gorm:"not null"`    // rental_application, property_listing, document_verification, showing_request
	Status          string      `json:"status" gorm:"default:'pending'"`  // pending, approved, rejected, under_review
	Priority        string      `json:"priority" gorm:"default:'medium'"` // low, medium, high
	ApplicantName   string      `json:"applicant_name"`
	PropertyAddress string      `json:"property_address"`
	Documents       StringArray `json:"documents" gorm:"type:json"`
	Notes           string      `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// Contact represents a contact inquiry from the website
type Contact struct {
	ID         uint                     `json:"id" gorm:"primaryKey"`
	Name       security.EncryptedString `json:"name" gorm:"not null"`
	Phone      security.EncryptedString `json:"phone" gorm:"not null"`
	Email      security.EncryptedString `json:"email"`
	Message    string                   `json:"message" gorm:"type:text"`
	PropertyID string                   `json:"property_id"` // Optional - specific property inquiry
	Urgent     bool                     `json:"urgent" gorm:"default:false"`
	Source     string                   `json:"source" gorm:"default:'website'"` // homepage_quick_contact, property_detail, etc.
	Status     string                   `json:"status" gorm:"default:'new'"`     // new, contacted, converted, closed
	FUBLeadID  string                   `json:"fub_lead_id"`                     // Follow Up Boss lead ID when synced
	FUBSynced  bool                     `json:"fub_synced" gorm:"default:false"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
}

func (c Contact) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":          c.ID,
		"name":        c.Name,
		"phone":       c.Phone,
		"email":       c.Email,
		"message":     c.Message,
		"property_id": c.PropertyID,
		"urgent":      c.Urgent,
		"source":      c.Source,
		"status":      c.Status,
		"fub_lead_id": c.FUBLeadID,
		"fub_synced":  c.FUBSynced,
		"created_at":  c.CreatedAt,
		"updated_at":  c.UpdatedAt,
	}
}

// CommissionAnalytics holds commission analytics data
type CommissionAnalytics struct {
	TotalLeases         int     `json:"total_leases"`
	TotalCommission     float64 `json:"total_commission"`
	AvgCommission       float64 `json:"avg_commission"`
	DoubleEndedCount    int     `json:"double_ended_count"`
	DoubleEndedRevenue  float64 `json:"double_ended_revenue"`
	DoubleEndedRate     float64 `json:"double_ended_rate"`
	ListingSideCount    int     `json:"listing_side_count"`
	ListingSideRevenue  float64 `json:"listing_side_revenue"`
	TenantSideCount     int     `json:"tenant_side_count"`
	TenantSideRevenue   float64 `json:"tenant_side_revenue"`
}

// MonthlyCommissionSummary holds monthly commission summary data
type MonthlyCommissionSummary struct {
	Year                     int     `json:"year"`
	Month                    int     `json:"month"`
	YearMonth                string  `json:"year_month"`
	TotalLeases              int     `json:"total_leases"`
	TotalCommission          float64 `json:"total_commission"`
	AvgCommission            float64 `json:"avg_commission"`
	DoubleEndedCount         int     `json:"double_ended_count"`
	ListingSideCount         int     `json:"listing_side_count"`
	TenantSideCount          int     `json:"tenant_side_count"`
	DoubleEndedCommission    float64 `json:"double_ended_commission"`
	ListingSideCommission    float64 `json:"listing_side_commission"`
	TenantSideCommission     float64 `json:"tenant_side_commission"`
}

// AgentCommissionPerformance holds agent performance data
type AgentCommissionPerformance struct {
	AgentID              uint    `json:"agent_id"`
	TotalDeals           int     `json:"total_deals"`
	TotalCommission      float64 `json:"total_commission"`
	AvgCommissionPerDeal float64 `json:"avg_commission_per_deal"`
	DoubleEndedDeals     int     `json:"double_ended_deals"`
	ListingSideDeals     int     `json:"listing_side_deals"`
	TenantSideDeals      int     `json:"tenant_side_deals"`
	DoubleEndedRate      float64 `json:"double_ended_rate"`
}
