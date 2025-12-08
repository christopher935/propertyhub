package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// FUBBidirectionalSync handles two-way sync between PropertyHub and Follow Up Boss
type FUBBidirectionalSync struct {
	db                *gorm.DB
	fubAPIKey         string
	fubBaseURL        string
	behavioralService *BehavioralEventService
	scoringEngine     *BehavioralScoringEngine
}

// NewFUBBidirectionalSync creates a new bi-directional sync service
func NewFUBBidirectionalSync(db *gorm.DB, fubAPIKey string) *FUBBidirectionalSync {
	return &FUBBidirectionalSync{
		db:                db,
		fubAPIKey:         fubAPIKey,
		fubBaseURL:        "https://api.followupboss.com/v1",
		behavioralService: NewBehavioralEventService(db),
		scoringEngine:     NewBehavioralScoringEngine(db),
	}
}

// ============================================================================
// PROPERTYHUB → FUB (Action Logging)
// ============================================================================

// LogCallToFUB logs a call made in PropertyHub to FUB
func (s *FUBBidirectionalSync) LogCallToFUB(leadID int64, duration int, notes string, agentID string) error {
	// Get lead's FUB ID
	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	// Create call log in FUB
	payload := map[string]interface{}{
		"personId":  fubPersonID,
		"type":      "call",
		"direction": "outbound",
		"duration":  duration,
		"notes":     notes,
		"userId":    agentID,
		"createdAt": time.Now().Format(time.RFC3339),
	}

	if err := s.sendToFUB("POST", "/events", payload); err != nil {
		return fmt.Errorf("failed to log call to FUB: %w", err)
	}

	// Track in PropertyHub behavioral events
	s.behavioralService.TrackEvent(leadID, "called", map[string]interface{}{
		"duration": duration,
		"agent_id": agentID,
	}, nil, "", "", "")

	log.Printf("✅ Logged call to FUB for lead %d (duration: %ds)", leadID, duration)
	return nil
}

// LogEmailToFUB logs an email sent in PropertyHub to FUB
func (s *FUBBidirectionalSync) LogEmailToFUB(leadID int64, subject string, body string, agentID string) error {
	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	payload := map[string]interface{}{
		"personId":  fubPersonID,
		"type":      "email",
		"direction": "outbound",
		"subject":   subject,
		"body":      body,
		"userId":    agentID,
		"createdAt": time.Now().Format(time.RFC3339),
	}

	if err := s.sendToFUB("POST", "/events", payload); err != nil {
		return fmt.Errorf("failed to log email to FUB: %w", err)
	}

	// Track in PropertyHub
	s.behavioralService.TrackEvent(leadID, "emailed", map[string]interface{}{
		"subject":  subject,
		"agent_id": agentID,
	}, nil, "", "", "")

	log.Printf("✅ Logged email to FUB for lead %d", leadID)
	return nil
}

// LogSMSToFUB logs an SMS sent in PropertyHub to FUB
func (s *FUBBidirectionalSync) LogSMSToFUB(leadID int64, message string, agentID string) error {
	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	payload := map[string]interface{}{
		"personId":  fubPersonID,
		"type":      "sms",
		"direction": "outbound",
		"body":      message,
		"userId":    agentID,
		"createdAt": time.Now().Format(time.RFC3339),
	}

	if err := s.sendToFUB("POST", "/events", payload); err != nil {
		return fmt.Errorf("failed to log SMS to FUB: %w", err)
	}

	// Track in PropertyHub
	s.behavioralService.TrackEvent(leadID, "sms_sent", map[string]interface{}{
		"agent_id": agentID,
	}, nil, "", "", "")

	log.Printf("✅ Logged SMS to FUB for lead %d", leadID)
	return nil
}

