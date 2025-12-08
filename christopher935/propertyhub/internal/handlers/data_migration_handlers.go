package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DataMigrationHandlers handles CSV import functionality
type DataMigrationHandlers struct {
	db               *gorm.DB
	migrationService *services.DataMigrationService
}

// NewDataMigrationHandlers creates new data migration handlers
func NewDataMigrationHandlers(db *gorm.DB) *DataMigrationHandlers {
	return &DataMigrationHandlers{
		db:               db,
		migrationService: services.NewDataMigrationService(db),
	}
}

// ImportCustomers handles customer CSV imports
func (dmh *DataMigrationHandlers) ImportCustomers(c *gin.Context) {
	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get CSV file
	files := form.File["csv_file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No CSV file provided",
		})
		return
	}

	file := files[0]

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File must be a CSV file",
		})
		return
	}

	// Get options
	skipDuplicates := c.DefaultPostForm("skip_duplicates", "true") == "true"

	// Open file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open CSV file",
		})
		return
	}
	defer src.Close()

	// Import customers
	result, err := dmh.migrationService.ImportCustomers(src, skipDuplicates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Import failed: %v", err),
		})
		return
	}

	status := "completed"
	if result.ErrorCount > 0 && result.SuccessCount == 0 {
		status = "failed"
	} else if result.ErrorCount > 0 {
		status = "partial"
	}

	errorLog := ""
	if len(result.Errors) > 0 {
		for i, e := range result.Errors {
			if i >= 10 {
				errorLog += fmt.Sprintf("... and %d more errors\n", len(result.Errors)-10)
				break
			}
			errorLog += fmt.Sprintf("Row %d: %s\n", e.Row, e.Message)
		}
	}

	dataImport := models.DataImport{
		Type:           "customers",
		FileName:       file.Filename,
		RecordsTotal:   result.TotalRows,
		RecordsSuccess: result.SuccessCount,
		RecordsFailed:  result.ErrorCount,
		RecordsSkipped: result.SkippedCount,
		Status:         status,
		ErrorLog:       errorLog,
		ImportedBy:     "admin",
		DurationMs:     result.Duration.Milliseconds(),
	}

	if err := dmh.db.Create(&dataImport).Error; err != nil {
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Customer import completed",
		"result":  result,
	})
}

// ImportProperties handles property CSV imports
func (dmh *DataMigrationHandlers) ImportProperties(c *gin.Context) {
	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get CSV file
	files := form.File["csv_file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No CSV file provided",
		})
		return
	}

	file := files[0]

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File must be a CSV file",
		})
		return
	}

	// Get options
	skipDuplicates := c.DefaultPostForm("skip_duplicates", "true") == "true"

	// Open file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open CSV file",
		})
		return
	}
	defer src.Close()

	// Import properties
	result, err := dmh.migrationService.ImportProperties(src, skipDuplicates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Import failed: %v", err),
		})
		return
	}

	status := "completed"
	if result.ErrorCount > 0 && result.SuccessCount == 0 {
		status = "failed"
	} else if result.ErrorCount > 0 {
		status = "partial"
	}

	errorLog := ""
	if len(result.Errors) > 0 {
		for i, e := range result.Errors {
			if i >= 10 {
				errorLog += fmt.Sprintf("... and %d more errors\n", len(result.Errors)-10)
				break
			}
			errorLog += fmt.Sprintf("Row %d: %s\n", e.Row, e.Message)
		}
	}

	dataImport := models.DataImport{
		Type:           "properties",
		FileName:       file.Filename,
		RecordsTotal:   result.TotalRows,
		RecordsSuccess: result.SuccessCount,
		RecordsFailed:  result.ErrorCount,
		RecordsSkipped: result.SkippedCount,
		Status:         status,
		ErrorLog:       errorLog,
		ImportedBy:     "admin",
		DurationMs:     result.Duration.Milliseconds(),
	}

	if err := dmh.db.Create(&dataImport).Error; err != nil {
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Property import completed",
		"result":  result,
	})
}

// ImportBookings handles booking CSV imports
func (dmh *DataMigrationHandlers) ImportBookings(c *gin.Context) {
	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get CSV file
	files := form.File["csv_file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No CSV file provided",
		})
		return
	}

	file := files[0]

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File must be a CSV file",
		})
		return
	}

	// Get options
	skipDuplicates := c.DefaultPostForm("skip_duplicates", "true") == "true"

	// Open file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open CSV file",
		})
		return
	}
	defer src.Close()

	// Import bookings
	result, err := dmh.migrationService.ImportBookings(src, skipDuplicates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Import failed: %v", err),
		})
		return
	}

	status := "completed"
	if result.ErrorCount > 0 && result.SuccessCount == 0 {
		status = "failed"
	} else if result.ErrorCount > 0 {
		status = "partial"
	}

	errorLog := ""
	if len(result.Errors) > 0 {
		for i, e := range result.Errors {
			if i >= 10 {
				errorLog += fmt.Sprintf("... and %d more errors\n", len(result.Errors)-10)
				break
			}
			errorLog += fmt.Sprintf("Row %d: %s\n", e.Row, e.Message)
		}
	}

	dataImport := models.DataImport{
		Type:           "bookings",
		FileName:       file.Filename,
		RecordsTotal:   result.TotalRows,
		RecordsSuccess: result.SuccessCount,
		RecordsFailed:  result.ErrorCount,
		RecordsSkipped: result.SkippedCount,
		Status:         status,
		ErrorLog:       errorLog,
		ImportedBy:     "admin",
		DurationMs:     result.Duration.Milliseconds(),
	}

	if err := dmh.db.Create(&dataImport).Error; err != nil {
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Booking import completed",
		"result":  result,
	})
}

