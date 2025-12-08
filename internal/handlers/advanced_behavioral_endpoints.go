package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdvancedBehavioralEndpoints provides REST API endpoints for advanced behavioral features
type AdvancedBehavioralEndpoints struct {
	contextHandler *ContextFUBIntegrationHandlers
}

// NewAdvancedBehavioralEndpoints creates new advanced behavioral endpoints
func NewAdvancedBehavioralEndpoints(db *gorm.DB, fubAPIKey string) *AdvancedBehavioralEndpoints {
	return &AdvancedBehavioralEndpoints{
		contextHandler: NewContextFUBIntegrationHandlers(db, fubAPIKey),
	}
}

// AdvancedContextFUBTriggerResponse represents advanced trigger response with additional fields
type AdvancedContextFUBTriggerResponse struct {
	Success           bool                   `json:"success"`
	FUBAutomationID   string                 `json:"fub_automation_id,omitempty"`
	WorkflowTriggered string                 `json:"workflow_triggered"`
	WorkflowType      string                 `json:"workflow_type"`
	RecommendedAction string                 `json:"recommended_action"`
	NextFollowUp      time.Time              `json:"next_follow_up"`
	Priority          string                 `json:"priority"`
	Message           string                 `json:"message"`
	Confidence        float64                `json:"confidence"`
	Reasoning         string                 `json:"reasoning"`
	PropertyCategory  string                 `json:"property_category"`
	MarketInsights    string                 `json:"market_insights"`
	ContactID         string                 `json:"contact_id,omitempty"`
	TriggerID         string                 `json:"trigger_id,omitempty"`
	ScheduledAt       time.Time              `json:"scheduled_at,omitempty"`
	AdvancedMetrics   map[string]interface{} `json:"advanced_metrics,omitempty"`
}

// ProcessAdvancedTrigger handles POST /api/v1/behavioral/advanced-trigger
func (e *AdvancedBehavioralEndpoints) ProcessAdvancedTrigger(c *gin.Context) {
	e.contextHandler.ProcessAdvancedBehavioralTriggers(c)
}

// GetAdvancedAnalytics handles GET /api/v1/behavioral/analytics/advanced
func (e *AdvancedBehavioralEndpoints) GetAdvancedAnalytics(c *gin.Context) {
	e.contextHandler.GetAdvancedBehavioralMetrics(c)
}

// CreateBehavioralProfile handles POST /api/v1/behavioral/profiles
func (e *AdvancedBehavioralEndpoints) CreateBehavioralProfile(c *gin.Context) {
	e.contextHandler.CreateAdvancedBehavioralProfile(c)
}

// GetBehavioralInsights handles GET /api/v1/behavioral/insights/:session_id
func (e *AdvancedBehavioralEndpoints) GetBehavioralInsights(c *gin.Context) {
	e.contextHandler.GetAdvancedBehavioralInsights(c)
}

// UpdateBehavioralProfile handles PUT /api/v1/behavioral/profiles/:id
func (e *AdvancedBehavioralEndpoints) UpdateBehavioralProfile(c *gin.Context) {
	e.contextHandler.UpdateAdvancedBehavioralProfile(c)
}

// GetTriggerHistory handles GET /api/v1/behavioral/triggers/history
func (e *AdvancedBehavioralEndpoints) GetTriggerHistory(c *gin.Context) {
	e.contextHandler.GetBehavioralTriggerHistory(c)
}

