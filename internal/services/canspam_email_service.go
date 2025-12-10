package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"html/template"
	
	"gorm.io/gorm"
)

// CANSPAMEmailService handles CAN-SPAM compliant email templating
type CANSPAMEmailService struct {
	db *gorm.DB
	companyInfo CompanyInfo
}

// CompanyInfo holds company information for CAN-SPAM compliance
type CompanyInfo struct {
	CompanyName    string `json:"company_name"`
	PhysicalAddress string `json:"physical_address"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Website        string `json:"website"`
}

// CANSPAMTemplate represents a CAN-SPAM compliant email template
type CANSPAMTemplate struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"uniqueIndex"`
	DisplayName      string    `json:"display_name"`
	Subject          string    `json:"subject"`
	HTMLContent      string    `json:"html_content" gorm:"type:text"`
	TextContent      string    `json:"text_content" gorm:"type:text"`
	TemplateType     string    `json:"template_type"` // marketing, transactional
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	HasUnsubscribe   bool      `json:"has_unsubscribe" gorm:"default:true"`
	HasPhysicalAddr  bool      `json:"has_physical_addr" gorm:"default:true"`
	HasClearSender   bool      `json:"has_clear_sender" gorm:"default:true"`
	ComplianceScore  int       `json:"compliance_score" gorm:"default:0"` // 0-100
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// EmailData represents data for email template rendering
type EmailData struct {
	RecipientName    string                 `json:"recipient_name"`
	RecipientEmail   string                 `json:"recipient_email"`
	PropertyAddress  string                 `json:"property_address,omitempty"`
	ShowingTime     string                 `json:"showing_time,omitempty"`
	UnsubscribeURL  string                 `json:"unsubscribe_url"`
	CustomData      map[string]interface{} `json:"custom_data"`
	CompanyInfo     CompanyInfo           `json:"company_info"`
	SentDate        string                `json:"sent_date"`
	UnsubscribeID   string                `json:"unsubscribe_id"`
}

// NewCANSPAMEmailService creates a new CAN-SPAM compliant email service
func NewCANSPAMEmailService(db *gorm.DB) *CANSPAMEmailService {
	// Auto-migrate the templates table
	db.AutoMigrate(&CANSPAMTemplate{})
	
	service := &CANSPAMEmailService{
		db: db,
		companyInfo: CompanyInfo{
			CompanyName:    "Landlords of Texas, LLC",
			PhysicalAddress: "Houston, TX 77002", // Update with actual address
			Phone:          "(713) 555-PROP",
			Email:          "info@landlordsoftexas.com", // Update with actual email
			Website:        "https://propertyhubtx.com",
		},
	}
	
	// Initialize default templates
	service.initializeDefaultTemplates()
	return service
}

// RenderTemplate renders a CAN-SPAM compliant email template
func (c *CANSPAMEmailService) RenderTemplate(templateName string, data EmailData) (*RenderedEmail, error) {
	var tpl CANSPAMTemplate
	if err := c.db.Where("name = ? AND is_active = ?", templateName, true).First(&tpl).Error; err != nil {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}
	
	// Ensure compliance data is set
	data.CompanyInfo = c.companyInfo
	data.SentDate = time.Now().Format("January 2, 2006")
	
	// Generate unsubscribe URL and ID if not provided
	if data.UnsubscribeURL == "" || data.UnsubscribeID == "" {
		unsubID := c.generateUnsubscribeID(data.RecipientEmail)
		data.UnsubscribeID = unsubID
		data.UnsubscribeURL = fmt.Sprintf("https://propertyhubtx.com/unsubscribe?id=%s&email=%s", 
			unsubID, data.RecipientEmail)
	}
	
	// Render HTML content
	htmlTemplate, err := template.New("html").Parse(tpl.HTMLContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %v", err)
	}
	
	var htmlBuffer strings.Builder
	if err := htmlTemplate.Execute(&htmlBuffer, data); err != nil {
		return nil, fmt.Errorf("failed to render HTML template: %v", err)
	}
	
	// Render text content
	textTemplate, err := template.New("text").Parse(tpl.TextContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text template: %v", err)
	}
	
	var textBuffer strings.Builder
	if err := textTemplate.Execute(&textBuffer, data); err != nil {
		return nil, fmt.Errorf("failed to render text template: %v", err)
	}
	
	// Render subject
	subjectTemplate, err := template.New("subject").Parse(tpl.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subject template: %v", err)
	}
	
	var subjectBuffer strings.Builder
	if err := subjectTemplate.Execute(&subjectBuffer, data); err != nil {
		return nil, fmt.Errorf("failed to render subject template: %v", err)
	}
	
	return &RenderedEmail{
		Subject:     subjectBuffer.String(),
		HTMLContent: htmlBuffer.String(),
		TextContent: textBuffer.String(),
		Headers: map[string]string{
			"List-Unsubscribe": fmt.Sprintf("<%s>", data.UnsubscribeURL),
			"List-Unsubscribe-Post": "List-Unsubscribe=One-Click",
		},
		ComplianceScore: tpl.ComplianceScore,
	}, nil
}

// RenderedEmail represents a fully rendered CAN-SPAM compliant email
type RenderedEmail struct {
	Subject         string            `json:"subject"`
	HTMLContent     string            `json:"html_content"`
	TextContent     string            `json:"text_content"`
	Headers         map[string]string `json:"headers"`
	ComplianceScore int               `json:"compliance_score"`
}

// generateUnsubscribeID generates a cryptographically secure unsubscribe ID
func (c *CANSPAMEmailService) generateUnsubscribeID(email string) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		b = []byte(fmt.Sprintf("%d%s", time.Now().UnixNano(), email))
	}

	h := hmac.New(sha256.New, []byte(c.companyInfo.Email))
	h.Write([]byte(email))
	h.Write(b)

	return base64.URLEncoding.EncodeToString(h.Sum(nil))[:32]
}

