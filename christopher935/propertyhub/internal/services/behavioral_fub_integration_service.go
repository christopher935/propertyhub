package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BehavioralFUBIntegrationService orchestrates behavioral intelligence with FUB operations
type BehavioralFUBIntegrationService struct {
	db        *gorm.DB
	apiClient *BehavioralFUBAPIClient
}

// NewBehavioralFUBIntegrationService creates a new behavioral FUB integration service
func NewBehavioralFUBIntegrationService(db *gorm.DB, apiKey string) *BehavioralFUBIntegrationService {
	return &BehavioralFUBIntegrationService{
		db:        db,
		apiClient: NewBehavioralFUBAPIClient(db, apiKey),
	}
}

// PropertyCategoryTriggerData represents property-specific trigger data
type PropertyCategoryTriggerData struct {
	PropertyCategory string                 `json:"property_category"`
	PropertyTier     string                 `json:"property_tier"`
	TargetDemo       string                 `json:"target_demo"`
	PriceRange       string                 `json:"price_range"`
	Location         string                 `json:"location"`
	BehavioralData   map[string]interface{} `json:"behavioral_data"`
	MarketContext    map[string]interface{} `json:"market_context"`
}

// ProcessBehavioralTrigger processes a behavioral intelligence trigger and executes FUB operations
func (service *BehavioralFUBIntegrationService) ProcessBehavioralTrigger(triggerData *PropertyCategoryTriggerData, contactInfo map[string]string) (*BehavioralTriggerResult, error) {
	log.Printf("🧠 Processing behavioral trigger for category: %s, tier: %s", triggerData.PropertyCategory, triggerData.PropertyTier)

	result := &BehavioralTriggerResult{
		ProcessedAt: time.Now(),
		PropertyCategory: map[string]interface{}{
			"category":    triggerData.PropertyCategory,
			"tier":        triggerData.PropertyTier,
			"target_demo": triggerData.TargetDemo,
			"price_range": triggerData.PriceRange,
			"location":    triggerData.Location,
		},
		BehavioralData: triggerData.BehavioralData,
		MarketContext:  triggerData.MarketContext,
	}

	// Step 1: Create or update contact
	contact, err := service.createOrUpdateContact(contactInfo, triggerData)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("Failed to create/update contact: %v", err)
		return result, err
	}
	result.ContactID = contact.ID

	// Step 2: Determine workflow type and priority based on behavioral intelligence
	workflowType, priority := service.determineWorkflowAndPriority(triggerData)
	result.WorkflowType = workflowType
	result.Priority = priority

	// Step 3: Create deal with behavioral intelligence data
	deal, err := service.createBehavioralDeal(contact.ID, triggerData)
	if err != nil {
		log.Printf("⚠️ Failed to create deal: %v", err) // Non-fatal
	} else {
		result.DealID = deal.ID
	}

	// Step 4: Assign appropriate action plan
	actionPlanID, err := service.assignBehavioralActionPlan(contact.ID, triggerData)
	if err != nil {
		log.Printf("⚠️ Failed to assign action plan: %v", err) // Non-fatal
	} else {
		result.ActionPlanID = actionPlanID
	}

	// Step 5: Add to appropriate pond
	pondID, err := service.assignToBehavioralPond(contact.ID, triggerData)
	if err != nil {
		log.Printf("⚠️ Failed to assign to pond: %v", err) // Non-fatal
	} else {
		result.PondID = pondID
	}

	// Step 6: Create immediate action task
	err = service.createImmediateActionTask(contact.ID, triggerData, priority)
	if err != nil {
		log.Printf("⚠️ Failed to create immediate action task: %v", err) // Non-fatal
	}

	// Step 7: Set recommended action and scheduling
	result.RecommendedAction = service.getRecommendedAction(triggerData)
	result.ScheduledAt = service.calculateOptimalResponseTime(triggerData, priority)

	result.Success = true
	log.Printf("✅ Successfully processed behavioral trigger - Contact: %s, Deal: %s, Priority: %s",
		result.ContactID, result.DealID, result.Priority)

	return result, nil
}

