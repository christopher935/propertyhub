package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	
	db.AutoMigrate(&SecurityEvent{}, &AuditLog{}, &DataAccessLog{}, &AdminAction{})
	
	return db
}

func TestSQLInjectionDetection(t *testing.T) {
	maliciousInputs := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"1; DELETE FROM properties",
		"admin'--",
		"' OR 1=1--",
		"1' UNION SELECT NULL, username, password FROM users--",
		"'; EXEC sp_MSForEachTable 'DROP TABLE ?'; --",
		"1' AND 1=(SELECT COUNT(*) FROM tabname); --",
	}

	detector := &InputValidator{}
	
	for _, input := range maliciousInputs {
		t.Run(input, func(t *testing.T) {
			isSafe := detector.ValidateNoSQLInjection(input)
			assert.False(t, isSafe, "Should detect SQL injection in: %s", input)
		})
	}
}

func TestXSSDetection(t *testing.T) {
	maliciousInputs := []string{
		"<script>alert('xss')</script>",
		"<img onerror='alert(1)' src='x'>",
		"javascript:alert(1)",
		"<iframe src='javascript:alert(1)'></iframe>",
		"<body onload='alert(1)'>",
		"<svg/onload=alert(1)>",
		"<object data='javascript:alert(1)'>",
		"\"><script>alert(1)</script>",
	}

	detector := &InputValidator{}
	
	for _, input := range maliciousInputs {
		t.Run(input, func(t *testing.T) {
			isSafe := detector.ValidateNoXSS(input)
			assert.False(t, isSafe, "Should detect XSS in: %s", input)
		})
	}
}

func TestSafeInputs(t *testing.T) {
	safeInputs := []string{
		"john.doe@example.com",
		"123 Main Street",
		"Houston, TX 77001",
		"Property with 3 bedrooms",
		"$350,000",
		"Call me at (713) 555-0123",
	}

	detector := &InputValidator{}
	
	for _, input := range safeInputs {
		t.Run(input, func(t *testing.T) {
			assert.True(t, detector.ValidateNoSQLInjection(input), "Should allow safe input: %s", input)
			assert.True(t, detector.ValidateNoXSS(input), "Should allow safe input: %s", input)
		})
	}
}

func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		password string
		expected bool
		reason   string
	}{
		{"short", false, "too short"},
		{"alllowercase123", false, "no uppercase"},
		{"ALLUPPERCASE123", false, "no lowercase"},
		{"NoNumbers!", false, "no numbers"},
		{"NoSpecial123", false, "no special characters"},
		{"ValidPass123!", true, "valid password"},
		{"AnotherGood1@", true, "valid password"},
		{"MySecureP@ss123", true, "valid password"},
	}

	validator := &PasswordValidator{}
	
	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := validator.IsStrong(tt.password)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}

