package security

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLProtectionMiddleware provides SQL injection protection
type SQLProtectionMiddleware struct {
	validator *InputValidator
	logger    *log.Logger
}

// NewSQLProtectionMiddleware creates a new SQL protection middleware
func NewSQLProtectionMiddleware(validator *InputValidator, logger *log.Logger) *SQLProtectionMiddleware {
	return &SQLProtectionMiddleware{
		validator: validator,
		logger:    logger,
	}
}

// Middleware returns the HTTP middleware function
func (spm *SQLProtectionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check URL parameters for SQL injection
		for key, values := range r.URL.Query() {
			for _, value := range values {
				if spm.validator.IsSQLInjectionAttempt(value) {
					spm.logger.Printf("🚨 SQL injection attempt detected in URL parameter '%s': %s from IP: %s", key, value, r.RemoteAddr)
					http.Error(w, "Invalid request parameters", http.StatusBadRequest)
					return
				}
			}
		}

		// Check form data for SQL injection
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if err := r.ParseForm(); err == nil {
				for key, values := range r.PostForm {
					for _, value := range values {
						if spm.validator.IsSQLInjectionAttempt(value) {
							spm.logger.Printf("🚨 SQL injection attempt detected in form data '%s': %s from IP: %s", key, value, r.RemoteAddr)
							http.Error(w, "Invalid form data", http.StatusBadRequest)
							return
						}
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SafeDB wraps GORM DB with additional safety checks
type SafeDB struct {
	*gorm.DB
	validator *InputValidator
	logger    *log.Logger
}

// NewSafeDB creates a new SafeDB wrapper
func NewSafeDB(db *gorm.DB, validator *InputValidator, logger *log.Logger) *SafeDB {
	return &SafeDB{
		DB:        db,
		validator: validator,
		logger:    logger,
	}
}

// SafeWhere performs a safe WHERE query with parameter validation
func (sdb *SafeDB) SafeWhere(query interface{}, args ...interface{}) *gorm.DB {
	// Validate query string for SQL injection
	if queryStr, ok := query.(string); ok {
		if sdb.validator.IsSQLInjectionAttempt(queryStr) {
			sdb.logger.Printf("🚨 Potential SQL injection in WHERE clause: %s", queryStr)
			// Return empty result set instead of executing dangerous query
			return sdb.DB.Where("1 = 0")
		}
	}

	// Validate arguments
	for i, arg := range args {
		if argStr, ok := arg.(string); ok {
			if sdb.validator.IsSQLInjectionAttempt(argStr) {
				sdb.logger.Printf("🚨 Potential SQL injection in WHERE argument %d: %s", i, argStr)
				// Return empty result set instead of executing dangerous query
				return sdb.DB.Where("1 = 0")
			}
		}
	}

	return sdb.DB.Where(query, args...)
}

// SafeFind performs a safe Find operation with validation
func (sdb *SafeDB) SafeFind(dest interface{}, conds ...interface{}) *gorm.DB {
	// Validate conditions
	for i, cond := range conds {
		if condStr, ok := cond.(string); ok {
			if sdb.validator.IsSQLInjectionAttempt(condStr) {
				sdb.logger.Printf("🚨 Potential SQL injection in Find condition %d: %s", i, condStr)
				// Return empty result set
				return sdb.DB.Where("1 = 0").Find(dest)
			}
		}
	}

	return sdb.DB.Find(dest, conds...)
}

// SafeFirst performs a safe First operation with validation
func (sdb *SafeDB) SafeFirst(dest interface{}, conds ...interface{}) *gorm.DB {
	// Validate conditions
	for i, cond := range conds {
		if condStr, ok := cond.(string); ok {
			if sdb.validator.IsSQLInjectionAttempt(condStr) {
				sdb.logger.Printf("🚨 Potential SQL injection in First condition %d: %s", i, condStr)
				// Return record not found
				return sdb.DB.Where("1 = 0").First(dest)
			}
		}
	}

	return sdb.DB.First(dest, conds...)
}

// SafeCreate performs a safe Create operation with validation
func (sdb *SafeDB) SafeCreate(value interface{}) *gorm.DB {
	// Additional validation could be added here for create operations
	return sdb.DB.Create(value)
}

// SafeUpdate performs a safe Update operation with validation
func (sdb *SafeDB) SafeUpdate(column string, value interface{}) *gorm.DB {
	// Validate column name
	if sdb.validator.IsSQLInjectionAttempt(column) {
		sdb.logger.Printf("🚨 Potential SQL injection in Update column: %s", column)
		// Return no-op update
		return sdb.DB.Where("1 = 0").Update(column, value)
	}

	// Validate value if it's a string
	if valueStr, ok := value.(string); ok {
		if sdb.validator.IsSQLInjectionAttempt(valueStr) {
			sdb.logger.Printf("🚨 Potential SQL injection in Update value: %s", valueStr)
			// Return no-op update
			return sdb.DB.Where("1 = 0").Update(column, value)
		}
	}

	return sdb.DB.Update(column, value)
}

// SafeDelete performs a safe Delete operation with validation
func (sdb *SafeDB) SafeDelete(value interface{}, conds ...interface{}) *gorm.DB {
	// Validate conditions
	for i, cond := range conds {
		if condStr, ok := cond.(string); ok {
			if sdb.validator.IsSQLInjectionAttempt(condStr) {
				sdb.logger.Printf("🚨 Potential SQL injection in Delete condition %d: %s", i, condStr)
				// Return no-op delete
				return sdb.DB.Where("1 = 0").Delete(value)
			}
		}
	}

	return sdb.DB.Delete(value, conds...)
}

// ConfigureSecureDB configures GORM with secure settings
func ConfigureSecureDB(db *gorm.DB) *gorm.DB {
	// Configure secure logger that doesn't log sensitive data
	secureLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error, // Only log errors, not queries
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return db.Session(&gorm.Session{
		Logger: secureLogger,
		// Disable prepared statements caching to prevent certain injection attacks
		PrepareStmt: false,
		// Set reasonable timeouts
		Context: ctx,
	})
}

// DatabaseQuerySanitizer provides additional query sanitization
type DatabaseQuerySanitizer struct {
	validator *InputValidator
}

// NewDatabaseQuerySanitizer creates a new query sanitizer
func NewDatabaseQuerySanitizer(validator *InputValidator) *DatabaseQuerySanitizer {
	return &DatabaseQuerySanitizer{validator: validator}
}

// SanitizeOrderBy sanitizes ORDER BY clauses
func (dqs *DatabaseQuerySanitizer) SanitizeOrderBy(orderBy string) string {
	if orderBy == "" {
		return ""
	}

	// Remove potentially dangerous characters
	orderBy = strings.ReplaceAll(orderBy, ";", "")
	orderBy = strings.ReplaceAll(orderBy, "--", "")
	orderBy = strings.ReplaceAll(orderBy, "/*", "")
	orderBy = strings.ReplaceAll(orderBy, "*/", "")

	// Check for SQL injection patterns
	if dqs.validator.IsSQLInjectionAttempt(orderBy) {
		return "id ASC" // Safe default
	}

	// Whitelist approach - only allow specific patterns
	allowedPattern := `^[a-zA-Z_][a-zA-Z0-9_]*(\s+(ASC|DESC|asc|desc))?$`
	if matched, _ := regexp.MatchString(allowedPattern, strings.TrimSpace(orderBy)); !matched {
		return "id ASC" // Safe default
	}

	return orderBy
}

// SanitizeLimit sanitizes LIMIT values
func (dqs *DatabaseQuerySanitizer) SanitizeLimit(limit int) int {
	if limit < 0 {
		return 10 // Safe default
	}
	if limit > 1000 {
		return 1000 // Reasonable maximum
	}
	return limit
}

// SanitizeOffset sanitizes OFFSET values
func (dqs *DatabaseQuerySanitizer) SanitizeOffset(offset int) int {
	if offset < 0 {
		return 0 // Safe default
	}
	if offset > 100000 {
		return 100000 // Reasonable maximum
	}
	return offset
}

// ValidateTableName validates table names for dynamic queries
func (dqs *DatabaseQuerySanitizer) ValidateTableName(tableName string) error {
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Check for SQL injection patterns
	if dqs.validator.IsSQLInjectionAttempt(tableName) {
		return fmt.Errorf("invalid table name")
	}

	// Whitelist approach - only allow alphanumeric and underscores
	allowedPattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	if matched, _ := regexp.MatchString(allowedPattern, tableName); !matched {
		return fmt.Errorf("invalid table name format")
	}

	// Length check
	if len(tableName) > 64 {
		return fmt.Errorf("table name too long")
	}

	return nil
}

// ValidateColumnName validates column names for dynamic queries
func (dqs *DatabaseQuerySanitizer) ValidateColumnName(columnName string) error {
	if columnName == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	// Check for SQL injection patterns
	if dqs.validator.IsSQLInjectionAttempt(columnName) {
		return fmt.Errorf("invalid column name")
	}

	// Whitelist approach - only allow alphanumeric and underscores
	allowedPattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	if matched, _ := regexp.MatchString(allowedPattern, columnName); !matched {
		return fmt.Errorf("invalid column name format")
	}

	// Length check
	if len(columnName) > 64 {
		return fmt.Errorf("column name too long")
	}

	return nil
}
