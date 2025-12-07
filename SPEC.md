# PropertyHub Complete Specification

**Version:** 1.0  
**Last Updated:** December 7, 2025  
**Purpose:** Landlords of Texas complete property management and lead intelligence platform

---

## Executive Summary

PropertyHub is an AI-powered property management platform that automates the entire rental lifecycle from lead generation through lease signing. The system is designed to handle 13,000+ leads at scale with minimal manual intervention, using behavioral intelligence and automated workflows to maximize conversions and prevent opportunities from slipping through the cracks.

**Core Value Proposition:**
- **Calendly for Real Estate** - Seamless showing booking system
- **FUB Integration** - Bidirectional sync with Follow Up Boss CRM
- **AI Decision Engine** - Automated opportunity detection and campaign execution
- **Zero Manual Work** - Automation handles routine tasks, humans intervene only where needed

---

## System Architecture

### Three-Tier Intelligence System

#### 1. Data Collection Layer
- **Behavioral Event Tracking** - Every user action tracked (views, saves, inquiries, applications)
- **Property Analytics** - Performance metrics per property (views, showings, applications, conversions)
- **Lead Scoring** - Real-time behavioral scores (0-100) with urgency/engagement/financial components
- **Session Tracking** - Anonymous and authenticated user journeys

#### 2. Intelligence Layer
- **Spiderweb AI Orchestrator** - Master coordinator of all AI modules
- **Relationship Intelligence Engine** - Cross-entity pattern detection
- **Funnel Analytics Service** - Conversion tracking and bottleneck identification
- **Property Matching Service** - Lead-to-property intelligent matching
- **Behavioral Scoring Engine** - Composite scoring with time decay
- **Insight Generator** - AI-generated recommendations with HTML formatting

#### 3. Automation Layer
- **Event Campaign Orchestrator** - Automated multi-channel campaigns
- **Abandonment Recovery** - Multi-step recovery sequences
- **Email/SMS Automation** - Triggered communications
- **Job Scheduler** - Cron-like scheduled tasks
- **FUB Batch Service** - Rate-limited API operations

---

## Core Features

### 1. Consumer Experience

#### Property Discovery
- **Property Browsing** - Grid view with lazy-loaded images
- **Advanced Search** - Address, neighborhood, ZIP code search
- **Smart Filters** - Price, bedrooms, bathrooms, location, property type
- **Quick View Modal** - Property preview without page navigation
- **AI Recommendations** - Behavioral-based property suggestions
- **Saved Properties** - Personal collection with session/email tracking
- **Property Alerts** - Email notifications for matching properties

#### Booking System (The "Calendly" Feature)
- **4-Step Booking Wizard**:
  - Step 1: Pre-qualification (income, criminal, rental history, credit requirements)
  - Step 2: Date/time selection (interactive calendar, hourly slots)
  - Step 3: Contact information (with consent tracking)
  - Step 4: Confirmation (with confirmation ID)
- **Availability Management** - 5-layer blackout system (global, property, recurring, vacation, capacity)
- **Auto-FUB Creation** - Every booking creates/updates FUB contact
- **Email/SMS Confirmations** - Automated reminders 2 hours and 1 hour before
- **Reschedule/Cancel** - Self-service booking management

#### User Account
- **Registration** - 4-step wizard (personal info, contact details, security, confirmation)
- **Login/Authentication** - Email/password with session management
- **Password Recovery** - Email-based reset flow
- **Email Verification** - Confirmation link system

### 2. Admin Experience

#### Dashboard
- **Executive KPIs** - Active leads, conversion rate, revenue MTD, bookings this week
- **Critical Alerts** - Pending bookings, pending applications, hot leads, system alerts
- **Activity Timeline** - Real-time feed of user actions
- **Top Opportunities** - AI-detected high-value leads (priority sorted)
- **Upcoming Tasks** - Stale leads, showings, applications

#### Property Management
- **CRUD Operations** - Create, read, update, delete properties
- **Photo Management** - Upload, reorder, set featured image
- **Status Management** - Active, pending_images, available, pending, sold, withdrawn, deleted
- **Property Analytics** - Views, saves, inquiries, showings, applications per property
- **HAR Sync** - Houston Association of Realtors market data integration
- **Property Valuation** - Automated valuation with comparable properties

#### Lead Management
- **Lead List** - Filterable, sortable, paginated lead database
- **Behavioral Scores** - Real-time engagement scores per lead
- **Lead-Property Matching** - AI-powered match recommendations
- **Lead Assignment** - Route leads to team members
- **Activity History** - Complete timeline of interactions
- **FUB Sync Status** - Bidirectional sync tracking

#### Application Workflow (Drag & Drop Kanban)
- **Property Grouping** - Properties with multiple "Application Numbers"
- **Unassigned Pool** - Applicants awaiting assignment
- **Drag-and-Drop** - Move applicants between applications
- **Status Progression** - submitted → review → further_review → rental_history_received → approved/denied/backup/cancelled
- **Agent Assignment** - Assign external agents with contact info
- **FUB Matching** - Automatic lead matching from FUB
- **Buildium Integration** - Auto-create applicants from Buildium email notifications
- **Audit Trail** - Who moved what, when, why

