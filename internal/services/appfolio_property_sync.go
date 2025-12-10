package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

type AppFolioPropertySync struct {
	db              *gorm.DB
	apiKey          string
	baseURL         string
	companyID       string
	lastSyncTime    time.Time
	syncErrors      []models.SyncError
}

type AppFolioProperty struct {
	ID            string   `json:"id"`
	Address       string   `json:"address"`
	City          string   `json:"city"`
	State         string   `json:"state"`
	ZipCode       string   `json:"zip_code"`
	PropertyType  string   `json:"property_type"`
	Bedrooms      int      `json:"bedrooms"`
	Bathrooms     float32  `json:"bathrooms"`
	SquareFeet    int      `json:"square_feet"`
	Status        string   `json:"status"`
	RentAmount    float64  `json:"rent_amount"`
	MarketRent    float64  `json:"market_rent"`
	VacantDate    *string  `json:"vacant_date"`
	CurrentTenant *string  `json:"current_tenant_id"`
	OwnerID       string   `json:"owner_id"`
	UnitNumber    string   `json:"unit_number"`
	IsVacant      bool     `json:"is_vacant"`
	LastUpdated   string   `json:"last_updated"`
}

type AppFolioPropertyResponse struct {
	Properties []AppFolioProperty `json:"properties"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	HasMore    bool               `json:"has_more"`
}

func NewAppFolioPropertySync(db *gorm.DB, apiKey, companyID string) *AppFolioPropertySync {
	return &AppFolioPropertySync{
		db:         db,
		apiKey:     apiKey,
		baseURL:    "https://api.appfolio.com/v1",
		companyID:  companyID,
		syncErrors: make([]models.SyncError, 0),
	}
}

func (s *AppFolioPropertySync) SyncProperties() (*models.PropertySyncResult, error) {
	log.Println("🏠 Starting AppFolio property sync...")
	startTime := time.Now()

	result := &models.PropertySyncResult{
		StartedAt: startTime,
		Source:    "appfolio",
	}

	page := 1
	perPage := 100
	hasMore := true

	for hasMore {
		properties, resp, err := s.fetchPropertiesPage(page, perPage)
		if err != nil {
			s.addSyncError("fetch_properties", fmt.Sprintf("page_%d", page), err.Error())
			result.Errors = append(result.Errors, models.SyncError{
				Entity:    "property",
				Operation: "fetch",
				Message:   err.Error(),
				Timestamp: time.Now(),
			})
			break
		}

		for _, afProp := range properties {
			syncErr := s.syncSingleProperty(afProp)
			if syncErr != nil {
				result.Failed++
				result.Errors = append(result.Errors, *syncErr)
			} else {
				result.Synced++
			}
		}

		hasMore = resp.HasMore
		page++
	}

	vacancies, err := s.syncVacancies()
	if err != nil {
		log.Printf("⚠️ Error syncing vacancies: %v", err)
	} else {
		result.VacanciesUpdated = vacancies
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(startTime)
	s.lastSyncTime = time.Now()

	log.Printf("✅ AppFolio property sync complete: %d synced, %d failed, %d vacancies updated",
		result.Synced, result.Failed, result.VacanciesUpdated)

	return result, nil
}

func (s *AppFolioPropertySync) fetchPropertiesPage(page, perPage int) ([]AppFolioProperty, *AppFolioPropertyResponse, error) {
	url := fmt.Sprintf("%s/properties?page=%d&per_page=%d&company_id=%s",
		s.baseURL, page, perPage, s.companyID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch properties: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("AppFolio API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response AppFolioPropertyResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Properties, &response, nil
}

func (s *AppFolioPropertySync) syncSingleProperty(afProp AppFolioProperty) *models.SyncError {
	var existingState models.PropertyState
	err := s.db.Where("appfolio_id = ?", afProp.ID).First(&existingState).Error

	if err == gorm.ErrRecordNotFound {
		newState := models.PropertyState{
			AppFolioID:      afProp.ID,
			Address:         afProp.Address,
			City:            afProp.City,
			State:           afProp.State,
			ZipCode:         afProp.ZipCode,
			PropertyType:    afProp.PropertyType,
			Bedrooms:        &afProp.Bedrooms,
			Bathrooms:       &afProp.Bathrooms,
			SquareFeet:      &afProp.SquareFeet,
			Status:          s.mapAppFolioStatus(afProp.Status),
			StatusSource:    "appfolio",
			StatusUpdatedAt: time.Now(),
			RentAmount:      &afProp.RentAmount,
			IsVacant:        afProp.IsVacant,
			AppFolioData: models.JSONB{
				"owner_id":       afProp.OwnerID,
				"unit_number":    afProp.UnitNumber,
				"market_rent":    afProp.MarketRent,
				"current_tenant": afProp.CurrentTenant,
				"vacant_date":    afProp.VacantDate,
			},
			LastSyncedAt: time.Now(),
		}

		if err := s.db.Create(&newState).Error; err != nil {
			return &models.SyncError{
				Entity:       "property",
				EntityID:     afProp.ID,
				Operation:    "create",
				Message:      err.Error(),
				Timestamp:    time.Now(),
				IsRetryable:  true,
			}
		}
		return nil
	} else if err != nil {
		return &models.SyncError{
			Entity:       "property",
			EntityID:     afProp.ID,
			Operation:    "lookup",
			Message:      err.Error(),
			Timestamp:    time.Now(),
			IsRetryable:  true,
		}
	}

	updates := map[string]interface{}{
		"status":            s.mapAppFolioStatus(afProp.Status),
		"status_source":     "appfolio",
		"status_updated_at": time.Now(),
		"is_vacant":         afProp.IsVacant,
		"rent_amount":       afProp.RentAmount,
		"last_synced_at":    time.Now(),
		"appfolio_data": models.JSONB{
			"owner_id":       afProp.OwnerID,
			"unit_number":    afProp.UnitNumber,
			"market_rent":    afProp.MarketRent,
			"current_tenant": afProp.CurrentTenant,
			"vacant_date":    afProp.VacantDate,
		},
	}

	if err := s.db.Model(&existingState).Updates(updates).Error; err != nil {
		return &models.SyncError{
			Entity:       "property",
			EntityID:     afProp.ID,
			Operation:    "update",
			Message:      err.Error(),
			Timestamp:    time.Now(),
			IsRetryable:  true,
		}
	}

	return nil
}

func (s *AppFolioPropertySync) syncVacancies() (int, error) {
	url := fmt.Sprintf("%s/vacancies?company_id=%s", s.baseURL, s.companyID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("AppFolio API error: status %d", resp.StatusCode)
	}

	var vacancies struct {
		Properties []struct {
			ID         string `json:"id"`
			VacantDate string `json:"vacant_date"`
		} `json:"vacancies"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&vacancies); err != nil {
		return 0, err
	}

	count := 0
	for _, v := range vacancies.Properties {
		result := s.db.Model(&models.PropertyState{}).
			Where("appfolio_id = ?", v.ID).
			Updates(map[string]interface{}{
				"is_vacant":         true,
				"status":            "vacant",
				"status_source":     "appfolio",
				"status_updated_at": time.Now(),
			})
		if result.RowsAffected > 0 {
			count++
		}
	}

	return count, nil
}

