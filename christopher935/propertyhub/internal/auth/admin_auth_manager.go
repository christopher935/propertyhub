package auth

import (
	"database/sql"
)

// AdminAuthManager is an alias for SimpleAuthManager to maintain compatibility
type AdminAuthManager = SimpleAuthManager

// NewAdminAuthManager creates a new admin auth manager (alias for SimpleAuthManager)
func NewAdminAuthManager(db *sql.DB) *AdminAuthManager {
	return NewSimpleAuthManager(db)
}
