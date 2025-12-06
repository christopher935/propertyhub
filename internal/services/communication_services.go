package services

import (
	"fmt"
	"log"
	"time"
	"gorm.io/gorm"
	"strconv"
	"chrisgross-ctrl-project/internal/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type EmailService struct {
	config       *config.Config
	sendgrid     *sendgrid.Client
	fromEmail    string
	fromName     string
	useSendGrid  bool
}

type SMSService struct {
	config       *config.Config
	twilioClient *twilio.RestClient
	fromPhone    string
	enabled      bool
}

type NotificationService struct {
	config       *config.Config
	emailService *EmailService
}

type LeadService struct {
}

type PropertyService struct {
}

type BehavioralLeadScoringService struct {
	db *gorm.DB
}

func NewEmailService(cfg *config.Config) *EmailService {
	service := &EmailService{
		config:    cfg,
		fromEmail: cfg.EmailFromAddress,
		fromName:  cfg.EmailFromName,
	}

	if cfg.SendGridAPIKey != "" {
		service.sendgrid = sendgrid.NewSendClient(cfg.SendGridAPIKey)
		service.useSendGrid = true
		log.Printf("✅ Email service initialized with SendGrid")
	} else {
		service.useSendGrid = false
		log.Printf("⚠️  Email service initialized in log-only mode (no SendGrid API key)")
	}

	return service
}

func NewSMSService(cfg *config.Config) *SMSService {
	service := &SMSService{
		config:    cfg,
		fromPhone: cfg.TwilioPhoneNumber,
	}

	if cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" {
		service.twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: cfg.TwilioAccountSID,
			Password: cfg.TwilioAuthToken,
		})
		service.enabled = true
		log.Printf("✅ SMS service initialized with Twilio")
	} else {
		service.enabled = false
		log.Printf("⚠️  SMS service initialized in log-only mode (no Twilio credentials)")
	}

	return service
}

func NewNotificationService(cfg *config.Config, emailService *EmailService) *NotificationService {
	return &NotificationService{
		config:       cfg,
		emailService: emailService,
	}
}

func NewLeadService() *LeadService {
	return &LeadService{}
}

func NewPropertyService() *PropertyService {
	return &PropertyService{}
}

func NewBehavioralLeadScoringService() *BehavioralLeadScoringService {
	return &BehavioralLeadScoringService{}
}

func (es *EmailService) SendEmail(to, subject, content string, metadata map[string]interface{}) error {
	log.Printf("📧 EMAIL: To=%s, Subject=%s", to, subject)

	if !es.useSendGrid {
		log.Printf("⚠️  SendGrid not configured, email logged only")
		return nil
	}

	from := mail.NewEmail(es.fromName, es.fromEmail)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, content, content)

	response, err := es.sendgrid.Send(message)
	if err != nil {
		log.Printf("❌ Email send failed: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		log.Printf("✅ Email sent successfully to %s, status: %d", to, response.StatusCode)
		return nil
	}

	log.Printf("❌ Email send failed with status %d: %s", response.StatusCode, response.Body)
	return fmt.Errorf("email send failed with status %d", response.StatusCode)
}

func (es *EmailService) SendTemplateEmail(to, subject, template string, data map[string]interface{}) error {
	log.Printf("📧 TEMPLATE EMAIL: To=%s, Subject=%s, Template=%s", to, subject, template)

	content := es.renderTemplate(template, data)
	return es.SendEmail(to, subject, content, nil)
}

