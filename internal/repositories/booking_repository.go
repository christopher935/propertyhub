package repositories

import (
	"chrisgross-ctrl-project/internal/models"
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

// BookingRepositoryImpl implements BookingRepository interface
type BookingRepositoryImpl struct {
	db *gorm.DB
}

// NewBookingRepository creates a new BookingRepository instance
func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &BookingRepositoryImpl{db: db}
}

// Create creates a new booking
func (r *BookingRepositoryImpl) Create(ctx context.Context, entity interface{}) error {
	booking, ok := entity.(*models.Booking)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Booking")
	}
	return r.db.WithContext(ctx).Create(booking).Error
}

// Update updates an existing booking
func (r *BookingRepositoryImpl) Update(ctx context.Context, entity interface{}) error {
	booking, ok := entity.(*models.Booking)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Booking")
	}
	return r.db.WithContext(ctx).Save(booking).Error
}

// Delete soft deletes a booking by ID
func (r *BookingRepositoryImpl) Delete(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).Delete(&models.Booking{}, id).Error
}

// FindByID finds a booking by ID
func (r *BookingRepositoryImpl) FindByID(ctx context.Context, id interface{}, entity interface{}) error {
	booking, ok := entity.(*models.Booking)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Booking")
	}
	return r.db.WithContext(ctx).First(booking, id).Error
}

// FindAll finds all bookings
func (r *BookingRepositoryImpl) FindAll(ctx context.Context, entities interface{}) error {
	bookings, ok := entities.(*[]*models.Booking)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *[]*models.Booking")
	}
	return r.db.WithContext(ctx).Find(bookings).Error
}

// List performs advanced booking search with filters and pagination
func (r *BookingRepositoryImpl) List(ctx context.Context, criteria BookingFilterCriteria) ([]*models.Booking, int64, error) {
	var bookings []*models.Booking
	var totalCount int64

	// Build optimized query with indexed filters
	query := r.db.WithContext(ctx).Model(&models.Booking{})

	// Apply indexed filters first for optimal performance
	if criteria.Status != "" {
		// Use idx_bookings_status if available
		query = query.Where("status = ?", criteria.Status)
	}

	if criteria.PropertyMLSID != "" {
		// Join with properties table using foreign key index
		query = query.Joins("JOIN properties ON properties.id = bookings.property_id").
			Where("properties.mls_id = ?", criteria.PropertyMLSID)
	}

	if criteria.CustomerEmail != "" {
		// Use idx_bookings_email if available
		query = query.Where("email = ?", criteria.CustomerEmail)
	}

	// Date range filters - use compound index idx_bookings_showing_date
	if criteria.DateFrom != "" {
		if dateFrom, err := time.Parse("2006-01-02", criteria.DateFrom); err == nil {
			query = query.Where("showing_date >= ?", dateFrom)
		}
	}
	if criteria.DateTo != "" {
		if dateTo, err := time.Parse("2006-01-02", criteria.DateTo); err == nil {
			// Add 24 hours to include the entire day
			dateTo = dateTo.Add(24 * time.Hour)
			query = query.Where("showing_date < ?", dateTo)
		}
	}

	// General search across multiple fields
	if criteria.Search != "" {
		searchTerm := "%" + strings.ToLower(criteria.Search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(reference_number) LIKE ? OR LOWER(fub_lead_id) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	// Get total count before pagination
	countQuery := query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count bookings: %v", err)
	}

	// Apply pagination and ordering
	orderBy := "showing_date"
	orderDir := "DESC"

	if criteria.OrderBy != "" {
		orderBy = criteria.OrderBy
	}
	if criteria.OrderDir != "" && (criteria.OrderDir == "ASC" || criteria.OrderDir == "DESC") {
		orderDir = criteria.OrderDir
	}

	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

	// Apply pagination with reasonable limits
	pageSize := criteria.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20 // Default page size
	}

	page := criteria.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize
	query = query.Limit(pageSize).Offset(offset)

	// Execute query with property preloading to avoid N+1 queries
	if err := query.Preload("Property").Find(&bookings).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list bookings: %v", err)
	}

	return bookings, totalCount, nil
}

