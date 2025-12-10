package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

const (
	MaintenancePriorityEmergency = "emergency"
	MaintenancePriorityHigh      = "high"
	MaintenancePriorityMedium    = "medium"
	MaintenancePriorityLow       = "low"

	MaintenanceCategoryPlumbing   = "plumbing"
	MaintenanceCategoryElectrical = "electrical"
	MaintenanceCategoryHVAC       = "hvac"
	MaintenanceCategoryAppliance  = "appliance"
	MaintenanceCategoryGeneral    = "general"
	MaintenanceCategoryStructural = "structural"
	MaintenanceCategoryPest       = "pest"

	MaintenanceStatusOpen       = "open"
	MaintenanceStatusInProgress = "in_progress"
	MaintenanceStatusCompleted  = "completed"
	MaintenanceStatusCancelled  = "cancelled"

	ResponseTimeImmediate = "immediate"
	ResponseTime24Hours   = "24h"
	ResponseTime48Hours   = "48h"
	ResponseTimeScheduled = "scheduled"
)

type MaintenanceRequest struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	AppFolioID      string     `json:"appfolio_id" gorm:"uniqueIndex;not null"`
	PropertyID      *uint      `json:"property_id" gorm:"index"`
	TenantID        *uint      `json:"tenant_id" gorm:"index"`
	TenantName      string     `json:"tenant_name"`
	TenantPhone     string     `json:"tenant_phone"`
	TenantEmail     string     `json:"tenant_email"`
	PropertyAddress string     `json:"property_address" gorm:"not null"`
	UnitNumber      string     `json:"unit_number"`
	Description     string     `json:"description" gorm:"type:text;not null"`
	Category        string     `json:"category" gorm:"default:'general'"`
	Priority        string     `json:"priority" gorm:"default:'medium';index"`
	Status          string     `json:"status" gorm:"default:'open';index"`
	SuggestedVendor string     `json:"suggested_vendor"`
	AssignedVendor  string     `json:"assigned_vendor"`
	AssignedVendorID *uint     `json:"assigned_vendor_id" gorm:"index"`
	AITriageResult  TriageJSON `json:"ai_triage_result" gorm:"type:jsonb"`
	EstimatedCost   *float64   `json:"estimated_cost"`
	ActualCost      *float64   `json:"actual_cost"`
	ResponseTime    string     `json:"response_time"`
	ScheduledDate   *time.Time `json:"scheduled_date"`
	CompletedDate   *time.Time `json:"completed_date"`
	Notes           string     `json:"notes" gorm:"type:text"`
	InternalNotes   string     `json:"internal_notes" gorm:"type:text"`
	PermissionToEnter bool     `json:"permission_to_enter" gorm:"default:false"`
	PetOnPremises   bool       `json:"pet_on_premises" gorm:"default:false"`
	LastSyncedAt    *time.Time `json:"last_synced_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	Property *Property `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
	Vendor   *Vendor   `json:"vendor,omitempty" gorm:"foreignKey:AssignedVendorID"`
}

func (MaintenanceRequest) TableName() string {
	return "maintenance_requests"
}

func (m MaintenanceRequest) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                  m.ID,
		"appfolio_id":         m.AppFolioID,
		"property_id":         m.PropertyID,
		"tenant_id":           m.TenantID,
		"tenant_name":         m.TenantName,
		"tenant_phone":        m.TenantPhone,
		"tenant_email":        m.TenantEmail,
		"property_address":    m.PropertyAddress,
		"unit_number":         m.UnitNumber,
		"description":         m.Description,
		"category":            m.Category,
		"priority":            m.Priority,
		"status":              m.Status,
		"suggested_vendor":    m.SuggestedVendor,
		"assigned_vendor":     m.AssignedVendor,
		"assigned_vendor_id":  m.AssignedVendorID,
		"ai_triage_result":    m.AITriageResult,
		"estimated_cost":      m.EstimatedCost,
		"actual_cost":         m.ActualCost,
		"response_time":       m.ResponseTime,
		"scheduled_date":      m.ScheduledDate,
		"completed_date":      m.CompletedDate,
		"notes":               m.Notes,
		"permission_to_enter": m.PermissionToEnter,
		"pet_on_premises":     m.PetOnPremises,
		"last_synced_at":      m.LastSyncedAt,
		"resolved_at":         m.ResolvedAt,
		"created_at":          m.CreatedAt,
		"updated_at":          m.UpdatedAt,
	}
}

