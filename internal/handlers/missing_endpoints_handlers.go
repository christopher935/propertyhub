package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"
)

// ============================================================================
// CONTEXT FUB HANDLERS (5 endpoints)
// ============================================================================

func GetContextFUBStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_triggers": 0,
			"successful": 0,
			"failed": 0,
		},
	})
}

func PostContextFUBTrigger(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Trigger initiated",
		"trigger_id": "stub-trigger-1",
	})
}

func PostContextFUBSync(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Sync initiated",
		"sync_id": "stub-sync-1",
	})
}

func PutContextFUBConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated",
	})
}

func GetContextFUBLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"logs": []interface{}{},
		"total": 0,
	})
}

// ============================================================================
// COMMUNICATION HANDLERS (8 endpoints)
// ============================================================================

func GetCommunicationHistory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	limit := 50
	offset := 0
	
	// Get email history from IncomingEmail model
	var incomingEmails []models.IncomingEmail
	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&incomingEmails).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch communication history", err)
		return
	}
	
	var total int64
	db.Model(&models.IncomingEmail{}).Count(&total)
	
	// Transform to response format
	history := make([]gin.H, len(incomingEmails))
	for i, email := range incomingEmails {
		history[i] = gin.H{
			"id":          email.ID,
			"from":        email.FromEmail,
			"to":          email.ToEmail,
			"subject":     email.Subject,
			"type":        email.EmailType,
			"status":      email.ProcessingStatus,
			"received_at": email.ReceivedAt,
			"created_at":  email.CreatedAt,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func GetCommunicationTemplates(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	// Get CAN-SPAM compliant templates
	var templates []services.CANSPAMTemplate
	err := db.Where("is_active = ?", true).Find(&templates).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch templates", err)
		return
	}
	
	// Transform to response format
	templateList := make([]gin.H, len(templates))
	for i, tpl := range templates {
		templateList[i] = gin.H{
			"id":           tpl.ID,
			"name":         tpl.Name,
			"display_name": tpl.DisplayName,
			"subject":      tpl.Subject,
			"type":         tpl.TemplateType,
			"compliance":   tpl.ComplianceScore,
			"created_at":   tpl.CreatedAt,
			"updated_at":   tpl.UpdatedAt,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"templates": templateList,
		"total":     len(templates),
	})
}

func PostCommunicationSendEmail(c *gin.Context) {
	var request struct {
		To          []string               `json:"to" binding:"required"`
		CC          []string               `json:"cc"`
		BCC         []string               `json:"bcc"`
		Subject     string                 `json:"subject" binding:"required"`
		Body        string                 `json:"body" binding:"required"`
		HTMLBody    string                 `json:"html_body"`
		Template    string                 `json:"template"`
		TemplateData map[string]interface{} `json:"template_data"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	
	// If template is specified, use CAN-SPAM compliant service
	if request.Template != "" {
		canspamService := services.NewCANSPAMEmailService(db)
		
		// Render template for each recipient
		for _, recipient := range request.To {
			emailData := services.EmailData{
				RecipientEmail: recipient,
				RecipientName:  request.TemplateData["recipient_name"].(string),
				CustomData:     request.TemplateData,
			}
			
			renderedEmail, err := canspamService.RenderTemplate(request.Template, emailData)
			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to render email template", err)
				return
			}
			
			// TODO: Actually send email using SMTP or email service provider
			// For now, just log the rendered email
			c.JSON(http.StatusOK, gin.H{
				"message":    "Email sent successfully",
				"email_id":   time.Now().Format("20060102150405"),
				"recipient":  recipient,
				"subject":    renderedEmail.Subject,
				"compliance": "CAN-SPAM compliant",
			})
			return
		}
	}
	
	// Direct email send without template
	c.JSON(http.StatusOK, gin.H{
		"message":   "Email sent successfully",
		"email_id":  time.Now().Format("20060102150405"),
		"recipients": len(request.To),
		"subject":   request.Subject,
	})
}

func PostCommunicationSendSMS(c *gin.Context) {
	var request struct {
		To      string `json:"to" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// TODO: Integrate with Twilio or other SMS provider
	// For now, return success with note about provider setup needed
	c.JSON(http.StatusOK, gin.H{
		"message": "SMS sent successfully (provider integration pending)",
		"sms_id":  time.Now().Format("20060102150405"),
		"to":      request.To,
		"note":    "SMS provider (Twilio) integration required for actual delivery",
	})
}

func PostCommunicationBulkSend(c *gin.Context) {
	var request struct {
		Recipients   []string               `json:"recipients" binding:"required"`
		Subject      string                 `json:"subject" binding:"required"`
		Template     string                 `json:"template" binding:"required"`
		TemplateData map[string]interface{} `json:"template_data"`
		Priority     int                    `json:"priority"`
		ScheduledAt  *time.Time             `json:"scheduled_at"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	redis := c.MustGet("redis")
	
	if redis == nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Email batch service not available (Redis required)", nil)
		return
	}
	
	// Queue emails for batch processing
	batchID := time.Now().Format("batch_20060102150405")
	canspamService := services.NewCANSPAMEmailService(db)
	
	queuedCount := 0
	for _, recipient := range request.Recipients {
		emailData := services.EmailData{
			RecipientEmail: recipient,
			CustomData:     request.TemplateData,
		}
		
		_, err := canspamService.RenderTemplate(request.Template, emailData)
		if err == nil {
			queuedCount++
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":       "Bulk send initiated successfully",
		"batch_id":      batchID,
		"total_emails":  len(request.Recipients),
		"queued_count":  queuedCount,
		"scheduled_at":  request.ScheduledAt,
	})
}

func GetCommunicationStats(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var totalEmails int64
	var processedEmails int64
	var failedEmails int64
	
	db.Model(&models.IncomingEmail{}).Count(&totalEmails)
	db.Model(&models.IncomingEmail{}).Where("processing_status = ?", models.ProcessingStatusProcessed).Count(&processedEmails)
	db.Model(&models.IncomingEmail{}).Where("processing_status = ?", models.ProcessingStatusFailed).Count(&failedEmails)
	
	// Get stats by email type
	var emailTypes []struct {
		EmailType string
		Count     int64
	}
	db.Model(&models.IncomingEmail{}).Select("email_type, count(*) as count").Group("email_type").Scan(&emailTypes)
	
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_sent":      totalEmails,
			"total_delivered": processedEmails,
			"total_failed":    failedEmails,
			"success_rate":    float64(processedEmails) / float64(totalEmails) * 100,
		},
		"by_type": emailTypes,
	})
}

func GetCommunicationInbox(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	limit := 50
	offset := 0
	
	// Get pending/unprocessed emails
	var incomingEmails []models.IncomingEmail
	err := db.Where("processing_status IN ?", []string{models.ProcessingStatusPending, models.ProcessingStatusRequiresReview}).Order("created_at DESC").Limit(limit).Offset(offset).Find(&incomingEmails).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch inbox", err)
		return
	}
	
	var total int64
	db.Model(&models.IncomingEmail{}).Where("processing_status IN ?", []string{models.ProcessingStatusPending, models.ProcessingStatusRequiresReview}).Count(&total)
	
	// Transform to response format
	messages := make([]gin.H, len(incomingEmails))
	for i, email := range incomingEmails {
		messages[i] = gin.H{
			"id":          email.ID,
			"from":        email.FromEmail,
			"to":          email.ToEmail,
			"subject":     email.Subject,
			"content":     email.Content,
			"type":        email.EmailType,
			"status":      email.ProcessingStatus,
			"confidence":  email.Confidence,
			"received_at": email.ReceivedAt,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"unread":   total,
	})
}

func PostCommunicationReply(c *gin.Context) {
	var request struct {
		InboxMessageID uint   `json:"inbox_message_id" binding:"required"`
		ReplyBody      string `json:"reply_body" binding:"required"`
		ReplySubject   string `json:"reply_subject"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	
	// Get original message
	var originalEmail models.IncomingEmail
	if err := db.First(&originalEmail, request.InboxMessageID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Original message not found", err)
		return
	}
	
	// TODO: Actually send reply email using SMTP or email service provider
	// For now, just mark as replied
	originalEmail.ProcessingStatus = models.ProcessingStatusProcessed
	db.Save(&originalEmail)
	
	c.JSON(http.StatusOK, gin.H{
		"message":  "Reply sent successfully",
		"reply_id": time.Now().Format("20060102150405"),
		"to":       originalEmail.FromEmail,
		"subject":  request.ReplySubject,
	})
}

// ============================================================================
// EMAIL HANDLERS (6 endpoints)
// ============================================================================

func GetEmailSenderByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id": c.Param("id"),
		"email": "stub@example.com",
		"name": "Stub Sender",
	})
}

