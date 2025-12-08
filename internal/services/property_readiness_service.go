package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// PropertyReadinessService handles automated property readiness validation
type PropertyReadinessService struct {
	db                     *gorm.DB
	photoProtectionService *PhotoProtectionService
}

// NewPropertyReadinessService creates a new property readiness service
func NewPropertyReadinessService(db *gorm.DB) *PropertyReadinessService {
	return &PropertyReadinessService{
		db:                     db,
		photoProtectionService: NewPhotoProtectionService(db),
	}
}

// PropertyReadinessStatus represents the readiness status of a property
type PropertyReadinessStatus struct {
	PropertyID     uint      `json:"property_id"`
	MLSId          string    `json:"mls_id"`
	IsReady        bool      `json:"is_ready"`
	ReadinessScore int       `json:"readiness_score"` // 0-100
	Status         string    `json:"status"`          // ready, pending, incomplete, blocked
	LastChecked    time.Time `json:"last_checked"`

	// Detailed checks
	PhotoReadiness PhotoReadinessCheck `json:"photo_readiness"`
	PropertyData   PropertyDataCheck   `json:"property_data"`
	SystemChecks   SystemChecksResult  `json:"system_checks"`

	// Actions needed
	RequiredActions []ReadinessAction `json:"required_actions"`
	Recommendations []ReadinessAction `json:"recommendations"`

	// Timeline
	EstimatedReadyDate *time.Time `json:"estimated_ready_date,omitempty"`
	BlockingIssues     []string   `json:"blocking_issues"`
}

// PhotoReadinessCheck represents photo-related readiness checks
type PhotoReadinessCheck struct {
	HasMinimumPhotos   bool                `json:"has_minimum_photos"`
	HasPrimaryPhoto    bool                `json:"has_primary_photo"`
	PhotoQualityScore  int                 `json:"photo_quality_score"`
	TotalPhotos        int                 `json:"total_photos"`
	HighQualityPhotos  int                 `json:"high_quality_photos"`
	RequiredPhotoTypes []RequiredPhotoType `json:"required_photo_types"`
	MissingPhotoTypes  []string            `json:"missing_photo_types"`
}

// PropertyDataCheck represents property data completeness checks
type PropertyDataCheck struct {
	HasCompleteAddress    bool `json:"has_complete_address"`
	HasValidPrice         bool `json:"has_valid_price"`
	HasDescription        bool `json:"has_description"`
	HasPropertyDetails    bool `json:"has_property_details"`
	DataCompletenessScore int  `json:"data_completeness_score"`
}

// SystemChecksResult represents system-level readiness checks
type SystemChecksResult struct {
	PropertyStatusActive    bool `json:"property_status_active"`
	BookingEligibilityValid bool `json:"booking_eligibility_valid"`
	NoConflictingBookings   bool `json:"no_conflicting_bookings"`
	SystemIntegrationsOK    bool `json:"system_integrations_ok"`
}

// RequiredPhotoType represents a required type of photo
type RequiredPhotoType struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Present     bool   `json:"present"`
	PhotoCount  int    `json:"photo_count"`
	MinRequired int    `json:"min_required"`
}

// ReadinessAction represents an action needed to improve readiness
type ReadinessAction struct {
	Type          string `json:"type"`     // required, recommended, optional
	Category      string `json:"category"` // photos, data, system
	Title         string `json:"title"`
	Description   string `json:"description"`
	Priority      int    `json:"priority"` // 1-5, 1 being highest
	EstimatedTime string `json:"estimated_time"`
	AutomatedFix  bool   `json:"automated_fix"`
}

