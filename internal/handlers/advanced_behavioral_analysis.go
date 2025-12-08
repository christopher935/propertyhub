package handlers

import (
	"strings"
)

// getCommunityGrowthInsights analyzes community growth patterns for enhanced behavioral triggers
func (h *ContextFUBIntegrationHandlers) getCommunityGrowthInsights(location string) map[string]interface{} {
	marketIntel := h.getHoustonMarketIntelligence(location, "", "")

	insights := map[string]interface{}{
		"growth_rate":       0.082,
		"development_index": 7.8,
		"investment_appeal": "high",
		"market_stability":  "strong",
	}

	if rentalMarket, exists := marketIntel["rental_market"]; exists {
		if rentalMap, ok := rentalMarket.(map[string]interface{}); ok {
			if neighborhoods, exists := rentalMap["popular_neighborhoods"]; exists {
				if neighSlice, ok := neighborhoods.([]string); ok && len(neighSlice) > 0 {
					insights["hottest_communities"] = neighSlice[:3]
				}
			}
		}
	}

	if strings.Contains(strings.ToLower(location), "heights") {
		insights["growth_rate"] = 0.095
		insights["character"] = "historic charm with modern appeal"
	} else if strings.Contains(strings.ToLower(location), "medical center") {
		insights["growth_rate"] = 0.089
		insights["character"] = "professional hub with strong rental demand"
	}

	return insights
}

// getCompetitiveFactors analyzes competitive market factors for property category decisions
func (h *ContextFUBIntegrationHandlers) getCompetitiveFactors(propertyCategory interface{}) map[string]interface{} {
	return map[string]interface{}{
		"market_competition":    "moderate-to-high",
		"buyer_advantage":       false,
		"seller_market":         true,
		"negotiation_leverage":  "seller-favored",
		"inventory_tightness":   0.73,
		"demand_pressure":       "high",
		"pricing_power":         "strong",
		"time_on_market_trend":  "decreasing",
		"offer_competition":     "multiple offers common",
		"cash_buyer_percentage": 0.28,
	}
}

// analyzeUrgencyFactors provides comprehensive urgency analysis for decision-making
func (h *ContextFUBIntegrationHandlers) analyzeUrgencyFactors(engagement, financial, timeline float64, propertyContext map[string]interface{}) map[string]interface{} {
	location := ""
	if loc, exists := propertyContext["location"]; exists {
		if locStr, ok := loc.(string); ok {
			location = locStr
		}
	}

	marketIntel := h.getHoustonMarketIntelligence(location, "", "")

	inventoryLevel := "balanced"
	monthsOfInventory := 2.8

	if salesMarket, exists := marketIntel["sales_market"]; exists {
		if salesMap, ok := salesMarket.(map[string]interface{}); ok {
			if inventory, exists := salesMap["months_of_inventory"]; exists {
				if invFloat, ok := inventory.(float64); ok {
					monthsOfInventory = invFloat
					if invFloat < 2.0 {
						inventoryLevel = "tight"
					} else if invFloat > 4.0 {
						inventoryLevel = "buyer_favorable"
					}
				}
			}
		}
	}

	urgencyScore := (engagement + financial + timeline) / 3.0

	if inventoryLevel == "tight" {
		urgencyScore *= 1.25
	}

	urgencyAnalysis := map[string]interface{}{
		"composite_urgency":       urgencyScore,
		"market_pressure":         inventoryLevel,
		"months_of_inventory":     monthsOfInventory,
		"action_timeline":         h.calculateRecommendedTimeline(urgencyScore),
		"competitive_environment": h.assessCompetitiveEnvironment(urgencyScore, inventoryLevel),
		"decision_factors": []string{
			"Market inventory levels",
			"Seasonal timing",
			"Interest rate environment",
			"Local demand patterns",
		},
	}

	if urgencyScore >= 0.8 {
		urgencyAnalysis["recommendation"] = "immediate_action_required"
		urgencyAnalysis["timeline"] = "within_24_hours"
	} else if urgencyScore >= 0.6 {
		urgencyAnalysis["recommendation"] = "prompt_follow_up"
		urgencyAnalysis["timeline"] = "within_48_hours"
	} else {
		urgencyAnalysis["recommendation"] = "standard_nurture"
		urgencyAnalysis["timeline"] = "within_week"
	}

	return urgencyAnalysis
}