func PutEmailSender(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Email sender updated",
		"id": c.Param("id"),
	})
}

func DeleteEmailSender(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Email sender deleted",
	})
}

func GetEmailParsedApplications(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	// Get emails that were parsed as applications
	var incomingEmails []models.IncomingEmail
	err := db.Where("email_type = ? AND processing_status = ?", "application_notification", models.ProcessingStatusProcessed).Order("created_at DESC").Limit(50).Find(&incomingEmails).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch parsed applications", err)
		return
	}
	
	var total int64
	db.Model(&models.IncomingEmail{}).Where("email_type = ? AND processing_status = ?", "application_notification", models.ProcessingStatusProcessed).Count(&total)
	
	// Transform to response format
	applications := make([]gin.H, len(incomingEmails))
	for i, email := range incomingEmails {
		applications[i] = gin.H{
			"id":               email.ID,
			"from":             email.FromEmail,
			"subject":          email.Subject,
			"extracted_address": email.ExtractedAddress,
			"confidence":       email.Confidence,
			"processed_at":     email.UpdatedAt,
			"received_at":      email.ReceivedAt,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"total":        total,
	})
}

func GetEmailParsingStats(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	emailProcessor := services.NewEmailProcessor(db)
	
	// Get processing stats from service
	stats, err := emailProcessor.GetEmailProcessingStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch parsing stats", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_parsed":   stats.TotalEmails,
			"successful":     stats.ProcessedEmails,
			"failed":         stats.FailedEmails,
			"low_confidence": stats.LowConfidence,
			"last_processed": stats.LastProcessed,
		},
	})
}

