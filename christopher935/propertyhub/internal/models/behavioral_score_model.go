package models

import "time"

// BehavioralScore represents the current behavioral score for a lead
// Schema matches existing behavioral_scores table in database
type BehavioralScore struct {
	ID              string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	VisitorID       *string    `json:"visitor_id,omitempty"`
	LeadID          *int       `json:"lead_id,omitempty"`
	EngagementScore int        `json:"engagement_score" gorm:"default:0;not null"`
	FinancialScore  int        `json:"financial_score" gorm:"default:0;not null"`
	UrgencyScore    int        `json:"urgency_score" gorm:"default:0;not null"`
	CompositeScore  int        `json:"composite_score" gorm:"default:0;not null"`
	ScoreFactors    JSONB      `json:"score_factors" gorm:"type:jsonb"`
	LastCalculated  time.Time  `json:"last_calculated" gorm:"default:now();not null"`
	CreatedAt       time.Time  `json:"created_at" gorm:"default:now();not null"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"default:now();not null"`
}

// TableName specifies the table name for GORM
func (BehavioralScore) TableName() string {
	return "behavioral_scores"
}

// GetSegment returns the segment based on composite score
func (s *BehavioralScore) GetSegment() string {
	if s.CompositeScore >= 70 {
		return "hot"
	} else if s.CompositeScore >= 40 {
		return "warm"
	} else if s.CompositeScore >= 10 {
		return "cold"
	}
	return "dormant"
}

// IsHot returns true if the lead is in the "hot" segment
func (s *BehavioralScore) IsHot() bool {
	return s.CompositeScore >= 70
}

// IsWarm returns true if the lead is in the "warm" segment
func (s *BehavioralScore) IsWarm() bool {
	return s.CompositeScore >= 40 && s.CompositeScore < 70
}

// IsCold returns true if the lead is in the "cold" segment
func (s *BehavioralScore) IsCold() bool {
	return s.CompositeScore >= 10 && s.CompositeScore < 40
}

// IsDormant returns true if the lead is in the "dormant" segment
func (s *BehavioralScore) IsDormant() bool {
	return s.CompositeScore < 10
}
