package main

import (
	"time"

	"chrisgross-ctrl-project/internal/handlers"
	"chrisgross-ctrl-project/internal/middleware"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers all API routes
func RegisterAPIRoutes(api *gin.RouterGroup, h *AllHandlers, propertyValuationHandler *handlers.PropertyValuationHandlers, emailAutomationHandler *handlers.EmailAutomationHandlers) {
	// ============================================================================
	// WEBSOCKET - Real-time Updates
	// ============================================================================
	api.GET("/ws", h.WebSocket.HandleWebSocket)
	api.GET("/ws/admin/activity", h.WebSocket.HandleAdminActivityFeed)

	// ============================================================================
	// ADMIN NOTIFICATIONS - Real-time Alerts
	// ============================================================================
	api.GET("/notifications/ws", h.AdminNotification.HandleWebSocket)
	api.GET("/notifications", h.AdminNotification.GetNotifications)
	api.GET("/notifications/unread-count", h.AdminNotification.GetUnreadCount)
	api.PUT("/notifications/:id/read", h.AdminNotification.MarkAsRead)
	api.PUT("/notifications/read-all", h.AdminNotification.MarkAllAsRead)

	// ============================================================================
	// TIERED STATS API - Dashboard Intelligence with Redis Caching
	// ============================================================================
	api.GET("/stats/live", h.TieredStats.GetLiveStats)
	api.GET("/stats/hot", h.TieredStats.GetHotStats)
	api.GET("/stats/warm", h.TieredStats.GetWarmStats)
	api.GET("/stats/daily", h.TieredStats.GetDailyStats)
	
	// Additional stats endpoints for admin dashboard
	api.GET("/stats/critical", func(c *gin.Context) {
		var pendingBookings int64
		var pendingApplications int64
		h.DB.Model(&models.BookingRequest{}).Where("status = ?", "pending").Count(&pendingBookings)
		h.DB.Model(&models.ApplicationNumber{}).Where("status = ?", "pending").Count(&pendingApplications)
		
		c.JSON(200, gin.H{
			"pending_bookings":      pendingBookings,
			"pending_applications":  pendingApplications,
			"system_alerts":         0,
		})
	})
	
	api.GET("/stats/key-metrics", func(c *gin.Context) {
		var totalLeads int64
		var hotLeads int64
		var warmLeads int64
		var confirmedBookings int64
		h.DB.Table("leads").Where("status = ?", "active").Count(&totalLeads)
		h.DB.Table("leads").Where("temperature = ?", "hot").Count(&hotLeads)
		h.DB.Table("leads").Where("temperature = ?", "warm").Count(&warmLeads)
		h.DB.Model(&models.BookingRequest{}).Where("status = ?", "confirmed").Count(&confirmedBookings)
		
		c.JSON(200, gin.H{
			"active_leads":        totalLeads,
			"active_leads_trend":  0,
			"conversion_rate":     0,
			"revenue_mtd":         0,
			"revenue_trend":       0,
			"bookings_week":       confirmedBookings,
			"confirmed_bookings":  confirmedBookings,
			"hot_leads":           hotLeads,
			"warm_leads":          warmLeads,
		})
	})
	
	api.GET("/stats/opportunities", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"opportunities": []gin.H{},
		})
	})
	
	api.GET("/stats/activity", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"activities": []gin.H{},
		})
	})

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

	// HAR Market API removed - HAR blocked access

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
	
	// Bulk Lead Operations
	api.POST("/leads/bulk/email", h.BulkOperations.BulkEmailLeads)
	api.POST("/leads/bulk/assign", h.BulkOperations.BulkAssignLeads)
	api.POST("/leads/bulk/status", h.BulkOperations.BulkUpdateLeadStatus)
	api.POST("/leads/bulk/tag", h.BulkOperations.BulkTagLeads)
	api.POST("/leads/bulk/archive", h.BulkOperations.BulkArchiveLeads)

	// Bulk Property Operations
	api.POST("/properties/bulk/status", h.BulkOperations.BulkUpdatePropertyStatus)
	api.POST("/properties/bulk/assign", h.BulkOperations.BulkAssignProperties)
	api.POST("/properties/bulk/featured", h.BulkOperations.BulkUpdateFeatured)
	api.POST("/properties/bulk/export", h.BulkOperations.BulkExportProperties)
	
	// Properties API
	api.GET("/properties", h.Properties.GetPropertiesGin)
	api.GET("/properties/:id", h.Properties.GetPropertyByIDGin)
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
	
	// API v1 Aliases - for backward compatibility with frontend JavaScript
	v1 := api.Group("/v1")
	{
		v1.GET("/properties", h.Properties.GetPropertiesGin)
		v1.GET("/properties/:id", h.Properties.GetPropertyByIDGin)
		v1.POST("/properties/search", h.Properties.SearchPropertiesPost)
		v1.POST("/properties/save", h.SavedProperties.SaveProperty)
		v1.DELETE("/properties/save/:id", h.SavedProperties.UnsaveProperty)
		v1.GET("/properties/saved", h.SavedProperties.GetSavedProperties)
		v1.GET("/properties/:id/is-saved", h.SavedProperties.CheckIfSaved)
		v1.GET("/properties/:id/similar", h.Recommendations.GetSimilarProperties)
		v1.POST("/bookings", middleware.BookingRateLimiter.RateLimit(), h.Booking.CreateBooking)
		v1.GET("/bookings/:id", h.Booking.GetBooking)
		v1.POST("/bookings/:id/cancel", h.Booking.CancelBooking)
		v1.GET("/bookings", h.Booking.ListBookings)
		v1.POST("/bookings/:id/complete", h.Booking.MarkCompleted)
		v1.POST("/bookings/:id/no-show", h.Booking.MarkNoShow)
		v1.PUT("/bookings/:id/reschedule", h.Booking.RescheduleBooking)
	}
	
	// Live Activity API (Admin Real-Time)
	api.GET("/admin/live-activity", h.LiveActivity.GetLiveActivity)
	api.GET("/admin/active-sessions", h.BehavioralSessions.GetActiveSessions)
	api.GET("/admin/session/:id", h.LiveActivity.GetSessionDetails)

	// Behavioral Sessions API (Admin Real-Time - Who's Browsing Now)

	// Behavioral Event Tracking API (Consumer-facing with real-time broadcasting)
	api.POST("/behavioral/track/property-view", h.BehavioralEvent.TrackPropertyView)
	api.POST("/behavioral/track/property-save", h.BehavioralEvent.TrackPropertySave)
	api.POST("/behavioral/track/inquiry", h.BehavioralEvent.TrackInquiry)
	api.POST("/behavioral/track/search", h.BehavioralEvent.TrackSearch)
	api.GET("/behavioral/active-count", h.BehavioralEvent.GetActiveSessionsCount)
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
	api.GET("/agent/dashboard", h.Team.GetAgentDashboard)

	// Team Management API
	api.GET("/admin/team/dashboard", h.Team.GetTeamDashboard)
	api.GET("/admin/team", h.Team.GetTeamMembers)
	api.GET("/admin/team/:id", h.Team.GetTeamMember)
	api.POST("/admin/team/add", h.Team.AddTeamMember)
	api.PUT("/admin/team/:id", h.Team.UpdateTeamMember)
	api.DELETE("/admin/team/:id", h.Team.DeleteTeamMember)

	// Lead Assignment API
	api.POST("/admin/leads/assign", h.Team.AssignLead)

	api.POST("/admin/leads/add", func(c *gin.Context) {
		// TODO: Add new lead manually
		c.JSON(200, gin.H{
			"success": true,
			"message": "Lead added successfully",
		})
	})

	// ============================================================================
	// CONTACT FORM API
	// ============================================================================
	api.POST("/contact", func(c *gin.Context) {
		var req struct {
			Name           string `json:"name" binding:"required"`
			Email          string `json:"email" binding:"required,email"`
			Phone          string `json:"phone"`
			Subject        string `json:"subject" binding:"required"`
			Message        string `json:"message" binding:"required"`
			RecaptchaToken string `json:"recaptcha_token"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid request data",
			})
			return
		}

		var encryptedName, encryptedEmail, encryptedPhone security.EncryptedString
		if h.EncryptionManager != nil {
			var err error
			encryptedName, err = h.EncryptionManager.Encrypt(req.Name)
			if err != nil {
				c.JSON(500, gin.H{"success": false, "message": "Encryption error"})
				return
			}
			encryptedEmail, err = h.EncryptionManager.Encrypt(req.Email)
			if err != nil {
				c.JSON(500, gin.H{"success": false, "message": "Encryption error"})
				return
			}
			encryptedPhone, err = h.EncryptionManager.Encrypt(req.Phone)
			if err != nil {
				c.JSON(500, gin.H{"success": false, "message": "Encryption error"})
				return
			}
		} else {
			encryptedName = security.EncryptedString(req.Name)
			encryptedEmail = security.EncryptedString(req.Email)
			encryptedPhone = security.EncryptedString(req.Phone)
		}

		contact := models.Contact{
			Name:      encryptedName,
			Email:     encryptedEmail,
			Phone:     encryptedPhone,
			Message:   req.Subject + "\n\n" + req.Message,
			Status:    "new",
			Source:    "contact_form",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := h.DB.Create(&contact).Error; err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to save contact request",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Thank you! We'll get back to you soon.",
		})
	})

	// ============================================================================
	// COOKIE CONSENT API - GDPR/CCPA Compliance
	// ============================================================================
	api.POST("/cookie-consent", middleware.CookieConsentHandler)

	// ============================================================================
	// AVAILABILITY API - Consolidated from ServeMux
	// ============================================================================
	v1.GET("/availability/check", h.Availability.CheckAvailabilityGin)
	v1.GET("/availability/blackouts", h.Availability.GetBlackoutDatesGin)
	v1.POST("/availability/blackouts", h.Availability.CreateBlackoutDateGin)
	v1.DELETE("/availability/blackouts/:id", h.Availability.RemoveBlackoutDateGin)
	v1.GET("/availability/blackouts/upcoming", h.Availability.GetUpcomingBlackoutsGin)
	v1.GET("/availability/stats", h.Availability.GetAvailabilityStatsGin)
	v1.POST("/availability/validate", h.Availability.ValidateBookingGin)
	v1.POST("/availability/cleanup", h.Availability.CleanupExpiredBlackoutsGin)

	// Also add without v1 prefix for JS compatibility
	api.GET("/availability/check", h.Availability.CheckAvailabilityGin)

	// ============================================================================
	// APPFOLIO INTEGRATION API - Tenant & Property Sync
	// ============================================================================
	handlers.RegisterAppFolioRoutes(v1, h.DB)
}
