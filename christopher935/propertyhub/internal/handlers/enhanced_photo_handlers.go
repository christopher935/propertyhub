package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"
)

// EnhancedPhotoHandlers provides enhanced photo management with smart protection
type EnhancedPhotoHandlers struct {
	db                       *gorm.DB
	photoProtectionService   *services.PhotoProtectionService
	propertyReadinessService *services.PropertyReadinessService
	encryptionManager        *security.EncryptionManager
}

// NewEnhancedPhotoHandlers creates a new enhanced photo handlers instance
func NewEnhancedPhotoHandlers(db *gorm.DB, encryptionManager *security.EncryptionManager) *EnhancedPhotoHandlers {
	return &EnhancedPhotoHandlers{
		db:                       db,
		photoProtectionService:   services.NewPhotoProtectionService(db),
		encryptionManager:        encryptionManager,
		propertyReadinessService: services.NewPropertyReadinessService(db),
	}
}

// EnhancedPhotoUploadResponse represents the enhanced response after photo upload
type EnhancedPhotoUploadResponse struct {
	Success          bool                              `json:"success"`
	Message          string                            `json:"message"`
	Photo            *models.PropertyPhoto             `json:"photo,omitempty"`
	ValidationResult *services.PhotoValidationResult   `json:"validation_result"`
	ReadinessStatus  *services.PropertyReadinessStatus `json:"readiness_status,omitempty"`
	BookingEligible  bool                              `json:"booking_eligible"`
	AutoFixesApplied []string                          `json:"auto_fixes_applied,omitempty"`
}

