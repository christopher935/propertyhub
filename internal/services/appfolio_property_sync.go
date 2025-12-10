package services

import (
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	ExternalSourceAppFolio = "appfolio"

	AppFolioStatusVacant      = "vacant"
	AppFolioStatusOccupied    = "occupied"
	AppFolioStatusOffMarket   = "off_market"
	AppFolioStatusMaintenance = "maintenance"
)

type AppFolioPropertySync struct {
	db             *gorm.DB
	appfolioClient *AppFolioAPIClient
}

type SyncResult struct {
	Success           bool      `json:"success"`
	Message           string    `json:"message"`
	PropertiesSynced  int       `json:"properties_synced"`
	PropertiesCreated int       `json:"properties_created"`
	PropertiesUpdated int       `json:"properties_updated"`
	PropertiesDeleted int       `json:"properties_deleted"`
	Errors            []string  `json:"errors,omitempty"`
	StartedAt         time.Time `json:"started_at"`
	CompletedAt       time.Time `json:"completed_at"`
	DurationMs        int64     `json:"duration_ms"`
}

func NewAppFolioPropertySync(db *gorm.DB, appfolioClient *AppFolioAPIClient) *AppFolioPropertySync {
	return &AppFolioPropertySync{
		db:             db,
		appfolioClient: appfolioClient,
	}
}

func (s *AppFolioPropertySync) SyncPropertiesFromAppFolio() (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		StartedAt: startTime,
	}

	syncStatus := models.SyncStatus{
		SyncType:  "full_sync",
		Status:    "running",
		StartedAt: startTime,
	}
	s.db.Create(&syncStatus)

	log.Printf("🔄 Starting AppFolio property sync...")

	afProperties, err := s.appfolioClient.GetAllProperties()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to fetch properties from AppFolio: %v", err)
		result.Errors = append(result.Errors, err.Error())
		s.updateSyncStatus(&syncStatus, "failed", result)
		return result, err
	}

	log.Printf("📥 Retrieved %d properties from AppFolio", len(afProperties))

	var syncedExternalIDs []string
	var errors []string

	for _, afProp := range afProperties {
		syncedExternalIDs = append(syncedExternalIDs, afProp.PropertyID)

		created, err := s.syncProperty(afProp)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to sync property %s: %v", afProp.PropertyID, err)
			errors = append(errors, errMsg)
			log.Printf("❌ %s", errMsg)
			continue
		}

		if created {
			result.PropertiesCreated++
		} else {
			result.PropertiesUpdated++
		}
		result.PropertiesSynced++
	}

	deletedCount, err := s.handleDeletedProperties(syncedExternalIDs)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Error handling deleted properties: %v", err))
	}
	result.PropertiesDeleted = deletedCount

	result.CompletedAt = time.Now()
	result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()
	result.Errors = errors
	result.Success = len(errors) == 0
	if result.Success {
		result.Message = fmt.Sprintf("Successfully synced %d properties (%d created, %d updated, %d deleted)",
			result.PropertiesSynced, result.PropertiesCreated, result.PropertiesUpdated, result.PropertiesDeleted)
	} else {
		result.Message = fmt.Sprintf("Sync completed with %d errors. Synced %d properties.",
			len(errors), result.PropertiesSynced)
	}

	s.updateSyncStatus(&syncStatus, "completed", result)

	log.Printf("✅ AppFolio sync completed: %s", result.Message)
	return result, nil
}

func (s *AppFolioPropertySync) SyncSingleProperty(appfolioID string) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		StartedAt: startTime,
	}

	log.Printf("🔄 Syncing single property: %s", appfolioID)

	afProperty, err := s.appfolioClient.GetProperty(appfolioID)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to fetch property %s from AppFolio: %v", appfolioID, err)
		return result, err
	}

	created, err := s.syncProperty(*afProperty)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to sync property %s: %v", appfolioID, err)
		return result, err
	}

	result.PropertiesSynced = 1
	if created {
		result.PropertiesCreated = 1
	} else {
		result.PropertiesUpdated = 1
	}

	result.CompletedAt = time.Now()
	result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()
	result.Success = true
	result.Message = fmt.Sprintf("Successfully synced property %s", appfolioID)

	log.Printf("✅ Single property sync completed: %s", result.Message)
	return result, nil
}

func (s *AppFolioPropertySync) GetVacancies() ([]AppFolioProperty, error) {
	vacantProps, err := s.appfolioClient.GetVacantProperties()
	if err != nil {
		return nil, fmt.Errorf("failed to get vacancies from AppFolio: %w", err)
	}

	log.Printf("📋 Retrieved %d vacant properties from AppFolio", len(vacantProps))
	return vacantProps, nil
}

func (s *AppFolioPropertySync) MapAppFolioToProperty(af AppFolioProperty) models.Property {
	now := time.Now()

	bedrooms := af.Bedrooms
	bathrooms := af.Bathrooms
	sqft := af.SquareFeet

	status := s.mapAppFolioStatus(af.Status)

	price := af.RentAmount
	if price == 0 {
		price = af.MarketRent
	}

	fullAddress := af.Address
	if af.Address2 != "" {
		fullAddress = fmt.Sprintf("%s %s", af.Address, af.Address2)
	}

	var images pq.StringArray
	for _, img := range af.Images {
		images = append(images, img)
	}

	return models.Property{
		ExternalID:       af.PropertyID,
		ExternalSource:   ExternalSourceAppFolio,
		LastSyncedAt:     &now,
		Address:          security.EncryptedString(fullAddress),
		City:             af.City,
		State:            af.State,
		ZipCode:          af.Zip,
		Bedrooms:         &bedrooms,
		Bathrooms:        &bathrooms,
		SquareFeet:       &sqft,
		Price:            price,
		Status:           status,
		PropertyType:     af.PropertyType,
		Description:      af.Description,
		YearBuilt:        af.YearBuilt,
		PropertyFeatures: af.Features,
		Images:           images,
		Source:           "AppFolio",
		ListingType:      "For Rent",
	}
}

