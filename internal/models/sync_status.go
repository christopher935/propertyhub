package models

import (
	"time"
)

type SyncStatus struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	SyncType          string     `json:"sync_type" gorm:"not null;index"`
	Status            string     `json:"status" gorm:"default:'pending'"`
	StartedAt         time.Time  `json:"started_at" gorm:"not null"`
	CompletedAt       *time.Time `json:"completed_at"`
	PropertiesSynced  int        `json:"properties_synced" gorm:"default:0"`
	PropertiesCreated int        `json:"properties_created" gorm:"default:0"`
	PropertiesUpdated int        `json:"properties_updated" gorm:"default:0"`
	PropertiesDeleted int        `json:"properties_deleted" gorm:"default:0"`
	ErrorCount        int        `json:"error_count" gorm:"default:0"`
	Errors            string     `json:"errors" gorm:"type:text"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (SyncStatus) TableName() string {
	return "sync_statuses"
}