// initializeDefaultTemplates creates default CAN-SPAM compliant email templates
func (c *CANSPAMEmailService) initializeDefaultTemplates() {
	templates := []CANSPAMTemplate{
		{
			Name:         "property_inquiry_response",
			DisplayName:  "Property Inquiry Response",
			Subject:      "Thank you for your interest in {{.PropertyAddress}}",
			TemplateType: "transactional",
			HTMLContent:  c.getPropertyInquiryHTMLTemplate(),
			TextContent:  c.getPropertyInquiryTextTemplate(),
			ComplianceScore: 100,
		},
		{
			Name:         "showing_confirmation",
			DisplayName:  "Showing Confirmation",
			Subject:      "Your showing is confirmed for {{.PropertyAddress}} on {{.ShowingTime}}",
			TemplateType: "transactional",
			HTMLContent:  c.getShowingConfirmationHTMLTemplate(),
			TextContent:  c.getShowingConfirmationTextTemplate(),
			ComplianceScore: 100,
		},
		{
			Name:         "property_availability_alert",
			DisplayName:  "Property Availability Alert",
			Subject:      "New property available that matches your criteria",
			TemplateType: "marketing",
			HTMLContent:  c.getPropertyAvailabilityHTMLTemplate(),
			TextContent:  c.getPropertyAvailabilityTextTemplate(),
			ComplianceScore: 100,
		},
		{
			Name:         "application_status_update",
			DisplayName:  "Application Status Update",
			Subject:      "Application Update for {{.PropertyAddress}}",
			TemplateType: "transactional",
			HTMLContent:  c.getApplicationStatusHTMLTemplate(),
			TextContent:  c.getApplicationStatusTextTemplate(),
			ComplianceScore: 100,
		},
	}
	
	for _, tpl := range templates {
		var existing CANSPAMTemplate
		result := c.db.Where("name = ?", tpl.Name).First(&existing)
		if result.Error != nil {
			// Template doesn't exist, create it
			c.db.Create(&tpl)
		}
	}
}

