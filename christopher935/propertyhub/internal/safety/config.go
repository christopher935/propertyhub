package safety

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// SafetyMode represents different levels of automation safety
type SafetyMode int

const (
	SafetyModeStrict   SafetyMode = iota // Maximum protection - all communications require approval
	SafetyModeModerate                   // Balanced protection - auto-approve low-risk, review medium/high
	SafetyModeRelaxed                    // Minimal protection - auto-approve most communications
	SafetyModeOff                        // No protection - all automation allowed (use with extreme caution)
)

// SafetyConfig contains all configurable safety settings
type SafetyConfig struct {
	// Core Settings
	Mode         SafetyMode `json:"mode"`
	Enabled      bool       `json:"enabled"`
	LastModified time.Time  `json:"last_modified"`
	ModifiedBy   string     `json:"modified_by"`

	// Approval Thresholds
	AutoApprovalThresholds AutoApprovalThresholds `json:"auto_approval_thresholds"`

	// Lead Protection Settings
	LeadProtection LeadProtectionSettings `json:"lead_protection"`

	// Communication Limits
	CommunicationLimits CommunicationLimits `json:"communication_limits"`

	// Emergency Controls
	EmergencyControls EmergencyControls `json:"emergency_controls"`

	// Override Settings
	OverrideSettings OverrideSettings `json:"override_settings"`
}

// AutoApprovalThresholds defines when communications can be auto-approved
type AutoApprovalThresholds struct {
	MaxRecipients               int    `json:"max_recipients"`                 // Max recipients for auto-approval
	MinSafetyScore              int    `json:"min_safety_score"`               // Minimum safety score required
	MaxRiskLevel                string `json:"max_risk_level"`                 // Maximum risk level for auto-approval
	RequireAllSafeRecipients    bool   `json:"require_all_safe_recipients"`    // All recipients must be "Safe" level
	BlockRecentlyContacted      bool   `json:"block_recently_contacted"`       // Block if contacted within threshold
	RecentContactThresholdHours int    `json:"recent_contact_threshold_hours"` // Hours since last contact
}

// LeadProtectionSettings defines how leads are protected
type LeadProtectionSettings struct {
	ProtectOldLeads        bool `json:"protect_old_leads"`        // Protect leads older than threshold
	OldLeadThresholdDays   int  `json:"old_lead_threshold_days"`  // Days to consider a lead "old"
	ProtectInactiveLeads   bool `json:"protect_inactive_leads"`   // Protect leads with no recent activity
	InactiveThresholdDays  int  `json:"inactive_threshold_days"`  // Days to consider a lead "inactive"
	RespectDoNotContact    bool `json:"respect_do_not_contact"`   // Always respect "Do Not Contact" status
	ProtectExistingTenants bool `json:"protect_existing_tenants"` // Protect existing tenants
	RequireOptIn           bool `json:"require_opt_in"`           // Require explicit opt-in for automation
}

// CommunicationLimits defines rate limits and frequency controls
type CommunicationLimits struct {
	MaxDailyEmails          int  `json:"max_daily_emails"`           // Max emails per day per lead
	MaxDailySMS             int  `json:"max_daily_sms"`              // Max SMS per day per lead
	MaxWeeklyContacts       int  `json:"max_weekly_contacts"`        // Max total contacts per week per lead
	MinHoursBetweenContacts int  `json:"min_hours_between_contacts"` // Minimum hours between any contacts
	RespectQuietHours       bool `json:"respect_quiet_hours"`        // Don't send during quiet hours
	QuietHoursStart         int  `json:"quiet_hours_start"`          // Quiet hours start (24-hour format)
	QuietHoursEnd           int  `json:"quiet_hours_end"`            // Quiet hours end (24-hour format)
}

// EmergencyControls defines emergency stop and safety mechanisms
type EmergencyControls struct {
	EnableEmergencyStop   bool    `json:"enable_emergency_stop"`     // Enable emergency stop functionality
	AutoStopOnHighFailure bool    `json:"auto_stop_on_high_failure"` // Auto-stop if failure rate exceeds threshold
	FailureRateThreshold  float64 `json:"failure_rate_threshold"`    // Failure rate that triggers auto-stop
	EnableComplaintStop   bool    `json:"enable_complaint_stop"`     // Auto-stop on complaints
	ComplaintThreshold    int     `json:"complaint_threshold"`       // Number of complaints that trigger stop
}

