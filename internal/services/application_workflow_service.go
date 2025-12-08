package services

import (
	"chrisgross-ctrl-project/internal/models"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// ApplicationWorkflowService handles Christopher's specific application workflow
type ApplicationWorkflowService struct {
	db *gorm.DB
}

// NewApplicationWorkflowService creates a new application workflow service
func NewApplicationWorkflowService(db *gorm.DB) *ApplicationWorkflowService {
	// Auto-migrate tables
	db.AutoMigrate(&models.PropertyApplicationGroup{})
	db.AutoMigrate(&models.ApplicationNumber{})
	db.AutoMigrate(&models.ApplicationApplicant{})
	db.AutoMigrate(&models.ApplicationStatusLog{})

	return &ApplicationWorkflowService{
		db: db,
	}
}

// ProcessBuildiumEmail creates unassigned applicant from Buildium email notification
func (aws *ApplicationWorkflowService) ProcessBuildiumEmail(applicantName, applicantEmail, propertyAddress string) error {
	// Create unassigned applicant record - Christopher's business model
	applicant := &models.ApplicationApplicant{
		ApplicantName:   applicantName,
		ApplicantEmail:  applicantEmail,
		ApplicationDate: time.Now(),
		SourceEmail:     "buildium_notification",
		FUBMatch:        false, // Will be updated by FUB matching process
		// ApplicationNumberID is nil - this makes them "unassigned"
	}

	// Attempt FUB matching
	fubLeadID, matchFound := aws.findFUBMatch(applicantEmail)
	if matchFound {
		applicant.FUBLeadID = fubLeadID
		applicant.FUBMatch = true
		applicant.MatchScore = 0.9 // High confidence match
	}

	// Add property address to applicant data for reference
	applicant.ApplicationData = models.JSONB{
		"property_address": propertyAddress,
		"source":           "buildium",
		"processed_at":     time.Now(),
	}

	return aws.db.Create(applicant).Error
}

// findFUBMatch attempts to find matching FUB lead by email
func (aws *ApplicationWorkflowService) findFUBMatch(email string) (string, bool) {
	// First check if we have a Lead record with this email
	var lead models.Lead
	if err := aws.db.Where("email = ?", email).First(&lead).Error; err == nil {
		if lead.FUBLeadID != "" {
			return lead.FUBLeadID, true
		}
	}

	// In production, this would also call FUB API to search for lead
	// For now, return placeholder
	return "", false
}

// GetUnassignedApplicants returns applicants not yet assigned to application numbers
func (aws *ApplicationWorkflowService) GetUnassignedApplicants() ([]models.ApplicationApplicant, error) {
	var applicants []models.ApplicationApplicant

	// Find applicants without application_number_id (unassigned)
	err := aws.db.Where("application_number_id IS NULL OR application_number_id = 0").
		Order("application_date DESC").
		Find(&applicants).Error

	return applicants, err
}

// GetPropertiesWithApplications returns all properties with their application numbers and applicants
func (aws *ApplicationWorkflowService) GetPropertiesWithApplications() ([]models.PropertyApplicationGroup, error) {
	var propertyGroups []models.PropertyApplicationGroup

	// Get all property groups with their application numbers
	result := aws.db.Preload("ApplicationNumbers", func(db *gorm.DB) *gorm.DB {
		return db.Order("application_number ASC")
	}).Find(&propertyGroups)

	// Handle empty results gracefully
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	// Return empty array if no data
	if len(propertyGroups) == 0 {
		return []models.PropertyApplicationGroup{}, nil
	}

	// For each property group, load applicants for each application number
	for i := range propertyGroups {
		propertyGroups[i].ApplicationNumbers = []models.ApplicationNumber{}

		// Load application numbers for this property
		var appNumbers []models.ApplicationNumber
		aws.db.Where("property_application_group_id = ? AND deleted_at IS NULL", propertyGroups[i].ID).
			Order("application_number ASC").
			Find(&appNumbers)

		// For each application number, load applicants
		for j, appNum := range appNumbers {
			applicants, _ := appNum.GetApplicantList(aws.db)
			appNumbers[j].Applicants = applicants
		}

		propertyGroups[i].ApplicationNumbers = appNumbers
	}

	return propertyGroups, nil
}

// MoveApplicantToApplication moves an applicant to a specific application number
func (aws *ApplicationWorkflowService) MoveApplicantToApplication(applicantID, targetApplicationID uint, movedBy, reason string) error {
	// Get applicant
	var applicant models.ApplicationApplicant
	if err := aws.db.First(&applicant, applicantID).Error; err != nil {
		return err
	}

	// Get target application
	var targetApplication models.ApplicationNumber
	if err := aws.db.First(&targetApplication, targetApplicationID).Error; err != nil {
		return err
	}

	// Update applicant assignment
	oldAppID := applicant.ApplicationNumberID
	applicant.ApplicationNumberID = targetApplicationID

	// Log the move in applicant notes
	note := fmt.Sprintf("Moved to %s by %s on %s",
		targetApplication.ApplicationName, movedBy, time.Now().Format("Jan 2, 2006"))
	if reason != "" {
		note += fmt.Sprintf(" - Reason: %s", reason)
	}

	if applicant.ApplicantNotes == "" {
		applicant.ApplicantNotes = note
	} else {
		applicant.ApplicantNotes += "\n\n" + note
	}

	// Save applicant
	if err := aws.db.Save(&applicant).Error; err != nil {
		return err
	}

	// Update applicant counts
	if oldAppID != 0 {
		aws.updateApplicationCount(oldAppID)
	}
	aws.updateApplicationCount(targetApplicationID)

	return nil
}

// updateApplicationCount updates the applicant count for an application number
func (aws *ApplicationWorkflowService) updateApplicationCount(applicationNumberID uint) {
	var count int64
	aws.db.Model(&models.ApplicationApplicant{}).
		Where("application_number_id = ? AND deleted_at IS NULL", applicationNumberID).
		Count(&count)

	aws.db.Model(&models.ApplicationNumber{}).
		Where("id = ?", applicationNumberID).
		Update("applicant_count", int(count))
}
