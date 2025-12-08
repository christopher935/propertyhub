package handlers

import (
	"github.com/gin-gonic/gin"
)

// AdminPageData is the standard data structure for ALL admin template pages
// This ensures consistent data availability across all admin templates
type AdminPageData struct {
	PageTitle         string
	User              UserData
	NotificationCount int
}

// UserData represents the current authenticated user
type UserData struct {
	ID        int64
	Name      string
	Email     string
	Role      string
	Initials  string
	AvatarURL string
}

// GetAdminPageData creates standard page data from Gin context
// Use this helper in ALL admin page handlers to ensure consistency
func GetAdminPageData(c *gin.Context, pageTitle string) AdminPageData {
	user := getUserFromContext(c)
	
	return AdminPageData{
		PageTitle:         pageTitle,
		User:              user,
		NotificationCount: getNotificationCount(c),
	}
}

// getUserFromContext extracts user data from Gin context
// Assumes auth middleware has set user info in context
func getUserFromContext(c *gin.Context) UserData {
	userID, exists := c.Get("user_id")
	if !exists {
		return UserData{
			Name:     "Guest",
			Role:     "Viewer",
			Initials: "GU",
		}
	}
	
	userName, _ := c.Get("user_name")
	userEmail, _ := c.Get("user_email")
	userRole, _ := c.Get("user_role")
	
	name := "Admin"
	if userName != nil {
		name = userName.(string)
	}
	
	role := "Administrator"
	if userRole != nil {
		role = userRole.(string)
	}
	
	initials := getInitials(name)
	
	return UserData{
		ID:       userID.(int64),
		Name:     name,
		Email:    userEmail.(string),
		Role:     role,
		Initials: initials,
	}
}

// getInitials extracts initials from a full name
func getInitials(name string) string {
	if name == "" {
		return "??"
	}
	
	parts := splitName(name)
	if len(parts) == 0 {
		return "??"
	}
	
	if len(parts) == 1 {
		if len(parts[0]) > 0 {
			return string(parts[0][0])
		}
		return "??"
	}
	
	first := ""
	last := ""
	if len(parts[0]) > 0 {
		first = string(parts[0][0])
	}
	if len(parts[len(parts)-1]) > 0 {
		last = string(parts[len(parts)-1][0])
	}
	
	return first + last
}

// splitName splits a full name into parts
func splitName(name string) []string {
	var parts []string
	current := ""
	
	for _, char := range name {
		if char == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

// getNotificationCount gets unread notification count for current user
func getNotificationCount(c *gin.Context) int {
	return 0
}