#### Friday Reports (Automated Weekly Intelligence)
Generated every Friday at 9:00 AM with email distribution:

**Sold Listings:**
- Property address
- Sold date
- Lease sent status
- Lease complete status
- Deposit received
- First month rent received
- Status summary
- Alert flags

**Active Listings:**
- MLS ID and address
- Price and days on market (DOM/CDOM)
- Total showings (external + booking platform)
- Week-over-week showing changes
- Total inquiries and lead count
- Week-over-week lead changes
- Applications submitted
- Week-over-week application changes
- Price reduction recommendations (hold/reduce)
- Marketing recommendations
- Contact attempts tracking
- Showing smart feedback:
  - Agent name
  - Interest level (high/medium/low)
  - Price opinion (competitive/high/low)
  - Comparison to comps
  - Comments/notes

**Pre-Listing Pipeline:**
- Properties in preparation
- Expected list dates
- Status tracking

**Weekly Summary:**
- Top performers (agents/properties)
- Key metrics (conversion rate, avg deal size, response time)
- Recommended actions (AI-generated priorities)
- Week range (e.g., "Dec 1-7, 2025")

#### Business Intelligence Dashboard
- **Executive Summary** - Sales volume, occupancy rate, conversion rate
- **Behavioral Intelligence** - High-score leads, active triggers, identified patterns
- **FUB Integration Status** - Safety recipient counts, sync health
- **Campaign Performance** - Email open rates, click rates, conversion rates
- **TREC Compliance** - Compliance scoring and gap identification
- **HAR Sync Status** - Data quality metrics, last sync timestamp
- **Revenue Analysis** - Charts for revenue trends, property mix, deal pipeline
- **Market Analytics** - Average price, total listings, days on market, occupancy percentage

#### Communication Center
- **Email Campaigns** - Create, edit, schedule campaigns
- **Email Templates** - Template library with variable substitution
- **SMS Automation** - Triggered SMS messages
- **Batch Email Sending** - Queue-based bulk email
- **Campaign Analytics** - Delivery rates, open rates, click rates, conversions
- **Email Processing** - Incoming email parsing and routing

#### Team Management
- **Team Members** - Add/edit/remove team members
- **Role Assignment** - Admin, agent, viewer roles
- **Performance Dashboard** - Metrics per team member
- **Lead Assignment** - Route leads to specific agents
- **Agent Dashboard** - Individual agent view with their leads/showings

### 3. AI & Automation Features

#### Spiderweb AI Orchestrator
**Master coordinator that runs intelligence cycles:**

**Intelligence Cycle Steps:**
1. Analyze cross-entity relationships (opportunities)
2. Analyze conversion funnel performance
3. Find new property matches
4. Execute automated campaigns

**Outputs:**
- Dashboard intelligence (top opportunities, funnel analysis, recent matches)
- Opportunity insights with action sequences
- Funnel performance with bottlenecks
- Property match recommendations

**Runs On:**
- Manual trigger via API
- Scheduled interval (configurable, e.g., every 30 minutes)

#### Relationship Intelligence Engine
**Detects 5 opportunity types:**

1. **Hot Leads** (behavioral score > 70)
   - Priority: 70-95
   - Action Sequence: 
     - Send personalized email (immediate, auto-execute)
     - Call if email opened (within 2 hours, manual)
     - Offer showing times (same day, manual)

2. **High View No Contact** (3+ views, 5+ days since last view, no showing booked)
   - Priority: 50-85
   - Action Sequence:
     - Send "Still interested?" email (immediate, auto-execute)
     - Offer specific showing times (within 24 hours, manual)

3. **Re-engagement** (30-90 days inactive, previous score > 30)
   - Priority: 30-50
   - Action Sequence:
     - Send re-engagement email (immediate, auto-execute)
     - Template selected based on days inactive (30-day, 60-day, 90-day)

4. **Conversion Ready** (showing completed OR application started)
   - Priority: 95
   - Action Sequence:
     - Send application reminder (immediate, auto-execute)
     - Personal call to close (within 4 hours, manual)

5. **Property Match** (new property matches lead preferences)
   - Priority: 40-70
   - Action Sequence:
     - Send "New match found" email (immediate, auto-execute)

**Calculations:**
- Conversion Probability = f(behavioral score, activity count, days since activity)
- Priority = (Urgency × 0.4) + (Conversion Probability × 40) + (Revenue Impact × 0.2)
- Revenue Estimate = Average Commission × Conversion Probability

**Insights:**
- AI-generated HTML insights with stats and recommendations
- Comparative analysis (e.g., "Similar leads converted at 85%")
- Urgency indicators (e.g., "Last view was 5 days ago - interest cooling!")

#### Event Campaign Orchestrator
**Automated campaign types:**

