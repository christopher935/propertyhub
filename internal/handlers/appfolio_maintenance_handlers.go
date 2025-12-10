package handlers

import (
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppFolioMaintenanceHandlers struct {
	db              *gorm.DB
	maintenanceSync *services.AppFolioMaintenanceSync
	aiTriage        *services.MaintenanceAITriage
}

func NewAppFolioMaintenanceHandlers(db *gorm.DB, maintenanceSync *services.AppFolioMaintenanceSync) *AppFolioMaintenanceHandlers {
	return &AppFolioMaintenanceHandlers{
		db:              db,
		maintenanceSync: maintenanceSync,
		aiTriage:        services.NewMaintenanceAITriage(db),
	}
}

func (h *AppFolioMaintenanceHandlers) GetMaintenanceRequests(c *gin.Context) {
	filters := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}

	requests, err := h.maintenanceSync.GetAllRequests(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch maintenance requests",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"requests": requests,
		"total":    len(requests),
	})
}

func (h *AppFolioMaintenanceHandlers) GetOpenRequests(c *gin.Context) {
	requests, err := h.maintenanceSync.GetOpenRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch open maintenance requests",
			"error":   err.Error(),
		})
		return
	}

	stats, _ := h.maintenanceSync.GetMaintenanceStats()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"requests": requests,
		"total":    len(requests),
		"stats":    stats,
	})
}

func (h *AppFolioMaintenanceHandlers) GetMaintenanceRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request ID",
		})
		return
	}

	request, err := h.maintenanceSync.GetRequestByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Maintenance request not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"request": request,
	})
}

func (h *AppFolioMaintenanceHandlers) RunTriage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request ID",
		})
		return
	}

	triageResult, err := h.maintenanceSync.RunTriageOnRequest(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to run triage",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Triage completed successfully",
		"triage_result": triageResult,
	})
}

func (h *AppFolioMaintenanceHandlers) AssignVendor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request ID",
		})
		return
	}

	var req struct {
		VendorID      uint       `json:"vendor_id" binding:"required"`
		ScheduledDate *time.Time `json:"scheduled_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	err = h.maintenanceSync.AssignVendor(uint(id), req.VendorID, req.ScheduledDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to assign vendor",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vendor assigned successfully",
	})
}

func (h *AppFolioMaintenanceHandlers) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request ID",
		})
		return
	}

	var req struct {
		Status    string `json:"status" binding:"required"`
		Notes     string `json:"notes"`
		ChangedBy string `json:"changed_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	validStatuses := []string{
		models.MaintenanceStatusOpen,
		models.MaintenanceStatusInProgress,
		models.MaintenanceStatusCompleted,
		models.MaintenanceStatusCancelled,
	}

	isValidStatus := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"message":        "Invalid status value",
			"valid_statuses": validStatuses,
		})
		return
	}

	changedBy := req.ChangedBy
	if changedBy == "" {
		changedBy = "Admin"
	}

	err = h.maintenanceSync.UpdateRequestStatus(uint(id), req.Status, req.Notes, changedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update status",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status updated successfully",
	})
}

func (h *AppFolioMaintenanceHandlers) SyncFromAppFolio(c *gin.Context) {
	result, err := h.maintenanceSync.SyncMaintenanceRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Sync failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sync completed",
		"result":  result,
	})
}

func (h *AppFolioMaintenanceHandlers) GetMaintenanceStats(c *gin.Context) {
	stats, err := h.maintenanceSync.GetMaintenanceStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get maintenance stats",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

func (h *AppFolioMaintenanceHandlers) GetVendors(c *gin.Context) {
	var vendors []models.Vendor
	query := h.db.Where("is_active = ?", true)

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if preferred := c.Query("preferred"); preferred == "true" {
		query = query.Where("is_preferred = ?", true)
	}

	query.Order("is_preferred DESC, rating DESC").Find(&vendors)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"vendors": vendors,
		"total":   len(vendors),
	})
}

