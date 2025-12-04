package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/config"
        "chrisgross-ctrl-project/internal/scraper"
	"log"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
)

// PreListingHandlers handles pre-listing workflow with AI property valuation and email processing
type PreListingHandlers struct {
	db                *gorm.DB
	valuationService  *services.PropertyValuationService
	preListingService *services.PreListingService
	emailProcessor    *services.EmailProcessor
	config            *config.Config
}

// NewPreListingHandlers creates new pre-listing handlers with all services
func NewPreListingHandlers(db *gorm.DB, config *config.Config, scraperService *scraper.ScraperService) *PreListingHandlers {
	return &PreListingHandlers{
		db:                db,
		valuationService:  services.NewPropertyValuationService(config, scraperService),
		preListingService: services.NewPreListingService(db),
		emailProcessor:    services.NewEmailProcessor(db),
		config:            config,
	}
}

// GetPropertyValuation provides AI-powered property valuation for pre-listing
// POST /api/v1/pre-listing/valuation
func (plh *PreListingHandlers) GetPropertyValuation(c *gin.Context) {
	var request services.PropertyValuationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Perform property valuation
	valuation, err := plh.valuationService.ValuateProperty(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to valuate property",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"valuation": valuation,
		"message":   "Property valuation completed successfully",
	})
}

// GetPreListingStats returns dashboard statistics
func (h *PreListingHandlers) GetPreListingStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.preListingService.GetPreListingStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetPreListingItems returns paginated list of pre-listing items
func (h *PreListingHandlers) GetPreListingItems(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	status := r.URL.Query().Get("status")

	items, total, err := h.preListingService.GetPreListingItems(page, limit, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items":       items,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetPreListingItem returns a specific pre-listing item
func (h *PreListingHandlers) GetPreListingItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/pre-listing/items/")
	idStr := strings.Split(path, "/")[0]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var item models.PreListingItem
	if err := h.db.Preload("EmailRecords").First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Pre-listing item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    item,
	})
}

// ProcessEmail handles incoming email processing
func (h *PreListingHandlers) ProcessEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req services.EmailProcessingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set received time if not provided
	if req.ReceivedAt.IsZero() {
		req.ReceivedAt = time.Now()
	}

	result, err := h.emailProcessor.ProcessEmail(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// GetEmailProcessingStats returns email processing statistics
func (h *PreListingHandlers) GetEmailProcessingStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.emailProcessor.GetEmailProcessingStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetIncomingEmails returns paginated list of incoming emails
func (h *PreListingHandlers) GetIncomingEmails(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	emailType := r.URL.Query().Get("email_type")
	status := r.URL.Query().Get("status")

	query := h.db.Model(&models.IncomingEmail{})

	if emailType != "" {
		query = query.Where("email_type = ?", emailType)
	}
	if status != "" {
		query = query.Where("processing_status = ?", status)
	}

	var total int64
	query.Count(&total)

	var emails []models.IncomingEmail
	offset := (page - 1) * limit
	err := query.Preload("PreListingItem").
		Order("received_at DESC").
		Limit(limit).Offset(offset).
		Find(&emails).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"emails":      emails,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CheckOverdueItems manually triggers overdue checking
func (h *PreListingHandlers) CheckOverdueItems(w http.ResponseWriter, r *http.Request) {
	if err := h.preListingService.CheckOverdueItems(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Overdue check completed",
	})
}

// GetOverdueItems returns items that are overdue
func (h *PreListingHandlers) GetOverdueItems(w http.ResponseWriter, r *http.Request) {
	var items []models.PreListingItem
	err := h.db.Where("is_overdue = ?", true).
		Order("created_at DESC").
		Find(&items).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    items,
	})
}

