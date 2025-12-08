package config

import (
        "database/sql"
        "log"
        "os"
        "strconv"
        "time"
        _ "github.com/lib/pq"
)

type Config struct {
        // Database configuration (bootstrap only)
        DatabaseURL      string
        DatabaseMaxConns int
        DatabaseTimeout  time.Duration

        // Server configuration (bootstrap only)
        Port        string
        Environment string
        LogLevel    string

        // Everything else from database
        JWTSecret          string
        EncryptionKey      string
        SessionTimeout     time.Duration
        MFARequired        bool
        RateLimitPerMinute int

        // External services (all from database)
        FUBAPIKey     string
        FUBAPIURL     string
        ScraperAPIKey string // Generic scraper for property data (not HAR-specific)

        // TREC compliance (from database)
        TRECComplianceEnabled bool
        AuditLogRetentionDays int

        // Redis configuration (from database)
        RedisURL      string
        RedisPassword string
        RedisDB       int

        // Email configuration (from database)
        SMTPHost     string
        SMTPPort     int
        SMTPUsername string
        SMTPPassword string

        // Email configuration (from database)
        EmailFromAddress  string
        EmailFromName     string

        // Features (from database)
        TwilioAccountSID string
        TwilioAuthToken  string
        TwilioPhoneNumber string

        // Business (from database)
        BusinessName    string
        BusinessPhone   string
        BusinessEmail   string
        BusinessAddress string
        TRECLicense     string
}

var AppConfig *Config

func LoadConfig() *Config {
        log.Printf("🔧 DEBUG: LoadConfig called")
        
        // Initialize database connection for config loading
        dbURL := os.Getenv("DATABASE_URL")
        if dbURL == "" {
                log.Fatal("❌ DATABASE_URL environment variable required for bootstrap")
        }
        log.Printf("🔧 DEBUG: DATABASE_URL loaded from env")

        // Load all settings from database
        dbSettings := loadAllDatabaseSettings(dbURL)
        log.Printf("🔧 DEBUG: loadAllDatabaseSettings returned %d settings", len(dbSettings))
        
        // Debug: Print JWT_SECRET specifically
        if jwtSecret, exists := dbSettings["JWT_SECRET"]; exists {
                if len(jwtSecret) > 10 {
                        log.Printf("🔧 DEBUG: JWT_SECRET found in settings: %s...", jwtSecret[:10])
                } else {
                        log.Printf("🔧 DEBUG: JWT_SECRET found in settings: %s", jwtSecret)
                }
        } else {
                log.Printf("❌ DEBUG: JWT_SECRET NOT found in settings")
        }

        config := &Config{
                // Bootstrap from environment (minimum required)
                DatabaseURL: dbURL,
                Port:        getEnv("PORT", "8080"),
                Environment: getEnv("ENVIRONMENT", "production"),
                LogLevel:    getEnv("LOG_LEVEL", "info"),

                // Database connection settings
                DatabaseMaxConns: getDbSettingInt(dbSettings, "DATABASE_MAX_CONNS", 25),
                DatabaseTimeout:  time.Duration(getDbSettingInt(dbSettings, "DATABASE_TIMEOUT_SECONDS", 30)) * time.Second,

                // Security (ALL from database)
                JWTSecret:          dbSettings["JWT_SECRET"],
                EncryptionKey:      dbSettings["ENCRYPTION_KEY"],
                SessionTimeout:     time.Duration(getDbSettingInt(dbSettings, "SESSION_TIMEOUT_MINUTES", 60)) * time.Minute,
                MFARequired:        getDbSettingBool(dbSettings, "MFA_REQUIRED", false),
                RateLimitPerMinute: getDbSettingInt(dbSettings, "RATE_LIMIT_REQUESTS_PER_MINUTE", 100),

                // External services (ALL from database)
                FUBAPIKey:     dbSettings["FUB_API_KEY"],
                FUBAPIURL:     getDbSetting(dbSettings, "FUB_API_URL", "https://api.followupboss.com"),
                ScraperAPIKey: dbSettings["SCRAPER_API_KEY"], // Generic scraper

                // TREC compliance
                TRECComplianceEnabled: getDbSettingBool(dbSettings, "TREC_COMPLIANCE_ENABLED", true),
                AuditLogRetentionDays: getDbSettingInt(dbSettings, "AUDIT_LOG_RETENTION_DAYS", 365),

                // Redis
                RedisURL:      getDbSetting(dbSettings, "REDIS_URL", "localhost:6379"),
                RedisPassword: dbSettings["REDIS_PASSWORD"],
                RedisDB:       getDbSettingInt(dbSettings, "REDIS_DB", 0),

                // Email
                SMTPHost:     getDbSetting(dbSettings, "SMTP_HOST", "localhost"),
                SMTPPort:     getDbSettingInt(dbSettings, "SMTP_PORT", 587),
                SMTPUsername: dbSettings["SMTP_USERNAME"],
                SMTPPassword: dbSettings["SMTP_PASSWORD"],

                // Email settings
                EmailFromAddress: getDbSetting(dbSettings, "EMAIL_FROM_ADDRESS", "info@llotschedule.online"),
                EmailFromName:    getDbSetting(dbSettings, "EMAIL_FROM_NAME", "PropertyHub"),

                // Twilio
                TwilioAccountSID:  dbSettings["TWILIO_ACCOUNT_SID"],
                TwilioAuthToken:   dbSettings["TWILIO_AUTH_TOKEN"],
                TwilioPhoneNumber: dbSettings["TWILIO_PHONE_NUMBER"],

                // Business info
                BusinessName:    getDbSetting(dbSettings, "BUSINESS_NAME", "PropertyHub"),
                BusinessPhone:   getDbSetting(dbSettings, "BUSINESS_PHONE", "(713) 555-0123"),
                BusinessEmail:   getDbSetting(dbSettings, "BUSINESS_EMAIL", "info@propertyhub.com"),
                BusinessAddress: getDbSetting(dbSettings, "BUSINESS_ADDRESS", "Houston, TX"),
                TRECLicense:     getDbSetting(dbSettings, "TREC_LICENSE", "#625244"),
        }

        if len(config.JWTSecret) > 10 {
                log.Printf("🔧 DEBUG: Config struct created with JWT: %s...", config.JWTSecret[:10])
        } else {
                log.Printf("🔧 DEBUG: Config struct created with JWT: %s", config.JWTSecret)
        }
        AppConfig = config
        return config
}

