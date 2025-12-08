package repositories

import (
	"chrisgross-ctrl-project/internal/models"
	"context"
	"fmt"
	"gorm.io/gorm"
)

// ContactRepositoryImpl implements ContactRepository interface
type ContactRepositoryImpl struct {
	db *gorm.DB
}

// NewContactRepository creates a new ContactRepository instance
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &ContactRepositoryImpl{db: db}
}

// Create creates a new contact
func (r *ContactRepositoryImpl) Create(ctx context.Context, entity interface{}) error {
	contact, ok := entity.(*models.Contact)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Contact")
	}
	return r.db.WithContext(ctx).Create(contact).Error
}

// Update updates an existing contact
func (r *ContactRepositoryImpl) Update(ctx context.Context, entity interface{}) error {
	contact, ok := entity.(*models.Contact)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Contact")
	}
	return r.db.WithContext(ctx).Save(contact).Error
}

// Delete soft deletes a contact by ID
func (r *ContactRepositoryImpl) Delete(ctx context.Context, id interface{}) error {
	return r.db.WithContext(ctx).Delete(&models.Contact{}, id).Error
}

// FindByID finds a contact by ID
func (r *ContactRepositoryImpl) FindByID(ctx context.Context, id interface{}, entity interface{}) error {
	contact, ok := entity.(*models.Contact)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *models.Contact")
	}
	return r.db.WithContext(ctx).First(contact, id).Error
}

// FindAll finds all contacts
func (r *ContactRepositoryImpl) FindAll(ctx context.Context, entities interface{}) error {
	contacts, ok := entities.(*[]*models.Contact)
	if !ok {
		return fmt.Errorf("invalid entity type, expected *[]*models.Contact")
	}
	return r.db.WithContext(ctx).Find(contacts).Error
}

// FindByEmail finds a contact by email address with optimized query
func (r *ContactRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.Contact, error) {
	var contact models.Contact

	// Use index idx_contacts_email for optimized performance
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&contact).Error
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// CreateOrUpdate creates a new contact or updates existing one based on email
func (r *ContactRepositoryImpl) CreateOrUpdate(ctx context.Context, contact *models.Contact) error {
	if contact == nil {
		return fmt.Errorf("contact cannot be nil")
	}

	// Use transaction for consistency
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingContact models.Contact

		// Check if contact exists by email
		err := tx.Where("email = ?", contact.Email).First(&existingContact).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new contact
				return tx.Create(contact).Error
			}
			return fmt.Errorf("failed to check existing contact: %v", err)
		}

		// Update existing contact
		contact.ID = existingContact.ID
		contact.CreatedAt = existingContact.CreatedAt // Preserve original creation time
		return tx.Save(contact).Error
	})
}

// GetRecentLeads returns recent leads ordered by creation date
func (r *ContactRepositoryImpl) GetRecentLeads(ctx context.Context, limit int) ([]*models.Contact, error) {
	var contacts []*models.Contact

	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	// Use idx_contacts_created_at for optimal performance
	err := r.db.WithContext(ctx).
		Where("status IN ?", []string{"new", "contacted"}).
		Order("created_at DESC").
		Limit(limit).
		Find(&contacts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent leads: %v", err)
	}

	return contacts, nil
}

// GetLeadsBySource returns contacts filtered by source
func (r *ContactRepositoryImpl) GetLeadsBySource(ctx context.Context, source string) ([]*models.Contact, error) {
	var contacts []*models.Contact

	// Use idx_contacts_source for optimization if available
	err := r.db.WithContext(ctx).
		Where("source = ?", source).
		Order("created_at DESC").
		Find(&contacts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get leads by source: %v", err)
	}

	return contacts, nil
}

// UpdateLeadStatus updates the status of a specific contact/lead
func (r *ContactRepositoryImpl) UpdateLeadStatus(ctx context.Context, contactID uint, status string) error {
	// Use optimized update with where clause
	result := r.db.WithContext(ctx).Model(&models.Contact{}).
		Where("id = ?", contactID).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update lead status: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("contact with ID %d not found", contactID)
	}

	return nil
}
