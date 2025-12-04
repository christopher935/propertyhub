package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// BehavioralFUBBridge connects behavioral intelligence handlers to FUB API integration
type BehavioralFUBBridge struct {
	db                 *gorm.DB
	integrationService *BehavioralFUBIntegrationService
}

// NewBehavioralFUBBridge creates a new bridge service
func NewBehavioralFUBBridge(db *gorm.DB, apiKey string) *BehavioralFUBBridge {
	return &BehavioralFUBBridge{
		db:                 db,
		integrationService: NewBehavioralFUBIntegrationService(db, apiKey),
	}
}

// BehavioralTriggerRequest represents incoming behavioral trigger from handlers
type BehavioralTriggerRequest struct {
	SessionID             string                 `json:"session_id"`
	Email                 string                 `json:"email"`
	Phone                 string                 `json:"phone"`
	Name                  string                 `json:"name"`
	PropertyID            int                    `json:"property_id"`
	TriggerType           string                 `json:"trigger_type"`
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

// ProcessBehavioralTriggerForFUB processes behavioral intelligence triggers and executes FUB operations
func (bridge *BehavioralFUBBridge) ProcessBehavioralTriggerForFUB(request *BehavioralTriggerRequest) (*BehavioralTriggerResult, error) {
	log.Printf("🌉 Processing behavioral trigger for FUB integration: %s", request.TriggerType)

	// Convert behavioral trigger request to property category trigger data
	triggerData := bridge.convertToPropertyCategoryData(request)

	// Extract contact information
	contactInfo := map[string]string{
		"name":  request.Name,
		"email": request.Email,
		"phone": request.Phone,
	}

	// Process through the behavioral integration service
	return bridge.integrationService.ProcessBehavioralTrigger(triggerData, contactInfo)
}

// convertToPropertyCategoryData converts behavioral trigger request to property category trigger data
func (bridge *BehavioralFUBBridge) convertToPropertyCategoryData(request *BehavioralTriggerRequest) *PropertyCategoryTriggerData {
	// Determine property category based on behavioral context and property type
	propertyCategory := bridge.determinePropertyCategory(request)
	propertyTier := bridge.determinePropertyTier(request)
	targetDemo := bridge.determineTargetDemographic(request)
	priceRange := bridge.determinePriceRange(request)
	location := bridge.extractLocationFromContext(request)

	// Combine behavioral data from rental and sales contexts
	behavioralData := make(map[string]interface{})
	behavioralData["urgency_score"] = request.UrgencyScore
	behavioralData["financial_readiness"] = request.FinancialQualScore
	behavioralData["engagement_depth"] = request.EngagementScore

	// Add context-specific behavioral data
	for key, value := range request.RentalBehaviorContext {
		behavioralData["rental_"+key] = value
	}
	for key, value := range request.SalesBehaviorContext {
		behavioralData["sales_"+key] = value
	}

	// Extract market context
	marketContext := bridge.extractMarketContext(request)

	return &PropertyCategoryTriggerData{
		PropertyCategory: propertyCategory,
		PropertyTier:     propertyTier,
		TargetDemo:       targetDemo,
		PriceRange:       priceRange,
		Location:         location,
		BehavioralData:   behavioralData,
		MarketContext:    marketContext,
	}
}

// determinePropertyCategory determines the property category from behavioral context
func (bridge *BehavioralFUBBridge) determinePropertyCategory(request *BehavioralTriggerRequest) string {
	// Check property context for explicit category
	if category, exists := request.PropertyContext["category"]; exists {
		if categoryStr, ok := category.(string); ok {
			return categoryStr
		}
	}

	// Determine from lead type and property type
	leadType := strings.ToLower(request.LeadType)

	// Check for luxury indicators
	isLuxury := bridge.isLuxuryProperty(request)

	// Check for investment indicators
	isInvestment := bridge.isInvestmentProperty(request)

	// Check for student housing indicators
	isStudent := bridge.isStudentHousing(request)

	switch {
	case isLuxury && (leadType == "buyer" || leadType == "seller"):
		return "luxury_sales"
	case isLuxury && (leadType == "tenant" || leadType == "landlord"):
		return "luxury_rental"
	case isInvestment:
		return "investment_property"
	case isStudent:
		return "student_housing"
	case leadType == "buyer" || leadType == "seller":
		return "starter_home"
	case leadType == "tenant" || leadType == "landlord":
		if bridge.isStarterRental(request) {
			return "starter_rental"
		}
		return "family_rental"
	default:
		return "general_property"
	}
}

// determinePropertyTier determines the property tier from context and behavioral signals
func (bridge *BehavioralFUBBridge) determinePropertyTier(request *BehavioralTriggerRequest) string {
	if bridge.isLuxuryProperty(request) {
		return "luxury"
	}
	if bridge.isInvestmentProperty(request) {
		return "investment_grade"
	}
	if bridge.isFirstTimeBuyer(request) {
		return "starter"
	}
	return "standard"
}

// determineTargetDemographic determines the target demographic from behavioral signals
func (bridge *BehavioralFUBBridge) determineTargetDemographic(request *BehavioralTriggerRequest) string {
	// Check for explicit demographic indicators in contexts
	if demo, exists := request.PropertyContext["target_demographic"]; exists {
		if demoStr, ok := demo.(string); ok {
			return demoStr
		}
	}

	// Infer from behavioral patterns and contexts
	if bridge.isFirstTimeBuyer(request) {
		return "first_time_buyer"
	}
	if bridge.isStudentHousing(request) {
		return "student"
	}
	if bridge.isLuxuryProperty(request) {
		return "affluent_professional"
	}
	if bridge.isInvestmentProperty(request) {
		return "investor"
	}
	if bridge.isYoungProfessional(request) {
		return "young_professional"
	}

	return "general_consumer"
}

// determinePriceRange determines the price range from property context
func (bridge *BehavioralFUBBridge) determinePriceRange(request *BehavioralTriggerRequest) string {
	// Check property context for price information
	if price, exists := request.PropertyContext["price"]; exists {
		if priceFloat, ok := price.(float64); ok {
			return bridge.categorizePriceRange(priceFloat)
		}
		if priceStr, ok := price.(string); ok {
			if priceFloat, err := strconv.ParseFloat(priceStr, 64); err == nil {
				return bridge.categorizePriceRange(priceFloat)
			}
		}
	}

	// Infer from property category
	if bridge.isLuxuryProperty(request) {
		return "luxury_1.5m_plus"
	}
	if bridge.isInvestmentProperty(request) {
		return "investment_500k_1m"
	}
	if bridge.isStarterRental(request) {
		return "under_1500"
	}

	return "market_rate"
}

// extractLocationFromContext extracts location information from contexts
func (bridge *BehavioralFUBBridge) extractLocationFromContext(request *BehavioralTriggerRequest) string {
	// Check property context for location
	if location, exists := request.PropertyContext["location"]; exists {
		if locationStr, ok := location.(string); ok {
			return bridge.normalizeHoustonLocation(locationStr)
		}
	}
	if city, exists := request.PropertyContext["city"]; exists {
		if cityStr, ok := city.(string); ok {
			return bridge.normalizeHoustonLocation(cityStr)
		}
	}
	if neighborhood, exists := request.PropertyContext["neighborhood"]; exists {
		if neighborhoodStr, ok := neighborhood.(string); ok {
			return bridge.normalizeHoustonLocation(neighborhoodStr)
		}
	}

	return "houston_general"
}

// extractMarketContext extracts market context information
func (bridge *BehavioralFUBBridge) extractMarketContext(request *BehavioralTriggerRequest) map[string]interface{} {
	marketContext := make(map[string]interface{})

	// Add timeline context as market context
	for key, value := range request.TimelineContext {
		marketContext[key] = value
	}

	// Add seasonal factors (this could be enhanced with real market data)
	marketContext["seasonal_factor"] = bridge.getCurrentSeasonalFactor()
	marketContext["market_conditions"] = "normal" // Could be enhanced with real market intelligence
	marketContext["inventory_level"] = "medium"   // Could be enhanced with real market data

	return marketContext
}

// Helper methods for property classification

func (bridge *BehavioralFUBBridge) isLuxuryProperty(request *BehavioralTriggerRequest) bool {
	// Check for luxury indicators in property context
	if luxury, exists := request.PropertyContext["luxury"]; exists {
		if luxuryBool, ok := luxury.(bool); ok && luxuryBool {
			return true
		}
	}

	// Check price thresholds
	if price, exists := request.PropertyContext["price"]; exists {
		if priceFloat, ok := price.(float64); ok && priceFloat > 1000000 {
			return true
		}
	}

	// Check for luxury amenities or neighborhoods
	if amenities, exists := request.PropertyContext["amenities"]; exists {
		amenitiesStr := fmt.Sprintf("%v", amenities)
		luxuryKeywords := []string{"pool", "golf", "concierge", "valet", "wine", "spa"}
		for _, keyword := range luxuryKeywords {
			if strings.Contains(strings.ToLower(amenitiesStr), keyword) {
				return true
			}
		}
	}

	// Check financial readiness - high financial qualification might indicate luxury
	return request.FinancialQualScore > 90
}

func (bridge *BehavioralFUBBridge) isInvestmentProperty(request *BehavioralTriggerRequest) bool {
	// Check explicit investment indicators
	if investment, exists := request.PropertyContext["investment"]; exists {
		if investmentBool, ok := investment.(bool); ok && investmentBool {
			return true
		}
	}

	// Check behavioral contexts for investment patterns
	if request.SalesBehaviorContext != nil {
		for key := range request.SalesBehaviorContext {
			if strings.Contains(strings.ToLower(key), "investment") ||
				strings.Contains(strings.ToLower(key), "roi") ||
				strings.Contains(strings.ToLower(key), "cash_flow") {
				return true
			}
		}
	}

	// Check if lead type suggests investment
	return strings.ToLower(request.LeadType) == "investor"
}

func (bridge *BehavioralFUBBridge) isStudentHousing(request *BehavioralTriggerRequest) bool {
	// Check property context for student indicators
	if student, exists := request.PropertyContext["student_housing"]; exists {
		if studentBool, ok := student.(bool); ok && studentBool {
			return true
		}
	}

	// Check rental context for student patterns
	if request.RentalBehaviorContext != nil {
		for key := range request.RentalBehaviorContext {
			if strings.Contains(strings.ToLower(key), "student") ||
				strings.Contains(strings.ToLower(key), "university") ||
				strings.Contains(strings.ToLower(key), "college") {
				return true
			}
		}
	}

	return false
}

func (bridge *BehavioralFUBBridge) isFirstTimeBuyer(request *BehavioralTriggerRequest) bool {
	// Check sales context for first-time buyer indicators
	if request.SalesBehaviorContext != nil {
		if firstTime, exists := request.SalesBehaviorContext["first_time_buyer"]; exists {
			if firstTimeBool, ok := firstTime.(bool); ok && firstTimeBool {
				return true
			}
		}

		// Check for first-time buyer behavioral patterns
		for key := range request.SalesBehaviorContext {
			if strings.Contains(strings.ToLower(key), "first_time") ||
				strings.Contains(strings.ToLower(key), "first_home") {
				return true
			}
		}
	}

	// Low financial qualification with high engagement might indicate first-time buyer needing education
	return request.FinancialQualScore < 60 && request.EngagementScore > 70
}

func (bridge *BehavioralFUBBridge) isStarterRental(request *BehavioralTriggerRequest) bool {
	// Check rental context for budget indicators
	if request.RentalBehaviorContext != nil {
		if budget, exists := request.RentalBehaviorContext["budget"]; exists {
			if budgetFloat, ok := budget.(float64); ok && budgetFloat < 1500 {
				return true
			}
		}

		// Check for starter rental keywords
		for key := range request.RentalBehaviorContext {
			if strings.Contains(strings.ToLower(key), "budget") ||
				strings.Contains(strings.ToLower(key), "affordable") ||
				strings.Contains(strings.ToLower(key), "starter") {
				return true
			}
		}
	}

	return false
}

func (bridge *BehavioralFUBBridge) isYoungProfessional(request *BehavioralTriggerRequest) bool {
	// Check contexts for young professional indicators
	contexts := []map[string]interface{}{
		request.RentalBehaviorContext,
		request.SalesBehaviorContext,
		request.PropertyContext,
	}

	for _, context := range contexts {
		if context != nil {
			for key := range context {
				if strings.Contains(strings.ToLower(key), "young") ||
					strings.Contains(strings.ToLower(key), "professional") ||
					strings.Contains(strings.ToLower(key), "career") {
					return true
				}
			}
		}
	}

	// High engagement with moderate financial readiness might indicate young professional
	return request.EngagementScore > 75 &&
		request.FinancialQualScore > 60 &&
		request.FinancialQualScore < 85
}

func (bridge *BehavioralFUBBridge) categorizePriceRange(price float64) string {
	switch {
	case price >= 3000000:
		return "luxury_3m_plus"
	case price >= 2000000:
		return "luxury_2m_3m"
	case price >= 1500000:
		return "luxury_1.5m_2m"
	case price >= 1000000:
		return "luxury_1m_1.5m"
	case price >= 750000:
		return "investment_750k_1m"
	case price >= 500000:
		return "investment_500k_750k"
	case price >= 350000:
		return "market_350k_500k"
	case price >= 250000:
		return "starter_250k_350k"
	case price >= 1500: // Monthly rent
		return "rental_1500_plus"
	case price >= 1000:
		return "rental_1000_1500"
	case price < 1000:
		return "under_1000"
	default:
		return "market_rate"
	}
}

func (bridge *BehavioralFUBBridge) normalizeHoustonLocation(location string) string {
	locationLower := strings.ToLower(location)

	// Map various location names to standardized Houston locations
	locationMap := map[string]string{
		"river oaks":      "river_oaks",
		"riveroaks":       "river_oaks",
		"memorial":        "memorial",
		"memorial city":   "memorial",
		"highlands":       "heights",
		"houston heights": "heights",
		"the heights":     "heights",
		"montrose":        "montrose",
		"midtown":         "midtown",
		"downtown":        "downtown",
		"uptown":          "uptown",
		"galleria":        "galleria",
		"west university": "west_university",
		"rice village":    "rice_village",
		"museum district": "museum_district",
		"medical center":  "medical_center",
		"sugar land":      "sugar_land",
		"the woodlands":   "woodlands",
		"katy":            "katy",
		"pearland":        "pearland",
		"friendswood":     "friendswood",
		"clear lake":      "clear_lake",
		"brookshire":      "brookshire",
		"waller":          "waller",
	}

	for key, value := range locationMap {
		if strings.Contains(locationLower, key) {
			return value
		}
	}

	return "houston_general"
}

func (bridge *BehavioralFUBBridge) getCurrentSeasonalFactor() float64 {
	// This could be enhanced with real seasonal market data
	// For now, return a base seasonal factor
	return 1.0 // No seasonal adjustment
}

// IntegrateBehavioralTriggerWithFUB is the main entry point for behavioral intelligence → FUB integration
func (bridge *BehavioralFUBBridge) IntegrateBehavioralTriggerWithFUB(triggerJSON []byte) (*BehavioralTriggerResult, error) {
	var request BehavioralTriggerRequest

	if err := json.Unmarshal(triggerJSON, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trigger request: %v", err)
	}

	return bridge.ProcessBehavioralTriggerForFUB(&request)
}
