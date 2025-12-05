package services

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// PhotoProtectionService provides enhanced photo validation and protection
type PhotoProtectionService struct {
	db *gorm.DB
}

// NewPhotoProtectionService creates a new photo protection service
func NewPhotoProtectionService(db *gorm.DB) *PhotoProtectionService {
	return &PhotoProtectionService{
		db: db,
	}
}

// PhotoValidationResult represents the result of photo validation
type PhotoValidationResult struct {
	IsValid            bool     `json:"is_valid"`
	Errors             []string `json:"errors"`
	Warnings           []string `json:"warnings"`
	PhotoQuality       string   `json:"photo_quality"` // excellent, good, fair, poor
	RecommendedActions []string `json:"recommended_actions"`

	// Technical details
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	AspectRatio      float64 `json:"aspect_ratio"`
	FileSize         int64   `json:"file_size"`
	Format           string  `json:"format"`
	IsDuplicate      bool    `json:"is_duplicate"`
	DuplicatePhotoID *uint   `json:"duplicate_photo_id,omitempty"`
}

// PhotoRequirements defines the requirements for property photos
type PhotoRequirements struct {
	MinWidth           int      `json:"min_width"`
	MinHeight          int      `json:"min_height"`
	MaxFileSize        int64    `json:"max_file_size"`
	MinAspectRatio     float64  `json:"min_aspect_ratio"`
	MaxAspectRatio     float64  `json:"max_aspect_ratio"`
	AllowedFormats     []string `json:"allowed_formats"`
	RequiredPhotoTypes []string `json:"required_photo_types"`
	MinPhotoCount      int      `json:"min_photo_count"`
	MaxPhotoCount      int      `json:"max_photo_count"`
}

// GetDefaultPhotoRequirements returns the default photo requirements
func (pps *PhotoProtectionService) GetDefaultPhotoRequirements() PhotoRequirements {
	return PhotoRequirements{
		MinWidth:           800,
		MinHeight:          600,
		MaxFileSize:        5 * 1024 * 1024, // 5MB
		MinAspectRatio:     0.75,            // 3:4 portrait
		MaxAspectRatio:     2.0,             // 2:1 landscape
		AllowedFormats:     []string{"image/jpeg", "image/jpg", "image/png"},
		RequiredPhotoTypes: []string{"exterior", "interior", "kitchen", "bathroom"},
		MinPhotoCount:      3,
		MaxPhotoCount:      25,
	}
}

// ValidatePhoto performs comprehensive photo validation
func (pps *PhotoProtectionService) ValidatePhoto(file multipart.File, fileHeader *multipart.FileHeader, mlsID string) (*PhotoValidationResult, error) {
	result := &PhotoValidationResult{
		IsValid:            true,
		Errors:             []string{},
		Warnings:           []string{},
		RecommendedActions: []string{},
		FileSize:           fileHeader.Size,
		Format:             fileHeader.Header.Get("Content-Type"),
	}

	requirements := pps.GetDefaultPhotoRequirements()

	// Reset file position
	file.Seek(0, 0)

	// 1. Validate file format
	if !pps.isValidFormat(result.Format, requirements.AllowedFormats) {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid file format: %s. Allowed formats: %v", result.Format, requirements.AllowedFormats))
	}

	// 2. Validate file size
	if fileHeader.Size > requirements.MaxFileSize {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("File size too large: %d bytes. Maximum allowed: %d bytes", fileHeader.Size, requirements.MaxFileSize))
		result.RecommendedActions = append(result.RecommendedActions, "Compress the image to reduce file size")
	}

	// 3. Validate image dimensions and quality
	if err := pps.validateImageDimensions(file, result, requirements); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not validate image dimensions: %v", err))
	}

	// 4. Check for duplicates
	file.Seek(0, 0)
	if err := pps.checkForDuplicates(file, mlsID, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check for duplicates: %v", err))
	}

	// 5. Assess overall photo quality
	pps.assessPhotoQuality(result)

	// 6. Check property photo count limits
	if err := pps.validatePhotoCount(mlsID, requirements, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not validate photo count: %v", err))
	}

	return result, nil
}

