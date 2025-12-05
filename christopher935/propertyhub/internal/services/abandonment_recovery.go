package services

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// AbandonmentRecoveryService handles sophisticated abandonment recovery campaigns
type AbandonmentRecoveryService struct {
	emailService      *EmailService
	smsService        *SMSService
	analyticsService  *AnalyticsAutomationService
	leadService       *LeadService
	propertyService   *PropertyService
	recoveryTemplates map[string]*RecoveryTemplate
	recoverySequences map[string]*RecoverySequence
	activeRecoveries  map[string]*ActiveRecovery
}

type RecoveryTemplate struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"` // "email" or "sms"
	Subject         string                 `json:"subject"`
	Content         string                 `json:"content"`
	Variables       []string               `json:"variables"`
	Personalization map[string]interface{} `json:"personalization"`
	CTAText         string                 `json:"ctaText"`
	CTALink         string                 `json:"ctaLink"`
	UrgencyLevel    int                    `json:"urgencyLevel"` // 1-5
	Timing          time.Duration          `json:"timing"`
}

type RecoverySequence struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Trigger    string         `json:"trigger"` // "property_selection", "date_selection", etc.
	Steps      []RecoveryStep `json:"steps"`
	Conditions []string       `json:"conditions"`
	Priority   int            `json:"priority"`
	Enabled    bool           `json:"enabled"`
}

type RecoveryStep struct {
	StepNumber int           `json:"stepNumber"`
	Delay      time.Duration `json:"delay"`
	TemplateID string        `json:"templateId"`
	Channel    string        `json:"channel"` // "email", "sms", "push"
	Conditions []string      `json:"conditions"`
	ABTest     *ABTestConfig `json:"abTest,omitempty"`
}

type ABTestConfig struct {
	Enabled    bool    `json:"enabled"`
	VariantA   string  `json:"variantA"`
	VariantB   string  `json:"variantB"`
	SplitRatio float64 `json:"splitRatio"` // 0.5 = 50/50 split
	MetricGoal string  `json:"metricGoal"` // "open_rate", "click_rate", "conversion"
}

type ActiveRecovery struct {
	UserID           string                 `json:"userId"`
	SessionID        string                 `json:"sessionId"`
	SequenceID       string                 `json:"sequenceId"`
	StartTime        time.Time              `json:"startTime"`
	CurrentStep      int                    `json:"currentStep"`
	CompletedSteps   []int                  `json:"completedSteps"`
	NextStepTime     time.Time              `json:"nextStepTime"`
	AbandonmentData  map[string]interface{} `json:"abandonmentData"`
	PersonalizedData map[string]interface{} `json:"personalizedData"`
	IsActive         bool                   `json:"isActive"`
	ConversionGoal   string                 `json:"conversionGoal"`
	ABTestVariant    string                 `json:"abTestVariant,omitempty"`
	Metrics          *RecoveryMetrics       `json:"metrics"`
}

type RecoveryMetrics struct {
	EmailsSent      int       `json:"emailsSent"`
	SMSSent         int       `json:"smsSent"`
	EmailsOpened    int       `json:"emailsOpened"`
	EmailsClicked   int       `json:"emailsClicked"`
	SMSClicked      int       `json:"smsClicked"`
	Conversions     int       `json:"conversions"`
	LastInteraction time.Time `json:"lastInteraction"`
	TotalSpent      float64   `json:"totalSpent"`
	ROI             float64   `json:"roi"`
}

// NewAbandonmentRecoveryService creates a new abandonment recovery service
func NewAbandonmentRecoveryService(emailService *EmailService, smsService *SMSService, analyticsService *AnalyticsAutomationService, leadService *LeadService, propertyService *PropertyService) *AbandonmentRecoveryService {
	service := &AbandonmentRecoveryService{
		emailService:      emailService,
		smsService:        smsService,
		analyticsService:  analyticsService,
		leadService:       leadService,
		propertyService:   propertyService,
		recoveryTemplates: make(map[string]*RecoveryTemplate),
		recoverySequences: make(map[string]*RecoverySequence),
		activeRecoveries:  make(map[string]*ActiveRecovery),
	}

	// Initialize default templates and sequences
	service.initializeDefaultTemplates()
	service.initializeDefaultSequences()

	// Start recovery processing routine
	go service.processRecoveryRoutine()

	return service
}

