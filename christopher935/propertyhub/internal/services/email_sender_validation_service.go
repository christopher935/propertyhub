package services

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// EmailSenderValidationService validates and manages trusted email senders
type EmailSenderValidationService struct {
	db *gorm.DB
}

// NewEmailSenderValidationService creates a new email sender validation service
func NewEmailSenderValidationService(db *gorm.DB) *EmailSenderValidationService {
	return &EmailSenderValidationService{
		db: db,
	}
}

// ValidateSender validates if an email sender is trusted and active
func (s *EmailSenderValidationService) ValidateSender(fromEmail string) (*models.TrustedEmailSender, error) {
	// Normalize email for comparison
	fromEmail = strings.ToLower(strings.TrimSpace(fromEmail))
	
	var sender models.TrustedEmailSender
	err := s.db.Where("LOWER(sender_email) = ? AND is_active = ? AND is_verified = ?", 
		fromEmail, true, true).First(&sender).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("sender not found or not active: %s", fromEmail)
		}
		return nil, fmt.Errorf("database error validating sender: %v", err)
	}
	
	return &sender, nil
}

// GetTrustedSenderByEmail retrieves a trusted sender by email address
func (s *EmailSenderValidationService) GetTrustedSenderByEmail(email string) (*models.TrustedEmailSender, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	
	var sender models.TrustedEmailSender
	err := s.db.Where("LOWER(sender_email) = ?", email).First(&sender).Error
	
	if err != nil {
		return nil, err
	}
	
	return &sender, nil
}

// GetSendersByType retrieves all trusted senders of a specific email type
func (s *EmailSenderValidationService) GetSendersByType(emailType string) ([]models.TrustedEmailSender, error) {
	var senders []models.TrustedEmailSender
	err := s.db.Where("email_type = ? AND is_active = ?", emailType, true).Find(&senders).Error
	return senders, err
}

// GetActiveSenders retrieves all active trusted senders
func (s *EmailSenderValidationService) GetActiveSenders() ([]models.TrustedEmailSender, error) {
	var senders []models.TrustedEmailSender
	err := s.db.Where("is_active = ? AND is_verified = ?", true, true).Find(&senders).Error
	return senders, err
}

// UpdateSenderLastActivity updates the last activity timestamp for a sender
func (s *EmailSenderValidationService) UpdateSenderLastActivity(senderID uint) error {
	now := time.Now()
	return s.db.Model(&models.TrustedEmailSender{}).
		Where("id = ?", senderID).
		Updates(map[string]interface{}{
			"last_email_at": &now,
			"email_count":   gorm.Expr("email_count + 1"),
		}).Error
}

// ValidateEmailForProcessing validates an email for processing based on sender rules
func (s *EmailSenderValidationService) ValidateEmailForProcessing(fromEmail, subject, content string) (*EmailValidationResult, error) {
	sender, err := s.ValidateSender(fromEmail)
	if err != nil {
		return &EmailValidationResult{
			IsValid:     false,
			Reason:      "Sender not trusted",
			Confidence:  0.0,
			Error:       err,
		}, err
	}
	
	// Validate processing mode
	if sender.ProcessingMode == "disabled" {
		return &EmailValidationResult{
			IsValid:    false,
			Reason:     "Sender processing disabled",
			Confidence: 0.0,
			Sender:     sender,
		}, fmt.Errorf("processing disabled for sender: %s", fromEmail)
	}
	
	// Calculate confidence based on sender history and email content
	confidence := s.calculateProcessingConfidence(sender, subject, content)
	
	return &EmailValidationResult{
		IsValid:    true,
		Reason:     "Sender validated and active",
		Confidence: confidence,
		Sender:     sender,
	}, nil
}

// EmailValidationResult represents the result of email validation
type EmailValidationResult struct {
	IsValid    bool                       `json:"is_valid"`
	Reason     string                     `json:"reason"`
	Confidence float64                    `json:"confidence"`
	Sender     *models.TrustedEmailSender `json:"sender,omitempty"`
	Error      error                      `json:"error,omitempty"`
}

// calculateProcessingConfidence calculates confidence score for email processing
func (s *EmailSenderValidationService) calculateProcessingConfidence(sender *models.TrustedEmailSender, subject, content string) float64 {
	baseConfidence := 0.8
	
	// Boost confidence for high-priority senders
	if sender.Priority == "high" {
		baseConfidence += 0.1
	}
	
	// Boost confidence for senders with good history
	if sender.EmailCount > 10 {
		baseConfidence += 0.05
	}
	
	// Reduce confidence for manual review mode
	if sender.ProcessingMode == "manual_review" {
		baseConfidence -= 0.2
	}
	
	// Analyze content patterns based on email type
	confidence := s.analyzeContentPatterns(sender.EmailType, subject, content, baseConfidence)
	
	// Ensure confidence stays within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}
	
	return confidence
}

