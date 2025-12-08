package models

import (
	"gorm.io/gorm"
	"time"
)

// LeadTemplate represents a reusable template for lead communication
type LeadTemplate struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"not null;uniqueIndex"`
	DisplayName  string         `json:"display_name" gorm:"not null"`
	TemplateType string         `json:"template_type" gorm:"not null;index"` // email, sms, letter
	Subject      string         `json:"subject" gorm:"type:text"`
	Body         string         `json:"body" gorm:"type:text;not null"`
	Variables    StringArray    `json:"variables" gorm:"type:json"`
	Category     string         `json:"category" gorm:"index"` // follow_up, nurture, reengagement, etc.
	IsActive     bool           `json:"is_active" gorm:"default:true;index"`
	UsageCount   int            `json:"usage_count" gorm:"default:0"`
	CreatedBy    string         `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for LeadTemplate
func (LeadTemplate) TableName() string {
	return "lead_templates"
}