// CheckPropertyReadiness performs comprehensive readiness validation
func (prs *PropertyReadinessService) CheckPropertyReadiness(mlsID string) (*PropertyReadinessStatus, error) {
	// Find property
	var property models.Property
	if err := prs.db.Where("mls_id = ?", mlsID).First(&property).Error; err != nil {
		return nil, fmt.Errorf("property not found: %v", err)
	}

	status := &PropertyReadinessStatus{
		PropertyID:      property.ID,
		MLSId:           mlsID,
		LastChecked:     time.Now(),
		RequiredActions: []ReadinessAction{},
		Recommendations: []ReadinessAction{},
		BlockingIssues:  []string{},
	}

	// Perform all readiness checks
	prs.checkPhotoReadiness(property, status)
	prs.checkPropertyData(property, status)
	prs.checkSystemReadiness(property, status)

	// Calculate overall readiness
	prs.calculateOverallReadiness(status)

	// Determine status and actions
	prs.determineStatusAndActions(status)

	return status, nil
}

// checkPhotoReadiness validates photo-related requirements
func (prs *PropertyReadinessService) checkPhotoReadiness(property models.Property, status *PropertyReadinessStatus) {
	photoCheck := &status.PhotoReadiness

	// Get photo statistics
	stats, err := prs.photoProtectionService.GetPhotoStatistics(property.MLSId)
	if err != nil {
		status.BlockingIssues = append(status.BlockingIssues, "Could not retrieve photo statistics")
		return
	}

	photoCheck.TotalPhotos = int(stats["total_photos"].(int64))
	photoCheck.HasPrimaryPhoto = stats["has_primary_photo"].(bool)

	// Check minimum photo requirements
	requirements := prs.photoProtectionService.GetDefaultPhotoRequirements()
	photoCheck.HasMinimumPhotos = photoCheck.TotalPhotos >= requirements.MinPhotoCount

	// Check required photo types
	photoCheck.RequiredPhotoTypes = []RequiredPhotoType{
		{Type: "exterior", Required: true, MinRequired: 2},
		{Type: "interior", Required: true, MinRequired: 3},
		{Type: "kitchen", Required: true, MinRequired: 1},
		{Type: "bathroom", Required: true, MinRequired: 1},
		{Type: "bedroom", Required: false, MinRequired: 1},
	}

	// Check for missing photo types (simplified - would need photo categorization)
	photoCheck.MissingPhotoTypes = []string{}
	for _, reqType := range photoCheck.RequiredPhotoTypes {
		if reqType.Required && reqType.PhotoCount < reqType.MinRequired {
			photoCheck.MissingPhotoTypes = append(photoCheck.MissingPhotoTypes, reqType.Type)
		}
	}

	// Calculate photo quality score (simplified)
	photoCheck.PhotoQualityScore = 85 // Placeholder - would analyze actual photo quality
	if photoCheck.TotalPhotos >= requirements.MinPhotoCount {
		photoCheck.HighQualityPhotos = photoCheck.TotalPhotos
	}

	// Add required actions for photo issues
	if !photoCheck.HasMinimumPhotos {
		status.RequiredActions = append(status.RequiredActions, ReadinessAction{
			Type:          "required",
			Category:      "photos",
			Title:         "Add More Photos",
			Description:   fmt.Sprintf("Property needs at least %d photos (currently has %d)", requirements.MinPhotoCount, photoCheck.TotalPhotos),
			Priority:      1,
			EstimatedTime: "15-30 minutes",
			AutomatedFix:  false,
		})
	}

	if !photoCheck.HasPrimaryPhoto {
		status.RequiredActions = append(status.RequiredActions, ReadinessAction{
			Type:          "required",
			Category:      "photos",
			Title:         "Set Primary Photo",
			Description:   "Property must have a primary photo for booking display",
			Priority:      1,
			EstimatedTime: "1 minute",
			AutomatedFix:  false,
		})
	}

	if len(photoCheck.MissingPhotoTypes) > 0 {
		status.Recommendations = append(status.Recommendations, ReadinessAction{
			Type:          "recommended",
			Category:      "photos",
			Title:         "Add Required Photo Types",
			Description:   fmt.Sprintf("Consider adding photos of: %v", photoCheck.MissingPhotoTypes),
			Priority:      2,
			EstimatedTime: "10-20 minutes",
			AutomatedFix:  false,
		})
	}
}

