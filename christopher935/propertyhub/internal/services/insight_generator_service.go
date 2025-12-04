package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InsightGeneratorService generates AI-powered insights from decision engines
type InsightGeneratorService struct {
	db            *gorm.DB
	scoringEngine *BehavioralScoringEngine
	biService     *BusinessIntelligenceService
}

// Insight represents a single AI-generated insight
type Insight struct {
	ID       string   `json:"id"`
	Message  string   `json:"message"`
	Type     string   `json:"type"`     // "success", "warning", "info", "critical"
	Priority int      `json:"priority"` // 1-5, 5 being highest
	Actions  []Action `json:"actions"`
	Source   string   `json:"source"` // Which engine generated this
}

// Action represents an actionable item from an insight
type Action struct {
	Label    string `json:"label"`
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
}

// DashboardInsights contains all insights for the dashboard
type DashboardInsights struct {
	Insights []Insight              `json:"insights"`
	Metrics  map[string]interface{} `json:"metrics"`
}

// NewInsightGeneratorService creates a new insight generator
func NewInsightGeneratorService(db *gorm.DB, scoringEngine *BehavioralScoringEngine, biService *BusinessIntelligenceService) *InsightGeneratorService {
	return &InsightGeneratorService{
		db:            db,
		scoringEngine: scoringEngine,
		biService:     biService,
	}
}

// GenerateDashboardInsights generates insights for the admin dashboard
func (s *InsightGeneratorService) GenerateDashboardInsights() (*DashboardInsights, error) {
	insights := []Insight{}

	// Generate insights from different engines
	leadInsights, err := s.generateLeadInsights()
	if err == nil {
		insights = append(insights, leadInsights...)
	}

	propertyInsights, err := s.generatePropertyInsights()
	if err == nil {
		insights = append(insights, propertyInsights...)
	}

	bookingInsights, err := s.generateBookingInsights()
	if err == nil {
		insights = append(insights, bookingInsights...)
	}

	emailInsights, err := s.generateEmailInsights()
	if err == nil {
		insights = append(insights, emailInsights...)
	}

	// Get metrics
	metrics, err := s.getDashboardMetrics()
	if err != nil {
		metrics = make(map[string]interface{})
	}

	return &DashboardInsights{
		Insights: insights,
		Metrics:  metrics,
	}, nil
}

// generateLeadInsights analyzes leads and generates insights
func (s *InsightGeneratorService) generateLeadInsights() ([]Insight, error) {
	insights := []Insight{}

	// Count hot leads (behavioral score > 70)
	var hotLeads int64
	err := s.db.Table("behavioral_scores").
		Where("overall_score > ?", 70).
		Where("created_at > ?", time.Now().AddDate(0, 0, -7)).
		Count(&hotLeads).Error

	if err == nil && hotLeads > 0 {
		insights = append(insights, Insight{
			ID:       "hot-leads-alert",
			Message:  fmt.Sprintf("You have %d hot leads with engagement scores above 70. Follow up now for best conversion.", hotLeads),
			Type:     "success",
			Priority: 5,
			Source:   "BehavioralScoring",
			Actions: []Action{
				{Label: "View Hot Leads", Endpoint: "/admin/leads?filter=hot", Method: "GET"},
				{Label: "Send Campaign", Endpoint: "/api/email/campaign", Method: "POST"},
			},
		})
	}

	// Count cold leads (no activity in 30 days)
	var coldLeads int64
	err = s.db.Table("leads").
		Where("last_contact < ?", time.Now().AddDate(0, 0, -30)).
		Where("status = ?", "active").
		Count(&coldLeads).Error

	if err == nil && coldLeads > 5 {
		insights = append(insights, Insight{
			ID:       "cold-leads-warning",
			Message:  fmt.Sprintf("%d leads haven't been contacted in 30+ days. Re-engagement recommended.", coldLeads),
			Type:     "warning",
			Priority: 3,
			Source:   "LeadReengagement",
			Actions: []Action{
				{Label: "View Cold Leads", Endpoint: "/admin/leads?filter=cold", Method: "GET"},
				{Label: "Start Re-engagement", Endpoint: "/api/leads/reengagement", Method: "POST"},
			},
		})
	}

	// Count new leads today
	var newLeadsToday int64
	err = s.db.Table("leads").
		Where("created_at > ?", time.Now().Truncate(24*time.Hour)).
		Count(&newLeadsToday).Error

	if err == nil && newLeadsToday > 0 {
		insights = append(insights, Insight{
			ID:       "new-leads-today",
			Message:  fmt.Sprintf("%d new leads captured today. Respond within 5 minutes for 400%% better conversion.", newLeadsToday),
			Type:     "info",
			Priority: 4,
			Source:   "LeadManagement",
			Actions: []Action{
				{Label: "View New Leads", Endpoint: "/admin/leads?filter=today", Method: "GET"},
			},
		})
	}

	return insights, nil
}

