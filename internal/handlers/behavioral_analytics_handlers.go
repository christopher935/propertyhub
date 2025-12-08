package handlers

import (
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BehavioralAnalyticsHandlers provides API endpoints for behavioral analytics
type BehavioralAnalyticsHandlers struct {
	db *gorm.DB
}

// NewBehavioralAnalyticsHandlers creates new behavioral analytics handlers
func NewBehavioralAnalyticsHandlers(db *gorm.DB) *BehavioralAnalyticsHandlers {
	return &BehavioralAnalyticsHandlers{db: db}
}

// ============================================================================
// GET /api/v1/behavioral/trends
// ============================================================================

// GetBehavioralTrends returns time-series data for trend charts
func (h *BehavioralAnalyticsHandlers) GetBehavioralTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "daily") // daily, weekly, monthly
	metric := c.DefaultQuery("metric", "all")   // active_users, engagement, conversions, all
	days := c.DefaultQuery("days", "30")        // Number of days to look back

	daysInt, _ := strconv.Atoi(days)
	startDate := time.Now().AddDate(0, 0, -daysInt)

	trends := make(map[string][]models.BehavioralTrendPoint)

	// Active Users Trend
	if metric == "active_users" || metric == "all" {
		var activeUsersTrend []models.BehavioralTrendPoint

		query := `
			SELECT 
				DATE(created_at) as date,
				COUNT(DISTINCT lead_id) as value
			FROM behavioral_events
			WHERE created_at >= ?
			GROUP BY DATE(created_at)
			ORDER BY date
		`

		type TrendResult struct {
			Date  time.Time
			Value int64
		}

		var results []TrendResult
		h.db.Raw(query, startDate).Scan(&results)

		for _, r := range results {
			activeUsersTrend = append(activeUsersTrend, models.BehavioralTrendPoint{
				Date:  r.Date.Format("2006-01-02"),
				Value: float64(r.Value),
				Label: "Active Users",
			})
		}

		trends["active_users"] = activeUsersTrend
	}

	// Engagement Trend (events per user)
	if metric == "engagement" || metric == "all" {
		var engagementTrend []models.BehavioralTrendPoint

		query := `
			SELECT 
				DATE(created_at) as date,
				COUNT(*)::float / NULLIF(COUNT(DISTINCT lead_id), 0) as value
			FROM behavioral_events
			WHERE created_at >= ?
			GROUP BY DATE(created_at)
			ORDER BY date
		`

		type TrendResult struct {
			Date  time.Time
			Value float64
		}

		var results []TrendResult
		h.db.Raw(query, startDate).Scan(&results)

		for _, r := range results {
			engagementTrend = append(engagementTrend, models.BehavioralTrendPoint{
				Date:  r.Date.Format("2006-01-02"),
				Value: r.Value,
				Label: "Avg Events/User",
			})
		}

		trends["engagement"] = engagementTrend
	}

	// Conversion Trend
	if metric == "conversions" || metric == "all" {
		var conversionTrend []models.BehavioralTrendPoint

		query := `
			SELECT 
				DATE(created_at) as date,
				COUNT(*) as value
			FROM behavioral_events
			WHERE created_at >= ? AND event_type = 'converted'
			GROUP BY DATE(created_at)
			ORDER BY date
		`

		type TrendResult struct {
			Date  time.Time
			Value int64
		}

		var results []TrendResult
		h.db.Raw(query, startDate).Scan(&results)

		for _, r := range results {
			conversionTrend = append(conversionTrend, models.BehavioralTrendPoint{
				Date:  r.Date.Format("2006-01-02"),
				Value: float64(r.Value),
				Label: "Conversions",
			})
		}

		trends["conversions"] = conversionTrend
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"period": period,
		"days":   daysInt,
	})
}

// ============================================================================
// GET /api/v1/behavioral/funnel
// ============================================================================