func (s *AppFolioPropertySync) syncProperty(af AppFolioProperty) (bool, error) {
	var existingProperty models.Property
	err := s.db.Where("external_id = ? AND external_source = ?", af.PropertyID, ExternalSourceAppFolio).First(&existingProperty).Error

	if err == gorm.ErrRecordNotFound {
		newProperty := s.MapAppFolioToProperty(af)
		if err := s.db.Create(&newProperty).Error; err != nil {
			return false, fmt.Errorf("failed to create property: %w", err)
		}
		log.Printf("➕ Created new property: %s (%s)", af.Address, af.PropertyID)
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to query existing property: %w", err)
	}

	mappedProperty := s.MapAppFolioToProperty(af)

	updates := map[string]interface{}{
		"address":           mappedProperty.Address,
		"city":              mappedProperty.City,
		"state":             mappedProperty.State,
		"zip_code":          mappedProperty.ZipCode,
		"bedrooms":          mappedProperty.Bedrooms,
		"bathrooms":         mappedProperty.Bathrooms,
		"square_feet":       mappedProperty.SquareFeet,
		"price":             mappedProperty.Price,
		"status":            mappedProperty.Status,
		"property_type":     mappedProperty.PropertyType,
		"description":       mappedProperty.Description,
		"year_built":        mappedProperty.YearBuilt,
		"property_features": mappedProperty.PropertyFeatures,
		"images":            mappedProperty.Images,
		"last_synced_at":    mappedProperty.LastSyncedAt,
	}

	if err := s.db.Model(&existingProperty).Updates(updates).Error; err != nil {
		return false, fmt.Errorf("failed to update property: %w", err)
	}

	log.Printf("🔄 Updated property: %s (%s)", af.Address, af.PropertyID)
	return false, nil
}

func (s *AppFolioPropertySync) handleDeletedProperties(syncedExternalIDs []string) (int, error) {
	if len(syncedExternalIDs) == 0 {
		return 0, nil
	}

	result := s.db.Model(&models.Property{}).
		Where("external_source = ?", ExternalSourceAppFolio).
		Where("external_id NOT IN ?", syncedExternalIDs).
		Update("status", "inactive")

	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark deleted properties as inactive: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("🗑️ Marked %d properties as inactive (removed from AppFolio)", result.RowsAffected)
	}

	return int(result.RowsAffected), nil
}

func (s *AppFolioPropertySync) mapAppFolioStatus(afStatus string) string {
	switch afStatus {
	case AppFolioStatusVacant:
		return "active"
	case AppFolioStatusOccupied:
		return "occupied"
	case AppFolioStatusOffMarket:
		return "inactive"
	case AppFolioStatusMaintenance:
		return "maintenance"
	default:
		return "active"
	}
}

func (s *AppFolioPropertySync) updateSyncStatus(syncStatus *models.SyncStatus, status string, result *SyncResult) {
	now := time.Now()
	syncStatus.Status = status
	syncStatus.CompletedAt = &now
	syncStatus.PropertiesSynced = result.PropertiesSynced
	syncStatus.PropertiesCreated = result.PropertiesCreated
	syncStatus.PropertiesUpdated = result.PropertiesUpdated
	syncStatus.PropertiesDeleted = result.PropertiesDeleted
	syncStatus.ErrorCount = len(result.Errors)

	if len(result.Errors) > 0 {
		errorStr := ""
		for i, err := range result.Errors {
			if i > 0 {
				errorStr += "\n"
			}
			errorStr += err
		}
		syncStatus.Errors = errorStr
	}

	s.db.Save(syncStatus)
}

func (s *AppFolioPropertySync) GetLastSyncStatus() (*models.SyncStatus, error) {
	var syncStatus models.SyncStatus
	err := s.db.Where("sync_type = ?", "full_sync").Order("created_at DESC").First(&syncStatus).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &syncStatus, nil
}

func (s *AppFolioPropertySync) GetSyncHistory(limit int) ([]models.SyncStatus, error) {
	var history []models.SyncStatus
	err := s.db.Order("created_at DESC").Limit(limit).Find(&history).Error
	if err != nil {
		return nil, err
	}
	return history, nil
}

func (s *AppFolioPropertySync) GetAppFolioProperties() ([]models.Property, error) {
	var properties []models.Property
	err := s.db.Where("external_source = ?", ExternalSourceAppFolio).Find(&properties).Error
	if err != nil {
		return nil, err
	}
	return properties, nil
}

func (s *AppFolioPropertySync) GetPropertyByExternalID(externalID string) (*models.Property, error) {
	var property models.Property
	err := s.db.Where("external_id = ? AND external_source = ?", externalID, ExternalSourceAppFolio).First(&property).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &property, nil
}