// GetStatistics returns comprehensive booking statistics using optimized queries
func (r *BookingRepositoryImpl) GetStatistics(ctx context.Context) (*BookingStatistics, error) {
	var stats BookingStatistics

	// Use indexed queries for better performance
	// Total bookings count
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).Count(&stats.TotalBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count total bookings: %v", err)
	}

	// Status-based counts - use idx_bookings_status for optimization
	statusCounts := make(map[string]int64)
	rows, err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Select("status, COUNT(*) as count").
		Group("status").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err == nil {
			statusCounts[status] = count
		}
	}

	// Map status counts
	stats.PendingBookings = statusCounts["scheduled"] + statusCounts["pending"]
	stats.ConfirmedBookings = statusCounts["confirmed"]
	stats.CompletedBookings = statusCounts["completed"]
	stats.CancelledBookings = statusCounts["cancelled"]

	// Time-based statistics - use idx_bookings_showing_date for optimization
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Today's bookings
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("showing_date >= ? AND showing_date < ?", today, today.Add(24*time.Hour)).
		Count(&stats.TodayBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count today bookings: %v", err)
	}

	// This week's bookings
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("showing_date >= ? AND showing_date < ?", weekStart, weekStart.Add(7*24*time.Hour)).
		Count(&stats.WeekBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count week bookings: %v", err)
	}

	// This month's bookings
	nextMonth := monthStart.AddDate(0, 1, 0)
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("showing_date >= ? AND showing_date < ?", monthStart, nextMonth).
		Count(&stats.MonthBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count month bookings: %v", err)
	}

	return &stats, nil
}

// GetByPropertyMLSID returns all bookings for a specific property using MLS ID
func (r *BookingRepositoryImpl) GetByPropertyMLSID(ctx context.Context, mlsID string) ([]*models.Booking, error) {
	var bookings []*models.Booking

	// Use JOIN with property table and MLS ID index for optimization
	err := r.db.WithContext(ctx).
		Joins("JOIN properties ON properties.id = bookings.property_id").
		Where("properties.mls_id = ?", mlsID).
		Order("showing_date DESC").
		Preload("Property").
		Find(&bookings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get bookings by MLS ID: %v", err)
	}

	return bookings, nil
}

// GetMetricsByDateRange returns booking metrics for a specific date range
func (r *BookingRepositoryImpl) GetMetricsByDateRange(ctx context.Context, startDate, endDate string) (*BookingStatistics, error) {
	var stats BookingStatistics

	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %v", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %v", err)
	}
	end = end.Add(24 * time.Hour) // Include the end date

	// Base query with date range filter
	baseQuery := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("showing_date >= ? AND showing_date < ?", start, end)

	// Total bookings in range
	if err := baseQuery.Count(&stats.TotalBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count total bookings: %v", err)
	}

	// Status-based counts for the date range
	statusCounts := make(map[string]int64)
	rows, err := baseQuery.Select("status, COUNT(*) as count").Group("status").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts by date range: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err == nil {
			statusCounts[status] = count
		}
	}

	// Map status counts
	stats.PendingBookings = statusCounts["scheduled"] + statusCounts["pending"]
	stats.ConfirmedBookings = statusCounts["confirmed"]
	stats.CompletedBookings = statusCounts["completed"]
	stats.CancelledBookings = statusCounts["cancelled"]

	return &stats, nil
}

// GetBookingTrends returns booking trends for the last N days
func (r *BookingRepositoryImpl) GetBookingTrends(ctx context.Context, days int) (map[string]int64, error) {
	if days <= 0 || days > 365 {
		days = 30 // Default to 30 days
	}

	trends := make(map[string]int64)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Query bookings grouped by date
	query := `
		SELECT DATE(showing_date) as booking_date, COUNT(*) as count
		FROM bookings 
		WHERE showing_date >= ? AND showing_date <= ?
		GROUP BY DATE(showing_date)
		ORDER BY booking_date ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get booking trends: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var date time.Time
		var count int64
		if err := rows.Scan(&date, &count); err == nil {
			dateStr := date.Format("2006-01-02")
			trends[dateStr] = count
		}
	}

	return trends, nil
}

// GetBookingsWithProperty returns bookings with property details to avoid N+1 queries
func (r *BookingRepositoryImpl) GetBookingsWithProperty(ctx context.Context, criteria BookingFilterCriteria) ([]*models.Booking, int64, error) {
	// This is essentially the same as List() method but ensures property preloading
	// We can just call List() which already includes property preloading
	return r.List(ctx, criteria)
}

// CheckPropertyBookingCount returns the booking count for a specific property
func (r *BookingRepositoryImpl) CheckPropertyBookingCount(ctx context.Context, propertyID uint) (int64, error) {
	var count int64

	// Use foreign key index for optimization
	err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("property_id = ?", propertyID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count property bookings: %v", err)
	}

	return count, nil
}