// analyzeContentPatterns analyzes email content patterns to adjust confidence
func (s *EmailSenderValidationService) analyzeContentPatterns(emailType, subject, content string, baseConfidence float64) float64 {
	subjectLower := strings.ToLower(subject)
	contentLower := strings.ToLower(content)
	
	switch emailType {
	case "application_notification":
		// Look for application-related keywords
		appKeywords := []string{"application", "applicant", "applied", "rental application"}
		for _, keyword := range appKeywords {
			if strings.Contains(subjectLower, keyword) || strings.Contains(contentLower, keyword) {
				baseConfidence += 0.05
				break
			}
		}
		
	case "pre_listing_alert", "broker_alert":
		// Look for pre-listing keywords
		preListingKeywords := []string{"pre-listing", "new listing", "property ready", "coming available"}
		for _, keyword := range preListingKeywords {
			if strings.Contains(subjectLower, keyword) || strings.Contains(contentLower, keyword) {
				baseConfidence += 0.05
				break
			}
		}
		
	case "vendor_completion":
		// Look for completion keywords
		completionKeywords := []string{"completed", "finished", "done", "installed", "lockbox", "photos"}
		for _, keyword := range completionKeywords {
			if strings.Contains(subjectLower, keyword) || strings.Contains(contentLower, keyword) {
				baseConfidence += 0.05
				break
			}
		}
		
	case "lease_update":
		// Look for lease-related keywords
		leaseKeywords := []string{"lease", "signed", "approved", "tenant", "move-in"}
		for _, keyword := range leaseKeywords {
			if strings.Contains(subjectLower, keyword) || strings.Contains(contentLower, keyword) {
				baseConfidence += 0.05
				break
			}
		}
	}
	
	return baseConfidence
}

// GetSenderStatistics returns processing statistics for a specific sender
func (s *EmailSenderValidationService) GetSenderStatistics(senderID uint) (*SenderStatistics, error) {
	var sender models.TrustedEmailSender
	if err := s.db.First(&sender, senderID).Error; err != nil {
		return nil, err
	}
	
	var stats SenderStatistics
	stats.SenderID = senderID
	stats.SenderName = sender.SenderName
	stats.EmailType = sender.EmailType
	stats.TotalEmails = sender.EmailCount
	
	// Get processing statistics from incoming emails
	var processedCount, failedCount int64
	s.db.Model(&models.IncomingEmail{}).
		Where("from_email = ? AND processing_status = ?", sender.SenderEmail, "processed").
		Count(&processedCount)
	
	s.db.Model(&models.IncomingEmail{}).
		Where("from_email = ? AND processing_status = ?", sender.SenderEmail, "failed").
		Count(&failedCount)
	
	stats.ProcessedEmails = int(processedCount)
	stats.FailedEmails = int(failedCount)
	
	if stats.TotalEmails > 0 {
		stats.SuccessRate = float64(stats.ProcessedEmails) / float64(stats.TotalEmails) * 100
	}
	
	if sender.LastEmailAt != nil {
		stats.LastEmailAt = *sender.LastEmailAt
	}
	
	return &stats, nil
}

// SenderStatistics represents processing statistics for a sender
type SenderStatistics struct {
	SenderID       uint      `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	EmailType      string    `json:"email_type"`
	TotalEmails    int       `json:"total_emails"`
	ProcessedEmails int      `json:"processed_emails"`
	FailedEmails   int       `json:"failed_emails"`
	SuccessRate    float64   `json:"success_rate"`
	LastEmailAt    time.Time `json:"last_email_at"`
}

// ValidateAndRouteDomainEmail validates an email from a domain-based sender
func (s *EmailSenderValidationService) ValidateAndRouteDomainEmail(fromEmail, toDomain string) (*models.TrustedEmailSender, error) {
	// For PropertyHub domain emails, check if we have sender configured for the domain
	domain := extractDomainFromEmail(fromEmail)
	
	var sender models.TrustedEmailSender
	err := s.db.Where("sender_email LIKE ? AND is_active = ?", "%@"+domain, true).First(&sender).Error
	
	if err != nil {
		return nil, fmt.Errorf("no active sender configuration for domain: %s", domain)
	}
	
	return &sender, nil
}

// extractDomainFromEmail extracts the domain part from an email address
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

// GetSendersByPriority returns senders grouped by priority level
func (s *EmailSenderValidationService) GetSendersByPriority(priority string) ([]models.TrustedEmailSender, error) {
	var senders []models.TrustedEmailSender
	err := s.db.Where("priority = ? AND is_active = ?", priority, true).
		Order("created_at DESC").Find(&senders).Error
	return senders, err
}

// GetSendersRequiringReview returns senders that need manual review
func (s *EmailSenderValidationService) GetSendersRequiringReview() ([]models.TrustedEmailSender, error) {
	var senders []models.TrustedEmailSender
	err := s.db.Where("processing_mode = ? OR is_verified = ?", "manual_review", false).
		Find(&senders).Error
	return senders, err
}