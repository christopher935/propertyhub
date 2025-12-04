package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

// EncryptionManager handles field-level encryption
type EncryptionManager struct {
	gcm   cipher.AEAD
	keyID string
	db    *gorm.DB
}

// EncryptionKey represents an encryption key in the database
type EncryptionKey struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	KeyID       string     `json:"key_id" gorm:"uniqueIndex;not null"`
	KeyHash     string     `json:"key_hash" gorm:"not null"` // SHA256 hash for verification
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at"`
	RotatedAt   *time.Time `json:"rotated_at"`
	Description string     `json:"description"`
}

// EncryptedField represents an encrypted database field
type EncryptedField struct {
	KeyID      string `json:"key_id"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager(db *gorm.DB) (*EncryptionManager, error) {
	// Get encryption key from environment or generate one
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		// Generate a new key for development
		key := make([]byte, 32) // 256-bit key
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		encryptionKey = base64.StdEncoding.EncodeToString(key)

		// In production, this should be set as an environment variable
		fmt.Printf("Generated encryption key (set ENCRYPTION_KEY env var): %s\n", encryptionKey)
	}

	// Decode the key
	key, err := base64.StdEncoding.DecodeString(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key format: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes (256 bits)")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate key ID
	keyHash := sha256.Sum256(key)
	keyID := base64.StdEncoding.EncodeToString(keyHash[:8]) // First 8 bytes as ID

	em := &EncryptionManager{
		gcm:   gcm,
		keyID: keyID,
		db:    db,
	}

	// Store key info in database
	if err := em.storeKeyInfo(key); err != nil {
		return nil, fmt.Errorf("failed to store key info: %w", err)
	}

	return em, nil
}

// EncryptedString is a custom type for encrypted string fields
type EncryptedString string

// Encrypt encrypts a plaintext string
func (em *EncryptionManager) Encrypt(plaintext string) (EncryptedString, error) {
	if plaintext == "" {
		return "", nil
	}

	// Generate a random nonce
	nonce := make([]byte, em.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := em.gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Create encrypted field structure
	encField := EncryptedField{
		KeyID:      em.keyID,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(encField)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted field: %w", err)
	}

	return EncryptedString(jsonData), nil
}

// Decrypt decrypts an encrypted string
func (em *EncryptionManager) Decrypt(encrypted EncryptedString) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	// Check if it's already plaintext (for backward compatibility)
	encryptedStr := string(encrypted)
	if !strings.HasPrefix(encryptedStr, "{") {
		// Assume it's plaintext for backward compatibility
		return encryptedStr, nil
	}

	// Parse encrypted field
	var encField EncryptedField
	if err := json.Unmarshal([]byte(encrypted), &encField); err != nil {
		// If JSON parsing fails, assume it's plaintext
		return encryptedStr, nil
	}

	// Verify key ID
	if encField.KeyID != em.keyID {
		return "", fmt.Errorf("key ID mismatch: expected %s, got %s", em.keyID, encField.KeyID)
	}

	// Decode nonce and ciphertext
	nonce, err := base64.StdEncoding.DecodeString(encField.Nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encField.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decrypt
	plaintext, err := em.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Value implements the driver.Valuer interface for database storage
func (es EncryptedString) Value() (driver.Value, error) {
	return string(es), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (es *EncryptedString) Scan(value interface{}) error {
	if value == nil {
		*es = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*es = EncryptedString(v)
	case []byte:
		*es = EncryptedString(v)
	default:
		return fmt.Errorf("cannot scan %T into EncryptedString", value)
	}

	return nil
}

// String returns the encrypted string (for debugging - never shows plaintext)
func (es EncryptedString) String() string {
	if es == "" {
		return ""
	}
	return "[ENCRYPTED]"
}

// MarshalJSON implements json.Marshaler (never exposes plaintext)
func (es EncryptedString) MarshalJSON() ([]byte, error) {
	if es == "" {
		return json.Marshal("")
	}
	return json.Marshal("[ENCRYPTED]")
}

// EncryptedJSON is a custom type for encrypted JSON fields
type EncryptedJSON map[string]interface{}

// Value implements the driver.Valuer interface
func (ej EncryptedJSON) Value() (driver.Value, error) {
	if len(ej) == 0 {
		return nil, nil
	}

	// This would need the encryption manager instance
	// For now, just marshal as regular JSON
	return json.Marshal(ej)
}

// Scan implements the sql.Scanner interface
func (ej *EncryptedJSON) Scan(value interface{}) error {
	if value == nil {
		*ej = make(EncryptedJSON)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, ej)
}

// EncryptionAuditLog tracks encryption operations
type EncryptionAuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Operation string    `json:"operation"` // "encrypt", "decrypt", "key_rotation"
	KeyID     string    `json:"key_id"`
	TableName string    `json:"table_name"`
	FieldName string    `json:"field_name"`
	RecordID  uint      `json:"record_id"`
	UserID    uint      `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	Success   bool      `json:"success"`
	ErrorMsg  string    `json:"error_msg"`
	CreatedAt time.Time `json:"created_at"`
}

// LogEncryptionOperation logs an encryption/decryption operation
func (em *EncryptionManager) LogEncryptionOperation(operation, tableName, fieldName string, recordID, userID uint, ipAddress string, success bool, errorMsg string) {
	auditLog := &EncryptionAuditLog{
		Operation: operation,
		KeyID:     em.keyID,
		TableName: tableName,
		FieldName: fieldName,
		RecordID:  recordID,
		UserID:    userID,
		IPAddress: ipAddress,
		Success:   success,
		ErrorMsg:  errorMsg,
	}

	em.db.Create(auditLog)
}