// ScheduleShowingInFUB creates a showing event in FUB
func (s *FUBBidirectionalSync) ScheduleShowingInFUB(leadID int64, propertyID int64, showingTime time.Time, agentID string) error {
	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	// Get property address for event title
	var property models.Property
	if err := s.db.First(&property, propertyID).Error; err != nil {
		return fmt.Errorf("failed to get property: %w", err)
	}

	payload := map[string]interface{}{
		"personId":    fubPersonID,
		"type":        "appointment",
		"title":       fmt.Sprintf("Property Showing: %s", property.Address),
		"description": fmt.Sprintf("Showing at %s", property.Address),
		"startTime":   showingTime.Format(time.RFC3339),
		"userId":      agentID,
	}

	if err := s.sendToFUB("POST", "/events", payload); err != nil {
		return fmt.Errorf("failed to schedule showing in FUB: %w", err)
	}

	// Track in PropertyHub
	s.behavioralService.TrackEvent(leadID, "showing_scheduled", map[string]interface{}{
		"property_id":  propertyID,
		"showing_time": showingTime,
		"agent_id":     agentID,
	}, &propertyID, "", "", "")

	log.Printf("✅ Scheduled showing in FUB for lead %d at property %d", leadID, propertyID)
	return nil
}

// AddNoteToFUB syncs a note from PropertyHub to FUB
func (s *FUBBidirectionalSync) AddNoteToFUB(leadID int64, note string, agentID string) error {
	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	payload := map[string]interface{}{
		"personId": fubPersonID,
		"body":     note,
		"userId":   agentID,
	}

	if err := s.sendToFUB("POST", "/notes", payload); err != nil {
		return fmt.Errorf("failed to add note to FUB: %w", err)
	}

	log.Printf("✅ Added note to FUB for lead %d", leadID)
	return nil
}

// UpdateLeadStatusInFUB updates lead status in FUB
func (s *FUBBidirectionalSync) UpdateLeadStatusInFUB(leadID int64, status string) error {
	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	payload := map[string]interface{}{
		"status": status,
	}

	if err := s.sendToFUB("PUT", fmt.Sprintf("/people/%s", fubPersonID), payload); err != nil {
		return fmt.Errorf("failed to update status in FUB: %w", err)
	}

	log.Printf("✅ Updated lead %d status to '%s' in FUB", leadID, status)
	return nil
}

// AssignAgentInFUB assigns an agent to a lead in FUB
func (s *FUBBidirectionalSync) AssignAgentInFUB(leadID int64, agentID string) error {
	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	payload := map[string]interface{}{
		"ownerId": agentID,
	}

	if err := s.sendToFUB("PUT", fmt.Sprintf("/people/%s", fubPersonID), payload); err != nil {
		return fmt.Errorf("failed to assign agent in FUB: %w", err)
	}

	log.Printf("✅ Assigned agent %s to lead %d in FUB", agentID, leadID)
	return nil
}

// SyncBehavioralScoreToFUB pushes behavioral score and segment to FUB custom fields
func (s *FUBBidirectionalSync) SyncBehavioralScoreToFUB(leadID int64) error {
	// Get current score
	score, err := s.scoringEngine.GetScore(leadID)
	if err != nil {
		return fmt.Errorf("failed to get behavioral score: %w", err)
	}

	fubPersonID, err := s.getFUBPersonID(leadID)
	if err != nil {
		return fmt.Errorf("failed to get FUB person ID: %w", err)
	}

	// Update FUB custom fields
	payload := map[string]interface{}{
		"customFields": map[string]interface{}{
			"behavioralScore":   score.CompositeScore,
			"engagementSegment": score.GetSegment(),
			"lastActivity":      score.LastCalculated.Format("2006-01-02"),
			"totalEvents":       0,
		},
	}

	if err := s.sendToFUB("PUT", fmt.Sprintf("/people/%s", fubPersonID), payload); err != nil {
		return fmt.Errorf("failed to sync score to FUB: %w", err)
	}

	log.Printf("✅ Synced behavioral score (%d, %s) to FUB for lead %d", score.CompositeScore, score.GetSegment(), leadID)
	return nil
}

// ============================================================================
// FUB → PROPERTYHUB (Webhook Handlers)
// ============================================================================

