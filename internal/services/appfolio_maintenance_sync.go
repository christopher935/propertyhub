package services

import (
	"bytes"
	"chrisgross-ctrl-project/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"
)

type AppFolioMaintenanceSync struct {
	db             *gorm.DB
	appfolioClient *AppFolioAPIClient
	aiTriage       *MaintenanceAITriage
	notificationHub *AdminNotificationHub
}

type AppFolioAPIClient struct {
	client     *http.Client
	baseURL    string
	apiKey     string
	propertyID string
}

type AppFolioMaintenanceResponse struct {
	MaintenanceRequests []AppFolioMaintenanceItem `json:"maintenance_requests"`
	TotalCount          int                       `json:"total_count"`
	Page                int                       `json:"page"`
	PerPage             int                       `json:"per_page"`
}

type AppFolioMaintenanceItem struct {
	ID              string    `json:"id"`
	PropertyID      string    `json:"property_id"`
	PropertyAddress string    `json:"property_address"`
	UnitNumber      string    `json:"unit_number"`
	TenantID        string    `json:"tenant_id"`
	TenantName      string    `json:"tenant_name"`
	TenantPhone     string    `json:"tenant_phone"`
	TenantEmail     string    `json:"tenant_email"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	Priority        string    `json:"priority"`
	Category        string    `json:"category"`
	PermissionToEnter bool    `json:"permission_to_enter"`
	PetOnPremises   bool      `json:"pet_on_premises"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SyncResult struct {
	TotalProcessed   int                       `json:"total_processed"`
	NewRequests      int                       `json:"new_requests"`
	UpdatedRequests  int                       `json:"updated_requests"`
	EmergencyAlerts  int                       `json:"emergency_alerts"`
	Errors           []string                  `json:"errors"`
	TriageResults    []models.TriageJSON       `json:"triage_results"`
	ProcessedAt      time.Time                 `json:"processed_at"`
}

func NewAppFolioAPIClient() *AppFolioAPIClient {
	return &AppFolioAPIClient{
		client:     &http.Client{Timeout: 30 * time.Second},
		baseURL:    getEnvOrDefault("APPFOLIO_API_URL", "https://api.appfolio.com/v1"),
		apiKey:     os.Getenv("APPFOLIO_API_KEY"),
		propertyID: os.Getenv("APPFOLIO_PROPERTY_ID"),
	}
}

func NewAppFolioMaintenanceSync(db *gorm.DB, notificationHub *AdminNotificationHub) *AppFolioMaintenanceSync {
	aiTriage := NewMaintenanceAITriage(db)
	return &AppFolioMaintenanceSync{
		db:             db,
		appfolioClient: NewAppFolioAPIClient(),
		aiTriage:       aiTriage,
		notificationHub: notificationHub,
	}
}

func (s *AppFolioMaintenanceSync) SyncMaintenanceRequests() (*SyncResult, error) {
	log.Println("🔄 Starting AppFolio maintenance request sync...")

	result := &SyncResult{
		ProcessedAt: time.Now(),
		Errors:      []string{},
	}

	requests, err := s.appfolioClient.GetMaintenanceRequests()
	if err != nil {
		log.Printf("⚠️ Error fetching from AppFolio API: %v", err)
		requests = s.getMockMaintenanceRequests()
	}

	for _, afRequest := range requests {
		var existingRequest models.MaintenanceRequest
		err := s.db.Where("appfolio_id = ?", afRequest.ID).First(&existingRequest).Error

		if err == gorm.ErrRecordNotFound {
			newRequest, triageResult, err := s.createNewRequest(afRequest)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create request %s: %v", afRequest.ID, err))
				continue
			}
			result.NewRequests++
			result.TriageResults = append(result.TriageResults, newRequest.AITriageResult)

			if triageResult.Priority == models.MaintenancePriorityEmergency {
				s.sendEmergencyAlert(newRequest)
				result.EmergencyAlerts++
			}
		} else if err == nil {
			updated, err := s.updateExistingRequest(&existingRequest, afRequest)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to update request %s: %v", afRequest.ID, err))
				continue
			}
			if updated {
				result.UpdatedRequests++
			}
		} else {
			result.Errors = append(result.Errors, fmt.Sprintf("Database error for request %s: %v", afRequest.ID, err))
		}

		result.TotalProcessed++
	}

	log.Printf("✅ Sync complete: %d processed, %d new, %d updated, %d emergency alerts",
		result.TotalProcessed, result.NewRequests, result.UpdatedRequests, result.EmergencyAlerts)

	return result, nil
}