func (h *AppFolioMaintenanceHandlers) GetVendor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid vendor ID",
		})
		return
	}

	var vendor models.Vendor
	if err := h.db.First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Vendor not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"vendor":  vendor,
	})
}

func (h *AppFolioMaintenanceHandlers) CreateVendor(c *gin.Context) {
	var vendor models.Vendor
	if err := c.ShouldBindJSON(&vendor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if vendor.Name == "" || vendor.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Name and category are required",
		})
		return
	}

	vendor.IsActive = true
	if err := h.db.Create(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create vendor",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Vendor created successfully",
		"vendor":  vendor,
	})
}

func (h *AppFolioMaintenanceHandlers) UpdateVendor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid vendor ID",
		})
		return
	}

	var vendor models.Vendor
	if err := h.db.First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Vendor not found",
		})
		return
	}

	var updates models.Vendor
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	h.db.Model(&vendor).Updates(updates)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vendor updated successfully",
		"vendor":  vendor,
	})
}

func (h *AppFolioMaintenanceHandlers) DeleteVendor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid vendor ID",
		})
		return
	}

	var vendor models.Vendor
	if err := h.db.First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Vendor not found",
		})
		return
	}

	vendor.IsActive = false
	h.db.Save(&vendor)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vendor deactivated successfully",
	})
}

func (h *AppFolioMaintenanceHandlers) GetMaintenanceHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request ID",
		})
		return
	}

	var statusLogs []models.MaintenanceStatusLog
	h.db.Where("maintenance_request_id = ?", id).
		Order("created_at DESC").
		Find(&statusLogs)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": statusLogs,
	})
}

