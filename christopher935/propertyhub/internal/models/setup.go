package models

import "time"

// SetupStatus represents the current setup state of the application
type SetupStatus struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	IsSetupComplete   bool      `gorm:"default:false" json:"is_setup_complete"`
	IsTestMode        bool      `gorm:"default:true" json:"is_test_mode"`
	IsProductionMode  bool      `gorm:"default:false" json:"is_production_mode"`
	DatabaseConfigured bool     `gorm:"default:false" json:"database_configured"`
	AdminUserCreated  bool      `gorm:"default:false" json:"admin_user_created"`
	APIKeysConfigured bool      `gorm:"default:false" json:"api_keys_configured"`
	SetupCompletedAt  *time.Time `json:"setup_completed_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SetupConfiguration stores configuration during setup
type SetupConfiguration struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	DatabaseURL     string `json:"database_url"`
	RedisURL        string `json:"redis_url"`
	FUBAPIKey       string `json:"fub_api_key"`
	ScraperAPIKey   string `json:"scraper_api_key"`
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPUsername    string `json:"smtp_username"`
	SMTPPassword    string `json:"smtp_password"`
	TwilioAccountSID string `json:"twilio_account_sid"`
	TwilioAuthToken string `json:"twilio_auth_token"`
	TwilioPhoneNumber string `json:"twilio_phone_number"`
	JWTSecret       string `json:"jwt_secret"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AdminSetup stores admin user setup information
type AdminSetup struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"` // Will be hashed
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