// OverrideSettings defines who can override safety settings
type OverrideSettings struct {
	AllowAdminOverride      bool     `json:"allow_admin_override"`      // Allow admin to override safety
	AllowTemporaryOverride  bool     `json:"allow_temporary_override"`  // Allow temporary overrides
	OverrideTimeoutHours    int      `json:"override_timeout_hours"`    // Hours before override expires
	RequireOverrideReason   bool     `json:"require_override_reason"`   // Require reason for overrides
	AuthorizedOverrideUsers []string `json:"authorized_override_users"` // Users authorized to override
}

// SafetyConfigManager manages safety configuration
type SafetyConfigManager struct {
	config *SafetyConfig
}

// NewSafetyConfigManager creates a new safety config manager with default settings
func NewSafetyConfigManager() *SafetyConfigManager {
	return &SafetyConfigManager{
		config: getDefaultSafetyConfig(),
	}
}

// getDefaultSafetyConfig returns the default safety configuration (Strict mode)
func getDefaultSafetyConfig() *SafetyConfig {
	return &SafetyConfig{
		Mode:         SafetyModeStrict,
		Enabled:      true,
		LastModified: time.Now(),
		ModifiedBy:   "system",

		AutoApprovalThresholds: AutoApprovalThresholds{
			MaxRecipients:               5,     // Very conservative in strict mode
			MinSafetyScore:              90,    // High safety score required
			MaxRiskLevel:                "Low", // Only low-risk auto-approved
			RequireAllSafeRecipients:    true,  // All must be safe
			BlockRecentlyContacted:      true,  // Block recent contacts
			RecentContactThresholdHours: 72,    // 3 days
		},

		LeadProtection: LeadProtectionSettings{
			ProtectOldLeads:        true,
			OldLeadThresholdDays:   90, // Protect leads older than 90 days
			ProtectInactiveLeads:   true,
			InactiveThresholdDays:  30,   // Protect inactive leads
			RespectDoNotContact:    true, // Always respect DNC
			ProtectExistingTenants: true, // Always protect tenants
			RequireOptIn:           true, // Require opt-in
		},

		CommunicationLimits: CommunicationLimits{
			MaxDailyEmails:          1,  // Very conservative
			MaxDailySMS:             1,  // Very conservative
			MaxWeeklyContacts:       3,  // Very conservative
			MinHoursBetweenContacts: 24, // 24 hours between contacts
			RespectQuietHours:       true,
			QuietHoursStart:         20, // 8 PM
			QuietHoursEnd:           8,  // 8 AM
		},

		EmergencyControls: EmergencyControls{
			EnableEmergencyStop:   true,
			AutoStopOnHighFailure: true,
			FailureRateThreshold:  0.1, // 10% failure rate triggers stop
			EnableComplaintStop:   true,
			ComplaintThreshold:    1, // Single complaint stops automation
		},

		OverrideSettings: OverrideSettings{
			AllowAdminOverride:      true,
			AllowTemporaryOverride:  true,
			OverrideTimeoutHours:    24, // 24-hour override timeout
			RequireOverrideReason:   true,
			AuthorizedOverrideUsers: []string{"admin"},
		},
	}
}

// GetConfig returns the current safety configuration
func (m *SafetyConfigManager) GetConfig() *SafetyConfig {
	return m.config
}

// UpdateMode updates the safety mode and adjusts settings accordingly
func (m *SafetyConfigManager) UpdateMode(mode SafetyMode, modifiedBy string) error {
	log.Printf("Updating safety mode from %s to %s by %s",
		m.getSafeModeString(m.config.Mode),
		m.getSafeModeString(mode),
		modifiedBy)

	m.config.Mode = mode
	m.config.LastModified = time.Now()
	m.config.ModifiedBy = modifiedBy

	// Adjust settings based on mode
	switch mode {
	case SafetyModeStrict:
		m.applyStrictModeSettings()
	case SafetyModeModerate:
		m.applyModerateModeSettings()
	case SafetyModeRelaxed:
		m.applyRelaxedModeSettings()
	case SafetyModeOff:
		m.applyOffModeSettings()
	}

	return nil
}

