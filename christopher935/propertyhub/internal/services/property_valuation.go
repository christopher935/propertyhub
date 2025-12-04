package services

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"chrisgross-ctrl-project/internal/config"
        "chrisgross-ctrl-project/internal/scraper"
)

// PropertyValuationService provides AI-powered property valuation for pre-listings
type PropertyValuationService struct {
	config          *config.Config
	scraperService  *scraper.ScraperService
	marketDataCache map[string]*MarketData
	cacheTTL        time.Duration
}

// PropertyValuationRequest represents a valuation request
type PropertyValuationRequest struct {
	Address         string  `json:"address"`
	City            string  `json:"city"`
	ZipCode         string  `json:"zip_code"`
	SquareFeet      int     `json:"square_feet"`
	Bedrooms        int     `json:"bedrooms"`
	Bathrooms       float32 `json:"bathrooms"`
	PropertyType    string  `json:"property_type"` // single_family, townhome, condo, etc.
	YearBuilt       int     `json:"year_built"`
	LotSize         float32 `json:"lot_size,omitempty"`
	Garage          int     `json:"garage,omitempty"`
	Pool            bool    `json:"pool,omitempty"`
	Fireplace       bool    `json:"fireplace,omitempty"`
	UpdatedKitchen  bool    `json:"updated_kitchen,omitempty"`
	UpdatedBathroom bool    `json:"updated_bathroom,omitempty"`
	HardwoodFloors  bool    `json:"hardwood_floors,omitempty"`
}

// PropertyValuation represents the valuation result
type PropertyValuation struct {
	EstimatedValue   int                     `json:"estimated_value"`
	ValueRange       ValueRange              `json:"value_range"`
	PricePerSqFt     float32                 `json:"price_per_sqft"`
	ConfidenceScore  float32                 `json:"confidence_score"` // 0.0 to 1.0
	MarketConditions MarketConditions        `json:"market_conditions"`
	Comparables      []ComparableProperty    `json:"comparables"`
	ValuationFactors []ValuationFactor       `json:"valuation_factors"`
	Recommendations  []PricingRecommendation `json:"recommendations"`
	LastUpdated      time.Time               `json:"last_updated"`
}

// ValueRange represents the estimated value range
type ValueRange struct {
	Low    int `json:"low"`
	High   int `json:"high"`
	Median int `json:"median"`
}

// MarketConditions represents current market state
type MarketConditions struct {
	MarketTrend        string  `json:"market_trend"`         // rising, stable, declining
	DaysOnMarket       int     `json:"days_on_market"`       // average for area
	PriceChangePercent float32 `json:"price_change_percent"` // vs last quarter
	InventoryLevel     string  `json:"inventory_level"`      // low, normal, high
	SeasonalAdjustment float32 `json:"seasonal_adjustment"`  // seasonal price modifier
}

// ComparableProperty represents a comparable sale
type ComparableProperty struct {
	Address         string    `json:"address"`
	Distance        float32   `json:"distance_miles"`
	SalePrice       int       `json:"sale_price"`
	SaleDate        time.Time `json:"sale_date"`
	SquareFeet      int       `json:"square_feet"`
	Bedrooms        int       `json:"bedrooms"`
	Bathrooms       float32   `json:"bathrooms"`
	YearBuilt       int       `json:"year_built"`
	PricePerSqFt    float32   `json:"price_per_sqft"`
	AdjustedPrice   int       `json:"adjusted_price"`   // adjusted for differences
	SimilarityScore float32   `json:"similarity_score"` // 0.0 to 1.0
}

// ValuationFactor represents factors affecting valuation
type ValuationFactor struct {
	Factor      string  `json:"factor"`
	Impact      string  `json:"impact"`     // positive, negative, neutral
	Adjustment  float32 `json:"adjustment"` // percentage adjustment
	Description string  `json:"description"`
}

// PricingRecommendation provides pricing strategy advice
type PricingRecommendation struct {
	Strategy   string `json:"strategy"`
	PricePoint int    `json:"price_point"`
	Reasoning  string `json:"reasoning"`
	Confidence string `json:"confidence"` // high, medium, low
}

