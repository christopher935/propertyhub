package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

type AppFolioTenantSync struct {
	db             *gorm.DB
	appfolioClient *AppFolioAPIClient
	propertySync   *AppFolioPropertySync
}

type TenantSyncResult struct {
	Created       int                    `json:"created"`
	Updated       int                    `json:"updated"`
	Skipped       int                    `json:"skipped"`
	Errors        []string               `json:"errors"`
	SyncedAt      time.Time              `json:"synced_at"`
	Direction     string                 `json:"direction"`
	TenantsSynced []TenantSyncedRecord   `json:"tenants_synced"`
}

type TenantSyncedRecord struct {
	PropertyHubID    uint   `json:"propertyhub_id,omitempty"`
	AppFolioID       string `json:"appfolio_id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Action           string `json:"action"`
}

type TenantSyncLog struct {
	ID             uint            `json:"id" gorm:"primaryKey"`
	Direction      string          `json:"direction" gorm:"not null"`
	TenantID       string          `json:"tenant_id"`
	LeadID         *uint           `json:"lead_id"`
	BookingID      *uint           `json:"booking_id"`
	Action         string          `json:"action"`
	Status         string          `json:"status" gorm:"default:'pending'"`
	RequestPayload json.RawMessage `json:"request_payload" gorm:"type:jsonb"`
	ResponseData   json.RawMessage `json:"response_data" gorm:"type:jsonb"`
	ErrorMessage   string          `json:"error_message"`
	SyncedAt       time.Time       `json:"synced_at"`
	CreatedAt      time.Time       `json:"created_at"`
}

func NewAppFolioTenantSync(db *gorm.DB, client *AppFolioAPIClient, propertySync *AppFolioPropertySync) *AppFolioTenantSync {
	return &AppFolioTenantSync{
		db:             db,
		appfolioClient: client,
		propertySync:   propertySync,
	}
}

func (s *AppFolioTenantSync) PushTenantToAppFolio(booking models.Booking) (*AppFolioTenant, error) {
	syncLog := &TenantSyncLog{
		Direction: "push",
		BookingID: &booking.ID,
		Action:    "create_tenant",
		Status:    "pending",
		SyncedAt:  time.Now(),
	}
	defer s.saveSyncLog(syncLog)

	existingTenant, err := s.GetTenantByEmail(string(booking.Email))
	if err != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = fmt.Sprintf("failed to check existing tenant: %v", err)
		return nil, fmt.Errorf("failed to check existing tenant: %w", err)
	}

	if existingTenant != nil {
		syncLog.Status = "skipped"
		syncLog.TenantID = existingTenant.ID
		syncLog.ErrorMessage = "tenant already exists in AppFolio"
		log.Printf("⚠️ Tenant already exists in AppFolio: %s (ID: %s)", existingTenant.Email, existingTenant.ID)
		return existingTenant, nil
	}

	var property models.Property
	if booking.PropertyID > 0 {
		if err := s.db.First(&property, booking.PropertyID).Error; err != nil {
			syncLog.Status = "error"
			syncLog.ErrorMessage = fmt.Sprintf("failed to get property: %v", err)
			return nil, fmt.Errorf("failed to get property: %w", err)
		}
	}

	appfolioPropertyID := ""
	if property.ID > 0 {
		propID, err := s.propertySync.GetAppFolioPropertyID(property.ID)
		if err != nil {
			log.Printf("⚠️ Could not find AppFolio property ID: %v", err)
		} else {
			appfolioPropertyID = propID
		}
	}

	leaseStart := booking.ShowingDate
	leaseEnd := leaseStart.AddDate(1, 0, 0)

	createReq := &CreateTenantRequest{
		Name:       string(booking.Name),
		Email:      string(booking.Email),
		Phone:      string(booking.Phone),
		PropertyID: appfolioPropertyID,
		LeaseStart: leaseStart,
		LeaseEnd:   leaseEnd,
	}

	reqJSON, _ := json.Marshal(createReq)
	syncLog.RequestPayload = reqJSON

	tenant, err := s.appfolioClient.CreateTenant(createReq)
	if err != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = err.Error()
		return nil, fmt.Errorf("failed to create tenant in AppFolio: %w", err)
	}

	respJSON, _ := json.Marshal(tenant)
	syncLog.ResponseData = respJSON
	syncLog.Status = "success"
	syncLog.TenantID = tenant.ID

	if err := s.updateLeadWithAppFolioTenant(booking, tenant); err != nil {
		log.Printf("⚠️ Failed to update lead with AppFolio tenant ID: %v", err)
	}

	log.Printf("✅ Pushed tenant to AppFolio: %s (ID: %s)", tenant.Name, tenant.ID)
	return tenant, nil
}

func (s *AppFolioTenantSync) PushTenantFromApplication(appNumber models.ApplicationNumber) (*AppFolioTenant, error) {
	syncLog := &TenantSyncLog{
		Direction: "push",
		Action:    "create_tenant_from_application",
		Status:    "pending",
		SyncedAt:  time.Now(),
	}
	defer s.saveSyncLog(syncLog)

	var applicants []models.ApplicationApplicant
	if err := s.db.Where("application_number_id = ?", appNumber.ID).Find(&applicants).Error; err != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = fmt.Sprintf("failed to get applicants: %v", err)
		return nil, fmt.Errorf("failed to get applicants: %w", err)
	}

	if len(applicants) == 0 {
		syncLog.Status = "error"
		syncLog.ErrorMessage = "no applicants found for application"
		return nil, fmt.Errorf("no applicants found for application %d", appNumber.ID)
	}

	primaryApplicant := applicants[0]

	existingTenant, err := s.GetTenantByEmail(primaryApplicant.ApplicantEmail)
	if err != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = fmt.Sprintf("failed to check existing tenant: %v", err)
		return nil, fmt.Errorf("failed to check existing tenant: %w", err)
	}

	if existingTenant != nil {
		syncLog.Status = "skipped"
		syncLog.TenantID = existingTenant.ID
		log.Printf("⚠️ Tenant already exists in AppFolio: %s (ID: %s)", existingTenant.Email, existingTenant.ID)
		return existingTenant, nil
	}

	var propertyGroup models.PropertyApplicationGroup
	if err := s.db.First(&propertyGroup, appNumber.PropertyApplicationGroupID).Error; err != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = fmt.Sprintf("failed to get property group: %v", err)
		return nil, fmt.Errorf("failed to get property group: %w", err)
	}

	appfolioPropertyID := ""
	if propertyGroup.PropertyID > 0 {
		propID, err := s.propertySync.GetAppFolioPropertyID(propertyGroup.PropertyID)
		if err != nil {
			log.Printf("⚠️ Could not find AppFolio property ID: %v", err)
		} else {
			appfolioPropertyID = propID
		}
	}

	leaseStart := time.Now()
	leaseEnd := leaseStart.AddDate(1, 0, 0)

	createReq := &CreateTenantRequest{
		Name:       primaryApplicant.ApplicantName,
		Email:      primaryApplicant.ApplicantEmail,
		Phone:      primaryApplicant.ApplicantPhone,
		PropertyID: appfolioPropertyID,
		LeaseStart: leaseStart,
		LeaseEnd:   leaseEnd,
	}

	reqJSON, _ := json.Marshal(createReq)
	syncLog.RequestPayload = reqJSON

	tenant, err := s.appfolioClient.CreateTenant(createReq)
	if err != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = err.Error()
		return nil, fmt.Errorf("failed to create tenant in AppFolio: %w", err)
	}

	respJSON, _ := json.Marshal(tenant)
	syncLog.ResponseData = respJSON
	syncLog.Status = "success"
	syncLog.TenantID = tenant.ID

	if err := s.updateApplicantWithAppFolioTenant(primaryApplicant, tenant); err != nil {
		log.Printf("⚠️ Failed to update applicant with AppFolio tenant ID: %v", err)
	}

	log.Printf("✅ Pushed tenant to AppFolio from application: %s (ID: %s)", tenant.Name, tenant.ID)
	return tenant, nil
}

func (s *AppFolioTenantSync) SyncTenantsFromAppFolio() (*TenantSyncResult, error) {
	result := &TenantSyncResult{
		SyncedAt:      time.Now(),
		Direction:     "pull",
		Errors:        make([]string, 0),
		TenantsSynced: make([]TenantSyncedRecord, 0),
	}

	syncLog := &TenantSyncLog{
		Direction: "pull",
		Action:    "full_sync",
		Status:    "pending",
		SyncedAt:  time.Now(),
	}
	defer s.saveSyncLog(syncLog)

	page := 1
	perPage := 100

	for {
		tenantsResp, err := s.appfolioClient.GetTenants(page, perPage)
		if err != nil {
			errMsg := fmt.Sprintf("failed to fetch page %d: %v", page, err)
			result.Errors = append(result.Errors, errMsg)
			break
		}

		if len(tenantsResp.Tenants) == 0 {
			break
		}

		for _, afTenant := range tenantsResp.Tenants {
			syncErr := s.syncSingleTenantToPropertyHub(afTenant, result)
			if syncErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("tenant %s: %v", afTenant.ID, syncErr))
			}
		}

		if len(tenantsResp.Tenants) < perPage {
			break
		}
		page++
	}

	if len(result.Errors) > 0 {
		syncLog.Status = "partial"
		syncLog.ErrorMessage = fmt.Sprintf("%d errors occurred", len(result.Errors))
	} else {
		syncLog.Status = "success"
	}

	respJSON, _ := json.Marshal(result)
	syncLog.ResponseData = respJSON

	log.Printf("✅ AppFolio tenant sync completed: %d created, %d updated, %d skipped, %d errors",
		result.Created, result.Updated, result.Skipped, len(result.Errors))

	return result, nil
}

func (s *AppFolioTenantSync) syncSingleTenantToPropertyHub(afTenant AppFolioTenant, result *TenantSyncResult) error {
	var existingLead models.Lead
	err := s.db.Where("email = ? OR appfolio_tenant_id = ?", afTenant.Email, afTenant.ID).First(&existingLead).Error

	if err == gorm.ErrRecordNotFound {
		newLead := models.Lead{
			FirstName:        extractFirstName(afTenant.Name),
			LastName:         extractLastName(afTenant.Name),
			Email:            afTenant.Email,
			Phone:            afTenant.Phone,
			Status:           mapAppFolioTenantStatusToLeadStatus(afTenant.Status),
			Source:           "AppFolio",
			AppFolioTenantID: afTenant.ID,
			TenantStatus:     afTenant.Status,
			LeaseStart:       &afTenant.LeaseStart,
			LeaseEnd:         &afTenant.LeaseEnd,
			AppFolioUnitID:   afTenant.UnitID,
		}

		if err := s.db.Create(&newLead).Error; err != nil {
			return fmt.Errorf("failed to create lead: %w", err)
		}

		result.Created++
		result.TenantsSynced = append(result.TenantsSynced, TenantSyncedRecord{
			PropertyHubID: newLead.ID,
			AppFolioID:    afTenant.ID,
			Name:          afTenant.Name,
			Email:         afTenant.Email,
			Action:        "created",
		})

		log.Printf("👤 Created lead from AppFolio tenant: %s", afTenant.Name)
	} else if err != nil {
		return fmt.Errorf("failed to query lead: %w", err)
	} else {
		existingLead.AppFolioTenantID = afTenant.ID
		existingLead.TenantStatus = afTenant.Status
		existingLead.LeaseStart = &afTenant.LeaseStart
		existingLead.LeaseEnd = &afTenant.LeaseEnd
		existingLead.AppFolioUnitID = afTenant.UnitID
		existingLead.Phone = afTenant.Phone

		if err := s.db.Save(&existingLead).Error; err != nil {
			return fmt.Errorf("failed to update lead: %w", err)
		}

		result.Updated++
		result.TenantsSynced = append(result.TenantsSynced, TenantSyncedRecord{
			PropertyHubID: existingLead.ID,
			AppFolioID:    afTenant.ID,
			Name:          afTenant.Name,
			Email:         afTenant.Email,
			Action:        "updated",
		})

		log.Printf("👤 Updated lead from AppFolio tenant: %s", afTenant.Name)
	}

	return nil
}

func (s *AppFolioTenantSync) GetTenantByEmail(email string) (*AppFolioTenant, error) {
	return s.appfolioClient.GetTenantByEmail(email)
}

func (s *AppFolioTenantSync) GetTenantByID(tenantID string) (*AppFolioTenant, error) {
	return s.appfolioClient.GetTenant(tenantID)
}

func (s *AppFolioTenantSync) UpdateTenantStatus(tenantID string, status string) error {
	syncLog := &TenantSyncLog{
		Direction: "push",
		TenantID:  tenantID,
		Action:    "update_status",
		Status:    "pending",
		SyncedAt:  time.Now(),
	}
	defer s.saveSyncLog(syncLog)

	reqPayload := map[string]string{"status": status}
	reqJSON, _ := json.Marshal(reqPayload)
	syncLog.RequestPayload = reqJSON

	if err := s.appfolioClient.UpdateTenantStatus(tenantID, status); err != nil {
		syncLog.Status = "error"
		syncLog.ErrorMessage = err.Error()
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	syncLog.Status = "success"
	log.Printf("✅ Updated tenant status in AppFolio: %s -> %s", tenantID, status)
	return nil
}

func (s *AppFolioTenantSync) SyncTenantStatusToAppFolio(leadID uint) error {
	var lead models.Lead
	if err := s.db.First(&lead, leadID).Error; err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	if lead.AppFolioTenantID == "" {
		return fmt.Errorf("lead %d has no AppFolio tenant ID", leadID)
	}

	afStatus := mapLeadStatusToAppFolioTenantStatus(lead.TenantStatus)
	return s.UpdateTenantStatus(lead.AppFolioTenantID, afStatus)
}

func (s *AppFolioTenantSync) GetSyncLogs(limit int) ([]TenantSyncLog, error) {
	var logs []TenantSyncLog
	err := s.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (s *AppFolioTenantSync) GetSyncLogsByDirection(direction string, limit int) ([]TenantSyncLog, error) {
	var logs []TenantSyncLog
	err := s.db.Where("direction = ?", direction).Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (s *AppFolioTenantSync) updateLeadWithAppFolioTenant(booking models.Booking, tenant *AppFolioTenant) error {
	var lead models.Lead
	err := s.db.Where("email = ? OR fub_lead_id = ?", string(booking.Email), booking.FUBLeadID).First(&lead).Error

	if err == gorm.ErrRecordNotFound {
		lead = models.Lead{
			FirstName:        extractFirstName(string(booking.Name)),
			LastName:         extractLastName(string(booking.Name)),
			Email:            string(booking.Email),
			Phone:            string(booking.Phone),
			FUBLeadID:        booking.FUBLeadID,
			Source:           "Booking",
			Status:           "tenant",
			AppFolioTenantID: tenant.ID,
			TenantStatus:     tenant.Status,
			LeaseStart:       &tenant.LeaseStart,
			LeaseEnd:         &tenant.LeaseEnd,
			AppFolioUnitID:   tenant.UnitID,
		}
		return s.db.Create(&lead).Error
	} else if err != nil {
		return err
	}

	lead.AppFolioTenantID = tenant.ID
	lead.TenantStatus = tenant.Status
	lead.LeaseStart = &tenant.LeaseStart
	lead.LeaseEnd = &tenant.LeaseEnd
	lead.AppFolioUnitID = tenant.UnitID
	lead.Status = "tenant"

	return s.db.Save(&lead).Error
}

func (s *AppFolioTenantSync) updateApplicantWithAppFolioTenant(applicant models.ApplicationApplicant, tenant *AppFolioTenant) error {
	var lead models.Lead
	err := s.db.Where("email = ?", applicant.ApplicantEmail).First(&lead).Error

	if err == gorm.ErrRecordNotFound {
		lead = models.Lead{
			FirstName:        extractFirstName(applicant.ApplicantName),
			LastName:         extractLastName(applicant.ApplicantName),
			Email:            applicant.ApplicantEmail,
			Phone:            applicant.ApplicantPhone,
			FUBLeadID:        applicant.FUBLeadID,
			Source:           "Application",
			Status:           "tenant",
			AppFolioTenantID: tenant.ID,
			TenantStatus:     tenant.Status,
			LeaseStart:       &tenant.LeaseStart,
			LeaseEnd:         &tenant.LeaseEnd,
			AppFolioUnitID:   tenant.UnitID,
		}
		return s.db.Create(&lead).Error
	} else if err != nil {
		return err
	}

	lead.AppFolioTenantID = tenant.ID
	lead.TenantStatus = tenant.Status
	lead.LeaseStart = &tenant.LeaseStart
	lead.LeaseEnd = &tenant.LeaseEnd
	lead.AppFolioUnitID = tenant.UnitID
	lead.Status = "tenant"

	return s.db.Save(&lead).Error
}

func (s *AppFolioTenantSync) saveSyncLog(log *TenantSyncLog) {
	if err := s.db.Create(log).Error; err != nil {
		fmt.Printf("⚠️ Failed to save sync log: %v\n", err)
	}
}

func extractFirstName(fullName string) string {
	if fullName == "" {
		return ""
	}
	for i, c := range fullName {
		if c == ' ' {
			return fullName[:i]
		}
	}
	return fullName
}

func extractLastName(fullName string) string {
	if fullName == "" {
		return ""
	}
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == ' ' {
			return fullName[i+1:]
		}
	}
	return ""
}

func mapAppFolioTenantStatusToLeadStatus(afStatus string) string {
	switch afStatus {
	case "prospect":
		return "new"
	case "applicant":
		return "qualified"
	case "tenant", "current":
		return "tenant"
	case "former", "past":
		return "closed"
	case "eviction":
		return "closed"
	default:
		return "new"
	}
}

func mapLeadStatusToAppFolioTenantStatus(leadStatus string) string {
	switch leadStatus {
	case "new", "prospect":
		return "prospect"
	case "qualified", "applicant":
		return "applicant"
	case "tenant", "active":
		return "current"
	case "closed", "former":
		return "former"
	default:
		return "prospect"
	}
}
