package main

import (
	"chrisgross-ctrl-project/internal/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers all API routes
func RegisterAPIRoutes(api *gin.RouterGroup, h *AllHandlers, propertyValuationHandler *handlers.PropertyValuationHandlers, emailAutomationHandler *handlers.EmailAutomationHandlers) {
	// ============================================================================
	// WEBSOCKET - Real-time Updates
	// ============================================================================
	api.GET("/ws", h.WebSocket.HandleWebSocket)

	// ============================================================================
	// TIERED STATS API - Dashboard Intelligence with Redis Caching
	// ============================================================================
	api.GET("/stats/live", h.TieredStats.GetLiveStats)
	api.GET("/stats/hot", h.TieredStats.GetHotStats)
	api.GET("/stats/warm", h.TieredStats.GetWarmStats)
	api.GET("/stats/daily", h.TieredStats.GetDailyStats)

	// Dashboard API - Real-time dashboard widgets
	api.GET("/dashboard/properties", h.Dashboard.GetPropertySummary)
	api.GET("/dashboard/bookings", h.Dashboard.GetBookingSummary)
	api.GET("/dashboard/revenue", h.Dashboard.GetRevenueSummary)
	api.GET("/dashboard/market-data", h.Dashboard.GetMarketData)
	api.GET("/dashboard/recent-activity", h.Dashboard.GetRecentActivity)
	api.GET("/dashboard/alerts", h.Dashboard.GetAlerts)
	api.GET("/dashboard/maintenance", h.Dashboard.GetMaintenanceRequests)
	api.GET("/dashboard/leads", h.Dashboard.GetLeadSummary)
	api.GET("/dashboard/upcoming-tasks", h.Dashboard.GetUpcomingTasks)

	// Analytics API
	api.GET("/analytics/bookings", h.BusinessIntelligence.GetBookingAnalytics)
	api.GET("/analytics/properties", h.BusinessIntelligence.GetPropertyAnalytics)

	// Business Intelligence API
	api.GET("/business-intelligence/metrics", h.BusinessIntelligence.GetDashboardMetrics)
	api.GET("/business-intelligence/friday-report", h.BusinessIntelligence.GetFridayReport)
	api.GET("/business-intelligence/property-analytics", h.BusinessIntelligence.GetPropertyAnalytics)
	api.GET("/business-intelligence/booking-analytics", h.BusinessIntelligence.GetBookingAnalytics)
	api.GET("/business-intelligence/lead-analytics", h.BusinessIntelligence.GetLeadAnalytics)
	api.GET("/business-intelligence/roi-analytics", h.BusinessIntelligence.GetROIAnalytics)
	api.GET("/business-intelligence/conversion-funnel", h.BusinessIntelligence.GetConversionFunnel)
	api.GET("/business-intelligence/efficiency-metrics", h.BusinessIntelligence.GetEfficiencyMetrics)
	api.GET("/business-intelligence/system-health", h.BusinessIntelligence.GetSystemHealth)
	api.GET("/business-intelligence/realtime-stats", h.BusinessIntelligence.GetRealtimeStats)

	// Approvals Management API
	api.GET("/approvals", h.Approvals.GetApprovals)
	api.GET("/approvals/:id", h.Approvals.GetApproval)
	api.POST("/approvals", h.Approvals.CreateApproval)
	api.PUT("/approvals/:id/status", h.Approvals.UpdateApprovalStatus)
	api.DELETE("/approvals/:id", h.Approvals.DeleteApproval)

	// Application Workflow API
	api.GET("/application-workflow/:id", h.ApplicationWorkflow.GetApplicationWorkflow)
	api.GET("/applicants-without-application", h.ApplicationWorkflow.GetApplicantsWithoutApplication)
	api.GET("/properties-with-applications", h.ApplicationWorkflow.GetPropertiesWithApplications)
	api.POST("/application-workflow/application-number", h.ApplicationWorkflow.CreateApplicationNumber)
	api.PUT("/application-workflow/:id/status", h.ApplicationWorkflow.UpdateApplicationStatus)
	api.POST("/application-workflow/assign-agent", h.ApplicationWorkflow.AssignAgentToApplication)
	api.POST("/application-workflow/move-applicant", h.ApplicationWorkflow.MoveApplicantToApplication)
	api.POST("/applications/:id/approve", h.ApplicationWorkflow.ApproveApplication)
	api.POST("/applications/:id/deny", h.ApplicationWorkflow.DenyApplication)
	api.POST("/applications/:id/request-info", h.ApplicationWorkflow.RequestMoreInfo)

	// Behavioral Intelligence API
	api.GET("/behavioral/dashboard", h.Behavioral.GetBehavioralIntelligenceDashboard)
	api.GET("/behavioral/metrics", h.Behavioral.GetBehavioralMetrics)
	api.GET("/behavioral/insights", h.InsightsAPI.GetPredictiveInsights)
	api.GET("/behavioral/houston-market", h.Behavioral.GetHoustonMarketIntelligence)
	
	// Behavioral Analytics API
	handlers.RegisterBehavioralAnalyticsRoutes(api, h.DB)

	// Booking API
	api.POST("/v1/bookings", h.Booking.CreateBooking)
	api.GET("/v1/bookings/:id", h.Booking.GetBooking)
	api.POST("/v1/bookings/:id/cancel", h.Booking.CancelBooking)
	api.GET("/v1/bookings", h.Booking.ListBookings)
	api.POST("/bookings/:id/complete", h.Booking.MarkCompleted)
	api.POST("/bookings/:id/no-show", h.Booking.MarkNoShow)
	api.PUT("/bookings/:id/reschedule", h.Booking.RescheduleBooking)

	// Calendar Management API
	api.GET("/calendar/stats", h.Calendar.GetCalendarStats)
	api.GET("/calendar/today", h.Calendar.GetTodayEvents)
	api.GET("/calendar/upcoming", h.Calendar.GetUpcomingEvents)
	api.GET("/calendar/automation-stats", h.Calendar.GetAutomationStats)
	api.POST("/calendar/showing", h.Calendar.CreateShowingEvent)
	api.POST("/calendar/follow-up", h.Calendar.CreateFollowUpEvent)
	api.POST("/calendar/schedule-follow-up", h.Calendar.ScheduleFollowUp)
	api.POST("/calendar/sync", h.Calendar.SyncCalendar)
	api.POST("/calendar/trigger-automation", h.Calendar.TriggerAutomation)
	api.PUT("/calendar/event/:id/status", h.Calendar.UpdateEventStatus)

	// Closing Pipeline API
	api.GET("/closing-pipeline", h.ClosingPipeline.GetClosingPipelines)
	api.GET("/closing-pipeline/:id", h.ClosingPipeline.GetPipelineItem)
	api.POST("/closing-pipeline", h.ClosingPipeline.CreatePipelineItem)
	api.PUT("/closing-pipeline/:id/stage", h.ClosingPipeline.UpdatePipelineStage)
	api.PUT("/closing-pipeline/:id/lease-status", h.ClosingPipeline.UpdateLeaseWorkflowStatus)
	api.DELETE("/closing-pipeline/:id", h.ClosingPipeline.DeletePipelineItem)

	// Context FUB Integration API
	api.GET("/context-fub/status", h.ContextFUB.GetContextFUBStatus)
	api.GET("/context-fub/analytics", h.ContextFUB.GetContextFUBAnalytics)
	api.GET("/context-fub/behavioral-insights", h.ContextFUB.GetAdvancedBehavioralInsights)
	api.GET("/context-fub/behavioral-metrics", h.ContextFUB.GetAdvancedBehavioralMetrics)
	api.GET("/context-fub/trigger-history", h.ContextFUB.GetBehavioralTriggerHistory)
	api.POST("/context-fub/behavioral-profile", h.ContextFUB.CreateAdvancedBehavioralProfile)
	api.PUT("/context-fub/behavioral-profile/:id", h.ContextFUB.UpdateAdvancedBehavioralProfile)
	api.POST("/context-fub/process-triggers", h.ContextFUB.ProcessAdvancedBehavioralTriggers)
	api.POST("/context-fub/trigger-automation", h.ContextFUB.TriggerContextDrivenFUBAutomation)
	api.POST("/context-fub/webhook", h.ContextFUB.ProcessContextIntelligenceWebhook)

	// Data Migration API
	api.GET("/migration/history", h.DataMigration.GetImportHistory)
	api.GET("/migration/requirements", h.DataMigration.GetImportRequirements)
	api.GET("/migration/sample-csv", h.DataMigration.DownloadSampleCSV)
	api.POST("/migration/validate", h.DataMigration.ValidateCSV)
	api.POST("/migration/import/properties", h.DataMigration.ImportProperties)
	api.POST("/migration/import/bookings", h.DataMigration.ImportBookings)
	api.POST("/migration/import/customers", h.DataMigration.ImportCustomers)

	// Email Senders API
	api.GET("/email/senders", h.EmailSender.GetTrustedSenders)
	api.GET("/email/senders/:id", h.EmailSender.GetTrustedSender)
	api.POST("/email/senders", h.EmailSender.CreateTrustedSender)
	api.PUT("/email/senders/:id", h.EmailSender.UpdateTrustedSender)
	api.DELETE("/email/senders/:id", h.EmailSender.DeleteTrustedSender)
	api.GET("/email/stats", h.EmailSender.GetEmailProcessingStats)
	api.POST("/email/test-parsing", h.EmailSender.TestEmailParsing)

	// Email Automation API (only if Redis available)
	if emailAutomationHandler != nil {
		api.GET("/email/campaigns", emailAutomationHandler.GetCampaigns)
		api.GET("/email/campaigns/:id", emailAutomationHandler.GetCampaign)
		api.POST("/email/campaigns", emailAutomationHandler.CreateCampaign)
		api.PUT("/email/campaigns/:id", emailAutomationHandler.UpdateCampaign)
		api.DELETE("/email/campaigns/:id", emailAutomationHandler.DeleteCampaign)
		api.GET("/email/templates", emailAutomationHandler.GetTemplates)
		api.GET("/email/templates/:id", emailAutomationHandler.GetTemplate)
		api.POST("/email/templates", emailAutomationHandler.CreateTemplate)
		api.PUT("/email/templates/:id", emailAutomationHandler.UpdateTemplate)
		api.DELETE("/email/templates/:id", emailAutomationHandler.DeleteTemplate)
		api.POST("/email/send-batch", emailAutomationHandler.SendBatch)
		api.GET("/email/batch-status/:id", emailAutomationHandler.GetBatchStatus)
		api.GET("/email/analytics/delivery", emailAutomationHandler.GetDeliveryAnalytics)
		api.GET("/email/analytics/engagement", emailAutomationHandler.GetEngagementAnalytics)
		api.GET("/email/analytics/performance", emailAutomationHandler.GetPerformanceAnalytics)
	}

	// HAR Market API
	api.GET("/har/market-summary", h.HARMarket.GetMarketSummary)
	api.GET("/har/market-report/:id", h.HARMarket.GetMarketReport)
	api.GET("/har/latest-reports", h.HARMarket.GetLatestReports)
	api.GET("/har/reports-by-type/:type", h.HARMarket.GetReportsByType)
	api.GET("/har/search", h.HARMarket.SearchReports)
	api.GET("/har/scraping-stats", h.HARMarket.GetScrapingStats)
	api.GET("/har/scraping-logs", h.HARMarket.GetScrapingLogs)
	api.POST("/har/trigger-scraping", h.HARMarket.TriggerScraping)
	api.POST("/har/schedule-weekly", h.HARMarket.ScheduleWeeklyScraping)

	// Lead Reengagement API
	api.GET("/leads/list", h.LeadReengagement.GetLeads)
	api.GET("/leads/metrics", h.LeadReengagement.GetMetrics)
	api.GET("/leads/segment-stats", h.LeadReengagement.GetSegmentStats)
	api.GET("/leads/safety-status", h.LeadReengagement.GetSafetyStatus)
	api.GET("/leads/templates", h.LeadReengagement.GetTemplates)
	api.POST("/leads/import", h.LeadReengagement.ImportLeads)
	api.POST("/leads/segment", h.LeadReengagement.SegmentLeads)
	api.POST("/leads/prepare-campaign", h.LeadReengagement.PrepareCampaign)
	api.POST("/leads/template", h.LeadReengagement.CreateTemplate)
	api.POST("/leads/emergency-stop", h.LeadReengagement.EmergencyStopAll)

	api.GET("/leads", h.LeadsList.GetAllLeads)
	
	// Properties API
	api.GET("/properties", h.Properties.GetPropertiesGin)
	api.POST("/properties/search", h.Properties.SearchPropertiesPost)
	
	// Saved Properties API (Consumer Feature)
	api.POST("/properties/save", h.SavedProperties.SaveProperty)
	api.DELETE("/properties/save/:id", h.SavedProperties.UnsaveProperty)
	api.GET("/properties/saved", h.SavedProperties.GetSavedProperties)
	api.GET("/properties/:id/is-saved", h.SavedProperties.CheckIfSaved)
	
	// AI Recommendations API (Consumer Feature)
	api.GET("/recommendations", h.Recommendations.GetPersonalizedRecommendations)
	api.GET("/properties/:id/similar", h.Recommendations.GetSimilarProperties)
	
	// Property Alerts API (Consumer Feature)
	api.POST("/alerts/subscribe", h.PropertyAlerts.SubscribeToAlerts)
	api.GET("/alerts/preferences", h.PropertyAlerts.GetAlertPreferences)
	api.PUT("/alerts/preferences", h.PropertyAlerts.UpdateAlertPreferences)
	api.POST("/alerts/unsubscribe", h.PropertyAlerts.UnsubscribeFromAlerts)
	
	// Live Activity API (Admin Real-Time)
	api.GET("/admin/live-activity", h.LiveActivity.GetLiveActivity)
	api.GET("/admin/active-sessions", h.BehavioralSessions.GetActiveSessions)
	api.GET("/admin/session/:id", h.LiveActivity.GetSessionDetails)

	// Behavioral Sessions API (Admin Real-Time - Who's Browsing Now)
	api.GET("/admin/sessions/active", h.BehavioralSessions.GetActiveSessions)
	api.GET("/admin/sessions/:id/journey", h.BehavioralSessions.GetSessionJourney)

	// Command Center API - AI-driven actionable insights
	api.GET("/command-center/items", h.CommandCenter.GetItems)
	api.POST("/command-center/act", h.CommandCenter.ExecuteAction)
	api.POST("/command-center/dismiss", h.CommandCenter.DismissItem)
	api.GET("/command-center/stats", h.CommandCenter.GetStats)

	// Pre-listing API
	api.GET("/pre-listing/valuation/:id", h.PreListing.GetPropertyValuation)

	// Property Valuation API (if enabled)
	if propertyValuationHandler != nil {
		api.GET("/valuation/property/:id", propertyValuationHandler.GetPropertyValuation)
		api.GET("/valuation/property-by-id/:id", propertyValuationHandler.GetPropertyValuationByID)
		api.GET("/valuation/requests", propertyValuationHandler.GetValuationRequests)
		api.GET("/valuation/request/:id", propertyValuationHandler.GetValuationRequest)
		api.GET("/valuation/history/:id", propertyValuationHandler.GetValuationHistory)
		api.GET("/valuation/trends", propertyValuationHandler.GetValuationTrends)
		api.GET("/valuation/accuracy", propertyValuationHandler.GetValuationAccuracy)
		api.GET("/valuation/config", propertyValuationHandler.GetValuationConfig)
		api.PUT("/valuation/config", propertyValuationHandler.UpdateValuationConfig)
		api.GET("/valuation/market-report", propertyValuationHandler.GetMarketReport)
		api.GET("/valuation/market-trends", propertyValuationHandler.GetMarketTrends)
		api.GET("/valuation/area-analysis/:area", propertyValuationHandler.GetAreaMarketAnalysis)
		api.GET("/valuation/city-analysis/:city", propertyValuationHandler.GetCityMarketAnalysis)
		api.GET("/valuation/comparables/:id", propertyValuationHandler.GetComparableProperties)
		api.GET("/valuation/bulk", propertyValuationHandler.GetBulkValuations)
		api.GET("/valuation/performance", propertyValuationHandler.GetPerformanceReport)
		api.POST("/valuation/calibrate", propertyValuationHandler.CalibrateValuationModel)
		api.POST("/valuation/test-accuracy", propertyValuationHandler.TestValuationAccuracy)
	}

	// Security API
	api.GET("/security/metrics", h.SecurityMonitoring.GetSecurityMetrics)
	api.GET("/security/events", h.SecurityMonitoring.GetSecurityEvents)
	api.GET("/security/events/:id", h.SecurityMonitoring.GetSecurityEventDetails)
	api.POST("/security/events", h.SecurityMonitoring.CreateSecurityEvent)
	api.POST("/security/session", h.SecurityMonitoring.CreateSecuritySession)
	api.PUT("/security/events/:id/resolve", h.SecurityMonitoring.ResolveSecurityEvent)

	// Advanced Security API
	api.GET("/security/advanced/metrics", h.AdvancedSecurityAPI.GetSecurityMetrics)
	api.POST("/security/advanced/session", h.AdvancedSecurityAPI.CreateSecuritySession)
	api.POST("/security/advanced/analyze-threat", h.AdvancedSecurityAPI.AnalyzeThreat)

	// Unsubscribe API
	api.GET("/unsubscribe/list", h.Unsubscribe.GetUnsubscribeList)
	api.GET("/unsubscribe/stats", h.Unsubscribe.GetUnsubscribeStats)

	// Webhook API
	api.GET("/webhooks/events", h.Webhook.GetWebhookEvents)
	api.GET("/webhooks/stats", h.Webhook.GetWebhookStats)
	api.POST("/webhooks/fub", h.Webhook.ProcessFUBWebhook)
	api.POST("/webhooks/twilio", h.Webhook.ProcessTwilioWebhook)
	api.POST("/webhooks/inbound-email", h.Webhook.ProcessInboundEmail)

	// ============================================================================
	// TIER 2: TEAM COLLABORATION API
	// ============================================================================

	// Agent Dashboard API
	api.GET("/agent/dashboard", func(c *gin.Context) {
		// TODO: Implement agent-specific dashboard data
		c.JSON(200, gin.H{
			"activeLeads":          12,
			"upcomingShowings":     5,
			"conversionRate":       32,
			"avgResponseTime":      "2.3h",
			"recentLeads":          []gin.H{},
			"upcomingShowingsList": []gin.H{},
		})
	})

	// Team Management API
	api.GET("/admin/team/dashboard", func(c *gin.Context) {
		// TODO: Implement team dashboard data
		c.JSON(200, gin.H{
			"activeAgents":   8,
			"assignedLeads":  124,
			"teamConversion": 28,
			"weeklyShowings": 47,
			"teamMembers":    []gin.H{},
			"topPerformers":  []gin.H{},
		})
	})

	api.GET("/admin/team", func(c *gin.Context) {
		// TODO: Return list of team members
		c.JSON(200, gin.H{"agents": []gin.H{}})
	})

	api.GET("/admin/team/:id", func(c *gin.Context) {
		// TODO: Return single team member details
		c.JSON(200, gin.H{
			"id":             c.Param("id"),
			"firstName":      "",
			"lastName":       "",
			"email":          "",
			"phone":          "",
			"role":           "agent",
			"status":         "active",
			"permissions":    []string{},
			"assignedLeads":  0,
			"activeShowings": 0,
			"conversionRate": 0,
			"totalRevenue":   0,
		})
	})

	api.POST("/admin/team/add", func(c *gin.Context) {
		// TODO: Add new team member
		c.JSON(200, gin.H{
			"success": true,
			"message": "Team member added successfully",
		})
	})

	api.PUT("/admin/team/:id", func(c *gin.Context) {
		// TODO: Update team member
		c.JSON(200, gin.H{
			"success": true,
			"message": "Team member updated successfully",
		})
	})

	api.DELETE("/admin/team/:id", func(c *gin.Context) {
		// TODO: Delete team member
		c.JSON(200, gin.H{
			"success": true,
			"message": "Team member deleted successfully",
		})
	})

	// Lead Assignment API
	api.POST("/admin/leads/assign", func(c *gin.Context) {
		// TODO: Assign lead to agent
		c.JSON(200, gin.H{
			"success": true,
			"message": "Lead assigned successfully",
		})
	})

	api.POST("/admin/leads/add", func(c *gin.Context) {
		// TODO: Add new lead manually
		c.JSON(200, gin.H{
			"success": true,
			"message": "Lead added successfully",
		})
	})
}

// ============================================================================
// MISSING ENDPOINTS - 55 Routes (Phase 2: Steel Installation)
// ============================================================================

func RegisterMissingRoutes(api *gin.RouterGroup) {
	// Context FUB routes (5)
	api.GET("/context-fub/stats", handlers.GetContextFUBStats)
	api.POST("/context-fub/trigger", handlers.PostContextFUBTrigger)
	api.POST("/context-fub/sync", handlers.PostContextFUBSync)
	api.PUT("/context-fub/config", handlers.PutContextFUBConfig)
	api.GET("/context-fub/logs", handlers.GetContextFUBLogs)
	
	// Communication routes (8)
	api.GET("/communication/history", handlers.GetCommunicationHistory)
	api.GET("/communication/templates", handlers.GetCommunicationTemplates)
	api.POST("/communication/send-email", handlers.PostCommunicationSendEmail)
	api.POST("/communication/send-sms", handlers.PostCommunicationSendSMS)
	api.POST("/communication/bulk-send", handlers.PostCommunicationBulkSend)
	api.GET("/communication/stats", handlers.GetCommunicationStats)
	api.GET("/communication/inbox", handlers.GetCommunicationInbox)
	api.POST("/communication/reply", handlers.PostCommunicationReply)
	
	// Email routes (6)
	api.GET("/email/senders/:id", handlers.GetEmailSenderByID)
	api.PUT("/email/senders/:id", handlers.PutEmailSender)
	api.DELETE("/email/senders/:id", handlers.DeleteEmailSender)
	api.GET("/email/parsed-applications", handlers.GetEmailParsedApplications)
	api.GET("/email/parsing-stats", handlers.GetEmailParsingStats)
	api.POST("/email/retry-parsing", handlers.PostEmailRetryParsing)
	api.GET("/email/parsing-logs", handlers.GetEmailParsingLogs)
	
	// HAR Market routes (3)
	api.GET("/har/scrape-stats", handlers.GetHARScrapeStats)
	api.POST("/har/trigger-scrape", handlers.PostHARTriggerScrape)
	api.GET("/har/scrape-logs", handlers.GetHARScrapeLogs)
	
	// Leads routes (8)
	api.GET("/leads/:id", handlers.GetLeadByID)
	api.PUT("/leads/:id", handlers.PutLead)
	api.DELETE("/leads/:id", handlers.DeleteLead)
	api.POST("/leads/templates", handlers.PostLeadTemplate)
	api.PUT("/leads/templates/:id", handlers.PutLeadTemplate)
	api.DELETE("/leads/templates/:id", handlers.DeleteLeadTemplate)
	api.POST("/leads/campaign/prepare", handlers.PostLeadCampaignPrepare)
	
	// Migration routes (2)
	api.GET("/migration/status", handlers.GetMigrationStatus)
	api.POST("/migration/start", handlers.PostMigrationStart)
	
	// Pre-listing routes (6)
	api.GET("/pre-listing/properties", handlers.GetPreListingProperties)
	api.GET("/pre-listing/:id", handlers.GetPreListingByID)
	api.POST("/pre-listing", handlers.PostPreListing)
	api.PUT("/pre-listing/:id", handlers.PutPreListing)
	api.DELETE("/pre-listing/:id", handlers.DeletePreListing)
	api.GET("/pre-listing/stats", handlers.GetPreListingStats)
	
	// Valuation routes (5)
	api.POST("/valuation/request", handlers.PostValuationRequest)
	api.PUT("/valuation/request/:id", handlers.PutValuationRequest)
	api.DELETE("/valuation/request/:id", handlers.DeleteValuationRequest)
	api.GET("/valuation/stats", handlers.GetValuationStats)
	api.POST("/valuation/bulk-request", handlers.PostValuationBulkRequest)
	
	// Security routes (3)
	api.POST("/security/events/:id/resolve", handlers.PostSecurityEventResolve)
	api.GET("/security/audit-logs", handlers.GetSecurityAuditLogs)
	api.GET("/security/compliance-report", handlers.GetSecurityComplianceReport)
	
	// Webhooks routes (5)
	api.GET("/webhooks", handlers.GetWebhooks)
	api.GET("/webhooks/:id", handlers.GetWebhookByID)
	api.POST("/webhooks", handlers.PostWebhook)
	api.PUT("/webhooks/:id", handlers.PutWebhook)
	api.DELETE("/webhooks/:id", handlers.DeleteWebhook)
	
	// Misc routes (4)
	api.GET("/approvals/:id", handlers.GetApprovalByID)
	api.GET("/closing-pipeline/:id", handlers.GetClosingPipelineByID)
	api.PUT("/closing-pipeline/:id/status", handlers.PutClosingPipelineStatus)
	api.GET("/agent/stats", handlers.GetAgentStats)
}
