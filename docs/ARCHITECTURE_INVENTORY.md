# PropertyHub Architecture Inventory
Last Updated: December 5, 2025

## How to Use This Document
Before creating any task or implementing any feature:
1. **Search this document** for related functionality using Ctrl+F
2. **Check the status** (✅ COMPLETE / ⚠️ PARTIAL / ❌ STUB)
3. **Review TODO comments** with line numbers to see what's missing
4. **Only implement what's actually missing** - avoid rebuilding working code

---

## Services Inventory (@internal/services/)

### ✅ COMPLETE - Do Not Rebuild

| Service | File | Key Functions | Notes |
|---------|------|---------------|-------|
| Abandonment Recovery | abandonment_recovery.go | StartRecovery(), processRecoveryRoutine(), executeRecoveryStep(), sendRecoveryEmail(), sendRecoverySMS() | Multi-step recovery campaigns with A/B testing and time-decay urgency |
| Analytics Automation | analytics_automation.go | ProcessEvent(), triggerAbandonmentRecovery(), triggerHighIntentAlert(), triggerVIPWelcome(), triggerExtendedEngagement() | Full automation rule engine with priorities and cooldowns; returns mock recommendations (line 703) |
| Analytics Cache | analytics_cache.go | GetPropertyStatistics(), GetCities(), GetFeaturedProperties(), GetBookingStatistics(), GetDashboardMetrics(), InvalidatePropertyAnalytics() | Full Redis integration with TTL management and cache hit/miss tracking |
| Application Service | application_service.go | CreateApplicationRecord(), UpdateApplicationFromLeaseUpdate(), GetApplicationsWithContact(), GetApplicationsByStatus(), GetApplicationStats() | Functional CRUD with lease status parsing and statistics calculation |
| Application Workflow | application_workflow_service.go | ProcessBuildiumEmail(), GetUnassignedApplicants(), GetPropertiesWithApplications(), MoveApplicantToApplication() | Full workflow orchestration; FUB matching is placeholder (line 70) |
| Asset Minifier | asset_minifier.go | MinifyCSS(), MinifyJS(), MinifyFile(), MinifyDirectory(), WatchDirectory(), CleanupOldFiles() | Full regex-based minification with compression tracking and file watching |
| Availability Service | availability_service.go | CheckAvailability(), CreateBlackoutDate(), CreateRecurringBlackout(), CreateVacationBlackout(), CreateGlobalBlackout(), matchesRecurringRule() | Sophisticated layered blackout system with 4 priority levels and recurring pattern support |
| Behavioral Event Service | behavioral_event_service.go | TrackEvent(), TrackPropertyView(), TrackPropertySave(), TrackInquiry(), TrackApplication(), TrackConversion(), StartSession(), EndSession(), TrackFunnelStage(), SaveBehavioralScores() | Full event tracking system with automatic scoring triggers and session lifecycle management |
| Behavioral FUB API Client | behavioral_fub_api_client.go | makeRequest(), CreateContact(), UpdateContact(), CreateDeal(), UpdateDeal(), AssignActionPlan(), TriggerAutomation(), AddToPond(), CreateTask(), GetAgents() | Full FUB API integration with comprehensive error handling and retry logic |
| Behavioral FUB Bridge | behavioral_fub_bridge.go | ProcessBehavioralTriggerForFUB(), convertToPropertyCategoryData(), determinePropertyCategory(), determinePropertyTier(), isLuxuryProperty(), isInvestmentProperty() | Sophisticated behavioral-to-property-category mapping with Houston-specific location normalization |
| Behavioral FUB Integration | behavioral_fub_integration_service.go | ProcessBehavioralTrigger(), createOrUpdateContact(), determineWorkflowAndPriority(), createBehavioralDeal(), calculateDealValue(), calculateRentalCommissionValue(), calculateSalesCommissionValue() | **DETAILED COMMISSION LOGIC**: RENTAL: Monthly rent - $100 per $1000; SALES: 2.5-3% × 60% agent split |
| Behavioral Scoring Engine | behavioral_scoring_engine.go | CalculateScore(), calculateUrgencyScore(), calculateEngagementScore(), calculateFinancialScore(), calculateDecayFactor(), determineSegment(), SaveScore(), RecalculateAllScores() | Full scoring with time decay (1/0.8/0.5/0.2/0.1). Weights: 40% urgency + 40% engagement + 20% financial. Thresholds: Hot (70+), Warm (40-69), Cold (10-39), Dormant (<10) |
| Booking Service | booking_webhook_services.go | CreateBooking(), GetBookings(), GetBookingByID(), UpdateBookingStatus(), CancelBooking(), RescheduleBooking(), GetBookingWithLeadData() | Full booking lifecycle with PII encryption and async FUB conversion |
| Calendar Integration | calendar_integration.go | CreateShowingEvent(), CreateFollowUpEvent(), UpdateEventStatus(), GetUpcomingEvents(), ScheduleAutomaticFollowUp(), GetCalendarStats() | Full calendar management with database persistence; FUB sync is mock (line 153) |
| Campaign Trigger Automation | campaign_trigger_automation.go | ProcessOpportunityTriggers(), ProcessPropertyMatchTriggers(), ProcessTimedTriggers(), ProcessBehavioralTriggers(), RunAllTriggers(), executeEmailCampaign() | Full trigger detection with SQL queries and template engine with variable substitution |
| Email Sender Validation | email_sender_validation_service.go | ValidateSender(), ValidateEmailForProcessing(), calculateProcessingConfidence(), analyzeContentPatterns(), GetTrustedSenderByEmail(), GetSenderStatistics() | Full sender validation with confidence scoring and content pattern analysis for 4 email types |
| FUB Error Handling | fub_error_handling.go | HandleHTTPResponse(), ExecuteWithRetry(), calculateBackoffDelay(), ValidateBehavioralTriggerRequest() | Comprehensive error handling with exponential backoff: 1s → 2s → 4s (max 30s). HTTP status mapping (400, 401, 403, 404, 422, 429, 500, 502-504) |
| Scoring Rules | scoring_rules.go | DefaultScoringRules(), DefaultSegmentThresholds(), GetPoints(), SetPoints() | Simple configuration file with comprehensive point assignments: viewed: 5, saved: 15, applied: 50, converted: 100 |

