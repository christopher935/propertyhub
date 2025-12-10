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
	"chrisgross-ctrl-project/internal/security"
	"gorm.io/gorm"
)

type AppFolioTenantSync struct {
	db                *gorm.DB
	apiKey            string
	baseURL           string
	companyID         string
	encryptionManager *security.EncryptionManager
	lastSyncTime      time.Time
	syncErrors        []models.SyncError
}

type AppFolioTenant struct {
	ID            string  `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	PropertyID    string  `json:"property_id"`
	UnitID        string  `json:"unit_id"`
	LeaseStart    string  `json:"lease_start"`
	LeaseEnd      string  `json:"lease_end"`
	RentAmount    float64 `json:"rent_amount"`
	DepositAmount float64 `json:"deposit_amount"`
	Status        string  `json:"status"`
	MoveInDate    *string `json:"move_in_date"`
	MoveOutDate   *string `json:"move_out_date"`
	Balance       float64 `json:"balance"`
	LastPayment   *string `json:"last_payment_date"`
	IsActive      bool    `json:"is_active"`
}

type AppFolioTenantResponse struct {
	Tenants []AppFolioTenant `json:"tenants"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	PerPage int              `json:"per_page"`
	HasMore bool             `json:"has_more"`
}

func NewAppFolioTenantSync(db *gorm.DB, apiKey, companyID string, encMgr *security.EncryptionManager) *AppFolioTenantSync {
	return &AppFolioTenantSync{
		db:                db,
		apiKey:            apiKey,
		baseURL:           "https://api.appfolio.com/v1",
		companyID:         companyID,
		encryptionManager: encMgr,
		syncErrors:        make([]models.SyncError, 0),
	}
}

