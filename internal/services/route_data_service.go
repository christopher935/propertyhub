package services

import (
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"gorm.io/gorm"
)

// RouteDataService provides centralized data access for all route handlers
// This service ensures consistent data queries across the application
type RouteDataService struct {
	DB                *gorm.DB
	EncryptionManager *security.EncryptionManager
}

// NewRouteDataService creates a new route data service
func NewRouteDataService(db *gorm.DB, encryptionManager *security.EncryptionManager) *RouteDataService {
	return &RouteDataService{
		DB:                db,
		EncryptionManager: encryptionManager,
	}
}

// DecryptPropertyAddress decrypts a property's address for display
func (s *RouteDataService) DecryptPropertyAddress(property *models.Property) string {
	if s.EncryptionManager == nil {
		return string(property.Address) // Return as-is if no encryption manager
	}
	decrypted, err := s.EncryptionManager.Decrypt(property.Address)
	if err != nil {
		return "[Address Unavailable]" // Fallback on error
	}
	return decrypted
}

// PropertyWithDecryptedAddress is a helper struct for consumer-facing display
type PropertyWithDecryptedAddress struct {
	models.Property
	DecryptedAddress string `json:"address"`
}

// DecryptPropertiesForDisplay decrypts addresses for a slice of properties
func (s *RouteDataService) DecryptPropertiesForDisplay(properties []models.Property) []PropertyWithDecryptedAddress {
	result := make([]PropertyWithDecryptedAddress, len(properties))
	for i, prop := range properties {
		result[i] = PropertyWithDecryptedAddress{
			Property:         prop,
			DecryptedAddress: s.DecryptPropertyAddress(&prop),
		}
	}
	return result
}

// ============================================================================
// PROPERTY METHODS
// ============================================================================

// GetAllProperties returns all properties with optional filters
func (s *RouteDataService) GetAllProperties() ([]models.Property, error) {
	var properties []models.Property
	err := s.DB.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&properties).Error
	return properties, err
}

// GetAvailableProperties returns only available properties for public listing
func (s *RouteDataService) GetAvailableProperties() ([]models.Property, error) {
	var properties []models.Property
	err := s.DB.Where("status IN ? AND deleted_at IS NULL",
		[]string{"https://schema.org/InStock", "active"}).
		Order("created_at DESC").
		Find(&properties).Error
	return properties, err
}

// GetPropertyByID returns a specific property by ID
func (s *RouteDataService) GetPropertyByID(id uint) (*models.Property, error) {
	var property models.Property
	err := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&property).Error
	if err != nil {
		return nil, err
	}
	return &property, nil
}

// SearchProperties searches properties by query string
func (s *RouteDataService) SearchProperties(query string) ([]models.Property, error) {
	var properties []models.Property
	searchPattern := "%" + query + "%"
	err := s.DB.Where("(address LIKE ? OR city LIKE ? OR zip_code LIKE ?) AND status IN ? AND deleted_at IS NULL",
		searchPattern, searchPattern, searchPattern,
		[]string{"https://schema.org/InStock", "active"}).
		Order("created_at DESC").
		Find(&properties).Error
	return properties, err
}

// ============================================================================
// BOOKING METHODS
// ============================================================================

// GetAllBookings returns all bookings
func (s *RouteDataService) GetAllBookings() ([]models.Booking, error) {
	var bookings []models.Booking
	err := s.DB.Order("showing_date DESC").Find(&bookings).Error
	return bookings, err
}