1. **Price Changed**
   - Target: Leads who viewed property in last 30 days
   - Template: "Price Reduced: {address} - Now ${new_price}!"
   - Cooldown: 24 hours

2. **New Listing**
   - Target: Leads with matching preferences (zip code, price range, behavioral score > 70)
   - Template: "New Listing Alert: {beds}/{baths} at {address}"
   - Cooldown: 12 hours

3. **Lead Scored Hot**
   - Target: Lead who just scored > 70
   - Template: "Your Houston Property Search - Let's Find Your Perfect Home"
   - Cooldown: 48 hours

4. **Showing Completed**
   - Target: Showing attendee
   - Template: "How was your showing at {address}?"
   - Cooldown: 2 hours

5. **Application Submitted**
   - Target: Applicant
   - Template: "Application Received - Next Steps"
   - Cooldown: 0 (no cooldown)

6. **Property Relisted**
   - Target: Previous interested leads (viewed/saved/inquired/applied)
   - Template: "Back on Market: {address}"
   - Cooldown: 24 hours

7. **Lead Dormant**
   - Target: Lead inactive for 30+ days
   - Template: "Still looking for a place in Houston?"
   - Cooldown: 168 hours (7 days)

8. **Lease Ending Soon**
   - Target: Current tenant
   - Template: "Time to Renew: {address}"
   - Cooldown: 72 hours

**Behavioral Targeting:**
- SQL queries against behavioral_events table
- Lead-property relationship analysis
- Scoring-based filtering
- Duplicate detection to prevent spam

**Template Engine:**
- Variable substitution ({{first_name}}, {{property_address}}, etc.)
- Personalized subject lines
- Dynamic content based on event data

#### Behavioral Scoring Engine
**Composite Score Formula:**
Composite Score = (Urgency × 0.4) + (Engagement × 0.4) + (Financial × 0.2)

**Urgency Score (0-100):**
- Recent activity (last 24 hours): +30 points
- Property saves: +20 points
- Application submissions: +40 points
- Inquiry submissions: +25 points

**Engagement Score (0-100):**
- Property views: 5 points each (max 40)
- Session duration (20+ minutes): +30 points
- Return visits: 5 points each (max 30)
- Saved properties: 10 points each (max 30)

**Financial Score (0-100):**
- Application completion: +50 points
- Income verification: +30 points
- Credit check: +20 points

**Time Decay:**
- Day 1: 100% (factor: 1.0)
- Day 2-7: 80% (factor: 0.8)
- Day 8-30: 50% (factor: 0.5)
- Day 31-90: 20% (factor: 0.2)
- Day 90+: 10% (factor: 0.1)

**Segment Thresholds:**
- 🔥 Hot: 70-100 (Immediate contact, call within 1 hour)
- 🌡️ Warm: 40-69 (Same-day contact, email + follow-up within 4 hours)
- ❄️ Cold: 10-39 (Nurture campaign, drip email, weekly touchpoints)
- 💤 Dormant: 0-9 (Re-engagement, abandonment recovery, special offers)

**Event Point Values:**
- Property Viewed: 5 points
- Property Saved: 15 points
- Inquiry Submitted: 30 points
- Application Started: 35 points
- Application Submitted: 50 points
- Showing Booked: 40 points
- Showing Attended: 45 points
- Conversion (Lease Signed): 100 points (no decay)
- Email Opened: 3 points
- Email Clicked: 8 points

#### Property Matching Service
**Scoring Algorithm:**

Weights:
- Bedrooms: 25%
- Bathrooms: 15%
- Price: 30%
- Location: 20%
- Property Type: 10%

**Bedroom Matching:**
- Exact match: 25 points
- ±1 bedroom: 15 points
- ±2+ bedrooms: 5 points

**Bathroom Matching:**
- Exact match: 15 points
- ±0.5 bathroom: 10 points
- ±1+ bathroom: 5 points

**Price Matching:**
- Within range: 30 points
- Within ±15%: 15-20 points (linear scale)
- Outside ±15%: 0 points

**Location Matching:**
- Same ZIP code: 20 points
- Same city: 15 points
- Different: 5 points

**Property Type Matching:**
- Exact match: 10 points
- Different: 0 points

**Final Score:**
- 0-100 scale
- Minimum 60 to include in matches
- Sorted by score descending

**Preference Inference:**
If explicit preferences missing, infer from last 10 property views:
- Bedroom mode (most common)
- Price range (min/max with ±20% buffer)
- Preferred locations (cities/zips viewed)

#### Funnel Analytics Service
**Tracked Stages:**
1. Visitor (landed on site)
2. Property View (viewed at least 1 property)
3. Property Save (saved a property)
4. Inquiry (contacted about property)
5. Showing Booked (scheduled tour)
6. Showing Attended (completed tour)
7. Application Started (began application)
8. Application Submitted (completed application)
9. Conversion (lease signed)

**Metrics Calculated:**
- Total leads per stage
- Conversion rate per stage
- Overall conversion rate (visitor → conversion)
- Average time in stage
- Drop-off rate per stage
- Bottleneck identification

