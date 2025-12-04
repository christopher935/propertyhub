package services

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type DashboardStatsService struct {
	db           *gorm.DB
	spiderwebAI  *SpiderwebAIOrchestrator
	cache        *IntelligenceCacheService
}

func NewDashboardStatsService(
	db *gorm.DB,
	spiderwebAI *SpiderwebAIOrchestrator,
	cache *IntelligenceCacheService,
) *DashboardStatsService {
	return &DashboardStatsService{
		db:          db,
		spiderwebAI: spiderwebAI,
		cache:       cache,
	}
}

func (dss *DashboardStatsService) GetLiveStats() (map[string]interface{}, error) {
	var activeUsers int64
	var unreadMessages int64
	var currentShowings int64
	
	fifteenMinutesAgo := time.Now().Add(-15 * time.Minute)
	dss.db.Table("users").
		Where("last_active_at > ?", fifteenMinutesAgo).
		Count(&activeUsers)
	
	dss.db.Table("messages").
		Where("read_at IS NULL").
		Count(&unreadMessages)
	
	now := time.Now()
	dss.db.Table("bookings").
		Where("start_time <= ? AND end_time >= ? AND status = ?", now, now, "confirmed").
		Count(&currentShowings)
	
	return map[string]interface{}{
		"active_users":     activeUsers,
		"unread_messages":  unreadMessages,
		"current_showings": currentShowings,
		"timestamp":        time.Now(),
	}, nil
}

func (dss *DashboardStatsService) GetHotStats() (map[string]interface{}, error) {
	if dss.cache != nil && dss.cache.IsAvailable() {
		cached, err := dss.cache.GetDashboardHot()
		if err == nil {
			log.Println("🎯 Dashboard hot stats: Cache HIT")
			return cached, nil
		}
		log.Println("❌ Dashboard hot stats: Cache MISS - Computing...")
	}
	
	return dss.computeHotStats()
}

func (dss *DashboardStatsService) computeHotStats() (map[string]interface{}, error) {
	var pendingBookings int64
	var pendingApplications int64
	
	dss.db.Table("bookings").Where("status = ?", "pending").Count(&pendingBookings)
	dss.db.Table("application_numbers").Where("status = ?", "submitted").Count(&pendingApplications)
	
	topOpportunities := []map[string]interface{}{}
	behavioralScores := map[string]interface{}{}
	
	if dss.spiderwebAI != nil {
		intelligence, err := dss.spiderwebAI.GetDashboardIntelligence()
		if err == nil {
			if opps, ok := intelligence["top_opportunities"].([]Opportunity); ok {
				for _, opp := range opps {
					topOpportunities = append(topOpportunities, map[string]interface{}{
						"id": opp.ID, "type": opp.Type, "priority": opp.Priority,
						"lead_name": opp.LeadName, "lead_email": opp.LeadEmail,
						"lead_id": opp.LeadID, "property_address": opp.PropertyAddress,
						"context": opp.Context, "insight": opp.Insight,
						"conversion_probability": opp.ConversionProbability,
						"urgency_score": opp.UrgencyScore,
						"revenue_estimate": opp.RevenueEstimate,
						"action_sequence": opp.ActionSequence,
						"detected_at": opp.DetectedAt, "expires_at": opp.ExpiresAt,
					})
				}
			}
		}
	}
	var avgScore float64
	var highScoreCount int64
	dss.db.Table("behavioral_scores").
		Select("COALESCE(AVG(composite_score), 0)").
		Scan(&avgScore)
	dss.db.Table("behavioral_scores").
		Where("composite_score > ?", 70).
		Count(&highScoreCount)
	behavioralScores = map[string]interface{}{
		"average_score":    avgScore,
		"high_score_count": highScoreCount,
	}
	hotStats := map[string]interface{}{
		"pending_bookings":     pendingBookings,
		"pending_applications": pendingApplications,
		"top_opportunities":    topOpportunities,
		"behavioral_scores":    behavioralScores,
		"timestamp":            time.Now(),
	}
	if dss.cache != nil && dss.cache.IsAvailable() {
		if err := dss.cache.SetDashboardHot(hotStats); err != nil {
			log.Printf("⚠️ Failed to cache dashboard hot stats: %v", err)
		} else {
			log.Println("✅ Dashboard hot stats cached (5min TTL)")
		}
	}
	return hotStats, nil
}

func (dss *DashboardStatsService) GetWarmStats() (map[string]interface{}, error) {
	if dss.cache != nil && dss.cache.IsAvailable() {
		cached, err := dss.cache.GetDashboardWarm()
		if err == nil {
			log.Println("🎯 Dashboard warm stats: Cache HIT")
			return cached, nil
		}
	}
	return dss.computeWarmStats()
}

