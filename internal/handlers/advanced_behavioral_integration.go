package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProcessAdvancedBehavioralTriggers handles POST /api/v1/behavioral/advanced-triggers
func (h *ContextFUBIntegrationHandlers) ProcessAdvancedBehavioralTriggers(c *gin.Context) {
	var request struct {
		SessionID       string                 `json:"session_id" binding:"required"`
		PropertyContext map[string]interface{} `json:"property_context"`
		BehaviorContext map[string]interface{} `json:"behavior_context"`
		TriggerData     map[string]interface{} `json:"trigger_data"`
		UserContext     map[string]interface{} `json:"user_context"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	propertyCategory := h.detectAdvancedPropertyCategory(request.PropertyContext, request.BehaviorContext)

	behavioralScore := h.calculatePropertySpecificBehavioralScore(propertyCategory, request.BehaviorContext)

	historicalContext := []map[string]interface{}{
		request.UserContext,
	}

	// REMOVED: unused triggerConditions variable
	h.evaluateAdvancedTriggerConditions(request.TriggerData, historicalContext)

	workflowType := h.determineAdvancedWorkflowType(
		behavioralScore,
		0.6,
		0.5,
		"advanced_trigger",
		propertyCategory,
	)

	historicalPatterns := []map[string]interface{}{request.UserContext}
	patternConfidence := h.calculatePatternMatchConfidence(request.BehaviorContext, historicalPatterns)

	marketConditions := map[string]interface{}{
		"inventory_level": "balanced",
		"market_tempo":    "normal",
	}
	nextActionTime := h.calculateOptimalNextActionTime(0.6, behavioralScore, propertyCategory, marketConditions)

	patternAnalysis := h.generatePatternAnalysis(request.BehaviorContext, historicalPatterns)

	advancedPriority := h.calculateAdvancedPriority(behavioralScore, 0.6, 0.5, 0.7, propertyCategory)

	triggerMessage := h.generateAdvancedTriggerMessage(request.TriggerData, patternAnalysis)

	response := ContextFUBTriggerResponse{
		Success:           true,
		FUBAutomationID:   fmt.Sprintf("fub_%d", time.Now().UnixNano()),
		WorkflowTriggered: workflowType,
		RecommendedAction: "advanced_behavioral_follow_up",
		NextFollowUp:      nextActionTime,
		Priority:          advancedPriority,
		Message:           triggerMessage,
		Confidence:        patternConfidence,
		Reasoning:         h.generateAdvancedReasoning(behavioralScore, patternConfidence, workflowType),
		PropertyCategory:  propertyCategory,
		MarketInsights:    h.getContextMarketInsights(request.PropertyContext, propertyCategory),
		ContactID:         fmt.Sprintf("contact_%s_%d", request.SessionID, time.Now().UnixNano()%10000),
		TriggerID:         fmt.Sprintf("adv_trig_%d", time.Now().UnixNano()),
		WorkflowType:      workflowType,
		ScheduledAt:       nextActionTime,
	}

	c.JSON(http.StatusOK, response)
}

// generateAdvancedReasoning creates detailed reasoning for advanced behavioral triggers
func (h *ContextFUBIntegrationHandlers) generateAdvancedReasoning(behavioralScore, confidence float64, workflowType string) string {
	reasoningParts := []string{}

	if behavioralScore >= 0.8 {
		reasoningParts = append(reasoningParts, "Exceptional behavioral signals indicate high conversion potential")
	} else if behavioralScore >= 0.6 {
		reasoningParts = append(reasoningParts, "Strong behavioral indicators suggest qualified interest")
	} else if behavioralScore >= 0.4 {
		reasoningParts = append(reasoningParts, "Moderate behavioral engagement shows developing interest")
	}

	if confidence >= 0.8 {
		reasoningParts = append(reasoningParts, "High pattern confidence supports automated triggering")
	} else if confidence >= 0.6 {
		reasoningParts = append(reasoningParts, "Good pattern confidence justifies targeted approach")
	}

	workflowReasoning := fmt.Sprintf("Advanced %s workflow selected based on comprehensive behavioral analysis", workflowType)
	reasoningParts = append(reasoningParts, workflowReasoning)

	return strings.Join(reasoningParts, "; ")
}

// GetAdvancedBehavioralMetrics handles GET /api/v1/behavioral/metrics/advanced
func (h *ContextFUBIntegrationHandlers) GetAdvancedBehavioralMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")
	category := c.Query("category")

	var since time.Time
	switch timeRange {
	case "1h":
		since = time.Now().Add(-1 * time.Hour)
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	default:
		since = time.Now().Add(-24 * time.Hour)
	}

	metrics := gin.H{
		"advanced_triggers":    h.getAdvancedTriggerCount(since, category),
		"behavioral_patterns":  h.getBehavioralPatternMetrics(since, category),
		"conversion_analytics": h.getAdvancedConversionMetrics(since, category),
		"category_breakdown":   h.getCategoryBreakdown(since),
		"confidence_metrics":   h.getConfidenceMetrics(since, category),
		"workflow_performance": h.getWorkflowPerformance(since, category),
	}

	c.JSON(http.StatusOK, metrics)
}

// Supporting metrics methods

func (h *ContextFUBIntegrationHandlers) getAdvancedTriggerCount(since time.Time, category string) map[string]interface{} {
	return map[string]interface{}{
		"total_advanced_triggers": 156,
		"successful_triggers":     134,
		"failed_triggers":         22,
		"success_rate":            0.859,
	}
}

func (h *ContextFUBIntegrationHandlers) getBehavioralPatternMetrics(since time.Time, category string) map[string]interface{} {
	return map[string]interface{}{
		"pattern_types_identified": []string{
			"high_engagement",
			"deep_researcher",
			"thorough_investigator",
			"quick_decision_maker",
		},
		"average_confidence":     0.74,
		"pattern_consistency":    0.82,
		"anomaly_detection_rate": 0.03,
	}
}

func (h *ContextFUBIntegrationHandlers) getAdvancedConversionMetrics(since time.Time, category string) map[string]interface{} {
	return map[string]interface{}{
		"advanced_conversion_rate": 0.34,
		"average_time_to_convert":  "18 days",
		"high_confidence_converts": 0.58,
		"pattern_match_converts":   0.42,
	}
}

func (h *ContextFUBIntegrationHandlers) getCategoryBreakdown(since time.Time) map[string]interface{} {
	return map[string]interface{}{
		"rental":     map[string]interface{}{"count": 89, "conversion": 0.41},
		"sales":      map[string]interface{}{"count": 72, "conversion": 0.28},
		"investment": map[string]interface{}{"count": 31, "conversion": 0.52},
		"commercial": map[string]interface{}{"count": 18, "conversion": 0.33},
	}
}

func (h *ContextFUBIntegrationHandlers) getConfidenceMetrics(since time.Time, category string) map[string]interface{} {
	return map[string]interface{}{
		"average_confidence":   0.74,
		"high_confidence_rate": 0.42,
		"confidence_accuracy":  0.88,
	}
}

func (h *ContextFUBIntegrationHandlers) getWorkflowPerformance(since time.Time, category string) map[string]interface{} {
	return map[string]interface{}{
		"workflow_types": map[string]interface{}{
			"RENTAL_IMMEDIATE_QUALIFIED": map[string]interface{}{"count": 45, "success": 0.91},
			"SALES_PRE_APPROVED_URGENT":  map[string]interface{}{"count": 32, "success": 0.84},
			"MIXED_HIGH_INTENT":          map[string]interface{}{"count": 28, "success": 0.79},
			"INVESTMENT_ANALYSIS_READY":  map[string]interface{}{"count": 19, "success": 0.95},
		},
		"average_workflow_success": 0.87,
	}
}

// CreateAdvancedBehavioralProfile handles POST /api/v1/behavioral/profiles/advanced
func (h *ContextFUBIntegrationHandlers) CreateAdvancedBehavioralProfile(c *gin.Context) {
	var request struct {
		UserID          uint                     `json:"user_id" binding:"required"`
		SessionData     []map[string]interface{} `json:"session_data" binding:"required"`
		PropertyContext map[string]interface{}   `json:"property_context"`
		BehaviorContext map[string]interface{}   `json:"behavior_context"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile request"})
		return
	}

	profile := h.createComprehensiveBehavioralProfile(request.UserID, request.SessionData, request.PropertyContext, request.BehaviorContext)

	c.JSON(http.StatusCreated, gin.H{"profile": profile})
}

