package security

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"image/png"
	"net/http"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// TOTPManager handles TOTP multi-factor authentication
type TOTPManager struct {
	db *gorm.DB
}

// NewTOTPManager creates a new TOTP manager
func NewTOTPManager(db *gorm.DB) *TOTPManager {
	return &TOTPManager{
		db: db,
	}
}

// TOTPSecret represents a user's TOTP secret
type TOTPSecret struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"uniqueIndex;not null"`
	Secret      string     `json:"-" gorm:"not null"`  // Never expose in JSON
	BackupCodes []string   `json:"-" gorm:"type:json"` // Never expose in JSON
	IsEnabled   bool       `json:"is_enabled" gorm:"default:false"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
}

// BackupCode represents a single-use backup code
type BackupCode struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id" gorm:"not null"`
	Code      string     `json:"-" gorm:"not null"` // Never expose in JSON
	IsUsed    bool       `json:"is_used" gorm:"default:false"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// DeviceFingerprint represents a trusted device
type DeviceFingerprint struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	Fingerprint string    `json:"fingerprint" gorm:"not null"`
	DeviceName  string    `json:"device_name"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`
	Location    string    `json:"location"`
	IsTrusted   bool      `json:"is_trusted" gorm:"default:false"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MFAAttempt logs MFA authentication attempts
type MFAAttempt struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	AttemptType string    `json:"attempt_type"` // "totp", "backup_code"
	Success     bool      `json:"success"`
	FailReason  string    `json:"fail_reason"`
	CreatedAt   time.Time `json:"created_at"`
}

// GenerateSecret creates a new TOTP secret for a user
func (tm *TOTPManager) GenerateSecret(userID uint, userEmail string) (*otp.Key, []string, error) {
	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "PropertyHub",
		AccountName: userEmail,
		SecretSize:  32,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate backup codes
	backupCodes, err := tm.generateBackupCodes(8)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Save to database
	totpSecret := &TOTPSecret{
		UserID:      userID,
		Secret:      key.Secret(),
		BackupCodes: backupCodes,
		IsEnabled:   false, // User must verify before enabling
	}

	if err := tm.db.Create(totpSecret).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	// Also save individual backup codes for tracking
	for _, code := range backupCodes {
		backupCode := &BackupCode{
			UserID: userID,
			Code:   code,
		}
		tm.db.Create(backupCode)
	}

	return key, backupCodes, nil
}

// VerifyTOTP verifies a TOTP code and enables MFA if first verification
func (tm *TOTPManager) VerifyTOTP(userID uint, code string, ipAddress, userAgent string) (bool, error) {
	var totpSecret TOTPSecret
	if err := tm.db.Where("user_id = ?", userID).First(&totpSecret).Error; err != nil {
		tm.logMFAAttempt(userID, ipAddress, userAgent, "totp", false, "secret_not_found")
		return false, fmt.Errorf("TOTP secret not found")
	}

	// Verify the code
	valid := totp.Validate(code, totpSecret.Secret)

	// Log the attempt
	failReason := ""
	if !valid {
		failReason = "invalid_code"
	}
	tm.logMFAAttempt(userID, ipAddress, userAgent, "totp", valid, failReason)

	if valid {
		// Enable MFA if this is the first successful verification
		if !totpSecret.IsEnabled {
			totpSecret.IsEnabled = true
		}

		// Update last used time
		now := time.Now()
		totpSecret.LastUsedAt = &now
		tm.db.Save(&totpSecret)
	}

	return valid, nil
}

// VerifyBackupCode verifies a backup code (single use)
func (tm *TOTPManager) VerifyBackupCode(userID uint, code string, ipAddress, userAgent string) (bool, error) {
	var backupCode BackupCode
	if err := tm.db.Where("user_id = ? AND code = ? AND is_used = false", userID, code).First(&backupCode).Error; err != nil {
		tm.logMFAAttempt(userID, ipAddress, userAgent, "backup_code", false, "code_not_found_or_used")
		return false, fmt.Errorf("backup code not found or already used")
	}

	// Mark code as used
	now := time.Now()
	backupCode.IsUsed = true
	backupCode.UsedAt = &now
	tm.db.Save(&backupCode)

	// Log successful attempt
	tm.logMFAAttempt(userID, ipAddress, userAgent, "backup_code", true, "")

	return true, nil
}

// IsMFAEnabled checks if MFA is enabled for a user
func (tm *TOTPManager) IsMFAEnabled(userID uint) bool {
	var totpSecret TOTPSecret
	if err := tm.db.Where("user_id = ? AND is_enabled = true", userID).First(&totpSecret).Error; err != nil {
		return false
	}
	return true
}

// DisableMFA disables MFA for a user (admin function)
func (tm *TOTPManager) DisableMFA(userID uint) error {
	// Disable TOTP
	if err := tm.db.Model(&TOTPSecret{}).Where("user_id = ?", userID).Update("is_enabled", false).Error; err != nil {
		return fmt.Errorf("failed to disable TOTP: %w", err)
	}

	// Mark all backup codes as used
	now := time.Now()
	if err := tm.db.Model(&BackupCode{}).Where("user_id = ? AND is_used = false", userID).Updates(map[string]interface{}{
		"is_used": true,
		"used_at": now,
	}).Error; err != nil {
		return fmt.Errorf("failed to invalidate backup codes: %w", err)
	}

	return nil
}

