package models

import (
	"time"

	"chrisgross-ctrl-project/internal/security"
)

// ClosingPipelineDataResponse represents a closing pipeline with decrypted sensitive fields
type ClosingPipelineDataResponse struct {
	ID                     uint        `json:"id"`
	PropertyID             *uint       `json:"property_id"`
	PropertyAddress        string      `json:"property_address"`
	MLSID                  string      `json:"mls_id"`
	SoldDate               time.Time   `json:"sold_date"`
	LeaseSentOut           bool        `json:"lease_sent_out"`
	LeaseSentDate          *time.Time  `json:"lease_sent_date,omitempty"`
	LeaseComplete          bool        `json:"lease_complete"`
	LeaseCompleteDate      *time.Time  `json:"lease_complete_date,omitempty"`
	DepositReceived        bool        `json:"deposit_received"`
	DepositReceivedDate    *time.Time  `json:"deposit_received_date,omitempty"`
	FirstMonthReceived     bool        `json:"first_month_received"`
	FirstMonthReceivedDate *time.Time  `json:"first_month_received_date,omitempty"`
	MoveInDate             *time.Time  `json:"move_in_date,omitempty"`
	MoveInDateSource       string      `json:"move_in_date_source"`
	DepositAmount          *float64    `json:"deposit_amount,omitempty"`
	MonthlyRent            *float64    `json:"monthly_rent,omitempty"`
	CommissionEarned       *float64    `json:"commission_earned,omitempty"`
	CommissionRate         *float64    `json:"commission_rate,omitempty"`
	DealType               string      `json:"deal_type"`
	ListingAgentID         *uint       `json:"listing_agent_id,omitempty"`
	TenantAgentID          *uint       `json:"tenant_agent_id,omitempty"`
	LeaseSignedDate        *time.Time  `json:"lease_signed_date,omitempty"`
	ApplicationDate        *time.Time  `json:"application_date,omitempty"`
	ApprovalDate           *time.Time  `json:"approval_date,omitempty"`
	TenantName             string      `json:"tenant_name"`  // Decrypted
	TenantEmail            string      `json:"tenant_email"` // Decrypted
	TenantPhone            string      `json:"tenant_phone"` // Decrypted
	Status                 string      `json:"status"`
	AlertFlags             StringArray `json:"alert_flags"`
	ProcessedEmails        StringArray `json:"processed_emails"`
	LastEmailProcessedAt   *time.Time  `json:"last_email_processed_at,omitempty"`
	AmendmentNotes         string      `json:"amendment_notes"`
	CreatedAt              time.Time   `json:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at"`
}

// ToClosingPipelineDataResponse converts a ClosingPipeline to ClosingPipelineDataResponse with decrypted fields
func ToClosingPipelineDataResponse(cp ClosingPipeline, encryptionManager *security.EncryptionManager) ClosingPipelineDataResponse {
	// Default to encrypted values
	tenantName := string(cp.TenantName)
	tenantEmail := string(cp.TenantEmail)
	tenantPhone := string(cp.TenantPhone)

	// Decrypt if encryption manager is available
	if encryptionManager != nil {
		if decrypted, err := encryptionManager.Decrypt(cp.TenantName); err == nil {
			tenantName = decrypted
		}
		if decrypted, err := encryptionManager.Decrypt(cp.TenantEmail); err == nil {
			tenantEmail = decrypted
		}
		if decrypted, err := encryptionManager.Decrypt(cp.TenantPhone); err == nil {
			tenantPhone = decrypted
		}
	}

	return ClosingPipelineDataResponse{
		ID:                     cp.ID,
		PropertyID:             cp.PropertyID,
		PropertyAddress:        cp.PropertyAddress,
		MLSID:                  cp.MLSID,
		SoldDate:               cp.SoldDate,
		LeaseSentOut:           cp.LeaseSentOut,
		LeaseSentDate:          cp.LeaseSentDate,
		LeaseComplete:          cp.LeaseComplete,
		LeaseCompleteDate:      cp.LeaseCompleteDate,
		DepositReceived:        cp.DepositReceived,
		DepositReceivedDate:    cp.DepositReceivedDate,
		FirstMonthReceived:     cp.FirstMonthReceived,
		FirstMonthReceivedDate: cp.FirstMonthReceivedDate,
		MoveInDate:             cp.MoveInDate,
		MoveInDateSource:       cp.MoveInDateSource,
		DepositAmount:          cp.DepositAmount,
		MonthlyRent:            cp.MonthlyRent,
		CommissionEarned:       cp.CommissionEarned,
		CommissionRate:         cp.CommissionRate,
		DealType:               cp.DealType,
		ListingAgentID:         cp.ListingAgentID,
		TenantAgentID:          cp.TenantAgentID,
		LeaseSignedDate:        cp.LeaseSignedDate,
		ApplicationDate:        cp.ApplicationDate,
		ApprovalDate:           cp.ApprovalDate,
		TenantName:             tenantName,
		TenantEmail:            tenantEmail,
		TenantPhone:            tenantPhone,
		Status:                 cp.Status,
		AlertFlags:             cp.AlertFlags,
		ProcessedEmails:        cp.ProcessedEmails,
		LastEmailProcessedAt:   cp.LastEmailProcessedAt,
		AmendmentNotes:         cp.AmendmentNotes,
		CreatedAt:              cp.CreatedAt,
		UpdatedAt:              cp.UpdatedAt,
	}
}

// ToClosingPipelineDataResponseList converts a slice of ClosingPipelines to ClosingPipelineDataResponses
func ToClosingPipelineDataResponseList(pipelines []ClosingPipeline, encryptionManager *security.EncryptionManager) []ClosingPipelineDataResponse {
	responses := make([]ClosingPipelineDataResponse, len(pipelines))
	for i, pipeline := range pipelines {
		responses[i] = ToClosingPipelineDataResponse(pipeline, encryptionManager)
	}
	return responses
}