func (h *AppFolioMaintenanceHandlers) CreateMaintenanceRequest(c *gin.Context) {
	var req struct {
		PropertyAddress   string `json:"property_address" binding:"required"`
		UnitNumber        string `json:"unit_number"`
		TenantName        string `json:"tenant_name"`
		TenantPhone       string `json:"tenant_phone"`
		TenantEmail       string `json:"tenant_email"`
		Description       string `json:"description" binding:"required"`
		PermissionToEnter bool   `json:"permission_to_enter"`
		PetOnPremises     bool   `json:"pet_on_premises"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	tempRequest := models.MaintenanceRequest{
		PropertyAddress:   req.PropertyAddress,
		UnitNumber:        req.UnitNumber,
		TenantName:        req.TenantName,
		TenantPhone:       req.TenantPhone,
		TenantEmail:       req.TenantEmail,
		Description:       req.Description,
		PermissionToEnter: req.PermissionToEnter,
		PetOnPremises:     req.PetOnPremises,
	}

	triageResult, err := h.aiTriage.TriageRequest(tempRequest)
	if err != nil {
		triageResult = &services.TriageResult{
			Priority:     models.MaintenancePriorityMedium,
			Category:     models.MaintenanceCategoryGeneral,
			ResponseTime: models.ResponseTime48Hours,
		}
	}

	now := time.Now()
	newRequest := models.MaintenanceRequest{
		AppFolioID:        "MANUAL-" + strconv.FormatInt(now.UnixNano(), 36),
		PropertyAddress:   req.PropertyAddress,
		UnitNumber:        req.UnitNumber,
		TenantName:        req.TenantName,
		TenantPhone:       req.TenantPhone,
		TenantEmail:       req.TenantEmail,
		Description:       req.Description,
		Category:          triageResult.Category,
		Priority:          triageResult.Priority,
		Status:            models.MaintenanceStatusOpen,
		SuggestedVendor:   triageResult.SuggestedVendor,
		ResponseTime:      triageResult.ResponseTime,
		EstimatedCost:     &triageResult.EstimatedCost,
		PermissionToEnter: req.PermissionToEnter,
		PetOnPremises:     req.PetOnPremises,
		AITriageResult: models.TriageJSON{
			Priority:        triageResult.Priority,
			Category:        triageResult.Category,
			SuggestedVendor: triageResult.SuggestedVendor,
			EstimatedCost:   triageResult.EstimatedCost,
			ResponseTime:    triageResult.ResponseTime,
			AIReasoning:     triageResult.AIReasoning,
			Keywords:        triageResult.Keywords,
			ConfidenceScore: triageResult.ConfidenceScore,
			TriagedAt:       now.Format(time.RFC3339),
		},
	}

	if err := h.db.Create(&newRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create maintenance request",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "Maintenance request created successfully",
		"request":       newRequest,
		"triage_result": triageResult,
	})
}

func (h *AppFolioMaintenanceHandlers) GetDashboardWidget(c *gin.Context) {
	var openRequests int64
	h.db.Model(&models.MaintenanceRequest{}).
		Where("status IN ?", []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Count(&openRequests)

	var emergencyRequests int64
	h.db.Model(&models.MaintenanceRequest{}).
		Where("priority = ? AND status IN ?", models.MaintenancePriorityEmergency, []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Count(&emergencyRequests)

	var highPriorityRequests int64
	h.db.Model(&models.MaintenanceRequest{}).
		Where("priority = ? AND status IN ?", models.MaintenancePriorityHigh, []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Count(&highPriorityRequests)

	var recentRequests []struct {
		ID              uint      `json:"id"`
		PropertyAddress string    `json:"property_address"`
		Description     string    `json:"description"`
		Priority        string    `json:"priority"`
		Status          string    `json:"status"`
		Category        string    `json:"category"`
		CreatedAt       time.Time `json:"created_at"`
	}
	h.db.Model(&models.MaintenanceRequest{}).
		Select("id, property_address, description, priority, status, category, created_at").
		Where("status IN ?", []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Order("CASE priority WHEN 'emergency' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 END").
		Order("created_at DESC").
		Limit(5).
		Find(&recentRequests)

	var avgResolutionTime float64
	h.db.Model(&models.MaintenanceRequest{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_date - created_at))/3600), 0)").
		Where("status = ? AND completed_date IS NOT NULL", models.MaintenanceStatusCompleted).
		Scan(&avgResolutionTime)

	c.JSON(http.StatusOK, gin.H{
		"success":             true,
		"open_requests":       openRequests,
		"emergency_requests":  emergencyRequests,
		"urgent_requests":     emergencyRequests + highPriorityRequests,
		"avg_resolution_time": avgResolutionTime,
		"recent_requests":     recentRequests,
	})
}

func RegisterAppFolioMaintenanceRoutes(r *gin.Engine, db *gorm.DB, maintenanceSync *services.AppFolioMaintenanceSync) {
	handlers := NewAppFolioMaintenanceHandlers(db, maintenanceSync)

	api := r.Group("/api/v1/appfolio")
	{
		maintenance := api.Group("/maintenance")
		{
			maintenance.GET("", handlers.GetMaintenanceRequests)
			maintenance.GET("/open", handlers.GetOpenRequests)
			maintenance.GET("/stats", handlers.GetMaintenanceStats)
			maintenance.GET("/dashboard-widget", handlers.GetDashboardWidget)
			maintenance.POST("/sync", handlers.SyncFromAppFolio)
			maintenance.POST("", handlers.CreateMaintenanceRequest)

			maintenance.GET("/:id", handlers.GetMaintenanceRequest)
			maintenance.GET("/:id/history", handlers.GetMaintenanceHistory)
			maintenance.POST("/:id/triage", handlers.RunTriage)
			maintenance.PUT("/:id/assign", handlers.AssignVendor)
			maintenance.PUT("/:id/status", handlers.UpdateStatus)
		}

		vendors := api.Group("/vendors")
		{
			vendors.GET("", handlers.GetVendors)
			vendors.GET("/:id", handlers.GetVendor)
			vendors.POST("", handlers.CreateVendor)
			vendors.PUT("/:id", handlers.UpdateVendor)
			vendors.DELETE("/:id", handlers.DeleteVendor)
		}
	}
}
