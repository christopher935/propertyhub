# PropertyHub Production Deployment Checklist

**Last Updated:** December 5, 2025  
**Version:** 1.0.0

Use this checklist to ensure all security and operational requirements are met before deploying PropertyHub to production.

## Table of Contents

- [Pre-Deployment Checklist](#pre-deployment-checklist)
- [Environment Variables](#environment-variables)
- [Database Configuration](#database-configuration)
- [Security Configuration](#security-configuration)
- [External Services](#external-services)
- [Infrastructure](#infrastructure)
- [Testing & Validation](#testing--validation)
- [Monitoring & Alerting](#monitoring--alerting)
- [Post-Deployment](#post-deployment)
- [Rollback Plan](#rollback-plan)

---

## Pre-Deployment Checklist

### Code Review & Testing

- [ ] All code changes reviewed and approved
- [ ] Unit tests passing (run `go test ./...`)
- [ ] Integration tests passing
- [ ] Security scans passing (govulncheck, gosec)
- [ ] No hardcoded secrets or credentials in code
- [ ] Dependencies up to date (`go get -u && go mod tidy`)
- [ ] `.gitignore` excludes sensitive files (.env, keys, etc.)

### Build & Compile

- [ ] Production build completed without errors
- [ ] Binary size reasonable (check for bloat)
- [ ] All templates validated and loading correctly
- [ ] Static assets minified and optimized
- [ ] Go version matches production (1.22.2+)

### Documentation

- [ ] README.md up to date
- [ ] SECURITY.md reviewed
- [ ] API documentation current
- [ ] Deployment procedures documented
- [ ] Runbook created for ops team

---

## Environment Variables

### ✅ Required Environment Variables

Copy this template to production `.env` file and fill in values:

```bash
# ============================================
# DATABASE CONFIGURATION (REQUIRED)
# ============================================
DATABASE_URL=postgres://username:password@host:5432/propertyhub?sslmode=require

# ============================================
# SECURITY - ENCRYPTION (REQUIRED)
# ============================================
# Generate with: openssl rand -base64 32
ENCRYPTION_KEY=<32-byte-base64-encoded-key>

# ============================================
# APPLICATION SETTINGS
# ============================================
PORT=8080
GIN_MODE=release
ENV=production

# ============================================
# AUTHENTICATION
# ============================================
SESSION_SECRET=<64-character-random-string>
SESSION_TIMEOUT=24h
JWT_SECRET=<64-character-random-string>
JWT_EXPIRATION=24h

# ============================================
# SMTP EMAIL CONFIGURATION (REQUIRED)
# ============================================
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=noreply@propertyhub.com
SMTP_PASSWORD=<smtp-password>
SMTP_FROM_ADDRESS=noreply@propertyhub.com
SMTP_FROM_NAME=PropertyHub

# ============================================
# REDIS (Optional but Recommended)
# ============================================
REDIS_URL=localhost:6379
REDIS_PASSWORD=<redis-password>
REDIS_DB=0

# ============================================
# EXTERNAL APIS
# ============================================
# FollowUp Boss Integration
FUB_API_KEY=<fub-api-key>
FUB_API_URL=https://api.followupboss.com/v1

# Property Data Scraping
SCRAPER_API_KEY=<scraper-api-key>

# ============================================
# SECURITY SETTINGS
# ============================================
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_REQUESTS_PER_HOUR=1000

SECURITY_MONITORING_ENABLED=true
SECURITY_ALERT_EMAIL=security@propertyhub.com

MFA_ISSUER=PropertyHub
MFA_ENFORCE_FOR_ADMINS=true

# ============================================
# CORS CONFIGURATION
# ============================================
ALLOWED_ORIGINS=https://propertyhub.com,https://www.propertyhub.com,https://admin.propertyhub.com

# ============================================
# TRUSTED PROXIES (if behind load balancer)
# ============================================
TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12

# ============================================
# LOGGING & MONITORING
# ============================================
LOG_LEVEL=info
LOG_FORMAT=json

# Sentry (Optional)
SENTRY_DSN=<sentry-dsn>
SENTRY_ENVIRONMENT=production

# ============================================
# FILE STORAGE
# ============================================
STORAGE_TYPE=s3
S3_BUCKET=propertyhub-production
S3_REGION=us-east-1
AWS_ACCESS_KEY_ID=<aws-access-key>
AWS_SECRET_ACCESS_KEY=<aws-secret-key>

# ============================================
# FEATURE FLAGS
# ============================================
ENABLE_PROPERTY_ALERTS=true
ENABLE_SAVED_PROPERTIES=true
ENABLE_AI_RECOMMENDATIONS=true
ENABLE_BEHAVIORAL_ANALYTICS=true
```

### Validation Steps

- [ ] All REQUIRED variables are set
- [ ] `ENCRYPTION_KEY` is 32 bytes (256 bits) base64-encoded
- [ ] `DATABASE_URL` includes `sslmode=require`
- [ ] Secrets are not duplicated from staging/dev
- [ ] SMTP credentials tested (send test email)
- [ ] Redis connection tested (if enabled)
- [ ] `ALLOWED_ORIGINS` matches production domains
- [ ] All passwords use strong, random values
- [ ] No variables contain `localhost` or `127.0.0.1`

---

## Database Configuration

### Pre-Deployment Database Tasks

- [ ] Database created with correct encoding (UTF-8)
- [ ] Database user created with appropriate permissions
- [ ] SSL/TLS enabled for database connections
- [ ] Connection pooling configured (max 25 connections)
- [ ] Backup schedule configured (daily minimum)
- [ ] Point-in-time recovery enabled
- [ ] Database monitoring enabled

### Database Migrations

```bash
# Review pending migrations
go run cmd/server/main.go --dry-run-migrations

# Run migrations
go run cmd/server/main.go --migrate

# Verify migration status
psql $DATABASE_URL -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;"
```

- [ ] All migrations reviewed and tested in staging
- [ ] Migrations run successfully in production
- [ ] Database indexes created
- [ ] Database statistics updated

### Database Security Tables

Ensure these security tables exist:

- [ ] `encryption_keys` - Encryption key tracking
- [ ] `audit_logs` - General audit trail
- [ ] `security_events` - Security-specific events
- [ ] `data_access_logs` - PII access tracking
- [ ] `admin_actions` - Administrative actions
- [ ] `sessions` - Active user sessions
- [ ] `device_info` - Device fingerprints
- [ ] `totp_secrets` - MFA secrets
- [ ] `backup_codes` - MFA backup codes
- [ ] `security_alerts` - Real-time alerts
- [ ] `security_metrics` - Security statistics
- [ ] `suspicious_activities` - Threat detection

```bash
# Verify tables exist
psql $DATABASE_URL -c "\dt security*; \dt *audit*; \dt *session*; \dt *totp*;"
```

---

## Security Configuration

### Encryption

- [ ] `ENCRYPTION_KEY` generated securely (NOT reused from other environments)
- [ ] Encryption key backed up to secure vault (1Password, AWS Secrets Manager, etc.)
- [ ] Encryption key documented in disaster recovery plan
- [ ] Key rotation schedule established (annually minimum)
- [ ] Encryption validation test passed

```bash
# Test encryption
curl https://your-domain.com/health/encryption
# Expected: {"status": "ok", "encryption": "operational"}
```

### SSL/TLS Certificates

- [ ] SSL certificate installed for primary domain
- [ ] SSL certificate installed for www subdomain
- [ ] SSL certificate installed for admin subdomain (if separate)
- [ ] Certificate auto-renewal configured (Let's Encrypt recommended)
- [ ] HSTS header configured (max-age=31536000)
- [ ] SSL Labs test passes with A+ rating (https://www.ssllabs.com/ssltest/)

### Security Headers

These are automatically set by PropertyHub middleware, but verify:

```bash
# Test security headers
curl -I https://your-domain.com

# Should include:
# X-Content-Type-Options: nosniff
# X-Frame-Options: DENY
# X-XSS-Protection: 1; mode=block
# Strict-Transport-Security: max-age=31536000
# Content-Security-Policy: ...
```

- [ ] All security headers present
- [ ] CSP policy tested and not blocking legitimate resources
- [ ] HSTS header includes `includeSubDomains` and `preload`

### Rate Limiting

- [ ] Rate limiting enabled (`RATE_LIMIT_ENABLED=true`)
- [ ] Rate limits appropriate for expected traffic
- [ ] Rate limit monitoring configured
- [ ] Alerts configured for excessive rate limiting

### Multi-Factor Authentication

- [ ] MFA enforced for admin accounts (`MFA_ENFORCE_FOR_ADMINS=true`)
- [ ] MFA issuer set correctly (`MFA_ISSUER=PropertyHub`)
- [ ] QR code generation tested
- [ ] Backup codes working correctly
- [ ] MFA recovery process documented

### IP Whitelisting/Blacklisting

If using IP restrictions:

- [ ] Admin panel IP whitelist configured (if applicable)
- [ ] VPN IP addresses whitelisted (if applicable)
- [ ] Office IP addresses whitelisted (if applicable)
- [ ] IP whitelist documented and maintained

### Firewall Rules

- [ ] Only ports 80 (HTTP), 443 (HTTPS) open to public
- [ ] PostgreSQL port (5432) restricted to application servers only
- [ ] Redis port (6379) restricted to application servers only
- [ ] SSH (22) restricted to specific IP addresses or VPN
- [ ] Application servers can't be accessed directly (behind load balancer)

---

## External Services

### FollowUp Boss Integration

- [ ] FUB API key configured (`FUB_API_KEY`)
- [ ] FUB API URL correct (`FUB_API_URL`)
- [ ] FUB connection tested
- [ ] FUB webhook endpoints configured
- [ ] FUB webhook secrets configured
- [ ] FUB rate limits reviewed and within bounds

### Property Data Scraping

- [ ] Scraper API key configured (`SCRAPER_API_KEY`)
- [ ] Scraper service tested
- [ ] Scraper rate limits reviewed
- [ ] HAR (Houston Association of Realtors) compliance verified

### Email Service (SMTP)

- [ ] SMTP credentials configured and tested
- [ ] SPF record configured in DNS
- [ ] DKIM configured and keys added to DNS
- [ ] DMARC policy configured
- [ ] Email templates tested
- [ ] Unsubscribe functionality working
- [ ] Email deliverability tested (not going to spam)

### File Storage (AWS S3)

- [ ] S3 bucket created with correct name
- [ ] S3 bucket region configured
- [ ] IAM user created for PropertyHub with minimal permissions
- [ ] Bucket policy configured (private, no public access)
- [ ] CORS configured if needed for direct uploads
- [ ] Lifecycle policies configured (for old file cleanup)
- [ ] Versioning enabled
- [ ] Bucket encryption enabled

### Redis (Optional)

- [ ] Redis server deployed and accessible
- [ ] Redis password configured
- [ ] Redis persistence enabled (AOF or RDB)
- [ ] Redis maxmemory policy configured (allkeys-lru recommended)
- [ ] Redis connection tested from application servers

---

## Infrastructure

### Server Requirements

**Minimum Production Specs:**
- CPU: 4 cores (8 cores recommended)
- RAM: 8 GB (16 GB recommended)
- Disk: 50 GB SSD (100 GB recommended)
- Network: 1 Gbps

- [ ] Server meets minimum specifications
- [ ] Operating system updated (Ubuntu 22.04 LTS recommended)
- [ ] Required packages installed (postgresql-client, curl, etc.)
- [ ] Time synchronization configured (NTP)
- [ ] Timezone set to UTC

### Load Balancer

- [ ] Load balancer deployed (AWS ALB, Nginx, Caddy, etc.)
- [ ] Health check configured (`/health` endpoint)
- [ ] SSL termination configured at load balancer
- [ ] Sticky sessions configured (if needed)
- [ ] Request timeout configured (60 seconds recommended)
- [ ] X-Forwarded-For header preserved
- [ ] X-Real-IP header set

### Container/Orchestration (if using Docker/Kubernetes)

- [ ] Docker image built and pushed to registry
- [ ] Image scanned for vulnerabilities
- [ ] Kubernetes manifests reviewed
- [ ] Resource limits set (CPU, memory)
- [ ] Liveness probe configured
- [ ] Readiness probe configured
- [ ] Horizontal pod autoscaling configured
- [ ] Pod security policies configured

### DNS Configuration

- [ ] A record points to load balancer/server
- [ ] AAAA record configured (IPv6) if applicable
- [ ] www CNAME points to primary domain
- [ ] admin subdomain configured (if separate)
- [ ] CAA records configured for SSL validation
- [ ] TTL set appropriately (300 seconds recommended for initial deployment)

### Backup Strategy

- [ ] Database backups configured (automated daily minimum)
- [ ] Encryption keys backed up to secure vault
- [ ] Application configuration backed up
- [ ] File storage backups configured (S3 versioning + replication)
- [ ] Backup restoration tested successfully
- [ ] Backup retention policy defined (30 days minimum)
- [ ] Off-site backup copy maintained

---

## Testing & Validation

### Functional Testing

- [ ] Homepage loads correctly
- [ ] Admin login working
- [ ] Consumer registration working
- [ ] Property search working
- [ ] Booking system functional
- [ ] Email notifications sending
- [ ] File uploads working
- [ ] MFA setup and verification working

### Security Testing

- [ ] Security scan passed (govulncheck)

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

- [ ] Static analysis passed (gosec)

```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

- [ ] SQL injection tests passed (manual testing with sqlmap or similar)
- [ ] XSS tests passed (manual testing)
- [ ] CSRF protection tested
- [ ] Rate limiting tested and working
- [ ] Session management tested
- [ ] Encryption tested (sensitive data encrypted at rest)

### Performance Testing

- [ ] Load testing completed (Apache Bench, k6, or similar)
- [ ] Response times acceptable (< 200ms for most endpoints)
- [ ] Database query performance acceptable
- [ ] No memory leaks detected
- [ ] CPU usage reasonable under load
- [ ] Autoscaling triggers tested (if using autoscaling)

### Compliance Testing

- [ ] TREC compliance verified
- [ ] GDPR requirements met (if applicable)
- [ ] CCPA requirements met (if applicable)
- [ ] Privacy policy deployed and linked
- [ ] Terms of service deployed and linked
- [ ] Cookie consent banner working (if applicable)

---

## Monitoring & Alerting

### Application Monitoring

- [ ] Application health check endpoint working (`/health`)
- [ ] Metrics endpoint configured (if using Prometheus)
- [ ] Application logs aggregated (CloudWatch, Datadog, etc.)
- [ ] Error tracking configured (Sentry, Rollbar, etc.)
- [ ] Performance monitoring configured (APM)

### Infrastructure Monitoring

- [ ] Server CPU monitoring
- [ ] Server memory monitoring
- [ ] Disk space monitoring
- [ ] Network monitoring
- [ ] Database connection pool monitoring
- [ ] Load balancer health monitoring

### Security Monitoring

- [ ] Security event logging enabled
- [ ] Failed login attempt monitoring
- [ ] Rate limit violation alerts
- [ ] Unusual access pattern detection
- [ ] Admin action logging
- [ ] Data access audit logging

### Alerting

Configure alerts for:

- [ ] Application errors (> 10 per minute)
- [ ] High response times (> 1 second)
- [ ] Database connection failures
- [ ] Disk space low (< 20% free)
- [ ] Memory usage high (> 90%)
- [ ] SSL certificate expiring (< 30 days)
- [ ] Failed login attempts (> 5 in 5 minutes from same IP)
- [ ] High-severity security events
- [ ] Backup failures

### Alert Destinations

- [ ] PagerDuty or similar on-call rotation configured
- [ ] Slack channel for alerts created and configured
- [ ] Email alerts configured for critical issues
- [ ] SMS alerts configured for critical issues (optional)

---

## Post-Deployment

### Immediate Post-Deployment (Day 1)

- [ ] Verify application is accessible from all domains
- [ ] Test critical user flows (registration, login, booking)
- [ ] Monitor error rates for first hour
- [ ] Monitor response times
- [ ] Check security logs for anomalies
- [ ] Verify all cron jobs/scheduled tasks running
- [ ] Test email functionality
- [ ] Verify external integrations working (FUB, etc.)

### First Week

- [ ] Daily review of error logs
- [ ] Daily review of security alerts
- [ ] Monitor performance metrics
- [ ] Check backup success
- [ ] User feedback collection
- [ ] Performance tuning as needed

### First Month

- [ ] Weekly security log reviews
- [ ] Performance optimization
- [ ] Capacity planning review
- [ ] User feedback analysis
- [ ] Security audit (internal or external)

---

## Rollback Plan

### Pre-Rollback Checklist

- [ ] Rollback decision documented with reason
- [ ] Stakeholders notified
- [ ] Database backup verified
- [ ] Previous version artifacts available

### Rollback Steps

1. **Database Rollback (if migrations were run)**

```bash
# Revert last migration
go run cmd/server/main.go --rollback-migration

# Or restore from backup
pg_restore --clean --if-exists -d propertyhub < backup.sql
```

2. **Application Rollback**

```bash
# If using Docker
docker pull propertyhub:previous-tag
docker-compose up -d

# If using Kubernetes
kubectl rollout undo deployment/propertyhub

# If using traditional deployment
systemctl stop propertyhub
cp /opt/propertyhub/previous/server /opt/propertyhub/current/
systemctl start propertyhub
```

3. **Verification**

- [ ] Application starts successfully
- [ ] Health check passes
- [ ] Critical flows working
- [ ] No errors in logs
- [ ] Users can access the system

4. **Post-Rollback**

- [ ] Incident report created
- [ ] Root cause analysis scheduled
- [ ] Fix planned for next deployment

---

## Sign-off

### Pre-Deployment Sign-off

- [ ] **Developer:** Code complete and tested
- [ ] **Security Team:** Security requirements met
- [ ] **QA Team:** All tests passed
- [ ] **DevOps:** Infrastructure ready
- [ ] **Product Manager:** Feature complete and ready
- [ ] **CTO/Technical Lead:** Final approval

### Deployment Details

**Deployment Date:** _______________  
**Deployed By:** _______________  
**Deployment Time:** _______________  
**Version/Git SHA:** _______________  
**Rollback Plan Tested:** ☐ Yes ☐ No

### Post-Deployment Sign-off

- [ ] **DevOps:** Deployment successful, monitoring active
- [ ] **QA:** Smoke tests passed in production
- [ ] **Security:** Security monitoring active, no alerts
- [ ] **Product Manager:** User-facing features verified

---

## Appendix

### Useful Commands

```bash
# Check application health
curl https://your-domain.com/health

# Check security encryption
curl https://your-domain.com/health/encryption

# View application logs (systemd)
journalctl -u propertyhub -f

# View database connections
psql $DATABASE_URL -c "SELECT count(*) FROM pg_stat_activity;"

# Test database connection
psql $DATABASE_URL -c "SELECT version();"

# Test Redis connection
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping

# Generate encryption key
openssl rand -base64 32

# Check SSL certificate expiration
echo | openssl s_client -servername your-domain.com -connect your-domain.com:443 2>/dev/null | openssl x509 -noout -dates

# Run security scan
govulncheck ./...
gosec ./...

# Load test
ab -n 1000 -c 10 https://your-domain.com/
```

### Emergency Contacts

- **On-Call Engineer:** See PagerDuty
- **Database Admin:** dba@propertyhub.com
- **Security Team:** security@propertyhub.com
- **DevOps Team:** devops@propertyhub.com
- **CTO/Tech Lead:** cto@propertyhub.com

### Runbook Links

- Security Incident Response: `docs/security-incident-response.md`
- Database Maintenance: `docs/database-maintenance.md`
- Backup and Restore: `docs/backup-restore.md`
- Scaling Guide: `docs/scaling-guide.md`

---

**Document Version:** 1.0.0  
**Last Review:** December 5, 2025  
**Next Review:** Quarterly or before next major deployment
