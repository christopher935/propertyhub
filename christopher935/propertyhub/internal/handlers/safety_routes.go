package handlers

import (
	"net/http"
	"gorm.io/gorm"
)

// RegisterSafetyRoutes registers all safety-related routes
func RegisterSafetyRoutes(mux *http.ServeMux, db *gorm.DB) {
	safetyHandlers := NewSafetyHandlers(db)

	// Safety configuration routes
	mux.HandleFunc("/api/v1/safety/config", safetyHandlers.GetSafetyConfig)
	mux.HandleFunc("/api/v1/safety/config/update", safetyHandlers.UpdateSafetySettings)

	// Safety mode routes
	mux.HandleFunc("/api/v1/safety/mode", safetyHandlers.UpdateSafetyMode)
	mux.HandleFunc("/api/v1/safety/mode/recommendation", safetyHandlers.GetModeRecommendation)
	mux.HandleFunc("/api/v1/safety/mode/history", safetyHandlers.GetTransitionHistory)
	mux.HandleFunc("/api/v1/safety/mode/plan", safetyHandlers.GetTransitionPlan)

	// Override routes
	mux.HandleFunc("/api/v1/safety/overrides", safetyHandlers.GetActiveOverrides)
	mux.HandleFunc("/api/v1/safety/overrides/request", safetyHandlers.RequestOverride)
	mux.HandleFunc("/api/v1/safety/overrides/expire", safetyHandlers.ExpireOverrides)
	mux.HandleFunc("/api/v1/safety/overrides/stats", safetyHandlers.GetOverrideStats)

	// Emergency routes
	mux.HandleFunc("/api/v1/safety/emergency", safetyHandlers.GetEmergencyState)
	mux.HandleFunc("/api/v1/safety/emergency/activate", safetyHandlers.ActivateEmergencyStop)
	mux.HandleFunc("/api/v1/safety/emergency/deactivate", safetyHandlers.DeactivateEmergencyStop)

	// Performance metrics
	mux.HandleFunc("/api/v1/safety/metrics", safetyHandlers.GetPerformanceMetrics)
}
