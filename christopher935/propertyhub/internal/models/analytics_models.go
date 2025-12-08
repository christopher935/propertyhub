package models

import (
	"time"
)

// AnalyticsEvent represents an analytics event tracked in the system
type AnalyticsEvent struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	EventType  string    `json:"event_type" gorm:"not null"`
	UserID     string    `json:"user_id,omitempty"`
	SessionID  string    `json:"session_id,omitempty"`
	PropertyID string    `json:"property_id,omitempty"`
	Properties JSONB     `json:"properties" gorm:"type:json"`
	Data       JSONB     `json:"data" gorm:"type:json"`
	Referrer   string    `json:"referrer,omitempty"`
	Path       string    `json:"path,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Timestamp  time.Time `json:"timestamp" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
}
