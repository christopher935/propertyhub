package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// SimpleAuthManager handles authentication with raw SQL (no JWT)
type SimpleAuthManager struct {
	db *sql.DB
}

// NewSimpleAuthManager creates a new simple auth manager
func NewSimpleAuthManager(db *sql.DB) *SimpleAuthManager {
	return &SimpleAuthManager{
		db: db,
	}
}

type AdminUser = models.AdminUser

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success   bool       `json:"success"`
	Token     string     `json:"token,omitempty"`
	User      *AdminUser `json:"user,omitempty"`
	ExpiresAt time.Time  `json:"expires_at,omitempty"`
	Message   string     `json:"message,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// AuthenticateUser authenticates a user with email/password
func (am *SimpleAuthManager) AuthenticateUser(email, password string) (*LoginResponse, error) {
	query := `SELECT id, username, email, password_hash, role, active, created_at, last_login, login_count 
			  FROM admin_users WHERE email = $1 AND active = true`

	var user AdminUser
	var lastLogin sql.NullTime

	err := am.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Active, &user.CreatedAt, &lastLogin, &user.LoginCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Log failed login attempt
		am.logSecurityEvent(user.ID, "failed_login", map[string]interface{}{
			"email":  email,
			"reason": "invalid_password",
		})
		return nil, fmt.Errorf("invalid credentials")
	}

	// Convert sql.NullTime to *time.Time
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	} else {
		user.LastLogin = nil
	}

	// Update last login and login count
	_, err = am.db.Exec(`UPDATE admin_users SET last_login = CURRENT_TIMESTAMP, login_count = login_count + 1 WHERE id = $1`, user.ID)
	if err != nil {
		log.Printf("Failed to update login info: %v", err)
	}

	// Generate session token
	sessionToken, expiresAt, err := am.GenerateSessionToken(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session: %v", err)
	}

	// Log successful login
	am.logSecurityEvent(user.ID, "login", map[string]interface{}{
		"email": email,
	})

	return &LoginResponse{
		Success:   true,
		Token:     sessionToken,
		User:      &user,
		ExpiresAt: expiresAt,
		Message:   "Login successful",
	}, nil
}

// GenerateSessionToken generates a simple session token for the user
func (am *SimpleAuthManager) GenerateSessionToken(user *AdminUser) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	sessionToken := am.generateSessionID()
	tokenHash := am.hashSessionToken(sessionToken)

	_, err := am.db.Exec(`INSERT INTO admin_sessions (user_id, token_hash, expires_at, ip_address, user_agent, active) 
						 VALUES ($1, $2, $3, NULL, NULL, true)`,
		user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to store session: %v", err)
	}

	return sessionToken, expiresAt, nil
}

// ValidateSessionToken validates a session token and returns the user
func (am *SimpleAuthManager) ValidateSessionToken(sessionToken string) (*AdminUser, error) {
	tokenHash := am.hashSessionToken(sessionToken)

	var userID string
	err := am.db.QueryRow(`SELECT user_id FROM admin_sessions WHERE token_hash = $1 AND active = true AND expires_at > CURRENT_TIMESTAMP`,
		tokenHash).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired: %v", err)
	}

	query := `SELECT id, username, email, role, active, created_at, last_login, login_count 
			  FROM admin_users WHERE id = $1 AND active = true`

	var user AdminUser
	var lastLogin sql.NullTime

	err = am.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.Active, &user.CreatedAt, &lastLogin, &user.LoginCount,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	} else {
		user.LastLogin = nil
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (am *SimpleAuthManager) GetUserByID(userID string) (*AdminUser, error) {
	query := `SELECT id, username, email, role, active, created_at, last_login, login_count 
			  FROM admin_users WHERE id = $1 AND active = true`

	var user AdminUser
	var lastLogin sql.NullTime

	err := am.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.Active, &user.CreatedAt, &lastLogin, &user.LoginCount,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	} else {
		user.LastLogin = nil
	}

	return &user, nil
}

// InvalidateSession invalidates a session token
func (am *SimpleAuthManager) InvalidateSession(sessionToken string) error {
	tokenHash := am.hashSessionToken(sessionToken)
	_, err := am.db.Exec(`UPDATE admin_sessions SET active = false WHERE token_hash = $1`, tokenHash)
	return err
}

// CreateSession creates a new session for a user (interface method)
func (am *SimpleAuthManager) CreateSession(user *AdminUser) (string, error) {
	sessionToken, _, err := am.GenerateSessionToken(user)
	return sessionToken, err
}

// RefreshSession refreshes an existing session (interface method)
func (am *SimpleAuthManager) RefreshSession(token string) (string, error) {
	user, err := am.ValidateSessionToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid session for refresh: %v", err)
	}

	_ = am.InvalidateSession(token)
	return am.CreateSession(user)
}

// GetAllUsers returns all admin users (interface method - returns pointers)
func (am *SimpleAuthManager) GetAllUsers() ([]*AdminUser, error) {
	query := `
		SELECT id, username, email, role, active, created_at, last_login, login_count
		FROM admin_users
		ORDER BY created_at DESC
	`

	rows, err := am.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %v", err)
	}
	defer rows.Close()

	var users []*AdminUser
	for rows.Next() {
		user := &AdminUser{}
		var lastLogin sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Role,
			&user.Active, &user.CreatedAt, &lastLogin, &user.LoginCount,
		)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// CreateUser creates a new admin user (interface method)
func (am *SimpleAuthManager) CreateUser(username, email, password, role string) (*AdminUser, error) {
	var existingID string
	err := am.db.QueryRow(`SELECT id FROM admin_users WHERE username = $1 OR email = $2`, username, email).Scan(&existingID)
	if err == nil {
		return nil, fmt.Errorf("user with username or email already exists")
	}

	hashedPassword, err := am.hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}

	var userID string
	err = am.db.QueryRow(`
		INSERT INTO admin_users (id, username, email, password_hash, role, active, created_at, login_count)
		VALUES (generate_random_uuid(), $1, $2, $3, $4, true, CURRENT_TIMESTAMP, 0)
		RETURNING id
	`, username, email, hashedPassword, role).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	user := &AdminUser{
		ID:         userID,
		Username:   username,
		Email:      email,
		Role:       role,
		Active:     true,
		CreatedAt:  time.Now(),
		LoginCount: 0,
	}

	am.logSecurityEvent(user.ID, "user_created", map[string]interface{}{
		"username": username,
		"email":    email,
		"role":     role,
	})

	return user, nil
}

// UpdateUser updates an existing admin user (interface method)
func (am *SimpleAuthManager) UpdateUser(userID string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	allowedFields := map[string]bool{
		"username": true,
		"email":    true,
		"role":     true,
		"password": true,
		"active":   true,
	}

	for field, value := range updates {
		if !allowedFields[field] {
			return fmt.Errorf("invalid field name: %s", field)
		}

		switch field {
		case "username", "email", "role":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		case "password":
			hashedPassword, err := am.hashPassword(value.(string))
			if err != nil {
				return fmt.Errorf("error hashing password: %v", err)
			}
			setParts = append(setParts, fmt.Sprintf("password_hash = $%d", argIndex))
			args = append(args, hashedPassword)
			argIndex++
		case "active":
			setParts = append(setParts, fmt.Sprintf("active = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	args = append(args, userID)
	query := fmt.Sprintf("UPDATE admin_users SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	_, err := am.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	am.logSecurityEvent(userID, "user_updated", updates)
	return nil
}

// DeactivateUser deactivates an admin user (interface method)
func (am *SimpleAuthManager) DeactivateUser(userID string) error {
	_, err := am.db.Exec(`UPDATE admin_users SET active = false WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("error deactivating user: %v", err)
	}

	_, err = am.db.Exec(`UPDATE admin_sessions SET active = false WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("Error invalidating user sessions: %v", err)
	}

	am.logSecurityEvent(userID, "user_deactivated", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// GetCacheHitRate returns cache hit rate (interface method)
func (am *SimpleAuthManager) GetCacheHitRate() float64 {
	return 0.0
}

// GetActiveSessionCount returns count of active sessions (interface method)
func (am *SimpleAuthManager) GetActiveSessionCount() int64 {
	var count int64
	err := am.db.QueryRow(`SELECT COUNT(*) FROM admin_sessions WHERE active = true AND expires_at > CURRENT_TIMESTAMP`).Scan(&count)
	if err != nil {
		log.Printf("Error getting active session count: %v", err)
		return 0
	}
	return count
}

// RequireAuthFunc middleware for protecting routes (HandlerFunc version)
func (am *SimpleAuthManager) RequireAuthFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
		if sessionToken == authHeader {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		user, err := am.ValidateSessionToken(sessionToken)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", user.ID)
		r.Header.Set("X-User-Role", user.Role)

		next(w, r)
	}
}

// RequireAuth middleware for protecting routes (Handler version for enterprise handlers)
func (am *SimpleAuthManager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
		if sessionToken == authHeader {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		user, err := am.ValidateSessionToken(sessionToken)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", user.ID)
		r.Header.Set("X-User-Role", user.Role)

		next.ServeHTTP(w, r)
	})
}

// GetDashboardMetrics provides dashboard metrics
func (am *SimpleAuthManager) GetDashboardMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	var totalUsers, activeUsers int
	err := am.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN active = true THEN 1 END) as active
		FROM admin_users
	`).Scan(&totalUsers, &activeUsers)
	if err != nil {
		log.Printf("Error getting user counts: %v", err)
		totalUsers = 1
		activeUsers = 1
	}

	var totalSessions, activeSessions int
	err = am.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN active = true AND expires_at > CURRENT_TIMESTAMP THEN 1 END) as active
		FROM admin_sessions
	`).Scan(&totalSessions, &activeSessions)
	if err != nil {
		log.Printf("Error getting session counts: %v", err)
		totalSessions = 0
		activeSessions = 0
	}

	var recentLogins int
	err = am.db.QueryRow(`
		SELECT COUNT(*) 
		FROM security_events 
		WHERE event_type = 'login' AND created_at > CURRENT_TIMESTAMP - INTERVAL '24 hours'
	`).Scan(&recentLogins)
	if err != nil {
		log.Printf("Error getting recent logins: %v", err)
		recentLogins = 0
	}

	metrics["total_users"] = totalUsers
	metrics["active_users"] = activeUsers
	metrics["total_sessions"] = totalSessions
	metrics["active_sessions"] = activeSessions
	metrics["recent_logins"] = recentLogins
	metrics["last_updated"] = time.Now()

	return metrics, nil
}

// GetPropertyInsights provides property-related authentication insights
func (am *SimpleAuthManager) GetPropertyInsights() (map[string]interface{}, error) {
	insights := make(map[string]interface{})

	var propertyViews, propertyEdits int
	err := am.db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN event_type = 'property_view' THEN 1 END) as views,
			COUNT(CASE WHEN event_type = 'property_edit' THEN 1 END) as edits
		FROM security_events 
		WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '7 days'
	`).Scan(&propertyViews, &propertyEdits)
	if err != nil {
		log.Printf("Error getting property insights: %v", err)
		propertyViews = 0
		propertyEdits = 0
	}

	insights["property_views_week"] = propertyViews
	insights["property_edits_week"] = propertyEdits
	insights["admin_activity_level"] = "moderate"

	if propertyViews > 50 {
		insights["admin_activity_level"] = "high"
	} else if propertyViews < 10 {
		insights["admin_activity_level"] = "low"
	}

	return insights, nil
}

