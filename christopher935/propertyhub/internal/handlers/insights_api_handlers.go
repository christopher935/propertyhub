package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"
)

type InsightsAPIHandlers struct {
	db                *gorm.DB
	insightsGenerator *services.IntelligentInsightsGenerator
}

func NewInsightsAPIHandlers(db *gorm.DB) *InsightsAPIHandlers {
	return &InsightsAPIHandlers{
		db:                db,
		insightsGenerator: services.NewIntelligentInsightsGenerator(db),
	}
}

func (h *InsightsAPIHandlers) GetPredictiveInsights(c *gin.Context) {
	insights, err := h.insightsGenerator.GenerateAllInsights()
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to generate insights", err)
		return
	}
	
	utils.SuccessResponse(c, gin.H{
		"insights": insights,
		"count":    len(insights),
	})
}
