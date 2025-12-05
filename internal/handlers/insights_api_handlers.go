package handlers

import (
	"github.com/gin-gonic/gin"
	
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"
)

type InsightsAPIHandlers struct {
	insightsGenerator *services.InsightGeneratorService
}

func NewInsightsAPIHandlers(insightsGenerator *services.InsightGeneratorService) *InsightsAPIHandlers {
	return &InsightsAPIHandlers{
		insightsGenerator: insightsGenerator,
	}
}

func (h *InsightsAPIHandlers) GetPredictiveInsights(c *gin.Context) {
	dashboardInsights, err := h.insightsGenerator.GenerateDashboardInsights()
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to generate insights", err)
		return
	}
	
	utils.SuccessResponse(c, gin.H{
		"insights": dashboardInsights.Insights,
		"metrics":  dashboardInsights.Metrics,
		"count":    len(dashboardInsights.Insights),
	})
}