// Property Inquiry Response Templates
func (c *CANSPAMEmailService) getPropertyInquiryHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Property Inquiry Response</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #003366; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .footer { background: #eee; padding: 20px; font-size: 12px; text-align: center; }
        .unsubscribe { margin-top: 20px; padding: 15px; background: #fff3cd; border: 1px solid #ffeaa7; }
        .btn { background: #003366; color: white; padding: 12px 25px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.CompanyInfo.CompanyName}}</h1>
        <p>Your Houston Area Rental Specialists</p>
    </div>
    
    <div class="content">
        <h2>Hi {{.RecipientName}},</h2>
        
        <p>Thank you for your inquiry about <strong>{{.PropertyAddress}}</strong>!</p>
        
        <p>We received your request and one of our rental specialists will contact you within 2 business hours with detailed information about this property, including:</p>
        
        <ul>
            <li>Complete property details and amenities</li>
            <li>Available showing times</li>
            <li>Lease terms and pricing</li>
            <li>Application requirements</li>
        </ul>
        
        <p>In the meantime, feel free to browse our other available properties on our website.</p>
        
        <p><a href="{{.CompanyInfo.Website}}" class="btn">View More Properties</a></p>
        
        <p>If you have immediate questions, please call us at <strong>{{.CompanyInfo.Phone}}</strong>.</p>
        
        <p>Best regards,<br>
        The PropertyHub Team<br>
        {{.CompanyInfo.CompanyName}}</p>
    </div>
    
    <div class="footer">
        <div class="unsubscribe">
            <strong>CAN-SPAM Compliance:</strong><br>
            This email was sent because you inquired about a rental property. You can 
            <a href="{{.UnsubscribeURL}}">unsubscribe from marketing emails</a> at any time.<br>
            <strong>One-click unsubscribe:</strong> <a href="{{.UnsubscribeURL}}">Click here to unsubscribe</a>
        </div>
        
        <div style="margin-top: 20px;">
            <strong>{{.CompanyInfo.CompanyName}}</strong><br>
            {{.CompanyInfo.PhysicalAddress}}<br>
            Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}<br>
            Website: <a href="{{.CompanyInfo.Website}}">{{.CompanyInfo.Website}}</a>
        </div>
        
        <div style="margin-top: 15px; font-size: 11px; color: #666;">
            Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}<br>
            Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244
        </div>
    </div>
</body>
</html>`
}

func (c *CANSPAMEmailService) getPropertyInquiryTextTemplate() string {
	return `{{.CompanyInfo.CompanyName}} - Your Houston Area Rental Specialists

Hi {{.RecipientName}},

Thank you for your inquiry about {{.PropertyAddress}}!

We received your request and one of our rental specialists will contact you within 2 business hours with detailed information about this property, including:

- Complete property details and amenities
- Available showing times  
- Lease terms and pricing
- Application requirements

In the meantime, feel free to browse our other available properties at {{.CompanyInfo.Website}}.

If you have immediate questions, please call us at {{.CompanyInfo.Phone}}.

Best regards,
The PropertyHub Team
{{.CompanyInfo.CompanyName}}

---
CAN-SPAM COMPLIANCE
This email was sent because you inquired about a rental property.
You can unsubscribe from marketing emails at: {{.UnsubscribeURL}}