### ⚠️ PARTIAL - Check Before Work

| Service | File | What's Done | What's Missing | TODO Lines |
|---------|------|-------------|----------------|------------|
| Business Intelligence | business_intelligence_service.go | Real database queries for behavioral scores and leads (lines 346-387); Dashboard metrics with calculations | Property metrics mostly hardcoded/TODO | Lines 461-475: "TODO: Query from Property model", "TODO: Query active listings", "TODO: Query pending sales", "TODO: Calculate average price", "TODO: Calculate from historical data", "TODO: Calculate booking conversion" |

### ❌ STUB - Needs Implementation

| Service | File | Returns | TODO Lines |
|---------|------|---------|------------|
| Webhook Service | booking_webhook_services.go | Nil/empty arrays for all methods | Lines 164, 171, 177: "TODO: Implement webhook processing logic", "TODO: Implement webhook event retrieval", "TODO: Implement webhook event reprocessing" |

### 📋 Services Not Fully Analyzed (Present in Directory)

The following 40 service files exist but were not fully analyzed in detail. They should be reviewed individually when working in their domain:

- canspam_email_service.go
- central_property_state_manager.go
- communication_services.go
- compliance_monitoring.go
- comprehensive_security_service.go
- daily_schedule_service.go
- dashboard_stats_service.go
- data_migration_service.go
- email_batch.go
- email_processor.go
- enhanced_fub_integration_service.go
- event_campaign_orchestrator.go
- event_processor.go
- fub_batch_service.go
- fub_bidirectional_sync.go
- fub_integration_test.go
- funnel_analytics_service.go
- har_market_scraper.go
- insight_generator_service.go
- intelligence_cache_service.go
- lead_routing.go
- lead_safety_filter.go
- override_manager.go
- performance_monitoring.go
- performance_optimization.go
- photo_processing.go
- photo_protection_service.go
- pre_listing_service.go
- property_alerts_service.go
- property_matching_service.go
- property_readiness_service.go
- property_valuation.go
- real_time_sync_service.go
- relationship_intelligence_engine.go
- route_data_service.go
- setup_service.go
- sms_email_automation.go
- spiderweb_ai_orchestrator.go
- storage_service.go
- template_function_service.go

---

## Handlers Inventory (@internal/handlers/)

### ✅ COMPLETE - Fully Functional

