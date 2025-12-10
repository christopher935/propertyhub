package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SMSEmailAutomationService handles automated SMS and email communication via FUB
type SMSEmailAutomationService struct {
	db           *gorm.DB
	fubAPIKey    string
	twilioSID    string
	twilioToken  string
	twilioPhone  string
	httpClient   *http.Client
	mutex        sync.RWMutex
	emailService *EmailService
	smsService   *SMSService
}

// AutomationRule defines when and how to send automated messages
type AutomationRule struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	TriggerType string `gorm:"not null" json:"trigger_type"` // "booking_created", "lead_imported", "time_delay"
	Conditions  string `gorm:"type:json" json:"conditions"`  // JSON conditions
	MessageType string `gorm:"not null" json:"message_type"` // "sms", "email", "both"
	Template    string `gorm:"type:text" json:"template"`
	DelayHours  int    `gorm:"default:0" json:"delay_hours"`
	Active      bool   `gorm:"default:true" json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AutomationExecution tracks automation executions
type AutomationExecution struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RuleID      uint      `gorm:"not null" json:"rule_id"`
	ContactID   string    `gorm:"not null" json:"contact_id"` // FUB contact ID
	TriggerData string    `gorm:"type:json" json:"trigger_data"`
	Status      string    `gorm:"default:'pending'" json:"status"` // "pending", "sent", "failed"
	SentAt      *time.Time `json:"sent_at,omitempty"`
	ErrorMsg    string    `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AutomationFUBContact represents a Follow Up Boss contact for automation
type AutomationFUBContact struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Source   string `json:"source"`
	Tags     []string `json:"tags"`
	CustomFields map[string]interface{} `json:"customFields"`
}

// FUBMessage represents a message to send via FUB
type FUBMessage struct {
	PersonID    string `json:"personId"`
	Subject     string `json:"subject,omitempty"`
	Body        string `json:"body"`
	MessageType string `json:"messageType"` // "email", "text"
}

// AutomationStats represents automation statistics
type AutomationStats struct {
	TotalRules           int64   `json:"total_rules"`
	ActiveRules          int64   `json:"active_rules"`
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	SuccessRate          float64 `json:"success_rate"`
	PendingExecutions    int64   `json:"pending_executions"`
}

