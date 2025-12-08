package repositories

import (
	"gorm.io/gorm"
)

// Repositories holds all repository instances for dependency injection
type Repositories struct {
	Property PropertyRepository
	Booking  BookingRepository
	Contact  ContactRepository
	Admin    AdminRepository
}

// NewRepositories creates and initializes all repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Property: NewPropertyRepository(db),
		Booking:  NewBookingRepository(db),
		Contact:  NewContactRepository(db),
		Admin:    NewAdminRepository(db),
	}
}

// RepositoryManager provides centralized repository access
type RepositoryManager struct {
	db           *gorm.DB
	repositories *Repositories
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(db *gorm.DB) *RepositoryManager {
	return &RepositoryManager{
		db:           db,
		repositories: NewRepositories(db),
	}
}

// GetRepositories returns all repository instances
func (rm *RepositoryManager) GetRepositories() *Repositories {
	return rm.repositories
}

// GetDB returns the database connection for raw queries if needed
func (rm *RepositoryManager) GetDB() *gorm.DB {
	return rm.db
}

// Property returns the PropertyRepository instance
func (rm *RepositoryManager) Property() PropertyRepository {
	return rm.repositories.Property
}

// Booking returns the BookingRepository instance
func (rm *RepositoryManager) Booking() BookingRepository {
	return rm.repositories.Booking
}

// Contact returns the ContactRepository instance
func (rm *RepositoryManager) Contact() ContactRepository {
	return rm.repositories.Contact
}

// Admin returns the AdminRepository instance
func (rm *RepositoryManager) Admin() AdminRepository {
	return rm.repositories.Admin
}
