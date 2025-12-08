package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type AdminNotificationHandler struct {
	hub *services.AdminNotificationHub
	db  *gorm.DB
}

func NewAdminNotificationHandler(hub *services.AdminNotificationHub, db *gorm.DB) *AdminNotificationHandler {
	return &AdminNotificationHandler{
		hub: hub,
		db:  db,
	}
}

var notificationUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *AdminNotificationHandler) HandleWebSocket(c *gin.Context) {
	conn, err := notificationUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Notification WebSocket upgrade error: %v", err)
		return
	}

	h.hub.Register(conn)

	go func() {
		defer h.hub.Unregister(conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("❌ Notification WebSocket error: %v", err)
				}
				break
			}
		}
	}()

	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *AdminNotificationHandler) GetNotifications(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	notifications, err := h.hub.GetRecentNotifications(limit)
	if err != nil {
		log.Printf("❌ Error getting notifications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

func (h *AdminNotificationHandler) GetUnreadCount(c *gin.Context) {
	count, err := h.hub.GetUnreadCount()
	if err != nil {
		log.Printf("❌ Error getting unread count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

func (h *AdminNotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.hub.MarkAsRead(uint(id)); err != nil {
		log.Printf("❌ Error marking notification as read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminNotificationHandler) MarkAllAsRead(c *gin.Context) {
	if err := h.hub.MarkAllAsRead(); err != nil {
		log.Printf("❌ Error marking all notifications as read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