// NewSMSEmailAutomationService creates a new automation service
func NewSMSEmailAutomationService(db *gorm.DB) *SMSEmailAutomationService {
	return &SMSEmailAutomationService{
		db:          db,
		fubAPIKey:   os.Getenv("FUB_API_KEY"),
		twilioSID:   os.Getenv("TWILIO_ACCOUNT_SID"),
		twilioToken: os.Getenv("TWILIO_AUTH_TOKEN"), 
		twilioPhone: os.Getenv("TWILIO_PHONE_NUMBER"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetServices sets the email and SMS services for unsubscribe checking
func (s *SMSEmailAutomationService) SetServices(emailService *EmailService, smsService *SMSService) {
	s.emailService = emailService
	s.smsService = smsService
}

// TriggerAutomation triggers automation rules for a specific event
func (s *SMSEmailAutomationService) TriggerAutomation(triggerType string, data map[string]interface{}) error {
	// Find matching automation rules
	var rules []AutomationRule
	err := s.db.Where("trigger_type = ? AND active = ?", triggerType, true).Find(&rules).Error
	if err != nil {
		return fmt.Errorf("failed to find automation rules: %v", err)
	}

	log.Printf("🤖 Found %d automation rules for trigger: %s", len(rules), triggerType)

	// Process each rule
	for _, rule := range rules {
		// Check if conditions match
		if !s.evaluateConditions(rule.Conditions, data) {
			continue
		}

		// Create execution record
		execution := AutomationExecution{
			RuleID:      rule.ID,
			ContactID:   fmt.Sprintf("%v", data["contact_id"]),
			TriggerData: s.marshalJSON(data),
			Status:      "pending",
		}

		if err := s.db.Create(&execution).Error; err != nil {
			log.Printf("Failed to create execution record: %v", err)
			continue
		}

		// Schedule execution (immediate or delayed)
		if rule.DelayHours > 0 {
			go s.scheduleDelayedExecution(execution.ID, time.Duration(rule.DelayHours)*time.Hour)
		} else {
			go s.executeAutomationRule(execution.ID, rule, data)
		}
	}

	return nil
}

// scheduleDelayedExecution schedules a delayed automation execution
func (s *SMSEmailAutomationService) scheduleDelayedExecution(executionID uint, delay time.Duration) {
	time.Sleep(delay)
	
	// Reload the execution to make sure it's still valid
	var execution AutomationExecution
	if err := s.db.First(&execution, executionID).Error; err != nil {
		log.Printf("Failed to reload execution %d: %v", executionID, err)
		return
	}

	// Reload the rule
	var rule AutomationRule
	if err := s.db.First(&rule, execution.RuleID).Error; err != nil {
		log.Printf("Failed to reload rule %d: %v", execution.RuleID, err)
		return
	}

	// Parse trigger data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(execution.TriggerData), &data); err != nil {
		log.Printf("Failed to parse trigger data: %v", err)
		return
	}

	s.executeAutomationRule(executionID, rule, data)
}

// executeAutomationRule executes a specific automation rule
func (s *SMSEmailAutomationService) executeAutomationRule(executionID uint, rule AutomationRule, data map[string]interface{}) {
	log.Printf("🚀 Executing automation rule: %s", rule.Name)

	// Get contact information
	contact, err := s.getFUBContact(fmt.Sprintf("%v", data["contact_id"]))
	if err != nil {
		s.markExecutionFailed(executionID, fmt.Sprintf("Failed to get contact: %v", err))
		return
	}

	// Render the message template
	message := s.renderTemplate(rule.Template, data, contact)

	// Send the message based on type
	var sendErr error
	switch rule.MessageType {
	case "email":
		if s.isEmailUnsubscribed(contact.Email) {
			log.Printf("Skipping unsubscribed email recipient: %s", contact.Email)
			s.markExecutionFailed(executionID, "recipient unsubscribed from email")
			return
		}
		sendErr = s.sendEmailViaFUB(contact, "PropertyHub Update", message)
	case "sms":
		if s.isPhoneUnsubscribed(contact.Phone) {
			log.Printf("Skipping unsubscribed SMS recipient: %s", contact.Phone)
			s.markExecutionFailed(executionID, "recipient unsubscribed from SMS")
			return
		}
		sendErr = s.sendSMSViaFUB(contact, message)
	case "both":
		var emailErr, smsErr error
		if s.isEmailUnsubscribed(contact.Email) {
			log.Printf("Skipping unsubscribed email recipient: %s", contact.Email)
			emailErr = fmt.Errorf("recipient unsubscribed from email")
		} else {
			emailErr = s.sendEmailViaFUB(contact, "PropertyHub Update", message)
		}
		if s.isPhoneUnsubscribed(contact.Phone) {
			log.Printf("Skipping unsubscribed SMS recipient: %s", contact.Phone)
			smsErr = fmt.Errorf("recipient unsubscribed from SMS")
		} else {
			smsErr = s.sendSMSViaFUB(contact, message)
		}
		if emailErr != nil && smsErr != nil {
			sendErr = fmt.Errorf("email: %v, sms: %v", emailErr, smsErr)
		}
	}

	if sendErr != nil {
		s.markExecutionFailed(executionID, sendErr.Error())
		return
	}

	// Mark execution as successful
	now := time.Now()
	s.db.Model(&AutomationExecution{}).
		Where("id = ?", executionID).
		Updates(map[string]interface{}{
			"status":  "sent",
			"sent_at": &now,
		})

	log.Printf("✅ Automation executed successfully: %s", rule.Name)
}

// getFUBContact retrieves a contact from Follow Up Boss
func (s *SMSEmailAutomationService) getFUBContact(contactID string) (*AutomationFUBContact, error) {
	if s.fubAPIKey == "" {
		return nil, fmt.Errorf("FUB API key not configured")
	}

	url := fmt.Sprintf("https://api.followupboss.com/v1/people/%s", contactID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.fubAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FUB API returned status %d", resp.StatusCode)
	}

	var contact AutomationFUBContact
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

// sendEmailViaFUB sends an email through Follow Up Boss
func (s *SMSEmailAutomationService) sendEmailViaFUB(contact *AutomationFUBContact, subject, body string) error {
	if s.fubAPIKey == "" {
		log.Printf("📧 Would send email to %s: %s", contact.Email, subject)
		return nil // Mock in development
	}

	message := FUBMessage{
		PersonID:    contact.ID,
		Subject:     subject,
		Body:        body,
		MessageType: "email",
	}

	return s.sendFUBMessage(message)
}

// sendSMSViaFUB sends an SMS through Follow Up Boss
func (s *SMSEmailAutomationService) sendSMSViaFUB(contact *AutomationFUBContact, message string) error {
	if s.fubAPIKey == "" {
		log.Printf("📱 Would send SMS to %s: %s", contact.Phone, message)
		return nil // Mock in development
	}

	fubMessage := FUBMessage{
		PersonID:    contact.ID,
		Body:        message,
		MessageType: "text",
	}

	return s.sendFUBMessage(fubMessage)
}

// sendFUBMessage sends a message via Follow Up Boss API
func (s *SMSEmailAutomationService) sendFUBMessage(message FUBMessage) error {
	url := "https://api.followupboss.com/v1/communications"
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.fubAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("FUB API returned status %d", resp.StatusCode)
	}

	return nil
}

// renderTemplate renders a message template with data
func (s *SMSEmailAutomationService) renderTemplate(template string, data map[string]interface{}, contact *AutomationFUBContact) string {
	message := template
	
	// Replace placeholders with actual data
	replacements := map[string]string{
		"{{name}}":        contact.Name,
		"{{first_name}}":  s.getFirstName(contact.Name),
		"{{email}}":       contact.Email,
		"{{phone}}":       contact.Phone,
		"{{agent_name}}":  "Christopher Gross",
		"{{agent_phone}}": "(713) 555-0123",
		"{{company}}":     "Landlords of Texas",
		"{{website}}":     "https://chrisgross-ctrl-project.com",
	}

	// Add dynamic data replacements
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacements[placeholder] = fmt.Sprintf("%v", value)
	}

	// Perform replacements
	for placeholder, value := range replacements {
		message = strings.ReplaceAll(message, placeholder, value)
	}

	return message
}

// evaluateConditions evaluates if automation conditions are met
func (s *SMSEmailAutomationService) evaluateConditions(conditions string, data map[string]interface{}) bool {
	if conditions == "" || conditions == "{}" {
		return true // No conditions means always execute
	}

	// Parse conditions JSON
	var conditionMap map[string]interface{}
	if err := json.Unmarshal([]byte(conditions), &conditionMap); err != nil {
		log.Printf("Failed to parse conditions: %v", err)
		return false
	}

	// Simple condition evaluation (can be enhanced)
	for key, expectedValue := range conditionMap {
		if actualValue, exists := data[key]; !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

// getFirstName extracts first name from full name
func (s *SMSEmailAutomationService) getFirstName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[0]
	}
	return fullName
}

// markExecutionFailed marks an automation execution as failed
func (s *SMSEmailAutomationService) markExecutionFailed(executionID uint, errorMsg string) {
	s.db.Model(&AutomationExecution{}).
		Where("id = ?", executionID).
		Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": errorMsg,
		})
	
	log.Printf("❌ Automation execution %d failed: %s", executionID, errorMsg)
}

