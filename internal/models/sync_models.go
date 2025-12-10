package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"chrisgross-ctrl-project/internal/security"
	"gorm.io/gorm"
)

type SyncError struct {
	Entity      string    `json:"entity"`
	EntityID    string    `json:"entity_id"`
	Operation   string    `json:"operation"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	IsRetryable bool      `json:"is_retryable"`
}

func (se SyncError) Error() string {
	return se.Message
}

type SyncReport struct {
	ID                  uint         `json:"id" gorm:"primaryKey"`
	StartedAt           time.Time    `json:"started_at" gorm:"not null"`
	CompletedAt         time.Time    `json:"completed_at"`
	Duration            time.Duration `json:"duration" gorm:"-"`
	DurationSeconds     float64      `json:"duration_seconds"`
	PropertiesSynced    int          `json:"properties_synced" gorm:"default:0"`
	TenantsSynced       int          `json:"tenants_synced" gorm:"default:0"`
	LeadsSynced         int          `json:"leads_synced" gorm:"default:0"`
	MaintenanceSynced   int          `json:"maintenance_synced" gorm:"default:0"`
	VacanciesUpdated    int          `json:"vacancies_updated" gorm:"default:0"`
	ErrorCount          int          `json:"error_count" gorm:"default:0"`
	Errors              SyncErrorList `json:"errors" gorm:"type:json"`
	Status              string       `json:"status" gorm:"default:'pending'"`
	SyncType            string       `json:"sync_type"`
	TriggeredBy         string       `json:"triggered_by"`
	FUBLastSync         *time.Time   `json:"fub_last_sync"`
	AppFolioLastSync    *time.Time   `json:"appfolio_last_sync"`
	Notes               string       `json:"notes" gorm:"type:text"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

type SyncErrorList []SyncError

func (sel SyncErrorList) Value() (driver.Value, error) {
	return json.Marshal(sel)
}

func (sel *SyncErrorList) Scan(value interface{}) error {
	if value == nil {
		*sel = make(SyncErrorList, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sel)
}

type PropertySyncResult struct {
	StartedAt        time.Time     `json:"started_at"`
	CompletedAt      time.Time     `json:"completed_at"`
	Duration         time.Duration `json:"duration"`
	Source           string        `json:"source"`
	Synced           int           `json:"synced"`
	Failed           int           `json:"failed"`
	VacanciesUpdated int           `json:"vacancies_updated"`
	Errors           []SyncError   `json:"errors"`
}

type TenantSyncResult struct {
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Duration    time.Duration `json:"duration"`
	Source      string        `json:"source"`
	Synced      int           `json:"synced"`
	Failed      int           `json:"failed"`
	Errors      []SyncError   `json:"errors"`
}

type MaintenanceSyncResult struct {
	StartedAt      time.Time     `json:"started_at"`
	CompletedAt    time.Time     `json:"completed_at"`
	Duration       time.Duration `json:"duration"`
	Source         string        `json:"source"`
	Synced         int           `json:"synced"`
	Failed         int           `json:"failed"`
	EmergencyCount int           `json:"emergency_count"`
	Errors         []SyncError   `json:"errors"`
}

type LeadSyncResult struct {
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Duration    time.Duration `json:"duration"`
	Source      string        `json:"source"`
	Synced      int           `json:"synced"`
	Failed      int           `json:"failed"`
	Errors      []SyncError   `json:"errors"`
}

type SyncQueueItem struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	EntityType  string    `json:"entity_type" gorm:"not null;index"`
	EntityID    string    `json:"entity_id" gorm:"not null"`
	Operation   string    `json:"operation" gorm:"not null"`
	Source      string    `json:"source" gorm:"not null"`
	Destination string    `json:"destination" gorm:"not null"`
	Payload     JSONB     `json:"payload" gorm:"type:json"`
	Priority    int       `json:"priority" gorm:"default:0;index"`
	Status      string    `json:"status" gorm:"default:'pending';index"`
	RetryCount  int       `json:"retry_count" gorm:"default:0"`
	MaxRetries  int       `json:"max_retries" gorm:"default:3"`
	LastError   string    `json:"last_error" gorm:"type:text"`
	ProcessedAt *time.Time `json:"processed_at"`
	ScheduledAt time.Time  `json:"scheduled_at" gorm:"index"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (s *SyncQueueItem) CanRetry() bool {
	return s.RetryCount < s.MaxRetries
}

func (s *SyncQueueItem) IncrementRetry(errMsg string) {
	s.RetryCount++
	s.LastError = errMsg
	s.Status = "retry_pending"
	if s.RetryCount >= s.MaxRetries {
		s.Status = "failed"
	}
}

type AppFolioTenant struct {
	ID            uint                     `json:"id" gorm:"primaryKey"`
	AppFolioID    string                   `json:"appfolio_id" gorm:"uniqueIndex;not null"`
	FirstName     string                   `json:"first_name"`
	LastName      string                   `json:"last_name"`
	Email         security.EncryptedString `json:"email"`
	Phone         security.EncryptedString `json:"phone"`
	PropertyID    string                   `json:"property_id" gorm:"index"`
	UnitID        string                   `json:"unit_id"`
	LeaseStart    time.Time                `json:"lease_start"`
	LeaseEnd      time.Time                `json:"lease_end"`
	RentAmount    float64                  `json:"rent_amount"`
	DepositAmount float64                  `json:"deposit_amount"`
	Status        string                   `json:"status" gorm:"index"`
	MoveInDate    *time.Time               `json:"move_in_date"`
	MoveOutDate   *time.Time               `json:"move_out_date"`
	Balance       float64                  `json:"balance"`
	IsActive      bool                     `json:"is_active" gorm:"index;default:true"`
	FUBContactID  string                   `json:"fub_contact_id"`
	LeadID        *uint                    `json:"lead_id"`
	AppFolioData  JSONB                    `json:"appfolio_data" gorm:"type:json"`
	LastSyncedAt  time.Time                `json:"last_synced_at"`
	SyncErrors    StringArray              `json:"sync_errors" gorm:"type:json"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	DeletedAt     gorm.DeletedAt           `json:"-" gorm:"index"`
}

type MaintenanceRequest struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	AppFolioID        string     `json:"appfolio_id" gorm:"uniqueIndex;not null"`
	PropertyID        string     `json:"property_id" gorm:"index;not null"`
	UnitID            string     `json:"unit_id"`
	TenantID          string     `json:"tenant_id" gorm:"index"`
	TenantName        string     `json:"tenant_name"`
	Category          string     `json:"category" gorm:"index"`
	Priority          string     `json:"priority" gorm:"index"`
	Status            string     `json:"status" gorm:"index;default:'open'"`
	Description       string     `json:"description" gorm:"type:text"`
	RequestedDate     time.Time  `json:"requested_date"`
	ScheduledDate     *time.Time `json:"scheduled_date"`
	CompletedDate     *time.Time `json:"completed_date"`
	AssignedTo        *string    `json:"assigned_to"`
	VendorID          *string    `json:"vendor_id"`
	EstimatedCost     *float64   `json:"estimated_cost"`
	ActualCost        *float64   `json:"actual_cost"`
	Notes             string     `json:"notes" gorm:"type:text"`
	IsEmergency       bool       `json:"is_emergency" gorm:"index;default:false"`
	TriagePriority    string     `json:"triage_priority"`
	AITriageAnalysis  string     `json:"ai_triage_analysis" gorm:"type:text"`
	SuggestedAction   string     `json:"suggested_action"`
	UrgencyScore      int        `json:"urgency_score" gorm:"default:0"`
	OwnerNotified     bool       `json:"owner_notified" gorm:"default:false"`
	ManagerNotified   bool       `json:"manager_notified" gorm:"default:false"`
	AppFolioData      JSONB      `json:"appfolio_data" gorm:"type:json"`
	LastSyncedAt      time.Time  `json:"last_synced_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

type MaintenanceStats struct {
	OpenCount          int64   `json:"open_count"`
	EmergencyCount     int64   `json:"emergency_count"`
	CompletedThisMonth int64   `json:"completed_this_month"`
	AvgResolutionDays  float64 `json:"avg_resolution_days"`
	Source             string  `json:"source,omitempty"`
}

type UnifiedDashboard struct {
	Properties     PropertyStats     `json:"properties"`
	Leads          LeadStats         `json:"leads"`
	Maintenance    MaintenanceStats  `json:"maintenance"`
	Revenue        RevenueStats      `json:"revenue"`
	LastSync       LastSyncInfo      `json:"last_sync"`
	SystemHealth   SystemHealthStats `json:"system_health"`
	GeneratedAt    time.Time         `json:"generated_at"`
}

type PropertyStats struct {
	Total     int64  `json:"total"`
	Vacant    int64  `json:"vacant"`
	Occupied  int64  `json:"occupied"`
	Listed    int64  `json:"listed"`
	Source    string `json:"source"`
}

type LeadStats struct {
	Total     int64  `json:"total"`
	Hot       int64  `json:"hot"`
	Warm      int64  `json:"warm"`
	Cold      int64  `json:"cold"`
	NewToday  int64  `json:"new_today"`
	Source    string `json:"source"`
}

type RevenueStats struct {
	Collected      float64 `json:"collected"`
	Pending        float64 `json:"pending"`
	ProjectedMonth float64 `json:"projected_month"`
	Source         string  `json:"source"`
}

type LastSyncInfo struct {
	FUB               *time.Time `json:"fub"`
	AppFolioProperty  *time.Time `json:"appfolio_property"`
	AppFolioTenant    *time.Time `json:"appfolio_tenant"`
	AppFolioMaintenance *time.Time `json:"appfolio_maintenance"`
	FullSync          *time.Time `json:"full_sync"`
}

type SystemHealthStats struct {
	FUBConnected           bool   `json:"fub_connected"`
	AppFolioConnected      bool   `json:"appfolio_connected"`
	QueuedSyncItems        int64  `json:"queued_sync_items"`
	FailedSyncItems        int64  `json:"failed_sync_items"`
	LastErrorMessage       string `json:"last_error_message,omitempty"`
}

type IntegrationEvent struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	EventType     string    `json:"event_type" gorm:"not null;index"`
	Source        string    `json:"source" gorm:"not null;index"`
	EntityType    string    `json:"entity_type" gorm:"index"`
	EntityID      string    `json:"entity_id"`
	Payload       JSONB     `json:"payload" gorm:"type:json"`
	ProcessedBy   StringArray `json:"processed_by" gorm:"type:json"`
	Status        string    `json:"status" gorm:"default:'pending';index"`
	ProcessedAt   *time.Time `json:"processed_at"`
	ErrorMessage  string    `json:"error_message" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

const (
	EventNewLead           = "new_lead"
	EventLeadUpdated       = "lead_updated"
	EventBookingCreated    = "booking_created"
	EventBookingCompleted  = "booking_completed"
	EventApplicationSubmitted = "application_submitted"
	EventApplicationApproved  = "application_approved"
	EventLeaseConversion   = "lease_conversion"
	EventPropertyVacancy   = "property_vacancy"
	EventPropertyOccupied  = "property_occupied"
	EventMaintenanceCreated = "maintenance_created"
	EventMaintenanceCompleted = "maintenance_completed"
	EventTenantMoveIn      = "tenant_move_in"
	EventTenantMoveOut     = "tenant_move_out"
)

const (
	SourcePropertyHub = "propertyhub"
	SourceFUB         = "fub"
	SourceAppFolio    = "appfolio"
	SourceWebhook     = "webhook"
	SourceManual      = "manual"
)

const (
	SyncStatusPending    = "pending"
	SyncStatusInProgress = "in_progress"
	SyncStatusSuccess    = "success"
	SyncStatusPartial    = "partial"
	SyncStatusFailed     = "failed"
)
