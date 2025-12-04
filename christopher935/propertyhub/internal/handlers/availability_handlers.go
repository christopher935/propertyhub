package handlers

import (
	"os"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"
)

type AvailabilityHandler struct {
	db                  *gorm.DB
	availabilityService *services.AvailabilityService
	response            *utils.ResponseHelper
}

func NewAvailabilityHandler(db *gorm.DB) *AvailabilityHandler {
	return &AvailabilityHandler{
		db:                  db,
		availabilityService: services.NewAvailabilityService(db),
		response:            utils.NewResponseHelper(),
	}
}

// CheckAvailability checks if a property is available for booking
func (h *AvailabilityHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	// Set standard headers for GET requests
	h.response.SetStandardHeaders(w, "GET", "OPTIONS")

	mlsID := r.URL.Query().Get("mls_id")
	dateStr := r.URL.Query().Get("date")

	if mlsID == "" || dateStr == "" {
		h.response.WriteBadRequestError(w, "mls_id and date parameters are required")
		return
	}

	requestedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.response.WriteBadRequestError(w, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	check, err := h.availabilityService.CheckAvailability(mlsID, requestedDate)
	if err != nil {
		h.response.WriteInternalServerError(w, fmt.Sprintf("Failed to check availability: %v", err))
		return
	}

	h.response.WriteSuccessResponse(w, "Availability check completed", check)
}

// CreateBlackoutDate creates a new blackout date
func (h *AvailabilityHandler) CreateBlackoutDate(w http.ResponseWriter, r *http.Request) {
	// Handle standard HTTP flow (OPTIONS, method validation, headers)
	if !h.response.HandleStandardHTTPFlow(w, r, "POST") {
		return
	}

	var request struct {
		MLSId     string `json:"mls_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Reason    string `json:"reason"`
		IsGlobal  bool   `json:"is_global"`
		CreatedBy string `json:"created_by"`
	}

	if !h.response.ParseJSONBody(w, r, &request) {
		return
	}

	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		h.response.WriteBadRequestError(w, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		h.response.WriteBadRequestError(w, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	if endDate.Before(startDate) {
		h.response.WriteBadRequestError(w, "End date must be after start date")
		return
	}

	if request.Reason == "" {
		h.response.WriteBadRequestError(w, "Reason is required")
		return
	}

	if request.CreatedBy == "" {
		request.CreatedBy = "admin"
	}

	blackout, err := h.availabilityService.CreateBlackoutDate(
		request.MLSId, startDate, endDate, request.Reason, request.CreatedBy, request.IsGlobal)
	if err != nil {
		h.response.WriteInternalServerError(w, fmt.Sprintf("Failed to create blackout date: %v", err))
		return
	}

	h.response.WriteJSONResponse(w, http.StatusCreated, utils.StandardResponse{
		Success: true,
		Message: "Blackout date created successfully",
		Data:    blackout,
	})
}

// GetBlackoutDates retrieves blackout dates
func (h *AvailabilityHandler) GetBlackoutDates(w http.ResponseWriter, r *http.Request) {
	// Set standard headers
	h.response.SetStandardHeaders(w, "GET", "OPTIONS")

	mlsID := r.URL.Query().Get("mls_id")

	blackouts, err := h.availabilityService.GetBlackoutDates(mlsID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get blackout dates: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    blackouts,
		"count":   len(blackouts),
	})
}

// RemoveBlackoutDate removes a blackout date
func (h *AvailabilityHandler) RemoveBlackoutDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	// Extract blackout ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	blackoutID, err := strconv.ParseUint(pathParts[5], 10, 32)
	if err != nil {
		http.Error(w, "Invalid blackout ID", http.StatusBadRequest)
		return
	}

	err = h.availabilityService.RemoveBlackoutDate(uint(blackoutID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Blackout date not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to remove blackout date: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Blackout date removed successfully",
	})
}

// GetUpcomingBlackouts returns upcoming blackout dates
func (h *AvailabilityHandler) GetUpcomingBlackouts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	blackouts, err := h.availabilityService.GetUpcomingBlackouts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get upcoming blackouts: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    blackouts,
		"count":   len(blackouts),
	})
}

// GetAvailabilityStats returns availability statistics
func (h *AvailabilityHandler) GetAvailabilityStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	stats, err := h.availabilityService.GetAvailabilityStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get availability stats: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// ValidateBooking validates a booking request against availability rules
func (h *AvailabilityHandler) ValidateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	var request struct {
		MLSId string `json:"mls_id"`
		Date  string `json:"date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requestedDate, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	err = h.availabilityService.ValidateBookingDate(request.MLSId, requestedDate)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"valid":    false,
			"message":  err.Error(),
			"can_book": false,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"valid":    true,
		"message":  "Booking date is available",
		"can_book": true,
	})
}

// CleanupExpiredBlackouts removes expired blackout dates
func (h *AvailabilityHandler) CleanupExpiredBlackouts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	err := h.availabilityService.CleanupExpiredBlackouts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to cleanup expired blackouts: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Expired blackout dates cleaned up successfully",
	})
}

// RegisterAvailabilityRoutes registers all availability-related routes
func RegisterAvailabilityRoutes(mux *http.ServeMux, db *gorm.DB) {
	handler := NewAvailabilityHandler(db)

	mux.HandleFunc("/api/v1/availability/check", handler.CheckAvailability)
	mux.HandleFunc("/api/v1/availability/blackouts", handler.GetBlackoutDates)
	mux.HandleFunc("/api/v1/availability/blackouts/create", handler.CreateBlackoutDate)
	mux.HandleFunc("/api/v1/availability/blackouts/", handler.RemoveBlackoutDate)
	mux.HandleFunc("/api/v1/availability/blackouts/upcoming", handler.GetUpcomingBlackouts)
	mux.HandleFunc("/api/v1/availability/stats", handler.GetAvailabilityStats)
	mux.HandleFunc("/api/v1/availability/validate", handler.ValidateBooking)
	mux.HandleFunc("/api/v1/availability/cleanup", handler.CleanupExpiredBlackouts)
}
