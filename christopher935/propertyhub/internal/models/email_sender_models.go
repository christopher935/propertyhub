package models

import (
	"fmt"
	"time"
	"gorm.io/gorm"
)

// TrustedEmailSender represents a trusted email sender for automated processing
type TrustedEmailSender struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Sender Information
	SenderEmail    string `json:"sender_email" gorm:"uniqueIndex;not null"`
	SenderName     string `json:"sender_name" gorm:"not null"`
	CompanyName    string `json:"company_name"`
	ContactPerson  string `json:"contact_person"`

	// Email Processing Configuration
	EmailType       string `json:"email_type" gorm:"not null"` // application_notification, pre_listing_alert, lease_update, vendor_completion, terry_alert
	ProcessingMode  string `json:"processing_mode" gorm:"default:'automatic'"` // automatic, manual_review, disabled
	ParsingTemplate string `json:"parsing_template" gorm:"type:text"` // JSON template for extracting data
	
	// Status and Control
	IsActive        bool `json:"is_active" gorm:"default:true"`
	IsVerified      bool `json:"is_verified" gorm:"default:false"`
	LastEmailAt     *time.Time `json:"last_email_at"`
	EmailCount      int  `json:"email_count" gorm:"default:0"`
	
	// Business Context
	BusinessPurpose string `json:"business_purpose" gorm:"type:text"` // Description of what emails from this sender accomplish
	Priority        string `json:"priority" gorm:"default:'medium'"` // high, medium, low (affects processing speed)
	
	// Administrative
	AddedBy       string `json:"added_by"`
	VerifiedBy    string `json:"verified_by"`
	Notes         string `json:"notes" gorm:"type:text"`
	ApprovalDate  *time.Time `json:"approval_date"`
}

// EmailProcessingRule represents parsing rules for specific email types
type EmailProcessingRule struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Rule Information
	TrustedSenderID uint               `json:"trusted_sender_id" gorm:"not null"`
	TrustedSender   TrustedEmailSender `json:"trusted_sender" gorm:"foreignKey:TrustedSenderID"`
	
	RuleName        string `json:"rule_name" gorm:"not null"`
	RuleDescription string `json:"rule_description"`
	EmailType       string `json:"email_type" gorm:"not null"`
	
	// Parsing Configuration
	SubjectPattern  string `json:"subject_pattern"` // Regex or keywords to match subject lines
	BodyPatterns    string `json:"body_patterns" gorm:"type:text"` // JSON array of patterns to extract data
	RequiredFields  string `json:"required_fields"` // JSON array of required extracted fields
	
	// Extraction Mapping
	FieldMappings   string `json:"field_mappings" gorm:"type:text"` // JSON mapping of extracted fields to PropertyHub fields
	
	// Rule Status
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	SuccessCount   int       `json:"success_count" gorm:"default:0"`
	FailureCount   int       `json:"failure_count" gorm:"default:0"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	
	// Administrative
	CreatedBy string `json:"created_by"`
	TestResults string `json:"test_results" gorm:"type:text"` // JSON of test parsing results
}

// EmailProcessingLog represents a log of email processing attempts
type EmailProcessingLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Email Reference
	IncomingEmailID   uint                 `json:"incoming_email_id" gorm:"not null"`
	IncomingEmail     IncomingEmail        `json:"incoming_email" gorm:"foreignKey:IncomingEmailID"`
	TrustedSenderID   *uint                `json:"trusted_sender_id"`
	TrustedSender     *TrustedEmailSender  `json:"trusted_sender" gorm:"foreignKey:TrustedSenderID"`
	
	// Processing Results
	ProcessingStatus  string               `json:"processing_status"` // success, failed, sender_not_trusted, parsing_failed
	ProcessingResult  string               `json:"processing_result" gorm:"type:text"` // JSON result of processing
	ExtractedData     string               `json:"extracted_data" gorm:"type:text"` // JSON of successfully extracted fields
	ErrorMessage      string               `json:"error_message"`
	ProcessingTimeMs  int                  `json:"processing_time_ms"`
	
	// Business Impact
	ActionTaken       string `json:"action_taken"` // pre_listing_created, application_matched, status_updated, manual_review_needed
	ImpactDescription string `json:"impact_description" gorm:"type:text"`
	
	// Quality Metrics
	ConfidenceScore   float64 `json:"confidence_score"` // 0.0-1.0 confidence in parsing accuracy
	RequiresReview    bool    `json:"requires_review" gorm:"default:false"`
	ReviewedBy        string  `json:"reviewed_by"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ReviewNotes       string  `json:"review_notes"`
}

// Validation methods
func (tes *TrustedEmailSender) Validate() error {
	if tes.SenderEmail == "" {
		return fmt.Errorf("sender email is required")
	}
	if tes.SenderName == "" {
		return fmt.Errorf("sender name is required")
	}
	if tes.EmailType == "" {
		return fmt.Errorf("email type is required")
	}
	
	// Validate email type
	validTypes := map[string]bool{
		"application_notification": true,
		"pre_listing_alert": true,
		"lease_update": true,
		"vendor_completion": true,
		"broker_alert": true,
		"maintenance_alert": true,
		"payment_notification": true,
	}
	if !validTypes[tes.EmailType] {
		return fmt.Errorf("invalid email type: %s", tes.EmailType)
	}
	
	return nil
}

// Business logic methods
func (tes *TrustedEmailSender) CanProcessEmail(fromEmail string) bool {
	return tes.IsActive && tes.IsVerified && tes.SenderEmail == fromEmail
}

func (tes *TrustedEmailSender) GetPriority() int {
	switch tes.Priority {
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 2
	}
}

func (tes *TrustedEmailSender) UpdateLastActivity() {
	now := time.Now()
	tes.LastEmailAt = &now
	tes.EmailCount++
}

// Helper methods for business logic
func (tes *TrustedEmailSender) IsPreListingSystem() bool {
	return tes.EmailType == "pre_listing_alert" || tes.EmailType == "broker_alert" || tes.EmailType == "vendor_completion"
}

func (tes *TrustedEmailSender) IsApplicationSystem() bool {
	return tes.EmailType == "application_notification" || tes.EmailType == "lease_update"
}

func (tes *TrustedEmailSender) RequiresImmediateProcessing() bool {
	return tes.Priority == "high" || tes.EmailType == "application_notification"
}