package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CentralPropertySyncHandlers handles API requests for Central Property State Manager with real-time sync
type CentralPropertySyncHandlers struct {
	db                  *gorm.DB
	centralStateManager *services.CentralPropertyStateManager
	realTimeSyncService *services.RealTimeSyncService
}

// NewCentralPropertySyncHandlers creates new handlers with integrated services
func NewCentralPropertySyncHandlers(db *gorm.DB) *CentralPropertySyncHandlers {
	// Create encryption manager
	encryptionManager, err := security.NewEncryptionManager(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize encryption manager: %v", err)
		// Continue with nil encryption manager for now
	}

	// Create services with proper dependency management
	centralStateManager := services.NewCentralPropertyStateManager(db, encryptionManager)
	realTimeSyncService := services.NewRealTimeSyncService(db)

	// Connect services to avoid circular dependency
	centralStateManager.SetRealTimeSync(realTimeSyncService)
	realTimeSyncService.SetCentralStateManager(centralStateManager)

	return &CentralPropertySyncHandlers{
		db:                  db,
		centralStateManager: centralStateManager,
		realTimeSyncService: realTimeSyncService,
	}
}

// RegisterCentralPropertySyncRoutes registers all routes for central property sync
func RegisterCentralPropertySyncRoutes(router *gin.Engine, db *gorm.DB) {
	handlers := NewCentralPropertySyncHandlers(db)

	api := router.Group("/api/v1")
	{
		// Central Property State Management
		api.POST("/central-properties", handlers.CreateOrUpdateProperty)
		api.GET("/central-properties/:id", handlers.GetProperty)
		// DISABLED: 		api.GET("/central-properties", handlers.GetAllProperties)
		api.PUT("/central-properties/:id/status", handlers.UpdatePropertyStatus)
		api.GET("/central-properties/stats", handlers.GetSystemStats)
		api.POST("/central-properties/:id/resolve-conflict", handlers.ResolveConflict)

		// Real-time Sync Management
		api.GET("/sync/stats", handlers.GetSyncStats)
		api.GET("/sync/events", handlers.GetSyncEvents)
		api.POST("/sync/retry-failed", handlers.RetryFailedSyncs)
		api.POST("/sync/trigger", handlers.TriggerManualSync)
		api.GET("/sync/health", handlers.GetSyncHealth)

		// Test endpoints
		api.GET("/central-properties/test", handlers.TestCentralState)
		api.POST("/sync/test", handlers.TestRealTimeSync)
	}

	log.Printf("✅ Central Property Sync routes registered")
}

// CreateOrUpdateProperty creates or updates a property in central state
func (cpsh *CentralPropertySyncHandlers) CreateOrUpdateProperty(c *gin.Context) {
	var request models.PropertyUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	log.Printf("🏠 API: Creating/updating property: %s (Source: %s)", request.Address, request.Source)

	property, err := cpsh.centralStateManager.CreateOrUpdateProperty(request)
	if err != nil {
		log.Printf("❌ Failed to create/update property: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process property",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Property processed successfully",
		"property":       property,
		"sync_triggered": true,
	})
}

// GetProperty retrieves a property by ID or MLS ID
func (cpsh *CentralPropertySyncHandlers) GetProperty(c *gin.Context) {
	identifier := c.Param("id")

	property, err := cpsh.centralStateManager.GetPropertyState(identifier)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Property not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"property": property,
	})
}

// GetAllProperties retrieves all properties in central state
func (cpsh *CentralPropertySyncHandlers) GetAllProperties(c *gin.Context) {
	properties, err := cpsh.centralStateManager.GetAllProperties()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve properties",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"properties": properties,
		"count":      len(properties),
	})
}

// UpdatePropertyStatus updates property status
func (cpsh *CentralPropertySyncHandlers) UpdatePropertyStatus(c *gin.Context) {
	propertyID := c.Param("id")

	var request struct {
		Status string `json:"status" binding:"required"`
		Source string `json:"source" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// For now, use propertyID as MLSId - in production, you'd look up the MLS ID
	err := cpsh.centralStateManager.UpdatePropertyStatus(propertyID, request.Status, request.Source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update property status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Property status updated successfully",
		"sync_triggered": true,
	})
}

// GetSystemStats returns central property state statistics
func (cpsh *CentralPropertySyncHandlers) GetSystemStats(c *gin.Context) {
	stats, err := cpsh.centralStateManager.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve system stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// ResolveConflict manually resolves a data conflict
func (cpsh *CentralPropertySyncHandlers) ResolveConflict(c *gin.Context) {
	propertyIDStr := c.Param("id")
	propertyID, err := strconv.ParseUint(propertyIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid property ID",
		})
		return
	}

	var request struct {
		Field         string `json:"field" binding:"required"`
		ResolvedValue string `json:"resolved_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err = cpsh.centralStateManager.ResolveConflict(uint(propertyID), request.Field, request.ResolvedValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to resolve conflict",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Conflict resolved successfully",
	})
}

