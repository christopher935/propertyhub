# Production Deployment Checklist

## Pre-Deployment Checklist

### 1. Environment Variables ✓

Verify all required environment variables are set:

#### Database Configuration
- [ ] `DATABASE_URL` - Production PostgreSQL connection string
- [ ] `DATABASE_MAX_CONNS` - Maximum database connections (recommended: 25)
- [ ] `DATABASE_TIMEOUT_SECONDS` - Query timeout (recommended: 30)

#### Security Configuration
- [ ] `JWT_SECRET` - Strong random secret (min 32 characters)
- [ ] `ENCRYPTION_KEY` - 32-byte AES encryption key
- [ ] `SESSION_TIMEOUT_MINUTES` - Session duration (recommended: 60)
- [ ] `MFA_REQUIRED` - Enable MFA (true/false)
- [ ] `RATE_LIMIT_REQUESTS_PER_MINUTE` - Global rate limit (recommended: 100)

#### Redis Configuration
- [ ] `REDIS_URL` - Production Redis instance
- [ ] `REDIS_PASSWORD` - Redis authentication password
- [ ] `REDIS_DB` - Redis database number (default: 0)

#### Email Configuration
- [ ] `SMTP_HOST` - SMTP server hostname
- [ ] `SMTP_PORT` - SMTP server port (587 for TLS)
- [ ] `SMTP_USERNAME` - SMTP authentication username
- [ ] `SMTP_PASSWORD` - SMTP authentication password

#### External Services
- [ ] `FUB_API_KEY` - Follow Up Boss API key
- [ ] `FUB_API_URL` - Follow Up Boss API URL
- [ ] `HAR_API_KEY` - HAR API key (if using)
- [ ] `HAR_API_URL` - HAR API URL
- [ ] `SCRAPER_API_KEY` - Web scraping API key (if using)

#### Business Information
- [ ] `BUSINESS_NAME` - Your business name
- [ ] `BUSINESS_PHONE` - Contact phone number
- [ ] `BUSINESS_EMAIL` - Contact email
- [ ] `BUSINESS_ADDRESS` - Business address
- [ ] `TREC_LICENSE` - Texas Real Estate Commission license number

### 2. Security Configuration ✓

#### Application Security
- [ ] GIN_MODE set to `release`
- [ ] HTTPS enforced (HSTS header enabled)
- [ ] Security headers configured (CSP, X-Frame-Options, etc.)
- [ ] CSRF protection enabled globally
- [ ] CORS origins restricted to production domains only
- [ ] Rate limiting enabled on all endpoints
- [ ] Brute force protection active on login endpoints

#### Secrets Management
- [ ] No secrets in source code
- [ ] All secrets in environment variables or secrets manager
- [ ] API keys rotated from development values
- [ ] Database credentials unique to production
- [ ] JWT secret is strong and unique

#### Access Control
- [ ] Admin accounts created with strong passwords
- [ ] MFA enabled for admin accounts
- [ ] Default/demo accounts removed
- [ ] Database user has appropriate permissions (no unnecessary privileges)

### 3. Infrastructure ✓

#### Server Configuration
- [ ] Firewall rules configured (only necessary ports open)
- [ ] SSH key-based authentication enabled
- [ ] Root login disabled
- [ ] Automatic security updates enabled
- [ ] Server timezone set to UTC

#### SSL/TLS
- [ ] SSL certificate installed and valid
- [ ] Certificate auto-renewal configured
- [ ] TLS 1.2+ enforced (TLS 1.0/1.1 disabled)
- [ ] Strong cipher suites configured

#### Database
- [ ] PostgreSQL production instance running
- [ ] Database backups configured
- [ ] Backup encryption enabled
- [ ] Point-in-time recovery enabled
- [ ] Connection pooling configured
- [ ] Slow query logging enabled

#### Redis
- [ ] Redis production instance running
- [ ] Redis password authentication enabled
- [ ] Redis persistence configured (AOF or RDB)
- [ ] Redis maxmemory policy set
- [ ] Redis backups configured

### 4. Monitoring & Logging ✓

#### Application Monitoring
- [ ] Error logging to external service (Sentry, Rollbar, etc.)
- [ ] Performance monitoring enabled (APM)
- [ ] Uptime monitoring configured (PingDom, UptimeRobot, etc.)
- [ ] Log aggregation configured (CloudWatch, Datadog, etc.)

#### Security Monitoring
- [ ] Security event monitoring active
- [ ] Failed login attempt monitoring
- [ ] Brute force detection alerts
- [ ] Suspicious activity alerts
- [ ] Audit log retention configured (365 days minimum)

#### Alerting
- [ ] Email alerts for critical errors
- [ ] Slack/Discord/Teams webhook for incidents
- [ ] Database connection failure alerts
- [ ] Redis connection failure alerts
- [ ] Disk space alerts (warn at 80%, critical at 90%)
- [ ] Memory usage alerts

### 5. Backup & Recovery ✓