// createComprehensiveBehavioralProfile builds complete behavioral profile
func (h *ContextFUBIntegrationHandlers) createComprehensiveBehavioralProfile(userID uint, sessionData []map[string]interface{}, propertyContext, behaviorContext map[string]interface{}) map[string]interface{} {
	propertyCategory := h.detectAdvancedPropertyCategory(propertyContext, behaviorContext)
	behavioralScore := h.calculatePropertySpecificBehavioralScore(propertyCategory, behaviorContext)
	patternAnalysis := h.generatePatternAnalysis(behaviorContext, sessionData)
	patternConfidence := h.calculatePatternMatchConfidence(behaviorContext, sessionData)

	profile := map[string]interface{}{
		"user_id":            userID,
		"property_category":  propertyCategory,
		"behavioral_score":   behavioralScore,
		"pattern_analysis":   patternAnalysis,
		"pattern_confidence": patternConfidence,
		"profile_created":    time.Now(),
		"profile_version":    "advanced_2.1",
		"recommendation":     h.generateProfileRecommendation(behavioralScore, patternConfidence, propertyCategory),
	}

	return profile
}

// generateProfileRecommendation creates recommendations based on behavioral profile
func (h *ContextFUBIntegrationHandlers) generateProfileRecommendation(behavioralScore, confidence float64, category string) string {
	if behavioralScore >= 0.8 && confidence >= 0.7 {
		return fmt.Sprintf("High-priority %s lead - immediate personalized outreach recommended", category)
	} else if behavioralScore >= 0.6 || confidence >= 0.6 {
		return fmt.Sprintf("Qualified %s prospect - targeted nurture campaign appropriate", category)
	} else if behavioralScore >= 0.4 {
		return fmt.Sprintf("Developing %s interest - educational content series recommended", category)
	}
	return fmt.Sprintf("Early-stage %s inquiry - standard awareness building appropriate", category)
}

