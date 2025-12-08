package services

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// OverrideManager handles agent manual overrides of automation
// This gives agents full control over Copilot automation
type OverrideManager struct {
	db *gorm.DB
}

// NewOverrideManager creates a new override manager
func NewOverrideManager(db *gorm.DB) *OverrideManager {
	return &OverrideManager{
		db: db,
	}
}

// EmergencyState represents the global emergency stop state
type EmergencyState struct {
	Active bool   `json:"active"`
	Reason string `json:"reason"`
	// ActivatedBy string  `json:"activated_by"`
	ActivatedAt time.Time `json:"activated_at"`
}

// GlobalOverride represents system-wide automation overrides
type GlobalOverride struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Type   string `gorm:"uniqueIndex" json:"type"` // "emergency_stop", "maintenance_mode"
	Active bool   `json:"active"`
	Reason string `json:"reason"`
	// ActivatedBy string    `json:"activated_by"`
	ActivatedAt   time.Time  `json:"activated_at"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`
}

// SetDoNotContact marks a lead as Do Not Contact
func (om *OverrideManager) SetDoNotContact(leadID string, reason string, setBy string) error {
	// Check if override already exists
	var override LeadOverride
	err := om.db.Where("lead_id = ?", leadID).First(&override).Error

	if err == gorm.ErrRecordNotFound {
		// Create new override
		override = LeadOverride{
			LeadID:       leadID,
			DoNotContact: true,
			Reason:       reason,
			SetBy:        setBy,
		}
		err = om.db.Create(&override).Error
	} else if err == nil {
		// Update existing override
		override.DoNotContact = true
		override.Reason = reason
		override.SetBy = setBy
		err = om.db.Save(&override).Error
	}

	if err != nil {
		log.Printf("❌ Error setting Do Not Contact for lead %s: %v", leadID, err)
		return err
	}

	log.Printf("🚫 Do Not Contact set for lead %s by %s: %s", leadID, setBy, reason)
	return nil
}

// ClearDoNotContact removes Do Not Contact flag from a lead
func (om *OverrideManager) ClearDoNotContact(leadID string, clearedBy string) error {
	var override LeadOverride
	err := om.db.Where("lead_id = ?", leadID).First(&override).Error

	if err == gorm.ErrRecordNotFound {
		return nil // Already clear
	} else if err != nil {
		return err
	}

	override.DoNotContact = false
	override.SetBy = clearedBy
	err = om.db.Save(&override).Error

	if err != nil {
		log.Printf("❌ Error clearing Do Not Contact for lead %s: %v", leadID, err)
		return err
	}

	log.Printf("✅ Do Not Contact cleared for lead %s by %s", leadID, clearedBy)
	return nil
}

// SetPauseUntil pauses automation for a lead until a specific date
func (om *OverrideManager) SetPauseUntil(leadID string, pauseUntil time.Time, reason string, setBy string) error {
	var override LeadOverride
	err := om.db.Where("lead_id = ?", leadID).First(&override).Error

	if err == gorm.ErrRecordNotFound {
		// Create new override
		override = LeadOverride{
			LeadID:     leadID,
			PauseUntil: &pauseUntil,
			Reason:     reason,
			SetBy:      setBy,
		}
		err = om.db.Create(&override).Error
	} else if err == nil {
		// Update existing override
		override.PauseUntil = &pauseUntil
		override.Reason = reason
		override.SetBy = setBy
		err = om.db.Save(&override).Error
	}

	if err != nil {
		log.Printf("❌ Error setting Pause Until for lead %s: %v", leadID, err)
		return err
	}

	log.Printf("⏸️  Automation paused for lead %s until %s by %s: %s",
		leadID, pauseUntil.Format("Jan 2, 3:04 PM"), setBy, reason)
	return nil
}

// ClearPauseUntil removes pause from a lead
func (om *OverrideManager) ClearPauseUntil(leadID string, clearedBy string) error {
	var override LeadOverride
	err := om.db.Where("lead_id = ?", leadID).First(&override).Error

	if err == gorm.ErrRecordNotFound {
		return nil // Already clear
	} else if err != nil {
		return err
	}

	override.PauseUntil = nil
	override.SetBy = clearedBy
	err = om.db.Save(&override).Error

	if err != nil {
		log.Printf("❌ Error clearing Pause Until for lead %s: %v", leadID, err)
		return err
	}

	log.Printf("▶️  Automation resumed for lead %s by %s", leadID, clearedBy)
	return nil
}