func (es *EmailService) renderTemplate(template string, data map[string]interface{}) string {
	content := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = replaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}
	return content
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		index := indexOf(s, old)
		if index == -1 {
			result += s
			break
		}
		result += s[:index] + new
		s = s[index+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (ss *SMSService) SendSMS(to, content string, metadata map[string]interface{}) error {
	log.Printf("📱 SMS: To=%s, Content=%s", to, content)

	if !ss.enabled {
		log.Printf("⚠️  Twilio not configured, SMS logged only")
		return nil
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(ss.fromPhone)
	params.SetBody(content)

	resp, err := ss.twilioClient.Api.CreateMessage(params)
	if err != nil {
		log.Printf("❌ SMS send failed: %v", err)
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	log.Printf("✅ SMS sent successfully to %s, SID: %s", to, *resp.Sid)
	return nil
}

func (ss *SMSService) SendSMSSimple(to, content string) error {
	return ss.SendSMS(to, content, nil)
}

func (ss *SMSService) SendTemplateSMS(to, template string, data map[string]interface{}) error {
	log.Printf("📱 TEMPLATE SMS: To=%s, Template=%s", to, template)

	content := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = replaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}

	return ss.SendSMS(to, content, nil)
}

func (ns *NotificationService) SendNotification(userID, title, content string, metadata map[string]interface{}) error {
	log.Printf("🔔 NOTIFICATION: UserID=%s, Title=%s", userID, title)

	if ns.emailService == nil {
		log.Printf("⚠️  Email service not available, notification logged only")
		return nil
	}

	userEmail := ""
	if metadata != nil {
		if email, ok := metadata["email"].(string); ok {
			userEmail = email
		}
	}

	if userEmail == "" {
		log.Printf("⚠️  No email address provided in metadata, notification logged only")
		return nil
	}

	emailContent := fmt.Sprintf("%s\n\n%s", title, content)
	return ns.emailService.SendEmail(userEmail, title, emailContent, nil)
}

func (ns *NotificationService) SendScheduledNotification(userID, title, content string, scheduleTime time.Time, metadata map[string]interface{}) error {
	log.Printf("⏰ SCHEDULED NOTIFICATION: UserID=%s, Title=%s, ScheduleTime=%s", userID, title, scheduleTime.Format(time.RFC3339))

	if time.Now().After(scheduleTime) {
		return ns.SendNotification(userID, title, content, metadata)
	}

	log.Printf("⚠️  Scheduled notifications not fully implemented, would send at %s", scheduleTime.Format(time.RFC3339))
	return nil
}

func (ns *NotificationService) SendAgentAlert(agentID, title, content string, metadata map[string]interface{}) error {
	log.Printf("🚨 AGENT ALERT: AgentID=%s, Title=%s", agentID, title)

	if ns.emailService == nil {
		log.Printf("⚠️  Email service not available, agent alert logged only")
		return nil
	}

	agentEmail := ""
	if metadata != nil {
		if email, ok := metadata["agent_email"].(string); ok {
			agentEmail = email
		}
	}

	if agentEmail == "" {
		agentEmail = ns.config.BusinessEmail
		log.Printf("⚠️  No agent email in metadata, using business email: %s", agentEmail)
	}

	emailContent := fmt.Sprintf("🚨 AGENT ALERT 🚨\n\nAgent ID: %s\n\n%s\n\n%s", agentID, title, content)
	return ns.emailService.SendEmail(agentEmail, fmt.Sprintf("[ALERT] %s", title), emailContent, nil)
}

func (ls *LeadService) GetLead(leadID string) (map[string]interface{}, error) {
	log.Printf("🎯 GET LEAD: ID=%s", leadID)
	return map[string]interface{}{
		"id":     leadID,
		"status": "active",
	}, nil
}

func (ls *LeadService) UpdateLead(leadID string, data map[string]interface{}) error {
	log.Printf("🎯 UPDATE LEAD: ID=%s", leadID)
	return nil
}

func (ls *LeadService) CreateLead(data map[string]interface{}) (string, error) {
	leadID := fmt.Sprintf("lead_%d", time.Now().UnixNano())
	log.Printf("🎯 CREATE LEAD: ID=%s", leadID)
	return leadID, nil
}

func (ls *LeadService) GetLeadByUserID(userID string) (map[string]interface{}, error) {
	log.Printf("🎯 GET LEAD BY USER ID: UserID=%s", userID)
	return map[string]interface{}{
		"id":         fmt.Sprintf("lead_%s", userID),
		"user_id":    userID,
		"status":     "active",
		"email":      fmt.Sprintf("user%s@example.com", userID),
		"phone":      "+1234567890",
		"first_name": "John",
		"last_name":  "Doe",
	}, nil
}

func (ps *PropertyService) GetProperty(propertyID string) (map[string]interface{}, error) {
	log.Printf("🏠 GET PROPERTY: ID=%s", propertyID)
	return map[string]interface{}{
		"id":     propertyID,
		"status": "active",
	}, nil
}

func (ps *PropertyService) UpdateProperty(propertyID string, data map[string]interface{}) error {
	log.Printf("🏠 UPDATE PROPERTY: ID=%s", propertyID)
	return nil
}

func (ps *PropertyService) GetPropertyByID(propertyID string) (map[string]interface{}, error) {
	log.Printf("🏠 GET PROPERTY BY ID: ID=%s", propertyID)
	return map[string]interface{}{
		"id":     propertyID,
		"status": "active",
		"type":   "residential",
	}, nil
}

func (blss *BehavioralLeadScoringService) ScoreLead(leadID string, behaviorData map[string]interface{}) (float64, error) {
	log.Printf("📊 SCORE LEAD: ID=%s", leadID)
	
	if blss.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	
	leadIDInt, err := strconv.Atoi(leadID)
	if err != nil {
		return 0, fmt.Errorf("invalid lead ID: %w", err)
	}
	
	var compositeScore int
	err = blss.db.Table("behavioral_scores").
		Select("composite_score").
		Where("lead_id = ?", leadIDInt).
		Order("last_calculated DESC").
		Limit(1).
		Scan(&compositeScore).Error
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("failed to query behavioral score: %w", err)
	}
	
	if err == gorm.ErrRecordNotFound {
		return 0.0, nil
	}
	
	return float64(compositeScore), nil
}

func (blss *BehavioralLeadScoringService) GetLeadScore(leadID string) (float64, error) {
	log.Printf("📊 GET LEAD SCORE: ID=%s", leadID)
	
	if blss.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	
	leadIDInt, err := strconv.Atoi(leadID)
	if err != nil {
		return 0, fmt.Errorf("invalid lead ID: %w", err)
	}
	
	var compositeScore int
	err = blss.db.Table("behavioral_scores").
		Select("composite_score").
		Where("lead_id = ?", leadIDInt).
		Order("last_calculated DESC").
		Limit(1).
		Scan(&compositeScore).Error
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("failed to query behavioral score: %w", err)
	}
	
	if err == gorm.ErrRecordNotFound {
		return 0.0, nil
	}
	
	return float64(compositeScore), nil
}

func (blss *BehavioralLeadScoringService) UpdateLeadScore(leadID string, score float64, reason string) error {
	log.Printf("📊 UPDATE LEAD SCORE: ID=%s, Score=%.1f, Reason=%s", leadID, score, reason)
	return nil
}
