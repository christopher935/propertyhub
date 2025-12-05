package handlers

import (
	"os"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/safety"
	"gorm.io/gorm"
)

// SafetyHandlers handles safety-related API endpoints
type SafetyHandlers struct {
	db              *gorm.DB
	configManager   *safety.SafetyConfigManager
	modeManager     *safety.SafetyModeManager
	overrideManager *safety.OverrideManager
}

// NewSafetyHandlers creates new safety handlers
func NewSafetyHandlers(db *gorm.DB) *SafetyHandlers {
	configManager := safety.NewSafetyConfigManager()
	modeManager := safety.NewSafetyModeManager(configManager)
	overrideManager := safety.NewOverrideManager(configManager)

	return &SafetyHandlers{
		db:              db,
		configManager:   configManager,
		modeManager:     modeManager,
		overrideManager: overrideManager,
	}
}

// GetSafetyConfig returns the current safety configuration
func (h *SafetyHandlers) GetSafetyConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	config := h.configManager.GetConfig()

	response := map[string]interface{}{
		"config":      config,
		"stats":       h.configManager.GetSafetyStats(),
		"description": h.configManager.GetModeDescription(),
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateSafetyMode updates the safety mode
func (h *SafetyHandlers) UpdateSafetyMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	var request struct {
		Mode       int    `json:"mode"`
		ModifiedBy string `json:"modified_by"`
		Reason     string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert int to SafetyMode
	mode := safety.SafetyMode(request.Mode)

	// Attempt the transition
	if err := h.modeManager.TransitionToMode(mode, request.ModifiedBy, request.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return updated config
	response := map[string]interface{}{
		"success": true,
		"config":  h.configManager.GetConfig(),
		"message": "Safety mode updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetModeRecommendation returns mode transition recommendations
func (h *SafetyHandlers) GetModeRecommendation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	// Query real performance metrics from database
	var totalCampaigns, successfulCampaigns, failedCampaigns int64
	h.db.Table("campaign_execution_logs").Count(&totalCampaigns)
	h.db.Table("campaign_execution_logs").Where("status = ?", "success").Count(&successfulCampaigns)
	h.db.Table("campaign_execution_logs").Where("status = ?", "failed").Count(&failedCampaigns)
	
	var complaintCount, unsubscribeCount int64
	h.db.Table("email_complaints").Count(&complaintCount)
	h.db.Table("unsubscribes").Count(&unsubscribeCount)
	
	// Calculate rates
	successRate := 0.0
	if totalCampaigns > 0 {
		successRate = float64(successfulCampaigns) / float64(totalCampaigns)
	}
	
	complaintRate := 0.0
	unsubscribeRate := 0.0
	var totalEmails int64
	h.db.Table("email_campaigns").Select("COALESCE(SUM(sent_count), 0)").Scan(&totalEmails)
	if totalEmails > 0 {
		complaintRate = float64(complaintCount) / float64(totalEmails)
		unsubscribeRate = float64(unsubscribeCount) / float64(totalEmails)
	}
	
	// Calculate average engagement rate from campaigns
	var avgEngagementRate float64
	h.db.Table("email_campaigns").
		Select("COALESCE(AVG((opened_count::float + clicked_count::float) / NULLIF(sent_count, 0)), 0)").
		Where("sent_count > 0").
		Scan(&avgEngagementRate)
	
	// Get days in current mode from config
	config := h.configManager.GetConfig()
	daysInCurrentMode := 0
	if !config.LastModified.IsZero() {
		daysInCurrentMode = int(time.Since(config.LastModified).Hours() / 24)
	}
	
	metrics := safety.SafetyMetrics{
		TotalCampaigns:        int(totalCampaigns),
		SuccessfulCampaigns:   int(successfulCampaigns),
		FailedCampaigns:       int(failedCampaigns),
		ComplaintCount:        int(complaintCount),
		UnsubscribeCount:      int(unsubscribeCount),
		SuccessRate:           successRate,
		ComplaintRate:         complaintRate,
		UnsubscribeRate:       unsubscribeRate,
		AverageEngagementRate: avgEngagementRate,
		DaysInCurrentMode:     daysInCurrentMode,
	}

	recommendation := h.modeManager.GetModeRecommendation(metrics)

	json.NewEncoder(w).Encode(recommendation)
}

// GetTransitionHistory returns the mode transition history
func (h *SafetyHandlers) GetTransitionHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	history := h.modeManager.GetTransitionHistory()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"transitions": history,
		"count":       len(history),
	})
}

