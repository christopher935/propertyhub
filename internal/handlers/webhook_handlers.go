package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WebhookHandlers provides comprehensive webhook processing capabilities
type WebhookHandlers struct {
	db          *gorm.DB
	twilioSID   string
	twilioToken string
	twilioPhone string
	fubAPIKey   string
}

// WebhookEvent represents an incoming webhook event
type WebhookEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Source      string    `gorm:"not null" json:"source"`      // "twilio", "fub", "har", etc.
	EventType   string    `gorm:"not null" json:"event_type"`  // "sms_received", "lead_updated", etc.
	Payload     string    `gorm:"type:text" json:"payload"`    // JSON payload
	Headers     string    `gorm:"type:text" json:"headers"`    // HTTP headers
	Signature   string    `json:"signature,omitempty"`         // Webhook signature
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Status      string    `gorm:"default:'pending'" json:"status"` // "pending", "processed", "failed"
	ErrorMsg    string    `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TwilioSMSWebhook represents Twilio SMS webhook payload
type TwilioSMSWebhook struct {
	MessageSid string `json:"MessageSid"`
	From       string `json:"From"`
	To         string `json:"To"`
	Body       string `json:"Body"`
	NumMedia   string `json:"NumMedia"`
	MediaUrl0  string `json:"MediaUrl0,omitempty"`
	AccountSid string `json:"AccountSid"`
	ApiVersion string `json:"ApiVersion"`
}

// FUBWebhook represents Follow Up Boss webhook payload
type FUBWebhook struct {
	Event     string      `json:"event"`
	EventID   string      `json:"eventId"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// InboundEmail represents parsed inbound email data
type InboundEmail struct {
	From     string            `json:"from"`
	To       string            `json:"to"`
	Subject  string            `json:"subject"`
	Body     string            `json:"body"`
	Headers  map[string]string `json:"headers"`
	ThreadID string            `json:"thread_id,omitempty"`
}

// NewWebhookHandlers creates new webhook handlers
func NewWebhookHandlers(db *gorm.DB) *WebhookHandlers {
	return &WebhookHandlers{
		db:          db,
		twilioSID:   os.Getenv("TWILIO_ACCOUNT_SID"),
		twilioToken: os.Getenv("TWILIO_AUTH_TOKEN"),
		twilioPhone: os.Getenv("TWILIO_PHONE_NUMBER"),
		fubAPIKey:   os.Getenv("FUB_API_KEY"),
	}
}

// ProcessTwilioWebhook handles incoming Twilio webhooks (SMS, Voice, etc.)
// POST /webhooks/twilio/sms
func (w *WebhookHandlers) ProcessTwilioWebhook(c *gin.Context) {
	// Verify Twilio signature for security
	signature := c.GetHeader("X-Twilio-Signature")
	if !w.verifyTwilioSignature(c.Request, signature) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid Twilio signature", nil)
		return
	}

	// Parse form data (Twilio sends form-encoded data)
	if err := c.Request.ParseForm(); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to parse form data", err)
		return
	}

	// Extract SMS data
	smsData := TwilioSMSWebhook{
		MessageSid: c.Request.FormValue("MessageSid"),
		From:       c.Request.FormValue("From"),
		To:         c.Request.FormValue("To"),
		Body:       c.Request.FormValue("Body"),
		NumMedia:   c.Request.FormValue("NumMedia"),
		MediaUrl0:  c.Request.FormValue("MediaUrl0"),
		AccountSid: c.Request.FormValue("AccountSid"),
		ApiVersion: c.Request.FormValue("ApiVersion"),
	}

	// Log the webhook event
	headersJSON, _ := json.Marshal(w.getRequestHeaders(c.Request))
	payloadJSON, _ := json.Marshal(smsData)

	event := WebhookEvent{
		Source:    "twilio",
		EventType: "sms_received",
		Payload:   string(payloadJSON),
		Headers:   string(headersJSON),
		Signature: signature,
		Status:    "pending",
	}

	if err := w.db.Create(&event).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to log webhook", err)
		return
	}

	// Process the SMS in the background
	go w.processTwilioSMS(event.ID, smsData)

	// Return TwiML response to acknowledge receipt
	response := `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Message>Thank you for your message! We'll get back to you soon. For immediate assistance, call (713) 555-0123.</Message>
</Response>`

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, response)
}