// RegenerateBackupCodes generates new backup codes for a user
func (tm *TOTPManager) RegenerateBackupCodes(userID uint) ([]string, error) {
	// Mark existing backup codes as used
	now := time.Now()
	tm.db.Model(&BackupCode{}).Where("user_id = ? AND is_used = false", userID).Updates(map[string]interface{}{
		"is_used": true,
		"used_at": now,
	})

	// Generate new backup codes
	backupCodes, err := tm.generateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Save new backup codes
	for _, code := range backupCodes {
		backupCode := &BackupCode{
			UserID: userID,
			Code:   code,
		}
		tm.db.Create(backupCode)
	}

	// Update TOTP secret with new backup codes
	tm.db.Model(&TOTPSecret{}).Where("user_id = ?", userID).Update("backup_codes", backupCodes)

	return backupCodes, nil
}

// GetQRCode generates a QR code image for TOTP setup
func (tm *TOTPManager) GetQRCode(key *otp.Key, w http.ResponseWriter) error {
	// Generate QR code
	img, err := key.Image(256, 256)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Set headers
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Write image
	return png.Encode(w, img)
}

// GetMFAStatus returns MFA status for a user
func (tm *TOTPManager) GetMFAStatus(userID uint) map[string]interface{} {
	var totpSecret TOTPSecret
	mfaEnabled := tm.db.Where("user_id = ? AND is_enabled = true", userID).First(&totpSecret).Error == nil

	var unusedBackupCodes int64
	tm.db.Model(&BackupCode{}).Where("user_id = ? AND is_used = false", userID).Count(&unusedBackupCodes)

	var trustedDevices int64
	tm.db.Model(&DeviceFingerprint{}).Where("user_id = ? AND is_trusted = true", userID).Count(&trustedDevices)

	var recentAttempts int64
	tm.db.Model(&MFAAttempt{}).Where("user_id = ? AND created_at > ?", userID, time.Now().Add(-24*time.Hour)).Count(&recentAttempts)

	return map[string]interface{}{
		"mfa_enabled":         mfaEnabled,
		"backup_codes_count":  unusedBackupCodes,
		"trusted_devices":     trustedDevices,
		"recent_attempts_24h": recentAttempts,
		"last_used":           totpSecret.LastUsedAt,
	}
}

// generateBackupCodes creates random backup codes
func (tm *TOTPManager) generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		// Generate 8 random bytes
		bytes := make([]byte, 8)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}

		// Convert to base32 and format
		code := base32.StdEncoding.EncodeToString(bytes)[:10] // Take first 10 chars
		codes[i] = fmt.Sprintf("%s-%s", code[:5], code[5:])   // Format as XXXXX-XXXXX
	}

	return codes, nil
}

// logMFAAttempt logs an MFA authentication attempt
func (tm *TOTPManager) logMFAAttempt(userID uint, ipAddress, userAgent, attemptType string, success bool, failReason string) {
	attempt := &MFAAttempt{
		UserID:      userID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		AttemptType: attemptType,
		Success:     success,
		FailReason:  failReason,
	}
	tm.db.Create(attempt)
}

// CleanupExpiredAttempts removes old MFA attempt logs
func (tm *TOTPManager) CleanupExpiredAttempts() error {
	// Remove attempts older than 90 days
	cutoff := time.Now().Add(-90 * 24 * time.Hour)
	return tm.db.Where("created_at < ?", cutoff).Delete(&MFAAttempt{}).Error
}

// GetMFAStatistics returns MFA usage statistics
func (tm *TOTPManager) GetMFAStatistics() map[string]interface{} {
	var totalUsers int64
	tm.db.Model(&TOTPSecret{}).Count(&totalUsers)

	var enabledUsers int64
	tm.db.Model(&TOTPSecret{}).Where("is_enabled = true").Count(&enabledUsers)

	var recentAttempts int64
	tm.db.Model(&MFAAttempt{}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).Count(&recentAttempts)

	var successfulAttempts int64
	tm.db.Model(&MFAAttempt{}).Where("created_at > ? AND success = true", time.Now().Add(-24*time.Hour)).Count(&successfulAttempts)

	var trustedDevices int64
	tm.db.Model(&DeviceFingerprint{}).Where("is_trusted = true").Count(&trustedDevices)

	successRate := float64(0)
	if recentAttempts > 0 {
		successRate = float64(successfulAttempts) / float64(recentAttempts) * 100
	}

	return map[string]interface{}{
		"total_users_with_mfa":  totalUsers,
		"enabled_users":         enabledUsers,
		"adoption_rate":         float64(enabledUsers) / float64(totalUsers) * 100,
		"attempts_24h":          recentAttempts,
		"success_rate_24h":      successRate,
		"trusted_devices_total": trustedDevices,
	}
}
