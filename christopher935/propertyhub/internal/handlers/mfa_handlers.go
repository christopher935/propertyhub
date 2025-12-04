package handlers

import (
	"log"
	"encoding/json"
	"net/http"
	"strconv"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/auth"
	"chrisgross-ctrl-project/internal/security"
)

type MFAHandler struct {
	db          *gorm.DB
	authManager auth.AuthenticationManager
	totpManager *security.TOTPManager
}

func NewMFAHandler(db *gorm.DB, authManager auth.AuthenticationManager) *MFAHandler {
	return &MFAHandler{
		db:          db,
		authManager: authManager,
		totpManager: security.NewTOTPManager(db),
	}
}

// SetupMFA handles MFA setup for users
func (h *MFAHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if MFA is already enabled
	if h.totpManager.IsMFAEnabled(uint(userID)) {
		http.Error(w, "MFA already enabled", http.StatusConflict)
		return
	}

	// Get user email for TOTP generation
	user, err := h.authManager.GetUserByID(userIDStr)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Generate secret and backup codes
	key, backupCodes, err := h.totpManager.GenerateSecret(uint(userID), user.Email)
	if err != nil {
		http.Error(w, "Failed to generate MFA secret", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"secret":       key.Secret(),
		"backup_codes": backupCodes,
		"qr_url":       key.URL(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyMFA handles MFA verification
func (h *MFAHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get IP and User Agent for audit
	ipAddress := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")

	// Verify TOTP code
	valid, err := h.totpManager.VerifyTOTP(uint(userID), request.Code, ipAddress, userAgent)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusInternalServerError)
		return
	}

	if !valid {
		// Try backup code verification
		validBackup, err := h.totpManager.VerifyBackupCode(uint(userID), request.Code, ipAddress, userAgent)
		if err != nil || !validBackup {
			http.Error(w, "Invalid code", http.StatusUnauthorized)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "MFA verification successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMFAStatus returns MFA status for user
func (h *MFAHandler) GetMFAStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get MFA status using actual TOTPManager methods
	status := h.totpManager.GetMFAStatus(uint(userID))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// DisableMFA handles MFA disabling
func (h *MFAHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.totpManager.DisableMFA(uint(userID))
	if err != nil {
		http.Error(w, "Failed to disable MFA", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "MFA disabled successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterMFARoutes registers all MFA routes
func RegisterMFARoutes(mux *http.ServeMux, db *gorm.DB, authManager auth.AuthenticationManager) {
	handler := NewMFAHandler(db, authManager)

	// MFA management routes
	mux.Handle("/api/v1/mfa/setup", authManager.RequireAuth(http.HandlerFunc(handler.SetupMFA)))
	mux.Handle("/api/v1/mfa/verify", authManager.RequireAuth(http.HandlerFunc(handler.VerifyMFA)))
	mux.Handle("/api/v1/mfa/status", authManager.RequireAuth(http.HandlerFunc(handler.GetMFAStatus)))
	mux.Handle("/api/v1/mfa/disable", authManager.RequireAuth(http.HandlerFunc(handler.DisableMFA)))

	log.Println("🔐 Enterprise MFA routes registered")
}
