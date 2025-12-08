package models

// PropertyCategory represents enhanced property categorization beyond rental/sales
type PropertyCategory struct {
	Type       string   `json:"type"`        // "rental", "sales", "mixed"
	Tier       string   `json:"tier"`        // "luxury", "mid_tier", "starter", "investment_grade"
	Segment    string   `json:"segment"`     // "single_family", "condo", "townhome", "apartment", "commercial"
	TargetDemo string   `json:"target_demo"` // "young_professional", "family", "student", "retiree", "investor"
	PriceRange string   `json:"price_range"` // "under_200k", "200k_400k", "400k_750k", "750k_1.5m", "luxury_1.5m_plus"
	Amenities  []string `json:"amenities"`   // Pool, gym, garage, etc.
	Location   string   `json:"location"`    // Specific Houston area classification
}

// BehavioralTriggerCondition represents multi-factor trigger conditions
type BehavioralTriggerCondition struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	RequiredSignals []string               `json:"required_signals"`
	OptionalSignals []string               `json:"optional_signals"`
	Thresholds      map[string]float64     `json:"thresholds"`
	PropertyTypes   []string               `json:"property_types"`
	MinConfidence   float64                `json:"min_confidence"`
	Actions         []string               `json:"actions"`
	Context         map[string]interface{} `json:"context"`
}

// AdvancedBehavioralEngine handles property-specific behavioral analysis
type AdvancedBehavioralEngine struct {
	PropertyCategories map[string]PropertyCategory           `json:"property_categories"`
	TriggerConditions  map[string]BehavioralTriggerCondition `json:"trigger_conditions"`
	BehavioralPatterns map[string]map[string]interface{}     `json:"behavioral_patterns"`
}