func (dss *DashboardStatsService) computeWarmStats() (map[string]interface{}, error) {
	var totalLeads int64
	var convertedLeads int64
	
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	dss.db.Table("contacts").Where("created_at > ?", thirtyDaysAgo).Count(&totalLeads)
	dss.db.Table("contacts").Where("created_at > ? AND status = ?", thirtyDaysAgo, "converted").Count(&convertedLeads)
	
	conversionRate := float64(0)
	if totalLeads > 0 {
		conversionRate = (float64(convertedLeads) / float64(totalLeads)) * 100
	}
	
	type SourceCount struct {
		Source string
		Count  int64
	}
	
	var sources []SourceCount
	dss.db.Table("contacts").
		Select("source, COUNT(*) as count").
		Where("created_at > ?", thirtyDaysAgo).
		Group("source").Order("count DESC").Limit(5).Find(&sources)
	
	leadSources := []map[string]interface{}{}
	for _, s := range sources {
		leadSources = append(leadSources, map[string]interface{}{"source": s.Source, "count": s.Count})
	}
	
	var totalVisitors int64
	var totalApplications int64
	var totalConversions int64
	
	dss.db.Table("behavioral_events").
		Where("created_at > ?", thirtyDaysAgo).
		Distinct("session_id").Count(&totalVisitors)
	
	dss.db.Table("application_numbers").Where("created_at > ?", thirtyDaysAgo).Count(&totalApplications)
	dss.db.Table("application_numbers").Where("created_at > ? AND status = ?", thirtyDaysAgo, "approved").Count(&totalConversions)
	
	funnelMetrics := map[string]interface{}{
		"visitors": totalVisitors, "leads": totalLeads,
		"applications": totalApplications, "conversions": totalConversions,
	}
	
	warmStats := map[string]interface{}{
		"total_leads": totalLeads, "converted_leads": convertedLeads,
		"conversion_rate": conversionRate, "lead_sources": leadSources,
		"funnel_metrics": funnelMetrics, "timestamp": time.Now(),
	}
	
	if dss.cache != nil && dss.cache.IsAvailable() {
		if err := dss.cache.SetDashboardWarm(warmStats); err != nil {
			log.Printf("⚠️ Failed to cache: %v", err)
		} else {
			log.Println("✅ Dashboard warm stats cached (1hr TTL)")
		}
	}
	
	return warmStats, nil
}

func (dss *DashboardStatsService) GetDailyStats() (map[string]interface{}, error) {
	if dss.cache != nil && dss.cache.IsAvailable() {
		cached, err := dss.cache.GetDashboardDaily()
		if err == nil {
			log.Println("🎯 Dashboard daily stats: Cache HIT")
			return cached, nil
		}
	}
	return dss.computeDailyStats()
}

func (dss *DashboardStatsService) computeDailyStats() (map[string]interface{}, error) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	
	var applicationsToday int64
	var applicationsYesterday int64
	var revenueToday float64
	var propertiesListed int64
	var uniqueLeads int64
	
	dss.db.Table("application_numbers").Where("created_at >= ?", today).Count(&applicationsToday)
	dss.db.Table("application_numbers").Where("created_at >= ? AND created_at < ?", yesterday, today).Count(&applicationsYesterday)
	dss.db.Table("properties").Where("created_at >= ?", today).Count(&propertiesListed)
	dss.db.Table("contacts").Where("created_at >= ?", today).Count(&uniqueLeads)
	
	dailyStats := map[string]interface{}{
		"applications_count": applicationsToday,
		"yesterday_applications": applicationsYesterday,
		"revenue_today": revenueToday,
		"properties_listed_today": propertiesListed,
		"unique_leads": uniqueLeads,
		"date": today,
		"timestamp": time.Now(),
	}
	
	if dss.cache != nil && dss.cache.IsAvailable() {
		if err := dss.cache.SetDashboardDaily(dailyStats); err != nil {
			log.Printf("⚠️ Failed to cache: %v", err)
		} else {
			log.Println("✅ Dashboard daily stats cached (24hr TTL)")
		}
	}
	
	return dailyStats, nil
}

func (dss *DashboardStatsService) WarmCache() error {
	log.Println("🔥 Warming dashboard cache...")
	
	if _, err := dss.computeHotStats(); err != nil {
		log.Printf("⚠️ Failed to warm hot cache: %v", err)
	}
	
	if _, err := dss.computeWarmStats(); err != nil {
		log.Printf("⚠️ Failed to warm warm cache: %v", err)
	}
	
	if _, err := dss.computeDailyStats(); err != nil {
		log.Printf("⚠️ Failed to warm daily cache: %v", err)
	}
	
	log.Println("✅ Dashboard cache warmed successfully")
	return nil
}