**Bottleneck Detection:**
Drop-off > 50% between stages = bottleneck

**Analysis Period:**
Configurable (default: 30 days)

#### Abandonment Recovery Service
**Triggers:**
- Cart abandonment (property selected but not booked)
- Form abandonment (booking started but not completed)
- Browse abandonment (viewed multiple properties, no action)

**Recovery Sequences:**
- Step 1: Email reminder (2 hours after abandonment)
- Step 2: SMS follow-up (24 hours after, if email not opened)
- Step 3: Discount offer (48 hours after)
- Step 4: Final reminder (72 hours after)

**A/B Testing:**
- Subject line variations
- Message content variations
- Timing variations
- Incentive variations

**Conditions:**
- Business hours check
- Weekday check
- High engagement check

#### Job Scheduler
**Scheduled Jobs:**

1. **Friday Reports**
   - Schedule: Every Friday at 9:00 AM
   - Timeout: 30 minutes
   - Generates comprehensive weekly report
   - Sends email to stakeholders

2. **HAR Sync**
   - Schedule: Daily at 2:00 AM
   - Timeout: 2 hours
   - Syncs Houston MLS data
   - Updates property information

3. **FUB Sync**
   - Schedule: Every 30 minutes
   - Timeout: 15 minutes
   - Bidirectional contact sync
   - Updates behavioral data

4. **Analytics Aggregation**
   - Schedule: Every hour
   - Timeout: 10 minutes
   - Aggregates dashboard metrics
   - Updates cached statistics

5. **Behavioral Score Recalculation**
   - Schedule: Every 15 minutes
   - Timeout: 5 minutes
   - Recalculates all lead scores
   - Updates segments

**Job Management:**
- Manual trigger via API
- Status tracking (pending, running, completed, failed)
- Retry logic (max 3 attempts)
- Execution logs
- Performance metrics

### 4. FUB Integration

#### Bidirectional Sync
**PropertyHub → FUB:**
- Booking created → Create/update FUB contact
- Call logged → Log to FUB timeline
- Email sent → Log to FUB timeline
- SMS sent → Log to FUB timeline
- Showing scheduled → Create FUB event
- Note added → Sync to FUB
- Lead status updated → Update FUB
- Agent assigned → Update FUB
- Behavioral score updated → Sync custom field

**FUB → PropertyHub:**
- email.opened → Track engagement (+5 score)
- email.clicked → Track engagement (+10 score)
- call.logged → Track engagement (+15 score)
- sms.replied → Track engagement (+10 score)
- note.added → Import note
- person.updated → Sync contact data

#### Deal Creation
**Trigger:** High-value behavioral trigger (showing booked, application submitted, hot lead)

**Deal Data:**
- Name: "{property_category} - {location}"
- Value: Commission calculation
- Stage: Based on behavioral scores
- Probability: Calculated from behavioral data
- Expected Close Date: Based on property category and urgency

**Commission Calculations:**

**Rental Properties:**
Formula: `Monthly Rent - ($100 per $1000 of rent)`
- $5,000/month → $4,500 commission
- $2,000/month → $1,800 commission
- $1,200/month → $1,080 commission

**Sales Properties:**
Formula: `Property Value × Rate × 0.60 (agent split)`
- Standard (≤$1M): 2.5% × 60% = 1.5% to agent
- Luxury (>$1M): 3.0% × 60% = 1.8% to agent

Examples:
- $3M property → $3M × 3% × 60% = $54,000
- $750K property → $750K × 2.5% × 60% = $11,250
- $350K property → $350K × 2.5% × 60% = $5,250

**Deal Stage Mapping:**
- Urgency ≥ 80 && Financial ≥ 80 → "ready_to_close"
- Urgency ≥ 60 && Financial ≥ 70 → "showing_scheduled"
- Engagement ≥ 70 && Financial ≥ 60 → "qualified"
- Engagement ≥ 50 → "contact_made"
- Default → "new_lead"

**Close Probability:**
Base: 50%
+ Urgency adjustment: ±30%
+ Financial adjustment: ±40%
+ Engagement adjustment: ±20%
Clamped: 10% - 95%

**Expected Close Date:**
Base timelines:
- Luxury: 60 days
- Investment: 45 days
- Student housing: 14 days
- Rental: 7 days
- Standard: 30 days

Adjusted by urgency:
- Urgency > 80: 30% faster
- Urgency < 40: 50% slower

#### Action Plan Assignment
**Based on segment:**
- Hot (70-100): "luxury_buyer_immediate_plan"
- Warm (40-69): "active_shopper_nurture_plan"
- Cold (10-39): "long_term_drip_campaign"
- Dormant (0-9): "reactivation_campaign"

**Based on property category:**
- Luxury: "luxury_buyer_premium_service"
- Investment: "investor_portfolio_analysis_plan"
- Student housing: "student_housing_semester_plan"
- First-time buyer: "first_time_buyer_education_plan"

