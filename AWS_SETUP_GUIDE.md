# AWS SES + SNS Setup Guide

**Last Updated:** December 7, 2025  
**Purpose:** Configure email (SES) and SMS (SNS) for PropertyHub

---

## Why AWS Instead of SendGrid/Twilio?

**AWS SES (Email):**
- ✅ No blocking issues (SendGrid won't unblock)
- ✅ $0.10 per 1,000 emails (10x cheaper than SendGrid)
- ✅ 50,000 emails/day sending limit (scalable)
- ✅ Excellent deliverability
- ✅ No daily quotas after verification

**AWS SNS (SMS):**
- ✅ $0.00645 per SMS (cheaper than Twilio)
- ✅ Reliable delivery
- ✅ Same provider as email (unified billing)
- ✅ Transactional SMS support

---

## Setup Steps

### Step 1: Create AWS Account (5 minutes)

1. Go to https://aws.amazon.com
2. Click "Create an AWS Account"
3. Enter email, password, AWS account name
4. Enter payment information (credit card required)
5. Verify phone number
6. Select "Basic Support - Free" plan

**Cost:** Free tier includes:
- 62,000 emails/month via SES (first 12 months)
- After free tier: $0.10 per 1,000 emails
- SMS: ~$0.00645 per message (no free tier)

---

### Step 2: Create IAM User (5 minutes)

1. Log into AWS Console
2. Go to **IAM** (Identity and Access Management)
3. Click **Users** → **Create user**
4. User name: `propertyhub-ses-sns`
5. Check "Provide user access to the AWS Management Console" → **NO** (programmatic access only)
6. Click **Next**
7. Permissions: **Attach policies directly**
8. Search and select:
   - `AmazonSESFullAccess`
   - `AmazonSNSFullAccess`
9. Click **Next** → **Create user**

---

### Step 3: Get Access Keys (2 minutes)

1. Click on the user you just created
2. Go to **Security credentials** tab
3. Scroll to **Access keys** section
4. Click **Create access key**
5. Purpose: **Application running outside AWS**
6. Click **Next** → **Create access key**
7. **IMPORTANT:** Download the CSV or copy:
   - Access Key ID: `AKIA...`
   - Secret Access Key: `wJal...` (only shown once!)
8. Store these securely (password manager)

---

### Step 4: Verify Sender Email in SES (10 minutes)

**Important:** SES starts in "sandbox mode" - you can only send to verified emails. After verification, request production access.

1. Go to **Amazon SES** in AWS Console
2. Select your region (e.g., **US East (N. Virginia)** = us-east-1)
3. Click **Verified identities** (left sidebar)
4. Click **Create identity**
5. Identity type: **Email address**
6. Email address: `noreply@landlords-of-texas.com` (or your domain)
7. Click **Create identity**
8. Check your email inbox for verification email from AWS
9. Click the verification link
10. Status changes to **Verified** ✅

**Alternative:** Verify entire domain (recommended for production)
- Instead of email address, select "Domain"
- Enter: `landlords-of-texas.com`
- Add DNS records (DKIM, SPF) shown in AWS console
- Wait for verification (1-72 hours)

---

### Step 5: Request Production Access (Optional, 24-48 hours)

While in sandbox mode, you can only send to verified emails. For production:

1. In SES console, click **Account dashboard**
2. Look for sandbox mode warning
3. Click **Request production access**
4. Fill out form:
   - **Mail type:** Transactional
   - **Website URL:** https://landlords-of-texas.com
   - **Use case description:**
     ```
     Property management platform sending transactional emails:
     - Booking confirmations for property showings
     - Appointment reminders
     - Application status updates
     - Property alerts for interested leads
     
     Expected volume: 500-1000 emails/day
     All recipients have opted in via website forms.
     ```
5. Submit request
6. Wait 24-48 hours for approval

**For testing:** You can skip this and just verify test email addresses

---

### Step 6: Configure PropertyHub (2 minutes)

Add to your `.env` file or database `system_settings` table:

```bash
# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=AKIA123456789EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# SES Configuration
SES_FROM_EMAIL=noreply@landlords-of-texas.com

# SNS Configuration (for SMS)
SNS_SENDER_ID=PropertyHub
```

**Security Note:** Never commit these to git. Use environment variables or AWS Secrets Manager.

---

### Step 7: Test Email Sending (5 minutes)

Create a test script or use the API:

**Test via curl:**
```bash
# After server is running
curl -X POST http://localhost:8080/api/v1/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "property_id": 1,
    "first_name": "Test",
    "last_name": "User",
    "email": "your-verified-email@example.com",
    "phone": "+15555551234",
    "showing_date": "2025-12-10T15:00:00Z"
  }'
```

**Check logs:**
```
✅ AWS Communication Service initialized (SES + SNS)
📧 Email sent via SES to your-email@example.com (MessageID: ...)
```

---

### Step 8: Test SMS Sending (5 minutes)

**Test via curl:**
```bash
# Simple SMS test (you'll need an endpoint that triggers SMS)
# SMS will be sent when booking reminder is triggered
```

**Check logs:**
```
✅ SMS sent via SNS to +15555551234 (MessageID: ...)
```

---

## Troubleshooting

### "Email not sent" - SES Not Verified

**Problem:** SES sandbox mode, recipient not verified

**Solution:**
1. Go to SES console → Verified identities
2. Verify the recipient email address
3. OR request production access (see Step 5)

---

### "InvalidParameterValue: Unverified email address"

**Problem:** Sender email not verified in SES

**Solution:**
1. Go to SES → Verified identities
2. Verify noreply@landlords-of-texas.com
3. Check inbox for verification email

---

### "AccessDenied" or "UnauthorizedOperation"

**Problem:** IAM user doesn't have SES/SNS permissions

**Solution:**
1. Go to IAM → Users → propertyhub-ses-sns
2. Permissions tab → Add permissions
3. Attach `AmazonSESFullAccess` and `AmazonSNSFullAccess`

---

### "InvalidClientTokenId" or "SignatureDoesNotMatch"

**Problem:** Wrong AWS credentials

**Solution:**
1. Verify AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
2. Regenerate access keys if needed (IAM console)
3. Update environment variables

---

### SMS not sending

**Problem:** Phone number not in E.164 format

**Solution:**
- Use format: `+1` (country code) + area code + number
- Example: `+15551234567` (USA)
- NOT: `555-123-4567` or `(555) 123-4567`

---

## Cost Estimates

### SES Email Costs
- First 62,000 emails/month: **FREE** (first 12 months)
- After free tier: **$0.10 per 1,000 emails**
- Attachments: $0.12 per GB

**Example:**
- 10,000 emails/month = $1.00/month
- 50,000 emails/month = $5.00/month
- 100,000 emails/month = $10.00/month

### SNS SMS Costs (USA)
- $0.00645 per SMS

**Example:**
- 100 SMS/month = $0.65/month
- 500 SMS/month = $3.23/month
- 1,000 SMS/month = $6.45/month

**Total estimated cost for PropertyHub:**
- 5,000 emails + 200 SMS/month = **$1.79/month**
- 20,000 emails + 500 SMS/month = **$5.23/month**
- 50,000 emails + 1,000 SMS/month = **$11.45/month**

**vs SendGrid + Twilio:**
- SendGrid Essentials: $19.95/month
- Twilio: $0.0079/SMS
- Total: ~$25/month minimum

**Savings:** ~$15-20/month with AWS

---

## Production Recommendations

### 1. Use AWS Secrets Manager for Credentials
Instead of environment variables:

```go
// Use AWS Secrets Manager to fetch credentials
// More secure than environment variables
```

Cost: $0.40/month per secret

---

### 2. Set up CloudWatch Alarms

Monitor:
- SES bounce rate (> 5% = problem)
- SES complaint rate (> 0.1% = problem)
- SNS failed deliveries
- Daily sending quota usage

---

### 3. Configure SES Sending Limits

Default limits after verification:
- 14 emails/second
- 50,000 emails/24 hours

Request limit increases if needed (AWS support ticket)

---

### 4. Add Bounce/Complaint Handling

Set up SNS topics to receive:
- Bounce notifications
- Complaint notifications
- Update DNC list automatically

---

## Quick Reference

### AWS Console URLs
- **SES Dashboard:** https://console.aws.amazon.com/ses/
- **SNS Dashboard:** https://console.aws.amazon.com/sns/
- **IAM Users:** https://console.aws.amazon.com/iam/home#/users

### Verification Status Check
```bash
aws ses list-identities --region us-east-1
aws ses get-identity-verification-attributes --identities noreply@landlords-of-texas.com --region us-east-1
```

### Test Email via AWS CLI
```bash
aws ses send-email \
  --from noreply@landlords-of-texas.com \
  --to your-email@example.com \
  --subject "Test Email" \
  --text "This is a test email from PropertyHub" \
  --region us-east-1
```

### Test SMS via AWS CLI
```bash
aws sns publish \
  --phone-number +15555551234 \
  --message "Test SMS from PropertyHub" \
  --region us-east-1
```

---

## What's Already Implemented ✅

- ✅ AWS SDK integration in `@internal/services/aws_communication_service.go`
- ✅ Email sending via SES with HTML support
- ✅ SMS sending via SNS with transactional priority
- ✅ Bulk email support (rate-limited)
- ✅ Bulk SMS support
- ✅ Error handling and logging
- ✅ Graceful degradation if AWS not configured
- ✅ Integration with existing EmailService and SMSService
- ✅ Safety controls integration

**All you need to do:** Add AWS credentials to environment variables and verify sender email.

---

**Total setup time:** 30-45 minutes (excluding production access request)
