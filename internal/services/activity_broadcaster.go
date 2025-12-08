package services

import (
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

type ActivityBroadcaster interface {
	BroadcastEvent(event ActivityEventData)
	BroadcastActiveCount(count int)
}

type ActivityEventData struct {
	Type        string                 `json:"type"`
	SessionID   string                 `json:"session_id"`
	UserID      int64                  `json:"user_id,omitempty"`
	UserEmail   string                 `json:"user_email,omitempty"`
	PropertyID  *int64                 `json:"property_id,omitempty"`
	Details     string                 `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	Score       int                    `json:"score,omitempty"`
	EventData   map[string]interface{} `json:"event_data,omitempty"`
}

type ActivityBroadcastService struct {
	db          *gorm.DB
	broadcaster ActivityBroadcaster
}

func NewActivityBroadcastService(db *gorm.DB) *ActivityBroadcastService {
	return &ActivityBroadcastService{
		db: db,
	}
}

func (s *ActivityBroadcastService) SetBroadcaster(broadcaster ActivityBroadcaster) {
	s.broadcaster = broadcaster
}

func (s *ActivityBroadcastService) BroadcastPropertyView(leadID int64, propertyID int64, sessionID string, eventData map[string]interface{}) {
	if s.broadcaster == nil {
		return
	}

	var lead models.Lead
	var property models.Property
	
	s.db.Where("id = ?", leadID).First(&lead)
	s.db.Where("id = ?", propertyID).First(&property)

	var score models.BehavioralScore
	s.db.Where("lead_id = ?", leadID).Order("last_calculated DESC").First(&score)

	details := fmt.Sprintf("Viewing %s", property.Address)
	if lead.Email != "" {
		details = fmt.Sprintf("%s viewed %s", lead.Email, property.Address)
	}

	event := ActivityEventData{
		Type:       "property_view",
		SessionID:  sessionID,
		UserID:     leadID,
		UserEmail:  lead.Email,
		PropertyID: &propertyID,
		Details:    details,
		Timestamp:  time.Now(),
		Score:      score.CompositeScore,
		EventData:  eventData,
	}

	log.Printf("📡 Broadcasting property view: %s", details)
	s.broadcaster.BroadcastEvent(event)
}

func (s *ActivityBroadcastService) BroadcastPropertySave(leadID int64, propertyID int64, sessionID string, eventData map[string]interface{}) {
	if s.broadcaster == nil {
		return
	}

	var lead models.Lead
	var property models.Property
	
	s.db.Where("id = ?", leadID).First(&lead)
	s.db.Where("id = ?", propertyID).First(&property)

	var score models.BehavioralScore
	s.db.Where("lead_id = ?", leadID).Order("last_calculated DESC").First(&score)

	details := fmt.Sprintf("Saved %s", property.Address)
	if lead.Email != "" {
		details = fmt.Sprintf("%s saved %s", lead.Email, property.Address)
	}

	event := ActivityEventData{
		Type:       "property_save",
		SessionID:  sessionID,
		UserID:     leadID,
		UserEmail:  lead.Email,
		PropertyID: &propertyID,
		Details:    details,
		Timestamp:  time.Now(),
		Score:      score.CompositeScore,
		EventData:  eventData,
	}

	log.Printf("📡 Broadcasting property save: %s", details)
	s.broadcaster.BroadcastEvent(event)
}

func (s *ActivityBroadcastService) BroadcastInquiry(leadID int64, propertyID *int64, inquiryType string, sessionID string, eventData map[string]interface{}) {
	if s.broadcaster == nil {
		return
	}

	var lead models.Lead
	s.db.Where("id = ?", leadID).First(&lead)

	var score models.BehavioralScore
	s.db.Where("lead_id = ?", leadID).Order("last_calculated DESC").First(&score)

	details := fmt.Sprintf("Sent %s inquiry", inquiryType)
	if lead.Email != "" {
		details = fmt.Sprintf("%s sent %s inquiry", lead.Email, inquiryType)
	}

	if propertyID != nil {
		var property models.Property
		s.db.Where("id = ?", *propertyID).First(&property)
		details += fmt.Sprintf(" for %s", property.Address)
	}

	event := ActivityEventData{
		Type:       "inquiry",
		SessionID:  sessionID,
		UserID:     leadID,
		UserEmail:  lead.Email,
		PropertyID: propertyID,
		Details:    details,
		Timestamp:  time.Now(),
		Score:      score.CompositeScore,
		EventData:  eventData,
	}

	log.Printf("📡 Broadcasting inquiry: %s", details)
	s.broadcaster.BroadcastEvent(event)
}

func (s *ActivityBroadcastService) BroadcastSearch(leadID int64, sessionID string, searchCriteria map[string]interface{}) {
	if s.broadcaster == nil {
		return
	}

	var lead models.Lead
	s.db.Where("id = ?", leadID).First(&lead)

	var score models.BehavioralScore
	s.db.Where("lead_id = ?", leadID).Order("last_calculated DESC").First(&score)

	details := "Searched properties"
	if lead.Email != "" {
		details = fmt.Sprintf("%s searched properties", lead.Email)
	}

	event := ActivityEventData{
		Type:      "search",
		SessionID: sessionID,
		UserID:    leadID,
		UserEmail: lead.Email,
		Details:   details,
		Timestamp: time.Now(),
		Score:     score.CompositeScore,
		EventData: searchCriteria,
	}

	log.Printf("📡 Broadcasting search: %s", details)
	s.broadcaster.BroadcastEvent(event)
}

func (s *ActivityBroadcastService) BroadcastApplication(leadID int64, propertyID int64, sessionID string, eventData map[string]interface{}) {
	if s.broadcaster == nil {
		return
	}

	var lead models.Lead
	var property models.Property
	
	s.db.Where("id = ?", leadID).First(&lead)
	s.db.Where("id = ?", propertyID).First(&property)

	var score models.BehavioralScore
	s.db.Where("lead_id = ?", leadID).Order("last_calculated DESC").First(&score)

	details := fmt.Sprintf("Submitted application for %s", property.Address)
	if lead.Email != "" {
		details = fmt.Sprintf("%s submitted application for %s", lead.Email, property.Address)
	}

	event := ActivityEventData{
		Type:       "application",
		SessionID:  sessionID,
		UserID:     leadID,
		UserEmail:  lead.Email,
		PropertyID: &propertyID,
		Details:    details,
		Timestamp:  time.Now(),
		Score:      score.CompositeScore,
		EventData:  eventData,
	}

	log.Printf("📡 Broadcasting application: %s", details)
	s.broadcaster.BroadcastEvent(event)
}
