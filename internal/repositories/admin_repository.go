package repositories

import (
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/security"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// AdminRepositoryImpl implements AdminRepository interface
type AdminRepositoryImpl struct {
	db *gorm.DB
}

// NewAdminRepository creates a new AdminRepository instance
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &AdminRepositoryImpl{db: db}
}

// Create creates a new admin user
func (r *AdminRepositoryImpl) Create(ctx context.Context, entity interface{}) error {
	admin, ok := entity.(*models.AdminUser)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.AdminUser")
	}
	return r.db.WithContext(ctx).Create(admin).Error
}

// Update updates an existing admin user
func (r *AdminRepositoryImpl) Update(ctx context.Context, entity interface{}) error {
	admin, ok := entity.(*models.AdminUser)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.AdminUser")
	}
	return r.db.WithContext(ctx).Save(admin).Error
}

// Delete soft deletes an admin user by ID
func (r *AdminRepositoryImpl) Delete(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).Delete(&models.AdminUser{}, id).Error
}

// FindByID finds an admin user by ID
func (r *AdminRepositoryImpl) FindByID(ctx context.Context, id interface{}, entity interface{}) error {
	admin, ok := entity.(*models.AdminUser)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.AdminUser")
	}
	return r.db.WithContext(ctx).First(admin, "id = ?", id).Error
}

// FindAll finds all admin users
func (r *AdminRepositoryImpl) FindAll(ctx context.Context, entities interface{}) error {
	admins, ok := entities.(*[]*models.AdminUser)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *[]*models.AdminUser")
	}
	return r.db.WithContext(ctx).Find(admins).Error
}

// FindActiveByEmail finds an active admin user by email with optimized query
func (r *AdminRepositoryImpl) FindActiveByEmail(ctx context.Context, email string) (*models.AdminUser, error) {
	var admin models.AdminUser

	// Use compound index idx_admin_users_email_active for optimization
	err := r.db.WithContext(ctx).
		Where("email = ? AND active = ?", email, true).
		First(&admin).Error

	if err != nil {
		return nil, err
	}

	return &admin, nil
}

// UpdateLoginStats updates login statistics for an admin user
func (r *AdminRepositoryImpl) UpdateLoginStats(ctx context.Context, adminID string) error {
	now := time.Now()

	// Use atomic update to increment login count and update last login
	result := r.db.WithContext(ctx).Model(&models.AdminUser{}).
		Where("id = ?", adminID).
		Updates(map[string]interface{}{
			"last_login":  now,
			"login_count": gorm.Expr("login_count + ?", 1),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update login stats: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("admin user with ID %s not found", adminID)
	}

	return nil
}

// GetDashboardMetrics returns dashboard metrics for admin interface
func (r *AdminRepositoryImpl) GetDashboardMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get total properties count
	var totalProperties int64
	if err := r.db.WithContext(ctx).Model(&models.Property{}).Count(&totalProperties).Error; err != nil {
		return nil, fmt.Errorf("failed to count properties: %v", err)
	}
	metrics["total_properties"] = totalProperties

	// Get active properties count
	var activeProperties int64
	if err := r.db.WithContext(ctx).Model(&models.Property{}).
		Where("status = ?", "active").Count(&activeProperties).Error; err != nil {
		return nil, fmt.Errorf("failed to count active properties: %v", err)
	}
	metrics["active_properties"] = activeProperties

	// Get total bookings count
	var totalBookings int64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).Count(&totalBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count bookings: %v", err)
	}
	metrics["total_bookings"] = totalBookings

	// Get pending bookings count
	var pendingBookings int64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("status IN ?", []string{"scheduled", "pending"}).Count(&pendingBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending bookings: %v", err)
	}
	metrics["pending_bookings"] = pendingBookings

	// Get new contacts count (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var newContacts int64
	if err := r.db.WithContext(ctx).Model(&models.Contact{}).
		Where("created_at >= ?", thirtyDaysAgo).Count(&newContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to count new contacts: %v", err)
	}
	metrics["new_contacts"] = newContacts

	// Get total contacts count
	var totalContacts int64
	if err := r.db.WithContext(ctx).Model(&models.Contact{}).Count(&totalContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total contacts: %v", err)
	}
	metrics["total_contacts"] = totalContacts

	// Get average property price
	var avgPrice struct {
		AvgPrice float64 `json:"avg_price"`
	}
	if err := r.db.WithContext(ctx).Model(&models.Property{}).
		Select("AVG(price) as avg_price").
		Where("price > 0").
		Scan(&avgPrice).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average price: %v", err)
	}
	metrics["avg_property_price"] = avgPrice.AvgPrice

	// Get recent activity metrics
	today := time.Now().Truncate(24 * time.Hour)

	// Today's bookings
	var todayBookings int64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("DATE(showing_date) = DATE(?)", today).Count(&todayBookings).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's bookings: %v", err)
	}
	metrics["today_bookings"] = todayBookings

	// Today's contacts
	var todayContacts int64
	if err := r.db.WithContext(ctx).Model(&models.Contact{}).
		Where("DATE(created_at) = DATE(?)", today).Count(&todayContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's contacts: %v", err)
	}
	metrics["today_contacts"] = todayContacts

	return metrics, nil
}

// CreateUser creates a new admin user with hashed password
func (r *AdminRepositoryImpl) CreateUser(ctx context.Context, username, email, password, role string) (*models.AdminUser, error) {
	// Validate inputs
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username, email, and password are required")
	}

	if role == "" {
		role = models.RoleUser // Default role
	}

	// Check if user already exists
	var existingUser models.AdminUser
	err := r.db.WithContext(ctx).Where("email = ? OR username = ?", email, username).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("user with email %s or username %s already exists", email, username)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Generate unique ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %v", err)
	}
	userID := hex.EncodeToString(idBytes)

	// Create new user
	newUser := &models.AdminUser{
		ID:           userID,
		Username:     username,
		Email:        security.EncryptedString(email), // Cast to EncryptedString type
		PasswordHash: string(hashedPassword),
		Role:         role,
		Active:       true,
		LoginCount:   0,
	}

	if err := r.db.WithContext(ctx).Create(newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return newUser, nil
}

// UpdateUser updates user information with selective field updates
func (r *AdminRepositoryImpl) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	// Validate userID
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Hash password if provided
	if password, ok := updates["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}
		updates["password_hash"] = string(hashedPassword)
		delete(updates, "password") // Remove plain password
	}

	// Perform update
	result := r.db.WithContext(ctx).Model(&models.AdminUser{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}

// DeactivateUser sets a user as inactive instead of deleting
func (r *AdminRepositoryImpl) DeactivateUser(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Model(&models.AdminUser{}).
		Where("id = ?", userID).
		Update("active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate user: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}

// GetAllUsers returns all admin users ordered by username
func (r *AdminRepositoryImpl) GetAllUsers(ctx context.Context) ([]*models.AdminUser, error) {
	var users []*models.AdminUser

	err := r.db.WithContext(ctx).
		Order("username ASC").
		Find(&users).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %v", err)
	}

	return users, nil
}