// MarketData represents cached market information
type MarketData struct {
	ZipCode             string    `json:"zip_code"`
	MedianPrice         int       `json:"median_price"`
	AveragePricePerSqFt float32   `json:"average_price_per_sqft"`
	DaysOnMarket        int       `json:"days_on_market"`
	PriceGrowth         float32   `json:"price_growth_percent"`
	InventoryCount      int       `json:"inventory_count"`
	LastUpdated         time.Time `json:"last_updated"`
}

// NewPropertyValuationService creates a new property valuation service
func NewPropertyValuationService(config *config.Config, scraperService *scraper.ScraperService) *PropertyValuationService {
	return &PropertyValuationService{
		config:          config,
		scraperService:  scraperService,
		marketDataCache: make(map[string]*MarketData),
		cacheTTL:        24 * time.Hour, // Cache market data for 24 hours
	}
}

// ValuateProperty performs comprehensive property valuation
func (pvs *PropertyValuationService) ValuateProperty(request PropertyValuationRequest) (*PropertyValuation, error) {
	log.Printf("🏠 Starting property valuation for %s", request.Address)

	// Get market data for the area
	marketData, err := pvs.getMarketData(request.ZipCode)
	if err != nil {
		log.Printf("Warning: Could not get market data: %v", err)
		// Continue with default market data
		marketData = pvs.getDefaultMarketData(request.ZipCode)
	}

	// Find comparable properties
	comparables, err := pvs.findComparables(request)
	if err != nil {
		log.Printf("Warning: Could not find comparables: %v", err)
		// Use market data for basic estimation
	}

	// Calculate base valuation using comparables
	baseValue := pvs.calculateBaseValue(request, comparables, marketData)

	// Apply adjustments based on property features
	adjustedValue, factors := pvs.applyFeatureAdjustments(baseValue, request)

	// Calculate confidence score
	confidence := pvs.calculateConfidenceScore(len(comparables), marketData)

	// Generate value range
	valueRange := pvs.calculateValueRange(adjustedValue, confidence)

	// Determine market conditions
	marketConditions := pvs.assessMarketConditions(marketData)

	// Generate pricing recommendations
	recommendations := pvs.generatePricingRecommendations(adjustedValue, marketConditions, confidence)

	valuation := &PropertyValuation{
		EstimatedValue:   adjustedValue,
		ValueRange:       valueRange,
		PricePerSqFt:     float32(adjustedValue) / float32(request.SquareFeet),
		ConfidenceScore:  confidence,
		MarketConditions: marketConditions,
		Comparables:      comparables,
		ValuationFactors: factors,
		Recommendations:  recommendations,
		LastUpdated:      time.Now(),
	}

	log.Printf("🎯 Property valuation complete: $%d (confidence: %.2f)", adjustedValue, confidence)
	return valuation, nil
}

// getMarketData retrieves market data for a specific zip code
func (pvs *PropertyValuationService) getMarketData(zipCode string) (*MarketData, error) {
	// Check cache first
	if cached, exists := pvs.marketDataCache[zipCode]; exists {
		if time.Since(cached.LastUpdated) < pvs.cacheTTL {
			return cached, nil
		}
	}

	// In production, this would fetch from HAR MLS API
	// For now, simulate with Houston-specific data
	marketData := pvs.getHoustonMarketData(zipCode)

	// Cache the result
	pvs.marketDataCache[zipCode] = marketData

	return marketData, nil
}

// findComparables finds comparable properties using scraper service
func (pvs *PropertyValuationService) findComparables(request PropertyValuationRequest) ([]ComparableProperty, error) {
	// For development, use mock data
	// In production, would use scraper service with parameters:
	// Location: fmt.Sprintf("%s, %s", request.City, request.ZipCode)
	// MinBedrooms: max(1, request.Bedrooms-1)
	// MaxBedrooms: request.Bedrooms + 1
	// MinBathrooms: maxFloat(1.0, request.Bathrooms-0.5)
	// MaxBathrooms: request.Bathrooms + 0.5
	comparables := pvs.generateMockComparables(request)

	// Sort by similarity score
	sort.Slice(comparables, func(i, j int) bool {
		return comparables[i].SimilarityScore > comparables[j].SimilarityScore
	})

	// Return top 10 comparables
	if len(comparables) > 10 {
		comparables = comparables[:10]
	}

	return comparables, nil
}

