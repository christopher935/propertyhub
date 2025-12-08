package models

import (
	"time"
)

// UserProfile represents extended profile information for admin users
type UserProfile struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string `gorm:"uniqueIndex;not null" json:"user_id"`
	FirstName string    `gorm:"size:100" json:"first_name"`
	LastName  string    `gorm:"size:100" json:"last_name"`
	Phone     string    `gorm:"size:50" json:"phone"`
	Company   string    `gorm:"size:200" json:"company"`
	Department string   `gorm:"size:100" json:"department"`
	JobTitle  string    `gorm:"size:100" json:"job_title"`
	AvatarURL string    `gorm:"type:text" json:"avatar_url"`
	Bio       string    `gorm:"type:text" json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for UserProfile
func (UserProfile) TableName() string {
	return "user_profiles"
}

// UserPreferences represents user preferences and notification settings
type UserPreferences struct {
	ID                   int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID               string `gorm:"uniqueIndex;not null" json:"user_id"`
	Timezone             string    `gorm:"size:50;default:'America/Chicago'" json:"timezone"`
	Language             string    `gorm:"size:10;default:'en'" json:"language"`
	DateFormat           string    `gorm:"size:20;default:'MM/DD/YYYY'" json:"date_format"`
	TimeFormat           string    `gorm:"size:10;default:'12h'" json:"time_format"`
	EmailNotifications   bool      `gorm:"default:true" json:"email_notifications"`
	SMSNotifications     bool      `gorm:"default:false" json:"sms_notifications"`
	DesktopNotifications bool      `gorm:"default:true" json:"desktop_notifications"`
	WeeklyReports        bool      `gorm:"default:true" json:"weekly_reports"`
	Theme                string    `gorm:"size:20;default:'light'" json:"theme"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// TableName specifies the table name for UserPreferences
func (UserPreferences) TableName() string {
	return "user_preferences"
}
