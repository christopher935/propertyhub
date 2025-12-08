package models

import (
	"time"
)

// PropertyValuationRecord represents a stored valuation report in the database
type PropertyValuationRecord struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PropertyID       *uint     `gorm:"index" json:"property_id"`
	EstimatedValue   float64   `gorm:"type:decimal(12,2);not null" json:"estimated_value"`
	ValueLow         float64   `gorm:"type:decimal(12,2);not null" json:"value_low"`
	ValueHigh        float64   `gorm:"type:decimal(12,2);not null" json:"value_high"`
	PricePerSqft     *float64  `gorm:"type:decimal(8,2)" json:"price_per_sqft"`
	Confidence       float64   `gorm:"type:decimal(5,2);not null" json:"confidence"`
	Comparables      JSONB     `gorm:"type:jsonb" json:"comparables"`
	Adjustments      JSONB     `gorm:"type:jsonb" json:"adjustments"`
	MarketAnalysis   JSONB     `gorm:"type:jsonb" json:"market_analysis"`
	ValuationFactors JSONB     `gorm:"type:jsonb" json:"valuation_factors"`
	Recommendations  JSONB     `gorm:"type:jsonb" json:"recommendations"`
	RequestedBy      string    `json:"requested_by"`
	ModelVersion     string    `gorm:"default:'v1.0'" json:"model_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (PropertyValuationRecord) TableName() string {
	return "property_valuations"
}