// createOrUpdateContact creates or updates a FUB contact with behavioral intelligence data
func (service *BehavioralFUBIntegrationService) createOrUpdateContact(contactInfo map[string]string, triggerData *PropertyCategoryTriggerData) (*FUBContact, error) {
	// Extract name components
	fullName := contactInfo["name"]
	nameParts := strings.Fields(fullName)
	firstName := nameParts[0]
	lastName := ""
	if len(nameParts) > 1 {
		lastName = strings.Join(nameParts[1:], " ")
	}

	// Create behavioral tags based on property category and triggers
	tags := service.generateBehavioralTags(triggerData)

	// Create custom fields with behavioral intelligence
	customFields := map[string]interface{}{
		"property_category":     triggerData.PropertyCategory,
		"property_tier":         triggerData.PropertyTier,
		"target_demographic":    triggerData.TargetDemo,
		"price_range":           triggerData.PriceRange,
		"behavioral_urgency":    triggerData.BehavioralData["urgency_score"],
		"behavioral_financial":  triggerData.BehavioralData["financial_readiness"],
		"behavioral_engagement": triggerData.BehavioralData["engagement_depth"],
		"market_conditions":     triggerData.MarketContext["seasonal_factor"],
		"houston_location":      triggerData.Location,
	}

	contact := &FUBContact{
		Name:         fullName,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        contactInfo["email"],
		Phone:        contactInfo["phone"],
		Source:       service.getSourceFromCategory(triggerData.PropertyCategory),
		Tags:         tags,
		CustomFields: customFields,
	}

	// Try to create the contact
	result, err := service.apiClient.CreateContact(contact)
	if err != nil {
		// If creation fails, it might already exist - try to find and update
		log.Printf("🔄 Contact creation failed, attempting update: %v", err)
		// For now, return the creation error - in production you'd implement contact lookup
		return nil, err
	}

	return result, nil
}

// determineWorkflowAndPriority determines the appropriate workflow and priority based on behavioral intelligence
func (service *BehavioralFUBIntegrationService) determineWorkflowAndPriority(triggerData *PropertyCategoryTriggerData) (string, string) {
	category := triggerData.PropertyCategory
	tier := triggerData.PropertyTier
	urgencyScore, _ := triggerData.BehavioralData["urgency_score"].(float64)

	// Determine workflow based on property category and behavioral signals
	switch {
	case strings.Contains(category, "luxury") && urgencyScore > 80:
		return "luxury_immediate_concierge", "URGENT"
	case strings.Contains(category, "luxury"):
		return "luxury_buyer_premium_service", "HIGH"
	case strings.Contains(category, "investment") && urgencyScore > 75:
		return "investment_rapid_analysis", "HIGH"
	case strings.Contains(category, "investment"):
		return "investment_portfolio_analysis", "MEDIUM"
	case strings.Contains(category, "student") && urgencyScore > 70:
		return "student_semester_emergency", "URGENT"
	case strings.Contains(category, "student"):
		return "student_housing_placement", "HIGH"
	case strings.Contains(category, "starter") && tier == "first_time":
		return "first_time_buyer_education", "MEDIUM"
	case strings.Contains(category, "rental") && urgencyScore > 75:
		return "urgent_tenant_placement", "URGENT"
	case strings.Contains(category, "rental"):
		return "rental_placement_standard", "MEDIUM"
	default:
		return "general_lead_nurture", "MEDIUM"
	}
}

// createBehavioralDeal creates a deal with behavioral intelligence data
func (service *BehavioralFUBIntegrationService) createBehavioralDeal(contactID string, triggerData *PropertyCategoryTriggerData) (*FUBDeal, error) {
	// Calculate deal value based on property category and market intelligence
	dealValue := service.calculateDealValue(triggerData)
	probability := service.calculateCloseProbability(triggerData)
	expectedClose := service.calculateExpectedClose(triggerData)

	deal := &FUBDeal{
		Name:          fmt.Sprintf("%s - %s", triggerData.PropertyCategory, triggerData.Location),
		ContactID:     contactID,
		Value:         dealValue,
		Stage:         service.getDealStageFromBehavioral(triggerData),
		Probability:   probability,
		ExpectedClose: expectedClose,
		Tags:          service.generateDealTags(triggerData),
		CustomFields: map[string]interface{}{
			"property_category":     triggerData.PropertyCategory,
			"behavioral_confidence": service.calculateBehavioralConfidence(triggerData),
			"market_conditions":     triggerData.MarketContext,
			"urgency_factor":        triggerData.BehavioralData["urgency_score"],
			"financial_readiness":   triggerData.BehavioralData["financial_readiness"],
		},
	}

	return service.apiClient.CreateDeal(deal)
}

