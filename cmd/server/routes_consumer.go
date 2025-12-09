package main

import (
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/config"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"

	"github.com/gin-gonic/gin"
)

func decryptPropertyAddresses(properties []models.Property, em *security.EncryptionManager) []gin.H {
	result := make([]gin.H, len(properties))
	for i, p := range properties {
		addr := p.Address
		if em != nil {
			if decrypted, err := em.Decrypt(p.Address); err == nil {
				addr = decrypted
			}
		}
		result[i] = gin.H{
			"ID": p.ID, "Address": addr, "City": p.City, "State": p.State,
			"ZipCode": p.ZipCode, "Price": p.Price, "Bedrooms": p.Bedrooms,
			"Bathrooms": p.Bathrooms, "SquareFeet": p.SquareFeet,
			"FeaturedImage": p.FeaturedImage, "Status": p.Status,
			"PropertyType": p.PropertyType, "Description": p.Description,
			"MLSId": p.MLSId, "YearBuilt": p.YearBuilt, "CreatedAt": p.CreatedAt,
		}
	}
	return result
}

// RegisterConsumerRoutes registers all consumer-facing routes
func RegisterConsumerRoutes(r *gin.Engine, h *AllHandlers, cfg *config.Config) {
	// Core website routes
	// Homepage - show 2 featured properties
	r.GET("/", func(c *gin.Context) {
		var properties []models.Property
		h.DB.Where("status = ?", "available").Order("created_at DESC").Limit(2).Find(&properties)
		decryptedProperties := decryptPropertyAddresses(properties, h.EncryptionManager)
		c.HTML(http.StatusOK, "consumer/pages/index.html", gin.H{
			"Title":      "PropertyHub",
			"Properties": decryptedProperties,
		})
	})
	r.GET("/home", func(c *gin.Context) {
		var properties []models.Property
		h.DB.Where("status = ?", "available").Order("created_at DESC").Limit(2).Find(&properties)
		decryptedProperties := decryptPropertyAddresses(properties, h.EncryptionManager)
		c.HTML(http.StatusOK, "consumer/pages/index.html", gin.H{
			"Title":      "Home",
			"Properties": decryptedProperties,
		})
	})
	r.GET("/properties", func(c *gin.Context) {
		var properties []models.Property
		h.DB.Where("status = ?", "available").Order("created_at DESC").Limit(50).Find(&properties)
		decryptedProperties := decryptPropertyAddresses(properties, h.EncryptionManager)
		c.HTML(http.StatusOK, "consumer/pages/properties-grid.html", gin.H{
			"Title":      "Properties",
			"Properties": decryptedProperties,
			"CSRFToken":  c.GetString("csrf_token"),
		})
	})
	r.GET("/saved-properties", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/saved-properties.html", gin.H{
			"Title":     "Saved Properties",
			"CSRFToken": c.GetString("csrf_token"),
		})
	})
	r.GET("/property-alerts", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/property-alerts.html", gin.H{
			"Title":     "Property Alerts",
			"CSRFToken": c.GetString("csrf_token"),
		})
	})
	r.GET("/property/:id", func(c *gin.Context) {
		propertyID := c.Param("id")
		var property models.Property
		if err := h.DB.First(&property, propertyID).Error; err != nil {
			c.HTML(http.StatusNotFound, "errors/pages/404.html", gin.H{"Title": "Property Not Found"})
			return
		}

		decryptedAddress := property.Address
		if h.EncryptionManager != nil {
			if addr, err := h.EncryptionManager.Decrypt(property.Address); err == nil {
				decryptedAddress = addr
			}
		}

		daysOnMarket := int(time.Since(property.CreatedAt).Hours() / 24)

		var similarProperties []models.Property
		h.DB.Where("city = ? AND id != ? AND price BETWEEN ? AND ?",
			property.City,
			property.ID,
			property.Price-500,
			property.Price+500,
		).Limit(3).Find(&similarProperties)

		c.HTML(http.StatusOK, "consumer/pages/property-detail.html", gin.H{
			"Property":          property,
			"PropertyAddress":   decryptedAddress,
			"DaysOnMarket":      daysOnMarket,
			"SimilarProperties": similarProperties,
			"ContactPhone":      "(281) 925-7222",
			"ListingAgent":      "Christopher Gross",
			"CSRFToken":         c.GetString("csrf_token"),
			"Agent": gin.H{
				"Name":          "Christopher Gross",
				"Initials":      "CG",
				"LicenseNumber": "0123456",
			},
		})
	})
	r.GET("/book-showing", func(c *gin.Context) {
		recaptchaSiteKey := ""
		if cfg != nil && cfg.RecaptchaSiteKey != "" {
			recaptchaSiteKey = cfg.RecaptchaSiteKey
		}
		c.HTML(http.StatusOK, "consumer/pages/book-showing.html", gin.H{
			"Title":           "Book Showing",
			"RecaptchaSiteKey": recaptchaSiteKey,
			"CSRFToken":       c.GetString("csrf_token"),
		})
	})
	r.GET("/booking-confirmation", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/booking-confirmation.html", gin.H{"Title": "Booking Confirmed"})
	})
	r.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/contact.html", gin.H{
			"Title":     "Contact",
			"CSRFToken": c.GetString("csrf_token"),
		})
	})
	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/about.html", gin.H{
			"Title":     "About",
			"CSRFToken": c.GetString("csrf_token"),
		})
	})

	r.GET("/booking", func(c *gin.Context) {
		var properties []models.Property
		h.DB.Where("status = ?", "available").Order("created_at DESC").Find(&properties)
		decryptedProperties := decryptPropertyAddresses(properties, h.EncryptionManager)
		c.HTML(http.StatusOK, "consumer/pages/booking.html", gin.H{
			"Title":      "Booking",
			"Properties": decryptedProperties,
			"CSRFToken":  c.GetString("csrf_token"),
		})
	})

	r.GET("/manage-booking", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/manage-booking.html", gin.H{
			"Title":     "Manage Booking",
			"CSRFToken": c.GetString("csrf_token"),
		})
	})

	// Error pages
	r.GET("/403", func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "errors/pages/403.html", gin.H{
			"Title": "403 Forbidden",
		})
	})

	r.GET("/404", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "errors/pages/404.html", gin.H{"Title": "Page Not Found"})
	})
	r.GET("/500", func(c *gin.Context) {
		c.HTML(http.StatusInternalServerError, "errors/pages/500.html", gin.H{"Title": "Server Error"})
	})
	r.GET("/503", func(c *gin.Context) {
		c.HTML(http.StatusServiceUnavailable, "errors/pages/503.html", gin.H{"Title": "Service Unavailable"})
	})

	// Legal pages
	r.GET("/terms", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/terms-of-service.html", gin.H{"Title": "Terms of Service"})
	})
	r.GET("/privacy", func(c *gin.Context) {
		c.HTML(http.StatusOK, "consumer/pages/privacy-policy.html", gin.H{"Title": "Privacy Policy"})
	})

	// Sitemap
	r.GET("/sitemap", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/sitemap.html", gin.H{"Title": "Sitemap"})
	})
	r.GET("/unsubscribe/error", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/unsubscribe_error.html", gin.H{"Title": "Unsubscribe Error"})
	})
	r.GET("/unsubscribe/success", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/unsubscribe_success.html", gin.H{"Title": "Unsubscribe Success"})
	})
	r.GET("/trec-compliance", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/trec-compliance.html", gin.H{"Title": "TREC Compliance"})
	})

	// Search
	r.GET("/search", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/search-results.html", gin.H{"Title": "Search Results", "CSRFToken": c.GetString("csrf_token")})
	})

	// Authentication routes
	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/login.html", gin.H{"Title": "Login", "CSRFToken": c.GetString("csrf_token")})
	})
	r.POST("/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")
		_ = c.PostForm("remember")

		if email == "" || password == "" {
			c.HTML(http.StatusBadRequest, "consumer/pages/login.html", gin.H{
				"Title":      "Login",
				"CSRFToken":  c.GetString("csrf_token"),
				"Error":      "Email and password are required",
				"Email":      email,
			})
			return
		}

		c.HTML(http.StatusOK, "consumer/pages/login.html", gin.H{
			"Title":      "Login",
			"CSRFToken":  c.GetString("csrf_token"),
			"Error":      "Authentication feature coming soon. Please check back later.",
			"Email":      email,
		})
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/register.html", gin.H{"Title": "Create Account", "CSRFToken": c.GetString("csrf_token")})
	})
	r.POST("/register", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")
		firstName := c.PostForm("first_name")
		lastName := c.PostForm("last_name")
		phone := c.PostForm("phone")

		if email == "" || password == "" {
			c.HTML(http.StatusBadRequest, "consumer/pages/register.html", gin.H{
				"Title":     "Create Account",
				"CSRFToken": c.GetString("csrf_token"),
				"Error":     "Email and password are required",
				"FormData": gin.H{
					"email":      email,
					"first_name": firstName,
					"last_name":  lastName,
					"phone":      phone,
				},
			})
			return
		}

		c.HTML(http.StatusOK, "consumer/pages/register.html", gin.H{
			"Title":     "Create Account",
			"CSRFToken": c.GetString("csrf_token"),
			"Error":     "Registration feature coming soon. Please check back later.",
			"FormData": gin.H{
				"email":      email,
				"first_name": firstName,
				"last_name":  lastName,
				"phone":      phone,
			},
		})
	})
	r.GET("/forgot-password", func(c *gin.Context) {
		c.HTML(200, "auth/pages/forgot-password.html", gin.H{"Title": "Reset Password", "CSRFToken": c.GetString("csrf_token")})
	})
	r.GET("/reset-password", func(c *gin.Context) {
		c.HTML(200, "auth/pages/reset-password.html", gin.H{"Title": "Reset Password", "CSRFToken": c.GetString("csrf_token")})
	})
	r.GET("/password-reset-success", func(c *gin.Context) {
		c.HTML(200, "auth/pages/password-reset-success.html", gin.H{"Title": "Password Reset Successful", "CSRFToken": c.GetString("csrf_token")})
	})
	r.GET("/email-verification", func(c *gin.Context) {
		c.HTML(200, "auth/pages/email-verification.html", gin.H{"Title": "Verify Email", "CSRFToken": c.GetString("csrf_token")})
	})

	// Success pages
	r.GET("/application-submitted", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/application-submitted.html", gin.H{"Title": "Application Submitted", "CSRFToken": c.GetString("csrf_token")})
	})
	r.GET("/booking-confirmed", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/booking-confirmed.html", gin.H{"Title": "Booking Confirmed", "CSRFToken": c.GetString("csrf_token")})
	})
}
