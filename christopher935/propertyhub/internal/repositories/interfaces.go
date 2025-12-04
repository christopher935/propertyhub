package repositories

import (
	"context"
	"chrisgross-ctrl-project/internal/models"
)

// BaseRepository defines common database operations
type BaseRepository interface {
	Create(ctx context.Context, entity interface{}) error
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id interface{}) error
	FindByID(ctx context.Context, id interface{}, entity interface{}) error
	FindAll(ctx context.Context, entities interface{}) error
}

// PaginationOptions represents pagination parameters
type PaginationOptions struct {
	Page     int
	PageSize int
	OrderBy  string
	OrderDir string // "ASC" or "DESC"
}

// FilterOptions represents common filtering options
type FilterOptions struct {
	Search   string
	Status   string
	DateFrom string
	DateTo   string
}

// PropertySearchCriteria represents property search parameters
type PropertySearchCriteria struct {
	Search       string
	City         string
	Status       string
	MinPrice     *float64
	MaxPrice     *float64
	Bedrooms     *int
	Bathrooms    *float64
	PropertyType string
	PaginationOptions
}

// PropertyStatistics represents property metrics
type PropertyStatistics struct {
	TotalProperties    int64
	ActiveProperties   int64
	InactiveProperties int64
	AvgPrice           float64
	TotalValue         float64
	CitiesCovered      int64
}

// PropertyRepository interface defines property-specific database operations
type PropertyRepository interface {
	BaseRepository

	// Core property operations
	FindByMLSID(ctx context.Context, mlsID string) (*models.Property, error)
	Search(ctx context.Context, criteria PropertySearchCriteria) ([]*models.Property, int64, error)
	GetStatistics(ctx context.Context) (*PropertyStatistics, error)
	GetCities(ctx context.Context) ([]string, error)

	// Batch operations
	UpdateStatusBatch(ctx context.Context, ids []uint, status string) error
	DeleteBatch(ctx context.Context, ids []uint) error

	// Advanced queries
	GetFeaturedProperties(ctx context.Context, limit int) ([]*models.Property, error)
	GetRecentProperties(ctx context.Context, limit int) ([]*models.Property, error)
	GetPropertyWithPhotos(ctx context.Context, id uint) (*models.Property, error)
}

// BookingStatistics represents booking metrics
type BookingStatistics struct {
	TotalBookings     int64
	PendingBookings   int64
	ConfirmedBookings int64
	CompletedBookings int64
	CancelledBookings int64
	TodayBookings     int64
	WeekBookings      int64
	MonthBookings     int64
}

// BookingFilterCriteria represents booking search parameters
type BookingFilterCriteria struct {
	Status        string
	PropertyMLSID string
	CustomerEmail string
	DateFrom      string
	DateTo        string
	Search        string // For cross-field search
	PaginationOptions
}

// BookingRepository interface defines booking-specific database operations
type BookingRepository interface {
	BaseRepository

	// Core booking operations
	List(ctx context.Context, criteria BookingFilterCriteria) ([]*models.Booking, int64, error)
	GetStatistics(ctx context.Context) (*BookingStatistics, error)
	GetByPropertyMLSID(ctx context.Context, mlsID string) ([]*models.Booking, error)

	// Analytics
	GetMetricsByDateRange(ctx context.Context, startDate, endDate string) (*BookingStatistics, error)
	GetBookingTrends(ctx context.Context, days int) (map[string]int64, error)

	// Relationship queries
	GetBookingsWithProperty(ctx context.Context, criteria BookingFilterCriteria) ([]*models.Booking, int64, error)
	CheckPropertyBookingCount(ctx context.Context, propertyID uint) (int64, error)
}

// ContactRepository interface defines contact-specific database operations
type ContactRepository interface {
	BaseRepository

	// Core contact operations
	FindByEmail(ctx context.Context, email string) (*models.Contact, error)
	CreateOrUpdate(ctx context.Context, contact *models.Contact) error

	// Lead management
	GetRecentLeads(ctx context.Context, limit int) ([]*models.Contact, error)
	GetLeadsBySource(ctx context.Context, source string) ([]*models.Contact, error)
	UpdateLeadStatus(ctx context.Context, contactID uint, status string) error
}

// AdminRepository interface defines admin user operations
type AdminRepository interface {
	BaseRepository

	// Authentication operations
	FindActiveByEmail(ctx context.Context, email string) (*models.AdminUser, error)
	UpdateLoginStats(ctx context.Context, adminID string) error
	GetDashboardMetrics(ctx context.Context) (map[string]interface{}, error)

	// User management
	CreateUser(ctx context.Context, username, email, password, role string) (*models.AdminUser, error)
	UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error
	DeactivateUser(ctx context.Context, userID string) error
	GetAllUsers(ctx context.Context) ([]*models.AdminUser, error)
}
