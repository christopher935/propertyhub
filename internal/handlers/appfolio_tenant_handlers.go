package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppFolioTenantHandlers struct {
	db           *gorm.DB
	tenantSync   *services.AppFolioTenantSync
	propertySync *services.AppFolioPropertySync
	apiClient    *services.AppFolioAPIClient
}

func NewAppFolioTenantHandlers(db *gorm.DB) *AppFolioTenantHandlers {
	clientID := os.Getenv("APPFOLIO_CLIENT_ID")
	clientSecret := os.Getenv("APPFOLIO_CLIENT_SECRET")
	baseURL := os.Getenv("APPFOLIO_API_URL")

	if clientID == "" || clientSecret == "" {
		log.Println("⚠️ AppFolio credentials not configured - tenant sync will be disabled")
		return &AppFolioTenantHandlers{db: db}
	}

	apiClient := services.NewAppFolioAPIClient(clientID, clientSecret, baseURL)
	propertySync := services.NewAppFolioPropertySync(db, apiClient)
	tenantSync := services.NewAppFolioTenantSync(db, apiClient, propertySync)

	return &AppFolioTenantHandlers{
		db:           db,
		tenantSync:   tenantSync,
		propertySync: propertySync,
		apiClient:    apiClient,
	}
}

func (h *AppFolioTenantHandlers) PushBookingToAppFolio(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	bookingIDStr := c.Param("booking_id")
	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid booking ID",
		})
		return
	}

	var booking models.Booking
	if err := h.db.Preload("Property").First(&booking, uint(bookingID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Booking not found",
		})
		return
	}

	tenant, err := h.tenantSync.PushTenantToAppFolio(booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant pushed to AppFolio successfully",
		"data": gin.H{
			"tenant_id": tenant.ID,
			"name":      tenant.Name,
			"email":     tenant.Email,
		},
	})
}

func (h *AppFolioTenantHandlers) PushApplicationToAppFolio(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	applicationIDStr := c.Param("application_id")
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid application ID",
		})
		return
	}

	var appNumber models.ApplicationNumber
	if err := h.db.Preload("PropertyApplicationGroup").First(&appNumber, uint(applicationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Application not found",
		})
		return
	}

	if appNumber.Status != models.AppStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Application must be approved before pushing to AppFolio",
		})
		return
	}

	tenant, err := h.tenantSync.PushTenantFromApplication(appNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant created in AppFolio from approved application",
		"data": gin.H{
			"tenant_id": tenant.ID,
			"name":      tenant.Name,
			"email":     tenant.Email,
		},
	})
}

func (h *AppFolioTenantHandlers) SyncTenantsFromAppFolio(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	result, err := h.tenantSync.SyncTenantsFromAppFolio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant sync completed",
		"data":    result,
	})
}

func (h *AppFolioTenantHandlers) GetTenantFromAppFolio(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant ID is required",
		})
		return
	}

	tenant, err := h.tenantSync.GetTenantByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if tenant == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tenant not found in AppFolio",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tenant,
	})
}

func (h *AppFolioTenantHandlers) GetTenantByEmail(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Email parameter is required",
		})
		return
	}

	tenant, err := h.tenantSync.GetTenantByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if tenant == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tenant not found in AppFolio",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tenant,
	})
}

func (h *AppFolioTenantHandlers) UpdateTenantStatus(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant ID is required",
		})
		return
	}

	var request struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Status is required",
		})
		return
	}

	if err := h.tenantSync.UpdateTenantStatus(tenantID, request.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tenant status updated in AppFolio",
	})
}

func (h *AppFolioTenantHandlers) SyncLeadStatusToAppFolio(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	leadIDStr := c.Param("lead_id")
	leadID, err := strconv.ParseUint(leadIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid lead ID",
		})
		return
	}

	if err := h.tenantSync.SyncTenantStatusToAppFolio(uint(leadID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Lead status synced to AppFolio",
	})
}

func (h *AppFolioTenantHandlers) GetSyncLogs(c *gin.Context) {
	if h.tenantSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	direction := c.Query("direction")

	var logs []services.TenantSyncLog
	var err error

	if direction != "" {
		logs, err = h.tenantSync.GetSyncLogsByDirection(direction, limit)
	} else {
		logs, err = h.tenantSync.GetSyncLogs(limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":  logs,
			"count": len(logs),
		},
	})
}

func (h *AppFolioTenantHandlers) GetSyncStatus(c *gin.Context) {
	configured := h.tenantSync != nil

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"configured":    configured,
			"client_id_set": os.Getenv("APPFOLIO_CLIENT_ID") != "",
			"api_url":       os.Getenv("APPFOLIO_API_URL"),
		},
	})
}

func (h *AppFolioTenantHandlers) SyncPropertiesFromAppFolio(c *gin.Context) {
	if h.propertySync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	result, err := h.propertySync.SyncPropertiesFromAppFolio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Property sync completed",
		"data":    result,
	})
}

func RegisterAppFolioRoutes(r *gin.RouterGroup, db *gorm.DB) {
	handler := NewAppFolioTenantHandlers(db)

	appfolio := r.Group("/appfolio")
	{
		appfolio.GET("/status", handler.GetSyncStatus)

		tenants := appfolio.Group("/tenants")
		{
			tenants.POST("/push/:booking_id", handler.PushBookingToAppFolio)
			tenants.POST("/push-application/:application_id", handler.PushApplicationToAppFolio)
			tenants.POST("/sync", handler.SyncTenantsFromAppFolio)
			tenants.GET("/:id", handler.GetTenantFromAppFolio)
			tenants.GET("/search", handler.GetTenantByEmail)
			tenants.PUT("/:id/status", handler.UpdateTenantStatus)
			tenants.POST("/sync-lead/:lead_id", handler.SyncLeadStatusToAppFolio)
		}

		properties := appfolio.Group("/properties")
		{
			properties.POST("/sync", handler.SyncPropertiesFromAppFolio)
		}

		appfolio.GET("/logs", handler.GetSyncLogs)
	}
}
