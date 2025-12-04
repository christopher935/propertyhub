package handlers

import (
	"chrisgross-ctrl-project/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIHandlers holds references to services
type APIHandlers struct {
	BIService *services.BusinessIntelligenceService
}

// NewAPIHandlers creates a new API handlers instance
func NewAPIHandlers(biService *services.BusinessIntelligenceService) *APIHandlers {
	return &APIHandlers{
		BIService: biService,
	}
}

// GetDashboardMetrics returns dashboard metrics as JSON
func (h *APIHandlers) GetDashboardMetrics(c *gin.Context) {
	metrics, err := h.BIService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch dashboard metrics",
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetLeadSegments returns lead segment breakdown
func (h *APIHandlers) GetLeadSegments(c *gin.Context) {
	metrics, err := h.BIService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch lead segments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hot_leads":     metrics.LeadMetrics.HotLeads,
		"warm_leads":    metrics.LeadMetrics.WarmLeads,
		"cold_leads":    metrics.LeadMetrics.ColdLeads,
		"dormant_leads": metrics.LeadMetrics.DormantLeads,
		"total_leads":   metrics.LeadMetrics.TotalLeads,
	})
}

// GetBookingMetrics returns booking statistics
func (h *APIHandlers) GetBookingMetrics(c *gin.Context) {
	metrics, err := h.BIService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch booking metrics",
		})
		return
	}

	c.JSON(http.StatusOK, metrics.BookingMetrics)
}

// GetPropertyMetrics returns property statistics
func (h *APIHandlers) GetPropertyMetrics(c *gin.Context) {
	metrics, err := h.BIService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch property metrics",
		})
		return
	}

	c.JSON(http.StatusOK, metrics.PropertyMetrics)
}

// GetSystemHealth returns system health status
func (h *APIHandlers) GetSystemHealth(c *gin.Context) {
	metrics, err := h.BIService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch system health",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"system_health":  metrics.SystemHealth,
		"system_metrics": metrics.SystemMetrics,
	})
}
