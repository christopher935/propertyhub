package safety

import (
	"log"
	"strings"
	"time"
)

// LeadSafetyLevel represents the automation safety level for a lead
type LeadSafetyLevel int

const (
	SafetyLevelBlocked LeadSafetyLevel = iota // Red - No automation allowed
	SafetyLevelReview                         // Yellow - Requires manual approval
	SafetyLevelSafe                           // Green - Full automation approved
)

// LeadSafetyClassification contains the safety assessment for a lead
type LeadSafetyClassification struct {
	LeadID           string          `json:"lead_id"`
	SafetyLevel      LeadSafetyLevel `json:"safety_level"`
	SafetyScore      int             `json:"safety_score"` // 0-100
	Reasons          []string        `json:"reasons"`
	LastContactDate  *time.Time      `json:"last_contact_date"`
	LastActivityDate *time.Time      `json:"last_activity_date"`
	CreatedDate      *time.Time      `json:"created_date"`
	Status           string          `json:"status"`
	DoNotContact     bool            `json:"do_not_contact"`
	IsExistingTenant bool            `json:"is_existing_tenant"`
	EngagementScore  int             `json:"engagement_score"`
	Flags            []string        `json:"flags"`
}

// LeadSafetyFilter provides methods for lead safety assessment
type LeadSafetyFilter struct {
	SystemActivationDate time.Time
}

// NewLeadSafetyFilter creates a new lead safety filter
func NewLeadSafetyFilter(activationDate time.Time) *LeadSafetyFilter {
	return &LeadSafetyFilter{
		SystemActivationDate: activationDate,
	}
}

// ClassifyLead performs comprehensive safety assessment of a lead
func (f *LeadSafetyFilter) ClassifyLead(lead map[string]interface{}) *LeadSafetyClassification {
	classification := &LeadSafetyClassification{
		LeadID:  getStringValue(lead, "id"),
		Reasons: []string{},
		Flags:   []string{},
	}

	// Extract lead data
	classification.Status = getStringValue(lead, "status")
	classification.DoNotContact = getBoolValue(lead, "do_not_contact")
	classification.IsExistingTenant = getBoolValue(lead, "is_existing_tenant")

	if createdStr := getStringValue(lead, "created"); createdStr != "" {
		if created, err := time.Parse(time.RFC3339, createdStr); err == nil {
			classification.CreatedDate = &created
		}
	}

	if lastContactStr := getStringValue(lead, "last_contact"); lastContactStr != "" {
		if lastContact, err := time.Parse(time.RFC3339, lastContactStr); err == nil {
			classification.LastContactDate = &lastContact
		}
	}

	if lastActivityStr := getStringValue(lead, "last_activity"); lastActivityStr != "" {
		if lastActivity, err := time.Parse(time.RFC3339, lastActivityStr); err == nil {
			classification.LastActivityDate = &lastActivity
		}
	}

	// Calculate safety score
	classification.SafetyScore = f.calculateSafetyScore(classification, lead)

	// Determine safety level
	classification.SafetyLevel = f.determineSafetyLevel(classification)

	// Add specific reasons and flags
	f.addSafetyReasons(classification, lead)

	return classification
}