// StartRecovery initiates an abandonment recovery sequence
func (s *AbandonmentRecoveryService) StartRecovery(userID, sessionID string, abandonmentData map[string]interface{}) error {
	log.Printf("🚨 Starting abandonment recovery for user %s, session %s", userID, sessionID)

	// Get user information
	user, err := s.leadService.GetLeadByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user info: %v", err)
	}

	// Determine appropriate recovery sequence
	sequenceID := s.selectRecoverySequence(abandonmentData)
	sequence, exists := s.recoverySequences[sequenceID]
	if !exists {
		return fmt.Errorf("recovery sequence not found: %s", sequenceID)
	}

	// Check if user already has an active recovery
	if existingRecovery, exists := s.activeRecoveries[userID]; exists && existingRecovery.IsActive {
		log.Printf("User %s already has active recovery, updating...", userID)
		return s.updateExistingRecovery(existingRecovery, abandonmentData)
	}

	// Create personalized data
	personalizedData := s.createPersonalizedData(user, abandonmentData)

	// Determine A/B test variant if applicable
	abTestVariant := s.determineABTestVariant(sequence)

	// Create active recovery
	recovery := &ActiveRecovery{
		UserID:           userID,
		SessionID:        sessionID,
		SequenceID:       sequenceID,
		StartTime:        time.Now(),
		CurrentStep:      0,
		CompletedSteps:   []int{},
		NextStepTime:     time.Now().Add(sequence.Steps[0].Delay),
		AbandonmentData:  abandonmentData,
		PersonalizedData: personalizedData,
		IsActive:         true,
		ConversionGoal:   "booking_completion",
		ABTestVariant:    abTestVariant,
		Metrics: &RecoveryMetrics{
			LastInteraction: time.Now(),
		},
	}

	s.activeRecoveries[userID] = recovery

	log.Printf("✅ Recovery sequence '%s' started for user %s", sequence.Name, userID)

	// Track recovery start
	s.trackRecoveryEvent("recovery_started", recovery, map[string]interface{}{
		"sequence_id":   sequenceID,
		"sequence_name": sequence.Name,
		"ab_variant":    abTestVariant,
	})

	return nil
}

func (s *AbandonmentRecoveryService) selectRecoverySequence(abandonmentData map[string]interface{}) string {
	abandonedStep, _ := abandonmentData["step_abandoned"].(string)
	engagementScore, _ := abandonmentData["engagement_score"].(float64)
	propertiesViewed, _ := abandonmentData["properties_viewed"].([]interface{})

	// High-value user sequences
	if engagementScore > 50 || len(propertiesViewed) > 3 {
		switch abandonedStep {
		case "property_selection":
			return "high_value_property_abandonment"
		case "date_time_selection":
			return "high_value_scheduling_abandonment"
		case "contact_information":
			return "high_value_contact_abandonment"
		default:
			return "high_value_generic_abandonment"
		}
	}

	// Standard sequences
	switch abandonedStep {
	case "property_selection":
		return "standard_property_abandonment"
	case "date_time_selection":
		return "standard_scheduling_abandonment"
	case "contact_information":
		return "standard_contact_abandonment"
	default:
		return "standard_generic_abandonment"
	}
}

func (s *AbandonmentRecoveryService) createPersonalizedData(user interface{}, abandonmentData map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})

	// Add user-specific data
	if userMap, ok := user.(map[string]interface{}); ok {
		data["user_name"] = userMap["name"]
		data["user_email"] = userMap["email"]
		data["user_phone"] = userMap["phone"]
	}

	// Add abandonment context
	data["abandoned_step"] = abandonmentData["step_abandoned"]
	data["time_spent"] = abandonmentData["time_before_abandonment"]

	// Add property context
	if propertiesViewed, ok := abandonmentData["properties_viewed"].([]interface{}); ok && len(propertiesViewed) > 0 {
		// Get details for the last viewed property
		if property, err := s.propertyService.GetPropertyByID("1"); err == nil {
			data["last_property"] = property
			data["property_address"] = property["address"]
			data["property_price"] = property["price"]
			data["property_type"] = property["type"]
		}

		data["properties_viewed_count"] = len(propertiesViewed)
		data["multiple_properties"] = len(propertiesViewed) > 1
	}

	// Add urgency factors
	data["urgency_message"] = s.generateUrgencyMessage(abandonmentData)
	data["scarcity_message"] = s.generateScarcityMessage(abandonmentData)

	// Add incentives
	data["special_offer"] = s.generateSpecialOffer(abandonmentData)

	// Add recovery link
	data["recovery_link"] = fmt.Sprintf("https://elitepropertyshowings.com/booking/resume?session=%s&recovery=true", abandonmentData["session_id"])

	// Add social proof
	data["social_proof"] = s.generateSocialProof()

	// Add agent information
	data["assigned_agent"] = s.getAssignedAgent(user)

	return data
}