// ProcessFUBWebhook handles Follow Up Boss webhooks
// POST /webhooks/fub
func (w *WebhookHandlers) ProcessFUBWebhook(c *gin.Context) {
	// Read the payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to read payload", err)
		return
	}

	// SECURITY: Always verify FUB webhook signature
	if !w.verifyFUBWebhook(c, body) {
		return
	}

	// Parse FUB webhook
	var fubData FUBWebhook
	if err := json.Unmarshal(body, &fubData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid FUB webhook format", err)
		return
	}

	// Log the webhook event
	headersJSON, _ := json.Marshal(w.getRequestHeaders(c.Request))
	event := WebhookEvent{
		Source:    "fub",
		EventType: fubData.Event,
		Payload:   string(body),
		Headers:   string(headersJSON),
		Signature: signature,
		Status:    "pending",
	}

	if err := w.db.Create(&event).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to log webhook", err)
		return
	}

	// Process the FUB event in the background
	go w.processFUBEvent(event.ID, fubData)

	utils.SuccessResponse(c, gin.H{
		"status":  "received",
		"eventId": fubData.EventID,
	})
}

// ProcessInboundEmail handles inbound emails (if using email processing)
// POST /webhooks/email
func (w *WebhookHandlers) ProcessInboundEmail(c *gin.Context) {
	var emailData InboundEmail
	if err := c.ShouldBindJSON(&emailData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid email format", err)
		return
	}

	// Log the webhook event
	payloadJSON, _ := json.Marshal(emailData)
	headersJSON, _ := json.Marshal(w.getRequestHeaders(c.Request))
	
	event := WebhookEvent{
		Source:    "email",
		EventType: "email_received",
		Payload:   string(payloadJSON),
		Headers:   string(headersJSON),
		Status:    "pending",
	}

	if err := w.db.Create(&event).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to log webhook", err)
		return
	}

	// Process the email in the background
	go w.processInboundEmail(event.ID, emailData)

	utils.SuccessResponse(c, gin.H{
		"status":    "received",
		"thread_id": emailData.ThreadID,
	})
}

// GetWebhookEvents retrieves webhook events with filtering
// GET /api/v1/webhooks/events
func (w *WebhookHandlers) GetWebhookEvents(c *gin.Context) {
	source := c.Query("source")
	status := c.Query("status")
	limit := c.DefaultQuery("limit", "50")

	query := w.db.Model(&WebhookEvent{})
	
	if source != "" {
		query = query.Where("source = ?", source)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var events []WebhookEvent
	err := query.Order("created_at DESC").
		Limit(parseInt(limit, 50)).
		Find(&events).Error

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve events", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"events": events,
		"total":  len(events),
	})
}