// SetCustomCooldown sets a custom cooldown period for a lead (in hours)
func (om *OverrideManager) SetCustomCooldown(leadID string, cooldownHours int, reason string, setBy string) error {
	var override LeadOverride
	err := om.db.Where("lead_id = ?", leadID).First(&override).Error

	if err == gorm.ErrRecordNotFound {
		// Create new override
		override = LeadOverride{
			LeadID:         leadID,
			CustomCooldown: &cooldownHours,
			Reason:         reason,
			SetBy:          setBy,
		}
		err = om.db.Create(&override).Error
	} else if err == nil {
		// Update existing override
		override.CustomCooldown = &cooldownHours
		override.Reason = reason
		override.SetBy = setBy
		err = om.db.Save(&override).Error
	}

	if err != nil {
		log.Printf("❌ Error setting custom cooldown for lead %s: %v", leadID, err)
		return err
	}

	log.Printf("⏱️  Custom cooldown set for lead %s: %d hours by %s: %s",
		leadID, cooldownHours, setBy, reason)
	return nil
}

// GetLeadOverride returns the current override settings for a lead
func (om *OverrideManager) GetLeadOverride(leadID string) (*LeadOverride, error) {
	var override LeadOverride
	err := om.db.Where("lead_id = ?", leadID).First(&override).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // No override set
	} else if err != nil {
		return nil, err
	}

	return &override, nil
}

// ActivateEmergencyStop activates emergency stop - ALL automation is paused
func (om *OverrideManager) ActivateEmergencyStop(reason string, activatedBy string) error {
	var globalOverride GlobalOverride
	err := om.db.Where("type = ?", "emergency_stop").First(&globalOverride).Error

	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// Create new emergency stop
		globalOverride = GlobalOverride{
			Type:   "emergency_stop",
			Active: true,
			Reason: reason,
			// ActivatedBy: activatedBy,
			ActivatedAt: now,
		}
		err = om.db.Create(&globalOverride).Error
	} else if err == nil {
		// Update existing emergency stop
		globalOverride.Active = true
		globalOverride.Reason = reason
		// ActivatedBy := activatedBy
		globalOverride.ActivatedAt = now
		globalOverride.DeactivatedAt = nil
		err = om.db.Save(&globalOverride).Error
	}

	if err != nil {
		log.Printf("❌ Error activating emergency stop: %v", err)
		return err
	}

	log.Printf("🚨 EMERGENCY STOP ACTIVATED by %s: %s", activatedBy, reason)
	return nil
}

// DeactivateEmergencyStop deactivates emergency stop - automation resumes
func (om *OverrideManager) DeactivateEmergencyStop(deactivatedBy string) error {
	var globalOverride GlobalOverride
	err := om.db.Where("type = ?", "emergency_stop").First(&globalOverride).Error

	if err == gorm.ErrRecordNotFound {
		return nil // No emergency stop active
	} else if err != nil {
		return err
	}

	now := time.Now()
	globalOverride.Active = false
	globalOverride.DeactivatedAt = &now
	err = om.db.Save(&globalOverride).Error

	if err != nil {
		log.Printf("❌ Error deactivating emergency stop: %v", err)
		return err
	}

	log.Printf("✅ Emergency stop deactivated by %s", deactivatedBy)
	return nil
}

// GetEmergencyState returns the current emergency stop state
func (om *OverrideManager) GetEmergencyState() EmergencyState {
	var globalOverride GlobalOverride
	err := om.db.Where("type = ? AND active = ?", "emergency_stop", true).First(&globalOverride).Error

	if err == gorm.ErrRecordNotFound {
		return EmergencyState{Active: false}
	} else if err != nil {
		log.Printf("❌ Error getting emergency state: %v", err)
		return EmergencyState{Active: false}
	}

	return EmergencyState{
		Active: globalOverride.Active,
		Reason: globalOverride.Reason,
		// ActivatedBy: globalOverride.ActivatedBy,
		ActivatedAt: globalOverride.ActivatedAt,
	}
}

// IsEmergencyStopActive returns true if emergency stop is currently active
func (om *OverrideManager) IsEmergencyStopActive() bool {
	state := om.GetEmergencyState()
	return state.Active
}

// GetAllLeadOverrides returns all leads with active overrides
func (om *OverrideManager) GetAllLeadOverrides() ([]LeadOverride, error) {
	var overrides []LeadOverride
	err := om.db.Where("do_not_contact = ? OR pause_until IS NOT NULL OR custom_cooldown IS NOT NULL",
		true).Find(&overrides).Error

	return overrides, err
}

// RemoveAllOverrides removes all overrides for a lead
func (om *OverrideManager) RemoveAllOverrides(leadID string, removedBy string) error {
	err := om.db.Where("lead_id = ?", leadID).Delete(&LeadOverride{}).Error

	if err != nil {
		log.Printf("❌ Error removing overrides for lead %s: %v", leadID, err)
		return err
	}

	log.Printf("🗑️  All overrides removed for lead %s by %s", leadID, removedBy)
	return nil
}

// AutoMigrate creates the necessary database tables
func (om *OverrideManager) AutoMigrate() error {
	return om.db.AutoMigrate(&GlobalOverride{})
}