func (s *AbandonmentRecoveryService) determineABTestVariant(sequence *RecoverySequence) string {
	for _, step := range sequence.Steps {
		if step.ABTest != nil && step.ABTest.Enabled {
			if rand.Float64() < step.ABTest.SplitRatio {
				return step.ABTest.VariantA
			}
			return step.ABTest.VariantB
		}
	}
	return ""
}

func (s *AbandonmentRecoveryService) processRecoveryRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.processActiveRecoveries()
	}
}

func (s *AbandonmentRecoveryService) processActiveRecoveries() {
	now := time.Now()

	for userID, recovery := range s.activeRecoveries {
		if !recovery.IsActive {
			continue
		}

		// Check if it's time for the next step
		if now.After(recovery.NextStepTime) {
			err := s.executeRecoveryStep(recovery)
			if err != nil {
				log.Printf("❌ Failed to execute recovery step for user %s: %v", userID, err)
				continue
			}
		}

		// Check for recovery timeout (7 days)
		if now.Sub(recovery.StartTime) > 7*24*time.Hour {
			s.completeRecovery(recovery, "timeout")
		}
	}
}

func (s *AbandonmentRecoveryService) executeRecoveryStep(recovery *ActiveRecovery) error {
	sequence := s.recoverySequences[recovery.SequenceID]
	if recovery.CurrentStep >= len(sequence.Steps) {
		s.completeRecovery(recovery, "sequence_completed")
		return nil
	}

	step := sequence.Steps[recovery.CurrentStep]

	// Check step conditions
	if !s.checkStepConditions(step, recovery) {
		log.Printf("⏭️ Skipping step %d for user %s - conditions not met", step.StepNumber, recovery.UserID)
		s.advanceToNextStep(recovery)
		return nil
	}

	// Get template (with A/B test variant if applicable)
	templateID := step.TemplateID
	if recovery.ABTestVariant != "" && step.ABTest != nil {
		if recovery.ABTestVariant == step.ABTest.VariantB {
			templateID = step.ABTest.VariantB
		}
	}

	template, exists := s.recoveryTemplates[templateID]
	if !exists {
		return fmt.Errorf("template not found: %s", templateID)
	}

	// Execute step based on channel
	switch step.Channel {
	case "email":
		err := s.sendRecoveryEmail(recovery, template)
		if err != nil {
			return fmt.Errorf("failed to send recovery email: %v", err)
		}
		recovery.Metrics.EmailsSent++

	case "sms":
		err := s.sendRecoverySMS(recovery, template)
		if err != nil {
			return fmt.Errorf("failed to send recovery SMS: %v", err)
		}
		recovery.Metrics.SMSSent++
	}

	// Track step execution
	s.trackRecoveryEvent("recovery_step_executed", recovery, map[string]interface{}{
		"step_number": step.StepNumber,
		"channel":     step.Channel,
		"template_id": templateID,
		"ab_variant":  recovery.ABTestVariant,
	})

	// Advance to next step
	s.advanceToNextStep(recovery)

	log.Printf("📧 Executed recovery step %d for user %s via %s", step.StepNumber, recovery.UserID, step.Channel)

	return nil
}