// GetWebhookStats provides webhook processing statistics
// GET /api/v1/webhooks/stats
func (w *WebhookHandlers) GetWebhookStats(c *gin.Context) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	type SourceCount struct {
		Source string `json:"source"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	var sourceCounts []SourceCount

	// Get counts by status
	w.db.Model(&WebhookEvent{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts)

	// Get counts by source
	w.db.Model(&WebhookEvent{}).
		Select("source, count(*) as count").
		Group("source").
		Find(&sourceCounts)

	// Get total count
	var totalCount int64
	w.db.Model(&WebhookEvent{}).Count(&totalCount)

	// Get recent activity (last 24 hours)
	var recentCount int64
	yesterday := time.Now().AddDate(0, 0, -1)
	w.db.Model(&WebhookEvent{}).
		Where("created_at > ?", yesterday).
		Count(&recentCount)

	utils.SuccessResponse(c, gin.H{
		"total_events":    totalCount,
		"recent_events":   recentCount,
		"status_breakdown": statusCounts,
		"source_breakdown": sourceCounts,
	})
}

// Background processing functions

func (w *WebhookHandlers) processTwilioSMS(eventID uint, smsData TwilioSMSWebhook) {
	defer w.markEventProcessed(eventID)

	// Extract useful information from SMS
	phoneNumber := strings.TrimSpace(smsData.From)
	message := strings.TrimSpace(smsData.Body)

	// Simple keyword processing
	response := w.generateSMSResponse(message)
	
	// Send response via Twilio
	if err := w.sendTwilioSMS(phoneNumber, response); err != nil {
		w.markEventFailed(eventID, fmt.Sprintf("Failed to send SMS response: %v", err))
		return
	}

	// Create or update contact in database
	if err := w.processInquiry(phoneNumber, message, "sms"); err != nil {
		// Log error but don't fail the webhook processing
		fmt.Printf("Failed to process inquiry: %v\n", err)
	}
}

func (w *WebhookHandlers) processFUBEvent(eventID uint, fubData FUBWebhook) {
	defer w.markEventProcessed(eventID)

	// Process different FUB event types
	switch fubData.Event {
	case "person.created":
		w.handleFUBPersonCreated(fubData.Data)
	case "person.updated":
		w.handleFUBPersonUpdated(fubData.Data)
	case "event.created":
		w.handleFUBEventCreated(fubData.Data)
	default:
		fmt.Printf("Unknown FUB event type: %s\n", fubData.Event)
	}
}

func (w *WebhookHandlers) processInboundEmail(eventID uint, emailData InboundEmail) {
	defer w.markEventProcessed(eventID)

	// Process the email content
	response := w.generateEmailResponse(emailData.Subject, emailData.Body)
	
	// Send auto-response if configured
	if response != "" {
		// Here you would integrate with your email sending service
		fmt.Printf("Would send email response to %s: %s\n", emailData.From, response)
	}

	// Create or update contact
	if err := w.processInquiry(emailData.From, emailData.Body, "email"); err != nil {
		fmt.Printf("Failed to process email inquiry: %v\n", err)
	}
}

// Verification functions

func (w *WebhookHandlers) verifyTwilioSignature(req *http.Request, signature string) bool {
	if w.twilioToken == "" || signature == "" {
		return false // Skip verification in development if token not set
	}

	// Build the signature base string
	url := req.URL.String()
	if req.Method == "POST" {
		req.ParseForm()
		params := req.PostForm
		for key, values := range params {
			for _, value := range values {
				url += key + value
			}
		}
	}

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(w.twilioToken))
	mac.Write([]byte(url))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

func (w *WebhookHandlers) verifyFUBSignature(payload []byte, signature string) bool {
	if w.fubAPIKey == "" {
		log.Printf("⚠️ FUB webhook rejected - API key not configured")
		return false
	}

	if signature == "" {
		log.Printf("⚠️ FUB webhook rejected - no signature provided")
		return false
	}

	if !strings.HasPrefix(signature, "sha256=") {
		log.Printf("⚠️ FUB webhook rejected - invalid signature format")
		return false
	}

	mac := hmac.New(sha256.New, []byte(w.fubAPIKey))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

func (w *WebhookHandlers) verifyFUBWebhook(c *gin.Context, body []byte) bool {
	signature := c.GetHeader("X-FUB-Signature")
	timestamp := c.GetHeader("X-FUB-Timestamp")

	if !w.verifyFUBSignature(body, signature) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or missing FUB signature", nil)
		return false
	}

	if timestamp != "" {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err == nil {
			webhookTime := time.Unix(ts, 0)
			if time.Since(webhookTime) > 5*time.Minute {
				log.Printf("⚠️ FUB webhook rejected - timestamp too old: %v", webhookTime)
				utils.ErrorResponse(c, http.StatusUnauthorized, "Webhook timestamp too old", nil)
				return false
			}
		}
	}

	return true
}

// Helper functions

func (w *WebhookHandlers) getRequestHeaders(req *http.Request) map[string]string {
	headers := make(map[string]string)
	for key, values := range req.Header {
		headers[key] = strings.Join(values, ", ")
	}
	return headers
}

func (w *WebhookHandlers) markEventProcessed(eventID uint) {
	now := time.Now()
	w.db.Model(&WebhookEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       "processed",
			"processed_at": &now,
		})
}

func (w *WebhookHandlers) markEventFailed(eventID uint, errorMsg string) {
	w.db.Model(&WebhookEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": errorMsg,
		})
}

func (w *WebhookHandlers) generateSMSResponse(message string) string {
	message = strings.ToLower(message)
	
	// Simple keyword-based responses
	if strings.Contains(message, "showing") || strings.Contains(message, "tour") {
		return "Great! I'd love to show you some properties. You can schedule a showing at https://chrisgross-ctrl-project.com/book-showing or reply with your preferred times."
	}
	
	if strings.Contains(message, "rent") || strings.Contains(message, "lease") {
		return "I have several rental properties available! Visit https://chrisgross-ctrl-project.com to see current listings or reply with your budget and preferred area."
	}
	
	if strings.Contains(message, "buy") || strings.Contains(message, "purchase") {
		return "I'd be happy to help you find the perfect home! Visit our listings at https://chrisgross-ctrl-project.com or reply with your budget and preferred areas."
	}
	
	// Default response
	return "Thank you for reaching out! I'm Christopher Gross, your Houston real estate expert. Reply with 'SHOWINGS' for tours, 'RENT' for rentals, or call (713) 555-0123."
}

func (w *WebhookHandlers) generateEmailResponse(subject, body string) string {
	// Generate contextual email responses based on content
	body = strings.ToLower(body)
	
	if strings.Contains(body, "showing") || strings.Contains(body, "tour") {
		return "Thank you for your interest in touring our properties! You can schedule a showing online at https://chrisgross-ctrl-project.com/book-showing"
	}
	
	// Return empty string if no auto-response needed
	return ""
}

func (w *WebhookHandlers) sendTwilioSMS(to, message string) error {
	if w.twilioSID == "" || w.twilioToken == "" {
		return fmt.Errorf("Twilio credentials not configured")
	}

	// Twilio API endpoint
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", w.twilioSID)

	// Prepare form data
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", w.twilioPhone)
	data.Set("Body", message)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(w.twilioSID, w.twilioToken)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Twilio API error: %s", string(body))
	}

	return nil
}

func (w *WebhookHandlers) processInquiry(contact, message, source string) error {
	// Create contact inquiry record
	// This would integrate with your existing contact/booking system
	fmt.Printf("Processing %s inquiry from %s: %s\n", source, contact, message)
	
	// Here you would:
	// 1. Create or update contact record
	// 2. Trigger FUB lead creation
	// 3. Set up appropriate follow-up sequences
	
	return nil
}

func (w *WebhookHandlers) handleFUBPersonCreated(data interface{}) {
	// Handle new person created in FUB
	fmt.Printf("FUB person created: %+v\n", data)
}

func (w *WebhookHandlers) handleFUBPersonUpdated(data interface{}) {
	// Handle person updated in FUB
	fmt.Printf("FUB person updated: %+v\n", data)
}

func (w *WebhookHandlers) handleFUBEventCreated(data interface{}) {
	// Handle new event created in FUB
	fmt.Printf("FUB event created: %+v\n", data)
}

func parseInt(s string, defaultVal int) int {
	if val, err := json.Number(s).Int64(); err == nil {
		return int(val)
	}
	return defaultVal
}

// RegisterWebhookRoutes registers all webhook routes
func RegisterWebhookRoutes(r *gin.Engine, db *gorm.DB) {
	webhooks := NewWebhookHandlers(db)
	
	// Webhook endpoints (no authentication for external services)
	webhookGroup := r.Group("/webhooks")
	{
		webhookGroup.POST("/twilio/sms", webhooks.ProcessTwilioWebhook)
		webhookGroup.POST("/fub", webhooks.ProcessFUBWebhook)
		webhookGroup.POST("/email", webhooks.ProcessInboundEmail)
	}
	
	// Admin API for webhook management
	api := r.Group("/api/v1")
	{
		webhookAPI := api.Group("/webhooks")
		{
			webhookAPI.GET("/events", webhooks.GetWebhookEvents)
			webhookAPI.GET("/stats", webhooks.GetWebhookStats)
		}
	}
}
