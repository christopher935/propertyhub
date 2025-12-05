package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
)

// ApplicationWorkflowHandlers handles Christopher's specific application workflow
type ApplicationWorkflowHandlers struct {
	db                  *gorm.DB
	service             *services.ApplicationWorkflowService
	behavioralService   *services.BehavioralEventService
	notificationService *services.AdminNotificationService
}

// NewApplicationWorkflowHandlers creates new application workflow handlers
func NewApplicationWorkflowHandlers(db *gorm.DB) *ApplicationWorkflowHandlers {
	return &ApplicationWorkflowHandlers{
		db:                db,
		service:           services.NewApplicationWorkflowService(db),
		behavioralService: services.NewBehavioralEventService(db),
	}
}

func (awh *ApplicationWorkflowHandlers) SetNotificationService(service *services.AdminNotificationService) {
	awh.notificationService = service
}

// GetApplicationWorkflow returns the application workflow view
func (awh *ApplicationWorkflowHandlers) GetApplicationWorkflow(c *gin.Context) {
	c.HTML(http.StatusOK, "application-workflow.html", gin.H{
		"Title": "Application Workflow - PropertyHub",
	})
}

// GetPropertiesWithApplications returns all properties with their application numbers and applicants
func (awh *ApplicationWorkflowHandlers) GetPropertiesWithApplications(c *gin.Context) {
	propertyGroups, err := awh.service.GetPropertiesWithApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch properties with applications",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"properties": propertyGroups,
		},
	})
}

// CreateApplicationNumber creates a new application number for a property
func (awh *ApplicationWorkflowHandlers) CreateApplicationNumber(c *gin.Context) {
	propertyID, err := strconv.ParseUint(c.Param("propertyId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid property ID",
		})
		return
	}
	
	// Find or create property application group
	var propertyGroup models.PropertyApplicationGroup
	result := awh.db.Where("property_id = ?", uint(propertyID)).First(&propertyGroup)
	
	if result.Error != nil {
		// Create new property group
		propertyGroup = models.PropertyApplicationGroup{
			PropertyID:         uint(propertyID),
			PropertyAddress:    c.PostForm("property_address"),
			ApplicationsCreated: 0,
			ActiveApplications:  0,
		}
		if err := awh.db.Create(&propertyGroup).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create property application group",
			})
			return
		}
	}
	
	// Create next application number
	appNumber, err := propertyGroup.CreateNextApplicationNumber(awh.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create application number",
		})
		return
	}

	// Trigger notification for new application
	if awh.notificationService != nil {
		go awh.notificationService.OnApplicationSubmitted(int64(appNumber.ID))
	}

	// ============ ADDED: BEHAVIORAL TRACKING ============
	// Track application submission
	if leadID := extractLeadIDFromGin(c); leadID > 0 {
		sessionID := extractSessionIDFromGin(c)
		ipAddress := c.ClientIP()
		userAgent := c.Request.UserAgent()
		
		propertyIDInt64 := int64(propertyID)
		
		// Track application event (non-blocking)
		go awh.behavioralService.TrackApplication(
			leadID,
			propertyIDInt64,
			appNumber.ApplicationName,
			sessionID,
			ipAddress,
			userAgent,
		)
	}
	// ============ END TRACKING ============
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"application_number": appNumber,
			"message": fmt.Sprintf("Created %s for %s", appNumber.ApplicationName, propertyGroup.PropertyAddress),
		},
	})
}

// MoveApplicantToApplication moves an applicant between application numbers (Christopher's drag-and-drop)
func (awh *ApplicationWorkflowHandlers) MoveApplicantToApplication(c *gin.Context) {
	var request struct {
		ApplicantID           uint   `json:"applicant_id"`
		TargetApplicationID   uint   `json:"target_application_id"`
		MovedBy               string `json:"moved_by"`
		Reason                string `json:"reason,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}
	
	// Use service to move applicant
	err := awh.service.MoveApplicantToApplication(request.ApplicantID, request.TargetApplicationID, 
		request.MovedBy, request.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to move applicant",
		})
		return
	}
	
	// Get applicant and application names for response
	var applicant models.ApplicationApplicant
	var application models.ApplicationNumber
	awh.db.First(&applicant, request.ApplicantID)
	awh.db.First(&application, request.TargetApplicationID)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("%s moved to %s", applicant.ApplicantName, application.ApplicationName),
	})
}

// AssignAgentToApplication assigns an external agent to an application number
func (awh *ApplicationWorkflowHandlers) AssignAgentToApplication(c *gin.Context) {
	var request struct {
		ApplicationNumberID uint   `json:"application_number_id"`
		AgentName           string `json:"agent_name"`
		AgentPhone          string `json:"agent_phone"`
		AgentEmail          string `json:"agent_email"`
		AssignedBy          string `json:"assigned_by"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}
	
	// Get application number
	var appNumber models.ApplicationNumber
	if err := awh.db.First(&appNumber, request.ApplicationNumberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application number not found",
		})
		return
	}
	
	// Assign agent
	err := appNumber.AssignAgent(awh.db, request.AgentName, request.AgentPhone, 
		request.AgentEmail, request.AssignedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign agent",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Agent %s assigned to %s", request.AgentName, appNumber.ApplicationName),
	})
}

