# PropertyHub Security Architecture

**Last Updated:** December 5, 2025  
**Security Level:** Enterprise-Grade  
**Compliance:** TREC, GDPR, CCPA

## Overview

PropertyHub implements a comprehensive, multi-layered security architecture with 13 integrated security services, real-time threat monitoring, and automated attack prevention. This document outlines the existing security infrastructure and provides operational guidance.

## Table of Contents

- [Security Architecture](#security-architecture)
- [Security Middleware](#security-middleware)
- [Security Services](#security-services)
- [Rate Limiting](#rate-limiting)
- [Configuration](#configuration)
- [Monitoring & Alerting](#monitoring--alerting)
- [Incident Response](#incident-response)
- [Compliance](#compliance)

---

## Security Architecture

### Multi-Layer Defense

PropertyHub implements defense-in-depth with multiple security layers:

```
┌─────────────────────────────────────────┐
│   Layer 1: Network Security Headers    │
│   (CSP, HSTS, X-Frame-Options, etc.)   │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│   Layer 2: Rate Limiting & Throttling  │
│   (Per-endpoint, Per-IP, Burst limits) │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│   Layer 3: Input Validation & Attack   │
│   Detection (SQL, XSS, CSRF, etc.)     │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│   Layer 4: Authentication & MFA         │
│   (TOTP, Session Management, Device)   │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│   Layer 5: Data Encryption              │
│   (AES-256 GCM, Field-level, TLS)      │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│   Layer 6: Audit Logging & Monitoring   │
│   (Real-time alerts, Threat detection) │
└─────────────────────────────────────────┘
```

---

## Security Middleware

### Enhanced Security Middleware
**Location:** `@internal/middleware/enhanced_security_middleware.go`

Provides comprehensive request-level security:

#### Features

1. **Security Headers**
   - Content Security Policy (CSP)
   - HTTP Strict Transport Security (HSTS)
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY
   - X-XSS-Protection: 1; mode=block
   - Referrer-Policy: strict-origin-when-cross-origin
   - Permissions-Policy

2. **Rate Limiting**
   - Per-client tracking with in-memory store
   - Configurable requests per minute/hour
   - Automatic blocking with cooldown periods
   - Burst protection

3. **IP Filtering**
   - Dynamic IP whitelist/blacklist
   - Automatic blacklist on brute force detection
   - Time-based auto-removal from blacklist

4. **Request Validation**
   - Request size limits (default: 10MB)
   - HTTP method validation
   - User-Agent blacklist
   - Suspicious pattern detection

5. **Attack Detection**
   - SQL injection pattern detection
   - XSS attempt detection
   - Path traversal detection
   - Command injection detection

6. **Brute Force Protection**
   - Tracks failed login attempts
   - Auto-blacklist after 5 failures in 15 minutes
   - 1-hour cooldown period

#### Configuration

```go
// Default Security Configuration
SecurityConfig{
    EnableRateLimit:      true,
    EnableIPFiltering:    true,
    EnableCSRFProtection: true,
    EnableXSSProtection:  true,
    EnableClickjacking:   true,
    MaxRequestSize:       10 * 1024 * 1024, // 10MB
    AllowedOrigins:       []string{"*"},
    BlockedUserAgents: []string{
        "bot", "crawler", "spider", "scraper",
        "curl", "wget", "python-requests",
    },
}

// Default Rate Limit Configuration
RateLimitConfig{
    RequestsPerMinute: 60,
    RequestsPerHour:   1000,
    BurstSize:         10,
    BlockDuration:     15 * time.Minute,
}
```

#### IP Management

```go
// Add IP to whitelist (programmatic)
securityMiddleware.AddToWhitelist("192.168.1.100")

// Add IP to blacklist
securityMiddleware.AddToBlacklist("203.0.113.45")

// Remove from lists
securityMiddleware.RemoveFromWhitelist("192.168.1.100")
securityMiddleware.RemoveFromBlacklist("203.0.113.45")

// Get current lists
whitelist := securityMiddleware.GetWhitelist()
blacklist := securityMiddleware.GetBlacklist()
```

### Endpoint Rate Limiters
**Location:** `@internal/middleware/endpoint_rate_limiter.go`

Pre-configured rate limiters for specific endpoints:

#### Rate Limiter Configurations

1. **BookingRateLimiter**
   - 5 requests per minute
   - 20 requests per hour
   - 15-minute block duration
   - Use for: Booking submissions, contact forms

2. **AdminLoginRateLimiter**
   - 3 requests per minute
   - 10 requests per hour
   - 30-minute block duration
   - Use for: Admin login, password reset

3. **PublicAPIRateLimiter**
   - 10 requests per minute
   - 50 requests per hour
   - 5-minute block duration
   - Use for: Public API endpoints

#### Usage

```go
// Apply to Gin routes
router.POST("/api/bookings", BookingRateLimiter.RateLimit(), bookingHandler)
router.POST("/admin/login", AdminLoginRateLimiter.RateLimit(), loginHandler)
router.GET("/api/properties", PublicAPIRateLimiter.RateLimit(), propertiesHandler)
```

### CSRF Protection
**Location:** `@internal/middleware/csrf_middleware.go`

Applied globally to all routes in main.go (line 403):
```go
r.Use(middleware.CSRFProtection())
```

### CORS Configuration
**Location:** `@internal/middleware/cors_middleware.go`

Controls cross-origin resource sharing for API security.

---

## Security Services

### 1. Field Encryption
**Location:** `@internal/security/field_encryption.go`

**Purpose:** Encrypts sensitive PII at the field level using AES-256-GCM

#### Features
- AES-256-GCM encryption with unique nonces
- Automatic key rotation support
- Backward compatibility with plaintext
- Key ID tracking for multi-key support
- Audit logging of encryption operations

#### Usage

```go
// Initialize
encryptionManager, err := security.NewEncryptionManager(db)

// Encrypt sensitive data
encryptedEmail, err := encryptionManager.EncryptEmail("user@example.com")
encryptedPhone, err := encryptionManager.EncryptPhone("555-1234")
encryptedSSN, err := encryptionManager.EncryptSSN("123-45-6789")

// Decrypt
plaintext, err := encryptionManager.Decrypt(encrypted)

// Validate encryption is working
err := encryptionManager.ValidateEncryption()

// Get statistics
stats := encryptionManager.GetEncryptionStatistics()
```

#### Environment Variables
```bash
# Required: 256-bit base64-encoded encryption key
ENCRYPTION_KEY="<base64-encoded-32-byte-key>"
```

**⚠️ CRITICAL:** Store encryption key securely. Loss of key means permanent data loss.

### 2. Audit Logging
**Location:** `@internal/security/audit_logging.go`

**Purpose:** Comprehensive audit trail for compliance and forensics

#### Log Types

1. **General Audit Logs** - All user actions
2. **Security Events** - Security-specific events with risk scoring
3. **Data Access Logs** - Sensitive data access tracking
4. **Admin Actions** - Administrative operations

#### Usage

```go
auditLogger := security.NewAuditLogger(db)

// Log general action
auditLogger.LogAction(userID, sessionID, ipAddress, userAgent, 
    "property_viewed", "properties", propertyID, true, "", metadata)

// Log security event
auditLogger.LogSecurityEvent("login_failure", userID, ipAddress, 
    userAgent, "Invalid password", details, 60) // Risk score 0-100

// Log data access
auditLogger.LogDataAccess(userID, "contacts", recordID, "read",
    []string{"email", "phone"}, oldValues, newValues, ipAddress, 
    userAgent, "Client inquiry")

// Log admin action
auditLogger.LogAdminAction(adminID, "delete_property", "property",
    propertyID, "Property removed", details, ipAddress, userAgent, true, "")
```

#### Retention Policy

```go
// Cleanup old logs (run periodically)
auditLogger.CleanupOldLogs(365) // Keep logs for 365 days
```

### 3. TOTP Multi-Factor Authentication
**Location:** `@internal/security/totp_mfa.go`

**Purpose:** Time-based one-time password (TOTP) for 2FA

#### Features
- RFC 6238 compliant TOTP
- QR code generation for authenticator apps
- 8 backup codes per user
- Device fingerprinting and trust
- MFA attempt logging

#### Usage

```go
totpManager := security.NewTOTPManager(db)

// Setup MFA for user
key, backupCodes, err := totpManager.GenerateSecret(userID, userEmail)
// Display QR code: key.URL() or use totpManager.GetQRCode(key, w)

// Verify TOTP code
valid, err := totpManager.VerifyTOTP(userID, code, ipAddress, userAgent)

// Verify backup code
valid, err := totpManager.VerifyBackupCode(userID, code, ipAddress, userAgent)

// Check if MFA enabled
enabled := totpManager.IsMFAEnabled(userID)

// Disable MFA (admin function)
err := totpManager.DisableMFA(userID)

// Regenerate backup codes
newCodes, err := totpManager.RegenerateBackupCodes(userID)

// Get MFA status
status := totpManager.GetMFAStatus(userID)
```

### 4. SQL Protection
**Location:** `@internal/security/sql_protection.go`

**Purpose:** Prevents SQL injection attacks

#### Features
- Pattern-based SQL injection detection
- Safe database wrapper (SafeDB)
- Query parameter validation
- Database query sanitization

#### Usage

```go
validator := security.NewInputValidator()
sqlProtection := security.NewSQLProtectionMiddleware(validator, logger)

// Apply as HTTP middleware
http.Handle("/api/", sqlProtection.Middleware(apiHandler))

// Use SafeDB wrapper
safeDB := security.NewSafeDB(db, validator, logger)
safeDB.SafeWhere("email = ?", userEmail).First(&user)
safeDB.SafeFind(&properties)

// Sanitize dynamic queries
sanitizer := security.NewDatabaseQuerySanitizer(validator)
safeOrderBy := sanitizer.SanitizeOrderBy(orderBy)
safeLimit := sanitizer.SanitizeLimit(limit)
safeOffset := sanitizer.SanitizeOffset(offset)
```

### 5. XSS Protection
**Location:** `@internal/security/xss_protection.go`

**Purpose:** Prevents cross-site scripting attacks

#### Features
- HTML sanitization
- JavaScript injection detection
- URL validation
- Attribute sanitization
- Content Security Policy generation

#### Usage

```go
xssProtector := security.NewXSSProtector()

// Sanitize user input
safeHTML := xssProtector.SanitizeHTML(userInput)
safeAttr := xssProtector.SanitizeHTMLAttribute(attrValue)
safeJS := xssProtector.SanitizeJavaScript(jsCode)
safeURL := xssProtector.SanitizeURL(url)
safeJSON := xssProtector.SanitizeForJSON(jsonString)

// For template output
safeOutput := xssProtector.SanitizeForHTML(userContent)

// Detect XSS attempts
isXSS, reason := xssProtector.DetectXSSAttempt(input)

// Apply as middleware
xssMiddleware := security.NewXSSProtectionMiddleware()
http.Handle("/api/", xssMiddleware.Middleware(handler))
```

### 6. Password Security
**Location:** `@internal/security/password.go`

**Purpose:** Secure password hashing using bcrypt

#### Usage

```go
// Hash password
hashedPassword, err := security.HashPassword(plainPassword)

// Verify password
valid := security.CheckPasswordHash(plainPassword, hashedPassword)
```

### 7. Input Validation
**Location:** `@internal/security/input_validation.go`

**Purpose:** Comprehensive input validation and sanitization

#### Features
- Email validation (RFC 5322 compliant)
- Phone number validation
- Name validation
- Address validation
- Text validation with length limits
- SQL injection detection
- XSS attempt detection

#### Usage

```go
validator := security.NewInputValidator()

// Validate individual fields
email, err := validator.ValidateEmail(rawEmail)
phone, err := validator.ValidatePhone(rawPhone)
name, err := validator.ValidateName(rawName)
address, err := validator.ValidateAddress(rawAddress)
text, err := validator.ValidateText(rawText, "description", 10, 500)

// Validate complete requests
result := validator.ValidateBookingRequest(requestMap)
if !result.IsValid {
    // Handle validation errors
    for _, err := range result.Errors {
        log.Printf("Field: %s, Error: %s", err.Field, err.Message)
    }
}

// Check for attacks
isSQLInjection := validator.IsSQLInjectionAttempt(input)
isXSS := validator.IsXSSAttempt(input)
```

### 8. Session Management
**Location:** `@internal/security/session_management.go`

**Purpose:** Advanced session tracking with device fingerprinting

#### Features
- Secure session ID generation
- Device fingerprinting
- Geolocation tracking
- Risk scoring (0-100)
- Session anomaly detection
- Trusted device management
- Activity logging

#### Usage

```go
sessionManager := security.NewSessionManager(db)

// Create session on login
session, err := sessionManager.CreateSession(userID, r, "password")

// Validate session on each request
session, err := sessionManager.ValidateSession(sessionID, r)

// Invalidate session
err := sessionManager.InvalidateSession(sessionID, "user_logout")

// Invalidate all user sessions (security event)
err := sessionManager.InvalidateAllUserSessions(userID, "password_reset")

// Get user's active sessions
sessions, err := sessionManager.GetUserSessions(userID)

// Get statistics
stats := sessionManager.GetSessionStatistics()
```

#### Session Risk Scoring

Sessions are automatically scored based on:
- New/untrusted device: +20
- Foreign country: +30
- Unusual hours (2am-6am): +10
- Multiple IPs in short time: +15 per occurrence
- IP/fingerprint change: +50

**High-risk sessions (60+)** trigger additional monitoring and may require re-authentication.

### 9. Real-time Monitoring
**Location:** `@internal/security/realtime_monitoring.go`

**Purpose:** Real-time threat detection and alerting

#### Features
- HTTP request analysis
- Pattern-based threat detection
- Security alert generation
- Automatic escalation
- Threat analytics
- Coordinated attack detection

#### Usage

```go
monitor := security.NewRealtimeMonitor(db, logger, auditLogger)

// Start monitoring
monitor.Start()
defer monitor.Stop()

// Process HTTP requests
alerts := monitor.ProcessHTTPRequest(r)

// Report specific threats
monitor.ReportBruteForceAttempt(username, ipAddress, userAgent, attemptCount)
monitor.ReportUnauthorizedAccess(resource, userID, ipAddress, userAgent)
monitor.ReportDataAccess(dataType, userID, ipAddress, userAgent, recordCount, suspicious)
monitor.ReportComplianceViolation(violationType, description, clientName, propertyAddress)

// Subscribe to alerts
alertChannel := monitor.SubscribeToAlerts("admin-dashboard")
go func() {
    for alert := range alertChannel {
        // Handle alert
        log.Printf("Security Alert: %s - %s", alert.AlertType, alert.Title)
    }
}()

// Get active alerts
alerts, err := monitor.GetActiveAlerts()

// Acknowledge/resolve alerts
monitor.AcknowledgeAlert(alertID, adminUsername)
monitor.ResolveAlert(alertID, adminUsername)
```

#### Alert Types

- `sql_injection` - SQL injection attempt detected
- `xss_attempt` - XSS attack attempt detected
- `brute_force` - Brute force attack detected
- `unauthorized_access` - Unauthorized resource access
- `data_breach` - Potential data breach
- `suspicious_ip` - Suspicious IP activity
- `rate_limit_exceeded` - Rate limit violations
- `suspicious_file_upload` - Dangerous file upload attempt
- `privilege_escalation` - Privilege escalation attempt
- `data_exfiltration` - Large data export detected
- `compliance_violation` - TREC compliance violation

#### Severity Levels

- **Low:** Monitoring only, no immediate action
- **Medium:** Log and monitor, review periodically
- **High:** Alert administrators, investigate
- **Critical:** Immediate action required, auto-block if configured

### 10. Document Encryption
**Location:** `@internal/security/document_encryption.go`

Encrypts documents and files at rest.

### 11. API Security
**Location:** `@internal/security/api_security.go`

API authentication and authorization controls.

### 12. TREC Compliance
**Location:** `@internal/security/trec_compliance.go`

Texas Real Estate Commission compliance monitoring and enforcement.

### 13. Template Security
**Location:** `@internal/security/template_security.go`

Secure template rendering with automatic XSS prevention.

---

## Rate Limiting

### Global Rate Limiting

Applied via enhanced security middleware:

```go
securityMiddleware.RateLimit(middleware.RateLimitConfig{
    RequestsPerMinute: 60,
    RequestsPerHour:   1000,
    BurstSize:         10,
    BlockDuration:     15 * time.Minute,
})
```

### Endpoint-Specific Rate Limiting

```go
// Booking endpoints
router.POST("/api/bookings", middleware.BookingRateLimiter.RateLimit(), handler)

// Admin authentication
router.POST("/admin/login", middleware.AdminLoginRateLimiter.RateLimit(), handler)

// Public API
router.GET("/api/properties", middleware.PublicAPIRateLimiter.RateLimit(), handler)
```

### Rate Limit Response

When rate limit is exceeded:
- **Status Code:** 429 Too Many Requests
- **Headers:** 
  - `X-RateLimit-Limit`: Limit configuration
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Unix timestamp when limit resets
  - `Retry-After`: Seconds until retry allowed
- **Body:** JSON with error details and retry_after

---

## Configuration

### Environment Variables

```bash
# ============================================
# SECURITY CONFIGURATION
# ============================================

# Encryption (REQUIRED)
ENCRYPTION_KEY=<base64-encoded-32-byte-key>

# Database (REQUIRED)
DATABASE_URL=postgres://user:pass@host:5432/dbname

# SMTP for security notifications
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=security@propertyhub.com
SMTP_PASSWORD=<smtp-password>

# Redis (Optional - for advanced caching)
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Session Security
SESSION_TIMEOUT=24h
TRUSTED_PROXIES=192.168.1.0/24

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_REQUESTS_PER_HOUR=1000

# Security Monitoring
SECURITY_MONITORING_ENABLED=true
SECURITY_ALERT_EMAIL=security@propertyhub.com

# CORS
ALLOWED_ORIGINS=https://propertyhub.com,https://www.propertyhub.com

# MFA
MFA_ISSUER=PropertyHub
MFA_ENFORCE_FOR_ADMINS=true
```

### Initialization in main.go

All security services are initialized in `@cmd/server/main.go`:

```go
// Line 52-55: Encryption Manager
encryptionManager, err := security.NewEncryptionManager(gormDB)

// Line 28: Audit Logger
auditLogger := security.NewAuditLogger(gormDB)

// Line 390-400: Security Headers
r.Use(func(c *gin.Context) {
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("X-Frame-Options", "DENY")
    c.Header("X-XSS-Protection", "1; mode=block")
    c.Header("Strict-Transport-Security", "max-age=31536000")
    c.Next()
})

// Line 403: CSRF Protection
r.Use(middleware.CSRFProtection())
```

---

## Monitoring & Alerting

### Real-time Security Dashboard

Monitor security metrics:

```go
// Get audit statistics
stats := auditLogger.GetAuditStatistics()
// Returns:
// - total_logs
// - successful_actions
// - failed_actions
// - success_rate
// - security_events
// - unresolved_security_events
// - high_risk_events
// - recent_logs_24h

// Get session statistics
sessionStats := sessionManager.GetSessionStatistics()
// Returns:
// - active_sessions
// - trusted_sessions
// - high_risk_sessions
// - trust_rate
// - suspicious_activities

// Get MFA statistics
mfaStats := totpManager.GetMFAStatistics()
// Returns:
// - total_users_with_mfa
// - enabled_users
// - adoption_rate
// - attempts_24h
// - success_rate_24h

// Get rate limit stats
rateLimitStats := securityMiddleware.GetRateLimitStats()
// Returns:
// - total_clients
// - blocked_clients
// - active_clients
```

### Security Event Queries

```go
// Get unresolved security events
filters := map[string]interface{}{
    "resolved": false,
    "severity": "high",
}
events, total, err := auditLogger.GetSecurityEvents(filters, 50, 0)

// Get audit logs with filtering
filters := map[string]interface{}{
    "user_id": userID,
    "action": "login",
    "start_date": time.Now().Add(-24 * time.Hour),
    "success": false,
}
logs, total, err := auditLogger.GetAuditLogs(filters, 50, 0)
```

### Automated Cleanup

```go
// Run periodically (e.g., daily cron job)
auditLogger.CleanupOldLogs(365) // Keep 1 year
encryptionManager.CleanupAuditLogs() // Keep 1 year
totpManager.CleanupExpiredAttempts() // Keep 90 days
sessionManager.CleanupExpiredSessions() // Keep 7 days
```

---

## Incident Response

### Immediate Response Actions

#### Suspected Compromise

```go
// 1. Invalidate all user sessions
sessionManager.InvalidateAllUserSessions(userID, "security_incident")

// 2. Disable user account (application-specific)
// userService.DisableAccount(userID, "security_incident")

// 3. Review recent activity
filters := map[string]interface{}{
    "user_id": userID,
    "start_date": time.Now().Add(-7 * 24 * time.Hour),
}
logs, _, _ := auditLogger.GetAuditLogs(filters, 1000, 0)

// 4. Check for data access
dataAccessLogs := auditLogger.GetDataAccessLogs(userID, 7) // Last 7 days

// 5. Force MFA re-setup
totpManager.DisableMFA(userID)
```

#### Brute Force Attack

```go
// 1. Add attacking IP to blacklist
securityMiddleware.AddToBlacklist(attackerIP)

// 2. Review all attempts from that IP
filters := map[string]interface{}{
    "ip_address": attackerIP,
    "start_date": time.Now().Add(-24 * time.Hour),
}
logs, _, _ := auditLogger.GetAuditLogs(filters, 1000, 0)

// 3. Check if any successful logins
filters["success"] = true
successfulLogins, _, _ := auditLogger.GetAuditLogs(filters, 100, 0)

// 4. If successful logins, invalidate those sessions
for _, log := range successfulLogins {
    if log.UserID != nil {
        sessionManager.InvalidateAllUserSessions(*log.UserID, "brute_force_incident")
    }
}
```

#### SQL Injection Attempt

```go
// 1. Review security events
filters := map[string]interface{}{
    "event_type": "sql_injection_attempt",
    "start_date": time.Now().Add(-1 * time.Hour),
}
events, _, _ := auditLogger.GetSecurityEvents(filters, 100, 0)

// 2. Add attacking IPs to blacklist
for _, event := range events {
    securityMiddleware.AddToBlacklist(event.IPAddress)
}

// 3. Verify database integrity
// Run manual SQL queries to check for unauthorized changes

// 4. Review audit logs for actual data access
// Check if any queries succeeded before detection
```

### Post-Incident

1. **Document the incident** - Create incident report with timeline
2. **Review logs** - Export relevant audit logs for analysis
3. **Update security rules** - Add new threat patterns if needed
4. **Notify stakeholders** - Follow compliance requirements (TREC, GDPR)
5. **Strengthen defenses** - Implement additional controls if needed

---

## Compliance

### TREC Compliance

PropertyHub implements TREC (Texas Real Estate Commission) compliance monitoring:

- Document retention requirements
- Privacy policy enforcement
- Consumer information protection
- Transaction logging
- License verification

See `@internal/security/trec_compliance.go` for implementation details.

### GDPR Compliance

- Right to access: Audit logs track all data access
- Right to erasure: Implement data deletion with audit trail
- Data portability: Export user data securely
- Breach notification: Real-time monitoring and alerting

### CCPA Compliance

- Consumer data rights
- Opt-out of data sale
- Data inventory and mapping
- Vendor management

---

## Security Best Practices

### For Developers

1. **Never disable security middleware** in production
2. **Always use parameterized queries** - avoid string concatenation
3. **Validate all user input** - use InputValidator for all forms
4. **Encrypt sensitive data** - use EncryptionManager for PII
5. **Log security events** - use AuditLogger for all sensitive operations
6. **Test with security tools** - run govulncheck and gosec regularly
7. **Review security logs** - monitor audit logs daily
8. **Keep dependencies updated** - run `go get -u` and test regularly

### For System Administrators

1. **Rotate encryption keys** annually using `encryptionManager.RotateKey()`
2. **Review blacklists** weekly and clear stale entries
3. **Monitor rate limits** - adjust if legitimate users are blocked
4. **Review security alerts** daily and respond to high/critical alerts
5. **Test backups** - ensure encrypted data can be restored
6. **Update IP whitelists** when infrastructure changes
7. **Enforce MFA** for all admin accounts
8. **Run security scans** - GitHub Actions runs weekly (see @.github/workflows/security.yml)

### For Security Teams

1. **Conduct penetration testing** quarterly
2. **Review audit logs** for anomalies
3. **Update threat patterns** based on new attack vectors
4. **Incident response drills** semi-annually
5. **Security training** for development team annually
6. **Compliance audits** per regulatory requirements
7. **Vendor security assessments** for third-party integrations

---

## Troubleshooting

### Users Getting Rate Limited

```go
// Check rate limit stats
stats := securityMiddleware.GetRateLimitStats()

// Clear rate limit data if needed (use with caution)
securityMiddleware.ClearRateLimitData()

// Or adjust rate limits for specific endpoints
newLimiter := middleware.NewEndpointRateLimiter(100, 500, 5*time.Minute)
```

### Encryption Key Issues

```bash
# Generate new encryption key (32 bytes = 256 bits)
openssl rand -base64 32

# Set in environment
export ENCRYPTION_KEY="<generated-key>"

# Verify encryption is working
curl http://localhost:8080/health/encryption
```

### Session Issues

```go
// Check session statistics
stats := sessionManager.GetSessionStatistics()

// Clear expired sessions manually
sessionManager.CleanupExpiredSessions()

// Review suspicious activities
var activities []security.SuspiciousActivity
db.Where("is_resolved = false").Order("created_at DESC").Find(&activities)
```

### High False Positive Rate

```go
// Adjust threat detection thresholds
// Edit threat patterns in realtime_monitoring.go

// Or whitelist specific patterns
// Add trusted IPs to whitelist
securityMiddleware.AddToWhitelist("203.0.113.10")
```

---

## Security Contacts

### Internal

- **Security Team:** security@propertyhub.com
- **On-Call Security:** See PagerDuty rotation
- **DevOps Team:** devops@propertyhub.com

### External

- **Vulnerability Reports:** security@propertyhub.com (PGP key available)
- **Bug Bounty Program:** https://propertyhub.com/security/bounty
- **Security Advisories:** https://propertyhub.com/security/advisories

---

## Change Log

### 2025-12-05
- Initial security architecture documentation
- Documented all 13 security services
- Added configuration and operational procedures
- Added incident response playbooks

---

**Document Version:** 1.0.0  
**Next Review:** 2026-03-05 (Quarterly review required)
