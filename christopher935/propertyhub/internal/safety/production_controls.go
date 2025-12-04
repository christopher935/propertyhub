package safety

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// ProductionMode represents the current system mode
type ProductionMode string

const (
	ModeDevelopment ProductionMode = "development"
	ModeDemo        ProductionMode = "demo"
	ModeProduction  ProductionMode = "production"
)

// SafetyControls manages all production safety settings
type SafetyControls struct {
	mu                     sync.RWMutex
	Mode                   ProductionMode `json:"mode"`
	AllowLiveCommunication bool           `json:"allow_live_communication"`
	AllowFUBIntegration    bool           `json:"allow_fub_integration"`
	AllowEmailSending      bool           `json:"allow_email_sending"`
	AllowSMSSending        bool           `json:"allow_sms_sending"`
	AllowLeadCreation      bool           `json:"allow_lead_creation"`
	AllowAutomation        bool           `json:"allow_automation"`
	TestModePrefix         string         `json:"test_mode_prefix"`
	MaxTestLeads           int            `json:"max_test_leads"`
	LastUpdated            time.Time      `json:"last_updated"`
	UpdatedBy              string         `json:"updated_by"`
}

var (
	globalSafetyControls *SafetyControls
	once                 sync.Once
)

// GetSafetyControls returns the global safety controls instance
func GetSafetyControls() *SafetyControls {
	once.Do(func() {
		globalSafetyControls = &SafetyControls{
			Mode:                   ModeDevelopment,
			AllowLiveCommunication: false,
			AllowFUBIntegration:    false,
			AllowEmailSending:      false,
			AllowSMSSending:        false,
			AllowLeadCreation:      false,
			AllowAutomation:        false,
			TestModePrefix:         "[TEST]",
			MaxTestLeads:           100,
			LastUpdated:            time.Now(),
			UpdatedBy:              "system",
		}

		// Load from environment or config file
		globalSafetyControls.loadFromEnvironment()
	})
	return globalSafetyControls
}

// loadFromEnvironment loads safety settings from environment variables
func (sc *SafetyControls) loadFromEnvironment() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Set mode from environment
	if mode := os.Getenv("PRODUCTION_MODE"); mode != "" {
		sc.Mode = ProductionMode(mode)
	}

	// Load boolean settings
	sc.AllowLiveCommunication = getBoolEnv("ALLOW_LIVE_COMMUNICATION", false)
	sc.AllowFUBIntegration = getBoolEnv("ALLOW_FUB_INTEGRATION", false)
	sc.AllowEmailSending = getBoolEnv("ALLOW_EMAIL_SENDING", false)
	sc.AllowSMSSending = getBoolEnv("ALLOW_SMS_SENDING", false)
	sc.AllowLeadCreation = getBoolEnv("ALLOW_LEAD_CREATION", false)
	sc.AllowAutomation = getBoolEnv("ALLOW_AUTOMATION", false)

	// Load string settings
	if prefix := os.Getenv("TEST_MODE_PREFIX"); prefix != "" {
		sc.TestModePrefix = prefix
	}

	// Load numeric settings
	if maxLeads := os.Getenv("MAX_TEST_LEADS"); maxLeads != "" {
		if val, err := strconv.Atoi(maxLeads); err == nil {
			sc.MaxTestLeads = val
		}
	}

	log.Printf("Safety Controls loaded: Mode=%s, LiveComm=%v, FUB=%v",
		sc.Mode, sc.AllowLiveCommunication, sc.AllowFUBIntegration)
}

// getBoolEnv gets a boolean environment variable with default
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// IsLiveCommunicationAllowed checks if live communication is permitted
func (sc *SafetyControls) IsLiveCommunicationAllowed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.AllowLiveCommunication && sc.Mode == ModeProduction
}

// IsFUBIntegrationAllowed checks if FUB integration is permitted
func (sc *SafetyControls) IsFUBIntegrationAllowed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.AllowFUBIntegration
}

// IsEmailSendingAllowed checks if email sending is permitted
func (sc *SafetyControls) IsEmailSendingAllowed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.AllowEmailSending && sc.IsLiveCommunicationAllowed()
}

// IsSMSSendingAllowed checks if SMS sending is permitted
func (sc *SafetyControls) IsSMSSendingAllowed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.AllowSMSSending && sc.IsLiveCommunicationAllowed()
}

// IsLeadCreationAllowed checks if lead creation is permitted
func (sc *SafetyControls) IsLeadCreationAllowed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.AllowLeadCreation
}

// IsAutomationAllowed checks if automation is permitted
func (sc *SafetyControls) IsAutomationAllowed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.AllowAutomation
}

