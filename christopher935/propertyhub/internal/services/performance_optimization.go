package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// PerformanceOptimizationService handles A/B testing and performance optimization
type PerformanceOptimizationService struct {
	analyticsService   *AnalyticsAutomationService
	leadScoringService *BehavioralLeadScoringService
	emailService       *EmailService
	smsService         *SMSService
	experiments        map[string]*Experiment
	activeTests        map[string]*ActiveTest
	optimizationRules  []OptimizationRule
	performanceMetrics *OptimizationMetrics
	mutex              sync.Mutex
}

// Experiment defines an A/B test configuration
type Experiment struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Type            string              `json:"type"`   // "email_subject", "landing_page", "form_flow", "pricing"
	Status          string              `json:"status"` // "draft", "running", "paused", "completed"
	StartDate       time.Time           `json:"startDate"`
	EndDate         time.Time           `json:"endDate"`
	TrafficSplit    float64             `json:"trafficSplit"` // 0.5 = 50/50 split
	Variants        []ExperimentVariant `json:"variants"`
	TargetMetric    string              `json:"targetMetric"` // "conversion_rate", "click_rate", "revenue"
	MinSampleSize   int                 `json:"minSampleSize"`
	ConfidenceLevel float64             `json:"confidenceLevel"` // 0.95 = 95%
	Results         *ExperimentResults  `json:"results"`
	CreatedBy       string              `json:"createdBy"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

// ExperimentVariant represents a variant in an A/B test
type ExperimentVariant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Content     map[string]interface{} `json:"content"`
	Weight      float64                `json:"weight"` // Traffic allocation weight
	IsControl   bool                   `json:"isControl"`
}

// ExperimentResults stores the results of an A/B test
type ExperimentResults struct {
	TotalParticipants int                        `json:"totalParticipants"`
	VariantResults    map[string]*VariantResults `json:"variantResults"`
	Winner            string                     `json:"winner,omitempty"`
	Confidence        float64                    `json:"confidence"`
	StatisticalSig    bool                       `json:"statisticalSignificance"`
	LiftPercent       float64                    `json:"liftPercent"`
	CompletedAt       time.Time                  `json:"completedAt,omitempty"`
}

// VariantResults stores results for a specific variant
type VariantResults struct {
	VariantID      string  `json:"variantId"`
	Participants   int     `json:"participants"`
	Conversions    int     `json:"conversions"`
	ConversionRate float64 `json:"conversionRate"`
	Revenue        float64 `json:"revenue"`
	AvgOrderValue  float64 `json:"avgOrderValue"`
	ClickRate      float64 `json:"clickRate"`
	OpenRate       float64 `json:"openRate"`
	BounceRate     float64 `json:"bounceRate"`
	TimeOnPage     float64 `json:"timeOnPage"`
	StandardError  float64 `json:"standardError"`
}

// ActiveTest represents a user's participation in an experiment
type ActiveTest struct {
	UserID       string            `json:"userId"`
	ExperimentID string            `json:"experimentId"`
	VariantID    string            `json:"variantId"`
	StartTime    time.Time         `json:"startTime"`
	Converted    bool              `json:"converted"`
	ConvertedAt  time.Time         `json:"convertedAt,omitempty"`
	Revenue      float64           `json:"revenue"`
	Events       []ExperimentEvent `json:"events"`
}

// ExperimentEvent tracks user actions during an experiment
type ExperimentEvent struct {
	EventType string                 `json:"eventType"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// OptimizationRule defines automatic optimization rules
type OptimizationRule struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Type      string                  `json:"type"` // "auto_winner", "traffic_allocation", "early_stop"
	Condition func(*Experiment) bool  `json:"-"`
	Action    func(*Experiment) error `json:"-"`
	Enabled   bool                    `json:"enabled"`
	Priority  int                     `json:"priority"`
}

// OptimizationMetrics tracks overall system performance
type OptimizationMetrics struct {
	TotalExperiments  int                     `json:"totalExperiments"`
	ActiveExperiments int                     `json:"activeExperiments"`
	CompletedTests    int                     `json:"completedTests"`
	OverallLift       float64                 `json:"overallLift"`
	TotalRevenueLift  float64                 `json:"totalRevenueLift"`
	AvgTestDuration   float64                 `json:"avgTestDuration"` // days
	SuccessfulTests   int                     `json:"successfulTests"`
	MetricsByType     map[string]*TypeMetrics `json:"metricsByType"`
	LastUpdated       time.Time               `json:"lastUpdated"`
}

