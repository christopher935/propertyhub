package models

import (
	"time"

	"chrisgross-ctrl-project/internal/security"
)

// ContactDataResponse represents a contact with decrypted sensitive fields
type ContactDataResponse struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`  // Decrypted
	Phone      string    `json:"phone"` // Decrypted
	Email      string    `json:"email"` // Decrypted
	Message    string    `json:"message"`
	PropertyID string    `json:"property_id"`
	Urgent     bool      `json:"urgent"`
	Source     string    `json:"source"`
	Status     string    `json:"status"`
	FUBLeadID  string    `json:"fub_lead_id"`
	FUBSynced  bool      `json:"fub_synced"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ToContactDataResponse converts a Contact to ContactDataResponse with decrypted fields
func ToContactDataResponse(contact Contact, encryptionManager *security.EncryptionManager) ContactDataResponse {
	// Default to encrypted values
	name := string(contact.Name)
	phone := string(contact.Phone)
	email := string(contact.Email)

	// Decrypt if encryption manager is available
	if encryptionManager != nil {
		if decrypted, err := encryptionManager.Decrypt(contact.Name); err == nil {
			name = decrypted
		}
		if decrypted, err := encryptionManager.Decrypt(contact.Phone); err == nil {
			phone = decrypted
		}
		if decrypted, err := encryptionManager.Decrypt(contact.Email); err == nil {
			email = decrypted
		}
	}

	return ContactDataResponse{
		ID:         contact.ID,
		Name:       name,
		Phone:      phone,
		Email:      email,
		Message:    contact.Message,
		PropertyID: contact.PropertyID,
		Urgent:     contact.Urgent,
		Source:     contact.Source,
		Status:     contact.Status,
		FUBLeadID:  contact.FUBLeadID,
		FUBSynced:  contact.FUBSynced,
		CreatedAt:  contact.CreatedAt,
		UpdatedAt:  contact.UpdatedAt,
	}
}

// ToContactDataResponseList converts a slice of Contacts to ContactDataResponses
func ToContactDataResponseList(contacts []Contact, encryptionManager *security.EncryptionManager) []ContactDataResponse {
	responses := make([]ContactDataResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = ToContactDataResponse(contact, encryptionManager)
	}
	return responses
}
