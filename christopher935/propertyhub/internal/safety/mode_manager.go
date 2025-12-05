package safety

import (
	"fmt"
	"log"
	"time"
)

// SafetyModeManager handles transitions between safety modes
type SafetyModeManager struct {
	configManager *SafetyConfigManager
	transitionLog []ModeTransition
}

// ModeTransition records a safety mode change
type ModeTransition struct {
	FromMode       SafetyMode `json:"from_mode"`
	ToMode         SafetyMode `json:"to_mode"`
	Timestamp      time.Time  `json:"timestamp"`
	ModifiedBy     string     `json:"modified_by"`
	Reason         string     `json:"reason"`
	AutoTransition bool       `json:"auto_transition"`
}

// ModeRecommendation suggests when to transition between modes
type ModeRecommendation struct {
	CurrentMode        SafetyMode    `json:"current_mode"`
	RecommendedMode    SafetyMode    `json:"recommended_mode"`
	Confidence         float64       `json:"confidence"` // 0-1 confidence in recommendation
	Reason             string        `json:"reason"`
	RequiredMetrics    []string      `json:"required_metrics"` // Metrics that support the recommendation
	RiskFactors        []string      `json:"risk_factors"`     // Factors that increase risk
	TimeInCurrentMode  time.Duration `json:"time_in_current_mode"`
	ReadyForTransition bool          `json:"ready_for_transition"`
}

// SafetyMetrics tracks automation performance for mode recommendations
type SafetyMetrics struct {
	TotalCampaigns        int        `json:"total_campaigns"`
	SuccessfulCampaigns   int        `json:"successful_campaigns"`
	FailedCampaigns       int        `json:"failed_campaigns"`
	ComplaintCount        int        `json:"complaint_count"`
	UnsubscribeCount      int        `json:"unsubscribe_count"`
	SuccessRate           float64    `json:"success_rate"`
	ComplaintRate         float64    `json:"complaint_rate"`
	UnsubscribeRate       float64    `json:"unsubscribe_rate"`
	AverageEngagementRate float64    `json:"average_engagement_rate"`
	DaysInCurrentMode     int        `json:"days_in_current_mode"`
	LastIncident          *time.Time `json:"last_incident,omitempty"`
}

// NewSafetyModeManager creates a new safety mode manager
func NewSafetyModeManager(configManager *SafetyConfigManager) *SafetyModeManager {
	return &SafetyModeManager{
		configManager: configManager,
		transitionLog: []ModeTransition{},
	}
}

// TransitionToMode safely transitions to a new safety mode
func (m *SafetyModeManager) TransitionToMode(newMode SafetyMode, modifiedBy, reason string) error {
	currentMode := m.configManager.GetConfig().Mode

	// Validate transition
	if err := m.validateTransition(currentMode, newMode); err != nil {
		return fmt.Errorf("invalid transition: %v", err)
	}

	// Log the transition
	transition := ModeTransition{
		FromMode:       currentMode,
		ToMode:         newMode,
		Timestamp:      time.Now(),
		ModifiedBy:     modifiedBy,
		Reason:         reason,
		AutoTransition: false,
	}

	// Perform the transition
	if err := m.configManager.UpdateMode(newMode, modifiedBy); err != nil {
		return fmt.Errorf("failed to update mode: %v", err)
	}

	// Record the transition
	m.transitionLog = append(m.transitionLog, transition)

	// Log the change
	log.Printf("Safety mode transition: %s → %s by %s. Reason: %s",
		m.getModeString(currentMode),
		m.getModeString(newMode),
		modifiedBy,
		reason)

	// Provide guidance for the new mode
	m.logModeGuidance(newMode)

	return nil
}