// marshalJSON safely marshals data to JSON string
func (s *SMSEmailAutomationService) marshalJSON(data interface{}) string {
	if jsonData, err := json.Marshal(data); err == nil {
		return string(jsonData)
	}
	return "{}"
}

// isEmailUnsubscribed checks if email is unsubscribed using the EmailService
func (s *SMSEmailAutomationService) isEmailUnsubscribed(email string) bool {
	if s.emailService != nil {
		return s.emailService.IsUnsubscribed(email, "marketing")
	}
	if s.db == nil {
		return false
	}
	email = strings.ToLower(strings.TrimSpace(email))
	var count int64
	s.db.Table("unsubscribe_records").
		Where("LOWER(email) = ? AND (unsubscribe_type = 'marketing' OR unsubscribe_type = 'all') AND is_active = ? AND resubscribe_date IS NULL",
			email, true).
		Count(&count)
	return count > 0
}

// isPhoneUnsubscribed checks if phone is unsubscribed using the SMSService
func (s *SMSEmailAutomationService) isPhoneUnsubscribed(phone string) bool {
	if s.smsService != nil {
		return s.smsService.IsPhoneUnsubscribed(phone)
	}
	if s.db == nil {
		return false
	}
	normalizedPhone := s.normalizePhone(phone)
	var count int64
	s.db.Table("unsubscribe_records").
		Where("email = ? AND (unsubscribe_type = 'sms' OR unsubscribe_type = 'all') AND is_active = ? AND resubscribe_date IS NULL",
			normalizedPhone, true).
		Count(&count)
	return count > 0
}

// normalizePhone removes non-digit characters from phone number
func (s *SMSEmailAutomationService) normalizePhone(phone string) string {
	var normalized strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			normalized.WriteRune(r)
		}
	}
	return normalized.String()
}

