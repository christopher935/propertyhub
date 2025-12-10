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

type AppFolioMaintenanceSync struct {
	db                  *gorm.DB
	apiKey              string
	baseURL             string
	companyID           string
	notificationService *NotificationService
	lastSyncTime        time.Time
	syncErrors          []models.SyncError
}

type AppFolioMaintenanceRequest struct {
	ID            string   `json:"id"`
	PropertyID    string   `json:"property_id"`
	UnitID        string   `json:"unit_id"`
	TenantID      string   `json:"tenant_id"`
	TenantName    string   `json:"tenant_name"`
	Category      string   `json:"category"`
	Priority      string   `json:"priority"`
	Status        string   `json:"status"`
	Description   string   `json:"description"`
	RequestedDate string   `json:"requested_date"`
	ScheduledDate *string  `json:"scheduled_date"`
	CompletedDate *string  `json:"completed_date"`
	AssignedTo    *string  `json:"assigned_to"`
	VendorID      *string  `json:"vendor_id"`
	EstimatedCost *float64 `json:"estimated_cost"`
	ActualCost    *float64 `json:"actual_cost"`
	Notes         string   `json:"notes"`
	Images        []string `json:"images"`
	IsEmergency   bool     `json:"is_emergency"`
}

type AppFolioMaintenanceResponse struct {
	Requests []AppFolioMaintenanceRequest `json:"maintenance_requests"`
	Total    int                          `json:"total"`
	Page     int                          `json:"page"`
	PerPage  int                          `json:"per_page"`
	HasMore  bool                         `json:"has_more"`
}

type MaintenanceTriageResult struct {
	RequestID        string
	TriagePriority   string
	SuggestedAction  string
	EstimatedUrgency int
	AIAnalysis       string
	NotifyOwner      bool
	NotifyManager    bool
}

func NewAppFolioMaintenanceSync(db *gorm.DB, apiKey, companyID string, notificationService *NotificationService) *AppFolioMaintenanceSync {
	return &AppFolioMaintenanceSync{
		db:                  db,
		apiKey:              apiKey,
		baseURL:             "https://api.appfolio.com/v1",
		companyID:           companyID,
		notificationService: notificationService,
		syncErrors:          make([]models.SyncError, 0),
	}
}

