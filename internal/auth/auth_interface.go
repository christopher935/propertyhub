package auth

import (
	"net/http"
)

// AuthenticationManager defines a unified interface for all authentication methods
type AuthenticationManager interface {
	// Core authentication methods
	AuthenticateUser(email, password string) (*LoginResponse, error)
	ValidateSessionToken(token string) (*AdminUser, error)
	GetUserByID(userID string) (*AdminUser, error)
	InvalidateSession(token string) error

	// Session management
	CreateSession(user *AdminUser) (string, error)
	RefreshSession(token string) (string, error)

	// Admin management
	GetAllUsers() ([]*AdminUser, error)
	CreateUser(username, email, password, role string) (*AdminUser, error)
	UpdateUser(userID string, updates map[string]interface{}) error
	DeactivateUser(userID string) error

	// Performance monitoring
	GetCacheHitRate() float64 // Returns 0.0 for non-cached implementations
	GetActiveSessionCount() int64

	// Authentication middleware
	RequireAuth(next http.Handler) http.Handler
}

// AuthenticationStats provides performance metrics
type AuthenticationStats struct {
	TotalAuthAttempts   int64   `json:"total_auth_attempts"`
	SuccessfulAuths     int64   `json:"successful_auths"`
	FailedAuths         int64   `json:"failed_auths"`
	ActiveSessions      int64   `json:"active_sessions"`
	CacheHitRate        float64 `json:"cache_hit_rate"`
	AverageResponseTime float64 `json:"average_response_time_ms"`
}

// AuthenticationManagerWithStats extends the basic interface with statistics
type AuthenticationManagerWithStats interface {
	AuthenticationManager
	GetStats() *AuthenticationStats
	ResetStats()
}
