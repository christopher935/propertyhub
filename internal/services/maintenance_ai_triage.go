package services

import (
	"chrisgross-ctrl-project/internal/models"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

type MaintenanceAITriage struct {
	db *gorm.DB
}

type TriageResult struct {
	Priority        string   `json:"priority"`
	Category        string   `json:"category"`
	SuggestedVendor string   `json:"suggested_vendor"`
	EstimatedCost   float64  `json:"estimated_cost"`
	ResponseTime    string   `json:"response_time"`
	AIReasoning     string   `json:"ai_reasoning"`
	Keywords        []string `json:"keywords_found"`
	ConfidenceScore float64  `json:"confidence_score"`
}

type keywordPattern struct {
	patterns []string
	weight   int
}

var emergencyKeywords = keywordPattern{
	patterns: []string{
		"flood", "flooding", "flooded",
		"fire", "smoke", "burning",
		"gas leak", "gas smell", "smell gas",
		"no heat", "no heating", "heater not working", "furnace not working",
		"sewage", "sewer backup", "sewage backup",
		"electrical fire", "sparking", "electrical spark",
		"carbon monoxide", "co detector",
		"burst pipe", "pipe burst", "broken pipe",
		"no water", "water shut off",
		"ceiling collapse", "roof collapse",
		"mold", "black mold",
		"security", "break in", "broken door lock", "broken window",
	},
	weight: 100,
}

var highPriorityKeywords = keywordPattern{
	patterns: []string{
		"leak", "leaking", "water leak",
		"no hot water", "hot water not working", "water heater",
		"broken lock", "lock broken", "can't lock",
		"pest infestation", "roaches", "rats", "mice", "bed bugs", "termites",
		"toilet not flushing", "toilet broken", "toilet clogged",
		"refrigerator not working", "fridge not working", "fridge broken",
		"stove not working", "oven not working",
		"ac not working", "air conditioning broken", "no ac", "no air",
		"electrical outlet not working", "no power",
		"garage door broken", "garage door not opening",
	},
	weight: 75,
}

var mediumPriorityKeywords = keywordPattern{
	patterns: []string{
		"appliance broken", "appliance not working",
		"clogged drain", "drain clogged", "slow drain",
		"running toilet", "toilet running",
		"dishwasher", "washing machine", "dryer",
		"garbage disposal", "disposal not working",
		"light fixture", "ceiling fan", "fan not working",
		"doorbell", "intercom",
		"window stuck", "door stuck",
		"thermostat", "hvac maintenance",
		"water pressure", "low water pressure",
	},
	weight: 50,
}

var lowPriorityKeywords = keywordPattern{
	patterns: []string{
		"cosmetic", "paint", "painting", "touch up",
		"minor repair", "small repair",
		"screen repair", "screen replacement",
		"caulking", "grout", "tile repair",
		"cabinet", "drawer", "handle",
		"weatherstripping", "door sweep",
		"light bulb", "replace bulb",
		"cleaning", "carpet cleaning",
		"landscaping", "yard",
		"blinds", "curtain rod",
	},
	weight: 25,
}

var categoryKeywords = map[string][]string{
	models.MaintenanceCategoryPlumbing: {
		"plumbing", "pipe", "leak", "faucet", "toilet", "drain", "water heater",
		"shower", "bathtub", "sink", "garbage disposal", "sewer", "sewage",
		"water pressure", "clog", "clogged", "flood", "flooding",
	},
	models.MaintenanceCategoryElectrical: {
		"electrical", "electric", "outlet", "switch", "light", "wiring",
		"circuit", "breaker", "fuse", "power", "spark", "shocking",
		"voltage", "fixture", "fan", "ceiling fan",
	},
	models.MaintenanceCategoryHVAC: {
		"hvac", "heating", "cooling", "air conditioning", "ac", "a/c",
		"furnace", "heater", "thermostat", "duct", "vent", "filter",
		"heat pump", "condenser", "compressor", "refrigerant",
	},
	models.MaintenanceCategoryAppliance: {
		"appliance", "refrigerator", "fridge", "stove", "oven", "range",
		"dishwasher", "microwave", "washer", "dryer", "washing machine",
		"ice maker", "freezer", "garbage disposal",
	},
	models.MaintenanceCategoryStructural: {
		"roof", "roofing", "ceiling", "wall", "floor", "foundation",
		"structural", "crack", "settling", "beam", "joist", "drywall",
		"stucco", "siding", "gutter", "downspout",
	},
	models.MaintenanceCategoryPest: {
		"pest", "bug", "insect", "roach", "roaches", "ant", "ants",
		"spider", "rat", "rats", "mice", "mouse", "termite", "termites",
		"bed bug", "bed bugs", "wasp", "bee", "rodent",
	},
	models.MaintenanceCategoryGeneral: {
		"door", "window", "lock", "key", "paint", "carpet", "blinds",
		"screen", "cabinet", "drawer", "handle", "knob", "hinge",
		"weatherstrip", "caulk", "grout", "tile",
	},
}

var estimatedCostByCategory = map[string]float64{
	models.MaintenanceCategoryPlumbing:   150.00,
	models.MaintenanceCategoryElectrical: 175.00,
	models.MaintenanceCategoryHVAC:       250.00,
	models.MaintenanceCategoryAppliance:  200.00,
	models.MaintenanceCategoryStructural: 500.00,
	models.MaintenanceCategoryPest:       125.00,
	models.MaintenanceCategoryGeneral:    100.00,
}

func NewMaintenanceAITriage(db *gorm.DB) *MaintenanceAITriage {
	return &MaintenanceAITriage{db: db}
}

func (t *MaintenanceAITriage) TriageRequest(request models.MaintenanceRequest) (*TriageResult, error) {
	log.Printf("🤖 Running AI triage for maintenance request: %s", request.AppFolioID)

	description := strings.ToLower(request.Description)
	title := strings.ToLower(request.PropertyAddress)
	fullText := description + " " + title

	priority, priorityKeywords, priorityScore := t.determinePriority(fullText)
	category, categoryKeywords := t.determineCategory(fullText)
	suggestedVendor := t.findSuggestedVendor(category)
	estimatedCost := t.estimateCost(category, priority)
	responseTime := t.determineResponseTime(priority)
	reasoning := t.generateReasoning(priority, category, priorityKeywords, categoryKeywords)

	allKeywords := append(priorityKeywords, categoryKeywords...)
	confidenceScore := t.calculateConfidence(priorityScore, len(allKeywords))

	result := &TriageResult{
		Priority:        priority,
		Category:        category,
		SuggestedVendor: suggestedVendor,
		EstimatedCost:   estimatedCost,
		ResponseTime:    responseTime,
		AIReasoning:     reasoning,
		Keywords:        allKeywords,
		ConfidenceScore: confidenceScore,
	}

	log.Printf("✅ Triage complete: Priority=%s, Category=%s, ResponseTime=%s, Confidence=%.2f",
		priority, category, responseTime, confidenceScore)

	return result, nil
}

func (t *MaintenanceAITriage) determinePriority(text string) (string, []string, int) {
	var foundKeywords []string
	highestScore := 0

	emergencyMatches := t.findKeywordMatches(text, emergencyKeywords.patterns)
	if len(emergencyMatches) > 0 {
		foundKeywords = append(foundKeywords, emergencyMatches...)
		highestScore = emergencyKeywords.weight
	}

	highMatches := t.findKeywordMatches(text, highPriorityKeywords.patterns)
	if len(highMatches) > 0 && highPriorityKeywords.weight > highestScore {
		foundKeywords = append(foundKeywords, highMatches...)
		if highestScore < highPriorityKeywords.weight {
			highestScore = highPriorityKeywords.weight
		}
	}

	mediumMatches := t.findKeywordMatches(text, mediumPriorityKeywords.patterns)
	if len(mediumMatches) > 0 {
		foundKeywords = append(foundKeywords, mediumMatches...)
		if highestScore < mediumPriorityKeywords.weight {
			highestScore = mediumPriorityKeywords.weight
		}
	}

	lowMatches := t.findKeywordMatches(text, lowPriorityKeywords.patterns)
	if len(lowMatches) > 0 {
		foundKeywords = append(foundKeywords, lowMatches...)
		if highestScore < lowPriorityKeywords.weight {
			highestScore = lowPriorityKeywords.weight
		}
	}

	if t.isWinterSeason() && containsAny(text, []string{"no heat", "heater", "furnace", "heating"}) {
		highestScore = emergencyKeywords.weight
	}
	if t.isSummerSeason() && containsAny(text, []string{"no ac", "no air", "air conditioning", "cooling"}) {
		highestScore = max(highestScore, highPriorityKeywords.weight)
	}

	priority := models.MaintenancePriorityLow
	switch {
	case highestScore >= emergencyKeywords.weight:
		priority = models.MaintenancePriorityEmergency
	case highestScore >= highPriorityKeywords.weight:
		priority = models.MaintenancePriorityHigh
	case highestScore >= mediumPriorityKeywords.weight:
		priority = models.MaintenancePriorityMedium
	default:
		priority = models.MaintenancePriorityLow
	}

	return priority, foundKeywords, highestScore
}

func (t *MaintenanceAITriage) determineCategory(text string) (string, []string) {
	categoryScores := make(map[string]int)
	categoryMatches := make(map[string][]string)

	for category, keywords := range categoryKeywords {
		matches := t.findKeywordMatches(text, keywords)
		categoryScores[category] = len(matches)
		categoryMatches[category] = matches
	}

	bestCategory := models.MaintenanceCategoryGeneral
	bestScore := 0
	for category, score := range categoryScores {
		if score > bestScore {
			bestCategory = category
			bestScore = score
		}
	}

	return bestCategory, categoryMatches[bestCategory]
}

func (t *MaintenanceAITriage) findKeywordMatches(text string, keywords []string) []string {
	var matches []string
	for _, keyword := range keywords {
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(keyword) + `\b`)
		if pattern.MatchString(text) {
			matches = append(matches, keyword)
		}
	}
	return matches
}