// generatePropertyInsights analyzes properties and generates insights
func (s *InsightGeneratorService) generatePropertyInsights() ([]Insight, error) {
	insights := []Insight{}

	// Count properties with no images
	var noImageCount int64
	err := s.db.Table("properties").
		Where("(images IS NULL OR images = '{}' OR images = '')").
		Where("status = ?", "https://schema.org/InStock").
		Count(&noImageCount).Error

	if err == nil && noImageCount > 0 {
		insights = append(insights, Insight{
			ID:       "properties-no-images",
			Message:  fmt.Sprintf("%d active properties are missing photos. Listings with photos get 3x more inquiries.", noImageCount),
			Type:     "warning",
			Priority: 4,
			Source:   "PropertyValuation",
			Actions: []Action{
				{Label: "View Properties", Endpoint: "/admin/properties?filter=no-images", Method: "GET"},
			},
		})
	}

	// Count properties with old pricing (not updated in 60 days)
	var stalePricing int64
	err = s.db.Table("properties").
		Where("updated_at < ?", time.Now().AddDate(0, 0, -60)).
		Where("status = ?", "https://schema.org/InStock").
		Count(&stalePricing).Error

	if err == nil && stalePricing > 0 {
		insights = append(insights, Insight{
			ID:       "stale-pricing",
			Message:  fmt.Sprintf("%d properties haven't had pricing reviewed in 60+ days. Market conditions may have changed.", stalePricing),
			Type:     "info",
			Priority: 2,
			Source:   "PropertyValuation",
			Actions: []Action{
				{Label: "Review Pricing", Endpoint: "/admin/properties?filter=stale-pricing", Method: "GET"},
			},
		})
	}

	return insights, nil
}

// generateBookingInsights analyzes bookings and generates insights
func (s *InsightGeneratorService) generateBookingInsights() ([]Insight, error) {
	insights := []Insight{}

	// Count upcoming showings today
	var showingsToday int64
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := s.db.Table("bookings").
		Where("scheduled_at >= ? AND scheduled_at < ?", today, tomorrow).
		Where("status = ?", "confirmed").
		Count(&showingsToday).Error

	if err == nil && showingsToday > 0 {
		insights = append(insights, Insight{
			ID:       "showings-today",
			Message:  fmt.Sprintf("You have %d property showings scheduled today. Preparation checklist recommended.", showingsToday),
			Type:     "info",
			Priority: 4,
			Source:   "Calendar",
			Actions: []Action{
				{Label: "View Schedule", Endpoint: "/admin/calendar?view=today", Method: "GET"},
			},
		})
	}

	// Count no-shows in last 7 days
	var noShows int64
	err = s.db.Table("bookings").
		Where("status = ?", "no-show").
		Where("scheduled_at > ?", time.Now().AddDate(0, 0, -7)).
		Count(&noShows).Error

	if err == nil && noShows > 2 {
		insights = append(insights, Insight{
			ID:       "high-no-show-rate",
			Message:  fmt.Sprintf("%d no-shows in the past week. Consider implementing confirmation reminders.", noShows),
			Type:     "warning",
			Priority: 3,
			Source:   "Calendar",
			Actions: []Action{
				{Label: "Enable Reminders", Endpoint: "/admin/settings/calendar", Method: "GET"},
			},
		})
	}

	return insights, nil
}

