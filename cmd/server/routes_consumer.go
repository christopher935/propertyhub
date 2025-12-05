package main

import (
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterConsumerRoutes registers all consumer-facing routes
func RegisterConsumerRoutes(r *gin.Engine, h *AllHandlers) {
	// Core website routes
	// Homepage - show 2 featured properties
	r.GET("/", func(c *gin.Context) {
		var properties []models.Property
		h.DB.Where("status = ?", "available").Order("created_at DESC").Limit(2).Find(&properties)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":      "PropertyHub",
			"Properties": properties,
		})
	})
	r.GET("/home", func(c *gin.Context) {
		var properties []models.Property
		h.DB.Where("status = ?", "available").Order("created_at DESC").Limit(2).Find(&properties)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":      "Home",
			"Properties": properties,
		})
	})
	r.GET("/properties", func(c *gin.Context) {
		c.HTML(http.StatusOK, "properties-grid.html", gin.H{"Title": "Properties"})
	})
	r.GET("/saved-properties", func(c *gin.Context) {
		c.HTML(http.StatusOK, "saved-properties.html", gin.H{"Title": "Saved Properties"})
	})
	r.GET("/property-alerts", func(c *gin.Context) {
		c.HTML(http.StatusOK, "property-alerts.html", gin.H{"Title": "Property Alerts"})
	})
	r.GET("/property/:id", func(c *gin.Context) {
		propertyID := c.Param("id")
		var property models.Property
		if err := h.DB.First(&property, propertyID).Error; err != nil {
			c.HTML(http.StatusNotFound, "404.html", gin.H{"Title": "Property Not Found"})
			return
		}

		// Calculate days on market
		daysOnMarket := int(time.Since(property.CreatedAt).Hours() / 24)

		// Get similar properties (same city, similar price range)
		var similarProperties []models.Property
		h.DB.Where("city = ? AND id != ? AND price BETWEEN ? AND ?",
			property.City,
			property.ID,
			property.Price-500,
			property.Price+500,
		).Limit(3).Find(&similarProperties)

		c.HTML(http.StatusOK, "property-detail.html", gin.H{
			"Property":          property,
			"DaysOnMarket":      daysOnMarket,
			"SimilarProperties": similarProperties,
			"ContactPhone":      "(281) 925-7222",
			"ListingAgent":      "Christopher Gross",
			"Agent": gin.H{
				"Name":          "Christopher Gross",
				"Initials":      "CG",
				"LicenseNumber": "0123456",
			},
		})
	})
	r.GET("/book-showing", func(c *gin.Context) {
		c.HTML(http.StatusOK, "book-showing.html", gin.H{"Title": "Book Showing"})
	})
	r.GET("/booking-confirmation", func(c *gin.Context) {
		c.HTML(http.StatusOK, "booking-confirmation.html", gin.H{"Title": "Booking Confirmed"})
	})
	r.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", gin.H{"Title": "Contact"})
	})
	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{"Title": "About"})
	})

	// Booking routes
	r.GET("/booking", func(c *gin.Context) {
		c.HTML(http.StatusOK, "booking.html", gin.H{"Title": "Booking"})
	})

	r.GET("/manage-booking", func(c *gin.Context) {
		c.HTML(http.StatusOK, "manage-booking.html", gin.H{"Title": "Manage Booking"})
	})

	// Error pages
	r.GET("/403", func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "403.html", gin.H{
			"Title": "403 Forbidden",
		})
	})

	r.GET("/404", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"Title": "Page Not Found"})
	})
	r.GET("/500", func(c *gin.Context) {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"Title": "Server Error"})
	})
	r.GET("/503", func(c *gin.Context) {
		c.HTML(http.StatusServiceUnavailable, "503.html", gin.H{"Title": "Service Unavailable"})
	})

	// Legal pages
	r.GET("/terms", func(c *gin.Context) {
		c.HTML(http.StatusOK, "terms-of-service.html", gin.H{"Title": "Terms of Service"})
	})
	r.GET("/privacy", func(c *gin.Context) {
		c.HTML(http.StatusOK, "privacy-policy.html", gin.H{"Title": "Privacy Policy"})
	})

	// Sitemap
	r.GET("/sitemap", func(c *gin.Context) {
		c.HTML(200, "sitemap.html", gin.H{"Title": "Sitemap"})
	})
	r.GET("/unsubscribe/error", func(c *gin.Context) {
		c.HTML(200, "unsubscribe_error.html", gin.H{"Title": "Unsubscribe Error"})
	})
	r.GET("/unsubscribe/success", func(c *gin.Context) {
		c.HTML(200, "unsubscribe_success.html", gin.H{"Title": "Unsubscribe Success"})
	})
	r.GET("/trec-compliance", func(c *gin.Context) {
		c.HTML(200, "trec-compliance.html", gin.H{"Title": "TREC Compliance"})
	})

	// Search
	r.GET("/search", func(c *gin.Context) {
		c.HTML(200, "search-results.html", gin.H{"Title": "Search Results", "CSRFToken": c.GetString("csrf_token")})
	})

	// Authentication routes
	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/login.html", gin.H{"Title": "Login", "CSRFToken": c.GetString("csrf_token")})
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(200, "consumer/pages/register.html", gin.H{"Title": "Create Account", "CSRFToken": c.GetString("csrf_token")})
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
		c.HTML(200, "application-submitted.html", gin.H{"Title": "Application Submitted", "CSRFToken": c.GetString("csrf_token")})
	})
	r.GET("/booking-confirmed", func(c *gin.Context) {
		c.HTML(200, "booking-confirmed.html", gin.H{"Title": "Booking Confirmed", "CSRFToken": c.GetString("csrf_token")})
	})
}