// analyzeEngagementDepth provides detailed engagement pattern analysis
func (h *ContextFUBIntegrationHandlers) analyzeEngagementDepth(behaviorData map[string]interface{}, sessionHistory []map[string]interface{}) map[string]interface{} {
	baseEngagement := 0.3

	if viewTime, exists := behaviorData["total_view_time"]; exists {
		if timeFloat, ok := viewTime.(float64); ok {
			baseEngagement += (timeFloat / 300.0) * 0.4
		}
	}

	if interactions, exists := behaviorData["interaction_count"]; exists {
		if countFloat, ok := interactions.(float64); ok {
			baseEngagement += (countFloat / 10.0) * 0.3
		}
	}

	sessionCount := float64(len(sessionHistory))
	sessionScore := sessionCount / 5.0 * 0.2

	if sessionScore > 0.2 {
		sessionScore = 0.2
	}

	finalScore := baseEngagement + sessionScore
	if finalScore > 1.0 {
		finalScore = 1.0
	}

	return map[string]interface{}{
		"engagement_score":    finalScore,
		"session_count":       sessionCount,
		"interaction_quality": h.assessInteractionQuality(behaviorData),
		"attention_span":      h.calculateSessionDepth(behaviorData),
		"information_seeking": h.calculateInformationDepth(behaviorData),
		"behavioral_signals":  h.checkBehavioralSignal(behaviorData),
	}
}

// assessFinancialReadiness evaluates financial qualification and readiness signals
func (h *ContextFUBIntegrationHandlers) assessFinancialReadiness(financialContext map[string]interface{}, behaviorIndicators map[string]interface{}) map[string]interface{} {
	baseScore := 0.2

	if preApproved, exists := financialContext["pre_approved"]; exists {
		if approved, ok := preApproved.(bool); ok && approved {
			baseScore += 0.4
		}
	}

	if creditScore, exists := financialContext["credit_score"]; exists {
		if score, ok := creditScore.(float64); ok {
			if score >= 750 {
				baseScore += 0.3
			} else if score >= 700 {
				baseScore += 0.2
			} else if score >= 650 {
				baseScore += 0.1
			}
		}
	}

	if incomeVerified, exists := financialContext["income_verified"]; exists {
		if verified, ok := incomeVerified.(bool); ok && verified {
			baseScore += 0.2
		}
	}

	if priceSearches, exists := behaviorIndicators["price_focused_searches"]; exists {
		if searches, ok := priceSearches.(float64); ok && searches > 5 {
			baseScore += 0.1
		}
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}

	readinessLevel := h.categorizeFinancialReadiness(baseScore)
	nextSteps := h.recommendFinancialNextSteps(baseScore, financialContext)

	return map[string]interface{}{
		"financial_score":      baseScore,
		"readiness_level":      readinessLevel,
		"qualification_status": h.determineQualificationStatus(baseScore),
		"recommended_actions":  nextSteps,
		"timeline_estimate":    h.estimateFinancialTimeline(baseScore),
		"risk_factors":         h.identifyFinancialRisks(financialContext),
	}
}

// analyzeConversionFactors evaluates factors influencing conversion likelihood
func (h *ContextFUBIntegrationHandlers) analyzeConversionFactors(engagementData, financialData, urgencyData map[string]interface{}) map[string]interface{} {
	conversionScore := 0.0
	factors := []string{}

	if engScore, exists := engagementData["engagement_score"]; exists {
		if score, ok := engScore.(float64); ok {
			conversionScore += score * 0.4
			if score >= 0.7 {
				factors = append(factors, "high_engagement")
			}
		}
	}

	if finScore, exists := financialData["financial_score"]; exists {
		if score, ok := finScore.(float64); ok {
			conversionScore += score * 0.4
			if score >= 0.6 {
				factors = append(factors, "financially_qualified")
			}
		}
	}

	if urgScore, exists := urgencyData["composite_urgency"]; exists {
		if score, ok := urgScore.(float64); ok {
			conversionScore += score * 0.2
			if score >= 0.8 {
				factors = append(factors, "high_urgency")
			}
		}
	}

	probability := h.categorizeConversionProbability(conversionScore)

	return map[string]interface{}{
		"conversion_score":     conversionScore,
		"probability_category": probability,
		"contributing_factors": factors,
		"estimated_timeline":   h.estimateConversionTimelineAdvanced(conversionScore, factors),
		"recommended_approach": h.determineConversionApproach(conversionScore, factors),
		"success_indicators":   h.identifySuccessIndicators(factors),
	}
}