// calculateBaseValue calculates base property value from comparables
func (pvs *PropertyValuationService) calculateBaseValue(request PropertyValuationRequest, comparables []ComparableProperty, marketData *MarketData) int {
	if len(comparables) == 0 {
		// Fallback to market data
		return int(marketData.AveragePricePerSqFt * float32(request.SquareFeet))
	}

	// Weight comparables by similarity and recency
	var weightedSum float64
	var totalWeight float64

	for _, comp := range comparables {
		// Calculate age weight (more recent = higher weight)
		daysSinceSale := time.Since(comp.SaleDate).Hours() / 24
		ageWeight := math.Max(0.1, 1.0-(daysSinceSale/365.0)*0.3) // Reduce weight by 30% per year

		// Calculate similarity weight
		similarityWeight := float64(comp.SimilarityScore)

		// Combined weight
		weight := ageWeight * similarityWeight

		weightedSum += float64(comp.AdjustedPrice) * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return int(marketData.AveragePricePerSqFt * float32(request.SquareFeet))
	}

	return int(weightedSum / totalWeight)
}

// applyFeatureAdjustments applies adjustments based on property features
func (pvs *PropertyValuationService) applyFeatureAdjustments(baseValue int, request PropertyValuationRequest) (int, []ValuationFactor) {
	adjustedValue := float64(baseValue)
	var factors []ValuationFactor

	// Age adjustment
	currentYear := time.Now().Year()
	propertyAge := currentYear - request.YearBuilt

	var ageAdjustment float64
	var ageImpact string

	if propertyAge < 5 {
		ageAdjustment = 0.05 // 5% premium for new construction
		ageImpact = "positive"
	} else if propertyAge > 30 {
		ageAdjustment = -0.02 // 2% discount for older homes
		ageImpact = "negative"
	} else {
		ageAdjustment = 0
		ageImpact = "neutral"
	}

	if ageAdjustment != 0 {
		adjustedValue *= (1 + ageAdjustment)
		factors = append(factors, ValuationFactor{
			Factor:      "Property Age",
			Impact:      ageImpact,
			Adjustment:  float32(ageAdjustment * 100),
			Description: fmt.Sprintf("Built in %d (%d years old)", request.YearBuilt, propertyAge),
		})
	}

	// Feature premiums
	featureAdjustments := []struct {
		hasFeature  bool
		adjustment  float64
		description string
	}{
		{request.Pool, 0.03, "Swimming pool adds value"},
		{request.Fireplace, 0.015, "Fireplace adds comfort value"},
		{request.UpdatedKitchen, 0.04, "Updated kitchen increases appeal"},
		{request.UpdatedBathroom, 0.025, "Updated bathroom improves value"},
		{request.HardwoodFloors, 0.02, "Hardwood floors premium"},
		{request.Garage >= 2, 0.015, "Two-car garage convenience"},
	}

	for _, adj := range featureAdjustments {
		if adj.hasFeature {
			adjustedValue *= (1 + adj.adjustment)
			factors = append(factors, ValuationFactor{
				Factor:      adj.description,
				Impact:      "positive",
				Adjustment:  float32(adj.adjustment * 100),
				Description: adj.description,
			})
		}
	}

	return int(adjustedValue), factors
}