{{.CompanyInfo.CompanyName}}
{{.CompanyInfo.PhysicalAddress}}
Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}
Website: {{.CompanyInfo.Website}}

Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}
Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244`
}

// Showing Confirmation Templates
func (c *CANSPAMEmailService) getShowingConfirmationHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Showing Confirmation</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #28a745; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .footer { background: #eee; padding: 20px; font-size: 12px; text-align: center; }
        .highlight { background: #e7f3ff; padding: 15px; margin: 15px 0; border-left: 4px solid #007bff; }
        .unsubscribe { margin-top: 20px; padding: 15px; background: #fff3cd; border: 1px solid #ffeaa7; }
    </style>
</head>
<body>
    <div class="header">
        <h1>✅ Showing Confirmed!</h1>
        <p>{{.CompanyInfo.CompanyName}}</p>
    </div>
    
    <div class="content">
        <h2>Hi {{.RecipientName}},</h2>
        
        <p>Your property showing has been confirmed!</p>
        
        <div class="highlight">
            <h3>📍 Showing Details:</h3>
            <strong>Property:</strong> {{.PropertyAddress}}<br>
            <strong>Date & Time:</strong> {{.ShowingTime}}<br>
            <strong>Duration:</strong> Approximately 15-20 minutes
        </div>
        
        <h3>What to Expect:</h3>
        <ul>
            <li>Please arrive 5 minutes early</li>
            <li>Bring a valid photo ID</li>
            <li>Feel free to take photos/videos</li>
            <li>Ask any questions about the property</li>
        </ul>
        
        <h3>Need to Reschedule?</h3>
        <p>Please call us at <strong>{{.CompanyInfo.Phone}}</strong> at least 2 hours before your scheduled showing time.</p>
        
        <p>We look forward to showing you this property!</p>
        
        <p>Best regards,<br>
        The PropertyHub Team<br>
        {{.CompanyInfo.CompanyName}}</p>
    </div>
    
    <div class="footer">
        <div class="unsubscribe">
            <strong>CAN-SPAM Compliance:</strong><br>
            This is a transactional email confirming your property showing appointment. You can 
            <a href="{{.UnsubscribeURL}}">unsubscribe from marketing emails</a> at any time.<br>
            <strong>One-click unsubscribe:</strong> <a href="{{.UnsubscribeURL}}">Click here to unsubscribe</a>
        </div>
        
        <div style="margin-top: 20px;">
            <strong>{{.CompanyInfo.CompanyName}}</strong><br>
            {{.CompanyInfo.PhysicalAddress}}<br>
            Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}<br>
            Website: <a href="{{.CompanyInfo.Website}}">{{.CompanyInfo.Website}}</a>
        </div>
        
        <div style="margin-top: 15px; font-size: 11px; color: #666;">
            Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}<br>
            Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244
        </div>
    </div>
</body>
</html>`
}

func (c *CANSPAMEmailService) getShowingConfirmationTextTemplate() string {
	return `✅ SHOWING CONFIRMED - {{.CompanyInfo.CompanyName}}

Hi {{.RecipientName}},

Your property showing has been confirmed!

📍 SHOWING DETAILS:
Property: {{.PropertyAddress}}
Date & Time: {{.ShowingTime}}
Duration: Approximately 15-20 minutes

WHAT TO EXPECT:
- Please arrive 5 minutes early
- Bring a valid photo ID
- Feel free to take photos/videos  
- Ask any questions about the property

NEED TO RESCHEDULE?
Please call us at {{.CompanyInfo.Phone}} at least 2 hours before your scheduled showing time.

We look forward to showing you this property!

Best regards,
The PropertyHub Team
{{.CompanyInfo.CompanyName}}

---
CAN-SPAM COMPLIANCE
This is a transactional email confirming your property showing appointment.
You can unsubscribe from marketing emails at: {{.UnsubscribeURL}}

{{.CompanyInfo.CompanyName}}
{{.CompanyInfo.PhysicalAddress}}
Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}
Website: {{.CompanyInfo.Website}}

Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}
Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244`
}

// Property Availability Alert Templates  
func (c *CANSPAMEmailService) getPropertyAvailabilityHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Property Available</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #ff6b35; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .footer { background: #eee; padding: 20px; font-size: 12px; text-align: center; }
        .btn { background: #ff6b35; color: white; padding: 12px 25px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 10px 0; }
        .unsubscribe { margin-top: 20px; padding: 15px; background: #fff3cd; border: 1px solid #ffeaa7; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🏠 New Property Available!</h1>
        <p>{{.CompanyInfo.CompanyName}}</p>
    </div>
    
    <div class="content">
        <h2>Hi {{.RecipientName}},</h2>
        
        <p>Great news! We have a new property available that matches your rental criteria:</p>
        
        <h3>📍 {{.PropertyAddress}}</h3>
        
        {{if .CustomData.rent}}<p><strong>Rent:</strong> ${{.CustomData.rent}}/month</p>{{end}}
        {{if .CustomData.bedrooms}}<p><strong>Bedrooms:</strong> {{.CustomData.bedrooms}} | <strong>Bathrooms:</strong> {{.CustomData.bathrooms}}</p>{{end}}
        {{if .CustomData.sqft}}<p><strong>Size:</strong> {{.CustomData.sqft}} sq ft</p>{{end}}
        {{if .CustomData.amenities}}<p><strong>Amenities:</strong> {{.CustomData.amenities}}</p>{{end}}
        
        <p>This property is available for immediate occupancy and matches your search preferences!</p>
        
        <p><a href="{{.CompanyInfo.Website}}/properties/{{.CustomData.property_id}}" class="btn">View Property Details</a></p>
        
        <p><strong>Want to schedule a showing?</strong><br>
        Call us at {{.CompanyInfo.Phone}} or book online to see this property.</p>
        
        <p>Best regards,<br>
        The PropertyHub Team<br>
        {{.CompanyInfo.CompanyName}}</p>
    </div>
    
    <div class="footer">
        <div class="unsubscribe">
            <strong>CAN-SPAM Compliance:</strong><br>
            This marketing email was sent because you subscribed to property alerts. You can 
            <a href="{{.UnsubscribeURL}}">unsubscribe from marketing emails</a> at any time.<br>
            <strong>One-click unsubscribe:</strong> <a href="{{.UnsubscribeURL}}">Click here to unsubscribe</a>
        </div>
        
        <div style="margin-top: 20px;">
            <strong>{{.CompanyInfo.CompanyName}}</strong><br>
            {{.CompanyInfo.PhysicalAddress}}<br>
            Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}<br>
            Website: <a href="{{.CompanyInfo.Website}}">{{.CompanyInfo.Website}}</a>
        </div>
        
        <div style="margin-top: 15px; font-size: 11px; color: #666;">
            Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}<br>
            Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244
        </div>
    </div>
