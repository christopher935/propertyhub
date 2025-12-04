package security

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

// TRECComplianceLogger handles Texas Real Estate Commission compliance logging
type TRECComplianceLogger struct {
	db          *gorm.DB
	auditLogger *AuditLogger
	logger      *log.Logger
	companyInfo TRECCompanyInfo
}

// TRECCompanyInfo stores company information for compliance
type TRECCompanyInfo struct {
	CompanyName       string `json:"company_name"`
	TRECLicenseNumber string `json:"trec_license_number"`
	PrincipalBroker   string `json:"principal_broker"`
	BrokerLicense     string `json:"broker_license"`
	CompanyAddress    string `json:"company_address"`
	CompanyPhone      string `json:"company_phone"`
}

// TRECEvent represents a TREC-compliant event log
type TRECEvent struct {
	ID                uint                   `json:"id" gorm:"primaryKey"`
	EventType         string                 `json:"event_type" gorm:"not null"`    // lead_generation, property_disclosure, commission_agreement, etc.
	TRECCategory      string                 `json:"trec_category" gorm:"not null"` // IABS, TRELA, Fair_Housing, etc.
	PropertyAddress   string                 `json:"property_address"`
	ClientName        string                 `json:"client_name"`
	AgentName         string                 `json:"agent_name"`
	AgentLicense      string                 `json:"agent_license"`
	ComplianceData    map[string]interface{} `json:"compliance_data" gorm:"type:json"`
	DocumentsAttached []string               `json:"documents_attached" gorm:"type:json"`

	// Audit trail
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	SessionID string `json:"session_id"`

	// Compliance timestamps
	CreatedAt      time.Time  `json:"created_at"`
	RetentionUntil time.Time  `json:"retention_until"` // TREC requires 4-year retention
	ArchivedAt     *time.Time `json:"archived_at"`
}

// TREC compliance event types
const (
	TRECEventLeadGeneration        = "lead_generation"
	TRECEventPropertyInquiry       = "property_inquiry"
	TRECEventShowingScheduled      = "showing_scheduled"
	TRECEventDisclosureProvided    = "disclosure_provided"
	TRECEventConsentCollected      = "consent_collected"
	TRECEventCommissionDisclosed   = "commission_disclosed"
	TRECEventIABSCompliance        = "iabs_compliance"
	TRECEventFairHousingCompliance = "fair_housing_compliance"
	TRECEventDataRetention         = "data_retention"
)

// TREC categories for different regulations
const (
	TRECCategoryIABS        = "IABS"  // Information About Brokerage Services
	TRECCategoryTRELA       = "TRELA" // Texas Real Estate License Act
	TRECCategoryFairHousing = "Fair_Housing"
	TRECCategoryDataPrivacy = "Data_Privacy"
	TRECCategoryMarketing   = "Marketing"
	TRECCategoryDisclosure  = "Disclosure"
)

// NewTRECComplianceLogger creates a new TREC compliance logger
func NewTRECComplianceLogger(db *gorm.DB, auditLogger *AuditLogger, logger *log.Logger) *TRECComplianceLogger {
	// Auto-migrate TREC events table
	db.AutoMigrate(&TRECEvent{})

	return &TRECComplianceLogger{
		db:          db,
		auditLogger: auditLogger,
		logger:      logger,
		companyInfo: TRECCompanyInfo{
			CompanyName:       "Landlords of Texas, LLC",
			TRECLicenseNumber: "TREC_LICENSE_HERE", // Replace with actual license
			PrincipalBroker:   "Christopher Gross",
			BrokerLicense:     "BROKER_LICENSE_HERE", // Replace with actual license
			CompanyAddress:    "Houston, TX",
			CompanyPhone:      "COMPANY_PHONE_HERE", // Replace with actual phone
		},
	}
}

// LogLeadGeneration logs when a new lead is generated through the website
func (t *TRECComplianceLogger) LogLeadGeneration(propertyAddress, clientName, clientEmail, clientPhone, source, ipAddress, userAgent, sessionID string) error {
	complianceData := map[string]interface{}{
		"source":                   source,
		"client_email":             clientEmail,
		"client_phone":             clientPhone,
		"iabs_disclosure_required": true,
		"lead_source_website":      true,
		"consent_status":           "pending",
		"marketing_consent":        false,
		"company_info":             t.companyInfo,
	}

	event := &TRECEvent{
		EventType:       TRECEventLeadGeneration,
		TRECCategory:    TRECCategoryIABS,
		PropertyAddress: propertyAddress,
		ClientName:      clientName,
		ComplianceData:  complianceData,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		SessionID:       sessionID,
		CreatedAt:       time.Now(),
		RetentionUntil:  time.Now().AddDate(4, 0, 0), // 4-year retention as required by TREC
	}

	if err := t.db.Create(event).Error; err != nil {
		t.logger.Printf("❌ TREC Compliance: Failed to log lead generation: %v", err)
		return err
	}

	// Also log to general audit system
	t.auditLogger.LogSecurityEvent("trec_lead_generation", nil, ipAddress, userAgent, fmt.Sprintf("TREC lead generation: %s", clientName), map[string]interface{}{
		"trec_event_id":    event.ID,
		"property_address": propertyAddress,
		"client_name":      clientName,
		"source":           source,
	}, 30)

	t.logger.Printf("✅ TREC Compliance: Lead generation logged for property %s, client %s", propertyAddress, clientName)
	return nil
}

