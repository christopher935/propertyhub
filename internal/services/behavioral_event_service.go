package services

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// BehavioralEventService handles behavioral event tracking
type BehavioralEventService struct {
	db                  *gorm.DB
	scoringEngine       *BehavioralScoringEngine
	notificationService *AdminNotificationService
}

// NewBehavioralEventService creates a new behavioral event service
func NewBehavioralEventService(db *gorm.DB) *BehavioralEventService {
	return &BehavioralEventService{
		db:            db,
		scoringEngine: NewBehavioralScoringEngine(db),
	}
}

// SetNotificationService sets the notification service (for avoiding circular dependency)
func (s *BehavioralEventService) SetNotificationService(notificationService *AdminNotificationService) {
	s.notificationService = notificationService
}

// ============================================================================
// EVENT TRACKING (WITH AUTOMATIC SCORING)
// ============================================================================

// TrackEvent logs a behavioral event and triggers score recalculation
func (s *BehavioralEventService) TrackEvent(leadID int64, eventType string, eventData map[string]interface{}, propertyID *int64, sessionID string, ipAddress string, userAgent string) error {
	event := models.BehavioralEvent{
		LeadID:     leadID,
		EventType:  eventType,
		EventData:  eventData,
		PropertyID: propertyID,
		SessionID:  sessionID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	if err := s.db.Create(&event).Error; err != nil {
		log.Printf("❌ Failed to track event %s for lead %d: %v", eventType, leadID, err)
		return err
	}

	log.Printf("✅ Tracked event: %s for lead %d", eventType, leadID)

	// Get previous score for comparison
	previousScore, _ := s.scoringEngine.GetScore(leadID)
	previousCompositeScore := 0
	if previousScore != nil {
		previousCompositeScore = previousScore.CompositeScore
	}

	// Trigger score recalculation and notifications asynchronously
	go func() {
		newScore, err := s.scoringEngine.CalculateScore(leadID)
		if err != nil {
			log.Printf("⚠️  Failed to recalculate score for lead %d: %v", leadID, err)
			return
		}

		if s.notificationService == nil {
			return
		}

		// Check for hot lead
		if newScore.CompositeScore >= 70 && previousCompositeScore < 70 {
			s.notificationService.OnHotLeadActive(leadID, sessionID)
		}

		// Check for engagement spike
		scoreDelta := newScore.CompositeScore - previousCompositeScore
		if scoreDelta >= 20 {
			s.notificationService.OnEngagementSpike(leadID, scoreDelta, newScore.CompositeScore)
		}

		// Check for multiple property views in session
		if eventType == "viewed" && sessionID != "" {
			var viewCount int64
			s.db.Model(&models.BehavioralEvent{}).
				Where("lead_id = ? AND session_id = ? AND event_type = ?", leadID, sessionID, "viewed").
				Count(&viewCount)
			if viewCount >= 5 {
				s.notificationService.OnMultiplePropertiesViewed(leadID, int(viewCount), sessionID)
			}
		}
	}()

	return nil
}

// TrackPropertyView logs a property view event
func (s *BehavioralEventService) TrackPropertyView(leadID int64, propertyID int64, sessionID string, ipAddress string, userAgent string) error {
	eventData := map[string]interface{}{
		"property_id": propertyID,
		"action":      "view",
	}
	return s.TrackEvent(leadID, "viewed", eventData, &propertyID, sessionID, ipAddress, userAgent)
}

// TrackPropertySave logs a property save event
func (s *BehavioralEventService) TrackPropertySave(leadID int64, propertyID int64, sessionID string, ipAddress string, userAgent string) error {
	eventData := map[string]interface{}{
		"property_id": propertyID,
		"action":      "save",
	}
	err := s.TrackEvent(leadID, "saved", eventData, &propertyID, sessionID, ipAddress, userAgent)
	if err == nil && s.notificationService != nil {
		go s.notificationService.OnPropertySaved(leadID, propertyID)
	}
	return err
}

// TrackInquiry logs an inquiry/contact form submission
func (s *BehavioralEventService) TrackInquiry(leadID int64, propertyID *int64, inquiryType string, sessionID string, ipAddress string, userAgent string) error {
	eventData := map[string]interface{}{
		"inquiry_type": inquiryType,
		"action":       "inquiry",
	}
	if propertyID != nil {
		eventData["property_id"] = *propertyID
	}
	err := s.TrackEvent(leadID, "inquired", eventData, propertyID, sessionID, ipAddress, userAgent)
	if err == nil && s.notificationService != nil {
		go s.notificationService.OnInquirySent(leadID, propertyID)
	}
	return err
}

// TrackApplication logs an application submission
func (s *BehavioralEventService) TrackApplication(leadID int64, propertyID int64, applicationID string, sessionID string, ipAddress string, userAgent string) error {
	eventData := map[string]interface{}{
		"property_id":    propertyID,
		"application_id": applicationID,
		"action":         "apply",
	}
	return s.TrackEvent(leadID, "applied", eventData, &propertyID, sessionID, ipAddress, userAgent)
}

// TrackConversion logs a conversion (lease signed)
func (s *BehavioralEventService) TrackConversion(leadID int64, propertyID int64, leaseID string, sessionID string, ipAddress string, userAgent string) error {
	eventData := map[string]interface{}{
		"property_id": propertyID,
		"lease_id":    leaseID,
		"action":      "convert",
	}
	return s.TrackEvent(leadID, "converted", eventData, &propertyID, sessionID, ipAddress, userAgent)
}

// TrackEmailEngagement logs email opens/clicks from FUB webhooks
func (s *BehavioralEventService) TrackEmailEngagement(leadID int64, eventType string, campaignID string, emailID string) error {
	eventData := map[string]interface{}{
		"campaign_id": campaignID,
		"email_id":    emailID,
		"source":      "fub_webhook",
	}
	return s.TrackEvent(leadID, eventType, eventData, nil, "", "", "")
}

// ============================================================================
// SESSION TRACKING
// ============================================================================

// StartSession creates a new behavioral session
func (s *BehavioralEventService) StartSession(leadID int64, ipAddress string, userAgent string, referrer string) (string, error) {
	sessionID := uuid.New().String()
	
	session := models.BehavioralSession{
		ID:        sessionID,
		LeadID:    leadID,
		StartTime: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referrer:  referrer,
	}

	if err := s.db.Create(&session).Error; err != nil {
		return "", err
	}

	// Track session start event
	eventData := map[string]interface{}{
		"action":   "session_start",
		"referrer": referrer,
	}
	go s.TrackEvent(leadID, "session_start", eventData, nil, sessionID, ipAddress, userAgent)

	return sessionID, nil
}

// EndSession closes a behavioral session
func (s *BehavioralEventService) EndSession(sessionID string) error {
	var session models.BehavioralSession
	if err := s.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return err
	}

	endTime := time.Now()
	duration := int(endTime.Sub(session.StartTime).Seconds())

	updates := map[string]interface{}{
		"end_time":         endTime,
		"duration_seconds": duration,
	}

	return s.db.Model(&session).Updates(updates).Error
}