// SetupDefaultAutomationRules creates default automation rules
func (s *SMSEmailAutomationService) SetupDefaultAutomationRules() error {
	defaultRules := []AutomationRule{
		{
			Name:        "Welcome New Booking",
			TriggerType: "booking_created",
			Conditions:  "{}",
			MessageType: "both",
			Template:    "Hi {{first_name}}! Thanks for scheduling a showing with {{agent_name}}. I'll call you 24 hours before your appointment to confirm. Questions? Call {{agent_phone}}.",
			DelayHours:  0,
			Active:      true,
		},
		{
			Name:        "Booking Reminder",
			TriggerType: "booking_reminder",
			Conditions:  "{}",
			MessageType: "sms",
			Template:    "Reminder: Your property showing is tomorrow at {{showing_time}}. I'll see you at {{property_address}}! - {{agent_name}}",
			DelayHours:  0,
			Active:      true,
		},
		{
			Name:        "New Lead Welcome",
			TriggerType: "lead_created",
			Conditions:  "{}",
			MessageType: "email",
			Template:    "Welcome {{name}}! I'm {{agent_name}}, your Houston real estate expert. I've received your inquiry and will follow up within 15 minutes. In the meantime, browse our latest properties at {{website}}.",
			DelayHours:  0,
			Active:      true,
		},
		{
			Name:        "Follow Up After Showing",
			TriggerType: "showing_completed",
			Conditions:  "{}",
			MessageType: "sms",
			Template:    "Hi {{first_name}}! How did you like the property today? I'd love to hear your thoughts and answer any questions. - {{agent_name}}",
			DelayHours:  2,
			Active:      true,
		},
	}

	for _, rule := range defaultRules {
		// Check if rule already exists
		var existing AutomationRule
		result := s.db.Where("name = ?", rule.Name).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			// Create new rule
			if err := s.db.Create(&rule).Error; err != nil {
				log.Printf("Failed to create automation rule '%s': %v", rule.Name, err)
			} else {
				log.Printf("✅ Created automation rule: %s", rule.Name)
			}
		}
	}

	return nil
}

// GetAutomationStats returns automation statistics
func (s *SMSEmailAutomationService) GetAutomationStats() (*AutomationStats, error) {
	var totalRules int64
	var activeRules int64
	var totalExecutions int64
	var successfulExecutions int64

	s.db.Model(&AutomationRule{}).Count(&totalRules)
	s.db.Model(&AutomationRule{}).Where("active = ?", true).Count(&activeRules)
	s.db.Model(&AutomationExecution{}).Count(&totalExecutions)
	s.db.Model(&AutomationExecution{}).Where("status = 'sent'").Count(&successfulExecutions)

	// Get failed and pending executions
	var failedExecutions int64
	var pendingExecutions int64
	s.db.Model(&AutomationExecution{}).Where("status = 'failed'").Count(&failedExecutions)
	s.db.Model(&AutomationExecution{}).Where("status = 'pending'").Count(&pendingExecutions)

	successRate := float64(0)
	if totalExecutions > 0 {
		successRate = (float64(successfulExecutions) / float64(totalExecutions)) * 100
	}

	return &AutomationStats{
		TotalRules:           totalRules,
		ActiveRules:          activeRules,
		TotalExecutions:      totalExecutions,
		SuccessfulExecutions: successfulExecutions,
		FailedExecutions:     failedExecutions,
		SuccessRate:          successRate,
		PendingExecutions:    pendingExecutions,
	}, nil
}

// ProcessPendingAutomations processes any pending automation executions
func (s *SMSEmailAutomationService) ProcessPendingAutomations() error {
	// This would be called by a background job to handle any stuck executions
	var pendingExecutions []AutomationExecution
	cutoff := time.Now().Add(-1 * time.Hour) // Process executions older than 1 hour

	err := s.db.Where("status = 'pending' AND created_at < ?", cutoff).
		Find(&pendingExecutions).Error
	
	if err != nil {
		return err
	}

	log.Printf("Processing %d pending automation executions", len(pendingExecutions))

	for _, execution := range pendingExecutions {
		// Reload the rule and trigger data
		var rule AutomationRule
		if err := s.db.First(&rule, execution.RuleID).Error; err != nil {
			s.markExecutionFailed(execution.ID, fmt.Sprintf("Rule not found: %v", err))
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(execution.TriggerData), &data); err != nil {
			s.markExecutionFailed(execution.ID, fmt.Sprintf("Invalid trigger data: %v", err))
			continue
		}

		// Execute in background
		go s.executeAutomationRule(execution.ID, rule, data)
	}

	return nil
}