// LogShowingScheduled logs when a property showing is scheduled
func (t *TRECComplianceLogger) LogShowingScheduled(propertyAddress, clientName, agentName, showingType, ipAddress, userAgent, sessionID string) error {
	complianceData := map[string]interface{}{
		"showing_type":             showingType,
		"agent_name":               agentName,
		"iabs_disclosure_provided": false, // Should be updated when provided
		"safety_protocols":         true,
		"insurance_coverage":       true,
		"liability_disclosure":     false, // Should be updated when provided
		"company_info":             t.companyInfo,
	}

	event := &TRECEvent{
		EventType:       TRECEventShowingScheduled,
		TRECCategory:    TRECCategoryTRELA,
		PropertyAddress: propertyAddress,
		ClientName:      clientName,
		AgentName:       agentName,
		ComplianceData:  complianceData,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		SessionID:       sessionID,
		CreatedAt:       time.Now(),
		RetentionUntil:  time.Now().AddDate(4, 0, 0),
	}

	if err := t.db.Create(event).Error; err != nil {
		t.logger.Printf("❌ TREC Compliance: Failed to log showing scheduled: %v", err)
		return err
	}

	t.auditLogger.LogSecurityEvent("trec_showing_scheduled", nil, ipAddress, userAgent, fmt.Sprintf("TREC showing scheduled: %s", propertyAddress), map[string]interface{}{
		"trec_event_id":    event.ID,
		"property_address": propertyAddress,
		"client_name":      clientName,
		"agent_name":       agentName,
		"showing_type":     showingType,
	}, 25)

	t.logger.Printf("✅ TREC Compliance: Showing scheduled logged for property %s", propertyAddress)
	return nil
}

// LogConsentCollected logs when client consent is collected for marketing/communications
func (t *TRECComplianceLogger) LogConsentCollected(clientName, consentType string, consentGiven bool, documentPath, ipAddress, userAgent, sessionID string) error {
	complianceData := map[string]interface{}{
		"consent_type":           consentType,
		"consent_given":          consentGiven,
		"consent_timestamp":      time.Now(),
		"consent_method":         "website_form",
		"document_path":          documentPath,
		"retention_period_years": 4,
		"can_revoke":             true,
		"company_info":           t.companyInfo,
	}

	event := &TRECEvent{
		EventType:      TRECEventConsentCollected,
		TRECCategory:   TRECCategoryDataPrivacy,
		ClientName:     clientName,
		ComplianceData: complianceData,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		SessionID:      sessionID,
		CreatedAt:      time.Now(),
		RetentionUntil: time.Now().AddDate(4, 0, 0),
	}

	if err := t.db.Create(event).Error; err != nil {
		t.logger.Printf("❌ TREC Compliance: Failed to log consent collection: %v", err)
		return err
	}

	t.auditLogger.LogSecurityEvent("trec_consent_collected", nil, ipAddress, userAgent, fmt.Sprintf("TREC consent collected: %s", consentType), map[string]interface{}{
		"trec_event_id": event.ID,
		"client_name":   clientName,
		"consent_type":  consentType,
		"consent_given": consentGiven,
	}, 35)

	t.logger.Printf("✅ TREC Compliance: Consent collection logged for client %s, type %s, given: %v", clientName, consentType, consentGiven)
	return nil
}

// LogIABSDisclosure logs when Information About Brokerage Services is provided
func (t *TRECComplianceLogger) LogIABSDisclosure(propertyAddress, clientName, agentName, disclosureMethod, documentPath, ipAddress, userAgent, sessionID string) error {
	complianceData := map[string]interface{}{
		"disclosure_method":       disclosureMethod,
		"document_path":           documentPath,
		"agent_name":              agentName,
		"disclosure_date":         time.Now(),
		"acknowledgment_received": false, // Should be updated when client acknowledges
		"trec_form_number":        "IABS-1-0",
		"company_info":            t.companyInfo,
	}

	documentsAttached := []string{}
	if documentPath != "" {
		documentsAttached = append(documentsAttached, documentPath)
	}

	event := &TRECEvent{
		EventType:         TRECEventDisclosureProvided,
		TRECCategory:      TRECCategoryIABS,
		PropertyAddress:   propertyAddress,
		ClientName:        clientName,
		AgentName:         agentName,
		ComplianceData:    complianceData,
		DocumentsAttached: documentsAttached,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		SessionID:         sessionID,
		CreatedAt:         time.Now(),
		RetentionUntil:    time.Now().AddDate(4, 0, 0),
	}

	if err := t.db.Create(event).Error; err != nil {
		t.logger.Printf("❌ TREC Compliance: Failed to log IABS disclosure: %v", err)
		return err
	}

	t.auditLogger.LogSecurityEvent("trec_iabs_disclosure", nil, ipAddress, userAgent, fmt.Sprintf("TREC IABS disclosure: %s", propertyAddress), map[string]interface{}{
		"trec_event_id":     event.ID,
		"property_address":  propertyAddress,
		"client_name":       clientName,
		"agent_name":        agentName,
		"disclosure_method": disclosureMethod,
	}, 40)

	t.logger.Printf("✅ TREC Compliance: IABS disclosure logged for property %s, client %s", propertyAddress, clientName)
	return nil
}