// TypeMetrics tracks metrics by experiment type
type TypeMetrics struct {
	TotalTests      int     `json:"totalTests"`
	SuccessfulTests int     `json:"successfulTests"`
	AvgLift         float64 `json:"avgLift"`
	AvgDuration     float64 `json:"avgDuration"`
	TotalRevenue    float64 `json:"totalRevenue"`
}

// NewPerformanceOptimizationService creates a new performance optimization service
func NewPerformanceOptimizationService(analyticsService *AnalyticsAutomationService, leadScoringService *BehavioralLeadScoringService, emailService *EmailService, smsService *SMSService) *PerformanceOptimizationService {
	service := &PerformanceOptimizationService{
		analyticsService:   analyticsService,
		leadScoringService: leadScoringService,
		emailService:       emailService,
		smsService:         smsService,
		experiments:        make(map[string]*Experiment),
		activeTests:        make(map[string]*ActiveTest),
		optimizationRules:  []OptimizationRule{},
		performanceMetrics: &OptimizationMetrics{
			MetricsByType: make(map[string]*TypeMetrics),
			LastUpdated:   time.Now(),
		},
	}

	// Initialize optimization rules
	service.initializeOptimizationRules()

	// Initialize default experiments
	service.initializeDefaultExperiments()

	// Start optimization processing routine
	go service.optimizationProcessingRoutine()

	return service
}

// CreateExperiment creates a new A/B test experiment
func (s *PerformanceOptimizationService) CreateExperiment(experiment *Experiment) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate experiment
	if err := s.validateExperiment(experiment); err != nil {
		return fmt.Errorf("experiment validation failed: %v", err)
	}

	// Generate ID if not provided
	if experiment.ID == "" {
		experiment.ID = s.generateExperimentID(experiment.Name)
	}

	// Set timestamps
	experiment.CreatedAt = time.Now()
	experiment.UpdatedAt = time.Now()
	experiment.Status = "draft"

	// Initialize results
	experiment.Results = &ExperimentResults{
		VariantResults: make(map[string]*VariantResults),
	}

	// Initialize variant results
	for _, variant := range experiment.Variants {
		experiment.Results.VariantResults[variant.ID] = &VariantResults{
			VariantID: variant.ID,
		}
	}

	s.experiments[experiment.ID] = experiment

	log.Printf("✅ Created experiment: %s (%s)", experiment.Name, experiment.ID)
	return nil
}

func (s *PerformanceOptimizationService) validateExperiment(experiment *Experiment) error {
	if experiment.Name == "" {
		return fmt.Errorf("experiment name is required")
	}

	if len(experiment.Variants) < 2 {
		return fmt.Errorf("experiment must have at least 2 variants")
	}

	if experiment.TrafficSplit <= 0 || experiment.TrafficSplit > 1 {
		return fmt.Errorf("traffic split must be between 0 and 1")
	}

	if experiment.ConfidenceLevel <= 0 || experiment.ConfidenceLevel >= 1 {
		return fmt.Errorf("confidence level must be between 0 and 1")
	}

	// Validate variant weights sum to 1
	totalWeight := 0.0
	controlCount := 0
	for _, variant := range experiment.Variants {
		totalWeight += variant.Weight
		if variant.IsControl {
			controlCount++
		}
	}

	if math.Abs(totalWeight-1.0) > 0.001 {
		return fmt.Errorf("variant weights must sum to 1.0")
	}

	if controlCount != 1 {
		return fmt.Errorf("experiment must have exactly one control variant")
	}

	return nil
}

func (s *PerformanceOptimizationService) generateExperimentID(name string) string {
	hash := md5.Sum([]byte(name + time.Now().String()))
	return "exp_" + hex.EncodeToString(hash[:])[:8]
}

// StartExperiment starts an A/B test
func (s *PerformanceOptimizationService) StartExperiment(experimentID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	experiment, exists := s.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment not found: %s", experimentID)
	}

	if experiment.Status != "draft" && experiment.Status != "paused" {
		return fmt.Errorf("experiment cannot be started from status: %s", experiment.Status)
	}

	experiment.Status = "running"
	experiment.StartDate = time.Now()
	experiment.UpdatedAt = time.Now()

	s.performanceMetrics.ActiveExperiments++
	s.performanceMetrics.TotalExperiments++

	log.Printf("🚀 Started experiment: %s", experiment.Name)
	return nil
}