// assignBehavioralActionPlan assigns the appropriate action plan based on behavioral intelligence
func (service *BehavioralFUBIntegrationService) assignBehavioralActionPlan(contactID string, triggerData *PropertyCategoryTriggerData) (string, error) {
	actionPlanName := service.getActionPlanFromCategory(triggerData)

	// In a real implementation, you'd look up the action plan ID by name
	// For now, we'll use the name as the ID (you'd need to map these properly)
	actionPlanID := strings.ReplaceAll(strings.ToLower(actionPlanName), " ", "_")

	err := service.apiClient.AssignActionPlan(contactID, actionPlanID)
	if err != nil {
		return "", err
	}

	return actionPlanID, nil
}

// assignToBehavioralPond assigns contact to appropriate pond based on behavioral intelligence
func (service *BehavioralFUBIntegrationService) assignToBehavioralPond(contactID string, triggerData *PropertyCategoryTriggerData) (string, error) {
	pondName := service.getPondFromCategory(triggerData)

	// In a real implementation, you'd look up the pond ID by name
	pondID := strings.ReplaceAll(strings.ToLower(pondName), " ", "_")

	err := service.apiClient.AddToPond(contactID, pondID)
	if err != nil {
		return "", err
	}

	return pondID, nil
}

// createImmediateActionTask creates an immediate action task for agents
func (service *BehavioralFUBIntegrationService) createImmediateActionTask(contactID string, triggerData *PropertyCategoryTriggerData, priority string) error {
	responseTime := service.calculateOptimalResponseTime(triggerData, priority)
	taskTitle := service.getTaskTitle(triggerData, priority)
	taskDescription := service.getTaskDescription(triggerData)

	return service.apiClient.CreateTask(contactID, taskTitle, taskDescription, responseTime, priority)
}

// Helper methods for behavioral intelligence mapping

func (service *BehavioralFUBIntegrationService) generateBehavioralTags(triggerData *PropertyCategoryTriggerData) []string {
	tags := []string{
		triggerData.PropertyCategory,
		triggerData.PropertyTier,
		"behavioral_intelligence",
	}

	// Add urgency tags
	if urgency, ok := triggerData.BehavioralData["urgency_score"].(float64); ok {
		if urgency > 80 {
			tags = append(tags, "high_urgency", "immediate_response")
		} else if urgency > 60 {
			tags = append(tags, "medium_urgency")
		}
	}

	// Add financial readiness tags
	if financial, ok := triggerData.BehavioralData["financial_readiness"].(float64); ok {
		if financial > 85 {
			tags = append(tags, "qualified_buyer", "high_financial_readiness")
		} else if financial > 60 {
			tags = append(tags, "moderate_financial_readiness")
		}
	}

	// Add location tags
	if triggerData.Location != "" {
		tags = append(tags, "houston_"+strings.ToLower(triggerData.Location))
	}

	return tags
}

func (service *BehavioralFUBIntegrationService) getSourceFromCategory(category string) string {
	switch {
	case strings.Contains(category, "luxury"):
		return "Luxury Property Intelligence"
	case strings.Contains(category, "investment"):
		return "Investment Property Analysis"
	case strings.Contains(category, "student"):
		return "Student Housing Portal"
	case strings.Contains(category, "rental"):
		return "Rental Property Platform"
	default:
		return "PropertyHub Behavioral Intelligence"
	}
}

