package models

import (
	"time"

	"gorm.io/gorm"
)

type AppFolioPropertyStatus string

const (
	AppFolioPropertyStatusActive   AppFolioPropertyStatus = "active"
	AppFolioPropertyStatusInactive AppFolioPropertyStatus = "inactive"
	AppFolioPropertyStatusPending  AppFolioPropertyStatus = "pending"
	AppFolioPropertyStatusSold     AppFolioPropertyStatus = "sold"
)

type AppFolioMaintenancePriority string

const (
	AppFolioMaintenancePriorityLow       AppFolioMaintenancePriority = "low"
	AppFolioMaintenancePriorityMedium    AppFolioMaintenancePriority = "medium"
	AppFolioMaintenancePriorityHigh      AppFolioMaintenancePriority = "high"
	AppFolioMaintenancePriorityUrgent    AppFolioMaintenancePriority = "urgent"
	AppFolioMaintenancePriorityEmergency AppFolioMaintenancePriority = "emergency"
)

type AppFolioMaintenanceStatus string

const (
	AppFolioMaintenanceStatusOpen       AppFolioMaintenanceStatus = "open"
	AppFolioMaintenanceStatusInProgress AppFolioMaintenanceStatus = "in_progress"
	AppFolioMaintenanceStatusPending    AppFolioMaintenanceStatus = "pending"
	AppFolioMaintenanceStatusCompleted  AppFolioMaintenanceStatus = "completed"
	AppFolioMaintenanceStatusCancelled  AppFolioMaintenanceStatus = "cancelled"
)

type AppFolioPaymentStatus string

const (
	AppFolioPaymentStatusPending   AppFolioPaymentStatus = "pending"
	AppFolioPaymentStatusPaid      AppFolioPaymentStatus = "paid"
	AppFolioPaymentStatusPartial   AppFolioPaymentStatus = "partial"
	AppFolioPaymentStatusOverdue   AppFolioPaymentStatus = "overdue"
	AppFolioPaymentStatusCancelled AppFolioPaymentStatus = "cancelled"
	AppFolioPaymentStatusRefunded  AppFolioPaymentStatus = "refunded"
)

type AppFolioLeaseStatus string

const (
	AppFolioLeaseStatusActive     AppFolioLeaseStatus = "active"
	AppFolioLeaseStatusExpired    AppFolioLeaseStatus = "expired"
	AppFolioLeaseStatusPending    AppFolioLeaseStatus = "pending"
	AppFolioLeaseStatusTerminated AppFolioLeaseStatus = "terminated"
)

type AppFolioAddress struct {
	Street1    string `json:"street_1"`
	Street2    string `json:"street_2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country,omitempty"`
}

func (a AppFolioAddress) FullAddress() string {
	addr := a.Street1
	if a.Street2 != "" {
		addr += ", " + a.Street2
	}
	addr += ", " + a.City + ", " + a.State + " " + a.PostalCode
	if a.Country != "" {
		addr += ", " + a.Country
	}
	return addr
}

type AppFolioProperty struct {
	ID                 uint                   `json:"id" gorm:"primaryKey"`
	AppFolioPropertyID string                 `json:"appfolio_property_id" gorm:"uniqueIndex;not null"`
	Name               string                 `json:"name"`
	PropertyType       string                 `json:"property_type"`
	Address            AppFolioAddress        `json:"address" gorm:"embedded;embeddedPrefix:address_"`
	UnitCount          int                    `json:"unit_count" gorm:"default:1"`
	Status             AppFolioPropertyStatus `json:"status" gorm:"default:'active'"`
	RentAmount         float64                `json:"rent_amount"`
	DepositAmount      float64                `json:"deposit_amount"`
	SquareFootage      int                    `json:"square_footage"`
	Bedrooms           int                    `json:"bedrooms"`
	Bathrooms          float64                `json:"bathrooms"`
	YearBuilt          int                    `json:"year_built"`
	Description        string                 `json:"description" gorm:"type:text"`
	Amenities          []string               `json:"amenities" gorm:"type:text[]"`
	OwnerID            string                 `json:"owner_id" gorm:"index"`
	PropertyManagerID  string                 `json:"property_manager_id"`
	MarketRent         float64                `json:"market_rent"`
	LastInspectionDate *time.Time             `json:"last_inspection_date"`
	NextInspectionDate *time.Time             `json:"next_inspection_date"`
	CustomFields       JSONMap                `json:"custom_fields" gorm:"type:jsonb"`
	AppFolioCreatedAt  time.Time              `json:"appfolio_created_at"`
	AppFolioUpdatedAt  time.Time              `json:"appfolio_updated_at"`
	LastSyncedAt       time.Time              `json:"last_synced_at"`
	SyncErrors         []string               `json:"sync_errors" gorm:"type:text[]"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	DeletedAt          gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`
}

