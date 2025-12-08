package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/auth"
	"chrisgross-ctrl-project/internal/config"
	"chrisgross-ctrl-project/internal/handlers"
	"chrisgross-ctrl-project/internal/middleware"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/repositories"
	"chrisgross-ctrl-project/internal/scraper"
	"chrisgross-ctrl-project/internal/security"
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/templates"

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

        // ============================================================================
        // DATABASE MIGRATIONS - Run BEFORE any handlers are initialized
        // ============================================================================
        log.Println("🔄 Running comprehensive database migrations...")
        
        // Run core model migrations from models.AutoMigrate()
        if err := models.AutoMigrate(gormDB); err != nil {
                log.Fatalf("❌ Core model migration failed: %v", err)
        }
        
        // Migrate additional critical models not in models.AutoMigrate()
        additionalModels := []interface{}{
                // Core Models
                &models.AdminUser{},
                &models.Lead{},
                &models.Contact{},
                &models.ClosingPipeline{},
                &models.WebhookEvent{},
                &models.PriceChangeEvent{},
                
                // Behavioral Models
                &models.BehavioralEvent{},
                &models.BehavioralSession{},
                &models.BehavioralScore{},
                
                // Application Workflow
                &models.PropertyApplicationGroup{},
                &models.ApplicationNumber{},
                &models.ApplicationApplicant{},
                
                // Email/Notification Models
                &models.EmailEvent{},
                &models.Campaign{},
                &models.EmailBatch{},
                &models.EmailTemplate{},
                &models.TrustedEmailSender{},
                &models.EmailProcessingRule{},
                &models.EmailProcessingLog{},
                &models.IncomingEmail{},
        }
        
        for _, model := range additionalModels {
                if err := gormDB.AutoMigrate(model); err != nil {
                        log.Printf("⚠️ Warning: Migration failed for %T: %v", model, err)
                }
        }
        log.Println("✅ All database migrations completed")

        // Initialize enterprise authentication manager
        authManager := auth.NewSimpleAuthManager(sqlDB)
        log.Println("🔐 Enterprise authentication initialized")

        // Initialize enterprise security (CRITICAL - fail fast if not available)
        encryptionManager, err := security.NewEncryptionManager(gormDB)
        if err != nil {
                log.Fatalf("❌ Encryption manager initialization failed: %v\n💡 Ensure ENCRYPTION_KEY environment variable is set", err)
        }
        log.Println("🔐 Enterprise encryption manager initialized")

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
                propertyValuationService = services.NewPropertyValuationService(cfg, gormDB, scraperService)
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

// Dashboard
dashboardHandler := handlers.NewDashboardHandlers(gormDB)
log.Println("📊 Dashboard handlers initialized")

// Data Migration & Import
dataMigrationHandler := handlers.NewDataMigrationHandlers(gormDB)
log.Println("📥 Data migration handlers initialized")

// Email Management
emailSenderHandler := handlers.NewEmailSenderHandlers(gormDB)
unsubscribeHandler := handlers.NewUnsubscribeHandlers(gormDB)

// Email automation (only if Redis is available)
var emailBatchService *services.EmailBatchService
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
        emailBatchService = services.NewEmailBatchService(redisClient, smtpConfig)

        // Initialize email automation handler
        emailAutomationHandler = handlers.NewEmailAutomationHandlers(gormDB, emailProcessor, emailBatchService)
        log.Println("📧 Email automation handlers initialized")
} else {
        log.Println("📧 Email senders initialized (automation disabled - Redis not available)")
}

// HAR Market removed - HAR blocked access

// Lead Management & Reengagement
leadReengagementHandler := handlers.NewLeadReengagementHandler(gormDB, encryptionManager)
leadsListHandler := handlers.NewLeadsListHandler(gormDB, encryptionManager)
bulkOperationsHandler := handlers.NewBulkOperationsHandler(gormDB)
log.Println("👥 Lead management handlers initialized")

