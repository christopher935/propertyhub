package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocumentEncryption handles encryption of real estate documents and PII data
type DocumentEncryption struct {
	db           *gorm.DB
	logger       *log.Logger
	masterKey    []byte
	documentPath string
	auditLogger  *AuditLogger
}

// EncryptedDocument represents an encrypted document in the database
type EncryptedDocument struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	DocumentName      string     `json:"document_name" gorm:"not null"`
	OriginalFilename  string     `json:"original_filename" gorm:"not null"`
	DocumentType      string     `json:"document_type" gorm:"not null"` // lease, contract, id_verification, etc.
	EncryptedPath     string     `json:"encrypted_path" gorm:"not null"`
	EncryptionKeyHash string     `json:"encryption_key_hash" gorm:"not null"` // Hash of the key used
	DocumentHash      string     `json:"document_hash" gorm:"not null"`       // Hash of original document
	PropertyAddress   string     `json:"property_address"`
	ClientName        string     `json:"client_name"`
	AccessLevel       string     `json:"access_level" gorm:"default:'private'"`   // public, restricted, private, confidential
	PIIClassification string     `json:"pii_classification" gorm:"default:'low'"` // low, medium, high, critical
	RetentionUntil    time.Time  `json:"retention_until"`
	AccessCount       int        `json:"access_count" gorm:"default:0"`
	LastAccessedAt    *time.Time `json:"last_accessed_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// DocumentAccessLog tracks access to encrypted documents for compliance
type DocumentAccessLog struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	DocumentID    uint      `json:"document_id" gorm:"not null"`
	AccessedBy    string    `json:"accessed_by" gorm:"not null"`
	AccessType    string    `json:"access_type" gorm:"not null"` // view, download, decrypt
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	AccessGranted bool      `json:"access_granted" gorm:"default:false"`
	DenialReason  string    `json:"denial_reason"`
	AccessedAt    time.Time `json:"accessed_at"`

	// Relationships
	Document EncryptedDocument `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

// PIIData represents personally identifiable information that needs encryption
type PIIData struct {
	SSN              string `json:"ssn,omitempty"`
	DriverLicenseNum string `json:"driver_license_number,omitempty"`
	BankAccountNum   string `json:"bank_account_number,omitempty"`
	CreditCardNum    string `json:"credit_card_number,omitempty"`
	TaxID            string `json:"tax_id,omitempty"`
	PassportNum      string `json:"passport_number,omitempty"`
	DateOfBirth      string `json:"date_of_birth,omitempty"`
	MedicalInfo      string `json:"medical_info,omitempty"`
	FinancialInfo    string `json:"financial_info,omitempty"`
	BackgroundCheck  string `json:"background_check,omitempty"`
}

// Document types for real estate
const (
	DocTypeLeaseAgreement     = "lease_agreement"
	DocTypePurchaseContract   = "purchase_contract"
	DocTypeIDVerification     = "id_verification"
	DocTypeIncomeVerification = "income_verification"
	DocTypeCreditReport       = "credit_report"
	DocTypeBackgroundCheck    = "background_check"
	DocTypeInsurance          = "insurance"
	DocTypeBankStatements     = "bank_statements"
	DocTypePayStubs           = "pay_stubs"
	DocTypeTaxReturns         = "tax_returns"
	DocTypeUtilityBills       = "utility_bills"
	DocTypePropertyDocs       = "property_documents"
	DocTypeDisclosures        = "disclosures"
	DocTypeApplications       = "applications"
)

// Access levels for document security
const (
	AccessLevelPublic       = "public"       // Marketing materials, property photos
	AccessLevelRestricted   = "restricted"   // Property details, basic applications
	AccessLevelPrivate      = "private"      // Complete applications, references
	AccessLevelConfidential = "confidential" // SSN, credit reports, background checks
)

// PII classification levels
const (
	PIILow      = "low"      // Name, email, phone
	PIIMedium   = "medium"   // Address, employment info
	PIIHigh     = "high"     // Financial info, references
	PIICritical = "critical" // SSN, credit reports, background checks
)

// NewDocumentEncryption creates a new document encryption service
func NewDocumentEncryption(db *gorm.DB, logger *log.Logger, auditLogger *AuditLogger, documentPath string) (*DocumentEncryption, error) {
	// Auto-migrate tables
	err := db.AutoMigrate(&EncryptedDocument{}, &DocumentAccessLog{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate encryption tables: %v", err)
	}

	// Get or generate master key
	masterKey, err := getMasterKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get master key: %v", err)
	}

	// Ensure document directory exists
	if err := os.MkdirAll(documentPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create document directory: %v", err)
	}

	return &DocumentEncryption{
		db:           db,
		logger:       logger,
		masterKey:    masterKey,
		documentPath: documentPath,
		auditLogger:  auditLogger,
	}, nil
}