</body>
</html>`
}

func (c *CANSPAMEmailService) getPropertyAvailabilityTextTemplate() string {
	return `🏠 NEW PROPERTY AVAILABLE - {{.CompanyInfo.CompanyName}}

Hi {{.RecipientName}},

Great news! We have a new property available that matches your rental criteria:

📍 {{.PropertyAddress}}

{{if .CustomData.rent}}Rent: ${{.CustomData.rent}}/month{{end}}
{{if .CustomData.bedrooms}}Bedrooms: {{.CustomData.bedrooms}} | Bathrooms: {{.CustomData.bathrooms}}{{end}}
{{if .CustomData.sqft}}Size: {{.CustomData.sqft}} sq ft{{end}}
{{if .CustomData.amenities}}Amenities: {{.CustomData.amenities}}{{end}}

This property is available for immediate occupancy and matches your search preferences!

View Property Details: {{.CompanyInfo.Website}}/properties/{{.CustomData.property_id}}

WANT TO SCHEDULE A SHOWING?
Call us at {{.CompanyInfo.Phone}} or book online to see this property.

Best regards,
The PropertyHub Team
{{.CompanyInfo.CompanyName}}

---
CAN-SPAM COMPLIANCE
This marketing email was sent because you subscribed to property alerts.
You can unsubscribe from marketing emails at: {{.UnsubscribeURL}}

{{.CompanyInfo.CompanyName}}
{{.CompanyInfo.PhysicalAddress}}
Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}
Website: {{.CompanyInfo.Website}}

Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}
Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244`
}

// Application Status Update Templates
func (c *CANSPAMEmailService) getApplicationStatusHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Application Status Update</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #6c757d; color: white; padding: 20px; text-align: center; }
        .content { background: #f9f9f9; padding: 30px; }
        .footer { background: #eee; padding: 20px; font-size: 12px; text-align: center; }
        .status-box { background: #e7f3ff; padding: 15px; margin: 15px 0; border-left: 4px solid #007bff; }
        .unsubscribe { margin-top: 20px; padding: 15px; background: #fff3cd; border: 1px solid #ffeaa7; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📋 Application Update</h1>
        <p>{{.CompanyInfo.CompanyName}}</p>
    </div>
    
    <div class="content">
        <h2>Hi {{.RecipientName}},</h2>
        
        <p>We have an update on your rental application:</p>
        
        <div class="status-box">
            <h3>📍 Application Details:</h3>
            <strong>Property:</strong> {{.PropertyAddress}}<br>
            <strong>Status:</strong> {{.CustomData.status}}<br>
            {{if .CustomData.next_step}}<strong>Next Step:</strong> {{.CustomData.next_step}}<br>{{end}}
            {{if .CustomData.timeline}}<strong>Expected Timeline:</strong> {{.CustomData.timeline}}<br>{{end}}
        </div>
        
        {{if .CustomData.message}}
        <h3>Additional Information:</h3>
        <p>{{.CustomData.message}}</p>
        {{end}}
        
        <p>If you have any questions about your application, please don't hesitate to contact us at {{.CompanyInfo.Phone}}.</p>
        
        <p>Thank you for choosing {{.CompanyInfo.CompanyName}}!</p>
        
        <p>Best regards,<br>
        The PropertyHub Team<br>
        {{.CompanyInfo.CompanyName}}</p>
    </div>
    
    <div class="footer">
        <div class="unsubscribe">
            <strong>CAN-SPAM Compliance:</strong><br>
            This is a transactional email about your rental application. You can 
            <a href="{{.UnsubscribeURL}}">unsubscribe from marketing emails</a> at any time.<br>
            <strong>One-click unsubscribe:</strong> <a href="{{.UnsubscribeURL}}">Click here to unsubscribe</a>
        </div>
        
        <div style="margin-top: 20px;">
            <strong>{{.CompanyInfo.CompanyName}}</strong><br>
            {{.CompanyInfo.PhysicalAddress}}<br>
            Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}<br>
            Website: <a href="{{.CompanyInfo.Website}}">{{.CompanyInfo.Website}}</a>
        </div>
        
        <div style="margin-top: 15px; font-size: 11px; color: #666;">
            Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}<br>
            Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244
        </div>
    </div>
</body>
</html>`
}

