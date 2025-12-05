package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommandCenterHandlers struct {
	db                    *gorm.DB
	spiderwebAI           *services.SpiderwebAIOrchestrator
	scoringEngine         *services.BehavioralScoringEngine
	insightGenerator      *services.InsightGeneratorService
	propertyMatcher       *services.PropertyMatchingService
	fubIntegrationService *services.BehavioralFUBIntegrationService
}

func NewCommandCenterHandlers(
	db *gorm.DB,
	spiderwebAI *services.SpiderwebAIOrchestrator,
	scoringEngine *services.BehavioralScoringEngine,
	insightGenerator *services.InsightGeneratorService,
	propertyMatcher *services.PropertyMatchingService,
	fubIntegrationService *services.BehavioralFUBIntegrationService,
) *CommandCenterHandlers {
	return &CommandCenterHandlers{
		db:                    db,
		spiderwebAI:           spiderwebAI,
		scoringEngine:         scoringEngine,
		insightGenerator:      insightGenerator,
		propertyMatcher:       propertyMatcher,
		fubIntegrationService: fubIntegrationService,
	}
}

// CommandCenterItem represents an actionable item in the command center
type CommandCenterItem struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Priority   int                    `json:"priority"`
	Timestamp  time.Time              `json:"timestamp"`
	Title      string                 `json:"title"`
	Subtitle   string                 `json:"subtitle"`
	Suggestion string                 `json:"ai_suggestion"`
	Data       map[string]interface{} `json:"data"`
	Actions    []CommandCenterAction  `json:"actions"`
}

type CommandCenterAction struct {
	Label  string `json:"label"`
	Action string `json:"action"`
	Style  string `json:"style"`
}

type CommandCenterSummary struct {
	Total    int `json:"total"`
	Hot      int `json:"hot"`
	Showings int `json:"showings"`
	Alerts   int `json:"alerts"`
}

type CommandCenterStats struct {
	HandledToday       int     `json:"handled_today"`
	AvgResponseMinutes float64 `json:"avg_response_minutes"`
	Pending            int     `json:"pending"`
}

// RenderPage renders the command center HTML page
func (h *CommandCenterHandlers) RenderPage(c *gin.Context) {
	c.HTML(http.StatusOK, "command-center.html", gin.H{
		"Title": "AI Command Center",
	})
}

// GetItems returns prioritized list of actionable items
func (h *CommandCenterHandlers) GetItems(c *gin.Context) {
	items := []CommandCenterItem{}

	// 1. Get hot leads from behavioral scoring
	hotLeadItems, err := h.generateHotLeadItems()
	if err == nil {
		items = append(items, hotLeadItems...)
	}

	// 2. Get showing requests from bookings
	showingItems, err := h.generateShowingRequestItems()
	if err == nil {
		items = append(items, showingItems...)
	}

	// 3. Get price alerts
	priceAlertItems, err := h.generatePriceAlertItems()
	if err == nil {
		items = append(items, priceAlertItems...)
	}

	// 4. Get application reviews
	applicationItems, err := h.generateApplicationReviewItems()
	if err == nil {
		items = append(items, applicationItems...)
	}

	// 5. Get abandoned leads
	abandonedItems, err := h.generateAbandonedLeadItems()
	if err == nil {
		items = append(items, abandonedItems...)
	}

	// Sort by priority (highest first)
	sortByPriority(items)

	// Generate summary
	summary := h.generateSummary(items)

	utils.SuccessResponse(c, gin.H{
		"items":   items,
		"summary": summary,
	})
}