// EncryptDocument encrypts and stores a document
func (de *DocumentEncryption) EncryptDocument(filename, documentType, propertyAddress, clientName string, data []byte, accessLevel, piiClassification string) (*EncryptedDocument, error) {
	// Generate document-specific encryption key
	docKey := make([]byte, 32)
	if _, err := rand.Read(docKey); err != nil {
		return nil, fmt.Errorf("failed to generate document key: %v", err)
	}

	// Encrypt the data
	encryptedData, err := de.encryptData(data, docKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt document: %v", err)
	}

	// Generate secure filename
	secureFilename := de.generateSecureFilename(filename)
	encryptedPath := filepath.Join(de.documentPath, secureFilename)

	// Write encrypted file
	if err := os.WriteFile(encryptedPath, encryptedData, 0600); err != nil {
		return nil, fmt.Errorf("failed to write encrypted file: %v", err)
	}

	// Calculate document hash
	docHash := fmt.Sprintf("%x", sha256.Sum256(data))
	keyHash := fmt.Sprintf("%x", sha256.Sum256(docKey))

	// Determine retention period based on document type
	retentionPeriod := de.getRetentionPeriod(documentType)

	// Create database record
	doc := &EncryptedDocument{
		DocumentName:      de.sanitizeFilename(filename),
		OriginalFilename:  filename,
		DocumentType:      documentType,
		EncryptedPath:     encryptedPath,
		EncryptionKeyHash: keyHash,
		DocumentHash:      docHash,
		PropertyAddress:   propertyAddress,
		ClientName:        clientName,
		AccessLevel:       accessLevel,
		PIIClassification: piiClassification,
		RetentionUntil:    time.Now().Add(retentionPeriod),
		CreatedAt:         time.Now(),
	}

	if err := de.db.Create(doc).Error; err != nil {
		// Clean up file if database insert fails
		os.Remove(encryptedPath)
		return nil, fmt.Errorf("failed to store document record: %v", err)
	}

	// Store the document key securely (encrypted with master key)
	if err := de.storeDocumentKey(doc.ID, docKey); err != nil {
		de.logger.Printf("⚠️ Warning: Failed to store document key for ID %d: %v", doc.ID, err)
	}

	// Log encryption event
	de.auditLogger.LogSecurityEvent("document_encrypted", nil, "", "", fmt.Sprintf("Document encrypted: %s", filename), map[string]interface{}{
		"document_id":        doc.ID,
		"document_type":      documentType,
		"access_level":       accessLevel,
		"pii_classification": piiClassification,
		"client_name":        clientName,
		"property_address":   propertyAddress,
	}, 40)

	de.logger.Printf("✅ Document encrypted and stored: %s (ID: %d)", filename, doc.ID)
	return doc, nil
}

// DecryptDocument decrypts and returns document data
func (de *DocumentEncryption) DecryptDocument(documentID uint, accessedBy, ipAddress, userAgent string) ([]byte, *EncryptedDocument, error) {
	// Get document record
	var doc EncryptedDocument
	if err := de.db.First(&doc, documentID).Error; err != nil {
		return nil, nil, fmt.Errorf("document not found: %v", err)
	}

	// Log access attempt
	de.logDocumentAccess(documentID, accessedBy, "decrypt", ipAddress, userAgent, true, "")

	// Get document key
	docKey, err := de.getDocumentKey(documentID)
	if err != nil {
		de.logDocumentAccess(documentID, accessedBy, "decrypt", ipAddress, userAgent, false, "key_retrieval_failed")
		return nil, &doc, fmt.Errorf("failed to get document key: %v", err)
	}

	// Read encrypted file
	encryptedData, err := os.ReadFile(doc.EncryptedPath)
	if err != nil {
		de.logDocumentAccess(documentID, accessedBy, "decrypt", ipAddress, userAgent, false, "file_read_failed")
		return nil, &doc, fmt.Errorf("failed to read encrypted file: %v", err)
	}

	// Decrypt the data
	decryptedData, err := de.decryptData(encryptedData, docKey)
	if err != nil {
		de.logDocumentAccess(documentID, accessedBy, "decrypt", ipAddress, userAgent, false, "decryption_failed")
		return nil, &doc, fmt.Errorf("failed to decrypt document: %v", err)
	}

	// Verify document integrity
	calculatedHash := fmt.Sprintf("%x", sha256.Sum256(decryptedData))
	if calculatedHash != doc.DocumentHash {
		de.logDocumentAccess(documentID, accessedBy, "decrypt", ipAddress, userAgent, false, "integrity_check_failed")
		return nil, &doc, fmt.Errorf("document integrity check failed")
	}

	// Update access statistics
	de.db.Model(&doc).Updates(map[string]interface{}{
		"access_count":     doc.AccessCount + 1,
		"last_accessed_at": time.Now(),
	})

	de.logger.Printf("✅ Document decrypted successfully: %s (ID: %d) by %s", doc.DocumentName, documentID, accessedBy)
	return decryptedData, &doc, nil
}