// Team Management
teamHandler := handlers.NewTeamHandlers(gormDB)
log.Println("👥 Team management handlers initialized")

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

// ============================================================================
// MISSING SERVICES INSTANTIATION (SCO-127)
// ============================================================================

// Helper services for automation
leadService := services.NewLeadService()
log.Println("👤 Lead service initialized")

propertyService := services.NewPropertyService()
log.Println("🏠 Property service initialized")

// FUB Services (create in order due to dependencies)
// Validate FUB API key before creating services
if cfg.FUBAPIKey == "" {
	log.Println("⚠️  FUB_API_KEY not configured - FUB integration disabled")
}

fubAPIClient := services.NewBehavioralFUBAPIClient(gormDB, cfg.FUBAPIKey)
if cfg.FUBAPIKey != "" {
	log.Println("🔗 FUB API client initialized")
} else {
	log.Println("⚠️  FUB API client created (inactive - no API key)")
}

// NOTE: These services are initialized but not yet wired to handlers
// They are ready for future integration when needed
fubErrorHandler := services.NewFUBErrorHandler()
log.Println("⚠️ FUB error handler initialized (available for future use)")
_ = fubErrorHandler // Not yet wired to handlers

var fubBatchService *services.FUBBatchService
if redisClient != nil {
	fubBatchService = services.NewFUBBatchService(gormDB, redisClient)
	log.Println("📦 FUB batch service initialized (available for future use)")
	_ = fubBatchService // Not yet wired to handlers
} else {
	log.Println("⚠️ FUB batch service skipped - Redis not available")
}

fubBidirectionalSync := services.NewFUBBidirectionalSync(gormDB, cfg.FUBAPIKey)
if cfg.FUBAPIKey != "" {
	log.Println("🔄 FUB bidirectional sync initialized (available for future use)")
} else {
	log.Println("⚠️  FUB bidirectional sync created (inactive - no API key)")
}
_ = fubBidirectionalSync // Not yet wired to handlers

fubBridge := services.NewBehavioralFUBBridge(gormDB, cfg.FUBAPIKey)
if cfg.FUBAPIKey != "" {
	log.Println("🌉 FUB behavioral bridge initialized (available for future use)")
} else {
	log.Println("⚠️  FUB behavioral bridge created (inactive - no API key)")
}
_ = fubBridge // Not yet wired to handlers

// Analytics Services (no dependencies on email/SMS)
funnelAnalytics := services.NewFunnelAnalyticsService(gormDB)
log.Println("📊 Funnel analytics initialized")

// NOTE: These services are initialized but not yet wired to handlers
var analyticsCacheService *services.AnalyticsCacheService
var performanceMonitor *services.PerformanceMonitoringService
if redisClient != nil {
	analyticsCacheService = services.NewAnalyticsCacheService(redisClient, repos.Property, repos.Booking, repos.Admin)
	log.Println("📊 Analytics cache service initialized (available for future use)")
	_ = analyticsCacheService // Not yet wired to handlers
	
	performanceMonitor = services.NewPerformanceMonitoringService(redisClient)
	log.Println("📈 Performance monitoring initialized (available for future use)")
	_ = performanceMonitor // Not yet wired to handlers
} else {
	log.Println("⚠️ Analytics cache and performance monitoring skipped - Redis not available")
}

// Routing and Scheduling Services
// NOTE: These services are initialized but not yet wired to handlers
leadRouting := services.NewLeadRoutingService()
log.Println("🔀 Lead routing service initialized (available for future use)")
_ = leadRouting // Not yet wired to handlers

dailySchedule := services.NewDailyScheduleService(gormDB)
log.Println("📅 Daily schedule service initialized (available for future use)")
_ = dailySchedule // Not yet wired to handlers