// validateTransition ensures the mode transition is safe and logical
func (m *SafetyModeManager) validateTransition(from, to SafetyMode) error {
	// Always allow transitions to stricter modes (emergency situations)
	if to < from {
		return nil
	}

	// Validate relaxing transitions
	switch from {
	case SafetyModeStrict:
		if to > SafetyModeModerate {
			return fmt.Errorf("cannot jump from Strict to %s - must transition through Moderate first", m.getModeString(to))
		}
	case SafetyModeModerate:
		if to == SafetyModeOff {
			return fmt.Errorf("cannot jump from Moderate to Off - must transition through Relaxed first")
		}
	case SafetyModeRelaxed:
		// Can transition to Off from Relaxed
		break
	case SafetyModeOff:
		// Already at minimum safety
		break
	}

	return nil
}

// GetModeRecommendation analyzes current performance and recommends mode changes
func (m *SafetyModeManager) GetModeRecommendation(metrics SafetyMetrics) *ModeRecommendation {
	currentMode := m.configManager.GetConfig().Mode

	recommendation := &ModeRecommendation{
		CurrentMode:        currentMode,
		RecommendedMode:    currentMode, // Default to no change
		Confidence:         0.5,
		RequiredMetrics:    []string{},
		RiskFactors:        []string{},
		TimeInCurrentMode:  time.Duration(metrics.DaysInCurrentMode) * 24 * time.Hour,
		ReadyForTransition: false,
	}

	// Analyze metrics for recommendations
	switch currentMode {
	case SafetyModeStrict:
		recommendation = m.analyzeStrictMode(metrics, recommendation)
	case SafetyModeModerate:
		recommendation = m.analyzeModerateMod(metrics, recommendation)
	case SafetyModeRelaxed:
		recommendation = m.analyzeRelaxedMode(metrics, recommendation)
	case SafetyModeOff:
		recommendation = m.analyzeOffMode(metrics, recommendation)
	}

	return recommendation
}

// analyzeStrictMode analyzes performance in strict mode
func (m *SafetyModeManager) analyzeStrictMode(metrics SafetyMetrics, rec *ModeRecommendation) *ModeRecommendation {
	// Criteria for moving from Strict to Moderate
	positiveFactors := 0

	// Check success rate
	if metrics.SuccessRate >= 0.95 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "High success rate (95%+)")
		positiveFactors++
	}

	// Check complaint rate
	if metrics.ComplaintRate <= 0.01 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "Low complaint rate (<1%)")
		positiveFactors++
	}

	// Check time in mode
	if metrics.DaysInCurrentMode >= 14 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "Sufficient time in strict mode (14+ days)")
		positiveFactors++
	}

	// Check campaign volume
	if metrics.TotalCampaigns >= 10 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "Sufficient campaign experience (10+ campaigns)")
		positiveFactors++
	}

	// Check for risk factors
	if metrics.ComplaintCount > 0 {
		rec.RiskFactors = append(rec.RiskFactors, "Recent complaints received")
	}

	if float64(metrics.FailedCampaigns) > float64(metrics.SuccessfulCampaigns)*0.1 {
		rec.RiskFactors = append(rec.RiskFactors, "High failure rate")
	}

	// Make recommendation
	if positiveFactors >= 3 && len(rec.RiskFactors) == 0 {
		rec.RecommendedMode = SafetyModeModerate
		rec.Confidence = 0.8
		rec.Reason = "Strong performance metrics support transition to Moderate mode"
		rec.ReadyForTransition = true
	} else if positiveFactors >= 2 {
		rec.Confidence = 0.6
		rec.Reason = "Good performance, but recommend more time in Strict mode"
	} else {
		rec.Confidence = 0.3
		rec.Reason = "Insufficient performance data or concerning metrics"
	}

	return rec
}

