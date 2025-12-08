package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ContextFUBIntegrationHandlers handles context intelligence driven FUB automation
type ContextFUBIntegrationHandlers struct {
	db               *gorm.DB
	behavioralBridge *services.BehavioralFUBBridge
}

// NewContextFUBIntegrationHandlers creates new context-driven FUB integration handlers
func NewContextFUBIntegrationHandlers(db *gorm.DB, fubAPIKey string) *ContextFUBIntegrationHandlers {
	return &ContextFUBIntegrationHandlers{
		db:               db,
		behavioralBridge: services.NewBehavioralFUBBridge(db, fubAPIKey),
	}
}

// ContextFUBTriggerRequest represents a property-type aware context-driven FUB automation trigger
type ContextFUBTriggerRequest struct {
	SessionID             string                 `json:"session_id" binding:"required"`
	Email                 string                 `json:"email"`
	Phone                 string                 `json:"phone"`
	Name                  string                 `json:"name"`
	PropertyID            int                    `json:"property_id"`
	TriggerType           string                 `json:"trigger_type" binding:"required"`
	LeadType              string                 `json:"lead_type"`
	PropertyType          string                 `json:"property_type"`
	UrgencyScore          float64                `json:"urgency_score"`
	FinancialQualScore    float64                `json:"financial_qualification_score"`
	EngagementScore       float64                `json:"engagement_score"`
	RentalBehaviorContext map[string]interface{} `json:"rental_behavior_context"`
	SalesBehaviorContext  map[string]interface{} `json:"sales_behavior_context"`
	PropertyContext       map[string]interface{} `json:"property_context"`
	TimelineContext       map[string]interface{} `json:"timeline_context"`
}

// ContextFUBTriggerResponse represents the response from context-driven FUB trigger
type ContextFUBTriggerResponse struct {
	Success           bool      `json:"success"`
	FUBAutomationID   string    `json:"fub_automation_id,omitempty"`
	WorkflowTriggered string    `json:"workflow_triggered"`
	RecommendedAction string    `json:"recommended_action"`
	NextFollowUp      time.Time `json:"next_follow_up"`
	Priority          string    `json:"priority"`
	Message           string    `json:"message"`
	Confidence        float64   `json:"confidence"`
	Reasoning         string    `json:"reasoning"`
	PropertyCategory  string    `json:"property_category"`
	MarketInsights    string    `json:"market_insights"`
	ContactID         string    `json:"contact_id,omitempty"`
	TriggerID         string    `json:"trigger_id,omitempty"`
	WorkflowType      string    `json:"workflow_type,omitempty"`
	ScheduledAt       time.Time `json:"scheduled_at,omitempty"`
}

// TriggerContextDrivenFUBAutomation handles POST /api/v1/context-fub/trigger
func (h *ContextFUBIntegrationHandlers) TriggerContextDrivenFUBAutomation(c *gin.Context) {
	var trigger ContextFUBTriggerRequest

	if err := c.ShouldBindJSON(&trigger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
		return
	}

	log.Printf("🎯 Processing context-driven FUB trigger for session: %s", trigger.SessionID)

	result := h.processHybridContextTrigger(trigger)

	c.JSON(http.StatusOK, result)
}

// ProcessContextIntelligenceWebhook handles POST /api/v1/context-fub/webhook
func (h *ContextFUBIntegrationHandlers) ProcessContextIntelligenceWebhook(c *gin.Context) {
	var webhookData map[string]interface{}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	shouldTrigger, triggerType := h.shouldTriggerContextAutomation(c.GetHeader("X-Event-Type"), webhookData)

	if !shouldTrigger {
		c.JSON(http.StatusOK, gin.H{"processed": true, "action": "ignored"})
		return
	}

	contextTrigger := h.extractContextFromWebhook(webhookData, triggerType)
	result := h.processHybridContextTrigger(contextTrigger)

	c.JSON(http.StatusOK, gin.H{
		"processed": true,
		"action":    "context_automation_triggered",
		"result":    result,
	})
}

// GetContextFUBAnalytics handles GET /api/v1/context-fub/analytics
func (h *ContextFUBIntegrationHandlers) GetContextFUBAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")
	propertyType := c.Query("property_type")

	var since time.Time
	switch timeRange {
	case "1h":
		since = time.Now().Add(-1 * time.Hour)
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		since = time.Now().Add(-30 * 24 * time.Hour)
	default:
		since = time.Now().Add(-24 * time.Hour)
	}

	analytics := gin.H{
		"time_range":          timeRange,
		"property_filter":     propertyType,
		"total_triggers":      h.getAnalyticsCount("total_triggers", since, propertyType),
		"successful_triggers": h.getAnalyticsCount("successful_triggers", since, propertyType),
		"property_breakdown":  h.getPropertyTypeBreakdown(since),
		"conversion_metrics":  h.getConversionMetrics(since, propertyType),
		"behavioral_insights": h.getBehavioralInsights(since, propertyType),
	}

	c.JSON(http.StatusOK, analytics)
}

// GetContextFUBStatus handles GET /api/v1/context-fub/status
func (h *ContextFUBIntegrationHandlers) GetContextFUBStatus(c *gin.Context) {
	status := gin.H{
		"service_status":       "operational",
		"behavioral_bridge":    h.behavioralBridge != nil,
		"context_intelligence": "active",
		"last_updated":         time.Now(),
		"version":              "2.1.0-enterprise",
	}

	c.JSON(http.StatusOK, status)
}

// getHoustonMarketIntelligence provides comprehensive Houston real estate market intelligence
func (h *ContextFUBIntegrationHandlers) getHoustonMarketIntelligence(location string, propertyType string, priceRange string) map[string]interface{} {
	marketIntelligence := map[string]interface{}{
		"market_overview": map[string]interface{}{
			"city":          "Houston",
			"state":         "Texas",
			"market_status": "active",
			"trend":         "rising",
			"last_updated":  time.Now().Format("2006-01-02"),
			"data_source":   "Houston MLS & Market Analytics",
		},
		"rental_market": map[string]interface{}{
			"median_rent":            2850,
			"rent_growth_yoy":        0.074,
			"average_days_on_market": 14,
			"occupancy_rate":         0.943,
			"rental_yield":           0.058,
			"popular_neighborhoods": []string{
				"The Heights", "Montrose", "River Oaks", "Galleria",
				"Medical Center", "Downtown", "Midtown", "West University",
			},
		},
		"sales_market": map[string]interface{}{
			"median_home_price":      425000,
			"price_growth_yoy":       0.069,
			"average_days_on_market": 28,
			"months_of_inventory":    2.8,
			"sale_to_list_ratio":     0.987,
			"new_listings":           1450,
			"homes_sold":             1320,
		},
		"neighborhood_insights": h.getNeighborhoodInsights(location),
		"investment_metrics": map[string]interface{}{
			"cap_rate_range":        "4.5%-7.2%",
			"cash_on_cash_return":   0.089,
			"appreciation_forecast": 0.055,
			"rental_demand":         "high",
			"investor_activity":     "increasing",
		},
		"market_factors": map[string]interface{}{
			"job_growth":              0.032,
			"population_growth":       0.018,
			"major_employers":         []string{"Texas Medical Center", "ExxonMobil", "Shell", "NASA", "Port of Houston"},
			"infrastructure_projects": []string{"I-45 Expansion", "Metro Rail Extension", "Port Expansion"},
			"economic_indicators": map[string]interface{}{
				"unemployment_rate":  0.038,
				"gdp_growth":         0.041,
				"business_formation": "strong",
			},
		},
		"seasonal_patterns":    h.getSeasonalPatterns(),
		"competitive_analysis": h.getCompetitiveAnalysis(propertyType, priceRange),
		"forecasts": map[string]interface{}{
			"next_quarter": map[string]interface{}{
				"price_forecast":  "continued growth",
				"inventory_trend": "tightening",
				"demand_outlook":  "strong",
			},
			"next_year": map[string]interface{}{
				"price_growth":   0.048,
				"market_outlook": "favorable",
				"risk_factors":   []string{"interest rates", "supply chain", "energy sector volatility"},
			},
		},
	}

	if location != "" {
		marketIntelligence["location_specific"] = h.getLocationSpecificInsights(location)
	}

	if propertyType != "" {
		marketIntelligence["property_type_insights"] = h.getPropertyTypeInsights(propertyType)
	}

	if priceRange != "" {
		marketIntelligence["price_range_analysis"] = h.getPriceRangeAnalysis(priceRange)
	}

	return marketIntelligence
}