#### List (Pond) Assignment
**Examples:**
- "Hot Leads - Immediate Follow-up"
- "Luxury Buyers - River Oaks/Memorial"
- "Investment Properties - Portfolio Builders"
- "Student Housing - Semester Rush"
- "Emergency Housing Needs"

#### Task Creation
**Priority-based:**
- HIGH: Hot leads, urgent showings, conversion-ready
- MEDIUM: Warm leads, scheduled follow-ups
- LOW: Cold leads, general nurture

**Example Tasks:**
- "💎 LUXURY: Premium client consultation required" (hot luxury lead)
- "🚨 URGENT: Tenant placement required - call within 5 minutes" (urgent rental)
- "📞 Follow up: student_housing lead in rice_village" (student housing)
- "⭐ HIGH PRIORITY: Qualified lead follow-up" (investment)

#### Batch Processing
**Queue System:**
- Priority 1: Lead creation (highest)
- Priority 2: Lead updates
- Priority 3: Note creation
- Priority 4: Bulk sync

**Rate Limiting:**
- 1 request per second
- Batch size: 10 operations
- Max retries: 3 attempts
- Exponential backoff: 1s → 2s → 4s (max 30s)

**Error Handling:**
Retryable: 429 (rate limit), 500, 502, 503, 504
Non-retryable: 400, 401, 403, 404, 422

### 5. Security & Compliance

#### Data Encryption
**Field-Level Encryption (EncryptedString type):**
- Email addresses
- Phone numbers
- Names (first, last)
- Property addresses (in bookings)
- SSN (if collected)

**Encryption Method:**
- AES-256-GCM
- Master key from environment variable
- Transparent encrypt/decrypt via GORM hooks
- Backward compatibility with unencrypted legacy data

**Document Encryption:**
- Lease agreements
- Applications
- Financial documents
- 4-7 year retention periods
- SHA256 integrity checking

#### Audit Logging
**Events Logged:**
- Data access (who accessed what PII, when)
- Admin actions (user creation, role changes, deletions)
- Security events (failed logins, suspicious activity)
- HTTP requests (IP, user agent, method, path)
- Configuration changes

**Retention:**
- Audit logs: 1 year
- Security events: 2 years
- Data access logs: 3 years (compliance requirement)

**Statistics:**
- Events by severity
- Events by category
- Events by user
- Timeline analysis

#### Authentication
**Admin Authentication:**
- JWT token-based
- bcrypt password hashing (cost: 12)
- Session tracking (device fingerprinting, geolocation, IP)
- "Remember me" functionality
- Last login tracking
- Login attempt tracking

**MFA (Multi-Factor Authentication):**
- TOTP (Time-based One-Time Password)
- Backup codes (single-use, XXXXX-XXXXX format)
- QR code generation for authenticator apps
- Enforce for admin accounts
- 90-day attempt log retention

**Session Management:**
- 24-hour default expiration
- Device fingerprinting
- Geolocation tracking (mock fallback)
- Risk scoring (trust level, location change, hours since login, IP change)
- Session hijacking detection
- User agent parsing

#### Input Validation & Protection
**SQL Injection Protection:**
- Parameterized queries via GORM
- Whitelist approach for table/column names
- Query timeout enforcement
- Safe empty results on injection attempts

**XSS Protection:**
- Template security with context-aware escaping
- HTML sanitization (allowlist approach)
- JavaScript sanitization
- URL sanitization (HTTP/HTTPS only)
- CSP (Content Security Policy) headers
- Response data sanitization

**Request Validation:**
- Email format validation
- Phone format validation
- Name validation (alphanumeric + spaces)
- Address validation
- Text sanitization
- ID validation
- Price validation
- SQL injection pattern detection
- XSS pattern detection

#### TREC Compliance
**Texas Real Estate Commission Requirements:**

**Events Logged:**
- Lead generation (source, timestamp, consent)
- Showing scheduled (property, date, attendee)
- Consent collected (type, timestamp, IP address)
- IABS disclosure shown (Information About Brokerage Services)

**Retention:**
- 4 years minimum
- Automatic archival after retention period

**Compliance Report:**
- Event counts by type
- Event timeline
- Compliance gaps (missing required events)
- Audit trail per lead