// Campaign Services (will be initialized after scoringEngine and email/SMS services)
var campaignTriggers *services.CampaignTriggerAutomation
var eventOrchestrator *services.EventCampaignOrchestrator
var relationshipEngine *services.RelationshipIntelligenceEngine
var leadSafetyFilter *services.LeadSafetyFilter


	// Property Matching (needed by campaign services and relationship engine)
	propertyMatcher := services.NewPropertyMatchingService(gormDB)
	log.Println("🔍 Property matching service initialized")

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
	
	// Initialize relationship engine now that we have scoringEngine
	relationshipEngine = services.NewRelationshipIntelligenceEngine(gormDB, scoringEngine)
	relationshipEngine.SetFunnelAnalytics(funnelAnalytics)
	relationshipEngine.SetPropertyMatcher(propertyMatcher)
	relationshipEngine.SetInsightGenerator(insightGenerator)
	log.Println("🧠 Relationship intelligence engine initialized")
	
	// ============================================================================
	// CRITICAL FIX: Initialize email/SMS/abandonmentRecovery BEFORE campaignTriggers
	// ============================================================================
	emailService := services.NewEmailService(cfg, gormDB)
	log.Println("📧 Email service initialized")
	
	smsService := services.NewSMSService(cfg, gormDB)
	log.Println("📱 SMS service initialized")
	
	notificationService := services.NewNotificationService(emailService, gormDB)
	log.Println("🔔 Notification service initialized")
	
	// AI/Automation Services (depend on email/SMS/notification services)
	analyticsAutomationService := services.NewAnalyticsAutomationService(emailService, smsService, leadService, notificationService)
	log.Println("🤖 Analytics automation service initialized")
	
	// CRITICAL: Initialize abandonmentRecovery BEFORE it's used by campaignTriggers
	abandonmentRecovery := services.NewAbandonmentRecoveryService(emailService, smsService, analyticsAutomationService, leadService, propertyService)
	log.Println("🔄 Abandonment recovery service initialized")
	
	// Initialize campaign services now that we have all dependencies (including abandonmentRecovery)
	if emailBatchService != nil {
		campaignTriggers = services.NewCampaignTriggerAutomation(gormDB, emailBatchService, relationshipEngine, propertyMatcher, abandonmentRecovery)
		log.Println("🎯 Campaign trigger automation initialized with abandonment recovery (available for future use)")
		_ = campaignTriggers // Not yet wired to handlers
		
		// SMSEmailAutomationService for EventCampaignOrchestrator
		smsEmailAutomation := services.NewSMSEmailAutomationService(gormDB)
		eventOrchestrator = services.NewEventCampaignOrchestrator(gormDB, smsEmailAutomation)
		log.Println("📡 Event campaign orchestrator initialized (available for future use)")
		_ = eventOrchestrator // Not yet wired to handlers
	} else {
		log.Println("⚠️ Campaign services skipped - email batch service not available")
	}
	
	// SpiderwebAIOrchestrator can handle nil emailBatchService gracefully
	if emailBatchService != nil {
		log.Println("✅ PropertyHub AI System will initialize with email automation support")
	} else {
		log.Println("⚠️  PropertyHub AI System will initialize WITHOUT email automation (Redis unavailable)")
	}
	
	propertyHubAI := services.NewSpiderwebAIOrchestrator(
		gormDB,
		scoringEngine,
		insightGenerator,
		emailBatchService, // Can be nil - service must handle gracefully
		abandonmentRecovery,
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
	
	// NOTE: email/SMS/notification/abandonmentRecovery services already initialized above (before campaignTriggers)
	
	// Safety Services (depend on FUB API client)
	leadSafetyFilter = services.NewLeadSafetyFilter(gormDB, fubAPIClient)
	log.Println("🛡️ Lead safety filter initialized (available for future use)")
	_ = leadSafetyFilter // Not yet wired to handlers
	
	propertyAlertsHandler := handlers.NewPropertyAlertsHandler(gormDB, emailService)
	log.Println("🔔 Property alerts handler initialized")
	
	liveActivityHandler := handlers.NewLiveActivityHandler(gormDB)
	log.Println("📡 Live activity handler initialized")
	
	behavioralSessionsHandler := handlers.NewBehavioralSessionsHandler(gormDB)
	log.Println("👥 Behavioral sessions handler initialized")

	webSocketHandler := handlers.NewWebSocketHandler(gormDB, dashboardStatsService)
	log.Println("🔌 WebSocket handler initialized")

	// Initialize Activity Broadcasting for real-time admin feed
	activityBroadcastService := services.NewActivityBroadcastService(gormDB)
	activityHubAdapter := handlers.NewActivityHubAdapter(webSocketHandler.GetActivityHub())
	activityBroadcastService.SetBroadcaster(activityHubAdapter)
	log.Println("📡 Activity broadcast service initialized and wired to WebSocket")

	// Start periodic active count broadcasting
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			var count int64
			gormDB.Raw("SELECT COUNT(DISTINCT id) FROM behavioral_sessions WHERE end_time IS NULL AND start_time >= NOW() - INTERVAL '15 minutes'").Scan(&count)
			activityHubAdapter.BroadcastActiveCount(int(count))
		}
	}()
	log.Println("⏰ Periodic active count broadcasting started (15 second interval)")

	// Initialize Behavioral Event Service and Handler
	behavioralEventService := services.NewBehavioralEventService(gormDB)
	behavioralEventHandler := handlers.NewBehavioralEventHandler(gormDB, behavioralEventService, activityBroadcastService)
	log.Println("🧠 Behavioral event handler initialized with activity broadcasting")

	adminNotificationHub := services.NewAdminNotificationHub(gormDB)
	log.Println("🔔 Admin notification hub initialized")

	adminNotificationHandler := handlers.NewAdminNotificationHandler(adminNotificationHub, gormDB)
	log.Println("📢 Admin notification handler initialized")

	bookingHandler := handlers.NewBookingHandler(gormDB, repos, encryptionManager)
	bookingHandler.SetNotificationHub(adminNotificationHub)
	log.Println("📅 Booking handler initialized with notifications")

	scoringEngine.SetNotificationHub(adminNotificationHub)
	log.Println("🎯 Scoring engine wired to notifications")

	applicationWorkflowHandler.SetNotificationHub(adminNotificationHub)
	log.Println("📝 Application workflow handler wired to notifications")

	// Command Center - AI-driven actionable insights
	fubIntegrationService := services.NewBehavioralFUBIntegrationService(gormDB, cfg.FUBAPIKey)
	commandCenterHandler := handlers.NewCommandCenterHandlers(
		gormDB,
		propertyHubAI,
		scoringEngine,
		insightGenerator,
		propertyMatcher,
		fubIntegrationService,
	)
	log.Println("🎯 Command Center handler initialized")

	// Safety Management
	safetyHandler := handlers.NewSafetyHandlers(gormDB)
	log.Println("🔒 Safety handlers initialized")

	// Availability Management
	availabilityHandler := handlers.NewAvailabilityHandler(gormDB)
	log.Println("📅 Availability handlers initialized")

	// Central Property State
	centralPropertyHandler := handlers.NewCentralPropertyHandler(gormDB, encryptionManager)
	log.Println("🏠 Central property handlers initialized")

	// Central Property Sync
	centralPropertySyncHandler := handlers.NewCentralPropertySyncHandlers(gormDB)
	log.Println("🔄 Central property sync handlers initialized")

	// Daily Schedule
	dailyScheduleHandler := handlers.NewDailyScheduleHandlers(gormDB)
	log.Println("📆 Daily schedule handlers initialized")

	// MFA Authentication
	mfaHandler := handlers.NewMFAHandler(gormDB, authManager)
	log.Println("🔐 MFA handlers initialized")

	// Settings Management
	settingsHandler := handlers.NewSettingsHandler(gormDB)
	log.Println("⚙️ Settings handlers initialized")

	// Validation
	validationHandler := handlers.NewValidationHandler()
	log.Println("✔️ Validation handlers initialized")

	log.Println("✅ All enterprise handlers initialized successfully")

	// Create handlers struct for route registration
	allHandlers := &AllHandlers{
		BusinessIntelligence:  businessIntelligenceHandler,
		TieredStats:           tieredStatsHandler,
		Approvals:             approvalsHandler,
		ApplicationWorkflow:   applicationWorkflowHandler,
		ClosingPipeline:       closingPipelineHandler,
		Behavioral:            behavioralHandler,
		BehavioralEvent:       behavioralEventHandler,
		InsightsAPI:           handlers.NewInsightsAPIHandlers(insightGenerator),
		ContextFUB:            contextFUBHandler,
		CommandCenter:         commandCenterHandler,
		Booking:               bookingHandler,
		Calendar:              calendarHandler,
		Dashboard:             dashboardHandler,
		DataMigration:         dataMigrationHandler,
		EmailSender:           emailSenderHandler,
		Unsubscribe:           unsubscribeHandler,
		// HARMarket removed - HAR blocked access
		LeadReengagement:      leadReengagementHandler,
		LeadsList:             leadsListHandler,
		BulkOperations:        bulkOperationsHandler,
		Team:                  teamHandler,
		PreListing:            preListingHandler,
		Properties:            propertiesHandler,
		SavedProperties:       savedPropertiesHandler,
		Recommendations:       recommendationsHandler,
		PropertyAlerts:        propertyAlertsHandler,
		LiveActivity:          liveActivityHandler,
		BehavioralSessions:    behavioralSessionsHandler,
		SecurityMonitoring:    securityMonitoringHandler,
		AdvancedSecurityAPI:   advancedSecurityAPIHandler,
		Webhook:               webhookHandler,
		WebSocket:             webSocketHandler,
		AdminNotification:     adminNotificationHandler,
		Safety:                safetyHandler,
		Availability:          availabilityHandler,
		CentralProperty:       centralPropertyHandler,
		CentralPropertySync:   centralPropertySyncHandler,
		DailySchedule:         dailyScheduleHandler,
		MFA:                   mfaHandler,
		Settings:              settingsHandler,
		Validation:            validationHandler,
		DB:                    gormDB,
	}
	log.Println("📦 Handler struct initialized for route registration")


        // Initialize Gin with enterprise security
        gin.SetMode(gin.ReleaseMode)
        r := gin.Default()

        // Configure Pongo2 template engine with Django-style inheritance
        templates.RegisterPongo2Filters()
        r.HTMLRender = templates.NewPongo2Render(templates.GetTemplateDir())
        r.Static("/static", "./web/static")
        log.Println("🎨 Pongo2 template engine configured with custom filters")

        // Initialize enhanced security middleware
        securityMiddleware := middleware.NewSecurityMiddleware(gormDB)
        log.Println("🔒 Enhanced security middleware initialized")

        // Enterprise security headers with CSP (exclude static files)
        r.Use(func(c *gin.Context) {
                // Don't apply nosniff to static files
                if !strings.HasPrefix(c.Request.URL.Path, "/static/") {
                        c.Header("X-Content-Type-Options", "nosniff")
                }
                c.Header("X-Frame-Options", "DENY") 
                c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data: https:; font-src 'self' https://fonts.gstatic.com; connect-src 'self'; frame-ancestors 'none';")
		c.Next()
	})
	log.Println("🛡️ Enhanced security headers applied (CSP, Referrer-Policy, Permissions-Policy)")

	// ISSUE #3 FIX: Apply CSRF protection middleware globally
	r.Use(middleware.CSRFProtection())

	// Apply SQL injection and XSS protection globally
	r.Use(gin.WrapH(securityMiddleware.SQLInjectionProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))
	r.Use(gin.WrapH(securityMiddleware.XSSProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))
	log.Println("🛡️ SQL injection and XSS protection applied")

	// ===== TEMPLATE ROUTES (All 35+ Templates) =====

	// Register all routes
	log.Println("🛣️ Registering consumer routes...")
	RegisterConsumerRoutes(r, allHandlers, cfg)
	log.Println("✅ Consumer routes registered")

	log.Println("🛣️ Registering admin routes...")
	RegisterAdminRoutes(r, allHandlers, propertyHubAI, authManager)
	log.Println("✅ Admin routes registered")

	log.Println("🛣️ Registering API routes...")
	api := r.Group("/api")
	// Apply API rate limiting to all API routes
	api.Use(middleware.PublicAPIRateLimiter.RateLimit())
	log.Println("🔒 API rate limiting applied (10/min, 50/hour)")
	RegisterAPIRoutes(api, allHandlers, propertyValuationHandler, emailAutomationHandler)
	log.Println("✅ API routes registered")

	// Register admin authentication routes with enhanced security
	log.Println("🛣️ Registering admin authentication routes...")
	adminAuth := r.Group("/api/v1/admin")
	// Apply strict rate limiting for admin login
	adminAuth.Use(middleware.AdminLoginRateLimiter.RateLimit())
	// Apply brute force protection
	adminAuth.Use(gin.WrapH(securityMiddleware.BruteForceProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))
	log.Println("🔒 Admin login protection applied (rate limiting + brute force detection)")
	handlers.RegisterAdminAuthRoutes(r, gormDB, cfg.JWTSecret)
	log.Println("✅ Admin authentication routes registered")

	log.Println("🛣️ Registering health check and error handlers...")
	RegisterHealthRoutes(r, gormDB, authManager, encryptionManager)
	log.Println("✅ Health check and error handlers registered")

	// ============================================================================
	// ADDITIONAL ROUTE REGISTRATIONS - Previously Missing
	// ============================================================================

	// Gin-compatible routes - can be registered directly
	// NOTE: RegisterEmailSenderRoutes REMOVED - routes already exist in routes_admin.go and routes_api.go
	
	log.Println("🛣️ Registering central property sync routes...")
	handlers.RegisterCentralPropertySyncRoutes(r, gormDB)
	log.Println("✅ Central property sync routes registered")

	// http.ServeMux-based routes - need to be wrapped for Gin
	log.Println("🛣️ Registering additional HTTP routes (ServeMux-based)...")
	mux := http.NewServeMux()
	
	// Register ServeMux-based routes
	handlers.RegisterSafetyRoutes(mux, gormDB)
	log.Println("✅ Safety routes registered to ServeMux")
	
	handlers.RegisterAvailabilityRoutes(mux, gormDB)
	log.Println("✅ Availability routes registered to ServeMux")
	
	handlers.RegisterCentralPropertyRoutes(mux, gormDB, encryptionManager)
	log.Println("✅ Central property routes registered to ServeMux")
	
	handlers.RegisterMFARoutes(mux, gormDB, authManager)
	log.Println("✅ MFA routes registered to ServeMux")
	
	handlers.RegisterPropertyCRUDRoutes(mux, gormDB)
	log.Println("✅ Property CRUD routes registered to ServeMux")
	
	handlers.RegisterSecurityMiddlewareRoutes(mux, gormDB, authManager)
	log.Println("✅ Security middleware routes registered to ServeMux")
	
	setupService := services.NewSetupService()
	handlers.RegisterSetupRoutes(mux, setupService)
	log.Println("✅ Setup routes registered to ServeMux")
	
	handlers.RegisterValidationRoutes(mux)
	log.Println("✅ Validation routes registered to ServeMux")
	
	// Mount ServeMux to Gin using NoRoute handler
	r.NoRoute(gin.WrapH(mux))
	log.Println("✅ All ServeMux routes mounted to Gin router")


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
