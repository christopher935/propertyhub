package models

import (
	"time"
	"gorm.io/gorm"
)

type AdminNotification struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	Type         string                 `json:"type" gorm:"not null;index"`
	Priority     string                 `json:"priority" gorm:"not null;default:'medium'"`
	Title        string                 `json:"title" gorm:"not null"`
	Message      string                 `json:"message" gorm:"type:text"`
	LeadID       *int                   `json:"lead_id,omitempty" gorm:"index"`
	LeadName     string                 `json:"lead_name,omitempty"`
	LeadEmail    string                 `json:"lead_email,omitempty"`
	LeadScore    int                    `json:"lead_score,omitempty"`
	PropertyID   *int                   `json:"property_id,omitempty" gorm:"index"`
	PropertyAddr string                 `json:"property_address,omitempty"`
	ActionURL    string                 `json:"action_url"`
	ActionLabel  string                 `json:"action_label"`
	Data         JSONB                  `json:"data" gorm:"type:jsonb"`
	CreatedAt    time.Time              `json:"created_at" gorm:"index"`
	ReadAt       *time.Time             `json:"read_at,omitempty"`
	DismissedAt  *time.Time             `json:"dismissed_at,omitempty"`
	DeletedAt    gorm.DeletedAt         `json:"-" gorm:"index"`

	Lead     *Lead     `json:"lead,omitempty" gorm:"foreignKey:LeadID"`
	Property *Property `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
}

func (an *AdminNotification) IsRead() bool {
	return an.ReadAt != nil
}

func (an *AdminNotification) IsDismissed() bool {
	return an.DismissedAt != nil
}

func (an *AdminNotification) MarkAsRead() {
	now := time.Now()
	an.ReadAt = &now
}

func (an *AdminNotification) Dismiss() {
	now := time.Now()
	an.DismissedAt = &now
}

func (an *AdminNotification) GetIcon() string {
	icons := map[string]string{
		"hot_lead":         "🔥",
		"application":      "📋",
		"booking":          "📅",
		"return_visitor":   "🔄",
		"engagement_spike": "📈",
		"property_saved":   "⭐",
		"inquiry":          "💬",
		"multiple_views":   "👀",
	}
	if icon, ok := icons[an.Type]; ok {
		return icon
	}
	return "🔔"
}

func (an *AdminNotification) GetPriorityColor() string {
	colors := map[string]string{
		"high":   "#EF4444",
		"medium": "#F59E0B",
		"low":    "#10B981",
	}
	if color, ok := colors[an.Priority]; ok {
		return color
	}
	return "#6B7280"
}

func (an *AdminNotification) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":               an.ID,
		"type":             an.Type,
		"priority":         an.Priority,
		"title":            an.Title,
		"message":          an.Message,
		"lead_id":          an.LeadID,
		"lead_name":        an.LeadName,
		"lead_email":       an.LeadEmail,
		"lead_score":       an.LeadScore,
		"property_id":      an.PropertyID,
		"property_address": an.PropertyAddr,
		"action_url":       an.ActionURL,
		"action_label":     an.ActionLabel,
		"data":             an.Data,
		"created_at":       an.CreatedAt,
		"read_at":          an.ReadAt,
		"dismissed_at":     an.DismissedAt,
		"is_read":          an.IsRead(),
		"is_dismissed":     an.IsDismissed(),
		"icon":             an.GetIcon(),
		"priority_color":   an.GetPriorityColor(),
	}
}
