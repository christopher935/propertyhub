package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// EventCampaignOrchestrator connects behavioral events to intelligent campaign execution
// This is the "symphony conductor" that orchestrates automation at scale for 13,000+ leads
type EventCampaignOrchestrator struct {
	db                 *gorm.DB
	automationService  *SMSEmailAutomationService
	eventMappings      map[string]CampaignMapping
}

// CampaignMapping defines how an event type maps to campaigns
type CampaignMapping struct {
	EventType       string
	CampaignName    string
	TargetingRules  func(map[string]interface{}) ([]string, error) // Returns FUB contact IDs
	TemplateBuilder func(map[string]interface{}) (string, string, error) // Returns subject, body
	Priority        int
	CooldownHours   int // Minimum hours between campaigns of this type per lead
}

// CampaignExecutionLog tracks orchestration actions for analytics
type CampaignExecutionLog struct {
	EventType       string    `json:"event_type"`
	CampaignName    string    `json:"campaign_name"`
	TargetsFound    int       `json:"targets_found"`
	MessagesSent    int       `json:"messages_sent"`
	ExecutedAt      time.Time `json:"executed_at"`
	EventData       string    `json:"event_data"`
}

// NewEventCampaignOrchestrator creates the orchestration service
func NewEventCampaignOrchestrator(db *gorm.DB, automationService *SMSEmailAutomationService) *EventCampaignOrchestrator {
	orchestrator := &EventCampaignOrchestrator{
		db:                db,
		automationService: automationService,
		eventMappings:     make(map[string]CampaignMapping),
	}

	// Initialize event→campaign mappings
	orchestrator.initializeEventMappings()

	log.Println("🎼 Event Campaign Orchestrator initialized - automation symphony ready")

	return orchestrator
}

// initializeEventMappings defines the intelligent automation flows
func (eco *EventCampaignOrchestrator) initializeEventMappings() {
	eco.eventMappings = map[string]CampaignMapping{
		// Property price reduced → Notify warm leads who viewed it
		"price_changed": {
			EventType:      "price_changed",
			CampaignName:   "Price Reduction Alert",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findLeadsWhoViewedProperty(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildPriceChangeTemplate(data) },
			Priority:       1,
			CooldownHours:  24,
		},
		
		// New property listed → Notify leads with matching preferences
		"new_listing": {
			EventType:      "new_listing",
			CampaignName:   "New Property Match",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findLeadsMatchingProperty(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildNewListingTemplate(data) },
			Priority:       2,
			CooldownHours:  12,
		},
		
		// Lead scored HOT → Agent alert + personalized outreach
		"lead_scored_hot": {
			EventType:      "lead_scored_hot",
			CampaignName:   "Hot Lead Engagement",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findHotLead(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildHotLeadTemplate(data) },
			Priority:       1,
			CooldownHours:  48,
		},
		
		// Showing completed → Follow-up with attendee
		"showing_completed": {
			EventType:      "showing_completed",
			CampaignName:   "Showing Follow-Up",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findShowingAttendee(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildShowingFollowUpTemplate(data) },
			Priority:       1,
			CooldownHours:  2,
		},
		
		// Application submitted → Confirmation + next steps
		"application_submitted": {
			EventType:      "application_submitted",
			CampaignName:   "Application Confirmation",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findApplicant(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildApplicationConfirmationTemplate(data) },
			Priority:       1,
			CooldownHours:  0,
		},
		
		// Property back on market → Notify previous interested leads
		"property_relisted": {
			EventType:      "property_relisted",
			CampaignName:   "Property Available Again",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findPreviouslyInterestedLeads(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildRelistedTemplate(data) },
			Priority:       2,
			CooldownHours:  24,
		},
		
		// Lead inactive for 30 days → Re-engagement campaign
		"lead_dormant": {
			EventType:      "lead_dormant",
			CampaignName:   "Re-Engagement Outreach",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findDormantLead(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildReengagementTemplate(data) },
			Priority:       3,
			CooldownHours:  168, // 7 days
		},
		
		// Lease ending soon (from Friday Report data) → Renewal campaign
		"lease_ending_soon": {
			EventType:      "lease_ending_soon",
			CampaignName:   "Lease Renewal Opportunity",
			TargetingRules: func(data map[string]interface{}) ([]string, error) { return eco.findTenantByProperty(data) },
			TemplateBuilder: func(data map[string]interface{}) (string, string, error) { return eco.buildLeaseRenewalTemplate(data) },
			Priority:       2,
			CooldownHours:  72,
		},
	}
}