// GetComplianceReport generates a compliance report for a specific time period
func (t *TRECComplianceLogger) GetComplianceReport(startDate, endDate time.Time) (map[string]interface{}, error) {
	var events []TRECEvent

	err := t.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve compliance events: %v", err)
	}

	report := map[string]interface{}{
		"report_period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"company_info": t.companyInfo,
		"summary": map[string]interface{}{
			"total_events":         len(events),
			"lead_generation":      0,
			"showings_scheduled":   0,
			"disclosures_provided": 0,
			"consents_collected":   0,
		},
		"events_by_category": make(map[string]int),
		"events_by_type":     make(map[string]int),
		"properties_tracked": make(map[string]int),
		"compliance_gaps":    []string{},
	}

	// Analyze events
	propertyCount := make(map[string]bool)
	for _, event := range events {
		// Count by type
		report["events_by_type"].(map[string]int)[event.EventType]++
		report["events_by_category"].(map[string]int)[event.TRECCategory]++

		// Track unique properties
		if event.PropertyAddress != "" {
			propertyCount[event.PropertyAddress] = true
		}

		// Update summary counters
		summary := report["summary"].(map[string]interface{})
		switch event.EventType {
		case TRECEventLeadGeneration:
			summary["lead_generation"] = summary["lead_generation"].(int) + 1
		case TRECEventShowingScheduled:
			summary["showings_scheduled"] = summary["showings_scheduled"].(int) + 1
		case TRECEventDisclosureProvided:
			summary["disclosures_provided"] = summary["disclosures_provided"].(int) + 1
		case TRECEventConsentCollected:
			summary["consents_collected"] = summary["consents_collected"].(int) + 1
		}
	}

	report["properties_tracked"] = len(propertyCount)

	// Check for compliance gaps
	gaps := t.identifyComplianceGaps(events)
	report["compliance_gaps"] = gaps

	return report, nil
}

// identifyComplianceGaps analyzes events to identify potential compliance issues
func (t *TRECComplianceLogger) identifyComplianceGaps(events []TRECEvent) []string {
	gaps := []string{}

	// Group events by client
	clientEvents := make(map[string][]TRECEvent)
	for _, event := range events {
		if event.ClientName != "" {
			clientEvents[event.ClientName] = append(clientEvents[event.ClientName], event)
		}
	}

	// Check each client for compliance gaps
	for clientName, clientEventList := range clientEvents {
		hasLeadGeneration := false
		hasIABSDisclosure := false
		hasConsentCollection := false

		for _, event := range clientEventList {
			switch event.EventType {
			case TRECEventLeadGeneration:
				hasLeadGeneration = true
			case TRECEventDisclosureProvided:
				if event.TRECCategory == TRECCategoryIABS {
					hasIABSDisclosure = true
				}
			case TRECEventConsentCollected:
				hasConsentCollection = true
			}
		}

		// Identify gaps
		if hasLeadGeneration && !hasIABSDisclosure {
			gaps = append(gaps, fmt.Sprintf("Client %s: Lead generated but IABS disclosure not provided", clientName))
		}
		if hasLeadGeneration && !hasConsentCollection {
			gaps = append(gaps, fmt.Sprintf("Client %s: Lead generated but consent not collected", clientName))
		}
	}

	return gaps
}

// ArchiveOldEvents archives events older than 4 years (TREC requirement)
func (t *TRECComplianceLogger) ArchiveOldEvents() error {
	cutoffDate := time.Now().AddDate(-4, 0, 0)

	result := t.db.Model(&TRECEvent{}).
		Where("created_at < ? AND archived_at IS NULL", cutoffDate).
		Update("archived_at", time.Now())

	if result.Error != nil {
		t.logger.Printf("❌ TREC Compliance: Failed to archive old events: %v", result.Error)
		return result.Error
	}

	t.logger.Printf("✅ TREC Compliance: Archived %d events older than 4 years", result.RowsAffected)
	return nil
}

// UpdateCompanyInfo updates the company information used in compliance logs
func (t *TRECComplianceLogger) UpdateCompanyInfo(info TRECCompanyInfo) {
	t.companyInfo = info
	t.logger.Printf("✅ TREC Compliance: Company information updated")
}
