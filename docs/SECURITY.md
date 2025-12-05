# PropertyHub Security Documentation

## Overview

PropertyHub implements enterprise-grade security measures to protect user data, prevent unauthorized access, and ensure compliance with industry regulations including TREC (Texas Real Estate Commission) requirements.

## Table of Contents

1. [Security Architecture](#security-architecture)
2. [Authentication & Authorization](#authentication--authorization)
3. [Data Protection](#data-protection)
4. [API Security](#api-security)
5. [Network Security](#network-security)
6. [Audit & Compliance](#audit--compliance)
7. [Incident Response](#incident-response)
8. [Security Best Practices](#security-best-practices)

## Security Architecture

### Defense in Depth

PropertyHub implements multiple layers of security:

```
┌─────────────────────────────────────────┐
│   Application Layer (Gin Framework)     │
│   - Input validation                    │
│   - Output encoding                     │
│   - CSRF protection                     │
└─────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────┐
│   Security Middleware                   │
│   - Rate limiting                       │
│   - Security headers                    │
│   - SQL injection prevention            │
│   - XSS protection                      │
└─────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────┐
│   Authentication & Authorization        │
│   - JWT tokens                          │
│   - Session management                  │
│   - MFA (TOTP)                          │
│   - Brute force protection              │
└─────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────┐
│   Data Layer                            │
│   - Field encryption (AES-256)          │
│   - Document encryption                 │
│   - Encrypted backups                   │
└─────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────┐
│   Infrastructure                        │
│   - TLS 1.2+                           │
│   - Firewall rules                      │
│   - Network isolation                   │
└─────────────────────────────────────────┘
```

### Security Components

| Component | Location | Purpose |
|-----------|----------|---------|
| Field Encryption | `internal/security/field_encryption.go` | Encrypts PII at rest |
| Document Encryption | `internal/security/document_encryption.go` | Encrypts uploaded documents |
| Password Hashing | `internal/security/password.go` | bcrypt password hashing |
| Session Management | `internal/security/session_management.go` | Secure session handling |
| SQL Protection | `internal/security/sql_protection.go` | SQL injection prevention |
| XSS Protection | `internal/security/xss_protection.go` | Cross-site scripting prevention |
| Input Validation | `internal/security/input_validation.go` | Input sanitization |
| MFA | `internal/security/totp_mfa.go` | Two-factor authentication |
| Audit Logging | `internal/security/audit_logging.go` | Comprehensive audit trails |
| TREC Compliance | `internal/security/trec_compliance.go` | Real estate compliance |
| Brute Force Protection | `internal/security/brute_force_protection.go` | Login attempt limiting |

## Authentication & Authorization

### JWT Tokens

PropertyHub uses JSON Web Tokens (JWT) for authentication:

**Token Configuration:**
- **Algorithm**: HS256 (HMAC with SHA-256)
- **Access Token Expiry**: 24 hours (configurable)
- **Refresh Token Expiry**: 7 days (configurable)
- **Secret**: Minimum 32 characters, stored in environment variables

**Token Structure:**
```json
{
  "user_id": "uuid",
  "username": "string",
  "email": "string",
  "role": "admin|agent|consumer",
  "exp": 1234567890
}
```

### Multi-Factor Authentication (MFA)

**Implementation:**
- TOTP (Time-based One-Time Password)
- Compatible with Google Authenticator, Authy, 1Password
- 30-second time window
- 6-digit codes
- QR code generation for easy setup

**Enrollment Flow:**
1. User enables MFA in account settings
2. System generates TOTP secret
3. QR code displayed for scanning
4. User enters verification code to confirm
5. Backup codes generated (10 single-use codes)

### Session Management

**Session Security:**
- Session tokens stored in Redis (if available) or database
- Automatic session expiration (default: 60 minutes)
- Session invalidation on password change
- Session refresh on activity
- Secure, httpOnly cookies (when using cookies)

### Brute Force Protection

**Login Protection:**
- **Max Attempts**: 5 failed attempts
- **Lockout Duration**: 15 minutes
- **Detection Window**: 15 minutes
- **Storage**: Redis (preferred) or database
- **IP Tracking**: Monitor suspicious IPs

**Implementation:**
```go
// Check login attempt
allowed, remaining, retryAfter, err := bfp.CheckLoginAttempt(username, ipAddress)
if !allowed {
    return error("Account locked. Retry after: " + retryAfter)
}

// Record failed attempt
bfp.RecordFailedAttempt(username, ipAddress, userAgent)

// Record successful login (clears attempts)
bfp.RecordSuccessfulLogin(username)
```

## Data Protection

### Encryption at Rest

**Field-Level Encryption:**
- **Algorithm**: AES-256-GCM
- **Key Management**: Environment variable or key management service
- **Encrypted Fields**:
  - Email addresses
  - Phone numbers
  - Physical addresses
  - SSN (if collected)
  - Bank account information
  - Credit card information

**Implementation:**
```go
// Encrypt sensitive field
encrypted, err := encMgr.EncryptField(plaintext)

// Decrypt when needed
plaintext, err := encMgr.DecryptField(encrypted)
```

### Document Encryption

**Document Security:**
- All uploaded documents encrypted
- Unique encryption key per document
- Master key stored securely
- Automatic encryption on upload
- Transparent decryption on download

### Password Security

**Password Requirements:**
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

**Password Storage:**
- bcrypt hashing with cost factor 12
- Never stored in plaintext
- Never logged or transmitted after initial submission

### Sensitive Data Redaction

**Audit Log Protection:**
Sensitive data automatically redacted from logs:

**Redacted Fields:**
- password, password_hash, new_password, old_password
- ssn, social_security
- credit_card, card_number, cvv, cvc
- bank_account, routing_number, account_number
- api_key, api_secret, secret, secret_key
- access_token, refresh_token, token
- authorization, bearer, jwt
- private_key, encryption_key
- session_id, cookie

**Example:**
```go
// Before redaction
data := map[string]interface{}{
    "username": "john",
    "password": "secret123",
    "email": "john@example.com",
}

// After redaction
sanitized := auditLogger.SanitizeForLog(data)
// Result: {"username": "john", "password": "[REDACTED]", "email": "john@example.com"}
```

## API Security

### Rate Limiting

**Tiered Rate Limits:**

| Tier | Requests/Min | Requests/Hour | Block Duration | Use Case |
|------|--------------|---------------|----------------|----------|
| Public | 100 | 1000 | 5 min | Anonymous users, property listings |
| Auth | 300 | 3000 | 5 min | Authenticated users |
| Sensitive | 5 | 20 | 30 min | Login, password reset, MFA |
| Admin | 50 | 500 | 15 min | Admin dashboard operations |
| API | 60 | 600 | 10 min | External API integrations |

**Implementation:**
```go
// Apply rate limiting to endpoint
r.Use(middleware.NewEndpointRateLimiterWithTier(middleware.TierSensitive).RateLimit())
```

### CORS Configuration

**Production CORS Settings:**
```go
AllowOrigins: []string{
    "https://yourdomain.com",
    "https://www.yourdomain.com",
}
AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token"}
AllowCredentials: true
MaxAge: 12 * time.Hour
```

### CSRF Protection

**Implementation:**
- CSRF tokens for all state-changing requests
- Double-submit cookie pattern
- Token validation on POST, PUT, DELETE, PATCH
- Automatic token refresh
- SameSite cookie attribute

### Security Headers

**HTTP Security Headers:**
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' ...
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

**CSP Configuration:**
- Restrict script sources to trusted CDNs
- Prevent inline script execution (where possible)
- Restrict frame ancestors
- Enable form action restrictions

## Network Security

### TLS/SSL Configuration

**Requirements:**
- TLS 1.2 minimum (TLS 1.3 preferred)
- Strong cipher suites only
- HSTS enabled with preload
- Certificate from trusted CA
- Automatic certificate renewal

**Cipher Suites (Recommended):**
```
TLS_AES_128_GCM_SHA256
TLS_AES_256_GCM_SHA384
TLS_CHACHA20_POLY1305_SHA256
TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
```

### Firewall Rules

**Required Open Ports:**
- 443 (HTTPS) - Public
- 80 (HTTP) - Public (redirect to 443)
- 22 (SSH) - Restricted to admin IPs only
- 5432 (PostgreSQL) - Internal network only
- 6379 (Redis) - Internal network only

**Blocked by Default:**
- All other incoming connections
- All outgoing connections except to trusted services

## Audit & Compliance

### Audit Logging

**Logged Events:**
- User authentication (login, logout, MFA)
- Data access (view, create, update, delete)
- Administrative actions
- Security events (failed logins, suspicious activity)
- System changes (configuration, permissions)

**Audit Log Fields:**
```go
type AuditLog struct {
    Timestamp    time.Time
    UserID       *uint
    SessionID    string
    IPAddress    string
    UserAgent    string
    Action       string
    Resource     string
    ResourceID   *uint
    Success      bool
    ErrorMessage string
    Severity     string  // low, medium, high, critical
    Category     string  // auth, data, admin, security, system
}
```

### TREC Compliance

**Requirements:**
- Display TREC license number on all pages
- Maintain audit logs for all transactions
- Secure storage of client information
- Proper disclosure of relationships
- Record retention: 4 years minimum

**Implementation:**
- Automated TREC license display
- Comprehensive audit logging
- Encrypted PII storage
- Document retention policies

### Data Retention

**Retention Policies:**
- Audit logs: 365 days (configurable)
- Security events: 730 days
- User data: Retained until account deletion
- Backup data: 30 days
- Deleted data: Permanently removed after 90 days

## Incident Response

### Security Incident Classification

| Severity | Definition | Response Time | Examples |
|----------|------------|---------------|----------|
| Critical | Data breach, system compromise | Immediate | SQL injection, unauthorized admin access |
| High | Security control failure | 1 hour | XSS vulnerability, authentication bypass |
| Medium | Potential security risk | 4 hours | Excessive failed logins, suspicious patterns |
| Low | Security anomaly | 24 hours | Unusual user behavior, configuration drift |

### Incident Response Process

1. **Detection & Identification**
   - Monitor security alerts
   - Analyze audit logs
   - Identify scope and impact

2. **Containment**
   - Isolate affected systems
   - Block malicious IPs
   - Revoke compromised credentials
   - Enable temporary restrictions

3. **Eradication**
   - Remove malicious code/access
   - Patch vulnerabilities
   - Update security rules
   - Rotate secrets/keys

4. **Recovery**
   - Restore from clean backups
   - Verify system integrity
   - Re-enable services
   - Monitor for recurrence

5. **Post-Incident Review**
   - Document incident details
   - Analyze root cause
   - Update security measures
   - Train team on lessons learned

### Emergency Contacts

**Security Incidents:**
1. Security Lead: [Contact Info]
2. System Administrator: [Contact Info]
3. Database Administrator: [Contact Info]
4. Legal Counsel: [Contact Info]

## Security Best Practices

### For Developers

1. **Never commit secrets to version control**
   - Use environment variables
   - Use `.env.example` for templates
   - Add sensitive files to `.gitignore`

2. **Always validate and sanitize input**
   - Validate on both client and server
   - Use parameterized queries
   - Escape output appropriately

3. **Follow least privilege principle**
   - Grant minimum necessary permissions
   - Use specific database users
   - Implement role-based access control

4. **Keep dependencies updated**
   - Run `go mod tidy` regularly
   - Monitor for security advisories
   - Use `govulncheck` in CI/CD

5. **Log security events, not sensitive data**
   - Use sanitization functions
   - Never log passwords or tokens
   - Redact PII in logs

### For Administrators

1. **Enable MFA for all admin accounts**
2. **Use strong, unique passwords**
3. **Regularly review audit logs**
4. **Keep software updated and patched**
5. **Monitor security alerts and metrics**
6. **Perform regular backups and test restoration**
7. **Implement IP whitelisting for admin access**
8. **Review and update firewall rules quarterly**

### For Users

1. **Use strong, unique passwords**
2. **Enable MFA when available**
3. **Be cautious of phishing attempts**
4. **Report suspicious activity**
5. **Keep contact information updated**
6. **Review account activity regularly**

## Security Testing

### Automated Testing

**Security Test Suite:**
```bash
# Run security tests
go test ./internal/security/... -v

# Run with coverage
go test ./internal/security/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Tests Include:**
- SQL injection detection
- XSS prevention
- Password strength validation
- Sensitive data redaction
- Brute force protection
- Encryption/decryption
- Audit logging

### Penetration Testing

**Recommended Schedule:**
- Annual third-party penetration test
- Quarterly internal security assessment
- Ad-hoc testing after major changes

**Test Scope:**
- Authentication and authorization
- Session management
- Input validation
- API security
- Data protection
- Infrastructure security

## Vulnerability Disclosure

### Reporting Security Issues

If you discover a security vulnerability, please report it to:

**Email**: security@yourdomain.com

**Please include:**
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if available)

**Do NOT:**
- Publicly disclose the vulnerability before we've addressed it
- Exploit the vulnerability beyond what's necessary to demonstrate it
- Access or modify user data without permission

### Response Process

1. **Acknowledgment**: Within 24 hours
2. **Initial Assessment**: Within 72 hours
3. **Fix Development**: Based on severity
4. **Deployment**: Coordinated with reporter
5. **Public Disclosure**: After fix is deployed (30-90 days)

## Compliance & Certifications

### Current Compliance

- ✓ TREC (Texas Real Estate Commission)
- ✓ OWASP Top 10 protections
- ✓ CIS Security Benchmarks

### Planned Certifications

- SOC 2 Type II
- ISO 27001
- GDPR (if serving EU users)

## Security Updates

This security documentation is reviewed and updated quarterly. Last update: December 5, 2024

For questions or concerns, contact: security@yourdomain.com

---

**Version**: 1.0  
**Last Updated**: December 5, 2024  
**Next Review**: March 5, 2025