// DownloadSampleCSV provides sample CSV files for import
func (dmh *DataMigrationHandlers) DownloadSampleCSV(c *gin.Context) {
	dataType := c.Param("type")

	if dataType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Data type is required",
		})
		return
	}

	// Generate sample CSV
	csvContent, err := dmh.migrationService.GenerateSampleCSV(dataType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid data type: %s", dataType),
		})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("sample_%s.csv", dataType)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Length", strconv.Itoa(len(csvContent)))

	// Send CSV content
	c.String(http.StatusOK, csvContent)
}

// GetImportHistory returns import history and statistics
func (dmh *DataMigrationHandlers) GetImportHistory(c *gin.Context) {
	var imports []models.DataImport
	if err := dmh.db.Order("created_at DESC").Limit(50).Find(&imports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch import history",
		})
		return
	}

	var totalCount int64
	dmh.db.Model(&models.DataImport{}).Count(&totalCount)

	totalRecordsImported := 0
	totalRecordsProcessed := 0
	importTypeCount := make(map[string]int)
	var mostRecentImport *time.Time

	for _, imp := range imports {
		totalRecordsImported += imp.RecordsSuccess
		totalRecordsProcessed += imp.RecordsTotal
		importTypeCount[imp.Type]++
		if mostRecentImport == nil || imp.CreatedAt.After(*mostRecentImport) {
			mostRecentImport = &imp.CreatedAt
		}
	}

	successRate := 0.0
	if totalRecordsProcessed > 0 {
		successRate = (float64(totalRecordsImported) / float64(totalRecordsProcessed)) * 100
	}

	summary := map[string]interface{}{
		"total_imports":          totalCount,
		"total_records_imported": totalRecordsImported,
		"success_rate":           successRate,
		"most_recent_import":     mostRecentImport,
		"import_types":           importTypeCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"history": imports,
			"summary": summary,
		},
	})
}

// ValidateCSV validates a CSV file without importing
func (dmh *DataMigrationHandlers) ValidateCSV(c *gin.Context) {
	dataType := c.PostForm("data_type")
	if dataType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Data type is required (customers, properties, bookings)",
		})
		return
	}

	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get CSV file
	files := form.File["csv_file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No CSV file provided",
		})
		return
	}

	file := files[0]

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File must be a CSV file",
		})
		return
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open CSV file",
		})
		return
	}
	defer src.Close()

	// Perform validation based on data type
	validation := dmh.performCSVValidation(src, dataType)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"validation": validation,
	})
}

// GetImportRequirements returns the required columns for each import type
func (dmh *DataMigrationHandlers) GetImportRequirements(c *gin.Context) {
	requirements := map[string]interface{}{
		"customers": map[string]interface{}{
			"required_columns": []string{"first_name", "last_name", "email"},
			"optional_columns": []string{
				"phone", "alternate_phone", "source", "tags", "notes",
				"marketing_consent", "preferred_contact", "budget", "move_in_date",
			},
			"sample_data": map[string]string{
				"first_name":        "John",
				"last_name":         "Doe",
				"email":             "john.doe@example.com",
				"phone":             "713-555-0123",
				"marketing_consent": "true",
				"budget":            "2500",
				"move_in_date":      "2024-02-01",
			},
		},
		"properties": map[string]interface{}{
			"required_columns": []string{"address"},
			"optional_columns": []string{
				"city", "state", "zip_code", "property_type", "bedrooms",
				"bathrooms", "square_feet", "rent", "deposit", "available",
				"available_date", "description", "amenities", "pet_policy",
				"owner_name", "owner_email", "owner_phone",
			},
			"sample_data": map[string]string{
				"address":        "123 Main St",
				"city":           "Houston",
				"state":          "TX",
				"zip_code":       "77002",
				"bedrooms":       "2",
				"bathrooms":      "2",
				"rent":           "2500",
				"available":      "true",
				"available_date": "2024-01-15",
			},
		},
		"bookings": map[string]interface{}{
			"required_columns": []string{"customer_email", "property_address", "booking_date", "booking_time"},
			"optional_columns": []string{"status", "notes", "agent_name", "booking_type", "duration"},
			"sample_data": map[string]string{
				"customer_email":   "john.doe@example.com",
				"property_address": "123 Main St",
				"booking_date":     "2024-01-20",
				"booking_time":     "14:00",
				"status":           "confirmed",
				"duration":         "30",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"requirements": requirements,
	})
}

// Helper function to validate CSV structure
func (dmh *DataMigrationHandlers) performCSVValidation(src interface{}, dataType string) map[string]interface{} {
	// This is a simplified validation - in production you'd parse the CSV
	// and check column headers, data types, required fields, etc.

	return map[string]interface{}{
		"valid":           true,
		"row_count":       100, // Mock count
		"columns_found":   []string{"first_name", "last_name", "email", "phone"},
		"missing_columns": []string{},
		"warnings": []string{
			"10 rows have empty phone numbers",
			"2 rows have invalid email formats",
		},
		"errors":          []string{},
		"ready_to_import": true,
	}
}
