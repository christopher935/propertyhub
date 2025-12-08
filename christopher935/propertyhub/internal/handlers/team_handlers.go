package handlers

import (
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
)

type TeamHandlers struct {
	db *gorm.DB
}

func NewTeamHandlers(db *gorm.DB) *TeamHandlers {
	return &TeamHandlers{db: db}
}

func (h *TeamHandlers) GetAgentDashboard(c *gin.Context) {
	agentID := c.GetString("user_id")
	if agentID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var activeLeads int64
	var upcomingShowings int64
	
	h.db.Model(&models.Lead{}).Where("assigned_agent_id = ? AND status != ?", agentID, "closed").Count(&activeLeads)
	h.db.Model(&models.Booking{}).Where("status = ? AND showing_date >= ?", "scheduled", time.Now()).Count(&upcomingShowings)
	
	var recentLeads []models.Lead
	h.db.Where("assigned_agent_id = ?", agentID).Order("created_at DESC").Limit(5).Find(&recentLeads)
	
	var upcomingShowingsList []models.Booking
	h.db.Where("status = ? AND showing_date >= ?", "scheduled", time.Now()).
		Order("showing_date ASC").Limit(5).Find(&upcomingShowingsList)
	
	var totalLeads int64
	var closedLeads int64
	h.db.Model(&models.Lead{}).Where("assigned_agent_id = ?", agentID).Count(&totalLeads)
	h.db.Model(&models.Lead{}).Where("assigned_agent_id = ? AND status = ?", agentID, "closed").Count(&closedLeads)
	
	conversionRate := 0
	if totalLeads > 0 {
		conversionRate = int((float64(closedLeads) / float64(totalLeads)) * 100)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"activeLeads":          activeLeads,
		"upcomingShowings":     upcomingShowings,
		"conversionRate":       conversionRate,
		"avgResponseTime":      "2.3h",
		"recentLeads":          recentLeads,
		"upcomingShowingsList": upcomingShowingsList,
	})
}

func (h *TeamHandlers) GetTeamDashboard(c *gin.Context) {
	var activeAgents int64
	var assignedLeads int64
	var weeklyShowings int64
	
	h.db.Model(&models.AdminUser{}).Where("active = ?", true).Count(&activeAgents)
	h.db.Model(&models.Lead{}).Where("assigned_agent_id IS NOT NULL AND assigned_agent_id != ''").Count(&assignedLeads)
	h.db.Model(&models.Booking{}).Where("created_at >= ?", time.Now().AddDate(0, 0, -7)).Count(&weeklyShowings)
	
	var teamMembers []gin.H
	var agents []models.AdminUser
	h.db.Where("active = ?", true).Find(&agents)
	
	for _, agent := range agents {
		var leadCount int64
		var showingCount int64
		h.db.Model(&models.Lead{}).Where("assigned_agent_id = ?", agent.ID).Count(&leadCount)
		h.db.Model(&models.Booking{}).Where("fub_lead_id IN (?)", 
			h.db.Table("leads").Select("fub_lead_id").Where("assigned_agent_id = ?", agent.ID)).Count(&showingCount)
		
		teamMembers = append(teamMembers, gin.H{
			"id":        agent.ID,
			"name":      agent.Username,
			"email":     agent.Email,
			"leads":     leadCount,
			"showings":  showingCount,
			"role":      agent.Role,
			"active":    agent.Active,
		})
	}
	
	var topPerformers []gin.H
	type AgentPerformance struct {
		AgentID string
		LeadCount int64
	}
	var performances []AgentPerformance
	h.db.Table("leads").
		Select("assigned_agent_id as agent_id, COUNT(*) as lead_count").
		Where("assigned_agent_id IS NOT NULL AND assigned_agent_id != ''").
		Group("assigned_agent_id").
		Order("lead_count DESC").
		Limit(5).
		Scan(&performances)
	
	for _, perf := range performances {
		var agent models.AdminUser
		if err := h.db.First(&agent, "id = ?", perf.AgentID).Error; err == nil {
			topPerformers = append(topPerformers, gin.H{
				"id":        agent.ID,
				"name":      agent.Username,
				"leads":     perf.LeadCount,
			})
		}
	}
	
	var totalLeads int64
	var closedLeads int64
	h.db.Model(&models.Lead{}).Count(&totalLeads)
	h.db.Model(&models.Lead{}).Where("status = ?", "closed").Count(&closedLeads)
	
	teamConversion := 0
	if totalLeads > 0 {
		teamConversion = int((float64(closedLeads) / float64(totalLeads)) * 100)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"activeAgents":   activeAgents,
		"assignedLeads":  assignedLeads,
		"teamConversion": teamConversion,
		"weeklyShowings": weeklyShowings,
		"teamMembers":    teamMembers,
		"topPerformers":  topPerformers,
	})
}