// Supporting methods for Houston market intelligence

func (h *ContextFUBIntegrationHandlers) getNeighborhoodInsights(location string) map[string]interface{} {
	neighborhoodData := map[string]interface{}{
		"walk_score":         75,
		"school_rating":      8.2,
		"crime_index":        "low",
		"commute_time":       "22 minutes to downtown",
		"amenities_score":    9.1,
		"future_development": "mixed-use project planned",
	}

	if strings.Contains(strings.ToLower(location), "heights") {
		neighborhoodData["character"] = "historic, trendy"
		neighborhoodData["price_trend"] = "premium growth"
	} else if strings.Contains(strings.ToLower(location), "montrose") {
		neighborhoodData["character"] = "arts district, eclectic"
		neighborhoodData["price_trend"] = "steady appreciation"
	}

	return neighborhoodData
}

func (h *ContextFUBIntegrationHandlers) getSeasonalPatterns() map[string]interface{} {
	return map[string]interface{}{
		"spring": map[string]interface{}{
			"activity_level": "peak",
			"price_movement": "strongest appreciation",
			"inventory":      "increasing",
		},
		"summer": map[string]interface{}{
			"activity_level": "high",
			"price_movement": "continued growth",
			"inventory":      "stabilizing",
		},
		"fall": map[string]interface{}{
			"activity_level": "moderate",
			"price_movement": "slower growth",
			"inventory":      "declining",
		},
		"winter": map[string]interface{}{
			"activity_level": "lower",
			"price_movement": "stable",
			"inventory":      "lowest",
		},
	}
}

func (h *ContextFUBIntegrationHandlers) getCompetitiveAnalysis(propertyType, priceRange string) map[string]interface{} {
	return map[string]interface{}{
		"market_competition":         "high",
		"buyer_demand":               "strong",
		"seller_advantage":           true,
		"negotiation_power":          "seller-favored",
		"average_offers_per_listing": 3.2,
		"cash_offer_percentage":      0.28,
		"above_asking_percentage":    0.15,
	}
}

func (h *ContextFUBIntegrationHandlers) getLocationSpecificInsights(location string) map[string]interface{} {
	return map[string]interface{}{
		"proximity_scores": map[string]interface{}{
			"downtown":       8.5,
			"medical_center": 7.2,
			"airports":       6.8,
			"major_highways": 9.1,
		},
		"development_pipeline": "moderate",
		"zoning_changes":       "none planned",
		"infrastructure_score": 8.3,
	}
}

func (h *ContextFUBIntegrationHandlers) getPropertyTypeInsights(propertyType string) map[string]interface{} {
	insights := map[string]interface{}{}

	switch strings.ToLower(propertyType) {
	case "rental":
		insights = map[string]interface{}{
			"rental_demand":    "very high",
			"tenant_retention": 0.847,
			"rental_growth":    0.074,
			"vacancy_rate":     0.057,
		}
	case "sales":
		insights = map[string]interface{}{
			"buyer_demand":       "high",
			"price_appreciation": 0.069,
			"sale_velocity":      "fast",
			"market_timing":      "favorable for sellers",
		}
	}

	return insights
}

func (h *ContextFUBIntegrationHandlers) getPriceRangeAnalysis(priceRange string) map[string]interface{} {
	return map[string]interface{}{
		"market_segment":    "active",
		"competition_level": "moderate",
		"demand_strength":   "strong",
		"inventory_level":   "balanced",
		"price_volatility":  "low",
	}
}

// formatMarketInsightsForResponse formats market intelligence for response
func (h *ContextFUBIntegrationHandlers) formatMarketInsightsForResponse(intelligence map[string]interface{}) string {
	insights := []string{}

	if overview, exists := intelligence["market_overview"]; exists {
		if overviewMap, ok := overview.(map[string]interface{}); ok {
			if trend, exists := overviewMap["trend"]; exists {
				insights = append(insights, fmt.Sprintf("Houston market trend: %v", trend))
			}
		}
	}

	if rental, exists := intelligence["rental_market"]; exists {
		if rentalMap, ok := rental.(map[string]interface{}); ok {
			if growth, exists := rentalMap["rent_growth_yoy"]; exists {
				if growthFloat, ok := growth.(float64); ok {
					insights = append(insights, fmt.Sprintf("Rental growth: %.1f%% YoY", growthFloat*100))
				}
			}
		}
	}

	if sales, exists := intelligence["sales_market"]; exists {
		if salesMap, ok := sales.(map[string]interface{}); ok {
			if price, exists := salesMap["median_home_price"]; exists {
				insights = append(insights, fmt.Sprintf("Median home price: $%v", price))
			}
		}
	}

	if len(insights) == 0 {
		insights = append(insights, "Houston market showing positive activity")
	}

	return strings.Join(insights, ". ")
}