// HandleFUBWebhook processes incoming webhooks from FUB
func (s *FUBBidirectionalSync) HandleFUBWebhook(webhookData map[string]interface{}) error {
	eventType, ok := webhookData["type"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook: missing type")
	}

	switch eventType {
	case "email.opened":
		return s.handleEmailOpened(webhookData)
	case "email.clicked":
		return s.handleEmailClicked(webhookData)
	case "call.logged":
		return s.handleCallLogged(webhookData)
	case "sms.replied":
		return s.handleSMSReplied(webhookData)
	case "note.added":
		return s.handleNoteAdded(webhookData)
	case "person.updated":
		return s.handlePersonUpdated(webhookData)
	default:
		log.Printf("⚠️  Unhandled FUB webhook type: %s", eventType)
		return nil
	}
}

func (s *FUBBidirectionalSync) handleEmailOpened(data map[string]interface{}) error {
	leadID, err := s.getLeadIDFromFUBPerson(data["personId"].(string))
	if err != nil {
		return err
	}

	// Track as behavioral event (+5 points)
	return s.behavioralService.TrackEvent(leadID, "email_opened", map[string]interface{}{
		"source": "fub_webhook",
	}, nil, "", "", "")
}

func (s *FUBBidirectionalSync) handleEmailClicked(data map[string]interface{}) error {
	leadID, err := s.getLeadIDFromFUBPerson(data["personId"].(string))
	if err != nil {
		return err
	}

	// Track as behavioral event (+10 points)
	return s.behavioralService.TrackEvent(leadID, "email_clicked", map[string]interface{}{
		"source": "fub_webhook",
		"url":    data["url"],
	}, nil, "", "", "")
}

func (s *FUBBidirectionalSync) handleCallLogged(data map[string]interface{}) error {
	leadID, err := s.getLeadIDFromFUBPerson(data["personId"].(string))
	if err != nil {
		return err
	}

	// Track as behavioral event (+15 points)
	return s.behavioralService.TrackEvent(leadID, "call_received", map[string]interface{}{
		"source":   "fub_webhook",
		"duration": data["duration"],
	}, nil, "", "", "")
}

func (s *FUBBidirectionalSync) handleSMSReplied(data map[string]interface{}) error {
	leadID, err := s.getLeadIDFromFUBPerson(data["personId"].(string))
	if err != nil {
		return err
	}

	// Track as behavioral event (+10 points)
	return s.behavioralService.TrackEvent(leadID, "sms_replied", map[string]interface{}{
		"source": "fub_webhook",
	}, nil, "", "", "")
}

func (s *FUBBidirectionalSync) handleNoteAdded(data map[string]interface{}) error {
	// Sync note from FUB to PropertyHub (if needed)
	log.Printf("📝 Note added in FUB: %v", data)
	return nil
}

func (s *FUBBidirectionalSync) handlePersonUpdated(data map[string]interface{}) error {
	// Sync person updates from FUB to PropertyHub
	log.Printf("👤 Person updated in FUB: %v", data)
	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getFUBPersonID retrieves the FUB person ID for a PropertyHub lead
func (s *FUBBidirectionalSync) getFUBPersonID(leadID int64) (string, error) {
	var lead models.Lead
	if err := s.db.First(&lead, leadID).Error; err != nil {
		return "", err
	}

	// Assume leads table has fub_id column
	// If not, you'll need to add it via migration
	if lead.FUBLeadID == "" {
		return "", fmt.Errorf("lead %d has no FUB ID", leadID)
	}

	return lead.FUBLeadID, nil
}

// getLeadIDFromFUBPerson retrieves the PropertyHub lead ID from a FUB person ID
func (s *FUBBidirectionalSync) getLeadIDFromFUBPerson(fubPersonID string) (int64, error) {
	var lead models.Lead
	if err := s.db.Where("fub_id = ?", fubPersonID).First(&lead).Error; err != nil {
		return 0, fmt.Errorf("no lead found for FUB person %s: %w", fubPersonID, err)
	}
	return int64(lead.ID), nil
}

// sendToFUB sends an API request to Follow Up Boss
func (s *FUBBidirectionalSync) sendToFUB(method string, endpoint string, payload map[string]interface{}) error {
	url := s.fubBaseURL + endpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+s.fubAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("FUB API error: status %d", resp.StatusCode)
	}

	return nil
}
