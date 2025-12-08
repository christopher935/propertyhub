package services

import (
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// PreListingService handles pre-listing business logic
type PreListingService struct {
	db *gorm.DB
}

// PreListingStats represents dashboard statistics
type PreListingStats struct {
	TotalItems        int64   `json:"total_items"`
	PendingItems      int64   `json:"pending_items"`
	CompletedItems    int64   `json:"completed_items"`
	OverdueItems      int64   `json:"overdue_items"`
	AvgCompletionDays float64 `json:"avg_completion_days"`
	SuccessRate       float32 `json:"success_rate"`
}

// NewPreListingService creates a new pre-listing service
func NewPreListingService(db *gorm.DB) *PreListingService {
	return &PreListingService{
		db: db,
	}
}

// GetPreListingStats returns dashboard statistics
func (pls *PreListingService) GetPreListingStats() (*PreListingStats, error) {
	var stats PreListingStats

	pls.db.Model(&models.PreListingItem{}).Count(&stats.TotalItems)
	pls.db.Model(&models.PreListingItem{}).Where("status NOT IN (?)", []string{"completed", "cancelled"}).Count(&stats.PendingItems)
	pls.db.Model(&models.PreListingItem{}).Where("status = ?", "completed").Count(&stats.CompletedItems)
	pls.db.Model(&models.PreListingItem{}).Where("is_overdue = ?", true).Count(&stats.OverdueItems)

	if stats.TotalItems > 0 {
		stats.SuccessRate = float32(stats.CompletedItems) / float32(stats.TotalItems) * 100
	}

	stats.AvgCompletionDays = 7.5 // Mock average

	return &stats, nil
}

// GetPreListingItems returns paginated list of pre-listing items
func (pls *PreListingService) GetPreListingItems(page, limit int, status string) ([]models.PreListingItem, int64, error) {
	var items []models.PreListingItem
	var total int64

	query := pls.db.Model(&models.PreListingItem{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("EmailRecords").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error

	return items, total, err
}

// CheckOverdueItems checks for overdue items and marks them
func (pls *PreListingService) CheckOverdueItems() error {
	// Mark items as overdue if they've been pending for more than 14 days
	cutoffDate := time.Now().AddDate(0, 0, -14)

	return pls.db.Model(&models.PreListingItem{}).
		Where("created_at < ? AND status NOT IN (?) AND is_overdue = ?",
			cutoffDate, []string{"completed", "cancelled"}, false).
		Update("is_overdue", true).Error
}