// ProcessEvent is the main orchestration entry point
func (eco *EventCampaignOrchestrator) ProcessEvent(eventType string, eventData map[string]interface{}) error {
	log.Printf("🎵 Orchestrating campaign for event: %s", eventType)

	// Get campaign mapping for this event type
	mapping, exists := eco.eventMappings[eventType]
	if !exists {
		log.Printf("⚠️  No campaign mapping for event type: %s", eventType)
		return nil // Not an error, just no automation defined yet
	}

	// Check if this is a duplicate recent event (prevent spam)
	if eco.isRecentDuplicate(eventType, eventData, mapping.CooldownHours) {
		log.Printf("⏭️  Skipping duplicate event within cooldown period: %s", eventType)
		return nil
	}

	// Find target audience using behavioral targeting
	targetContactIDs, err := mapping.TargetingRules(eventData)
	if err != nil {
		return fmt.Errorf("targeting rules failed: %v", err)
	}

	if len(targetContactIDs) == 0 {
		log.Printf("📭 No targets found for campaign: %s", mapping.CampaignName)
		return nil
	}

	log.Printf("🎯 Found %d targets for campaign: %s", len(targetContactIDs), mapping.CampaignName)

	// Build campaign message using template builder
	subject, body, err := mapping.TemplateBuilder(eventData)
	if err != nil {
		return fmt.Errorf("template builder failed: %v", err)
	}

	// Execute campaign for each target
	successCount := 0
	for _, contactID := range targetContactIDs {
		// Create campaign data for automation service
		campaignData := map[string]interface{}{
			"contact_id":    contactID,
			"event_type":    eventType,
			"campaign_name": mapping.CampaignName,
			"subject":       subject,
			"body":          body,
		}

		// Add event-specific data for template personalization
		for key, value := range eventData {
			campaignData[key] = value
		}

		// Trigger automation via existing automation service
		if err := eco.triggerCampaign(contactID, subject, body, campaignData); err != nil {
			log.Printf("❌ Failed to send campaign to %s: %v", contactID, err)
			continue
		}

		// Log execution to campaign_executions table
		eco.logCampaignExecution(contactID, eventType, mapping.CampaignName, eventData)

		successCount++
	}

	log.Printf("✅ Campaign executed: %s (%d/%d sent)", mapping.CampaignName, successCount, len(targetContactIDs))

	// Record execution summary
	eco.recordExecutionSummary(eventType, mapping.CampaignName, len(targetContactIDs), successCount, eventData)

	return nil
}

// ============================================================================
// BEHAVIORAL TARGETING QUERIES (The Intelligence Layer)
// ============================================================================

// findLeadsWhoViewedProperty - For price changes, notify warm leads who viewed the property
func (eco *EventCampaignOrchestrator) findLeadsWhoViewedProperty(eventData map[string]interface{}) ([]string, error) {
	propertyIDFloat, ok := eventData["property_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("property_id missing from event data")
	}
	propertyID := int64(propertyIDFloat)

	// Query behavioral_events for leads who viewed this property in last 30 days
	var events []models.BehavioralEvent
	err := eco.db.Raw(`
		SELECT DISTINCT lead_id 
		FROM behavioral_events 
		WHERE property_id = ? 
		AND event_type IN ('viewed', 'saved', 'inquired')
		AND created_at > NOW() - INTERVAL '30 days'
		AND lead_id IS NOT NULL
	`, propertyID).Scan(&events).Error

	if err != nil {
		return nil, err
	}

	// Convert lead IDs to FUB contact IDs
	contactIDs := []string{}
	for _, event := range events {
		lead, err := eco.getFUBContactIDForLead(event.LeadID)
		if err == nil && lead != "" {
			contactIDs = append(contactIDs, lead)
		}
	}

	return eco.deduplicateAndFilter(contactIDs), nil
}

