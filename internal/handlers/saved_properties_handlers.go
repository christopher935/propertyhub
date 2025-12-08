package handlers

import (
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SavedPropertiesHandler struct {
	db *gorm.DB
}

func NewSavedPropertiesHandler(db *gorm.DB) *SavedPropertiesHandler {
	return &SavedPropertiesHandler{db: db}
}

type SavePropertyRequest struct {
	PropertyID uint   `json:"property_id" binding:"required"`
	SessionID  string `json:"session_id"`
	Email      string `json:"email"`
	Notes      string `json:"notes"`
}

func (h *SavedPropertiesHandler) SaveProperty(c *gin.Context) {
	var req SavePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = c.GetString("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	var property models.Property
	if err := h.db.First(&property, req.PropertyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	savedProperty := models.SavedProperty{
		SessionID:  sessionID,
		PropertyID: req.PropertyID,
		Email:      req.Email,
		SavedAt:    time.Now(),
		Notes:      req.Notes,
	}

	if err := h.db.Create(&savedProperty).Error; err != nil {
		if err.Error() == "duplicate key value violates unique constraint" ||
			err.Error() == "UNIQUE constraint failed" {
			c.JSON(http.StatusConflict, gin.H{"error": "Property already saved"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save property"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Property saved successfully",
		"saved_property": savedProperty,
	})
}

func (h *SavedPropertiesHandler) UnsaveProperty(c *gin.Context) {
	propertyIDStr := c.Param("id")
	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	result := h.db.Where("session_id = ? AND property_id = ?", sessionID, uint(propertyID)).
		Delete(&models.SavedProperty{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsave property"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Saved property not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Property unsaved successfully"})
}

func (h *SavedPropertiesHandler) GetSavedProperties(c *gin.Context) {
	sessionID := c.Query("session_id")
	email := c.Query("email")

	if sessionID == "" {
		sessionID = c.GetString("session_id")
	}

	if sessionID == "" && email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID or email required"})
		return
	}

	var savedProperties []models.SavedProperty
	query := h.db.Preload("Property").Order("saved_at DESC")

	if email != "" {
		query = query.Where("email = ?", email)
	} else {
		query = query.Where("session_id = ?", sessionID)
	}

	if err := query.Find(&savedProperties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve saved properties"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"saved_properties": savedProperties,
		"count":            len(savedProperties),
	})
}

func (h *SavedPropertiesHandler) CheckIfSaved(c *gin.Context) {
	propertyIDStr := c.Param("id")
	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
	}

	if sessionID == "" {
		c.JSON(http.StatusOK, gin.H{"is_saved": false})
		return
	}

	var count int64
	h.db.Model(&models.SavedProperty{}).
		Where("session_id = ? AND property_id = ?", sessionID, uint(propertyID)).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{"is_saved": count > 0})
}