// GetAlerts returns email alerts
func (h *PreListingHandlers) GetAlerts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	alertType := r.URL.Query().Get("alert_type")
	resolved := r.URL.Query().Get("resolved")

	query := h.db.Model(&models.EmailAlert{})

	if alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}
	if resolved != "" {
		isResolved := resolved == "true"
		query = query.Where("is_resolved = ?", isResolved)
	}

	var total int64
	query.Count(&total)

	var alerts []models.EmailAlert
	offset := (page - 1) * limit
	err := query.Preload("PreListingItem").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&alerts).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"alerts":      alerts,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateManualPreListing creates a pre-listing item manually
func (h *PreListingHandlers) CreateManualPreListing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address           string     `json:"address"`
		City              string     `json:"city"`
		State             string     `json:"state"`
		ZipCode           string     `json:"zip_code"`
		TargetListingDate *time.Time `json:"target_listing_date"`
		Notes             string     `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	item := &models.PreListingItem{
		Address:           req.Address,
		City:              req.City,
		State:             req.State,
		ZipCode:           req.ZipCode,
		TargetListingDate: req.TargetListingDate,
		AdminNotes:        req.Notes,
		Status:            "email_received",
		ManualOverride:    true,
		OverrideReason:    "Manually created",
	}

	// Set full address
	if req.City != "" && req.State != "" {
		item.FullAddress = req.Address + ", " + req.City + ", " + req.State
		if req.ZipCode != "" {
			item.FullAddress += " " + req.ZipCode
		}
	} else {
		item.FullAddress = req.Address
	}

	if err := h.db.Create(item).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    item,
	})
}

// UpdatePreListingItem updates a pre-listing item
func (h *PreListingHandlers) UpdatePreListingItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/pre-listing/items/")
	idStr := strings.Split(path, "/")[0]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var item models.PreListingItem
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Pre-listing item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update allowed fields
	if status, ok := updateData["status"].(string); ok {
		item.Status = status
	}
	if notes, ok := updateData["admin_notes"].(string); ok {
		item.AdminNotes = notes
	}
	if override, ok := updateData["manual_override"].(bool); ok {
		item.ManualOverride = override
	}
	if reason, ok := updateData["override_reason"].(string); ok {
		item.OverrideReason = reason
	}

	// Handle date fields
	if targetDate, ok := updateData["target_listing_date"].(string); ok && targetDate != "" {
		if parsed, err := time.Parse(time.RFC3339, targetDate); err == nil {
			item.TargetListingDate = &parsed
		}
	}

	if err := h.db.Save(&item).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    item,
	})
}

// DeletePreListingItem deletes a pre-listing item
func (h *PreListingHandlers) DeletePreListingItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/pre-listing/items/")
	idStr := strings.Split(path, "/")[0]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.db.Delete(&models.PreListingItem{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Pre-listing item deleted",
	})
}

// GetPreListingTimeline returns timeline events for a pre-listing item
func (h *PreListingHandlers) GetPreListingTimeline(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/pre-listing/items/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get the pre-listing item
	var item models.PreListingItem
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Pre-listing item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get related emails
	var emails []models.IncomingEmail
	h.db.Where("pre_listing_item_id = ?", id).
		Order("received_at ASC").
		Find(&emails)

	// Build timeline
	timeline := []map[string]interface{}{}

	// Add creation event
	timeline = append(timeline, map[string]interface{}{
		"type":        "created",
		"timestamp":   item.CreatedAt,
		"title":       "Pre-listing created",
		"description": "Pre-listing item created in system",
	})

	// Add email events
	for _, email := range emails {
		timeline = append(timeline, map[string]interface{}{
			"type":        "email",
			"timestamp":   email.ReceivedAt,
			"title":       email.Subject,
			"description": "Email received from " + email.FromEmail,
			"email_type":  email.EmailType,
			"confidence":  email.Confidence,
		})
	}

	// Add status change events based on dates
	if item.LockboxPlacedDate != nil {
		timeline = append(timeline, map[string]interface{}{
			"type":        "lockbox_placed",
			"timestamp":   *item.LockboxPlacedDate,
			"title":       "Lockbox placed",
			"description": "Lockbox placed at property",
		})
	}

	if item.PhotoScheduledDate != nil {
		timeline = append(timeline, map[string]interface{}{
			"type":        "photos_scheduled",
			"timestamp":   *item.PhotoScheduledDate,
			"title":       "Photos scheduled",
			"description": "Photography session scheduled",
		})
	}

	if item.PhotoCompletedDate != nil {
		timeline = append(timeline, map[string]interface{}{
			"type":        "photos_completed",
			"timestamp":   *item.PhotoCompletedDate,
			"title":       "Photos completed",
			"description": "Photography session completed",
		})
	}

	if item.MLSListedDate != nil {
		timeline = append(timeline, map[string]interface{}{
			"type":        "listed",
			"timestamp":   *item.MLSListedDate,
			"title":       "Property listed",
			"description": "Property listed on MLS",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"item":     item,
			"timeline": timeline,
		},
	})
}

// ReprocessEmail reprocesses a failed or low-confidence email
func (h *PreListingHandlers) ReprocessEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/pre-listing/emails/")
	idStr := strings.TrimSuffix(path, "/reprocess")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var email models.IncomingEmail
	if err := h.db.First(&email, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Email not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Create processing request from stored email
	req := &services.EmailProcessingRequest{
		From:       email.FromEmail,
		To:         email.ToEmail,
		Subject:    email.Subject,
		Content:    email.Content,
		ReceivedAt: email.ReceivedAt,
	}

	result, err := h.emailProcessor.ProcessEmail(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// ResolveAlert marks an alert as resolved
func (h *PreListingHandlers) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/pre-listing/alerts/")
	idStr := strings.TrimSuffix(path, "/resolve")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var alert models.EmailAlert
	if err := h.db.First(&alert, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Alert not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	now := time.Now()
	alert.IsResolved = true
	alert.ResolvedAt = &now

	if err := h.db.Save(&alert).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    alert,
	})
}

// RegisterPreListingRoutes registers all pre-listing workflow routes
func RegisterPreListingRoutes(mux *http.ServeMux, db *gorm.DB, config *config.Config, scraperService *scraper.ScraperService) {
	handler := NewPreListingHandlers(db, config, scraperService)
	
	// Pre-listing management endpoints
	mux.HandleFunc("/api/v1/pre-listing/stats", handler.GetPreListingStats)
	mux.HandleFunc("/api/v1/pre-listing/items", handler.GetPreListingItems)
	mux.HandleFunc("/api/v1/pre-listing/items/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && len(r.URL.Path) > len("/api/v1/pre-listing/items/") {
			handler.GetPreListingItem(w, r)
		} else if r.Method == "PUT" {
			handler.UpdatePreListingItem(w, r)
		} else if r.Method == "DELETE" {
			handler.DeletePreListingItem(w, r)
		}
	})
	mux.HandleFunc("/api/v1/pre-listing/create", handler.CreateManualPreListing)
	mux.HandleFunc("/api/v1/pre-listing/timeline/", handler.GetPreListingTimeline)
	
	// Email processing endpoints
	mux.HandleFunc("/api/v1/pre-listing/email/process", handler.ProcessEmail)
	mux.HandleFunc("/api/v1/pre-listing/email/stats", handler.GetEmailProcessingStats)
	mux.HandleFunc("/api/v1/pre-listing/email/incoming", handler.GetIncomingEmails)
	mux.HandleFunc("/api/v1/pre-listing/email/reprocess/", handler.ReprocessEmail)
	
	// Overdue and alert management
	mux.HandleFunc("/api/v1/pre-listing/overdue/check", handler.CheckOverdueItems)
	mux.HandleFunc("/api/v1/pre-listing/overdue/items", handler.GetOverdueItems)
	mux.HandleFunc("/api/v1/pre-listing/alerts", handler.GetAlerts)
	mux.HandleFunc("/api/v1/pre-listing/alerts/resolve/", handler.ResolveAlert)
	
	log.Println("✅ Pre-listing workflow routes registered successfully")
}