func (t *MaintenanceAITriage) findSuggestedVendor(category string) string {
	var vendor models.Vendor
	err := t.db.Where("category = ? AND is_active = ? AND is_preferred = ?", category, true, true).
		Order("rating DESC, total_jobs DESC").
		First(&vendor).Error

	if err != nil {
		err = t.db.Where("category = ? AND is_active = ?", category, true).
			Order("rating DESC, total_jobs DESC").
			First(&vendor).Error
	}

	if err != nil {
		err = t.db.Where("is_active = ? AND is_preferred = ?", true, true).
			Order("rating DESC").
			First(&vendor).Error
	}

	if err != nil {
		return ""
	}

	return vendor.Name
}

func (t *MaintenanceAITriage) estimateCost(category string, priority string) float64 {
	baseCost, exists := estimatedCostByCategory[category]
	if !exists {
		baseCost = 100.00
	}

	switch priority {
	case models.MaintenancePriorityEmergency:
		baseCost *= 1.5
	case models.MaintenancePriorityHigh:
		baseCost *= 1.25
	}

	return baseCost
}

func (t *MaintenanceAITriage) determineResponseTime(priority string) string {
	switch priority {
	case models.MaintenancePriorityEmergency:
		return models.ResponseTimeImmediate
	case models.MaintenancePriorityHigh:
		return models.ResponseTime24Hours
	case models.MaintenancePriorityMedium:
		return models.ResponseTime48Hours
	default:
		return models.ResponseTimeScheduled
	}
}

