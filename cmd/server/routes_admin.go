package main

import (
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/auth"
	"chrisgross-ctrl-project/internal/middleware"
	"chrisgross-ctrl-project/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes registers all admin-facing routes
func RegisterAdminRoutes(r *gin.Engine, h *AllHandlers, propertyHubAI *services.SpiderwebAIOrchestrator, authManager *auth.SimpleAuthManager) {
	// Public admin login page (no auth required)
	r.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin-login.html", gin.H{"Title": "Admin Login"})
	})

	// Admin login POST endpoint
	r.POST("/admin/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")
		rememberMe := c.PostForm("remember") == "1"

		if email == "" || password == "" {
			c.HTML(http.StatusBadRequest, "admin-login.html", gin.H{
				"Title": "Admin Login",
				"Error": "Email and password are required",
			})
			return
		}

		// Authenticate user
		loginResp, err := authManager.AuthenticateUser(email, password)
		if err != nil || !loginResp.Success {
			c.HTML(http.StatusUnauthorized, "admin-login.html", gin.H{
				"Title": "Admin Login",
				"Error": "Invalid email or password",
			})
			return
		}

		// Set session cookie
		maxAge := 86400 // 24 hours
		if rememberMe {
			maxAge = 2592000 // 30 days
		}

		c.SetCookie(
			"admin_session_token",
			loginResp.Token,
			maxAge,
			"/",
			"",
			false, // secure (set to true in production with HTTPS)
			true,  // httpOnly
		)

		// Redirect to dashboard
		c.Redirect(http.StatusFound, "/admin/dashboard")
	})

	// Admin logout POST endpoint
	r.POST("/admin/logout", func(c *gin.Context) {
		sessionToken, err := c.Cookie("admin_session_token")
		if err == nil && sessionToken != "" {
			// Invalidate the session in the database
			authManager.InvalidateSession(sessionToken)
		}

		// Clear the cookie
		c.SetCookie(
			"admin_session_token",
			"",
			-1,
			"/",
			"",
			false,
			true,
		)

		// Redirect to login page
		c.Redirect(http.StatusFound, "/admin")
	})

	// ===== PROTECTED ADMIN ROUTES (Require Authentication) =====
	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired(authManager))
	{
		// 1. Dashboard - Overview
		admin.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin-dashboard.html", gin.H{"Title": "Dashboard"})
	})

		// 2. Calendar - Showing Management
		admin.GET("/calendar", h.Calendar.CalendarManagementDashboard)

		// 2. Properties & Showings (with tabs: All Properties, Property Performance, Photos, Showing Calendar, Booking Management, HAR Sync, Pre-Listing Manager, Valuations)
		admin.GET("/property-list", func(c *gin.Context) {
		c.HTML(200, "property-list.html", gin.H{"Title": "Properties & Showings"})
	})

		// 3. Leads & Conversion (with tabs: Lead Management, Lead Sources, Conversion Funnel, Client Portal, Behavioral Intelligence)
		admin.GET("/lead-management", func(c *gin.Context) {
		c.HTML(200, "lead-management.html", gin.H{"Title": "Leads & Conversion"})
	})

		// 4. Communications (with tabs: Communication Center, Email Management)
		admin.GET("/communication-center", func(c *gin.Context) {
		c.HTML(200, "communication-center.html", gin.H{"Title": "Communications"})
	})

		// 5. Workflow (with tabs: Applications, Approvals, Closing Pipeline)
		admin.GET("/application-workflow", func(c *gin.Context) {
		c.HTML(200, "application-workflow.html", gin.H{"Title": "Workflow"})
	})

		// 6. Analytics & Reports (with tabs: Business Intelligence, Market Analytics, Performance Reports, Custom Reports, Friday Reports)
		admin.GET("/business-intelligence", func(c *gin.Context) {
		c.HTML(200, "business-intelligence.html", gin.H{"Title": "Analytics & Reports"})
	})

		// 7. Team
		admin.GET("/team-dashboard", func(c *gin.Context) {
		c.HTML(200, "team-dashboard.html", gin.H{"Title": "Team"})
	})

		// 8. System (with tabs: Settings, Security, User Management, Access Control, Audit Logs, Integrations, Data Migration)
		admin.GET("/settings", func(c *gin.Context) {
		c.HTML(200, "system-settings.html", gin.H{"Title": "System"})
	})

		// Supporting detail/action routes
		admin.GET("/property-detail/:id", func(c *gin.Context) {
		c.HTML(200, "property-detail.html", gin.H{"Title": "Property Details"})
	})
		admin.GET("/property-create", func(c *gin.Context) {
		c.HTML(200, "property-create.html", gin.H{"Title": "Add Property"})
	})
		admin.GET("/property-edit/:id", func(c *gin.Context) {
		c.HTML(200, "property-edit.html", gin.H{"Title": "Edit Property"})
	})
		admin.GET("/booking-detail/:id", func(c *gin.Context) {
		c.HTML(200, "booking-detail.html", gin.H{"Title": "Showing Details"})
	})
		admin.GET("/application-detail/:id", func(c *gin.Context) {
		c.HTML(200, "application-detail.html", gin.H{"Title": "Application Review"})
	})
		admin.GET("/lead-detail/:id", func(c *gin.Context) {
		c.HTML(200, "lead-detail.html", gin.H{"Title": "Lead Detail"})
	})
		admin.GET("/commission-detail/:id", func(c *gin.Context) {
		c.HTML(200, "commission-detail.html", gin.H{"Title": "Commission Detail"})
	})

		// Handler-based routes (keep these as-is)
		admin.GET("/behavioral-intelligence", h.Behavioral.BehavioralIntelligencePage)
		admin.GET("/email-senders", h.EmailSender.AdminEmailSendersPage)
		admin.GET("/command-center", h.CommandCenter.RenderPage)

		// Customer Feedback
		admin.GET("/customer-feedback", func(c *gin.Context) {
		c.HTML(200, "customer-feedback.html", gin.H{"Title": "Customer Feedback"})
	})
		admin.GET("/feedback-detail/:id", func(c *gin.Context) {
		c.HTML(200, "feedback-detail.html", gin.H{"Title": "Feedback Detail"})
	})

		// Admin Confirmation Pages
		admin.GET("/property-added-success", func(c *gin.Context) {
		c.HTML(200, "property-added-success.html", gin.H{"Title": "Property Added"})
	})
		admin.GET("/property-updated-success", func(c *gin.Context) {
		c.HTML(200, "property-updated-success.html", gin.H{"Title": "Property Updated"})
	})
		admin.GET("/lead-added-success", func(c *gin.Context) {
		c.HTML(200, "lead-added-success.html", gin.H{"Title": "Lead Added"})
	})
		admin.GET("/commission-updated-success", func(c *gin.Context) {
		c.HTML(200, "commission-updated-success.html", gin.H{"Title": "Commission Updated"})
	})

		// ============================================================================
		// TIER 2: TEAM COLLABORATION ROUTES
		// ============================================================================

		// Agent Dashboard (for agents to view their own data)
		admin.GET("/agent-dashboard", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{"Title": "My Dashboard"})
	})

		// Admin Team Management
		admin.GET("/team", func(c *gin.Context) {
		c.HTML(200, "team-dashboard.html", gin.H{"Title": "Team Dashboard"})
	})
		admin.GET("/team/add", func(c *gin.Context) {
		c.HTML(200, "team-member-add.html", gin.H{"Title": "Add Team Member"})
	})
		admin.GET("/team/edit/:id", func(c *gin.Context) {
		c.HTML(200, "team-member-edit.html", gin.H{"Title": "Edit Team Member"})
	})

		// Lead Assignment
		admin.GET("/leads/assign", func(c *gin.Context) {
		c.HTML(200, "lead-assignment.html", gin.H{"Title": "Lead Assignment"})
	})
		admin.GET("/leads/add", func(c *gin.Context) {
		c.HTML(200, "lead-add.html", gin.H{"Title": "Add Lead"})
	})

		// Team Confirmation Pages
		admin.GET("/team-member-added-success", func(c *gin.Context) {
		c.HTML(200, "team-member-added-success.html", gin.H{"Title": "Team Member Added"})
	})
		admin.GET("/team-member-updated-success", func(c *gin.Context) {
		c.HTML(200, "team-member-updated-success.html", gin.H{"Title": "Team Member Updated"})
	})

		// ============================================================================
		// PROPERTYHUB AI INTELLIGENCE ROUTES
		// ============================================================================

		// Intelligence Dashboard - Main AI insights endpoint
		admin.GET("/intelligence/dashboard", func(c *gin.Context) {
		data, err := propertyHubAI.GetDashboardIntelligence()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

		// Top Opportunities - High-priority leads and actions
		admin.GET("/intelligence/opportunities", func(c *gin.Context) {
		opportunities, err := propertyHubAI.GetOpportunityInsights()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"opportunities": opportunities})
	})

		// Funnel Analytics - Conversion funnel insights
		admin.GET("/intelligence/funnel", func(c *gin.Context) {
		funnel, err := propertyHubAI.GetFunnelInsights(30) // Last 30 days
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, funnel)
	})

		// Property Matches - AI-recommended properties for a lead
		admin.GET("/intelligence/matches/:lead_id", func(c *gin.Context) {
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
		admin.GET("/intelligence/analyze/:lead_id", func(c *gin.Context) {
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
		admin.POST("/intelligence/cycle/trigger", func(c *gin.Context) {
			go propertyHubAI.RunIntelligenceCycle()
			c.JSON(http.StatusOK, gin.H{"message": "Intelligence cycle triggered"})
		})
	}
}
