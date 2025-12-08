package repositories

import (
	"chrisgross-ctrl-project/internal/models"
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

// PropertyRepositoryImpl implements PropertyRepository interface
type PropertyRepositoryImpl struct {
	db *gorm.DB
}

// NewPropertyRepository creates a new PropertyRepository instance
func NewPropertyRepository(db *gorm.DB) PropertyRepository {
	return &PropertyRepositoryImpl{db: db}
}

// Create creates a new property
func (r *PropertyRepositoryImpl) Create(ctx context.Context, entity interface{}) error {
	property, ok := entity.(*models.Property)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Property")
	}
	return r.db.WithContext(ctx).Create(property).Error
}

// Update updates an existing property
func (r *PropertyRepositoryImpl) Update(ctx context.Context, entity interface{}) error {
	property, ok := entity.(*models.Property)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Property")
	}
	return r.db.WithContext(ctx).Save(property).Error
}

// Delete soft deletes a property by ID
func (r *PropertyRepositoryImpl) Delete(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).Delete(&models.Property{}, id).Error
}

// FindByID finds a property by ID
func (r *PropertyRepositoryImpl) FindByID(ctx context.Context, id interface{}, entity interface{}) error {
	property, ok := entity.(*models.Property)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Property")
	}
	return r.db.WithContext(ctx).First(property, id).Error
}

// FindAll finds all properties
func (r *PropertyRepositoryImpl) FindAll(ctx context.Context, entities interface{}) error {
	properties, ok := entities.(*[]*models.Property)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *[]*models.Property")
	}
	return r.db.WithContext(ctx).Find(properties).Error
}

// FindByMLSID finds a property by MLS ID with optimized query
func (r *PropertyRepositoryImpl) FindByMLSID(ctx context.Context, mlsID string) (*models.Property, error) {
	var property models.Property
	// Use index idx_properties_mls_id for optimized performance
	err := r.db.WithContext(ctx).Where("mls_id = ?", mlsID).First(&property).Error
	if err != nil {
		return nil, err
	}
	return &property, nil
}

// Search performs advanced property search with optimized queries
func (r *PropertyRepositoryImpl) Search(ctx context.Context, criteria PropertySearchCriteria) ([]*models.Property, int64, error) {
	var properties []*models.Property
	var totalCount int64

	// Build base query with optimized joins and indexes
	query := r.db.WithContext(ctx).Model(&models.Property{})

	// Apply filters with indexed fields first for optimal performance
	if criteria.City != "" {
		// Use idx_properties_city if available
		query = query.Where("city ILIKE ?", "%"+criteria.City+"%")
	}

	if criteria.Status != "" {
		// Use idx_properties_status if available
		query = query.Where("status = ?", criteria.Status)
	}

	// Price range filters - use compound index idx_properties_price_status
	if criteria.MinPrice != nil {
		query = query.Where("price >= ?", *criteria.MinPrice)
	}
	if criteria.MaxPrice != nil {
		query = query.Where("price <= ?", *criteria.MaxPrice)
	}

	// Property specifications
	if criteria.Bedrooms != nil {
		query = query.Where("bedrooms >= ?", *criteria.Bedrooms)
	}
	if criteria.Bathrooms != nil {
		query = query.Where("bathrooms >= ?", *criteria.Bathrooms)
	}
	if criteria.PropertyType != "" {
		query = query.Where("property_type = ?", criteria.PropertyType)
	}

	// General search across multiple fields - use GIN index if available
	if criteria.Search != "" {
		searchTerm := "%" + strings.ToLower(criteria.Search) + "%"
		query = query.Where(
			"LOWER(address) LIKE ? OR LOWER(description) LIKE ? OR LOWER(property_features) LIKE ? OR mls_id LIKE ?",
			searchTerm, searchTerm, searchTerm, "%"+criteria.Search+"%",
		)
	}

	// Get total count before pagination
	countQuery := query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count properties: %v", err)
	}

	// Apply pagination and ordering with SQL injection protection
	orderBy := "updated_at"
	orderDir := "DESC"

	if criteria.OrderBy != "" {
		// Whitelist allowed columns for ORDER BY to prevent SQL injection
		allowedColumns := map[string]bool{
			"id":            true,
			"price":         true,
			"created_at":    true,
			"updated_at":    true,
			"city":          true,
			"status":        true,
			"bedrooms":      true,
			"bathrooms":     true,
			"square_feet":   true,
			"property_type": true,
			"mls_id":        true,
			"address":       true,
			"year_built":    true,
		}

		if allowedColumns[criteria.OrderBy] {
			orderBy = criteria.OrderBy
		}
		// If not in whitelist, keep default "updated_at"
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

	// Execute query
	if err := query.Find(&properties).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search properties: %v", err)
	}

	return properties, totalCount, nil
}

