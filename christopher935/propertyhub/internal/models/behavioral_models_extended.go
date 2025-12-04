package models

import (
	"fmt"
    "database/sql/driver"
    "encoding/json"
	"time"
)

// ============================================================================
// BEHAVIORAL EVENTS
// ============================================================================

// BehavioralEvent represents a single user behavioral event

// JSONB is a custom type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
    if value == nil {
        *j = make(map[string]interface{})
        return nil
    }
    
    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
    }
    
    if len(bytes) == 0 {
        *j = make(map[string]interface{})
        return nil
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(bytes, &result); err != nil {
        return err
    }
    
    *j = result
    return nil
}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
    if j == nil {
        return nil, nil
    }
    return json.Marshal(j)
}


type BehavioralEvent struct {
	ID         int64                  `json:"id" gorm:"primaryKey"`
	LeadID     int64                  `json:"lead_id"`
	EventType  string                 `json:"event_type"` // viewed, saved, inquired, applied, converted, session_start, session_end
	EventData  JSONB `json:"event_data" gorm:"type:jsonb"`
	PropertyID *int64                 `json:"property_id,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (BehavioralEvent) TableName() string {
	return "behavioral_events"
}

// ============================================================================
// BEHAVIORAL SESSIONS
// ============================================================================

// BehavioralSession represents a user session
type BehavioralSession struct {
	ID              string     `json:"id" gorm:"primaryKey"` // UUID
	LeadID          int64      `json:"lead_id"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	DurationSeconds int        `json:"duration_seconds"`
	PageViews       int        `json:"page_views"`
	Interactions    int        `json:"interactions"`
	DeviceType      string     `json:"device_type,omitempty"` // desktop, mobile, tablet
	Browser         string     `json:"browser,omitempty"`
	UserAgent       string     `json:"user_agent,omitempty"`
	IPAddress       string     `json:"ip_address,omitempty"`
	Referrer        string     `json:"referrer,omitempty"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (BehavioralSession) TableName() string {
	return "behavioral_sessions"
}

// ============================================================================
// BEHAVIORAL SCORES HISTORY
// ============================================================================

// BehavioralScoreHistory represents historical behavioral scores
type BehavioralScoreHistory struct {
	ID              int64                  `json:"id" gorm:"primaryKey"`
	LeadID          int64                  `json:"lead_id"`
	UrgencyScore    float64                `json:"urgency_score"`    // 0.00 to 1.00
	FinancialScore  float64                `json:"financial_score"`  // 0.00 to 1.00
	EngagementScore float64                `json:"engagement_score"` // 0.00 to 1.00
	OverallScore    int                    `json:"overall_score"`    // 0 to 100
	ScoreFactors    map[string]interface{} `json:"score_factors" gorm:"type:jsonb"`
	CalculatedAt    time.Time              `json:"calculated_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (BehavioralScoreHistory) TableName() string {
	return "behavioral_scores_history"
}

// ============================================================================
// CONVERSION FUNNEL EVENTS
// ============================================================================