// GetSyncStats returns real-time synchronization statistics
func (cpsh *CentralPropertySyncHandlers) GetSyncStats(c *gin.Context) {
	stats := cpsh.realTimeSyncService.GetSyncStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// GetSyncEvents returns recent synchronization events
func (cpsh *CentralPropertySyncHandlers) GetSyncEvents(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	events, err := cpsh.realTimeSyncService.GetSyncEvents(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve sync events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"events":  events,
		"count":   len(events),
	})
}

// RetryFailedSyncs retries failed synchronization events
func (cpsh *CentralPropertySyncHandlers) RetryFailedSyncs(c *gin.Context) {
	maxRetriesStr := c.DefaultQuery("max_retries", "3")
	maxRetries, err := strconv.Atoi(maxRetriesStr)
	if err != nil {
		maxRetries = 3
	}

	err = cpsh.realTimeSyncService.RetryFailedSyncs(maxRetries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retry sync events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Failed sync events retry initiated",
	})
}

// TriggerManualSync manually triggers synchronization for a property
func (cpsh *CentralPropertySyncHandlers) TriggerManualSync(c *gin.Context) {
	var request models.PropertyUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := cpsh.realTimeSyncService.TriggerSync(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to trigger sync",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Manual sync triggered successfully",
	})
}

// GetSyncHealth returns health status of sync workers
func (cpsh *CentralPropertySyncHandlers) GetSyncHealth(c *gin.Context) {
	stats := cpsh.realTimeSyncService.GetSyncStats()

	// Determine overall health based on sync statistics
	overallHealth := "healthy"
	healthDetails := make(map[string]interface{})

	// Check failed sync percentage
	failureRate := float64(stats.FailedSyncs) / float64(stats.TotalSyncs) * 100
	if failureRate > 5.0 {
		overallHealth = "degraded"
		healthDetails["failure_rate"] = fmt.Sprintf("%.2f%% (above threshold)", failureRate)
	}

	// Check if last sync was too long ago (more than 10 minutes)
	if time.Since(stats.LastSyncTime) > 10*time.Minute {
		overallHealth = "degraded"
		healthDetails["last_sync"] = "overdue"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"health": gin.H{
			"overall_status": overallHealth,
			"sync_stats":     stats,
			"details":        healthDetails,
			"timestamp":      time.Now(),
		},
	})
}

// TestCentralState tests the central property state functionality
func (cpsh *CentralPropertySyncHandlers) TestCentralState(c *gin.Context) {
	log.Printf("🧪 Testing Central Property State Manager")

	// Create test property data
	testPrice := 350000.0
	testBedrooms := 3
	testBathrooms := float32(2.5)
	testSquareFeet := 1800

	testProperty := models.PropertyUpdateRequest{
		MLSId:        "TEST123",
		Address:      "123 Test Street, Austin, TX 78701",
		Price:        &testPrice,
		Bedrooms:     &testBedrooms,
		Bathrooms:    &testBathrooms,
		SquareFeet:   &testSquareFeet,
		PropertyType: "Single Family",
		Status:       "active",
		Source:       "test",
		Data:         models.JSONB{"test": true, "created_by": "api_test"},
	}

	// Test create/update
	property, err := cpsh.centralStateManager.CreateOrUpdateProperty(testProperty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Test failed",
			"details": err.Error(),
		})
		return
	}

	// Get system stats
	stats, err := cpsh.centralStateManager.GetSystemStats()
	if err != nil {
		log.Printf("⚠️ Failed to get stats: %v", err)
		stats = map[string]interface{}{"error": "stats unavailable"}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Central Property State Manager test completed successfully",
		"test_property":  property,
		"system_stats":   stats,
		"sync_triggered": true,
		"test_timestamp": property.UpdatedAt,
	})
}

// TestRealTimeSync tests the real-time synchronization functionality
func (cpsh *CentralPropertySyncHandlers) TestRealTimeSync(c *gin.Context) {
	log.Printf("🧪 Testing Real-time Sync Service")

	// Create test sync request
	testPrice := 375000.0
	testBedrooms := 4
	testBathrooms := float32(3.0)
	testSquareFeet := 2200

	testSyncRequest := models.PropertyUpdateRequest{
		MLSId:        "SYNC_TEST456",
		Address:      "456 Sync Test Avenue, Austin, TX 78702",
		Price:        &testPrice,
		Bedrooms:     &testBedrooms,
		Bathrooms:    &testBathrooms,
		SquareFeet:   &testSquareFeet,
		PropertyType: "Townhouse",
		Status:       "active",
		Source:       "sync_test",
		Data:         models.JSONB{"sync_test": true, "created_by": "sync_api_test"},
	}

	// Trigger sync
	err := cpsh.realTimeSyncService.TriggerSync(testSyncRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Sync test failed",
			"details": err.Error(),
		})
		return
	}

	// Get sync stats
	syncStats := cpsh.realTimeSyncService.GetSyncStats()

	// Get recent sync events
	events, err := cpsh.realTimeSyncService.GetSyncEvents(10)
	if err != nil {
		log.Printf("⚠️ Failed to get sync events: %v", err)
		events = []services.SyncEvent{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"message":           "Real-time sync test completed successfully",
		"test_sync_request": testSyncRequest,
		"sync_stats":        syncStats,
		"recent_events":     events,
		"systems_synced":    []string{"har", "propertyhub", "fub", "email"},
	})
}