// findLeadsMatchingProperty - For new listings, find leads with matching search preferences
func (eco *EventCampaignOrchestrator) findLeadsMatchingProperty(eventData map[string]interface{}) ([]string, error) {
	// Extract property attributes
	zipCode, _ := eventData["zip_code"].(string)
	_ , _ = eventData["price"].(float64)
	_, _ = eventData["bedrooms"].(float64)

	if zipCode == "" {
		return nil, fmt.Errorf("zip_code missing from event data")
	}

	// Query leads with behavioral data showing interest in this area/price range
	var leadIDs []int64
	err := eco.db.Raw(`
		SELECT DISTINCT be.lead_id
		FROM behavioral_events be
		WHERE be.event_data->>'zip_code' = ?
		AND be.created_at > NOW() - INTERVAL '90 days'
		AND be.lead_id IS NOT NULL
		ORDER BY be.created_at DESC
		LIMIT 50
	`, zipCode).Scan(&leadIDs).Error

	if err != nil {
		return nil, err
	}

	// Also include leads with high behavioral scores who are active
	var hotLeadIDs []int64
	eco.db.Raw(`
		SELECT lead_id 
		FROM behavioral_scores 
		WHERE overall_score >= 70 
		AND updated_at > NOW() - INTERVAL '7 days'
		LIMIT 20
	`).Scan(&hotLeadIDs)

	// Combine and convert to contact IDs
	allLeadIDs := append(leadIDs, hotLeadIDs...)
	contactIDs := []string{}
	for _, leadID := range allLeadIDs {
		contactID, err := eco.getFUBContactIDForLead(leadID)
		if err == nil && contactID != "" {
			contactIDs = append(contactIDs, contactID)
		}
	}

	return eco.deduplicateAndFilter(contactIDs), nil
}

// findHotLead - For hot lead scoring, return the lead who just got scored
func (eco *EventCampaignOrchestrator) findHotLead(eventData map[string]interface{}) ([]string, error) {
	leadIDFloat, ok := eventData["lead_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("lead_id missing from event data")
	}

	contactID, err := eco.getFUBContactIDForLead(int64(leadIDFloat))
	if err != nil {
		return nil, err
	}

	return []string{contactID}, nil
}

// findShowingAttendee - For showing follow-ups, return the attendee
func (eco *EventCampaignOrchestrator) findShowingAttendee(eventData map[string]interface{}) ([]string, error) {
	contactID, ok := eventData["contact_id"].(string)
	if !ok || contactID == "" {
		return nil, fmt.Errorf("contact_id missing from event data")
	}

	return []string{contactID}, nil
}

// findApplicant - For application confirmations, return the applicant
func (eco *EventCampaignOrchestrator) findApplicant(eventData map[string]interface{}) ([]string, error) {
	contactID, ok := eventData["contact_id"].(string)
	if !ok || contactID == "" {
		return nil, fmt.Errorf("contact_id missing from event data")
	}

	return []string{contactID}, nil
}

