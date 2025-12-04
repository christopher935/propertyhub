package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/auth"
	"chrisgross-ctrl-project/internal/config"
	"chrisgross-ctrl-project/internal/handlers"
	"chrisgross-ctrl-project/internal/middleware"
	"chrisgross-ctrl-project/internal/repositories"
	"chrisgross-ctrl-project/internal/scraper"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	_ "github.com/lib/pq"
)

func main() {
        log.Println("🚀 Starting PropertyHub Enterprise System v2.0...")

        // Load enterprise configuration
        cfg := config.LoadConfig()
        log.Println("⚙️ Enterprise configuration loaded")

        // Initialize enterprise database
        gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
        if err != nil {
                log.Fatalf("❌ Database connection failed: %v", err)
        }
        log.Println("📊 Enterprise database connected")

        // Initialize SQL database for auth manager
        sqlDB, _ := gormDB.DB()

        // Initialize enterprise authentication manager
        authManager := auth.NewSimpleAuthManager(sqlDB)
        log.Println("🔐 Enterprise authentication initialized")

        // Initialize enterprise security
        encryptionManager, err := security.NewEncryptionManager(gormDB)
        if err != nil {
                log.Printf("Warning: Encryption manager initialization failed: %v", err)
        }

          // Initialize repositories  
        repos := repositories.NewRepositories(gormDB)
        log.Println("📚 Enterprise repositories initialized")

        // Initialize email processor
        emailProcessor := services.NewEmailProcessor(gormDB)
        log.Println("📧 Enterprise email processor initialized")

        // Initialize scraper service (required for valuation)
        var scraperService *scraper.ScraperService
        var propertyValuationService *services.PropertyValuationService
        if cfg.ScraperAPIKey != "" {
                scraperService = scraper.NewScraperService(cfg)
                propertyValuationService = services.NewPropertyValuationService(cfg, scraperService)
                log.Println("🕷️ Enterprise scraper service initialized")
                log.Println("💰 Enterprise property valuation initialized")
        }
// ============================================================================
// PHASE 1: COMPLETE HANDLER INITIALIZATIONS
// Insert after line 73 (after scraper/valuation initialization)
// ============================================================================

// Initialize Redis if configured
var redisClient *redis.Client
if cfg.RedisURL != "" {
        redisClient = redis.NewClient(&redis.Options{
                Addr:     cfg.RedisURL,
                Password: cfg.RedisPassword,
                DB:       cfg.RedisDB,
        })

        // Test Redis connection
        if err := redisClient.Ping(context.Background()).Err(); err != nil {
                log.Printf("⚠️  Redis connection failed: %v (continuing without Redis)", err)
                redisClient = nil
        } else {
                log.Println("🔴 Redis connected")
        }
}

// Initialize all enterprise handlers
log.Println("🔧 Initializing enterprise handlers...")

// Analytics & Business Intelligence
analyticsAPIHandler := handlers.NewAnalyticsAPIHandlers(gormDB)
biService := services.NewBusinessIntelligenceService(gormDB)
	businessIntelligenceHandler := handlers.NewBusinessIntelligenceHandlers(gormDB)
	log.Println("📊 Analytics handlers initialized")

// Approvals & Workflow Management
approvalsHandler := handlers.NewApprovalsManagementHandlers(gormDB)
applicationWorkflowHandler := handlers.NewApplicationWorkflowHandlers(gormDB)
closingPipelineHandler := handlers.NewClosingPipelineHandlers(gormDB)
log.Println("✅ Workflow handlers initialized")

// Behavioral Intelligence & FUB Integration
behavioralHandler := handlers.NewBehavioralIntelligenceHandlers(gormDB)
contextFUBHandler := handlers.NewContextFUBIntegrationHandlers(gormDB, cfg.FUBAPIKey)
log.Println("🧠 Behavioral intelligence handlers initialized")

// Calendar & Scheduling
calendarHandler := handlers.NewCalendarHandlers(gormDB)
log.Println("📅 Calendar handlers initialized")

// Data Migration & Import
dataMigrationHandler := handlers.NewDataMigrationHandlers(gormDB)
log.Println("📥 Data migration handlers initialized")

// Email Management
emailSenderHandler := handlers.NewEmailSenderHandlers(gormDB)
unsubscribeHandler := handlers.NewUnsubscribeHandlers(gormDB)

// Email automation (only if Redis is available)
var emailAutomationHandler *handlers.EmailAutomationHandlers
if redisClient != nil {
        // Create SMTP config
        smtpConfig := services.SMTPConfig{
                Host:     cfg.SMTPHost,
                Port:     cfg.SMTPPort,
                Username: cfg.SMTPUsername,
                Password: cfg.SMTPPassword,
        }

        // Initialize email batch service
        emailBatchService := services.NewEmailBatchService(redisClient, smtpConfig)

        // Initialize email automation handler
        emailAutomationHandler = handlers.NewEmailAutomationHandlers(gormDB, emailProcessor, emailBatchService)
        log.Println("📧 Email automation handlers initialized")
} else {
        log.Println("📧 Email senders initialized (automation disabled - Redis not available)")
}

// HAR Market & Reports
harMarketHandler := handlers.NewHARMarketHandlers(gormDB, cfg.ScraperAPIKey)
log.Println("🏘️ HAR market handlers initialized")

// Lead Management & Reengagement
leadReengagementHandler := handlers.NewLeadReengagementHandler(gormDB, encryptionManager)
leadsListHandler := handlers.NewLeadsListHandler(gormDB, encryptionManager)
log.Println("👥 Lead management handlers initialized")

// Pre-listing Management
preListingHandler := handlers.NewPreListingHandlers(gormDB, cfg, scraperService)
log.Println("📝 Pre-listing handlers initialized")

// Property Valuation
var propertyValuationHandler *handlers.PropertyValuationHandlers
if propertyValuationService != nil {
        propertyValuationHandler = handlers.NewPropertyValuationHandlers(gormDB, propertyValuationService)
        log.Println("💰 Property valuation handlers initialized")
}

// Security & Monitoring
securityMonitoringHandler := handlers.NewSecurityMonitoringHandlers(gormDB)
advancedSecurityAPIHandler := handlers.NewAdvancedSecurityAPIHandlers(gormDB, encryptionManager)
log.Println("🔒 Security handlers initialized")

// Webhook Integrations
webhookHandler := handlers.NewWebhookHandlers(gormDB)
log.Println("🔗 Webhook handlers initialized")


	// PropertyHub AI Intelligence System
	// PropertyHub AI Intelligence System with Redis Caching
	log.Println("🏠 Initializing PropertyHub AI Intelligence System...")
	
	intelligenceCache := services.NewIntelligenceCacheService(redisClient)
	if intelligenceCache.IsAvailable() {
		log.Println("✅ Intelligence cache service initialized (Redis)")
	} else {
		log.Println("⚠️ Intelligence cache unavailable - running without cache")
	}
	
	scoringEngine := services.NewBehavioralScoringEngine(gormDB)
	insightGenerator := services.NewInsightGeneratorService(gormDB, scoringEngine, biService)
	propertyHubAI := services.NewSpiderwebAIOrchestrator(
		gormDB,
		scoringEngine,
		insightGenerator,
		nil,
		nil,
		intelligenceCache,
	)
	log.Println("✅ PropertyHub AI System initialized")
	
	dashboardStatsService := services.NewDashboardStatsService(gormDB, propertyHubAI, intelligenceCache)
	log.Println("✅ Dashboard stats service initialized")
	
	tieredStatsHandler := handlers.NewTieredStatsHandlers(gormDB, dashboardStatsService)
	log.Println("✅ Tiered stats handler initialized")
	
	go propertyHubAI.StartAutomatedIntelligence(5)
	log.Println("🤖 Automated intelligence cycle started (5 minute interval)")
	propertiesHandler := handlers.NewPropertiesHandler(gormDB, repos, encryptionManager)
	log.Println("🏠 Properties handler initialized with decryption")
	
	savedPropertiesHandler := handlers.NewSavedPropertiesHandler(gormDB)
	log.Println("💾 Saved properties handler initialized")
	
	recommendationsHandler := handlers.NewRecommendationsHandler(gormDB, scoringEngine)
	log.Println("🤖 AI recommendations handler initialized")
	
	emailService := services.NewEmailService()
	propertyAlertsHandler := handlers.NewPropertyAlertsHandler(gormDB, emailService)
	log.Println("🔔 Property alerts handler initialized")
	
	liveActivityHandler := handlers.NewLiveActivityHandler(gormDB)
	log.Println("📡 Live activity handler initialized")

	log.Println("✅ All enterprise handlers initialized successfully")

	// Create handlers struct for route registration
	allHandlers := &AllHandlers{
		AnalyticsAPI:          analyticsAPIHandler,
		BusinessIntelligence:  businessIntelligenceHandler,
		TieredStats:           tieredStatsHandler,
		Approvals:             approvalsHandler,
		ApplicationWorkflow:   applicationWorkflowHandler,
		ClosingPipeline:       closingPipelineHandler,
		Behavioral:            behavioralHandler,
		InsightsAPI:           handlers.NewInsightsAPIHandlers(gormDB),
		ContextFUB:            contextFUBHandler,
		Calendar:              calendarHandler,
		DataMigration:         dataMigrationHandler,
		EmailSender:           emailSenderHandler,
		Unsubscribe:           unsubscribeHandler,
		HARMarket:             harMarketHandler,
		LeadReengagement:      leadReengagementHandler,
		LeadsList:             leadsListHandler,
		PreListing:            preListingHandler,
		Properties:            propertiesHandler,
		SavedProperties:       savedPropertiesHandler,
		Recommendations:       recommendationsHandler,
		PropertyAlerts:        propertyAlertsHandler,
		LiveActivity:          liveActivityHandler,
		SecurityMonitoring:    securityMonitoringHandler,
		AdvancedSecurityAPI:   advancedSecurityAPIHandler,
		Webhook:               webhookHandler,
		DB:                    gormDB,
	}
	log.Println("📦 Handler struct initialized for route registration")


        // Initialize Gin with enterprise security
        gin.SetMode(gin.ReleaseMode)
        r := gin.Default()

        // Enterprise template functions
        r.SetFuncMap(template.FuncMap{
                "safeHTML": func(html string) template.HTML { return template.HTML(html) },
                "safeCSS": func(css string) template.CSS { return template.CSS(css) },
                "safeURL": func(rawURL string) template.URL {
                        if u, err := url.Parse(rawURL); err == nil {
                                return template.URL(u.String())
                        }
                        return template.URL("")
                },
                "formatPrice": func(price interface{}) string {
                        switch v := price.(type) {
                        case int:
                                return fmt.Sprintf("$%,d", v)
                        case int64:
                                return fmt.Sprintf("$%,d", v)
                        case float64:
                                return fmt.Sprintf("$%,.2f", v)
                        case float32:
                                return fmt.Sprintf("$%,.2f", v)
                        case string:
                                if f, err := strconv.ParseFloat(v, 64); err == nil {
                                        return fmt.Sprintf("$%,.2f", f)
                                }
                                return v
                        default:
                                return fmt.Sprintf("%v", v)
                        }
                },
                "formatDate": func(t interface{}) string {
                        switch v := t.(type) {
                        case time.Time:
                                return v.Format("January 2, 2006")
                        case *time.Time:
                                if v != nil {
                                        return v.Format("January 2, 2006")
                                }
                                return ""
                        default:
                                return fmt.Sprintf("%v", v)
                        }
                },
                "currentYear": func() int {
                        return time.Now().Year()
                },
                "formatNumber": func(num interface{}) string {
                        switch v := num.(type) {
                        case int:
                                return fmt.Sprintf("%,d", v)
                        case int64:
                                return fmt.Sprintf("%,d", v)
                        case float64:
                                return fmt.Sprintf("%,.0f", v)
                        case float32:
                                return fmt.Sprintf("%,.0f", v)
                        default:
                                return fmt.Sprintf("%v", v)
                        }
                },
                "upper": func(s string) string {
                        return strings.ToUpper(s)
                },
                "lower": func(s string) string {
                        return strings.ToLower(s)
                },
                "title": func(s string) string {
                        return strings.Title(s)
                },
                "urlEncode": func(s string) string {
                        return url.QueryEscape(s)
                },
        })

        // Load all templates including components  
        // Use filepath.Walk to load ALL templates recursively
        templateFiles, err := filepath.Glob("web/templates/**/**/*.html")
        if err == nil && len(templateFiles) > 0 {
                r.LoadHTMLFiles(templateFiles...)
        } else {
                // Fallback: load templates manually
                var templates []string
                filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
                        if err == nil && !info.IsDir() && filepath.Ext(path) == ".html" {
                                templates = append(templates, path)
                        }
                        return nil
                })
                if len(templates) > 0 {
                        r.LoadHTMLFiles(templates...)
                        log.Printf("📄 Loaded %d templates", len(templates))
                }
        }
        r.Static("/static", "./web/static")

        // Enterprise security middleware (exclude static files)
        r.Use(func(c *gin.Context) {
                // Don't apply nosniff to static files
                if !strings.HasPrefix(c.Request.URL.Path, "/static/") {
                        c.Header("X-Content-Type-Options", "nosniff")
                }
                c.Header("X-Frame-Options", "DENY") 
                c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Strict-Transport-Security", "max-age=31536000")
		c.Next()
	})

	// ISSUE #3 FIX: Apply CSRF protection middleware globally
	r.Use(middleware.CSRFProtection())

	// ===== TEMPLATE ROUTES (All 35+ Templates) =====

	// Register all routes
	log.Println("🛣️ Registering consumer routes...")
	RegisterConsumerRoutes(r, allHandlers)
	log.Println("✅ Consumer routes registered")

	log.Println("🛣️ Registering admin routes...")
	RegisterAdminRoutes(r, allHandlers, propertyHubAI)
	log.Println("✅ Admin routes registered")

	log.Println("🛣️ Registering API routes...")
	api := r.Group("/api")
	RegisterAPIRoutes(api, allHandlers, propertyValuationHandler, emailAutomationHandler)
// DISABLED - causes duplicate routes: 	RegisterMissingRoutes(api) // 55 missing endpoints
	log.Println("✅ API routes registered")

	log.Println("🛣️ Registering health check and error handlers...")
	RegisterHealthRoutes(r, gormDB, authManager, encryptionManager)
	log.Println("✅ Health check and error handlers registered")


	// Start enterprise system
        port := os.Getenv("PORT")
        if port == "" {
                port = "8080"
        }

        log.Printf("🌐 PropertyHub Enterprise System running at http://localhost:%s", port)
        log.Println("🏠 All enterprise handlers registered and active")
        log.Println("📊 Database: Connected and ready")
        log.Println("🔐 Security: Enterprise level")
        log.Printf("🔗 Integrations: %d external services active", func() int {
                count := 0
                if scraperService != nil { count++ }
                if cfg.FUBAPIKey != "" { count++ }
                return count
        }())


	// Start server
	r.Run(":" + port)
}
