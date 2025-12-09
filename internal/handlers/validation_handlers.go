package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ValidationHandler handles form validation operations
type ValidationHandler struct{}

// NewValidationHandler creates a new validation handler
func NewValidationHandler() *ValidationHandler {
	return &ValidationHandler{}
}

// ValidationRequest represents a validation request
type ValidationRequest struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Rules []string    `json:"rules"`
}

// ValidationResponse represents a validation response
type ValidationResponse struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
}

// ValidateEmail validates email addresses
func (h *ValidationHandler) ValidateEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := h.validateEmailAddress(request.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidatePhone validates phone numbers
func (h *ValidationHandler) ValidatePhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Phone string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := h.validatePhoneNumber(request.Phone)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateBookingForm validates complete booking form
func (h *ValidationHandler) ValidateBookingForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
		ShowingDate   string `json:"showingDate"`
		ShowingTime   string `json:"showingTime"`
		AttendeeCount string `json:"attendeeCount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errors := []string{}

	// Validate required fields
	if strings.TrimSpace(request.FirstName) == "" {
		errors = append(errors, "First name is required")
	}

	if strings.TrimSpace(request.LastName) == "" {
		errors = append(errors, "Last name is required")
	}

	// Validate email
	emailValidation := h.validateEmailAddress(request.Email)
	if !emailValidation.Valid {
		errors = append(errors, emailValidation.Errors...)
	}

	// Validate phone
	phoneValidation := h.validatePhoneNumber(request.Phone)
	if !phoneValidation.Valid {
		errors = append(errors, phoneValidation.Errors...)
	}

	// Validate showing date
	if request.ShowingDate == "" {
		errors = append(errors, "Showing date is required")
	} else {
		if _, err := time.Parse("2006-01-02", request.ShowingDate); err != nil {
			errors = append(errors, "Invalid showing date format")
		} else {
			// Check if date is in the future
			showingDate, _ := time.Parse("2006-01-02", request.ShowingDate)
			if showingDate.Before(time.Now().Truncate(24 * time.Hour)) {
				errors = append(errors, "Showing date must be in the future")
			}
		}
	}

	// Validate showing time
	if request.ShowingTime == "" {
		errors = append(errors, "Showing time is required")
	} else {
		if _, err := time.Parse("15:04", request.ShowingTime); err != nil {
			errors = append(errors, "Invalid showing time format")
		}
	}

	// Validate attendee count
	if request.AttendeeCount == "" {
		errors = append(errors, "Attendee count is required")
	} else {
		if count, err := strconv.Atoi(request.AttendeeCount); err != nil {
			errors = append(errors, "Invalid attendee count")
		} else if count < 1 || count > 10 {
			errors = append(errors, "Attendee count must be between 1 and 10")
		}
	}

	valid := len(errors) == 0
	message := "Form validation successful"
	if !valid {
		message = "Form validation failed"
	}

	response := ValidationResponse{
		Valid:   valid,
		Errors:  errors,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateContactForm validates contact form
func (h *ValidationHandler) ValidateContactForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errors := []string{}

	// Validate required fields
	if strings.TrimSpace(request.Name) == "" {
		errors = append(errors, "Name is required")
	}

	if strings.TrimSpace(request.Message) == "" {
		errors = append(errors, "Message is required")
	}

	// Validate email
	emailValidation := h.validateEmailAddress(request.Email)
	if !emailValidation.Valid {
		errors = append(errors, emailValidation.Errors...)
	}

	// Validate phone (optional but format check if provided)
	if request.Phone != "" {
		phoneValidation := h.validatePhoneNumber(request.Phone)
		if !phoneValidation.Valid {
			errors = append(errors, phoneValidation.Errors...)
		}
	}

	valid := len(errors) == 0
	message := "Contact form validation successful"
	if !valid {
		message = "Contact form validation failed"
	}

	response := ValidationResponse{
		Valid:   valid,
		Errors:  errors,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper validation methods

func (h *ValidationHandler) validateEmailAddress(email string) ValidationResponse {
	email = strings.TrimSpace(email)
	
	if email == "" {
		return ValidationResponse{
			Valid:  false,
			Errors: []string{"Email is required"},
		}
	}

	// Email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ValidationResponse{
			Valid:  false,
			Errors: []string{"Please enter a valid email address"},
		}
	}

	// Check email length
	if len(email) > 254 {
		return ValidationResponse{
			Valid:  false,
			Errors: []string{"Email address is too long"},
		}
	}

	return ValidationResponse{
		Valid:   true,
		Message: "Valid email address",
	}
}

func (h *ValidationHandler) validatePhoneNumber(phone string) ValidationResponse {
	phone = strings.TrimSpace(phone)
	
	if phone == "" {
		return ValidationResponse{
			Valid:  false,
			Errors: []string{"Phone number is required"},
		}
	}

	// Remove common formatting characters
	cleanPhone := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Check length (US phone numbers)
	if len(cleanPhone) < 10 {
		return ValidationResponse{
			Valid:  false,
			Errors: []string{"Phone number is too short"},
		}
	}

	if len(cleanPhone) > 11 {
		return ValidationResponse{
			Valid:  false,
			Errors: []string{"Phone number is too long"},
		}
	}

	// US phone number pattern
	phoneRegex := regexp.MustCompile(`^(\+?1)?[2-9]\d{2}[2-9]\d{2}\d{4}$`)
	if !phoneRegex.MatchString(cleanPhone) {
		return ValidationResponse{
			Valid:  false,
			Errors: []string{"Please enter a valid US phone number"},
		}
	}

	return ValidationResponse{
		Valid:   true,
		Message: "Valid phone number",
	}
}

// RegisterValidationRoutes registers all validation routes
// Deprecated: Use Gin routes instead
func RegisterValidationRoutes(mux *http.ServeMux) {
	h := NewValidationHandler()

	mux.HandleFunc("/api/v1/validation/email", h.ValidateEmail)
	mux.HandleFunc("/api/v1/validation/phone", h.ValidatePhone)
	mux.HandleFunc("/api/v1/validation/booking-form", h.ValidateBookingForm)
	mux.HandleFunc("/api/v1/validation/contact-form", h.ValidateContactForm)

	log.Println("✅ Validation routes registered")
}

// ============================================================================
// GIN-COMPATIBLE METHODS
// ============================================================================

func (h *ValidationHandler) ValidateEmailGin(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	response := h.validateEmailAddress(request.Email)
	c.JSON(http.StatusOK, response)
}

func (h *ValidationHandler) ValidatePhoneGin(c *gin.Context) {
	var request struct {
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	response := h.validatePhoneNumber(request.Phone)
	c.JSON(http.StatusOK, response)
}

func (h *ValidationHandler) ValidateBookingFormGin(c *gin.Context) {
	var request struct {
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
		ShowingDate   string `json:"showingDate"`
		ShowingTime   string `json:"showingTime"`
		AttendeeCount string `json:"attendeeCount"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	errors := []string{}
	if strings.TrimSpace(request.FirstName) == "" {
		errors = append(errors, "First name is required")
	}
	if strings.TrimSpace(request.LastName) == "" {
		errors = append(errors, "Last name is required")
	}
	if ev := h.validateEmailAddress(request.Email); !ev.Valid {
		errors = append(errors, ev.Errors...)
	}
	if pv := h.validatePhoneNumber(request.Phone); !pv.Valid {
		errors = append(errors, pv.Errors...)
	}
	if request.ShowingDate == "" {
		errors = append(errors, "Showing date is required")
	} else if sd, err := time.Parse("2006-01-02", request.ShowingDate); err != nil {
		errors = append(errors, "Invalid showing date format")
	} else if sd.Before(time.Now().Truncate(24 * time.Hour)) {
		errors = append(errors, "Showing date must be in the future")
	}
	if request.ShowingTime == "" {
		errors = append(errors, "Showing time is required")
	} else if _, err := time.Parse("15:04", request.ShowingTime); err != nil {
		errors = append(errors, "Invalid showing time format")
	}
	if request.AttendeeCount == "" {
		errors = append(errors, "Attendee count is required")
	} else if count, err := strconv.Atoi(request.AttendeeCount); err != nil {
		errors = append(errors, "Invalid attendee count")
	} else if count < 1 || count > 10 {
		errors = append(errors, "Attendee count must be between 1 and 10")
	}
	valid := len(errors) == 0
	message := "Form validation successful"
	if !valid {
		message = "Form validation failed"
	}
	c.JSON(http.StatusOK, ValidationResponse{Valid: valid, Errors: errors, Message: message})
}

func (h *ValidationHandler) ValidateContactFormGin(c *gin.Context) {
	var request struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	errors := []string{}
	if strings.TrimSpace(request.Name) == "" {
		errors = append(errors, "Name is required")
	}
	if strings.TrimSpace(request.Message) == "" {
		errors = append(errors, "Message is required")
	}
	if ev := h.validateEmailAddress(request.Email); !ev.Valid {
		errors = append(errors, ev.Errors...)
	}
	if request.Phone != "" {
		if pv := h.validatePhoneNumber(request.Phone); !pv.Valid {
			errors = append(errors, pv.Errors...)
		}
	}
	valid := len(errors) == 0
	message := "Contact form validation successful"
	if !valid {
		message = "Contact form validation failed"
	}
	c.JSON(http.StatusOK, ValidationResponse{Valid: valid, Errors: errors, Message: message})
}
