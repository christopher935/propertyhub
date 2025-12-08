package services

import (
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// ApplicationService handles application records (uses only actual model fields)
type ApplicationService struct {
	db *gorm.DB
}

// ApplicationCreateRequest represents a request to create an application record
type ApplicationCreateRequest struct {
	ApplicantName   string `json:"applicant_name"`
	PropertyAddress string `json:"property_address"`
	ApprovalType    string `json:"approval_type"`
	Status          string `json:"status"`
	Notes           string `json:"notes"`
}

// NewApplicationService creates a new application service
func NewApplicationService(db *gorm.DB) *ApplicationService {
	return &ApplicationService{
		db: db,
	}
}

// CreateApplicationRecord creates a new application record using only existing model fields
func (as *ApplicationService) CreateApplicationRecord(req *ApplicationCreateRequest) (*models.Approval, error) {
	// Check for existing application
	var existingApp models.Approval
	err := as.db.Where("applicant_name ILIKE ? AND property_address ILIKE ?",
		"%"+req.ApplicantName+"%", "%"+req.PropertyAddress+"%").First(&existingApp).Error

	if err == nil {
		// Update existing application with new information
		if req.Notes != "" {
			if existingApp.Notes != "" {
				existingApp.Notes += "\n" + req.Notes
			} else {
				existingApp.Notes = req.Notes
			}
		}

		existingApp.UpdatedAt = time.Now()

		if err := as.db.Save(&existingApp).Error; err != nil {
			return nil, err
		}

		return &existingApp, nil
	}

	// Create new application record using only actual model fields
	application := models.Approval{
		ApplicantName:   req.ApplicantName,
		PropertyAddress: req.PropertyAddress,
		ApprovalType:    req.ApprovalType,
		Status:          req.Status,
		Notes:           req.Notes,
	}

	if err := as.db.Create(&application).Error; err != nil {
		return nil, err
	}

	return &application, nil
}

// UpdateApplicationFromLeaseUpdate updates application status from lease updates
func (as *ApplicationService) UpdateApplicationFromLeaseUpdate(applicantName, propertyAddress, leaseStatus string) (*ApplicationUpdateResult, error) {
	var application models.Approval
	err := as.db.Where("applicant_name ILIKE ? AND property_address ILIKE ?",
		"%"+applicantName+"%", "%"+propertyAddress+"%").First(&application).Error

	if err != nil {
		// Application not found - this is normal for applications not processed through our system
		return &ApplicationUpdateResult{
			Found:         false,
			StatusChanged: false,
		}, nil
	}

	originalStatus := application.Status
	statusChanged := false

	// Update status based on lease update
	leaseStatusLower := strings.ToLower(leaseStatus)

	switch {
	case strings.Contains(leaseStatusLower, "approved") || strings.Contains(leaseStatusLower, "accepted"):
		if application.Status != "approved" {
			application.Status = "approved"
			statusChanged = true
		}

	case strings.Contains(leaseStatusLower, "signed"):
		if application.Status != "approved" {
			application.Status = "approved"
			statusChanged = true
		}
		// Note: Lease signed date stored in Notes field

	case strings.Contains(leaseStatusLower, "rejected") || strings.Contains(leaseStatusLower, "denied"):
		if application.Status != "rejected" {
			application.Status = "rejected"
			statusChanged = true
		}

	case strings.Contains(leaseStatusLower, "move-in") || strings.Contains(leaseStatusLower, "moved in"):
		if application.Status != "completed" {
			application.Status = "completed"
			statusChanged = true
		}
		// Note: Move-in date stored in Notes field
	}

	// Add lease update note
	leaseNote := "Lease update: " + leaseStatus
	if application.Notes != "" {
		application.Notes += "\n" + leaseNote
	} else {
		application.Notes = leaseNote
	}

	application.UpdatedAt = time.Now()

	if err := as.db.Save(&application).Error; err != nil {
		return nil, err
	}

	return &ApplicationUpdateResult{
		Found:         true,
		StatusChanged: statusChanged,
		OldStatus:     originalStatus,
		NewStatus:     application.Status,
		Application:   &application,
	}, nil
}

// ApplicationUpdateResult represents the result of updating an application
type ApplicationUpdateResult struct {
	Found         bool             `json:"found"`
	StatusChanged bool             `json:"status_changed"`
	OldStatus     string           `json:"old_status,omitempty"`
	NewStatus     string           `json:"new_status,omitempty"`
	Application   *models.Approval `json:"application,omitempty"`
}

// GetApplicationsWithContact returns applications that have contact information
func (as *ApplicationService) GetApplicationsWithContact() ([]models.Approval, error) {
	var applications []models.Approval
	err := as.db.Where("applicant_name != '' AND property_address != ''").Find(&applications).Error
	return applications, err
}

// GetApplicationsByStatus returns applications filtered by status
func (as *ApplicationService) GetApplicationsByStatus(status string) ([]models.Approval, error) {
	var applications []models.Approval
	err := as.db.Where("status = ?", status).Find(&applications).Error
	return applications, err
}

// GetApplicationStats returns basic application statistics
func (as *ApplicationService) GetApplicationStats() (*ApplicationStats, error) {
	stats := &ApplicationStats{}

	// Count totals using only existing fields
	as.db.Model(&models.Approval{}).Where("approval_type = ?", "rental_application").Count(&stats.TotalApplications)
	as.db.Model(&models.Approval{}).Where("approval_type = ? AND status = ?", "rental_application", "approved").Count(&stats.Approved)
	as.db.Model(&models.Approval{}).Where("approval_type = ? AND status = ?", "rental_application", "rejected").Count(&stats.Rejected)
	as.db.Model(&models.Approval{}).Where("approval_type = ? AND status = ?", "rental_application", "pending").Count(&stats.Pending)

	// Calculate rates
	if stats.TotalApplications > 0 {
		stats.ApprovalRate = (float64(stats.Approved) / float64(stats.TotalApplications)) * 100
		stats.RejectionRate = (float64(stats.Rejected) / float64(stats.TotalApplications)) * 100
	}

	return stats, nil
}

// ApplicationStats represents application statistics
type ApplicationStats struct {
	TotalApplications int64   `json:"total_applications"`
	Approved          int64   `json:"approved"`
	Rejected          int64   `json:"rejected"`
	Pending           int64   `json:"pending"`
	ApprovalRate      float64 `json:"approval_rate"`
	RejectionRate     float64 `json:"rejection_rate"`
}