// analyzeModerateMod analyzes performance in moderate mode
func (m *SafetyModeManager) analyzeModerateMod(metrics SafetyMetrics, rec *ModeRecommendation) *ModeRecommendation {
	positiveFactors := 0

	// Higher bar for moving to Relaxed
	if metrics.SuccessRate >= 0.97 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "Very high success rate (97%+)")
		positiveFactors++
	}

	if metrics.ComplaintRate <= 0.005 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "Very low complaint rate (<0.5%)")
		positiveFactors++
	}

	if metrics.DaysInCurrentMode >= 21 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "Extended time in moderate mode (21+ days)")
		positiveFactors++
	}

	if metrics.TotalCampaigns >= 25 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "Extensive campaign experience (25+ campaigns)")
		positiveFactors++
	}

	if metrics.AverageEngagementRate >= 0.15 {
		rec.RequiredMetrics = append(rec.RequiredMetrics, "High engagement rate (15%+)")
		positiveFactors++
	}

	// Check for risk factors
	if metrics.ComplaintCount > 1 {
		rec.RiskFactors = append(rec.RiskFactors, "Multiple complaints received")
	}

	if metrics.UnsubscribeRate > 0.02 {
		rec.RiskFactors = append(rec.RiskFactors, "High unsubscribe rate")
	}

	// Recommendation logic
	if positiveFactors >= 4 && len(rec.RiskFactors) == 0 {
		rec.RecommendedMode = SafetyModeRelaxed
		rec.Confidence = 0.85
		rec.Reason = "Excellent performance metrics support transition to Relaxed mode"
		rec.ReadyForTransition = true
	} else if metrics.SuccessRate < 0.9 || metrics.ComplaintRate > 0.02 {
		rec.RecommendedMode = SafetyModeStrict
		rec.Confidence = 0.7
		rec.Reason = "Performance concerns suggest returning to Strict mode"
		rec.ReadyForTransition = true
	} else {
		rec.Confidence = 0.5
		rec.Reason = "Continue in Moderate mode to build more performance history"
	}

	return rec
}

// analyzeRelaxedMode analyzes performance in relaxed mode
func (m *SafetyModeManager) analyzeRelaxedMode(metrics SafetyMetrics, rec *ModeRecommendation) *ModeRecommendation {
	// Very high bar for turning off safety
	if metrics.SuccessRate >= 0.98 &&
		metrics.ComplaintRate <= 0.001 &&
		metrics.DaysInCurrentMode >= 30 &&
		metrics.TotalCampaigns >= 50 &&
		metrics.ComplaintCount == 0 {

		rec.RecommendedMode = SafetyModeOff
		rec.Confidence = 0.9
		rec.Reason = "Exceptional performance history supports disabling safety protections"
		rec.ReadyForTransition = true
		rec.RequiredMetrics = append(rec.RequiredMetrics,
			"98%+ success rate",
			"<0.1% complaint rate",
			"30+ days experience",
			"50+ campaigns",
			"Zero recent complaints")
	} else if metrics.SuccessRate < 0.95 || metrics.ComplaintRate > 0.01 {
		rec.RecommendedMode = SafetyModeModerate
		rec.Confidence = 0.8
		rec.Reason = "Performance decline suggests returning to Moderate mode"
		rec.ReadyForTransition = true
		rec.RiskFactors = append(rec.RiskFactors, "Declining performance metrics")
	} else {
		rec.Confidence = 0.6
		rec.Reason = "Continue building performance history in Relaxed mode"
	}

	return rec
}

// analyzeOffMode analyzes performance with safety off
func (m *SafetyModeManager) analyzeOffMode(metrics SafetyMetrics, rec *ModeRecommendation) *ModeRecommendation {
	// Monitor for any issues that require re-enabling safety
	if metrics.ComplaintCount > 0 || metrics.SuccessRate < 0.95 {
		rec.RecommendedMode = SafetyModeRelaxed
		rec.Confidence = 0.9
		rec.Reason = "Performance issues detected - recommend re-enabling safety protections"
		rec.ReadyForTransition = true
		rec.RiskFactors = append(rec.RiskFactors, "Performance degradation with safety off")
	} else {
		rec.Confidence = 0.7
		rec.Reason = "Continue monitoring performance with safety disabled"
	}

	return rec
}

