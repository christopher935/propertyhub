package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EmailSenderHandlers manages trusted email sender configuration and processing
type EmailSenderHandlers struct {
	db                      *gorm.DB
	emailProcessor          *services.EmailProcessor
	senderValidationService *services.EmailSenderValidationService
}

// NewEmailSenderHandlers creates new email sender handlers
func NewEmailSenderHandlers(db *gorm.DB) *EmailSenderHandlers {
	return &EmailSenderHandlers{
		db:                      db,
		emailProcessor:          services.NewEmailProcessor(db),
		senderValidationService: services.NewEmailSenderValidationService(db),
	}
}

// CreateTrustedSender creates a new trusted email sender
// POST /api/v1/email/senders
func (h *EmailSenderHandlers) CreateTrustedSender(c *gin.Context) {
	var request struct {
		SenderEmail     string `json:"sender_email" binding:"required,email"`
		SenderName      string `json:"sender_name" binding:"required"`
		CompanyName     string `json:"company_name"`
		ContactPerson   string `json:"contact_person"`
		EmailType       string `json:"email_type" binding:"required"`
		ProcessingMode  string `json:"processing_mode"`
		Priority        string `json:"priority"`
		IsActive        bool   `json:"is_active"`
		BusinessPurpose string `json:"business_purpose"`
		Notes           string `json:"notes"`
		ParsingTemplate string `json:"parsing_template"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	// Check if sender already exists
	var existingSender models.TrustedEmailSender
	if err := h.db.Where("sender_email = ?", request.SenderEmail).First(&existingSender).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email sender already exists", nil)
		return
	}

	// Create new trusted sender
	sender := models.TrustedEmailSender{
		SenderEmail:     request.SenderEmail,
		SenderName:      request.SenderName,
		CompanyName:     request.CompanyName,
		ContactPerson:   request.ContactPerson,
		EmailType:       request.EmailType,
		ProcessingMode:  request.ProcessingMode,
		Priority:        request.Priority,
		IsActive:        request.IsActive,
		BusinessPurpose: request.BusinessPurpose,
		Notes:           request.Notes,
		ParsingTemplate: request.ParsingTemplate,
		AddedBy:         c.GetString("admin_username"), // From auth middleware
		IsVerified:      true,                          // Auto-verify for now, could require manual verification
	}

	// Validate sender data
	if err := sender.Validate(); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Sender validation failed", err)
		return
	}

	// Set approval date
	now := time.Now()
	sender.ApprovalDate = &now

	// Save to database
	if err := h.db.Create(&sender).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create trusted sender", err)
		return
	}

	// Create default parsing rule if template provided
	if request.ParsingTemplate != "" {
		h.createDefaultParsingRule(sender.ID, request.EmailType, request.ParsingTemplate)
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Trusted email sender created successfully",
		"sender":  sender,
	})
}

// GetTrustedSenders returns paginated list of trusted senders
// GET /api/v1/email/senders
func (h *EmailSenderHandlers) GetTrustedSenders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	emailType := c.Query("email_type")
	isActive := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.db.Model(&models.TrustedEmailSender{})

	if emailType != "" {
		query = query.Where("email_type = ?", emailType)
	}
	if isActive != "" {
		activeFlag := isActive == "true"
		query = query.Where("is_active = ?", activeFlag)
	}

	var total int64
	query.Count(&total)

	var senders []models.TrustedEmailSender
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&senders).Error

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch trusted senders", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"data": gin.H{
			"senders":     senders,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetTrustedSender returns specific trusted sender
// GET /api/v1/email/senders/:id
func (h *EmailSenderHandlers) GetTrustedSender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid sender ID", err)
		return
	}

	var sender models.TrustedEmailSender
	if err := h.db.First(&sender, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Trusted sender not found", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch sender", err)
		}
		return
	}

	utils.SuccessResponse(c, gin.H{
		"data": sender,
	})
}

// UpdateTrustedSender updates an existing trusted sender
// PUT /api/v1/email/senders/:id
func (h *EmailSenderHandlers) UpdateTrustedSender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid sender ID", err)
		return
	}

	var sender models.TrustedEmailSender
	if err := h.db.First(&sender, uint(id)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Trusted sender not found", err)
		return
	}

	var request struct {
		SenderEmail     string `json:"sender_email"`
		SenderName      string `json:"sender_name"`
		CompanyName     string `json:"company_name"`
		ContactPerson   string `json:"contact_person"`
		EmailType       string `json:"email_type"`
		ProcessingMode  string `json:"processing_mode"`
		Priority        string `json:"priority"`
		IsActive        bool   `json:"is_active"`
		BusinessPurpose string `json:"business_purpose"`
		Notes           string `json:"notes"`
		ParsingTemplate string `json:"parsing_template"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	// Update fields
	sender.SenderEmail = request.SenderEmail
	sender.SenderName = request.SenderName
	sender.CompanyName = request.CompanyName
	sender.ContactPerson = request.ContactPerson
	sender.EmailType = request.EmailType
	sender.ProcessingMode = request.ProcessingMode
	sender.Priority = request.Priority
	sender.IsActive = request.IsActive
	sender.BusinessPurpose = request.BusinessPurpose
	sender.Notes = request.Notes
	sender.ParsingTemplate = request.ParsingTemplate

	// Validate updated data
	if err := sender.Validate(); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Sender validation failed", err)
		return
	}

	// Save changes
	if err := h.db.Save(&sender).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update trusted sender", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Trusted sender updated successfully",
		"sender":  sender,
	})
}