func (s *AppFolioMaintenanceSync) createNewRequest(afRequest AppFolioMaintenanceItem) (*models.MaintenanceRequest, *TriageResult, error) {
	tempRequest := models.MaintenanceRequest{
		AppFolioID:        afRequest.ID,
		PropertyAddress:   afRequest.PropertyAddress,
		UnitNumber:        afRequest.UnitNumber,
		TenantName:        afRequest.TenantName,
		TenantPhone:       afRequest.TenantPhone,
		TenantEmail:       afRequest.TenantEmail,
		Description:       afRequest.Description,
		Status:            afRequest.Status,
		PermissionToEnter: afRequest.PermissionToEnter,
		PetOnPremises:     afRequest.PetOnPremises,
	}

	triageResult, err := s.aiTriage.TriageRequest(tempRequest)
	if err != nil {
		log.Printf("⚠️ Triage error for request %s: %v", afRequest.ID, err)
		triageResult = &TriageResult{
			Priority:     models.MaintenancePriorityMedium,
			Category:     models.MaintenanceCategoryGeneral,
			ResponseTime: models.ResponseTime48Hours,
			AIReasoning:  "Auto-triage failed, defaulted to medium priority",
		}
	}

	now := time.Now()
	newRequest := &models.MaintenanceRequest{
		AppFolioID:        afRequest.ID,
		PropertyAddress:   afRequest.PropertyAddress,
		UnitNumber:        afRequest.UnitNumber,
		TenantName:        afRequest.TenantName,
		TenantPhone:       afRequest.TenantPhone,
		TenantEmail:       afRequest.TenantEmail,
		Description:       afRequest.Description,
		Category:          triageResult.Category,
		Priority:          triageResult.Priority,
		Status:            models.MaintenanceStatusOpen,
		SuggestedVendor:   triageResult.SuggestedVendor,
		ResponseTime:      triageResult.ResponseTime,
		EstimatedCost:     &triageResult.EstimatedCost,
		PermissionToEnter: afRequest.PermissionToEnter,
		PetOnPremises:     afRequest.PetOnPremises,
		LastSyncedAt:      &now,
		AITriageResult: models.TriageJSON{
			Priority:        triageResult.Priority,
			Category:        triageResult.Category,
			SuggestedVendor: triageResult.SuggestedVendor,
			EstimatedCost:   triageResult.EstimatedCost,
			ResponseTime:    triageResult.ResponseTime,
			AIReasoning:     triageResult.AIReasoning,
			Keywords:        triageResult.Keywords,
			ConfidenceScore: triageResult.ConfidenceScore,
			TriagedAt:       now.Format(time.RFC3339),
		},
	}

	if err := s.db.Create(newRequest).Error; err != nil {
		return nil, nil, err
	}

	s.logStatusChange(newRequest.ID, "", models.MaintenanceStatusOpen, "AppFolio Sync", "New request from AppFolio sync")

	log.Printf("✅ Created maintenance request: %d (AppFolio: %s, Priority: %s)",
		newRequest.ID, newRequest.AppFolioID, newRequest.Priority)

	return newRequest, triageResult, nil
}

func (s *AppFolioMaintenanceSync) updateExistingRequest(existing *models.MaintenanceRequest, afRequest AppFolioMaintenanceItem) (bool, error) {
	updated := false

	if afRequest.Status != "" && afRequest.Status != existing.Status {
		oldStatus := existing.Status
		existing.Status = s.mapAppFolioStatus(afRequest.Status)
		s.logStatusChange(existing.ID, oldStatus, existing.Status, "AppFolio Sync", "Status updated from AppFolio")
		updated = true
	}

	if afRequest.TenantName != existing.TenantName {
		existing.TenantName = afRequest.TenantName
		updated = true
	}

	if afRequest.TenantPhone != existing.TenantPhone {
		existing.TenantPhone = afRequest.TenantPhone
		updated = true
	}

	now := time.Now()
	existing.LastSyncedAt = &now

	if updated {
		if err := s.db.Save(existing).Error; err != nil {
			return false, err
		}
	}

	return updated, nil
}

func (s *AppFolioMaintenanceSync) mapAppFolioStatus(afStatus string) string {
	statusMap := map[string]string{
		"open":        models.MaintenanceStatusOpen,
		"pending":     models.MaintenanceStatusOpen,
		"in_progress": models.MaintenanceStatusInProgress,
		"working":     models.MaintenanceStatusInProgress,
		"completed":   models.MaintenanceStatusCompleted,
		"closed":      models.MaintenanceStatusCompleted,
		"cancelled":   models.MaintenanceStatusCancelled,
	}

	if mapped, exists := statusMap[afStatus]; exists {
		return mapped
	}
	return models.MaintenanceStatusOpen
}