// GetConversionFunnel returns conversion funnel data
func (h *BehavioralAnalyticsHandlers) GetConversionFunnel(c *gin.Context) {
	days := c.DefaultQuery("days", "30")
	daysInt, _ := strconv.Atoi(days)
	startDate := time.Now().AddDate(0, 0, -daysInt)

	// Define funnel stages in order
	stages := []string{"viewed", "saved", "inquired", "applied", "converted"}

	var funnelStages []models.BehavioralFunnelStage
	var totalViewed int64

	for i, stage := range stages {
		var count int64
		h.db.Model(&models.BehavioralEvent{}).
			Where("event_type = ? AND created_at >= ?", stage, startDate).
			Count(&count)

		if i == 0 {
			totalViewed = count
		}

		percentage := 0.0
		if totalViewed > 0 {
			percentage = (float64(count) / float64(totalViewed)) * 100
		}

		conversionRate := 0.0
		if i > 0 && funnelStages[i-1].Count > 0 {
			conversionRate = (float64(count) / float64(funnelStages[i-1].Count)) * 100
		}

		// Calculate average time in stage
		var avgTime float64
		h.db.Model(&models.ConversionFunnelEvent{}).
			Where("stage = ? AND entered_at >= ?", stage, startDate).
			Select("AVG(time_in_stage_seconds)").
			Scan(&avgTime)

		funnelStages = append(funnelStages, models.BehavioralFunnelStage{
			Stage:          stage,
			Count:          int(count),
			Percentage:     percentage,
			ConversionRate: conversionRate,
			AvgTimeInStage: int(avgTime),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"funnel":       funnelStages,
		"total_viewed": totalViewed,
		"conversion_rate": func() float64 {
			if totalViewed > 0 && len(funnelStages) > 0 {
				return (float64(funnelStages[len(funnelStages)-1].Count) / float64(totalViewed)) * 100
			}
			return 0.0
		}(),
		"days": daysInt,
	})
}

// ============================================================================
// GET /api/v1/behavioral/segments
// ============================================================================

// GetBehavioralSegments returns behavioral segment breakdown
func (h *BehavioralAnalyticsHandlers) GetBehavioralSegments(c *gin.Context) {
	segments := []string{"high_engagement", "medium_engagement", "low_engagement", "dormant"}

	var segmentSummaries []models.BehavioralSegmentSummary

	for _, segment := range segments {
		var count int64
		h.db.Model(&models.BehavioralSegment{}).
			Where("segment = ?", segment).
			Count(&count)

		// Calculate conversion rate for this segment
		var conversions int64
		h.db.Raw(`
			SELECT COUNT(DISTINCT be.lead_id)
			FROM behavioral_events be
			INNER JOIN behavioral_segments bs ON be.lead_id = bs.lead_id
			WHERE bs.segment = ? AND be.event_type = 'converted'
		`, segment).Scan(&conversions)

		conversionRate := 0.0
		if count > 0 {
			conversionRate = (float64(conversions) / float64(count)) * 100
		}

		// Calculate average engagement score
		var avgEngagement float64
		h.db.Raw(`
			SELECT AVG(engagement_score)
			FROM behavioral_scores_history bsh
			INNER JOIN behavioral_segments bs ON bsh.lead_id = bs.lead_id
			WHERE bs.segment = ?
		`, segment).Scan(&avgEngagement)

		// Calculate average session length
		var avgSessionLength float64
		h.db.Raw(`
			SELECT AVG(duration_seconds)
			FROM behavioral_sessions bsess
			INNER JOIN behavioral_segments bs ON bsess.lead_id = bs.lead_id
			WHERE bs.segment = ?
		`, segment).Scan(&avgSessionLength)

		segmentSummaries = append(segmentSummaries, models.BehavioralSegmentSummary{
			Segment:            segment,
			LeadCount:          int(count),
			ConversionRate:     conversionRate,
			AvgEngagementScore: avgEngagement,
			AvgSessionLength:   int(avgSessionLength),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"segments": segmentSummaries,
	})
}

// ============================================================================
// GET /api/v1/behavioral/heatmap
// ============================================================================

// GetActivityHeatmap returns activity heatmap data (day-of-week x hour-of-day)
func (h *BehavioralAnalyticsHandlers) GetActivityHeatmap(c *gin.Context) {
	days := c.DefaultQuery("days", "30")
	daysInt, _ := strconv.Atoi(days)
	startDate := time.Now().AddDate(0, 0, -daysInt)

	query := `
		SELECT 
			EXTRACT(DOW FROM created_at)::int as day_of_week,
			EXTRACT(HOUR FROM created_at)::int as hour,
			COUNT(*) as activity
		FROM behavioral_events
		WHERE created_at >= ?
		GROUP BY day_of_week, hour
		ORDER BY day_of_week, hour
	`

	type HeatmapResult struct {
		DayOfWeek int
		Hour      int
		Activity  int
	}

	var results []HeatmapResult
	h.db.Raw(query, startDate).Scan(&results)

	// Find max activity for normalization
	maxActivity := 0
	for _, r := range results {
		if r.Activity > maxActivity {
			maxActivity = r.Activity
		}
	}

	// Convert to heatmap cells with intensity
	var heatmapCells []models.BehavioralHeatmapCell
	for _, r := range results {
		intensity := 0.0
		if maxActivity > 0 {
			intensity = float64(r.Activity) / float64(maxActivity)
		}

		heatmapCells = append(heatmapCells, models.BehavioralHeatmapCell{
			DayOfWeek: r.DayOfWeek,
			Hour:      r.Hour,
			Activity:  r.Activity,
			Intensity: intensity,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"heatmap":      heatmapCells,
		"max_activity": maxActivity,
		"days":         daysInt,
	})
}

// ============================================================================
// GET /api/v1/behavioral/anomalies
// ============================================================================

// GetBehavioralAnomalies returns detected anomalies
func (h *BehavioralAnalyticsHandlers) GetBehavioralAnomalies(c *gin.Context) {
	severity := c.DefaultQuery("severity", "all") // critical, high, medium, low, all
	limit := c.DefaultQuery("limit", "20")

	limitInt, _ := strconv.Atoi(limit)

	query := h.db.Model(&models.BehavioralAnomaly{}).
		Where("resolved_at IS NULL").
		Order("detected_at DESC").
		Limit(limitInt)

	if severity != "all" {
		query = query.Where("severity = ?", severity)
	}

	var anomalies []models.BehavioralAnomaly
	query.Find(&anomalies)

	c.JSON(http.StatusOK, gin.H{
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}

// ============================================================================
// GET /api/v1/behavioral/recommendations
// ============================================================================

// GetAIRecommendations returns AI-generated recommendations
func (h *BehavioralAnalyticsHandlers) GetAIRecommendations(c *gin.Context) {
	status := c.DefaultQuery("status", "pending") // pending, accepted, rejected, expired, completed
	priority := c.DefaultQuery("priority", "all") // urgent, high, medium, low, all
	limit := c.DefaultQuery("limit", "20")

	limitInt, _ := strconv.Atoi(limit)

	query := h.db.Model(&models.BehavioralRecommendation{}).
		Where("status = ?", status).
		Order("priority DESC, created_at DESC").
		Limit(limitInt)

	if priority != "all" {
		query = query.Where("priority = ?", priority)
	}

	var recommendations []models.BehavioralRecommendation
	query.Find(&recommendations)

	// Enrich with lead data
	type EnrichedRecommendation struct {
		models.BehavioralRecommendation
		LeadName  string `json:"lead_name"`
		LeadEmail string `json:"lead_email"`
	}

	var enriched []EnrichedRecommendation
	for _, rec := range recommendations {
		var lead models.Lead
		h.db.First(&lead, rec.LeadID)

		enriched = append(enriched, EnrichedRecommendation{
			BehavioralRecommendation: rec,
			LeadName:                 lead.FirstName + " " + lead.LastName,
			LeadEmail:                lead.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": enriched,
		"count":           len(enriched),
	})
}

// ============================================================================
// GET /api/v1/behavioral/cohorts
// ============================================================================

// GetBehavioralCohorts returns cohort retention analysis
func (h *BehavioralAnalyticsHandlers) GetBehavioralCohorts(c *gin.Context) {
	cohortType := c.DefaultQuery("type", "weekly") // weekly, monthly
	limit := c.DefaultQuery("limit", "12")

	limitInt, _ := strconv.Atoi(limit)

	var cohorts []models.BehavioralCohort
	h.db.Model(&models.BehavioralCohort{}).
		Where("cohort_type = ?", cohortType).
		Order("cohort_date DESC").
		Limit(limitInt).
		Find(&cohorts)

	// Calculate retention for each cohort
	var cohortRetention []models.BehavioralCohortRetention
	for _, cohort := range cohorts {
		retention := models.BehavioralCohortRetention{
			CohortName: cohort.CohortName,
			CohortDate: cohort.CohortDate,
			Week0:      cohort.LeadCount,
		}

		// Calculate retention for weeks 1-4
		for week := 1; week <= 4; week++ {
			weekStart := cohort.CohortDate.AddDate(0, 0, week*7)
			weekEnd := weekStart.AddDate(0, 0, 7)

			var activeCount int64
			h.db.Raw(`
				SELECT COUNT(DISTINCT be.lead_id)
				FROM behavioral_events be
				INNER JOIN behavioral_cohort_members bcm ON be.lead_id = bcm.lead_id
				WHERE bcm.cohort_id = ? 
				AND be.created_at >= ? 
				AND be.created_at < ?
			`, cohort.ID, weekStart, weekEnd).Scan(&activeCount)

			switch week {
			case 1:
				retention.Week1 = int(activeCount)
			case 2:
				retention.Week2 = int(activeCount)
			case 3:
				retention.Week3 = int(activeCount)
			case 4:
				retention.Week4 = int(activeCount)
			}
		}

		cohortRetention = append(cohortRetention, retention)
	}

	c.JSON(http.StatusOK, gin.H{
		"cohorts": cohortRetention,
		"type":    cohortType,
	})
}

// ============================================================================
// REGISTER ROUTES
// ============================================================================

// RegisterBehavioralAnalyticsRoutes registers all behavioral analytics routes
func RegisterBehavioralAnalyticsRoutes(r *gin.RouterGroup, db *gorm.DB) {
	handler := NewBehavioralAnalyticsHandlers(db)

	r.GET("/behavioral/trends", handler.GetBehavioralTrends)
	r.GET("/behavioral/funnel", handler.GetConversionFunnel)
	r.GET("/behavioral/segments", handler.GetBehavioralSegments)
	r.GET("/behavioral/heatmap", handler.GetActivityHeatmap)
	r.GET("/behavioral/anomalies", handler.GetBehavioralAnomalies)
	r.GET("/behavioral/recommendations", handler.GetAIRecommendations)
	r.GET("/behavioral/cohorts", handler.GetBehavioralCohorts)
}