func (c *CANSPAMEmailService) getApplicationStatusTextTemplate() string {
	return `📋 APPLICATION UPDATE - {{.CompanyInfo.CompanyName}}

Hi {{.RecipientName}},

We have an update on your rental application:

📍 APPLICATION DETAILS:
Property: {{.PropertyAddress}}
Status: {{.CustomData.status}}
{{if .CustomData.next_step}}Next Step: {{.CustomData.next_step}}{{end}}
{{if .CustomData.timeline}}Expected Timeline: {{.CustomData.timeline}}{{end}}

{{if .CustomData.message}}
ADDITIONAL INFORMATION:
{{.CustomData.message}}
{{end}}

If you have any questions about your application, please don't hesitate to contact us at {{.CompanyInfo.Phone}}.

Thank you for choosing {{.CompanyInfo.CompanyName}}!

Best regards,
The PropertyHub Team
{{.CompanyInfo.CompanyName}}

---
CAN-SPAM COMPLIANCE
This is a transactional email about your rental application.
You can unsubscribe from marketing emails at: {{.UnsubscribeURL}}

{{.CompanyInfo.CompanyName}}
{{.CompanyInfo.PhysicalAddress}}
Phone: {{.CompanyInfo.Phone}} | Email: {{.CompanyInfo.Email}}
Website: {{.CompanyInfo.Website}}

Email sent on {{.SentDate}} | Unsubscribe ID: {{.UnsubscribeID}}
Licensed Real Estate Broker | TREC License #9008008 | Christopher Gross #625244`
}

// ValidateTemplate checks if a template is CAN-SPAM compliant
func (c *CANSPAMEmailService) ValidateTemplate(tpl *CANSPAMTemplate) (bool, []string) {
	var issues []string
	score := 100
	
	// Check for unsubscribe link
	if !strings.Contains(strings.ToLower(tpl.HTMLContent), "unsubscribe") {
		issues = append(issues, "Missing unsubscribe link")
		score -= 30
	}
	
	// Check for physical address
	if !strings.Contains(strings.ToLower(tpl.HTMLContent), "physicaladdress") {
		issues = append(issues, "Missing physical address")
		score -= 25
	}
	
	// Check for clear sender identification
	if !strings.Contains(strings.ToLower(tpl.HTMLContent), "companyinfo") {
		issues = append(issues, "Missing clear sender identification")
		score -= 20
	}
	
	// Check for truthful subject line (basic check)
	if strings.Contains(strings.ToLower(tpl.Subject), "free") && tpl.TemplateType == "marketing" {
		issues = append(issues, "Potentially misleading subject line")
		score -= 15
	}
	
	tpl.ComplianceScore = score
	return len(issues) == 0, issues
}

// UpdateCompanyInfo updates company information for CAN-SPAM compliance
func (c *CANSPAMEmailService) UpdateCompanyInfo(info CompanyInfo) {
	c.companyInfo = info
}