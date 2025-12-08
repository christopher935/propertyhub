package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// RBAC Models
type SecurityUser struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"unique;not null"`
	Email     string         `json:"email" gorm:"unique;not null"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Roles     []SecurityRole `json:"roles" gorm:"many2many:user_roles;"`
}

type SecurityRole struct {
	ID          uint                 `json:"id" gorm:"primaryKey"`
	Name        string               `json:"name" gorm:"unique;not null"`
	Description string               `json:"description"`
	IsActive    bool                 `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Permissions []SecurityPermission `json:"permissions" gorm:"many2many:role_permissions;"`
}

type SecurityPermission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"unique;not null"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RBACHandler handles role-based access control
type RBACHandler struct {
	db *gorm.DB
}

func NewRBACHandler(db *gorm.DB) *RBACHandler {
	return &RBACHandler{db: db}
}

// Initialize RBAC tables
func (h *RBACHandler) InitializeTables() error {
	return h.db.AutoMigrate(&SecurityUser{}, &SecurityRole{}, &SecurityPermission{})
}

// Role Management
func (h *RBACHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	var roles []SecurityRole
	if err := h.db.Preload("Permissions").Find(&roles).Error; err != nil {
		http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"roles":   roles,
	})
}

func (h *RBACHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		PermissionIDs []uint `json:"permission_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := SecurityRole{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	if err := h.db.Create(&role).Error; err != nil {
		http.Error(w, "Failed to create role", http.StatusInternalServerError)
		return
	}

	// Assign permissions
	if len(req.PermissionIDs) > 0 {
		var permissions []SecurityPermission
		h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions)
		h.db.Model(&role).Association("Permissions").Append(permissions)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"role":    role,
		"message": "Role created successfully",
	})
}

// Permission Management
func (h *RBACHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	var permissions []SecurityPermission
	if err := h.db.Find(&permissions).Error; err != nil {
		http.Error(w, "Failed to fetch permissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"permissions": permissions,
	})
}

func (h *RBACHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	permission := SecurityPermission{
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}

	if err := h.db.Create(&permission).Error; err != nil {
		http.Error(w, "Failed to create permission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"permission": permission,
		"message":    "Permission created successfully",
	})
}

// User Role Management
func (h *RBACHandler) UpdateUserRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		RoleIDs []uint `json:"role_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user SecurityUser
	if err := h.db.First(&user, uint(userID)).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Clear existing roles and assign new ones
	h.db.Model(&user).Association("Roles").Clear()
	if len(req.RoleIDs) > 0 {
		var roles []SecurityRole
		h.db.Where("id IN ?", req.RoleIDs).Find(&roles)
		h.db.Model(&user).Association("Roles").Append(roles)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User roles updated successfully",
	})
}

// Permission Check
func (h *RBACHandler) CheckUserPermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	resource := r.URL.Query().Get("resource")
	action := r.URL.Query().Get("action")

	var user SecurityUser
	if err := h.db.Preload("Roles.Permissions").First(&user, uint(userID)).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	hasPermission := false
	userPermissions := make([]string, 0)

	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissionKey := fmt.Sprintf("%s:%s", permission.Resource, permission.Action)
			userPermissions = append(userPermissions, permissionKey)

			if permission.Resource == resource && permission.Action == action {
				hasPermission = true
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":          true,
		"has_permission":   hasPermission,
		"user_permissions": userPermissions,
	})
}