// GetBookingByID returns a specific booking by ID
func (s *RouteDataService) GetBookingByID(id uint) (*models.Booking, error) {
	var booking models.Booking
	err := s.DB.Preload("Property").First(&booking, id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// GetBookingByReference returns a booking by reference number
func (s *RouteDataService) GetBookingByReference(ref string) (*models.Booking, error) {
	var booking models.Booking
	err := s.DB.Where("reference_number = ?", ref).Preload("Property").First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// GetUserBookings returns all bookings for a specific user (by email)
func (s *RouteDataService) GetUserBookings(email string) ([]models.Booking, error) {
	var bookings []models.Booking
	err := s.DB.Where("email = ?", email).
		Order("showing_date DESC").
		Preload("Property").
		Find(&bookings).Error
	return bookings, err
}

// GetFeaturedProperties returns featured properties for homepage
func (s *RouteDataService) GetFeaturedProperties(limit int) ([]models.Property, error) {
	var properties []models.Property
	err := s.DB.Where("status IN ? AND deleted_at IS NULL",
		[]string{"https://schema.org/InStock", "active"}).
		Order("created_at DESC").
		Limit(limit).
		Find(&properties).Error
	return properties, err
}

// GetSimilarProperties returns properties similar to the given property
func (s *RouteDataService) GetSimilarProperties(propertyID uint, city string, price float64) ([]models.Property, error) {
	var properties []models.Property
	priceMin := price - 500
	priceMax := price + 500
	err := s.DB.Where("city = ? AND id != ? AND price BETWEEN ? AND ? AND status IN ? AND deleted_at IS NULL",
		city, propertyID, priceMin, priceMax,
		[]string{"https://schema.org/InStock", "active"}).
		Limit(3).
		Find(&properties).Error
	return properties, err
}

// ============================================================================
// DASHBOARD METHODS
// ============================================================================

// RouteDashboardMetrics holds dashboard statistics
type RouteDashboardMetrics struct {
	TotalProperties     int64
	AvailableProperties int64
	TotalBookings       int64
	PendingBookings     int64
	TotalApplications   int64
	PendingApplications int64
	TotalLeads          int64
	RecentProperties    []models.Property
	RecentBookings      []models.Booking
	TotalRevenue        float64
	MonthlyRevenue      float64
}

// GetDashboardMetrics returns comprehensive dashboard statistics
func (s *RouteDataService) GetDashboardMetrics() (*RouteDashboardMetrics, error) {
	metrics := &RouteDashboardMetrics{}

	// Count properties
	s.DB.Model(&models.Property{}).Where("deleted_at IS NULL").Count(&metrics.TotalProperties)
	s.DB.Model(&models.Property{}).Where("status IN ? AND deleted_at IS NULL",
		[]string{"https://schema.org/InStock", "active"}).Count(&metrics.AvailableProperties)

	// Count bookings
	s.DB.Model(&models.Booking{}).Count(&metrics.TotalBookings)
	s.DB.Model(&models.Booking{}).Where("status = ?", "scheduled").Count(&metrics.PendingBookings)

	// Get recent properties
	s.DB.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(5).
		Find(&metrics.RecentProperties)

	// Get recent bookings
	s.DB.Order("created_at DESC").
		Limit(5).
		Preload("Property").
		Find(&metrics.RecentBookings)

	return metrics, nil
}

// ============================================================================
// TEAM METHODS
// ============================================================================

// GetAllLeads returns all leads (placeholder - table may not exist yet)
func (s *RouteDataService) GetAllLeads() ([]interface{}, error) {
	// Return empty for now - leads table may not exist
	return []interface{}{}, nil
}

// GetAllTeamMembers returns all admin users/team members
func (s *RouteDataService) GetAllTeamMembers() ([]interface{}, error) {
	var users []models.AdminUser
	err := s.DB.Where("active = ?", true).
		Order("created_at DESC").
		Find(&users).Error
	// Convert to []interface{}
	result := make([]interface{}, len(users))
	for i, u := range users {
		result[i] = u
	}
	return result, err
}

// GetTeamMemberByID returns a specific team member by ID
func (s *RouteDataService) GetTeamMemberByID(id string) (*models.AdminUser, error) {
	var user models.AdminUser
	err := s.DB.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// AgentStats holds statistics for a specific agent
type AgentStats struct {
	TotalProperties    int64
	ActiveListings     int64
	TotalBookings      int64
	CompletedBookings  int64
	TotalCommissions   float64
	MonthlyCommissions float64
}

// GetAgentStats returns statistics for a specific agent
func (s *RouteDataService) GetAgentStats(agentID uint) (*AgentStats, error) {
	stats := &AgentStats{}

	// Count properties by agent
	s.DB.Model(&models.Property{}).
		Where("listing_agent_id = ? AND deleted_at IS NULL", agentID).
		Count(&stats.TotalProperties)

	s.DB.Model(&models.Property{}).
		Where("listing_agent_id = ? AND status IN ? AND deleted_at IS NULL",
			agentID, []string{"https://schema.org/InStock", "active"}).
		Count(&stats.ActiveListings)

	// Count bookings (would need agent_id field in bookings table)
	s.DB.Model(&models.Booking{}).Count(&stats.TotalBookings)
	s.DB.Model(&models.Booking{}).Where("status = ?", "completed").Count(&stats.CompletedBookings)

	return stats, nil
}

// ============================================================================
// BUSINESS INTELLIGENCE METHODS
// ============================================================================

// BIMetrics holds business intelligence metrics
type BIMetrics struct {
	TotalRevenue          float64
	MonthlyRevenue        float64
	AveragePropertyPrice  float64
	TotalProperties       int64
	AvailableProperties   int64
	SoldProperties        int64
	TotalBookings         int64
	CompletedBookings     int64
	ConversionRate        float64
	AverageTimeToClose    float64
	TopPerformingAgents   []AgentPerformance
	PropertyTypeBreakdown map[string]int64
	CityBreakdown         map[string]int64
}

// AgentPerformance holds performance metrics for an agent
type AgentPerformance struct {
	AgentName      string
	TotalListings  int64
	SoldListings   int64
	TotalRevenue   float64
	ConversionRate float64
}

// GetBIMetrics returns comprehensive business intelligence metrics
func (s *RouteDataService) GetBIMetrics() (*BIMetrics, error) {
	metrics := &BIMetrics{}

	// Property counts
	s.DB.Model(&models.Property{}).Where("deleted_at IS NULL").Count(&metrics.TotalProperties)
	s.DB.Model(&models.Property{}).Where("status IN ? AND deleted_at IS NULL",
		[]string{"https://schema.org/InStock", "active"}).Count(&metrics.AvailableProperties)
	s.DB.Model(&models.Property{}).Where("status = ?", "https://schema.org/SoldOut").Count(&metrics.SoldProperties)

	// Booking counts
	s.DB.Model(&models.Booking{}).Count(&metrics.TotalBookings)
	s.DB.Model(&models.Booking{}).Where("status = ?", "completed").Count(&metrics.CompletedBookings)

	// Calculate conversion rate
	if metrics.TotalBookings > 0 {
		metrics.ConversionRate = (float64(metrics.CompletedBookings) / float64(metrics.TotalBookings)) * 100
	}

	// Average property price
	s.DB.Model(&models.Property{}).
		Where("deleted_at IS NULL AND price > 0").
		Select("AVG(price)").
		Row().Scan(&metrics.AveragePropertyPrice)

	// Property type breakdown
	metrics.PropertyTypeBreakdown = make(map[string]int64)
	rows, _ := s.DB.Model(&models.Property{}).
		Select("property_type, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("property_type").
		Rows()
	defer rows.Close()

	for rows.Next() {
		var propType string
		var count int64
		rows.Scan(&propType, &count)
		if propType != "" {
			metrics.PropertyTypeBreakdown[propType] = count
		}
	}

	// City breakdown
	metrics.CityBreakdown = make(map[string]int64)
	cityRows, _ := s.DB.Model(&models.Property{}).
		Select("city, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("city").
		Rows()
	defer cityRows.Close()

	for cityRows.Next() {
		var city string
		var count int64
		cityRows.Scan(&city, &count)
		if city != "" {
			metrics.CityBreakdown[city] = count
		}
	}

	return metrics, nil
}

// ============================================================================
// SETTINGS METHODS
// ============================================================================

// ContactInfo holds contact information
type ContactInfo struct {
	Phone       string
	Email       string
	Address     string
	OfficeHours string
}

// GetContactInfo returns contact information
func (s *RouteDataService) GetContactInfo() *ContactInfo {
	return &ContactInfo{
		Phone:       "(281) 925-7222",
		Email:       "info@landlordsoftexas.com",
		Address:     "Houston, TX",
		OfficeHours: "Monday - Friday: 9:00 AM - 6:00 PM",
	}
}

// GetCompanyInfo returns company information as a map
func (s *RouteDataService) GetCompanyInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Landlords of Texas, LLC",
		"description": "Premier real estate services in Houston, Texas",
		"founded":     "2020",
		"teamSize":    5,
	}
}

// ============================================================================
// UTILITY METHODS
// ============================================================================

// GetSitemap returns all public routes for sitemap
func (s *RouteDataService) GetSitemap() []map[string]string {
	return []map[string]string{
		{"url": "/", "priority": "1.0"},
		{"url": "/properties", "priority": "0.9"},
		{"url": "/about", "priority": "0.7"},
		{"url": "/contact", "priority": "0.7"},
		{"url": "/booking", "priority": "0.8"},
		{"url": "/login", "priority": "0.5"},
		{"url": "/register", "priority": "0.5"},
	}
}

// GetRecentActivity returns recent system activity
func (s *RouteDataService) GetRecentActivity() []map[string]interface{} {
	activity := []map[string]interface{}{}

	// Get recent properties
	var recentProps []models.Property
	s.DB.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(5).
		Find(&recentProps)

	for _, prop := range recentProps {
		activity = append(activity, map[string]interface{}{
			"type":      "property",
			"action":    "added",
			"item":      prop.Address,
			"timestamp": prop.CreatedAt,
		})
	}

	// Get recent bookings
	var recentBookings []models.Booking
	s.DB.Order("created_at DESC").
		Limit(5).
		Find(&recentBookings)

	for _, booking := range recentBookings {
		activity = append(activity, map[string]interface{}{
			"type":      "booking",
			"action":    "created",
			"item":      booking.ReferenceNumber,
			"timestamp": booking.ShowingDate,
		})
	}

	return activity
}
