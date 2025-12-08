package models

import (
	"gorm.io/gorm"
	"time"
)

// PropertyNotification represents a notification about a property
type PropertyNotification struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Property Reference
	PropertyID      *uint  `json:"property_id"` // Optional - may not exist in our system
	PropertyMLSID   string `json:"property_mls_id" gorm:"not null;index"`
	PropertyAddress string `json:"property_address" gorm:"not null"`

	// Notification Details
	NotificationType string `json:"notification_type" gorm:"not null"` // price_change, status_change, new_listing, sold, etc.
	Title            string `json:"title" gorm:"not null"`
	Message          string `json:"message" gorm:"type:text"`

	// Change Details
	PreviousValue    string   `json:"previous_value"`
	NewValue         string   `json:"new_value"`
	ChangeAmount     *float64 `json:"change_amount"`     // For price changes
	ChangePercentage *float64 `json:"change_percentage"` // For price changes

	// Notification Target
	RecipientType   string   `json:"recipient_type" gorm:"not null"` // agent, client, admin, system
	RecipientIDs    []string `json:"recipient_ids" gorm:"type:text[]"`
	RecipientEmails []string `json:"recipient_emails" gorm:"type:text[]"`

	// Delivery Status
	Status           string     `json:"status" gorm:"default:'pending'"` // pending, sent, failed, read
	SentAt           *time.Time `json:"sent_at"`
	ReadAt           *time.Time `json:"read_at"`
	DeliveryAttempts int        `json:"delivery_attempts" gorm:"default:0"`

	// Delivery Channels
	EmailSent       bool `json:"email_sent" gorm:"default:false"`
	SMSSent         bool `json:"sms_sent" gorm:"default:false"`
	PushSent        bool `json:"push_sent" gorm:"default:false"`
	WebNotification bool `json:"web_notification" gorm:"default:false"`

	// Context Data
	ContextData JSONB  `json:"context_data" gorm:"type:jsonb"`
	ActionURL   string `json:"action_url"`
	ActionLabel string `json:"action_label"`

	// Priority and Urgency
	Priority  string     `json:"priority" gorm:"default:'medium'"` // low, medium, high, urgent
	IsUrgent  bool       `json:"is_urgent" gorm:"default:false"`
	ExpiresAt *time.Time `json:"expires_at"`

	// Source and Tracking
	Source   string `json:"source" gorm:"default:'system'"` // system, manual, api, webhook
	SourceID string `json:"source_id"`                      // ID from source system
	BatchID  string `json:"batch_id"`                       // For grouping related notifications

	// Error Handling
	LastError  string     `json:"last_error" gorm:"type:text"`
	RetryAfter *time.Time `json:"retry_after"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Property *Property `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
}

// PropertyNotificationLog represents a log of notification delivery attempts
type PropertyNotificationLog struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Notification Reference
	PropertyNotificationID uint                 `json:"property_notification_id" gorm:"not null;index"`
	PropertyNotification   PropertyNotification `json:"property_notification" gorm:"foreignKey:PropertyNotificationID"`

	// Delivery Details
	Channel        string    `json:"channel" gorm:"not null"` // email, sms, push, web
	Recipient      string    `json:"recipient" gorm:"not null"`
	DeliveryStatus string    `json:"delivery_status" gorm:"not null"` // sent, delivered, failed, bounced, opened, clicked
	DeliveryTime   time.Time `json:"delivery_time" gorm:"not null"`

	// Delivery Metadata
	Provider          string `json:"provider"` // sendgrid, twilio, firebase, etc.
	ProviderMessageID string `json:"provider_message_id"`
	ResponseCode      string `json:"response_code"`
	ResponseMessage   string `json:"response_message" gorm:"type:text"`

	// Engagement Tracking
	OpenedAt       *time.Time `json:"opened_at"`
	ClickedAt      *time.Time `json:"clicked_at"`
	RepliedAt      *time.Time `json:"replied_at"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at"`

	// Context
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	DeviceType string `json:"device_type"`

	CreatedAt time.Time `json:"created_at"`
}