func (s *AppFolioMaintenanceSync) SyncMaintenanceRequests() (*models.MaintenanceSyncResult, error) {
	log.Println("🔧 Starting AppFolio maintenance sync...")
	startTime := time.Now()

	result := &models.MaintenanceSyncResult{
		StartedAt: startTime,
		Source:    "appfolio",
	}

	page := 1
	perPage := 100
	hasMore := true

	for hasMore {
		requests, resp, err := s.fetchMaintenancePage(page, perPage)
		if err != nil {
			s.addSyncError("fetch_maintenance", fmt.Sprintf("page_%d", page), err.Error())
			result.Errors = append(result.Errors, models.SyncError{
				Entity:    "maintenance",
				Operation: "fetch",
				Message:   err.Error(),
				Timestamp: time.Now(),
			})
			break
		}

		for _, afReq := range requests {
			syncErr := s.syncSingleRequest(afReq)
			if syncErr != nil {
				result.Failed++
				result.Errors = append(result.Errors, *syncErr)
			} else {
				result.Synced++
				if afReq.IsEmergency {
					result.EmergencyCount++
				}
			}
		}

		hasMore = resp.HasMore
		page++
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(startTime)
	s.lastSyncTime = time.Now()

	log.Printf("✅ AppFolio maintenance sync complete: %d synced, %d failed, %d emergencies",
		result.Synced, result.Failed, result.EmergencyCount)

	return result, nil
}

func (s *AppFolioMaintenanceSync) fetchMaintenancePage(page, perPage int) ([]AppFolioMaintenanceRequest, *AppFolioMaintenanceResponse, error) {
	url := fmt.Sprintf("%s/maintenance_requests?page=%d&per_page=%d&company_id=%s&status=open,in_progress",
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
		return nil, nil, fmt.Errorf("failed to fetch maintenance requests: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("AppFolio API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response AppFolioMaintenanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Requests, &response, nil
}

func (s *AppFolioMaintenanceSync) syncSingleRequest(afReq AppFolioMaintenanceRequest) *models.SyncError {
	var existingReq models.MaintenanceRequest
	err := s.db.Where("appfolio_id = ?", afReq.ID).First(&existingReq).Error

	if err == gorm.ErrRecordNotFound {
		newReq := models.MaintenanceRequest{
			AppFolioID:    afReq.ID,
			PropertyID:    afReq.PropertyID,
			UnitID:        afReq.UnitID,
			TenantID:      afReq.TenantID,
			TenantName:    afReq.TenantName,
			Category:      afReq.Category,
			Priority:      afReq.Priority,
			Status:        afReq.Status,
			Description:   afReq.Description,
			RequestedDate: s.parseDate(afReq.RequestedDate),
			ScheduledDate: s.parseDatePtr(afReq.ScheduledDate),
			CompletedDate: s.parseDatePtr(afReq.CompletedDate),
			AssignedTo:    afReq.AssignedTo,
			VendorID:      afReq.VendorID,
			EstimatedCost: afReq.EstimatedCost,
			ActualCost:    afReq.ActualCost,
			Notes:         afReq.Notes,
			IsEmergency:   afReq.IsEmergency,
			LastSyncedAt:  time.Now(),
			AppFolioData: models.JSONB{
				"images": afReq.Images,
			},
		}

		if err := s.db.Create(&newReq).Error; err != nil {
			return &models.SyncError{
				Entity:      "maintenance",
				EntityID:    afReq.ID,
				Operation:   "create",
				Message:     err.Error(),
				Timestamp:   time.Now(),
				IsRetryable: true,
			}
		}

		triageResult := s.triageRequest(&newReq)
		s.handleTriageResult(&newReq, triageResult)

		return nil
	} else if err != nil {
		return &models.SyncError{
			Entity:      "maintenance",
			EntityID:    afReq.ID,
			Operation:   "lookup",
			Message:     err.Error(),
			Timestamp:   time.Now(),
			IsRetryable: true,
		}
	}

	updates := map[string]interface{}{
		"status":         afReq.Status,
		"priority":       afReq.Priority,
		"assigned_to":    afReq.AssignedTo,
		"scheduled_date": s.parseDatePtr(afReq.ScheduledDate),
		"completed_date": s.parseDatePtr(afReq.CompletedDate),
		"actual_cost":    afReq.ActualCost,
		"notes":          afReq.Notes,
		"last_synced_at": time.Now(),
	}

	if err := s.db.Model(&existingReq).Updates(updates).Error; err != nil {
		return &models.SyncError{
			Entity:      "maintenance",
			EntityID:    afReq.ID,
			Operation:   "update",
			Message:     err.Error(),
			Timestamp:   time.Now(),
			IsRetryable: true,
		}
	}

	return nil
}

func (s *AppFolioMaintenanceSync) triageRequest(req *models.MaintenanceRequest) *MaintenanceTriageResult {
	result := &MaintenanceTriageResult{
		RequestID:   req.AppFolioID,
		NotifyOwner: false,
		NotifyManager: true,
	}

	emergencyCategories := map[string]bool{
		"plumbing_leak":       true,
		"no_heat":             true,
		"no_ac":               true,
		"electrical_hazard":   true,
		"gas_leak":            true,
		"security":            true,
		"flood":               true,
		"fire_damage":         true,
		"broken_lock":         true,
	}

	if req.IsEmergency || emergencyCategories[req.Category] {
		result.TriagePriority = "emergency"
		result.EstimatedUrgency = 100
		result.SuggestedAction = "Dispatch emergency service immediately"
		result.NotifyOwner = true
		result.AIAnalysis = fmt.Sprintf("Emergency request detected: %s. Requires immediate attention.", req.Category)
		return result
	}

	highPriorityCategories := map[string]bool{
		"hvac":        true,
		"plumbing":    true,
		"appliance":   true,
		"electrical":  true,
		"pest":        true,
	}

	if highPriorityCategories[req.Category] || req.Priority == "high" {
		result.TriagePriority = "high"
		result.EstimatedUrgency = 75
		result.SuggestedAction = "Schedule within 24-48 hours"
		result.AIAnalysis = fmt.Sprintf("High priority %s request. Should be addressed promptly.", req.Category)
		return result
	}

	mediumPriorityCategories := map[string]bool{
		"doors_windows": true,
		"flooring":      true,
		"painting":      true,
	}

	if mediumPriorityCategories[req.Category] || req.Priority == "medium" {
		result.TriagePriority = "medium"
		result.EstimatedUrgency = 50
		result.SuggestedAction = "Schedule within 1 week"
		result.AIAnalysis = fmt.Sprintf("Medium priority %s request. Can be scheduled for routine maintenance.", req.Category)
		return result
	}

	result.TriagePriority = "low"
	result.EstimatedUrgency = 25
	result.SuggestedAction = "Add to routine maintenance schedule"
	result.AIAnalysis = fmt.Sprintf("Low priority %s request. Can be addressed during next routine visit.", req.Category)

	return result
}

func (s *AppFolioMaintenanceSync) handleTriageResult(req *models.MaintenanceRequest, triage *MaintenanceTriageResult) {
	s.db.Model(req).Updates(map[string]interface{}{
		"triage_priority":    triage.TriagePriority,
		"ai_triage_analysis": triage.AIAnalysis,
		"suggested_action":   triage.SuggestedAction,
		"urgency_score":      triage.EstimatedUrgency,
	})

	if s.notificationService != nil {
		if triage.NotifyOwner || triage.TriagePriority == "emergency" {
			s.notificationService.SendAgentAlert("owner", "Maintenance Alert", fmt.Sprintf(
				"(%s): %s - %s at property %s",
				triage.TriagePriority, req.Category, req.Description, req.PropertyID,
			), map[string]interface{}{"property_id": req.PropertyID, "priority": triage.TriagePriority})
		}

		if triage.NotifyManager {
			s.notificationService.SendAgentAlert("manager", "Maintenance Request", fmt.Sprintf(
				"New request: %s (%s priority) at property %s",
				req.Category, triage.TriagePriority, req.PropertyID,
			), map[string]interface{}{"property_id": req.PropertyID, "priority": triage.TriagePriority})
		}
	}
}

func (s *AppFolioMaintenanceSync) UpdateMaintenanceStatus(appfolioID, status string, notes string) error {
	url := fmt.Sprintf("%s/maintenance_requests/%s", s.baseURL, appfolioID)

	payload := map[string]interface{}{
		"status": status,
	}
	if notes != "" {
		payload["notes"] = notes
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

	s.db.Model(&models.MaintenanceRequest{}).
		Where("appfolio_id = ?", appfolioID).
		Updates(map[string]interface{}{
			"status":         status,
			"last_synced_at": time.Now(),
		})

	return nil
}

func (s *AppFolioMaintenanceSync) HandleMaintenanceWebhook(webhookData map[string]interface{}) error {
	requestID, ok := webhookData["maintenance_request_id"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook: missing maintenance_request_id")
	}

	eventType, _ := webhookData["event_type"].(string)

	switch eventType {
	case "maintenance.created":
		return s.handleMaintenanceCreated(webhookData)
	case "maintenance.updated":
		return s.handleMaintenanceUpdated(requestID, webhookData)
	case "maintenance.completed":
		return s.handleMaintenanceCompleted(requestID, webhookData)
	case "maintenance.assigned":
		return s.handleMaintenanceAssigned(requestID, webhookData)
	default:
		log.Printf("⚠️ Unhandled AppFolio maintenance event: %s", eventType)
	}

	return nil
}

func (s *AppFolioMaintenanceSync) handleMaintenanceCreated(data map[string]interface{}) error {
	afReq := AppFolioMaintenanceRequest{
		ID:            data["maintenance_request_id"].(string),
		PropertyID:    data["property_id"].(string),
		Category:      data["category"].(string),
		Priority:      data["priority"].(string),
		Description:   data["description"].(string),
		Status:        "open",
		RequestedDate: time.Now().Format("2006-01-02"),
	}

	if isEmergency, ok := data["is_emergency"].(bool); ok {
		afReq.IsEmergency = isEmergency
	}
	if tenantID, ok := data["tenant_id"].(string); ok {
		afReq.TenantID = tenantID
	}

	if syncErr := s.syncSingleRequest(afReq); syncErr != nil {
		return syncErr
	}
	return nil
}

func (s *AppFolioMaintenanceSync) handleMaintenanceUpdated(requestID string, data map[string]interface{}) error {
	updates := map[string]interface{}{
		"last_synced_at": time.Now(),
	}

	if status, ok := data["status"].(string); ok {
		updates["status"] = status
	}
	if priority, ok := data["priority"].(string); ok {
		updates["priority"] = priority
	}
	if notes, ok := data["notes"].(string); ok {
		updates["notes"] = notes
	}
	if scheduledDate, ok := data["scheduled_date"].(string); ok {
		if t := s.parseDatePtr(&scheduledDate); t != nil {
			updates["scheduled_date"] = t
		}
	}

	result := s.db.Model(&models.MaintenanceRequest{}).
		Where("appfolio_id = ?", requestID).
		Updates(updates)

	return result.Error
}

func (s *AppFolioMaintenanceSync) handleMaintenanceCompleted(requestID string, data map[string]interface{}) error {
	completedDate := time.Now()
	if dateStr, ok := data["completed_date"].(string); ok {
		if t := s.parseDatePtr(&dateStr); t != nil {
			completedDate = *t
		}
	}

	updates := map[string]interface{}{
		"status":         "completed",
		"completed_date": completedDate,
		"last_synced_at": time.Now(),
	}

	if actualCost, ok := data["actual_cost"].(float64); ok {
		updates["actual_cost"] = actualCost
	}

	result := s.db.Model(&models.MaintenanceRequest{}).
		Where("appfolio_id = ?", requestID).
		Updates(updates)

	return result.Error
}

func (s *AppFolioMaintenanceSync) handleMaintenanceAssigned(requestID string, data map[string]interface{}) error {
	updates := map[string]interface{}{
		"last_synced_at": time.Now(),
	}

	if assignedTo, ok := data["assigned_to"].(string); ok {
		updates["assigned_to"] = assignedTo
	}
	if vendorID, ok := data["vendor_id"].(string); ok {
		updates["vendor_id"] = vendorID
	}
	if scheduledDate, ok := data["scheduled_date"].(string); ok {
		if t := s.parseDatePtr(&scheduledDate); t != nil {
			updates["scheduled_date"] = t
		}
	}

	result := s.db.Model(&models.MaintenanceRequest{}).
		Where("appfolio_id = ?", requestID).
		Updates(updates)

	return result.Error
}

func (s *AppFolioMaintenanceSync) GetOpenMaintenanceRequests() ([]models.MaintenanceRequest, error) {
	var requests []models.MaintenanceRequest
	err := s.db.Where("status IN ?", []string{"open", "in_progress"}).
		Order("urgency_score DESC, requested_date ASC").
		Find(&requests).Error
	return requests, err
}

func (s *AppFolioMaintenanceSync) GetEmergencyRequests() ([]models.MaintenanceRequest, error) {
	var requests []models.MaintenanceRequest
	err := s.db.Where("is_emergency = ? OR triage_priority = ?", true, "emergency").
		Where("status NOT IN ?", []string{"completed", "cancelled"}).
		Order("requested_date ASC").
		Find(&requests).Error
	return requests, err
}

func (s *AppFolioMaintenanceSync) GetMaintenanceByProperty(propertyID string) ([]models.MaintenanceRequest, error) {
	var requests []models.MaintenanceRequest
	err := s.db.Where("property_id = ?", propertyID).
		Order("requested_date DESC").
		Find(&requests).Error
	return requests, err
}

func (s *AppFolioMaintenanceSync) GetMaintenanceStats() (*models.MaintenanceStats, error) {
	var stats models.MaintenanceStats

	s.db.Model(&models.MaintenanceRequest{}).
		Where("status IN ?", []string{"open", "in_progress"}).
		Count(&stats.OpenCount)

	s.db.Model(&models.MaintenanceRequest{}).
		Where("is_emergency = ? AND status NOT IN ?", true, []string{"completed", "cancelled"}).
		Count(&stats.EmergencyCount)

	s.db.Model(&models.MaintenanceRequest{}).
		Where("status = ? AND completed_date >= ?", "completed", time.Now().AddDate(0, -1, 0)).
		Count(&stats.CompletedThisMonth)

	var avgDays float64
	s.db.Model(&models.MaintenanceRequest{}).
		Select("AVG(EXTRACT(EPOCH FROM (completed_date - requested_date))/86400)").
		Where("status = ? AND completed_date IS NOT NULL", "completed").
		Scan(&avgDays)
	stats.AvgResolutionDays = avgDays

	return &stats, nil
}

func (s *AppFolioMaintenanceSync) parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (s *AppFolioMaintenanceSync) parseDatePtr(dateStr *string) *time.Time {
	if dateStr == nil || *dateStr == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return nil
	}
	return &t
}

func (s *AppFolioMaintenanceSync) GetLastSyncTime() time.Time {
	return s.lastSyncTime
}

func (s *AppFolioMaintenanceSync) GetSyncErrors() []models.SyncError {
	return s.syncErrors
}

func (s *AppFolioMaintenanceSync) addSyncError(operation, entityID, message string) {
	s.syncErrors = append(s.syncErrors, models.SyncError{
		Entity:    "maintenance",
		EntityID:  entityID,
		Operation: operation,
		Message:   message,
		Timestamp: time.Now(),
	})
}
