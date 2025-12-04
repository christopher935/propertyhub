package services

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
)

// CentralPropertyStateManager manages the single source of truth for all property data
type CentralPropertyStateManager struct {
	db                *gorm.DB
	encryptionManager *security.EncryptionManager
	realTimeSync      interface{}
}

// NewCentralPropertyStateManager creates a new central property state manager
func NewCentralPropertyStateManager(db *gorm.DB, encryptionManager *security.EncryptionManager) *CentralPropertyStateManager {
	return &CentralPropertyStateManager{
		db:                db,
		encryptionManager: encryptionManager,
	}
}

// CreateOrUpdateProperty creates or updates property - EXACT handler signature
func (cpsm *CentralPropertyStateManager) CreateOrUpdateProperty(request models.PropertyUpdateRequest) (*models.PropertyState, error) {
	log.Printf("🔄 Creating/updating property: %s", request.Address)

	// Convert PropertyUpdateRequest to PropertyState
	propertyState := &models.PropertyState{
		MLSId:        request.MLSId,
		Address:      request.Address,
		Price:        request.Price,
		Bedrooms:     request.Bedrooms,
		Bathrooms:    request.Bathrooms,
		SquareFeet:   request.SquareFeet,
		PropertyType: request.PropertyType,
		Status:       request.Status,
		StatusSource: request.Source,
		StatusUpdatedAt: time.Now(),
	}

	// Check if property exists
	var existingProperty models.PropertyState
	result := cpsm.db.Where("mls_id = ?", request.MLSId).First(&existingProperty)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new property
		if err := cpsm.db.Create(propertyState).Error; err != nil {
			return nil, fmt.Errorf("failed to create property: %v", err)
		}
		log.Printf("✅ Created property: %s", propertyState.Address)
		return propertyState, nil
	} else if result.Error == nil {
		// Update existing property
		existingProperty.Price = request.Price
		existingProperty.Bedrooms = request.Bedrooms
		existingProperty.Bathrooms = request.Bathrooms
		existingProperty.SquareFeet = request.SquareFeet
		existingProperty.Status = request.Status
		existingProperty.StatusSource = request.Source
		existingProperty.StatusUpdatedAt = time.Now()

		if err := cpsm.db.Save(&existingProperty).Error; err != nil {
			return nil, fmt.Errorf("failed to update property: %v", err)
		}
		log.Printf("✅ Updated property: %s", existingProperty.Address)
		return &existingProperty, nil
	}

	return nil, fmt.Errorf("database error: %v", result.Error)
}

// GetSystemStats returns enterprise system statistics - EXACT handler signature
func (cpsm *CentralPropertyStateManager) GetSystemStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalProperties, activeProperties, pendingProperties int64
	
	cpsm.db.Model(&models.PropertyState{}).Count(&totalProperties)
	cpsm.db.Model(&models.PropertyState{}).Where("status = ?", "active").Count(&activeProperties)
	cpsm.db.Model(&models.PropertyState{}).Where("status = ?", "pending").Count(&pendingProperties)

	stats["total_properties"] = totalProperties
	stats["active_properties"] = activeProperties
	stats["pending_properties"] = pendingProperties
	stats["sync_health"] = "operational"
	stats["last_updated"] = time.Now()
	stats["system_version"] = "enterprise"

	return stats, nil
}

// SetRealTimeSync sets the real-time sync service
func (cpsm *CentralPropertyStateManager) SetRealTimeSync(syncService interface{}) {
	cpsm.realTimeSync = syncService
	log.Printf("🔄 Real-time sync service attached to central state manager")
}

// ResolveConflict resolves conflicts - EXACT handler signature (3 parameters)
func (cpsm *CentralPropertyStateManager) ResolveConflict(propertyID uint, conflictType string, resolution string) error {
	log.Printf("🔧 Resolving conflict for property ID %d: %s using %s", propertyID, conflictType, resolution)

	var property models.PropertyState
	if err := cpsm.db.First(&property, propertyID).Error; err != nil {
		return fmt.Errorf("property not found for conflict resolution: %v", err)
	}

	switch resolution {
	case "manual_override":
		property.StatusSource = "manual"
		property.StatusUpdatedAt = time.Now()
	case "har_authoritative":
		property.StatusSource = "har"
		property.StatusUpdatedAt = time.Now()
	case "fub_authoritative":
		property.StatusSource = "fub"
		property.StatusUpdatedAt = time.Now()
	default:
		return fmt.Errorf("unknown resolution strategy: %s", resolution)
	}

	if err := cpsm.db.Save(&property).Error; err != nil {
		return fmt.Errorf("failed to save conflict resolution: %v", err)
	}

	log.Printf("✅ Conflict resolved for property: %s", property.Address)
	return nil
}

// GetPropertyState retrieves the current state of a property
func (cpsm *CentralPropertyStateManager) GetPropertyState(mlsID string) (*models.PropertyState, error) {
	var property models.PropertyState
	if err := cpsm.db.Where("mls_id = ?", mlsID).First(&property).Error; err != nil {
		return nil, fmt.Errorf("property not found: %v", err)
	}
	return &property, nil
}

// GetAllProperties retrieves all properties in the central state
func (cpsm *CentralPropertyStateManager) GetAllProperties() ([]models.PropertyState, error) {
	var properties []models.PropertyState
	if err := cpsm.db.Find(&properties).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve properties: %v", err)
	}
	return properties, nil
}


// GetPublicProperties retrieves only properties available for public showing (with photos)
func (cpsm *CentralPropertyStateManager) GetPublicProperties() ([]models.PropertyState, error) {
	var propertyStates []models.PropertyState
	
	// Use Raw SQL to ensure the WHERE clause is applied
	err := cpsm.db.Raw(`
		SELECT * FROM property_states 
		WHERE is_available_for_showing = true
	`).Scan(&propertyStates).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve public properties: %v", err)
	}
	
	log.Printf("🔍 GetPublicProperties returned %d properties", len(propertyStates))
	return propertyStates, nil
}