// generateEmailInsights analyzes email campaigns and generates insights
func (s *InsightGeneratorService) generateEmailInsights() ([]Insight, error) {
	insights := []Insight{}

	// Count active campaigns
	var activeCampaigns int64
	err := s.db.Table("email_campaigns").
		Where("status = ?", "active").
		Count(&activeCampaigns).Error

	if err == nil && activeCampaigns > 0 {
		// Get campaign with low open rate
		var lowOpenRate struct {
			Name     string
			OpenRate float64
		}
		err = s.db.Table("email_campaigns").
			Select("name, open_rate").
			Where("status = ?", "active").
			Where("open_rate < ?", 15.0).
			Order("open_rate ASC").
			First(&lowOpenRate).Error

		if err == nil {
			insights = append(insights, Insight{
				ID:       "low-email-open-rate",
				Message:  fmt.Sprintf("Campaign '%s' has only %.1f%% open rate. Try A/B testing subject lines.", lowOpenRate.Name, lowOpenRate.OpenRate),
				Type:     "warning",
				Priority: 2,
				Source:   "EmailAutomation",
				Actions: []Action{
					{Label: "View Campaign", Endpoint: "/admin/email/campaigns", Method: "GET"},
				},
			})
		}
	}

	// Count scheduled emails
	var scheduledEmails int64
	err = s.db.Table("email_campaigns").
		Where("status = ?", "scheduled").
		Where("scheduled_at > ?", time.Now()).
		Count(&scheduledEmails).Error

	if err == nil && scheduledEmails > 0 {
		insights = append(insights, Insight{
			ID:       "scheduled-emails",
			Message:  fmt.Sprintf("%d email campaigns are scheduled to send. Review before they go out.", scheduledEmails),
			Type:     "info",
			Priority: 2,
			Source:   "EmailAutomation",
			Actions: []Action{
				{Label: "Review Campaigns", Endpoint: "/admin/email/campaigns?filter=scheduled", Method: "GET"},
			},
		})
	}

	return insights, nil
}

// getDashboardMetrics retrieves key metrics for the dashboard
func (s *InsightGeneratorService) getDashboardMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Total leads
	var totalLeads int64
	s.db.Table("leads").Count(&totalLeads)
	metrics["total_leads"] = totalLeads

	// Active properties
	var activeProperties int64
	s.db.Table("properties").
		Where("status = ?", "https://schema.org/InStock").
		Count(&activeProperties)
	metrics["active_properties"] = activeProperties

	// Bookings this week
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	var bookingsThisWeek int64
	s.db.Table("bookings").
		Where("created_at >= ?", weekStart).
		Count(&bookingsThisWeek)
	metrics["bookings_this_week"] = bookingsThisWeek

	// Applications pending
	var pendingApplications int64
	s.db.Table("applications").
		Where("status = ?", "pending").
		Count(&pendingApplications)
	metrics["pending_applications"] = pendingApplications

	// Cold leads (no activity in 30 days)
	var coldLeads int64
	s.db.Table("leads").
		Where("last_contact < ?", time.Now().AddDate(0, 0, -30)).
		Where("status = ?", "active").
		Count(&coldLeads)
	metrics["cold_leads"] = coldLeads

	// Showings this week
	var showingsThisWeek int64
	s.db.Table("bookings").
		Where("booking_type = ?", "showing").
		Where("created_at >= ?", weekStart).
		Count(&showingsThisWeek)
	metrics["showings_this_week"] = showingsThisWeek

	// Pending approvals
	var pendingApprovals int64
	s.db.Table("applications").
		Where("status = ?", "awaiting_approval").
		Count(&pendingApprovals)
	metrics["pending_approvals"] = pendingApprovals

	// Total active leads
	var totalActiveLeads int64
	s.db.Table("leads").
		Where("status = ?", "active").
		Count(&totalActiveLeads)
	metrics["total_active_leads"] = totalActiveLeads

	// Average response time (mock for now)
	metrics["avg_response_time_minutes"] = 18

	// Conversion rate (mock for now)
	metrics["conversion_rate_percent"] = 23.5

	return metrics, nil
}