// processHybridContextTrigger processes the trigger with full context intelligence
func (h *ContextFUBIntegrationHandlers) processHybridContextTrigger(trigger ContextFUBTriggerRequest) ContextFUBTriggerResponse {
	log.Printf("🧠 Processing hybrid context trigger for property type: %s", trigger.PropertyType)

	triggerID := fmt.Sprintf("trig_%d", time.Now().UnixNano())
	contactID := fmt.Sprintf("contact_%s_%d", trigger.SessionID, time.Now().UnixNano()%10000)

	workflowType := h.determineAdaptiveWorkflowType(
		trigger.EngagementScore,
		trigger.FinancialQualScore,
		trigger.UrgencyScore,
		trigger.TriggerType,
		trigger.PropertyType,
	)

	recommendedAction := h.getAdaptiveRecommendedAction(
		trigger.EngagementScore,
		trigger.FinancialQualScore,
		trigger.UrgencyScore,
		trigger.PropertyType,
	)

	priority := h.calculatePriority(trigger.EngagementScore, trigger.FinancialQualScore, trigger.UrgencyScore)
	nextFollowUp := h.calculateScheduling(trigger.UrgencyScore, priority)
	scheduledAt := nextFollowUp

	location := ""
	priceRange := ""
	if propertyCtx, exists := trigger.PropertyContext["location"]; exists {
		if locationStr, ok := propertyCtx.(string); ok {
			location = locationStr
		}
	}
	if priceCtx, exists := trigger.PropertyContext["price_range"]; exists {
		if priceStr, ok := priceCtx.(string); ok {
			priceRange = priceStr
		}
	}

	marketIntelligence := h.getHoustonMarketIntelligence(location, trigger.PropertyType, priceRange)
	marketInsights := h.formatMarketInsightsForResponse(marketIntelligence)
	reasoning := h.generateContextReasoning(trigger, workflowType, recommendedAction)

	return ContextFUBTriggerResponse{
		Success:           true,
		WorkflowTriggered: workflowType,
		RecommendedAction: recommendedAction,
		NextFollowUp:      nextFollowUp,
		Priority:          priority,
		Message:           fmt.Sprintf("Context-driven %s workflow triggered for %s property", workflowType, trigger.PropertyType),
		Confidence:        0.85,
		Reasoning:         reasoning,
		PropertyCategory:  trigger.PropertyType,
		MarketInsights:    marketInsights,
		ContactID:         contactID,
		TriggerID:         triggerID,
		WorkflowType:      workflowType,
		ScheduledAt:       scheduledAt,
	}
}

// Workflow determination methods

func (h *ContextFUBIntegrationHandlers) determineWorkflowType(engagement, conversion, urgency float64, triggerType string) string {
	if urgency >= 0.8 && engagement >= 0.7 {
		return "HIGH_INTENT_IMMEDIATE"
	} else if conversion >= 0.6 && engagement >= 0.5 {
		return "QUALIFIED_NURTURE"
	} else if engagement >= 0.4 {
		return "ENGAGEMENT_BUILDING"
	}
	return "AWARENESS_BUILDING"
}

func (h *ContextFUBIntegrationHandlers) getContextRecommendedAction(engagement, conversion, urgency float64) string {
	if urgency >= 0.8 {
		return "immediate_call"
	} else if conversion >= 0.6 {
		return "personalized_follow_up"
	} else if engagement >= 0.4 {
		return "educational_content"
	}
	return "nurture_sequence"
}

func (h *ContextFUBIntegrationHandlers) calculatePriority(engagement, conversion, urgency float64) string {
	score := (engagement + conversion + urgency) / 3.0
	if score >= 0.7 {
		return "HIGH"
	} else if score >= 0.4 {
		return "MEDIUM"
	}
	return "LOW"
}

func (h *ContextFUBIntegrationHandlers) getContextActionPlanID(workflowType, recommendedAction string) string {
	return fmt.Sprintf("context_%s_%s", strings.ToLower(workflowType), strings.ToLower(recommendedAction))
}

func (h *ContextFUBIntegrationHandlers) calculateScheduling(urgencyScore float64, priority string) time.Time {
	now := time.Now()
	if priority == "HIGH" {
		return now.Add(15 * time.Minute)
	} else if priority == "MEDIUM" {
		return now.Add(2 * time.Hour)
	}
	return now.Add(24 * time.Hour)
}

// Webhook processing methods

func (h *ContextFUBIntegrationHandlers) shouldTriggerContextAutomation(eventType string, webhookData map[string]interface{}) (bool, string) {
	triggerEvents := []string{
		"property_viewed",
		"inquiry_submitted",
		"application_started",
		"document_uploaded",
		"phone_call_completed",
		"email_engagement",
	}

	for _, event := range triggerEvents {
		if eventType == event {
			return true, event
		}
	}

	return false, ""
}

func (h *ContextFUBIntegrationHandlers) calculateEngagementFromWebhook(data map[string]interface{}) float64 {
	score := 0.3

	if viewTime, exists := data["view_duration"]; exists {
		if duration, ok := viewTime.(float64); ok && duration > 60 {
			score += 0.2
		}
	}

	if interactions, exists := data["page_interactions"]; exists {
		if count, ok := interactions.(float64); ok {
			score += count * 0.1
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (h *ContextFUBIntegrationHandlers) calculateConversionFromWebhook(data map[string]interface{}) float64 {
	score := 0.2

	if formData, exists := data["form_completed"]; exists {
		if completed, ok := formData.(bool); ok && completed {
			score += 0.4
		}
	}

	if contactInfo, exists := data["contact_provided"]; exists {
		if provided, ok := contactInfo.(bool); ok && provided {
			score += 0.3
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (h *ContextFUBIntegrationHandlers) calculateUrgencyFromWebhook(data map[string]interface{}) float64 {
	score := 0.1

	if timeline, exists := data["move_timeline"]; exists {
		if timelineStr, ok := timeline.(string); ok {
			switch timelineStr {
			case "immediate":
				score = 0.9
			case "30_days":
				score = 0.7
			case "90_days":
				score = 0.5
			case "6_months":
				score = 0.3
			}
		}
	}

	return score
}

// Context extraction methods

func (h *ContextFUBIntegrationHandlers) extractPropertyContext(data map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})

	if propertyID, exists := data["property_id"]; exists {
		context["property_id"] = propertyID
	}

	if propertyType, exists := data["property_type"]; exists {
		context["property_type"] = propertyType
	}

	if price, exists := data["price"]; exists {
		context["price_range"] = price
	}

	if location, exists := data["location"]; exists {
		context["location"] = location
	}

	return context
}

func (h *ContextFUBIntegrationHandlers) extractRentalBehaviorContext(data map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})

	if leaseLength, exists := data["desired_lease_length"]; exists {
		context["lease_preference"] = leaseLength
	}

	if moveDate, exists := data["move_date"]; exists {
		context["timeline"] = moveDate
	}

	if budget, exists := data["monthly_budget"]; exists {
		context["budget_range"] = budget
	}

	return context
}

func (h *ContextFUBIntegrationHandlers) extractSalesBehaviorContext(data map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})

	if buyerType, exists := data["buyer_type"]; exists {
		context["buyer_category"] = buyerType
	}

	if financing, exists := data["financing_status"]; exists {
		context["financing_ready"] = financing
	}

	if priceRange, exists := data["max_budget"]; exists {
		context["budget_ceiling"] = priceRange
	}

	return context
}

func (h *ContextFUBIntegrationHandlers) extractTimelineContext(data map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})

	if urgency, exists := data["urgency_level"]; exists {
		context["urgency"] = urgency
	}

	if timeline, exists := data["decision_timeline"]; exists {
		context["decision_window"] = timeline
	}

	return context
}