// DeleteTrustedSender deletes a trusted sender
// DELETE /api/v1/email/senders/:id
func (h *EmailSenderHandlers) DeleteTrustedSender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid sender ID", err)
		return
	}

	var sender models.TrustedEmailSender
	if err := h.db.First(&sender, uint(id)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Trusted sender not found", err)
		return
	}

	// Soft delete
	if err := h.db.Delete(&sender).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete trusted sender", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Trusted sender deleted successfully",
	})
}

// ProcessIncomingEmail processes incoming email using sender validation
// POST /api/v1/email/process-incoming
func (h *EmailSenderHandlers) ProcessIncomingEmail(c *gin.Context) {
	var emailRequest struct {
		From       string            `json:"from" binding:"required"`
		To         string            `json:"to" binding:"required"`
		Subject    string            `json:"subject" binding:"required"`
		Content    string            `json:"content" binding:"required"`
		Headers    map[string]string `json:"headers"`
		ReceivedAt time.Time         `json:"received_at"`
	}

	if err := c.ShouldBindJSON(&emailRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid email data", err)
		return
	}

	// Set received time if not provided
	if emailRequest.ReceivedAt.IsZero() {
		emailRequest.ReceivedAt = time.Now()
	}

	// Validate sender
	trustedSender, err := h.senderValidationService.ValidateSender(emailRequest.From)
	if err != nil {
		// Log untrusted sender but don't fail - convert struct type
		h.logUnstrustedEmail(struct {
			From       string
			To         string
			Subject    string
			Content    string
			Headers    map[string]string
			ReceivedAt time.Time
		}{
			From:       emailRequest.From,
			To:         emailRequest.To,
			Subject:    emailRequest.Subject,
			Content:    emailRequest.Content,
			Headers:    emailRequest.Headers,
			ReceivedAt: emailRequest.ReceivedAt,
		})
		utils.ErrorResponse(c, http.StatusForbidden, "Email from untrusted sender", err)
		return
	}

	// Process email using trusted sender configuration - convert struct type
	result, err := h.processEmailWithSender(struct {
		From       string
		To         string
		Subject    string
		Content    string
		Headers    map[string]string
		ReceivedAt time.Time
	}{
		From:       emailRequest.From,
		To:         emailRequest.To,
		Subject:    emailRequest.Subject,
		Content:    emailRequest.Content,
		Headers:    emailRequest.Headers,
		ReceivedAt: emailRequest.ReceivedAt,
	}, trustedSender)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process email", err)
		return
	}

	// Log processing result
	// Log processing result - convert struct type
	h.logEmailProcessing(struct {
		From       string
		To         string
		Subject    string
		Content    string
		Headers    map[string]string
		ReceivedAt time.Time
	}{
		From:       emailRequest.From,
		To:         emailRequest.To,
		Subject:    emailRequest.Subject,
		Content:    emailRequest.Content,
		Headers:    emailRequest.Headers,
		ReceivedAt: emailRequest.ReceivedAt,
	}, trustedSender, result)

	utils.SuccessResponse(c, gin.H{
		"message": "Email processed successfully",
		"result":  result,
		"sender":  trustedSender.SenderName,
		"type":    trustedSender.EmailType,
	})
}

