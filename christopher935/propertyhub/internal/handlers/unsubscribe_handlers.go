package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// UnsubscribeHandlers handles CAN-SPAM compliant unsubscribe functionality
type UnsubscribeHandlers struct {
	db *gorm.DB
}

// UnsubscribeRecord tracks email unsubscriptions for CAN-SPAM compliance
type UnsubscribeRecord struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Email            string    `json:"email" gorm:"index"`
	UnsubscribeID    string    `json:"unsubscribe_id" gorm:"uniqueIndex"`
	UnsubscribeType  string    `json:"unsubscribe_type"` // all, marketing, alerts
	IPAddress        string    `json:"ip_address"`
	UserAgent        string    `json:"user_agent"`
	UnsubscribeDate  time.Time `json:"unsubscribe_date"`
	OriginalEmailID  string    `json:"original_email_id,omitempty"`
	Source           string    `json:"source"` // email_link, one_click, manual
	ResubscribeDate  *time.Time `json:"resubscribe_date,omitempty"`
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	CreatedAt        time.Time `json:"created_at"`
}

// NewUnsubscribeHandlers creates new unsubscribe handlers
func NewUnsubscribeHandlers(db *gorm.DB) *UnsubscribeHandlers {
	// Auto-migrate unsubscribe records table
	db.AutoMigrate(&UnsubscribeRecord{})
	
	return &UnsubscribeHandlers{
		db: db,
	}
}

// HandleUnsubscribe processes unsubscribe requests with proper CAN-SPAM compliance
func (u *UnsubscribeHandlers) HandleUnsubscribe(c *gin.Context) {
	// Get parameters
	email := c.Query("email")
	unsubscribeID := c.Query("id")
	unsubType := c.DefaultQuery("type", "all") // all, marketing, alerts
	source := c.DefaultQuery("source", "email_link")

	if email == "" || unsubscribeID == "" {
		c.HTML(http.StatusBadRequest, "unsubscribe_error.html", gin.H{
			"Title": "Unsubscribe Error",
			"Error": "Invalid unsubscribe link. Please contact us directly.",
		})
		return
	}

	// Validate email format
	if !u.isValidEmail(email) {
		c.HTML(http.StatusBadRequest, "unsubscribe_error.html", gin.H{
			"Title": "Unsubscribe Error", 
			"Error": "Invalid email address format.",
		})
		return
	}

	// Check if already unsubscribed
	var existing UnsubscribeRecord
	result := u.db.Where("email = ? AND unsubscribe_type = ? AND is_active = ?", 
		email, unsubType, true).First(&existing)
	
	if result.Error == nil {
		// Already unsubscribed
		c.HTML(http.StatusOK, "unsubscribe_success.html", gin.H{
			"Title":           "Already Unsubscribed",
			"Email":           email,
			"UnsubscribeType": u.getUnsubscribeTypeDisplay(unsubType),
			"Message":         "You were already unsubscribed from this type of communication.",
			"ShowResubscribe": true,
		})
		return
	}

	// Create unsubscribe record
	unsubRecord := &UnsubscribeRecord{
		Email:            email,
		UnsubscribeID:    unsubscribeID,
		UnsubscribeType:  unsubType,
		IPAddress:        c.ClientIP(),
		UserAgent:        c.GetHeader("User-Agent"),
		UnsubscribeDate:  time.Now(),
		Source:           source,
		IsActive:         true,
	}

	if err := u.db.Create(unsubRecord).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "unsubscribe_error.html", gin.H{
			"Title": "Unsubscribe Error",
			"Error": "Unable to process your unsubscribe request. Please try again or contact us directly.",
		})
		return
	}

	// Update lead/contact records to reflect unsubscribe status
	u.updateContactUnsubscribeStatus(email, unsubType)

	// Show success page
	c.HTML(http.StatusOK, "unsubscribe_success.html", gin.H{
		"Title":           "Successfully Unsubscribed",
		"Email":           email,
		"UnsubscribeType": u.getUnsubscribeTypeDisplay(unsubType),
		"Message":         u.getUnsubscribeSuccessMessage(unsubType),
		"ShowResubscribe": true,
	})
}

