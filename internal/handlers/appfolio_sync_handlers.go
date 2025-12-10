package handlers

import (
	"log"
	"net/http"
	"os"

	"chrisgross-ctrl-project/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppFolioSyncHandlers struct {
	db          *gorm.DB
	syncService *services.AppFolioPropertySync
}

func NewAppFolioSyncHandlers(db *gorm.DB) *AppFolioSyncHandlers {
	appfolioBaseURL := os.Getenv("APPFOLIO_API_URL")
	appfolioAPIKey := os.Getenv("APPFOLIO_API_KEY")

	if appfolioBaseURL == "" {
		appfolioBaseURL = "https://api.appfolio.com"
	}

	var syncService *services.AppFolioPropertySync
	if appfolioAPIKey != "" {
		client := services.NewAppFolioAPIClient(appfolioBaseURL, appfolioAPIKey)
		syncService = services.NewAppFolioPropertySync(db, client)
		log.Printf("✅ AppFolio sync service initialized")
	} else {
		log.Printf("⚠️ AppFolio API key not configured - sync endpoints will return errors")
	}

	return &AppFolioSyncHandlers{
		db:          db,
		syncService: syncService,
	}
}

func (h *AppFolioSyncHandlers) TriggerFullSync(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured. Please set APPFOLIO_API_KEY environment variable.",
		})
		return
	}

	log.Printf("🔄 AppFolio full sync triggered via API")

	result, err := h.syncService.SyncPropertiesFromAppFolio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"result":  result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": result.Success,
		"message": result.Message,
		"result":  result,
	})
}

func (h *AppFolioSyncHandlers) GetSyncStatus(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success":    false,
			"error":      "AppFolio integration not configured",
			"configured": false,
		})
		return
	}

	lastSync, err := h.syncService.GetLastSyncStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	history, err := h.syncService.GetSyncHistory(10)
	if err != nil {
		log.Printf("Warning: failed to get sync history: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"configured": true,
		"last_sync":  lastSync,
		"history":    history,
	})
}

func (h *AppFolioSyncHandlers) SyncSingleProperty(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	appfolioID := c.Param("id")
	if appfolioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "AppFolio property ID is required",
		})
		return
	}

	log.Printf("🔄 Syncing single property: %s", appfolioID)

	result, err := h.syncService.SyncSingleProperty(appfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"result":  result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": result.Success,
		"message": result.Message,
		"result":  result,
	})
}

func (h *AppFolioSyncHandlers) GetVacancies(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	vacancies, err := h.syncService.GetVacancies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"vacancies": vacancies,
		"count":     len(vacancies),
	})
}

func (h *AppFolioSyncHandlers) GetSyncedProperties(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	properties, err := h.syncService.GetAppFolioProperties()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"properties": properties,
		"count":      len(properties),
	})
}

func (h *AppFolioSyncHandlers) GetPropertyByExternalID(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "AppFolio integration not configured",
		})
		return
	}

	externalID := c.Param("external_id")
	if externalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "External ID is required",
		})
		return
	}

	property, err := h.syncService.GetPropertyByExternalID(externalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if property == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Property not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"property": property,
	})
}

func (h *AppFolioSyncHandlers) TestConnection(c *gin.Context) {
	appfolioBaseURL := os.Getenv("APPFOLIO_API_URL")
	appfolioAPIKey := os.Getenv("APPFOLIO_API_KEY")

	if appfolioAPIKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"success":    false,
			"configured": false,
			"message":    "AppFolio API key not configured",
		})
		return
	}

	if appfolioBaseURL == "" {
		appfolioBaseURL = "https://api.appfolio.com"
	}

	client := services.NewAppFolioAPIClient(appfolioBaseURL, appfolioAPIKey)
	err := client.TestConnection()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":    false,
			"configured": true,
			"connected":  false,
			"error":      err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"configured": true,
		"connected":  true,
		"message":    "Successfully connected to AppFolio API",
	})
}