// storeKeyInfo stores encryption key information in the database
func (em *EncryptionManager) storeKeyInfo(key []byte) error {
	keyHash := sha256.Sum256(key)
	keyHashStr := base64.StdEncoding.EncodeToString(keyHash[:])

	// Check if key already exists
	var existingKey EncryptionKey
	if err := em.db.Where("key_id = ?", em.keyID).First(&existingKey).Error; err == nil {
		// Key exists, verify hash
		if existingKey.KeyHash != keyHashStr {
			return fmt.Errorf("key ID collision: different key with same ID exists")
		}
		return nil
	}

	// Create new key record
	encKey := &EncryptionKey{
		KeyID:       em.keyID,
		KeyHash:     keyHashStr,
		IsActive:    true,
		Description: "AES-256 encryption key for sensitive data",
	}

	return em.db.Create(encKey).Error
}

// RotateKey rotates the encryption key (advanced feature)
func (em *EncryptionManager) RotateKey() error {
	// Mark current key as inactive
	now := time.Now()
	if err := em.db.Model(&EncryptionKey{}).Where("key_id = ?", em.keyID).Updates(map[string]interface{}{
		"is_active":  false,
		"rotated_at": now,
	}).Error; err != nil {
		return fmt.Errorf("failed to deactivate old key: %w", err)
	}

	// Generate new key
	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	// Update encryption manager with new key
	block, err := aes.NewCipher(newKey)
	if err != nil {
		return fmt.Errorf("failed to create new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create new GCM: %w", err)
	}

	// Generate new key ID
	keyHash := sha256.Sum256(newKey)
	newKeyID := base64.StdEncoding.EncodeToString(keyHash[:8])

	// Update manager
	em.gcm = gcm
	em.keyID = newKeyID

	// Store new key info
	return em.storeKeyInfo(newKey)
}

// GetEncryptionStatistics returns encryption usage statistics
func (em *EncryptionManager) GetEncryptionStatistics() map[string]interface{} {
	var totalOperations int64
	em.db.Model(&EncryptionAuditLog{}).Count(&totalOperations)

	var encryptOperations int64
	em.db.Model(&EncryptionAuditLog{}).Where("operation = 'encrypt'").Count(&encryptOperations)

	var decryptOperations int64
	em.db.Model(&EncryptionAuditLog{}).Where("operation = 'decrypt'").Count(&decryptOperations)

	var failedOperations int64
	em.db.Model(&EncryptionAuditLog{}).Where("success = false").Count(&failedOperations)

	var recentOperations int64
	em.db.Model(&EncryptionAuditLog{}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).Count(&recentOperations)

	var activeKeys int64
	em.db.Model(&EncryptionKey{}).Where("is_active = true").Count(&activeKeys)

	successRate := float64(100)
	if totalOperations > 0 {
		successRate = float64(totalOperations-failedOperations) / float64(totalOperations) * 100
	}

	return map[string]interface{}{
		"total_operations":      totalOperations,
		"encrypt_operations":    encryptOperations,
		"decrypt_operations":    decryptOperations,
		"failed_operations":     failedOperations,
		"success_rate":          successRate,
		"recent_operations_24h": recentOperations,
		"active_keys":           activeKeys,
		"current_key_id":        em.keyID,
	}
}

// CleanupAuditLogs removes old encryption audit logs
func (em *EncryptionManager) CleanupAuditLogs() error {
	// Remove logs older than 1 year
	cutoff := time.Now().Add(-365 * 24 * time.Hour)
	return em.db.Where("created_at < ?", cutoff).Delete(&EncryptionAuditLog{}).Error
}

// ValidateEncryption validates that encryption is working correctly
func (em *EncryptionManager) ValidateEncryption() error {
	testData := "test-encryption-validation-" + time.Now().Format("20060102150405")

	// Encrypt
	encrypted, err := em.Encrypt(testData)
	if err != nil {
		return fmt.Errorf("encryption validation failed: %w", err)
	}

	// Decrypt
	decrypted, err := em.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("decryption validation failed: %w", err)
	}

	// Verify
	if decrypted != testData {
		return fmt.Errorf("encryption validation failed: data mismatch")
	}

	return nil
}

// Helper functions for common encrypted fields

// EncryptEmail encrypts an email address
func (em *EncryptionManager) EncryptEmail(email string) (EncryptedString, error) {
	return em.Encrypt(email)
}

// DecryptEmail decrypts an email address
func (em *EncryptionManager) DecryptEmail(encrypted EncryptedString) (string, error) {
	return em.Decrypt(encrypted)
}

// EncryptPhone encrypts a phone number
func (em *EncryptionManager) EncryptPhone(phone string) (EncryptedString, error) {
	return em.Encrypt(phone)
}

// DecryptPhone decrypts a phone number
func (em *EncryptionManager) DecryptPhone(encrypted EncryptedString) (string, error) {
	return em.Decrypt(encrypted)
}

// EncryptSSN encrypts a social security number
func (em *EncryptionManager) EncryptSSN(ssn string) (EncryptedString, error) {
	return em.Encrypt(ssn)
}

// DecryptSSN decrypts a social security number
func (em *EncryptionManager) DecryptSSN(encrypted EncryptedString) (string, error) {
	return em.Decrypt(encrypted)
}
