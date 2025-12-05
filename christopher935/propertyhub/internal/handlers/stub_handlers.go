package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// Stub handlers for missing endpoints
// These return empty data with 200 OK to prevent frontend errors

type StubHandlers struct{}

func NewStubHandlers() *StubHandlers {
	return &StubHandlers{}
}

func (s *StubHandlers) GetMigrationSampleCSV(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Sample CSV generation coming soon",
		"data": "",
	})
}

func (s *StubHandlers) TestEmailParsing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Email parsing test coming soon",
		"result": "pending",
	})
}

func (s *StubHandlers) GetCalendarAutomationStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_automations": 0,
		"active_automations": 0,
		"triggered_today": 0,
	})
}

func (s *StubHandlers) GetContextFUBTriggerHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"triggers": []interface{}{},
		"total": 0,
	})
}

func (s *StubHandlers) GetWebhooksStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_webhooks": 0,
		"successful": 0,
		"failed": 0,
		"pending": 0,
	})
}

func (s *StubHandlers) GetAdminTeamDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"team_members": []interface{}{},
		"total_members": 0,
		"active_members": 0,
	})
}

func (s *StubHandlers) GetAgentDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"appointments_today": 0,
		"upcoming_appointments": []interface{}{},
		"total_leads": 0,
	})
}

func (s *StubHandlers) GetBehavioralHoustonMarket(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"market_data": []interface{}{},
		"message": "Houston market analysis coming soon",
	})
}