func (service *BehavioralFUBIntegrationService) calculateDealValue(triggerData *PropertyCategoryTriggerData) float64 {
	// RENTAL PROPERTIES: Use rental commission structure (rent minus $100 per $1000)
	if strings.Contains(triggerData.PropertyCategory, "rental") || triggerData.PropertyTier == "rental" || strings.Contains(triggerData.PropertyCategory, "student") {
		return service.calculateRentalCommissionValue(triggerData)
	}

	// SALES PROPERTIES: Use sales commission structure (60/40 split with broker)
	propertyValue := service.getSalesPropertyValue(triggerData)
	return service.calculateSalesCommissionValue(propertyValue)
}

// getSalesPropertyValue determines the property value for sales transactions
func (service *BehavioralFUBIntegrationService) getSalesPropertyValue(triggerData *PropertyCategoryTriggerData) float64 {
	switch triggerData.PropertyTier {
	case "luxury":
		switch triggerData.PriceRange {
		case "luxury_3m_plus":
			return 3500000
		case "luxury_2m_3m":
			return 2500000
		case "luxury_1.5m_2m":
			return 1750000
		default:
			return 1500000
		}
	case "investment_grade":
		return 750000 // Average investment property value for SALE
	case "starter_home":
		return 350000 // Houston median home price for SALE
	default:
		return 400000 // Default Houston property value for SALE
	}
}

// calculateSalesCommissionValue calculates agent's actual commission for SALES (60% of total commission)
func (service *BehavioralFUBIntegrationService) calculateSalesCommissionValue(propertyValue float64) float64 {
	// Standard real estate sales commission rate in Houston
	commissionRate := 0.025 // 2.5% of property value

	// Luxury properties might have slightly higher rates
	if propertyValue > 1000000 {
		commissionRate = 0.03 // 3% for luxury sales
	}

	// Calculate total commission on the sale
	totalSalesCommission := propertyValue * commissionRate

	// Agent receives 60% of total commission, broker gets 40%
	agentCommission := totalSalesCommission * 0.60

	log.Printf("🏠 Calculated SALES commission deal value: $%.0f (Property: $%.0f, Total commission: $%.0f @ %.1f%%, Agent's 60%%: $%.0f)",
		agentCommission, propertyValue, totalSalesCommission, commissionRate*100, agentCommission)

	return agentCommission
}

// calculateRentalCommissionValue calculates deal value for RENTAL PROPERTIES ONLY
// RENTAL Commission Structure: Monthly Rent - ($100 for every $1000 in rent)
// This is the FULL commission that goes to the agent (no broker split)
func (service *BehavioralFUBIntegrationService) calculateRentalCommissionValue(triggerData *PropertyCategoryTriggerData) float64 {
	monthlyRent := service.extractMonthlyRent(triggerData)
	if monthlyRent == 0 {
		return 0 // No rental value found
	}

	// RENTAL Commission: Monthly rent - ($100 for every $1000 in rent)
	// Example: $3500/month → $3500 - $350 = $3150 commission
	reductionPer1000 := (monthlyRent / 1000) * 100
	commissionValue := monthlyRent - reductionPer1000

	// Ensure commission is never negative
	if commissionValue < 0 {
		commissionValue = 0
	}

	log.Printf("🏠 Calculated RENTAL commission deal value: $%.0f (Monthly rent: $%.0f - $%.0f reduction)",
		commissionValue, monthlyRent, reductionPer1000)

	return commissionValue
}

