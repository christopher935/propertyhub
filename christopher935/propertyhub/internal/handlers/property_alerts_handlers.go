package handlers

import (
	"net/http"
	
	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PropertyAlertsHandler struct {
	db            *gorm.DB
	alertsService *services.PropertyAlertsService
}

func NewPropertyAlertsHandler(db *gorm.DB, emailService *services.EmailService) *PropertyAlertsHandler {
	return &PropertyAlertsHandler{
		db:            db,
		alertsService: services.NewPropertyAlertsService(db, emailService),
	}
}

func (h *PropertyAlertsHandler) SubscribeToAlerts(c *gin.Context) {
	var pref services.AlertPreferences
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}
	
	if pref.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	
	if err := h.alertsService.SaveAlertPreferences(&pref); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Alert preferences saved successfully",
		"preferences": pref,
	})
}

func (h *PropertyAlertsHandler) GetAlertPreferences(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter required"})
		return
	}
	
	pref, err := h.alertsService.GetAlertPreferences(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No alert preferences found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"preferences": pref})
}

func (h *PropertyAlertsHandler) UpdateAlertPreferences(c *gin.Context) {
	var pref services.AlertPreferences
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	if err := h.alertsService.SaveAlertPreferences(&pref); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Preferences updated successfully",
		"preferences": pref,
	})
}

func (h *PropertyAlertsHandler) UnsubscribeFromAlerts(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter required"})
		return
	}
	
	if err := h.alertsService.UnsubscribeFromAlerts(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Successfully unsubscribed from property alerts"})
}