// findPreviouslyInterestedLeads - For relisted properties, find leads who showed interest before
func (eco *EventCampaignOrchestrator) findPreviouslyInterestedLeads(eventData map[string]interface{}) ([]string, error) {
	propertyIDFloat, ok := eventData["property_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("property_id missing from event data")
	}
	propertyID := int64(propertyIDFloat)

	// Find leads who viewed/saved/inquired about this property in the past
	var leadIDs []int64
	err := eco.db.Raw(`
		SELECT DISTINCT lead_id 
		FROM behavioral_events 
		WHERE property_id = ? 
		AND event_type IN ('viewed', 'saved', 'inquired', 'applied')
		AND lead_id IS NOT NULL
		ORDER BY created_at DESC
		LIMIT 25
	`, propertyID).Scan(&leadIDs).Error

	if err != nil {
		return nil, err
	}

	contactIDs := []string{}
	for _, leadID := range leadIDs {
		contactID, err := eco.getFUBContactIDForLead(leadID)
		if err == nil && contactID != "" {
			contactIDs = append(contactIDs, contactID)
		}
	}

	return eco.deduplicateAndFilter(contactIDs), nil
}

// findDormantLead - For re-engagement, return the dormant lead
func (eco *EventCampaignOrchestrator) findDormantLead(eventData map[string]interface{}) ([]string, error) {
	leadIDFloat, ok := eventData["lead_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("lead_id missing from event data")
	}

	contactID, err := eco.getFUBContactIDForLead(int64(leadIDFloat))
	if err != nil {
		return nil, err
	}

	return []string{contactID}, nil
}

// findTenantByProperty - For lease renewals, find current tenant
func (eco *EventCampaignOrchestrator) findTenantByProperty(eventData map[string]interface{}) ([]string, error) {
	propertyAddress, ok := eventData["property_address"].(string)
	if !ok || propertyAddress == "" {
		return nil, fmt.Errorf("property_address missing from event data")
	}

	// Query closing_pipeline for current tenant (this would be enhanced with tenant tracking)
	var contactID string
	err := eco.db.Raw(`
		SELECT contact_id 
		FROM closing_pipeline 
		WHERE address = ? 
		AND status = 'rented'
		AND contact_id IS NOT NULL
		LIMIT 1
	`, propertyAddress).Scan(&contactID).Error

	if err != nil || contactID == "" {
		return []string{}, nil // No tenant found
	}

	return []string{contactID}, nil
}

// ============================================================================
// TEMPLATE BUILDERS (The Message Composers)
// ============================================================================

// buildPriceChangeTemplate - Creates personalized price reduction message
func (eco *EventCampaignOrchestrator) buildPriceChangeTemplate(eventData map[string]interface{}) (string, string, error) {
	address, _ := eventData["property_address"].(string)
	oldPriceFloat, _ := eventData["old_price"].(float64)
	newPriceFloat, _ := eventData["new_price"].(float64)
	oldPrice := int(oldPriceFloat)
	newPrice := int(newPriceFloat)
	reduction := oldPrice - newPrice

	subject := fmt.Sprintf("🏡 Price Reduced: %s - Now $%d!", address, newPrice)
	
	body := fmt.Sprintf(`Hi {{first_name}},

Great news! The property you viewed at %s just reduced its price by $%d!

New Price: $%d (was $%d)

This property won't last long at this price. Want to schedule a showing?

Book your tour: https://chrisgross-ctrl-project.com/booking?property=%s

Best,
Christopher Gross
Landlords of Texas
(713) 555-0123`, address, reduction, newPrice, oldPrice, strings.ReplaceAll(address, " ", "-"))

	return subject, body, nil
}

// buildNewListingTemplate - Creates new property match notification
func (eco *EventCampaignOrchestrator) buildNewListingTemplate(eventData map[string]interface{}) (string, string, error) {
	address, _ := eventData["property_address"].(string)
	priceFloat, _ := eventData["price"].(float64)
	bedsFloat, _ := eventData["bedrooms"].(float64)
	bathsFloat, _ := eventData["bathrooms"].(float64)
	price := int(priceFloat)
	beds := int(bedsFloat)
	baths := int(bathsFloat)

	subject := fmt.Sprintf("🆕 New Listing Alert: %d/%d at %s", beds, baths, address)
	
	body := fmt.Sprintf(`Hi {{first_name}},

A new property just hit the market that matches your search preferences!

📍 %s
🛏️  %d bed / %d bath
💰 $%d/month

This is a hot property in a great neighborhood. Be the first to see it!

Schedule your showing: https://chrisgross-ctrl-project.com/booking?property=%s

Christopher Gross
Landlords of Texas`, address, beds, baths, price, strings.ReplaceAll(address, " ", "-"))

	return subject, body, nil
}