// checkPropertyData validates property data completeness
func (prs *PropertyReadinessService) checkPropertyData(property models.Property, status *PropertyReadinessStatus) {
	dataCheck := &status.PropertyData

	// Check address completeness
	dataCheck.HasCompleteAddress = property.Address != "" && property.City != "" && property.State != "" && property.ZipCode != ""

	// Check price validity
	dataCheck.HasValidPrice = property.Price > 0

	// Check description
	dataCheck.HasDescription = property.Description != ""

	// Check property details (simplified)
	dataCheck.HasPropertyDetails = property.Bedrooms != nil && *property.Bedrooms > 0 &&
		property.Bathrooms != nil && *property.Bathrooms > 0 &&
		property.SquareFeet != nil && *property.SquareFeet > 0

	// Calculate data completeness score
	score := 0
	if dataCheck.HasCompleteAddress {
		score += 25
	}
	if dataCheck.HasValidPrice {
		score += 25
	}
	if dataCheck.HasDescription {
		score += 20
	}
	if dataCheck.HasPropertyDetails {
		score += 20
	}
	dataCheck.DataCompletenessScore = score

	// Add required actions for data issues
	if !dataCheck.HasCompleteAddress {
		status.RequiredActions = append(status.RequiredActions, ReadinessAction{
			Type:          "required",
			Category:      "data",
			Title:         "Complete Address Information",
			Description:   "Property must have complete address (street, city, state, zip)",
			Priority:      1,
			EstimatedTime: "2-5 minutes",
			AutomatedFix:  true,
		})
	}

	if !dataCheck.HasValidPrice {
		status.RequiredActions = append(status.RequiredActions, ReadinessAction{
			Type:          "required",
			Category:      "data",
			Title:         "Set Valid Price",
			Description:   "Property must have a valid price greater than $0",
			Priority:      1,
			EstimatedTime: "1-2 minutes",
			AutomatedFix:  true,
		})
	}

}

// checkSystemReadiness validates system-level requirements
func (prs *PropertyReadinessService) checkSystemReadiness(property models.Property, status *PropertyReadinessStatus) {
	systemCheck := &status.SystemChecks

	// Check property status
	systemCheck.PropertyStatusActive = property.Status == "active"

	// Check booking eligibility
	var eligibility models.PropertyBookingEligibility
	err := prs.db.Where("property_id = ?", property.ID).First(&eligibility).Error
	if err == nil {
		systemCheck.BookingEligibilityValid = eligibility.IsBookable
	}

	// Check for conflicting bookings (simplified)
	systemCheck.NoConflictingBookings = true // Placeholder

	// Check system integrations
	systemCheck.SystemIntegrationsOK = true // Placeholder

	// Add required actions for system issues
	if !systemCheck.PropertyStatusActive {
		status.BlockingIssues = append(status.BlockingIssues, "Property status is not active")
		status.RequiredActions = append(status.RequiredActions, ReadinessAction{
			Type:          "required",
			Category:      "system",
			Title:         "Wait for Active Status",
			Description:   "Property must be in 'active' status to accept bookings",
			Priority:      1,
			EstimatedTime: "Depends on property status",
			AutomatedFix:  false,
		})
	}
}

// calculateOverallReadiness calculates the overall readiness score and status
func (prs *PropertyReadinessService) calculateOverallReadiness(status *PropertyReadinessStatus) {
	// Weight different aspects of readiness
	photoScore := 0
	if status.PhotoReadiness.HasMinimumPhotos {
		photoScore += 30
	}
	if status.PhotoReadiness.HasPrimaryPhoto {
		photoScore += 20
	}
	photoScore += status.PhotoReadiness.PhotoQualityScore / 5 // Max 20 points

	dataScore := status.PropertyData.DataCompletenessScore / 2 // Max 50 points

	systemScore := 0
	if status.SystemChecks.PropertyStatusActive {
		systemScore += 15
	}
	if status.SystemChecks.BookingEligibilityValid {
		systemScore += 10
	}
	if status.SystemChecks.NoConflictingBookings {
		systemScore += 3
	}
	if status.SystemChecks.SystemIntegrationsOK {
		systemScore += 2
	}

	status.ReadinessScore = photoScore + dataScore + systemScore

	// Determine if property is ready
	status.IsReady = status.ReadinessScore >= 80 && len(status.BlockingIssues) == 0
}