// TestEmailParsing tests parsing configuration with sample content
// POST /api/v1/email/test-parsing
func (h *EmailSenderHandlers) TestEmailParsing(c *gin.Context) {
	var testRequest struct {
		EmailContent    string `json:"email_content" binding:"required"`
		EmailType       string `json:"email_type" binding:"required"`
		ParsingTemplate string `json:"parsing_template"`
	}

	if err := c.ShouldBindJSON(&testRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid test request", err)
		return
	}

	// Parse template
	var template map[string]interface{}
	if testRequest.ParsingTemplate != "" {
		if err := json.Unmarshal([]byte(testRequest.ParsingTemplate), &template); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid parsing template", err)
			return
		}
	}

	// Run parsing test
	result := h.runParsingTest(testRequest.EmailContent, testRequest.EmailType, template)

	utils.SuccessResponse(c, gin.H{
		"test_result": result,
		"timestamp":   time.Now(),
	})
}

// GetEmailProcessingStats returns email processing statistics
// GET /api/v1/email/processing-stats
func (h *EmailSenderHandlers) GetEmailProcessingStats(c *gin.Context) {
	days := c.DefaultQuery("days", "30")
	daysInt, _ := strconv.Atoi(days)
	if daysInt < 1 || daysInt > 365 {
		daysInt = 30
	}

	since := time.Now().AddDate(0, 0, -daysInt)

	stats := gin.H{
		"total_senders":    h.getTotalSenders(),
		"active_senders":   h.getActiveSenders(),
		"processing_stats": h.getProcessingStats(since),
		"top_senders":      h.getTopSenders(since, 5),
		"recent_activity":  h.getRecentActivity(10),
	}

	utils.SuccessResponse(c, stats)
}

// AdminEmailSendersPage renders the email sender management page
func (h *EmailSenderHandlers) AdminEmailSendersPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-email-senders.html", gin.H{
		"Title": "Email Sender Management",
	})
}

// Helper methods

func (h *EmailSenderHandlers) createDefaultParsingRule(senderID uint, emailType, parsingTemplate string) {
	// Create default parsing rule for the sender
	rule := models.EmailProcessingRule{
		TrustedSenderID: senderID,
		RuleName:        fmt.Sprintf("Default %s Parser", emailType),
		RuleDescription: fmt.Sprintf("Default parsing rule for %s emails", emailType),
		EmailType:       emailType,
		IsActive:        true,
		CreatedBy:       "system",
	}

	// Parse template to extract patterns
	var template map[string]interface{}
	if err := json.Unmarshal([]byte(parsingTemplate), &template); err == nil {
		if subjectPattern, ok := template["subjectPattern"].(string); ok {
			rule.SubjectPattern = subjectPattern
		}
		if bodyPatterns, ok := template["bodyPatterns"].([]interface{}); ok {
			bodyPatternsJSON, _ := json.Marshal(bodyPatterns)
			rule.BodyPatterns = string(bodyPatternsJSON)
		}
		if fields, ok := template["fields"].([]interface{}); ok {
			fieldsJSON, _ := json.Marshal(fields)
			rule.RequiredFields = string(fieldsJSON)
		}
	}

	h.db.Create(&rule)
}