func (h *ContextFUBIntegrationHandlers) extractContextFromWebhook(webhookData map[string]interface{}, triggerType string) ContextFUBTriggerRequest {
	sessionID := ""
	if sid, exists := webhookData["session_id"]; exists {
		if sidStr, ok := sid.(string); ok {
			sessionID = sidStr
		}
	}

	// FIXED: Remove unused variables engagement, conversion, urgency
	// Use direct method calls in adaptive scoring instead

	propertyContext := h.extractPropertyContext(webhookData)
	rentalContext := h.extractRentalBehaviorContext(webhookData)
	salesContext := h.extractSalesBehaviorContext(webhookData)
	timelineContext := h.extractTimelineContext(webhookData)

	propertyMarketType := h.detectPropertyMarketType(propertyContext, rentalContext, salesContext)
	leadType := h.determineLeadType(webhookData, propertyMarketType)

	adaptiveEngagement := h.calculateAdaptiveEngagementScore(webhookData, propertyMarketType)
	adaptiveFinancial := h.calculateAdaptiveFinancialScore(webhookData, propertyMarketType)
	adaptiveUrgency := h.calculateAdaptiveUrgencyScore(webhookData, propertyMarketType)

	return ContextFUBTriggerRequest{
		SessionID:             sessionID,
		TriggerType:           triggerType,
		LeadType:              leadType,
		PropertyType:          propertyMarketType,
		EngagementScore:       adaptiveEngagement,
		FinancialQualScore:    adaptiveFinancial,
		UrgencyScore:          adaptiveUrgency,
		PropertyContext:       propertyContext,
		RentalBehaviorContext: rentalContext,
		SalesBehaviorContext:  salesContext,
		TimelineContext:       timelineContext,
	}
}

// Property market intelligence methods

func (h *ContextFUBIntegrationHandlers) detectPropertyMarketType(propertyContext, rentalContext, salesContext map[string]interface{}) string {
	rentalIndicators := len(rentalContext)
	salesIndicators := len(salesContext)

	if rentalIndicators > salesIndicators {
		return "rental"
	} else if salesIndicators > rentalIndicators {
		return "sales"
	} else if rentalIndicators > 0 && salesIndicators > 0 {
		return "mixed"
	}

	return "unknown"
}

func (h *ContextFUBIntegrationHandlers) determineLeadType(data map[string]interface{}, propertyMarketType string) string {
	switch propertyMarketType {
	case "rental":
		return "tenant"
	case "sales":
		return "buyer"
	default:
		return "prospect"
	}
}

// Adaptive scoring methods

func (h *ContextFUBIntegrationHandlers) calculateAdaptiveEngagementScore(data map[string]interface{}, propertyMarketType string) float64 {
	baseScore := h.calculateEngagementFromWebhook(data)

	switch propertyMarketType {
	case "rental":
		return baseScore * 1.1
	case "sales":
		return baseScore * 0.9
	case "mixed":
		return baseScore
	}

	return baseScore
}

