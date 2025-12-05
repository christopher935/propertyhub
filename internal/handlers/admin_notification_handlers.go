package handlers

import (
	"net/http"
	"strconv"

	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminNotificationHandler struct {
	db      *gorm.DB
	service *services.AdminNotificationService
}

func NewAdminNotificationHandler(db *gorm.DB, service *services.AdminNotificationService) *AdminNotificationHandler {
	return &AdminNotificationHandler{
		db:      db,
		service: service,
	}
}

func (h *AdminNotificationHandler) GetNotifications(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	notifications, err := h.service.GetRecentNotifications(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch notifications",
		})
		return
	}

	notificationDicts := make([]map[string]interface{}, len(notifications))
	for i, notif := range notifications {
		notificationDicts[i] = notif.ToDict()
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"notifications": notificationDicts,
		"total":         len(notificationDicts),
	})
}

func (h *AdminNotificationHandler) GetUnreadCount(c *gin.Context) {
	count, err := h.service.GetUnreadCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch unread count",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"unread_count":  count,
	})
}

func (h *AdminNotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid notification ID",
		})
		return
	}

	if err := h.service.MarkAsRead(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark notification as read",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification marked as read",
	})
}

func (h *AdminNotificationHandler) Dismiss(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid notification ID",
		})
		return
	}

	if err := h.service.Dismiss(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to dismiss notification",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification dismissed",
	})
}

func (h *AdminNotificationHandler) DismissAll(c *gin.Context) {
	if err := h.service.DismissAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to dismiss all notifications",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All notifications dismissed",
	})
}