// NotificationTemplate represents reusable notification templates
type NotificationTemplate struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Template Identity
	Name        string `json:"name" gorm:"not null;uniqueIndex"`
	DisplayName string `json:"display_name" gorm:"not null"`
	Description string `json:"description" gorm:"type:text"`

	// Template Configuration
	NotificationType string `json:"notification_type" gorm:"not null"`
	Category         string `json:"category"` // marketing, transactional, alert
	Language         string `json:"language" gorm:"default:'en'"`

	// Template Content
	Subject       string `json:"subject" gorm:"not null"`
	EmailTemplate string `json:"email_template" gorm:"type:text"`
	SMSTemplate   string `json:"sms_template" gorm:"type:text"`
	PushTemplate  string `json:"push_template" gorm:"type:text"`
	WebTemplate   string `json:"web_template" gorm:"type:text"`

	// Template Variables
	Variables     []string `json:"variables" gorm:"type:text[]"`
	DefaultValues JSONB    `json:"default_values" gorm:"type:jsonb"`

	// Delivery Configuration
	DefaultPriority string   `json:"default_priority" gorm:"default:'medium'"`
	DefaultChannels []string `json:"default_channels" gorm:"type:text[]"`
	DeliveryDelay   int      `json:"delivery_delay" gorm:"default:0"` // seconds
	ExpirationHours int      `json:"expiration_hours" gorm:"default:24"`

	// Template Status
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	Version    string     `json:"version" gorm:"default:'1.0'"`
	LastUsed   *time.Time `json:"last_used"`
	UsageCount int        `json:"usage_count" gorm:"default:0"`

	// Approval and Compliance
	IsApproved      bool       `json:"is_approved" gorm:"default:false"`
	ApprovedBy      string     `json:"approved_by"`
	ApprovedAt      *time.Time `json:"approved_at"`
	ComplianceNotes string     `json:"compliance_notes" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// NotificationStats represents aggregated notification statistics
type NotificationStats struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Time Period
	StatsDate  time.Time `json:"stats_date" gorm:"not null;index"`
	PeriodType string    `json:"period_type" gorm:"not null"` // daily, weekly, monthly

	// Template/Type Breakdown
	NotificationType string                `json:"notification_type"`
	TemplateID       *uint                 `json:"template_id"`
	Template         *NotificationTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`

	// Volume Metrics
	TotalSent      int `json:"total_sent" gorm:"default:0"`
	TotalDelivered int `json:"total_delivered" gorm:"default:0"`
	TotalFailed    int `json:"total_failed" gorm:"default:0"`
	TotalBounced   int `json:"total_bounced" gorm:"default:0"`

	// Engagement Metrics
	TotalOpened       int `json:"total_opened" gorm:"default:0"`
	TotalClicked      int `json:"total_clicked" gorm:"default:0"`
	TotalReplied      int `json:"total_replied" gorm:"default:0"`
	TotalUnsubscribed int `json:"total_unsubscribed" gorm:"default:0"`

	// Channel Breakdown
	EmailSent int `json:"email_sent" gorm:"default:0"`
	SMSSent   int `json:"sms_sent" gorm:"default:0"`
	PushSent  int `json:"push_sent" gorm:"default:0"`
	WebSent   int `json:"web_sent" gorm:"default:0"`

	// Calculated Rates
	DeliveryRate    float64 `json:"delivery_rate" gorm:"default:0"`    // delivered/sent
	OpenRate        float64 `json:"open_rate" gorm:"default:0"`        // opened/delivered
	ClickRate       float64 `json:"click_rate" gorm:"default:0"`       // clicked/opened
	BounceRate      float64 `json:"bounce_rate" gorm:"default:0"`      // bounced/sent
	UnsubscribeRate float64 `json:"unsubscribe_rate" gorm:"default:0"` // unsubscribed/delivered

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Helper methods for PropertyNotification