// validateImageDimensions validates image dimensions and calculates aspect ratio
func (pps *PhotoProtectionService) validateImageDimensions(file multipart.File, result *PhotoValidationResult, req PhotoRequirements) error {
	file.Seek(0, 0)

	// Decode image to get dimensions
	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	result.Width = bounds.Dx()
	result.Height = bounds.Dy()
	result.AspectRatio = float64(result.Width) / float64(result.Height)
	result.Format = "image/" + format

	// Validate minimum dimensions
	if result.Width < req.MinWidth {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Image width too small: %dpx. Minimum required: %dpx", result.Width, req.MinWidth))
		result.RecommendedActions = append(result.RecommendedActions, "Use a higher resolution image")
	}

	if result.Height < req.MinHeight {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Image height too small: %dpx. Minimum required: %dpx", result.Height, req.MinHeight))
		result.RecommendedActions = append(result.RecommendedActions, "Use a higher resolution image")
	}

	// Validate aspect ratio
	if result.AspectRatio < req.MinAspectRatio || result.AspectRatio > req.MaxAspectRatio {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Unusual aspect ratio: %.2f. Recommended range: %.2f - %.2f", result.AspectRatio, req.MinAspectRatio, req.MaxAspectRatio))
		result.RecommendedActions = append(result.RecommendedActions, "Consider cropping the image to a more standard aspect ratio")
	}

	return nil
}