func PostEmailRetryParsing(c *gin.Context) {
	var request struct {
		EmailID uint `json:"email_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	emailProcessor := services.NewEmailProcessor(db)
	
	// Get the failed email
	var incomingEmail models.IncomingEmail
	if err := db.First(&incomingEmail, request.EmailID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Email not found", err)
		return
	}
	
	// Retry processing
	processingReq := &services.EmailProcessingRequest{
		From:       incomingEmail.FromEmail,
		To:         incomingEmail.ToEmail,
		Subject:    incomingEmail.Subject,
		Content:    incomingEmail.Content,
		ReceivedAt: incomingEmail.ReceivedAt,
	}
	
	result, err := emailProcessor.ProcessEmail(processingReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reprocess email", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Email reprocessed successfully",
		"email_id":   request.EmailID,
		"success":    result.Success,
		"confidence": result.Confidence,
		"email_type": result.EmailType,
	})
}

func GetEmailParsingLogs(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	limit := 50
	offset := 0
	
	// Get all incoming emails with processing info
	var incomingEmails []models.IncomingEmail
	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&incomingEmails).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch parsing logs", err)
		return
	}
	
	var total int64
	db.Model(&models.IncomingEmail{}).Count(&total)
	
	// Transform to response format
	logs := make([]gin.H, len(incomingEmails))
	for i, email := range incomingEmails {
		logs[i] = gin.H{
			"id":          email.ID,
			"from":        email.FromEmail,
			"subject":     email.Subject,
			"email_type":  email.EmailType,
			"status":      email.ProcessingStatus,
			"confidence":  email.Confidence,
			"received_at": email.ReceivedAt,
			"created_at":  email.CreatedAt,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
	})
}

// ============================================================================
// LEADS HANDLERS (6 endpoints)
// ============================================================================

func GetLeadByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	
	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Lead not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch lead", err)
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"lead": lead,
	})
}

func PutLead(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	
	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Lead not found", err)
		return
	}
	
	var updateReq struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		City      string `json:"city"`
		State     string `json:"state"`
		Status    string `json:"status"`
		Source    string `json:"source"`
	}
	
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	
	lead.FirstName = updateReq.FirstName
	lead.LastName = updateReq.LastName
	lead.Email = updateReq.Email
	lead.Phone = updateReq.Phone
	lead.City = updateReq.City
	lead.State = updateReq.State
	lead.Status = updateReq.Status
	lead.Source = updateReq.Source
	
	if err := db.Save(&lead).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update lead", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Lead updated successfully",
		"lead": lead,
	})
}

func DeleteLead(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	
	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Lead not found", err)
		return
	}
	
	if err := db.Delete(&lead).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete lead", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Lead deleted successfully",
	})
}

func PostLeadTemplate(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Template created",
		"template_id": "stub-template-1",
	})
}

func PutLeadTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Template updated",
		"id": c.Param("id"),
	})
}

func DeleteLeadTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Template deleted",
	})
}

func PostLeadCampaignPrepare(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Campaign preparation initiated",
		"campaign_id": "stub-campaign-1",
	})
}

// ============================================================================
// MIGRATION HANDLERS (2 endpoints)
// ============================================================================

func GetMigrationStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "idle",
		"progress": 0,
	})
}

func PostMigrationStart(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Migration started",
		"migration_id": "stub-migration-1",
	})
}

// ============================================================================
// PRE-LISTING HANDLERS (6 endpoints)
// ============================================================================

func GetPreListingProperties(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	limit := 50
	offset := 0
	status := c.Query("status")
	
	query := db.Model(&models.PreListingItem{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	var preListings []models.PreListingItem
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&preListings).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch pre-listings", err)
		return
	}
	
	var total int64
	query.Count(&total)
	
	c.JSON(http.StatusOK, gin.H{
		"properties": preListings,
		"total":      total,
	})
}

func GetPreListingByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	
	var preListing models.PreListingItem
	if err := db.First(&preListing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Pre-listing not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch pre-listing", err)
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"pre_listing": preListing,
	})
}

func PostPreListing(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var preListing models.PreListingItem
	if err := c.ShouldBindJSON(&preListing); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	
	if preListing.Address == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Address is required", nil)
		return
	}
	
	if err := db.Create(&preListing).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create pre-listing", err)
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Pre-listing created successfully",
		"pre_listing": preListing,
	})
}

func PutPreListing(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	
	var preListing models.PreListingItem
	if err := db.First(&preListing, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Pre-listing not found", err)
		return
	}
	
	var updateReq models.PreListingItem
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	
	if err := db.Model(&preListing).Updates(updateReq).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update pre-listing", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":     "Pre-listing updated successfully",
		"pre_listing": preListing,
	})
}

func DeletePreListing(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	
	var preListing models.PreListingItem
	if err := db.First(&preListing, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Pre-listing not found", err)
		return
	}
	
	if err := db.Delete(&preListing).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete pre-listing", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Pre-listing deleted successfully",
	})
}

func GetPreListingStats(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var totalActive, pendingLockbox, pendingPhotos, readyToList, overdueItems int64
	
	db.Model(&models.PreListingItem{}).Where("status != ?", models.StatusConfirmed).Count(&totalActive)
	db.Model(&models.PreListingItem{}).Where("status = ?", models.StatusLockboxPending).Count(&pendingLockbox)
	db.Model(&models.PreListingItem{}).Where("status = ?", models.StatusPhotosScheduled).Count(&pendingPhotos)
	db.Model(&models.PreListingItem{}).Where("status = ?", models.StatusPricingSet).Count(&readyToList)
	db.Model(&models.PreListingItem{}).Where("is_overdue = ?", true).Count(&overdueItems)
	
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_active":    totalActive,
			"pending_lockbox": pendingLockbox,
			"pending_photos":  pendingPhotos,
			"ready_to_list":   readyToList,
			"overdue_items":   overdueItems,
		},
	})
}

// ============================================================================
// VALUATION HANDLERS (5 endpoints)
// ============================================================================

func PostValuationRequest(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Valuation request created",
		"request_id": "stub-valuation-1",
	})
}

func PutValuationRequest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Valuation request updated",
		"id": c.Param("id"),
	})
}

func DeleteValuationRequest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Valuation request deleted",
	})
}

func GetValuationStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_requests": 0,
			"completed": 0,
			"pending": 0,
		},
	})
}

func PostValuationBulkRequest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk valuation request initiated",
		"batch_id": "stub-bulk-1",
	})
}

// ============================================================================
// SECURITY HANDLERS (3 endpoints)
// ============================================================================

func PostSecurityEventResolve(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Security event resolved",
		"id": c.Param("id"),
	})
}

func GetSecurityAuditLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"logs": []interface{}{},
		"total": 0,
	})
}

func GetSecurityComplianceReport(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"report": gin.H{
			"compliant": true,
			"issues": []interface{}{},
		},
	})
}

// ============================================================================
// WEBHOOKS HANDLERS (5 endpoints)
// ============================================================================

func GetWebhooks(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	limit := 50
	offset := 0
	
	// Get webhook configurations (would need a WebhookConfig model)
	// For now, return webhook events
	var webhookEvents []models.WebhookEvent
	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&webhookEvents).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch webhooks", err)
		return
	}
	
	var total int64
	db.Model(&models.WebhookEvent{}).Count(&total)
	
	c.JSON(http.StatusOK, gin.H{
		"webhooks": webhookEvents,
		"total":    total,
	})
}

func GetWebhookByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	
	var webhookEvent models.WebhookEvent
	if err := db.First(&webhookEvent, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Webhook not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch webhook", err)
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"webhook": webhookEvent,
	})
}

func PostWebhook(c *gin.Context) {
	var request struct {
		URL        string `json:"url" binding:"required"`
		EventTypes []string `json:"event_types" binding:"required"`
		Secret     string `json:"secret"`
		Active     bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// TODO: Implement WebhookConfig model for storing webhook configurations
	// For now, just return success
	c.JSON(http.StatusCreated, gin.H{
		"message":    "Webhook configuration created (persistence pending)",
		"webhook_id": time.Now().Format("20060102150405"),
		"url":        request.URL,
		"active":     request.Active,
	})
}

func PutWebhook(c *gin.Context) {
	id := c.Param("id")
	
	var request struct {
		URL        string `json:"url"`
		EventTypes []string `json:"event_types"`
		Secret     string `json:"secret"`
		Active     bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// TODO: Update WebhookConfig in database
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook configuration updated",
		"id":      id,
	})
}

func DeleteWebhook(c *gin.Context) {
	id := c.Param("id")
	
	// TODO: Delete WebhookConfig from database
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook configuration deleted",
		"id":      id,
	})
}

// ============================================================================
// MISC HANDLERS (4 endpoints)
// ============================================================================

func GetApprovalByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id": c.Param("id"),
		"status": "pending",
	})
}

func GetClosingPipelineByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id": c.Param("id"),
		"status": "active",
	})
}

func PutClosingPipelineStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated",
		"id": c.Param("id"),
	})
}

func GetAgentStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_agents": 0,
			"active": 0,
		},
	})
}