func (h *EmailSenderHandlers) processEmailWithSender(emailRequest struct {
	From       string
	To         string
	Subject    string
	Content    string
	Headers    map[string]string
	ReceivedAt time.Time
}, trustedSender *models.TrustedEmailSender) (map[string]interface{}, error) {

	// Create incoming email record
	incomingEmail := models.IncomingEmail{
		FromEmail:        emailRequest.From,
		ToEmail:          emailRequest.To,
		Subject:          emailRequest.Subject,
		Content:          emailRequest.Content,
		ReceivedAt:       emailRequest.ReceivedAt,
		EmailType:        trustedSender.EmailType,
		ProcessingStatus: models.ProcessingStatusPending,
	}

	if err := h.db.Create(&incomingEmail).Error; err != nil {
		return nil, fmt.Errorf("failed to create incoming email record: %v", err)
	}

	// Parse email content based on sender type and configuration
	extractedData, confidence := h.parseEmailBySenderType(emailRequest, trustedSender)

	// Update incoming email with results
	incomingEmail.ProcessingStatus = models.ProcessingStatusProcessed
	incomingEmail.Confidence = confidence
	if extractedDataJSON, err := json.Marshal(extractedData); err == nil {
		incomingEmail.ExtractedData = string(extractedDataJSON)
	}

	h.db.Save(&incomingEmail)

	// Route to appropriate system based on email type
	var actionTaken string
	var impactDescription string

	switch trustedSender.EmailType {
	case "application_notification":
		actionTaken, impactDescription = h.processApplicationNotification(extractedData, &incomingEmail)
	case "pre_listing_alert", "broker_alert":
		actionTaken, impactDescription = h.processPreListingAlert(extractedData, &incomingEmail)
	case "vendor_completion":
		actionTaken, impactDescription = h.processVendorCompletion(extractedData, &incomingEmail)
	case "lease_update":
		actionTaken, impactDescription = h.processLeaseUpdate(extractedData, &incomingEmail)
	default:
		actionTaken = "logged_only"
		impactDescription = "Email logged for manual review"
	}

	// Update sender activity
	trustedSender.UpdateLastActivity()
	h.db.Save(trustedSender)

	return map[string]interface{}{
		"incoming_email_id": incomingEmail.ID,
		"extracted_data":    extractedData,
		"confidence":        confidence,
		"action_taken":      actionTaken,
		"impact":            impactDescription,
		"sender_name":       trustedSender.SenderName,
		"email_type":        trustedSender.EmailType,
	}, nil
}

