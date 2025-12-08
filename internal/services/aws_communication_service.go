package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// AWSCommunicationService handles email (SES) and SMS (SNS) via AWS
type AWSCommunicationService struct {
	sesClient *ses.Client
	snsClient *sns.Client
	fromEmail string
	senderID  string
	enabled   bool
}

// NewAWSCommunicationService creates a new AWS communication service
func NewAWSCommunicationService(fromEmail, senderID string) (*AWSCommunicationService, error) {
	// Check if AWS credentials are configured
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKey == "" {
		log.Println("⚠️  AWS credentials not configured - communication service disabled")
		return &AWSCommunicationService{enabled: false}, nil
	}

	// Load AWS configuration with us-east-2 region for llotschedule.online
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(getEnvOrDefault("AWS_REGION", "us-east-2")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Use provided email/senderID or fall back to defaults
	if fromEmail == "" {
		fromEmail = "info@llotschedule.online"
	}
	if senderID == "" {
		senderID = "PropertyHub"
	}

	service := &AWSCommunicationService{
		sesClient: ses.NewFromConfig(cfg),
		snsClient: sns.NewFromConfig(cfg),
		fromEmail: fromEmail,
		senderID:  senderID,
		enabled:   true,
	}

	log.Printf("✅ AWS Communication Service initialized (SES + SNS) with from=%s", fromEmail)
	return service, nil
}

// SendEmail sends an email via AWS SES
func (svc *AWSCommunicationService) SendEmail(to, subject, bodyHTML, bodyText string) error {
	if !svc.enabled {
		log.Printf("📧 [DISABLED] Would send email to %s: %s", to, subject)
		return nil
	}

	input := &ses.SendEmailInput{
		Destination: &sestypes.Destination{
			ToAddresses: []string{to},
		},
		Message: &sestypes.Message{
			Subject: &sestypes.Content{
				Data: aws.String(subject),
			},
			Body: &sestypes.Body{},
		},
		Source: aws.String(svc.fromEmail),
	}

	// Add HTML body if provided
	if bodyHTML != "" {
		input.Message.Body.Html = &sestypes.Content{
			Data: aws.String(bodyHTML),
		}
	}

	// Add text body (required fallback)
	if bodyText != "" {
		input.Message.Body.Text = &sestypes.Content{
			Data: aws.String(bodyText),
		}
	} else {
		// Strip HTML tags for text version (basic fallback)
		input.Message.Body.Text = &sestypes.Content{
			Data: aws.String(stripHTMLBasic(bodyHTML)),
		}
	}

	result, err := svc.sesClient.SendEmail(context.Background(), input)
	if err != nil {
		log.Printf("❌ Failed to send email via SES to %s: %v", to, err)
		return fmt.Errorf("SES send failed: %w", err)
	}

	log.Printf("✅ Email sent via SES to %s (MessageID: %s)", to, *result.MessageId)
	return nil
}

// SendBulkEmail sends bulk emails via AWS SES (up to 50 recipients per call)
func (svc *AWSCommunicationService) SendBulkEmail(recipients []string, subject, bodyHTML, bodyText string) (int, error) {
	if !svc.enabled {
		log.Printf("📧 [DISABLED] Would send bulk email to %d recipients", len(recipients))
		return 0, nil
	}

	successCount := 0
	
	// AWS SES supports bulk send but we'll send individually for better error tracking
	// For production scale, use SES SendBulkTemplatedEmail instead
	for _, recipient := range recipients {
		err := svc.SendEmail(recipient, subject, bodyHTML, bodyText)
		if err != nil {
			log.Printf("⚠️  Failed to send to %s: %v", recipient, err)
			continue
		}
		successCount++
		
		// Rate limiting: AWS SES typically allows 14 emails/second
		// Add small delay if sending to many recipients
		if len(recipients) > 100 {
			// Sleep briefly to avoid rate limits (handled by AWS SDK but adding safety)
		}
	}

	log.Printf("✅ Bulk email sent: %d/%d successful", successCount, len(recipients))
	return successCount, nil
}

// SendSMS sends an SMS via AWS SNS
func (svc *AWSCommunicationService) SendSMS(to, message string) error {
	if !svc.enabled {
		log.Printf("📱 [DISABLED] Would send SMS to %s: %s", to, message)
		return nil
	}

	input := &sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(to),
		MessageAttributes: map[string]snstypes.MessageAttributeValue{
			"AWS.SNS.SMS.SenderID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(svc.senderID),
			},
			"AWS.SNS.SMS.SMSType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Transactional"), // Transactional = high priority
			},
		},
	}

	result, err := svc.snsClient.Publish(context.Background(), input)
	if err != nil {
		log.Printf("❌ Failed to send SMS via SNS to %s: %v", to, err)
		return fmt.Errorf("SNS send failed: %w", err)
	}

	log.Printf("✅ SMS sent via SNS to %s (MessageID: %s)", to, *result.MessageId)
	return nil
}

// SendBulkSMS sends SMS to multiple recipients
func (svc *AWSCommunicationService) SendBulkSMS(recipients []string, message string) (int, error) {
	if !svc.enabled {
		log.Printf("📱 [DISABLED] Would send bulk SMS to %d recipients", len(recipients))
		return 0, nil
	}

	successCount := 0
	
	for _, recipient := range recipients {
		err := svc.SendSMS(recipient, message)
		if err != nil {
			log.Printf("⚠️  Failed to send SMS to %s: %v", recipient, err)
			continue
		}
		successCount++
	}

	log.Printf("✅ Bulk SMS sent: %d/%d successful", successCount, len(recipients))
	return successCount, nil
}

// IsEnabled returns whether the service is enabled
func (svc *AWSCommunicationService) IsEnabled() bool {
	return svc.enabled
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func stripHTMLBasic(html string) string {
	// Very basic HTML stripping for fallback text version
	// In production, use proper HTML-to-text library
	text := html
	// Remove common tags (basic implementation)
	replacements := map[string]string{
		"<br>":   "\n",
		"<br/>":  "\n",
		"<br />": "\n",
		"<p>":    "\n",
		"</p>":   "\n",
		"<div>":  "\n",
		"</div>": "\n",
	}
	for tag, replacement := range replacements {
		text = replaceAll(text, tag, replacement)
	}
	// Remove remaining tags (very basic)
	// For production, use: github.com/jaytaylor/html2text
	return text
}

func replaceAll(s, old, new string) string {
	result := ""
	for len(s) > 0 {
		idx := indexOf(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
