package main

import (
	"chrisgross-ctrl-project/internal/handlers"
	"chrisgross-ctrl-project/internal/security"
	"gorm.io/gorm"
)

// AllHandlers contains all initialized handlers for route registration
type AllHandlers struct {
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

	// Command Center
	CommandCenter         *handlers.CommandCenterHandlers

	// Booking
	Booking               *handlers.BookingHandler

	// Calendar & Scheduling
	Calendar              *handlers.CalendarHandlers

	// Dashboard
	Dashboard             *handlers.DashboardHandlers

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
	BehavioralSessions    *handlers.BehavioralSessionsHandler

	// Security
	SecurityMonitoring    *handlers.SecurityMonitoringHandlers
	AdvancedSecurityAPI   *handlers.AdvancedSecurityAPIHandlers

	// Webhooks
	Webhook               *handlers.WebhookHandlers

	// WebSocket
	WebSocket             *handlers.WebSocketHandler

	// Database (for inline handlers that need it)
	DB                    *gorm.DB
	
	// Encryption Manager (for decrypting addresses)
	EncryptionManager     *security.EncryptionManager
}