func TestSensitiveDataRedaction(t *testing.T) {
	db := setupTestDB()
	logger := NewAuditLogger(db)

	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Password redaction",
			input: map[string]interface{}{
				"username": "john",
				"password": "secret123",
				"email":    "john@example.com",
			},
			expected: map[string]interface{}{
				"username": "john",
				"password": "[REDACTED]",
				"email":    "john@example.com",
			},
		},
		{
			name: "Multiple sensitive fields",
			input: map[string]interface{}{
				"api_key":      "sk-abc123xyz",
				"access_token": "token-xyz-123",
				"user_data":    "public info",
			},
			expected: map[string]interface{}{
				"api_key":      "[REDACTED]",
				"access_token": "[REDACTED]",
				"user_data":    "public info",
			},
		},
		{
			name: "Nested sensitive data",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "John",
					"password": "secret",
					"ssn":      "123-45-6789",
				},
			},
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "John",
					"password": "[REDACTED]",
					"ssn":      "[REDACTED]",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := logger.SanitizeForLog(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBruteForceProtection(t *testing.T) {
	db := setupTestDB()
	bfp := NewBruteForceProtection(db, nil)

	identifier := "test@example.com"
	ipAddress := "192.168.1.1"

	for i := 1; i <= MaxLoginAttempts; i++ {
		allowed, remaining, _, err := bfp.CheckLoginAttempt(identifier, ipAddress)
		assert.NoError(t, err)
		assert.True(t, allowed, "Attempt %d should be allowed", i)
		assert.Equal(t, MaxLoginAttempts-i, remaining)

		err = bfp.RecordFailedAttempt(identifier, ipAddress, "test-agent")
		assert.NoError(t, err)
	}

	allowed, remaining, retryAfter, err := bfp.CheckLoginAttempt(identifier, ipAddress)
	assert.NoError(t, err)
	assert.False(t, allowed, "Should be locked out after max attempts")
	assert.Equal(t, 0, remaining)
	assert.Greater(t, retryAfter, time.Duration(0))
}

func TestAuditLogStatistics(t *testing.T) {
	db := setupTestDB()
	logger := NewAuditLogger(db)

	userID := uint(1)
	logger.LogAction(&userID, "session-1", "192.168.1.1", "test-agent", "login", "user", &userID, true, "", map[string]interface{}{})
	logger.LogAction(&userID, "session-1", "192.168.1.1", "test-agent", "update", "profile", &userID, true, "", map[string]interface{}{})
	logger.LogAction(&userID, "session-1", "192.168.1.1", "test-agent", "delete", "property", &userID, false, "permission denied", map[string]interface{}{})

	stats := logger.GetAuditStatistics()
	
	assert.Equal(t, int64(3), stats["total_logs"])
	assert.Equal(t, int64(2), stats["successful_actions"])
	assert.Equal(t, int64(1), stats["failed_actions"])
	assert.Greater(t, stats["success_rate"].(float64), 60.0)
}

func TestEncryptionKeyRotation(t *testing.T) {
	db := setupTestDB()
	
	encMgr, err := NewEncryptionManager(db)
	assert.NoError(t, err)

	plaintext := "sensitive data"
	
	encrypted, err := encMgr.EncryptField(plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := encMgr.DecryptField(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestSecurityEventClassification(t *testing.T) {
	tests := []struct {
		eventType string
		riskScore int
		severity  string
	}{
		{"login_failure", 40, "medium"},
		{"brute_force_detected", 90, "critical"},
		{"sql_injection_attempt", 95, "critical"},
		{"xss_attempt", 85, "high"},
		{"suspicious_activity", 70, "high"},
		{"password_reset", 30, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			severity := classifySeverity(tt.riskScore)
			assert.Equal(t, tt.severity, severity)
		})
	}
}

func classifySeverity(riskScore int) string {
	if riskScore >= 80 {
		return "critical"
	} else if riskScore >= 60 {
		return "high"
	} else if riskScore >= 40 {
		return "medium"
	}
	return "low"
}

type InputValidator struct{}

func (iv *InputValidator) ValidateNoSQLInjection(input string) bool {
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/",
		"union", "select", "insert", "update", "delete", "drop", "create", "alter",
		"exec", "execute", "sp_", "xp_",
	}
	
	lowerInput := input
	for _, pattern := range sqlPatterns {
		if contains(lowerInput, pattern) {
			return false
		}
	}
	return true
}

func (iv *InputValidator) ValidateNoXSS(input string) bool {
	xssPatterns := []string{
		"<script", "</script>", "<iframe", "</iframe>",
		"<object", "</object>", "<embed", "</embed>",
		"javascript:", "vbscript:", "data:text/html",
		"onload=", "onerror=", "onclick=",
	}
	
	lowerInput := input
	for _, pattern := range xssPatterns {
		if contains(lowerInput, pattern) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type PasswordValidator struct{}

func (pv *PasswordValidator) IsStrong(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case c >= '!' && c <= '/' || c >= ':' && c <= '@' || c >= '[' && c <= '`' || c >= '{' && c <= '~':
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}