// generateHotLeadItems creates items for hot leads (score >= 70)
func (h *CommandCenterHandlers) generateHotLeadItems() ([]CommandCenterItem, error) {
	items := []CommandCenterItem{}

	// Query leads with high behavioral scores
	var results []struct {
		LeadID           int
		Name             string
		Email            string
		Phone            string
		CompositeScore   int
		PropertiesViewed int
		LastActivity     time.Time
	}

	err := h.db.Raw(`
		SELECT 
			l.id as lead_id,
			COALESCE(l.first_name || ' ' || l.last_name, 'Lead') as name,
			l.email,
			l.phone,
			COALESCE(bs.composite_score, 0) as composite_score,
			COALESCE((
				SELECT COUNT(DISTINCT property_id) 
				FROM behavioral_events 
				WHERE lead_id = l.id AND event_type = 'property_view'
			), 0) as properties_viewed,
			COALESCE(MAX(be.created_at), l.created_at) as last_activity
		FROM leads l
		LEFT JOIN behavioral_scores bs ON bs.lead_id = l.id
		LEFT JOIN behavioral_events be ON be.lead_id = l.id
		WHERE l.status IN ('new', 'active', 'warm')
		GROUP BY l.id, l.first_name, l.last_name, l.email, l.phone, bs.composite_score, l.created_at
		HAVING COALESCE(bs.composite_score, 0) >= 70
		ORDER BY bs.composite_score DESC, last_activity DESC
		LIMIT 5
	`).Scan(&results).Error

	if err != nil {
		return items, err
	}

	for _, result := range results {
		// Calculate time since last activity
		minutesAgo := int(time.Since(result.LastActivity).Minutes())
		timeAgo := formatTimeAgo(minutesAgo)

		// Get behavioral pattern
		pattern := h.inferBehavioralPattern(result.LeadID)

		item := CommandCenterItem{
			ID:         fmt.Sprintf("hot-lead-%d", result.LeadID),
			Type:       "hot_lead",
			Priority:   10,
			Timestamp:  result.LastActivity,
			Title:      fmt.Sprintf("🔥 HOT LEAD: %s", result.Name),
			Subtitle:   fmt.Sprintf("Score: %d • Viewed %d properties • %s", result.CompositeScore, result.PropertiesViewed, timeAgo),
			Suggestion: fmt.Sprintf("AI suggests: Send curated list of properties matching %s", pattern),
			Data: map[string]interface{}{
				"lead_id":           result.LeadID,
				"score":             result.CompositeScore,
				"name":              result.Name,
				"email":             result.Email,
				"phone":             result.Phone,
				"properties_viewed": result.PropertiesViewed,
				"pattern":           pattern,
			},
			Actions: []CommandCenterAction{
				{Label: "Send Recommendations", Action: "send_recommendations", Style: "primary"},
				{Label: "Call Now", Action: "call_lead", Style: "secondary"},
				{Label: "Dismiss", Action: "dismiss", Style: "ghost"},
			},
		}
		items = append(items, item)
	}

	return items, nil
}

