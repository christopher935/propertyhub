package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/repositories"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookingHandler struct {
	db                  *gorm.DB
	repos               *repositories.Repositories
	encryptionManager   *security.EncryptionManager
	fubBatchService     *services.FUBBatchService
	availabilityService *services.AvailabilityService
	notificationHub     *services.AdminNotificationHub
	calendarService     *services.CalendarIntegrationService
	automationService   *services.SMSEmailAutomationService
}

func NewBookingHandler(db *gorm.DB, repos *repositories.Repositories, em *security.EncryptionManager) *BookingHandler {
	return &BookingHandler{
		db:                  db,
		repos:               repos,
		encryptionManager:   em,
		fubBatchService:     nil,
		availabilityService: services.NewAvailabilityService(db),
		calendarService:     services.NewCalendarIntegrationService(db),
		automationService:   services.NewSMSEmailAutomationService(db),
	}
}

func (h *BookingHandler) SetNotificationHub(hub *services.AdminNotificationHub) {
	h.notificationHub = hub
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if err := req.Validate(); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	ctx := context.Background()

	propertyID, err := strconv.ParseUint(req.PropertyID, 10, 32)
	var mlsID string
	if err != nil {
		var property *models.Property
		property, err = h.repos.Property.FindByMLSID(ctx, req.PropertyID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Property not found", err)
			return
		}
		propertyID = uint64(property.ID)
		mlsID = req.PropertyID
	} else {
		var property models.Property
		if err := h.repos.Property.FindByID(ctx, uint(propertyID), &property); err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Property not found", err)
			return
		}
		mlsID = property.MLSId
	}

	dateFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006-1-2",
	}
	var parsedDate time.Time
	for _, format := range dateFormats {
		parsedDate, err = time.Parse(format, req.ShowingDate)
		if err == nil {
			break
		}
	}
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid date format", err)
		return
	}

	availCheck, err := h.availabilityService.CheckAvailability(mlsID, parsedDate)
	if err != nil {
		log.Printf("Warning: Availability check failed: %v", err)
	} else if !availCheck.IsAvailable {
		c.JSON(http.StatusConflict, gin.H{
			"success":            false,
			"message":            "Time slot not available",
			"reasons":            availCheck.BlockingReasons,
			"alternative_slots": availCheck.AlternativeSlots,
		})
		return
	}

	fubLeadID, err := h.createOrUpdateFUBLead(&req)
	if err != nil {
		log.Printf("Warning: FUB lead creation failed: %v", err)
		fubLeadID = fmt.Sprintf("PENDING_%d", time.Now().Unix())
	}

	booking, err := req.ToBooking(uint(propertyID), fubLeadID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create booking", err)
		return
	}

	booking.ReferenceNumber = generateBookingReference()
	booking.Status = "scheduled"

	if h.encryptionManager != nil {
		encryptedEmail, err := h.encryptionManager.Encrypt(req.Email)
		if err != nil {
			log.Printf("Warning: Email encryption failed: %v", err)
			booking.Email = security.EncryptedString(req.Email)
		} else {
			booking.Email = security.EncryptedString(encryptedEmail)
		}

		encryptedName, err := h.encryptionManager.Encrypt(req.FirstName + " " + req.LastName)
		if err != nil {
			log.Printf("Warning: Name encryption failed: %v", err)
			booking.Name = security.EncryptedString(req.FirstName + " " + req.LastName)
		} else {
			booking.Name = security.EncryptedString(encryptedName)
		}

		encryptedPhone, err := h.encryptionManager.Encrypt(req.Phone)
		if err != nil {
			log.Printf("Warning: Phone encryption failed: %v", err)
			booking.Phone = security.EncryptedString(req.Phone)
		} else {
			booking.Phone = security.EncryptedString(encryptedPhone)
		}
	} else {
		booking.Email = security.EncryptedString(req.Email)
		booking.Name = security.EncryptedString(req.FirstName + " " + req.LastName)
		booking.Phone = security.EncryptedString(req.Phone)
	}

	if err := h.repos.Booking.Create(ctx, booking); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save booking", err)
		return
	}

	if h.calendarService != nil {
		var property models.Property
		if err := h.repos.Property.FindByID(ctx, uint(propertyID), &property); err == nil {
			decryptedAddress := string(property.Address)
			if h.encryptionManager != nil {
				decryptedAddress, _ = h.encryptionManager.Decrypt(property.Address)
			}
			
			bookingData := services.BookingCalendarData{
				PropertyAddress: decryptedAddress,
				ShowingTime:     booking.ShowingDate,
				ContactName:     req.FirstName + " " + req.LastName,
				ContactPhone:    req.Phone,
				ContactEmail:    req.Email,
				SpecialRequests: req.Notes,
			}
			
			if _, err := h.calendarService.CreateShowingEvent(bookingData); err != nil {
				log.Printf("Warning: Calendar event creation failed: %v", err)
			}
		}
	}

	if h.automationService != nil {
		var property models.Property
		if err := h.repos.Property.FindByID(ctx, uint(propertyID), &property); err == nil {
			decryptedAddress := string(property.Address)
			if h.encryptionManager != nil {
				decryptedAddress, _ = h.encryptionManager.Decrypt(property.Address)
			}
			
			automationData := map[string]interface{}{
				"booking_id":       booking.ID,
				"reference_number": booking.ReferenceNumber,
				"first_name":       req.FirstName,
				"last_name":        req.LastName,
				"name":             req.FirstName + " " + req.LastName,
				"email":            req.Email,
				"phone":            req.Phone,
				"property_address": decryptedAddress,
				"showing_date":     booking.ShowingDate.Format("Monday, January 2, 2006"),
				"showing_time":     booking.ShowingDate.Format("3:04 PM"),
				"showing_type":     req.ShowingType,
			}
			
			if err := h.automationService.TriggerAutomation("booking_created", automationData); err != nil {
				log.Printf("Warning: Booking automation failed: %v", err)
			}
		}
	}

	if h.notificationHub != nil {
		var property models.Property
		if err := h.repos.Property.FindByID(ctx, uint(propertyID), &property); err == nil {
			decryptedName := string(booking.Name)
			decryptedAddress := string(property.Address)
			if h.encryptionManager != nil {
				decryptedAddress, _ = h.encryptionManager.Decrypt(property.Address)
				decryptedName, _ = h.encryptionManager.Decrypt(booking.Name)
			}
			h.notificationHub.SendBookingAlert(decryptedAddress, decryptedName, booking.ID)
		}
	}

	utils.SuccessResponse(c, gin.H{
		"booking_id":        booking.ID,
		"reference_number":  booking.ReferenceNumber,
		"showing_date":      booking.ShowingDate,
		"status":            booking.Status,
		"fub_lead_id":       booking.FUBLeadID,
		"message":           "Booking created successfully",
	})
}

