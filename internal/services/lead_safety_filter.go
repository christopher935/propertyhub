package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// LeadSafetyFilter checks if automation actions are safe for a given lead
// This prevents edge cases like spamming closing-stage leads, recommending rejected properties, etc.
type LeadSafetyFilter struct {
	db        *gorm.DB
	fubClient *BehavioralFUBAPIClient
}

// NewLeadSafetyFilter creates a new lead safety filter
func NewLeadSafetyFilter(db *gorm.DB, fubClient *BehavioralFUBAPIClient) *LeadSafetyFilter {
	return &LeadSafetyFilter{
		db:        db,
		fubClient: fubClient,
	}
}

// SafetyClassification represents the safety assessment of a lead for automation
type SafetyClassification struct {
	Safe          bool     `json:"safe"`           // Overall safety - true if action is safe
	DoNotContact  bool     `json:"do_not_contact"` // Lead has Do Not Contact flag
	Reasons       []string `json:"reasons"`        // Reasons why action is blocked
	Warnings      []string `json:"warnings"`       // Non-blocking warnings
	Stage         string   `json:"stage"`          // Lead stage (new, active, closing, etc.)
	ClosingRisk   bool     `json:"closing_risk"`   // Lead is in closing stage
	RejectionRisk bool     `json:"rejection_risk"` // Action might recommend rejected property
}

// IsBlocked returns true if the lead is blocked from automation
func (sc *SafetyClassification) IsBlocked() bool {
	return !sc.Safe || sc.DoNotContact
}

// LeadOverride represents agent-set overrides for a lead
type LeadOverride struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	LeadID          string     `gorm:"uniqueIndex" json:"lead_id"`
	DoNotContact    bool       `gorm:"default:false" json:"do_not_contact"`
	PauseUntil      *time.Time `json:"pause_until,omitempty"`
	CustomCooldown  *int       `json:"custom_cooldown,omitempty"` // Custom cooldown in hours
	Reason          string     `json:"reason"`
	SetBy           string     `json:"set_by"`     // Agent who set the override
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// LeadRejection tracks properties that a lead has explicitly rejected
type LeadRejection struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	LeadID     string    `gorm:"index" json:"lead_id"`
	PropertyID string    `gorm:"index" json:"property_id"`
	Reason     string    `json:"reason"` // "too_small", "too_expensive", "wrong_location", "other"
	Notes      string    `json:"notes"`
	RejectedAt time.Time `gorm:"index" json:"rejected_at"`
	RejectedBy string    `json:"rejected_by"` // Agent who recorded the rejection
}

// LeadStageHistory tracks lead stage changes for detecting closing stage
type LeadStageHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	LeadID    string    `gorm:"index" json:"lead_id"`
	Stage     string    `json:"stage"` // "new", "active", "showing", "application", "closing", "signed", "cold"
	Source    string    `json:"source"` // "fub", "propertyhub", "manual"
	ChangedAt time.Time `gorm:"index" json:"changed_at"`
	ChangedBy string    `json:"changed_by"`
}

// ClassifyLead performs comprehensive safety assessment for a lead
func (lsf *LeadSafetyFilter) ClassifyLead(leadData map[string]interface{}) SafetyClassification {
	classification := SafetyClassification{
		Safe:     true,
		Reasons:  []string{},
		Warnings: []string{},
	}

	leadID, ok := leadData["id"].(string)
	if !ok || leadID == "" {
		classification.Safe = false
		classification.Reasons = append(classification.Reasons, "Invalid lead ID")
		return classification
	}

	// Check 1: Do Not Contact flag
	if lsf.checkDoNotContact(leadID) {
		classification.Safe = false
		classification.DoNotContact = true
		classification.Reasons = append(classification.Reasons, "Lead marked Do Not Contact")
		return classification // Absolute block
	}

	// Check 2: Pause Until date
	if pausedUntil := lsf.checkPauseUntil(leadID); pausedUntil != nil {
		if time.Now().Before(*pausedUntil) {
			classification.Safe = false
			classification.Reasons = append(classification.Reasons, 
				fmt.Sprintf("Lead paused until %s", pausedUntil.Format("Jan 2, 3:04 PM")))
			return classification
		}
	}

	// Check 3: Lead stage (detect closing stage)
	stage := lsf.detectLeadStage(leadID)
	classification.Stage = stage
	if stage == "closing" || stage == "signed" {
		classification.ClosingRisk = true
		classification.Warnings = append(classification.Warnings, 
			"Lead is in closing stage - avoid property recommendations")
	}

	// Check 4: Application status (from PropertyHub or FUB)
	appStatus := lsf.checkApplicationStatus(leadID)
	if appStatus == "approved" || appStatus == "lease_sent" {
		classification.ClosingRisk = true
		classification.Warnings = append(classification.Warnings, 
			fmt.Sprintf("Application status: %s - lead is closing", appStatus))
	}

	// Check 5: Recent agent contact (don't interfere with active conversations)
	if recentContact := lsf.checkRecentAgentContact(leadID); recentContact {
		classification.Warnings = append(classification.Warnings, 
			"Agent contacted lead recently - consider delaying automation")
	}

	return classification
}