**Company Info:**
- License: #9008008
- Principal Broker: Terry Evans (#615707)
- Location: Houston, TX
- Phone: (281) 925-7222

#### CAN-SPAM Compliance
**Requirements:**
- Unsubscribe link in every email
- Process unsubscribes within 10 days
- Physical address in emails
- Truthful subject lines
- Identify as advertisement where required

**Do Not Contact (DNC) List:**
- Hard bounces automatically added
- Previous unsubscribes honored
- Suppression list checked before sending
- Manual DNC additions supported

**Consent Tracking:**
- Express consent (checkbox)
- Implied consent (business relationship)
- Unknown consent status
- Revoked consent

#### Rate Limiting
**Endpoint-Specific Limits:**

**Booking API:**
- 5 requests per minute
- 20 requests per hour

**Admin Login:**
- 5 attempts per minute
- 20 attempts per hour
- Temporary IP block after 10 failed attempts

**Public API:**
- 60 requests per minute
- 1000 requests per hour

**Rate Limit Headers:**
- X-RateLimit-Limit
- X-RateLimit-Remaining
- X-RateLimit-Reset

**Enforcement:**
- 429 Too Many Requests response
- Automatic cleanup of old tracking data

#### Security Headers
**Automatically Set:**
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Strict-Transport-Security: max-age=31536000
- Content-Security-Policy: (configurable)

---

## Data Models

### Core Entities

#### Property
```
- MLS_ID (unique identifier)
- Address, City, State, Zip
- Bedrooms, Bathrooms, Square Footage
- Price, Listing Type, Property Type
- Description, Images[], Featured Image
- Agent Info, Office Info
- Status (active, pending, sold, withdrawn, deleted)
- View Count, Last Viewed
- Created/Updated timestamps
```

#### Lead
```
- First Name, Last Name
- Email (encrypted)
- Phone (encrypted)
- Assigned Agent ID
- Behavioral Score (composite)
- Status (active, inactive, converted, unqualified)
- Source (website, referral, ad, etc.)
- Contact History
- Created/Updated timestamps
- Last Contact timestamp
```

#### Booking
```
- Reference Number (unique, BKxxxxxxxx format)
- Property ID (or external property)
- Property Address
- FUB Lead ID
- Email (encrypted)
- Name (encrypted)
- Phone (encrypted)
- Showing Date/Time
- Duration (minutes)
- Status (scheduled, confirmed, completed, cancelled, no_show)
- Notes, Special Requests
- Showing Type (in-person, virtual, self-guided)
- Attendee Count
- FUB Action Plan ID
- Completed At, Cancellation Reason
- Rescheduled From (reference to previous booking)
- Consent Tracking (consent given, marketing consent, terms accepted)
- Created/Updated timestamps
```

#### BehavioralScore
```
- Lead ID (foreign key)
- Composite Score (0-100)
- Urgency Score (0-100)
- Engagement Score (0-100)
- Financial Score (0-100)
- Segment (hot, warm, cold, dormant)
- Last Calculated
- Created/Updated timestamps
```

#### BehavioralEvent
```
- Lead ID (nullable for anonymous)
- Session ID
- Property ID (nullable)
- Event Type (viewed, saved, inquired, applied, converted, etc.)
- Event Data (JSONB)
- IP Address
- User Agent
- Created timestamp
```

#### SavedProperty
```
- Session ID
- Property ID
- Email (nullable for anonymous)
- Notes
- Created/Updated timestamps
```

#### PropertyAlert
```
- Email
- Session ID
- Min/Max Price
- Min/Max Bedrooms
- Min/Max Bathrooms
- Preferred Cities, Zips, Property Types
- Alert Frequency (instant, daily, weekly)
- Active status
- Last Notified
- Created/Updated timestamps
```

#### ApplicationNumber
```
- Property Application Group ID
- Application Number (1, 2, 3...)
- Application Name ("Application 1", "Application 2"...)
- Status (submitted, review, further_review, rental_history_received, approved, denied, backup, cancelled)
- Assigned Agent Name, Phone, Email
- Applicant Count
- Application Notes
- Internal Notes
- Created/Updated timestamps
```

#### ApplicationApplicant
```
- Application Number ID (nullable, unassigned initially)
- Applicant Name
- Applicant Email
- Applicant Phone
- Application Date
- Source Email (e.g., "buildium_notification")
- FUB Lead ID (if matched)
- FUB Match (boolean)
- Match Score (0.0-1.0)
- Created/Updated timestamps
```

### Supporting Models

#### PropertyNotification
```
- Property ID
- Email
- Notification Type (price_change, status_change, new_listing)
- Channel (email, sms, push, web)
- Match Score
- Sent status, Opened status, Clicked status
- Sent At timestamp
```

#### PropertyValuation
```
- Property ID
- Valuation Amount
- Confidence Score
- Comparable Properties (JSON)
- Market Data (JSON)
- Valuation Date
- Created/Updated timestamps
```

#### WebhookEvent
```
- Source (fub, buildium, stripe, twilio)
- Event Type
- Event ID
- Payload (JSON)
- Processing Status (pending, processing, completed, failed)
- Retry Count
- Processed At, Failed At
- Error Message
- Created/Updated timestamps
```

#### CampaignExecution
```
- Lead Reengagement ID (nullable)
- Campaign Template ID (nullable)
- Scheduled For
- Executed At
- Status (pending, sent, failed)
- Email Opened, Email Clicked, Responded
- Retry Count, Error Message
- FUB Action Plan ID, FUB Step ID
- Created/Updated timestamps
```

---

## API Endpoints

### Consumer API

#### Properties
```
GET    /api/v1/properties                 - List properties (filtered, sorted, paginated)
GET    /api/v1/properties/:id             - Get property details
POST   /api/v1/properties/search          - Advanced search
GET    /api/recommendations               - AI recommendations
GET    /api/properties/:id/similar        - Similar properties
```

#### Bookings
```
POST   /api/v1/bookings                   - Create booking
GET    /api/v1/bookings/:id               - Get booking
DELETE /api/v1/bookings/:id               - Cancel booking
GET    /api/v1/bookings                   - List bookings
PUT    /api/bookings/:id/reschedule       - Reschedule booking
POST   /api/bookings/:id/complete         - Mark completed
POST   /api/bookings/:id/no-show          - Mark no-show
```

#### Saved Properties & Alerts
```
POST   /api/v1/properties/save/:id        - Save property
DELETE /api/v1/properties/save/:id        - Unsave property
GET    /api/v1/properties/saved           - List saved properties
POST   /alerts/subscribe                  - Subscribe to alerts
GET    /alerts/preferences                - Get alert preferences
PUT    /alerts/update                     - Update preferences
DELETE /alerts/unsubscribe                - Unsubscribe
```

#### Availability
```
GET    /api/v1/availability/check         - Check availability
```

### Admin API

#### Dashboard & Analytics
```
GET    /api/v1/dashboard/metrics          - Dashboard KPIs
GET    /api/v1/bi/dashboard               - BI dashboard data
GET    /api/business-intelligence/friday-report  - Friday report
GET    /api/v1/bi/properties/analytics    - Property analytics
GET    /api/v1/bi/leads/funnel            - Conversion funnel
```

#### Properties
```
POST   /api/v1/properties                 - Create property
PUT    /api/v1/properties/:id             - Update property
DELETE /api/v1/properties/:id             - Delete property
POST   /api/v1/properties/:id/photos      - Upload photos
PUT    /api/v1/properties/:id/status      - Update status
POST   /api/v1/properties/sync/har        - Trigger HAR sync
```

#### Bookings
```
GET    /api/v1/bookings                   - List all bookings (admin view)
PUT    /api/v1/bookings/:id/status        - Update booking status
```

#### Application Workflow
```
GET    /api/properties-with-applications  - Get properties with application numbers
GET    /api/applicants-without-application - Get unassigned applicants
POST   /api/application-workflow/application-number - Create application number
POST   /api/application-workflow/move-applicant - Move applicant (drag-drop)
POST   /api/application-workflow/assign-agent - Assign agent
PUT    /api/application-workflow/:id/status - Update application status
POST   /api/application-workflow/:id/approve - Approve application
POST   /api/application-workflow/:id/deny - Deny application
POST   /api/application-workflow/:id/request-info - Request more info
```

#### Availability Management
```
GET    /api/v1/availability/blackouts     - List blackouts
POST   /api/v1/availability/blackouts/create - Create blackout
DELETE /api/v1/availability/blackouts/:id - Delete blackout
POST   /api/v1/availability/cleanup       - Remove expired blackouts
```

#### FUB Integration
```
POST   /api/v1/fub/webhook                - FUB webhook receiver
POST   /api/v1/fub/log-call               - Log call to FUB
POST   /api/v1/fub/log-email              - Log email to FUB
GET    /api/fub/sync-status               - Sync status dashboard
GET    /api/fub/sync-status/health        - Health check
POST   /api/fub/sync-status/retry         - Retry failed syncs
```

#### Lead Management
```
GET    /api/v1/leads                      - List leads
POST   /api/v1/leads                      - Create lead
GET    /api/v1/leads/:id                  - Get lead details
GET    /leads/{id}/matches                - Get property matches for lead
```

#### Behavioral Intelligence
```
GET    /api/v1/behavioral/trends          - Behavioral trends
GET    /api/v1/behavioral/funnel          - Funnel analysis
GET    /api/v1/behavioral/segments        - Segment analysis
GET    /api/v1/behavioral/heatmap         - Activity heatmap
GET    /api/v1/behavioral/sessions/active - Active sessions
GET    /api/v1/behavioral/sessions/:id/journey - Session journey
```

#### Campaign Management
```
GET    /api/v1/reengagement/campaigns     - List campaigns
POST   /api/v1/reengagement/campaigns/prepare - Prepare campaign
POST   /api/v1/reengagement/campaigns/activate - Activate campaign
GET    /api/v1/reengagement/campaigns/:id/status - Campaign status
POST   /api/v1/reengagement/emergency/stop - Emergency stop all
POST   /api/v1/reengagement/emergency/stop/:id - Stop specific campaign
```

#### Email Management
```
GET    /api/v1/email-automation/campaigns - List campaigns
POST   /api/v1/email-automation/campaigns - Create campaign
PUT    /api/v1/email-automation/campaigns/:id - Update campaign
DELETE /api/v1/email-automation/campaigns/:id - Delete campaign
GET    /api/v1/email-automation/processing-history - Processing history
GET    /api/v1/email-automation/analytics - Email analytics
```

### Admin Pages (HTML)
```
GET    /admin/dashboard                   - Main dashboard
GET    /admin/property-list               - Property list
GET    /admin/property-create             - Create property form
GET    /admin/property-edit/:id           - Edit property form
GET    /admin/property-detail/:id         - Property detail view
GET    /admin/booking-detail/:id          - Booking detail view
GET    /admin/application-workflow        - Application workflow Kanban
GET    /admin/business-intelligence       - BI dashboard
GET    /admin/communication-center        - Communication center
GET    /admin/calendar                    - Calendar management
GET    /admin/team-dashboard              - Team management
GET    /admin/settings                    - System settings
```

### Consumer Pages (HTML)
```
GET    /                                  - Homepage
GET    /properties                        - Property grid
GET    /property/:id                      - Property detail
GET    /book-showing                      - Booking wizard
GET    /booking-confirmation              - Booking confirmation
GET    /saved-properties                  - Saved properties
GET    /property-alerts                   - Alert preferences
GET    /login                             - Consumer login
GET    /register                          - Consumer registration
GET    /about                             - About page
GET    /contact                           - Contact page
GET    /privacy                           - Privacy policy
GET    /terms                             - Terms of service
GET    /trec-compliance                   - TREC compliance info
```

---

## Technology Stack

### Backend
- **Language:** Go 1.22+
- **Framework:** Gin (web framework)
- **Database:** PostgreSQL with GORM ORM
- **Cache:** Redis (optional but recommended)
- **Authentication:** JWT tokens, bcrypt password hashing
- **Encryption:** AES-256-GCM

### Frontend
- **Templates:** Go HTML templates
- **JavaScript:** Alpine.js for interactivity, vanilla JS for complex features
- **CSS:** Custom design system with CSS variables
- **Fonts:** Playfair Display (serif), Inter (sans-serif), IBM Plex Mono (mono)
- **Icons:** Heroicons SVG sprite

### External Services
- **Email:** AWS SES (Simple Email Service) - Recommended
- **SMS:** AWS SNS (Simple Notification Service) - Recommended
- **CRM:** Follow Up Boss (fully integrated)
- **Property Data:** Generic scraper (HAR blocked access)

### Infrastructure
- **Server:** Linux/Ubuntu 22.04
- **Deployment:** Docker or traditional binary deployment
- **Reverse Proxy:** Nginx or Caddy
- **SSL:** Let's Encrypt
- **Monitoring:** (TBD - Sentry, Datadog, etc.)

---

## Environment Variables

### Required
```
DATABASE_URL=postgres://user:pass@host:5432/propertyhub?sslmode=require
ENCRYPTION_KEY=<32-byte-base64-key>
JWT_SECRET=<64-char-random-string>
SESSION_SECRET=<64-char-random-string>
FUB_API_KEY=<fub-api-key>
```

### Optional
```
PORT=8080
GIN_MODE=release
REDIS_URL=localhost:6379
REDIS_PASSWORD=<redis-password>
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<aws-access-key>
AWS_SECRET_ACCESS_KEY=<aws-secret-key>
SES_FROM_EMAIL=noreply@landlords-of-texas.com
SNS_SENDER_ID=PropertyHub
```

---

## Success Metrics

### Lead Management
- **Response Time:** < 15 minutes for hot leads
- **Conversion Rate:** > 15% (visitor to lease signed)
- **Lead Score Accuracy:** > 90% (hot leads should convert)

### Automation
- **Campaign Open Rate:** > 25%
- **Campaign Click Rate:** > 10%
- **Campaign Conversion Rate:** > 5%
- **Automated Actions:** > 80% of routine tasks handled without human intervention

### System Performance
- **API Response Time:** < 200ms (95th percentile)
- **Page Load Time:** < 2 seconds
- **Uptime:** > 99.5%
- **Error Rate:** < 0.1%

### Business Outcomes
- **Opportunities Identified:** > 100/week from 13K lead database
- **Revenue Per Lead:** Tracked via FUB deal values
- **Time Saved:** > 20 hours/week vs manual processes
- **Lead Reactivation Rate:** > 10% of dormant leads re-engaged

---

## Deployment Checklist

See @docs/PRODUCTION_CHECKLIST.md for complete deployment checklist including:
- Environment variable validation
- Database migration verification
- Security configuration
- SSL certificate setup
- Rate limiting configuration
- Email/SMS service testing
- FUB integration testing
- Backup strategy
- Monitoring setup

---

## Future Enhancements (Out of Scope for v1.0)

- Google Calendar / Outlook Calendar integration
- Stripe payment processing for deposits
- Mobile app (iOS/Android)
- Tenant portal (post-lease management)
- Maintenance request system
- Lease renewal automation
- Accounting integration (QuickBooks, Xero)
- Advanced ML models for lead scoring
- Predictive analytics (days to lease, optimal pricing)
- Multi-language support
- White-label customization

---

**Document Owner:** Christopher Gross  
**Maintained By:** Development Team  
**Review Frequency:** Quarterly or before major releases
