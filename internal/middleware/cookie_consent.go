package middleware

import (
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
)

// CookieConsentPreferences represents user's cookie consent choices
type CookieConsentPreferences struct {
	Essential  bool   `json:"essential"`
	Analytics  bool   `json:"analytics"`
	Behavioral bool   `json:"behavioral"`
	Marketing  bool   `json:"marketing"`
	Timestamp  string `json:"timestamp"`
}

// HasBehavioralConsent checks if the user has consented to behavioral tracking
// Returns true if consent is granted or not yet requested (opt-in by default for backwards compatibility)
func HasBehavioralConsent(c *gin.Context) bool {
	// Check for consent cookie
	consentCookie, err := c.Cookie("phCookieConsent")
	if err != nil {
		// No consent cookie found - check localStorage via custom header
		// Frontend should send consent preference in header if available
		consentHeader := c.GetHeader("X-Cookie-Consent")
		if consentHeader == "" {
			// No consent information available - default to FALSE for privacy
			// Users must explicitly opt-in
			return false
		}
		
		var prefs CookieConsentPreferences
		if err := json.Unmarshal([]byte(consentHeader), &prefs); err != nil {
			return false
		}
		return prefs.Behavioral
	}
	
	// Parse consent cookie
	var prefs CookieConsentPreferences
	if err := json.Unmarshal([]byte(consentCookie), &prefs); err != nil {
		return false
	}
	
	return prefs.Behavioral
}

// HasAnalyticsConsent checks if the user has consented to analytics tracking
func HasAnalyticsConsent(c *gin.Context) bool {
	consentCookie, err := c.Cookie("phCookieConsent")
	if err != nil {
		consentHeader := c.GetHeader("X-Cookie-Consent")
		if consentHeader == "" {
			return false
		}
		
		var prefs CookieConsentPreferences
		if err := json.Unmarshal([]byte(consentHeader), &prefs); err != nil {
			return false
		}
		return prefs.Analytics
	}
	
	var prefs CookieConsentPreferences
	if err := json.Unmarshal([]byte(consentCookie), &prefs); err != nil {
		return false
	}
	
	return prefs.Analytics
}

// HasMarketingConsent checks if the user has consented to marketing tracking
func HasMarketingConsent(c *gin.Context) bool {
	consentCookie, err := c.Cookie("phCookieConsent")
	if err != nil {
		consentHeader := c.GetHeader("X-Cookie-Consent")
		if consentHeader == "" {
			return false
		}
		
		var prefs CookieConsentPreferences
		if err := json.Unmarshal([]byte(consentHeader), &prefs); err != nil {
			return false
		}
		return prefs.Marketing
	}
	
	var prefs CookieConsentPreferences
	if err := json.Unmarshal([]byte(consentCookie), &prefs); err != nil {
		return false
	}
	
	return prefs.Marketing
}

// SaveConsentCookie saves the user's cookie consent preferences
func SaveConsentCookie(c *gin.Context, prefs CookieConsentPreferences) error {
	prefsJSON, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	
	// Set cookie for 365 days
	c.SetCookie(
		"phCookieConsent",
		string(prefsJSON),
		365*24*60*60, // maxAge in seconds (1 year)
		"/",
		"",
		false, // secure (set to true in production with HTTPS)
		false, // httpOnly (false so JS can read it)
	)
	
	return nil
}

// CookieConsentHandler handles POST /api/cookie-consent
func CookieConsentHandler(c *gin.Context) {
	var prefs CookieConsentPreferences
	if err := c.BindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consent data"})
		return
	}
	
	// Save consent to cookie
	if err := SaveConsentCookie(c, prefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save consent"})
		return
	}
	
	// TODO: Log consent to database for compliance audit trail
	// This would involve creating a consent_log table and storing:
	// - User ID (if authenticated)
	// - IP Address
	// - Timestamp
	// - Consent preferences
	// - User Agent
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cookie consent preferences saved",
	})
}
