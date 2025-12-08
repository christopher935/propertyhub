package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// PropertyApplicationGroup represents a property with its application groups
type PropertyApplicationGroup struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Property Information
	PropertyID      uint   `json:"property_id" gorm:"not null;index"`
	PropertyAddress string `json:"property_address" gorm:"not null"`

	// Application Management
	TotalApplications   int `json:"total_applications" gorm:"default:0"`
	ActiveApplications  int `json:"active_applications" gorm:"default:0"`
	ApplicationsCreated int `json:"applications_created" gorm:"default:0"`

	// Relationship to application numbers
	ApplicationNumbers []ApplicationNumber `json:"application_numbers,omitempty" gorm:"foreignKey:PropertyApplicationGroupID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ApplicationNumber represents a numbered application slot for a property
type ApplicationNumber struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Linking
	PropertyApplicationGroupID uint                      `json:"property_application_group_id" gorm:"not null;index"`
	PropertyApplicationGroup   *PropertyApplicationGroup `json:"property_application_group,omitempty" gorm:"foreignKey:PropertyApplicationGroupID"`

	// Application Number Info
	ApplicationNumber int    `json:"application_number" gorm:"not null"` // 1, 2, 3, etc.
	ApplicationName   string `json:"application_name"`                   // "Application 1", "Application 2"

	// Status Tracking
	Status          string     `json:"status" gorm:"default:'submitted'"` // submitted, review, further_review, rental_history_received, approved, denied, backup, cancelled
	StatusUpdatedAt *time.Time `json:"status_updated_at"`
	StatusUpdatedBy string     `json:"status_updated_by"`

	// Agent Assignment (External agents with contact info)
	AssignedAgentName  string     `json:"assigned_agent_name"`
	AssignedAgentPhone string     `json:"assigned_agent_phone"`
	AssignedAgentEmail string     `json:"assigned_agent_email"`
	AgentAssignedAt    *time.Time `json:"agent_assigned_at"`

	// Notes and Tracking
	ApplicationNotes string `json:"application_notes" gorm:"type:text"`
	InternalNotes    string `json:"internal_notes" gorm:"type:text"`

	// Applicant Count
	ApplicantCount int `json:"applicant_count" gorm:"default:0"`

	// Runtime fields for template rendering
	Applicants []ApplicationApplicant `json:"applicants,omitempty" gorm:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ApplicationApplicant represents individual applicants within an application number
type ApplicationApplicant struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Linking
	ApplicationNumberID uint               `json:"application_number_id" gorm:"not null;index"`
	ApplicationNumber   *ApplicationNumber `json:"application_number,omitempty" gorm:"foreignKey:ApplicationNumberID"`

	// Applicant Information
	ApplicantName  string `json:"applicant_name" gorm:"not null"`
	ApplicantEmail string `json:"applicant_email" gorm:"not null;index"`
	ApplicantPhone string `json:"applicant_phone"`

	// FUB Integration
	FUBLeadID  string  `json:"fub_lead_id"`
	FUBMatch   bool    `json:"fub_match" gorm:"default:false"`
	MatchScore float64 `json:"match_score" gorm:"default:0"`

	// Application Details
	ApplicationDate time.Time `json:"application_date"`
	SourceEmail     string    `json:"source_email"` // Original Buildium email
	ApplicationData JSONB     `json:"application_data" gorm:"type:jsonb"`

	// Individual Notes
	ApplicantNotes   string  `json:"applicant_notes" gorm:"type:text"`
	CreditScore      int     `json:"credit_score"`
	Income           float64 `json:"income"`
	EmploymentStatus string  `json:"employment_status"`
	RentalHistory    string  `json:"rental_history"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Application status constants
const (
	AppStatusSubmitted             = "submitted"
	AppStatusReview                = "review"
	AppStatusFurtherReview         = "further_review"
	AppStatusRentalHistoryReceived = "rental_history_received"
	AppStatusApproved              = "approved"
	AppStatusDenied                = "denied"
	AppStatusBackup                = "backup"
	AppStatusCancelled             = "cancelled"
)

// ApplicationStatusLog tracks status changes for audit trail
type ApplicationStatusLog struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Reference
	ApplicationNumberID uint `json:"application_number_id" gorm:"not null;index"`

	// Status Change Information
	PreviousStatus string `json:"previous_status"`
	NewStatus      string `json:"new_status"`
	ChangedBy      string `json:"changed_by"`
	ChangeReason   string `json:"change_reason"`
	Notes          string `json:"notes" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
}

// Helper methods for ApplicationNumber

// GetApplicantList returns all applicants in this application number
func (an *ApplicationNumber) GetApplicantList(db *gorm.DB) ([]ApplicationApplicant, error) {
	var applicants []ApplicationApplicant
	err := db.Where("application_number_id = ? AND deleted_at IS NULL", an.ID).Find(&applicants).Error
	return applicants, err
}

// AddApplicant adds an applicant to this application number
func (an *ApplicationNumber) AddApplicant(db *gorm.DB, applicant *ApplicationApplicant) error {
	applicant.ApplicationNumberID = an.ID

	// Create applicant
	if err := db.Create(applicant).Error; err != nil {
		return err
	}

	// Update applicant count
	var count int64
	db.Model(&ApplicationApplicant{}).Where("application_number_id = ? AND deleted_at IS NULL", an.ID).Count(&count)
	an.ApplicantCount = int(count)

	return db.Save(an).Error
}