// EncryptPIIData encrypts personally identifiable information
func (de *DocumentEncryption) EncryptPIIData(piiData PIIData) (string, error) {
	// Convert PII data to JSON
	jsonData, err := json.Marshal(piiData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal PII data: %v", err)
	}

	// Generate PII-specific key
	piiKey := make([]byte, 32)
	if _, err := rand.Read(piiKey); err != nil {
		return "", fmt.Errorf("failed to generate PII key: %v", err)
	}

	// Encrypt PII data
	encryptedData, err := de.encryptData(jsonData, piiKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt PII data: %v", err)
	}

	// Encrypt the PII key with master key
	encryptedKey, err := de.encryptData(piiKey, de.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt PII key: %v", err)
	}

	// Combine encrypted key and data
	combined := append(encryptedKey, encryptedData...)

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptPIIData decrypts personally identifiable information
func (de *DocumentEncryption) DecryptPIIData(encryptedPII string) (PIIData, error) {
	var piiData PIIData

	// Decode base64
	combined, err := base64.StdEncoding.DecodeString(encryptedPII)
	if err != nil {
		return piiData, fmt.Errorf("failed to decode encrypted PII: %v", err)
	}

	// Split encrypted key and data (first 48 bytes are encrypted key)
	if len(combined) < 48 {
		return piiData, fmt.Errorf("invalid encrypted PII data length")
	}

	encryptedKey := combined[:48]
	encryptedData := combined[48:]

	// Decrypt the PII key
	piiKey, err := de.decryptData(encryptedKey, de.masterKey)
	if err != nil {
		return piiData, fmt.Errorf("failed to decrypt PII key: %v", err)
	}

	// Decrypt the PII data
	jsonData, err := de.decryptData(encryptedData, piiKey)
	if err != nil {
		return piiData, fmt.Errorf("failed to decrypt PII data: %v", err)
	}

	// Unmarshal JSON data
	if err := json.Unmarshal(jsonData, &piiData); err != nil {
		return piiData, fmt.Errorf("failed to unmarshal PII data: %v", err)
	}

	return piiData, nil
}

// GetDocumentsByClient returns all documents for a specific client
func (de *DocumentEncryption) GetDocumentsByClient(clientName string) ([]EncryptedDocument, error) {
	var documents []EncryptedDocument

	err := de.db.Where("client_name = ?", clientName).
		Order("created_at DESC").
		Find(&documents).Error

	return documents, err
}

// GetDocumentsByProperty returns all documents for a specific property
func (de *DocumentEncryption) GetDocumentsByProperty(propertyAddress string) ([]EncryptedDocument, error) {
	var documents []EncryptedDocument

	err := de.db.Where("property_address = ?", propertyAddress).
		Order("created_at DESC").
		Find(&documents).Error

	return documents, err
}

// DeleteDocument securely deletes a document
func (de *DocumentEncryption) DeleteDocument(documentID uint, deletedBy, reason, ipAddress, userAgent string) error {
	var doc EncryptedDocument
	if err := de.db.First(&doc, documentID).Error; err != nil {
		return fmt.Errorf("document not found: %v", err)
	}

	// Remove encrypted file
	if err := os.Remove(doc.EncryptedPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete encrypted file: %v", err)
	}

	// Remove document key
	de.deleteDocumentKey(documentID)

	// Remove database record
	if err := de.db.Delete(&doc).Error; err != nil {
		return fmt.Errorf("failed to delete document record: %v", err)
	}

	// Log deletion
	de.auditLogger.LogSecurityEvent("document_deleted", nil, ipAddress, userAgent, fmt.Sprintf("Document deleted: %s", doc.DocumentName), map[string]interface{}{
		"document_id":      documentID,
		"document_name":    doc.DocumentName,
		"deleted_by":       deletedBy,
		"deletion_reason":  reason,
		"client_name":      doc.ClientName,
		"property_address": doc.PropertyAddress,
	}, 60)

	de.logger.Printf("✅ Document securely deleted: %s (ID: %d) by %s", doc.DocumentName, documentID, deletedBy)
	return nil
}

// Helper methods

func (de *DocumentEncryption) encryptData(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)

	// Prepend nonce to ciphertext
	return append(nonce, ciphertext...), nil
}

