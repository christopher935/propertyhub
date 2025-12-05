package handlers

import (
	"chrisgross-ctrl-project/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminHandlers holds references to services for admin pages
type AdminHandlers struct {
	BIService *services.BusinessIntelligenceService
}

// NewAdminHandlers creates a new admin handlers instance
func NewAdminHandlers(biService *services.BusinessIntelligenceService) *AdminHandlers {
	return &AdminHandlers{
		BIService: biService,
	}
}

// AdminDashboard renders the admin dashboard page
func (h *AdminHandlers) AdminDashboard(c *gin.Context) {
	// Render the dashboard template
	// The template will use Alpine.js to fetch data from /api/metrics/dashboard
	c.HTML(http.StatusOK, "admin-dashboard.html", gin.H{
		"Title": "Dashboard - PropertyHub",
	})
}

// AdminLeads renders the leads management page
func (h *AdminHandlers) AdminLeads(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-leads.html", gin.H{
		"Title": "Leads - PropertyHub",
	})
}

// AdminProperties renders the properties management page
func (h *AdminHandlers) AdminProperties(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-properties.html", gin.H{
		"Title": "Properties - PropertyHub",
	})
}

// AdminBookings renders the bookings management page
func (h *AdminHandlers) AdminBookings(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-bookings.html", gin.H{
		"Title": "Bookings - PropertyHub",
	})
}

// AdminReports renders the reports page
func (h *AdminHandlers) AdminReports(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-reports.html", gin.H{
		"Title": "Reports - PropertyHub",
	})
}

// AdminSettings renders the settings page
func (h *AdminHandlers) AdminSettings(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-settings.html", gin.H{
		"Title": "Settings - PropertyHub",
	})
}
