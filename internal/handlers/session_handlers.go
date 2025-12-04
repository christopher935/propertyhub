package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/security"
)

type SessionHandler struct {
	db             *gorm.DB
	sessionManager *security.SessionManager
}

func NewSessionHandler(db *gorm.DB) *SessionHandler {
	return &SessionHandler{
		db:             db,
		sessionManager: security.NewSessionManager(db),
	}
}

// GetUserSessions returns active sessions for a user
func (h *SessionHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	sessions, err := h.sessionManager.GetUserSessions(uint(userID))
	if err != nil {
		log.Printf("Error fetching user sessions: %v", err)
		http.Error(w, "Failed to fetch sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    sessions,
	})
}

// GetAllActiveSessions returns all active sessions (admin only)
func (h *SessionHandler) GetAllActiveSessions(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	var sessions []security.Session
	var total int64

	query := h.db.Model(&security.Session{}).Where("is_active = true AND expires_at > ?", time.Now())

	// Apply filters
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			query = query.Where("user_id = ?", uint(id))
		}
	}

	if ipAddress := r.URL.Query().Get("ip_address"); ipAddress != "" {
		query = query.Where("ip_address = ?", ipAddress)
	}

	if riskLevel := r.URL.Query().Get("risk_level"); riskLevel != "" {
		switch riskLevel {
		case "low":
			query = query.Where("risk_score < 30")
		case "medium":
			query = query.Where("risk_score >= 30 AND risk_score < 60")
		case "high":
			query = query.Where("risk_score >= 60")
		}
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("last_activity DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	if err != nil {
		log.Printf("Error fetching active sessions: %v", err)
		http.Error(w, "Failed to fetch sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"sessions": sessions,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
		},
	})
}

// InvalidateSession invalidates a specific session
func (h *SessionHandler) InvalidateSession(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		SessionID string `json:"session_id"`
		Reason    string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.SessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	if requestData.Reason == "" {
		requestData.Reason = "Admin action"
	}

	if err := h.sessionManager.InvalidateSession(requestData.SessionID, requestData.Reason); err != nil {
		log.Printf("Error invalidating session: %v", err)
		http.Error(w, "Failed to invalidate session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Session invalidated successfully",
	})
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (h *SessionHandler) InvalidateAllUserSessions(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		UserID uint   `json:"user_id"`
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.UserID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if requestData.Reason == "" {
		requestData.Reason = "Admin action"
	}

	if err := h.sessionManager.InvalidateAllUserSessions(requestData.UserID, requestData.Reason); err != nil {
		log.Printf("Error invalidating user sessions: %v", err)
		http.Error(w, "Failed to invalidate user sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All user sessions invalidated successfully",
	})
}

// GetSessionStatistics returns session statistics
func (h *SessionHandler) GetSessionStatistics(w http.ResponseWriter, r *http.Request) {
	stats := h.sessionManager.GetSessionStatistics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetDeviceInfo returns device information for users
func (h *SessionHandler) GetDeviceInfo(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	var devices []security.DeviceInfo
	var total int64

	query := h.db.Model(&security.DeviceInfo{})

	// Apply filters
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			query = query.Where("user_id = ?", uint(id))
		}
	}

	if deviceType := r.URL.Query().Get("device_type"); deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}

	if trusted := r.URL.Query().Get("trusted"); trusted != "" {
		query = query.Where("is_trusted = ?", trusted == "true")
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("last_seen DESC").Limit(limit).Offset(offset).Find(&devices).Error
	if err != nil {
		log.Printf("Error fetching device info: %v", err)
		http.Error(w, "Failed to fetch device info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"devices": devices,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		},
	})
}

// TrustDevice marks a device as trusted
func (h *SessionHandler) TrustDevice(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		DeviceFingerprint string `json:"device_fingerprint"`
		UserID            uint   `json:"user_id"`
		Trusted           bool   `json:"trusted"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.DeviceFingerprint == "" || requestData.UserID == 0 {
		http.Error(w, "Device fingerprint and user ID are required", http.StatusBadRequest)
		return
	}

	// Update device trust status
	err := h.db.Model(&security.DeviceInfo{}).
		Where("device_fingerprint = ? AND user_id = ?", requestData.DeviceFingerprint, requestData.UserID).
		Update("is_trusted", requestData.Trusted).Error

	if err != nil {
		log.Printf("Error updating device trust: %v", err)
		http.Error(w, "Failed to update device trust", http.StatusInternalServerError)
		return
	}

	// Update related sessions
	h.db.Model(&security.Session{}).
		Where("device_fingerprint = ? AND user_id = ? AND is_active = true", requestData.DeviceFingerprint, requestData.UserID).
		Update("is_trusted", requestData.Trusted)

	action := "untrusted"
	if requestData.Trusted {
		action = "trusted"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Device " + action + " successfully",
	})
}

// GetSuspiciousActivities returns suspicious activities
func (h *SessionHandler) GetSuspiciousActivities(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	var activities []security.SuspiciousActivity
	var total int64

	query := h.db.Model(&security.SuspiciousActivity{})

	// Apply filters
	if resolved := r.URL.Query().Get("resolved"); resolved != "" {
		query = query.Where("is_resolved = ?", resolved == "true")
	}

	if activityType := r.URL.Query().Get("activity_type"); activityType != "" {
		query = query.Where("activity_type = ?", activityType)
	}

	if minRiskScore := r.URL.Query().Get("min_risk_score"); minRiskScore != "" {
		if score, err := strconv.Atoi(minRiskScore); err == nil {
			query = query.Where("risk_score >= ?", score)
		}
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&activities).Error
	if err != nil {
		log.Printf("Error fetching suspicious activities: %v", err)
		http.Error(w, "Failed to fetch suspicious activities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"activities": activities,
			"total":      total,
			"limit":      limit,
			"offset":     offset,
		},
	})
}

// ResolveSuspiciousActivity marks a suspicious activity as resolved
func (h *SessionHandler) ResolveSuspiciousActivity(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		ActivityID uint     `json:"activity_id"`
		Actions    []string `json:"actions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.ActivityID == 0 {
		http.Error(w, "Activity ID is required", http.StatusBadRequest)
		return
	}

	// Update suspicious activity
	now := time.Now()
	updates := map[string]interface{}{
		"is_resolved": true,
		"resolved_at": now,
		"actions":     requestData.Actions,
	}

	if err := h.db.Model(&security.SuspiciousActivity{}).Where("id = ?", requestData.ActivityID).Updates(updates).Error; err != nil {
		log.Printf("Error resolving suspicious activity: %v", err)
		http.Error(w, "Failed to resolve suspicious activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Suspicious activity resolved successfully",
	})
}