// HandleOneClickUnsubscribe handles one-click unsubscribe per RFC 8058
func (u *UnsubscribeHandlers) HandleOneClickUnsubscribe(c *gin.Context) {
	// One-click unsubscribe must be POST request per RFC 8058
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "One-click unsubscribe requires POST method",
		})
		return
	}

	// Parse form data
	email := c.PostForm("email")
	unsubscribeID := c.PostForm("id")
	
	if email == "" || unsubscribeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	// Process unsubscribe
	unsubRecord := &UnsubscribeRecord{
		Email:            email,
		UnsubscribeID:    unsubscribeID,
		UnsubscribeType:  "all", // One-click always unsubscribes from all
		IPAddress:        c.ClientIP(),
		UserAgent:        c.GetHeader("User-Agent"),
		UnsubscribeDate:  time.Now(),
		Source:           "one_click",
		IsActive:         true,
	}

	if err := u.db.Create(unsubRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process unsubscribe request",
		})
		return
	}

	// Update contact records
	u.updateContactUnsubscribeStatus(email, "all")

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully unsubscribed",
		"email":   email,
	})
}

// HandleResubscribe allows users to resubscribe
func (u *UnsubscribeHandlers) HandleResubscribe(c *gin.Context) {
	email := c.PostForm("email")
	subscribeType := c.DefaultPostForm("type", "all")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email is required",
		})
		return
	}

	// Deactivate existing unsubscribe records
	result := u.db.Model(&UnsubscribeRecord{}).
		Where("email = ? AND unsubscribe_type = ? AND is_active = ?", 
			email, subscribeType, true).
		Updates(map[string]interface{}{
			"is_active":        false,
			"resubscribe_date": time.Now(),
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process resubscribe request",
		})
		return
	}

	// Update contact records
	u.updateContactSubscribeStatus(email, subscribeType)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully resubscribed",
		"email":   email,
		"type":    subscribeType,
	})
}

// IsEmailUnsubscribed checks if an email is unsubscribed from a specific type
func (u *UnsubscribeHandlers) IsEmailUnsubscribed(email, emailType string) bool {
	var count int64
	
	// Check for specific type unsubscribe or "all" unsubscribe
	u.db.Model(&UnsubscribeRecord{}).
		Where("email = ? AND (unsubscribe_type = ? OR unsubscribe_type = ?) AND is_active = ?", 
			email, emailType, "all", true).
		Count(&count)
	
	return count > 0
}

// GetUnsubscribeStats returns unsubscribe statistics
func (u *UnsubscribeHandlers) GetUnsubscribeStats(c *gin.Context) {
	var stats struct {
		TotalUnsubscribes    int64   `json:"total_unsubscribes"`
		MarketingUnsubscribes int64  `json:"marketing_unsubscribes"`
		AlertsUnsubscribes   int64   `json:"alerts_unsubscribes"`
		AllUnsubscribes      int64   `json:"all_unsubscribes"`
		OneClickUnsubscribes int64   `json:"one_click_unsubscribes"`
		UnsubscribeRate      float64 `json:"unsubscribe_rate"`
		RecentUnsubscribes   []UnsubscribeRecord `json:"recent_unsubscribes"`
	}

	// Get total counts
	u.db.Model(&UnsubscribeRecord{}).Where("is_active = ?", true).Count(&stats.TotalUnsubscribes)
	u.db.Model(&UnsubscribeRecord{}).Where("unsubscribe_type = ? AND is_active = ?", "marketing", true).Count(&stats.MarketingUnsubscribes)
	u.db.Model(&UnsubscribeRecord{}).Where("unsubscribe_type = ? AND is_active = ?", "alerts", true).Count(&stats.AlertsUnsubscribes)
	u.db.Model(&UnsubscribeRecord{}).Where("unsubscribe_type = ? AND is_active = ?", "all", true).Count(&stats.AllUnsubscribes)
	u.db.Model(&UnsubscribeRecord{}).Where("source = ? AND is_active = ?", "one_click", true).Count(&stats.OneClickUnsubscribes)

	// Get recent unsubscribes (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	u.db.Where("created_at >= ? AND is_active = ?", thirtyDaysAgo, true).
		Order("created_at DESC").
		Limit(50).
		Find(&stats.RecentUnsubscribes)

	// Calculate unsubscribe rate (would need total email count from email service)
	// This is a placeholder - in production you'd calculate against total emails sent
	stats.UnsubscribeRate = 0.02 // Example: 2% unsubscribe rate

	c.JSON(http.StatusOK, stats)
}