// UploadPropertyPhotoEnhancedHandler handles enhanced property photo uploads with validation
func (eph *EnhancedPhotoHandlers) UploadPropertyPhotoEnhancedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		eph.sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	mlsID := r.FormValue("mls_id")
	if mlsID == "" {
		eph.sendErrorResponse(w, "MLS ID is required", http.StatusBadRequest)
		return
	}

	isPrimary := r.FormValue("is_primary") == "true"
	caption := r.FormValue("caption")
	altText := r.FormValue("alt_text")
	displayOrder, _ := strconv.Atoi(r.FormValue("display_order"))
	autoOptimize := r.FormValue("auto_optimize") != "false" // Default to true

	// Get uploaded file
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		eph.sendErrorResponse(w, "No photo file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate photo using protection service
	validationResult, err := eph.photoProtectionService.ValidatePhoto(file, fileHeader, mlsID)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to validate photo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If validation failed, return validation errors
	if !validationResult.IsValid {
		response := EnhancedPhotoUploadResponse{
			Success:          false,
			Message:          "Photo validation failed",
			ValidationResult: validationResult,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Find or create property
	property, err := eph.findOrCreateProperty(mlsID)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to find property: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate unique filename and file hash
	fileName := eph.generateFileName(fileHeader.Filename, mlsID)
	filePath := filepath.Join("uploads", "properties", fileName)

	// Calculate file hash for duplicate detection
	file.Seek(0, 0)
	hash := md5.New()
	io.Copy(hash, file)
	fileHash := fmt.Sprintf("%x", hash.Sum(nil))
	file.Seek(0, 0)

	// Ensure upload directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		eph.sendErrorResponse(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	fileSize, err := io.Copy(dst, file)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create photo record with enhanced fields
	photo := models.PropertyPhoto{
		PropertyID:     property.ID,
		MLSId:          mlsID,
		FileName:       fileName,
		OriginalName:   fileHeader.Filename,
		FilePath:       filePath,
		FileSize:       fileSize,
		MimeType:       fileHeader.Header.Get("Content-Type"),
		FileHash:       fileHash,
		IsPrimary:      isPrimary,
		DisplayOrder:   displayOrder,
		Caption:        caption,
		AltText:        altText,
		IsActive:       true,
		UploadedBy:     "agent", // TODO: Get from authentication
		PropertyStatus: property.Status,
	}

	// Save photo to database
	if err := eph.db.Create(&photo).Error; err != nil {
		// Clean up file if database save fails
		os.Remove(filePath)
		eph.sendErrorResponse(w, "Failed to save photo record", http.StatusInternalServerError)
		return
	}

	// Auto-optimize photo if requested
	if autoOptimize {
		if err := eph.photoProtectionService.GeneratePhotoVariants(filePath, mlsID); err != nil {
			// Log warning but don't fail the upload
			fmt.Printf("Warning: Failed to generate photo variants: %v\n", err)
		}
	}

	// Set as primary if requested
	if isPrimary {
		if err := photo.SetAsPrimary(eph.db); err != nil {
			// Photo is saved but primary setting failed
			response := EnhancedPhotoUploadResponse{
				Success:          false,
				Message:          "Photo uploaded but failed to set as primary",
				Photo:            &photo,
				ValidationResult: validationResult,
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Check property readiness after photo upload
	readinessStatus, err := eph.propertyReadinessService.CheckPropertyReadiness(mlsID)
	if err != nil {
		// Photo uploaded successfully but readiness check failed
		response := EnhancedPhotoUploadResponse{
			Success:          true,
			Message:          "Photo uploaded but failed to check property readiness",
			Photo:            &photo,
			ValidationResult: validationResult,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Try to auto-fix any readiness issues
	autoFixedStatus, err := eph.propertyReadinessService.AutoFixReadinessIssues(mlsID)
	autoFixesApplied := []string{}
	if err == nil && autoFixedStatus.ReadinessScore > readinessStatus.ReadinessScore {
		readinessStatus = autoFixedStatus
		autoFixesApplied = append(autoFixesApplied, "Applied automatic readiness fixes")
	}

	response := EnhancedPhotoUploadResponse{
		Success:          true,
		Message:          "Photo uploaded successfully",
		Photo:            &photo,
		ValidationResult: validationResult,
		ReadinessStatus:  readinessStatus,
		BookingEligible:  readinessStatus.IsReady,
		AutoFixesApplied: autoFixesApplied,
	}

	json.NewEncoder(w).Encode(response)
}

// ValidatePhotoHandler validates a photo without uploading it
func (eph *EnhancedPhotoHandlers) ValidatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	mlsID := r.FormValue("mls_id")
	if mlsID == "" {
		eph.sendErrorResponse(w, "MLS ID is required", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		eph.sendErrorResponse(w, "No photo file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate photo
	validationResult, err := eph.photoProtectionService.ValidatePhoto(file, fileHeader, mlsID)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to validate photo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":           true,
		"message":           "Photo validation completed",
		"validation_result": validationResult,
	}

	json.NewEncoder(w).Encode(response)
}

// GetPropertyReadinessHandler returns the readiness status of a property
func (eph *EnhancedPhotoHandlers) GetPropertyReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	mlsID := r.URL.Query().Get("mls_id")
	if mlsID == "" {
		eph.sendErrorResponse(w, "MLS ID is required", http.StatusBadRequest)
		return
	}

	readinessStatus, err := eph.propertyReadinessService.CheckPropertyReadiness(mlsID)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to check property readiness: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":          true,
		"message":          "Property readiness checked successfully",
		"readiness_status": readinessStatus,
	}

	json.NewEncoder(w).Encode(response)
}

// AutoFixPropertyReadinessHandler attempts to automatically fix property readiness issues
func (eph *EnhancedPhotoHandlers) AutoFixPropertyReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		MLSId string `json:"mls_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		eph.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.MLSId == "" {
		eph.sendErrorResponse(w, "MLS ID is required", http.StatusBadRequest)
		return
	}

	readinessStatus, err := eph.propertyReadinessService.AutoFixReadinessIssues(request.MLSId)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to auto-fix readiness issues: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":          true,
		"message":          "Auto-fix completed",
		"readiness_status": readinessStatus,
	}

	json.NewEncoder(w).Encode(response)
}

// GetReadinessReportHandler generates a readiness report for multiple properties
func (eph *EnhancedPhotoHandlers) GetReadinessReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		MLSIds []string `json:"mls_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		eph.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.MLSIds) == 0 {
		eph.sendErrorResponse(w, "At least one MLS ID is required", http.StatusBadRequest)
		return
	}

	report, err := eph.propertyReadinessService.GetReadinessReport(request.MLSIds)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to generate readiness report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Readiness report generated successfully",
		"report":  report,
		"count":   len(report),
	}

	json.NewEncoder(w).Encode(response)
}

// GetPhotoStatisticsHandler returns detailed photo statistics for a property
func (eph *EnhancedPhotoHandlers) GetPhotoStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	mlsID := r.URL.Query().Get("mls_id")
	if mlsID == "" {
		eph.sendErrorResponse(w, "MLS ID is required", http.StatusBadRequest)
		return
	}

	stats, err := eph.photoProtectionService.GetPhotoStatistics(mlsID)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to get photo statistics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    "Photo statistics retrieved successfully",
		"statistics": stats,
	}

	json.NewEncoder(w).Encode(response)
}

// CleanupInactivePhotosHandler cleans up inactive photos
func (eph *EnhancedPhotoHandlers) CleanupInactivePhotosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	origin := os.Getenv("CORS_ALLOWED_ORIGIN"); if origin == "" { origin = "http://localhost:8080" }; w.Header().Set("Access-Control-Allow-Origin", origin)

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Default to 30 days for cleanup
	inactiveDuration := 30 * 24 * time.Hour

	// Allow custom duration from query parameter
	if daysStr := r.URL.Query().Get("inactive_days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			inactiveDuration = time.Duration(days) * 24 * time.Hour
		}
	}

	err := eph.photoProtectionService.CleanupInactivePhotos(inactiveDuration)
	if err != nil {
		eph.sendErrorResponse(w, "Failed to cleanup inactive photos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Cleaned up photos inactive for more than %v", inactiveDuration),
	}

	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (eph *EnhancedPhotoHandlers) findOrCreateProperty(mlsID string) (*models.Property, error) {
	var property models.Property

	err := eph.db.Where("mls_id = ?", mlsID).First(&property).Error
	if err == gorm.ErrRecordNotFound {
		// Create placeholder property (will be updated by scraper)
		placeholderAddress := "Property " + mlsID
		encryptedAddress, err := eph.encryptionManager.Encrypt(placeholderAddress)
		if err != nil {
			return nil, err
		}

		property = models.Property{
			MLSId:   mlsID,
			Address: encryptedAddress,
			City:    "Houston",
			State:   "TX",
			ZipCode: "77000",
			Price:   0,
			Status:  "pending_scrape",
		}

		if err := eph.db.Create(&property).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &property, nil
}

func (eph *EnhancedPhotoHandlers) generateFileName(originalName, mlsID string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d%s", mlsID, timestamp, ext)
}

func (eph *EnhancedPhotoHandlers) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// RegisterEnhancedPhotoRoutes registers all enhanced photo-related routes
func RegisterEnhancedPhotoRoutes(mux *http.ServeMux, db *gorm.DB, encryptionManager *security.EncryptionManager) {
	handlers := NewEnhancedPhotoHandlers(db, encryptionManager)

	// Enhanced photo upload with validation
	mux.HandleFunc("/api/v1/photos/upload-enhanced", handlers.UploadPropertyPhotoEnhancedHandler)

	// Photo validation without upload
	mux.HandleFunc("/api/v1/photos/validate", handlers.ValidatePhotoHandler)

	// Property readiness checking
	mux.HandleFunc("/api/v1/photos/readiness", handlers.GetPropertyReadinessHandler)
	mux.HandleFunc("/api/v1/photos/readiness/autofix", handlers.AutoFixPropertyReadinessHandler)
	mux.HandleFunc("/api/v1/photos/readiness/report", handlers.GetReadinessReportHandler)

	// Photo statistics and management
	mux.HandleFunc("/api/v1/photos/statistics", handlers.GetPhotoStatisticsHandler)
	mux.HandleFunc("/api/v1/photos/cleanup", handlers.CleanupInactivePhotosHandler)
}
