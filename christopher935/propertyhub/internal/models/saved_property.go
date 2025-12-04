package models

import (
	"time"
)

type SavedProperty struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SessionID  string    `json:"session_id" gorm:"index;not null"`
	PropertyID uint      `json:"property_id" gorm:"index;not null"`
	Email      string    `json:"email" gorm:"index"`
	SavedAt    time.Time `json:"saved_at" gorm:"not null;index"`
	Notes      string    `json:"notes" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	Property   Property  `json:"property,omitempty" gorm:"foreignKey:PropertyID"`
}

func (SavedProperty) TableName() string {
	return "saved_properties"
}