// StopExperiment stops an A/B test
func (s *PerformanceOptimizationService) StopExperiment(experimentID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	experiment, exists := s.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment not found: %s", experimentID)
	}

	if experiment.Status != "running" {
		return fmt.Errorf("experiment is not running")
	}

	experiment.Status = "completed"
	experiment.EndDate = time.Now()
	experiment.UpdatedAt = time.Now()
	experiment.Results.CompletedAt = time.Now()

	// Calculate final results
	s.calculateExperimentResults(experiment)

	s.performanceMetrics.ActiveExperiments--
	s.performanceMetrics.CompletedTests++

	log.Printf("🏁 Stopped experiment: %s", experiment.Name)
	return nil
}

// AssignUserToExperiment assigns a user to an experiment variant
func (s *PerformanceOptimizationService) AssignUserToExperiment(userID, experimentID string) (*ExperimentVariant, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	experiment, exists := s.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	if experiment.Status != "running" {
		return nil, fmt.Errorf("experiment is not running")
	}

	// Check if user is already assigned
	testKey := userID + "_" + experimentID
	if activeTest, exists := s.activeTests[testKey]; exists {
		// Return existing variant
		for _, variant := range experiment.Variants {
			if variant.ID == activeTest.VariantID {
				return &variant, nil
			}
		}
	}

	// Assign user to variant based on weights
	variant := s.selectVariantForUser(userID, experiment)

	// Create active test
	activeTest := &ActiveTest{
		UserID:       userID,
		ExperimentID: experimentID,
		VariantID:    variant.ID,
		StartTime:    time.Now(),
		Events:       []ExperimentEvent{},
	}

	s.activeTests[testKey] = activeTest

	// Update experiment results
	experiment.Results.TotalParticipants++
	experiment.Results.VariantResults[variant.ID].Participants++

	log.Printf("👤 Assigned user %s to experiment %s, variant %s", userID, experimentID, variant.Name)
	return variant, nil
}

func (s *PerformanceOptimizationService) selectVariantForUser(userID string, experiment *Experiment) *ExperimentVariant {
	// Use consistent hashing to ensure same user always gets same variant
	hash := md5.Sum([]byte(userID + experiment.ID))
	hashValue := float64(hash[0]) / 255.0

	// Select variant based on cumulative weights
	cumulative := 0.0
	for _, variant := range experiment.Variants {
		cumulative += variant.Weight
		if hashValue <= cumulative {
			return &variant
		}
	}

	// Fallback to control variant
	for _, variant := range experiment.Variants {
		if variant.IsControl {
			return &variant
		}
	}

	return &experiment.Variants[0]
}