func (s *AbandonmentRecoveryService) sendRecoveryEmail(recovery *ActiveRecovery, template *RecoveryTemplate) error {
	// Personalize email content
	subject := s.personalizeContent(template.Subject, recovery.PersonalizedData)
	content := s.personalizeContent(template.Content, recovery.PersonalizedData)

	// Add tracking parameters
	trackingParams := map[string]string{
		"recovery_id": recovery.UserID,
		"step":        fmt.Sprintf("%d", recovery.CurrentStep),
		"variant":     recovery.ABTestVariant,
	}

	emailData := map[string]interface{}{
		"subject":           subject,
		"content":           content,
		"cta_text":          template.CTAText,
		"cta_link":          s.addTrackingToLink(template.CTALink, trackingParams),
		"personalized_data": recovery.PersonalizedData,
		"template_id":       template.ID,
		"recovery_id":       recovery.UserID,
	}

	userEmail := recovery.PersonalizedData["user_email"].(string)
	return s.emailService.SendTemplateEmail(userEmail, subject, template.ID, emailData)
}

func (s *AbandonmentRecoveryService) sendRecoverySMS(recovery *ActiveRecovery, template *RecoveryTemplate) error {
	// Personalize SMS content
	content := s.personalizeContent(template.Content, recovery.PersonalizedData)

	// Add tracking to links
	if template.CTALink != "" {
		trackingParams := map[string]string{
			"recovery_id": recovery.UserID,
			"step":        fmt.Sprintf("%d", recovery.CurrentStep),
			"variant":     recovery.ABTestVariant,
			"channel":     "sms",
		}
		content = strings.Replace(content, template.CTALink, s.addTrackingToLink(template.CTALink, trackingParams), 1)
	}

	userPhone := recovery.PersonalizedData["user_phone"].(string)
	return s.smsService.SendSMS(userPhone, content, map[string]interface{}{
		"recovery_id": recovery.UserID,
		"template_id": template.ID,
	})
}

func (s *AbandonmentRecoveryService) personalizeContent(content string, data map[string]interface{}) string {
	result := content

	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		if valueStr, ok := value.(string); ok {
			result = strings.Replace(result, placeholder, valueStr, -1)
		}
	}

	return result
}

func (s *AbandonmentRecoveryService) addTrackingToLink(link string, params map[string]string) string {
	if link == "" {
		return link
	}

	separator := "?"
	if strings.Contains(link, "?") {
		separator = "&"
	}

	var trackingParams []string
	for key, value := range params {
		trackingParams = append(trackingParams, fmt.Sprintf("%s=%s", key, value))
	}

	return fmt.Sprintf("%s%s%s", link, separator, strings.Join(trackingParams, "&"))
}

func (s *AbandonmentRecoveryService) checkStepConditions(step RecoveryStep, recovery *ActiveRecovery) bool {
	// Check if user has already converted
	if s.hasUserConverted(recovery.UserID) {
		return false
	}

	// Check custom conditions
	for _, condition := range step.Conditions {
		if !s.evaluateCondition(condition, recovery) {
			return false
		}
	}

	return true
}

func (s *AbandonmentRecoveryService) hasUserConverted(userID string) bool {
	// Check if user has completed a booking since recovery started
	// This would integrate with your booking service
	return false
}

func (s *AbandonmentRecoveryService) evaluateCondition(condition string, recovery *ActiveRecovery) bool {
	switch condition {
	case "high_engagement":
		if score, ok := recovery.AbandonmentData["engagement_score"].(float64); ok {
			return score > 30
		}
	case "multiple_properties":
		if props, ok := recovery.AbandonmentData["properties_viewed"].([]interface{}); ok {
			return len(props) > 1
		}
	case "business_hours":
		now := time.Now()
		hour := now.Hour()
		return hour >= 9 && hour <= 17 // 9 AM to 5 PM
	case "weekday":
		now := time.Now()
		weekday := now.Weekday()
		return weekday >= time.Monday && weekday <= time.Friday
	}
	return true
}

func (s *AbandonmentRecoveryService) advanceToNextStep(recovery *ActiveRecovery) {
	recovery.CompletedSteps = append(recovery.CompletedSteps, recovery.CurrentStep)
	recovery.CurrentStep++

	sequence := s.recoverySequences[recovery.SequenceID]
	if recovery.CurrentStep < len(sequence.Steps) {
		nextStep := sequence.Steps[recovery.CurrentStep]
		recovery.NextStepTime = time.Now().Add(nextStep.Delay)
	} else {
		s.completeRecovery(recovery, "sequence_completed")
	}
}