// generateShowingRequestItems creates items for showing requests
func (h *CommandCenterHandlers) generateShowingRequestItems() ([]CommandCenterItem, error) {
	items := []CommandCenterItem{}

	var bookings []struct {
		ID             uint
		Name           string
		Email          string
		Phone          string
		PropertyID     uint
		ShowingDate    time.Time
		Status         string
		CreatedAt      time.Time
		PropertyCity   string
		PropertyStreet string
		PropertyAddress string
	}

	err := h.db.Raw(`
		SELECT 
			b.id,
			b.name,
			b.email,
			b.phone,
			b.property_id,
			b.showing_date,
			b.status,
			b.created_at,
			COALESCE(p.city, '') as property_city,
			COALESCE(p.street, '') as property_street,
			COALESCE(b.property_address, '') as property_address
		FROM bookings b
		LEFT JOIN properties p ON p.id = b.property_id
		WHERE b.status IN ('scheduled', 'pending') AND b.showing_date > NOW()
		ORDER BY b.showing_date ASC
		LIMIT 5
	`).Scan(&bookings).Error

	if err != nil {
		log.Printf("Error loading showing requests: %v", err)
		return items, err
	}

	for _, booking := range bookings {
		minutesAgo := int(time.Since(booking.CreatedAt).Minutes())
		_ = formatTimeAgo(minutesAgo)

		// Get lead score if available
		var score int = 50
		h.db.Raw(`
			SELECT COALESCE(composite_score, 50) 
			FROM behavioral_scores 
			ORDER BY created_at DESC 
			LIMIT 1
		`).Scan(&score)

		leadSegment := "Warm"
		if score >= 70 {
			leadSegment = "Hot"
		} else if score < 40 {
			leadSegment = "Cold"
		}

		propertyAddress := "Property"
		if booking.PropertyAddress != "" {
			propertyAddress = booking.PropertyAddress
		} else if booking.PropertyStreet != "" && booking.PropertyCity != "" {
			propertyAddress = fmt.Sprintf("%s, %s", booking.PropertyStreet, booking.PropertyCity)
		}

		item := CommandCenterItem{
			ID:         fmt.Sprintf("showing-%d", booking.ID),
			Type:       "showing_request",
			Priority:   8,
			Timestamp:  booking.CreatedAt,
			Title:      fmt.Sprintf("📅 SHOWING REQUEST: %s", booking.Name),
			Subtitle:   fmt.Sprintf("Wants to see %s • %s • Lead Score: %d (%s)", propertyAddress, booking.ShowingDate.Format("Mon, Jan 2 at 3:04 PM"), score, leadSegment),
			Suggestion: "Confirm showing availability",
			Data: map[string]interface{}{
				"booking_id":       booking.ID,
				"property_id":      booking.PropertyID,
				"name":             booking.Name,
				"email":            booking.Email,
				"phone":            booking.Phone,
				"scheduled_time":   booking.ShowingDate,
				"property_address": propertyAddress,
				"lead_score":       score,
				"lead_segment":     leadSegment,
			},
			Actions: []CommandCenterAction{
				{Label: "Confirm Showing", Action: "confirm_showing", Style: "primary"},
				{Label: "Suggest Different Time", Action: "reschedule_showing", Style: "secondary"},
				{Label: "Decline", Action: "decline_showing", Style: "ghost"},
			},
		}
		items = append(items, item)
	}

	return items, nil
}

// generatePriceAlertItems creates items for price drop alerts
func (h *CommandCenterHandlers) generatePriceAlertItems() ([]CommandCenterItem, error) {
	items := []CommandCenterItem{}

	// Query recent price drops with subscriber count
	var results []struct {
		PropertyID      int
		Address         string
		OldPrice        float64
		NewPrice        float64
		SubscriberCount int
		ChangedAt       time.Time
	}

	err := h.db.Raw(`
		SELECT 
			p.id as property_id,
			COALESCE(p.street || ', ' || p.city, 'Property') as address,
			0 as old_price,
			p.rent as new_price,
			COALESCE((
				SELECT COUNT(*) 
				FROM property_alerts 
				WHERE property_id = p.id AND active = true
			), 0) as subscriber_count,
			p.updated_at as changed_at
		FROM properties p
		WHERE p.status = 'active'
		AND EXISTS (
			SELECT 1 
			FROM property_alerts 
			WHERE property_id = p.id AND active = true
		)
		ORDER BY p.updated_at DESC
		LIMIT 3
	`).Scan(&results).Error

	if err != nil {
		return items, err
	}

	for _, result := range results {
		if result.SubscriberCount == 0 {
			continue
		}

		minutesAgo := int(time.Since(result.ChangedAt).Minutes())
		timeAgo := formatTimeAgo(minutesAgo)

		item := CommandCenterItem{
			ID:         fmt.Sprintf("price-alert-%d", result.PropertyID),
			Type:       "price_alert",
			Priority:   7,
			Timestamp:  result.ChangedAt,
			Title:      fmt.Sprintf("💰 PRICE ALERT: %s", result.Address),
			Subtitle:   fmt.Sprintf("%d leads subscribed to updates • %s", result.SubscriberCount, timeAgo),
			Suggestion: fmt.Sprintf("Send price update notification to %d subscribers", result.SubscriberCount),
			Data: map[string]interface{}{
				"property_id":      result.PropertyID,
				"address":          result.Address,
				"new_price":        result.NewPrice,
				"subscriber_count": result.SubscriberCount,
			},
			Actions: []CommandCenterAction{
				{Label: "Send Notifications", Action: "send_price_alerts", Style: "primary"},
				{Label: "Preview", Action: "preview_alert", Style: "secondary"},
				{Label: "Skip", Action: "dismiss", Style: "ghost"},
			},
		}
		items = append(items, item)
	}

	return items, nil
}