// TrackExperimentEvent tracks an event for an active test
func (s *PerformanceOptimizationService) TrackExperimentEvent(userID, experimentID, eventType string, data map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	testKey := userID + "_" + experimentID
	activeTest, exists := s.activeTests[testKey]
	if !exists {
		return fmt.Errorf("no active test found for user %s in experiment %s", userID, experimentID)
	}

	// Add event to test
	event := ExperimentEvent{
		EventType: eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	activeTest.Events = append(activeTest.Events, event)

	// Handle conversion events
	if eventType == "conversion" || eventType == "booking_completed" {
		if !activeTest.Converted {
			activeTest.Converted = true
			activeTest.ConvertedAt = time.Now()

			if revenue, ok := data["revenue"].(float64); ok {
				activeTest.Revenue = revenue
			}

			// Update experiment results
			experiment := s.experiments[experimentID]
			if experiment != nil {
				experiment.Results.VariantResults[activeTest.VariantID].Conversions++
				experiment.Results.VariantResults[activeTest.VariantID].Revenue += activeTest.Revenue
				s.updateConversionRates(experiment)
			}
		}
	}

	log.Printf("📊 Tracked event %s for user %s in experiment %s", eventType, userID, experimentID)
	return nil
}

func (s *PerformanceOptimizationService) updateConversionRates(experiment *Experiment) {
	if experiment.Results == nil {
		return
	}
	for _, results := range experiment.Results.VariantResults {
		if results.Participants > 0 {
			results.ConversionRate = float64(results.Conversions) / float64(results.Participants)
			if results.Conversions > 0 {
				results.AvgOrderValue = results.Revenue / float64(results.Conversions)
			}
			results.StandardError = math.Sqrt(results.ConversionRate * (1 - results.ConversionRate) / float64(results.Participants))
		}
	}
}

// GetExperimentForUser returns the active experiment variant for a user
func (s *PerformanceOptimizationService) GetExperimentForUser(userID, experimentType string) (*ExperimentVariant, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find running experiment of the specified type
	for _, experiment := range s.experiments {
		if experiment.Status == "running" && experiment.Type == experimentType {
			variant, err := s.AssignUserToExperiment(userID, experiment.ID)
			if err != nil {
				continue
			}
			return variant, nil
		}
	}

	return nil, fmt.Errorf("no active experiment found for type: %s", experimentType)
}

// calculateExperimentResults calculates statistical significance and winner
func (s *PerformanceOptimizationService) calculateExperimentResults(experiment *Experiment) {
	if len(experiment.Results.VariantResults) < 2 {
		return
	}

	// Find control variant
	var controlResults *VariantResults
	for _, variant := range experiment.Variants {
		if variant.IsControl {
			controlResults = experiment.Results.VariantResults[variant.ID]
			break
		}
	}

	if controlResults == nil {
		return
	}

	// Calculate statistical significance and lift for each variant
	bestVariantID := ""
	bestConversionRate := 0.0

	for variantID, results := range experiment.Results.VariantResults {
		if variantID == controlResults.VariantID {
			continue
		}

		// Calculate z-score for statistical significance
		zScore := s.calculateZScore(controlResults, results)

		// Check if statistically significant (z > 1.96 for 95% confidence)
		if math.Abs(zScore) > 1.96 {
			experiment.Results.StatisticalSig = true
		}

		// Track best performing variant
		if results.ConversionRate > bestConversionRate {
			bestConversionRate = results.ConversionRate
			bestVariantID = variantID
		}
	}

	// Calculate lift
	if controlResults.ConversionRate > 0 && bestConversionRate > controlResults.ConversionRate {
		experiment.Results.LiftPercent = ((bestConversionRate - controlResults.ConversionRate) / controlResults.ConversionRate) * 100
		experiment.Results.Winner = bestVariantID
	}

	// Calculate confidence
	if experiment.Results.StatisticalSig {
		experiment.Results.Confidence = 0.95
	} else {
		experiment.Results.Confidence = 0.5 // Not significant
	}
}

func (s *PerformanceOptimizationService) calculateZScore(control, variant *VariantResults) float64 {
	if control.Participants == 0 || variant.Participants == 0 {
		return 0
	}

	p1 := control.ConversionRate
	p2 := variant.ConversionRate
	n1 := float64(control.Participants)
	n2 := float64(variant.Participants)

	// Pooled standard error
	pooledP := (float64(control.Conversions) + float64(variant.Conversions)) / (n1 + n2)
	se := math.Sqrt(pooledP * (1 - pooledP) * (1/n1 + 1/n2))

	if se == 0 {
		return 0
	}

	return (p2 - p1) / se
}

// Optimization processing routine
func (s *PerformanceOptimizationService) optimizationProcessingRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.processOptimizationRules()
		s.updatePerformanceMetrics()
	}
}

func (s *PerformanceOptimizationService) processOptimizationRules() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, rule := range s.optimizationRules {
		if !rule.Enabled {
			continue
		}

		for _, experiment := range s.experiments {
			if experiment.Status == "running" && rule.Condition(experiment) {
				err := rule.Action(experiment)
				if err != nil {
					log.Printf("❌ Failed to execute optimization rule %s: %v", rule.Name, err)
				} else {
					log.Printf("⚡ Executed optimization rule %s for experiment %s", rule.Name, experiment.Name)
				}
			}
		}
	}
}

