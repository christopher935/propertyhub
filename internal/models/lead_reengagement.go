package models

import (
	"chrisgross-ctrl-project/internal/security"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// LeadSegment represents the classification of a lead for re-engagement
type LeadSegment string

const (
	SegmentActive     LeadSegment = "active"     // Active in last 12 months
	SegmentDormant    LeadSegment = "dormant"    // 1-5 years old
	SegmentUnknown    LeadSegment = "unknown"    // Source unclear
	SegmentSuppressed LeadSegment = "suppressed" // High-risk, do not contact
)

// RiskLevel represents the risk assessment for contacting a lead
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"    // Safe to contact
	RiskMedium RiskLevel = "medium" // Proceed with caution
	RiskHigh   RiskLevel = "high"   // Do not contact
)

// ConsentStatus represents the documented consent level
type ConsentStatus string

const (
	ConsentExpress ConsentStatus = "express" // Documented express written consent
	ConsentImplied ConsentStatus = "implied" // Business relationship implied consent
	ConsentUnknown ConsentStatus = "unknown" // Consent status unclear
	ConsentRevoked ConsentStatus = "revoked" // Previously opted out
)

// CampaignStatus represents the current state in re-engagement campaign
type CampaignStatus string

const (
	CampaignPending    CampaignStatus = "pending"    // Ready for campaign
	CampaignActive     CampaignStatus = "active"     // Currently in campaign
	CampaignCompleted  CampaignStatus = "completed"  // Campaign finished
	CampaignSuppressed CampaignStatus = "suppressed" // Removed from campaign
)

// LeadReengagement represents a lead prepared for re-engagement campaign
type LeadReengagement struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Lead Identification
	FUBContactID string                   `json:"fub_contact_id" gorm:"uniqueIndex;not null"` // FUB contact ID
	Email        security.EncryptedString `json:"email" gorm:"index"`
	Phone        security.EncryptedString `json:"phone" gorm:"index"`
	FirstName    security.EncryptedString `json:"first_name"`
	LastName     security.EncryptedString `json:"last_name"`

	// Segmentation Data
	Segment       LeadSegment   `json:"segment" gorm:"index;not null"`
	RiskLevel     RiskLevel     `json:"risk_level" gorm:"index;not null"`
	ConsentStatus ConsentStatus `json:"consent_status" gorm:"index;not null"`

	// Historical Data
	LastActivity   *time.Time `json:"last_activity,omitempty"`
	FirstContact   *time.Time `json:"first_contact,omitempty"`
	SourceCampaign string     `json:"source_campaign"`
	OriginalSource string     `json:"original_source"`

	// Risk Assessment Factors
	HasEmail            bool `json:"has_email"`
	EmailValid          bool `json:"email_valid"`
	HardBounce          bool `json:"hard_bounce"`
	PreviousUnsubscribe bool `json:"previous_unsubscribe"`
	OnDNCList           bool `json:"on_dnc_list"`

	// Campaign Management
	CampaignStatus    CampaignStatus `json:"campaign_status" gorm:"index;default:'pending'"`
	CampaignStarted   *time.Time     `json:"campaign_started,omitempty"`
	CampaignCompleted *time.Time     `json:"campaign_completed,omitempty"`
	EmailsSent        int            `json:"emails_sent" gorm:"default:0"`
	LastEmailSent     *time.Time     `json:"last_email_sent,omitempty"`

	// Engagement Tracking
	EmailsOpened  int        `json:"emails_opened" gorm:"default:0"`
	EmailsClicked int        `json:"emails_clicked" gorm:"default:0"`
	Responded     bool       `json:"responded" gorm:"default:false"`
	ResponseDate  *time.Time `json:"response_date,omitempty"`
	OptedIn       bool       `json:"opted_in" gorm:"default:false"`
	OptInDate     *time.Time `json:"opt_in_date,omitempty"`

	// Compliance Tracking
	ConsentDocumented bool       `json:"consent_documented" gorm:"default:false"`
	ConsentDate       *time.Time `json:"consent_date,omitempty"`
	ConsentMethod     string     `json:"consent_method"` // "website_form", "phone_call", "email_reply", etc.

	// Notes and Tags
	Notes string `json:"notes"`
	Tags  string `json:"tags"` // JSON array of tags
}

// CampaignTemplate represents email templates for re-engagement
type CampaignTemplate struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name        string `json:"name" gorm:"not null"`
	EmailNumber int    `json:"email_number" gorm:"not null"` // 1, 2, or 3
	Subject     string `json:"subject" gorm:"not null"`
	Body        string `json:"body" gorm:"type:text;not null"`
	DaysDelay   int    `json:"days_delay" gorm:"default:0"` // Days after previous email

	// Template Variables
	Variables string `json:"variables"` // JSON array of available variables

	// Performance Tracking
	TimesSent    int     `json:"times_sent" gorm:"default:0"`
	OpenRate     float64 `json:"open_rate" gorm:"default:0"`
	ClickRate    float64 `json:"click_rate" gorm:"default:0"`
	ResponseRate float64 `json:"response_rate" gorm:"default:0"`
}

