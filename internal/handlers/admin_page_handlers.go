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
	data := GetAdminPageData(c, "Overview")
	c.HTML(http.StatusOK, "admin-dashboard.html", data)
}

// AdminLeads renders the leads management page
func (h *AdminHandlers) AdminLeads(c *gin.Context) {
	data := GetAdminPageData(c, "Leads Management")
	c.HTML(http.StatusOK, "admin-leads.html", data)
}

// AdminProperties renders the properties management page
func (h *AdminHandlers) AdminProperties(c *gin.Context) {
	data := GetAdminPageData(c, "Properties & Showings")
	c.HTML(http.StatusOK, "admin-properties.html", data)
}

// AdminBookings renders the bookings management page
func (h *AdminHandlers) AdminBookings(c *gin.Context) {
	data := GetAdminPageData(c, "Booking Management")
	c.HTML(http.StatusOK, "admin-bookings.html", data)
}

// AdminReports renders the reports page
func (h *AdminHandlers) AdminReports(c *gin.Context) {
	data := GetAdminPageData(c, "Reports")
	c.HTML(http.StatusOK, "admin-reports.html", data)
}

// AdminSettings renders the settings page
func (h *AdminHandlers) AdminSettings(c *gin.Context) {
	data := GetAdminPageData(c, "Settings")
	c.HTML(http.StatusOK, "admin-settings.html", data)
}