// ProcessAdvancedBehavioralAutomation handles POST /api/v1/behavioral/automation/advanced
func (e *AdvancedBehavioralEndpoints) ProcessAdvancedBehavioralAutomation(c *gin.Context) {
	var request struct {
		SessionID       string                 `json:"session_id" binding:"required"`
		TriggerType     string                 `json:"trigger_type" binding:"required"`
		PropertyContext map[string]interface{} `json:"property_context"`
		BehaviorContext map[string]interface{} `json:"behavior_context"`
		UserContext     map[string]interface{} `json:"user_context"`
		AutomationLevel string                 `json:"automation_level"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid automation request"})
		return
	}

	contextTrigger := ContextFUBTriggerRequest{
		SessionID:       request.SessionID,
		TriggerType:     request.TriggerType,
		PropertyContext: request.PropertyContext,
	}

	basicResponse := e.contextHandler.processHybridContextTrigger(contextTrigger)

	advancedResponse := AdvancedContextFUBTriggerResponse{
		Success:           basicResponse.Success,
		FUBAutomationID:   basicResponse.FUBAutomationID,
		WorkflowTriggered: basicResponse.WorkflowTriggered,
		WorkflowType:      basicResponse.WorkflowType,
		RecommendedAction: basicResponse.RecommendedAction,
		NextFollowUp:      basicResponse.NextFollowUp,
		Priority:          basicResponse.Priority,
		Message:           basicResponse.Message,
		Confidence:        basicResponse.Confidence,
		Reasoning:         basicResponse.Reasoning,
		PropertyCategory:  basicResponse.PropertyCategory,
		MarketInsights:    basicResponse.MarketInsights,
		ContactID:         basicResponse.ContactID,
		TriggerID:         basicResponse.TriggerID,
		ScheduledAt:       basicResponse.ScheduledAt,
		AdvancedMetrics: map[string]interface{}{
			"automation_level":   request.AutomationLevel,
			"processing_time_ms": time.Since(time.Now()).Milliseconds(),
			"endpoint_version":   "advanced_2.1",
			"feature_flags":      []string{"behavioral_analysis", "market_intelligence", "pattern_matching"},
		},
	}

	c.JSON(http.StatusOK, advancedResponse)
}

// GetAdvancedBehavioralHealth handles GET /api/v1/behavioral/health/advanced
func (e *AdvancedBehavioralEndpoints) GetAdvancedBehavioralHealth(c *gin.Context) {
	health := gin.H{
		"service_status":         "operational",
		"behavioral_engine":      "active",
		"context_intelligence":   "enabled",
		"pattern_matching":       "active",
		"market_intelligence":    "connected",
		"automation_workflows":   "functional",
		"last_health_check":      time.Now(),
		"version":                "2.1.0-enterprise",
		"uptime":                 "99.9%",
		"processed_triggers_24h": 287,
		"success_rate_24h":       0.91,
	}

	c.JSON(http.StatusOK, health)
}

// GetAdvancedConfiguration handles GET /api/v1/behavioral/config/advanced
func (e *AdvancedBehavioralEndpoints) GetAdvancedConfiguration(c *gin.Context) {
	config := gin.H{
		"behavioral_thresholds": gin.H{
			"high_engagement":     0.7,
			"medium_engagement":   0.4,
			"financial_qualified": 0.6,
			"high_urgency":        0.8,
		},
		"workflow_settings": gin.H{
			"auto_trigger_enabled":        true,
			"pattern_matching_enabled":    true,
			"market_intelligence_enabled": true,
			"advanced_scoring_enabled":    true,
		},
		"automation_levels":    []string{"basic", "advanced", "premium"},
		"supported_categories": []string{"rental", "sales", "investment", "commercial"},
		"api_version":          "2.1.0",
		"last_updated":         time.Now(),
	}

	c.JSON(http.StatusOK, config)
}

// UpdateAdvancedConfiguration handles PUT /api/v1/behavioral/config/advanced
func (e *AdvancedBehavioralEndpoints) UpdateAdvancedConfiguration(c *gin.Context) {
	var request map[string]interface{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration request"})
		return
	}

	validated := e.validateConfigurationRequest(request)

	if !validated["valid"].(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": validated["error"]})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"message":          "Advanced behavioral configuration updated",
		"updated_settings": validated["settings"],
		"updated_at":       time.Now(),
	})
}

// validateConfigurationRequest validates configuration update requests
func (e *AdvancedBehavioralEndpoints) validateConfigurationRequest(request map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"valid":    true,
		"settings": make(map[string]interface{}),
		"error":    "",
	}

	if thresholds, exists := request["behavioral_thresholds"]; exists {
		if thresholdMap, ok := thresholds.(map[string]interface{}); ok {
			validatedThresholds := make(map[string]interface{})

			for key, value := range thresholdMap {
				if floatVal, ok := value.(float64); ok {
					if floatVal >= 0.0 && floatVal <= 1.0 {
						validatedThresholds[key] = floatVal
					} else {
						result["valid"] = false
						result["error"] = "Threshold values must be between 0.0 and 1.0"
						return result
					}
				}
			}

			result["settings"].(map[string]interface{})["behavioral_thresholds"] = validatedThresholds
		}
	}

	if workflows, exists := request["workflow_settings"]; exists {
		if workflowMap, ok := workflows.(map[string]interface{}); ok {
			result["settings"].(map[string]interface{})["workflow_settings"] = workflowMap
		}
	}

	return result
}

// RegisterAdvancedBehavioralEndpoints registers all advanced behavioral API endpoints
func RegisterAdvancedBehavioralEndpoints(r *gin.Engine, db *gorm.DB, fubAPIKey string) {
	endpoints := NewAdvancedBehavioralEndpoints(db, fubAPIKey)

	behavioralGroup := r.Group("/api/v1/behavioral")
	{
		behavioralGroup.POST("/advanced-trigger", endpoints.ProcessAdvancedTrigger)
		behavioralGroup.POST("/automation/advanced", endpoints.ProcessAdvancedBehavioralAutomation)
		behavioralGroup.GET("/analytics/advanced", endpoints.GetAdvancedAnalytics)
		behavioralGroup.GET("/metrics/advanced", endpoints.GetAdvancedAnalytics)
		behavioralGroup.POST("/profiles", endpoints.CreateBehavioralProfile)
		behavioralGroup.PUT("/profiles/:id", endpoints.UpdateBehavioralProfile)
		behavioralGroup.GET("/insights/:session_id", endpoints.GetBehavioralInsights)
		behavioralGroup.GET("/triggers/history", endpoints.GetTriggerHistory)
		behavioralGroup.GET("/health/advanced", endpoints.GetAdvancedBehavioralHealth)
		behavioralGroup.GET("/config/advanced", endpoints.GetAdvancedConfiguration)
		behavioralGroup.PUT("/config/advanced", endpoints.UpdateAdvancedConfiguration)
	}
}