func (s *AbandonmentRecoveryService) completeRecovery(recovery *ActiveRecovery, reason string) {
	recovery.IsActive = false

	// Track recovery completion
	s.trackRecoveryEvent("recovery_completed", recovery, map[string]interface{}{
		"completion_reason": reason,
		"steps_completed":   len(recovery.CompletedSteps),
		"total_duration":    time.Since(recovery.StartTime).Seconds(),
		"final_metrics":     recovery.Metrics,
	})

	log.Printf("🏁 Recovery completed for user %s - reason: %s", recovery.UserID, reason)
}

func (s *AbandonmentRecoveryService) trackRecoveryEvent(eventName string, recovery *ActiveRecovery, additionalData map[string]interface{}) {
	eventData := map[string]interface{}{
		"user_id":          recovery.UserID,
		"session_id":       recovery.SessionID,
		"sequence_id":      recovery.SequenceID,
		"current_step":     recovery.CurrentStep,
		"ab_variant":       recovery.ABTestVariant,
		"recovery_metrics": recovery.Metrics,
	}

	for key, value := range additionalData {
		eventData[key] = value
	}

	// This would send to your analytics system
	log.Printf("📊 Recovery Event: %s - %v", eventName, eventData)
}

// Utility methods for generating dynamic content

func (s *AbandonmentRecoveryService) generateUrgencyMessage(abandonmentData map[string]interface{}) string {
	messages := []string{
		"Properties are booking fast in your area!",
		"Don't miss out on your perfect home!",
		"Limited availability - book your showing today!",
		"Other buyers are viewing this property too!",
	}
	return messages[rand.Intn(len(messages))]
}

func (s *AbandonmentRecoveryService) generateScarcityMessage(abandonmentData map[string]interface{}) string {
	messages := []string{
		"Only 3 showing slots left this week!",
		"This property has 5 other interested buyers!",
		"Price may increase next week!",
		"Similar properties sold within 48 hours!",
	}
	return messages[rand.Intn(len(messages))]
}

func (s *AbandonmentRecoveryService) generateSpecialOffer(abandonmentData map[string]interface{}) map[string]interface{} {
	offers := []map[string]interface{}{
		{
			"type":        "Free Home Inspection",
			"value":       "$500 value",
			"description": "Complimentary professional home inspection",
			"expires":     time.Now().Add(48 * time.Hour).Format("January 2, 2006"),
		},
		{
			"type":        "Priority Booking",
			"value":       "Exclusive access",
			"description": "Skip the line and book your preferred time",
			"expires":     time.Now().Add(24 * time.Hour).Format("January 2, 2006"),
		},
		{
			"type":        "Market Analysis",
			"value":       "$300 value",
			"description": "Free comparative market analysis report",
			"expires":     time.Now().Add(72 * time.Hour).Format("January 2, 2006"),
		},
	}
	return offers[rand.Intn(len(offers))]
}

func (s *AbandonmentRecoveryService) generateSocialProof() map[string]interface{} {
	proofs := []map[string]interface{}{
		{
			"type":    "recent_booking",
			"message": "Sarah J. just booked a showing for a similar property",
			"time":    "2 hours ago",
		},
		{
			"type":    "testimonial",
			"message": "\"Found my dream home in just 3 showings!\" - Mike T.",
			"rating":  5,
		},
		{
			"type":    "statistic",
			"message": "97% of our clients find their perfect home within 30 days",
			"source":  "Elite Property Showings",
		},
	}
	return proofs[rand.Intn(len(proofs))]
}

func (s *AbandonmentRecoveryService) getAssignedAgent(user interface{}) map[string]interface{} {
	// This would integrate with your agent assignment system
	agents := []map[string]interface{}{
		{
			"name":   "Sarah Johnson",
			"title":  "Senior Property Specialist",
			"phone":  "(555) 123-4567",
			"email":  "sarah@elitepropertyshowings.com",
			"photo":  "https://example.com/agents/sarah.jpg",
			"rating": 4.9,
		},
		{
			"name":   "Michael Chen",
			"title":  "Luxury Property Expert",
			"phone":  "(555) 987-6543",
			"email":  "michael@elitepropertyshowings.com",
			"photo":  "https://example.com/agents/michael.jpg",
			"rating": 4.8,
		},
	}
	return agents[rand.Intn(len(agents))]
}