// checkForDuplicates checks if the photo is a duplicate of an existing photo
func (pps *PhotoProtectionService) checkForDuplicates(file multipart.File, mlsID string, result *PhotoValidationResult) error {
	file.Seek(0, 0)

	// Calculate MD5 hash of the file
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	fileHash := fmt.Sprintf("%x", hash.Sum(nil))

	// Check if any existing photo has the same hash
	var existingPhoto models.PropertyPhoto
	err := pps.db.Where("mls_id = ? AND is_active = ? AND file_hash = ?", mlsID, true, fileHash).First(&existingPhoto).Error

	if err == nil {
		// Duplicate found
		result.IsDuplicate = true
		result.DuplicatePhotoID = &existingPhoto.ID
		result.IsValid = false
		result.Errors = append(result.Errors, "This photo has already been uploaded for this property")
		result.RecommendedActions = append(result.RecommendedActions, "Choose a different photo or update the existing one")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	return nil
}

// assessPhotoQuality assesses the overall quality of the photo
func (pps *PhotoProtectionService) assessPhotoQuality(result *PhotoValidationResult) {
	score := 100

	// Deduct points for various issues
	if result.Width < 1200 || result.Height < 900 {
		score -= 20
		result.Warnings = append(result.Warnings, "Image resolution is below recommended standards (1200x900)")
	}

	if result.FileSize < 100*1024 { // Less than 100KB
		score -= 15
		result.Warnings = append(result.Warnings, "File size is very small, which may indicate low quality")
	}

	if result.AspectRatio < 0.8 || result.AspectRatio > 1.8 {
		score -= 10
	}

	// Determine quality rating
	if score >= 90 {
		result.PhotoQuality = "excellent"
	} else if score >= 75 {
		result.PhotoQuality = "good"
	} else if score >= 60 {
		result.PhotoQuality = "fair"
		result.RecommendedActions = append(result.RecommendedActions, "Consider using a higher quality image for better presentation")
	} else {
		result.PhotoQuality = "poor"
		result.RecommendedActions = append(result.RecommendedActions, "This image quality is below standards. Please use a higher quality photo")
	}
}

// validatePhotoCount checks if the property meets photo count requirements
func (pps *PhotoProtectionService) validatePhotoCount(mlsID string, req PhotoRequirements, result *PhotoValidationResult) error {
	var photoCount int64
	if err := pps.db.Model(&models.PropertyPhoto{}).Where("mls_id = ? AND is_active = ?", mlsID, true).Count(&photoCount).Error; err != nil {
		return err
	}

	currentCount := int(photoCount)

	if currentCount >= req.MaxPhotoCount {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Maximum photo limit reached: %d/%d", currentCount, req.MaxPhotoCount))
		result.RecommendedActions = append(result.RecommendedActions, "Remove some existing photos before adding new ones")
	} else if currentCount < req.MinPhotoCount {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Property needs more photos: %d/%d minimum", currentCount, req.MinPhotoCount))
		result.RecommendedActions = append(result.RecommendedActions, fmt.Sprintf("Add %d more photos to meet minimum requirements", req.MinPhotoCount-currentCount))
	}

	return nil
}

// isValidFormat checks if the file format is allowed
func (pps *PhotoProtectionService) isValidFormat(format string, allowedFormats []string) bool {
	for _, allowed := range allowedFormats {
		if strings.EqualFold(format, allowed) {
			return true
		}
	}
	return false
}

// OptimizePhoto optimizes a photo for web display
func (pps *PhotoProtectionService) OptimizePhoto(inputPath, outputPath string, maxWidth, maxHeight int, quality int) error {
	// Open the input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	// Decode the image
	img, format, err := image.Decode(inputFile)
	if err != nil {
		return err
	}

	// Calculate new dimensions while maintaining aspect ratio
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > maxWidth || height > maxHeight {
		ratio := float64(width) / float64(height)

		if width > height {
			width = maxWidth
			height = int(float64(maxWidth) / ratio)
		} else {
			height = maxHeight
			width = int(float64(maxHeight) * ratio)
		}
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Encode based on format
	switch format {
	case "jpeg", "jpg":
		return jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(outputFile, img)
	default:
		return jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	}
}

// GeneratePhotoVariants generates different sizes of a photo for responsive display
func (pps *PhotoProtectionService) GeneratePhotoVariants(originalPath string, mlsID string) error {
	baseDir := filepath.Dir(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
	ext := filepath.Ext(originalPath)

	variants := map[string]struct {
		width   int
		height  int
		quality int
	}{
		"thumbnail": {300, 200, 80},
		"medium":    {800, 600, 85},
		"large":     {1200, 900, 90},
	}

	for variantName, config := range variants {
		variantPath := filepath.Join(baseDir, fmt.Sprintf("%s_%s%s", baseName, variantName, ext))

		if err := pps.OptimizePhoto(originalPath, variantPath, config.width, config.height, config.quality); err != nil {
			return fmt.Errorf("failed to generate %s variant: %v", variantName, err)
		}
	}

	return nil
}

// CleanupInactivePhotos removes files for photos that have been inactive for a specified duration
func (pps *PhotoProtectionService) CleanupInactivePhotos(inactiveDuration time.Duration) error {
	var inactivePhotos []models.PropertyPhoto

	cutoffTime := time.Now().Add(-inactiveDuration)

	err := pps.db.Where("is_active = ? AND updated_at < ?", false, cutoffTime).Find(&inactivePhotos).Error
	if err != nil {
		return err
	}

	for _, photo := range inactivePhotos {
		// Remove file from disk
		if err := os.Remove(photo.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but continue with cleanup
			fmt.Printf("Warning: Could not remove file %s: %v\n", photo.FilePath, err)
		}

		// Remove variants
		baseDir := filepath.Dir(photo.FilePath)
		baseName := strings.TrimSuffix(filepath.Base(photo.FilePath), filepath.Ext(photo.FilePath))
		ext := filepath.Ext(photo.FilePath)

		variants := []string{"thumbnail", "medium", "large"}
		for _, variant := range variants {
			variantPath := filepath.Join(baseDir, fmt.Sprintf("%s_%s%s", baseName, variant, ext))
			os.Remove(variantPath) // Ignore errors for variants
		}

		// Permanently delete from database
		if err := pps.db.Unscoped().Delete(&photo).Error; err != nil {
			fmt.Printf("Warning: Could not delete photo record %d: %v\n", photo.ID, err)
		}
	}

	return nil
}

// GetPhotoStatistics returns statistics about property photos
func (pps *PhotoProtectionService) GetPhotoStatistics(mlsID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total photos
	var totalCount int64
	pps.db.Model(&models.PropertyPhoto{}).Where("mls_id = ?", mlsID).Count(&totalCount)
	stats["total_photos"] = totalCount

	// Active photos
	var activeCount int64
	pps.db.Model(&models.PropertyPhoto{}).Where("mls_id = ? AND is_active = ?", mlsID, true).Count(&activeCount)
	stats["active_photos"] = activeCount

	// Primary photo
	var hasPrimary bool
	err := pps.db.Where("mls_id = ? AND is_primary = ? AND is_active = ?", mlsID, true, true).First(&models.PropertyPhoto{}).Error
	hasPrimary = err == nil
	stats["has_primary_photo"] = hasPrimary

	// Quality distribution - placeholder for future implementation
	stats["quality_distribution"] = map[string]interface{}{
		"high":   0,
		"medium": 0,
		"low":    0,
	}

	// File size statistics
	var avgFileSize float64
	pps.db.Model(&models.PropertyPhoto{}).Where("mls_id = ? AND is_active = ?", mlsID, true).Select("AVG(file_size)").Scan(&avgFileSize)
	stats["average_file_size"] = avgFileSize

	return stats, nil
}
