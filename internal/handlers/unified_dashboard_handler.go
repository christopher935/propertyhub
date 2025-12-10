package handlers

import (
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
)

type UnifiedDashboardHandler struct {
	orchestrator *services.IntegrationOrchestrator
}

func NewUnifiedDashboardHandler(orchestrator *services.IntegrationOrchestrator) *UnifiedDashboardHandler {
	return &UnifiedDashboardHandler{
		orchestrator: orchestrator,
	}
}

func (h *UnifiedDashboardHandler) GetUnifiedDashboard(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	dashboard, err := h.orchestrator.GetUnifiedDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate dashboard",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *UnifiedDashboardHandler) TriggerFullSync(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	report, err := h.orchestrator.RunFullSync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Sync failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Full sync completed",
		"report": report,
	})
}

func (h *UnifiedDashboardHandler) TriggerPropertySync(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	result, err := h.orchestrator.SyncPropertiesFromAppFolio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Property sync failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Property sync completed",
		"result": result,
	})
}

func (h *UnifiedDashboardHandler) TriggerMaintenanceSync(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	result, err := h.orchestrator.SyncMaintenanceFromAppFolio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Maintenance sync failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Maintenance sync completed",
		"result": result,
	})
}

func (h *UnifiedDashboardHandler) TriggerLeadSync(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	synced, err := h.orchestrator.SyncLeadsWithFUB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Lead sync failed",
			"details": err.Error(),
			"synced": synced,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lead sync completed",
		"synced": synced,
	})
}

func (h *UnifiedDashboardHandler) RetryFailedSyncs(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	report, err := h.orchestrator.RetryFailedSyncs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Retry failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Retry completed",
		"report": report,
	})
}

func (h *UnifiedDashboardHandler) GetSyncHistory(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	reports, err := h.orchestrator.GetSyncHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sync history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"count": len(reports),
	})
}

func (h *UnifiedDashboardHandler) GetLastSyncReport(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	report, err := h.orchestrator.GetLastSyncReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get last sync report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *UnifiedDashboardHandler) HandleWebhook(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Integration orchestrator not configured",
		})
		return
	}

	source := c.Param("source")
	if source == "" {
		source = c.Query("source")
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payload",
			"details": err.Error(),
		})
		return
	}

	eventType, _ := payload["event_type"].(string)
	if eventType == "" {
		eventType, _ = payload["type"].(string)
	}

	err := h.orchestrator.HandleWebhook(source, eventType, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Webhook processing failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook processed",
		"source": source,
		"event_type": eventType,
	})
}

func (h *UnifiedDashboardHandler) GetIntegrationStatus(c *gin.Context) {
	if h.orchestrator == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "not_configured",
			"fub_connected": false,
			"appfolio_connected": false,
		})
		return
	}

	dashboard, _ := h.orchestrator.GetUnifiedDashboard()

	c.JSON(http.StatusOK, gin.H{
		"status": "operational",
		"fub_connected": dashboard.SystemHealth.FUBConnected,
		"appfolio_connected": dashboard.SystemHealth.AppFolioConnected,
		"queued_items": dashboard.SystemHealth.QueuedSyncItems,
		"failed_items": dashboard.SystemHealth.FailedSyncItems,
		"last_sync": dashboard.LastSync,
		"checked_at": time.Now(),
	})
}

func (h *UnifiedDashboardHandler) RegisterRoutes(router *gin.RouterGroup) {
	dashboard := router.Group("/dashboard")
	{
		dashboard.GET("/unified", h.GetUnifiedDashboard)
		dashboard.GET("/status", h.GetIntegrationStatus)
	}

	sync := router.Group("/sync")
	{
		sync.POST("/full", h.TriggerFullSync)
		sync.POST("/properties", h.TriggerPropertySync)
		sync.POST("/maintenance", h.TriggerMaintenanceSync)
		sync.POST("/leads", h.TriggerLeadSync)
		sync.POST("/retry", h.RetryFailedSyncs)
		sync.GET("/history", h.GetSyncHistory)
		sync.GET("/last", h.GetLastSyncReport)
	}

	router.POST("/webhook/:source", h.HandleWebhook)
	router.POST("/webhook", h.HandleWebhook)
}