// Initialize default templates and sequences

func (s *AbandonmentRecoveryService) initializeDefaultTemplates() {
	// Email templates
	s.recoveryTemplates["property_abandonment_email_1"] = &RecoveryTemplate{
		ID:      "property_abandonment_email_1",
		Name:    "Property Selection Abandonment - Email 1",
		Type:    "email",
		Subject: "{{user_name}}, don't miss out on {{property_address}}!",
		Content: `Hi {{user_name}},

I noticed you were looking at {{property_address}} but didn't complete your booking. This property is getting a lot of interest!

{{urgency_message}}

{{special_offer.description}} - but only if you book within the next 48 hours.

Ready to schedule your showing?`,
		CTAText:      "Complete My Booking",
		CTALink:      "{{recovery_link}}",
		UrgencyLevel: 3,
		Timing:       15 * time.Minute,
	}

	s.recoveryTemplates["property_abandonment_sms_1"] = &RecoveryTemplate{
		ID:           "property_abandonment_sms_1",
		Name:         "Property Selection Abandonment - SMS 1",
		Type:         "sms",
		Content:      "Hi {{user_name}}! The property at {{property_address}} you were viewing is in high demand. Complete your booking: {{recovery_link}}",
		CTALink:      "{{recovery_link}}",
		UrgencyLevel: 4,
		Timing:       30 * time.Minute,
	}

	// Add more templates...
	s.addMoreRecoveryTemplates()
}

func (s *AbandonmentRecoveryService) addMoreRecoveryTemplates() {
	// High-value user templates
	s.recoveryTemplates["high_value_abandonment_email"] = &RecoveryTemplate{
		ID:      "high_value_abandonment_email",
		Name:    "High Value User Abandonment Email",
		Type:    "email",
		Subject: "{{user_name}}, your VIP booking is waiting",
		Content: `Dear {{user_name}},

As one of our valued VIP clients, I wanted to personally reach out about your property booking.

I've reserved {{special_offer.description}} exclusively for you.

Your dedicated agent {{assigned_agent.name}} is standing by to assist.

Let's get your dream home tour scheduled!`,
		CTAText:      "Complete VIP Booking",
		CTALink:      "{{recovery_link}}",
		UrgencyLevel: 5,
		Timing:       10 * time.Minute,
	}

	// Final attempt templates
	s.recoveryTemplates["final_attempt_email"] = &RecoveryTemplate{
		ID:      "final_attempt_email",
		Name:    "Final Attempt Email",
		Type:    "email",
		Subject: "Last chance: {{property_address}} booking expires soon",
		Content: `{{user_name}},

This is my final attempt to help you secure your property showing.

Your booking for {{property_address}} will expire in 24 hours.

{{scarcity_message}}

If you're no longer interested, please let me know so I can help other buyers.

Otherwise, complete your booking now:`,
		CTAText:      "Secure My Booking Now",
		CTALink:      "{{recovery_link}}",
		UrgencyLevel: 5,
		Timing:       4 * 24 * time.Hour,
	}
}

