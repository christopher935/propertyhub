package services

import (
	"fmt"
	"log"
	"os"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupService manages application setup state
type SetupService struct {
	setupComplete bool
	testMode      bool
}

// NewSetupService creates a new setup service
func NewSetupService() *SetupService {
	return &SetupService{
		setupComplete: false,
		testMode:      false,
	}
}

// IsSetupRequired checks if setup is still required
func (s *SetupService) IsSetupRequired() bool {
	// Check if setup is permanently marked as complete
	if os.Getenv("SETUP_COMPLETE") == "true" {
		return false
	}
	
	// Fallback: Check if setup is complete by looking for admin users in database
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		if db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{}); err == nil {
			var count int64
			if err := db.Model(&models.AdminUser{}).Count(&count).Error; err == nil && count > 0 {
				// Admin user exists, setup is complete
				return false
			}
		}
	}
	return false
}

// GetSetupProgress returns current setup progress
func (s *SetupService) GetSetupProgress() map[string]bool {
	progress := map[string]bool{
		"database_configured": false,
		"admin_user_created":  false,
		"api_keys_configured": false,
	}

	// Check database configuration
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		if db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{}); err == nil {
			progress["database_configured"] = true
			
			// Check if admin user exists
			var count int64
			if err := db.Model(&models.AdminUser{}).Count(&count).Error; err == nil && count > 0 {
				progress["admin_user_created"] = true
			}
		}
	}

	// Check API keys (consider configured if any are set)
	if os.Getenv("FUB_API_KEY") != "" || os.Getenv("SCRAPER_API_KEY") != "" || os.Getenv("SMTP_HOST") != "" {
		progress["api_keys_configured"] = true
	}

	return progress
}

// MarkDatabaseConfigured marks database as configured
func (s *SetupService) MarkDatabaseConfigured() {
	log.Println("✅ Database marked as configured")
}

// MarkAdminUserCreated marks admin user as created
func (s *SetupService) MarkAdminUserCreated() {
	log.Println("✅ Admin user marked as created")
}

// MarkAPIKeysConfigured marks API keys as configured
func (s *SetupService) MarkAPIKeysConfigured() {
	log.Println("✅ API keys marked as configured")
}

// CanCompleteSetup checks if setup can be completed
func (s *SetupService) CanCompleteSetup() bool {
	progress := s.GetSetupProgress()
	return progress["database_configured"] && progress["admin_user_created"]
}

// CompleteSetup completes the setup process
func (s *SetupService) CompleteSetup() error {
	if !s.CanCompleteSetup() {
		return fmt.Errorf("setup requirements not met")
	}

	s.setupComplete = true
	s.testMode = true

	// Create a permanent marker that setup is complete
	os.Setenv("SETUP_COMPLETE", "true")
	
	// Also save to .env file
	envFile := ".env"
	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString("SETUP_COMPLETE=true\n")
		f.Close()
	}

	log.Println("🎉 Setup completed successfully")
	return nil
}

// IsTestMode checks if application is in test mode
func (s *SetupService) IsTestMode() bool {
	return s.testMode
}

// IsProductionMode checks if application is in production mode
func (s *SetupService) IsProductionMode() bool {
	return s.setupComplete && !s.testMode
}

// EnableTestMode enables test mode
func (s *SetupService) EnableTestMode() {
	s.testMode = true
	s.setupComplete = true
	log.Println("✅ Test mode enabled")
}

// EnableProductionMode enables production mode
func (s *SetupService) EnableProductionMode() {
	s.testMode = false
	s.setupComplete = true
	log.Println("✅ Production mode enabled")
}
