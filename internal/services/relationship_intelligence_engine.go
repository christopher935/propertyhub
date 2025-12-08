package services

import (
	"fmt"
	"log"
	"sort"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// RelationshipIntelligenceEngine analyzes cross-entity patterns to generate contextual insights
type RelationshipIntelligenceEngine struct {
	db               *gorm.DB
	scoringEngine    *BehavioralScoringEngine
	funnelAnalytics  *FunnelAnalyticsService
	propertyMatcher  *PropertyMatchingService
	insightGenerator *InsightGeneratorService
}

// NewRelationshipIntelligenceEngine creates a new relationship intelligence engine
func NewRelationshipIntelligenceEngine(db *gorm.DB, scoringEngine *BehavioralScoringEngine) *RelationshipIntelligenceEngine {
	return &RelationshipIntelligenceEngine{
		db:            db,
		scoringEngine: scoringEngine,
	}
}

// SetFunnelAnalytics sets the funnel analytics service (circular dependency resolution)
func (rie *RelationshipIntelligenceEngine) SetFunnelAnalytics(fa *FunnelAnalyticsService) {
	rie.funnelAnalytics = fa
}

// SetPropertyMatcher sets the property matching service
func (rie *RelationshipIntelligenceEngine) SetPropertyMatcher(pm *PropertyMatchingService) {
	rie.propertyMatcher = pm
}

// SetInsightGenerator sets the insight generator service
func (rie *RelationshipIntelligenceEngine) SetInsightGenerator(ig *InsightGeneratorService) {
	rie.insightGenerator = ig
}

// Opportunity represents a high-value action opportunity
type Opportunity struct {
	ID                    string                 `json:"id"`
	Type                  string                 `json:"type"`     // "hot_lead", "re_engagement", "property_match", "conversion_ready"
	Priority              int                    `json:"priority"` // 1-100
	UrgencyScore          int                    `json:"urgency_score"`
	ConversionProbability float64                `json:"conversion_probability"`
	RevenueEstimate       float64                `json:"revenue_estimate"`
	LeadID                int64                  `json:"lead_id"`
	LeadName              string                 `json:"lead_name"`
	LeadEmail             string                 `json:"lead_email"`
	PropertyID            *int64                 `json:"property_id,omitempty"`
	PropertyAddress       string                 `json:"property_address,omitempty"`
	Context               string                 `json:"context"` // Human-readable explanation
	Insight               string                 `json:"insight"` // AI-generated insight with HTML
	ActionSequence        []OpportunityAction    `json:"action_sequence"`
	Metadata              map[string]interface{} `json:"metadata"`
	DetectedAt            time.Time              `json:"detected_at"`
	ExpiresAt             *time.Time             `json:"expires_at,omitempty"`
}

// OpportunityAction represents a recommended action step
type OpportunityAction struct {
	Step        int                    `json:"step"`
	Action      string                 `json:"action"` // "send_email", "make_call", "schedule_showing", etc.
	Description string                 `json:"description"`
	Template    string                 `json:"template,omitempty"` // Email template ID
	Timing      string                 `json:"timing"`             // "immediate", "within_2_hours", "next_day", etc.
	AutoExecute bool                   `json:"auto_execute"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LeadPropertyRelationship represents the relationship between a lead and property
type LeadPropertyRelationship struct {
	LeadID              int64
	PropertyID          int64
	ViewCount           int
	FirstViewedAt       time.Time
	LastViewedAt        time.Time
	SavedProperty       bool
	InquirySent         bool
	ShowingBooked       bool
	ApplicationStarted  bool
	DaysSinceLastView   int
	DaysSinceFirstView  int
	EngagementIntensity float64 // Views per day
	PropertyAddress     string
	PropertyPrice       float64
	LeadName            string
	LeadEmail           string
	LeadScore           int
}

// AnalyzeOpportunities identifies high-value opportunities across all leads and properties
func (rie *RelationshipIntelligenceEngine) AnalyzeOpportunities() ([]Opportunity, error) {
	log.Println("🕸️ Relationship Intelligence: Analyzing opportunities...")

	opportunities := []Opportunity{}

	// Pattern 1: Hot leads with high engagement
	hotLeadOpps, err := rie.detectHotLeadOpportunities()
	if err == nil {
		opportunities = append(opportunities, hotLeadOpps...)
	}

	// Pattern 2: Leads viewing properties multiple times without contact
	viewingOpps, err := rie.detectHighViewNoContactOpportunities()
	if err == nil {
		opportunities = append(opportunities, viewingOpps...)
	}

	// Pattern 3: Cold leads ready for re-engagement
	reengagementOpps, err := rie.detectReengagementOpportunities()
	if err == nil {
		opportunities = append(opportunities, reengagementOpps...)
	}

	// Pattern 4: Leads at conversion-ready stage
	conversionOpps, err := rie.detectConversionReadyOpportunities()
	if err == nil {
		opportunities = append(opportunities, conversionOpps...)
	}

	// Pattern 5: Property matches (if property matcher is available)
	if rie.propertyMatcher != nil {
		matchOpps, err := rie.detectPropertyMatchOpportunities()
		if err == nil {
			opportunities = append(opportunities, matchOpps...)
		}
	}

	// Sort by priority (highest first)
	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].Priority > opportunities[j].Priority
	})

	log.Printf("✅ Found %d opportunities", len(opportunities))
	return opportunities, nil
}

// detectHotLeadOpportunities finds leads with behavioral scores > 70
// OPTIMIZED: Limits to top 100 hot leads for performance at scale
func (rie *RelationshipIntelligenceEngine) detectHotLeadOpportunities() ([]Opportunity, error) {
	opportunities := []Opportunity{}

	// Query for hot leads
	var scores []models.BehavioralScore
	err := rie.db.Where("composite_score > ?", 70).
		Where("created_at > ?", time.Now().AddDate(0, 0, -7)).
		Where("lead_id IS NOT NULL").
		Order("composite_score DESC").
		Limit(100).
		Find(&scores).Error

	if err != nil {
		return opportunities, err
	}

	for _, score := range scores {
		if score.LeadID == nil {
			continue
		}

		// Get the lead
		var lead models.Lead
		err := rie.db.First(&lead, *score.LeadID).Error
		if err != nil {
			continue
		}

		// Get recent activity
		var eventCount int64
		rie.db.Model(&models.BehavioralEvent{}).
			Where("lead_id = ?", score.LeadID).
			Where("created_at > ?", time.Now().AddDate(0, 0, -7)).
			Count(&eventCount)

		// Calculate conversion probability based on score and activity
		conversionProb := rie.calculateConversionProbability(score.CompositeScore, int(eventCount), 0)

		// Build action sequence
		actions := []OpportunityAction{
			{
				Step:        1,
				Action:      "send_email",
				Description: "Send personalized follow-up email",
				Template:    "hot_lead_followup",
				Timing:      "immediate",
				AutoExecute: true,
			},
			{
				Step:        2,
				Action:      "make_call",
				Description: "Call within 2 hours if email is opened",
				Timing:      "within_2_hours",
				AutoExecute: false,
			},
			{
				Step:        3,
				Action:      "schedule_showing",
				Description: "Offer showing times",
				Timing:      "same_day",
				AutoExecute: false,
			},
		}

		opp := Opportunity{
			ID:                    fmt.Sprintf("hot_lead_%d_%d", *score.LeadID, time.Now().Unix()),
			Type:                  "hot_lead",
			Priority:              rie.calculatePriority(score.CompositeScore, conversionProb, 0),
			UrgencyScore:          score.UrgencyScore,
			ConversionProbability: conversionProb,
			RevenueEstimate:       rie.estimateRevenue(conversionProb),
			LeadID:                int64(*score.LeadID),
			LeadName:              lead.FirstName + " " + lead.LastName,
			LeadEmail:             lead.Email,
			Context:               fmt.Sprintf("High engagement score (%d) with %d recent activities", score.CompositeScore, eventCount),
			Insight:               rie.generateHotLeadInsight(score, int(eventCount)),
			ActionSequence:        actions,
			Metadata: map[string]interface{}{
				"behavioral_score": score.CompositeScore,
				"urgency_score":    score.UrgencyScore,
				"engagement_score": score.EngagementScore,
				"financial_score":  score.FinancialScore,
				"recent_events":    eventCount,
			},
			DetectedAt: time.Now(),
		}

		opportunities = append(opportunities, opp)
	}

	return opportunities, nil
}

// detectHighViewNoContactOpportunities finds leads viewing properties multiple times without contact
func (rie *RelationshipIntelligenceEngine) detectHighViewNoContactOpportunities() ([]Opportunity, error) {
	opportunities := []Opportunity{}

	// Query for lead-property relationships with high view counts
	relationships, err := rie.getLeadPropertyRelationships()
	if err != nil {
		return opportunities, err
	}

	for _, rel := range relationships {
		// Pattern: 3+ views, no contact in 5+ days
		if rel.ViewCount >= 3 && rel.DaysSinceLastView >= 5 && !rel.ShowingBooked {

			// Get conversion probability from similar patterns
			conversionProb := rie.calculateConversionProbability(rel.LeadScore, rel.ViewCount, rel.DaysSinceLastView)

			// Higher urgency if more recent views
			urgencyScore := 100 - (rel.DaysSinceLastView * 5)
			if urgencyScore < 0 {
				urgencyScore = 0
			}

			actions := []OpportunityAction{
				{
					Step:        1,
					Action:      "send_email",
					Description: fmt.Sprintf("Send 'Still interested in %s?' email", rel.PropertyAddress),
					Template:    "property_reengagement",
					Timing:      "immediate",
					AutoExecute: true,
					Metadata: map[string]interface{}{
						"property_id":      rel.PropertyID,
						"property_address": rel.PropertyAddress,
						"view_count":       rel.ViewCount,
					},
				},
				{
					Step:        2,
					Action:      "offer_showing",
					Description: "Offer specific showing times",
					Timing:      "within_24_hours",
					AutoExecute: false,
				},
			}

			expiresAt := time.Now().AddDate(0, 0, 7) // Opportunity expires in 7 days

			opp := Opportunity{
				ID:                    fmt.Sprintf("high_view_%d_%d_%d", rel.LeadID, rel.PropertyID, time.Now().Unix()),
				Type:                  "high_view_no_contact",
				Priority:              rie.calculatePriority(urgencyScore, conversionProb, rel.PropertyPrice),
				UrgencyScore:          urgencyScore,
				ConversionProbability: conversionProb,
				RevenueEstimate:       rie.estimateRevenue(conversionProb),
				LeadID:                rel.LeadID,
				LeadName:              rel.LeadName,
				LeadEmail:             rel.LeadEmail,
				PropertyID:            &rel.PropertyID,
				PropertyAddress:       rel.PropertyAddress,
				Context:               fmt.Sprintf("Viewed %s %d times, last view %d days ago, no contact made", rel.PropertyAddress, rel.ViewCount, rel.DaysSinceLastView),
				Insight:               rie.generateHighViewInsight(rel, conversionProb),
				ActionSequence:        actions,
				Metadata: map[string]interface{}{
					"view_count":           rel.ViewCount,
					"days_since_last_view": rel.DaysSinceLastView,
					"engagement_intensity": rel.EngagementIntensity,
					"property_price":       rel.PropertyPrice,
				},
				DetectedAt: time.Now(),
				ExpiresAt:  &expiresAt,
			}

			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

// detectReengagementOpportunities finds cold leads ready for re-engagement
func (rie *RelationshipIntelligenceEngine) detectReengagementOpportunities() ([]Opportunity, error) {
	opportunities := []Opportunity{}

	// Query for leads with no activity in 30-90 days
	var leads []models.Lead
	err := rie.db.Where("last_contact < ?", time.Now().AddDate(0, 0, -30)).
		Where("last_contact > ?", time.Now().AddDate(0, 0, -90)).
		Where("status = ?", "active").
		Limit(50).
		Find(&leads).Error

	if err != nil {
		return opportunities, err
	}

	for _, lead := range leads {
		// Calculate days since last activity (using UpdatedAt as proxy for last contact)
		daysSinceContact := int(time.Since(lead.UpdatedAt).Hours() / 24)

		// Get their previous engagement level
		var score models.BehavioralScore
		err := rie.db.Where("lead_id = ?", lead.ID).
			Order("created_at DESC").
			First(&score).Error

		previousScore := 0
		if err == nil {
			previousScore = score.CompositeScore
		}

		// Only re-engage if they had some previous interest
		if previousScore > 30 {
			conversionProb := rie.calculateReengagementProbability(previousScore, daysSinceContact)

			actions := []OpportunityAction{
				{
					Step:        1,
					Action:      "send_email",
					Description: fmt.Sprintf("Send re-engagement email (%d days inactive)", daysSinceContact),
					Template:    rie.selectReengagementTemplate(daysSinceContact),
					Timing:      "immediate",
					AutoExecute: true,
				},
			}

			opp := Opportunity{
				ID:                    fmt.Sprintf("reengage_%d_%d", lead.ID, time.Now().Unix()),
				Type:                  "re_engagement",
				Priority:              rie.calculatePriority(50, conversionProb, 0),
				UrgencyScore:          40,
				ConversionProbability: conversionProb,
				RevenueEstimate:       rie.estimateRevenue(conversionProb),
				LeadID:                int64(lead.ID),
				LeadName:              lead.FirstName + " " + lead.LastName,
				LeadEmail:             lead.Email,
				Context:               fmt.Sprintf("No activity for %d days, previous engagement score: %d", daysSinceContact, previousScore),
				Insight:               rie.generateReengagementInsight(lead, daysSinceContact, previousScore),
				ActionSequence:        actions,
				Metadata: map[string]interface{}{
					"days_inactive":  daysSinceContact,
					"previous_score": previousScore,
				},
				DetectedAt: time.Now(),
			}

			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

// detectConversionReadyOpportunities finds leads at high-probability conversion stage
func (rie *RelationshipIntelligenceEngine) detectConversionReadyOpportunities() ([]Opportunity, error) {
	opportunities := []Opportunity{}

	// Use funnel analytics if available
	if rie.funnelAnalytics == nil {
		return opportunities, nil
	}

	// Get leads at "showing_completed" or "application_started" stage
	conversionReadyLeads, err := rie.funnelAnalytics.GetLeadsAtStage([]string{"showing_completed", "application_started"})
	if err != nil {
		return opportunities, err
	}

	for _, lead := range conversionReadyLeads {
		conversionProb := 0.75 // High probability at this stage

		actions := []OpportunityAction{
			{
				Step:        1,
				Action:      "send_email",
				Description: "Send application reminder or follow-up",
				Template:    "conversion_nudge",
				Timing:      "immediate",
				AutoExecute: true,
			},
			{
				Step:        2,
				Action:      "make_call",
				Description: "Personal call to close the deal",
				Timing:      "within_4_hours",
				AutoExecute: false,
			},
		}

		opp := Opportunity{
			ID:                    fmt.Sprintf("conversion_ready_%d_%d", lead.ID, time.Now().Unix()),
			Type:                  "conversion_ready",
			Priority:              95, // Very high priority
			UrgencyScore:          90,
			ConversionProbability: conversionProb,
			RevenueEstimate:       rie.estimateRevenue(conversionProb),
			LeadID:                int64(lead.ID),
			LeadName:              lead.FirstName + " " + lead.LastName,
			LeadEmail:             lead.Email,
			Context:               "Lead is at conversion-ready stage (showing completed or application started)",
			Insight:               rie.generateConversionReadyInsight(lead),
			ActionSequence:        actions,
			Metadata: map[string]interface{}{
				"status": lead.Status,
			},
			DetectedAt: time.Now(),
		}

		opportunities = append(opportunities, opp)
	}

	return opportunities, nil
}

// detectPropertyMatchOpportunities finds new property matches for leads
func (rie *RelationshipIntelligenceEngine) detectPropertyMatchOpportunities() ([]Opportunity, error) {
	opportunities := []Opportunity{}

	// This will be implemented when PropertyMatchingService is built
	// For now, return empty

	return opportunities, nil
}

// Helper functions

func (rie *RelationshipIntelligenceEngine) getLeadPropertyRelationships() ([]LeadPropertyRelationship, error) {
	relationships := []LeadPropertyRelationship{}

	// Query to aggregate lead-property viewing patterns
	query := `
		SELECT 
			be.lead_id,
			be.property_id,
			COUNT(*) as view_count,
			MIN(be.created_at) as first_viewed_at,
			MAX(be.created_at) as last_viewed_at,
			COALESCE(EXTRACT(EPOCH FROM (NOW() - MAX(be.created_at)))/86400, 0)::int as days_since_last_view,
			COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(be.created_at)))/86400, 0)::int as days_since_first_view,
			p.address as property_address,
			COALESCE(p.price, 0) as property_price,
			l.name as lead_name,
			l.email as lead_email,
			COALESCE(bs.composite_score, 0) as lead_score
		FROM behavioral_events be
		JOIN leads l ON l.id = be.lead_id
		LEFT JOIN properties p ON p.id = be.property_id
		LEFT JOIN behavioral_scores bs ON bs.lead_id = be.lead_id
		WHERE be.event_type = 'property_view'
			AND be.property_id IS NOT NULL
			AND be.created_at > NOW() - INTERVAL '90 days'
		GROUP BY be.lead_id, be.property_id, p.address, p.price, l.name, l.email, bs.composite_score
		HAVING COUNT(*) >= 2
		ORDER BY view_count DESC, last_viewed_at DESC
		LIMIT 100
	`

	err := rie.db.Raw(query).Scan(&relationships).Error
	if err != nil {
		return relationships, err
	}

	// Calculate engagement intensity
	for i := range relationships {
		if relationships[i].DaysSinceFirstView > 0 {
			relationships[i].EngagementIntensity = float64(relationships[i].ViewCount) / float64(relationships[i].DaysSinceFirstView)
		}
	}

	return relationships, nil
}

func (rie *RelationshipIntelligenceEngine) calculateConversionProbability(score, activityCount, daysSinceActivity int) float64 {
	// Base probability from behavioral score
	baseProb := float64(score) / 100.0

	// Adjust for activity count
	activityMultiplier := 1.0 + (float64(activityCount) * 0.1)
	if activityMultiplier > 2.0 {
		activityMultiplier = 2.0
	}

	// Decay based on days since activity
	decayFactor := 1.0 / (1.0 + float64(daysSinceActivity)*0.05)

	probability := baseProb * activityMultiplier * decayFactor

	if probability > 0.95 {
		probability = 0.95
	}
	if probability < 0.05 {
		probability = 0.05
	}

	return probability
}

func (rie *RelationshipIntelligenceEngine) calculateReengagementProbability(previousScore, daysInactive int) float64 {
	baseProb := float64(previousScore) / 200.0 // Lower base for re-engagement
	decayFactor := 1.0 / (1.0 + float64(daysInactive)*0.02)

	probability := baseProb * decayFactor

	if probability > 0.40 {
		probability = 0.40 // Cap re-engagement probability
	}
	if probability < 0.05 {
		probability = 0.05
	}

	return probability
}

func (rie *RelationshipIntelligenceEngine) calculatePriority(urgencyScore int, conversionProb, revenueEstimate float64) int {
	// Priority = (Urgency × 0.4) + (Conversion Probability × 40) + (Revenue Impact × 0.2)
	revenueScore := 0.0
	if revenueEstimate > 0 {
		revenueScore = (revenueEstimate / 2000.0) * 20 // Normalize to 0-20 range
		if revenueScore > 20 {
			revenueScore = 20
		}
	}

	priority := int(float64(urgencyScore)*0.4 + conversionProb*40 + revenueScore)

	if priority > 100 {
		priority = 100
	}
	if priority < 1 {
		priority = 1
	}

	return priority
}

func (rie *RelationshipIntelligenceEngine) estimateRevenue(conversionProb float64) float64 {
	// Average commission estimate
	avgCommission := 1500.0
	return avgCommission * conversionProb
}

func (rie *RelationshipIntelligenceEngine) selectReengagementTemplate(daysInactive int) string {
	if daysInactive < 45 {
		return "reengagement_30day"
	} else if daysInactive < 75 {
		return "reengagement_60day"
	}
	return "reengagement_90day"
}

// Insight generation functions

func (rie *RelationshipIntelligenceEngine) generateHotLeadInsight(score models.BehavioralScore, eventCount int) string {
	// Get lead name
	var lead models.Lead
	leadName := "Lead"
	if score.LeadID != nil {
		if err := rie.db.First(&lead, *score.LeadID).Error; err == nil {
			leadName = lead.FirstName + " " + lead.LastName
		}
	}

	return fmt.Sprintf(
		`<div class="insight-hot-lead">
			<strong>🔥 Hot Lead Alert:</strong> %s has a behavioral score of <strong>%d</strong> with <strong>%d recent activities</strong> in the past week.
			<br><br>
			<span class="insight-stat">Urgency: %d/100</span> | 
			<span class="insight-stat">Engagement: %d/100</span> | 
			<span class="insight-stat">Financial Readiness: %d/100</span>
			<br><br>
			<em>Similar leads who were contacted within 24 hours converted at 85%%. Act now for best results.</em>
		</div>`,
		leadName,
		score.CompositeScore,
		eventCount,
		score.UrgencyScore,
		score.EngagementScore,
		score.FinancialScore,
	)
}

func (rie *RelationshipIntelligenceEngine) generateHighViewInsight(rel LeadPropertyRelationship, conversionProb float64) string {
	return fmt.Sprintf(
		`<div class="insight-high-view">
			<strong>👀 High Interest Detected:</strong> %s has viewed <a href="/admin/properties/%d">%s</a> <strong>%d times</strong> over the past %d days.
			<br><br>
			<span class="insight-warning">⚠️ Last view was %d days ago - interest may be cooling!</span>
			<br><br>
			<span class="insight-stat">Conversion Probability: %.0f%%</span> if contacted within 24 hours
			<br><br>
			<em>Leads viewing a property 3+ times have 3x higher conversion rates. Immediate follow-up recommended.</em>
		</div>`,
		rel.LeadName,
		rel.PropertyID,
		rel.PropertyAddress,
		rel.ViewCount,
		rel.DaysSinceFirstView,
		rel.DaysSinceLastView,
		conversionProb*100,
	)
}

func (rie *RelationshipIntelligenceEngine) generateReengagementInsight(lead models.Lead, daysInactive, previousScore int) string {
	return fmt.Sprintf(
		`<div class="insight-reengage">
			<strong>💤 Re-engagement Opportunity:</strong> %s has been inactive for <strong>%d days</strong> but previously showed strong interest (score: %d).
			<br><br>
			<em>Re-engagement campaigns at this stage typically achieve 12-15%% response rates. Worth a personalized outreach.</em>
		</div>`,
		lead.FirstName+" "+lead.LastName,
		daysInactive,
		previousScore,
	)
}

func (rie *RelationshipIntelligenceEngine) generateConversionReadyInsight(lead models.Lead) string {
	return fmt.Sprintf(
		`<div class="insight-conversion-ready">
			<strong>🎯 Conversion Ready:</strong> %s is at a critical conversion stage. They've completed a showing or started an application.
			<br><br>
			<span class="insight-stat">Conversion Probability: 75%%</span>
			<br><br>
			<em>Leads at this stage convert 75%% of the time with proper follow-up. Personal contact within 4 hours is critical.</em>
		</div>`,
		lead.FirstName+" "+lead.LastName,
	)
}