// AutoTransitionCheck checks if automatic transitions should be triggered
func (m *SafetyModeManager) AutoTransitionCheck(metrics SafetyMetrics) error {
	recommendation := m.GetModeRecommendation(metrics)

	// Only auto-transition to stricter modes (safety concerns)
	if recommendation.RecommendedMode < recommendation.CurrentMode &&
		recommendation.Confidence >= 0.8 {

		err := m.TransitionToMode(
			recommendation.RecommendedMode,
			"system",
			fmt.Sprintf("Auto-transition due to: %s", recommendation.Reason))

		if err != nil {
			return fmt.Errorf("auto-transition failed: %v", err)
		}

		// Mark as auto-transition
		if len(m.transitionLog) > 0 {
			m.transitionLog[len(m.transitionLog)-1].AutoTransition = true
		}

		log.Printf("Auto-transition executed: %s → %s",
			m.getModeString(recommendation.CurrentMode),
			m.getModeString(recommendation.RecommendedMode))
	}

	return nil
}

// GetTransitionHistory returns the history of mode transitions
func (m *SafetyModeManager) GetTransitionHistory() []ModeTransition {
	return m.transitionLog
}

// logModeGuidance provides guidance for operating in the new mode
func (m *SafetyModeManager) logModeGuidance(mode SafetyMode) {
	switch mode {
	case SafetyModeStrict:
		log.Printf("STRICT MODE GUIDANCE: All campaigns require approval. Start with small, low-risk communications to build confidence.")
	case SafetyModeModerate:
		log.Printf("MODERATE MODE GUIDANCE: Low-risk campaigns auto-approved. Monitor performance closely for 2-3 weeks.")
	case SafetyModeRelaxed:
		log.Printf("RELAXED MODE GUIDANCE: Most campaigns auto-approved. Continue monitoring engagement and complaint rates.")
	case SafetyModeOff:
		log.Printf("SAFETY OFF WARNING: All protections disabled. Monitor performance very closely and be prepared to re-enable safety.")
	}
}

// getModeString returns string representation of safety mode
func (m *SafetyModeManager) getModeString(mode SafetyMode) string {
	switch mode {
	case SafetyModeStrict:
		return "Strict"
	case SafetyModeModerate:
		return "Moderate"
	case SafetyModeRelaxed:
		return "Relaxed"
	case SafetyModeOff:
		return "Off"
	default:
		return "Unknown"
	}
}

// GetModeTransitionPlan returns a plan for gradually relaxing safety
func (m *SafetyModeManager) GetModeTransitionPlan() map[string]interface{} {
	currentMode := m.configManager.GetConfig().Mode

	plan := map[string]interface{}{
		"current_mode": m.getModeString(currentMode),
		"phases":       []map[string]interface{}{},
	}

	switch currentMode {
	case SafetyModeStrict:
		plan["phases"] = []map[string]interface{}{
			{
				"phase":    1,
				"mode":     "Strict",
				"duration": "2-4 weeks",
				"goals": []string{
					"Build confidence with manual approvals",
					"Achieve 95%+ success rate",
					"Complete 10+ campaigns",
					"Zero complaints",
				},
				"next_mode": "Moderate",
			},
			{
				"phase":    2,
				"mode":     "Moderate",
				"duration": "3-4 weeks",
				"goals": []string{
					"Maintain 97%+ success rate",
					"Complete 25+ campaigns",
					"<0.5% complaint rate",
					"High engagement rates",
				},
				"next_mode": "Relaxed",
			},
			{
				"phase":    3,
				"mode":     "Relaxed",
				"duration": "4+ weeks",
				"goals": []string{
					"Maintain 98%+ success rate",
					"Complete 50+ campaigns",
					"<0.1% complaint rate",
					"Consistent performance",
				},
				"next_mode": "Off (Optional)",
			},
		}
	case SafetyModeModerate:
		plan["phases"] = []map[string]interface{}{
			{
				"phase":    1,
				"mode":     "Moderate",
				"duration": "Current",
				"goals": []string{
					"Maintain excellent performance",
					"Build campaign experience",
				},
				"next_mode": "Relaxed",
			},
		}
	}

	return plan
}

// EmergencyRevert immediately reverts to strict mode
func (m *SafetyModeManager) EmergencyRevert(reason string) error {
	return m.TransitionToMode(SafetyModeStrict, "emergency-system", fmt.Sprintf("Emergency revert: %s", reason))
}
