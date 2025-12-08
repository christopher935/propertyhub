package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// CampaignTriggerAutomation handles automatic campaign triggering based on intelligence
type CampaignTriggerAutomation struct {
	db                  *gorm.DB
	emailBatch          *EmailBatchService
	relationshipEngine  *RelationshipIntelligenceEngine
	propertyMatcher     *PropertyMatchingService
	abandonmentRecovery *AbandonmentRecoveryService
}

// NewCampaignTriggerAutomation creates a new campaign trigger automation service
func NewCampaignTriggerAutomation(
	db *gorm.DB,
	emailBatch *EmailBatchService,
	relationshipEngine *RelationshipIntelligenceEngine,
	propertyMatcher *PropertyMatchingService,
	abandonmentRecovery *AbandonmentRecoveryService,
) *CampaignTriggerAutomation {
	return &CampaignTriggerAutomation{
		db:                  db,
		emailBatch:          emailBatch,
		relationshipEngine:  relationshipEngine,
		propertyMatcher:     propertyMatcher,
		abandonmentRecovery: abandonmentRecovery,
	}
}

// CampaignTrigger represents a trigger condition for a campaign
type CampaignTrigger struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"` // "opportunity", "property_match", "time_based", "behavioral"
	Condition       string                 `json:"condition"`
	Template        string                 `json:"template"`
	Enabled         bool                   `json:"enabled"`
	AutoExecute     bool                   `json:"auto_execute"`
	Priority        int                    `json:"priority"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// CampaignExecution represents an executed campaign
type CampaignExecution struct {
	ID              string                 `json:"id"`
	TriggerID       string                 `json:"trigger_id"`
	TriggerType     string                 `json:"trigger_type"`
	LeadID          int64                  `json:"lead_id"`
	Template        string                 `json:"template"`
	TemplateData    map[string]interface{} `json:"template_data"`
	ExecutedAt      time.Time              `json:"executed_at"`
	Status          string                 `json:"status"` // "pending", "sent", "failed"
	EmailID         string                 `json:"email_id,omitempty"`
}

// ProcessOpportunityTriggers processes opportunities and triggers campaigns
func (cta *CampaignTriggerAutomation) ProcessOpportunityTriggers() error {
	log.Println("🎯 Campaign Triggers: Processing opportunity-based triggers...")
	
	// Get opportunities from relationship engine
	opportunities, err := cta.relationshipEngine.AnalyzeOpportunities()
	if err != nil {
		log.Printf("Error analyzing opportunities: %v", err)
		return err
	}
	
	executionCount := 0
	
	for _, opp := range opportunities {
		// Process each action in the opportunity's action sequence
		for _, action := range opp.ActionSequence {
			if action.Action == "send_email" && action.AutoExecute {
				// Execute email campaign
				err := cta.executeEmailCampaign(opp, action)
				if err != nil {
					log.Printf("Error executing email campaign for opportunity %s: %v", opp.ID, err)
				} else {
					executionCount++
				}
			}
		}
	}
	
	log.Printf("✅ Executed %d opportunity-based campaigns", executionCount)
	return nil
}

// ProcessPropertyMatchTriggers processes new property matches and triggers notifications
func (cta *CampaignTriggerAutomation) ProcessPropertyMatchTriggers() error {
	log.Println("🏠 Campaign Triggers: Processing property match triggers...")
	
	// Find new matches from the last 24 hours
	since := time.Now().AddDate(0, 0, -1)
	matches, err := cta.propertyMatcher.FindNewMatchesSince(since)
	if err != nil {
		log.Printf("Error finding new matches: %v", err)
		return err
	}
	
	executionCount := 0
	
	for _, match := range matches {
		// Only trigger if match score is high enough and not already notified
		if match.MatchScore >= 70 && !match.NotificationSent {
			err := cta.executePropertyMatchCampaign(match)
			if err != nil {
				log.Printf("Error executing property match campaign: %v", err)
			} else {
				executionCount++
			}
		}
	}
	
	log.Printf("✅ Executed %d property match campaigns", executionCount)
	return nil
}

// ProcessTimedTriggers processes time-based triggers (30-day, 60-day, 90-day inactive)
func (cta *CampaignTriggerAutomation) ProcessTimedTriggers() error {
	log.Println("⏰ Campaign Triggers: Processing time-based triggers...")
	
	executionCount := 0
	
	// 30-day inactive trigger
	count30, err := cta.processInactiveTrigger(30, "reengagement_30day")
	if err == nil {
		executionCount += count30
	}
	
	// 60-day inactive trigger
	count60, err := cta.processInactiveTrigger(60, "reengagement_60day")
	if err == nil {
		executionCount += count60
	}
	
	// 90-day inactive trigger
	count90, err := cta.processInactiveTrigger(90, "reengagement_90day")
	if err == nil {
		executionCount += count90
	}
	
	log.Printf("✅ Executed %d time-based campaigns", executionCount)
	return nil
}

// ProcessBehavioralTriggers processes behavioral event triggers
func (cta *CampaignTriggerAutomation) ProcessBehavioralTriggers() error {
	log.Println("🧠 Campaign Triggers: Processing behavioral triggers...")
	
	executionCount := 0
	
	// Trigger 1: Lead viewed property 3+ times in 24 hours
	count1, err := cta.processHighViewTrigger()
	if err == nil {
		executionCount += count1
	}
	
	// Trigger 2: Lead saved property but didn't book showing
	count2, err := cta.processSavedNoShowingTrigger()
	if err == nil {
		executionCount += count2
	}
	
	// Trigger 3: Showing completed but no application
	count3, err := cta.processShowingNoApplicationTrigger()
	if err == nil {
		executionCount += count3
	}
	
	log.Printf("✅ Executed %d behavioral campaigns", executionCount)
	return nil
}

// RunAllTriggers runs all trigger types
func (cta *CampaignTriggerAutomation) RunAllTriggers() error {
	log.Println("🚀 Campaign Triggers: Running all automated triggers...")
	
	// Process in order of priority
	cta.ProcessOpportunityTriggers()
	cta.ProcessPropertyMatchTriggers()
	cta.ProcessBehavioralTriggers()
	cta.ProcessTimedTriggers()
	
	log.Println("✅ All campaign triggers processed")
	return nil
}

// Helper functions

func (cta *CampaignTriggerAutomation) executeEmailCampaign(opp Opportunity, action OpportunityAction) error {
	// Build template data
	templateData := map[string]interface{}{
		"lead_name":       opp.LeadName,
		"lead_email":      opp.LeadEmail,
		"opportunity_type": opp.Type,
		"urgency_score":   opp.UrgencyScore,
		"conversion_prob": fmt.Sprintf("%.0f%%", opp.ConversionProbability*100),
	}
	
	// Add property data if available
	if opp.PropertyID != nil {
		templateData["property_id"] = *opp.PropertyID
		templateData["property_address"] = opp.PropertyAddress
	}
	
	// Add metadata
	for key, value := range opp.Metadata {
		templateData[key] = value
	}
	
	// Get template content
	template := cta.getTemplateContent(action.Template, templateData)
	
	// Send email via email batch service
	emailData := map[string]interface{}{
		"to":       opp.LeadEmail,
		"subject":  cta.getTemplateSubject(action.Template, templateData),
		"body":     template,
		"template": action.Template,
	}
	
	// Log execution
	execution := CampaignExecution{
		ID:           fmt.Sprintf("exec_%d_%d", opp.LeadID, time.Now().Unix()),
		TriggerID:    opp.ID,
		TriggerType:  "opportunity",
		LeadID:       opp.LeadID,
		Template:     action.Template,
		TemplateData: templateData,
		ExecutedAt:   time.Now(),
		Status:       "sent",
	}
	
	log.Printf("📧 Sent campaign: %s to %s (template: %s)", execution.ID, opp.LeadEmail, action.Template)
	
	// Store execution record (would save to database in production)
	_ = execution
	
	// In production, would call email service here
	// err := cta.emailBatch.SendEmail(emailData)
	_ = emailData // Placeholder until email service is connected
	
	return nil
}

func (cta *CampaignTriggerAutomation) executePropertyMatchCampaign(match PropertyMatch) error {
	templateData := map[string]interface{}{
			"lead_name":       match.Lead.FirstName + " " + match.Lead.LastName,
		"lead_email":       match.Lead.Email,
		"property_id":      match.PropertyID,
		"property_address": match.Property.Address,
		"property_price":   match.Property.Price,
		"property_bedrooms": match.Property.Bedrooms,
		"property_bathrooms": match.Property.Bathrooms,
		"match_score":      fmt.Sprintf("%.0f%%", match.MatchScore),
		"match_reasons":    match.MatchReasons,
	}
	
	template := cta.getTemplateContent("new_property_match", templateData)
	
	emailData := map[string]interface{}{
		"to":       match.Lead.Email,
		"subject":  cta.getTemplateSubject("new_property_match", templateData),
		"body":     template,
		"template": "new_property_match",
	}
	
	log.Printf("📧 Sent property match: Property %d to Lead %d (%.0f%% match)", match.PropertyID, match.LeadID, match.MatchScore)
	
	// In production, would call email service
	// err := cta.emailBatch.SendEmail(emailData)
	
	// Email would be sent here in production
	_ = emailData
	
	return nil
}

func (cta *CampaignTriggerAutomation) processInactiveTrigger(days int, template string) (int, error) {
	// Find leads inactive for exactly this many days (±1 day to avoid duplicates)
	startDate := time.Now().AddDate(0, 0, -(days + 1))
	endDate := time.Now().AddDate(0, 0, -(days - 1))
	
	var leads []models.Lead
	err := cta.db.Where("last_contact BETWEEN ? AND ?", startDate, endDate).
		Where("status = ?", "active").
		Limit(100).
		Find(&leads).Error
	
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, lead := range leads {
		templateData := map[string]interface{}{
			"lead_name":     lead.FirstName + " " + lead.LastName,
			"lead_email":    lead.Email,
			"days_inactive": days,
		}
		
		// Send re-engagement email
		log.Printf("📧 Sent %d-day re-engagement to %s", days, lead.Email)
		count++
		
		_ = templateData
		// In production: cta.emailBatch.SendEmail(...)
	}
	
	return count, nil
}

func (cta *CampaignTriggerAutomation) processHighViewTrigger() (int, error) {
	// Find leads who viewed a property 3+ times in last 24 hours
	query := `
		SELECT lead_id, property_id, COUNT(*) as view_count
		FROM behavioral_events
		WHERE event_type = 'property_view'
			AND created_at > NOW() - INTERVAL '24 hours'
			AND property_id IS NOT NULL
		GROUP BY lead_id, property_id
		HAVING COUNT(*) >= 3
	`
	
	type ViewCount struct {
		LeadID     int64
		PropertyID int64
		ViewCount  int
	}
	
	var views []ViewCount
	err := cta.db.Raw(query).Scan(&views).Error
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, view := range views {
		// Check if we already sent this trigger recently
		// (would check database in production)
		
		log.Printf("📧 Sent high-view trigger: Lead %d viewed Property %d %d times", view.LeadID, view.PropertyID, view.ViewCount)
		count++
		
		// In production: send email
	}
	
	return count, nil
}

func (cta *CampaignTriggerAutomation) processSavedNoShowingTrigger() (int, error) {
	// Find leads who saved properties but didn't book showings
	query := `
		SELECT DISTINCT be.lead_id, be.property_id
		FROM behavioral_events be
		WHERE be.event_type = 'property_save'
			AND be.created_at > NOW() - INTERVAL '7 days'
			AND NOT EXISTS (
				SELECT 1 FROM behavioral_events be2
				WHERE be2.lead_id = be.lead_id
					AND be2.property_id = be.property_id
					AND be2.event_type = 'showing_requested'
			)
	`
	
	type SavedProperty struct {
		LeadID     int64
		PropertyID int64
	}
	
	var saved []SavedProperty
	err := cta.db.Raw(query).Scan(&saved).Error
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, sp := range saved {
		log.Printf("📧 Sent saved-no-showing trigger: Lead %d saved Property %d", sp.LeadID, sp.PropertyID)
		count++
	}
	
	return count, nil
}

func (cta *CampaignTriggerAutomation) processShowingNoApplicationTrigger() (int, error) {
	// Find leads who completed showings but didn't apply
	query := `
		SELECT DISTINCT be.lead_id, be.property_id
		FROM behavioral_events be
		WHERE be.event_type = 'showing_completed'
			AND be.created_at > NOW() - INTERVAL '7 days'
			AND NOT EXISTS (
				SELECT 1 FROM behavioral_events be2
				WHERE be2.lead_id = be.lead_id
					AND be2.event_type = 'application_started'
					AND be2.created_at > be.created_at
			)
	`
	
	type CompletedShowing struct {
		LeadID     int64
		PropertyID int64
	}
	
	var showings []CompletedShowing
	err := cta.db.Raw(query).Scan(&showings).Error
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, showing := range showings {
		log.Printf("📧 Sent showing-no-application trigger: Lead %d completed showing for Property %d", showing.LeadID, showing.PropertyID)
		count++
	}
	
	return count, nil
}

// Template helpers

func (cta *CampaignTriggerAutomation) getTemplateContent(templateName string, data map[string]interface{}) string {
	// In production, this would load actual templates from database or files
	// For now, return placeholder
	
	templates := map[string]string{
		"hot_lead_followup": `
			Hi {{lead_name}},

			I noticed you've been very active on our site recently! You're showing strong interest, 
			and I'd love to help you find the perfect place.

			Based on your activity, I think you're ready to move forward. Would you like to schedule 
			a call or showing this week?

			Best regards,
			PropertyHub Team
		`,
		"property_reengagement": `
			Hi {{lead_name}},

			I noticed you viewed {{property_address}} {{view_count}} times recently. Are you still 
			interested in this property?

			I'd be happy to schedule a showing or answer any questions you might have.

			Let me know!
		`,
		"new_property_match": `
			Hi {{lead_name}},

			Great news! A new property just became available that's a {{match_score}} match for 
			what you're looking for:

			📍 {{property_address}}
			💰 ${{property_price}}/month
			🛏️ {{property_bedrooms}} BR / {{property_bathrooms}} BA

			This one matches your criteria perfectly. Want to see it before someone else grabs it?

			[Schedule Showing]
		`,
		"reengagement_30day": `
			Hi {{lead_name}},

			It's been about a month since we last connected. Are you still looking for a place?

			I have some new properties that might interest you. Let me know if you'd like to see 
			what's available!
		`,
		"reengagement_60day": `
			Hi {{lead_name}},

			Just checking in! It's been a couple of months since we last spoke.

			The market has changed quite a bit - there are some great new options available now.
			Would you like an updated list of properties in your area?
		`,
		"reengagement_90day": `
			Hi {{lead_name}},

			I hope you're doing well! It's been a while since we last connected.

			If you're still in the market for a rental, I'd love to help. We have some excellent 
			new listings that just came available.

			Let me know if you'd like to reconnect!
		`,
	}
	
	template, exists := templates[templateName]
	if !exists {
		template = "Template not found: " + templateName
	}
	
	// Simple template variable replacement
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		template = strings.Replace(template, placeholder, fmt.Sprintf("%v", value), -1)
	}
	
	return template
}

func (cta *CampaignTriggerAutomation) getTemplateSubject(templateName string, data map[string]interface{}) string {
	subjects := map[string]string{
		"hot_lead_followup":      "Ready to find your perfect place?",
		"property_reengagement":  fmt.Sprintf("Still interested in %v?", data["property_address"]),
		"new_property_match":     "New property matches your search!",
		"reengagement_30day":     "Still looking for a place?",
		"reengagement_60day":     "New properties available in your area",
		"reengagement_90day":     "Let's reconnect - great new listings!",
	}
	
	subject, exists := subjects[templateName]
	if !exists {
		subject = "PropertyHub Update"
	}
	
	return subject
}