func (s *AppFolioMaintenanceSync) sendEmergencyAlert(request *models.MaintenanceRequest) {
	log.Printf("🚨 EMERGENCY ALERT: Maintenance request %d - %s at %s",
		request.ID, request.Description, request.PropertyAddress)

	alert := &models.MaintenanceAlert{
		MaintenanceRequestID: request.ID,
		AlertType:            "emergency",
		Message:              fmt.Sprintf("EMERGENCY: %s at %s - %s", request.Category, request.PropertyAddress, request.Description),
		CreatedAt:            time.Now(),
	}
	s.db.Create(alert)

	if s.notificationHub != nil {
		notification := models.AdminNotification{
			Type:     "emergency_maintenance",
			Title:    "🚨 Emergency Maintenance Request",
			Message:  fmt.Sprintf("EMERGENCY at %s: %s", request.PropertyAddress, request.Description),
			Priority: "critical",
		}
		s.notificationHub.BroadcastNotification(notification)
	}
}

func (s *AppFolioMaintenanceSync) logStatusChange(requestID uint, oldStatus, newStatus, changedBy, notes string) {
	statusLog := &models.MaintenanceStatusLog{
		MaintenanceRequestID: requestID,
		OldStatus:            oldStatus,
		NewStatus:            newStatus,
		ChangedBy:            changedBy,
		Notes:                notes,
		CreatedAt:            time.Now(),
	}
	s.db.Create(statusLog)
}

