package services

import (
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// FunnelAnalyticsService analyzes conversion funnel performance
type FunnelAnalyticsService struct {
	db *gorm.DB
}

// NewFunnelAnalyticsService creates a new funnel analytics service
func NewFunnelAnalyticsService(db *gorm.DB) *FunnelAnalyticsService {
	return &FunnelAnalyticsService{
		db: db,
	}
}

// FunnelStage represents a stage in the conversion funnel
type FunnelStage struct {
	Name            string  `json:"name"`
	LeadCount       int     `json:"lead_count"`
	ConversionRate  float64 `json:"conversion_rate"` // % that move to next stage
	DropOffRate     float64 `json:"drop_off_rate"`
	AvgTimeInStage  float64 `json:"avg_time_in_stage_hours"`
	BottleneckScore int     `json:"bottleneck_score"` // 0-100, higher = bigger bottleneck
}

// FunnelAnalysis contains complete funnel performance data
type FunnelAnalysis struct {
	Stages            []FunnelStage  `json:"stages"`
	OverallConversion float64        `json:"overall_conversion_rate"`
	TotalLeads        int            `json:"total_leads"`
	Converted         int            `json:"converted"`
	DropOffs          map[string]int `json:"drop_offs"`   // Stage name -> count
	Bottlenecks       []string       `json:"bottlenecks"` // Stages with issues
	Insights          []string       `json:"insights"`
	AnalyzedAt        time.Time      `json:"analyzed_at"`
}

// DropOffAnalysis contains detailed drop-off information
type DropOffAnalysis struct {
	Stage           string          `json:"stage"`
	DropOffCount    int             `json:"drop_off_count"`
	DropOffRate     float64         `json:"drop_off_rate"`
	CommonReasons   []DropOffReason `json:"common_reasons"`
	AffectedLeads   []models.Lead   `json:"affected_leads"`
	Recommendations []string        `json:"recommendations"`
}

// DropOffReason represents why leads drop off at a stage
type DropOffReason struct {
	Reason     string  `json:"reason"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// AnalyzeFunnel performs complete funnel analysis
func (fas *FunnelAnalyticsService) AnalyzeFunnel(timeRange int) (*FunnelAnalysis, error) {
	log.Println("📊 Funnel Analytics: Analyzing conversion funnel...")

	startDate := time.Now().AddDate(0, 0, -timeRange)

	// Define funnel stages
	stageNames := []string{
		"inquiry",
		"property_view",
		"property_save",
		"showing_requested",
		"showing_completed",
		"application_started",
		"application_submitted",
		"lease_signed",
	}

	stages := []FunnelStage{}
	dropOffs := make(map[string]int)
	var totalLeads, converted int

	// Analyze each stage
	for i, stageName := range stageNames {
		stage, dropOffCount, err := fas.analyzeStage(stageName, startDate)
		if err != nil {
			log.Printf("Error analyzing stage %s: %v", stageName, err)
			continue
		}

		stages = append(stages, stage)
		dropOffs[stageName] = dropOffCount

		if i == 0 {
			totalLeads = stage.LeadCount
		}
		if stageName == "lease_signed" {
			converted = stage.LeadCount
		}
	}

	// Calculate overall conversion
	overallConversion := 0.0
	if totalLeads > 0 {
		overallConversion = (float64(converted) / float64(totalLeads)) * 100
	}

	// Identify bottlenecks
	bottlenecks := fas.identifyBottlenecks(stages)

	// Generate insights
	insights := fas.generateFunnelInsights(stages, bottlenecks, overallConversion)

	analysis := &FunnelAnalysis{
		Stages:            stages,
		OverallConversion: overallConversion,
		TotalLeads:        totalLeads,
		Converted:         converted,
		DropOffs:          dropOffs,
		Bottlenecks:       bottlenecks,
		Insights:          insights,
		AnalyzedAt:        time.Now(),
	}

	log.Printf("✅ Funnel analysis complete: %.1f%% overall conversion, %d bottlenecks identified", overallConversion, len(bottlenecks))

	return analysis, nil
}

// analyzeStage analyzes a single funnel stage
func (fas *FunnelAnalyticsService) analyzeStage(stageName string, startDate time.Time) (FunnelStage, int, error) {
	stage := FunnelStage{Name: stageName}

	// Count leads at this stage
	var leadCount int64
	err := fas.db.Model(&models.BehavioralEvent{}).
		Where("event_type = ?", stageName).
		Where("created_at >= ?", startDate).
		Distinct("lead_id").
		Count(&leadCount).Error

	if err != nil {
		return stage, 0, err
	}

	stage.LeadCount = int(leadCount)

	// Get next stage for conversion calculation
	nextStage := fas.getNextStage(stageName)
	if nextStage != "" {
		var nextStageCount int64
		fas.db.Model(&models.BehavioralEvent{}).
			Where("event_type = ?", nextStage).
			Where("created_at >= ?", startDate).
			Distinct("lead_id").
			Count(&nextStageCount)

		if leadCount > 0 {
			stage.ConversionRate = (float64(nextStageCount) / float64(leadCount)) * 100
			stage.DropOffRate = 100 - stage.ConversionRate
		}
	}

	// Calculate average time in stage
	avgTime, err := fas.calculateAvgTimeInStage(stageName, nextStage, startDate)
	if err == nil {
		stage.AvgTimeInStage = avgTime
	}

	// Calculate bottleneck score
	stage.BottleneckScore = fas.calculateBottleneckScore(stage)

	dropOffCount := int(leadCount) - int(float64(leadCount)*stage.ConversionRate/100)

	return stage, dropOffCount, nil
}

// calculateAvgTimeInStage calculates average time leads spend in a stage
func (fas *FunnelAnalyticsService) calculateAvgTimeInStage(currentStage, nextStage string, startDate time.Time) (float64, error) {
	if nextStage == "" {
		return 0, nil
	}

	query := `
		SELECT AVG(EXTRACT(EPOCH FROM (next_event.created_at - current_event.created_at))/3600) as avg_hours
		FROM behavioral_events current_event
		JOIN behavioral_events next_event ON current_event.lead_id = next_event.lead_id
		WHERE current_event.event_type = ?
			AND next_event.event_type = ?
			AND current_event.created_at >= ?
			AND next_event.created_at > current_event.created_at
	`

	var avgHours float64
	err := fas.db.Raw(query, currentStage, nextStage, startDate).Scan(&avgHours).Error

	return avgHours, err
}

// calculateBottleneckScore calculates how much of a bottleneck this stage is
func (fas *FunnelAnalyticsService) calculateBottleneckScore(stage FunnelStage) int {
	score := 0

	// High drop-off rate = bottleneck
	if stage.DropOffRate > 50 {
		score += 40
	} else if stage.DropOffRate > 30 {
		score += 20
	}

	// Long time in stage = bottleneck
	if stage.AvgTimeInStage > 72 { // 3 days
		score += 30
	} else if stage.AvgTimeInStage > 48 { // 2 days
		score += 15
	}

	// Low conversion rate = bottleneck
	if stage.ConversionRate < 30 {
		score += 30
	} else if stage.ConversionRate < 50 {
		score += 15
	}

	if score > 100 {
		score = 100
	}

	return score
}

// identifyBottlenecks identifies stages that are bottlenecks
func (fas *FunnelAnalyticsService) identifyBottlenecks(stages []FunnelStage) []string {
	bottlenecks := []string{}

	for _, stage := range stages {
		if stage.BottleneckScore >= 50 {
			bottlenecks = append(bottlenecks, stage.Name)
		}
	}

	return bottlenecks
}

// generateFunnelInsights generates actionable insights from funnel analysis
func (fas *FunnelAnalyticsService) generateFunnelInsights(stages []FunnelStage, bottlenecks []string, overallConversion float64) []string {
	insights := []string{}

	// Overall conversion insight
	if overallConversion < 5 {
		insights = append(insights, fmt.Sprintf("⚠️ Overall conversion rate is very low (%.1f%%). Multiple stages need optimization.", overallConversion))
	} else if overallConversion < 10 {
		insights = append(insights, fmt.Sprintf("📊 Overall conversion rate is %.1f%%. Industry average is 10-15%%. Room for improvement.", overallConversion))
	} else {
		insights = append(insights, fmt.Sprintf("✅ Overall conversion rate is healthy at %.1f%%.", overallConversion))
	}

	// Bottleneck insights
	for _, bottleneck := range bottlenecks {
		for _, stage := range stages {
			if stage.Name == bottleneck {
				if stage.DropOffRate > 50 {
					insights = append(insights, fmt.Sprintf("🚨 %s stage has %.1f%% drop-off rate - this is a critical bottleneck!", stage.Name, stage.DropOffRate))
				}
				if stage.AvgTimeInStage > 72 {
					insights = append(insights, fmt.Sprintf("⏱️ Leads spend %.1f hours in %s stage - consider faster follow-up", stage.AvgTimeInStage, stage.Name))
				}
			}
		}
	}

	// Stage-specific insights
	for i, stage := range stages {
		if i > 0 {
			prevStage := stages[i-1]
			if stage.LeadCount < prevStage.LeadCount/2 {
				insights = append(insights, fmt.Sprintf("📉 Major drop-off between %s and %s (%d → %d leads)", prevStage.Name, stage.Name, prevStage.LeadCount, stage.LeadCount))
			}
		}
	}

	return insights
}

// GetLeadsAtStage returns leads currently at a specific stage
func (fas *FunnelAnalyticsService) GetLeadsAtStage(stages []string) ([]models.Lead, error) {
	leads := []models.Lead{}

	// Get lead IDs at these stages
	var leadIDs []int64
	err := fas.db.Model(&models.BehavioralEvent{}).
		Where("event_type IN ?", stages).
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)).
		Distinct("lead_id").
		Pluck("lead_id", &leadIDs).Error

	if err != nil {
		return leads, err
	}

	// Get full lead records
	err = fas.db.Where("id IN ?", leadIDs).Find(&leads).Error

	return leads, err
}

// AnalyzeDropOff analyzes why leads drop off at a specific stage
func (fas *FunnelAnalyticsService) AnalyzeDropOff(stageName string, timeRange int) (*DropOffAnalysis, error) {
	startDate := time.Now().AddDate(0, 0, -timeRange)

	// Get leads that reached this stage
	var stageLeadIDs []int64
	fas.db.Model(&models.BehavioralEvent{}).
		Where("event_type = ?", stageName).
		Where("created_at >= ?", startDate).
		Distinct("lead_id").
		Pluck("lead_id", &stageLeadIDs)

	// Get leads that progressed to next stage
	nextStage := fas.getNextStage(stageName)
	var progressedLeadIDs []int64
	if nextStage != "" {
		fas.db.Model(&models.BehavioralEvent{}).
			Where("event_type = ?", nextStage).
			Where("created_at >= ?", startDate).
			Distinct("lead_id").
			Pluck("lead_id", &progressedLeadIDs)
	}

	// Find drop-offs (leads at stage but not at next stage)
	dropOffIDs := []int64{}
	progressedMap := make(map[int64]bool)
	for _, id := range progressedLeadIDs {
		progressedMap[id] = true
	}
	for _, id := range stageLeadIDs {
		if !progressedMap[id] {
			dropOffIDs = append(dropOffIDs, id)
		}
	}

	dropOffCount := len(dropOffIDs)
	dropOffRate := 0.0
	if len(stageLeadIDs) > 0 {
		dropOffRate = (float64(dropOffCount) / float64(len(stageLeadIDs))) * 100
	}

	// Get affected leads
	var affectedLeads []models.Lead
	if len(dropOffIDs) > 0 {
		fas.db.Where("id IN ?", dropOffIDs).Limit(50).Find(&affectedLeads)
	}

	// Analyze common reasons (based on behavioral patterns)
	reasons := fas.analyzeDropOffReasons(dropOffIDs, stageName)

	// Generate recommendations
	recommendations := fas.generateDropOffRecommendations(stageName, dropOffRate, reasons)

	analysis := &DropOffAnalysis{
		Stage:           stageName,
		DropOffCount:    dropOffCount,
		DropOffRate:     dropOffRate,
		CommonReasons:   reasons,
		AffectedLeads:   affectedLeads,
		Recommendations: recommendations,
	}

	return analysis, nil
}

// analyzeDropOffReasons analyzes common patterns in drop-offs
func (fas *FunnelAnalyticsService) analyzeDropOffReasons(leadIDs []int64, stage string) []DropOffReason {
	reasons := []DropOffReason{}

	if len(leadIDs) == 0 {
		return reasons
	}

	total := len(leadIDs)

	// Reason 1: No follow-up contact
	var noContactCount int64
	fas.db.Model(&models.Lead{}).
		Where("id IN ?", leadIDs).
		Where("last_contact < ?", time.Now().AddDate(0, 0, -7)).
		Count(&noContactCount)

	if noContactCount > 0 {
		reasons = append(reasons, DropOffReason{
			Reason:     "No follow-up contact in 7+ days",
			Count:      int(noContactCount),
			Percentage: (float64(noContactCount) / float64(total)) * 100,
		})
	}

	// Reason 2: Low engagement score
	var lowEngagementCount int64
	fas.db.Table("behavioral_scores").
		Where("lead_id IN ?", leadIDs).
		Where("engagement_score < ?", 30).
		Count(&lowEngagementCount)

	if lowEngagementCount > 0 {
		reasons = append(reasons, DropOffReason{
			Reason:     "Low engagement score (< 30)",
			Count:      int(lowEngagementCount),
			Percentage: (float64(lowEngagementCount) / float64(total)) * 100,
		})
	}

	// Reason 3: Long response time
	// This would require tracking response times - placeholder for now

	return reasons
}

// generateDropOffRecommendations generates recommendations to reduce drop-offs
func (fas *FunnelAnalyticsService) generateDropOffRecommendations(stage string, dropOffRate float64, reasons []DropOffReason) []string {
	recommendations := []string{}

	if dropOffRate > 50 {
		recommendations = append(recommendations, fmt.Sprintf("🚨 Critical: %.1f%% drop-off at %s stage. Immediate action required.", dropOffRate, stage))
	}

	for _, reason := range reasons {
		if reason.Percentage > 30 {
			switch {
			case reason.Reason == "No follow-up contact in 7+ days":
				recommendations = append(recommendations, "📧 Implement automated follow-up emails at this stage")
				recommendations = append(recommendations, "⏰ Set up reminders to contact leads within 24-48 hours")
			case reason.Reason == "Low engagement score (< 30)":
				recommendations = append(recommendations, "🎯 Improve lead qualification to focus on higher-intent leads")
				recommendations = append(recommendations, "💡 Create more engaging content for this stage")
			}
		}
	}

	// Stage-specific recommendations
	switch stage {
	case "property_view":
		recommendations = append(recommendations, "📸 Ensure properties have high-quality photos")
		recommendations = append(recommendations, "📝 Add detailed property descriptions")
	case "showing_requested":
		recommendations = append(recommendations, "⚡ Reduce response time to showing requests")
		recommendations = append(recommendations, "📅 Offer flexible showing times")
	case "application_started":
		recommendations = append(recommendations, "📋 Simplify application process")
		recommendations = append(recommendations, "💬 Provide application assistance")
	}

	return recommendations
}

// getNextStage returns the next stage in the funnel
func (fas *FunnelAnalyticsService) getNextStage(currentStage string) string {
	stageOrder := map[string]string{
		"inquiry":               "property_view",
		"property_view":         "property_save",
		"property_save":         "showing_requested",
		"showing_requested":     "showing_completed",
		"showing_completed":     "application_started",
		"application_started":   "application_submitted",
		"application_submitted": "lease_signed",
		"lease_signed":          "",
	}

	return stageOrder[currentStage]
}

// GetConversionProbabilityForStage returns the probability of converting from a given stage
func (fas *FunnelAnalyticsService) GetConversionProbabilityForStage(stage string) float64 {
	// Historical conversion rates by stage
	conversionRates := map[string]float64{
		"inquiry":               0.15,
		"property_view":         0.25,
		"property_save":         0.40,
		"showing_requested":     0.55,
		"showing_completed":     0.75,
		"application_started":   0.85,
		"application_submitted": 0.95,
		"lease_signed":          1.00,
	}

	if rate, exists := conversionRates[stage]; exists {
		return rate
	}

	return 0.10 // Default low probability
}