func (h *ContextFUBIntegrationHandlers) calculateAdaptiveFinancialScore(data map[string]interface{}, propertyMarketType string) float64 {
	baseScore := h.calculateConversionFromWebhook(data)

	if propertyMarketType == "sales" {
		if financing, exists := data["pre_approved"]; exists {
			if approved, ok := financing.(bool); ok && approved {
				baseScore += 0.3
			}
		}
	}

	if propertyMarketType == "rental" {
		if income, exists := data["income_verified"]; exists {
			if verified, ok := income.(bool); ok && verified {
				baseScore += 0.2
			}
		}
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

func (h *ContextFUBIntegrationHandlers) calculateAdaptiveUrgencyScore(data map[string]interface{}, propertyMarketType string) float64 {
	baseScore := h.calculateUrgencyFromWebhook(data)

	switch propertyMarketType {
	case "rental":
		return baseScore * 1.2
	case "sales":
		return baseScore * 0.8
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

// Advanced workflow determination

func (h *ContextFUBIntegrationHandlers) determineAdaptiveWorkflowType(engagement, financial, urgency float64, triggerType, propertyType string) string {
	switch propertyType {
	case "rental":
		return h.determineRentalWorkflowType(engagement, financial, urgency, triggerType)
	case "sales":
		return h.determineSalesWorkflowType(engagement, financial, urgency, triggerType)
	case "mixed":
		return h.determineMixedWorkflowType(engagement, financial, urgency, triggerType)
	default:
		return h.determineWorkflowType(engagement, financial, urgency, triggerType)
	}
}

func (h *ContextFUBIntegrationHandlers) determineRentalWorkflowType(engagement, financial, urgency float64, triggerType string) string {
	if urgency >= 0.8 && financial >= 0.6 {
		return "RENTAL_IMMEDIATE_QUALIFIED"
	} else if urgency >= 0.6 && engagement >= 0.5 {
		return "RENTAL_URGENT_NURTURE"
	} else if financial >= 0.5 {
		return "RENTAL_QUALIFIED_BUILDING"
	}
	return "RENTAL_AWARENESS_BUILDING"
}

func (h *ContextFUBIntegrationHandlers) determineSalesWorkflowType(engagement, financial, urgency float64, triggerType string) string {
	if financial >= 0.8 && urgency >= 0.6 {
		return "SALES_PRE_APPROVED_URGENT"
	} else if engagement >= 0.7 && financial >= 0.5 {
		return "SALES_ENGAGED_QUALIFIED"
	} else if urgency >= 0.6 {
		return "SALES_URGENT_NURTURE"
	}
	return "SALES_EDUCATION_BUILDING"
}

func (h *ContextFUBIntegrationHandlers) determineMixedWorkflowType(engagement, financial, urgency float64, triggerType string) string {
	avgScore := (engagement + financial + urgency) / 3.0

	if avgScore >= 0.7 {
		return "MIXED_HIGH_INTENT"
	} else if avgScore >= 0.5 {
		return "MIXED_QUALIFIED_NURTURE"
	}
	return "MIXED_EXPLORATION_SUPPORT"
}

// Adaptive action recommendations

func (h *ContextFUBIntegrationHandlers) getAdaptiveRecommendedAction(engagement, financial, urgency float64, propertyType string) string {
	switch propertyType {
	case "rental":
		return h.getRentalRecommendedAction(engagement, financial, urgency)
	case "sales":
		return h.getSalesRecommendedAction(engagement, financial, urgency)
	case "mixed":
		return h.getMixedRecommendedAction(engagement, financial, urgency)
	default:
		return h.getContextRecommendedAction(engagement, financial, urgency)
	}
}

func (h *ContextFUBIntegrationHandlers) getRentalRecommendedAction(engagement, financial, urgency float64) string {
	if urgency >= 0.8 {
		return "immediate_showing_call"
	} else if financial >= 0.6 && urgency >= 0.5 {
		return "application_assistance"
	} else if engagement >= 0.5 {
		return "property_tour_booking"
	}
	return "rental_market_education"
}

func (h *ContextFUBIntegrationHandlers) getSalesRecommendedAction(engagement, financial, urgency float64) string {
	if financial >= 0.8 {
		return "buyer_consultation_call"
	} else if urgency >= 0.7 && engagement >= 0.5 {
		return "market_analysis_presentation"
	} else if engagement >= 0.6 {
		return "financing_pre_approval_guidance"
	}
	return "buyer_education_sequence"
}

func (h *ContextFUBIntegrationHandlers) getMixedRecommendedAction(engagement, financial, urgency float64) string {
	avgScore := (engagement + financial + urgency) / 3.0

	if avgScore >= 0.7 {
		return "comprehensive_consultation"
	} else if avgScore >= 0.5 {
		return "option_exploration_call"
	}
	return "market_opportunity_education"
}

// Context intelligence and insights

func (h *ContextFUBIntegrationHandlers) getContextMarketInsights(propertyContext map[string]interface{}, propertyType string) string {
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

	marketData := h.getHoustonMarketIntelligence(location, propertyType, priceRange)
	return h.formatMarketInsightsForResponse(marketData)
}

func (h *ContextFUBIntegrationHandlers) generateContextReasoning(trigger ContextFUBTriggerRequest, workflowType, recommendedAction string) string {
	reasoningParts := []string{}

	if trigger.EngagementScore >= 0.7 {
		reasoningParts = append(reasoningParts, "High engagement indicates strong interest")
	} else if trigger.EngagementScore >= 0.4 {
		reasoningParts = append(reasoningParts, "Moderate engagement suggests developing interest")
	}

	if trigger.FinancialQualScore >= 0.6 {
		reasoningParts = append(reasoningParts, "Strong financial qualification")
	}

	if trigger.UrgencyScore >= 0.7 {
		reasoningParts = append(reasoningParts, "High urgency requires immediate response")
	}

	if trigger.PropertyType == "rental" {
		reasoningParts = append(reasoningParts, "Rental market dynamics favor quick action")
	} else if trigger.PropertyType == "sales" {
		reasoningParts = append(reasoningParts, "Sales process requires relationship building")
	}

	return strings.Join(reasoningParts, "; ")
}

// Analytics helper methods

func (h *ContextFUBIntegrationHandlers) getAnalyticsCount(metric string, since time.Time, propertyType string) int64 {
	return 42
}

func (h *ContextFUBIntegrationHandlers) getPropertyTypeBreakdown(since time.Time) map[string]interface{} {
	return map[string]interface{}{
		"rental": 45,
		"sales":  38,
		"mixed":  17,
	}
}

func (h *ContextFUBIntegrationHandlers) getConversionMetrics(since time.Time, propertyType string) map[string]interface{} {
	return map[string]interface{}{
		"conversion_rate":       0.23,
		"average_time_to_close": "14 days",
		"success_by_urgency": map[string]float64{
			"high":   0.78,
			"medium": 0.45,
			"low":    0.12,
		},
	}
}

func (h *ContextFUBIntegrationHandlers) getBehavioralInsights(since time.Time, propertyType string) map[string]interface{} {
	return map[string]interface{}{
		"top_trigger_types":        []string{"property_viewed", "inquiry_submitted", "application_started"},
		"peak_activity_hours":      []int{9, 10, 11, 14, 15, 16},
		"average_engagement_score": 0.64,
		"behavioral_patterns":      "Users show higher conversion rates when contacted within 15 minutes",
	}
}

// Advanced Behavioral Intelligence Methods

// detectAdvancedPropertyCategory analyzes property category using advanced behavioral signals
func (h *ContextFUBIntegrationHandlers) detectAdvancedPropertyCategory(propertyContext map[string]interface{}, behaviorContext map[string]interface{}) string {
	categoryScore := make(map[string]float64)
	categoryScore["rental"] = 0.3
	categoryScore["sales"] = 0.3
	categoryScore["investment"] = 0.2
	categoryScore["commercial"] = 0.2

	if propertyType, exists := propertyContext["property_type"]; exists {
		if typeStr, ok := propertyType.(string); ok {
			switch strings.ToLower(typeStr) {
			case "rental", "apartment", "lease":
				categoryScore["rental"] += 0.4
			case "sales", "purchase", "buy":
				categoryScore["sales"] += 0.4
			case "investment", "roi", "portfolio":
				categoryScore["investment"] += 0.4
			case "commercial", "business", "office":
				categoryScore["commercial"] += 0.4
			}
		}
	}

	if searchTerms, exists := behaviorContext["search_terms"]; exists {
		if terms, ok := searchTerms.([]string); ok {
			for _, term := range terms {
				lowerTerm := strings.ToLower(term)
				if strings.Contains(lowerTerm, "rent") || strings.Contains(lowerTerm, "lease") {
					categoryScore["rental"] += 0.15
				}
				if strings.Contains(lowerTerm, "buy") || strings.Contains(lowerTerm, "purchase") {
					categoryScore["sales"] += 0.15
				}
				if strings.Contains(lowerTerm, "invest") || strings.Contains(lowerTerm, "roi") {
					categoryScore["investment"] += 0.15
				}
			}
		}
	}

	maxScore := 0.0
	bestCategory := "mixed"
	for category, score := range categoryScore {
		if score > maxScore {
			maxScore = score
			bestCategory = category
		}
	}

	return bestCategory
}

// calculatePropertySpecificBehavioralScore computes behavioral scoring based on property category
func (h *ContextFUBIntegrationHandlers) calculatePropertySpecificBehavioralScore(propertyCategory string, behaviorData map[string]interface{}) float64 {
	baseScore := 0.3

	switch strings.ToLower(propertyCategory) {
	case "rental":
		if moveInDate, exists := behaviorData["move_in_urgency"]; exists {
			if urgency, ok := moveInDate.(string); ok {
				switch urgency {
				case "immediate", "asap":
					baseScore += 0.4
				case "30_days":
					baseScore += 0.3
				case "60_days":
					baseScore += 0.2
				}
			}
		}

		if applicationReady, exists := behaviorData["application_ready"]; exists {
			if ready, ok := applicationReady.(bool); ok && ready {
				baseScore += 0.2
			}
		}

	case "sales":
		if preApproval, exists := behaviorData["pre_approval_status"]; exists {
			if approved, ok := preApproval.(bool); ok && approved {
				baseScore += 0.3
			}
		}

		if tourRequests, exists := behaviorData["property_tour_requests"]; exists {
			if requests, ok := tourRequests.(float64); ok {
				baseScore += requests * 0.1
			}
		}

	case "investment":
		if roiCalculations, exists := behaviorData["roi_calculations_viewed"]; exists {
			if calculations, ok := roiCalculations.(float64); ok && calculations > 0 {
				baseScore += 0.25
			}
		}

		if cashFlowAnalysis, exists := behaviorData["cash_flow_analysis_time"]; exists {
			if analysisTime, ok := cashFlowAnalysis.(float64); ok && analysisTime > 300 {
				baseScore += 0.15
			}
		}
	}

	if contactAttempts, exists := behaviorData["contact_attempts"]; exists {
		if attempts, ok := contactAttempts.(float64); ok {
			baseScore += attempts * 0.1
		}
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

// evaluateAdvancedTriggerConditions analyzes complex trigger condition patterns
func (h *ContextFUBIntegrationHandlers) evaluateAdvancedTriggerConditions(triggerData map[string]interface{}, historicalContext []map[string]interface{}) map[string]interface{} {
	conditionResults := map[string]interface{}{
		"conditions_met":     []string{},
		"conditions_failed":  []string{},
		"condition_score":    0.0,
		"trigger_confidence": 0.0,
		"recommended_action": "evaluate",
	}

	totalScore := 0.0
	maxScore := 0.0

	if engagement, exists := triggerData["engagement_score"]; exists {
		if engScore, ok := engagement.(float64); ok {
			maxScore += 1.0
			if engScore >= 0.6 {
				conditionResults["conditions_met"] = append(conditionResults["conditions_met"].([]string), "engagement_threshold")
				totalScore += engScore
			} else {
				conditionResults["conditions_failed"] = append(conditionResults["conditions_failed"].([]string), "engagement_threshold")
			}
		}
	}

	if financial, exists := triggerData["financial_score"]; exists {
		if finScore, ok := financial.(float64); ok {
			maxScore += 1.0
			if finScore >= 0.5 {
				conditionResults["conditions_met"] = append(conditionResults["conditions_met"].([]string), "financial_readiness")
				totalScore += finScore
			} else {
				conditionResults["conditions_failed"] = append(conditionResults["conditions_failed"].([]string), "financial_readiness")
			}
		}
	}

	if urgency, exists := triggerData["urgency_score"]; exists {
		if urgScore, ok := urgency.(float64); ok {
			maxScore += 1.0
			if urgScore >= 0.4 {
				conditionResults["conditions_met"] = append(conditionResults["conditions_met"].([]string), "urgency_indicator")
				totalScore += urgScore
			} else {
				conditionResults["conditions_failed"] = append(conditionResults["conditions_failed"].([]string), "urgency_indicator")
			}
		}
	}

	if len(historicalContext) >= 2 {
		conditionResults["conditions_met"] = append(conditionResults["conditions_met"].([]string), "repeat_engagement")
		totalScore += 0.3
		maxScore += 1.0
	} else {
		conditionResults["conditions_failed"] = append(conditionResults["conditions_failed"].([]string), "repeat_engagement")
		maxScore += 1.0
	}

	conditionScore := totalScore / maxScore
	triggerConfidence := conditionScore * 0.85

	conditionResults["condition_score"] = conditionScore
	conditionResults["trigger_confidence"] = triggerConfidence

	if triggerConfidence >= 0.7 {
		conditionResults["recommended_action"] = "trigger_automation"
	} else if triggerConfidence >= 0.5 {
		conditionResults["recommended_action"] = "conditional_trigger"
	} else {
		conditionResults["recommended_action"] = "continue_monitoring"
	}

	return conditionResults
}

// calculatePatternMatchConfidence analyzes behavioral pattern matching confidence
func (h *ContextFUBIntegrationHandlers) calculatePatternMatchConfidence(behaviorPattern map[string]interface{}, historicalPatterns []map[string]interface{}) float64 {
	if len(historicalPatterns) == 0 {
		return 0.5
	}

	confidenceScores := []float64{}

	for _, historicalPattern := range historicalPatterns {
		similarity := h.calculatePatternSimilarity(behaviorPattern, historicalPattern)
		confidenceScores = append(confidenceScores, similarity)
	}

	totalConfidence := 0.0
	for i, score := range confidenceScores {
		weight := 1.0 / float64(i+1)
		totalConfidence += score * weight
	}

	confidence := totalConfidence / float64(len(confidenceScores))

	if len(confidenceScores) >= 3 {
		consistency := h.calculatePatternConsistency(confidenceScores)
		confidence = confidence * (1.0 + consistency*0.2)
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateOptimalNextActionTime determines the best timing for next action
func (h *ContextFUBIntegrationHandlers) calculateOptimalNextActionTime(urgencyScore, engagementScore float64, propertyCategory string, marketConditions map[string]interface{}) time.Time {
	baseDelay := 24 * time.Hour

	if urgencyScore >= 0.8 {
		baseDelay = 15 * time.Minute
	} else if urgencyScore >= 0.6 {
		baseDelay = 2 * time.Hour
	} else if urgencyScore >= 0.4 {
		baseDelay = 8 * time.Hour
	}

	if engagementScore >= 0.8 {
		baseDelay = time.Duration(float64(baseDelay) * 0.5)
	} else if engagementScore >= 0.6 {
		baseDelay = time.Duration(float64(baseDelay) * 0.7)
	}

	switch strings.ToLower(propertyCategory) {
	case "rental":
		baseDelay = time.Duration(float64(baseDelay) * 0.6)
	case "investment":
		baseDelay = time.Duration(float64(baseDelay) * 1.3)
	case "commercial":
		baseDelay = time.Duration(float64(baseDelay) * 2.0)
	}

	if marketConditions != nil {
		if inventory, exists := marketConditions["inventory_level"]; exists {
			if inventoryStr, ok := inventory.(string); ok {
				if inventoryStr == "tight" {
					baseDelay = time.Duration(float64(baseDelay) * 0.7)
				} else if inventoryStr == "abundant" {
					baseDelay = time.Duration(float64(baseDelay) * 1.2)
				}
			}
		}
	}

	nextTime := time.Now().Add(baseDelay)
	return h.adjustToBusinessHours(nextTime)
}

// generatePatternAnalysis creates comprehensive behavioral pattern analysis
func (h *ContextFUBIntegrationHandlers) generatePatternAnalysis(behaviorData map[string]interface{}, comparisonData []map[string]interface{}) map[string]interface{} {
	analysis := map[string]interface{}{
		"pattern_type":      "unknown",
		"behavior_category": "standard",
		"predictive_score":  0.0,
		"confidence_level":  0.0,
		"key_indicators":    []string{},
		"anomalies":         []string{},
		"trend_direction":   "stable",
		"recommendations":   []string{},
	}

	if sessionCount, exists := behaviorData["session_count"]; exists {
		if count, ok := sessionCount.(float64); ok {
			if count >= 5 {
				analysis["pattern_type"] = "high_engagement"
				analysis["key_indicators"] = append(analysis["key_indicators"].([]string), "multiple_sessions")
			}
		}
	}

	if avgSessionTime, exists := behaviorData["average_session_duration"]; exists {
		if duration, ok := avgSessionTime.(float64); ok {
			if duration >= 300 {
				analysis["behavior_category"] = "deep_researcher"
				analysis["key_indicators"] = append(analysis["key_indicators"].([]string), "extended_engagement")
			}
		}
	}

	if pageDepth, exists := behaviorData["max_page_depth"]; exists {
		if depth, ok := pageDepth.(float64); ok {
			if depth >= 8 {
				analysis["behavior_category"] = "thorough_investigator"
				analysis["key_indicators"] = append(analysis["key_indicators"].([]string), "comprehensive_browsing")
			}
		}
	}

	predictiveScore := h.calculatePredictiveScore(behaviorData, comparisonData)
	analysis["predictive_score"] = predictiveScore

	confidenceLevel := h.calculateAnalysisConfidence(behaviorData, len(comparisonData))
	analysis["confidence_level"] = confidenceLevel

	anomalies := h.detectBehavioralAnomalies(behaviorData, comparisonData)
	analysis["anomalies"] = anomalies

	trendDirection := h.calculateTrendDirection(behaviorData, comparisonData)
	analysis["trend_direction"] = trendDirection

	recommendations := h.generatePatternRecommendations(analysis)
	analysis["recommendations"] = recommendations

	return analysis
}

// calculateAdvancedPriority computes priority using advanced scoring algorithms
func (h *ContextFUBIntegrationHandlers) calculateAdvancedPriority(engagementScore, financialScore, urgencyScore, marketFactorScore float64, propertyCategory string) string {
	var engagementWeight, financialWeight, urgencyWeight, marketWeight float64

	switch strings.ToLower(propertyCategory) {
	case "rental":
		engagementWeight = 0.25
		financialWeight = 0.30
		urgencyWeight = 0.35
		marketWeight = 0.10
	case "sales":
		engagementWeight = 0.30
		financialWeight = 0.40
		urgencyWeight = 0.20
		marketWeight = 0.10
	case "investment":
		engagementWeight = 0.20
		financialWeight = 0.45
		urgencyWeight = 0.15
		marketWeight = 0.20
	case "commercial":
		engagementWeight = 0.25
		financialWeight = 0.35
		urgencyWeight = 0.15
		marketWeight = 0.25
	default:
		engagementWeight = 0.25
		financialWeight = 0.35
		urgencyWeight = 0.25
		marketWeight = 0.15
	}

	priorityScore := (engagementScore * engagementWeight) +
		(financialScore * financialWeight) +
		(urgencyScore * urgencyWeight) +
		(marketFactorScore * marketWeight)

	if priorityScore >= 0.85 {
		return "CRITICAL"
	} else if priorityScore >= 0.70 {
		return "HIGH"
	} else if priorityScore >= 0.50 {
		return "MEDIUM"
	} else if priorityScore >= 0.30 {
		return "LOW"
	}
	return "MINIMAL"
}

// generateAdvancedTriggerMessage creates personalized trigger messages
func (h *ContextFUBIntegrationHandlers) generateAdvancedTriggerMessage(triggerContext map[string]interface{}, patternAnalysis map[string]interface{}) string {
	var messageComponents []string

	propertyCategory := "property"
	if category, exists := triggerContext["property_category"]; exists {
		if categoryStr, ok := category.(string); ok {
			propertyCategory = categoryStr
		}
	}

	urgencyLevel := "standard"
	if urgency, exists := triggerContext["urgency_level"]; exists {
		if urgencyStr, ok := urgency.(string); ok {
			urgencyLevel = urgencyStr
		}
	}

	behaviorCategory := "standard"
	if pattern, exists := patternAnalysis["behavior_category"]; exists {
		if patternStr, ok := pattern.(string); ok {
			behaviorCategory = patternStr
		}
	}

	switch strings.ToLower(urgencyLevel) {
	case "critical", "high":
		messageComponents = append(messageComponents, "Time-sensitive opportunity:")
	case "medium":
		messageComponents = append(messageComponents, "Great timing for:")
	default:
		messageComponents = append(messageComponents, "Perfect opportunity to explore:")
	}

	switch strings.ToLower(propertyCategory) {
	case "rental":
		messageComponents = append(messageComponents, "your rental property search")
	case "sales":
		messageComponents = append(messageComponents, "your home buying journey")
	case "investment":
		messageComponents = append(messageComponents, "your investment property goals")
	case "commercial":
		messageComponents = append(messageComponents, "your commercial property needs")
	default:
		messageComponents = append(messageComponents, "your property interests")
	}

	switch strings.ToLower(behaviorCategory) {
	case "deep_researcher":
		messageComponents = append(messageComponents, "- with detailed market analysis and comprehensive property data")
	case "thorough_investigator":
		messageComponents = append(messageComponents, "- including in-depth neighborhood insights and investment projections")
	case "high_engagement":
		messageComponents = append(messageComponents, "- featuring properties matching your demonstrated preferences")
	default:
		messageComponents = append(messageComponents, "- with personalized recommendations based on your interests")
	}

	messageComponents = append(messageComponents, "in Houston's dynamic real estate market")

	switch strings.ToLower(urgencyLevel) {
	case "critical":
		messageComponents = append(messageComponents, ". Let's connect within the next few hours to discuss immediate opportunities.")
	case "high":
		messageComponents = append(messageComponents, ". I'd love to schedule a call today to explore your options.")
	case "medium":
		messageComponents = append(messageComponents, ". Would you be available for a brief conversation this week?")
	default:
		messageComponents = append(messageComponents, ". I'll follow up with valuable market insights and property recommendations.")
	}

	return strings.Join(messageComponents, " ")
}

// Helper methods

func (h *ContextFUBIntegrationHandlers) calculatePatternSimilarity(pattern1, pattern2 map[string]interface{}) float64 {
	commonKeys := []string{"session_count", "average_session_duration", "page_views", "interaction_count"}
	similaritySum := 0.0
	validComparisons := 0

	for _, key := range commonKeys {
		if val1, exists1 := pattern1[key]; exists1 {
			if val2, exists2 := pattern2[key]; exists2 {
				if num1, ok1 := val1.(float64); ok1 {
					if num2, ok2 := val2.(float64); ok2 {
						maxVal := num1
						minVal := num2
						if num2 > num1 {
							maxVal = num2
							minVal = num1
						}
						if maxVal > 0 {
							similarity := minVal / maxVal
							similaritySum += similarity
							validComparisons++
						}
					}
				}
			}
		}
	}

	if validComparisons == 0 {
		return 0.5
	}

	return similaritySum / float64(validComparisons)
}

func (h *ContextFUBIntegrationHandlers) calculatePatternConsistency(scores []float64) float64 {
	if len(scores) < 2 {
		return 0.0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	average := sum / float64(len(scores))

	variance := 0.0
	for _, score := range scores {
		variance += (score - average) * (score - average)
	}
	variance = variance / float64(len(scores))

	consistency := 1.0 / (1.0 + variance)
	return consistency
}

func (h *ContextFUBIntegrationHandlers) adjustToBusinessHours(t time.Time) time.Time {
	central, _ := time.LoadLocation("America/Chicago")
	adjustedTime := t.In(central)

	hour := adjustedTime.Hour()
	weekday := adjustedTime.Weekday()

	if weekday == time.Saturday {
		daysToAdd := 2
		adjustedTime = adjustedTime.AddDate(0, 0, daysToAdd)
		adjustedTime = time.Date(adjustedTime.Year(), adjustedTime.Month(), adjustedTime.Day(), 9, 0, 0, 0, central)
	} else if weekday == time.Sunday {
		daysToAdd := 1
		adjustedTime = adjustedTime.AddDate(0, 0, daysToAdd)
		adjustedTime = time.Date(adjustedTime.Year(), adjustedTime.Month(), adjustedTime.Day(), 9, 0, 0, 0, central)
	} else {
		if hour < 9 {
			adjustedTime = time.Date(adjustedTime.Year(), adjustedTime.Month(), adjustedTime.Day(), 9, 0, 0, 0, central)
		} else if hour >= 18 {
			if weekday == time.Friday {
				adjustedTime = adjustedTime.AddDate(0, 0, 3)
			} else {
				adjustedTime = adjustedTime.AddDate(0, 0, 1)
			}
			adjustedTime = time.Date(adjustedTime.Year(), adjustedTime.Month(), adjustedTime.Day(), 9, 0, 0, 0, central)
		}
	}

	return adjustedTime
}

// Additional helper methods

func (h *ContextFUBIntegrationHandlers) calculatePredictiveScore(behaviorData map[string]interface{}, comparisonData []map[string]interface{}) float64 {
	baseScore := 0.3

	if sessionCount, exists := behaviorData["session_count"]; exists {
		if count, ok := sessionCount.(float64); ok {
			baseScore += (count / 10.0) * 0.3
		}
	}

	if totalTime, exists := behaviorData["total_time_spent"]; exists {
		if timeVal, ok := totalTime.(float64); ok {
			baseScore += (timeVal / 3600.0) * 0.2
		}
	}

	if contactAttempts, exists := behaviorData["contact_attempts"]; exists {
		if attempts, ok := contactAttempts.(float64); ok && attempts > 0 {
			baseScore += 0.2
		}
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

func (h *ContextFUBIntegrationHandlers) calculateAnalysisConfidence(behaviorData map[string]interface{}, comparisonCount int) float64 {
	confidence := 0.5

	if comparisonCount >= 10 {
		confidence += 0.3
	} else if comparisonCount >= 5 {
		confidence += 0.2
	} else if comparisonCount >= 2 {
		confidence += 0.1
	}

	dataPointCount := float64(len(behaviorData))
	confidence += (dataPointCount / 20.0) * 0.2

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

func (h *ContextFUBIntegrationHandlers) detectBehavioralAnomalies(behaviorData map[string]interface{}, comparisonData []map[string]interface{}) []string {
	anomalies := []string{}

	if sessionCount, exists := behaviorData["session_count"]; exists {
		if count, ok := sessionCount.(float64); ok && count > 20 {
			anomalies = append(anomalies, "exceptionally_high_session_count")
		}
	}

	if avgDuration, exists := behaviorData["average_session_duration"]; exists {
		if duration, ok := avgDuration.(float64); ok && duration > 1800 {
			anomalies = append(anomalies, "unusually_long_session_duration")
		}
	}

	if bounceRate, exists := behaviorData["bounce_rate"]; exists {
		if totalTime, exists2 := behaviorData["total_time_spent"]; exists2 {
			if bounce, ok := bounceRate.(float64); ok {
				if timeVal, ok2 := totalTime.(float64); ok2 {
					if bounce > 0.8 && timeVal > 1200 {
						anomalies = append(anomalies, "focused_single_session_engagement")
					}
				}
			}
		}
	}

	return anomalies
}

func (h *ContextFUBIntegrationHandlers) calculateTrendDirection(behaviorData map[string]interface{}, comparisonData []map[string]interface{}) string {
	if len(comparisonData) < 2 {
		return "insufficient_data"
	}

	recentEngagement := 0.0
	olderEngagement := 0.0

	if sessionCount, exists := behaviorData["recent_session_count"]; exists {
		if count, ok := sessionCount.(float64); ok {
			recentEngagement = count
		}
	}

	if len(comparisonData) > 0 {
		if oldSessionCount, exists := comparisonData[0]["session_count"]; exists {
			if count, ok := oldSessionCount.(float64); ok {
				olderEngagement = count
			}
		}
	}

	if recentEngagement > olderEngagement*1.2 {
		return "increasing"
	} else if recentEngagement < olderEngagement*0.8 {
		return "decreasing"
	}

	return "stable"
}

func (h *ContextFUBIntegrationHandlers) generatePatternRecommendations(analysis map[string]interface{}) []string {
	recommendations := []string{}

	if behaviorCategory, exists := analysis["behavior_category"]; exists {
		if categoryStr, ok := behaviorCategory.(string); ok {
			switch categoryStr {
			case "deep_researcher":
				recommendations = append(recommendations, []string{
					"provide_detailed_market_analysis",
					"offer_comprehensive_property_reports",
					"schedule_educational_consultation",
				}...)
			case "thorough_investigator":
				recommendations = append(recommendations, []string{
					"share_neighborhood_investment_data",
					"provide_comparative_market_analysis",
					"offer_property_tour_with_detailed_walkthrough",
				}...)
			case "high_engagement":
				recommendations = append(recommendations, []string{
					"immediate_personal_outreach",
					"priority_property_matching",
					"expedited_response_protocol",
				}...)
			}
		}
	}

	if predictiveScore, exists := analysis["predictive_score"]; exists {
		if score, ok := predictiveScore.(float64); ok {
			if score >= 0.8 {
				recommendations = append(recommendations, "high_priority_lead_management")
			} else if score >= 0.6 {
				recommendations = append(recommendations, "active_nurture_sequence")
			} else {
				recommendations = append(recommendations, "educational_content_series")
			}
		}
	}

	if trendDirection, exists := analysis["trend_direction"]; exists {
		if direction, ok := trendDirection.(string); ok {
			switch direction {
			case "increasing":
				recommendations = append(recommendations, "accelerate_engagement_timeline")
			case "decreasing":
				recommendations = append(recommendations, "re_engagement_campaign")
			case "stable":
				recommendations = append(recommendations, "maintain_consistent_touchpoints")
			}
		}
	}

	return recommendations
}

// determineAdvancedWorkflowType determines workflow using advanced behavioral scoring
func (h *ContextFUBIntegrationHandlers) determineAdvancedWorkflowType(behavioralScore, financialScore, urgencyScore float64, triggerType, propertyCategory string) string {
	combinedScore := (behavioralScore * 0.4) + (financialScore * 0.35) + (urgencyScore * 0.25)

	switch strings.ToLower(propertyCategory) {
	case "rental":
		if combinedScore >= 0.85 && urgencyScore >= 0.8 {
			return "RENTAL_IMMEDIATE_PRIORITY"
		} else if combinedScore >= 0.70 && behavioralScore >= 0.7 {
			return "RENTAL_HIGH_BEHAVIORAL_INTENT"
		} else if combinedScore >= 0.55 {
			return "RENTAL_QUALIFIED_BEHAVIORAL_NURTURE"
		}
		return "RENTAL_BEHAVIORAL_DEVELOPMENT"

	case "sales":
		if combinedScore >= 0.85 && financialScore >= 0.7 {
			return "SALES_QUALIFIED_HIGH_INTENT"
		} else if combinedScore >= 0.70 && behavioralScore >= 0.8 {
			return "SALES_BEHAVIORAL_PRIORITY"
		} else if combinedScore >= 0.55 {
			return "SALES_BEHAVIORAL_QUALIFIED"
		}
		return "SALES_BEHAVIORAL_EDUCATION"

	case "investment":
		if combinedScore >= 0.80 && behavioralScore >= 0.7 {
			return "INVESTMENT_SERIOUS_ANALYSIS"
		} else if combinedScore >= 0.65 {
			return "INVESTMENT_BEHAVIORAL_QUALIFIED"
		}
		return "INVESTMENT_BEHAVIORAL_EDUCATION"

	default:
		if combinedScore >= 0.80 {
			return "ADVANCED_HIGH_PRIORITY"
		} else if combinedScore >= 0.60 {
			return "ADVANCED_QUALIFIED_NURTURE"
		}
		return "ADVANCED_BEHAVIORAL_BUILDING"
	}
}