// CampaignExecution represents the execution log of re-engagement campaigns
type CampaignExecution struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	LeadReengagementID uint             `json:"lead_reengagement_id" gorm:"not null"`
	LeadReengagement   LeadReengagement `json:"lead_reengagement" gorm:"foreignKey:LeadReengagementID"`

	CampaignTemplateID uint             `json:"campaign_template_id" gorm:"not null"`
	CampaignTemplate   CampaignTemplate `json:"campaign_template" gorm:"foreignKey:CampaignTemplateID"`

	// Execution Details
	ScheduledFor time.Time  `json:"scheduled_for"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	Status       string     `json:"status" gorm:"default:'scheduled'"` // scheduled, sent, failed, skipped

	// FUB Integration
	FUBActionPlanID string `json:"fub_action_plan_id"`
	FUBStepID       string `json:"fub_step_id"`

	// Results Tracking
	EmailOpened  bool   `json:"email_opened" gorm:"default:false"`
	EmailClicked bool   `json:"email_clicked" gorm:"default:false"`
	Responded    bool   `json:"responded" gorm:"default:false"`
	ResponseType string `json:"response_type"` // "opt_in", "opt_out", "inquiry", "complaint"

	// Error Handling
	ErrorMessage string     `json:"error_message"`
	RetryCount   int        `json:"retry_count" gorm:"default:0"`
	NextRetry    *time.Time `json:"next_retry,omitempty"`
}

// ReengagementMetrics represents campaign performance metrics
type ReengagementMetrics struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Date Range
	MetricDate time.Time `json:"metric_date" gorm:"uniqueIndex"`

	// Volume Metrics
	TotalLeads      int `json:"total_leads"`
	ActiveLeads     int `json:"active_leads"`
	DormantLeads    int `json:"dormant_leads"`
	SuppressedLeads int `json:"suppressed_leads"`

	// Campaign Metrics
	EmailsSent    int `json:"emails_sent"`
	EmailsOpened  int `json:"emails_opened"`
	EmailsClicked int `json:"emails_clicked"`
	Responses     int `json:"responses"`
	OptIns        int `json:"opt_ins"`
	OptOuts       int `json:"opt_outs"`

	// Calculated Rates
	OpenRate     float64 `json:"open_rate"`
	ClickRate    float64 `json:"click_rate"`
	ResponseRate float64 `json:"response_rate"`
	OptInRate    float64 `json:"opt_in_rate"`

	// Reputation Metrics
	SpamComplaints  int     `json:"spam_complaints"`
	HardBounces     int     `json:"hard_bounces"`
	ReputationScore float64 `json:"reputation_score"`
}

// Validation methods for LeadReengagement
func (lr *LeadReengagement) Validate() error {
	if lr.FUBContactID == "" {
		return fmt.Errorf("FUB contact ID is required")
	}

	if lr.Segment == "" {
		return fmt.Errorf("segment classification is required")
	}

	if lr.RiskLevel == "" {
		return fmt.Errorf("risk level assessment is required")
	}

	if lr.ConsentStatus == "" {
		return fmt.Errorf("consent status is required")
	}

	return nil
}

// IsEligibleForCampaign checks if lead is eligible for re-engagement
func (lr *LeadReengagement) IsEligibleForCampaign() bool {
	// Must have email
	if !lr.HasEmail || !lr.EmailValid {
		return false
	}

	// Must not be high risk
	if lr.RiskLevel == RiskHigh {
		return false
	}

	// Must not be suppressed
	if lr.Segment == SegmentSuppressed {
		return false
	}

	// Must not have hard bounced
	if lr.HardBounce {
		return false
	}

	// Must not have previously unsubscribed
	if lr.PreviousUnsubscribe {
		return false
	}

	// Must not be in active campaign
	if lr.CampaignStatus == CampaignActive {
		return false
	}

	return true
}

// GetNextEmailTemplate determines which email template to send next
func (lr *LeadReengagement) GetNextEmailTemplate() int {
	switch lr.EmailsSent {
	case 0:
		return 1 // Permission Reset email
	case 1:
		return 2 // Value Add email
	case 2:
		return 3 // Last Chance email
	default:
		return 0 // Campaign complete
	}
}

// CalculateSegment determines the appropriate segment for a lead
func (lr *LeadReengagement) CalculateSegment() LeadSegment {
	now := time.Now()

	// Check for suppression conditions first
	if !lr.HasEmail || lr.HardBounce || lr.PreviousUnsubscribe || lr.OnDNCList {
		return SegmentSuppressed
	}

	// Determine segment based on last activity
	if lr.LastActivity != nil {
		daysSinceActivity := int(now.Sub(*lr.LastActivity).Hours() / 24)

		if daysSinceActivity <= 365 {
			return SegmentActive
		} else if daysSinceActivity <= 1825 { // 5 years
			return SegmentDormant
		}
	}

	// If no last activity or very old, check first contact
	if lr.FirstContact != nil {
		daysSinceFirst := int(now.Sub(*lr.FirstContact).Hours() / 24)

		if daysSinceFirst <= 1825 { // 5 years
			return SegmentDormant
		}
	}

	// Default to unknown if we can't determine
	return SegmentUnknown
}

// CalculateRiskLevel assesses the risk of contacting this lead
func (lr *LeadReengagement) CalculateRiskLevel() RiskLevel {
	riskFactors := 0

	// High risk factors
	if !lr.HasEmail {
		riskFactors += 10
	}
	if !lr.EmailValid {
		riskFactors += 8
	}
	if lr.HardBounce {
		riskFactors += 10
	}
	if lr.PreviousUnsubscribe {
		riskFactors += 10
	}
	if lr.OnDNCList {
		riskFactors += 10
	}

	// Medium risk factors
	if lr.ConsentStatus == ConsentUnknown {
		riskFactors += 3
	}
	if lr.LastActivity != nil {
		daysSinceActivity := int(time.Now().Sub(*lr.LastActivity).Hours() / 24)
		if daysSinceActivity > 1095 { // 3 years
			riskFactors += 2
		}
	}

	// Determine risk level
	if riskFactors >= 8 {
		return RiskHigh
	} else if riskFactors >= 3 {
		return RiskMedium
	}

	return RiskLow
}