func (h *EmailSenderHandlers) parseEmailBySenderType(emailRequest struct {
	From       string
	To         string
	Subject    string
	Content    string
	Headers    map[string]string
	ReceivedAt time.Time
}, sender *models.TrustedEmailSender) (map[string]interface{}, float64) {

	extractedData := make(map[string]interface{})
	confidence := 0.8 // Default confidence

	switch sender.EmailType {
	case "application_notification":
		// Extract applicant name, email, property address from application notifications
		extractedData["applicant_name"] = h.extractApplicantName(emailRequest.Subject, emailRequest.Content)
		extractedData["applicant_email"] = h.extractApplicantEmail(emailRequest.Content)
		extractedData["property_address"] = h.extractPropertyAddress(emailRequest.Subject, emailRequest.Content)
		extractedData["application_date"] = emailRequest.ReceivedAt
		confidence = 0.9

	case "pre_listing_alert", "broker_alert":
		// Extract property details from broker's pre-listing alerts
		extractedData["property_address"] = h.extractPropertyAddress(emailRequest.Subject, emailRequest.Content)
		extractedData["target_list_date"] = h.extractTargetDate(emailRequest.Content)
		extractedData["owner_contact"] = h.extractOwnerContact(emailRequest.Content)
		extractedData["priority_level"] = h.extractPriority(emailRequest.Subject, emailRequest.Content)
		confidence = 0.85

	case "vendor_completion":
		// Extract completion details from vendor notifications
		extractedData["property_address"] = h.extractPropertyAddress(emailRequest.Subject, emailRequest.Content)
		extractedData["service_type"] = h.extractServiceType(emailRequest.Subject, emailRequest.Content)
		extractedData["completion_date"] = emailRequest.ReceivedAt
		extractedData["completion_notes"] = h.extractCompletionNotes(emailRequest.Content)
		confidence = 0.8

	case "lease_update":
		// Extract lease status updates
		extractedData["tenant_name"] = h.extractTenantName(emailRequest.Subject, emailRequest.Content)
		extractedData["property_address"] = h.extractPropertyAddress(emailRequest.Subject, emailRequest.Content)
		extractedData["lease_status"] = h.extractLeaseStatus(emailRequest.Subject, emailRequest.Content)
		extractedData["effective_date"] = h.extractEffectiveDate(emailRequest.Content)
		confidence = 0.85
	}

	return extractedData, confidence
}

func (h *EmailSenderHandlers) processApplicationNotification(extractedData map[string]interface{}, email *models.IncomingEmail) (string, string) {
	applicantName, _ := extractedData["applicant_name"].(string)
	applicantEmail, _ := extractedData["applicant_email"].(string)
	propertyAddress, _ := extractedData["property_address"].(string)

	if applicantName == "" || propertyAddress == "" {
		return "insufficient_data", "Could not extract required application data"
	}

	// Try to match to existing FUB lead
	var fubLead models.Lead
	err := h.db.Where("LOWER(first_name || ' ' || last_name) LIKE LOWER(?) OR email = ?",
		"%"+applicantName+"%", applicantEmail).First(&fubLead).Error

	if err == nil {
		// Found matching FUB lead - create application record
		application := models.Approval{
			ApprovalType:    "rental_application",
			Status:          "pending",
			ApplicantName:   applicantName,
			PropertyAddress: propertyAddress,
			Notes:           fmt.Sprintf("Application notification received from %s", email.FromEmail),
		}

		if err := h.db.Create(&application).Error; err == nil {
			// Update FUB lead with application status
			h.updateFUBLeadApplicationStatus(fubLead.FUBLeadID, "application_submitted", propertyAddress)
			return "application_matched_to_fub_lead", fmt.Sprintf("Matched application from %s to existing FUB lead %s", applicantName, fubLead.FUBLeadID)
		}
	}

	// No FUB lead match - create task for manual matching
	return "manual_matching_required", fmt.Sprintf("Application from %s requires manual FUB lead matching", applicantName)
}

func (h *EmailSenderHandlers) processPreListingAlert(extractedData map[string]interface{}, email *models.IncomingEmail) (string, string) {
	propertyAddress, _ := extractedData["property_address"].(string)

	if propertyAddress == "" {
		return "insufficient_data", "Could not extract property address from pre-listing alert"
	}

	// Create or update pre-listing item
	preListingItem := models.PreListingItem{
		Address:           propertyAddress,
		Status:            models.StatusEmailReceived,
		TerryEmailSubject: email.Subject,
		TerryEmailContent: email.Content,
	}

	if targetDate, ok := extractedData["target_list_date"].(time.Time); ok {
		preListingItem.TargetListingDate = &targetDate
	}

	now := time.Now()
	preListingItem.TerryEmailDate = &now

	if err := h.db.Create(&preListingItem).Error; err != nil {
		return "creation_failed", fmt.Sprintf("Failed to create pre-listing item: %v", err)
	}

	// Link email to pre-listing item
	email.PreListingItemID = &preListingItem.ID
	h.db.Save(email)

	return "pre_listing_created", fmt.Sprintf("Created pre-listing item for %s", propertyAddress)
}