// UpdateApplicationStatus updates the status of an application number
func (awh *ApplicationWorkflowHandlers) UpdateApplicationStatus(c *gin.Context) {
	var request struct {
		ApplicationNumberID uint   `json:"application_number_id"`
		Status              string `json:"status"`
		UpdatedBy           string `json:"updated_by"`
		Reason              string `json:"reason,omitempty"`
		Notes               string `json:"notes,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}
	
	// Get application number
	var appNumber models.ApplicationNumber
	if err := awh.db.First(&appNumber, request.ApplicationNumberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application number not found",
		})
		return
	}
	
	// Update status
	err := appNumber.UpdateStatus(awh.db, request.Status, request.UpdatedBy, request.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update status",
		})
		return
	}
	
	// Add notes if provided
	if request.Notes != "" {
		appNumber.InternalNotes = appendNote(appNumber.InternalNotes, 
			fmt.Sprintf("Notes added by %s: %s", request.UpdatedBy, request.Notes))
		awh.db.Save(&appNumber)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("%s status updated to %s", appNumber.ApplicationName, request.Status),
	})
}

// GetApplicantsWithoutApplication returns applicants that haven't been assigned to an application number
func (awh *ApplicationWorkflowHandlers) GetApplicantsWithoutApplication(c *gin.Context) {
	unassignedApplicants, err := awh.service.GetUnassignedApplicants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch unassigned applicants",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"unassigned_applicants": unassignedApplicants,
		},
	})
}

// ProcessBuildiumEmail processes incoming Buildium notification to create applicant
func (awh *ApplicationWorkflowHandlers) ProcessBuildiumEmail(applicantName, applicantEmail, propertyAddress string) error {
	// Find or create property group
	var propertyGroup models.PropertyApplicationGroup
	result := awh.db.Where("property_address = ?", propertyAddress).First(&propertyGroup)
	
	if result.Error != nil {
		// Property group doesn't exist yet - applicant will be unassigned until admin creates application numbers
		// This is fine - they'll show up in the "unassigned applicants" list
	}
	
	// Create applicant record (initially unassigned)
	applicant := &models.ApplicationApplicant{
		ApplicantName:   applicantName,
		ApplicantEmail:  applicantEmail,
		ApplicationDate: time.Now(),
		SourceEmail:     "buildium_notification",
		FUBMatch:        false, // Will be updated by FUB matching process
	}
	
	// Attempt FUB matching (placeholder for actual FUB integration)
	// This would call your FUB service to find matching lead
	fubLeadID, matchFound := awh.findFUBMatch(applicantEmail)
	if matchFound {
		applicant.FUBLeadID = fubLeadID
		applicant.FUBMatch = true
		applicant.MatchScore = 0.9 // High confidence match
	}
	
	return awh.db.Create(applicant).Error
}

// Helper methods

func (awh *ApplicationWorkflowHandlers) updateApplicantCounts(applicationNumberID uint) {
	var count int64
	awh.db.Model(&models.ApplicationApplicant{}).
		Where("application_number_id = ? AND deleted_at IS NULL", applicationNumberID).
		Count(&count)
	
	awh.db.Model(&models.ApplicationNumber{}).
		Where("id = ?", applicationNumberID).
		Update("applicant_count", int(count))
}

func (awh *ApplicationWorkflowHandlers) findFUBMatch(email string) (string, bool) {
	// Placeholder for FUB API integration
	// In production, this would call FUB API to find lead by email
	// Return leadID and whether match was found
	return "", false
}

// appendNote helper function
func appendNote(existingNotes, newNote string) string {
	if existingNotes == "" {
		return newNote
	}
	return existingNotes + "\n\n" + newNote
}

// ============================================================================
// BEHAVIORAL TRACKING HELPER FUNCTIONS FOR GIN
// ============================================================================

// extractLeadIDFromGin gets lead_id from Gin context or cookie
func extractLeadIDFromGin(c *gin.Context) int64 {
	// Try Gin context first
	if leadIDVal, exists := c.Get("lead_id"); exists {
		if id, ok := leadIDVal.(int64); ok {
			return id
		}
		if id, ok := leadIDVal.(uint); ok {
			return int64(id)
		}
	}

	// Try cookie
	if leadIDStr, err := c.Cookie("lead_id"); err == nil {
		if id, err := strconv.ParseInt(leadIDStr, 10, 64); err == nil {
			return id
		}
	}

	return 0 // Anonymous visitor
}

// extractSessionIDFromGin gets session_id from Gin cookie
func extractSessionIDFromGin(c *gin.Context) string {
	if sessionID, err := c.Cookie("session_id"); err == nil {
		return sessionID
	}
	return ""
}