// applyStrictModeSettings applies strict mode safety settings
func (m *SafetyConfigManager) applyStrictModeSettings() {
	m.config.AutoApprovalThresholds = AutoApprovalThresholds{
		MaxRecipients:               5,
		MinSafetyScore:              90,
		MaxRiskLevel:                "Low",
		RequireAllSafeRecipients:    true,
		BlockRecentlyContacted:      true,
		RecentContactThresholdHours: 72,
	}

	m.config.CommunicationLimits.MaxDailyEmails = 1
	m.config.CommunicationLimits.MaxDailySMS = 1
	m.config.CommunicationLimits.MaxWeeklyContacts = 3
	m.config.CommunicationLimits.MinHoursBetweenContacts = 24

	m.config.EmergencyControls.ComplaintThreshold = 1
	m.config.EmergencyControls.FailureRateThreshold = 0.1
}

// applyModerateModeSettings applies moderate mode safety settings
func (m *SafetyConfigManager) applyModerateModeSettings() {
	m.config.AutoApprovalThresholds = AutoApprovalThresholds{
		MaxRecipients:               25,
		MinSafetyScore:              75,
		MaxRiskLevel:                "Medium",
		RequireAllSafeRecipients:    false,
		BlockRecentlyContacted:      true,
		RecentContactThresholdHours: 48,
	}

	m.config.CommunicationLimits.MaxDailyEmails = 3
	m.config.CommunicationLimits.MaxDailySMS = 2
	m.config.CommunicationLimits.MaxWeeklyContacts = 7
	m.config.CommunicationLimits.MinHoursBetweenContacts = 12

	m.config.EmergencyControls.ComplaintThreshold = 3
	m.config.EmergencyControls.FailureRateThreshold = 0.2
}

// applyRelaxedModeSettings applies relaxed mode safety settings
func (m *SafetyConfigManager) applyRelaxedModeSettings() {
	m.config.AutoApprovalThresholds = AutoApprovalThresholds{
		MaxRecipients:               100,
		MinSafetyScore:              60,
		MaxRiskLevel:                "High",
		RequireAllSafeRecipients:    false,
		BlockRecentlyContacted:      false,
		RecentContactThresholdHours: 24,
	}

	m.config.CommunicationLimits.MaxDailyEmails = 5
	m.config.CommunicationLimits.MaxDailySMS = 3
	m.config.CommunicationLimits.MaxWeeklyContacts = 15
	m.config.CommunicationLimits.MinHoursBetweenContacts = 6

	m.config.EmergencyControls.ComplaintThreshold = 5
	m.config.EmergencyControls.FailureRateThreshold = 0.3
}

// applyOffModeSettings disables most safety protections
func (m *SafetyConfigManager) applyOffModeSettings() {
	log.Printf("WARNING: Safety mode set to OFF - automation protections disabled")

	m.config.AutoApprovalThresholds = AutoApprovalThresholds{
		MaxRecipients:               1000,
		MinSafetyScore:              0,
		MaxRiskLevel:                "Critical",
		RequireAllSafeRecipients:    false,
		BlockRecentlyContacted:      false,
		RecentContactThresholdHours: 0,
	}

	m.config.CommunicationLimits.MaxDailyEmails = 50
	m.config.CommunicationLimits.MaxDailySMS = 20
	m.config.CommunicationLimits.MaxWeeklyContacts = 100
	m.config.CommunicationLimits.MinHoursBetweenContacts = 0

	m.config.EmergencyControls.ComplaintThreshold = 20
	m.config.EmergencyControls.FailureRateThreshold = 0.5

	// Still respect absolute protections
	m.config.LeadProtection.RespectDoNotContact = true
	m.config.LeadProtection.ProtectExistingTenants = true
}

// IsAutomationAllowed checks if automation is allowed based on current settings
func (m *SafetyConfigManager) IsAutomationAllowed(recipientCount int, riskLevel string, safetyScore int) bool {
	if !m.config.Enabled {
		return false
	}

	if m.config.Mode == SafetyModeOff {
		return true // All automation allowed when safety is off
	}

	thresholds := m.config.AutoApprovalThresholds

	// Check recipient count
	if recipientCount > thresholds.MaxRecipients {
		return false
	}

	// Check safety score
	if safetyScore < thresholds.MinSafetyScore {
		return false
	}

	// Check risk level
	if !m.isRiskLevelAllowed(riskLevel, thresholds.MaxRiskLevel) {
		return false
	}

	return true
}

