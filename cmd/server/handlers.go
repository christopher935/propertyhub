package main

import (
	"chrisgross-ctrl-project/internal/handlers"
	"chrisgross-ctrl-project/internal/security"
	"gorm.io/gorm"
)

// AllHandlers contains all initialized handlers for route registration
type AllHandlers struct {
	// API Handlers
	AnalyticsAPI          *handlers.AnalyticsAPIHandlers

	// Analytics & Business Intelligence
	BusinessIntelligence  *handlers.BusinessIntelligenceHandlers
	TieredStats           *handlers.TieredStatsHandlers

	// Approvals & Workflow
	Approvals             *handlers.ApprovalsManagementHandlers
	ApplicationWorkflow   *handlers.ApplicationWorkflowHandlers
	ClosingPipeline       *handlers.ClosingPipelineHandlers

	// Behavioral Intelligence & FUB
	Behavioral            *handlers.BehavioralIntelligenceHandlers
	InsightsAPI           *handlers.InsightsAPIHandlers
	ContextFUB            *handlers.ContextFUBIntegrationHandlers

	// Calendar & Scheduling
	Calendar              *handlers.CalendarHandlers

	// Data Migration
	DataMigration         *handlers.DataMigrationHandlers

	// Email
	EmailSender           *handlers.EmailSenderHandlers
	Unsubscribe           *handlers.UnsubscribeHandlers

	// HAR Market
	HARMarket             *handlers.HARMarketHandlers

	// HAR Property Sync

	// Lead Management
	LeadReengagement      *handlers.LeadReengagementHandler
	LeadsList             *handlers.LeadsListHandler

	// Pre-Listing
	PreListing            *handlers.PreListingHandlers

	// Properties
	Properties            *handlers.PropertiesHandler
	SavedProperties       *handlers.SavedPropertiesHandler
	Recommendations       *handlers.RecommendationsHandler
	PropertyAlerts        *handlers.PropertyAlertsHandler
	LiveActivity          *handlers.LiveActivityHandler

	// Security
	SecurityMonitoring    *handlers.SecurityMonitoringHandlers
	AdvancedSecurityAPI   *handlers.AdvancedSecurityAPIHandlers

	// Webhooks
	Webhook               *handlers.WebhookHandlers

	// Database (for inline handlers that need it)
	DB                    *gorm.DB
	
	// Encryption Manager (for decrypting addresses)
	EncryptionManager     *security.EncryptionManager
}