// extractMonthlyRent extracts monthly rent from property context and behavioral data
// extractMonthlyRent extracts monthly rent from behavioral data and market context
func (service *BehavioralFUBIntegrationService) extractMonthlyRent(triggerData *PropertyCategoryTriggerData) float64 {
	// Check behavioral data for rent information
	if triggerData.BehavioralData != nil {
		// Try rent_budget first (most common)
		if rent, exists := triggerData.BehavioralData["rent_budget"]; exists {
			if rentFloat, ok := rent.(float64); ok {
				return rentFloat
			}
		}
		// Try monthly_rent
		if rent, exists := triggerData.BehavioralData["monthly_rent"]; exists {
			if rentFloat, ok := rent.(float64); ok {
				return rentFloat
			}
		}
		// Try rent
		if rent, exists := triggerData.BehavioralData["rent"]; exists {
			if rentFloat, ok := rent.(float64); ok {
				return rentFloat
			}
		}
		// Try price (for rental properties)
		if price, exists := triggerData.BehavioralData["price"]; exists {
			if priceFloat, ok := price.(float64); ok {
				// If price is reasonable for monthly rent (under $50K), use it
				if priceFloat < 50000 {
					return priceFloat
				}
			}
		}
	}

	// Check market context for rent information
	if triggerData.MarketContext != nil {
		if rent, exists := triggerData.MarketContext["rent_budget"]; exists {
			if rentFloat, ok := rent.(float64); ok {
				return rentFloat
			}
		}
		if rent, exists := triggerData.MarketContext["monthly_rent"]; exists {
			if rentFloat, ok := rent.(float64); ok {
				return rentFloat
			}
		}
	}

	// Check behavioral contexts for rent information
	if triggerData.BehavioralData != nil {
		for key, value := range triggerData.BehavioralData {
			if strings.Contains(strings.ToLower(key), "rent") {
				if rentFloat, ok := value.(float64); ok {
					return rentFloat
				}
			}
			if strings.Contains(strings.ToLower(key), "budget") {
				if budgetFloat, ok := value.(float64); ok {
					return budgetFloat
				}
			}
		}
	}

	// Default rental values based on property category and location
	return service.getDefaultRentalValue(triggerData)
}

// getDefaultRentalValue provides default rental values based on property category and location
func (service *BehavioralFUBIntegrationService) getDefaultRentalValue(triggerData *PropertyCategoryTriggerData) float64 {
	category := triggerData.PropertyCategory
	location := triggerData.Location

	// Luxury rentals
	if strings.Contains(category, "luxury") {
		if location == "river_oaks" || location == "memorial" {
			return 8000 // Premium luxury rental
		}
		return 5000 // Standard luxury rental
	}

	// Student housing
	if strings.Contains(category, "student") {
		return 1200 // Typical student housing rent
	}

	// Location-based defaults
	switch location {
	case "river_oaks", "memorial":
		return 3500 // High-end areas
	case "downtown", "midtown", "uptown":
		return 2500 // Urban core
	case "heights", "montrose":
		return 2000 // Trendy neighborhoods
	case "katy", "sugar_land", "woodlands":
		return 2200 // Suburban areas
	default:
		return 1800 // Houston general market average
	}
}

func (service *BehavioralFUBIntegrationService) calculateCloseProbability(triggerData *PropertyCategoryTriggerData) float64 {
	baseProbability := 0.5

	// Adjust based on urgency
	if urgency, ok := triggerData.BehavioralData["urgency_score"].(float64); ok {
		baseProbability += (urgency - 50) / 100 * 0.3 // +/- 30% based on urgency
	}

	// Adjust based on financial readiness
	if financial, ok := triggerData.BehavioralData["financial_readiness"].(float64); ok {
		baseProbability += (financial - 50) / 100 * 0.4 // +/- 40% based on financial readiness
	}

	// Adjust based on engagement
	if engagement, ok := triggerData.BehavioralData["engagement_depth"].(float64); ok {
		baseProbability += (engagement - 50) / 100 * 0.2 // +/- 20% based on engagement
	}

	// Clamp between 0.1 and 0.95
	if baseProbability < 0.1 {
		baseProbability = 0.1
	}
	if baseProbability > 0.95 {
		baseProbability = 0.95
	}

	return baseProbability
}

func (service *BehavioralFUBIntegrationService) calculateExpectedClose(triggerData *PropertyCategoryTriggerData) *time.Time {
	now := time.Now()

	// Base timeline based on property category
	var baseDays int
	switch {
	case strings.Contains(triggerData.PropertyCategory, "luxury"):
		baseDays = 60 // Luxury properties take longer
	case strings.Contains(triggerData.PropertyCategory, "investment"):
		baseDays = 45 // Investment properties need analysis time
	case strings.Contains(triggerData.PropertyCategory, "student"):
		baseDays = 14 // Student housing is semester-driven
	case strings.Contains(triggerData.PropertyCategory, "rental"):
		baseDays = 7 // Rentals are fastest
	default:
		baseDays = 30 // Standard home purchase
	}

	// Adjust based on urgency
	if urgency, ok := triggerData.BehavioralData["urgency_score"].(float64); ok {
		if urgency > 80 {
			baseDays = int(float64(baseDays) * 0.7) // 30% faster
		} else if urgency < 40 {
			baseDays = int(float64(baseDays) * 1.5) // 50% slower
		}
	}

	expectedClose := now.AddDate(0, 0, baseDays)
	return &expectedClose
}