// GetTestModePrefix returns the prefix for test mode items
func (sc *SafetyControls) GetTestModePrefix() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.TestModePrefix
}

// UpdateMode updates the production mode
func (sc *SafetyControls) UpdateMode(mode ProductionMode, updatedBy string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.Mode = mode
	sc.LastUpdated = time.Now()
	sc.UpdatedBy = updatedBy

	// Auto-configure based on mode
	switch mode {
	case ModeDevelopment:
		sc.AllowLiveCommunication = false
		sc.AllowFUBIntegration = false
		sc.AllowEmailSending = false
		sc.AllowSMSSending = false
		sc.AllowLeadCreation = false
		sc.AllowAutomation = false
	case ModeDemo:
		sc.AllowLiveCommunication = false
		sc.AllowFUBIntegration = true // Allow FUB but no live communication
		sc.AllowEmailSending = false
		sc.AllowSMSSending = false
		sc.AllowLeadCreation = true // Allow test lead creation
		sc.AllowAutomation = false
	case ModeProduction:
		// Production mode requires explicit enabling of each feature
		// Don't auto-enable anything for safety
	}

	log.Printf("Production mode updated to %s by %s", mode, updatedBy)
	return nil
}

// EnableFeature enables a specific feature
func (sc *SafetyControls) EnableFeature(feature string, enabled bool, updatedBy string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	switch feature {
	case "live_communication":
		sc.AllowLiveCommunication = enabled
	case "fub_integration":
		sc.AllowFUBIntegration = enabled
	case "email_sending":
		sc.AllowEmailSending = enabled
	case "sms_sending":
		sc.AllowSMSSending = enabled
	case "lead_creation":
		sc.AllowLeadCreation = enabled
	case "automation":
		sc.AllowAutomation = enabled
	default:
		return fmt.Errorf("unknown feature: %s", feature)
	}

	sc.LastUpdated = time.Now()
	sc.UpdatedBy = updatedBy

	log.Printf("Feature %s set to %v by %s", feature, enabled, updatedBy)
	return nil
}

// GetStatus returns the current safety status
func (sc *SafetyControls) GetStatus() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return map[string]interface{}{
		"mode":                     sc.Mode,
		"allow_live_communication": sc.AllowLiveCommunication,
		"allow_fub_integration":    sc.AllowFUBIntegration,
		"allow_email_sending":      sc.AllowEmailSending,
		"allow_sms_sending":        sc.AllowSMSSending,
		"allow_lead_creation":      sc.AllowLeadCreation,
		"allow_automation":         sc.AllowAutomation,
		"test_mode_prefix":         sc.TestModePrefix,
		"max_test_leads":           sc.MaxTestLeads,
		"last_updated":             sc.LastUpdated,
		"updated_by":               sc.UpdatedBy,
		"is_safe_mode":             !sc.AllowLiveCommunication,
	}
}

// ToJSON returns the safety controls as JSON
func (sc *SafetyControls) ToJSON() ([]byte, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return json.Marshal(sc)
}

// ValidateAction checks if an action is allowed based on current safety settings
func (sc *SafetyControls) ValidateAction(action string) error {
	switch action {
	case "send_email":
		if !sc.IsEmailSendingAllowed() {
			return fmt.Errorf("email sending is disabled in %s mode", sc.Mode)
		}
	case "send_sms":
		if !sc.IsSMSSendingAllowed() {
			return fmt.Errorf("SMS sending is disabled in %s mode", sc.Mode)
		}
	case "create_fub_lead":
		if !sc.IsFUBIntegrationAllowed() {
			return fmt.Errorf("FUB integration is disabled in %s mode", sc.Mode)
		}
		if !sc.IsLeadCreationAllowed() {
			return fmt.Errorf("lead creation is disabled in %s mode", sc.Mode)
		}
	case "trigger_automation":
		if !sc.IsAutomationAllowed() {
			return fmt.Errorf("automation is disabled in %s mode", sc.Mode)
		}
	case "live_communication":
		if !sc.IsLiveCommunicationAllowed() {
			return fmt.Errorf("live communication is disabled in %s mode", sc.Mode)
		}
	}
	return nil
}

// GetSafetyWarnings returns any active safety warnings
func (sc *SafetyControls) GetSafetyWarnings() []string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var warnings []string

	if sc.Mode == ModeProduction && !sc.AllowLiveCommunication {
		warnings = append(warnings, "Production mode active but live communication disabled")
	}

	if sc.AllowFUBIntegration && !sc.AllowLeadCreation {
		warnings = append(warnings, "FUB integration enabled but lead creation disabled")
	}

	if sc.AllowAutomation && !sc.AllowLiveCommunication {
		warnings = append(warnings, "Automation enabled but live communication disabled")
	}

	return warnings
}
