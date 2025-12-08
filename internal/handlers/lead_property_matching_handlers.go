package handlers

import (
	"fmt"
	"net/http"

	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LeadPropertyMatchingHandler struct {
	db              *gorm.DB
	matchingService *services.PropertyMatchingService
}

func NewLeadPropertyMatchingHandler(db *gorm.DB) *LeadPropertyMatchingHandler {
	return &LeadPropertyMatchingHandler{
		db:              db,
		matchingService: services.NewPropertyMatchingService(db),
	}
}

func (h *LeadPropertyMatchingHandler) GetMatchedPropertiesForLead(c *gin.Context) {
	leadID := c.Param("id")
	leadIDInt := parseInt64(leadID)

	matches, err := h.matchingService.FindMatchesForLead(leadIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"count":   len(matches),
	})
}

func parseInt64(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}