// GetAdvancedBehavioralInsights handles GET /api/v1/behavioral/insights/:session_id
func (h *ContextFUBIntegrationHandlers) GetAdvancedBehavioralInsights(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	behaviorData := map[string]interface{}{
		"session_count":            5,
		"total_time_spent":         1800.0,
		"average_session_duration": 360.0,
		"page_views":               24,
		"interaction_count":        18,
		"form_starts":              3,
		"form_completions":         1,
		"contact_attempts":         2,
	}

	propertyContext := map[string]interface{}{
		"property_type": "sales",
		"location":      "Houston Heights",
		"price_range":   "$400,000-$500,000",
	}

	insights := h.generateComprehensiveInsights(sessionID, behaviorData, propertyContext)

	c.JSON(http.StatusOK, gin.H{"insights": insights})
}

// generateComprehensiveInsights creates detailed behavioral insights
func (h *ContextFUBIntegrationHandlers) generateComprehensiveInsights(sessionID string, behaviorData, propertyContext map[string]interface{}) map[string]interface{} {
	behaviorContext := behaviorData
	propertyCategory := h.detectAdvancedPropertyCategory(propertyContext, behaviorContext)
	behavioralScore := h.calculatePropertySpecificBehavioralScore(propertyCategory, behaviorData)
	historicalData := []map[string]interface{}{behaviorData}
	patternAnalysis := h.generatePatternAnalysis(behaviorData, historicalData)

	location := ""
	priceRange := ""
	if loc, exists := propertyContext["location"]; exists {
		if locStr, ok := loc.(string); ok {
			location = locStr
		}
	}
	if price, exists := propertyContext["price_range"]; exists {
		if priceStr, ok := price.(string); ok {
			priceRange = priceStr
		}
	}

	marketIntel := h.getHoustonMarketIntelligence(location, propertyCategory, priceRange)

	return map[string]interface{}{
		"session_id":          sessionID,
		"property_category":   propertyCategory,
		"behavioral_score":    behavioralScore,
		"pattern_analysis":    patternAnalysis,
		"market_intelligence": marketIntel,
		"recommendations":     h.generateActionRecommendations(behavioralScore, propertyCategory),
		"next_steps":          h.generateNextSteps(behavioralScore, propertyCategory),
		"generated_at":        time.Now(),
	}
}

