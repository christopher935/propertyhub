package services

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// PropertyMatchingService matches properties to leads based on criteria
type PropertyMatchingService struct {
	db *gorm.DB
}

// NewPropertyMatchingService creates a new property matching service
func NewPropertyMatchingService(db *gorm.DB) *PropertyMatchingService {
	return &PropertyMatchingService{
		db: db,
	}
}

// LeadCriteria represents a lead's property search criteria
type LeadCriteria struct {
	LeadID          int64
	MinBedrooms     int
	MaxBedrooms     int
	MinBathrooms    float64
	MaxBathrooms    float64
	MinPrice        float64
	MaxPrice        float64
	PreferredZips   []string
	PreferredCities []string
	PropertyType    string // "apartment", "house", "condo", etc.
	MustHaveFeatures []string // "parking", "pool", "pets_allowed", etc.
}

// PropertyMatch represents a property matched to a lead
type PropertyMatch struct {
	PropertyID      int64                  `json:"property_id"`
	LeadID          int64                  `json:"lead_id"`
	MatchScore      float64                `json:"match_score"` // 0-100
	MatchReasons    []string               `json:"match_reasons"`
	Property        *models.Property       `json:"property"`
	Lead            *models.Lead           `json:"lead"`
	MatchedAt       time.Time              `json:"matched_at"`
	NotificationSent bool                  `json:"notification_sent"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// FindMatchesForProperty finds all leads that match a new property
func (pms *PropertyMatchingService) FindMatchesForProperty(propertyID int64) ([]PropertyMatch, error) {
	log.Printf("🔍 Property Matching: Finding matches for property %d", propertyID)
	
	// Get the property
	var property models.Property
	err := pms.db.First(&property, propertyID).Error
	if err != nil {
		return nil, err
	}
	
	// Get all active leads
	var leads []models.Lead
	err = pms.db.Where("status = ?", "active").Find(&leads).Error
	if err != nil {
		return nil, err
	}
	
	matches := []PropertyMatch{}
	
	for _, lead := range leads {
		// Get lead criteria
		criteria := pms.extractLeadCriteria(lead)
		
		// Calculate match score
		matchScore, reasons := pms.calculateMatchScore(property, criteria)
		
		// Only include if match score > 60
		if matchScore >= 60 {
			match := PropertyMatch{
				PropertyID:   int64(property.ID),
				LeadID:       int64(lead.ID),
				MatchScore:   matchScore,
				MatchReasons: reasons,
				Property:     &property,
				Lead:         &lead,
				MatchedAt:    time.Now(),
				Metadata: map[string]interface{}{
					"property_address": property.Address,
					"property_price":   property.Price,
					"lead_name":        lead.FirstName + " " + lead.LastName,
					"lead_email":       lead.Email,
				},
			}
			
			matches = append(matches, match)
		}
	}
	
	log.Printf("✅ Found %d matches for property %d", len(matches), propertyID)
	
	return matches, nil
}

// FindMatchesForLead finds all properties that match a lead's criteria
func (pms *PropertyMatchingService) FindMatchesForLead(leadID int64) ([]PropertyMatch, error) {
	log.Printf("🔍 Property Matching: Finding matches for lead %d", leadID)
	
	// Get the lead
	var lead models.Lead
	err := pms.db.First(&lead, leadID).Error
	if err != nil {
		return nil, err
	}
	
	// Extract criteria
	criteria := pms.extractLeadCriteria(lead)
	
	// Get active properties
	var properties []models.Property
	query := pms.db.Where("status = ?", "https://schema.org/InStock")
	
	// Apply basic filters
	if criteria.MinPrice > 0 && criteria.MaxPrice > 0 {
		query = query.Where("price BETWEEN ? AND ?", criteria.MinPrice, criteria.MaxPrice)
	}
	if criteria.MinBedrooms > 0 {
		query = query.Where("bedrooms >= ?", criteria.MinBedrooms)
	}
	if criteria.MaxBedrooms > 0 {
		query = query.Where("bedrooms <= ?", criteria.MaxBedrooms)
	}
	
	err = query.Find(&properties).Error
	if err != nil {
		return nil, err
	}
	
	matches := []PropertyMatch{}
	
	for _, property := range properties {
		matchScore, reasons := pms.calculateMatchScore(property, criteria)
		
		if matchScore >= 60 {
			match := PropertyMatch{
				PropertyID:   int64(property.ID),
				LeadID:       int64(lead.ID),
				MatchScore:   matchScore,
				MatchReasons: reasons,
				Property:     &property,
				Lead:         &lead,
				MatchedAt:    time.Now(),
				Metadata: map[string]interface{}{
					"property_address": property.Address,
					"property_price":   property.Price,
				},
			}
			
			matches = append(matches, match)
		}
	}
	
	log.Printf("✅ Found %d matches for lead %d", len(matches), leadID)
	
	return matches, nil
}

// FindNewMatchesSince finds new property matches created since a given time
func (pms *PropertyMatchingService) FindNewMatchesSince(since time.Time) ([]PropertyMatch, error) {
	log.Printf("🔍 Property Matching: Finding new matches since %s", since.Format("2006-01-02"))
	
	// Get properties added since the given time
	var newProperties []models.Property
	err := pms.db.Where("created_at >= ?", since).
		Where("status = ?", "https://schema.org/InStock").
		Find(&newProperties).Error
	
	if err != nil {
		return nil, err
	}
	
	allMatches := []PropertyMatch{}
	
	for _, property := range newProperties {
		matches, err := pms.FindMatchesForProperty(int64(property.ID))
		if err == nil {
			allMatches = append(allMatches, matches...)
		}
	}
	
	log.Printf("✅ Found %d new matches from %d new properties", len(allMatches), len(newProperties))
	
	return allMatches, nil
}

// extractLeadCriteria extracts search criteria from lead data
func (pms *PropertyMatchingService) extractLeadCriteria(lead models.Lead) LeadCriteria {
	criteria := LeadCriteria{
		LeadID: int64(lead.ID),
	}
	
	// Extract from lead preferences (stored in CustomFields JSONB field)
	if lead.CustomFields != nil {
		prefs := lead.CustomFields
		
		// Bedrooms
		if minBed, ok := prefs["min_bedrooms"].(float64); ok {
			criteria.MinBedrooms = int(minBed)
		}
		if maxBed, ok := prefs["max_bedrooms"].(float64); ok {
			criteria.MaxBedrooms = int(maxBed)
		}
		
		// Bathrooms
		if minBath, ok := prefs["min_bathrooms"].(float64); ok {
			criteria.MinBathrooms = minBath
		}
		if maxBath, ok := prefs["max_bathrooms"].(float64); ok {
			criteria.MaxBathrooms = maxBath
		}
		
		// Price
		if minPrice, ok := prefs["min_price"].(float64); ok {
			criteria.MinPrice = minPrice
		}
		if maxPrice, ok := prefs["max_price"].(float64); ok {
			criteria.MaxPrice = maxPrice
		}
		
		// Location
		if zips, ok := prefs["preferred_zips"].([]interface{}); ok {
			for _, zip := range zips {
				if zipStr, ok := zip.(string); ok {
					criteria.PreferredZips = append(criteria.PreferredZips, zipStr)
				}
			}
		}
		
		if cities, ok := prefs["preferred_cities"].([]interface{}); ok {
			for _, city := range cities {
				if cityStr, ok := city.(string); ok {
					criteria.PreferredCities = append(criteria.PreferredCities, cityStr)
				}
			}
		}
		
		// Property type
		if propType, ok := prefs["property_type"].(string); ok {
			criteria.PropertyType = propType
		}
		
		// Features
		if features, ok := prefs["must_have_features"].([]interface{}); ok {
			for _, feature := range features {
				if featureStr, ok := feature.(string); ok {
					criteria.MustHaveFeatures = append(criteria.MustHaveFeatures, featureStr)
				}
			}
		}
	}
	
	// If no explicit criteria, infer from behavioral data
	if criteria.MinBedrooms == 0 && criteria.MaxBedrooms == 0 {
		criteria.MinBedrooms, criteria.MaxBedrooms = pms.inferBedroomPreference(lead)
	}
	if criteria.MinPrice == 0 && criteria.MaxPrice == 0 {
		criteria.MinPrice, criteria.MaxPrice = pms.inferPriceRange(lead)
	}
	
	return criteria
}

// inferBedroomPreference infers bedroom preference from viewed properties
func (pms *PropertyMatchingService) inferBedroomPreference(lead models.Lead) (int, int) {
	// Get properties the lead has viewed
	var events []models.BehavioralEvent
	pms.db.Where("lead_id = ?", lead.ID).
		Where("event_type = ?", "property_view").
		Where("property_id IS NOT NULL").
		Limit(10).
		Find(&events)
	
	if len(events) == 0 {
		return 1, 4 // Default range
	}
	
	// Get property IDs
	propertyIDs := []int64{}
	for _, event := range events {
		if event.PropertyID != nil {
			propertyIDs = append(propertyIDs, *event.PropertyID)
		}
	}
	
	// Get bedroom counts
	var properties []models.Property
	pms.db.Where("id IN ?", propertyIDs).Find(&properties)
	
	if len(properties) == 0 {
		return 1, 4
	}
	
	// Calculate average and range
	total := 0
	min := 999
	max := 0
	for _, prop := range properties {
		if prop.Bedrooms != nil {
			total += *prop.Bedrooms
			if *prop.Bedrooms < min {
				min = *prop.Bedrooms
			}
			if *prop.Bedrooms > max {
				max = *prop.Bedrooms
			}
		}
	}
	
	// Return range with some flexibility
	if min > 0 {
		min = min - 1
		if min < 1 {
			min = 1
		}
	}
	max = max + 1
	
	return min, max
}

// inferPriceRange infers price range from viewed properties
func (pms *PropertyMatchingService) inferPriceRange(lead models.Lead) (float64, float64) {
	// Similar logic to bedroom inference
	var events []models.BehavioralEvent
	pms.db.Where("lead_id = ?", lead.ID).
		Where("event_type = ?", "property_view").
		Where("property_id IS NOT NULL").
		Limit(10).
		Find(&events)
	
	if len(events) == 0 {
		return 800, 3000 // Default range
	}
	
	propertyIDs := []int64{}
	for _, event := range events {
		if event.PropertyID != nil {
			propertyIDs = append(propertyIDs, *event.PropertyID)
		}
	}
	
	var properties []models.Property
	pms.db.Where("id IN ?", propertyIDs).Find(&properties)
	
	if len(properties) == 0 {
		return 800, 3000
	}
	
	// Calculate range
	var prices []float64
	for _, prop := range properties {
		if prop.Price > 0 {
			prices = append(prices, prop.Price)
		}
	}
	
	if len(prices) == 0 {
		return 800, 3000
	}
	
	// Find min and max
	minPrice := prices[0]
	maxPrice := prices[0]
	for _, price := range prices {
		if price < minPrice {
			minPrice = price
		}
		if price > maxPrice {
			maxPrice = price
		}
	}
	
	// Add 20% buffer
	minPrice = minPrice * 0.8
	maxPrice = maxPrice * 1.2
	
	return minPrice, maxPrice
}

// calculateMatchScore calculates how well a property matches lead criteria
func (pms *PropertyMatchingService) calculateMatchScore(property models.Property, criteria LeadCriteria) (float64, []string) {
	score := 0.0
	maxScore := 0.0
	reasons := []string{}
	
	// Bedroom match (weight: 25)
	maxScore += 25
	if property.Bedrooms != nil && (criteria.MinBedrooms > 0 || criteria.MaxBedrooms > 0) {
		bedrooms := *property.Bedrooms
		if bedrooms >= criteria.MinBedrooms && (criteria.MaxBedrooms == 0 || bedrooms <= criteria.MaxBedrooms) {
			score += 25
			reasons = append(reasons, fmt.Sprintf("✓ %d bedrooms matches preference", bedrooms))
		} else {
			// Partial score if close
			if math.Abs(float64(bedrooms-criteria.MinBedrooms)) <= 1 {
				score += 15
				reasons = append(reasons, fmt.Sprintf("~ %d bedrooms close to preference", bedrooms))
			}
		}
	} else {
		score += 12.5 // No preference specified
	}
	
	// Bathroom match (weight: 15)
	maxScore += 15
	if property.Bathrooms != nil && (criteria.MinBathrooms > 0 || criteria.MaxBathrooms > 0) {
		bathrooms := float64(*property.Bathrooms)
		if bathrooms >= criteria.MinBathrooms && (criteria.MaxBathrooms == 0 || bathrooms <= criteria.MaxBathrooms) {
			score += 15
			reasons = append(reasons, fmt.Sprintf("✓ %.1f bathrooms matches preference", bathrooms))
		} else if math.Abs(bathrooms-criteria.MinBathrooms) <= 0.5 {
			score += 10
		}
	} else {
		score += 7.5
	}
	
	// Price match (weight: 30)
	maxScore += 30
	if criteria.MinPrice > 0 || criteria.MaxPrice > 0 {
		if property.Price >= criteria.MinPrice && (criteria.MaxPrice == 0 || property.Price <= criteria.MaxPrice) {
			score += 30
			reasons = append(reasons, fmt.Sprintf("✓ $%.0f/mo within budget", property.Price))
		} else {
			// Partial score if within 15%
			if property.Price < criteria.MinPrice {
				diff := (criteria.MinPrice - property.Price) / criteria.MinPrice
				if diff <= 0.15 {
					score += 20
					reasons = append(reasons, fmt.Sprintf("✓ $%.0f/mo slightly below budget (good value!)", property.Price))
				}
			} else if property.Price > criteria.MaxPrice {
				diff := (property.Price - criteria.MaxPrice) / criteria.MaxPrice
				if diff <= 0.15 {
					score += 15
					reasons = append(reasons, fmt.Sprintf("~ $%.0f/mo slightly above budget", property.Price))
				}
			}
		}
	} else {
		score += 15
	}
	
	// Location match (weight: 20)
	maxScore += 20
	locationMatch := false
	if len(criteria.PreferredZips) > 0 {
		for _, zip := range criteria.PreferredZips {
			if strings.Contains(string(property.Address), zip) {
				score += 20
				reasons = append(reasons, "✓ In preferred ZIP code")
				locationMatch = true
				break
			}
		}
	}
	if !locationMatch && len(criteria.PreferredCities) > 0 {
		for _, city := range criteria.PreferredCities {
			if strings.Contains(strings.ToLower(string(property.Address)), strings.ToLower(city)) {
				score += 15
				reasons = append(reasons, "✓ In preferred city")
				locationMatch = true
				break
			}
		}
	}
	if !locationMatch {
		score += 5 // Some base score
	}
	
	// Property type match (weight: 10)
	maxScore += 10
	if criteria.PropertyType != "" {
		if strings.EqualFold(property.PropertyType, criteria.PropertyType) {
			score += 10
			reasons = append(reasons, fmt.Sprintf("✓ %s type matches", property.PropertyType))
		}
	} else {
		score += 5
	}
	
	// Normalize to 0-100
	finalScore := (score / maxScore) * 100
	
	return finalScore, reasons
}

// GetMatchInsight generates an HTML insight for a property match
func (pms *PropertyMatchingService) GetMatchInsight(match PropertyMatch) string {
	return fmt.Sprintf(
		`<div class="insight-property-match">
			<strong>🏠 New Property Match:</strong> <a href="/admin/properties/%d">%s</a> is a <strong>%.0f%% match</strong> for %s
			<br><br>
			<div class="match-reasons">
				%s
			</div>
			<br>
			<span class="insight-stat">Price: $%.0f/mo</span> | 
			<span class="insight-stat">%d BR / %.1f BA</span>
			<br><br>
			<em>Send notification now for best engagement. Leads respond 3x faster to new property alerts.</em>
		</div>`,
		match.PropertyID,
		match.Property.Address,
		match.MatchScore,
		match.Lead.FirstName + " " + match.Lead.LastName,
		strings.Join(match.MatchReasons, "<br>"),
		match.Property.Price,
		match.Property.Bedrooms,
		match.Property.Bathrooms,
	)
}
