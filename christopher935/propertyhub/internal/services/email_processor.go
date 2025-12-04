package services

import (
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// EmailProcessor handles email processing for pre-listings
type EmailProcessor struct {
	db *gorm.DB
}

// EmailProcessingRequest represents an email to be processed
type EmailProcessingRequest struct {
	From       string    `json:"from"`
	To         string    `json:"to"`
	Subject    string    `json:"subject"`
	Content    string    `json:"content"`
	ReceivedAt time.Time `json:"received_at"`
}

// EmailProcessingResult represents the result of email processing
type EmailProcessingResult struct {
	Success        bool                   `json:"success"`
	EmailType      string                 `json:"email_type"`
	Confidence     float32                `json:"confidence"`
	ExtractedData  map[string]interface{} `json:"extracted_data"`
	PreListingItem *models.PreListingItem `json:"pre_listing_item,omitempty"`
	IncomingEmail  *models.IncomingEmail  `json:"incoming_email"`
}

// EmailProcessingStats represents email processing statistics
type EmailProcessingStats struct {
	TotalEmails     int64     `json:"total_emails"`
	ProcessedEmails int64     `json:"processed_emails"`
	FailedEmails    int64     `json:"failed_emails"`
	LowConfidence   int64     `json:"low_confidence"`
	LastProcessed   time.Time `json:"last_processed"`
}

// NewEmailProcessor creates a new email processor
func NewEmailProcessor(db *gorm.DB) *EmailProcessor {
	return &EmailProcessor{
		db: db,
	}
}

// ProcessEmail processes an incoming email
func (ep *EmailProcessor) ProcessEmail(req *EmailProcessingRequest) (*EmailProcessingResult, error) {
	// Create incoming email record
	incomingEmail := &models.IncomingEmail{
		FromEmail:        req.From,
		ToEmail:          req.To,
		Subject:          req.Subject,
		Content:          req.Content,
		ReceivedAt:       req.ReceivedAt,
		ProcessingStatus: "processed",
		EmailType:        "pre_listing",
		Confidence:       0.8,
	}

	if err := ep.db.Create(incomingEmail).Error; err != nil {
		return nil, err
	}

	// Mock processing result
	result := &EmailProcessingResult{
		Success:    true,
		EmailType:  "pre_listing",
		Confidence: 0.8,
		ExtractedData: map[string]interface{}{
			"address": "Sample Address",
		},
		IncomingEmail: incomingEmail,
	}

	return result, nil
}

// GetEmailProcessingStats returns processing statistics
func (ep *EmailProcessor) GetEmailProcessingStats() (*EmailProcessingStats, error) {
	var stats EmailProcessingStats

	ep.db.Model(&models.IncomingEmail{}).Count(&stats.TotalEmails)
	ep.db.Model(&models.IncomingEmail{}).Where("processing_status = ?", "processed").Count(&stats.ProcessedEmails)
	ep.db.Model(&models.IncomingEmail{}).Where("processing_status = ?", "failed").Count(&stats.FailedEmails)
	ep.db.Model(&models.IncomingEmail{}).Where("confidence < ?", 0.5).Count(&stats.LowConfidence)

	var lastEmail models.IncomingEmail
	if err := ep.db.Order("received_at DESC").First(&lastEmail).Error; err == nil {
		stats.LastProcessed = lastEmail.ReceivedAt
	}

	return &stats, nil
}
