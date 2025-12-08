package handlers

import (
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// Stub handlers for missing endpoints
// These return empty data with 200 OK to prevent frontend errors

type StubHandlers struct{}

func NewStubHandlers() *StubHandlers {
	return &StubHandlers{}
}

func (s *StubHandlers) GetMigrationSampleCSV(c *gin.Context) {
	dataType := c.DefaultQuery("type", "customers")

	db := c.MustGet("db").(*gorm.DB)
	migrationService := services.NewDataMigrationService(db)

	// Generate sample CSV
	csvData, err := migrationService.GenerateSampleCSV(dataType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data type", err)
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=sample_"+dataType+".csv")
	c.String(http.StatusOK, csvData)
}

func (s *StubHandlers) TestEmailParsing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Email parsing test coming soon",
		"result":  "pending",
	})
}

func (s *StubHandlers) GetCalendarAutomationStats(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Query actual automation data
	var totalAutomations int64
	var activeAutomations int64
	var triggeredToday int64

	// Count calendar events
	db.Model(&models.CalendarEvent{}).Count(&totalAutomations)
	db.Model(&models.CalendarEvent{}).Where("status = ?", "active").Count(&activeAutomations)
	db.Model(&models.CalendarEvent{}).Where("DATE(created_at) = CURRENT_DATE").Count(&triggeredToday)

	c.JSON(http.StatusOK, gin.H{
		"total_automations":  totalAutomations,
		"active_automations": activeAutomations,
		"triggered_today":    triggeredToday,
	})
}

func (s *StubHandlers) GetContextFUBTriggerHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"triggers": []interface{}{},
		"total":    0,
	})
}

func (s *StubHandlers) GetWebhooksStats(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var totalWebhooks, successfulWebhooks, failedWebhooks, pendingWebhooks int64

	db.Model(&models.WebhookEvent{}).Count(&totalWebhooks)
	db.Model(&models.WebhookEvent{}).Where("processed = ?", true).Count(&successfulWebhooks)
	// Failed webhooks would be tracked in a separate failure log table if implemented
	failedWebhooks = 0
	db.Model(&models.WebhookEvent{}).Where("processed = ?", false).Count(&pendingWebhooks)

	c.JSON(http.StatusOK, gin.H{
		"total_webhooks": totalWebhooks,
		"successful":     successfulWebhooks,
		"failed":         failedWebhooks,
		"pending":        pendingWebhooks,
	})
}

func (s *StubHandlers) GetAdminTeamDashboard(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Query admin users table for team data
	var teamMembers []map[string]interface{}
	var totalMembers int64
	var activeMembers int64

	db.Model(&models.AdminUser{}).Count(&totalMembers)
	db.Model(&models.AdminUser{}).Where("is_active = ?", true).Count(&activeMembers)

	// Get team member details
	var users []models.AdminUser
	db.Where("is_active = ?", true).Limit(50).Find(&users)

	for _, user := range users {
		teamMembers = append(teamMembers, map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"team_members":   teamMembers,
		"total_members":  totalMembers,
		"active_members": activeMembers,
	})
}

func (s *StubHandlers) GetAgentDashboard(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Query agent-specific metrics
	var appointmentsToday int64
	var totalLeads int64

	// Count today's bookings
	db.Model(&models.Booking{}).Where("DATE(start_time) = CURRENT_DATE").Count(&appointmentsToday)

	// Count total leads
	db.Model(&models.Lead{}).Count(&totalLeads)

	// Get upcoming appointments
	var upcomingAppointments []map[string]interface{}
	var bookings []models.Booking
	db.Where("start_time > NOW()").Order("start_time ASC").Limit(10).Find(&bookings)

	for _, booking := range bookings {
		upcomingAppointments = append(upcomingAppointments, map[string]interface{}{
			"id":         booking.ID,
			"start_time": booking.StartTime,
			"end_time":   booking.EndTime,
			"status":     booking.Status,
			"customer":   booking.CustomerName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments_today":    appointmentsToday,
		"upcoming_appointments": upcomingAppointments,
		"total_leads":           totalLeads,
	})
}

func (s *StubHandlers) GetBehavioralHoustonMarket(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"market_data": []interface{}{},
		"message":     "Houston market analysis coming soon",
	})
}