func (service *BehavioralFUBIntegrationService) getDealStageFromBehavioral(triggerData *PropertyCategoryTriggerData) string {
	urgency, _ := triggerData.BehavioralData["urgency_score"].(float64)
	financial, _ := triggerData.BehavioralData["financial_readiness"].(float64)
	engagement, _ := triggerData.BehavioralData["engagement_depth"].(float64)

	switch {
	case urgency > 80 && financial > 80:
		return "ready_to_close"
	case urgency > 60 && financial > 70:
		return "showing_scheduled"
	case engagement > 70 && financial > 60:
		return "qualified"
	case engagement > 50:
		return "contact_made"
	default:
		return "new_lead"
	}
}

func (service *BehavioralFUBIntegrationService) generateDealTags(triggerData *PropertyCategoryTriggerData) []string {
	tags := []string{
		triggerData.PropertyCategory,
		"behavioral_intelligence",
	}

	// Add value-based tags
	dealValue := service.calculateDealValue(triggerData)
	if dealValue > 1000000 {
		tags = append(tags, "high_value", "million_plus")
	} else if dealValue > 500000 {
		tags = append(tags, "medium_value")
	}

	return tags
}

func (service *BehavioralFUBIntegrationService) calculateBehavioralConfidence(triggerData *PropertyCategoryTriggerData) float64 {
	// Average the behavioral scores for overall confidence
	urgency, _ := triggerData.BehavioralData["urgency_score"].(float64)
	financial, _ := triggerData.BehavioralData["financial_readiness"].(float64)
	engagement, _ := triggerData.BehavioralData["engagement_depth"].(float64)

	return (urgency + financial + engagement) / 3.0
}

func (service *BehavioralFUBIntegrationService) getActionPlanFromCategory(triggerData *PropertyCategoryTriggerData) string {
	category := triggerData.PropertyCategory
	urgency, _ := triggerData.BehavioralData["urgency_score"].(float64)

	switch {
	case strings.Contains(category, "luxury") && urgency > 80:
		return "luxury_buyer_immediate_plan"
	case strings.Contains(category, "luxury"):
		return "luxury_rental_concierge_plan"
	case strings.Contains(category, "investment"):
		return "investor_portfolio_analysis_plan"
	case strings.Contains(category, "student"):
		return "student_housing_semester_plan"
	case strings.Contains(category, "starter") && triggerData.TargetDemo == "first_time_buyer":
		return "first_time_buyer_education_plan"
	case strings.Contains(category, "rental") && urgency > 75:
		return "emergency_placement_plan"
	case strings.Contains(category, "rental"):
		return "starter_rental_assistance_plan"
	default:
		return "general_lead_nurture_plan"
	}
}

func (service *BehavioralFUBIntegrationService) getPondFromCategory(triggerData *PropertyCategoryTriggerData) string {
	category := triggerData.PropertyCategory
	location := triggerData.Location
	urgency, _ := triggerData.BehavioralData["urgency_score"].(float64)

	switch {
	case strings.Contains(category, "luxury") && (location == "river_oaks" || location == "memorial"):
		return "Luxury Buyers - River Oaks/Memorial"
	case strings.Contains(category, "luxury"):
		return "Luxury Rentals - High-End"
	case strings.Contains(category, "investment"):
		return "Investment Properties - Portfolio Builders"
	case strings.Contains(category, "student"):
		return "Student Housing - Semester Rush"
	case triggerData.TargetDemo == "first_time_buyer":
		return "First-Time Buyers - Education Track"
	case urgency > 80:
		return "Emergency Housing Needs"
	case strings.Contains(category, "rental") && triggerData.PriceRange == "under_1500":
		return "Starter Rentals - Budget Focus"
	default:
		return "General Leads - Behavioral Intelligence"
	}
}