func (h *BookingHandler) GetBooking(c *gin.Context) {
	idOrRef := c.Param("id")
	ctx := context.Background()

	var booking models.Booking

	id, err := strconv.ParseUint(idOrRef, 10, 32)
	if err == nil {
		err = h.repos.Booking.FindByID(ctx, uint(id), &booking)
	} else {
		err = h.db.Where("reference_number = ?", idOrRef).First(&booking).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Booking not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve booking", err)
		}
		return
	}

	var decryptedEmail, decryptedName, decryptedPhone string
	if h.encryptionManager != nil {
		decryptedEmail, _ = h.encryptionManager.Decrypt(booking.Email)
		decryptedName, _ = h.encryptionManager.Decrypt(booking.Name)
		decryptedPhone, _ = h.encryptionManager.Decrypt(booking.Phone)
	} else {
		decryptedEmail = string(booking.Email)
		decryptedName = string(booking.Name)
		decryptedPhone = string(booking.Phone)
	}

	utils.SuccessResponse(c, gin.H{
		"booking": gin.H{
			"id":                booking.ID,
			"reference_number":  booking.ReferenceNumber,
			"property_id":       booking.PropertyID,
			"fub_lead_id":       booking.FUBLeadID,
			"email":             decryptedEmail,
			"name":              decryptedName,
			"phone":             decryptedPhone,
			"showing_date":      booking.ShowingDate,
			"duration_minutes":  booking.DurationMinutes,
			"status":            booking.Status,
			"showing_type":      booking.ShowingType,
			"attendee_count":    booking.AttendeeCount,
			"notes":             booking.Notes,
			"created_at":        booking.CreatedAt,
		},
	})
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID", err)
		return
	}

	var cancelReq struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&cancelReq); err != nil {
		cancelReq.Reason = "Cancelled by user"
	}

	ctx := context.Background()
	var booking models.Booking
	if err := h.repos.Booking.FindByID(ctx, uint(id), &booking); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found", err)
		return
	}

	if booking.Status == "cancelled" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Booking already cancelled", nil)
		return
	}

	booking.Status = "cancelled"
	booking.CancellationReason = cancelReq.Reason

	if err := h.repos.Booking.Update(ctx, &booking); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel booking", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":          "Booking cancelled successfully",
		"booking_id":       booking.ID,
		"reference_number": booking.ReferenceNumber,
		"status":           booking.Status,
	})
}