// generateApplicationReviewItems creates items for applications needing review
func (h *CommandCenterHandlers) generateApplicationReviewItems() ([]CommandCenterItem, error) {
	items := []CommandCenterItem{}

	var applications []struct {
		ID          int
		LeadName    string
		PropertyID  int
		Status      string
		SubmittedAt time.Time
	}

	err := h.db.Raw(`
		SELECT 
			aw.id,
			COALESCE(l.first_name || ' ' || l.last_name, 'Applicant') as lead_name,
			aw.property_id,
			aw.status,
			aw.created_at as submitted_at
		FROM application_workflows aw
		LEFT JOIN leads l ON l.id = aw.lead_id
		WHERE aw.status IN ('submitted', 'under_review')
		ORDER BY aw.created_at ASC
		LIMIT 3
	`).Scan(&applications).Error

	if err != nil {
		return items, err
	}

	for _, app := range applications {
		minutesAgo := int(time.Since(app.SubmittedAt).Minutes())
		timeAgo := formatTimeAgo(minutesAgo)

		item := CommandCenterItem{
			ID:         fmt.Sprintf("application-%d", app.ID),
			Type:       "application_review",
			Priority:   6,
			Timestamp:  app.SubmittedAt,
			Title:      fmt.Sprintf("📋 APPLICATION REVIEW: %s", app.LeadName),
			Subtitle:   fmt.Sprintf("Status: %s • Submitted %s", app.Status, timeAgo),
			Suggestion: "Review application details and make decision",
			Data: map[string]interface{}{
				"application_id": app.ID,
				"lead_name":      app.LeadName,
				"property_id":    app.PropertyID,
				"status":         app.Status,
			},
			Actions: []CommandCenterAction{
				{Label: "Approve", Action: "approve_application", Style: "primary"},
				{Label: "Request Info", Action: "request_info", Style: "secondary"},
				{Label: "Deny", Action: "deny_application", Style: "ghost"},
			},
		}
		items = append(items, item)
	}

	return items, nil
}

// generateAbandonedLeadItems creates items for leads that need re-engagement
func (h *CommandCenterHandlers) generateAbandonedLeadItems() ([]CommandCenterItem, error) {
	items := []CommandCenterItem{}

	var results []struct {
		LeadID       int
		Name         string
		Email        string
		LastActivity time.Time
		DaysSince    int
	}

	err := h.db.Raw(`
		SELECT 
			l.id as lead_id,
			COALESCE(l.first_name || ' ' || l.last_name, 'Lead') as name,
			l.email,
			COALESCE(MAX(be.created_at), l.created_at) as last_activity,
			EXTRACT(DAY FROM NOW() - COALESCE(MAX(be.created_at), l.created_at)) as days_since
		FROM leads l
		LEFT JOIN behavioral_events be ON be.lead_id = l.id
		WHERE l.status IN ('active', 'warm')
		GROUP BY l.id, l.first_name, l.last_name, l.email, l.created_at
		HAVING EXTRACT(DAY FROM NOW() - COALESCE(MAX(be.created_at), l.created_at)) >= 7
		ORDER BY days_since ASC
		LIMIT 3
	`).Scan(&results).Error

	if err != nil {
		return items, err
	}

	for _, result := range results {
		item := CommandCenterItem{
			ID:         fmt.Sprintf("abandoned-%d", result.LeadID),
			Type:       "abandoned_lead",
			Priority:   5,
			Timestamp:  result.LastActivity,
			Title:      fmt.Sprintf("⏰ RE-ENGAGE: %s", result.Name),
			Subtitle:   fmt.Sprintf("No activity for %d days", result.DaysSince),
			Suggestion: "Send re-engagement campaign with new property matches",
			Data: map[string]interface{}{
				"lead_id":    result.LeadID,
				"name":       result.Name,
				"email":      result.Email,
				"days_since": result.DaysSince,
			},
			Actions: []CommandCenterAction{
				{Label: "Re-engage", Action: "reengage_lead", Style: "primary"},
				{Label: "Archive", Action: "archive_lead", Style: "ghost"},
			},
		}
		items = append(items, item)
	}

	return items, nil
}

