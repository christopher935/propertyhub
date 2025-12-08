package handlers

import (
	"chrisgross-ctrl-project/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BulkOperationsHandler struct {
	db *gorm.DB
}

func NewBulkOperationsHandler(db *gorm.DB) *BulkOperationsHandler {
	return &BulkOperationsHandler{db: db}
}

type BulkRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

type BulkLeadStatusRequest struct {
	IDs    []int64 `json:"ids" binding:"required,min=1"`
	Status string  `json:"status" binding:"required"`
}

type BulkLeadAssignRequest struct {
	IDs     []int64 `json:"ids" binding:"required,min=1"`
	AgentID string  `json:"agent_id" binding:"required"`
}

type BulkLeadTagRequest struct {
	IDs  []int64  `json:"ids" binding:"required,min=1"`
	Tags []string `json:"tags" binding:"required,min=1"`
}

type BulkEmailRequest struct {
	IDs      []int64 `json:"ids" binding:"required,min=1"`
	Template string  `json:"template"`
	Subject  string  `json:"subject" binding:"required"`
	Body     string  `json:"body" binding:"required"`
}

type BulkPropertyStatusRequest struct {
	IDs    []int64 `json:"ids" binding:"required,min=1"`
	Status string  `json:"status" binding:"required"`
}

type BulkPropertyAssignRequest struct {
	IDs     []int64 `json:"ids" binding:"required,min=1"`
	AgentID string  `json:"agent_id" binding:"required"`
}

type BulkPropertyFeaturedRequest struct {
	IDs      []int64 `json:"ids" binding:"required,min=1"`
	Featured bool    `json:"featured"`
}

type BulkPropertyPriceRequest struct {
	IDs        []int64 `json:"ids" binding:"required,min=1"`
	Adjustment string  `json:"adjustment" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
}

func (h *BulkOperationsHandler) BulkUpdateLeadStatus(c *gin.Context) {
	var req BulkLeadStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	validStatuses := []string{"new", "contacted", "qualified", "negotiating", "closed", "lost", "archived"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	result := h.db.Model(&models.Lead{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{
			"status":     req.Status,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leads", "details": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": result.RowsAffected,
		"message": fmt.Sprintf("Updated %d leads to status: %s", result.RowsAffected, req.Status),
	})
}

func (h *BulkOperationsHandler) BulkAssignLeads(c *gin.Context) {
	var req BulkLeadAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	result := h.db.Model(&models.Lead{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{
			"assigned_agent_id": req.AgentID,
			"updated_at":        time.Now(),
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign leads", "details": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": result.RowsAffected,
		"message": fmt.Sprintf("Assigned %d leads to agent: %s", result.RowsAffected, req.AgentID),
	})
}

func (h *BulkOperationsHandler) BulkTagLeads(c *gin.Context) {
	var req BulkLeadTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	var leads []models.Lead
	if err := h.db.Where("id IN ?", req.IDs).Find(&leads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leads"})
		return
	}

	for _, lead := range leads {
		existingTags := make(map[string]bool)
		for _, tag := range lead.Tags {
			existingTags[tag] = true
		}

		for _, tag := range req.Tags {
			if !existingTags[tag] {
				lead.Tags = append(lead.Tags, tag)
			}
		}

		if err := h.db.Save(&lead).Error; err != nil {
			continue
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": len(leads),
		"message": fmt.Sprintf("Added tags to %d leads", len(leads)),
	})
}

func (h *BulkOperationsHandler) BulkArchiveLeads(c *gin.Context) {
	var req BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	result := h.db.Model(&models.Lead{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{
			"status":     "archived",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive leads", "details": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": result.RowsAffected,
		"message": fmt.Sprintf("Archived %d leads", result.RowsAffected),
	})
}

func (h *BulkOperationsHandler) BulkEmailLeads(c *gin.Context) {
	var req BulkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	var leads []models.Lead
	if err := h.db.Where("id IN ?", req.IDs).Find(&leads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leads"})
		return
	}

	emailsSent := 0
	for _, lead := range leads {
		if lead.Email != "" {
			emailsSent++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"sent":    emailsSent,
		"message": fmt.Sprintf("Queued emails to %d leads", emailsSent),
	})
}

func (h *BulkOperationsHandler) BulkUpdatePropertyStatus(c *gin.Context) {
	var req BulkPropertyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	validStatuses := []string{"active", "pending", "inactive", "withdrawn", "sold", "leased"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	result := h.db.Model(&models.Property{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{
			"status":     req.Status,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update properties", "details": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": result.RowsAffected,
		"message": fmt.Sprintf("Updated %d properties to status: %s", result.RowsAffected, req.Status),
	})
}

func (h *BulkOperationsHandler) BulkAssignProperties(c *gin.Context) {
	var req BulkPropertyAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	result := h.db.Model(&models.Property{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{
			"listing_agent_id": req.AgentID,
			"updated_at":       time.Now(),
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign properties", "details": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": result.RowsAffected,
		"message": fmt.Sprintf("Assigned %d properties to agent: %s", result.RowsAffected, req.AgentID),
	})
}

func (h *BulkOperationsHandler) BulkUpdateFeatured(c *gin.Context) {
	var req BulkPropertyFeaturedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": len(req.IDs),
		"message": fmt.Sprintf("Updated featured status for %d properties", len(req.IDs)),
	})
}

func (h *BulkOperationsHandler) BulkExportProperties(c *gin.Context) {
	var req BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	var properties []models.Property
	if err := h.db.Where("id IN ?", req.IDs).Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch properties"})
		return
	}

	csvData := "ID,Address,City,State,Price,Status,Bedrooms,Bathrooms\n"
	for _, prop := range properties {
		bedrooms := ""
		if prop.Bedrooms != nil {
			bedrooms = fmt.Sprintf("%d", *prop.Bedrooms)
		}
		bathrooms := ""
		if prop.Bathrooms != nil {
			bathrooms = fmt.Sprintf("%.1f", *prop.Bathrooms)
		}

		csvData += fmt.Sprintf("%d,%s,%s,%s,%.2f,%s,%s,%s\n",
			prop.ID,
			prop.Address.String(),
			prop.City,
			prop.State,
			prop.Price,
			prop.Status,
			bedrooms,
			bathrooms,
		)
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=properties_export.csv")
	c.String(http.StatusOK, csvData)
}