func (s *AppFolioTenantSync) SyncTenants() (*models.TenantSyncResult, error) {
	log.Println("👥 Starting AppFolio tenant sync...")
	startTime := time.Now()

	result := &models.TenantSyncResult{
		StartedAt: startTime,
		Source:    "appfolio",
	}

	page := 1
	perPage := 100
	hasMore := true

	for hasMore {
		tenants, resp, err := s.fetchTenantsPage(page, perPage)
		if err != nil {
			s.addSyncError("fetch_tenants", fmt.Sprintf("page_%d", page), err.Error())
			result.Errors = append(result.Errors, models.SyncError{
				Entity:    "tenant",
				Operation: "fetch",
				Message:   err.Error(),
				Timestamp: time.Now(),
			})
			break
		}

		for _, afTenant := range tenants {
			syncErr := s.syncSingleTenant(afTenant)
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

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(startTime)
	s.lastSyncTime = time.Now()

	log.Printf("✅ AppFolio tenant sync complete: %d synced, %d failed", result.Synced, result.Failed)
	return result, nil
}

func (s *AppFolioTenantSync) fetchTenantsPage(page, perPage int) ([]AppFolioTenant, *AppFolioTenantResponse, error) {
	url := fmt.Sprintf("%s/tenants?page=%d&per_page=%d&company_id=%s",
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
		return nil, nil, fmt.Errorf("failed to fetch tenants: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("AppFolio API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response AppFolioTenantResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Tenants, &response, nil
}

func (s *AppFolioTenantSync) syncSingleTenant(afTenant AppFolioTenant) *models.SyncError {
	var existingTenant models.AppFolioTenant
	err := s.db.Where("appfolio_id = ?", afTenant.ID).First(&existingTenant).Error

	encName := security.EncryptedString(fmt.Sprintf("%s %s", afTenant.FirstName, afTenant.LastName))
	encEmail := security.EncryptedString(afTenant.Email)
	encPhone := security.EncryptedString(afTenant.Phone)

	if err == gorm.ErrRecordNotFound {
		newTenant := models.AppFolioTenant{
			AppFolioID:       afTenant.ID,
			FirstName:        afTenant.FirstName,
			LastName:         afTenant.LastName,
			Email:            encEmail,
			Phone:            encPhone,
			PropertyID:       afTenant.PropertyID,
			UnitID:           afTenant.UnitID,
			LeaseStart:       s.parseDate(afTenant.LeaseStart),
			LeaseEnd:         s.parseDate(afTenant.LeaseEnd),
			RentAmount:       afTenant.RentAmount,
			DepositAmount:    afTenant.DepositAmount,
			Status:           afTenant.Status,
			MoveInDate:       s.parseDatePtr(afTenant.MoveInDate),
			MoveOutDate:      s.parseDatePtr(afTenant.MoveOutDate),
			Balance:          afTenant.Balance,
			IsActive:         afTenant.IsActive,
			LastSyncedAt:     time.Now(),
			AppFolioData: models.JSONB{
				"last_payment": afTenant.LastPayment,
			},
		}

		if err := s.db.Create(&newTenant).Error; err != nil {
			return &models.SyncError{
				Entity:      "tenant",
				EntityID:    afTenant.ID,
				Operation:   "create",
				Message:     err.Error(),
				Timestamp:   time.Now(),
				IsRetryable: true,
			}
		}

		s.linkTenantToClosingPipeline(&newTenant, encName, encEmail, encPhone)
		return nil
	} else if err != nil {
		return &models.SyncError{
			Entity:      "tenant",
			EntityID:    afTenant.ID,
			Operation:   "lookup",
			Message:     err.Error(),
			Timestamp:   time.Now(),
			IsRetryable: true,
		}
	}

	updates := map[string]interface{}{
		"status":         afTenant.Status,
		"balance":        afTenant.Balance,
		"is_active":      afTenant.IsActive,
		"rent_amount":    afTenant.RentAmount,
		"lease_end":      s.parseDate(afTenant.LeaseEnd),
		"move_out_date":  s.parseDatePtr(afTenant.MoveOutDate),
		"last_synced_at": time.Now(),
	}

	if err := s.db.Model(&existingTenant).Updates(updates).Error; err != nil {
		return &models.SyncError{
			Entity:      "tenant",
			EntityID:    afTenant.ID,
			Operation:   "update",
			Message:     err.Error(),
			Timestamp:   time.Now(),
			IsRetryable: true,
		}
	}

	return nil
}

func (s *AppFolioTenantSync) linkTenantToClosingPipeline(tenant *models.AppFolioTenant, name, email, phone security.EncryptedString) {
	var propState models.PropertyState
	err := s.db.Where("appfolio_id = ?", tenant.PropertyID).First(&propState).Error
	if err != nil {
		return
	}

	var pipeline models.ClosingPipeline
	err = s.db.Where("property_id = ? AND status != ?", propState.ID, "completed").First(&pipeline).Error
	if err == gorm.ErrRecordNotFound {
		return
	}

	updates := map[string]interface{}{
		"tenant_name":  name,
		"tenant_email": email,
		"tenant_phone": phone,
	}

	if tenant.MoveInDate != nil {
		updates["move_in_date"] = tenant.MoveInDate
	}
	if tenant.DepositAmount > 0 {
		updates["deposit_amount"] = tenant.DepositAmount
		updates["deposit_received"] = true
		updates["deposit_received_date"] = time.Now()
	}
	if tenant.RentAmount > 0 {
		updates["monthly_rent"] = tenant.RentAmount
	}

	s.db.Model(&pipeline).Updates(updates)
}

func (s *AppFolioTenantSync) CreateTenantInAppFolio(lead *models.Lead, booking *models.Booking, property *models.PropertyState) (*AppFolioTenant, error) {
	log.Printf("📤 Creating tenant in AppFolio from lead %d", lead.ID)

	payload := map[string]interface{}{
		"first_name":    lead.FirstName,
		"last_name":     lead.LastName,
		"email":         string(lead.Email),
		"phone":         string(lead.Phone),
		"property_id":   property.AppFolioID,
		"lease_start":   time.Now().Format("2006-01-02"),
		"rent_amount":   property.RentAmount,
		"status":        "pending",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tenant data: %w", err)
	}

	url := fmt.Sprintf("%s/tenants?company_id=%s", s.baseURL, s.companyID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AppFolio API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tenant AppFolioTenant
	if err := json.NewDecoder(resp.Body).Decode(&tenant); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("✅ Created tenant in AppFolio: %s", tenant.ID)
	return &tenant, nil
}

func (s *AppFolioTenantSync) UpdateTenantStatus(appfolioID, status string) error {
	url := fmt.Sprintf("%s/tenants/%s", s.baseURL, appfolioID)

	payload := map[string]interface{}{
		"status": status,
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

	return nil
}

func (s *AppFolioTenantSync) HandleTenantWebhook(webhookData map[string]interface{}) error {
	tenantID, ok := webhookData["tenant_id"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook: missing tenant_id")
	}

	eventType, _ := webhookData["event_type"].(string)

	switch eventType {
	case "tenant.created":
		return s.handleTenantCreated(webhookData)
	case "tenant.updated":
		return s.handleTenantUpdated(tenantID, webhookData)
	case "tenant.moved_out":
		return s.handleTenantMoveOut(tenantID, webhookData)
	case "payment.received":
		return s.handlePaymentReceived(tenantID, webhookData)
	default:
		log.Printf("⚠️ Unhandled AppFolio tenant event: %s", eventType)
	}

	return nil
}

func (s *AppFolioTenantSync) handleTenantCreated(data map[string]interface{}) error {
	log.Printf("👤 New tenant created in AppFolio: %v", data["tenant_id"])
	return nil
}

func (s *AppFolioTenantSync) handleTenantUpdated(tenantID string, data map[string]interface{}) error {
	updates := make(map[string]interface{})

	if status, ok := data["status"].(string); ok {
		updates["status"] = status
	}
	if balance, ok := data["balance"].(float64); ok {
		updates["balance"] = balance
	}

	updates["last_synced_at"] = time.Now()

	result := s.db.Model(&models.AppFolioTenant{}).
		Where("appfolio_id = ?", tenantID).
		Updates(updates)

	return result.Error
}

func (s *AppFolioTenantSync) handleTenantMoveOut(tenantID string, data map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":        "moved_out",
		"is_active":     false,
		"last_synced_at": time.Now(),
	}

	if moveOutDate, ok := data["move_out_date"].(string); ok {
		if t, err := time.Parse("2006-01-02", moveOutDate); err == nil {
			updates["move_out_date"] = t
		}
	}

	result := s.db.Model(&models.AppFolioTenant{}).
		Where("appfolio_id = ?", tenantID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	var tenant models.AppFolioTenant
	if err := s.db.Where("appfolio_id = ?", tenantID).First(&tenant).Error; err == nil {
		s.db.Model(&models.PropertyState{}).
			Where("appfolio_id = ?", tenant.PropertyID).
			Updates(map[string]interface{}{
				"is_vacant":         true,
				"status":            "vacant",
				"status_source":     "appfolio_webhook",
				"status_updated_at": time.Now(),
			})
	}

	return nil
}

func (s *AppFolioTenantSync) handlePaymentReceived(tenantID string, data map[string]interface{}) error {
	amount, _ := data["amount"].(float64)
	paymentDate, _ := data["payment_date"].(string)

	var tenant models.AppFolioTenant
	if err := s.db.Where("appfolio_id = ?", tenantID).First(&tenant).Error; err != nil {
		return err
	}

	var propState models.PropertyState
	if err := s.db.Where("appfolio_id = ?", tenant.PropertyID).First(&propState).Error; err != nil {
		return nil
	}

	var pipeline models.ClosingPipeline
	err := s.db.Where("property_id = ? AND status != ?", propState.ID, "completed").First(&pipeline).Error
	if err != nil {
		return nil
	}

	updates := make(map[string]interface{})

	if !pipeline.DepositReceived && amount >= tenant.DepositAmount {
		updates["deposit_received"] = true
		updates["deposit_received_date"] = s.parseDate(paymentDate)
		updates["deposit_amount"] = amount
	}

	if pipeline.DepositReceived && !pipeline.FirstMonthReceived {
		updates["first_month_received"] = true
		updates["first_month_received_date"] = s.parseDate(paymentDate)
	}

	if len(updates) > 0 {
		s.db.Model(&pipeline).Updates(updates)
	}

	return nil
}

func (s *AppFolioTenantSync) GetActiveTenants() ([]models.AppFolioTenant, error) {
	var tenants []models.AppFolioTenant
	err := s.db.Where("is_active = ?", true).Find(&tenants).Error
	return tenants, err
}

func (s *AppFolioTenantSync) GetTenantsByProperty(propertyID string) ([]models.AppFolioTenant, error) {
	var tenants []models.AppFolioTenant
	err := s.db.Where("property_id = ? AND is_active = ?", propertyID, true).Find(&tenants).Error
	return tenants, err
}

func (s *AppFolioTenantSync) parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (s *AppFolioTenantSync) parseDatePtr(dateStr *string) *time.Time {
	if dateStr == nil || *dateStr == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return nil
	}
	return &t
}

func (s *AppFolioTenantSync) GetLastSyncTime() time.Time {
	return s.lastSyncTime
}

func (s *AppFolioTenantSync) GetSyncErrors() []models.SyncError {
	return s.syncErrors
}

func (s *AppFolioTenantSync) addSyncError(operation, entityID, message string) {
	s.syncErrors = append(s.syncErrors, models.SyncError{
		Entity:    "tenant",
		EntityID:  entityID,
		Operation: operation,
		Message:   message,
		Timestamp: time.Now(),
	})
}
