package services

import (
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/handlers"
	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

type AdminNotificationService struct {
	db      *gorm.DB
	wsHub   *handlers.WebSocketHub
	scoring *BehavioralScoringEngine
}

func NewAdminNotificationService(db *gorm.DB, wsHub *handlers.WebSocketHub) *AdminNotificationService {
	return &AdminNotificationService{
		db:      db,
		wsHub:   wsHub,
		scoring: NewBehavioralScoringEngine(db),
	}
}

func (s *AdminNotificationService) TriggerNotification(notifType string, data map[string]interface{}) error {
	notification := &models.AdminNotification{
		Type:      notifType,
		Priority:  s.getPriority(notifType),
		Title:     s.getTitle(notifType, data),
		Message:   s.getMessage(notifType, data),
		Data:      data,
		CreatedAt: time.Now(),
	}

	s.populateNotificationFields(notification, data)

	if err := s.db.Create(notification).Error; err != nil {
		log.Printf("❌ Failed to create notification: %v", err)
		return err
	}

	if s.wsHub != nil {
		if err := s.wsHub.BroadcastToAdmins("notification", notification.ToDict()); err != nil {
			log.Printf("⚠️ Failed to broadcast notification: %v", err)
		} else {
			log.Printf("✅ Notification broadcast: %s - %s", notification.Type, notification.Title)
		}
	}

	return nil
}

func (s *AdminNotificationService) OnHotLeadActive(leadID int64, sessionID string) error {
	var lead models.Lead
	if err := s.db.First(&lead, leadID).Error; err != nil {
		return err
	}

	score, _ := s.scoring.GetScore(leadID)
	compositeScore := 0
	if score != nil {
		compositeScore = score.CompositeScore
	}

	data := map[string]interface{}{
		"lead_id":    leadID,
		"lead_name":  fmt.Sprintf("%s %s", lead.FirstName, lead.LastName),
		"lead_email": lead.Email,
		"lead_score": compositeScore,
		"session_id": sessionID,
	}

	return s.TriggerNotification("hot_lead", data)
}

func (s *AdminNotificationService) OnApplicationSubmitted(applicationID int64) error {
	var application models.ApplicationNumber
	if err := s.db.Preload("PropertyGroup").First(&application, applicationID).Error; err != nil {
		return err
	}

	data := map[string]interface{}{
		"application_id":   applicationID,
		"application_name": application.ApplicationName,
		"property_id":      application.PropertyID,
	}

	if application.PropertyGroup != nil {
		data["property_address"] = application.PropertyGroup.PropertyAddress
	}

	return s.TriggerNotification("application", data)
}

func (s *AdminNotificationService) OnBookingCreated(bookingID int64) error {
	var booking models.Booking
	if err := s.db.Preload("Property").First(&booking, bookingID).Error; err != nil {
		return err
	}

	data := map[string]interface{}{
		"booking_id":       bookingID,
		"reference_number": booking.ReferenceNumber,
		"property_id":      booking.PropertyID,
		"showing_date":     booking.ShowingDate,
	}

	if booking.Property != nil {
		data["property_address"] = fmt.Sprintf("%s, %s", booking.Property.Address, booking.Property.City)
	}

	return s.TriggerNotification("booking", data)
}

func (s *AdminNotificationService) OnReturnVisitor(leadID int64, daysSinceLastVisit int) error {
	var lead models.Lead
	if err := s.db.First(&lead, leadID).Error; err != nil {
		return err
	}

	score, _ := s.scoring.GetScore(leadID)
	compositeScore := 0
	if score != nil {
		compositeScore = score.CompositeScore
	}

	data := map[string]interface{}{
		"lead_id":              leadID,
		"lead_name":            fmt.Sprintf("%s %s", lead.FirstName, lead.LastName),
		"lead_email":           lead.Email,
		"lead_score":           compositeScore,
		"days_since_last_visit": daysSinceLastVisit,
	}

	return s.TriggerNotification("return_visitor", data)
}

func (s *AdminNotificationService) OnEngagementSpike(leadID int64, scoreDelta int, newScore int) error {
	var lead models.Lead
	if err := s.db.First(&lead, leadID).Error; err != nil {
		return err
	}

	data := map[string]interface{}{
		"lead_id":     leadID,
		"lead_name":   fmt.Sprintf("%s %s", lead.FirstName, lead.LastName),
		"lead_email":  lead.Email,
		"lead_score":  newScore,
		"score_delta": scoreDelta,
	}

	return s.TriggerNotification("engagement_spike", data)
}

func (s *AdminNotificationService) OnPropertySaved(leadID int64, propertyID int64) error {
	var lead models.Lead
	var property models.Property

	if err := s.db.First(&lead, leadID).Error; err != nil {
		return err
	}

	if err := s.db.First(&property, propertyID).Error; err != nil {
		return err
	}

	data := map[string]interface{}{
		"lead_id":          leadID,
		"lead_name":        fmt.Sprintf("%s %s", lead.FirstName, lead.LastName),
		"lead_email":       lead.Email,
		"property_id":      propertyID,
		"property_address": fmt.Sprintf("%s, %s", property.Address, property.City),
	}

	return s.TriggerNotification("property_saved", data)
}

func (s *AdminNotificationService) OnInquirySent(leadID int64, propertyID *int64) error {
	var lead models.Lead
	if err := s.db.First(&lead, leadID).Error; err != nil {
		return err
	}

	data := map[string]interface{}{
		"lead_id":    leadID,
		"lead_name":  fmt.Sprintf("%s %s", lead.FirstName, lead.LastName),
		"lead_email": lead.Email,
	}

	if propertyID != nil {
		var property models.Property
		if err := s.db.First(&property, *propertyID).Error; err == nil {
			data["property_id"] = *propertyID
			data["property_address"] = fmt.Sprintf("%s, %s", property.Address, property.City)
		}
	}

	return s.TriggerNotification("inquiry", data)
}

func (s *AdminNotificationService) OnMultiplePropertiesViewed(leadID int64, viewCount int, sessionID string) error {
	var lead models.Lead
	if err := s.db.First(&lead, leadID).Error; err != nil {
		return err
	}

	data := map[string]interface{}{
		"lead_id":    leadID,
		"lead_name":  fmt.Sprintf("%s %s", lead.FirstName, lead.LastName),
		"lead_email": lead.Email,
		"view_count": viewCount,
		"session_id": sessionID,
	}

	return s.TriggerNotification("multiple_views", data)
}

func (s *AdminNotificationService) populateNotificationFields(notification *models.AdminNotification, data map[string]interface{}) {
	if leadID, ok := data["lead_id"].(int64); ok {
		leadIDInt := int(leadID)
		notification.LeadID = &leadIDInt
	}

	if leadName, ok := data["lead_name"].(string); ok {
		notification.LeadName = leadName
	}

	if leadEmail, ok := data["lead_email"].(string); ok {
		notification.LeadEmail = leadEmail
	}

	if leadScore, ok := data["lead_score"].(int); ok {
		notification.LeadScore = leadScore
	}

	if propertyID, ok := data["property_id"].(int64); ok {
		propertyIDInt := int(propertyID)
		notification.PropertyID = &propertyIDInt
	} else if propertyID, ok := data["property_id"].(uint); ok {
		propertyIDInt := int(propertyID)
		notification.PropertyID = &propertyIDInt
	}

	if propertyAddr, ok := data["property_address"].(string); ok {
		notification.PropertyAddr = propertyAddr
	}

	notification.ActionURL = s.getActionURL(notification.Type, data)
	notification.ActionLabel = s.getActionLabel(notification.Type)
}

func (s *AdminNotificationService) getPriority(notifType string) string {
	highPriority := map[string]bool{
		"hot_lead":         true,
		"application":      true,
		"booking":          true,
		"return_visitor":   true,
		"engagement_spike": true,
	}

	if highPriority[notifType] {
		return "high"
	}
	return "medium"
}

func (s *AdminNotificationService) getTitle(notifType string, data map[string]interface{}) string {
	titles := map[string]string{
		"hot_lead":         "🔥 Hot Lead Active",
		"application":      "📋 New Application Submitted",
		"booking":          "📅 New Tour Booking",
		"return_visitor":   "🔄 Hot Lead Returned",
		"engagement_spike": "📈 Engagement Spike Detected",
		"property_saved":   "⭐ Property Saved",
		"inquiry":          "💬 New Inquiry",
		"multiple_views":   "👀 Multiple Properties Viewed",
	}

	if title, ok := titles[notifType]; ok {
		return title
	}
	return "New Notification"
}

func (s *AdminNotificationService) getMessage(notifType string, data map[string]interface{}) string {
	leadName := ""
	if name, ok := data["lead_name"].(string); ok {
		leadName = name
	}

	switch notifType {
	case "hot_lead":
		score := 0
		if s, ok := data["lead_score"].(int); ok {
			score = s
		}
		return fmt.Sprintf("%s (score: %d) is actively browsing properties", leadName, score)

	case "application":
		appName := ""
		if name, ok := data["application_name"].(string); ok {
			appName = name
		}
		return fmt.Sprintf("New application %s has been submitted", appName)

	case "booking":
		refNum := ""
		if ref, ok := data["reference_number"].(string); ok {
			refNum = ref
		}
		return fmt.Sprintf("Tour booking %s created", refNum)

	case "return_visitor":
		days := 0
		if d, ok := data["days_since_last_visit"].(int); ok {
			days = d
		}
		return fmt.Sprintf("%s returned after %d days", leadName, days)

	case "engagement_spike":
		delta := 0
		if d, ok := data["score_delta"].(int); ok {
			delta = d
		}
		return fmt.Sprintf("%s's score increased by %d points", leadName, delta)

	case "property_saved":
		addr := ""
		if a, ok := data["property_address"].(string); ok {
			addr = a
		}
		return fmt.Sprintf("%s saved %s", leadName, addr)

	case "inquiry":
		return fmt.Sprintf("%s sent an inquiry", leadName)

	case "multiple_views":
		count := 0
		if c, ok := data["view_count"].(int); ok {
			count = c
		}
		return fmt.Sprintf("%s viewed %d properties in one session", leadName, count)

	default:
		return "New notification"
	}
}

func (s *AdminNotificationService) getActionURL(notifType string, data map[string]interface{}) string {
	switch notifType {
	case "hot_lead", "return_visitor", "engagement_spike":
		if leadID, ok := data["lead_id"].(int64); ok {
			return fmt.Sprintf("/admin/leads/%d", leadID)
		}
	case "application":
		if appID, ok := data["application_id"].(int64); ok {
			return fmt.Sprintf("/admin/applications/%d", appID)
		}
		return "/admin/applications"
	case "booking":
		if bookingID, ok := data["booking_id"].(int64); ok {
			return fmt.Sprintf("/admin/bookings/%d", bookingID)
		}
		return "/admin/bookings"
	case "property_saved", "inquiry", "multiple_views":
		if leadID, ok := data["lead_id"].(int64); ok {
			return fmt.Sprintf("/admin/leads/%d", leadID)
		}
	}
	return "/admin/dashboard"
}

func (s *AdminNotificationService) getActionLabel(notifType string) string {
	labels := map[string]string{
		"hot_lead":         "View Lead",
		"application":      "View Application",
		"booking":          "View Booking",
		"return_visitor":   "View Lead",
		"engagement_spike": "View Lead",
		"property_saved":   "View Lead",
		"inquiry":          "View Inquiry",
		"multiple_views":   "View Lead",
	}

	if label, ok := labels[notifType]; ok {
		return label
	}
	return "View Details"
}

func (s *AdminNotificationService) GetRecentNotifications(limit int) ([]models.AdminNotification, error) {
	var notifications []models.AdminNotification
	err := s.db.Where("dismissed_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (s *AdminNotificationService) GetUnreadCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.AdminNotification{}).
		Where("read_at IS NULL AND dismissed_at IS NULL").
		Count(&count).Error
	return count, err
}

func (s *AdminNotificationService) MarkAsRead(notificationID uint) error {
	now := time.Now()
	return s.db.Model(&models.AdminNotification{}).
		Where("id = ?", notificationID).
		Update("read_at", now).Error
}

func (s *AdminNotificationService) Dismiss(notificationID uint) error {
	now := time.Now()
	return s.db.Model(&models.AdminNotification{}).
		Where("id = ?", notificationID).
		Update("dismissed_at", now).Error
}

func (s *AdminNotificationService) DismissAll() error {
	now := time.Now()
	return s.db.Model(&models.AdminNotification{}).
		Where("dismissed_at IS NULL").
		Update("dismissed_at", now).Error
}