// getPropertySpecificActions generates property-type specific action recommendations
func (h *ContextFUBIntegrationHandlers) getPropertySpecificActions(propertyCategory interface{}, behavioralScore float64, urgencyLevel string) []string {
	actions := []string{}

	categoryStr := ""
	if category, ok := propertyCategory.(string); ok {
		categoryStr = strings.ToLower(category)
	}

	baseActions := []string{
		"property_information_packet",
		"market_analysis_report",
		"consultation_scheduling",
	}

	switch categoryStr {
	case "rental":
		actions = append(actions, []string{
			"lease_application_preparation",
			"rental_market_comparison",
			"move_in_timeline_discussion",
			"rental_terms_explanation",
		}...)
	case "sales":
		actions = append(actions, []string{
			"buyer_qualification_review",
			"financing_options_discussion",
			"comparable_sales_analysis",
			"purchase_process_explanation",
		}...)
	case "investment":
		actions = append(actions, []string{
			"roi_calculation_worksheet",
			"investment_property_analysis",
			"financing_strategy_review",
			"market_timing_assessment",
		}...)
	default:
		actions = append(actions, baseActions...)
	}

	if urgencyLevel == "high" || behavioralScore >= 0.8 {
		actions = append([]string{"immediate_callback_scheduling"}, actions...)
	}

	return actions
}

// getMarketIntelligentActions provides market-informed action recommendations
func (h *ContextFUBIntegrationHandlers) getMarketIntelligentActions(propertyCategory interface{}, location string) []string {
	categoryStr := ""
	if category, ok := propertyCategory.(string); ok {
		categoryStr = category
	}

	marketIntel := h.getHoustonMarketIntelligence(location, categoryStr, "")
	actions := []string{}

	if rentalMarket, exists := marketIntel["rental_market"]; exists {
		if rentalMap, ok := rentalMarket.(map[string]interface{}); ok {
			if avgDays, exists := rentalMap["average_days_on_market"]; exists {
				if days, ok := avgDays.(int); ok && days < 20 {
					actions = append(actions, "fast_moving_market_alert")
				}
			}
		}
	}

	if salesMarket, exists := marketIntel["sales_market"]; exists {
		if salesMap, ok := salesMarket.(map[string]interface{}); ok {
			if inventory, exists := salesMap["months_of_inventory"]; exists {
				if months, ok := inventory.(float64); ok && months < 3.0 {
					actions = append(actions, "tight_inventory_strategy")
				}
			}
		}
	}

	actions = append(actions, []string{
		"current_market_conditions_briefing",
		"neighborhood_trend_analysis",
		"pricing_strategy_discussion",
		"timing_optimization_review",
	}...)

	return actions
}

// prioritizeAndDeduplicateActions optimizes action recommendations
func (h *ContextFUBIntegrationHandlers) prioritizeAndDeduplicateActions(propertyActions, marketActions []string, urgencyScore, engagementScore float64) []string {
	actionMap := make(map[string]int)

	for i, action := range propertyActions {
		priority := len(propertyActions) - i
		actionMap[action] = priority
	}

	for i, action := range marketActions {
		priority := len(marketActions) - i
		if existing, exists := actionMap[action]; exists {
			actionMap[action] = existing + priority + 5
		} else {
			actionMap[action] = priority
		}
	}

	urgencyMultiplier := 1.0
	if urgencyScore >= 0.8 {
		urgencyMultiplier = 2.0
	} else if urgencyScore >= 0.6 {
		urgencyMultiplier = 1.5
	}

	engagementMultiplier := 1.0
	if engagementScore >= 0.7 {
		engagementMultiplier = 1.8
	} else if engagementScore >= 0.5 {
		engagementMultiplier = 1.3
	}

	type actionPriority struct {
		action   string
		priority float64
	}

	var sortedActions []actionPriority
	for action, priority := range actionMap {
		adjustedPriority := float64(priority) * urgencyMultiplier * engagementMultiplier
		sortedActions = append(sortedActions, actionPriority{action, adjustedPriority})
	}

	for i := 0; i < len(sortedActions)-1; i++ {
		for j := i + 1; j < len(sortedActions); j++ {
			if sortedActions[i].priority < sortedActions[j].priority {
				sortedActions[i], sortedActions[j] = sortedActions[j], sortedActions[i]
			}
		}
	}

	result := []string{}
	maxActions := 8
	if len(sortedActions) < maxActions {
		maxActions = len(sortedActions)
	}

	for i := 0; i < maxActions; i++ {
		result = append(result, sortedActions[i].action)
	}

	return result
}

// calculateRecommendedTimeline determines action timeline based on urgency
func (h *ContextFUBIntegrationHandlers) calculateRecommendedTimeline(urgencyScore float64) string {
	if urgencyScore >= 0.8 {
		return "within_4_hours"
	} else if urgencyScore >= 0.6 {
		return "within_24_hours"
	} else if urgencyScore >= 0.4 {
		return "within_48_hours"
	}
	return "within_week"
}