func (de *DocumentEncryption) decryptData(encryptedData, key []byte) ([]byte, error) {
	if len(encryptedData) < 12 {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:12]
	ciphertext := encryptedData[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, ciphertext, nil)
}

func (de *DocumentEncryption) generateSecureFilename(originalFilename string) string {
	// Generate random filename to prevent path traversal
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)

	// Get file extension
	ext := filepath.Ext(originalFilename)

	return fmt.Sprintf("%x%s", randomBytes, ext)
}

func (de *DocumentEncryption) sanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	cleaned := strings.ReplaceAll(filename, "/", "_")
	cleaned = strings.ReplaceAll(cleaned, "\\", "_")
	cleaned = strings.ReplaceAll(cleaned, "..", "_")
	return cleaned
}

func (de *DocumentEncryption) getRetentionPeriod(documentType string) time.Duration {
	// Set retention periods based on real estate law requirements
	switch documentType {
	case DocTypeLeaseAgreement:
		return time.Hour * 24 * 365 * 7 // 7 years
	case DocTypePurchaseContract:
		return time.Hour * 24 * 365 * 7 // 7 years
	case DocTypeCreditReport, DocTypeBackgroundCheck:
		return time.Hour * 24 * 365 * 5 // 5 years
	case DocTypeIncomeVerification, DocTypeBankStatements, DocTypePayStubs:
		return time.Hour * 24 * 365 * 5 // 5 years
	case DocTypeTaxReturns:
		return time.Hour * 24 * 365 * 7 // 7 years
	case DocTypeIDVerification:
		return time.Hour * 24 * 365 * 4 // 4 years (TREC requirement)
	default:
		return time.Hour * 24 * 365 * 4 // 4 years default
	}
}

func (de *DocumentEncryption) storeDocumentKey(documentID uint, key []byte) error {
	// Encrypt key with master key and store in secure location
	encryptedKey, err := de.encryptData(key, de.masterKey)
	if err != nil {
		return err
	}

	keyPath := filepath.Join(de.documentPath, "keys", fmt.Sprintf("%d.key", documentID))
	os.MkdirAll(filepath.Dir(keyPath), 0700)

	return os.WriteFile(keyPath, encryptedKey, 0600)
}

func (de *DocumentEncryption) getDocumentKey(documentID uint) ([]byte, error) {
	keyPath := filepath.Join(de.documentPath, "keys", fmt.Sprintf("%d.key", documentID))

	encryptedKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	return de.decryptData(encryptedKey, de.masterKey)
}

func (de *DocumentEncryption) deleteDocumentKey(documentID uint) {
	keyPath := filepath.Join(de.documentPath, "keys", fmt.Sprintf("%d.key", documentID))
	os.Remove(keyPath)
}

func (de *DocumentEncryption) logDocumentAccess(documentID uint, accessedBy, accessType, ipAddress, userAgent string, granted bool, denialReason string) {
	accessLog := &DocumentAccessLog{
		DocumentID:    documentID,
		AccessedBy:    accessedBy,
		AccessType:    accessType,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		AccessGranted: granted,
		DenialReason:  denialReason,
		AccessedAt:    time.Now(),
	}

	de.db.Create(accessLog)
}

func getMasterKey() ([]byte, error) {
	keyPath := os.Getenv("DOCUMENT_MASTER_KEY_PATH")
	if keyPath == "" {
		keyPath = "/etc/chrisgross-ctrl-project/master.key"
	}

	// Try to read existing key
	if key, err := os.ReadFile(keyPath); err == nil {
		if len(key) == 32 {
			return key, nil
		}
	}

	// Generate new master key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(keyPath), 0700)

	// Save master key
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save master key: %v", err)
	}

	return key, nil
}