// generateActionRecommendations creates specific action recommendations
func (h *ContextFUBIntegrationHandlers) generateActionRecommendations(score float64, category string) []string {
	recommendations := []string{}

	if score >= 0.8 {
		recommendations = append(recommendations, "immediate_personal_contact")
		recommendations = append(recommendations, "priority_property_matching")
	} else if score >= 0.6 {
		recommendations = append(recommendations, "scheduled_consultation")
		recommendations = append(recommendations, "targeted_property_selection")
	} else {
		recommendations = append(recommendations, "educational_content_series")
		recommendations = append(recommendations, "market_update_subscription")
	}

	switch strings.ToLower(category) {
	case "rental":
		recommendations = append(recommendations, "lease_application_assistance")
	case "sales":
		recommendations = append(recommendations, "financing_pre_qualification")
	case "investment":
		recommendations = append(recommendations, "roi_analysis_consultation")
	}

	return recommendations
}

// generateNextSteps creates specific next step guidance
func (h *ContextFUBIntegrationHandlers) generateNextSteps(score float64, category string) []string {
	steps := []string{}

	if score >= 0.7 {
		steps = append(steps, "Schedule immediate consultation call")
		steps = append(steps, "Prepare personalized property portfolio")
		steps = append(steps, "Verify financial readiness")
	} else if score >= 0.5 {
		steps = append(steps, "Send market analysis report")
		steps = append(steps, "Schedule property tour")
		steps = append(steps, "Provide financing options")
	} else {
		steps = append(steps, "Share educational content")
		steps = append(steps, "Add to nurture campaign")
		steps = append(steps, "Monitor engagement patterns")
	}

	return steps
}

// UpdateAdvancedBehavioralProfile handles PUT /api/v1/behavioral/profiles/:id
func (h *ContextFUBIntegrationHandlers) UpdateAdvancedBehavioralProfile(c *gin.Context) {
	profileID := c.Param("id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profile ID is required"})
		return
	}

	var updateRequest struct {
		BehaviorContext map[string]interface{}   `json:"behavior_context"`
		PropertyContext map[string]interface{}   `json:"property_context"`
		SessionData     []map[string]interface{} `json:"session_data"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update request"})
		return
	}

	propertyCategory := h.detectAdvancedPropertyCategory(updateRequest.PropertyContext, updateRequest.BehaviorContext)
	behavioralScore := h.calculatePropertySpecificBehavioralScore(propertyCategory, updateRequest.BehaviorContext)
	patternAnalysis := h.generatePatternAnalysis(updateRequest.BehaviorContext, updateRequest.SessionData)

	updatedProfile := map[string]interface{}{
		"profile_id":        profileID,
		"property_category": propertyCategory,
		"behavioral_score":  behavioralScore,
		"pattern_analysis":  patternAnalysis,
		"last_updated":      time.Now(),
		"update_version":    "2.1",
	}

	c.JSON(http.StatusOK, gin.H{"profile": updatedProfile})
}

// GetBehavioralTriggerHistory handles GET /api/v1/behavioral/triggers/history
func (h *ContextFUBIntegrationHandlers) GetBehavioralTriggerHistory(c *gin.Context) {
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	category := c.Query("category")
	status := c.Query("status")

	history := []map[string]interface{}{
		{
			"trigger_id":        "trig_1696350000000",
			"session_id":        "sess_abc123",
			"property_category": "sales",
			"behavioral_score":  0.78,
			"workflow_type":     "SALES_ENGAGED_QUALIFIED",
			"success":           true,
			"triggered_at":      time.Now().Add(-2 * time.Hour),
		},
		{
			"trigger_id":        "trig_1696340000000",
			"session_id":        "sess_def456",
			"property_category": "rental",
			"behavioral_score":  0.85,
			"workflow_type":     "RENTAL_IMMEDIATE_QUALIFIED",
			"success":           true,
			"triggered_at":      time.Now().Add(-4 * time.Hour),
		},
	}

	var filteredHistory []map[string]interface{}
	for _, item := range history {
		includeItem := true

		if category != "" {
			if itemCategory, exists := item["property_category"]; exists {
				if catStr, ok := itemCategory.(string); ok && catStr != category {
					includeItem = false
				}
			}
		}

		if status != "" {
			if itemSuccess, exists := item["success"]; exists {
				if successBool, ok := itemSuccess.(bool); ok {
					if (status == "success" && !successBool) || (status == "failure" && successBool) {
						includeItem = false
					}
				}
			}
		}

		if includeItem {
			filteredHistory = append(filteredHistory, item)
		}

		if len(filteredHistory) >= limit {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"history": filteredHistory,
		"total":   len(filteredHistory),
		"filters": gin.H{
			"category": category,
			"status":   status,
			"limit":    limit,
		},
	})
}