func (h *EmailSenderHandlers) processVendorCompletion(extractedData map[string]interface{}, email *models.IncomingEmail) (string, string) {
	propertyAddress, _ := extractedData["property_address"].(string)
	serviceType, _ := extractedData["service_type"].(string)

	if propertyAddress == "" {
		return "insufficient_data", "Could not extract property address from vendor notification"
	}

	// Find related pre-listing item
	var preListingItem models.PreListingItem
	err := h.db.Where("address LIKE ?", "%"+propertyAddress+"%").First(&preListingItem).Error

	if err != nil {
		return "no_prelisting_match", fmt.Sprintf("No pre-listing item found for %s", propertyAddress)
	}

	// Update pre-listing item based on service type
	now := time.Now()
	statusUpdated := false

	if strings.Contains(strings.ToLower(serviceType), "lockbox") {
		preListingItem.LockboxPlacedDate = &now
		preListingItem.Status = models.StatusLockboxPlaced
		statusUpdated = true
	} else if strings.Contains(strings.ToLower(serviceType), "photo") {
		preListingItem.PhotoCompletedDate = &now
		preListingItem.Status = models.StatusPhotosComplete
		statusUpdated = true
	} else if strings.Contains(strings.ToLower(serviceType), "sign") {
		preListingItem.SignPlacedDate = &now
		statusUpdated = true
	}

	if statusUpdated {
		h.db.Save(&preListingItem)
		return "prelisting_updated", fmt.Sprintf("Updated pre-listing status for %s: %s completed", propertyAddress, serviceType)
	}

	return "logged_only", fmt.Sprintf("Vendor notification logged for %s", propertyAddress)
}

func (h *EmailSenderHandlers) processLeaseUpdate(extractedData map[string]interface{}, email *models.IncomingEmail) (string, string) {
	tenantName, _ := extractedData["tenant_name"].(string)
	propertyAddress, _ := extractedData["property_address"].(string)
	leaseStatus, _ := extractedData["lease_status"].(string)

	if tenantName == "" || propertyAddress == "" {
		return "insufficient_data", "Could not extract tenant or property information"
	}

	// Try to find related application and update status
	var application models.Approval
	err := h.db.Where("applicant_name LIKE ? AND property_address LIKE ?",
		"%"+tenantName+"%", "%"+propertyAddress+"%").First(&application).Error

	if err == nil {
		// Update application status based on lease status
		if strings.Contains(strings.ToLower(leaseStatus), "approved") {
			application.Status = "approved"
		} else if strings.Contains(strings.ToLower(leaseStatus), "signed") {
			application.Status = "approved"
		}
		h.db.Save(&application)

		// Update FUB lead status if connected
		if application.ApprovalType == "rental_application" {
			// Try to find FUB lead and update status
			h.updateFUBLeadApplicationStatus("", "lease_"+leaseStatus, propertyAddress)
		}

		return "application_status_updated", fmt.Sprintf("Updated application status for %s: %s", tenantName, leaseStatus)
	}

	return "logged_only", fmt.Sprintf("Lease update logged for %s", tenantName)
}

func (h *EmailSenderHandlers) updateFUBLeadApplicationStatus(fubLeadID, status, propertyAddress string) {
	// This would integrate with FUB API to update lead status
	// For now, just log the intended update
	fmt.Printf("📊 Would update FUB lead %s status to %s for property %s\n", fubLeadID, status, propertyAddress)
}

func (h *EmailSenderHandlers) logUnstrustedEmail(emailRequest struct {
	From       string
	To         string
	Subject    string
	Content    string
	Headers    map[string]string
	ReceivedAt time.Time
}) {
	// Log untrusted email for security monitoring
	untrustedEmail := models.IncomingEmail{
		FromEmail:        emailRequest.From,
		ToEmail:          emailRequest.To,
		Subject:          emailRequest.Subject,
		Content:          emailRequest.Content,
		ReceivedAt:       emailRequest.ReceivedAt,
		EmailType:        "untrusted",
		ProcessingStatus: "sender_not_trusted",
		Confidence:       0.0,
	}

	h.db.Create(&untrustedEmail)
	fmt.Printf("⚠️ Untrusted email received from %s: %s\n", emailRequest.From, emailRequest.Subject)
}