// assessCompetitiveEnvironment evaluates market competitiveness
func (h *ContextFUBIntegrationHandlers) assessCompetitiveEnvironment(urgencyScore float64, inventoryLevel string) string {
	if inventoryLevel == "tight" && urgencyScore >= 0.7 {
		return "highly_competitive"
	} else if inventoryLevel == "tight" || urgencyScore >= 0.6 {
		return "moderately_competitive"
	}
	return "balanced_market"
}

// calculateSessionDepth analyzes session interaction depth
func (h *ContextFUBIntegrationHandlers) calculateSessionDepth(behaviorData map[string]interface{}) float64 {
	depth := 0.3

	if pageViews, exists := behaviorData["pages_viewed"]; exists {
		if views, ok := pageViews.(float64); ok {
			depth += (views / 20.0) * 0.4
		}
	}

	if timeSpent, exists := behaviorData["session_duration"]; exists {
		if duration, ok := timeSpent.(float64); ok {
			depth += (duration / 600.0) * 0.3
		}
	}

	if depth > 1.0 {
		depth = 1.0
	}

	return depth
}

// calculateInformationDepth evaluates information-seeking behavior
func (h *ContextFUBIntegrationHandlers) calculateInformationDepth(behaviorData map[string]interface{}) float64 {
	infoScore := 0.2

	if downloads, exists := behaviorData["document_downloads"]; exists {
		if count, ok := downloads.(float64); ok {
			infoScore += count * 0.15
		}
	}

	if contactViews, exists := behaviorData["contact_form_views"]; exists {
		if views, ok := contactViews.(float64); ok {
			infoScore += views * 0.1
		}
	}

	if faqTime, exists := behaviorData["faq_time"]; exists {
		if timeVal, ok := faqTime.(float64); ok {
			infoScore += (timeVal / 120.0) * 0.2
		}
	}

	if infoScore > 1.0 {
		infoScore = 1.0
	}

	return infoScore
}

// assessInteractionQuality evaluates the quality of user interactions
func (h *ContextFUBIntegrationHandlers) assessInteractionQuality(behaviorData map[string]interface{}) string {
	qualityScore := 0.0

	if formStarts, exists := behaviorData["form_starts"]; exists {
		if starts, ok := formStarts.(float64); ok {
			qualityScore += starts * 0.2
		}
	}

	// FIXED: Remove unused formCompletions variable, use completions directly
	if formCompletions, exists := behaviorData["form_completions"]; exists {
		if completions, ok := formCompletions.(float64); ok {
			qualityScore += completions * 0.4
		}
	}

	if contentShares, exists := behaviorData["content_shares"]; exists {
		if shares, ok := contentShares.(float64); ok {
			qualityScore += shares * 0.3
		}
	}

	if qualityScore >= 0.8 {
		return "high_quality"
	} else if qualityScore >= 0.5 {
		return "moderate_quality"
	}
	return "basic_interaction"
}

// checkBehavioralSignal identifies specific behavioral signals
func (h *ContextFUBIntegrationHandlers) checkBehavioralSignal(behaviorData map[string]interface{}) []string {
	signals := []string{}

	if returnVisits, exists := behaviorData["return_visits"]; exists {
		if visits, ok := returnVisits.(float64); ok && visits >= 3 {
			signals = append(signals, "high_interest")
		}
	}

	if timeSpent, exists := behaviorData["average_session_time"]; exists {
		if timeVal, ok := timeSpent.(float64); ok && timeVal >= 300 {
			signals = append(signals, "deep_engagement")
		}
	}

	if contactAttempts, exists := behaviorData["contact_attempts"]; exists {
		if attempts, ok := contactAttempts.(float64); ok && attempts > 0 {
			signals = append(signals, "ready_to_connect")
		}
	}

	if len(signals) == 0 {
		signals = append(signals, "standard_browsing")
	}

	return signals
}

// Categorization methods

func (h *ContextFUBIntegrationHandlers) categorizeUrgencyLevel(score float64) string {
	if score >= 0.8 {
		return "critical"
	} else if score >= 0.6 {
		return "high"
	} else if score >= 0.4 {
		return "moderate"
	}
	return "low"
}

func (h *ContextFUBIntegrationHandlers) categorizeEngagementLevel(score float64) string {
	if score >= 0.8 {
		return "highly_engaged"
	} else if score >= 0.6 {
		return "engaged"
	} else if score >= 0.4 {
		return "moderately_engaged"
	}
	return "low_engagement"
}

