package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/utils"
)

// ClosingPipelineHandlers manages the closing pipeline workflow
type ClosingPipelineHandlers struct {
	db                *gorm.DB
	encryptionManager *security.EncryptionManager
}

// NewClosingPipelineHandlers creates new closing pipeline handlers
func NewClosingPipelineHandlers(db *gorm.DB) *ClosingPipelineHandlers {
	encryptionManager, err := security.NewEncryptionManager(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize encryption manager: %v", err)
	}
	return &ClosingPipelineHandlers{
		db:                db,
		encryptionManager: encryptionManager,
	}
}

// GetClosingPipelines retrieves all closing pipeline items
// GET /api/v1/admin/closing-pipeline
func (h *ClosingPipelineHandlers) GetClosingPipelines(c *gin.Context) {
	status := c.Query("status")

	var pipelines []models.ClosingPipeline
	query := h.db.Order("created_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&pipelines).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch pipeline items", err)
		return
	}

	// Convert to response format with decrypted fields
	pipelineResponses := models.ToClosingPipelineDataResponseList(pipelines, h.encryptionManager)

	utils.SuccessResponse(c, gin.H{
		"pipelines": pipelineResponses,
		"total":     len(pipelineResponses),
	})
}

// UpdatePipelineStage updates the stage of a closing pipeline item
// PUT /api/v1/admin/closing-pipeline/:id/stage
func (h *ClosingPipelineHandlers) UpdatePipelineStage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid pipeline ID", err)
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

	var pipeline models.ClosingPipeline
	if err := h.db.First(&pipeline, uint(id)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Pipeline item not found", err)
		return
	}

	// Validate status
	validStatuses := []string{"pending", "in_progress", "ready", "completed"}
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

	previousStatus := pipeline.Status
	pipeline.Status = request.NewStatus

	if err := h.db.Save(&pipeline).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update pipeline stage", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":         "Pipeline status updated successfully",
		"previous_status": previousStatus,
		"new_status":      request.NewStatus,
		"pipeline":        pipeline,
	})
}

// GetPipelineItem retrieves a single closing pipeline item
// GET /api/v1/admin/closing-pipeline/:id
func (h *ClosingPipelineHandlers) GetPipelineItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid pipeline ID", err)
		return
	}

	var pipeline models.ClosingPipeline
	if err := h.db.First(&pipeline, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Pipeline item not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch pipeline item", err)
		}
		return
	}

	utils.SuccessResponse(c, gin.H{
		"pipeline": pipeline,
	})
}

// CreatePipelineItem creates a new closing pipeline item
// POST /api/v1/admin/closing-pipeline
func (h *ClosingPipelineHandlers) CreatePipelineItem(c *gin.Context) {
	var pipeline models.ClosingPipeline
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	// Set defaults
	if pipeline.Status == "" {
		pipeline.Status = "pending"
	}
	if pipeline.SoldDate.IsZero() {
		pipeline.SoldDate = time.Now()
	}

	if err := h.db.Create(&pipeline).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create pipeline item", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":  "Pipeline item created successfully",
		"pipeline": pipeline,
	})
}

// DeletePipelineItem deletes a closing pipeline item
// DELETE /api/v1/admin/closing-pipeline/:id
func (h *ClosingPipelineHandlers) DeletePipelineItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid pipeline ID", err)
		return
	}

	if err := h.db.Delete(&models.ClosingPipeline{}, uint(id)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete pipeline item", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Pipeline item deleted successfully",
	})
}

// UpdateLeaseWorkflowStatus updates specific lease workflow fields
// PUT /api/v1/admin/closing-pipeline/:id/lease-status
func (h *ClosingPipelineHandlers) UpdateLeaseWorkflowStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid pipeline ID", err)
		return
	}

	var request struct {
		LeaseSentOut       *bool `json:"lease_sent_out"`
		LeaseComplete      *bool `json:"lease_complete"`
		DepositReceived    *bool `json:"deposit_received"`
		FirstMonthReceived *bool `json:"first_month_received"`
		DepositAmount      *float64 `json:"deposit_amount"`
		MonthlyRent        *float64 `json:"monthly_rent"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	var pipeline models.ClosingPipeline
	if err := h.db.First(&pipeline, uint(id)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Pipeline item not found", err)
		return
	}

	now := time.Now()

	// Update fields if provided
	if request.LeaseSentOut != nil {
		pipeline.LeaseSentOut = *request.LeaseSentOut
		if *request.LeaseSentOut {
			pipeline.LeaseSentDate = &now
		}
	}

	if request.LeaseComplete != nil {
		pipeline.LeaseComplete = *request.LeaseComplete
		if *request.LeaseComplete {
			pipeline.LeaseCompleteDate = &now
		}
	}

	if request.DepositReceived != nil {
		pipeline.DepositReceived = *request.DepositReceived
		if *request.DepositReceived {
			pipeline.DepositReceivedDate = &now
		}
	}

	if request.FirstMonthReceived != nil {
		pipeline.FirstMonthReceived = *request.FirstMonthReceived
		if *request.FirstMonthReceived {
			pipeline.FirstMonthReceivedDate = &now
		}
	}

	if request.DepositAmount != nil {
		pipeline.DepositAmount = request.DepositAmount
	}

	if request.MonthlyRent != nil {
		pipeline.MonthlyRent = request.MonthlyRent
	}

	if err := h.db.Save(&pipeline).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update lease workflow", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":  "Lease workflow updated successfully",
		"pipeline": pipeline,
	})
}