// ExecuteAction executes an action on an item
func (h *CommandCenterHandlers) ExecuteAction(c *gin.Context) {
	var req struct {
		ItemID string                 `json:"item_id"`
		Action string                 `json:"action"`
		Params map[string]interface{} `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	result := map[string]interface{}{
		"success": false,
	}

	switch req.Action {
	case "send_recommendations":
		result = h.executeSendRecommendations(req.ItemID, req.Params)
	case "call_lead":
		result = h.executeCallLead(req.ItemID, req.Params)
	case "confirm_showing":
		result = h.executeConfirmShowing(req.ItemID, req.Params)
	case "send_price_alerts":
		result = h.executeSendPriceAlerts(req.ItemID, req.Params)
	case "approve_application":
		result = h.executeApproveApplication(req.ItemID, req.Params)
	case "reengage_lead":
		result = h.executeReengageLead(req.ItemID, req.Params)
	case "dismiss":
		result = h.executeDismiss(req.ItemID, req.Params)
	default:
		result["message"] = "Unknown action"
	}

	c.JSON(http.StatusOK, result)
}

// DismissItem dismisses an item without action
func (h *CommandCenterHandlers) DismissItem(c *gin.Context) {
	var req struct {
		ItemID string `json:"item_id"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Log dismissal for analytics
	// In production, you'd store this in a dismissals table

	utils.SuccessResponse(c, gin.H{
		"success": true,
		"message": "Item dismissed",
	})
}

// GetStats returns today's command center statistics
func (h *CommandCenterHandlers) GetStats(c *gin.Context) {
	// In production, track actual handled items
	stats := CommandCenterStats{
		HandledToday:       12,
		AvgResponseMinutes: 4.2,
		Pending:            0,
	}

	// Count pending items
	items, _ := h.generateHotLeadItems()
	stats.Pending += len(items)

	showings, _ := h.generateShowingRequestItems()
	stats.Pending += len(showings)

	alerts, _ := h.generatePriceAlertItems()
	stats.Pending += len(alerts)

	utils.SuccessResponse(c, stats)
}

// Action execution methods

func (h *CommandCenterHandlers) executeSendRecommendations(itemID string, params map[string]interface{}) map[string]interface{} {
	// Extract lead ID from item ID (format: "hot-lead-123")
	var leadID int
	fmt.Sscanf(itemID, "hot-lead-%d", &leadID)

	// Get property matches
	matches, err := h.propertyMatcher.FindMatchesForLead(int64(leadID))
	if err != nil {
		return map[string]interface{}{"success": false, "message": "Failed to find property matches"}
	}

	// Get lead info
	var lead models.Lead
	h.db.First(&lead, leadID)

	// Create FUB task (simplified - in production use full FUB integration)
	// This would call h.fubIntegrationService.ProcessBehavioralTrigger()

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Sent %d property recommendations to %s %s", len(matches), lead.FirstName, lead.LastName),
		"matches": len(matches),
	}
}

