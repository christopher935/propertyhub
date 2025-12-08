package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApprovalsManagementHandlers handles approval-related operations
type ApprovalsManagementHandlers struct {
	db *gorm.DB
}

// NewApprovalsManagementHandlers creates a new approvals management handlers instance
func NewApprovalsManagementHandlers(db *gorm.DB) *ApprovalsManagementHandlers {
	return &ApprovalsManagementHandlers{db: db}
}

// GetApprovals retrieves all approvals with optional filtering
func (h *ApprovalsManagementHandlers) GetApprovals(c *gin.Context) {
	var approvals []models.Approval
	query := h.db.Model(&models.Approval{})

	// Apply filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if approvalType := c.Query("type"); approvalType != "" {
		query = query.Where("approval_type = ?", approvalType)
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Execute query with pagination
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&approvals).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch approvals", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"approvals": approvals,
		"count":     len(approvals),
	})
}

// GetApproval retrieves a single approval by ID
func (h *ApprovalsManagementHandlers) GetApproval(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid approval ID", err)
		return
	}

	var approval models.Approval
	if err := h.db.First(&approval, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Approval not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch approval", err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"approval": approval,
	})
}

// UpdateApprovalStatus updates the status of an approval
func (h *ApprovalsManagementHandlers) UpdateApprovalStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid approval ID", err)
		return
	}

	var request struct {
		NewStatus string `json:"new_status" binding:"required"`
		Notes     string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	// Validate status
	validStatuses := []string{"pending", "approved", "rejected", "under_review"}
	isValid := false
	for _, status := range validStatuses {
		if status == request.NewStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid status value", nil)
		return
	}

	// Get existing approval
	var approval models.Approval
	if err := h.db.First(&approval, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Approval not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch approval", err)
		}
		return
	}

	// Store previous status for audit trail
	previousStatus := approval.Status
	approval.Status = request.NewStatus

	// Add status change note with timestamp
	now := time.Now()
	statusNote := fmt.Sprintf("Status changed from %s to %s at %s", previousStatus, request.NewStatus, now.Format("2006-01-02 15:04:05"))
	if request.Notes != "" {
		statusNote += " - " + request.Notes
	}

	if approval.Notes != "" {
		approval.Notes += "\n" + statusNote
	} else {
		approval.Notes = statusNote
	}

	if err := h.db.Save(&approval).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update application status", err)
		return
	}

	// Create closing pipeline entry if approved
	if request.NewStatus == "approved" {
		h.createClosingPipelineEntry(&approval)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  fmt.Sprintf("Application status updated to %s", request.NewStatus),
		"approval": approval,
	})
}

// createClosingPipelineEntry creates a closing pipeline entry for approved applications
func (h *ApprovalsManagementHandlers) createClosingPipelineEntry(approval *models.Approval) {
	now := time.Now()
	pipeline := models.ClosingPipeline{
		PropertyAddress: approval.PropertyAddress,
		SoldDate:        now,
		Status:          "pending", // pending, in_progress, ready, completed
	}

	if err := h.db.Create(&pipeline).Error; err != nil {
		fmt.Printf("Warning: Failed to create closing pipeline entry: %v\n", err)
	}
}

// CreateApproval creates a new approval request
func (h *ApprovalsManagementHandlers) CreateApproval(c *gin.Context) {
	var approval models.Approval
	if err := c.ShouldBindJSON(&approval); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	// Set defaults
	if approval.Status == "" {
		approval.Status = "pending"
	}
	if approval.Priority == "" {
		approval.Priority = "medium"
	}

	if err := h.db.Create(&approval).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create approval", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"message":  "Approval created successfully",
		"approval": approval,
	})
}

// DeleteApproval deletes an approval by ID
func (h *ApprovalsManagementHandlers) DeleteApproval(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid approval ID", err)
		return
	}

	if err := h.db.Delete(&models.Approval{}, uint(id)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete approval", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Approval deleted successfully",
	})
}