func (s *AppFolioPropertySync) GetVacantProperties() ([]models.PropertyState, error) {
	var properties []models.PropertyState
	err := s.db.Where("is_vacant = ? AND status_source = ?", true, "appfolio").Find(&properties).Error
	return properties, err
}

func (s *AppFolioPropertySync) PushPropertyUpdate(propertyState *models.PropertyState) error {
	if propertyState.AppFolioID == "" {
		return fmt.Errorf("property has no AppFolio ID")
	}

	url := fmt.Sprintf("%s/properties/%s", s.baseURL, propertyState.AppFolioID)

	payload := map[string]interface{}{
		"status":       propertyState.Status,
		"is_listed":    propertyState.IsBookable,
		"market_rent":  propertyState.RentAmount,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AppFolio API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ Pushed property update to AppFolio: %s", propertyState.AppFolioID)
	return nil
}

func (s *AppFolioPropertySync) HandleVacancyWebhook(webhookData map[string]interface{}) error {
	propertyID, ok := webhookData["property_id"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook: missing property_id")
	}

	eventType, _ := webhookData["event_type"].(string)

	var isVacant bool
	var status string

	switch eventType {
	case "tenant.moved_out", "lease.ended":
		isVacant = true
		status = "vacant"
	case "tenant.moved_in", "lease.started":
		isVacant = false
		status = "occupied"
	default:
		log.Printf("⚠️ Unhandled AppFolio vacancy event: %s", eventType)
		return nil
	}

	result := s.db.Model(&models.PropertyState{}).
		Where("appfolio_id = ?", propertyID).
		Updates(map[string]interface{}{
			"is_vacant":         isVacant,
			"status":            status,
			"status_source":     "appfolio_webhook",
			"status_updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	log.Printf("🏠 Property %s status updated via webhook: %s (vacant: %v)", propertyID, status, isVacant)
	return nil
}

func (s *AppFolioPropertySync) mapAppFolioStatus(afStatus string) string {
	statusMap := map[string]string{
		"vacant":    "vacant",
		"occupied":  "occupied",
		"available": "active",
		"rented":    "occupied",
		"off_market": "off_market",
		"maintenance": "maintenance",
		"pending":   "pending",
	}

	if mapped, ok := statusMap[afStatus]; ok {
		return mapped
	}
	return afStatus
}

func (s *AppFolioPropertySync) GetLastSyncTime() time.Time {
	return s.lastSyncTime
}

func (s *AppFolioPropertySync) GetSyncErrors() []models.SyncError {
	return s.syncErrors
}

func (s *AppFolioPropertySync) addSyncError(operation, entityID, message string) {
	s.syncErrors = append(s.syncErrors, models.SyncError{
		Entity:    "property",
		EntityID:  entityID,
		Operation: operation,
		Message:   message,
		Timestamp: time.Now(),
	})
}

func (s *AppFolioPropertySync) ClearSyncErrors() {
	s.syncErrors = make([]models.SyncError, 0)
}