func (s *PerformanceOptimizationService) updatePerformanceMetrics() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metrics := s.performanceMetrics
	metrics.TotalExperiments = len(s.experiments)
	metrics.ActiveExperiments = 0
	metrics.CompletedTests = 0
	metrics.SuccessfulTests = 0
	totalLift := 0.0
	totalDuration := 0.0

	// Reset type metrics
	for key := range metrics.MetricsByType {
		delete(metrics.MetricsByType, key)
	}

	for _, experiment := range s.experiments {
		switch experiment.Status {
		case "running":
			metrics.ActiveExperiments++
		case "completed":
			metrics.CompletedTests++

			if experiment.Results.StatisticalSig && experiment.Results.LiftPercent > 0 {
				metrics.SuccessfulTests++
				totalLift += experiment.Results.LiftPercent
			}

			// Calculate duration
			if !experiment.EndDate.IsZero() {
				duration := experiment.EndDate.Sub(experiment.StartDate).Hours() / 24
				totalDuration += duration
			}

			// Update type metrics
			typeMetrics, exists := metrics.MetricsByType[experiment.Type]
			if !exists {
				typeMetrics = &TypeMetrics{}
				metrics.MetricsByType[experiment.Type] = typeMetrics
			}
			typeMetrics.TotalTests++
			if experiment.Results.StatisticalSig && experiment.Results.LiftPercent > 0 {
				typeMetrics.SuccessfulTests++
				typeMetrics.AvgLift += experiment.Results.LiftPercent
			}
		}
	}

	// Calculate averages
	if metrics.SuccessfulTests > 0 {
		metrics.OverallLift = totalLift / float64(metrics.SuccessfulTests)
	}

	if metrics.CompletedTests > 0 {
		metrics.AvgTestDuration = totalDuration / float64(metrics.CompletedTests)
	}

	// Calculate type averages
	for _, typeMetrics := range metrics.MetricsByType {
		if typeMetrics.SuccessfulTests > 0 {
			typeMetrics.AvgLift /= float64(typeMetrics.SuccessfulTests)
		}
	}

	metrics.LastUpdated = time.Now()
}

// Initialize optimization rules
func (s *PerformanceOptimizationService) initializeOptimizationRules() {
	s.optimizationRules = []OptimizationRule{
		{
			ID:   "auto_winner_selection",
			Name: "Auto Winner Selection",
			Type: "auto_winner",
			Condition: func(experiment *Experiment) bool {
				// Auto-select winner if statistically significant and running for 7+ days
				if time.Since(experiment.StartDate) < 7*24*time.Hour {
					return false
				}

				s.calculateExperimentResults(experiment)
				return experiment.Results.StatisticalSig && experiment.Results.LiftPercent > 5
			},
			Action: func(experiment *Experiment) error {
				log.Printf("🏆 Auto-selecting winner for experiment %s: %s", experiment.Name, experiment.Results.Winner)
				return s.StopExperiment(experiment.ID)
			},
			Enabled:  true,
			Priority: 1,
		},
		{
			ID:   "early_stop_poor_performance",
			Name: "Early Stop Poor Performance",
			Type: "early_stop",
			Condition: func(experiment *Experiment) bool {
				// Stop if running for 14+ days with no statistical significance
				if time.Since(experiment.StartDate) < 14*24*time.Hour {
					return false
				}

				s.calculateExperimentResults(experiment)
				return !experiment.Results.StatisticalSig && experiment.Results.TotalParticipants > experiment.MinSampleSize
			},
			Action: func(experiment *Experiment) error {
				log.Printf("⏹️ Early stopping experiment %s due to poor performance", experiment.Name)
				return s.StopExperiment(experiment.ID)
			},
			Enabled:  true,
			Priority: 2,
		},
		{
			ID:   "traffic_reallocation",
			Name: "Dynamic Traffic Reallocation",
			Type: "traffic_allocation",
			Condition: func(experiment *Experiment) bool {
				// Reallocate traffic if one variant is clearly winning
				if time.Since(experiment.StartDate) < 3*24*time.Hour {
					return false
				}

				s.calculateExperimentResults(experiment)
				return experiment.Results.LiftPercent > 10 && experiment.Results.Confidence > 0.8
			},
			Action: func(experiment *Experiment) error {
				// Increase traffic to winning variant
				winnerID := experiment.Results.Winner
				for i, variant := range experiment.Variants {
					if variant.ID == winnerID {
						experiment.Variants[i].Weight = 0.7 // 70% to winner
					} else if variant.IsControl {
						experiment.Variants[i].Weight = 0.3 // 30% to control
					} else {
						experiment.Variants[i].Weight = 0.0 // 0% to other variants
					}
				}
				log.Printf("🔄 Reallocated traffic for experiment %s to favor winner", experiment.Name)
				return nil
			},
			Enabled:  false, // Disabled by default as it can affect statistical validity
			Priority: 3,
		},
	}
}