// determineStatusAndActions determines the overall status and timeline
func (prs *PropertyReadinessService) determineStatusAndActions(status *PropertyReadinessStatus) {
	if len(status.BlockingIssues) > 0 {
		status.Status = "blocked"
	} else if status.IsReady {
		status.Status = "ready"
	} else if status.ReadinessScore >= 60 {
		status.Status = "pending"

		// Estimate ready date based on required actions
		estimatedDays := 0
		for _, action := range status.RequiredActions {
			if action.AutomatedFix {
				estimatedDays += 1
			} else {
				estimatedDays += 2
			}
		}

		if estimatedDays > 0 {
			estimatedDate := time.Now().Add(time.Duration(estimatedDays) * 24 * time.Hour)
			status.EstimatedReadyDate = &estimatedDate
		}
	} else {
		status.Status = "incomplete"
	}
}

// AutoFixReadinessIssues attempts to automatically fix issues that can be resolved programmatically
func (prs *PropertyReadinessService) AutoFixReadinessIssues(mlsID string) (*PropertyReadinessStatus, error) {
	// Get current readiness status
	status, err := prs.CheckPropertyReadiness(mlsID)
	if err != nil {
		return nil, err
	}

	var property models.Property
	if err := prs.db.Where("mls_id = ?", mlsID).First(&property).Error; err != nil {
		return nil, err
	}

	fixedIssues := []string{}

	// Auto-fix data issues
	// Auto-fix system issues
	for _, action := range status.RequiredActions {
		if action.AutomatedFix && action.Category == "system" {
			switch action.Title {
			case "Update Booking Eligibility":
				// Update booking eligibility
				eligibility, err := models.FindOrCreateBookingEligibility(prs.db, property.ID, mlsID)
				if err == nil {
					eligibility.UpdateEligibility(prs.db)
					fixedIssues = append(fixedIssues, "Updated booking eligibility")
				}
			}
		}
	}

	// Re-check readiness after fixes
	updatedStatus, err := prs.CheckPropertyReadiness(mlsID)
	if err != nil {
		return status, err // Return original status if re-check fails
	}

	// Add information about what was fixed
	if len(fixedIssues) > 0 {
		updatedStatus.Recommendations = append(updatedStatus.Recommendations, ReadinessAction{
			Type:          "info",
			Category:      "system",
			Title:         "Auto-fixes Applied",
			Description:   fmt.Sprintf("Automatically fixed: %v", fixedIssues),
			Priority:      5,
			EstimatedTime: "Completed",
			AutomatedFix:  true,
		})
	}

	return updatedStatus, nil
}

// GetReadinessReport generates a comprehensive readiness report for multiple properties
func (prs *PropertyReadinessService) GetReadinessReport(mlsIDs []string) (map[string]*PropertyReadinessStatus, error) {
	report := make(map[string]*PropertyReadinessStatus)

	for _, mlsID := range mlsIDs {
		status, err := prs.CheckPropertyReadiness(mlsID)
		if err != nil {
			// Include error in report
			report[mlsID] = &PropertyReadinessStatus{
				MLSId:          mlsID,
				IsReady:        false,
				Status:         "error",
				BlockingIssues: []string{err.Error()},
				LastChecked:    time.Now(),
			}
		} else {
			report[mlsID] = status
		}
	}

	return report, nil
}

// ScheduleReadinessCheck schedules automatic readiness checks for properties
func (prs *PropertyReadinessService) ScheduleReadinessCheck(mlsID string, interval time.Duration) error {
	// This would integrate with a job scheduler
	// For now, just validate the property exists
	var property models.Property
	return prs.db.Where("mls_id = ?", mlsID).First(&property).Error
}