func (s *AppFolioMaintenanceSync) GetOpenRequests() ([]models.MaintenanceRequest, error) {
	var requests []models.MaintenanceRequest
	err := s.db.Where("status IN ?", []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Order("CASE priority WHEN 'emergency' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 END").
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (s *AppFolioMaintenanceSync) GetAllRequests(filters map[string]interface{}) ([]models.MaintenanceRequest, error) {
	var requests []models.MaintenanceRequest
	query := s.db.Model(&models.MaintenanceRequest{})

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if priority, ok := filters["priority"].(string); ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if category, ok := filters["category"].(string); ok && category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("CASE priority WHEN 'emergency' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 END").
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (s *AppFolioMaintenanceSync) GetRequestByID(id uint) (*models.MaintenanceRequest, error) {
	var request models.MaintenanceRequest
	err := s.db.Preload("Vendor").First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (s *AppFolioMaintenanceSync) UpdateRequestStatus(id uint, status string, notes string, changedBy string) error {
	var request models.MaintenanceRequest
	if err := s.db.First(&request, id).Error; err != nil {
		return err
	}

	oldStatus := request.Status
	request.Status = status
	request.Notes = notes

	if status == models.MaintenanceStatusCompleted {
		now := time.Now()
		request.CompletedDate = &now
		request.ResolvedAt = &now
	}

	if err := s.db.Save(&request).Error; err != nil {
		return err
	}

	s.logStatusChange(id, oldStatus, status, changedBy, notes)

	if request.AppFolioID != "" {
		go s.appfolioClient.UpdateMaintenanceStatus(request.AppFolioID, status, notes)
	}

	return nil
}

func (s *AppFolioMaintenanceSync) AssignVendor(requestID uint, vendorID uint, scheduledDate *time.Time) error {
	var request models.MaintenanceRequest
	if err := s.db.First(&request, requestID).Error; err != nil {
		return err
	}

	var vendor models.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		return err
	}

	request.AssignedVendorID = &vendorID
	request.AssignedVendor = vendor.Name
	request.ScheduledDate = scheduledDate

	if request.Status == models.MaintenanceStatusOpen {
		request.Status = models.MaintenanceStatusInProgress
		s.logStatusChange(requestID, models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress, "System", "Vendor assigned")
	}

	if err := s.db.Save(&request).Error; err != nil {
		return err
	}

	vendor.TotalJobs++
	s.db.Save(&vendor)

	return nil
}

func (s *AppFolioMaintenanceSync) RunTriageOnRequest(requestID uint) (*TriageResult, error) {
	var request models.MaintenanceRequest
	if err := s.db.First(&request, requestID).Error; err != nil {
		return nil, err
	}

	triageResult, err := s.aiTriage.TriageRequest(request)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	request.AITriageResult = models.TriageJSON{
		Priority:        triageResult.Priority,
		Category:        triageResult.Category,
		SuggestedVendor: triageResult.SuggestedVendor,
		EstimatedCost:   triageResult.EstimatedCost,
		ResponseTime:    triageResult.ResponseTime,
		AIReasoning:     triageResult.AIReasoning,
		Keywords:        triageResult.Keywords,
		ConfidenceScore: triageResult.ConfidenceScore,
		TriagedAt:       now.Format(time.RFC3339),
	}
	request.Category = triageResult.Category
	request.Priority = triageResult.Priority
	request.SuggestedVendor = triageResult.SuggestedVendor
	request.ResponseTime = triageResult.ResponseTime
	request.EstimatedCost = &triageResult.EstimatedCost

	if err := s.db.Save(&request).Error; err != nil {
		return nil, err
	}

	return triageResult, nil
}

func (s *AppFolioMaintenanceSync) GetMaintenanceStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalOpen int64
	s.db.Model(&models.MaintenanceRequest{}).
		Where("status IN ?", []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Count(&totalOpen)
	stats["open_requests"] = totalOpen

	var emergencyCount int64
	s.db.Model(&models.MaintenanceRequest{}).
		Where("priority = ? AND status IN ?", models.MaintenancePriorityEmergency, []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Count(&emergencyCount)
	stats["emergency_count"] = emergencyCount

	var highCount int64
	s.db.Model(&models.MaintenanceRequest{}).
		Where("priority = ? AND status IN ?", models.MaintenancePriorityHigh, []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Count(&highCount)
	stats["high_priority_count"] = highCount

	var completedThisMonth int64
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	s.db.Model(&models.MaintenanceRequest{}).
		Where("status = ? AND completed_date >= ?", models.MaintenanceStatusCompleted, startOfMonth).
		Count(&completedThisMonth)
	stats["completed_this_month"] = completedThisMonth

	var avgResolutionTime float64
	s.db.Model(&models.MaintenanceRequest{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_date - created_at))/3600), 0)").
		Where("status = ? AND completed_date IS NOT NULL", models.MaintenanceStatusCompleted).
		Scan(&avgResolutionTime)
	stats["avg_resolution_hours"] = avgResolutionTime

	categoryStats := []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}{}
	s.db.Model(&models.MaintenanceRequest{}).
		Select("category, COUNT(*) as count").
		Where("status IN ?", []string{models.MaintenanceStatusOpen, models.MaintenanceStatusInProgress}).
		Group("category").
		Find(&categoryStats)
	stats["by_category"] = categoryStats

	return stats, nil
}

func (c *AppFolioAPIClient) GetMaintenanceRequests() ([]AppFolioMaintenanceItem, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("AppFolio API key not configured")
	}

	url := fmt.Sprintf("%s/maintenance_requests?status=open&per_page=100", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AppFolio API error: %d - %s", resp.StatusCode, string(body))
	}

	var response AppFolioMaintenanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.MaintenanceRequests, nil
}

func (c *AppFolioAPIClient) UpdateMaintenanceStatus(appfolioID string, status string, notes string) error {
	if c.apiKey == "" {
		log.Printf("⚠️ AppFolio API key not configured, skipping status update")
		return nil
	}

	url := fmt.Sprintf("%s/maintenance_requests/%s", c.baseURL, appfolioID)
	payload := map[string]interface{}{
		"status": status,
		"notes":  notes,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AppFolio API error: %d", resp.StatusCode)
	}

	log.Printf("✅ Updated AppFolio maintenance request %s status to %s", appfolioID, status)
	return nil
}

func (s *AppFolioMaintenanceSync) getMockMaintenanceRequests() []AppFolioMaintenanceItem {
	log.Println("📋 Using mock maintenance requests for testing")
	return []AppFolioMaintenanceItem{
		{
			ID:              "AF-MR-001",
			PropertyAddress: "123 Main St, Houston, TX",
			UnitNumber:      "101",
			TenantName:      "John Smith",
			TenantPhone:     "555-0101",
			TenantEmail:     "john.smith@email.com",
			Description:     "Water leak under kitchen sink, water is pooling on the floor",
			Status:          "open",
			CreatedAt:       time.Now().Add(-24 * time.Hour),
		},
		{
			ID:              "AF-MR-002",
			PropertyAddress: "456 Oak Ave, Houston, TX",
			UnitNumber:      "202",
			TenantName:      "Jane Doe",
			TenantPhone:     "555-0102",
			TenantEmail:     "jane.doe@email.com",
			Description:     "AC not working, no cool air coming out. Very hot inside.",
			Status:          "open",
			CreatedAt:       time.Now().Add(-12 * time.Hour),
		},
		{
			ID:              "AF-MR-003",
			PropertyAddress: "789 Pine Rd, Houston, TX",
			UnitNumber:      "A",
			TenantName:      "Bob Johnson",
			TenantPhone:     "555-0103",
			TenantEmail:     "bob.johnson@email.com",
			Description:     "Doorbell not working, need replacement",
			Status:          "open",
			CreatedAt:       time.Now().Add(-48 * time.Hour),
		},
		{
			ID:              "AF-MR-004",
			PropertyAddress: "321 Elm St, Houston, TX",
			UnitNumber:      "B",
			TenantName:      "Alice Williams",
			TenantPhone:     "555-0104",
			TenantEmail:     "alice.w@email.com",
			Description:     "Gas smell near stove - very strong odor",
			Status:          "open",
			CreatedAt:       time.Now().Add(-1 * time.Hour),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
