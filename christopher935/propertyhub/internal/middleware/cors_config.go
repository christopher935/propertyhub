package middleware

import (
	"net/http"
	"os"
	"strings"
)

// GetAllowedOrigin returns the CORS allowed origin based on environment
func GetAllowedOrigin() string {
	// Get domain from environment variable
	if origin := os.Getenv("CORS_ALLOWED_ORIGIN"); origin != "" {
		return origin
	}
	
	// Check for common domain environment variables
	if domain := os.Getenv("DOMAIN"); domain != "" {
		return domain
	}
	
	if domain := os.Getenv("BASE_URL"); domain != "" {
		return domain
	}
	
	// Safe fallback for development - localhost only
	return "http://localhost:8080"
}

// SetCORSHeaders sets secure CORS headers for API responses
func SetCORSHeaders(w http.ResponseWriter, methods ...string) {
	origin := GetAllowedOrigin()
	w.Header().Set("Access-Control-Allow-Origin", origin)
	
	if len(methods) > 0 {
		methodsList := strings.Join(methods, ", ")
		w.Header().Set("Access-Control-Allow-Methods", methodsList)
	}
	
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
