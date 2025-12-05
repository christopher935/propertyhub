package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminAuthHandlers struct {
	db                  *gorm.DB
	jwtSecret           string
	bruteForceProtection *security.BruteForceProtection
}

type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me,omitempty"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

type AdminUser struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	Username     string     `json:"username" gorm:"uniqueIndex"`
	Email        string     `json:"email" gorm:"uniqueIndex"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	Active       bool       `json:"active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLogin    *time.Time `json:"last_login"`
	LoginCount   int64      `json:"login_count"`
}

func NewAdminAuthHandlers(db *gorm.DB, jwtSecret string) *AdminAuthHandlers {
	return &AdminAuthHandlers{
		db:                  db,
		jwtSecret:           jwtSecret,
		bruteForceProtection: security.NewBruteForceProtection(db, nil),
	}
}

func NewAdminAuthHandlersWithRedis(db *gorm.DB, jwtSecret string, redisClient *redis.Client) *AdminAuthHandlers {
	return &AdminAuthHandlers{
		db:                  db,
		jwtSecret:           jwtSecret,
		bruteForceProtection: security.NewBruteForceProtection(db, redisClient),
	}
}

func (h *AdminAuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.Username == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Username or email required",
		})
		return
	}

	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
			clientIP = r.RemoteAddr[:idx]
		} else {
			clientIP = r.RemoteAddr
		}
	} else {
		if idx := strings.Index(clientIP, ","); idx != -1 {
			clientIP = strings.TrimSpace(clientIP[:idx])
		}
	}

	allowed, remaining, retryAfter, err := h.bruteForceProtection.CheckLoginAttempt(req.Username, clientIP)
	if err != nil {
		log.Printf("Brute force check error: %v", err)
	}

	if !allowed {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", retryAfter.String())
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Too many failed login attempts. Account temporarily locked. Please try again later.",
		})
		return
	}

	if remaining > 0 && remaining <= 2 {
		log.Printf("⚠️ Warning: Only %d login attempts remaining for %s", remaining, req.Username)
	}

	var user AdminUser
	var query string
	var searchValue string

	if strings.Contains(req.Username, "@") {
		query = "email = ? AND active = ?"
		searchValue = req.Username
	} else {
		query = "username = ? AND active = ?"
		searchValue = req.Username
	}

	result := h.db.Where(query, searchValue, true).First(&user)
	if result.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	bcryptErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if bcryptErr != nil {
		h.bruteForceProtection.RecordFailedAttempt(req.Username, clientIP, r.Header.Get("User-Agent"))
		log.Printf("Failed login attempt for user: %s from IP: %s", user.Email, clientIP)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	now := time.Now()
	h.db.Model(&user).Updates(map[string]interface{}{
		"last_login":  &now,
		"login_count": user.LoginCount + 1,
		"updated_at":  now,
	})

	h.bruteForceProtection.RecordSuccessfulLogin(req.Username)
	log.Printf("✅ Successful login for user: %s (Role: %s) from IP: %s", user.Email, user.Role, clientIP)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Success: true,
		Token:   tokenString,
		Message: "Login successful",
	})
}

func (h *AdminAuthHandlers) AuthStatus(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
			"message":       "No authorization header",
		})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
			"message":       "Invalid token",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"message":       "Valid session",
	})
}

func RegisterAdminAuthRoutes(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	h := NewAdminAuthHandlers(db, jwtSecret)
	r.POST("/api/v1/admin/login", gin.WrapF(h.Login))
	r.GET("/admin/auth/status", gin.WrapF(h.AuthStatus))

	log.Println("✅ Enterprise admin authentication routes registered successfully")
}
