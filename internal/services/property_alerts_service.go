package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	
	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

type PropertyAlertsService struct {
	db               *gorm.DB
	emailService     *EmailService
	matchingService  *PropertyMatchingService
}

func NewPropertyAlertsService(db *gorm.DB, emailService *EmailService) *PropertyAlertsService {
	return &PropertyAlertsService{
		db:              db,
		emailService:    emailService,
		matchingService: NewPropertyMatchingService(db),
	}
}

type AlertPreferences struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Email            string    `json:"email" gorm:"index;not null"`
	SessionID        string    `json:"session_id" gorm:"index"`
	MinPrice         float64   `json:"min_price"`
	MaxPrice         float64   `json:"max_price"`
	MinBedrooms      int       `json:"min_bedrooms"`
	MaxBedrooms      int       `json:"max_bedrooms"`
	MinBathrooms     float64   `json:"min_bathrooms"`
	PreferredCities  string    `json:"preferred_cities" gorm:"type:text"`
	PreferredZips    string    `json:"preferred_zips" gorm:"type:text"`
	PropertyTypes    string    `json:"property_types" gorm:"type:text"`
	AlertFrequency   string    `json:"alert_frequency" gorm:"default:'instant'"`
	Active           bool      `json:"active" gorm:"default:true"`
	LastNotified     *time.Time `json:"last_notified"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (AlertPreferences) TableName() string {
	return "alert_preferences"
}

type PropertyAlert struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	PropertyID        uint      `json:"property_id" gorm:"index;not null"`
	AlertPreferenceID uint      `json:"alert_preference_id" gorm:"index;not null"`
	Email             string    `json:"email" gorm:"index;not null"`
	MatchScore        float64   `json:"match_score"`
	Sent              bool      `json:"sent" gorm:"default:false"`
	SentAt            *time.Time `json:"sent_at"`
	Opened            bool      `json:"opened" gorm:"default:false"`
	Clicked           bool      `json:"clicked" gorm:"default:false"`
	CreatedAt         time.Time `json:"created_at"`
}

func (PropertyAlert) TableName() string {
	return "property_alerts"
}

func (s *PropertyAlertsService) ProcessNewProperty(propertyID uint) error {
	log.Printf("🔔 Property Alerts: Processing new property %d", propertyID)
	
	var property models.Property
	if err := s.db.First(&property, propertyID).Error; err != nil {
		return err
	}
	
	var preferences []AlertPreferences
	query := s.db.Where("active = ?", true)
	
	if property.Price > 0 {
		query = query.Where("(min_price = 0 OR min_price <= ?) AND (max_price = 0 OR max_price >= ?)", 
			property.Price, property.Price)
	}
	
	query.Find(&preferences)
	
	log.Printf("📊 Found %d alert preferences to check", len(preferences))
	
	alertsSent := 0
	for _, pref := range preferences {
		if s.propertyMatchesPreferences(property, pref) {
			if err := s.sendPropertyAlert(property, pref); err == nil {
				alertsSent++
			}
		}
	}
	
	log.Printf("✅ Sent %d property alerts for new property %d", alertsSent, propertyID)
	return nil
}

func (s *PropertyAlertsService) propertyMatchesPreferences(property models.Property, pref AlertPreferences) bool {
	if pref.MinBedrooms > 0 && property.Bedrooms != nil && *property.Bedrooms < pref.MinBedrooms {
		return false
	}
	
	if pref.MaxBedrooms > 0 && property.Bedrooms != nil && *property.Bedrooms > pref.MaxBedrooms {
		return false
	}
	
	if pref.MinBathrooms > 0 && property.Bathrooms != nil && *property.Bathrooms < float32(pref.MinBathrooms) {
		return false
	}
	
	if pref.PreferredCities != "" {
		cities := strings.Split(pref.PreferredCities, ",")
		matched := false
		for _, city := range cities {
			if strings.TrimSpace(strings.ToLower(city)) == strings.ToLower(property.City) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	if pref.PreferredZips != "" {
		zips := strings.Split(pref.PreferredZips, ",")
		matched := false
		for _, zip := range zips {
			if strings.TrimSpace(zip) == property.ZipCode {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	if pref.PropertyTypes != "" {
		types := strings.Split(pref.PropertyTypes, ",")
		matched := false
		for _, pType := range types {
			if strings.TrimSpace(strings.ToLower(pType)) == strings.ToLower(property.PropertyType) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	return true
}

func (s *PropertyAlertsService) sendPropertyAlert(property models.Property, pref AlertPreferences) error {
	alert := PropertyAlert{
		PropertyID:        property.ID,
		AlertPreferenceID: pref.ID,
		Email:             pref.Email,
		MatchScore:        85.0,
		CreatedAt:         time.Now(),
	}
	
	if err := s.db.Create(&alert).Error; err != nil {
		return err
	}
	
	subject := fmt.Sprintf("🏠 New Property Alert: %s", property.Address)
	
	bedrooms := "—"
	bathrooms := "—"
	sqft := "—"
	
	if property.Bedrooms != nil {
		bedrooms = fmt.Sprintf("%d", *property.Bedrooms)
	}
	if property.Bathrooms != nil {
		bathrooms = fmt.Sprintf("%.1f", *property.Bathrooms)
	}
	if property.SquareFeet != nil {
		sqft = fmt.Sprintf("%d", *property.SquareFeet)
	}
	
	body := fmt.Sprintf(`
		<h2>New Property Matching Your Criteria</h2>
		<div style="background:#f9fafb;padding:20px;border-radius:8px;margin:20px 0;">
			<h3 style="color:#1e3a8a;margin:0 0 8px 0;">%s</h3>
			<p style="color:#6b7280;margin:0 0 12px 0;">%s, %s %s</p>
			<p style="font-size:28px;font-weight:700;color:#c4a053;margin:0 0 16px 0;">$%s</p>
			<div style="display:flex;gap:16px;margin-bottom:16px;">
				<span>🛏️ %s bed</span>
				<span>🛁 %s bath</span>
				<span>📐 %s sq ft</span>
			</div>
			<p style="color:#374151;margin-bottom:20px;">%s</p>
			<a href="http://209.38.116.238:8080/property/%d" style="display:inline-block;background:#1e3a8a;color:white;padding:12px 24px;text-decoration:none;border-radius:8px;font-weight:600;">View Property Details</a>
		</div>
		<p style="color:#9ca3af;font-size:12px;margin-top:20px;">
			You're receiving this because you signed up for property alerts.
			<a href="http://209.38.116.238:8080/alerts/unsubscribe?email=%s" style="color:#6b7280;">Unsubscribe</a>
		</p>
	`,
		property.Address,
		property.City,
		property.State,
		property.ZipCode,
		fmt.Sprintf("%.0f", property.Price),
		bedrooms,
		bathrooms,
		sqft,
		property.Description,
		property.ID,
		pref.Email,
	)
	
	metadata := map[string]interface{}{
		"property_id":  property.ID,
		"alert_id":     alert.ID,
		"match_score":  alert.MatchScore,
		"campaign_type": "property_alert",
	}
	
	if err := s.emailService.SendEmail(pref.Email, subject, body, metadata); err != nil {
		log.Printf("❌ Failed to send property alert to %s: %v", pref.Email, err)
		return err
	}
	
	s.db.Model(&alert).Updates(map[string]interface{}{
		"sent": true,
		"sent_at": time.Now(),
	})
	
	s.db.Model(&pref).Update("last_notified", time.Now())
	
	log.Printf("✅ Sent property alert to %s for property %d", pref.Email, property.ID)
	return nil
}

func (s *PropertyAlertsService) GetAlertPreferences(email string) (*AlertPreferences, error) {
	var pref AlertPreferences
	err := s.db.Where("email = ?", email).First(&pref).Error
	return &pref, err
}

func (s *PropertyAlertsService) SaveAlertPreferences(pref *AlertPreferences) error {
	if pref.ID == 0 {
		return s.db.Create(pref).Error
	}
	return s.db.Save(pref).Error
}

func (s *PropertyAlertsService) UnsubscribeFromAlerts(email string) error {
	return s.db.Model(&AlertPreferences{}).
		Where("email = ?", email).
		Update("active", false).Error
}