// Initialize default experiments
func (s *PerformanceOptimizationService) initializeDefaultExperiments() {
	// Email Subject Line Test
	emailSubjectTest := &Experiment{
		ID:           "email_subject_test_001",
		Name:         "Welcome Email Subject Line Test",
		Description:  "Testing different subject lines for welcome emails",
		Type:         "email_subject",
		TrafficSplit: 0.5,
		Variants: []ExperimentVariant{
			{
				ID:          "control",
				Name:        "Control - Welcome to Elite Property Showings",
				Description: "Original subject line",
				Content: map[string]interface{}{
					"subject": "Welcome to Elite Property Showings",
				},
				Weight:    0.5,
				IsControl: true,
			},
			{
				ID:          "variant_a",
				Name:        "Variant A - Your Dream Home Awaits",
				Description: "More emotional subject line",
				Content: map[string]interface{}{
					"subject": "Your Dream Home Awaits - Welcome!",
				},
				Weight:    0.5,
				IsControl: false,
			},
		},
		TargetMetric:    "open_rate",
		MinSampleSize:   100,
		ConfidenceLevel: 0.95,
	}
	s.CreateExperiment(emailSubjectTest)

	// Landing Page CTA Test
	ctaTest := &Experiment{
		ID:           "landing_cta_test_001",
		Name:         "Landing Page CTA Button Test",
		Description:  "Testing different CTA button text and colors",
		Type:         "landing_page",
		TrafficSplit: 0.5,
		Variants: []ExperimentVariant{
			{
				ID:          "control",
				Name:        "Control - Book Your Showing",
				Description: "Original CTA button",
				Content: map[string]interface{}{
					"cta_text":  "Book Your Showing",
					"cta_color": "#3B82F6",
				},
				Weight:    0.5,
				IsControl: true,
			},
			{
				ID:          "variant_a",
				Name:        "Variant A - Schedule Free Tour",
				Description: "More compelling CTA",
				Content: map[string]interface{}{
					"cta_text":  "Schedule Free Tour",
					"cta_color": "#EF4444",
				},
				Weight:    0.5,
				IsControl: false,
			},
		},
		TargetMetric:    "conversion_rate",
		MinSampleSize:   200,
		ConfidenceLevel: 0.95,
	}
	s.CreateExperiment(ctaTest)

	// Booking Form Flow Test
	formFlowTest := &Experiment{
		ID:           "booking_form_test_001",
		Name:         "Booking Form Flow Test",
		Description:  "Testing single-page vs multi-step booking form",
		Type:         "form_flow",
		TrafficSplit: 0.5,
		Variants: []ExperimentVariant{
			{
				ID:          "control",
				Name:        "Control - Multi-step Form",
				Description: "Original 3-step booking form",
				Content: map[string]interface{}{
					"form_type": "multi_step",
					"steps":     3,
				},
				Weight:    0.5,
				IsControl: true,
			},
			{
				ID:          "variant_a",
				Name:        "Variant A - Single Page Form",
				Description: "All fields on one page",
				Content: map[string]interface{}{
					"form_type": "single_page",
					"steps":     1,
				},
				Weight:    0.5,
				IsControl: false,
			},
		},
		TargetMetric:    "conversion_rate",
		MinSampleSize:   300,
		ConfidenceLevel: 0.95,
	}
	s.CreateExperiment(formFlowTest)
}

// Public API methods

func (s *PerformanceOptimizationService) GetAllExperiments() map[string]*Experiment {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.experiments
}

func (s *PerformanceOptimizationService) GetExperiment(experimentID string) (*Experiment, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	experiment, exists := s.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}
	return experiment, nil
}

func (s *PerformanceOptimizationService) GetActiveExperiments() []*Experiment {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var active []*Experiment
	for _, experiment := range s.experiments {
		if experiment.Status == "running" {
			active = append(active, experiment)
		}
	}
	return active
}

func (s *PerformanceOptimizationService) GetPerformanceMetrics() *OptimizationMetrics {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.performanceMetrics
}