// ConversionFunnelEvent represents a user's progression through the funnel
type ConversionFunnelEvent struct {
	ID                 int64                  `json:"id" gorm:"primaryKey"`
	LeadID             int64                  `json:"lead_id"`
	Stage              string                 `json:"stage"` // viewed, saved, inquired, applied, converted
	PropertyID         *int64                 `json:"property_id,omitempty"`
	EnteredAt          time.Time              `json:"entered_at" gorm:"autoCreateTime"`
	ExitedAt           *time.Time             `json:"exited_at,omitempty"`
	Converted          bool                   `json:"converted"`
	TimeInStageSeconds int                    `json:"time_in_stage_seconds"`
	Metadata           map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName specifies the table name for GORM
func (ConversionFunnelEvent) TableName() string {
	return "conversion_funnel_events"
}

// ============================================================================
// BEHAVIORAL TRIGGERS LOG
// ============================================================================

// BehavioralTriggerLog represents a logged behavioral trigger execution
type BehavioralTriggerLog struct {
	ID                int64                  `json:"id" gorm:"primaryKey"`
	LeadID            int64                  `json:"lead_id"`
	TriggerType       string                 `json:"trigger_type"` // high_urgency, financial_qualified, engagement_spike, etc.
	TriggerData       map[string]interface{} `json:"trigger_data" gorm:"type:jsonb"`
	FUBAutomationID   string                 `json:"fub_automation_id,omitempty"`
	FUBActionPlanID   string                 `json:"fub_action_plan_id,omitempty"`
	FUBContactID      string                 `json:"fub_contact_id,omitempty"`
	Success           bool                   `json:"success"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	ResponseData      map[string]interface{} `json:"response_data" gorm:"type:jsonb"`
	TriggeredAt       time.Time              `json:"triggered_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (BehavioralTriggerLog) TableName() string {
	return "behavioral_triggers_log"
}

// ============================================================================
// BEHAVIORAL RECOMMENDATIONS
// ============================================================================

// BehavioralRecommendation represents an AI-generated recommendation
type BehavioralRecommendation struct {
	ID                 int64                  `json:"id" gorm:"primaryKey"`
	LeadID             int64                  `json:"lead_id"`
	RecommendationType string                 `json:"recommendation_type"` // contact_now, assign_plan, recommend_property, re_engage, etc.
	RecommendationData map[string]interface{} `json:"recommendation_data" gorm:"type:jsonb"`
	Confidence         float64                `json:"confidence"` // 0.00 to 1.00
	Reasoning          string                 `json:"reasoning"`
	Priority           string                 `json:"priority"` // urgent, high, medium, low
	Status             string                 `json:"status"`   // pending, accepted, rejected, expired, completed
	CreatedBy          string                 `json:"created_by"`
	AssignedTo         string                 `json:"assigned_to,omitempty"`
	CreatedAt          time.Time              `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt          time.Time              `json:"expires_at"`
	ActedOnAt          *time.Time             `json:"acted_on_at,omitempty"`
	Outcome            string                 `json:"outcome,omitempty"` // success, failed, ignored
}

// TableName specifies the table name for GORM
func (BehavioralRecommendation) TableName() string {
	return "behavioral_recommendations"
}

// ============================================================================
// BEHAVIORAL SEGMENTS
// ============================================================================

// BehavioralSegment represents a lead's assignment to a behavioral segment
type BehavioralSegment struct {
	ID          int64                  `json:"id" gorm:"primaryKey"`
	LeadID      int64                  `json:"lead_id"`
	Segment     string                 `json:"segment"` // high_engagement, medium_engagement, low_engagement, dormant, hot_lead, cold_lead
	SegmentData map[string]interface{} `json:"segment_data" gorm:"type:jsonb"`
	AssignedAt  time.Time              `json:"assigned_at" gorm:"autoCreateTime"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// TableName specifies the table name for GORM
func (BehavioralSegment) TableName() string {
	return "behavioral_segments"
}

// ============================================================================
// BEHAVIORAL COHORTS
// ============================================================================

// BehavioralCohort represents a cohort for retention analysis
type BehavioralCohort struct {
	ID         int64                  `json:"id" gorm:"primaryKey"`
	CohortName string                 `json:"cohort_name"` // e.g., "2025-W44"
	CohortType string                 `json:"cohort_type"` // weekly, monthly, campaign, source
	CohortDate time.Time              `json:"cohort_date"`
	LeadCount  int                    `json:"lead_count"`
	Metadata   map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (BehavioralCohort) TableName() string {
	return "behavioral_cohorts"
}

// BehavioralCohortMember links leads to cohorts
type BehavioralCohortMember struct {
	ID       int64     `json:"id" gorm:"primaryKey"`
	CohortID int64     `json:"cohort_id"`
	LeadID   int64     `json:"lead_id"`
	JoinedAt time.Time `json:"joined_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (BehavioralCohortMember) TableName() string {
	return "behavioral_cohort_members"
}

// ============================================================================
// BEHAVIORAL ANOMALIES
// ============================================================================

// BehavioralAnomaly represents a detected anomaly in behavioral patterns
type BehavioralAnomaly struct {
	ID                   int64      `json:"id" gorm:"primaryKey"`
	AnomalyType          string     `json:"anomaly_type"` // engagement_drop, conversion_spike, unusual_activity, etc.
	Severity             string     `json:"severity"`     // critical, high, medium, low
	Description          string     `json:"description"`
	AffectedEntity       string     `json:"affected_entity"` // lead, property, agent, system
	EntityID             int        `json:"entity_id"`
	MetricName           string     `json:"metric_name"`
	ExpectedValue        float64    `json:"expected_value"`
	ActualValue          float64    `json:"actual_value"`
	DeviationPercentage  float64    `json:"deviation_percentage"`
	DetectionMethod      string     `json:"detection_method"` // statistical, ml, rule_based
	DetectedAt           time.Time  `json:"detected_at" gorm:"autoCreateTime"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes      string     `json:"resolution_notes,omitempty"`
}

// TableName specifies the table name for GORM
func (BehavioralAnomaly) TableName() string {
	return "behavioral_anomalies"
}

// ============================================================================
// HELPER STRUCTS FOR API RESPONSES
// ============================================================================

// BehavioralTrendPoint represents a single point in a trend chart
type BehavioralTrendPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
	Label string  `json:"label,omitempty"`
}

// BehavioralFunnelStage represents a stage in the conversion funnel
type BehavioralFunnelStage struct {
	Stage           string  `json:"stage"`
	Count           int     `json:"count"`
	Percentage      float64 `json:"percentage"`
	ConversionRate  float64 `json:"conversion_rate,omitempty"`
	AvgTimeInStage  int     `json:"avg_time_in_stage,omitempty"` // seconds
}

// BehavioralSegmentSummary represents a summary of a behavioral segment
type BehavioralSegmentSummary struct {
	Segment            string  `json:"segment"`
	LeadCount          int     `json:"lead_count"`
	ConversionRate     float64 `json:"conversion_rate"`
	AvgEngagementScore float64 `json:"avg_engagement_score"`
	AvgSessionLength   int     `json:"avg_session_length"` // seconds
}

// BehavioralHeatmapCell represents a cell in the activity heatmap
type BehavioralHeatmapCell struct {
	DayOfWeek int     `json:"day_of_week"` // 0=Sunday, 6=Saturday
	Hour      int     `json:"hour"`        // 0-23
	Activity  int     `json:"activity"`    // Count of events
	Intensity float64 `json:"intensity"`   // 0.0 to 1.0
}

// BehavioralCohortRetention represents retention data for a cohort
type BehavioralCohortRetention struct {
	CohortName string    `json:"cohort_name"`
	CohortDate time.Time `json:"cohort_date"`
	Week0      int       `json:"week_0"` // Initial size
	Week1      int       `json:"week_1"`
	Week2      int       `json:"week_2"`
	Week3      int       `json:"week_3"`
	Week4      int       `json:"week_4"`
}