func (s *AbandonmentRecoveryService) initializeDefaultSequences() {
	// Standard property abandonment sequence
	s.recoverySequences["standard_property_abandonment"] = &RecoverySequence{
		ID:       "standard_property_abandonment",
		Name:     "Standard Property Selection Abandonment",
		Trigger:  "property_selection",
		Priority: 1,
		Enabled:  true,
		Steps: []RecoveryStep{
			{
				StepNumber: 1,
				Delay:      15 * time.Minute,
				TemplateID: "property_abandonment_email_1",
				Channel:    "email",
				Conditions: []string{"business_hours"},
			},
			{
				StepNumber: 2,
				Delay:      30 * time.Minute,
				TemplateID: "property_abandonment_sms_1",
				Channel:    "sms",
				Conditions: []string{},
			},
			{
				StepNumber: 3,
				Delay:      24 * time.Hour,
				TemplateID: "property_abandonment_email_2",
				Channel:    "email",
				Conditions: []string{},
				ABTest: &ABTestConfig{
					Enabled:    true,
					VariantA:   "property_abandonment_email_2",
					VariantB:   "property_abandonment_email_2_variant",
					SplitRatio: 0.5,
					MetricGoal: "click_rate",
				},
			},
			{
				StepNumber: 4,
				Delay:      72 * time.Hour,
				TemplateID: "final_attempt_email",
				Channel:    "email",
				Conditions: []string{},
			},
		},
	}

	// High-value user sequence
	s.recoverySequences["high_value_property_abandonment"] = &RecoverySequence{
		ID:       "high_value_property_abandonment",
		Name:     "High Value Property Abandonment",
		Trigger:  "property_selection",
		Priority: 0, // Highest priority
		Enabled:  true,
		Steps: []RecoveryStep{
			{
				StepNumber: 1,
				Delay:      10 * time.Minute,
				TemplateID: "high_value_abandonment_email",
				Channel:    "email",
				Conditions: []string{},
			},
			{
				StepNumber: 2,
				Delay:      20 * time.Minute,
				TemplateID: "high_value_abandonment_sms",
				Channel:    "sms",
				Conditions: []string{},
			},
			{
				StepNumber: 3,
				Delay:      2 * time.Hour,
				TemplateID: "agent_personal_outreach",
				Channel:    "email",
				Conditions: []string{"high_engagement"},
			},
		},
	}

	// Add more sequences for different abandonment points
	s.addMoreRecoverySequences()
}

func (s *AbandonmentRecoveryService) addMoreRecoverySequences() {
	// Contact information abandonment
	s.recoverySequences["standard_contact_abandonment"] = &RecoverySequence{
		ID:       "standard_contact_abandonment",
		Name:     "Contact Information Abandonment",
		Trigger:  "contact_information",
		Priority: 1,
		Enabled:  true,
		Steps: []RecoveryStep{
			{
				StepNumber: 1,
				Delay:      5 * time.Minute,
				TemplateID: "contact_abandonment_email_1",
				Channel:    "email",
				Conditions: []string{},
			},
			{
				StepNumber: 2,
				Delay:      15 * time.Minute,
				TemplateID: "contact_abandonment_sms_1",
				Channel:    "sms",
				Conditions: []string{},
			},
		},
	}
}

// Public API methods

func (s *AbandonmentRecoveryService) GetActiveRecoveries() map[string]*ActiveRecovery {
	return s.activeRecoveries
}

func (s *AbandonmentRecoveryService) GetRecoveryMetrics(userID string) (*RecoveryMetrics, error) {
	recovery, exists := s.activeRecoveries[userID]
	if !exists {
		return nil, fmt.Errorf("no active recovery found for user: %s", userID)
	}
	return recovery.Metrics, nil
}

func (s *AbandonmentRecoveryService) StopRecovery(userID string, reason string) error {
	recovery, exists := s.activeRecoveries[userID]
	if !exists {
		return fmt.Errorf("no active recovery found for user: %s", userID)
	}

	s.completeRecovery(recovery, reason)
	return nil
}

func (s *AbandonmentRecoveryService) UpdateRecoveryData(userID string, newData map[string]interface{}) error {
	recovery, exists := s.activeRecoveries[userID]
	if !exists {
		return fmt.Errorf("no active recovery found for user: %s", userID)
	}

	// Update personalized data
	for key, value := range newData {
		recovery.PersonalizedData[key] = value
	}

	return nil
}

func (s *AbandonmentRecoveryService) updateExistingRecovery(recovery *ActiveRecovery, newAbandonmentData map[string]interface{}) error {
	// Update abandonment data with new information
	for key, value := range newAbandonmentData {
		recovery.AbandonmentData[key] = value
	}

	// Reset timing if user showed new engagement
	if newStep, ok := newAbandonmentData["step_abandoned"].(string); ok {
		if newStep != recovery.AbandonmentData["step_abandoned"] {
			// User progressed further, adjust recovery sequence
			recovery.CurrentStep = 0
			sequence := s.recoverySequences[recovery.SequenceID]
			if len(sequence.Steps) > 0 {
				recovery.NextStepTime = time.Now().Add(sequence.Steps[0].Delay)
			}
		}
	}

	log.Printf("🔄 Updated existing recovery for user %s", recovery.UserID)
	return nil
}