| Handler | File | Endpoints | Notes |
|---------|------|-----------|-------|
| Admin Auth | admin_auth_handlers.go | POST /api/v1/admin/login, GET /admin/auth/status | JWT-based authentication with bcrypt password hashing |
| Admin Pages | admin_page_handlers.go | GET /admin/dashboard, /admin/leads, /admin/properties, /admin/bookings, /admin/reports, /admin/settings | HTML template rendering for admin pages |
| Advanced Security API | advanced_security_api_handlers.go | POST /api/v1/security/analyze-threat, GET /api/v1/security/metrics/advanced, POST /api/v1/security/sessions/advanced | Full threat analysis with IP reputation, user agent analysis, request pattern analysis |
| Analytics Handler | analytics_handler.go | ShowAnalyticsDashboard(), getAnalyticsData(), getTotalSalesVolume(), getRevenueChartData() | Full database query implementation for analytics collection and rendering |
| API Handlers | api_handlers.go | GET /api/v1/dashboard/metrics, /api/v1/leads/segments, /api/v1/bookings/metrics, /api/v1/properties/metrics, /api/v1/system/health | All handlers delegate to BusinessIntelligenceService |
| Approvals Management | approvals_management_handlers.go | GET/POST/PATCH/DELETE /api/v1/approvals | Full CRUD with status validation and closing pipeline integration |
| Availability | availability_handlers.go | GET/POST/DELETE /api/v1/availability/*, /api/v1/availability/blackouts/* | Full availability checking with blackout date management |
| Behavioral Analytics | behavioral_analytics_handlers.go | GET /api/v1/behavioral/trends, /api/v1/behavioral/funnel, /api/v1/behavioral/segments, /api/v1/behavioral/heatmap, /api/v1/behavioral/cohorts | Real database queries with trend analysis, funnel calculations, segment analysis, cohort retention |
| Behavioral Sessions | behavioral_sessions_handlers.go | GET /api/v1/behavioral/sessions/active, GET /api/v1/behavioral/sessions/:id/journey | Full session tracking with behavioral score enrichment and journey timeline reconstruction |
| BI Admin | bi_admin_handlers.go | GET /admin/dashboard, /admin/communication-center, POST /api/v1/fub/log-call, POST /api/v1/fub/webhook | Real BI data with complex SQL queries for metrics and behavioral scoring integration |
| Booking | booking_handlers.go | POST/GET/DELETE /api/v1/bookings, GET /api/v1/bookings/:id | Full booking management with FUB lead creation, encryption, availability validation |
| Calendar | calendar_handlers.go | POST /api/v1/calendar/showing, POST /api/v1/calendar/followup, GET /api/v1/calendar/upcoming, PATCH /api/v1/calendar/events/:id/status | Full calendar management with FUB sync and automation triggers |

### ⚠️ PARTIAL - Has Real Implementation + Some Gaps

| Handler | File | Working | TODO/Missing | Line Numbers |
|---------|------|---------|--------------|--------------|
| Advanced Behavioral Analysis | advanced_behavioral_analysis.go | analyzeUrgencyFactors(), analyzeEngagementDepth(), assessFinancialReadiness(), getPropertySpecificActions() | getCommunityGrowthInsights(), getCompetitiveFactors() return hardcoded Houston market data | Returns mock data for market insights |
| Advanced Behavioral Endpoints | advanced_behavioral_endpoints.go | ProcessAdvancedTrigger(), CreateBehavioralProfile(), GetBehavioralInsights() | GetAdvancedAnalytics(), GetAdvancedBehavioralHealth(), GetAdvancedConfiguration() return mock data | Lines 126-164, 316-341: hardcoded data |
| Advanced Behavioral Integration | advanced_behavioral_integration.go | ProcessAdvancedBehavioralTriggers(), CreateAdvancedBehavioralProfile() | GetAdvancedBehavioralMetrics() returns hardcoded metrics | Lines 138-197: mock data |
| Application Workflow | application_workflow_handlers.go | All CRUD operations, behavioral tracking (lines 95-114) | findFUBMatch() is placeholder | Line 316-321: placeholder FUB match |
| Base Handlers | base_handlers.go | GetDashboardStats() works | GetPropertyInsights() returns empty mock data | Lines 55-66: empty insights |
| Behavioral Intelligence | behavioral_intelligence_handlers.go | calculateUrgencyScore(), calculateFinancialScore(), calculateEngagementScore() with real calculations | getActiveBehavioralTriggers(), getIdentifiedPatterns(), getHoustonNeighborhoodIntelligence() return hardcoded data | Lines 274-445: mock trigger/pattern data |
| Business Intelligence | business_intelligence_handlers.go | Real data from services for most endpoints | GetRealtimeStats() returns mock data | Lines 316-341: hardcoded real-time stats |
| Central Property | central_property_handlers.go | CreateOrUpdateProperty(), GetProperty(), GetPropertiesByStatus() | UpdatePropertyStatus() method missing in CentralPropertyStateManager | Line 237: TODO comment |
| Central Property Sync | central_property_sync_handlers.go | Basic sync handlers | UpdatePropertyStatus() method missing | Line 164: TODO comment |
| Enhanced Photo | enhanced_photo_handlers.go | Photo upload/processing working | Authentication context needs implementation | Line 159: "TODO: Get from authentication" |

### ❌ STUB - Returns Mock Data or Minimal Implementation

| Handler | File | Stub Count | TODO Lines |
|---------|------|------------|------------|
| Daily Schedule | daily_schedule_handlers.go | Completion/snooze logic missing | Lines 227, 275: "TODO: Implement actual completion logic", "TODO: Implement actual snooze logic" |
| Data Migration | data_migration_handlers.go | Returns mock history data | Lines 240-297: mock migration history |
| Missing Endpoints | missing_endpoints_handlers.go | 39 functions; email/SMS not sent; webhook config not implemented | Lines 168, 201, 350: Email/SMS stubs; Lines 961, 986, 996: WebhookConfig TODOs |
| Property Valuation | property_valuation_handlers.go | Returns mock valuation data | Lines 143, 160, 322: "TODO: Implement actual property lookup and valuation", "TODO: Implement actual market analysis", "TODO: Implement actual request lookup" |
| Stub Handlers | stub_handlers.go | 8 stub functions | Various placeholder implementations |

### 📋 Handlers Present (60 Total Files)
Additional handlers not fully documented above:
- closing_pipeline_handlers.go
- command_center_handlers.go
- context_fub_integration_handlers.go
- dashboard_handlers.go
- email_automation_handlers.go
- email_sender_handlers.go
- encryption_handlers.go
- fub_sync_status_handlers.go
- har_market_handlers.go
- insights_api_handlers.go
- lead_property_matching_handlers.go
- lead_reengagement_handlers.go
- leads_list_handler.go
- live_activity_handlers.go
- mfa_handlers.go
- pre_listing_handlers.go
- properties_handlers.go
- property_alerts_handlers.go
- property_crud_handlers.go
- rbac_handlers.go
- recommendations_handlers.go
- response_filter.go
- safety_handlers.go
- safety_routes.go
- saved_properties_handlers.go
- security_middleware_handlers.go
- security_monitoring_handlers.go
- security_routes.go
- session_handlers.go
- settings_handlers.go
- setup_handlers.go
- tiered_stats_handlers.go
- unsubscribe_handlers.go
- validation_handlers.go
- webhook_handlers.go
- websocket_handlers.go

---

## Security Inventory (@internal/security/)

### ✅ ALL COMPLETE - Production Ready

| Service | File | Key Functions | Notes |
|---------|------|---------------|-------|
| API Security | api_security.go | CreateAPIKey(), ValidateAPIRequest(), ValidateScope(), GenerateWebhookSignature(), VerifyWebhookSignature(), GetAPIUsageStats(), RevokeAPIKey() | Rate limiting, IP whitelisting, signature verification, async logging, audit trail |
| Audit Logging | audit_logging.go | LogAction(), LogHTTPRequest(), LogSecurityEvent(), LogDataAccess(), LogAdminAction(), GetAuditLogs(), GetSecurityEvents(), GetAuditStatistics(), CleanupOldLogs() | Full audit trail with severity/category determination, sanitization, statistics, retention policy |
| Document Encryption | document_encryption.go | EncryptDocument(), DecryptDocument(), EncryptPIIData(), DecryptPIIData(), GetDocumentsByClient(), DeleteDocument() | AES-GCM encryption, master key management, SHA256 integrity checking, retention periods (4-7 years), PII classification levels |
| Field Encryption | field_encryption.go | Encrypt(), Decrypt(), EncryptEmail(), DecryptEmail(), EncryptPhone(), DecryptPhone(), EncryptSSN(), DecryptSSN(), RotateKey(), GetEncryptionStatistics(), ValidateEncryption() | AES-256-GCM field-level encryption, custom EncryptedString type with GORM support, backward compatibility, custom JSON marshaling, audit logging, 1-year log retention |
| Input Validation | input_validation.go | ValidateEmail(), ValidatePhone(), ValidateName(), ValidateAddress(), ValidateText(), ValidateID(), ValidatePrice(), SanitizeHTML(), IsSQLInjectionAttempt(), IsXSSAttempt(), ValidateBookingRequest(), ValidatePropertyNotificationRequest() | Comprehensive regex patterns for SQL injection and XSS, context-specific sanitization, field type detection, real estate domain-specific validation |
| Password | password.go | HashPassword(), CheckPasswordHash() | Minimal bcrypt implementation with DefaultCost (12 rounds), constant-time comparison |
| Realtime Monitoring | realtime_monitoring.go | NewRealtimeMonitor(), Start(), Stop(), ProcessHTTPRequest(), ReportBruteForceAttempt(), ReportUnauthorizedAccess(), ReportDataAccess(), ReportComplianceViolation(), SubscribeToAlerts(), GetActiveAlerts(), AcknowledgeAlert(), ResolveAlert() | Comprehensive threat detection (SQL injection, XSS, file uploads), background goroutines for alerts/metrics/analysis/cleanup, alert escalation after 30min, coordinated attack detection, channel-based real-time alerts |
| Session Management | session_management.go | CreateSession(), ValidateSession(), InvalidateSession(), InvalidateAllUserSessions(), GetUserSessions(), GetSessionStatistics(), CleanupExpiredSessions() | Device fingerprinting, geolocation tracking (mock fallback), risk scoring based on trust/location/hours/IP changes, session hijacking detection, user agent parsing, 24-hour default expiration |
| SQL Protection | sql_protection.go | NewSQLProtectionMiddleware(), Middleware(), NewSafeDB(), SafeWhere(), SafeFind(), SafeFirst(), SafeCreate(), SafeUpdate(), SafeDelete(), ConfigureSecureDB(), SanitizeOrderBy(), SanitizeLimit(), SanitizeOffset(), ValidateTableName(), ValidateColumnName() | Middleware-based HTTP validation, all GORM operations wrapped, whitelist approach for dynamic identifiers, parameterized query enforcement, safe empty results on injection attempts, query timeout settings |
| Template Security | template_security.go | GetSecureFuncMap(), SafeHTML(), SafeText(), SafeAttribute(), SafeURL(), SafeJavaScript(), SafeJSONAttribute(), FormatPrice(), FormatAddress(), TruncateText(), JSONEncode(), StripHTML(), IsValidURL(), DetectXSS(), SanitizeTemplateData() | Template function map for Go templates, context-aware escaping (HTML, JS, URL, JSON), type-aware formatting, nested map sanitization, comprehensive XSS protection |
| TOTP MFA | totp_mfa.go | GenerateSecret(), VerifyTOTP(), VerifyBackupCode(), IsMFAEnabled(), DisableMFA(), RegenerateBackupCodes(), GetQRCode(), GetMFAStatus(), CleanupExpiredAttempts(), GetMFAStatistics() | TOTP with github.com/pquerna/otp, backup codes (XXXXX-XXXXX format), single-use enforcement, MFA attempt tracking, QR code generation, device fingerprinting, 90-day attempt log retention, metrics on adoption rates |
| TREC Compliance | trec_compliance.go | LogLeadGeneration(), LogShowingScheduled(), LogConsentCollected(), LogIABSDisclosure(), GetComplianceReport(), ArchiveOldEvents(), UpdateCompanyInfo() | TREC compliance event logging with 4-year retention, event categorization (IABS, TRELA, Fair Housing, Data Privacy), compliance gap identification, audit trail per event, report generation, automatic archival |
| XSS Protection | xss_protection.go | SanitizeHTML(), SanitizeHTMLAttribute(), SanitizeJavaScript(), SanitizeURL(), SanitizeForJSON(), NewHTMLSanitizer(), StripAllHTML(), AllowBasicHTML(), NewJavaScriptSanitizer(), RemoveDangerousPatterns(), IsJavaScriptInjection(), NewURLSanitizer(), IsValidURL(), NewContentSecurityPolicy(), GenerateHeader(), SanitizeResponseData(), DetectXSSAttempt() | Multiple specialized sanitizers (HTML, JS, URL, Attributes), regex-based pattern detection/removal, dangerous tag/protocol/event handler removal, URL scheme validation (HTTP/HTTPS only), CSP generation, XSS protection middleware with headers, recursive data sanitization |

**Total Security Services**: 13  
**TODO Comments**: 0  
**Incomplete Implementations**: 0  
**Status**: All production-ready with comprehensive error handling, audit logging, and integration

---

## Middleware Inventory (@internal/middleware/)

### ✅ COMPLETE Middleware

| Middleware | File | Functions | Notes |
|------------|------|-----------|-------|
| Asset Optimization | asset_optimization.go | NewAssetOptimizationMiddleware(), Middleware(), getAssetConfig(), setCacheHeaders(), setSecurityHeaders(), supportsCompression(), serveCompressed(), GetStats() | Gzip compression, cache headers by file type, statistics tracking, security headers for assets |
| Auth Required | auth_required.go | AuthRequired() | Gin middleware for session token validation, supports SimpleAuthManager and CachedSessionManager, redirects to login on failure |
| CORS Config | cors_config.go | GetAllowedOrigin(), SetCORSHeaders() | Reads CORS_ALLOWED_ORIGIN/DOMAIN/BASE_URL, fallback to localhost:8080, sets standard CORS headers |
| CORS Middleware | cors_middleware.go | NewCORSMiddleware(), Middleware(), handlePreflight(), handleActualRequest(), isOriginAllowed(), AddAllowedOrigin(), GetCORSStats(), SecureCORSMiddleware(), DevelopmentCORSMiddleware() | Comprehensive CORS handling with preflight support, wildcard pattern matching, configurable methods/headers/credentials, suspicious pattern detection, production and development presets |
| Endpoint Rate Limiter | endpoint_rate_limiter.go | NewEndpointRateLimiter(), RateLimit(), checkRateLimit(), cleanup(), BookingRateLimiter, AdminLoginRateLimiter, PublicAPIRateLimiter | Per-endpoint rate limiting, per-minute and per-hour tracking, temporary blocking, automatic cleanup, X-RateLimit-* headers, pre-configured limiters |
| Enhanced Security | enhanced_security_middleware.go | NewSecurityMiddleware(), SecurityHeaders(), RateLimit(), IPFiltering(), RequestValidation(), BruteForceProtection(), SQLInjectionProtection(), XSSProtection(), AddToWhitelist(), GetRateLimitStats() | OWASP security headers, rate limiting, IP filtering with whitelist/blacklist, request size validation, user-agent blacklist, brute force detection, SQL injection/XSS pattern detection, audit logging, statistics |
| Gin Wrappers | gin_optimized_wrappers.go | GinCORSWrapper(), GinSecurityWrapper(), GinRateLimitWrapper(), GinSQLProtectionWrapper(), GinXSSProtectionWrapper(), GinBruteForceWrapper(), GinAssetOptimizationWrapper() | Gin-specific wrappers using gin.WrapH() for http.Handler conversion, conditional logic for asset optimization |
| Validation | validation.go | ValidateIDParam(), ValidateAreaParam(), ValidateCityParam(), ValidateTypeParam() | Route parameter validation with positive integer check, stores parsed ID in context |

### ⚠️ PARTIAL Middleware

| Middleware | File | Working | Missing | Line Numbers |
|------------|------|---------|---------|--------------|
| CSRF Protection | csrf_middleware.go | CSRFProtection(), constantTimeEquals(), generateCSRFToken(), isCSRFExemptPath() | Issue #4 marked as fixed in comment but check if implementation is complete | Line 29: Comment about Issue #4 fix |
| JSON Protection | json_protection.go | Validates request size, JSON structure constraints (depth, array length, object keys, string length), detects script/SQL/command injection patterns, sanitization capability | GetValidationStats() returns empty structure - statistics collection not implemented | Lines 370-373: Empty stats return |

### 📋 Note on Duplicate Files
- **enhanced_security_wrapper.go** appears to be a Gin-specific wrapper/duplicate of enhanced_security_middleware.go functionality

**Total Middleware Files**: 11  
**Complete**: 9  
**Partial**: 2  
**TODO Comments**: 2 (CSRF comment, JSON stats comment)

---

## Template Inventory (@web/templates/)

### 📂 Template Organization

| Category | Location | Count | Status |
|----------|----------|-------|--------|
| Admin Pages | admin/pages/ | 27+ templates | ⚠️ Monolithic - each has embedded sidebar/header |
| Consumer Pages | consumer/pages/ | 20+ templates | ⚠️ Monolithic |
| Auth Pages | auth/pages/ | 6 templates | ⚠️ Monolithic |
| Commission Pages | commissions/pages/ | 2 templates | ⚠️ Monolithic |
| Lead Pages | leads/pages/ | 2 templates | ⚠️ Monolithic |
| Error Pages | errors/pages/ | 4 templates | ⚠️ Monolithic |
| Shared Components | shared/components/ | 3 components (admin-header, admin-sidebar, consumer-header) | ❌ NOT USED by most templates |
| Shared Includes | shared/includes/ | 1 component (cookie-consent-banner) | ❌ NOT USED - duplicated across pages |

### ⚠️ CRITICAL TEMPLATE ISSUES

#### 1. Component Reuse NOT Implemented
**Problem**: Shared components exist but are NOT referenced via Go template syntax
- @shared/components/admin-header.html exists but is not included via `{{template}}`
- @shared/components/admin-sidebar.html exists but is not included via `{{template}}`
- Each of 27+ admin templates has **fully embedded sidebar and header inline**
- **Impact**: Changing admin header requires updating 25+ files
- **Line Reference**: @web/templates/README.md documents inheritance pattern that is NOT implemented in actual templates

#### 2. Hardcoded User Data
**Problem**: User information is hardcoded in components
- User name "Christopher Gross" hardcoded at @shared/components/admin-header.html:30
- User initials "CG" hardcoded at line 28
- User role "Administrator" hardcoded
- **Impact**: Multi-user system will display wrong user name for all users

#### 3. No Template Inheritance
**Problem**: README.md documents base template pattern with `{{define}}` and `{{block}}` that is NOT implemented
- Each template includes full `<!DOCTYPE html>` and `<html>` structure
- No base templates in use (admin-base, consumer-base)
- Duplicated across 55+ template files
- **Impact**: Maintenance burden, no single source of truth for page structure

#### 4. CSS File Inconsistency
**Problem**: Mix of minified and non-minified CSS references
- @commissions/pages/commissions-dashboard.html uses `.min.css` files
- Most other admin templates use non-minified `.css` files
- **Impact**: Inconsistent asset loading, potential version mismatches

#### 5. Repository Clutter
**Files to clean up**:
- @admin/admin-dashboard.html.save
- @admin/admin-dashboard.html.save.1
- @admin/backup-20251116/ directory
- @auth/login.html.disabled
- @auth/register.html.disabled
- @consumer/about-fix.tar.gz
- @consumer/fix.tar.gz
- @consumer/html-fix.tar.gz

### ✅ Template Strengths
- Well-organized by function
- Consistent naming conventions
- Modern tech stack (Alpine.js for interactivity)
- Accessibility features (error pages, proper HTTP status handling)
- Security awareness (CSRF tokens in forms)
- SVG sprite approach reduces HTTP requests

### 🔧 Recommended Template Refactor
1. **High Priority**: Implement the Go template base system documented in README
   - Create true `admin-base.html` with `{{block}}` directives
   - Reference components via `{{template "admin-header" .}}`
   - Remove inline sidebar/header duplication from all 27+ admin pages
2. **High Priority**: Fix hardcoded user data - make admin-header.html template-driven with `{{.UserName}}`, `{{.UserRole}}`
3. **Medium Priority**: Standardize CSS references (all minified or all non-minified)
4. **Medium Priority**: Clean up backup files, disabled templates, and tar.gz archives from repository
5. **Low Priority**: Document actual template architecture or implement documented approach

---

## Caching Strategy

### Redis-Based Caching (@internal/services/analytics_cache.go)

| Cache Type | Service Method | TTL | Redis Key Pattern | Notes |
|------------|----------------|-----|-------------------|-------|
| Property Statistics | GetPropertyStatistics() | 5 minutes | `analytics:property_stats` | Fallback to repository on miss |
| Cities List | GetCities() | 1 hour | `analytics:cities` | City listings change infrequently |
| Featured Properties | GetFeaturedProperties() | 10 minutes | `analytics:featured_properties` | Updated more frequently |
| Booking Statistics | GetBookingStatistics() | 5 minutes | `analytics:booking_stats` | Real-time booking data |
| Booking Trends | GetBookingTrends() | 15 minutes | `analytics:booking_trends:{period}` | Hourly/daily/weekly/monthly trends |
| Dashboard Metrics | GetDashboardMetrics() | 2 minutes | `analytics:dashboard_metrics` | High-frequency access |

### Intelligence Cache (@internal/services/intelligence_cache_service.go)

| Cache Type | Expected TTL | Purpose |
|------------|--------------|---------|
| Dashboard Hot | Short (minutes) | Real-time dashboard data |
| Dashboard Warm | Medium (hours) | Less critical dashboard metrics |
| Daily Reports | 24 hours | Daily aggregated reports |

**Cache Invalidation**:
- `InvalidatePropertyAnalytics()` - called on property updates
- `InvalidateBookingAnalytics()` - called on booking updates
- Cache hit/miss tracking enabled for monitoring

---

## Commission Calculation Rules

### Implemented in @internal/services/behavioral_fub_integration_service.go

#### Rental Properties
**Formula**: `Monthly Rent - ($100 for every $1000 of monthly rent)`
- Example: $2000/month rent = $2000 - $200 = $1800 agent commission
- Full commission goes to agent (100%)
- Calculated by `calculateRentalCommissionValue()`

#### Sales Properties
**Formula**: `Property Value × Commission Rate × Agent Split`
- Commission Rate: 2.5% - 3.0% (varies by property value/location)
- Agent Split: 60%
- Broker Split: 40%
- Example: $300,000 property at 3% = $9,000 total commission × 60% = $5,400 agent commission
- Calculated by `calculateSalesCommissionValue()`

#### Default Values by Location
- Houston properties: 3.0% commission rate
- Other locations: 2.5% commission rate
- Luxury properties (>$500k): May have negotiated rates

**Used For**:
- FUB deal value calculation
- Agent compensation tracking
- Performance metrics in BI dashboard

---

## Behavioral Scoring System

### Scoring Components (@internal/services/behavioral_scoring_engine.go)

#### Score Calculation
**Composite Score** = (Urgency × 0.4) + (Engagement × 0.4) + (Financial × 0.2)
- Scale: 0-100
- Real-time recalculation on events

#### Urgency Score (0-100)
Factors:
- Recent activity (last 24 hours: +30 points)
- Property saves (+20 points)
- Application submissions (+40 points)
- Inquiry submissions (+25 points)

#### Engagement Score (0-100)
Factors:
- Number of property views (5 points each, max 40)
- Session duration (20+ minutes: +30 points)
- Return visits (5 points each, max 30)
- Saved properties (10 points each, max 30)

#### Financial Score (0-100)
Factors:
- Application completion: +50 points
- Income verification: +30 points
- Credit check: +20 points

#### Time Decay Factors
Event scores decay over time:
- Day 1: 100% (factor: 1.0)
- Day 2-7: 80% (factor: 0.8)
- Day 8-30: 50% (factor: 0.5)
- Day 31-90: 20% (factor: 0.2)
- Day 90+: 10% (factor: 0.1)

### Segment Thresholds (@internal/services/scoring_rules.go)

| Segment | Score Range | Priority | Actions |
|---------|-------------|----------|---------|
| 🔥 Hot | 70-100 | Immediate contact | Call within 1 hour, VIP treatment |
| 🌡️ Warm | 40-69 | Same-day contact | Email + follow-up within 4 hours |
| ❄️ Cold | 10-39 | Nurture campaign | Drip email campaign, weekly touchpoints |
| 💤 Dormant | 0-9 | Re-engagement | Abandonment recovery, special offers |

### Event Point Values (@internal/services/scoring_rules.go)

| Event Type | Points | Decay Applied |
|------------|--------|---------------|
| Property Viewed | 5 | Yes |
| Property Saved | 15 | Yes |
| Inquiry Submitted | 30 | Yes |
| Application Started | 35 | Yes |
| Application Submitted | 50 | Yes |
| Showing Booked | 40 | Yes |
| Showing Attended | 45 | Yes |
| Conversion (Lease Signed) | 100 | No |
| Email Opened | 3 | Yes |
| Email Clicked | 8 | Yes |

---

## FUB Integration Architecture

### API Client (@internal/services/behavioral_fub_api_client.go)
- Authentication: Basic auth with API key
- Retry Logic: Exponential backoff (1s → 2s → 4s, max 30s)
- Error Handling: HTTP status mapping (400, 401, 403, 404, 422, 429, 500, 502-504)

### Operations Supported
- Contact Management: CreateContact(), UpdateContact()
- Deal Management: CreateDeal(), UpdateDeal()
- Action Plans: AssignActionPlan(), CreateActionPlan()
- Automation: TriggerAutomation()
- Lists: AddToPond(), CreatePond()
- Tasks: CreateTask()
- Data Retrieval: GetAgents(), GetActionPlans(), GetPonds()

### Behavioral Bridge (@internal/services/behavioral_fub_bridge.go)
**Property Category Mapping**:
- Luxury properties: >$500k
- Investment properties: Multiple property views, price comparison behavior
- Student housing: Near universities, specific date patterns
- First-time buyers: Budget-conscious behavior, lots of questions

**Location Normalization**: Houston-specific location mapping to FUB-compatible names

### Integration Service (@internal/services/behavioral_fub_integration_service.go)
**Workflow**:
1. ProcessBehavioralTrigger() orchestrates full workflow
2. createOrUpdateContact() with behavioral data
3. determineWorkflowAndPriority() based on segment/score
4. createBehavioralDeal() with commission calculation
5. assignBehavioralActionPlan() based on segment
6. assignToBehavioralPond() for list management
7. createImmediateActionTask() with behavioral recommendations

---

## Database Models & Relationships

### Core Entities (Inferred from Services)

#### User/Lead Models
- User (authentication)
- Lead (customer tracking)
- BehavioralScore (scoring data)
- BehavioralEvent (event tracking)
- BehavioralSession (session tracking)
- FunnelStage (funnel position)

#### Property Models
- Property (core property data)
- PropertyAnalytics (cached analytics)
- PropertyAlert (user alerts)
- CentralPropertyState (multi-source consolidation)

#### Booking Models
- Booking (showing bookings)
- BlackoutDate (availability blocking)
- CalendarEvent (calendar integration)

#### Application Models
- Application (lease applications)
- Applicant (individual applicants)

#### Communication Models
- EmailEvent (email tracking)
- SMSEvent (SMS tracking)
- TrustedSender (email sender validation)
- Campaign (campaign automation)

#### Security Models
- Session (user sessions)
- APIKey (API authentication)
- AuditLog (security audit trail)
- SecurityEvent (security incidents)
- TRECComplianceEvent (regulatory compliance)
- MFASecret (TOTP secrets)

#### Analytics Models
- DashboardMetric (cached metrics)
- BookingStatistic (booking analytics)
- PropertyStatistic (property analytics)

---

## API Endpoints Summary

### Public API (@internal/handlers/)

#### Properties
- GET /api/v1/properties - List properties
- GET /api/v1/properties/:id - Get property details
- GET /api/v1/properties/search - Search properties

#### Bookings
- POST /api/v1/bookings - Create booking
- GET /api/v1/bookings/:id - Get booking
- DELETE /api/v1/bookings/:id - Cancel booking
- GET /api/v1/bookings - List bookings

#### Availability
- GET /api/v1/availability/check - Check availability
- POST /api/v1/availability/blackouts/create - Create blackout

#### Behavioral Tracking
- POST /api/v1/behavioral/advanced-trigger - Process trigger
- GET /api/v1/behavioral/analytics/advanced - Get analytics
- POST /api/v1/behavioral/profiles - Create profile
- GET /api/v1/behavioral/insights/:session_id - Get insights

### Admin API

#### Dashboard & BI
- GET /api/v1/dashboard/metrics - Dashboard metrics
- GET /api/v1/bi/dashboard - BI dashboard
- GET /api/v1/bi/properties/analytics - Property analytics
- GET /api/v1/bi/bookings/analytics - Booking analytics
- GET /api/v1/bi/leads/funnel - Conversion funnel

#### Application Workflow
- GET /api/v1/application-workflow - Get workflow
- POST /api/v1/property/:propertyId/application - Create application
- POST /api/v1/applicant/move - Move applicant

#### Calendar
- POST /api/v1/calendar/showing - Create showing event
- GET /api/v1/calendar/upcoming - Get upcoming events
- PATCH /api/v1/calendar/events/:id/status - Update event status

#### FUB Integration
- POST /api/v1/fub/log-call - Log call to FUB
- POST /api/v1/fub/log-email - Log email to FUB
- POST /api/v1/fub/webhook - FUB webhook receiver

#### Security
- POST /api/v1/security/analyze-threat - Analyze threat
- GET /api/v1/security/metrics/advanced - Security metrics

### Admin Pages (HTML)
- GET /admin/dashboard
- GET /admin/properties
- GET /admin/leads
- GET /admin/bookings
- GET /admin/reports
- GET /admin/settings
- GET /admin/communication-center
- GET /admin/behavioral-intelligence

---

## Environment Variables Referenced

### Core
- DOMAIN - Primary domain
- BASE_URL - Application base URL
- CORS_ALLOWED_ORIGIN - CORS origin whitelist

### Database
- DATABASE_URL - Database connection string
- REDIS_URL - Redis connection string

### FUB Integration
- FUB_API_KEY - Follow Up Boss API key
- FUB_API_BASE_URL - FUB API endpoint

### Security
- JWT_SECRET - JWT signing secret
- ENCRYPTION_KEY - Field encryption key
- MASTER_ENCRYPTION_KEY - Document encryption master key

### External Services
- SMTP_HOST - Email server
- SMTP_PORT - Email port
- SMTP_USER - Email username
- SMTP_PASSWORD - Email password
- TWILIO_ACCOUNT_SID - SMS service (if implemented)
- TWILIO_AUTH_TOKEN - SMS auth token (if implemented)

---

## Testing Status

### Unit Tests
- @internal/services/fub_integration_test.go exists
- Other unit test coverage unknown - recommend audit

### Integration Tests
- No integration tests identified in codebase scan

### E2E Tests
- No E2E tests identified in codebase scan

**Recommendation**: Establish testing strategy for critical paths:
- Behavioral scoring calculations
- Commission calculations
- FUB integration workflows
- Security middleware chain
- Payment/booking workflows

---

## Known Technical Debt

### High Priority
1. **Template duplication** - 27+ admin templates with embedded sidebar/header (see Template Inventory)
2. **WebhookService stub** - @internal/services/booking_webhook_services.go lines 164, 171, 177
3. **Property metrics TODOs** - @internal/services/business_intelligence_service.go lines 461-475
4. **UpdatePropertyStatus missing** - Referenced in @internal/handlers/central_property_handlers.go line 237
5. **Email/SMS not sent** - @internal/handlers/missing_endpoints_handlers.go lines 168, 201, 350
6. **Property valuation stubs** - @internal/handlers/property_valuation_handlers.go lines 143, 160, 322

### Medium Priority
7. **Mock data in handlers** - Multiple handlers return hardcoded data (see Handler Inventory PARTIAL section)
8. **JSON stats incomplete** - @internal/middleware/json_protection.go lines 370-373
9. **Daily schedule logic** - @internal/handlers/daily_schedule_handlers.go lines 227, 275
10. **WebhookConfig model** - @internal/handlers/missing_endpoints_handlers.go lines 961, 986, 996

### Low Priority
11. **Repository cleanup** - Remove .save files, .tar.gz archives, disabled templates (see Template Inventory)
12. **CSS file standardization** - Mix of minified and non-minified references
13. **Hardcoded user data** - @web/templates/shared/components/admin-header.html lines 28-32

---

## Update Process

### When Completing a Task:

1. **Update this document immediately**
   - Change status from ⚠️ PARTIAL → ✅ COMPLETE
   - Remove TODO line numbers if resolved
   - Add any new files created to appropriate section

2. **Document new functionality**
   - Add new services to Services Inventory
   - Add new handlers to Handlers Inventory
   - Add new endpoints to API Endpoints Summary
   - Update Caching Strategy if cache added

3. **Update related sections**
   - If adding middleware, update Middleware Inventory
   - If adding security features, update Security Inventory
   - If touching templates, update Template Inventory

4. **Commit message format**
   ```
   [ARCHITECTURE] Brief description of change
   
   - Updated ARCHITECTURE_INVENTORY.md
   - Marked [service/handler] as COMPLETE
   - Added [new functionality] to inventory
   ```

5. **Before creating new tasks**
   - Search this document (Ctrl+F) for related functionality
   - Check if feature already exists (may just need bug fix)
   - Check PARTIAL implementations for similar work in progress
   - Review TODO comments for existing placeholders

---

## Quick Reference: Where to Find Things

| What You Need | Where to Look |
|---------------|---------------|
| Behavioral scoring logic | @internal/services/behavioral_scoring_engine.go |
| Commission calculations | @internal/services/behavioral_fub_integration_service.go |
| FUB API integration | @internal/services/behavioral_fub_api_client.go |
| Event tracking | @internal/services/behavioral_event_service.go |
| Campaign automation | @internal/services/campaign_trigger_automation.go |
| Booking management | @internal/services/booking_webhook_services.go (BookingService) |
| Availability checking | @internal/services/availability_service.go |
| Security middleware | @internal/middleware/enhanced_security_middleware.go |
| API authentication | @internal/security/api_security.go |
| Audit logging | @internal/security/audit_logging.go |
| Template security | @internal/security/template_security.go |
| XSS protection | @internal/security/xss_protection.go |
| SQL injection protection | @internal/security/sql_protection.go |
| Dashboard metrics | @internal/handlers/business_intelligence_handlers.go |
| Admin pages | @internal/handlers/bi_admin_handlers.go |
| Behavioral analytics | @internal/handlers/behavioral_analytics_handlers.go |
| Property handlers | @internal/handlers/properties_handlers.go |
| Booking handlers | @internal/handlers/booking_handlers.go |
| Admin templates | @web/templates/admin/pages/ |
| Consumer templates | @web/templates/consumer/pages/ |
| Shared components | @web/templates/shared/components/ |

---

## Document Maintenance

**Last Updated**: December 5, 2025  
**Maintainer**: Development Team  
**Update Frequency**: After every feature completion  
**Review Frequency**: Monthly for accuracy

**Change Log**:
- 2025-12-05: Initial comprehensive architecture inventory created
  - Scanned 59 service files
  - Scanned 60 handler files
  - Documented 13 security services
  - Documented 11 middleware files
  - Analyzed template structure (55+ templates)
  - Identified 14+ TODOs across codebase
  - Documented commission calculations, scoring rules, caching strategy

---

*This document is the single source of truth for PropertyHub architecture. Keep it updated!*