// LogPropertyAccess logs when an admin accesses property data
func (am *SimpleAuthManager) LogPropertyAccess(userID, propertyID string, action string) error {
	eventData := map[string]interface{}{
		"property_id": propertyID,
		"action":      action,
	}

	am.logSecurityEvent(userID, "property_access", eventData)
	return nil
}

// GetUserActivity returns recent user activity
func (am *SimpleAuthManager) GetUserActivity(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT 
			se.event_type,
			se.event_data,
			se.created_at,
			au.username,
			au.email
		FROM security_events se
		JOIN admin_users au ON se.user_id = au.id
		ORDER BY se.created_at DESC
		LIMIT $1
	`

	rows, err := am.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying user activity: %v", err)
	}
	defer rows.Close()

	var activities []map[string]interface{}
	for rows.Next() {
		var eventType, eventDataStr, username, email string
		var createdAt time.Time

		err := rows.Scan(&eventType, &eventDataStr, &createdAt, &username, &email)
		if err != nil {
			log.Printf("Error scanning activity row: %v", err)
			continue
		}

		activity := map[string]interface{}{
			"event_type": eventType,
			"event_data": eventDataStr,
			"created_at": createdAt,
			"username":   username,
			"email":      email,
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// Helper functions
func (am *SimpleAuthManager) generateSessionID() string {
	bytes := make([]byte, 64)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Printf("Failed to generate secure session token: %v", err)
		bytes = make([]byte, 32)
		rand.Read(bytes)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func (am *SimpleAuthManager) hashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (am *SimpleAuthManager) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (am *SimpleAuthManager) logSecurityEvent(userID, eventType string, eventData map[string]interface{}) {
	eventDataJSON := "{}"
	if eventData != nil && len(eventData) > 0 {
		jsonParts := []string{}
		for key, value := range eventData {
			jsonParts = append(jsonParts, fmt.Sprintf(`"%s":"%v"`, key, value))
		}
		eventDataJSON = "{" + strings.Join(jsonParts, ",") + "}"
	}

	_, err := am.db.Exec(`INSERT INTO security_events (user_id, event_type, event_data, created_at) 
						  VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`,
		userID, eventType, eventDataJSON)
	if err != nil {
		log.Printf("Failed to log security event: %v", err)
	}
}
