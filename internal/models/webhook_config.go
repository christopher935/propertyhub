package models

import (
	"gorm.io/gorm"
	"time"
)

// WebhookConfig represents a webhook configuration for external integrations
type WebhookConfig struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	URL          string         `json:"url" gorm:"not null"`
	EventTypes   StringArray    `json:"event_types" gorm:"type:json"`
	Secret       string         `json:"secret" gorm:"type:text"`
	Active       bool           `json:"active" gorm:"default:true;index"`
	Description  string         `json:"description" gorm:"type:text"`
	LastTested   *time.Time     `json:"last_tested"`
	LastSuccess  *time.Time     `json:"last_success"`
	FailureCount int            `json:"failure_count" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for WebhookConfig
func (WebhookConfig) TableName() string {
	return "webhook_configs"
}