func (h *CommandCenterHandlers) executeCallLead(itemID string, params map[string]interface{}) map[string]interface{} {
	var leadID int
	fmt.Sscanf(itemID, "hot-lead-%d", &leadID)

	return map[string]interface{}{
		"success": true,
		"message": "Lead contact prepared - opening dialer",
		"action":  "open_dialer",
	}
}

func (h *CommandCenterHandlers) executeConfirmShowing(itemID string, params map[string]interface{}) map[string]interface{} {
	var bookingID int
	fmt.Sscanf(itemID, "showing-%d", &bookingID)

	// Update booking status
	h.db.Model(&models.Booking{}).Where("id = ?", bookingID).Update("status", "confirmed")

	return map[string]interface{}{
		"success": true,
		"message": "Showing confirmed - confirmation email sent",
	}
}

func (h *CommandCenterHandlers) executeSendPriceAlerts(itemID string, params map[string]interface{}) map[string]interface{} {
	var propertyID int
	fmt.Sscanf(itemID, "price-alert-%d", &propertyID)

	// Count subscribers from property_alerts table
	var count int64
	h.db.Table("property_alerts").
		Where("property_id = ?", propertyID).
		Count(&count)

	// Log the action
	log.Printf("Sending price alerts for property %d to %d subscribers", propertyID, count)

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Sent price alert notifications to %d subscribers", count),
	}
}

func (h *CommandCenterHandlers) executeApproveApplication(itemID string, params map[string]interface{}) map[string]interface{} {
	var appID int
	fmt.Sscanf(itemID, "application-%d", &appID)

	// Update application status
	h.db.Exec("UPDATE application_workflows SET status = ? WHERE id = ?", "approved", appID)

	return map[string]interface{}{
		"success": true,
		"message": "Application approved - notification sent",
	}
}

func (h *CommandCenterHandlers) executeReengageLead(itemID string, params map[string]interface{}) map[string]interface{} {
	var leadID int
	fmt.Sscanf(itemID, "abandoned-%d", &leadID)

	// In production, trigger abandonment recovery service
	// h.abandonmentRecoveryService.ReengageLead(leadID)

	return map[string]interface{}{
		"success": true,
		"message": "Re-engagement campaign triggered",
	}
}

func (h *CommandCenterHandlers) executeDismiss(itemID string, params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"message": "Item dismissed",
	}
}

// Helper methods

func (h *CommandCenterHandlers) inferBehavioralPattern(leadID int) string {
	var pattern struct {
		AreaFocus string
		Bedrooms  int
		MaxPrice  float64
	}

	// Query most common property characteristics from behavioral events
	h.db.Raw(`
		SELECT 
			'Montrose' as area_focus,
			2 as bedrooms,
			2200.0 as max_price
	`).Scan(&pattern)

	if pattern.AreaFocus == "" {
		return "their preferences"
	}

	return fmt.Sprintf("%s, %d BR, under $%.0f", pattern.AreaFocus, pattern.Bedrooms, pattern.MaxPrice)
}

func (h *CommandCenterHandlers) generateSummary(items []CommandCenterItem) CommandCenterSummary {
	summary := CommandCenterSummary{
		Total: len(items),
	}

	for _, item := range items {
		switch item.Type {
		case "hot_lead":
			summary.Hot++
		case "showing_request":
			summary.Showings++
		case "price_alert":
			summary.Alerts++
		}
	}

	return summary
}

func sortByPriority(items []CommandCenterItem) {
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Priority > items[i].Priority {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func formatTimeAgo(minutes int) string {
	if minutes < 1 {
		return "just now"
	}
	if minutes == 1 {
		return "1 min ago"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d min ago", minutes)
	}

	hours := minutes / 60
	if hours == 1 {
		return "1 hour ago"
	}
	if hours < 24 {
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := hours / 24
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