// CleanupExpiredSessions removes expired sessions
func (h *SessionHandler) CleanupExpiredSessions(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionManager.CleanupExpiredSessions(); err != nil {
		log.Printf("Error cleaning up expired sessions: %v", err)
		http.Error(w, "Failed to cleanup expired sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Expired sessions cleaned up successfully",
	})
}

// GetSessionActivity returns session activity logs
func (h *SessionHandler) GetSessionActivity(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	var activities []security.SessionActivity
	err := h.db.Where("session_id = ?", sessionID).Order("created_at DESC").Limit(limit).Find(&activities).Error
	if err != nil {
		log.Printf("Error fetching session activity: %v", err)
		http.Error(w, "Failed to fetch session activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    activities,
	})
}

// RegisterSessionRoutes registers all session-related routes
func (h *SessionHandler) RegisterSessionRoutes(mux *http.ServeMux, authManager interface{}) {
	// Type assertion to get the auth manager
	auth, ok := authManager.(interface {
		RequireAuth(http.Handler) http.Handler
	})
	if !ok {
		log.Fatal("CRITICAL SECURITY ERROR: Auth manager does not implement RequireAuth method - SERVER CANNOT START SAFELY")
		return // Never expose admin routes without proper authentication
	}

	// Register authenticated routes
	mux.Handle("/api/v1/admin/sessions/user", auth.RequireAuth(http.HandlerFunc(h.GetUserSessions)))
	mux.Handle("/api/v1/admin/sessions/active", auth.RequireAuth(http.HandlerFunc(h.GetAllActiveSessions)))
	mux.Handle("/api/v1/admin/sessions/invalidate", auth.RequireAuth(http.HandlerFunc(h.InvalidateSession)))
	mux.Handle("/api/v1/admin/sessions/invalidate-all", auth.RequireAuth(http.HandlerFunc(h.InvalidateAllUserSessions)))
	mux.Handle("/api/v1/admin/sessions/statistics", auth.RequireAuth(http.HandlerFunc(h.GetSessionStatistics)))
	mux.Handle("/api/v1/admin/sessions/devices", auth.RequireAuth(http.HandlerFunc(h.GetDeviceInfo)))
	mux.Handle("/api/v1/admin/sessions/trust-device", auth.RequireAuth(http.HandlerFunc(h.TrustDevice)))
	mux.Handle("/api/v1/admin/sessions/suspicious", auth.RequireAuth(http.HandlerFunc(h.GetSuspiciousActivities)))
	mux.Handle("/api/v1/admin/sessions/resolve-suspicious", auth.RequireAuth(http.HandlerFunc(h.ResolveSuspiciousActivity)))
	mux.Handle("/api/v1/admin/sessions/cleanup", auth.RequireAuth(http.HandlerFunc(h.CleanupExpiredSessions)))
	mux.Handle("/api/v1/admin/sessions/activity", auth.RequireAuth(http.HandlerFunc(h.GetSessionActivity)))

	log.Println("🔐 Session management routes registered:")
	log.Println("   • GET /api/v1/admin/sessions/user - Get user sessions")
	log.Println("   • GET /api/v1/admin/sessions/active - Get all active sessions")
	log.Println("   • POST /api/v1/admin/sessions/invalidate - Invalidate session")
	log.Println("   • POST /api/v1/admin/sessions/invalidate-all - Invalidate all user sessions")
	log.Println("   • GET /api/v1/admin/sessions/statistics - Get session statistics")
	log.Println("   • GET /api/v1/admin/sessions/devices - Get device information")
	log.Println("   • POST /api/v1/admin/sessions/trust-device - Trust/untrust device")
	log.Println("   • GET /api/v1/admin/sessions/suspicious - Get suspicious activities")
	log.Println("   • POST /api/v1/admin/sessions/resolve-suspicious - Resolve suspicious activity")
	log.Println("   • POST /api/v1/admin/sessions/cleanup - Cleanup expired sessions")
	log.Println("   • GET /api/v1/admin/sessions/activity - Get session activity")
}