func (p AppFolioProperty) IsActive() bool {
	return p.Status == AppFolioPropertyStatusActive && p.DeletedAt.Time.IsZero()
}

func (p AppFolioProperty) GetFullAddress() string {
	return p.Address.FullAddress()
}

func (p AppFolioProperty) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                   p.ID,
		"appfolio_property_id": p.AppFolioPropertyID,
		"name":                 p.Name,
		"property_type":        p.PropertyType,
		"address":              p.Address,
		"full_address":         p.GetFullAddress(),
		"unit_count":           p.UnitCount,
		"status":               p.Status,
		"rent_amount":          p.RentAmount,
		"deposit_amount":       p.DepositAmount,
		"square_footage":       p.SquareFootage,
		"bedrooms":             p.Bedrooms,
		"bathrooms":            p.Bathrooms,
		"year_built":           p.YearBuilt,
		"description":          p.Description,
		"amenities":            p.Amenities,
		"owner_id":             p.OwnerID,
		"property_manager_id":  p.PropertyManagerID,
		"market_rent":          p.MarketRent,
		"last_inspection_date": p.LastInspectionDate,
		"next_inspection_date": p.NextInspectionDate,
		"custom_fields":        p.CustomFields,
		"appfolio_created_at":  p.AppFolioCreatedAt,
		"appfolio_updated_at":  p.AppFolioUpdatedAt,
		"last_synced_at":       p.LastSyncedAt,
		"is_active":            p.IsActive(),
		"created_at":           p.CreatedAt,
		"updated_at":           p.UpdatedAt,
	}
}