// buildHotLeadTemplate - Creates personalized VIP outreach for hot leads
func (eco *EventCampaignOrchestrator) buildHotLeadTemplate(eventData map[string]interface{}) (string, string, error) {
	_, _ = eventData["overall_score"].(float64)

	subject := "🌟 Your Houston Property Search - Let's Find Your Perfect Home"
	
	body := fmt.Sprintf(`Hi {{first_name}},

I noticed you've been actively searching for properties in Houston. I'd love to help you find the perfect place!

Based on your activity, I have several properties that might be exactly what you're looking for. Can we schedule a quick 10-minute call to discuss your needs?

Reply to this email or call me directly: (713) 555-0123

I'm here to make your property search effortless.

Best regards,
Christopher Gross
Landlords of Texas

P.S. - I have access to properties before they hit the public market. Let's talk!`)

	return subject, body, nil
}

// buildShowingFollowUpTemplate - Creates follow-up after property showing
func (eco *EventCampaignOrchestrator) buildShowingFollowUpTemplate(eventData map[string]interface{}) (string, string, error) {
	propertyAddress, _ := eventData["property_address"].(string)

	subject := fmt.Sprintf("How was your showing at %s?", propertyAddress)
	
	body := fmt.Sprintf(`Hi {{first_name}},

Thanks for viewing %s today! I hope you loved it as much as I thought you would.

Do you have any questions about the property? Ready to submit an application?

📝 Apply now: https://chrisgross-ctrl-project.com/apply?property=%s
📞 Call me: (713) 555-0123
📧 Reply to this email

Don't wait - great properties get snatched up fast in Houston!

Best,
Christopher Gross`, propertyAddress, strings.ReplaceAll(propertyAddress, " ", "-"))

	return subject, body, nil
}

// buildApplicationConfirmationTemplate - Confirms application submission
func (eco *EventCampaignOrchestrator) buildApplicationConfirmationTemplate(eventData map[string]interface{}) (string, string, error) {
	propertyAddress, _ := eventData["property_address"].(string)

	subject := "✅ Application Received - Next Steps"
	
	body := fmt.Sprintf(`Hi {{first_name}},

Great news! We've received your application for %s.

Here's what happens next:
1. We'll review your application within 24-48 hours
2. We'll run background and credit checks
3. You'll hear from us with a decision

Keep your phone handy - I'll be calling you soon!

Questions? Call me anytime: (713) 555-0123

Excited to (hopefully!) welcome you as a tenant!

Christopher Gross
Landlords of Texas`, propertyAddress)

	return subject, body, nil
}

// buildRelistedTemplate - Notifies about property back on market
func (eco *EventCampaignOrchestrator) buildRelistedTemplate(eventData map[string]interface{}) (string, string, error) {
	address, _ := eventData["property_address"].(string)
	priceFloat, _ := eventData["price"].(float64)
	price := int(priceFloat)

	subject := fmt.Sprintf("🔄 Back on Market: %s", address)
	
	body := fmt.Sprintf(`Hi {{first_name}},

Remember %s that you were interested in?

It's back on the market at $%d/month!

The previous applicant fell through - this is your second chance to grab this property.

Book a showing NOW: https://chrisgross-ctrl-project.com/booking?property=%s

Act fast - it won't be available for long!

Christopher Gross
(713) 555-0123`, address, price, strings.ReplaceAll(address, " ", "-"))

	return subject, body, nil
}