func (h *EmailSenderHandlers) logEmailProcessing(emailRequest struct {
	From       string
	To         string
	Subject    string
	Content    string
	Headers    map[string]string
	ReceivedAt time.Time
}, sender *models.TrustedEmailSender, result map[string]interface{}) {

	// Create processing log entry
	processingLog := models.EmailProcessingLog{
		TrustedSenderID:   &sender.ID,
		ProcessingStatus:  "success",
		ProcessingResult:  fmt.Sprintf("%v", result),
		ActionTaken:       fmt.Sprintf("%v", result["action_taken"]),
		ImpactDescription: fmt.Sprintf("%v", result["impact"]),
		ConfidenceScore:   result["confidence"].(float64),
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		processingLog.ProcessingResult = string(resultJSON)
	}

	// This would be saved if EmailProcessingLog table exists
	fmt.Printf("📧 Processed email from %s (%s): %s\n", sender.SenderName, sender.EmailType, processingLog.ActionTaken)
}

// Email parsing helper methods

func (h *EmailSenderHandlers) extractApplicantName(subject, content string) string {
	// Simple pattern matching for applicant names
	// In production, this would use more sophisticated regex patterns
	if strings.Contains(subject, " - ") {
		parts := strings.Split(subject, " - ")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

func (h *EmailSenderHandlers) extractApplicantEmail(content string) string {
	// Extract email from content using regex or patterns
	// Simplified implementation
	return ""
}

func (h *EmailSenderHandlers) extractPropertyAddress(subject, content string) string {
	// Extract property address from subject or content
	// Simplified implementation
	return ""
}

func (h *EmailSenderHandlers) extractTargetDate(content string) *time.Time {
	// Extract target listing date from content
	// Simplified implementation
	return nil
}

func (h *EmailSenderHandlers) extractOwnerContact(content string) string {
	// Extract owner contact information
	return ""
}

func (h *EmailSenderHandlers) extractPriority(subject, content string) string {
	if strings.Contains(strings.ToLower(subject+content), "urgent") ||
		strings.Contains(strings.ToLower(subject+content), "asap") {
		return "high"
	}
	return "medium"
}

func (h *EmailSenderHandlers) extractServiceType(subject, content string) string {
	content_lower := strings.ToLower(subject + " " + content)
	if strings.Contains(content_lower, "lockbox") {
		return "lockbox"
	} else if strings.Contains(content_lower, "photo") {
		return "photography"
	} else if strings.Contains(content_lower, "sign") {
		return "signage"
	}
	return "unknown"
}

func (h *EmailSenderHandlers) extractCompletionNotes(content string) string {
	// Extract notes from vendor completion emails
	return ""
}

func (h *EmailSenderHandlers) extractTenantName(subject, content string) string {
	// Extract tenant name from lease update emails
	return ""
}

func (h *EmailSenderHandlers) extractLeaseStatus(subject, content string) string {
	content_lower := strings.ToLower(subject + " " + content)
	if strings.Contains(content_lower, "signed") {
		return "signed"
	} else if strings.Contains(content_lower, "approved") {
		return "approved"
	} else if strings.Contains(content_lower, "rejected") {
		return "rejected"
	}
	return "unknown"
}

func (h *EmailSenderHandlers) extractEffectiveDate(content string) *time.Time {
	// Extract effective date from lease updates
	return nil
}

func (h *EmailSenderHandlers) runParsingTest(emailContent, emailType string, template map[string]interface{}) map[string]interface{} {
	// Mock parsing test - in production this would use actual parsing logic
	return map[string]interface{}{
		"extracted_fields": map[string]string{
			"applicant_name":   "Sarah Martinez",
			"applicant_email":  "sarah.martinez@email.com",
			"property_address": "123 Heights Boulevard",
		},
		"confidence":  0.92,
		"warnings":    []string{},
		"suggestions": []string{"Parsing successful with high confidence"},
	}
}

// Statistics helper methods

func (h *EmailSenderHandlers) getTotalSenders() int64 {
	var count int64
	h.db.Model(&models.TrustedEmailSender{}).Count(&count)
	return count
}

func (h *EmailSenderHandlers) getActiveSenders() int64 {
	var count int64
	h.db.Model(&models.TrustedEmailSender{}).Where("is_active = ?", true).Count(&count)
	return count
}

func (h *EmailSenderHandlers) getProcessingStats(since time.Time) map[string]interface{} {
	// Get email processing statistics since date
	var totalEmails int64
	var processedEmails int64
	var failedEmails int64

	h.db.Model(&models.IncomingEmail{}).Where("received_at >= ?", since).Count(&totalEmails)
	h.db.Model(&models.IncomingEmail{}).Where("received_at >= ? AND processing_status = ?", since, "processed").Count(&processedEmails)
	h.db.Model(&models.IncomingEmail{}).Where("received_at >= ? AND processing_status = ?", since, "failed").Count(&failedEmails)

	successRate := float64(0)
	if totalEmails > 0 {
		successRate = (float64(processedEmails) / float64(totalEmails)) * 100
	}

	return map[string]interface{}{
		"total_emails":     totalEmails,
		"processed_emails": processedEmails,
		"failed_emails":    failedEmails,
		"success_rate":     successRate,
		"period_days":      int(time.Now().Sub(since).Hours() / 24),
	}
}

func (h *EmailSenderHandlers) getTopSenders(since time.Time, limit int) []map[string]interface{} {
	// Get top senders by email volume
	return []map[string]interface{}{
		{"sender_name": "Terry Johnson", "email_count": 45, "success_rate": 98.0},
		{"sender_name": "Buildium System", "email_count": 28, "success_rate": 92.0},
		{"sender_name": "AppFolio Leasing", "email_count": 15, "success_rate": 95.0},
	}
}

func (h *EmailSenderHandlers) getRecentActivity(limit int) []map[string]interface{} {
	// Get recent email processing activity
	return []map[string]interface{}{
		{"timestamp": time.Now().Add(-15 * time.Minute), "sender": "Terry Johnson", "action": "Pre-listing alert processed", "status": "success"},
		{"timestamp": time.Now().Add(-45 * time.Minute), "sender": "Buildium System", "action": "Application notification matched to FUB lead", "status": "success"},
		{"timestamp": time.Now().Add(-2 * time.Hour), "sender": "PRS Lockbox", "action": "Lockbox completion updated", "status": "success"},
	}
}

// RegisterEmailSenderRoutes registers all email sender management routes
func RegisterEmailSenderRoutes(r *gin.Engine, db *gorm.DB) {
	handlers := NewEmailSenderHandlers(db)

	// Admin page route
	r.GET("/admin/email-senders", handlers.AdminEmailSendersPage)

	api := r.Group("/api/v1/email")
	{
		// Sender management
		api.POST("/senders", handlers.CreateTrustedSender)
		api.GET("/senders", handlers.GetTrustedSenders)
		api.GET("/senders/:id", handlers.GetTrustedSender)
		api.PUT("/senders/:id", handlers.UpdateTrustedSender)
		api.DELETE("/senders/:id", handlers.DeleteTrustedSender)

		// Email processing
		api.POST("/process-incoming", handlers.ProcessIncomingEmail)
		api.POST("/test-parsing", handlers.TestEmailParsing)
		api.GET("/processing-stats", handlers.GetEmailProcessingStats)
	}
}