type AppFolioTenant struct {
	ID                uint                `json:"id" gorm:"primaryKey"`
	AppFolioTenantID  string              `json:"appfolio_tenant_id" gorm:"uniqueIndex;not null"`
	FirstName         string              `json:"first_name"`
	LastName          string              `json:"last_name"`
	Email             string              `json:"email" gorm:"index"`
	Phone             string              `json:"phone" gorm:"index"`
	AlternatePhone    string              `json:"alternate_phone"`
	Address           AppFolioAddress     `json:"address" gorm:"embedded;embeddedPrefix:address_"`
	UnitID            string              `json:"unit_id" gorm:"index"`
	PropertyID        string              `json:"property_id" gorm:"index"`
	LeaseStatus       AppFolioLeaseStatus `json:"lease_status" gorm:"default:'pending'"`
	LeaseStartDate    *time.Time          `json:"lease_start_date"`
	LeaseEndDate      *time.Time          `json:"lease_end_date"`
	MoveInDate        *time.Time          `json:"move_in_date"`
	MoveOutDate       *time.Time          `json:"move_out_date"`
	RentAmount        float64             `json:"rent_amount"`
	SecurityDeposit   float64             `json:"security_deposit"`
	Balance           float64             `json:"balance"`
	EmergencyContact  string              `json:"emergency_contact"`
	EmergencyPhone    string              `json:"emergency_phone"`
	Notes             string              `json:"notes" gorm:"type:text"`
	CustomFields      JSONMap             `json:"custom_fields" gorm:"type:jsonb"`
	AppFolioCreatedAt time.Time           `json:"appfolio_created_at"`
	AppFolioUpdatedAt time.Time           `json:"appfolio_updated_at"`
	LastSyncedAt      time.Time           `json:"last_synced_at"`
	SyncErrors        []string            `json:"sync_errors" gorm:"type:text[]"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	DeletedAt         gorm.DeletedAt      `json:"deleted_at,omitempty" gorm:"index"`
}

func (t AppFolioTenant) GetFullName() string {
	if t.FirstName == "" && t.LastName == "" {
		return "Unknown"
	}
	if t.FirstName == "" {
		return t.LastName
	}
	if t.LastName == "" {
		return t.FirstName
	}
	return t.FirstName + " " + t.LastName
}

func (t AppFolioTenant) HasContactInfo() bool {
	return t.Email != "" || t.Phone != ""
}

func (t AppFolioTenant) IsLeaseActive() bool {
	return t.LeaseStatus == AppFolioLeaseStatusActive && t.DeletedAt.Time.IsZero()
}

func (t AppFolioTenant) IsLeaseExpiringSoon(days int) bool {
	if t.LeaseEndDate == nil {
		return false
	}
	expirationThreshold := time.Now().AddDate(0, 0, days)
	return t.LeaseEndDate.Before(expirationThreshold) && t.LeaseEndDate.After(time.Now())
}

func (t AppFolioTenant) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                  t.ID,
		"appfolio_tenant_id":  t.AppFolioTenantID,
		"first_name":          t.FirstName,
		"last_name":           t.LastName,
		"full_name":           t.GetFullName(),
		"email":               t.Email,
		"phone":               t.Phone,
		"alternate_phone":     t.AlternatePhone,
		"address":             t.Address,
		"unit_id":             t.UnitID,
		"property_id":         t.PropertyID,
		"lease_status":        t.LeaseStatus,
		"lease_start_date":    t.LeaseStartDate,
		"lease_end_date":      t.LeaseEndDate,
		"move_in_date":        t.MoveInDate,
		"move_out_date":       t.MoveOutDate,
		"rent_amount":         t.RentAmount,
		"security_deposit":    t.SecurityDeposit,
		"balance":             t.Balance,
		"emergency_contact":   t.EmergencyContact,
		"emergency_phone":     t.EmergencyPhone,
		"notes":               t.Notes,
		"custom_fields":       t.CustomFields,
		"appfolio_created_at": t.AppFolioCreatedAt,
		"appfolio_updated_at": t.AppFolioUpdatedAt,
		"last_synced_at":      t.LastSyncedAt,
		"is_lease_active":     t.IsLeaseActive(),
		"has_contact_info":    t.HasContactInfo(),
		"created_at":          t.CreatedAt,
		"updated_at":          t.UpdatedAt,
	}
}

type AppFolioMaintenanceRequest struct {
	ID                uint                        `json:"id" gorm:"primaryKey"`
	AppFolioRequestID string                      `json:"appfolio_request_id" gorm:"uniqueIndex;not null"`
	PropertyID        string                      `json:"property_id" gorm:"index"`
	UnitID            string                      `json:"unit_id" gorm:"index"`
	TenantID          string                      `json:"tenant_id" gorm:"index"`
	RequestedBy       string                      `json:"requested_by"`
	RequestedByEmail  string                      `json:"requested_by_email"`
	RequestedByPhone  string                      `json:"requested_by_phone"`
	Category          string                      `json:"category"`
	SubCategory       string                      `json:"sub_category"`
	Description       string                      `json:"description" gorm:"type:text"`
	Priority          AppFolioMaintenancePriority `json:"priority" gorm:"default:'medium'"`
	Status            AppFolioMaintenanceStatus   `json:"status" gorm:"default:'open'"`
	AssignedTo        string                      `json:"assigned_to"`
	AssignedVendorID  string                      `json:"assigned_vendor_id"`
	ScheduledDate     *time.Time                  `json:"scheduled_date"`
	CompletedDate     *time.Time                  `json:"completed_date"`
	EstimatedCost     float64                     `json:"estimated_cost"`
	ActualCost        float64                     `json:"actual_cost"`
	PermissionToEnter bool                        `json:"permission_to_enter" gorm:"default:false"`
	EntryInstructions string                      `json:"entry_instructions" gorm:"type:text"`
	InternalNotes     string                      `json:"internal_notes" gorm:"type:text"`
	ResolutionNotes   string                      `json:"resolution_notes" gorm:"type:text"`
	Attachments       []string                    `json:"attachments" gorm:"type:text[]"`
	CustomFields      JSONMap                     `json:"custom_fields" gorm:"type:jsonb"`
	AppFolioCreatedAt time.Time                   `json:"appfolio_created_at"`
	AppFolioUpdatedAt time.Time                   `json:"appfolio_updated_at"`
	LastSyncedAt      time.Time                   `json:"last_synced_at"`
	SyncErrors        []string                    `json:"sync_errors" gorm:"type:text[]"`
	CreatedAt         time.Time                   `json:"created_at"`
	UpdatedAt         time.Time                   `json:"updated_at"`
	DeletedAt         gorm.DeletedAt              `json:"deleted_at,omitempty" gorm:"index"`
}

func (m AppFolioMaintenanceRequest) IsOpen() bool {
	return m.Status == AppFolioMaintenanceStatusOpen || m.Status == AppFolioMaintenanceStatusInProgress
}

func (m AppFolioMaintenanceRequest) IsHighPriority() bool {
	return m.Priority == AppFolioMaintenancePriorityHigh ||
		m.Priority == AppFolioMaintenancePriorityUrgent ||
		m.Priority == AppFolioMaintenancePriorityEmergency
}

func (m AppFolioMaintenanceRequest) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                  m.ID,
		"appfolio_request_id": m.AppFolioRequestID,
		"property_id":         m.PropertyID,
		"unit_id":             m.UnitID,
		"tenant_id":           m.TenantID,
		"requested_by":        m.RequestedBy,
		"requested_by_email":  m.RequestedByEmail,
		"requested_by_phone":  m.RequestedByPhone,
		"category":            m.Category,
		"sub_category":        m.SubCategory,
		"description":         m.Description,
		"priority":            m.Priority,
		"status":              m.Status,
		"assigned_to":         m.AssignedTo,
		"assigned_vendor_id":  m.AssignedVendorID,
		"scheduled_date":      m.ScheduledDate,
		"completed_date":      m.CompletedDate,
		"estimated_cost":      m.EstimatedCost,
		"actual_cost":         m.ActualCost,
		"permission_to_enter": m.PermissionToEnter,
		"entry_instructions":  m.EntryInstructions,
		"internal_notes":      m.InternalNotes,
		"resolution_notes":    m.ResolutionNotes,
		"attachments":         m.Attachments,
		"custom_fields":       m.CustomFields,
		"appfolio_created_at": m.AppFolioCreatedAt,
		"appfolio_updated_at": m.AppFolioUpdatedAt,
		"last_synced_at":      m.LastSyncedAt,
		"is_open":             m.IsOpen(),
		"is_high_priority":    m.IsHighPriority(),
		"created_at":          m.CreatedAt,
		"updated_at":          m.UpdatedAt,
	}
}

type AppFolioPayment struct {
	ID                uint                  `json:"id" gorm:"primaryKey"`
	AppFolioPaymentID string                `json:"appfolio_payment_id" gorm:"uniqueIndex;not null"`
	TenantID          string                `json:"tenant_id" gorm:"index"`
	PropertyID        string                `json:"property_id" gorm:"index"`
	UnitID            string                `json:"unit_id" gorm:"index"`
	LeaseID           string                `json:"lease_id" gorm:"index"`
	PaymentType       string                `json:"payment_type"`
	Amount            float64               `json:"amount"`
	DueDate           time.Time             `json:"due_date"`
	PaidDate          *time.Time            `json:"paid_date"`
	Status            AppFolioPaymentStatus `json:"status" gorm:"default:'pending'"`
	PaymentMethod     string                `json:"payment_method"`
	TransactionID     string                `json:"transaction_id"`
	CheckNumber       string                `json:"check_number"`
	BankAccountLast4  string                `json:"bank_account_last_4"`
	LateFee           float64               `json:"late_fee"`
	LateFeeApplied    bool                  `json:"late_fee_applied" gorm:"default:false"`
	Notes             string                `json:"notes" gorm:"type:text"`
	ReceiptURL        string                `json:"receipt_url"`
	CustomFields      JSONMap               `json:"custom_fields" gorm:"type:jsonb"`
	AppFolioCreatedAt time.Time             `json:"appfolio_created_at"`
	AppFolioUpdatedAt time.Time             `json:"appfolio_updated_at"`
	LastSyncedAt      time.Time             `json:"last_synced_at"`
	SyncErrors        []string              `json:"sync_errors" gorm:"type:text[]"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	DeletedAt         gorm.DeletedAt        `json:"deleted_at,omitempty" gorm:"index"`
}

