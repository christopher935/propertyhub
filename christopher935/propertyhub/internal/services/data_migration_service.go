package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
)

// DataMigrationService handles CSV imports using only actual model fields
type DataMigrationService struct {
	db *gorm.DB
}

// NewDataMigrationService creates a new data migration service
func NewDataMigrationService(db *gorm.DB) *DataMigrationService {
	return &DataMigrationService{
		db: db,
	}
}

// MigrationResult represents the result of a data migration
type MigrationResult struct {
	TotalRows     int                    `json:"total_rows"`
	SuccessCount  int                    `json:"success_count"`
	ErrorCount    int                    `json:"error_count"`
	SkippedCount  int                    `json:"skipped_count"`
	Errors        []MigrationError       `json:"errors"`
	Summary       map[string]interface{} `json:"summary"`
	ProcessedAt   time.Time              `json:"processed_at"`
	Duration      time.Duration          `json:"duration"`
}

// MigrationError represents an error during migration
type MigrationError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
	Type    string `json:"type"` // validation, parsing, database
}

// ImportCustomers imports customers from CSV using only actual Lead model fields
func (dms *DataMigrationService) ImportCustomers(csvReader io.Reader, skipDuplicates bool) (*MigrationResult, error) {
	startTime := time.Now()
	result := &MigrationResult{
		Errors:      []MigrationError{},
		ProcessedAt: startTime,
	}

	// Parse CSV
	reader := csv.NewReader(csvReader)
	reader.FieldsPerRecord = -1

	// Read header
	header, err := reader.Read()
	if err != nil {
		return result, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Create column map
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Validate required columns
	requiredColumns := []string{"first_name", "last_name", "email"}
	for _, required := range requiredColumns {
		if _, exists := columnMap[required]; !exists {
			return result, fmt.Errorf("required column '%s' not found in CSV", required)
		}
	}

	// Process rows
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				Row:     rowNum,
				Message: fmt.Sprintf("CSV parsing error: %v", err),
				Type:    "parsing",
			})
			rowNum++
			continue
		}

		result.TotalRows++
		
		// Extract basic fields that exist in Lead model
		firstName := dms.getColumnValue(record, columnMap, "first_name")
		lastName := dms.getColumnValue(record, columnMap, "last_name")
		email := strings.ToLower(strings.TrimSpace(dms.getColumnValue(record, columnMap, "email")))
		phone := dms.getColumnValue(record, columnMap, "phone")
		source := dms.getColumnValue(record, columnMap, "source")
		
		// Validate required fields
		if firstName == "" || lastName == "" || email == "" || !strings.Contains(email, "@") {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				Row:     rowNum,
				Message: "Missing required field or invalid email",
				Type:    "validation",
			})
			rowNum++
			continue
		}

		// Check for duplicates
		if skipDuplicates {
			var existingLead models.Lead
			if err := dms.db.Where("email = ?", email).First(&existingLead).Error; err == nil {
				result.SkippedCount++
				rowNum++
				continue
			}
		}

		// Create lead using only actual model fields
		lead := &models.Lead{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Phone:     phone,
			Source:    source,
			Status:    "new",
		}

		// Parse tags if available
		if tagsStr := dms.getColumnValue(record, columnMap, "tags"); tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			lead.Tags = make(models.StringArray, len(tags))
			for i, tag := range tags {
				lead.Tags[i] = strings.TrimSpace(tag)
			}
		}

		// Save to database
		if err := dms.db.Create(lead).Error; err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				Row:     rowNum,
				Message: fmt.Sprintf("Database error: %v", err),
				Type:    "database",
			})
		} else {
			result.SuccessCount++
		}

		rowNum++
	}

	result.Duration = time.Since(startTime)
	result.Summary = map[string]interface{}{
		"imported_customers": result.SuccessCount,
		"duplicate_emails":   result.SkippedCount,
		"total_processed":    result.TotalRows,
	}

	log.Printf("Customer import completed: %d success, %d errors, %d skipped", 
		result.SuccessCount, result.ErrorCount, result.SkippedCount)

	return result, nil
}

