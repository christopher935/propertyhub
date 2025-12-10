package services

import (
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

type AppFolioPropertySync struct {
	db             *gorm.DB
	appfolioClient *AppFolioAPIClient
}

type PropertySyncResult struct {
	Created   int      `json:"created"`
	Updated   int      `json:"updated"`
	Skipped   int      `json:"skipped"`
	Errors    []string `json:"errors"`
	SyncedAt  time.Time `json:"synced_at"`
}

func NewAppFolioPropertySync(db *gorm.DB, client *AppFolioAPIClient) *AppFolioPropertySync {
	return &AppFolioPropertySync{
		db:             db,
		appfolioClient: client,
	}
}

func (s *AppFolioPropertySync) SyncPropertiesFromAppFolio() (*PropertySyncResult, error) {
	result := &PropertySyncResult{
		SyncedAt: time.Now(),
		Errors:   make([]string, 0),
	}

	page := 1
	perPage := 100

	for {
		properties, err := s.appfolioClient.GetProperties(page, perPage)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to fetch page %d: %v", page, err))
			break
		}

		if len(properties) == 0 {
			break
		}

		for _, afProperty := range properties {
			syncErr := s.syncSingleProperty(afProperty, result)
			if syncErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("property %s: %v", afProperty.ID, syncErr))
			}
		}

		if len(properties) < perPage {
			break
		}
		page++
	}

	log.Printf("✅ AppFolio property sync completed: %d created, %d updated, %d skipped, %d errors",
		result.Created, result.Updated, result.Skipped, len(result.Errors))

	return result, nil
}

func (s *AppFolioPropertySync) syncSingleProperty(afProperty AppFolioProperty, result *PropertySyncResult) error {
	var existingProperty models.Property
	err := s.db.Where("mls_id = ? OR address = ?", afProperty.ID, afProperty.Address).First(&existingProperty).Error

	if err == gorm.ErrRecordNotFound {
		newProperty := models.Property{
			MLSId:        afProperty.ID,
			City:         afProperty.City,
			State:        afProperty.State,
			ZipCode:      afProperty.ZipCode,
			PropertyType: afProperty.PropertyType,
			Price:        afProperty.MonthlyRent,
			Status:       mapAppFolioStatusToPropertyHub(afProperty.Status),
			Source:       "AppFolio",
		}

		if err := s.db.Create(&newProperty).Error; err != nil {
			return fmt.Errorf("failed to create property: %w", err)
		}
		result.Created++
		log.Printf("📦 Created property from AppFolio: %s", afProperty.Address)
	} else if err != nil {
		return fmt.Errorf("failed to query property: %w", err)
	} else {
		existingProperty.Price = afProperty.MonthlyRent
		existingProperty.Status = mapAppFolioStatusToPropertyHub(afProperty.Status)
		existingProperty.PropertyType = afProperty.PropertyType
		existingProperty.UpdatedAt = time.Now()

		if err := s.db.Save(&existingProperty).Error; err != nil {
			return fmt.Errorf("failed to update property: %w", err)
		}
		result.Updated++
		log.Printf("📦 Updated property from AppFolio: %s", afProperty.Address)
	}

	return nil
}

func (s *AppFolioPropertySync) PushPropertyToAppFolio(property models.Property) (*AppFolioProperty, error) {
	existing, err := s.appfolioClient.GetPropertyByAddress(string(property.Address))
	if err != nil {
		return nil, fmt.Errorf("failed to check existing property: %w", err)
	}

	if existing != nil {
		log.Printf("📦 Property already exists in AppFolio: %s (ID: %s)", existing.Address, existing.ID)
		return existing, nil
	}

	log.Printf("⚠️ Property creation in AppFolio requires manual setup for: %s", property.Address)
	return nil, fmt.Errorf("property creation not supported - requires manual setup in AppFolio")
}

func (s *AppFolioPropertySync) GetAppFolioPropertyID(propertyID uint) (string, error) {
	var property models.Property
	if err := s.db.First(&property, propertyID).Error; err != nil {
		return "", fmt.Errorf("property not found: %w", err)
	}

	if property.MLSId != "" {
		afProperty, err := s.appfolioClient.GetProperty(property.MLSId)
		if err == nil && afProperty != nil {
			return afProperty.ID, nil
		}
	}

	afProperty, err := s.appfolioClient.GetPropertyByAddress(string(property.Address))
	if err != nil {
		return "", fmt.Errorf("failed to find property in AppFolio: %w", err)
	}

	if afProperty == nil {
		return "", fmt.Errorf("property not found in AppFolio: %s", property.Address)
	}

	return afProperty.ID, nil
}

func mapAppFolioStatusToPropertyHub(afStatus string) string {
	switch afStatus {
	case "available", "vacant":
		return "active"
	case "occupied", "leased":
		return "rented"
	case "pending":
		return "pending"
	case "maintenance", "offline":
		return "inactive"
	default:
		return "active"
	}
}

func mapPropertyHubStatusToAppFolio(phStatus string) string {
	switch phStatus {
	case "active":
		return "available"
	case "rented", "leased":
		return "occupied"
	case "pending":
		return "pending"
	case "inactive":
		return "offline"
	default:
		return "available"
	}
}