func (p AppFolioPayment) IsPaid() bool {
	return p.Status == AppFolioPaymentStatusPaid
}

func (p AppFolioPayment) IsOverdue() bool {
	if p.Status == AppFolioPaymentStatusPaid {
		return false
	}
	return time.Now().After(p.DueDate)
}

func (p AppFolioPayment) DaysOverdue() int {
	if !p.IsOverdue() {
		return 0
	}
	return int(time.Since(p.DueDate).Hours() / 24)
}

func (p AppFolioPayment) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                  p.ID,
		"appfolio_payment_id": p.AppFolioPaymentID,
		"tenant_id":           p.TenantID,
		"property_id":         p.PropertyID,
		"unit_id":             p.UnitID,
		"lease_id":            p.LeaseID,
		"payment_type":        p.PaymentType,
		"amount":              p.Amount,
		"due_date":            p.DueDate,
		"paid_date":           p.PaidDate,
		"status":              p.Status,
		"payment_method":      p.PaymentMethod,
		"transaction_id":      p.TransactionID,
		"check_number":        p.CheckNumber,
		"bank_account_last_4": p.BankAccountLast4,
		"late_fee":            p.LateFee,
		"late_fee_applied":    p.LateFeeApplied,
		"notes":               p.Notes,
		"receipt_url":         p.ReceiptURL,
		"custom_fields":       p.CustomFields,
		"appfolio_created_at": p.AppFolioCreatedAt,
		"appfolio_updated_at": p.AppFolioUpdatedAt,
		"last_synced_at":      p.LastSyncedAt,
		"is_paid":             p.IsPaid(),
		"is_overdue":          p.IsOverdue(),
		"days_overdue":        p.DaysOverdue(),
		"created_at":          p.CreatedAt,
		"updated_at":          p.UpdatedAt,
	}
}