func (pn *PropertyNotification) GetFormattedTitle() string {
	if pn.Title == "" {
		return "Property Update Notification"
	}
	return pn.Title
}

func (pn *PropertyNotification) IsExpired() bool {
	if pn.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*pn.ExpiresAt)
}

func (pn *PropertyNotification) GetDeliveryStatus() string {
	switch pn.Status {
	case "pending":
		return "Waiting to send"
	case "sent":
		return "Successfully sent"
	case "failed":
		return "Delivery failed"
	case "read":
		return "Read by recipient"
	default:
		return pn.Status
	}
}

func (pn *PropertyNotification) HasBeenDelivered() bool {
	return pn.EmailSent || pn.SMSSent || pn.PushSent || pn.WebNotification
}

func (pn *PropertyNotification) GetPriorityColor() string {
	switch pn.Priority {
	case "urgent":
		return "red"
	case "high":
		return "orange"
	case "medium":
		return "yellow"
	case "low":
		return "green"
	default:
		return "gray"
	}
}

// ToDict converts PropertyNotification to a map for JSON responses
func (pn *PropertyNotification) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                 pn.ID,
		"property_id":        pn.PropertyID,
		"property_mls_id":    pn.PropertyMLSID,
		"property_address":   pn.PropertyAddress,
		"notification_type":  pn.NotificationType,
		"title":              pn.GetFormattedTitle(),
		"message":            pn.Message,
		"previous_value":     pn.PreviousValue,
		"new_value":          pn.NewValue,
		"change_amount":      pn.ChangeAmount,
		"change_percentage":  pn.ChangePercentage,
		"recipient_type":     pn.RecipientType,
		"status":             pn.Status,
		"delivery_status":    pn.GetDeliveryStatus(),
		"sent_at":            pn.SentAt,
		"read_at":            pn.ReadAt,
		"priority":           pn.Priority,
		"priority_color":     pn.GetPriorityColor(),
		"is_urgent":          pn.IsUrgent,
		"expires_at":         pn.ExpiresAt,
		"is_expired":         pn.IsExpired(),
		"has_been_delivered": pn.HasBeenDelivered(),
		"email_sent":         pn.EmailSent,
		"sms_sent":           pn.SMSSent,
		"push_sent":          pn.PushSent,
		"web_notification":   pn.WebNotification,
		"created_at":         pn.CreatedAt,
		"updated_at":         pn.UpdatedAt,
	}
}

// Helper methods for NotificationTemplate

func (nt *NotificationTemplate) CanBeUsed() bool {
	return nt.IsActive && nt.IsApproved
}

func (nt *NotificationTemplate) GetVersionInfo() string {
	if nt.Version == "" {
		return "1.0"
	}
	return nt.Version
}

func (nt *NotificationTemplate) IncrementUsage() {
	nt.UsageCount++
	now := time.Now()
	nt.LastUsed = &now
}

// ToDict converts NotificationTemplate to a map for JSON responses
func (nt *NotificationTemplate) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                nt.ID,
		"name":              nt.Name,
		"display_name":      nt.DisplayName,
		"description":       nt.Description,
		"notification_type": nt.NotificationType,
		"category":          nt.Category,
		"language":          nt.Language,
		"subject":           nt.Subject,
		"variables":         nt.Variables,
		"default_priority":  nt.DefaultPriority,
		"default_channels":  nt.DefaultChannels,
		"is_active":         nt.IsActive,
		"version":           nt.GetVersionInfo(),
		"last_used":         nt.LastUsed,
		"usage_count":       nt.UsageCount,
		"is_approved":       nt.IsApproved,
		"can_be_used":       nt.CanBeUsed(),
		"created_at":        nt.CreatedAt,
		"updated_at":        nt.UpdatedAt,
	}
}
