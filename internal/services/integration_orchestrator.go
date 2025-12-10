package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

type IntegrationOrchestrator struct {
	db                    *gorm.DB
	fubSync               *FUBBidirectionalSync
	appfolioPropertySync  *AppFolioPropertySync
	appfolioTenantSync    *AppFolioTenantSync
	appfolioMaintenance   *AppFolioMaintenanceSync
	behavioralService     *BehavioralEventService
	scoringEngine         *BehavioralScoringEngine
	notificationService   *NotificationService
	propertyMatcher       *PropertyMatchingService
	syncQueue             chan *models.SyncQueueItem
	eventChan             chan *models.IntegrationEvent
	isRunning             bool
	mutex                 sync.RWMutex
	lastFullSync          *time.Time
	lastFUBSync           *time.Time
	lastAppFolioPropertySync *time.Time
	lastAppFolioTenantSync   *time.Time
	lastAppFolioMaintenanceSync *time.Time
}

func NewIntegrationOrchestrator(
	db *gorm.DB,
	fubSync *FUBBidirectionalSync,
	appfolioPropertySync *AppFolioPropertySync,
	appfolioTenantSync *AppFolioTenantSync,
	appfolioMaintenance *AppFolioMaintenanceSync,
	behavioralService *BehavioralEventService,
	scoringEngine *BehavioralScoringEngine,
	notificationService *NotificationService,
	propertyMatcher *PropertyMatchingService,
) *IntegrationOrchestrator {
	log.Println("🔄 Initializing Integration Orchestrator (3-Way Sync)...")

	orchestrator := &IntegrationOrchestrator{
		db:                   db,
		fubSync:              fubSync,
		appfolioPropertySync: appfolioPropertySync,
		appfolioTenantSync:   appfolioTenantSync,
		appfolioMaintenance:  appfolioMaintenance,
		behavioralService:    behavioralService,
		scoringEngine:        scoringEngine,
		notificationService:  notificationService,
		propertyMatcher:      propertyMatcher,
		syncQueue:            make(chan *models.SyncQueueItem, 1000),
		eventChan:            make(chan *models.IntegrationEvent, 500),
	}

	log.Println("✅ Integration Orchestrator initialized")
	return orchestrator
}

func (o *IntegrationOrchestrator) Start() {
	o.mutex.Lock()
	if o.isRunning {
		o.mutex.Unlock()
		return
	}
	o.isRunning = true
	o.mutex.Unlock()

	log.Println("🚀 Starting Integration Orchestrator background processes...")

	go o.processEventQueue()
	go o.processSyncQueue()
	go o.runReconciliationLoop()

	log.Println("✅ Integration Orchestrator running")
}

func (o *IntegrationOrchestrator) Stop() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.isRunning = false
	log.Println("🛑 Integration Orchestrator stopped")
}