type AppFolioOwner struct {
	ID                 uint            `json:"id" gorm:"primaryKey"`
	AppFolioOwnerID    string          `json:"appfolio_owner_id" gorm:"uniqueIndex;not null"`
	FirstName          string          `json:"first_name"`
	LastName           string          `json:"last_name"`
	CompanyName        string          `json:"company_name"`
	Email              string          `json:"email" gorm:"index"`
	Phone              string          `json:"phone" gorm:"index"`
	AlternatePhone     string          `json:"alternate_phone"`
	Address            AppFolioAddress `json:"address" gorm:"embedded;embeddedPrefix:address_"`
	TaxID              string          `json:"tax_id"`
	PropertyIDs        []string        `json:"property_ids" gorm:"type:text[]"`
	PropertyCount      int             `json:"property_count" gorm:"default:0"`
	TotalUnits         int             `json:"total_units" gorm:"default:0"`
	ManagementFeeRate  float64         `json:"management_fee_rate"`
	PaymentMethod      string          `json:"payment_method"`
	BankName           string          `json:"bank_name"`
	BankAccountLast4   string          `json:"bank_account_last_4"`
	RoutingNumberLast4 string          `json:"routing_number_last_4"`
	Notes              string          `json:"notes" gorm:"type:text"`
	CustomFields       JSONMap         `json:"custom_fields" gorm:"type:jsonb"`
	AppFolioCreatedAt  time.Time       `json:"appfolio_created_at"`
	AppFolioUpdatedAt  time.Time       `json:"appfolio_updated_at"`
	LastSyncedAt       time.Time       `json:"last_synced_at"`
	SyncErrors         []string        `json:"sync_errors" gorm:"type:text[]"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`
}

func (o AppFolioOwner) GetDisplayName() string {
	if o.CompanyName != "" {
		return o.CompanyName
	}
	if o.FirstName == "" && o.LastName == "" {
		return "Unknown"
	}
	if o.FirstName == "" {
		return o.LastName
	}
	if o.LastName == "" {
		return o.FirstName
	}
	return o.FirstName + " " + o.LastName
}

func (o AppFolioOwner) HasContactInfo() bool {
	return o.Email != "" || o.Phone != ""
}