func (service *BehavioralFUBIntegrationService) calculateOptimalResponseTime(triggerData *PropertyCategoryTriggerData, priority string) time.Time {
	now := time.Now()

	// Base response times based on priority and category
	var minutes int
	switch priority {
	case "URGENT":
		if strings.Contains(triggerData.PropertyCategory, "rental") {
			minutes = 5 // Urgent tenant placement
		} else {
			minutes = 10 // Urgent luxury buyers
		}
	case "HIGH":
		if strings.Contains(triggerData.PropertyCategory, "luxury") {
			minutes = 15 // Luxury service standard
		} else {
			minutes = 30 // High priority standard
		}
	case "MEDIUM":
		if strings.Contains(triggerData.PropertyCategory, "investment") {
			minutes = 45 // Investment analysis prep time
		} else {
			minutes = 60 // Standard response
		}
	default:
		minutes = 120 // Low priority
	}

	return now.Add(time.Duration(minutes) * time.Minute)
}

func (service *BehavioralFUBIntegrationService) getTaskTitle(triggerData *PropertyCategoryTriggerData, priority string) string {
	category := triggerData.PropertyCategory
	location := triggerData.Location

	switch priority {
	case "URGENT":
		if strings.Contains(category, "rental") {
			return "🚨 URGENT: Tenant placement required - call within 5 minutes"
		} else {
			return "🚨 URGENT: High-intent buyer - call within 10 minutes"
		}
	case "HIGH":
		if strings.Contains(category, "luxury") {
			return "💎 LUXURY: Premium client consultation required"
		} else {
			return "⭐ HIGH PRIORITY: Qualified lead follow-up"
		}
	default:
		return fmt.Sprintf("📞 Follow up: %s lead in %s", category, location)
	}
}

func (service *BehavioralFUBIntegrationService) getTaskDescription(triggerData *PropertyCategoryTriggerData) string {
	urgency, _ := triggerData.BehavioralData["urgency_score"].(float64)
	financial, _ := triggerData.BehavioralData["financial_readiness"].(float64)
	engagement, _ := triggerData.BehavioralData["engagement_depth"].(float64)

	description := fmt.Sprintf(
		"Behavioral Intelligence Analysis:\n"+
			"• Property Category: %s\n"+
			"• Property Tier: %s\n"+
			"• Target Demo: %s\n"+
			"• Location: %s\n"+
			"• Urgency Score: %.1f/100\n"+
			"• Financial Readiness: %.1f/100\n"+
			"• Engagement Depth: %.1f/100\n\n",
		triggerData.PropertyCategory,
		triggerData.PropertyTier,
		triggerData.TargetDemo,
		triggerData.Location,
		urgency,
		financial,
		engagement,
	)

	// Add specific recommendations
	description += "Recommended Approach:\n"
	if urgency > 80 {
		description += "• URGENT response required - lead has time pressure\n"
	}
	if financial > 80 {
		description += "• Highly qualified buyer - ready to proceed\n"
	}
	if engagement > 80 {
		description += "• Highly engaged - detailed information available\n"
	}

	if strings.Contains(triggerData.PropertyCategory, "luxury") {
		description += "• Use luxury specialist and premium service protocols\n"
	}
	if strings.Contains(triggerData.PropertyCategory, "investment") {
		description += "• Prepare ROI analysis and cash flow projections\n"
	}

	return description
}

func (service *BehavioralFUBIntegrationService) getRecommendedAction(triggerData *PropertyCategoryTriggerData) string {
	category := triggerData.PropertyCategory
	urgency, _ := triggerData.BehavioralData["urgency_score"].(float64)

	switch {
	case strings.Contains(category, "luxury") && urgency > 80:
		return "Immediate luxury specialist assignment with private consultation scheduling"
	case strings.Contains(category, "luxury"):
		return "Premium market analysis preparation and exclusive listing access"
	case strings.Contains(category, "investment"):
		return "Investment strategy consultation and ROI analysis preparation"
	case strings.Contains(category, "student"):
		return "University proximity property search and group viewing coordination"
	case urgency > 75:
		return "Emergency placement protocols - immediate response required"
	default:
		return "Standard lead nurture with behavioral intelligence monitoring"
	}
}