// calculateSafetyScore calculates a 0-100 safety score for automation
func (f *LeadSafetyFilter) calculateSafetyScore(classification *LeadSafetyClassification, lead map[string]interface{}) int {
	score := 50 // Start with neutral score

	// BLOCKING FACTORS (immediate disqualification)
	if classification.DoNotContact {
		return 0 // Absolute block
	}

	if classification.IsExistingTenant {
		return 0 // Absolute block
	}

	// Check for blocked statuses
	blockedStatuses := []string{"closed", "dead", "unqualified", "spam", "duplicate"}
	for _, blocked := range blockedStatuses {
		if strings.ToLower(classification.Status) == blocked {
			return 0
		}
	}

	// POSITIVE FACTORS

	// New leads (created after system activation) get high scores
	if classification.CreatedDate != nil && classification.CreatedDate.After(f.SystemActivationDate) {
		score += 30
		classification.Flags = append(classification.Flags, "new_lead")
	}

	// Recent activity boosts score
	now := time.Now()
	if classification.LastActivityDate != nil {
		daysSinceActivity := int(now.Sub(*classification.LastActivityDate).Hours() / 24)
		switch {
		case daysSinceActivity <= 7:
			score += 25
			classification.Flags = append(classification.Flags, "recent_activity")
		case daysSinceActivity <= 30:
			score += 15
			classification.Flags = append(classification.Flags, "moderate_activity")
		case daysSinceActivity <= 90:
			score += 5
			classification.Flags = append(classification.Flags, "old_activity")
		default:
			score -= 20
			classification.Flags = append(classification.Flags, "stale_activity")
		}
	}

	// Recent contact affects score
	if classification.LastContactDate != nil {
		daysSinceContact := int(now.Sub(*classification.LastContactDate).Hours() / 24)
		switch {
		case daysSinceContact <= 3:
			score -= 10 // Recently contacted, avoid spam
			classification.Flags = append(classification.Flags, "recently_contacted")
		case daysSinceContact <= 14:
			score += 5 // Good follow-up timing
		case daysSinceContact >= 90:
			score += 10 // Long time since contact, safe to re-engage
			classification.Flags = append(classification.Flags, "long_since_contact")
		}
	}

	// Status-based scoring
	switch strings.ToLower(classification.Status) {
	case "new", "open", "active":
		score += 20
	case "qualified", "hot":
		score += 25
	case "warm":
		score += 15
	case "cold":
		score += 5
	case "nurture":
		score += 10
	default:
		score -= 5
	}

	// Engagement scoring (if available)
	if engagementData, ok := lead["engagement"]; ok {
		if engagement, ok := engagementData.(map[string]interface{}); ok {
			engagementScore := f.calculateEngagementScore(engagement)
			classification.EngagementScore = engagementScore
			score += engagementScore / 5 // Add up to 20 points for high engagement
		}
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateEngagementScore calculates engagement score from FUB data
func (f *LeadSafetyFilter) calculateEngagementScore(engagement map[string]interface{}) int {
	score := 0

	// Email engagement
	if emailOpens := getIntValue(engagement, "email_opens"); emailOpens > 0 {
		score += min(emailOpens*2, 20) // Up to 20 points for email opens
	}

	if emailClicks := getIntValue(engagement, "email_clicks"); emailClicks > 0 {
		score += min(emailClicks*5, 25) // Up to 25 points for email clicks
	}

	// SMS engagement
	if smsReplies := getIntValue(engagement, "sms_replies"); smsReplies > 0 {
		score += min(smsReplies*10, 30) // Up to 30 points for SMS replies
	}

	// Call engagement
	if callsAnswered := getIntValue(engagement, "calls_answered"); callsAnswered > 0 {
		score += min(callsAnswered*15, 25) // Up to 25 points for answered calls
	}

	return min(score, 100)
}

// determineSafetyLevel determines the safety level based on score
func (f *LeadSafetyFilter) determineSafetyLevel(classification *LeadSafetyClassification) LeadSafetyLevel {
	switch {
	case classification.SafetyScore >= 80:
		return SafetyLevelSafe
	case classification.SafetyScore >= 50:
		return SafetyLevelReview
	default:
		return SafetyLevelBlocked
	}
}

// addSafetyReasons adds human-readable reasons for the safety classification
func (f *LeadSafetyFilter) addSafetyReasons(classification *LeadSafetyClassification, lead map[string]interface{}) {
	switch classification.SafetyLevel {
	case SafetyLevelBlocked:
		if classification.DoNotContact {
			classification.Reasons = append(classification.Reasons, "Lead marked as 'Do Not Contact'")
		}
		if classification.IsExistingTenant {
			classification.Reasons = append(classification.Reasons, "Lead is existing tenant")
		}
		if classification.SafetyScore < 50 {
			classification.Reasons = append(classification.Reasons, "Low safety score due to inactivity or status")
		}

	case SafetyLevelReview:
		classification.Reasons = append(classification.Reasons, "Moderate safety score - requires manual review")
		if classification.LastActivityDate != nil {
			daysSince := int(time.Now().Sub(*classification.LastActivityDate).Hours() / 24)
			if daysSince > 30 {
				classification.Reasons = append(classification.Reasons, "Lead has been inactive for over 30 days")
			}
		}

	case SafetyLevelSafe:
		classification.Reasons = append(classification.Reasons, "High safety score - approved for automation")
		if classification.CreatedDate != nil && classification.CreatedDate.After(f.SystemActivationDate) {
			classification.Reasons = append(classification.Reasons, "New lead created after system activation")
		}
	}
}

// IsAutomationAllowed checks if automation is allowed for a lead
func (classification *LeadSafetyClassification) IsAutomationAllowed() bool {
	return classification.SafetyLevel == SafetyLevelSafe
}

// RequiresApproval checks if automation requires manual approval
func (classification *LeadSafetyClassification) RequiresApproval() bool {
	return classification.SafetyLevel == SafetyLevelReview
}

// IsBlocked checks if automation is completely blocked
func (classification *LeadSafetyClassification) IsBlocked() bool {
	return classification.SafetyLevel == SafetyLevelBlocked
}

// GetSafetyLevelString returns human-readable safety level
func (classification *LeadSafetyClassification) GetSafetyLevelString() string {
	switch classification.SafetyLevel {
	case SafetyLevelSafe:
		return "Safe"
	case SafetyLevelReview:
		return "Requires Review"
	case SafetyLevelBlocked:
		return "Blocked"
	default:
		return "Unknown"
	}
}

// Helper functions
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolValue(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getIntValue(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BatchSafetyAnalysis analyzes multiple leads for safety
func (f *LeadSafetyFilter) BatchSafetyAnalysis(leads []map[string]interface{}) map[string]*LeadSafetyClassification {
	results := make(map[string]*LeadSafetyClassification)

	for _, lead := range leads {
		classification := f.ClassifyLead(lead)
		results[classification.LeadID] = classification

		log.Printf("Lead %s classified as %s (score: %d)",
			classification.LeadID,
			classification.GetSafetyLevelString(),
			classification.SafetyScore)
	}

	return results
}

// GetSafeLeadsForAutomation returns only leads that are safe for automation
func (f *LeadSafetyFilter) GetSafeLeadsForAutomation(leads []map[string]interface{}) []map[string]interface{} {
	var safeLeads []map[string]interface{}

	for _, lead := range leads {
		classification := f.ClassifyLead(lead)
		if classification.IsAutomationAllowed() {
			safeLeads = append(safeLeads, lead)
		}
	}

	return safeLeads
}