func (o AppFolioOwner) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                    o.ID,
		"appfolio_owner_id":     o.AppFolioOwnerID,
		"first_name":            o.FirstName,
		"last_name":             o.LastName,
		"company_name":          o.CompanyName,
		"display_name":          o.GetDisplayName(),
		"email":                 o.Email,
		"phone":                 o.Phone,
		"alternate_phone":       o.AlternatePhone,
		"address":               o.Address,
		"tax_id":                o.TaxID,
		"property_ids":          o.PropertyIDs,
		"property_count":        o.PropertyCount,
		"total_units":           o.TotalUnits,
		"management_fee_rate":   o.ManagementFeeRate,
		"payment_method":        o.PaymentMethod,
		"bank_name":             o.BankName,
		"bank_account_last_4":   o.BankAccountLast4,
		"routing_number_last_4": o.RoutingNumberLast4,
		"notes":                 o.Notes,
		"custom_fields":         o.CustomFields,
		"appfolio_created_at":   o.AppFolioCreatedAt,
		"appfolio_updated_at":   o.AppFolioUpdatedAt,
		"last_synced_at":        o.LastSyncedAt,
		"has_contact_info":      o.HasContactInfo(),
		"created_at":            o.CreatedAt,
		"updated_at":            o.UpdatedAt,
	}
}

type AppFolioUnit struct {
	ID                uint                   `json:"id" gorm:"primaryKey"`
	AppFolioUnitID    string                 `json:"appfolio_unit_id" gorm:"uniqueIndex;not null"`
	PropertyID        string                 `json:"property_id" gorm:"index;not null"`
	UnitNumber        string                 `json:"unit_number"`
	UnitType          string                 `json:"unit_type"`
	Status            AppFolioPropertyStatus `json:"status" gorm:"default:'active'"`
	Bedrooms          int                    `json:"bedrooms"`
	Bathrooms         float64                `json:"bathrooms"`
	SquareFootage     int                    `json:"square_footage"`
	MarketRent        float64                `json:"market_rent"`
	CurrentRent       float64                `json:"current_rent"`
	DepositAmount     float64                `json:"deposit_amount"`
	IsOccupied        bool                   `json:"is_occupied" gorm:"default:false"`
	CurrentTenantID   string                 `json:"current_tenant_id"`
	FloorPlan         string                 `json:"floor_plan"`
	Amenities         []string               `json:"amenities" gorm:"type:text[]"`
	Notes             string                 `json:"notes" gorm:"type:text"`
	CustomFields      JSONMap                `json:"custom_fields" gorm:"type:jsonb"`
	AppFolioCreatedAt time.Time              `json:"appfolio_created_at"`
	AppFolioUpdatedAt time.Time              `json:"appfolio_updated_at"`
	LastSyncedAt      time.Time              `json:"last_synced_at"`
	SyncErrors        []string               `json:"sync_errors" gorm:"type:text[]"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	DeletedAt         gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`
}

func (u AppFolioUnit) IsAvailable() bool {
	return !u.IsOccupied && u.Status == AppFolioPropertyStatusActive
}

func (u AppFolioUnit) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                  u.ID,
		"appfolio_unit_id":    u.AppFolioUnitID,
		"property_id":         u.PropertyID,
		"unit_number":         u.UnitNumber,
		"unit_type":           u.UnitType,
		"status":              u.Status,
		"bedrooms":            u.Bedrooms,
		"bathrooms":           u.Bathrooms,
		"square_footage":      u.SquareFootage,
		"market_rent":         u.MarketRent,
		"current_rent":        u.CurrentRent,
		"deposit_amount":      u.DepositAmount,
		"is_occupied":         u.IsOccupied,
		"current_tenant_id":   u.CurrentTenantID,
		"floor_plan":          u.FloorPlan,
		"amenities":           u.Amenities,
		"notes":               u.Notes,
		"custom_fields":       u.CustomFields,
		"appfolio_created_at": u.AppFolioCreatedAt,
		"appfolio_updated_at": u.AppFolioUpdatedAt,
		"last_synced_at":      u.LastSyncedAt,
		"is_available":        u.IsAvailable(),
		"created_at":          u.CreatedAt,
		"updated_at":          u.UpdatedAt,
	}
}

type AppFolioAPIListResponse[T any] struct {
	Data       []T  `json:"data"`
	TotalCount int  `json:"total_count"`
	PageSize   int  `json:"page_size"`
	Page       int  `json:"page"`
	HasMore    bool `json:"has_more"`
}

type AppFolioAPISingleResponse[T any] struct {
	Data T `json:"data"`
}

type AppFolioAPIErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type AppFolioWebhookEvent struct {
	ID           string                 `json:"id"`
	EventType    string                 `json:"event_type"`
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Data         map[string]interface{} `json:"data"`
	Timestamp    time.Time              `json:"timestamp"`
	Signature    string                 `json:"signature"`
}
