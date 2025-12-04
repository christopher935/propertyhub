package services

// ScoringRules defines the point values for different behavioral events
type ScoringRules struct {
	EventPoints map[string]int
}

// DefaultScoringRules returns the default scoring configuration
func DefaultScoringRules() *ScoringRules {
	return &ScoringRules{
		EventPoints: map[string]int{
			// Browsing behavior
			"viewed":   5,  // Viewed a property
			"browsed":  3,  // Browsed property list
			
			// Engagement actions
			"saved":    15, // Saved/favorited a property
			"shared":   10, // Shared a property
			
			// High-intent actions
			"inquired": 25, // Submitted inquiry/contact form
			"applied":  50, // Submitted rental application
			"scheduled": 30, // Scheduled a tour/viewing
			
			// Conversion
			"converted": 100, // Signed lease
			
			// Session engagement
			"session_start": 2,  // Started a session
			"long_session":  5,  // Session > 5 minutes
			
			// Email engagement (from FUB webhooks)
			"email_opened":  3,  // Opened marketing email
			"email_clicked": 10, // Clicked link in email
			
			// Negative signals
			"unsubscribed": -20, // Unsubscribed from emails
		},
	}
}

// GetPoints returns the point value for an event type
func (r *ScoringRules) GetPoints(eventType string) int {
	if points, exists := r.EventPoints[eventType]; exists {
		return points
	}
	return 0 // Unknown event types get 0 points
}

// SetPoints allows customizing point values
func (r *ScoringRules) SetPoints(eventType string, points int) {
	r.EventPoints[eventType] = points
}

// ============================================================================
// SEGMENT THRESHOLDS
// ============================================================================

// SegmentThresholds defines the score ranges for each segment
type SegmentThresholds struct {
	Hot     int // Minimum score for "hot" segment
	Warm    int // Minimum score for "warm" segment
	Cold    int // Minimum score for "cold" segment
	Dormant int // Below this is "dormant"
}

// DefaultSegmentThresholds returns the default segment configuration
func DefaultSegmentThresholds() *SegmentThresholds {
	return &SegmentThresholds{
		Hot:     70,  // 70-100: High intent, recent activity
		Warm:    40,  // 40-69: Moderate interest
		Cold:    10,  // 10-39: Low engagement
		Dormant: 10,  // 0-9: Minimal or no activity
	}
}

// ============================================================================
// SCORING EXAMPLES
// ============================================================================

/*
EXAMPLE SCENARIOS:

1. HOT LEAD (Score: 85)
   - Viewed 10 properties (50 points)
   - Saved 2 properties (30 points)
   - Submitted inquiry (25 points)
   - All within last 7 days (no decay)
   = 105 points (capped at 100)
   Segment: HOT

2. WARM LEAD (Score: 55)
   - Viewed 5 properties (25 points)
   - Saved 1 property (15 points)
   - Opened 3 emails (9 points)
   - Activity over last 14 days (70% decay)
   = 34 * 0.7 = 24 points (recent) + 25 (older) = 49 points
   Segment: WARM

3. COLD LEAD (Score: 20)
   - Viewed 3 properties (15 points)
   - Opened 1 email (3 points)
   - Activity 60 days ago (40% decay)
   = 18 * 0.4 = 7 points
   Segment: COLD

4. DORMANT LEAD (Score: 0)
   - No activity in 120 days
   = 0 points
   Segment: DORMANT

5. APPLICATION SUBMITTED (Score: 95)
   - Viewed 8 properties (40 points)
   - Saved 1 property (15 points)
   - Submitted application (50 points)
   - All within last 3 days
   = 105 points (capped at 100)
   Segment: HOT
*/