// ImportProperties imports properties using only actual model fields
func (dms *DataMigrationService) ImportProperties(csvReader io.Reader, skipDuplicates bool) (*MigrationResult, error) {
	startTime := time.Now()
	result := &MigrationResult{
		Errors:      []MigrationError{},
		ProcessedAt: startTime,
	}

	reader := csv.NewReader(csvReader)
	reader.FieldsPerRecord = -1

	// Read header
	header, err := reader.Read()
	if err != nil {
		return result, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Create column map
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Validate required columns
	if _, exists := columnMap["address"]; !exists {
		return result, fmt.Errorf("required column 'address' not found in CSV")
	}

	// Process rows
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				Row:     rowNum,
				Message: fmt.Sprintf("CSV parsing error: %v", err),
				Type:    "parsing",
			})
			rowNum++
			continue
		}

		result.TotalRows++

		// Extract address
		address := strings.TrimSpace(dms.getColumnValue(record, columnMap, "address"))
		if address == "" {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				Row:     rowNum,
				Field:   "address",
				Message: "Address is required",
				Type:    "validation",
			})
			rowNum++
			continue
		}

		// Check for duplicates
		if skipDuplicates {
			var existingProperty models.Property
			if err := dms.db.Where("address = ?", security.EncryptedString(address)).First(&existingProperty).Error; err == nil {
				result.SkippedCount++
				rowNum++
				continue
			}
		}

		// Create property using actual model fields only
		property := &models.Property{
			Address:      security.EncryptedString(address),
			City:         dms.getColumnValue(record, columnMap, "city"),
			State:        dms.getColumnValue(record, columnMap, "state"),
			ZipCode:      dms.getColumnValue(record, columnMap, "zip_code"),
			PropertyType: dms.getColumnValue(record, columnMap, "property_type"),
		}

		// Parse numeric fields safely
		if bedroomsStr := dms.getColumnValue(record, columnMap, "bedrooms"); bedroomsStr != "" {
			if bedrooms, err := strconv.Atoi(bedroomsStr); err == nil {
				property.Bedrooms = &bedrooms
			}
		}

		if bathroomsStr := dms.getColumnValue(record, columnMap, "bathrooms"); bathroomsStr != "" {
			if bathrooms, err := strconv.ParseFloat(bathroomsStr, 64); err == nil {
				bathroomsFloat32 := float32(bathrooms)
				property.Bathrooms = &bathroomsFloat32
			}
		}

		if priceStr := dms.getColumnValue(record, columnMap, "rent"); priceStr != "" {
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				property.Price = price
			}
		}

		// Save to database
		if err := dms.db.Create(property).Error; err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				Row:     rowNum,
				Message: fmt.Sprintf("Database error: %v", err),
				Type:    "database",
			})
		} else {
			result.SuccessCount++
		}

		rowNum++
	}

	result.Duration = time.Since(startTime)
	result.Summary = map[string]interface{}{
		"imported_properties": result.SuccessCount,
		"duplicate_addresses": result.SkippedCount,
		"total_processed":     result.TotalRows,
	}

	return result, nil
}

// Helper function to get column value safely
func (dms *DataMigrationService) getColumnValue(record []string, columnMap map[string]int, fieldName string) string {
	if idx, exists := columnMap[fieldName]; exists && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

// GenerateSampleCSV generates sample CSV files for import templates
func (dms *DataMigrationService) GenerateSampleCSV(dataType string) (string, error) {
	switch dataType {
	case "customers":
		return `first_name,last_name,email,phone,source,tags
John,Doe,john.doe@example.com,713-555-0123,website,"interested,high-priority"
Jane,Smith,jane.smith@example.com,281-555-0456,referral,"returning,vip"`, nil
	case "properties":
		return `address,city,state,zip_code,property_type,bedrooms,bathrooms,rent
"123 Main St",Houston,TX,77002,apartment,2,2,2500
"456 Oak Ave",Houston,TX,77001,house,3,2,3200`, nil
	default:
		return "", fmt.Errorf("unknown data type: %s", dataType)
	}
}