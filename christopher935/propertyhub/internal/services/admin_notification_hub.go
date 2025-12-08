package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type AdminNotificationHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan *models.AdminNotification
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
	db         *gorm.DB
}

func NewAdminNotificationHub(db *gorm.DB) *AdminNotificationHub {
	hub := &AdminNotificationHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan *models.AdminNotification, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		db:         db,
	}
	go hub.run()
	return hub
}

func (h *AdminNotificationHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("🔔 Admin notification client registered. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Printf("🔔 Admin notification client unregistered. Total: %d", len(h.clients))

		case notification := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				err := client.WriteJSON(notification.ToDict())
				if err != nil {
					log.Printf("❌ Error sending notification to client: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *AdminNotificationHub) Register(conn *websocket.Conn) {
	h.register <- conn
}

func (h *AdminNotificationHub) Unregister(conn *websocket.Conn) {
	h.unregister <- conn
}

func (h *AdminNotificationHub) Broadcast(notification *models.AdminNotification) {
	if err := h.db.Create(notification).Error; err != nil {
		log.Printf("❌ Failed to save notification: %v", err)
		return
	}

	h.broadcast <- notification
	log.Printf("📢 Broadcast notification: %s - %s", notification.Type, notification.Title)
}

func (h *AdminNotificationHub) SendHotLeadAlert(leadName string, score int, leadID int64) {
	data, _ := json.Marshal(map[string]interface{}{
		"lead_id":   leadID,
		"lead_name": leadName,
		"score":     score,
	})

	notification := &models.AdminNotification{
		Type:     "hot_lead",
		Title:    "🔥 Hot Lead Alert",
		Message:  leadName + " just became a hot lead!",
		Priority: "high",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) SendBookingAlert(propertyAddress string, leadName string, bookingID uint) {
	data, _ := json.Marshal(map[string]interface{}{
		"booking_id":       bookingID,
		"property_address": propertyAddress,
		"lead_name":        leadName,
	})

	notification := &models.AdminNotification{
		Type:     "booking_created",
		Title:    "📅 New Booking Scheduled",
		Message:  leadName + " scheduled a showing at " + propertyAddress,
		Priority: "normal",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) SendApplicationAlert(propertyAddress string, applicantName string, applicationID uint) {
	data, _ := json.Marshal(map[string]interface{}{
		"application_id":   applicationID,
		"property_address": propertyAddress,
		"applicant_name":   applicantName,
	})

	notification := &models.AdminNotification{
		Type:     "application_submitted",
		Title:    "📝 New Application Submitted",
		Message:  applicantName + " submitted an application for " + propertyAddress,
		Priority: "high",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) SendReturnVisitorAlert(leadName string, visitCount int, leadID int64) {
	data, _ := json.Marshal(map[string]interface{}{
		"lead_id":     leadID,
		"lead_name":   leadName,
		"visit_count": visitCount,
	})

	notification := &models.AdminNotification{
		Type:     "return_visitor",
		Title:    "🔁 High-Value Lead Returned",
		Message:  leadName + " is back on site (visit #" + string(rune(visitCount)) + ")",
		Priority: "normal",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) SendEngagementSpikeAlert(leadName string, eventCount int, leadID int64) {
	data, _ := json.Marshal(map[string]interface{}{
		"lead_id":     leadID,
		"lead_name":   leadName,
		"event_count": eventCount,
	})

	notification := &models.AdminNotification{
		Type:     "engagement_spike",
		Title:    "⚡ Engagement Spike Detected",
		Message:  leadName + " is very active right now",
		Priority: "high",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) SendPropertySavedAlert(leadName string, propertyAddress string, leadID int64, propertyID int64) {
	data, _ := json.Marshal(map[string]interface{}{
		"lead_id":          leadID,
		"lead_name":        leadName,
		"property_id":      propertyID,
		"property_address": propertyAddress,
	})

	notification := &models.AdminNotification{
		Type:     "property_saved",
		Title:    "💾 Property Saved",
		Message:  leadName + " saved " + propertyAddress,
		Priority: "normal",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) SendInquiryAlert(leadName string, propertyAddress string, leadID int64) {
	data, _ := json.Marshal(map[string]interface{}{
		"lead_id":          leadID,
		"lead_name":        leadName,
		"property_address": propertyAddress,
	})

	notification := &models.AdminNotification{
		Type:     "inquiry_sent",
		Title:    "💬 New Inquiry Received",
		Message:  leadName + " sent an inquiry about " + propertyAddress,
		Priority: "high",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) SendMultiplePropertiesAlert(leadName string, propertyCount int, leadID int64) {
	data, _ := json.Marshal(map[string]interface{}{
		"lead_id":        leadID,
		"lead_name":      leadName,
		"property_count": propertyCount,
	})

	notification := &models.AdminNotification{
		Type:     "multiple_properties",
		Title:    "👀 Browsing Multiple Properties",
		Message:  leadName + " viewed " + string(rune(propertyCount)) + "+ properties",
		Priority: "normal",
		Data:     data,
	}

	h.Broadcast(notification)
}

func (h *AdminNotificationHub) GetRecentNotifications(limit int) ([]models.AdminNotification, error) {
	var notifications []models.AdminNotification
	err := h.db.Order("created_at DESC").Limit(limit).Find(&notifications).Error
	return notifications, err
}

func (h *AdminNotificationHub) GetUnreadCount() (int64, error) {
	var count int64
	err := h.db.Model(&models.AdminNotification{}).Where("read_at IS NULL").Count(&count).Error
	return count, err
}

func (h *AdminNotificationHub) MarkAsRead(id uint) error {
	now := time.Now()
	return h.db.Model(&models.AdminNotification{}).Where("id = ?", id).Update("read_at", now).Error
}

func (h *AdminNotificationHub) MarkAllAsRead() error {
	now := time.Now()
	return h.db.Model(&models.AdminNotification{}).Where("read_at IS NULL").Update("read_at", now).Error
}