func (h *BookingHandler) ListBookings(c *gin.Context) {
	ctx := context.Background()

	email := c.Query("email")
	status := c.Query("status")
	propertyID := c.Query("property_id")

	criteria := repositories.BookingFilterCriteria{
		CustomerEmail: email,
		Status:        status,
		PaginationOptions: repositories.PaginationOptions{
			Page:     1,
			PageSize: 20,
		},
	}

	if propertyID != "" {
		criteria.PropertyMLSID = propertyID
	}

	pageStr := c.Query("page")
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			criteria.PaginationOptions.Page = page
		}
	}

	pageSizeStr := c.Query("page_size")
	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			criteria.PaginationOptions.PageSize = pageSize
		}
	}

	bookings, total, err := h.repos.Booking.List(ctx, criteria)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve bookings", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"bookings":   bookings,
		"total":      total,
		"page":       criteria.PaginationOptions.Page,
		"page_size":  criteria.PaginationOptions.PageSize,
		"total_pages": (total + int64(criteria.PaginationOptions.PageSize) - 1) / int64(criteria.PaginationOptions.PageSize),
	})
}

func (h *BookingHandler) createOrUpdateFUBLead(req *models.BookingRequest) (string, error) {
	if h.fubBatchService == nil {
		return "", fmt.Errorf("FUB service not available")
	}

	leadData := req.ToFUBLead()
	leadData["showing_date"] = req.ShowingDate
	leadData["showing_time"] = req.ShowingTime
	leadData["property_id"] = req.PropertyID
	leadData["showing_type"] = req.ShowingType
	leadData["has_agent"] = req.HasAgent

	contactID := uint(time.Now().Unix())

	if err := h.fubBatchService.QueueCreateLead(contactID, leadData); err != nil {
		return "", err
	}

	return fmt.Sprintf("FUB_%d", contactID), nil
}

func generateBookingReference() string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	
	return fmt.Sprintf("BK%d%s", timestamp, randomHex[:6])
}

func (h *BookingHandler) MarkCompleted(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID", err)
		return
	}

	ctx := context.Background()
	var booking models.Booking
	if err := h.repos.Booking.FindByID(ctx, uint(id), &booking); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found", err)
		return
	}

	booking.Status = "completed"

	if err := h.repos.Booking.Update(ctx, &booking); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark booking as completed", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":          "Booking marked as completed successfully",
		"booking_id":       booking.ID,
		"reference_number": booking.ReferenceNumber,
		"status":           booking.Status,
	})
}

func (h *BookingHandler) MarkNoShow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID", err)
		return
	}

	ctx := context.Background()
	var booking models.Booking
	if err := h.repos.Booking.FindByID(ctx, uint(id), &booking); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found", err)
		return
	}

	booking.Status = "no-show"

	if err := h.repos.Booking.Update(ctx, &booking); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark booking as no-show", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":          "Booking marked as no-show successfully",
		"booking_id":       booking.ID,
		"reference_number": booking.ReferenceNumber,
		"status":           booking.Status,
	})
}

func (h *BookingHandler) RescheduleBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID", err)
		return
	}

	var request struct {
		NewDate string `json:"new_date" binding:"required"`
		NewTime string `json:"new_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	ctx := context.Background()
	var booking models.Booking
	if err := h.repos.Booking.FindByID(ctx, uint(id), &booking); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found", err)
		return
	}

	dateFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006-1-2",
	}
	var parsedDate time.Time
	for _, format := range dateFormats {
		parsedDate, err = time.Parse(format, request.NewDate)
		if err == nil {
			break
		}
	}
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid date format", err)
		return
	}

	booking.ShowingDate = parsedDate
	booking.Status = "rescheduled"

	if err := h.repos.Booking.Update(ctx, &booking); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reschedule booking", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":          "Booking rescheduled successfully",
		"booking_id":       booking.ID,
		"reference_number": booking.ReferenceNumber,
		"new_date":         request.NewDate,
		"new_time":         request.NewTime,
		"status":           booking.Status,
	})
}