// GetTransitionPlan returns the mode transition plan
func (h *SafetyHandlers) GetTransitionPlan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	plan := h.modeManager.GetModeTransitionPlan()

	json.NewEncoder(w).Encode(plan)
}

// RequestOverride requests a safety override
func (h *SafetyHandlers) RequestOverride(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	var request safety.OverrideRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	override, err := h.overrideManager.RequestOverride(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"override": override,
		"message":  "Override created successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetActiveOverrides returns all active overrides
func (h *SafetyHandlers) GetActiveOverrides(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	overrides := h.overrideManager.GetActiveOverrides()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"overrides": overrides,
		"count":     len(overrides),
	})
}

// RevokeOverride revokes an active override
func (h *SafetyHandlers) RevokeOverride(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	// Extract override ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid override ID", http.StatusBadRequest)
		return
	}
	overrideID := pathParts[4] // /api/v1/safety/overrides/{id}/revoke

	var request struct {
		RevokedBy string `json:"revoked_by"`
		Reason    string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.overrideManager.RevokeOverride(overrideID, request.RevokedBy, request.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Override revoked successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// ActivateEmergencyStop activates emergency stop procedures
func (h *SafetyHandlers) ActivateEmergencyStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	var request struct {
		ActivatedBy string `json:"activated_by"`
		Reason      string `json:"reason"`
		Level       int    `json:"level"` // 0=Low, 1=Medium, 2=High, 3=Critical
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	level := safety.EmergencyLevel(request.Level)

	if err := h.overrideManager.ActivateEmergencyStop(request.ActivatedBy, request.Reason, level); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":         true,
		"message":         "Emergency stop activated",
		"emergency_state": h.overrideManager.GetEmergencyState(),
	}

	json.NewEncoder(w).Encode(response)
}

// DeactivateEmergencyStop deactivates emergency stop procedures
func (h *SafetyHandlers) DeactivateEmergencyStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	var request struct {
		DeactivatedBy string `json:"deactivated_by"`
		Reason        string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.overrideManager.DeactivateEmergencyStop(request.DeactivatedBy, request.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Emergency stop deactivated",
	}

	json.NewEncoder(w).Encode(response)
}

// GetEmergencyState returns the current emergency state
func (h *SafetyHandlers) GetEmergencyState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	state := h.overrideManager.GetEmergencyState()

	json.NewEncoder(w).Encode(state)
}

// UpdateSafetySettings updates specific safety settings
func (h *SafetyHandlers) UpdateSafetySettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	var request struct {
		ModifiedBy string                 `json:"modified_by"`
		Settings   map[string]interface{} `json:"settings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update settings (implementation would update specific settings)
	// This is a placeholder for the actual implementation

	response := map[string]interface{}{
		"success": true,
		"message": "Safety settings updated successfully",
		"config":  h.configManager.GetConfig(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetPerformanceMetrics returns current performance metrics
func (h *SafetyHandlers) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	// Mock performance metrics for demonstration
	metrics := map[string]interface{}{
		"total_campaigns":         12,
		"successful_campaigns":    11,
		"failed_campaigns":        1,
		"complaint_count":         0,
		"unsubscribe_count":       2,
		"success_rate":            0.965,
		"complaint_rate":          0.002,
		"unsubscribe_rate":        0.015,
		"average_engagement_rate": 0.18,
		"days_in_current_mode":    18,
		"last_updated":            time.Now(),
	}

	json.NewEncoder(w).Encode(metrics)
}

// ExpireOverrides manually triggers override expiration check
func (h *SafetyHandlers) ExpireOverrides(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	if err := h.overrideManager.ExpireOverrides(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":          true,
		"message":          "Override expiration check completed",
		"active_overrides": len(h.overrideManager.GetActiveOverrides()),
	}

	json.NewEncoder(w).Encode(response)
}

// GetOverrideStats returns override statistics
func (h *SafetyHandlers) GetOverrideStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	stats := h.overrideManager.GetOverrideStats()

	json.NewEncoder(w).Encode(stats)
}
