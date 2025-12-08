package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/auth"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"gorm.io/gorm"
)

// GetDashboardStats provides dashboard statistics
func GetDashboardStats(w http.ResponseWriter, r *http.Request, db *gorm.DB, authManager auth.AuthenticationManager) {
	w.Header().Set("Content-Type", "application/json")

	stats := map[string]interface{}{
		"total_properties": 0,
		"active_bookings":  0,
		"total_contacts":   0,
		"recent_activity":  0,
		"system_health":    "operational",
		"last_updated":     "just now",
	}

	if db != nil {
		var propertyCount, bookingCount, contactCount int64
		db.Model(&models.Property{}).Where("status = ?", "active").Count(&propertyCount)
		db.Model(&models.Booking{}).Where("status = ?", "scheduled").Count(&bookingCount)
		db.Model(&models.Contact{}).Count(&contactCount)

		stats["total_properties"] = propertyCount
		stats["active_bookings"] = bookingCount
		stats["total_contacts"] = contactCount
		stats["recent_activity"] = propertyCount + bookingCount + contactCount
	}

	if authManager != nil {
		stats["active_sessions"] = authManager.GetActiveSessionCount()
		stats["cache_hit_rate"] = authManager.GetCacheHitRate()
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetPropertyInsights provides property insights
func GetPropertyInsights(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	w.Header().Set("Content-Type", "application/json")

	insights := map[string]interface{}{
		"total_views":       0,
		"average_price":     0,
		"most_popular_type": "Unknown",
		"top_cities":        []string{},
		"last_updated":      "just now",
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    insights,
	})
}

// CreateContact handles contact form submissions
func CreateContact(w http.ResponseWriter, r *http.Request, db *gorm.DB, encryptionManager *security.EncryptionManager) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	message := r.FormValue("message")

	if firstName == "" || lastName == "" || email == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create contact with proper encryption
	var encryptedName, encryptedEmail, encryptedPhone security.EncryptedString

	if encryptionManager != nil {
		encryptedName, err = encryptionManager.Encrypt(fmt.Sprintf("%s %s", firstName, lastName))
		if err != nil {
			http.Error(w, "Encryption error", http.StatusInternalServerError)
			return
		}

		encryptedEmail, err = encryptionManager.Encrypt(email)
		if err != nil {
			http.Error(w, "Encryption error", http.StatusInternalServerError)
			return
		}

		encryptedPhone, err = encryptionManager.Encrypt(phone)
		if err != nil {
			http.Error(w, "Encryption error", http.StatusInternalServerError)
			return
		}
	} else {
		encryptedName = security.EncryptedString(fmt.Sprintf("%s %s", firstName, lastName))
		encryptedEmail = security.EncryptedString(email)
		encryptedPhone = security.EncryptedString(phone)
	}

	contact := models.Contact{
		Name:      encryptedName,
		Email:     encryptedEmail,
		Phone:     encryptedPhone,
		Message:   message,
		Status:    "new",
		Source:    "website_contact_form",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if db != nil {
		if err := db.Create(&contact).Error; err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Contact submitted successfully",
	})
}
