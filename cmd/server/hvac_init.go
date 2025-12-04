package main

import (
	"chrisgross-ctrl-project/internal/jobs"
	"chrisgross-ctrl-project/internal/services"
	"log"
	"gorm.io/gorm"
)

// InitializeHVACSystem sets up the HVAC (scheduled tasks) system
// This is installed but not activated - will be turned on when building is complete
func InitializeHVACSystem(db *gorm.DB, scraperAPIKey, fubAPIKey string) *jobs.JobManager {
	log.Println("🌡️ Installing HVAC system (not activating)...")
	
	// Initialize services
	harService := services.NewHARMarketScraper(db, scraperAPIKey)
	fubService := services.NewEnhancedFUBIntegrationService(db, fubAPIKey)
	biService := services.NewBusinessIntelligenceService(db)
	
	// Create JobManager (installed but not started)
	jobManager := jobs.NewJobManager(db, harService, fubService, biService, nil)
	
	log.Println("✅ HVAC system installed (dormant)")
	return jobManager
}

// ActivateHVAC starts the HVAC system
// Call this when ready to turn on scheduled tasks
func ActivateHVAC(jobManager *jobs.JobManager) {
	log.Println("🌡️ Activating HVAC system...")
	jobManager.StartScheduledJobs()
	log.Println("✅ HVAC operational")
}
