package main

import (
	"net/http"
	"strconv"

	"chrisgross-ctrl-project/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes registers all admin-facing routes
func RegisterAdminRoutes(r *gin.Engine, h *AllHandlers, propertyHubAI *services.SpiderwebAIOrchestrator) {
	// Admin login
	r.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth/pages/admin-login.html", gin.H{"Title": "Admin Login"})
	})

	// ===== SIMPLIFIED ADMIN ROUTES (8 Main Pages) =====

	// 1. Dashboard - Overview
	r.GET("/admin/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin-dashboard.html", gin.H{"Title": "Dashboard"})
	})

	// 2. Properties & Showings (with tabs: All Properties, Property Performance, Photos, Showing Calendar, Booking Management, HAR Sync, Pre-Listing Manager, Valuations)
	r.GET("/admin/property-list", func(c *gin.Context) {
		c.HTML(200, "property-list.html", gin.H{"Title": "Properties & Showings"})
	})

	// 3. Leads & Conversion (with tabs: Lead Management, Lead Sources, Conversion Funnel, Client Portal, Behavioral Intelligence)
	r.GET("/admin/lead-management", func(c *gin.Context) {
		c.HTML(200, "lead-management.html", gin.H{"Title": "Leads & Conversion"})
	})

	// 4. Communications (with tabs: Communication Center, Email Management)
	r.GET("/admin/communication-center", func(c *gin.Context) {
		c.HTML(200, "communication-center.html", gin.H{"Title": "Communications"})
	})

	// 5. Workflow (with tabs: Applications, Approvals, Closing Pipeline)
	r.GET("/admin/application-workflow", func(c *gin.Context) {
		c.HTML(200, "application-workflow.html", gin.H{"Title": "Workflow"})
	})

	// 6. Analytics & Reports (with tabs: Business Intelligence, Market Analytics, Performance Reports, Custom Reports, Friday Reports)
	r.GET("/admin/business-intelligence", func(c *gin.Context) {
		c.HTML(200, "business-intelligence.html", gin.H{"Title": "Analytics & Reports"})
	})

	// 7. Team
	r.GET("/admin/team-dashboard", func(c *gin.Context) {
		c.HTML(200, "team-dashboard.html", gin.H{"Title": "Team"})
	})

	// 8. System (with tabs: Settings, Security, User Management, Access Control, Audit Logs, Integrations, Data Migration)
	r.GET("/admin/settings", func(c *gin.Context) {
		c.HTML(200, "system-settings.html", gin.H{"Title": "System"})
	})

	// Supporting detail/action routes
	r.GET("/admin/property-detail/:id", func(c *gin.Context) {
		c.HTML(200, "property-detail.html", gin.H{"Title": "Property Details"})
	})
	r.GET("/admin/property-create", func(c *gin.Context) {
		c.HTML(200, "property-create.html", gin.H{"Title": "Add Property"})
	})
	r.GET("/admin/property-edit/:id", func(c *gin.Context) {
		c.HTML(200, "property-edit.html", gin.H{"Title": "Edit Property"})
	})
	r.GET("/admin/booking-detail/:id", func(c *gin.Context) {
		c.HTML(200, "booking-detail.html", gin.H{"Title": "Showing Details"})
	})
	r.GET("/admin/application-detail/:id", func(c *gin.Context) {
		c.HTML(200, "application-detail.html", gin.H{"Title": "Application Review"})
	})
	r.GET("/admin/lead-detail/:id", func(c *gin.Context) {
		c.HTML(200, "lead-detail.html", gin.H{"Title": "Lead Detail"})
	})
	r.GET("/admin/commission-detail/:id", func(c *gin.Context) {
		c.HTML(200, "commission-detail.html", gin.H{"Title": "Commission Detail"})
	})

	// Handler-based routes (keep these as-is)
	r.GET("/admin/behavioral-intelligence", h.Behavioral.BehavioralIntelligencePage)
	r.GET("/admin/email-senders", h.EmailSender.AdminEmailSendersPage)

	// Customer Feedback
	r.GET("/admin/customer-feedback", func(c *gin.Context) {
		c.HTML(200, "customer-feedback.html", gin.H{"Title": "Customer Feedback"})
	})
	r.GET("/admin/feedback-detail/:id", func(c *gin.Context) {
		c.HTML(200, "feedback-detail.html", gin.H{"Title": "Feedback Detail"})
	})

	// Admin Confirmation Pages
	r.GET("/admin/property-added-success", func(c *gin.Context) {
		c.HTML(200, "property-added-success.html", gin.H{"Title": "Property Added"})
	})
	r.GET("/admin/property-updated-success", func(c *gin.Context) {
		c.HTML(200, "property-updated-success.html", gin.H{"Title": "Property Updated"})
	})
	r.GET("/admin/lead-added-success", func(c *gin.Context) {
		c.HTML(200, "lead-added-success.html", gin.H{"Title": "Lead Added"})
	})
	r.GET("/admin/commission-updated-success", func(c *gin.Context) {
		c.HTML(200, "commission-updated-success.html", gin.H{"Title": "Commission Updated"})
	})

	// ============================================================================
	// TIER 2: TEAM COLLABORATION ROUTES
	// ============================================================================

	// Agent Dashboard (for agents to view their own data)
	r.GET("/agent/dashboard", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{"Title": "My Dashboard"})
	})

	// Admin Team Management
	r.GET("/admin/team", func(c *gin.Context) {
		c.HTML(200, "team-dashboard.html", gin.H{"Title": "Team Dashboard"})
	})
	r.GET("/admin/team/add", func(c *gin.Context) {
		c.HTML(200, "team-member-add.html", gin.H{"Title": "Add Team Member"})
	})
	r.GET("/admin/team/edit/:id", func(c *gin.Context) {
		c.HTML(200, "team-member-edit.html", gin.H{"Title": "Edit Team Member"})
	})

	// Lead Assignment
	r.GET("/admin/leads/assign", func(c *gin.Context) {
		c.HTML(200, "lead-assignment.html", gin.H{"Title": "Lead Assignment"})
	})
	r.GET("/admin/leads/add", func(c *gin.Context) {
		c.HTML(200, "lead-add.html", gin.H{"Title": "Add Lead"})
	})

	// Team Confirmation Pages
	r.GET("/admin/team-member-added-success", func(c *gin.Context) {
		c.HTML(200, "team-member-added-success.html", gin.H{"Title": "Team Member Added"})
	})
	r.GET("/admin/team-member-updated-success", func(c *gin.Context) {
		c.HTML(200, "team-member-updated-success.html", gin.H{"Title": "Team Member Updated"})
	})

	// ============================================================================
	// PROPERTYHUB AI INTELLIGENCE ROUTES
	// ============================================================================

	// Intelligence Dashboard - Main AI insights endpoint
	r.GET("/admin/intelligence/dashboard", func(c *gin.Context) {
		data, err := propertyHubAI.GetDashboardIntelligence()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	// Top Opportunities - High-priority leads and actions
	r.GET("/admin/intelligence/opportunities", func(c *gin.Context) {
		opportunities, err := propertyHubAI.GetOpportunityInsights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"opportunities": opportunities})
	})

	// Funnel Analytics - Conversion funnel insights
	r.GET("/admin/intelligence/funnel", func(c *gin.Context) {
		funnel, err := propertyHubAI.GetFunnelInsights(30) // Last 30 days
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, funnel)
	})

	// Property Matches - AI-recommended properties for a lead
	r.GET("/admin/intelligence/matches/:lead_id", func(c *gin.Context) {
		leadIDStr := c.Param("lead_id")
		leadID, err := strconv.ParseInt(leadIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lead ID"})
			return
		}
		matches, err := propertyHubAI.GetPropertyMatches(leadID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"matches": matches})
	})

	// Lead Analysis - Deep dive into a specific lead
	r.GET("/admin/intelligence/analyze/:lead_id", func(c *gin.Context) {
		leadIDStr := c.Param("lead_id")
		leadID, err := strconv.ParseInt(leadIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lead ID"})
			return
		}
		analysis, err := propertyHubAI.AnalyzeLeadOpportunity(leadID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, analysis)
	})

	// Trigger Intelligence Cycle - Manual trigger for AI processing
	r.POST("/admin/intelligence/cycle/trigger", func(c *gin.Context) {
		go propertyHubAI.RunIntelligenceCycle()
		c.JSON(http.StatusOK, gin.H{"message": "Intelligence cycle triggered"})
	})
}