func (o *IntegrationOrchestrator) RunFullSync() (*models.SyncReport, error) {
	log.Println("🔄 ═══════════════════════════════════════════════════════")
	log.Println("🔄 Starting Full 3-Way Integration Sync")
	log.Println("🔄 ═══════════════════════════════════════════════════════")

	startTime := time.Now()
	report := &models.SyncReport{
		StartedAt:   startTime,
		SyncType:    "full",
		TriggeredBy: "orchestrator",
		Status:      models.SyncStatusInProgress,
	}

	if err := o.db.Create(report).Error; err != nil {
		log.Printf("⚠️ Failed to create sync report: %v", err)
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	allErrors := make([]models.SyncError, 0)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if o.appfolioPropertySync != nil {
			log.Println("📊 Step 1: Syncing properties from AppFolio...")
			result, err := o.appfolioPropertySync.SyncProperties()
			mutex.Lock()
			if err != nil {
				allErrors = append(allErrors, models.SyncError{
					Entity:    "property_sync",
					Operation: "appfolio_sync",
					Message:   err.Error(),
					Timestamp: time.Now(),
				})
			} else {
				report.PropertiesSynced = result.Synced
				report.VacanciesUpdated = result.VacanciesUpdated
				allErrors = append(allErrors, result.Errors...)
				now := time.Now()
				o.lastAppFolioPropertySync = &now
			}
			mutex.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if o.appfolioTenantSync != nil {
			log.Println("👥 Step 2: Syncing tenants from AppFolio...")
			result, err := o.appfolioTenantSync.SyncTenants()
			mutex.Lock()
			if err != nil {
				allErrors = append(allErrors, models.SyncError{
					Entity:    "tenant_sync",
					Operation: "appfolio_sync",
					Message:   err.Error(),
					Timestamp: time.Now(),
				})
			} else {
				report.TenantsSynced = result.Synced
				allErrors = append(allErrors, result.Errors...)
				now := time.Now()
				o.lastAppFolioTenantSync = &now
			}
			mutex.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if o.appfolioMaintenance != nil {
			log.Println("🔧 Step 3: Syncing maintenance from AppFolio...")
			result, err := o.appfolioMaintenance.SyncMaintenanceRequests()
			mutex.Lock()
			if err != nil {
				allErrors = append(allErrors, models.SyncError{
					Entity:    "maintenance_sync",
					Operation: "appfolio_sync",
					Message:   err.Error(),
					Timestamp: time.Now(),
				})
			} else {
				report.MaintenanceSynced = result.Synced
				allErrors = append(allErrors, result.Errors...)
				now := time.Now()
				o.lastAppFolioMaintenanceSync = &now
			}
			mutex.Unlock()
		}
	}()

	wg.Wait()

	log.Println("📞 Step 4: Syncing leads with FUB...")
	leadsSynced, leadErrors := o.syncLeadsWithFUB()
	report.LeadsSynced = leadsSynced
	allErrors = append(allErrors, leadErrors...)
	now := time.Now()
	o.lastFUBSync = &now

	log.Println("🔗 Step 5: Running cross-system reconciliation...")
	reconcileErrors := o.runCrossSystemReconciliation()
	allErrors = append(allErrors, reconcileErrors...)

	report.CompletedAt = time.Now()
	report.DurationSeconds = report.CompletedAt.Sub(startTime).Seconds()
	report.Errors = allErrors
	report.ErrorCount = len(allErrors)
	report.FUBLastSync = o.lastFUBSync
	report.AppFolioLastSync = o.lastAppFolioPropertySync

	if report.ErrorCount == 0 {
		report.Status = models.SyncStatusSuccess
	} else if report.PropertiesSynced > 0 || report.TenantsSynced > 0 || report.LeadsSynced > 0 {
		report.Status = models.SyncStatusPartial
	} else {
		report.Status = models.SyncStatusFailed
	}

	o.db.Save(report)
	o.lastFullSync = &now

	log.Println("🔄 ═══════════════════════════════════════════════════════")
	log.Printf("🔄 Full Sync Complete: Properties=%d, Tenants=%d, Leads=%d, Maintenance=%d, Errors=%d",
		report.PropertiesSynced, report.TenantsSynced, report.LeadsSynced, report.MaintenanceSynced, report.ErrorCount)
	log.Printf("🔄 Duration: %.2f seconds, Status: %s", report.DurationSeconds, report.Status)
	log.Println("🔄 ═══════════════════════════════════════════════════════")

	return report, nil
}

func (o *IntegrationOrchestrator) HandleNewLead(lead models.Lead) error {
	log.Printf("🆕 Processing new lead: %s %s (ID: %d)", lead.FirstName, lead.LastName, lead.ID)

	if o.behavioralService != nil {
		o.behavioralService.TrackEvent(int64(lead.ID), "lead_created", map[string]interface{}{
			"source": lead.Source,
			"email":  string(lead.Email),
		}, nil, "", "", "")
	}

	if o.scoringEngine != nil {
		go func() {
			_, err := o.scoringEngine.CalculateScore(int64(lead.ID))
			if err != nil {
				log.Printf("⚠️ Failed to calculate initial score for lead %d: %v", lead.ID, err)
			}
		}()
	}

	if o.fubSync != nil && lead.FUBLeadID == "" {
		go func() {
			err := o.createLeadInFUB(&lead)
			if err != nil {
				log.Printf("⚠️ Failed to create lead %d in FUB: %v", lead.ID, err)
				o.queueForRetry("lead", fmt.Sprintf("%d", lead.ID), "create_fub", err.Error())
			}
		}()
	}

	o.emitEvent(models.EventNewLead, models.SourcePropertyHub, "lead", fmt.Sprintf("%d", lead.ID), map[string]interface{}{
		"first_name": lead.FirstName,
		"last_name":  lead.LastName,
		"email":      string(lead.Email),
		"source":     lead.Source,
	})

	return nil
}

func (o *IntegrationOrchestrator) HandleBookingCreated(booking *models.Booking) error {
	log.Printf("📅 Processing new booking: %s for property %d", booking.ReferenceNumber, booking.PropertyID)

	if o.behavioralService != nil {
		leadID := o.getLeadIDFromFUBID(booking.FUBLeadID)
		if leadID > 0 {
			propID := int64(booking.PropertyID)
			o.behavioralService.TrackEvent(leadID, "showing_scheduled", map[string]interface{}{
				"booking_id":   booking.ID,
				"showing_date": booking.ShowingDate,
				"showing_type": booking.ShowingType,
			}, &propID, "", "", "")
		}
	}

	if o.fubSync != nil && booking.FUBLeadID != "" {
		go func() {
			leadID := o.getLeadIDFromFUBID(booking.FUBLeadID)
			if leadID > 0 {
				err := o.fubSync.ScheduleShowingInFUB(leadID, int64(booking.PropertyID), booking.ShowingDate, "")
				if err != nil {
					log.Printf("⚠️ Failed to sync booking to FUB: %v", err)
				}
			}
		}()
	}

	o.emitEvent(models.EventBookingCreated, models.SourcePropertyHub, "booking", fmt.Sprintf("%d", booking.ID), map[string]interface{}{
		"property_id":  booking.PropertyID,
		"fub_lead_id":  booking.FUBLeadID,
		"showing_date": booking.ShowingDate,
	})

	return nil
}

func (o *IntegrationOrchestrator) HandleLeaseConversion(booking models.Booking, closingPipeline *models.ClosingPipeline) error {
	log.Printf("🎉 Processing lease conversion for booking %s", booking.ReferenceNumber)

	var lead models.Lead
	if err := o.db.Where("fub_lead_id = ?", booking.FUBLeadID).First(&lead).Error; err != nil {
		log.Printf("⚠️ Lead not found for FUB ID %s", booking.FUBLeadID)
	}

	var propertyState models.PropertyState
	if err := o.db.Where("id = ?", booking.PropertyID).First(&propertyState).Error; err == nil {
		if o.appfolioTenantSync != nil && propertyState.AppFolioID != "" {
			go func() {
				tenant, err := o.appfolioTenantSync.CreateTenantInAppFolio(&lead, &booking, &propertyState)
				if err != nil {
					log.Printf("⚠️ Failed to create tenant in AppFolio: %v", err)
					o.queueForRetry("lease_conversion", fmt.Sprintf("%d", booking.ID), "create_tenant_appfolio", err.Error())
				} else {
					log.Printf("✅ Created tenant in AppFolio: %s", tenant.ID)
				}
			}()
		}
	}

	if o.fubSync != nil {
		go func() {
			leadID := o.getLeadIDFromFUBID(booking.FUBLeadID)
			if leadID > 0 {
				err := o.fubSync.UpdateLeadStatusInFUB(leadID, "Tenant")
				if err != nil {
					log.Printf("⚠️ Failed to update lead status in FUB: %v", err)
				}

				err = o.fubSync.AddNoteToFUB(leadID, fmt.Sprintf(
					"Lease conversion completed. Property: %s. Move-in date: %v",
					booking.PropertyAddress, closingPipeline.MoveInDate,
				), "")
				if err != nil {
					log.Printf("⚠️ Failed to add conversion note to FUB: %v", err)
				}
			}
		}()
	}

	go func() {
		o.db.Model(&models.PropertyState{}).
			Where("id = ?", booking.PropertyID).
			Updates(map[string]interface{}{
				"status":            "occupied",
				"status_source":     "lease_conversion",
				"status_updated_at": time.Now(),
				"is_vacant":         false,
				"is_bookable":       false,
			})
	}()

	if o.behavioralService != nil {
		leadID := o.getLeadIDFromFUBID(booking.FUBLeadID)
		if leadID > 0 {
			propID := int64(booking.PropertyID)
			o.behavioralService.TrackEvent(leadID, "converted", map[string]interface{}{
				"conversion_type": "lease",
				"property_id":     booking.PropertyID,
			}, &propID, "", "", "")
		}
	}

	o.emitEvent(models.EventLeaseConversion, models.SourcePropertyHub, "booking", fmt.Sprintf("%d", booking.ID), map[string]interface{}{
		"property_id":    booking.PropertyID,
		"fub_lead_id":    booking.FUBLeadID,
		"pipeline_id":    closingPipeline.ID,
	})

	return nil
}

func (o *IntegrationOrchestrator) HandleMaintenanceRequest(request models.MaintenanceRequest) error {
	log.Printf("🔧 Processing maintenance request: %s (Category: %s, Priority: %s)",
		request.AppFolioID, request.Category, request.Priority)

	if request.IsEmergency || request.TriagePriority == "emergency" {
		if o.notificationService != nil {
			o.notificationService.SendAgentAlert("admin", "Emergency Maintenance", fmt.Sprintf(
				"🚨 %s at property %s - %s",
				request.Category, request.PropertyID, request.Description,
			), map[string]interface{}{"property_id": request.PropertyID, "is_emergency": true})
		}
	}

	var propertyState models.PropertyState
	if err := o.db.Where("appfolio_id = ?", request.PropertyID).First(&propertyState).Error; err == nil {
		if o.notificationService != nil {
			o.notificationService.SendAgentAlert("owner", "Maintenance Request", fmt.Sprintf(
				"%s (%s priority) at %s - %s",
				request.Category, request.Priority, propertyState.Address, request.Description,
			), map[string]interface{}{"property_id": propertyState.AppFolioID, "priority": request.Priority})
		}
	}

	o.emitEvent(models.EventMaintenanceCreated, models.SourceAppFolio, "maintenance", request.AppFolioID, map[string]interface{}{
		"property_id":  request.PropertyID,
		"category":     request.Category,
		"priority":     request.Priority,
		"is_emergency": request.IsEmergency,
	})

	return nil
}

func (o *IntegrationOrchestrator) HandlePropertyVacancy(propertyID string) error {
	log.Printf("🏠 Processing property vacancy: %s", propertyID)

	o.db.Model(&models.PropertyState{}).
		Where("appfolio_id = ?", propertyID).
		Updates(map[string]interface{}{
			"status":            "vacant",
			"status_source":     "appfolio",
			"status_updated_at": time.Now(),
			"is_vacant":         true,
			"is_bookable":       true,
		})

	var propertyState models.PropertyState
	if err := o.db.Where("appfolio_id = ?", propertyID).First(&propertyState).Error; err == nil {
		if o.propertyMatcher != nil {
			go func() {
				var leads []models.Lead
				o.db.Limit(50).Find(&leads)

				notifiedCount := 0
				for _, lead := range leads {
					matches, err := o.propertyMatcher.FindMatchesForLead(int64(lead.ID))
					if err != nil {
						continue
					}
					for _, match := range matches {
						if match.PropertyID == int64(propertyState.ID) && o.fubSync != nil {
							note := fmt.Sprintf(
								"New vacancy alert! Property at %s is now available. This matches your search criteria.",
								propertyState.Address,
							)
							o.fubSync.AddNoteToFUB(int64(lead.ID), note, "")
							notifiedCount++
							break
						}
					}
				}

				log.Printf("📧 Notified %d leads about vacancy at %s", notifiedCount, propertyState.Address)
			}()
		}
	}

	o.emitEvent(models.EventPropertyVacancy, models.SourceAppFolio, "property", propertyID, map[string]interface{}{
		"status": "vacant",
	})

	return nil
}

func (o *IntegrationOrchestrator) HandleApplicationApproved(lead *models.Lead, propertyID uint) error {
	log.Printf("✅ Processing approved application for lead %d on property %d", lead.ID, propertyID)

	var propertyState models.PropertyState
	if err := o.db.First(&propertyState, propertyID).Error; err != nil {
		return fmt.Errorf("property not found: %w", err)
	}

	if o.appfolioTenantSync != nil && propertyState.AppFolioID != "" {
		tenant, err := o.appfolioTenantSync.CreateTenantInAppFolio(lead, nil, &propertyState)
		if err != nil {
			log.Printf("⚠️ Failed to create tenant in AppFolio: %v", err)
			o.queueForRetry("application", fmt.Sprintf("%d", lead.ID), "create_tenant_appfolio", err.Error())
		} else {
			log.Printf("✅ Created tenant in AppFolio: %s", tenant.ID)
		}
	}

	if o.fubSync != nil {
		err := o.fubSync.UpdateLeadStatusInFUB(int64(lead.ID), "Application Approved")
		if err != nil {
			log.Printf("⚠️ Failed to update FUB status: %v", err)
		}

		o.fubSync.AddNoteToFUB(int64(lead.ID), fmt.Sprintf(
			"Application approved for property: %s",
			propertyState.Address,
		), "")
	}

	o.emitEvent(models.EventApplicationApproved, models.SourcePropertyHub, "lead", fmt.Sprintf("%d", lead.ID), map[string]interface{}{
		"property_id": propertyID,
	})

	return nil
}

func (o *IntegrationOrchestrator) GetUnifiedDashboard() (*models.UnifiedDashboard, error) {
	log.Println("📊 Generating unified dashboard...")

	dashboard := &models.UnifiedDashboard{
		GeneratedAt: time.Now(),
	}

	var propertyStats models.PropertyStats
	o.db.Model(&models.PropertyState{}).Count(&propertyStats.Total)
	o.db.Model(&models.PropertyState{}).Where("is_vacant = ?", true).Count(&propertyStats.Vacant)
	o.db.Model(&models.PropertyState{}).Where("is_vacant = ? OR status = ?", false, "occupied").Count(&propertyStats.Occupied)
	o.db.Model(&models.PropertyState{}).Where("is_bookable = ?", true).Count(&propertyStats.Listed)
	propertyStats.Source = "appfolio"
	dashboard.Properties = propertyStats

	var leadStats models.LeadStats
	o.db.Model(&models.Lead{}).Count(&leadStats.Total)

	var hotLeads, warmLeads, coldLeads int64
	o.db.Table("behavioral_scores").
		Joins("JOIN leads ON leads.id = behavioral_scores.lead_id").
		Where("behavioral_scores.composite_score >= 80").
		Count(&hotLeads)
	o.db.Table("behavioral_scores").
		Joins("JOIN leads ON leads.id = behavioral_scores.lead_id").
		Where("behavioral_scores.composite_score >= 50 AND behavioral_scores.composite_score < 80").
		Count(&warmLeads)
	o.db.Table("behavioral_scores").
		Joins("JOIN leads ON leads.id = behavioral_scores.lead_id").
		Where("behavioral_scores.composite_score < 50").
		Count(&coldLeads)

	leadStats.Hot = hotLeads
	leadStats.Warm = warmLeads
	leadStats.Cold = coldLeads

	today := time.Now().Truncate(24 * time.Hour)
	o.db.Model(&models.Lead{}).Where("created_at >= ?", today).Count(&leadStats.NewToday)
	leadStats.Source = "propertyhub+fub"
	dashboard.Leads = leadStats

	if o.appfolioMaintenance != nil {
		maintenanceStats, err := o.appfolioMaintenance.GetMaintenanceStats()
		if err == nil {
			dashboard.Maintenance = *maintenanceStats
		}
	}
	dashboard.Maintenance.Source = "appfolio"

	var revenueStats models.RevenueStats
	var totalRent float64
	o.db.Model(&models.AppFolioTenant{}).
		Where("is_active = ?", true).
		Select("COALESCE(SUM(rent_amount), 0)").
		Scan(&totalRent)
	revenueStats.ProjectedMonth = totalRent

	var collectedRent float64
	o.db.Model(&models.ClosingPipeline{}).
		Where("first_month_received = ?", true).
		Where("created_at >= ?", time.Now().AddDate(0, -1, 0)).
		Select("COALESCE(SUM(monthly_rent), 0)").
		Scan(&collectedRent)
	revenueStats.Collected = collectedRent

	var pendingRent float64
	o.db.Model(&models.AppFolioTenant{}).
		Where("is_active = ? AND balance > 0", true).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&pendingRent)
	revenueStats.Pending = pendingRent
	revenueStats.Source = "appfolio"
	dashboard.Revenue = revenueStats

	dashboard.LastSync = models.LastSyncInfo{
		FUB:                 o.lastFUBSync,
		AppFolioProperty:    o.lastAppFolioPropertySync,
		AppFolioTenant:      o.lastAppFolioTenantSync,
		AppFolioMaintenance: o.lastAppFolioMaintenanceSync,
		FullSync:            o.lastFullSync,
	}

	var queuedCount, failedCount int64
	o.db.Model(&models.SyncQueueItem{}).Where("status = ?", "pending").Count(&queuedCount)
	o.db.Model(&models.SyncQueueItem{}).Where("status = ?", "failed").Count(&failedCount)

	dashboard.SystemHealth = models.SystemHealthStats{
		FUBConnected:      o.fubSync != nil,
		AppFolioConnected: o.appfolioPropertySync != nil,
		QueuedSyncItems:   queuedCount,
		FailedSyncItems:   failedCount,
	}

	return dashboard, nil
}

func (o *IntegrationOrchestrator) SyncPropertiesFromAppFolio() (*models.PropertySyncResult, error) {
	if o.appfolioPropertySync == nil {
		return nil, fmt.Errorf("AppFolio property sync not configured")
	}

	result, err := o.appfolioPropertySync.SyncProperties()
	if err == nil {
		now := time.Now()
		o.lastAppFolioPropertySync = &now
	}
	return result, err
}

func (o *IntegrationOrchestrator) SyncMaintenanceFromAppFolio() (*models.MaintenanceSyncResult, error) {
	if o.appfolioMaintenance == nil {
		return nil, fmt.Errorf("AppFolio maintenance sync not configured")
	}

	result, err := o.appfolioMaintenance.SyncMaintenanceRequests()
	if err == nil {
		now := time.Now()
		o.lastAppFolioMaintenanceSync = &now
	}
	return result, err
}

func (o *IntegrationOrchestrator) SyncLeadsWithFUB() (int, error) {
	synced, errors := o.syncLeadsWithFUB()
	if len(errors) > 0 {
		return synced, fmt.Errorf("sync completed with %d errors", len(errors))
	}
	return synced, nil
}

func (o *IntegrationOrchestrator) RetryFailedSyncs() (*models.SyncReport, error) {
	log.Println("🔄 Retrying failed sync items...")

	var failedItems []models.SyncQueueItem
	o.db.Where("status = ? AND retry_count < max_retries", "failed").
		Order("priority DESC, created_at ASC").
		Limit(100).
		Find(&failedItems)

	report := &models.SyncReport{
		StartedAt:   time.Now(),
		SyncType:    "retry",
		TriggeredBy: "orchestrator",
	}

	for _, item := range failedItems {
		err := o.processSyncItem(&item)
		if err != nil {
			item.IncrementRetry(err.Error())
			report.ErrorCount++
		} else {
			item.Status = "completed"
			now := time.Now()
			item.ProcessedAt = &now
		}
		o.db.Save(&item)
	}

	report.CompletedAt = time.Now()
	if report.ErrorCount == 0 {
		report.Status = models.SyncStatusSuccess
	} else {
		report.Status = models.SyncStatusPartial
	}

	return report, nil
}

func (o *IntegrationOrchestrator) HandleWebhook(source string, eventType string, payload map[string]interface{}) error {
	log.Printf("📥 Processing webhook from %s: %s", source, eventType)

	event := &models.IntegrationEvent{
		EventType:  eventType,
		Source:     source,
		Payload:    payload,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	if err := o.db.Create(event).Error; err != nil {
		log.Printf("⚠️ Failed to store webhook event: %v", err)
	}

	switch source {
	case "fub":
		if o.fubSync != nil {
			return o.fubSync.HandleFUBWebhook(payload)
		}
	case "appfolio":
		return o.handleAppFolioWebhook(eventType, payload)
	}

	return nil
}

func (o *IntegrationOrchestrator) handleAppFolioWebhook(eventType string, payload map[string]interface{}) error {
	switch {
	case eventType == "property.vacancy" || eventType == "tenant.moved_out":
		if o.appfolioPropertySync != nil {
			if err := o.appfolioPropertySync.HandleVacancyWebhook(payload); err != nil {
				return err
			}
		}
		if propertyID, ok := payload["property_id"].(string); ok {
			return o.HandlePropertyVacancy(propertyID)
		}

	case eventType == "tenant.created" || eventType == "tenant.updated" || eventType == "tenant.moved_out":
		if o.appfolioTenantSync != nil {
			return o.appfolioTenantSync.HandleTenantWebhook(payload)
		}

	case eventType == "maintenance.created" || eventType == "maintenance.updated" || eventType == "maintenance.completed":
		if o.appfolioMaintenance != nil {
			return o.appfolioMaintenance.HandleMaintenanceWebhook(payload)
		}
	}

	return nil
}

func (o *IntegrationOrchestrator) GetLastSyncReport() (*models.SyncReport, error) {
	var report models.SyncReport
	err := o.db.Order("created_at DESC").First(&report).Error
	return &report, err
}

func (o *IntegrationOrchestrator) GetSyncHistory(limit int) ([]models.SyncReport, error) {
	var reports []models.SyncReport
	err := o.db.Order("created_at DESC").Limit(limit).Find(&reports).Error
	return reports, err
}

func (o *IntegrationOrchestrator) syncLeadsWithFUB() (int, []models.SyncError) {
	var errors []models.SyncError
	synced := 0

	var leads []models.Lead
	o.db.Where("fub_lead_id IS NOT NULL AND fub_lead_id != ''").
		Where("updated_at > ?", time.Now().Add(-2*time.Hour)).
		Find(&leads)

	for _, lead := range leads {
		if o.fubSync != nil {
			err := o.fubSync.SyncBehavioralScoreToFUB(int64(lead.ID))
			if err != nil {
				errors = append(errors, models.SyncError{
					Entity:    "lead",
					EntityID:  fmt.Sprintf("%d", lead.ID),
					Operation: "sync_score_to_fub",
					Message:   err.Error(),
					Timestamp: time.Now(),
				})
			} else {
				synced++
			}
		}
	}

	return synced, errors
}

func (o *IntegrationOrchestrator) runCrossSystemReconciliation() []models.SyncError {
	var errors []models.SyncError

	var vacantInAppFolio []models.PropertyState
	o.db.Where("is_vacant = ? AND status_source = ?", true, "appfolio").
		Where("is_bookable = ?", false).
		Find(&vacantInAppFolio)

	for _, prop := range vacantInAppFolio {
		o.db.Model(&prop).Updates(map[string]interface{}{
			"is_bookable": true,
		})
	}

	var activeTenants []models.AppFolioTenant
	o.db.Where("is_active = ? AND fub_contact_id = ''", true).Find(&activeTenants)

	for _, tenant := range activeTenants {
		var lead models.Lead
		if err := o.db.Where("email = ?", string(tenant.Email)).First(&lead).Error; err == nil {
			o.db.Model(&tenant).Update("fub_contact_id", lead.FUBLeadID)
			o.db.Model(&tenant).Update("lead_id", lead.ID)
		}
	}

	return errors
}

func (o *IntegrationOrchestrator) createLeadInFUB(lead *models.Lead) error {
	return nil
}

func (o *IntegrationOrchestrator) processEventQueue() {
	for {
		select {
		case event := <-o.eventChan:
			o.processIntegrationEvent(event)
		}
	}
}

func (o *IntegrationOrchestrator) processSyncQueue() {
	for {
		select {
		case item := <-o.syncQueue:
			err := o.processSyncItem(item)
			if err != nil {
				item.IncrementRetry(err.Error())
			} else {
				item.Status = "completed"
				now := time.Now()
				item.ProcessedAt = &now
			}
			o.db.Save(item)
		}
	}
}

func (o *IntegrationOrchestrator) processSyncItem(item *models.SyncQueueItem) error {
	switch item.EntityType {
	case "lead":
		switch item.Operation {
		case "create_fub":
			return nil
		case "sync_score":
			leadID := o.parseEntityID(item.EntityID)
			if o.fubSync != nil && leadID > 0 {
				return o.fubSync.SyncBehavioralScoreToFUB(leadID)
			}
		}
	case "tenant":
		switch item.Operation {
		case "create_appfolio":
			return nil
		}
	}
	return nil
}

func (o *IntegrationOrchestrator) processIntegrationEvent(event *models.IntegrationEvent) {
	log.Printf("Processing event: %s from %s", event.EventType, event.Source)

	now := time.Now()
	event.ProcessedAt = &now
	event.Status = "processed"
	o.db.Save(event)
}

func (o *IntegrationOrchestrator) runReconciliationLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		o.mutex.RLock()
		running := o.isRunning
		o.mutex.RUnlock()

		if !running {
			return
		}

		o.runCrossSystemReconciliation()
	}
}