#### Database Backups
- [ ] Automated daily backups scheduled
- [ ] Backup retention policy defined (30+ days)
- [ ] Backups stored in separate location/region
- [ ] Backup encryption verified
- [ ] Backup restoration tested successfully

#### Application Backups
- [ ] Code repository backed up (GitHub/GitLab)
- [ ] Environment configuration backed up
- [ ] SSL certificates backed up
- [ ] Documentation backed up

#### Disaster Recovery
- [ ] Recovery Time Objective (RTO) defined
- [ ] Recovery Point Objective (RPO) defined
- [ ] Disaster recovery plan documented
- [ ] DR plan tested within last 6 months

### 6. Performance Optimization ✓

#### Database Optimization
- [ ] Database indexes created for frequent queries
- [ ] Query performance analyzed and optimized
- [ ] Connection pooling configured
- [ ] Database statistics updated

#### Application Optimization
- [ ] Static assets minified
- [ ] Gzip compression enabled
- [ ] CDN configured for static assets (optional)
- [ ] Image optimization enabled
- [ ] Cache headers configured appropriately

#### Load Testing
- [ ] Load tests performed at expected peak traffic
- [ ] Identified bottlenecks resolved
- [ ] Auto-scaling configured (if applicable)

### 7. Compliance & Legal ✓

#### TREC Compliance
- [ ] TREC license displayed on website
- [ ] Required disclosures present
- [ ] Audit logging meets compliance requirements
- [ ] Data retention policy complies with regulations

#### Data Protection
- [ ] Privacy policy published
- [ ] Terms of service published
- [ ] Cookie consent implemented (if applicable)
- [ ] GDPR compliance verified (if serving EU users)
- [ ] Data encryption at rest verified
- [ ] Data encryption in transit verified

### 8. Testing ✓

#### Security Testing
- [ ] Penetration testing completed
- [ ] Vulnerability scanning completed
- [ ] SQL injection tests passed
- [ ] XSS injection tests passed
- [ ] CSRF protection verified
- [ ] Authentication bypass tests passed

#### Functional Testing
- [ ] All critical user flows tested
- [ ] Payment processing tested (if applicable)
- [ ] Email delivery tested
- [ ] Form submissions tested
- [ ] File uploads tested
- [ ] Mobile responsiveness verified

#### Integration Testing
- [ ] Follow Up Boss integration tested
- [ ] HAR API integration tested (if using)
- [ ] Email service integration tested
- [ ] SMS service integration tested (if using)

### 9. Documentation ✓

- [ ] API documentation complete
- [ ] Deployment procedures documented
- [ ] Rollback procedures documented
- [ ] Incident response plan documented
- [ ] Admin user guide created
- [ ] End-user documentation available

### 10. Go-Live Preparation ✓

#### Final Checks
- [ ] All tests passing in production environment
- [ ] No debug/verbose logging in production
- [ ] No test data in production database
- [ ] Monitoring dashboards configured
- [ ] Team notified of deployment schedule
- [ ] Support team briefed on new features

#### Post-Deployment
- [ ] Smoke tests completed successfully
- [ ] Critical paths verified working
- [ ] Performance metrics within acceptable ranges
- [ ] No critical errors in logs
- [ ] Monitoring alerts functioning
- [ ] Team available for immediate support

## Security Incident Response

### Immediate Actions
1. **Detect**: Monitor security alerts and anomalies
2. **Contain**: Isolate affected systems
3. **Investigate**: Analyze logs and audit trails
4. **Remediate**: Apply patches, rotate credentials
5. **Document**: Record incident details and response
6. **Review**: Post-incident analysis and improvements

### Contact Information
- **Security Lead**: [Name] - [Email] - [Phone]
- **System Admin**: [Name] - [Email] - [Phone]
- **Database Admin**: [Name] - [Email] - [Phone]

## Maintenance Schedule

### Daily
- Monitor error rates and response times
- Review security event logs
- Check backup completion status

### Weekly
- Review audit logs for suspicious activity
- Analyze performance metrics
- Update dependency vulnerabilities

### Monthly
- Review and rotate API keys
- Security patch updates
- Database performance tuning
- Backup restoration test

### Quarterly
- Security audit
- Disaster recovery drill
- Performance load testing
- Compliance review

### Annually
- Penetration testing
- SSL certificate renewal
- Security policy review
- Incident response plan review

## Emergency Contacts

### Critical Services
- **Hosting Provider**: [Provider Name] - [Support URL] - [Phone]
- **Database Provider**: [Provider Name] - [Support URL] - [Phone]
- **CDN Provider**: [Provider Name] - [Support URL] - [Phone]
- **Email Service**: [Provider Name] - [Support URL] - [Phone]

### Escalation Path
1. **Level 1**: On-call engineer
2. **Level 2**: Senior engineer / Team lead
3. **Level 3**: CTO / Engineering director
4. **Level 4**: CEO / Legal counsel (security incidents)

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2024-12-05 | Initial production checklist | Security Team |

---

**Note**: This checklist should be reviewed and updated quarterly to reflect changes in infrastructure, security best practices, and compliance requirements.