// UpdateSession updates session metrics
func (s *BehavioralEventService) UpdateSession(sessionID string, pageViews int, interactions int) error {
	updates := map[string]interface{}{
		"page_views":   pageViews,
		"interactions": interactions,
	}
	return s.db.Model(&models.BehavioralSession{}).
		Where("id = ?", sessionID).
		Updates(updates).Error
}

// ============================================================================
// FUNNEL TRACKING
// ============================================================================

// TrackFunnelStage logs progression through the conversion funnel
func (s *BehavioralEventService) TrackFunnelStage(leadID int64, stage string, propertyID *int64, metadata map[string]interface{}) error {
	funnelEvent := models.ConversionFunnelEvent{
		LeadID:     leadID,
		Stage:      stage,
		PropertyID: propertyID,
		Metadata:   metadata,
	}

	return s.db.Create(&funnelEvent).Error
}

// CompleteFunnelStage marks a funnel stage as complete
func (s *BehavioralEventService) CompleteFunnelStage(leadID int64, stage string, converted bool) error {
	exitTime := time.Now()
	
	var funnelEvent models.ConversionFunnelEvent
	if err := s.db.Where("lead_id = ? AND stage = ? AND exited_at IS NULL", leadID, stage).
		First(&funnelEvent).Error; err != nil {
		return err
	}

	timeInStage := int(exitTime.Sub(funnelEvent.EnteredAt).Seconds())

	updates := map[string]interface{}{
		"exited_at":            exitTime,
		"converted":            converted,
		"time_in_stage_seconds": timeInStage,
	}

	return s.db.Model(&funnelEvent).Updates(updates).Error
}

// ============================================================================
// SEGMENT ASSIGNMENT
// ============================================================================

// AssignSegment assigns a lead to a behavioral segment
func (s *BehavioralEventService) AssignSegment(leadID int64, segment string, segmentData map[string]interface{}) error {
	segmentRecord := models.BehavioralSegment{
		LeadID:      leadID,
		Segment:     segment,
		SegmentData: segmentData,
	}

	return s.db.Create(&segmentRecord).Error
}

// ============================================================================
// SCORE MANAGEMENT
// ============================================================================

// SaveBehavioralScores saves behavioral scores to history
func (s *BehavioralEventService) SaveBehavioralScores(leadID int64, urgencyScore float64, financialScore float64, engagementScore float64, overallScore int, scoreFactors map[string]interface{}) error {
	scoreHistory := models.BehavioralScoreHistory{
		LeadID:          leadID,
		UrgencyScore:    urgencyScore,
		FinancialScore:  financialScore,
		EngagementScore: engagementScore,
		OverallScore:    overallScore,
		ScoreFactors:    scoreFactors,
	}

	return s.db.Create(&scoreHistory).Error
}

// GetCurrentScore retrieves the current score for a lead
func (s *BehavioralEventService) GetCurrentScore(leadID int64) (*models.BehavioralScore, error) {
	return s.scoringEngine.GetScore(leadID)
}

// RecalculateScore manually triggers score recalculation for a lead
func (s *BehavioralEventService) RecalculateScore(leadID int64) (*models.BehavioralScore, error) {
	return s.scoringEngine.CalculateScore(leadID)
}