// CheckPropertyRecommendationSafety checks if it's safe to recommend a property to a lead
func (lsf *LeadSafetyFilter) CheckPropertyRecommendationSafety(leadID string, propertyID string) (bool, string) {
	// Check 1: Is lead in closing stage?
	stage := lsf.detectLeadStage(leadID)
	if stage == "closing" || stage == "signed" {
		return false, "Lead is in closing stage - do not send property recommendations"
	}

	// Check 2: Has lead rejected this property?
	var rejection LeadRejection
	err := lsf.db.Where("lead_id = ? AND property_id = ?", leadID, propertyID).
		First(&rejection).Error
	
	if err == nil {
		return false, fmt.Sprintf("Lead rejected this property: %s", rejection.Reason)
	}

	// Check 3: Is property still available?
	// TODO: Check property availability from database
	// For now, assume available

	return true, "Safe to recommend"
}

// CheckCommunicationSafety checks if it's safe to send communication to a lead
func (lsf *LeadSafetyFilter) CheckCommunicationSafety(leadID string, messageType string, timeOfDay time.Time) (bool, string) {
	// Check 1: Time of day restrictions
	hour := timeOfDay.Hour()
	if hour >= 21 || hour < 8 {
		return false, fmt.Sprintf("Outside communication hours (9pm-8am) - current time: %s", 
			timeOfDay.Format("3:04 PM"))
	}

	// Check 2: Do Not Contact
	if lsf.checkDoNotContact(leadID) {
		return false, "Lead marked Do Not Contact"
	}

	// Check 3: Pause Until
	if pausedUntil := lsf.checkPauseUntil(leadID); pausedUntil != nil {
		if time.Now().Before(*pausedUntil) {
			return false, fmt.Sprintf("Lead paused until %s", pausedUntil.Format("Jan 2, 3:04 PM"))
		}
	}

	return true, "Safe to communicate"
}

// checkDoNotContact checks if lead has Do Not Contact flag
func (lsf *LeadSafetyFilter) checkDoNotContact(leadID string) bool {
	var override LeadOverride
	err := lsf.db.Where("lead_id = ? AND do_not_contact = ?", leadID, true).
		First(&override).Error
	
	return err == nil // Found = Do Not Contact is set
}

// checkPauseUntil checks if lead is paused until a specific date
func (lsf *LeadSafetyFilter) checkPauseUntil(leadID string) *time.Time {
	var override LeadOverride
	err := lsf.db.Where("lead_id = ?", leadID).
		First(&override).Error
	
	if err == nil && override.PauseUntil != nil {
		return override.PauseUntil
	}
	
	return nil
}

// detectLeadStage determines the current stage of a lead
func (lsf *LeadSafetyFilter) detectLeadStage(leadID string) string {
	// Priority 1: Check PropertyHub stage history (most recent)
	var stageHistory LeadStageHistory
	err := lsf.db.Where("lead_id = ?", leadID).
		Order("changed_at DESC").
		First(&stageHistory).Error
	
	if err == nil {
		return stageHistory.Stage
	}

	// Priority 2: Check FUB stage (if FUB client is available)
	if lsf.fubClient != nil {
		fubStage := lsf.getFUBStage(leadID)
		if fubStage != "" {
			return lsf.normalizeFUBStage(fubStage)
		}
	}

	// Priority 3: Infer from application status
	appStatus := lsf.checkApplicationStatus(leadID)
	if appStatus == "approved" || appStatus == "lease_sent" {
		return "closing"
	}
	if appStatus == "submitted" || appStatus == "in_review" {
		return "application"
	}

	// Default: assume active
	return "active"
}

