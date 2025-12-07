package services

import (
	"fmt"
	"log"
	"time"
	"gorm.io/gorm"
	"strconv"
	"chrisgross-ctrl-project/internal/config"
	"chrisgross-ctrl-project/internal/safety"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type EmailService struct {
	db            *gorm.DB
	sendgrid      *sendgrid.Client
	fromEmail     string
	fromName      string
	isConfigured  bool
}

type SMSService struct {
	db           *gorm.DB
	client       *twilio.RestClient
	from         string
	isConfigured bool
}

type NotificationService struct {
	emailService *EmailService
	db           *gorm.DB
}

type LeadService struct {
}

type PropertyService struct {
}

type BehavioralLeadScoringService struct {
	db *gorm.DB
}

func NewEmailService(cfg *config.Config, db *gorm.DB) *EmailService {
	if cfg.SendGridAPIKey == "" {
		log.Printf("⚠️  WARNING: SENDGRID_API_KEY not configured - email sending will be disabled")
		return &EmailService{
			db:           db,
			isConfigured: false,
		}
	}

	return &EmailService{
		db:           db,
		sendgrid:     sendgrid.NewSendClient(cfg.SendGridAPIKey),
		fromEmail:    cfg.EmailFromAddress,
		fromName:     cfg.EmailFromName,
		isConfigured: true,
	}
}

func NewSMSService(cfg *config.Config, db *gorm.DB) *SMSService {
	if cfg.TwilioAccountSID == "" || cfg.TwilioAuthToken == "" {
		log.Printf("⚠️  WARNING: Twilio credentials not configured - SMS sending will be disabled")
		return &SMSService{
			db:           db,
			isConfigured: false,
		}
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.TwilioAccountSID,
		Password: cfg.TwilioAuthToken,
	})

	return &SMSService{
		db:           db,
		client:       client,
		from:         cfg.TwilioPhoneNumber,
		isConfigured: true,
	}
}

func NewNotificationService(emailService *EmailService, db *gorm.DB) *NotificationService {
	return &NotificationService{
		emailService: emailService,
		db:           db,
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
	if !es.isConfigured {
		log.Printf("⚠️  Email not configured - would have sent: To=%s, Subject=%s", to, subject)
		return fmt.Errorf("email service not configured")
	}

	// Check safety controls before sending
	safetyControls := safety.GetSafetyControls()
	if err := safetyControls.ValidateAction("send_email"); err != nil {
		log.Printf("🚫 BLOCKED: Email to %s blocked by safety controls - %v", to, err)
		return fmt.Errorf("email sending blocked: %w", err)
	}

	from := mail.NewEmail(es.fromName, es.fromEmail)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, content, content)

	response, err := es.sendgrid.Send(message)
	if err != nil {
		log.Printf("❌ Email send failed: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		log.Printf("❌ Email send failed with status %d: %s", response.StatusCode, response.Body)
		return fmt.Errorf("email send failed with status %d", response.StatusCode)
	}

	log.Printf("📧 Email sent to %s, status: %d", to, response.StatusCode)
	return nil
}

func (es *EmailService) SendTemplateEmail(to, subject, template string, data map[string]interface{}) error {
	if !es.isConfigured {
		log.Printf("⚠️  Email not configured - would have sent template: To=%s, Subject=%s, Template=%s", to, subject, template)
		return fmt.Errorf("email service not configured")
	}

	content := fmt.Sprintf("Template: %s\nData: %v", template, data)
	
	return es.SendEmail(to, subject, content, data)
}

func (ss *SMSService) SendSMS(to, content string, metadata map[string]interface{}) error {
	if !ss.isConfigured {
		log.Printf("⚠️  SMS not configured - would have sent: To=%s, Content=%s", to, content)
		return fmt.Errorf("SMS service not configured")
	}

	// Check safety controls before sending
	safetyControls := safety.GetSafetyControls()
	if err := safetyControls.ValidateAction("send_sms"); err != nil {
		log.Printf("🚫 BLOCKED: SMS to %s blocked by safety controls - %v", to, err)
		return fmt.Errorf("SMS sending blocked: %w", err)
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(ss.from)
	params.SetBody(content)

	resp, err := ss.client.Api.CreateMessage(params)
	if err != nil {
		log.Printf("❌ SMS send failed: %v", err)
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	log.Printf("📱 SMS sent to %s, SID: %s", to, *resp.Sid)
	return nil
}

func (ss *SMSService) SendSMSSimple(to, content string) error {
	return ss.SendSMS(to, content, nil)
}

func (ss *SMSService) SendTemplateSMS(to, template string, data map[string]interface{}) error {
	if !ss.isConfigured {
		log.Printf("⚠️  SMS not configured - would have sent template: To=%s, Template=%s", to, template)
		return fmt.Errorf("SMS service not configured")
	}

	content := fmt.Sprintf("Template: %s - Data: %v", template, data)
	
	return ss.SendSMS(to, content, data)
}

func (ns *NotificationService) SendNotification(userID, title, content string, metadata map[string]interface{}) error {
	log.Printf("🔔 NOTIFICATION: UserID=%s, Title=%s", userID, title)
	
	if ns.db == nil {
		log.Printf("⚠️  Database not available for notification")
		return fmt.Errorf("database not configured")
	}

	var userEmail string
	err := ns.db.Table("users").Select("email").Where("id = ?", userID).Scan(&userEmail).Error
	if err != nil {
		log.Printf("⚠️  Could not find user email for notification: %v", err)
		return fmt.Errorf("user not found: %w", err)
	}

	if ns.emailService != nil && ns.emailService.isConfigured {
		return ns.emailService.SendEmail(userEmail, title, content, metadata)
	}

	log.Printf("⚠️  Email service not available - notification not sent")
	return fmt.Errorf("email service not configured")
}

func (ns *NotificationService) SendScheduledNotification(userID, title, content string, scheduleTime time.Time, metadata map[string]interface{}) error {
	log.Printf("⏰ SCHEDULED NOTIFICATION: UserID=%s, Title=%s, ScheduleTime=%s", userID, title, scheduleTime.Format(time.RFC3339))
	return nil
}

func (ns *NotificationService) SendAgentAlert(agentID, title, content string, metadata map[string]interface{}) error {
	log.Printf("🚨 AGENT ALERT: AgentID=%s, Title=%s", agentID, title)
	
	if ns.db == nil {
		log.Printf("⚠️  Database not available for agent alert")
		return fmt.Errorf("database not configured")
	}

	var agentEmail string
	err := ns.db.Table("users").Select("email").Where("id = ? AND role = 'agent'", agentID).Scan(&agentEmail).Error
	if err != nil {
		log.Printf("⚠️  Could not find agent email for alert: %v", err)
		return fmt.Errorf("agent not found: %w", err)
	}

	if ns.emailService != nil && ns.emailService.isConfigured {
		return ns.emailService.SendEmail(agentEmail, title, content, metadata)
	}

	log.Printf("⚠️  Email service not available - agent alert not sent")
	return fmt.Errorf("email service not configured")
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