// isRiskLevelAllowed checks if the risk level is within allowed limits
func (m *SafetyConfigManager) isRiskLevelAllowed(riskLevel, maxAllowed string) bool {
	riskLevels := map[string]int{
		"Low":      1,
		"Medium":   2,
		"High":     3,
		"Critical": 4,
	}

	current, exists := riskLevels[riskLevel]
	if !exists {
		return false
	}

	max, exists := riskLevels[maxAllowed]
	if !exists {
		return false
	}

	return current <= max
}

// EnableSafety enables the safety system
func (m *SafetyConfigManager) EnableSafety(modifiedBy string) {
	m.config.Enabled = true
	m.config.LastModified = time.Now()
	m.config.ModifiedBy = modifiedBy
	log.Printf("Safety system enabled by %s", modifiedBy)
}

// DisableSafety disables the safety system (use with extreme caution)
func (m *SafetyConfigManager) DisableSafety(modifiedBy, reason string) {
	m.config.Enabled = false
	m.config.LastModified = time.Now()
	m.config.ModifiedBy = modifiedBy
	log.Printf("WARNING: Safety system DISABLED by %s. Reason: %s", modifiedBy, reason)
}

// GetModeDescription returns a human-readable description of the current mode
func (m *SafetyConfigManager) GetModeDescription() string {
	switch m.config.Mode {
	case SafetyModeStrict:
		return "Maximum protection - All communications require manual approval except very small, low-risk campaigns"
	case SafetyModeModerate:
		return "Balanced protection - Auto-approve low/medium risk campaigns, require approval for high-risk"
	case SafetyModeRelaxed:
		return "Minimal protection - Auto-approve most campaigns, basic safety checks only"
	case SafetyModeOff:
		return "⚠️ NO PROTECTION - All automation allowed (use with extreme caution)"
	default:
		return "Unknown safety mode"
	}
}

// getSafeModeString returns string representation of safety mode
func (m *SafetyConfigManager) getSafeModeString(mode SafetyMode) string {
	switch mode {
	case SafetyModeStrict:
		return "Strict"
	case SafetyModeModerate:
		return "Moderate"
	case SafetyModeRelaxed:
		return "Relaxed"
	case SafetyModeOff:
		return "Off"
	default:
		return "Unknown"
	}
}

// ExportConfig exports the current configuration as JSON
func (m *SafetyConfigManager) ExportConfig() ([]byte, error) {
	return json.MarshalIndent(m.config, "", "  ")
}

// ImportConfig imports configuration from JSON
func (m *SafetyConfigManager) ImportConfig(configJSON []byte, modifiedBy string) error {
	var newConfig SafetyConfig
	if err := json.Unmarshal(configJSON, &newConfig); err != nil {
		return fmt.Errorf("failed to parse configuration: %v", err)
	}

	newConfig.LastModified = time.Now()
	newConfig.ModifiedBy = modifiedBy

	m.config = &newConfig
	log.Printf("Safety configuration imported by %s", modifiedBy)
	return nil
}

// GetSafetyStats returns statistics about current safety settings
func (m *SafetyConfigManager) GetSafetyStats() map[string]interface{} {
	return map[string]interface{}{
		"mode":                   m.getSafeModeString(m.config.Mode),
		"enabled":                m.config.Enabled,
		"max_auto_recipients":    m.config.AutoApprovalThresholds.MaxRecipients,
		"min_safety_score":       m.config.AutoApprovalThresholds.MinSafetyScore,
		"max_risk_level":         m.config.AutoApprovalThresholds.MaxRiskLevel,
		"daily_email_limit":      m.config.CommunicationLimits.MaxDailyEmails,
		"daily_sms_limit":        m.config.CommunicationLimits.MaxDailySMS,
		"emergency_stop_enabled": m.config.EmergencyControls.EnableEmergencyStop,
		"last_modified":          m.config.LastModified,
		"modified_by":            m.config.ModifiedBy,
	}
}
