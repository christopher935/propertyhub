package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandlers struct {
	DB *gorm.DB
}

func NewDashboardHandlers(db *gorm.DB) *DashboardHandlers {
	return &DashboardHandlers{DB: db}
}

func (h *DashboardHandlers) GetPropertySummary(c *gin.Context) {
	var totalProperties int64
	var occupiedUnits int64
	var vacantUnits int64

	h.DB.Table("properties").Count(&totalProperties)
	h.DB.Table("properties").Where("status = ?", "leased").Count(&occupiedUnits)
	h.DB.Table("properties").Where("status IN (?)", []string{"active", "available"}).Count(&vacantUnits)

	occupancyRate := float64(0)
	if totalProperties > 0 {
		occupancyRate = (float64(occupiedUnits) / float64(totalProperties)) * 100
	}

	type RecentProperty struct {
		ID            uint      `json:"id"`
		Address       string    `json:"address"`
		City          string    `json:"city"`
		Price         float64   `json:"price"`
		Status        string    `json:"status"`
		Bedrooms      *int      `json:"bedrooms"`
		Bathrooms     *float32  `json:"bathrooms"`
		PropertyType  string    `json:"propertyType"`
		FeaturedImage string    `json:"featuredImage"`
		CreatedAt     time.Time `json:"createdAt"`
	}

	var recentProperties []RecentProperty
	h.DB.Table("properties").
		Select("id, address, city, price, status, bedrooms, bathrooms, property_type, featured_image, created_at").
		Order("created_at DESC").
		Limit(5).
		Find(&recentProperties)

	c.JSON(http.StatusOK, gin.H{
		"totalProperties": totalProperties,
		"occupiedUnits":   occupiedUnits,
		"vacantUnits":     vacantUnits,
		"occupancyRate":   occupancyRate,
		"recentProperties": recentProperties,
	})
}

func (h *DashboardHandlers) GetBookingSummary(c *gin.Context) {
	var activeBookings int64
	var pendingBookings int64
	var todayCheckIns int64
	var todayCheckOuts int64

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	h.DB.Table("bookings").Where("status = ?", "confirmed").Count(&activeBookings)
	h.DB.Table("bookings").Where("status = ?", "pending").Count(&pendingBookings)
	h.DB.Table("bookings").
		Where("status = ? AND showing_date >= ? AND showing_date < ?", "confirmed", today, tomorrow).
		Count(&todayCheckIns)

	type RecentBooking struct {
		ID              uint      `json:"id"`
		ReferenceNumber string    `json:"referenceNumber"`
		PropertyAddress string    `json:"propertyAddress"`
		Name            string    `json:"name"`
		Email           string    `json:"email"`
		Phone           string    `json:"phone"`
		ShowingDate     time.Time `json:"showingDate"`
		Status          string    `json:"status"`
		ShowingType     string    `json:"showingType"`
	}

	var recentBookings []RecentBooking
	h.DB.Table("bookings").
		Select("id, reference_number, property_address, name, email, phone, showing_date, status, showing_type").
		Order("created_at DESC").
		Limit(5).
		Find(&recentBookings)

	c.JSON(http.StatusOK, gin.H{
		"activeBookings":   activeBookings,
		"pendingBookings":  pendingBookings,
		"todayCheckIns":    todayCheckIns,
		"todayCheckOuts":   todayCheckOuts,
		"recentBookings":   recentBookings,
	})
}