// RemoveApplicant removes an applicant from this application number
func (an *ApplicationNumber) RemoveApplicant(db *gorm.DB, applicantID uint) error {
	// Soft delete applicant
	if err := db.Delete(&ApplicationApplicant{}, applicantID).Error; err != nil {
		return err
	}

	// Update applicant count
	var count int64
	db.Model(&ApplicationApplicant{}).Where("application_number_id = ? AND deleted_at IS NULL", an.ID).Count(&count)
	an.ApplicantCount = int(count)

	return db.Save(an).Error
}

// UpdateStatus updates the application number status with audit logging
func (an *ApplicationNumber) UpdateStatus(db *gorm.DB, newStatus, updatedBy, reason string) error {
	oldStatus := an.Status

	// Create audit log
	statusLog := &ApplicationStatusLog{
		ApplicationNumberID: an.ID,
		PreviousStatus:      oldStatus,
		NewStatus:           newStatus,
		ChangedBy:           updatedBy,
		ChangeReason:        reason,
		CreatedAt:           time.Now(),
	}

	if err := db.Create(statusLog).Error; err != nil {
		return err
	}

	// Update application status
	now := time.Now()
	an.Status = newStatus
	an.StatusUpdatedAt = &now
	an.StatusUpdatedBy = updatedBy

	return db.Save(an).Error
}

// AssignAgent assigns an external agent to this application number
func (an *ApplicationNumber) AssignAgent(db *gorm.DB, agentName, agentPhone, agentEmail, assignedBy string) error {
	now := time.Now()
	an.AssignedAgentName = agentName
	an.AssignedAgentPhone = agentPhone
	an.AssignedAgentEmail = agentEmail
	an.AgentAssignedAt = &now

	// Log the assignment
	note := fmt.Sprintf("Agent %s (%s, %s) assigned by %s on %s",
		agentName, agentPhone, agentEmail, assignedBy, now.Format("Jan 2, 2006"))
	an.ApplicationNotes = appendNote(an.ApplicationNotes, note)

	return db.Save(an).Error
}

// RemoveAgent removes agent assignment from this application number
func (an *ApplicationNumber) RemoveAgent(db *gorm.DB, removedBy string) error {
	if an.AssignedAgentName == "" {
		return nil // No agent to remove
	}

	// Log the removal
	note := fmt.Sprintf("Agent %s (%s) removed by %s on %s",
		an.AssignedAgentName, an.AssignedAgentPhone, removedBy, time.Now().Format("Jan 2, 2006"))
	an.ApplicationNotes = appendNote(an.ApplicationNotes, note)

	// Clear agent assignment
	an.AssignedAgentName = ""
	an.AssignedAgentPhone = ""
	an.AssignedAgentEmail = ""
	an.AgentAssignedAt = nil

	return db.Save(an).Error
}

// UpdateAgentInfo updates existing agent contact information
func (an *ApplicationNumber) UpdateAgentInfo(db *gorm.DB, agentName, agentPhone, agentEmail, updatedBy string) error {
	oldInfo := fmt.Sprintf("%s (%s)", an.AssignedAgentName, an.AssignedAgentPhone)

	an.AssignedAgentName = agentName
	an.AssignedAgentPhone = agentPhone
	an.AssignedAgentEmail = agentEmail

	// Log the update
	note := fmt.Sprintf("Agent updated from %s to %s (%s, %s) by %s on %s",
		oldInfo, agentName, agentPhone, agentEmail, updatedBy, time.Now().Format("Jan 2, 2006"))
	an.ApplicationNotes = appendNote(an.ApplicationNotes, note)

	return db.Save(an).Error
}

// Helper function to append notes
func appendNote(existingNotes, newNote string) string {
	if existingNotes == "" {
		return newNote
	}
	return existingNotes + "\n\n" + newNote
}

// Helper methods for PropertyApplicationGroup

// CreateNextApplicationNumber creates the next application number for this property
func (pag *PropertyApplicationGroup) CreateNextApplicationNumber(db *gorm.DB) (*ApplicationNumber, error) {
	// Get the highest application number
	var maxNumber int
	db.Model(&ApplicationNumber{}).
		Where("property_application_group_id = ?", pag.ID).
		Select("COALESCE(MAX(application_number), 0)").
		Scan(&maxNumber)

	nextNumber := maxNumber + 1

	appNumber := &ApplicationNumber{
		PropertyApplicationGroupID: pag.ID,
		ApplicationNumber:          nextNumber,
		ApplicationName:            fmt.Sprintf("Application %d", nextNumber),
		Status:                     AppStatusSubmitted,
		ApplicantCount:             0,
	}

	if err := db.Create(appNumber).Error; err != nil {
		return nil, err
	}

	// Update group counts
	pag.ApplicationsCreated++
	pag.ActiveApplications++
	db.Save(pag)

	return appNumber, nil
}

// GetAllApplicationNumbers returns all application numbers for this property
func (pag *PropertyApplicationGroup) GetAllApplicationNumbers(db *gorm.DB) ([]ApplicationNumber, error) {
	var applications []ApplicationNumber
	err := db.Where("property_application_group_id = ? AND deleted_at IS NULL", pag.ID).
		Order("application_number ASC").
		Find(&applications).Error
	return applications, err
}