func (o *IntegrationOrchestrator) emitEvent(eventType, source, entityType, entityID string, payload map[string]interface{}) {
	event := &models.IntegrationEvent{
		EventType:  eventType,
		Source:     source,
		EntityType: entityType,
		EntityID:   entityID,
		Payload:    payload,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	select {
	case o.eventChan <- event:
	default:
		o.db.Create(event)
	}
}

func (o *IntegrationOrchestrator) queueForRetry(entityType, entityID, operation, errorMsg string) {
	item := &models.SyncQueueItem{
		EntityType:  entityType,
		EntityID:    entityID,
		Operation:   operation,
		Source:      "propertyhub",
		Destination: "fub",
		Priority:    5,
		Status:      "pending",
		MaxRetries:  3,
		LastError:   errorMsg,
		ScheduledAt: time.Now().Add(5 * time.Minute),
		CreatedAt:   time.Now(),
	}

	o.db.Create(item)
}

func (o *IntegrationOrchestrator) getLeadIDFromFUBID(fubID string) int64 {
	var lead models.Lead
	if err := o.db.Where("fub_lead_id = ?", fubID).First(&lead).Error; err != nil {
		return 0
	}
	return int64(lead.ID)
}

func (o *IntegrationOrchestrator) parseEntityID(id string) int64 {
	var result int64
	fmt.Sscanf(id, "%d", &result)
	return result
}