// GetStatistics returns comprehensive property statistics using optimized queries
func (r *PropertyRepositoryImpl) GetStatistics(ctx context.Context) (*PropertyStatistics, error) {
	var stats PropertyStatistics

	// Use indexed queries for better performance
	// Total properties count
	if err := r.db.WithContext(ctx).Model(&models.Property{}).Count(&stats.TotalProperties).Error; err != nil {
		return nil, fmt.Errorf("failed to count total properties: %v", err)
	}

	// Active properties count - use idx_properties_status
	if err := r.db.WithContext(ctx).Model(&models.Property{}).
		Where("status = ?", "active").Count(&stats.ActiveProperties).Error; err != nil {
		return nil, fmt.Errorf("failed to count active properties: %v", err)
	}

	// Inactive properties count
	stats.InactiveProperties = stats.TotalProperties - stats.ActiveProperties

	// Average and total price calculations - optimize with aggregate query
	var result struct {
		AvgPrice   float64
		TotalValue float64
	}

	if err := r.db.WithContext(ctx).Model(&models.Property{}).
		Select("AVG(price) as avg_price, SUM(price) as total_value").
		Where("price > 0").
		Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate price statistics: %v", err)
	}

	stats.AvgPrice = result.AvgPrice
	stats.TotalValue = result.TotalValue

	// Cities covered - use DISTINCT with index
	if err := r.db.WithContext(ctx).Model(&models.Property{}).
		Distinct("city").Count(&stats.CitiesCovered).Error; err != nil {
		return nil, fmt.Errorf("failed to count cities: %v", err)
	}

	return &stats, nil
}

// GetCities returns list of unique cities with properties
func (r *PropertyRepositoryImpl) GetCities(ctx context.Context) ([]string, error) {
	var cities []string

	// Use index on city column for optimization
	err := r.db.WithContext(ctx).Model(&models.Property{}).
		Distinct("city").
		Where("city != '' AND city IS NOT NULL").
		Order("city ASC").
		Pluck("city", &cities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get cities: %v", err)
	}

	return cities, nil
}

// UpdateStatusBatch updates status for multiple properties in a single transaction
func (r *PropertyRepositoryImpl) UpdateStatusBatch(ctx context.Context, ids []uint, status string) error {
	if len(ids) == 0 {
		return nil
	}

	// Use transaction for consistency and batch update for performance
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(&models.Property{}).
			Where("id IN ?", ids).
			Update("status", status).Error
	})
}

// DeleteBatch soft deletes multiple properties efficiently
func (r *PropertyRepositoryImpl) DeleteBatch(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Delete(&models.Property{}, ids).Error
}

// GetFeaturedProperties returns featured properties with optimized query
func (r *PropertyRepositoryImpl) GetFeaturedProperties(ctx context.Context, limit int) ([]*models.Property, error) {
	var properties []*models.Property

	if limit <= 0 || limit > 50 {
		limit = 10 // Default limit
	}

	// Use compound index for optimal performance
	err := r.db.WithContext(ctx).
		Where("status = ? AND price > 0", "active").
		Order("view_count DESC, updated_at DESC").
		Limit(limit).
		Find(&properties).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get featured properties: %v", err)
	}

	return properties, nil
}

// GetRecentProperties returns recently added/updated properties
func (r *PropertyRepositoryImpl) GetRecentProperties(ctx context.Context, limit int) ([]*models.Property, error) {
	var properties []*models.Property

	if limit <= 0 || limit > 50 {
		limit = 10 // Default limit
	}

	// Use idx_properties_updated_at for optimal performance
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("updated_at DESC").
		Limit(limit).
		Find(&properties).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent properties: %v", err)
	}

	return properties, nil
}

// GetPropertyWithPhotos returns property with eagerly loaded relationships
func (r *PropertyRepositoryImpl) GetPropertyWithPhotos(ctx context.Context, id uint) (*models.Property, error) {
	var property models.Property

	// Use optimized preload to avoid N+1 queries
	err := r.db.WithContext(ctx).
		Preload("Bookings", func(db *gorm.DB) *gorm.DB {
			return db.Order("showing_date DESC").Limit(5) // Only recent bookings
		}).
		First(&property, id).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get property with photos: %v", err)
	}

	return &property, nil
}