func (h *DashboardHandlers) GetRevenueSummary(c *gin.Context) {
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
	currentYearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	type RevenueResult struct {
		Total float64
	}

	var currentMonthRevenue RevenueResult
	h.DB.Table("properties").
		Select("COALESCE(SUM(price), 0) as total").
		Where("status = ? AND updated_at >= ?", "leased", currentMonthStart).
		Scan(&currentMonthRevenue)

	var lastMonthRevenue RevenueResult
	h.DB.Table("properties").
		Select("COALESCE(SUM(price), 0) as total").
		Where("status = ? AND updated_at >= ? AND updated_at < ?", "leased", lastMonthStart, currentMonthStart).
		Scan(&lastMonthRevenue)

	var yearlyRevenue RevenueResult
	h.DB.Table("properties").
		Select("COALESCE(SUM(price), 0) as total").
		Where("status = ? AND updated_at >= ?", "leased", currentYearStart).
		Scan(&yearlyRevenue)

	monthlyChange := float64(0)
	if lastMonthRevenue.Total > 0 {
		monthlyChange = ((currentMonthRevenue.Total - lastMonthRevenue.Total) / lastMonthRevenue.Total) * 100
	}

	var totalProperties int64
	h.DB.Table("properties").Where("status = ?", "leased").Count(&totalProperties)

	avgDailyRate := float64(0)
	if totalProperties > 0 {
		avgDailyRate = currentMonthRevenue.Total / float64(totalProperties) / 30
	}

	revPAR := avgDailyRate

	trendData := []map[string]interface{}{}
	for i := 5; i >= 0; i-- {
		monthStart := currentMonthStart.AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		var monthRevenue RevenueResult
		h.DB.Table("properties").
			Select("COALESCE(SUM(price), 0) as total").
			Where("status = ? AND updated_at >= ? AND updated_at < ?", "leased", monthStart, monthEnd).
			Scan(&monthRevenue)

		trendData = append(trendData, map[string]interface{}{
			"month":   monthStart.Format("Jan 2006"),
			"revenue": monthRevenue.Total,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"monthlyRevenue":  currentMonthRevenue.Total,
		"yearlyRevenue":   yearlyRevenue.Total,
		"monthlyChange":   monthlyChange,
		"yearlyChange":    0,
		"avgDailyRate":    avgDailyRate,
		"revPAR":          revPAR,
		"trendData":       trendData,
	})
}

func (h *DashboardHandlers) GetMarketData(c *gin.Context) {
	type MarketStats struct {
		AvgPrice       float64
		TotalListings  int64
		AvgDaysOnMarket float64
	}

	var stats MarketStats
	h.DB.Table("properties").
		Select("COALESCE(AVG(price), 0) as avg_price, COUNT(*) as total_listings, COALESCE(AVG(days_on_market), 0) as avg_days_on_market").
		Where("status IN (?)", []string{"active", "available"}).
		Scan(&stats)

	var totalProperties int64
	var leasedProperties int64
	h.DB.Table("properties").Count(&totalProperties)
	h.DB.Table("properties").Where("status = ?", "leased").Count(&leasedProperties)

	marketOccupancy := float64(0)
	if totalProperties > 0 {
		marketOccupancy = (float64(leasedProperties) / float64(totalProperties)) * 100
	}

	insights := []map[string]interface{}{}
	
	if stats.AvgPrice > 0 {
		insights = append(insights, map[string]interface{}{
			"type":    "info",
			"message": "Market is active with steady pricing",
		})
	}

	if marketOccupancy > 85 {
		insights = append(insights, map[string]interface{}{
			"type":    "success",
			"message": "High occupancy rate - strong market demand",
		})
	} else if marketOccupancy < 70 {
		insights = append(insights, map[string]interface{}{
			"type":    "warning",
			"message": "Low occupancy - consider pricing adjustments",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"averageRent":      stats.AvgPrice,
		"marketOccupancy":  marketOccupancy,
		"daysOnMarket":     stats.AvgDaysOnMarket,
		"insights":         insights,
		"lastUpdated":      time.Now().Format(time.RFC3339),
	})
}

func (h *DashboardHandlers) GetRecentActivity(c *gin.Context) {
	type Activity struct {
		Type        string    `json:"type"`
		Description string    `json:"description"`
		Timestamp   time.Time `json:"timestamp"`
		EventType   string    `json:"eventType"`
		SessionID   string    `json:"sessionId"`
	}

	var activities []Activity
	h.DB.Table("behavioral_events").
		Select("event_type as type, event_type as event_type, session_id, created_at as timestamp").
		Order("created_at DESC").
		Limit(10).
		Find(&activities)

	for i := range activities {
		switch activities[i].EventType {
		case "property_view":
			activities[i].Description = "User viewed a property"
		case "search":
			activities[i].Description = "User performed a search"
		case "booking_created":
			activities[i].Description = "New booking created"
		case "lead_created":
			activities[i].Description = "New lead generated"
		default:
			activities[i].Description = "Activity: " + activities[i].EventType
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
	})
}

func (h *DashboardHandlers) GetAlerts(c *gin.Context) {
	alerts := []map[string]interface{}{}
	criticalCount := 0

	var hotLeadsCount int64
	h.DB.Table("leads").Where("behavioral_score > ?", 70).Count(&hotLeadsCount)
	if hotLeadsCount > 0 {
		alerts = append(alerts, map[string]interface{}{
			"id":        "hot-leads",
			"title":     "Hot Leads Need Attention",
			"message":   "You have hot leads with high behavioral scores",
			"priority":  "high",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		criticalCount++
	}

	var propertiesNoImages int64
	h.DB.Table("properties").
		Where("(images IS NULL OR images = '{}') AND status IN (?)", []string{"active", "available"}).
		Count(&propertiesNoImages)
	if propertiesNoImages > 0 {
		alerts = append(alerts, map[string]interface{}{
			"id":        "no-images",
			"title":     "Properties Missing Images",
			"message":   "Some active properties don't have images",
			"priority":  "medium",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	var pendingBookings int64
	h.DB.Table("bookings").Where("status = ?", "pending").Count(&pendingBookings)
	if pendingBookings > 0 {
		alerts = append(alerts, map[string]interface{}{
			"id":        "pending-bookings",
			"title":     "Bookings Need Confirmation",
			"message":   "You have pending bookings awaiting confirmation",
			"priority":  "high",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		criticalCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"totalCount":    len(alerts),
		"criticalCount": criticalCount,
		"alerts":        alerts,
	})
}

func (h *DashboardHandlers) GetMaintenanceRequests(c *gin.Context) {
	var openRequests int64
	h.DB.Table("maintenance_requests").
		Where("status IN ?", []string{"open", "in_progress"}).
		Count(&openRequests)

	var emergencyRequests int64
	h.DB.Table("maintenance_requests").
		Where("priority = ? AND status IN ?", "emergency", []string{"open", "in_progress"}).
		Count(&emergencyRequests)

	var highPriorityRequests int64
	h.DB.Table("maintenance_requests").
		Where("priority = ? AND status IN ?", "high", []string{"open", "in_progress"}).
		Count(&highPriorityRequests)

	type RecentMaintenanceRequest struct {
		ID              uint      `json:"id"`
		PropertyAddress string    `json:"propertyAddress"`
		Description     string    `json:"description"`
		Priority        string    `json:"priority"`
		Status          string    `json:"status"`
		Category        string    `json:"category"`
		TenantName      string    `json:"tenantName"`
		SuggestedVendor string    `json:"suggestedVendor"`
		ResponseTime    string    `json:"responseTime"`
		CreatedAt       time.Time `json:"createdAt"`
	}

	var recentRequests []RecentMaintenanceRequest
	h.DB.Table("maintenance_requests").
		Select("id, property_address, description, priority, status, category, tenant_name, suggested_vendor, response_time, created_at").
		Where("status IN ?", []string{"open", "in_progress"}).
		Order("CASE priority WHEN 'emergency' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 END").
		Order("created_at DESC").
		Limit(5).
		Find(&recentRequests)

	var avgResolutionHours float64
	h.DB.Table("maintenance_requests").
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_date - created_at))/3600), 0)").
		Where("status = ? AND completed_date IS NOT NULL", "completed").
		Scan(&avgResolutionHours)

	avgResponseTime := "N/A"
	if avgResolutionHours > 0 {
		if avgResolutionHours < 24 {
			avgResponseTime = fmt.Sprintf("%.1f hours", avgResolutionHours)
		} else {
			avgResponseTime = fmt.Sprintf("%.1f days", avgResolutionHours/24)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"openRequests":    openRequests,
		"urgentRequests":  emergencyRequests + highPriorityRequests,
		"emergencyCount":  emergencyRequests,
		"highPriorityCount": highPriorityRequests,
		"avgResponseTime": avgResponseTime,
		"recentRequests":  recentRequests,
	})
}

func (h *DashboardHandlers) GetLeadSummary(c *gin.Context) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var totalLeads int64
	var newLeads int64
	var hotLeads int64

	h.DB.Table("leads").Count(&totalLeads)
	h.DB.Table("leads").Where("created_at > ?", thirtyDaysAgo).Count(&newLeads)
	h.DB.Table("leads").Where("behavioral_score > ?", 70).Count(&hotLeads)

	var totalConversions int64
	h.DB.Table("applications").Where("status = ?", "approved").Count(&totalConversions)

	conversionRate := float64(0)
	if totalLeads > 0 {
		conversionRate = (float64(totalConversions) / float64(totalLeads)) * 100
	}

	type LeadSource struct {
		Source string `json:"name"`
		Count  int64  `json:"count"`
	}

	var leadSources []LeadSource
	h.DB.Table("leads").
		Select("source as source, COUNT(*) as count").
		Group("source").
		Order("count DESC").
		Limit(5).
		Find(&leadSources)

	c.JSON(http.StatusOK, gin.H{
		"newLeads":       newLeads,
		"hotLeads":       hotLeads,
		"conversionRate": conversionRate,
		"leadSources":    leadSources,
	})
}

func (h *DashboardHandlers) GetUpcomingTasks(c *gin.Context) {
	tasks := []map[string]interface{}{}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var staleLeadsCount int64
	h.DB.Table("leads").
		Where("last_contact_at < ? OR last_contact_at IS NULL", sevenDaysAgo).
		Where("behavioral_score > ?", 50).
		Count(&staleLeadsCount)

	if staleLeadsCount > 0 {
		tasks = append(tasks, map[string]interface{}{
			"title":       "Follow up with stale leads",
			"description": "Multiple leads haven't been contacted in over 7 days",
			"dueDate":     time.Now().Format(time.RFC3339),
			"priority":    "high",
		})
	}

	tomorrow := time.Now().AddDate(0, 0, 1)
	dayAfter := tomorrow.AddDate(0, 0, 1)
	var upcomingShowings int64
	h.DB.Table("bookings").
		Where("showing_date >= ? AND showing_date < ?", tomorrow, dayAfter).
		Where("status = ?", "confirmed").
		Count(&upcomingShowings)

	if upcomingShowings > 0 {
		tasks = append(tasks, map[string]interface{}{
			"title":       "Upcoming property showings",
			"description": "You have showings scheduled for tomorrow",
			"dueDate":     tomorrow.Format(time.RFC3339),
			"priority":    "medium",
		})
	}

	var pendingApplications int64
	h.DB.Table("applications").Where("status = ?", "pending").Count(&pendingApplications)
	if pendingApplications > 0 {
		tasks = append(tasks, map[string]interface{}{
			"title":       "Review pending applications",
			"description": "Applications awaiting review",
			"dueDate":     time.Now().Format(time.RFC3339),
			"priority":    "high",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}