func (h *ContextFUBIntegrationHandlers) categorizeFinancialReadiness(score float64) string {
	if score >= 0.8 {
		return "highly_qualified"
	} else if score >= 0.6 {
		return "qualified"
	} else if score >= 0.4 {
		return "potentially_qualified"
	}
	return "needs_qualification"
}

func (h *ContextFUBIntegrationHandlers) calculateRecommendedResponseTime(urgency, engagement float64) string {
	combinedScore := (urgency + engagement) / 2.0

	if combinedScore >= 0.8 {
		return "immediate_response"
	} else if combinedScore >= 0.6 {
		return "priority_response"
	} else if combinedScore >= 0.4 {
		return "standard_response"
	}
	return "scheduled_follow_up"
}

func (h *ContextFUBIntegrationHandlers) recommendFinancialNextSteps(score float64, context map[string]interface{}) []string {
	steps := []string{}

	if score < 0.4 {
		steps = append(steps, []string{
			"financial_pre_qualification",
			"lender_referral_program",
			"credit_improvement_guidance",
		}...)
	} else if score < 0.7 {
		steps = append(steps, []string{
			"document_preparation_checklist",
			"pre_approval_assistance",
			"financing_options_review",
		}...)
	} else {
		steps = append(steps, []string{
			"purchase_readiness_confirmation",
			"rate_lock_discussion",
			"closing_timeline_planning",
		}...)
	}

	return steps
}

func (h *ContextFUBIntegrationHandlers) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (h *ContextFUBIntegrationHandlers) categorizeConversionProbability(score float64) string {
	if score >= 0.8 {
		return "very_high"
	} else if score >= 0.6 {
		return "high"
	} else if score >= 0.4 {
		return "moderate"
	}
	return "low"
}

func (h *ContextFUBIntegrationHandlers) estimateConversionTimelineAdvanced(score float64, factors []string) string {
	hasHighUrgency := h.containsString(factors, "high_urgency")
	hasFinancialQual := h.containsString(factors, "financially_qualified")

	if score >= 0.8 && hasHighUrgency && hasFinancialQual {
		return "1-2_weeks"
	} else if score >= 0.6 && (hasHighUrgency || hasFinancialQual) {
		return "2-4_weeks"
	} else if score >= 0.4 {
		return "1-3_months"
	}
	return "3-6_months"
}

func (h *ContextFUBIntegrationHandlers) determineQualificationStatus(score float64) string {
	if score >= 0.8 {
		return "pre_approved_ready"
	} else if score >= 0.6 {
		return "qualification_in_progress"
	} else if score >= 0.4 {
		return "needs_pre_qualification"
	}
	return "financial_consultation_required"
}

func (h *ContextFUBIntegrationHandlers) estimateFinancialTimeline(score float64) string {
	if score >= 0.8 {
		return "ready_to_transact"
	} else if score >= 0.6 {
		return "2_weeks_to_ready"
	} else if score >= 0.4 {
		return "1_month_to_ready"
	}
	return "3_months_to_ready"
}

func (h *ContextFUBIntegrationHandlers) identifyFinancialRisks(context map[string]interface{}) []string {
	risks := []string{}

	if creditScore, exists := context["credit_score"]; exists {
		if score, ok := creditScore.(float64); ok && score < 650 {
			risks = append(risks, "credit_score_concerns")
		}
	}

	if employment, exists := context["employment_stability"]; exists {
		if stable, ok := employment.(bool); ok && !stable {
			risks = append(risks, "employment_instability")
		}
	}

	if downPayment, exists := context["down_payment_saved"]; exists {
		if saved, ok := downPayment.(bool); ok && !saved {
			risks = append(risks, "insufficient_down_payment")
		}
	}

	if len(risks) == 0 {
		risks = append(risks, "low_financial_risk")
	}

	return risks
}

func (h *ContextFUBIntegrationHandlers) determineConversionApproach(score float64, factors []string) string {
	if score >= 0.8 {
		return "aggressive_pursuit"
	} else if score >= 0.6 {
		return "active_engagement"
	} else if score >= 0.4 {
		return "nurture_and_educate"
	}
	return "long_term_cultivation"
}

func (h *ContextFUBIntegrationHandlers) identifySuccessIndicators(factors []string) []string {
	indicators := []string{"continued_engagement", "response_to_outreach"}

	if h.containsString(factors, "financially_qualified") {
		indicators = append(indicators, "financial_documentation_completion")
	}

	if h.containsString(factors, "high_urgency") {
		indicators = append(indicators, "property_viewing_requests")
	}

	if h.containsString(factors, "high_engagement") {
		indicators = append(indicators, "referral_potential")
	}

	return indicators
}