func (m MaintenanceRequest) IsEmergency() bool {
	return m.Priority == MaintenancePriorityEmergency
}

func (m MaintenanceRequest) IsOpen() bool {
	return m.Status == MaintenanceStatusOpen || m.Status == MaintenanceStatusInProgress
}

func (m MaintenanceRequest) GetPriorityColor() string {
	switch m.Priority {
	case MaintenancePriorityEmergency:
		return "red"
	case MaintenancePriorityHigh:
		return "orange"
	case MaintenancePriorityMedium:
		return "yellow"
	case MaintenancePriorityLow:
		return "green"
	default:
		return "gray"
	}
}

type TriageJSON struct {
	Priority        string   `json:"priority"`
	Category        string   `json:"category"`
	SuggestedVendor string   `json:"suggested_vendor"`
	EstimatedCost   float64  `json:"estimated_cost"`
	ResponseTime    string   `json:"response_time"`
	AIReasoning     string   `json:"ai_reasoning"`
	Keywords        []string `json:"keywords"`
	ConfidenceScore float64  `json:"confidence_score"`
	TriagedAt       string   `json:"triaged_at"`
}

func (t TriageJSON) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TriageJSON) Scan(value interface{}) error {
	if value == nil {
		*t = TriageJSON{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

type Vendor struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	CompanyName  string    `json:"company_name"`
	Category     string    `json:"category" gorm:"not null;index"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	Address      string    `json:"address"`
	HourlyRate   float64   `json:"hourly_rate"`
	MinimumCharge float64  `json:"minimum_charge"`
	IsPreferred  bool      `json:"is_preferred" gorm:"default:false;index"`
	IsActive     bool      `json:"is_active" gorm:"default:true;index"`
	Rating       float64   `json:"rating" gorm:"default:0"`
	TotalJobs    int       `json:"total_jobs" gorm:"default:0"`
	AvgResponseTime int    `json:"avg_response_time"`
	Notes        string    `json:"notes" gorm:"type:text"`
	LicenseNumber string   `json:"license_number"`
	InsuranceExpiry *time.Time `json:"insurance_expiry"`
	ServiceAreas StringArray `json:"service_areas" gorm:"type:json"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Vendor) TableName() string {
	return "vendors"
}

func (v Vendor) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":               v.ID,
		"name":             v.Name,
		"company_name":     v.CompanyName,
		"category":         v.Category,
		"phone":            v.Phone,
		"email":            v.Email,
		"address":          v.Address,
		"hourly_rate":      v.HourlyRate,
		"minimum_charge":   v.MinimumCharge,
		"is_preferred":     v.IsPreferred,
		"is_active":        v.IsActive,
		"rating":           v.Rating,
		"total_jobs":       v.TotalJobs,
		"avg_response_time": v.AvgResponseTime,
		"notes":            v.Notes,
		"license_number":   v.LicenseNumber,
		"insurance_expiry": v.InsuranceExpiry,
		"service_areas":    v.ServiceAreas,
		"created_at":       v.CreatedAt,
		"updated_at":       v.UpdatedAt,
	}
}

type MaintenanceStatusLog struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	MaintenanceRequestID uint      `json:"maintenance_request_id" gorm:"not null;index"`
	OldStatus            string    `json:"old_status"`
	NewStatus            string    `json:"new_status" gorm:"not null"`
	ChangedBy            string    `json:"changed_by"`
	Notes                string    `json:"notes" gorm:"type:text"`
	CreatedAt            time.Time `json:"created_at"`
}

func (MaintenanceStatusLog) TableName() string {
	return "maintenance_status_logs"
}

type MaintenanceAlert struct {
	ID                   uint       `json:"id" gorm:"primaryKey"`
	MaintenanceRequestID uint       `json:"maintenance_request_id" gorm:"not null;index"`
	AlertType            string     `json:"alert_type" gorm:"not null"`
	Message              string     `json:"message" gorm:"type:text"`
	SentAt               *time.Time `json:"sent_at"`
	AcknowledgedAt       *time.Time `json:"acknowledged_at"`
	AcknowledgedBy       string     `json:"acknowledged_by"`
	CreatedAt            time.Time  `json:"created_at"`
}

func (MaintenanceAlert) TableName() string {
	return "maintenance_alerts"
}