// calculateConfidenceScore calculates confidence in the valuation
func (pvs *PropertyValuationService) calculateConfidenceScore(comparablesCount int, marketData *MarketData) float32 {
	baseConfidence := float32(0.5)

	// More comparables = higher confidence
	comparableBonus := float32(comparablesCount) * 0.05
	if comparableBonus > 0.3 {
		comparableBonus = 0.3 // Cap at 30%
	}

	// Recent market data = higher confidence
	dataAge := time.Since(marketData.LastUpdated).Hours() / 24
	dataConfidence := float32(math.Max(0.0, 1.0-dataAge/30.0) * 0.2) // Up to 20% bonus for recent data

	confidence := baseConfidence + comparableBonus + dataConfidence

	// Cap at 95%
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

// calculateValueRange calculates the estimated value range
func (pvs *PropertyValuationService) calculateValueRange(estimatedValue int, confidence float32) ValueRange {
	// Lower confidence = wider range
	rangePercent := (1.0 - confidence) * 0.2 // 0% to 20% range

	rangeDollar := int(float32(estimatedValue) * rangePercent)

	return ValueRange{
		Low:    estimatedValue - rangeDollar,
		High:   estimatedValue + rangeDollar,
		Median: estimatedValue,
	}
}

// assessMarketConditions determines current market conditions
func (pvs *PropertyValuationService) assessMarketConditions(marketData *MarketData) MarketConditions {
	var trend string
	if marketData.PriceGrowth > 5 {
		trend = "rising"
	} else if marketData.PriceGrowth < -2 {
		trend = "declining"
	} else {
		trend = "stable"
	}

	var inventoryLevel string
	if marketData.InventoryCount < 100 {
		inventoryLevel = "low"
	} else if marketData.InventoryCount > 300 {
		inventoryLevel = "high"
	} else {
		inventoryLevel = "normal"
	}

	// Seasonal adjustment (Houston market patterns)
	month := time.Now().Month()
	var seasonalAdjustment float32
	switch {
	case month >= 4 && month <= 6: // Spring buying season
		seasonalAdjustment = 0.02
	case month >= 7 && month <= 8: // Summer
		seasonalAdjustment = 0.01
	case month >= 11 || month <= 2: // Winter slowdown
		seasonalAdjustment = -0.01
	default:
		seasonalAdjustment = 0
	}

	return MarketConditions{
		MarketTrend:        trend,
		DaysOnMarket:       marketData.DaysOnMarket,
		PriceChangePercent: marketData.PriceGrowth,
		InventoryLevel:     inventoryLevel,
		SeasonalAdjustment: seasonalAdjustment,
	}
}

// generatePricingRecommendations generates strategic pricing recommendations
func (pvs *PropertyValuationService) generatePricingRecommendations(estimatedValue int, conditions MarketConditions, confidence float32) []PricingRecommendation {
	var recommendations []PricingRecommendation

	// Conservative pricing
	conservativePrice := int(float32(estimatedValue) * 0.97)
	recommendations = append(recommendations, PricingRecommendation{
		Strategy:   "Conservative",
		PricePoint: conservativePrice,
		Reasoning:  "Price slightly below market value for quick sale and multiple offers",
		Confidence: "high",
	})

	// Market pricing
	recommendations = append(recommendations, PricingRecommendation{
		Strategy:   "Market Value",
		PricePoint: estimatedValue,
		Reasoning:  "Price at estimated market value based on comparable sales",
		Confidence: getConfidenceLevel(confidence),
	})

	// Aggressive pricing (only in rising markets)
	if conditions.MarketTrend == "rising" {
		aggressivePrice := int(float32(estimatedValue) * 1.05)
		recommendations = append(recommendations, PricingRecommendation{
			Strategy:   "Aggressive",
			PricePoint: aggressivePrice,
			Reasoning:  "Price above market in rising market - test demand ceiling",
			Confidence: "medium",
		})
	}

	return recommendations
}

// Helper functions

func (pvs *PropertyValuationService) getHoustonMarketData(zipCode string) *MarketData {
	// Simulate Houston market data with some variation by zip code
	basePrice := 300000
	basePricePerSqFt := float32(150)

	// Adjust based on zip code (simplified)
	switch {
	case zipCode >= "77001" && zipCode <= "77019": // Inner loop/Downtown
		basePrice = 450000
		basePricePerSqFt = 225
	case zipCode >= "77025" && zipCode <= "77096": // Galleria/River Oaks area
		basePrice = 550000
		basePricePerSqFt = 275
	case zipCode >= "77379" && zipCode <= "77389": // Spring/The Woodlands
		basePrice = 380000
		basePricePerSqFt = 190
	}

	return &MarketData{
		ZipCode:             zipCode,
		MedianPrice:         basePrice,
		AveragePricePerSqFt: basePricePerSqFt,
		DaysOnMarket:        35,
		PriceGrowth:         3.2,
		InventoryCount:      150,
		LastUpdated:         time.Now(),
	}
}

func (pvs *PropertyValuationService) getDefaultMarketData(zipCode string) *MarketData {
	return &MarketData{
		ZipCode:             zipCode,
		MedianPrice:         350000,
		AveragePricePerSqFt: 175,
		DaysOnMarket:        40,
		PriceGrowth:         2.5,
		InventoryCount:      200,
		LastUpdated:         time.Now(),
	}
}

func (pvs *PropertyValuationService) generateMockComparables(request PropertyValuationRequest) []ComparableProperty {
	var comparables []ComparableProperty

	// Generate 8-12 mock comparables with realistic data
	for i := 0; i < 10; i++ {
		// Vary the properties slightly
		sqftVariation := int(float32(request.SquareFeet) * (0.9 + float32(i)*0.02))
		bedVariation := request.Bedrooms
		if i%3 == 0 && request.Bedrooms > 2 {
			bedVariation = request.Bedrooms - 1
		} else if i%4 == 0 {
			bedVariation = request.Bedrooms + 1
		}

		// Calculate similarity score
		similarity := pvs.calculateSimilarityScore(request, sqftVariation, bedVariation, request.Bathrooms)

		// Mock sale price based on similar properties
		basePricePerSqFt := float32(180 + i*5) // Vary price per sq ft
		salePrice := int(basePricePerSqFt * float32(sqftVariation))

		// Adjust price for time (mock sales from 1-12 months ago)
		monthsAgo := i + 1
		saleDate := time.Now().AddDate(0, -monthsAgo, -i*3)

		comparable := ComparableProperty{
			Address:         fmt.Sprintf("%d Example St #%d, Houston, TX", 1000+i*100, i+1),
			Distance:        0.3 + float32(i)*0.1,
			SalePrice:       salePrice,
			SaleDate:        saleDate,
			SquareFeet:      sqftVariation,
			Bedrooms:        bedVariation,
			Bathrooms:       request.Bathrooms,
			YearBuilt:       request.YearBuilt + (i-5)*2, // Vary year built slightly
			PricePerSqFt:    basePricePerSqFt,
			AdjustedPrice:   salePrice, // In real implementation, adjust for differences
			SimilarityScore: similarity,
		}

		comparables = append(comparables, comparable)
	}

	return comparables
}

func (pvs *PropertyValuationService) calculateSimilarityScore(request PropertyValuationRequest, sqft, beds int, baths float32) float32 {
	score := 1.0

	// Square footage similarity (most important factor)
	sqftDiff := math.Abs(float64(request.SquareFeet-sqft)) / float64(request.SquareFeet)
	score -= sqftDiff * 0.4

	// Bedroom similarity
	bedDiff := math.Abs(float64(request.Bedrooms-beds)) / float64(request.Bedrooms+1)
	score -= bedDiff * 0.3

	// Bathroom similarity
	bathDiff := math.Abs(float64(request.Bathrooms-baths)) / float64(request.Bathrooms+1)
	score -= bathDiff * 0.3

	// Ensure score is between 0 and 1
	if score < 0 {
		score = 0
	}

	return float32(score)
}

func getConfidenceLevel(confidence float32) string {
	switch {
	case confidence >= 0.8:
		return "high"
	case confidence >= 0.6:
		return "medium"
	default:
		return "low"
	}
}



// GetValuationHistory returns historical valuations for a property
func (pvs *PropertyValuationService) GetValuationHistory(address string) ([]PropertyValuation, error) {
	// In production, this would query a database of historical valuations
	// For now, return empty slice
	return []PropertyValuation{}, nil
}

// UpdateMarketData manually updates market data for a zip code
func (pvs *PropertyValuationService) UpdateMarketData(zipCode string, data *MarketData) {
	pvs.marketDataCache[zipCode] = data
	log.Printf("📊 Updated market data for zip code %s", zipCode)
}
