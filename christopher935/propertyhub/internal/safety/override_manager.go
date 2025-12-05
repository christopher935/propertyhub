package safety

import (
	"fmt"
	"log"
	"time"
)

// OverrideManager handles safety overrides and emergency controls
type OverrideManager struct {
	configManager   *SafetyConfigManager
	activeOverrides map[string]*Override
	emergencyState  *EmergencyState
}

// Override represents a temporary safety override
type Override struct {
	ID               string                 `json:"id"`
	Type             OverrideType           `json:"type"`
	CreatedBy        string                 `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	ExpiresAt        time.Time              `json:"expires_at"`
	Reason           string                 `json:"reason"`
	Scope            string                 `json:"scope"`     // "global", "campaign", "lead"
	TargetID         string                 `json:"target_id"` // ID of campaign/lead if scoped
	OriginalSettings map[string]interface{} `json:"original_settings"`
	OverrideSettings map[string]interface{} `json:"override_settings"`
	Active           bool                   `json:"active"`
	UsageCount       int                    `json:"usage_count"`
	LastUsed         *time.Time             `json:"last_used,omitempty"`
}

// OverrideType defines the type of override
type OverrideType int

const (
	OverrideTypeTemporary   OverrideType = iota // Temporary override with expiration
	OverrideTypeEmergency                       // Emergency override (immediate, short duration)
	OverrideTypeMaintenance                     // Maintenance override (scheduled, longer duration)
	OverrideTypeAdmin                           // Admin override (manual, flexible duration)
)

// EmergencyState tracks the current emergency status
type EmergencyState struct {
	Active              bool           `json:"active"`
	ActivatedAt         time.Time      `json:"activated_at"`
	ActivatedBy         string         `json:"activated_by"`
	Reason              string         `json:"reason"`
	Level               EmergencyLevel `json:"level"`
	AutomationStopped   bool           `json:"automation_stopped"`
	AffectedSystems     []string       `json:"affected_systems"`
	ResolutionSteps     []string       `json:"resolution_steps"`
	EstimatedResolution *time.Time     `json:"estimated_resolution,omitempty"`
}

// EmergencyLevel defines the severity of an emergency
type EmergencyLevel int

const (
	EmergencyLevelLow      EmergencyLevel = iota // Minor issues, partial restrictions
	EmergencyLevelMedium                         // Moderate issues, significant restrictions
	EmergencyLevelHigh                           // Major issues, most automation stopped
	EmergencyLevelCritical                       // Critical issues, all automation stopped
)

// OverrideRequest represents a request for a safety override
type OverrideRequest struct {
	Type             OverrideType           `json:"type"`
	Reason           string                 `json:"reason"`
	RequestedBy      string                 `json:"requested_by"`
	DurationHours    int                    `json:"duration_hours"`
	Scope            string                 `json:"scope"`
	TargetID         string                 `json:"target_id,omitempty"`
	OverrideSettings map[string]interface{} `json:"override_settings"`
	Justification    string                 `json:"justification"`
	ApprovalRequired bool                   `json:"approval_required"`
}

// NewOverrideManager creates a new override manager
func NewOverrideManager(configManager *SafetyConfigManager) *OverrideManager {
	return &OverrideManager{
		configManager:   configManager,
		activeOverrides: make(map[string]*Override),
		emergencyState:  &EmergencyState{Active: false},
	}
}

// RequestOverride requests a safety override
func (m *OverrideManager) RequestOverride(request OverrideRequest) (*Override, error) {
	// Validate the request
	if err := m.validateOverrideRequest(request); err != nil {
		return nil, fmt.Errorf("invalid override request: %v", err)
	}

	// Check authorization
	if !m.isAuthorizedForOverride(request.RequestedBy, request.Type) {
		return nil, fmt.Errorf("user %s not authorized for %s override",
			request.RequestedBy, m.getOverrideTypeString(request.Type))
	}

	// Create the override
	override := &Override{
		ID:               m.generateOverrideID(),
		Type:             request.Type,
		CreatedBy:        request.RequestedBy,
		CreatedAt:        time.Now(),
		ExpiresAt:        time.Now().Add(time.Duration(request.DurationHours) * time.Hour),
		Reason:           request.Reason,
		Scope:            request.Scope,
		TargetID:         request.TargetID,
		OriginalSettings: m.getCurrentSettings(request.Scope, request.TargetID),
		OverrideSettings: request.OverrideSettings,
		Active:           true,
		UsageCount:       0,
	}

	// Apply the override
	if err := m.applyOverride(override); err != nil {
		return nil, fmt.Errorf("failed to apply override: %v", err)
	}

	// Store the override
	m.activeOverrides[override.ID] = override

	// Log the override
	log.Printf("Safety override created: %s by %s. Reason: %s",
		override.ID, override.CreatedBy, override.Reason)

	return override, nil
}

// ActivateEmergencyStop activates emergency stop procedures
func (m *OverrideManager) ActivateEmergencyStop(activatedBy, reason string, level EmergencyLevel) error {
	log.Printf("EMERGENCY STOP ACTIVATED by %s: %s", activatedBy, reason)

	m.emergencyState = &EmergencyState{
		Active:            true,
		ActivatedAt:       time.Now(),
		ActivatedBy:       activatedBy,
		Reason:            reason,
		Level:             level,
		AutomationStopped: true,
		AffectedSystems:   []string{"automation", "campaigns", "communications"},
		ResolutionSteps:   m.getEmergencyResolutionSteps(level),
	}

	// Stop all automation based on emergency level
	switch level {
	case EmergencyLevelCritical:
		// Stop everything immediately
		if err := m.stopAllAutomation(); err != nil {
			log.Printf("Error stopping automation during emergency: %v", err)
		}
		// Revert to strictest safety mode
		if err := m.configManager.UpdateMode(SafetyModeStrict, "emergency-system"); err != nil {
			log.Printf("Error reverting to strict mode during emergency: %v", err)
		}

	case EmergencyLevelHigh:
		// Stop most automation, allow only manual operations
		if err := m.stopAutomatedCampaigns(); err != nil {
			log.Printf("Error stopping automated campaigns during emergency: %v", err)
		}

	case EmergencyLevelMedium:
		// Increase safety restrictions significantly
		if err := m.configManager.UpdateMode(SafetyModeStrict, "emergency-system"); err != nil {
			log.Printf("Error increasing safety during emergency: %v", err)
		}

	case EmergencyLevelLow:
		// Apply additional safety checks
		if err := m.enableAdditionalSafetyChecks(); err != nil {
			log.Printf("Error enabling additional safety checks: %v", err)
		}
	}

	// Notify relevant parties
	m.notifyEmergencyActivation()

	return nil
}

// DeactivateEmergencyStop deactivates emergency stop procedures
func (m *OverrideManager) DeactivateEmergencyStop(deactivatedBy, reason string) error {
	if !m.emergencyState.Active {
		return fmt.Errorf("no active emergency to deactivate")
	}

	log.Printf("Emergency stop deactivated by %s: %s", deactivatedBy, reason)

	// Restore normal operations gradually
	if err := m.restoreNormalOperations(); err != nil {
		return fmt.Errorf("failed to restore normal operations: %v", err)
	}

	// Update emergency state
	m.emergencyState.Active = false

	// Log the deactivation
	log.Printf("Emergency state cleared. Duration: %v",
		time.Since(m.emergencyState.ActivatedAt))

	return nil
}

// ExpireOverrides removes expired overrides
func (m *OverrideManager) ExpireOverrides() error {
	now := time.Now()
	expiredCount := 0

	for id, override := range m.activeOverrides {
		if override.Active && now.After(override.ExpiresAt) {
			// Revert the override
			if err := m.revertOverride(override); err != nil {
				log.Printf("Error reverting expired override %s: %v", id, err)
				continue
			}

			// Mark as inactive
			override.Active = false
			expiredCount++

			log.Printf("Override %s expired and reverted", id)
		}
	}

	if expiredCount > 0 {
		log.Printf("Expired %d safety overrides", expiredCount)
	}

	return nil
}

// GetActiveOverrides returns all active overrides
func (m *OverrideManager) GetActiveOverrides() []*Override {
	var active []*Override
	for _, override := range m.activeOverrides {
		if override.Active {
			active = append(active, override)
		}
	}
	return active
}

// GetEmergencyState returns the current emergency state
func (m *OverrideManager) GetEmergencyState() *EmergencyState {
	return m.emergencyState
}

// RevokeOverride manually revokes an active override
func (m *OverrideManager) RevokeOverride(overrideID, revokedBy, reason string) error {
	override, exists := m.activeOverrides[overrideID]
	if !exists {
		return fmt.Errorf("override %s not found", overrideID)
	}

	if !override.Active {
		return fmt.Errorf("override %s is not active", overrideID)
	}

	// Check authorization to revoke
	if !m.isAuthorizedToRevoke(revokedBy, override) {
		return fmt.Errorf("user %s not authorized to revoke override %s", revokedBy, overrideID)
	}

	// Revert the override
	if err := m.revertOverride(override); err != nil {
		return fmt.Errorf("failed to revert override: %v", err)
	}

	// Mark as inactive
	override.Active = false

	log.Printf("Override %s revoked by %s. Reason: %s", overrideID, revokedBy, reason)

	return nil
}

// validateOverrideRequest validates an override request
func (m *OverrideManager) validateOverrideRequest(request OverrideRequest) error {
	if request.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	if request.RequestedBy == "" {
		return fmt.Errorf("requested_by is required")
	}

	if request.DurationHours <= 0 || request.DurationHours > 168 { // Max 1 week
		return fmt.Errorf("duration must be between 1 and 168 hours")
	}

	if request.Scope != "global" && request.Scope != "campaign" && request.Scope != "lead" {
		return fmt.Errorf("scope must be 'global', 'campaign', or 'lead'")
	}

	if (request.Scope == "campaign" || request.Scope == "lead") && request.TargetID == "" {
		return fmt.Errorf("target_id is required for %s scope", request.Scope)
	}

	return nil
}

// isAuthorizedForOverride checks if a user is authorized for an override type
func (m *OverrideManager) isAuthorizedForOverride(user string, overrideType OverrideType) bool {
	config := m.configManager.GetConfig()

	// Check if user is in authorized list
	for _, authorizedUser := range config.OverrideSettings.AuthorizedOverrideUsers {
		if user == authorizedUser {
			return true
		}
	}

	// Admin can always override
	if user == "admin" {
		return true
	}

	// Emergency overrides allowed for system
	if overrideType == OverrideTypeEmergency && user == "emergency-system" {
		return true
	}

	return false
}

// isAuthorizedToRevoke checks if a user can revoke an override
func (m *OverrideManager) isAuthorizedToRevoke(user string, override *Override) bool {
	// Creator can revoke their own override
	if user == override.CreatedBy {
		return true
	}

	// Admin can revoke any override
	if user == "admin" {
		return true
	}

	// System can revoke emergency overrides
	if user == "emergency-system" && override.Type == OverrideTypeEmergency {
		return true
	}

	return false
}

// applyOverride applies the override settings
func (m *OverrideManager) applyOverride(override *Override) error {
	// Implementation would apply the override settings
	// This is a placeholder for the actual implementation
	log.Printf("Applying override %s with scope %s", override.ID, override.Scope)
	return nil
}

// revertOverride reverts an override to original settings
func (m *OverrideManager) revertOverride(override *Override) error {
	// Implementation would revert to original settings
	// This is a placeholder for the actual implementation
	log.Printf("Reverting override %s", override.ID)
	return nil
}

// getCurrentSettings gets current settings for the specified scope
func (m *OverrideManager) getCurrentSettings(scope, targetID string) map[string]interface{} {
	// Implementation would get current settings
	// This is a placeholder for the actual implementation
	return map[string]interface{}{
		"scope":     scope,
		"target_id": targetID,
		"timestamp": time.Now(),
	}
}

// stopAllAutomation stops all automation immediately
func (m *OverrideManager) stopAllAutomation() error {
	log.Printf("STOPPING ALL AUTOMATION - Emergency Level Critical")
	// Implementation would stop all automation
	return nil
}

// stopAutomatedCampaigns stops automated campaigns
func (m *OverrideManager) stopAutomatedCampaigns() error {
	log.Printf("STOPPING AUTOMATED CAMPAIGNS - Emergency Level High")
	// Implementation would stop automated campaigns
	return nil
}

// enableAdditionalSafetyChecks enables additional safety checks
func (m *OverrideManager) enableAdditionalSafetyChecks() error {
	log.Printf("ENABLING ADDITIONAL SAFETY CHECKS - Emergency Level Medium")
	// Implementation would enable additional checks
	return nil
}

// restoreNormalOperations gradually restores normal operations
func (m *OverrideManager) restoreNormalOperations() error {
	log.Printf("RESTORING NORMAL OPERATIONS - Emergency Deactivated")
	// Implementation would restore normal operations
	return nil
}

// notifyEmergencyActivation notifies relevant parties of emergency
func (m *OverrideManager) notifyEmergencyActivation() {
	log.Printf("EMERGENCY NOTIFICATION SENT - Level: %s",
		m.getEmergencyLevelString(m.emergencyState.Level))
	// Implementation would send notifications
}

// getEmergencyResolutionSteps returns resolution steps for emergency level
func (m *OverrideManager) getEmergencyResolutionSteps(level EmergencyLevel) []string {
	switch level {
	case EmergencyLevelCritical:
		return []string{
			"Investigate root cause immediately",
			"Contact system administrator",
			"Review all recent changes",
			"Verify data integrity",
			"Test systems before reactivation",
		}
	case EmergencyLevelHigh:
		return []string{
			"Review recent campaign performance",
			"Check for complaint spikes",
			"Verify lead data quality",
			"Test automation workflows",
		}
	case EmergencyLevelMedium:
		return []string{
			"Monitor performance metrics",
			"Review safety thresholds",
			"Check system logs",
			"Validate recent changes",
		}
	case EmergencyLevelLow:
		return []string{
			"Monitor for 30 minutes",
			"Review performance trends",
			"Check for anomalies",
		}
	default:
		return []string{"Review system status"}
	}
}

// generateOverrideID generates a unique override ID
func (m *OverrideManager) generateOverrideID() string {
	return fmt.Sprintf("override_%d", time.Now().UnixNano())
}

// getOverrideTypeString returns string representation of override type
func (m *OverrideManager) getOverrideTypeString(overrideType OverrideType) string {
	switch overrideType {
	case OverrideTypeTemporary:
		return "Temporary"
	case OverrideTypeEmergency:
		return "Emergency"
	case OverrideTypeMaintenance:
		return "Maintenance"
	case OverrideTypeAdmin:
		return "Admin"
	default:
		return "Unknown"
	}
}

// getEmergencyLevelString returns string representation of emergency level
func (m *OverrideManager) getEmergencyLevelString(level EmergencyLevel) string {
	switch level {
	case EmergencyLevelLow:
		return "Low"
	case EmergencyLevelMedium:
		return "Medium"
	case EmergencyLevelHigh:
		return "High"
	case EmergencyLevelCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// GetOverrideStats returns statistics about overrides
func (m *OverrideManager) GetOverrideStats() map[string]interface{} {
	activeCount := len(m.GetActiveOverrides())
	totalCount := len(m.activeOverrides)

	return map[string]interface{}{
		"active_overrides": activeCount,
		"total_overrides":  totalCount,
		"emergency_active": m.emergencyState.Active,
		"emergency_level":  m.getEmergencyLevelString(m.emergencyState.Level),
		"last_emergency":   m.emergencyState.ActivatedAt,
	}
}
