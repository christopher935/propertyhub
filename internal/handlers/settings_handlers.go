package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"
)

// SettingsHandler handles all settings-related API endpoints
type SettingsHandler struct {
	db                *gorm.DB
	validator         *security.InputValidator
	encryptionManager *security.EncryptionManager
	storageService    *services.StorageService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	encryptionManager, err := security.NewEncryptionManager(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize encryption manager: %v", err)
	}

	storageService, err := services.NewStorageService()
	if err != nil {
		log.Printf("Warning: Failed to initialize storage service: %v", err)
	}

	return &SettingsHandler{
		db:                db,
		validator:         security.NewInputValidator(),
		encryptionManager: encryptionManager,
		storageService:    storageService,
	}
}

// secureDecodeJSON safely decodes JSON with size limits
func (h *SettingsHandler) secureDecodeJSON(r *http.Request, dst interface{}, maxSize int64) error {
	limitedReader := http.MaxBytesReader(nil, r.Body, maxSize)
	decoder := json.NewDecoder(limitedReader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

// sendSuccessResponse sends a success response
func (h *SettingsHandler) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// sendErrorResponse sends an error response
func (h *SettingsHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, errorCode string, message string) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    errorCode,
			"message": message,
		},
	})
}

// GetProfile handles GET /api/admin/settings/profile
func (h *SettingsHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		h.sendErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDStr.(string)

	var user models.AdminUser
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	var profile models.UserProfile
	err := h.db.Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		profile = models.UserProfile{UserID: userID}
		h.db.Create(&profile)
	} else if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch profile")
		return
	}

	// Return combined data
	h.sendSuccessResponse(w, map[string]interface{}{
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"first_name": profile.FirstName,
		"last_name":  profile.LastName,
		"phone":      profile.Phone,
		"company":    profile.Company,
		"department": profile.Department,
		"job_title":  profile.JobTitle,
		"avatar_url": profile.AvatarURL,
		"bio":        profile.Bio,
	})
}

// UpdateProfile handles POST /api/admin/settings/profile
func (h *SettingsHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		h.sendErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDStr.(string)

	var req struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Phone      string `json:"phone"`
		Company    string `json:"company"`
		Department string `json:"department"`
		JobTitle   string `json:"job_title"`
		Bio        string `json:"bio"`
	}

	if err := h.secureDecodeJSON(r, &req, 1024*1024); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON format")
		return
	}

	// Validate inputs
	inputMap := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"phone":      req.Phone,
		"company":    req.Company,
	}
	validationResult := h.validator.ValidateAll(inputMap)
	if !validationResult.IsValid {
		h.sendErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", validationResult.FirstError())
		return
	}

	// Update or create profile
	var profile models.UserProfile
	err := h.db.Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		profile = models.UserProfile{UserID: userID}
	} else if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch profile")
		return
	}

	profile.FirstName = req.FirstName
	profile.LastName = req.LastName
	profile.Phone = req.Phone
	profile.Company = req.Company
	profile.Department = req.Department
	profile.JobTitle = req.JobTitle
	profile.Bio = req.Bio

	if err := h.db.Save(&profile).Error; err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to save profile")
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Profile updated successfully"})
}

// UploadProfilePhoto handles POST /api/admin/settings/profile/photo
func (h *SettingsHandler) UploadProfilePhoto(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		h.sendErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDStr.(string)

	// Parse multipart form (10MB limit)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("photo")
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "MISSING_FILE", "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file
	if err := services.ValidateImageFile(header); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_FILE", err.Error())
		return
	}

	// Upload to Spaces
	url, err := h.storageService.UploadProfilePhoto(file, header.Filename, userID)
	if err != nil {
		log.Printf("Failed to upload profile photo: %v", err)
		h.sendErrorResponse(w, http.StatusInternalServerError, "UPLOAD_FAILED", "Failed to upload photo")
		return
	}

	// Get or create profile
	var profile models.UserProfile
	err = h.db.Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		profile = models.UserProfile{UserID: userID}
	} else if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch profile")
		return
	}

	// Delete old photo if exists
	if profile.AvatarURL != "" {
		h.storageService.DeleteFile(profile.AvatarURL)
	}

	// Save new photo URL
	profile.AvatarURL = url
	if err := h.db.Save(&profile).Error; err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to save profile")
		return
	}

	h.sendSuccessResponse(w, map[string]string{"avatar_url": url})
}

// GetPreferences handles GET /api/admin/settings/preferences
func (h *SettingsHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		h.sendErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDStr.(string)

	var prefs models.UserPreferences
	err := h.db.Where("user_id = ?", userID).First(&prefs).Error
	if err == gorm.ErrRecordNotFound {
		// Create default preferences
		prefs = models.UserPreferences{
			UserID:               userID,
			Timezone:             "America/Chicago",
			Language:             "en",
			DateFormat:           "MM/DD/YYYY",
			TimeFormat:           "12h",
			EmailNotifications:   true,
			SMSNotifications:     false,
			DesktopNotifications: true,
			WeeklyReports:        true,
			Theme:                "light",
		}
		h.db.Create(&prefs)
	} else if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch preferences")
		return
	}

	h.sendSuccessResponse(w, prefs)
}

// UpdatePreferences handles POST /api/admin/settings/preferences
func (h *SettingsHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		h.sendErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDStr.(string)

	var req models.UserPreferences
	if err := h.secureDecodeJSON(r, &req, 1024*1024); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON format")
		return
	}

	// Get or create preferences
	var prefs models.UserPreferences
	err := h.db.Where("user_id = ?", userID).First(&prefs).Error
	if err == gorm.ErrRecordNotFound {
		prefs = models.UserPreferences{UserID: userID}
	} else if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch preferences")
		return
	}

	// Update fields
	prefs.Timezone = req.Timezone
	prefs.Language = req.Language
	prefs.DateFormat = req.DateFormat
	prefs.TimeFormat = req.TimeFormat
	prefs.EmailNotifications = req.EmailNotifications
	prefs.SMSNotifications = req.SMSNotifications
	prefs.DesktopNotifications = req.DesktopNotifications
	prefs.WeeklyReports = req.WeeklyReports
	prefs.Theme = req.Theme

	if err := h.db.Save(&prefs).Error; err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to save preferences")
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Preferences updated successfully"})
}

// ChangePassword handles POST /api/admin/settings/password
func (h *SettingsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		h.sendErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDStr.(string)

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := h.secureDecodeJSON(r, &req, 1024*1024); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON format")
		return
	}

	// Validate password strength (min 8 chars)
	if len(req.NewPassword) < 8 {
		h.sendErrorResponse(w, http.StatusBadRequest, "WEAK_PASSWORD", "Password must be at least 8 characters")
		return
	}

	var user models.AdminUser
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	// Verify current password
	if !security.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		h.sendErrorResponse(w, http.StatusUnauthorized, "INVALID_PASSWORD", "Current password is incorrect")
		return
	}

	// Hash new password
	newHash, err := security.HashPassword(req.NewPassword)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "HASH_ERROR", "Failed to hash password")
		return
	}

	// Update password
	if err := h.db.Table("admin_users").Where("id = ?", userID).Update("password_hash", newHash).Error; err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update password")
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Password changed successfully"})
}
