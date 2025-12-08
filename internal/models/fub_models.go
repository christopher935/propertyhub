package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONMap for flexible JSON storage in FUB models
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// FUBLead represents a lead from FollowUp Boss
// This is the central model for all FUB lead data
type FUBLead struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Core FUB identifiers
	FUBLeadID   string `json:"fub_lead_id" gorm:"uniqueIndex;not null"`
	FUBPersonID string `json:"fub_person_id" gorm:"index"`

	// Lead Information
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" gorm:"index"`
	Phone     string `json:"phone" gorm:"index"`

	// FUB-specific data
	Status       string   `json:"status"`
	Stage        string   `json:"stage"`
	Source       string   `json:"source"`
	Tags         []string `json:"tags" gorm:"type:text[]"`
	CustomFields JSONMap  `json:"custom_fields" gorm:"type:jsonb"`

	// Agent assignment
	AgentID    string `json:"agent_id"`
	AgentEmail string `json:"agent_email"`

	// Timestamps from FUB
	FUBCreatedAt time.Time  `json:"fub_created_at"`
	FUBUpdatedAt time.Time  `json:"fub_updated_at"`
	LastActivity *time.Time `json:"last_activity"`

	// Sync tracking
	LastSyncedAt time.Time `json:"last_synced_at"`
	SyncErrors   []string  `json:"sync_errors" gorm:"type:text[]"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// FUBAutomationTrigger handles FUB automation rule triggers
type FUBAutomationTrigger struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Trigger identification
	TriggerName string `json:"trigger_name" gorm:"not null"`
	FUBRuleID   string `json:"fub_rule_id"`

	// Conditions
	Conditions JSONMap `json:"conditions" gorm:"type:jsonb"`

	// Actions to execute
	Actions []string `json:"actions" gorm:"type:text[]"`

	// Targeting
	LeadFilters    JSONMap `json:"lead_filters" gorm:"type:jsonb"`
	PropertyFilter JSONMap `json:"property_filter" gorm:"type:jsonb"`

	// Execution tracking
	LastExecuted   *time.Time `json:"last_executed"`
	ExecutionCount int        `json:"execution_count" gorm:"default:0"`
	SuccessCount   int        `json:"success_count" gorm:"default:0"`
	ErrorCount     int        `json:"error_count" gorm:"default:0"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Helper methods for FUBLead

func (fl FUBLead) GetFullName() string {
	if fl.FirstName == "" && fl.LastName == "" {
		return "Unknown"
	}
	if fl.FirstName == "" {
		return fl.LastName
	}
	if fl.LastName == "" {
		return fl.FirstName
	}
	return fl.FirstName + " " + fl.LastName
}

func (fl FUBLead) HasContactInfo() bool {
	return fl.Email != "" || fl.Phone != ""
}

func (fl FUBLead) IsActive() bool {
	return fl.Status != "archived" && fl.DeletedAt.Time.IsZero()
}

// ToDict converts FUBLead to a map for JSON responses
func (fl FUBLead) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":             fl.ID,
		"fub_lead_id":    fl.FUBLeadID,
		"fub_person_id":  fl.FUBPersonID,
		"first_name":     fl.FirstName,
		"last_name":      fl.LastName,
		"full_name":      fl.GetFullName(),
		"email":          fl.Email,
		"phone":          fl.Phone,
		"status":         fl.Status,
		"stage":          fl.Stage,
		"source":         fl.Source,
		"tags":           fl.Tags,
		"custom_fields":  fl.CustomFields,
		"agent_id":       fl.AgentID,
		"agent_email":    fl.AgentEmail,
		"fub_created_at": fl.FUBCreatedAt,
		"fub_updated_at": fl.FUBUpdatedAt,
		"last_activity":  fl.LastActivity,
		"last_synced_at": fl.LastSyncedAt,
		"sync_errors":    fl.SyncErrors,
		"created_at":     fl.CreatedAt,
		"updated_at":     fl.UpdatedAt,
		"is_active":      fl.IsActive(),
		"has_contact":    fl.HasContactInfo(),
	}
}

// Helper methods for FUBAutomationTrigger

func (fat FUBAutomationTrigger) GetSuccessRate() float64 {
	if fat.ExecutionCount == 0 {
		return 0.0
	}
	return float64(fat.SuccessCount) / float64(fat.ExecutionCount) * 100
}

func (fat FUBAutomationTrigger) GetErrorRate() float64 {
	if fat.ExecutionCount == 0 {
		return 0.0
	}
	return float64(fat.ErrorCount) / float64(fat.ExecutionCount) * 100
}

// ToDict converts FUBAutomationTrigger to a map for JSON responses
func (fat FUBAutomationTrigger) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":              fat.ID,
		"trigger_name":    fat.TriggerName,
		"fub_rule_id":     fat.FUBRuleID,
		"conditions":      fat.Conditions,
		"actions":         fat.Actions,
		"lead_filters":    fat.LeadFilters,
		"property_filter": fat.PropertyFilter,
		"last_executed":   fat.LastExecuted,
		"execution_count": fat.ExecutionCount,
		"success_count":   fat.SuccessCount,
		"error_count":     fat.ErrorCount,
		"success_rate":    fat.GetSuccessRate(),
		"error_rate":      fat.GetErrorRate(),
		"is_active":       fat.IsActive,
		"created_at":      fat.CreatedAt,
		"updated_at":      fat.UpdatedAt,
	}
}