// buildReengagementTemplate - Re-engages dormant leads
func (eco *EventCampaignOrchestrator) buildReengagementTemplate(eventData map[string]interface{}) (string, string, error) {
	subject := "Still looking for a place in Houston? 🏡"
	
	body := `Hi {{first_name}},

It's been a while since we last connected! Are you still searching for a property in Houston?

The market has changed a LOT recently - I have new properties that might be perfect for you.

Let's reconnect:
📞 Quick call: (713) 555-0123
📧 Reply to this email
🏡 Browse new listings: https://chrisgross-ctrl-project.com/properties

Looking forward to helping you find your perfect place!

Christopher Gross
Landlords of Texas`

	return subject, body, nil
}

// buildLeaseRenewalTemplate - Sends lease renewal opportunity
func (eco *EventCampaignOrchestrator) buildLeaseRenewalTemplate(eventData map[string]interface{}) (string, string, error) {
	propertyAddress, _ := eventData["property_address"].(string)
	leaseEndDate, _ := eventData["lease_end_date"].(string)

	subject := fmt.Sprintf("Time to Renew: %s", propertyAddress)
	
	body := fmt.Sprintf(`Hi {{first_name}},

Your lease at %s ends on %s.

We'd love to have you stay! Let's discuss your renewal options.

I can offer:
✅ Flexible lease terms
✅ Competitive renewal rates
✅ Priority maintenance

Let's chat: Call me at (713) 555-0123 or reply to this email.

Best,
Christopher Gross
Landlords of Texas`, propertyAddress, leaseEndDate)

	return subject, body, nil
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// triggerCampaign sends the campaign via automation service
func (eco *EventCampaignOrchestrator) triggerCampaign(contactID, subject, body string, data map[string]interface{}) error {
	// Use existing automation service infrastructure
	return eco.automationService.TriggerAutomation("campaign_triggered", data)
}

// logCampaignExecution logs to campaign_executions table
func (eco *EventCampaignOrchestrator) logCampaignExecution(contactID, eventType, campaignName string, eventData map[string]interface{}) {
	// Store in CampaignExecution model for tracking
	execution := models.CampaignExecution{
		Status: "sent",
	}
	
	executedAt := time.Now()
	execution.ExecutedAt = &executedAt
	
	eco.db.Table("campaign_executions").Create(&execution)
}

// recordExecutionSummary records high-level campaign metrics
func (eco *EventCampaignOrchestrator) recordExecutionSummary(eventType, campaignName string, targetsFound, sent int, eventData map[string]interface{}) {
	eventDataJSON, _ := json.Marshal(eventData)
	
	log := CampaignExecutionLog{
		EventType:    eventType,
		CampaignName: campaignName,
		TargetsFound: targetsFound,
		MessagesSent: sent,
		ExecutedAt:   time.Now(),
		EventData:    string(eventDataJSON),
	}

	// Store in system_alerts or custom campaign_logs table
	eco.db.Table("campaign_execution_logs").Create(&log)
}

// isRecentDuplicate checks if same event was processed recently (prevents spam)
func (eco *EventCampaignOrchestrator) isRecentDuplicate(eventType string, eventData map[string]interface{}, cooldownHours int) bool {
	cutoff := time.Now().Add(-time.Duration(cooldownHours) * time.Hour)
	
	// Check if similar campaign was executed recently
	var count int64
	eco.db.Table("campaign_execution_logs").
		Where("event_type = ? AND executed_at > ?", eventType, cutoff).
		Count(&count)
	
	return count > 0
}

// getFUBContactIDForLead converts internal lead_id to FUB contact ID
func (eco *EventCampaignOrchestrator) getFUBContactIDForLead(leadID int64) (string, error) {
	var contactID string
	err := eco.db.Table("leads").
		Select("fub_contact_id").
		Where("id = ?", leadID).
		Scan(&contactID).Error
	
	if err != nil {
		return "", err
	}
	
	return contactID, nil
}

// deduplicateAndFilter removes duplicates and filters out invalid contact IDs
func (eco *EventCampaignOrchestrator) deduplicateAndFilter(contactIDs []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, id := range contactIDs {
		if id != "" && !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	
	return result
}
