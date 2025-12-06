package services

import (
	"log"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// BehavioralScoringEngine calculates and manages behavioral scores for leads
type BehavioralScoringEngine struct {
	db           *gorm.DB
	scoringRules *ScoringRules
	notificationHub *AdminNotificationHub
}

// NewBehavioralScoringEngine creates a new scoring engine
func NewBehavioralScoringEngine(db *gorm.DB) *BehavioralScoringEngine {
	return &BehavioralScoringEngine{
		db:           db,
		scoringRules: DefaultScoringRules(),
	}
}

// SetNotificationHub sets the notification hub for sending real-time alerts
func (e *BehavioralScoringEngine) SetNotificationHub(hub *AdminNotificationHub) {
	e.notificationHub = hub
}

// CalculateScore calculates the behavioral score for a lead based on all their events
func (e *BehavioralScoringEngine) CalculateScore(leadID int64) (*models.BehavioralScore, error) {
	// Get all events for this lead
	var events []models.BehavioralEvent
	if err := e.db.Where("lead_id = ?", leadID).
		Order("created_at DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}

	// Calculate component scores (0-100 each)
	urgencyScore := e.calculateUrgencyScore(events)
	engagementScore := e.calculateEngagementScore(events)
	financialScore := e.calculateFinancialScore(events)
	
	// Calculate composite score (weighted average, 0-100)
	compositeScore := int(
		(float64(urgencyScore) * 0.40) +
		(float64(engagementScore) * 0.40) +
		(float64(financialScore) * 0.20),
	)

	// Build score factors JSON
	scoreFactors := map[string]interface{}{
		"urgency_score":    urgencyScore,
		"engagement_score": engagementScore,
		"financial_score":  financialScore,
		"total_events":     len(events),
		"segment":          e.determineSegment(compositeScore),
	}

	// Create or update score record
	leadIDInt := int(leadID)
	score := &models.BehavioralScore{
		LeadID:          &leadIDInt,
		UrgencyScore:    urgencyScore,
		EngagementScore: engagementScore,
		FinancialScore:  financialScore,
		CompositeScore:  compositeScore,
		ScoreFactors:    scoreFactors,
		LastCalculated:  time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := e.SaveScore(score); err != nil {
		return nil, err
	}

	log.Printf("✅ Calculated score for lead %d: %d (Segment: %s)", leadID, compositeScore, e.determineSegment(compositeScore))
	return score, nil
}

// calculateUrgencyScore measures how urgent/hot the lead is (0-100)
func (e *BehavioralScoringEngine) calculateUrgencyScore(events []models.BehavioralEvent) int {
	if len(events) == 0 {
		return 0
	}

	score := 0.0
	now := time.Now()

	for _, event := range events {
		points := e.scoringRules.GetPoints(event.EventType)
		daysSince := now.Sub(event.CreatedAt).Hours() / 24
		decayFactor := e.calculateDecayFactor(daysSince)
		score += float64(points) * decayFactor
	}

	// Normalize to 0-100
	if score > 100 {
		score = 100
	}
	return int(score)
}

// calculateEngagementScore measures overall engagement level (0-100)
func (e *BehavioralScoringEngine) calculateEngagementScore(events []models.BehavioralEvent) int {
	if len(events) == 0 {
		return 0
	}

	// Frequency score (more events = higher engagement)
	frequencyScore := float64(len(events)) * 2.0
	if frequencyScore > 50 {
		frequencyScore = 50
	}

	// Recency score (recent activity = higher engagement)
	recencyScore := 0.0
	if len(events) > 0 {
		daysSinceLastActivity := time.Since(events[0].CreatedAt).Hours() / 24
		if daysSinceLastActivity < 1 {
			recencyScore = 50
		} else if daysSinceLastActivity < 7 {
			recencyScore = 30
		} else if daysSinceLastActivity < 30 {
			recencyScore = 10
		}
	}

	score := frequencyScore + recencyScore
	if score > 100 {
		score = 100
	}
	return int(score)
}

// calculateFinancialScore estimates financial readiness (0-100)
func (e *BehavioralScoringEngine) calculateFinancialScore(events []models.BehavioralEvent) int {
	// Check for high-intent actions
	score := 0
	for _, event := range events {
		if event.EventType == "application" {
			score += 50
		} else if event.EventType == "inquiry" {
			score += 20
		}
	}

	if score > 100 {
		score = 100
	}
	return score
}

// calculateDecayFactor applies time decay to event scores
func (e *BehavioralScoringEngine) calculateDecayFactor(daysSince float64) float64 {
	if daysSince < 1 {
		return 1.0
	} else if daysSince < 7 {
		return 0.8
	} else if daysSince < 30 {
		return 0.5
	} else if daysSince < 90 {
		return 0.2
	}
	return 0.1
}

// determineSegment assigns a segment based on composite score
func (e *BehavioralScoringEngine) determineSegment(compositeScore int) string {
	if compositeScore >= 70 {
		return "hot"
	} else if compositeScore >= 40 {
		return "warm"
	} else if compositeScore >= 10 {
		return "cold"
	}
	return "dormant"
}

// SaveScore saves or updates a behavioral score
func (e *BehavioralScoringEngine) SaveScore(score *models.BehavioralScore) error {
	// Check if score exists for this lead
	var existing models.BehavioralScore
	err := e.db.Where("lead_id = ?", score.LeadID).First(&existing).Error
	
	newSegment := e.determineSegment(score.CompositeScore)
	var previousSegment string
	
	if err == gorm.ErrRecordNotFound {
		// Create new score
		score.CreatedAt = time.Now()
		if err := e.db.Create(score).Error; err != nil {
			return err
		}
		previousSegment = ""
	} else if err != nil {
		return err
	} else {
		// Update existing score
		previousSegment = e.determineSegment(existing.CompositeScore)
		score.ID = existing.ID
		score.CreatedAt = existing.CreatedAt
		if err := e.db.Save(score).Error; err != nil {
			return err
		}
	}
	
	// Check if lead just became hot
	if e.notificationHub != nil && newSegment == "hot" && previousSegment != "hot" {
		var lead models.Lead
		if err := e.db.First(&lead, score.LeadID).Error; err == nil {
			leadName := lead.FirstName + " " + lead.LastName
			e.notificationHub.SendHotLeadAlert(leadName, score.CompositeScore, int64(*score.LeadID))
		}
	}
	
	return nil
}

// GetScore retrieves the current score for a lead
func (e *BehavioralScoringEngine) GetScore(leadID int64) (*models.BehavioralScore, error) {
	var score models.BehavioralScore
	leadIDInt := int(leadID)
	err := e.db.Where("lead_id = ?", &leadIDInt).First(&score).Error
	return &score, err
}

// RecalculateAllScores recalculates scores for all leads (batch operation)
func (e *BehavioralScoringEngine) RecalculateAllScores() error {
	var leads []models.Lead
	if err := e.db.Find(&leads).Error; err != nil {
		return err
	}

	log.Printf("🔄 Recalculating scores for %d leads...", len(leads))
	
	for _, lead := range leads {
		if _, err := e.CalculateScore(int64(lead.ID)); err != nil {
			log.Printf("⚠️ Failed to recalculate score for lead %d: %v", lead.ID, err)
		}
	}

	log.Printf("✅ Batch recalculation complete")
	return nil
}
