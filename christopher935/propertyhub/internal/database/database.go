package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitDB initializes the database connection
func InitDB() error {
	// Use secure SSL mode for production, disable only for local development
	sslMode := getEnv("DB_SSL_MODE", "require")
	if getEnv("ENVIRONMENT", "production") == "development" {
		sslMode = "disable"
		log.Println("WARNING: Database SSL disabled for development environment")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Chicago",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "landlords_of_texas"),
		getEnv("DB_PORT", "5432"),
		sslMode,
	)

	var err error
	// Use appropriate logging level based on environment
	logLevel := logger.Info
	if getEnv("ENVIRONMENT", "production") == "production" {
		logLevel = logger.Error // Reduce logging in production
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// Production-optimized connection pool settings
	sqlDB.SetMaxOpenConns(25)                  // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)                   // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum connection lifetime
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Maximum idle time

	// Verify database connection with health check
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("📊 Database connection established successfully with optimized pool settings")
	log.Printf("🔧 Pool config: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v", 25, 5, time.Hour)
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("Database not initialized. Call InitDB() first.")
	}
	return db
}

// CloseDB closes the database connection gracefully
func CloseDB() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		log.Println("🔌 Closing database connections gracefully...")
		return sqlDB.Close()
	}
	return nil
}

// HealthCheck performs a database health check
func HealthCheck() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// Ping with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	return nil
}

// GetConnectionStats returns database connection statistics
func GetConnectionStats() map[string]interface{} {
	if db == nil {
		return map[string]interface{}{"error": "database not initialized"}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return map[string]interface{}{"error": "failed to get sql.DB"}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
		"max_open_connections": stats.MaxOpenConnections,
	}
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