func (s *PerformanceOptimizationService) GetExperimentResults(experimentID string) (*ExperimentResults, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	experiment, exists := s.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	// Calculate latest results
	s.calculateExperimentResults(experiment)
	return experiment.Results, nil
}

func (s *PerformanceOptimizationService) GetUserTests(userID string) []*ActiveTest {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var userTests []*ActiveTest
	for _, test := range s.activeTests {
		if test.UserID == userID {
			userTests = append(userTests, test)
		}
	}
	return userTests
}

// Advanced analytics methods

func (s *PerformanceOptimizationService) GetConversionFunnel(experimentID string) (map[string][]FunnelStep, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	experiment, exists := s.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	funnel := make(map[string][]FunnelStep)

	// Analyze conversion funnel for each variant
	for _, variant := range experiment.Variants {
		steps := s.calculateFunnelSteps(experimentID, variant.ID)
		funnel[variant.ID] = steps
	}

	return funnel, nil
}

type FunnelStep struct {
	Step        string  `json:"step"`
	Users       int     `json:"users"`
	Conversions int     `json:"conversions"`
	Rate        float64 `json:"rate"`
}

func (s *PerformanceOptimizationService) calculateFunnelSteps(experimentID, variantID string) []FunnelStep {
	// This would analyze user events to build conversion funnel
	// For now, returning mock data
	return []FunnelStep{
		{Step: "Landing", Users: 1000, Conversions: 800, Rate: 0.8},
		{Step: "Form Start", Users: 800, Conversions: 400, Rate: 0.5},
		{Step: "Form Complete", Users: 400, Conversions: 320, Rate: 0.8},
		{Step: "Booking", Users: 320, Conversions: 256, Rate: 0.8},
	}
}

func (s *PerformanceOptimizationService) GetSegmentedResults(experimentID string, segmentBy string) (map[string]*ExperimentResults, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// This would segment results by different criteria (device, location, source, etc.)
	// For now, returning mock segmented data
	segments := map[string]*ExperimentResults{
		"desktop": {
			TotalParticipants: 500,
			VariantResults: map[string]*VariantResults{
				"control":   {ConversionRate: 0.12, Participants: 250, Conversions: 30},
				"variant_a": {ConversionRate: 0.16, Participants: 250, Conversions: 40},
			},
		},
		"mobile": {
			TotalParticipants: 300,
			VariantResults: map[string]*VariantResults{
				"control":   {ConversionRate: 0.08, Participants: 150, Conversions: 12},
				"variant_a": {ConversionRate: 0.14, Participants: 150, Conversions: 21},
			},
		},
	}

	return segments, nil
}

// Integration with other services

func (s *PerformanceOptimizationService) OptimizeEmailCampaign(campaignID string) error {
	// Get best performing email variant and apply to campaign
	experiment, err := s.GetExperimentForUser("system", "email_subject")
	if err != nil {
		return err
	}

	// Apply winning variant to email campaign
	log.Printf("📧 Optimizing email campaign %s with variant %s", campaignID, experiment.Name)
	return nil
}

func (s *PerformanceOptimizationService) OptimizeLandingPage(pageID string) error {
	// Get best performing landing page variant
	experiment, err := s.GetExperimentForUser("system", "landing_page")
	if err != nil {
		return err
	}

	// Apply winning variant to landing page
	log.Printf("🎯 Optimizing landing page %s with variant %s", pageID, experiment.Name)
	return nil
}

func (s *PerformanceOptimizationService) GetOptimizationRecommendations() []OptimizationRecommendation {
	recommendations := []OptimizationRecommendation{
		{
			Type:        "email_optimization",
			Title:       "Improve Email Open Rates",
			Description: "Test personalized subject lines to increase open rates by 15-25%",
			Impact:      "High",
			Effort:      "Low",
			Priority:    1,
		},
		{
			Type:        "form_optimization",
			Title:       "Simplify Booking Form",
			Description: "Reduce form fields to increase completion rates by 20-30%",
			Impact:      "High",
			Effort:      "Medium",
			Priority:    2,
		},
		{
			Type:        "cta_optimization",
			Title:       "Optimize Call-to-Action Buttons",
			Description: "Test different button colors and text to improve click rates",
			Impact:      "Medium",
			Effort:      "Low",
			Priority:    3,
		},
	}

	return recommendations
}

type OptimizationRecommendation struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
	Priority    int    `json:"priority"`
}