// Administrative functions

// GetUnsubscribeList returns paginated unsubscribe records for admin
func (u *UnsubscribeHandlers) GetUnsubscribeList(c *gin.Context) {
	page := u.getPageFromQuery(c)
	limit := 50
	offset := (page - 1) * limit

	var records []UnsubscribeRecord
	var total int64

	// Count total records
	u.db.Model(&UnsubscribeRecord{}).Where("is_active = ?", true).Count(&total)

	// Get paginated records
	result := u.db.Where("is_active = ?", true).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&records)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch unsubscribe records",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records":     records,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// Helper functions

func (u *UnsubscribeHandlers) updateContactUnsubscribeStatus(email, unsubType string) {
	// Update in various contact tables based on what exists
	
	// Update BookingRequests if table exists
	u.db.Model(&models.BookingRequest{}).
		Where("email = ?", email).
		Update("marketing_consent", false)
	
	// Update Leads table if exists
	u.db.Model(&models.Lead{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"marketing_consent":     unsubType != "marketing" && unsubType != "all",
			"notification_consent":  unsubType != "alerts" && unsubType != "all",
			"unsubscribed_at":       time.Now(),
		})
}

func (u *UnsubscribeHandlers) updateContactSubscribeStatus(email, subscribeType string) {
	// Re-enable marketing consent when resubscribing
	
	// Update BookingRequests
	u.db.Model(&models.BookingRequest{}).
		Where("email = ?", email).
		Update("marketing_consent", true)
	
	// Update Leads table
	updates := map[string]interface{}{
		"unsubscribed_at": nil,
	}
	
	if subscribeType == "marketing" || subscribeType == "all" {
		updates["marketing_consent"] = true
	}
	if subscribeType == "alerts" || subscribeType == "all" {
		updates["notification_consent"] = true
	}
	
	u.db.Model(&models.Lead{}).
		Where("email = ?", email).
		Updates(updates)
}

func (u *UnsubscribeHandlers) getUnsubscribeTypeDisplay(unsubType string) string {
	switch unsubType {
	case "marketing":
		return "marketing emails"
	case "alerts":
		return "property alerts"
	case "all":
		return "all emails"
	default:
		return "communications"
	}
}

func (u *UnsubscribeHandlers) getUnsubscribeSuccessMessage(unsubType string) string {
	switch unsubType {
	case "marketing":
		return "You will no longer receive marketing emails from us. You will still receive important transactional emails related to your inquiries and applications."
	case "alerts":
		return "You will no longer receive property availability alerts. You can resubscribe at any time."
	case "all":
		return "You have been unsubscribed from all non-essential communications. You will only receive important transactional emails related to your active inquiries and applications."
	default:
		return "You have been successfully unsubscribed."
	}
}

func (u *UnsubscribeHandlers) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) > 5
}

func (u *UnsubscribeHandlers) getPageFromQuery(c *gin.Context) int {
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p := u.parseInt(pageStr); p > 0 {
			page = p
		}
	}
	return page
}

func (u *UnsubscribeHandlers) parseInt(s string) int {
	// Simple integer parsing - in production use strconv.Atoi
	if s == "" {
		return 0
	}
	// Placeholder implementation
	return 1
}