func (t *MaintenanceAITriage) generateReasoning(priority, category string, priorityKeywords, categoryKeywords []string) string {
	var reasoning strings.Builder

	reasoning.WriteString("AI Triage Analysis:\n")
	reasoning.WriteString("-------------------\n")

	if len(priorityKeywords) > 0 {
		reasoning.WriteString("Priority Keywords Found: ")
		reasoning.WriteString(strings.Join(priorityKeywords, ", "))
		reasoning.WriteString("\n")
	}

	if len(categoryKeywords) > 0 {
		reasoning.WriteString("Category Keywords Found: ")
		reasoning.WriteString(strings.Join(categoryKeywords, ", "))
		reasoning.WriteString("\n")
	}

	reasoning.WriteString("\nClassification:\n")
	reasoning.WriteString("- Priority: " + strings.ToUpper(priority))
	switch priority {
	case models.MaintenancePriorityEmergency:
		reasoning.WriteString(" (Immediate attention required - safety/habitability concern)")
	case models.MaintenancePriorityHigh:
		reasoning.WriteString(" (Address within 24 hours - significant impact on tenant)")
	case models.MaintenancePriorityMedium:
		reasoning.WriteString(" (Address within 48 hours - moderate inconvenience)")
	case models.MaintenancePriorityLow:
		reasoning.WriteString(" (Schedule at convenience - minor issue)")
	}
	reasoning.WriteString("\n")

	reasoning.WriteString("- Category: " + strings.ToUpper(category) + "\n")

	if t.isWinterSeason() {
		reasoning.WriteString("\n⚠️ Winter season: Heating issues are elevated to higher priority\n")
	}
	if t.isSummerSeason() {
		reasoning.WriteString("\n⚠️ Summer season: Cooling issues are elevated to higher priority\n")
	}

	return reasoning.String()
}

func (t *MaintenanceAITriage) calculateConfidence(priorityScore int, keywordCount int) float64 {
	confidence := 0.5

	if keywordCount > 0 {
		confidence += float64(keywordCount) * 0.05
	}

	if priorityScore >= emergencyKeywords.weight {
		confidence += 0.2
	} else if priorityScore >= highPriorityKeywords.weight {
		confidence += 0.15
	} else if priorityScore >= mediumPriorityKeywords.weight {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 0.99
	}

	return confidence
}

func (t *MaintenanceAITriage) isWinterSeason() bool {
	month := time.Now().Month()
	return month >= time.November || month <= time.March
}

func (t *MaintenanceAITriage) isSummerSeason() bool {
	month := time.Now().Month()
	return month >= time.June && month <= time.September
}

func (t *MaintenanceAITriage) BatchTriageRequests(requests []models.MaintenanceRequest) ([]TriageResult, error) {
	results := make([]TriageResult, len(requests))
	for i, req := range requests {
		result, err := t.TriageRequest(req)
		if err != nil {
			log.Printf("⚠️ Error triaging request %s: %v", req.AppFolioID, err)
			continue
		}
		results[i] = *result
	}
	return results, nil
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