func loadAllDatabaseSettings(dbURL string) map[string]string {
        log.Printf("🔧 DEBUG: Starting loadAllDatabaseSettings")
        settings := make(map[string]string)

        db, err := sql.Open("postgres", dbURL)
        if err != nil {
                log.Printf("❌ WARNING: Could not connect to database for settings: %v", err)
                return settings
        }
        defer db.Close()
        log.Printf("🔧 DEBUG: Database connection opened successfully")

        // Test connection first
        if err := db.Ping(); err != nil {
                log.Printf("❌ WARNING: Database ping failed: %v", err)
                return settings
        }
        log.Printf("🔧 DEBUG: Database ping successful")

        rows, err := db.Query("SELECT key, value FROM system_settings")
        if err != nil {
                log.Printf("❌ WARNING: Could not load settings from database: %v", err)
                return settings
        }
        defer rows.Close()
        log.Printf("🔧 DEBUG: Settings query executed successfully")

        for rows.Next() {
                var key, value string
                if err := rows.Scan(&key, &value); err != nil {
                        log.Printf("❌ WARNING: Could not scan setting row: %v", err)
                        continue
                }
                settings[key] = value
                
                // Safe debug logging - avoid slice bounds crash
                if len(value) > 10 {
                        log.Printf("🔧 DEBUG: Loaded setting: %s = %s...", key, value[:10])
                } else {
                        log.Printf("🔧 DEBUG: Loaded setting: %s = %s", key, value)
                }
        }

        if err := rows.Err(); err != nil {
                log.Printf("❌ WARNING: Error iterating settings rows: %v", err)
        }

        log.Printf("✅ Loaded %d settings from database", len(settings))
        return settings
}

func getDbSetting(settings map[string]string, key, defaultValue string) string {
        if value, exists := settings[key]; exists && value != "" {
                return value
        }
        return defaultValue
}

func getDbSettingInt(settings map[string]string, key string, defaultValue int) int {
        if value, exists := settings[key]; exists && value != "" {
                if intVal, err := strconv.Atoi(value); err == nil {
                        return intVal
                }
        }
        return defaultValue
}

func getDbSettingBool(settings map[string]string, key string, defaultValue bool) bool {
        if value, exists := settings[key]; exists {
                return value == "true" || value == "1"
        }
        return defaultValue
}

func (c *Config) IsDevelopment() bool {
        return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
        return c.Environment == "production"
}

func (c *Config) HasJWTSecret() bool {
        return c.JWTSecret != ""
}

// Bootstrap helpers (only for DATABASE_URL, PORT, ENVIRONMENT)
func getEnv(key, defaultValue string) string {
        if value := os.Getenv(key); value != "" {
                return value
        }
        return defaultValue
}

func GetConfig() *Config {
        if AppConfig == nil {
                return LoadConfig()
        }
        return AppConfig
}