func (h *TeamHandlers) GetTeamMembers(c *gin.Context) {
	var agents []models.AdminUser
	h.db.Where("active = ?", true).Find(&agents)
	
	var result []gin.H
	for _, agent := range agents {
		var leadCount int64
		h.db.Model(&models.Lead{}).Where("assigned_agent_id = ?", agent.ID).Count(&leadCount)
		
		result = append(result, gin.H{
			"id":          agent.ID,
			"username":    agent.Username,
			"email":       agent.Email,
			"role":        agent.Role,
			"active":      agent.Active,
			"leadCount":   leadCount,
			"lastLogin":   agent.LastLogin,
			"loginCount":  agent.LoginCount,
			"createdAt":   agent.CreatedAt,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{"agents": result})
}

func (h *TeamHandlers) GetTeamMember(c *gin.Context) {
	id := c.Param("id")
	
	var agent models.AdminUser
	if err := h.db.First(&agent, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}
	
	var leadCount, showingCount, closedDeals int64
	h.db.Model(&models.Lead{}).Where("assigned_agent_id = ?", id).Count(&leadCount)
	h.db.Model(&models.Booking{}).Where("fub_lead_id IN (?)", 
		h.db.Table("leads").Select("fub_lead_id").Where("assigned_agent_id = ?", id)).Count(&showingCount)
	h.db.Model(&models.ClosingPipeline{}).
		Where("(listing_agent_id = ? OR tenant_agent_id = ?) AND status = ?", id, id, "completed").
		Count(&closedDeals)
	
	var recentLeads []models.Lead
	h.db.Where("assigned_agent_id = ?", id).Order("created_at DESC").Limit(10).Find(&recentLeads)
	
	c.JSON(http.StatusOK, gin.H{
		"id":             agent.ID,
		"username":       agent.Username,
		"email":          agent.Email,
		"role":           agent.Role,
		"active":         agent.Active,
		"lastLogin":      agent.LastLogin,
		"loginCount":     agent.LoginCount,
		"assignedLeads":  leadCount,
		"activeShowings": showingCount,
		"closedDeals":    closedDeals,
		"recentLeads":    recentLeads,
		"createdAt":      agent.CreatedAt,
		"updatedAt":      agent.UpdatedAt,
	})
}

func (h *TeamHandlers) AddTeamMember(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Role     string `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var existingUser models.AdminUser
	if err := h.db.Where("username = ? OR email = ?", input.Username, input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}
	
	agent := models.AdminUser{
		Username: input.Username,
		Email:    security.EncryptedString(input.Email),
		Role:     input.Role,
		Active:   true,
	}
	
	if err := h.db.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team member"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Team member added successfully",
		"id":      agent.ID,
	})
}

func (h *TeamHandlers) UpdateTeamMember(c *gin.Context) {
	id := c.Param("id")
	
	var agent models.AdminUser
	if err := h.db.First(&agent, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}
	
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Active   *bool  `json:"active"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	updates := make(map[string]interface{})
	if input.Username != "" {
		updates["username"] = input.Username
	}
	if input.Email != "" {
		updates["email"] = security.EncryptedString(input.Email)
	}
	if input.Role != "" {
		updates["role"] = input.Role
	}
	if input.Active != nil {
		updates["active"] = *input.Active
	}
	
	if err := h.db.Model(&agent).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team member"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Team member updated"})
}

func (h *TeamHandlers) DeleteTeamMember(c *gin.Context) {
	id := c.Param("id")
	
	var agent models.AdminUser
	if err := h.db.First(&agent, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}
	
	if err := h.db.Model(&agent).Update("active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team member"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Team member deleted"})
}

func (h *TeamHandlers) AssignLead(c *gin.Context) {
	var input struct {
		LeadID  uint   `json:"leadId" binding:"required"`
		AgentID string `json:"agentId" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var agent models.AdminUser
	if err := h.db.First(&agent, "id = ?", input.AgentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	
	var lead models.Lead
	if err := h.db.First(&lead, input.LeadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead not found"})
		return
	}
	
	if err := h.db.Model(&lead).Update("assigned_agent_id", input.AgentID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign lead"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Lead assigned successfully"})
}