// getFUBStage queries FUB API for lead stage
func (lsf *LeadSafetyFilter) getFUBStage(leadID string) string {
	if lsf.db == nil {
		return ""
	}

	var fubLead models.FUBLead
	if err := lsf.db.Where("fub_lead_id = ?", leadID).First(&fubLead).Error; err != nil {
		log.Printf("Warning: Could not find FUB lead %s: %v", leadID, err)
		return ""
	}

	return fubLead.Stage
}

// normalizeFUBStage converts FUB stage names to PropertyHub stage names
func (lsf *LeadSafetyFilter) normalizeFUBStage(fubStage string) string {
	// Map FUB stages to PropertyHub stages
	stageMap := map[string]string{
		"New Lead":        "new",
		"Active":          "active",
		"Showing":         "showing",
		"Application":     "application",
		"Closing":         "closing",
		"Signed":          "signed",
		"Cold":            "cold",
		"Lost":            "cold",
		"Dead":            "cold",
	}

	normalized, exists := stageMap[fubStage]
	if exists {
		return normalized
	}

	// Default: lowercase and remove spaces
	return strings.ToLower(strings.ReplaceAll(fubStage, " ", "_"))
}

// checkApplicationStatus checks the application status for a lead
func (lsf *LeadSafetyFilter) checkApplicationStatus(leadID string) string {
	// TODO: Query applications table
	// For now, return empty (will be implemented when application system is built)
	
	// Placeholder query structure:
	// var application Application
	// err := lsf.db.Where("lead_id = ?", leadID).
	// 	Order("created_at DESC").
	// 	First(&application).Error
	// if err == nil {
	// 	return application.Status
	// }
	
	return ""
}

// checkRecentAgentContact checks if agent has contacted lead recently
func (lsf *LeadSafetyFilter) checkRecentAgentContact(leadID string) bool {
	// Check decision log for recent agent-initiated actions
	var recentDecision struct {
		CreatedAt time.Time
	}
	
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	
	err := lsf.db.Table("decision_log_entries").
		Select("created_at").
		Where("lead_id = ? AND initiated_by = ? AND created_at > ?", 
			leadID, "admin_manual", twoHoursAgo).
		Order("created_at DESC").
		First(&recentDecision).Error
	
	return err == nil // Found recent agent contact
}

// RecordRejection records that a lead rejected a property
func (lsf *LeadSafetyFilter) RecordRejection(leadID string, propertyID string, reason string, notes string, rejectedBy string) error {
	rejection := LeadRejection{
		LeadID:     leadID,
		PropertyID: propertyID,
		Reason:     reason,
		Notes:      notes,
		RejectedAt: time.Now(),
		RejectedBy: rejectedBy,
	}

	err := lsf.db.Create(&rejection).Error
	if err != nil {
		log.Printf("❌ Error recording rejection: %v", err)
		return err
	}

	log.Printf("✅ Recorded rejection: Lead %s rejected property %s (%s)", leadID, propertyID, reason)
	return nil
}

// RecordStageChange records a change in lead stage
func (lsf *LeadSafetyFilter) RecordStageChange(leadID string, newStage string, source string, changedBy string) error {
	stageHistory := LeadStageHistory{
		LeadID:    leadID,
		Stage:     newStage,
		Source:    source,
		ChangedAt: time.Now(),
		ChangedBy: changedBy,
	}

	err := lsf.db.Create(&stageHistory).Error
	if err != nil {
		log.Printf("❌ Error recording stage change: %v", err)
		return err
	}

	log.Printf("✅ Recorded stage change: Lead %s → %s (source: %s)", leadID, newStage, source)
	return nil
}

// GetLeadRejections returns all properties rejected by a lead
func (lsf *LeadSafetyFilter) GetLeadRejections(leadID string) ([]LeadRejection, error) {
	var rejections []LeadRejection
	err := lsf.db.Where("lead_id = ?", leadID).
		Order("rejected_at DESC").
		Find(&rejections).Error
	
	return rejections, err
}

// GetLeadStageHistory returns the stage history for a lead
func (lsf *LeadSafetyFilter) GetLeadStageHistory(leadID string) ([]LeadStageHistory, error) {
	var history []LeadStageHistory
	err := lsf.db.Where("lead_id = ?", leadID).
		Order("changed_at DESC").
		Find(&history).Error
	
	return history, err
}

// AutoMigrate creates the necessary database tables
func (lsf *LeadSafetyFilter) AutoMigrate() error {
	return lsf.db.AutoMigrate(
		&LeadOverride{},
		&LeadRejection{},
		&LeadStageHistory{},
	)
}
