package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
 )

// SetupHandlers handles setup wizard requests
type SetupHandlers struct {
	setupService *services.SetupService
}

// NewSetupHandlers creates new setup handlers
func NewSetupHandlers(setupService *services.SetupService) *SetupHandlers {
	return &SetupHandlers{
		setupService: setupService,
	}
}

// SetupWizardPage shows the main setup wizard page
func (h *SetupHandlers) SetupWizardPage(w http.ResponseWriter, r *http.Request ) {
	if !h.setupService.IsSetupRequired() {
		http.Redirect(w, r, "/admin/dashboard", http.StatusFound )
		return
	}

	progress := h.setupService.GetSetupProgress()
	
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>PropertyHub Setup Wizard</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .setup-step { border: 1px solid #ddd; margin: 10px 0; padding: 20px; border-radius: 5px; }
        .step-complete { background-color: #d4edda; border-color: #c3e6cb; }
        .step-pending { background-color: #f8f9fa; border-color: #dee2e6; }
        .btn { padding: 10px 20px; background-color: #007bff; color: white; border: none; border-radius: 3px; cursor: pointer; }
        .btn:hover { background-color: #0056b3; }
        .btn-success { background-color: #28a745; }
        .btn-success:hover { background-color: #1e7e34; }
        .form-group { margin: 15px 0; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .form-group input, .form-group select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 3px; }
        .progress-bar { width: 100%; height: 20px; background-color: #f8f9fa; border-radius: 10px; margin: 20px 0; }
        .progress-fill { height: 100%; background-color: #28a745; border-radius: 10px; transition: width 0.3s; }
        .header { text-align: center; margin-bottom: 30px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .status-info { background-color: #d1ecf1; border: 1px solid #bee5eb; color: #0c5460; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🏠 PropertyHub Setup Wizard</h1>
        <p>Welcome! Let's get your PropertyHub system configured and ready to go.</p>
        <div class="progress-bar">
            <div class="progress-fill" style="width: {{.ProgressPercent}}%"></div>
        </div>
        <p>Setup Progress: {{.ProgressPercent}}%</p>
    </div>

    <div class="status status-info">
        <strong>Current Status:</strong> Setup Required - Your PropertyHub is running but needs configuration to unlock all features.
    </div>

    <!-- Step 1: Database Configuration -->
    <div class="setup-step {{if .DatabaseConfigured}}step-complete{{else}}step-pending{{end}}">
        <h3>Step 1: Database Configuration {{if .DatabaseConfigured}}✅{{else}}⏳{{end}}</h3>
        {{if .DatabaseConfigured}}
            <p>✅ Database connection configured successfully!</p>
        {{else}}
            <p>Configure your PostgreSQL database connection.</p>
            <form action="/setup/database" method="POST">
                <div class="form-group">
                    <label>Database URL:</label>
                    <input type="text" name="database_url" placeholder="postgresql://user:password@host:port/database" required>
                    <small>You can find this in your DigitalOcean database settings</small>
                </div>
                <div class="form-group">
                    <label>Redis URL (Optional):</label>
                    <input type="text" name="redis_url" placeholder="redis://user:password@host:port">
                    <small>For caching and performance (optional)</small>
                </div>
                <button type="submit" class="btn">Configure Database</button>
            </form>
        {{end}}
    </div>

    <!-- Step 2: Admin User Creation -->
    <div class="setup-step {{if .AdminUserCreated}}step-complete{{else}}step-pending{{end}}">
        <h3>Step 2: Admin User Creation {{if .AdminUserCreated}}✅{{else}}⏳{{end}}</h3>
        {{if .AdminUserCreated}}
            <p>✅ Admin user created successfully!</p>
        {{else}}
            {{if not .DatabaseConfigured}}
                <p>⏸️ Complete database configuration first</p>
            {{else}}
                <p>Create your admin user account.</p>
                <form action="/setup/admin" method="POST">
                    <div class="form-group">
                        <label>Username:</label>
                        <input type="text" name="username" required>
                    </div>
                    <div class="form-group">
                        <label>Email:</label>
                        <input type="email" name="email" required>
                    </div>
                    <div class="form-group">
                        <label>Password:</label>
                        <input type="password" name="password" required>
                    </div>
                    <div class="form-group">
                        <label>First Name:</label>
                        <input type="text" name="first_name" required>
                    </div>
                    <div class="form-group">
                        <label>Last Name:</label>
                        <input type="text" name="last_name" required>
                    </div>
                    <button type="submit" class="btn">Create Admin User</button>
                </form>
            {{end}}
        {{end}}
    </div>

    <!-- Step 3: API Keys Configuration -->
    <div class="setup-step {{if .APIKeysConfigured}}step-complete{{else}}step-pending{{end}}">
        <h3>Step 3: API Keys Configuration (Optional) {{if .APIKeysConfigured}}✅{{else}}⏳{{end}}</h3>
        {{if .APIKeysConfigured}}
            <p>✅ API keys configured successfully!</p>
        {{else}}
            <p>Configure external service API keys (you can skip this and configure later).</p>
            <form action="/setup/apikeys" method="POST">
                <div class="form-group">
                    <label>Follow Up Boss API Key:</label>
                    <input type="text" name="fub_api_key" placeholder="Optional - for CRM integration">
                </div>
                <div class="form-group">
                    <label>SMTP Host:</label>
                    <input type="text" name="smtp_host" placeholder="Optional - for email notifications">
                </div>
                <div class="form-group">
                    <label>SMTP Port:</label>
                    <input type="number" name="smtp_port" placeholder="587">
                </div>
                <div class="form-group">
                    <label>SMTP Username:</label>
                    <input type="text" name="smtp_username" placeholder="Optional">
                </div>
                <div class="form-group">
                    <label>SMTP Password:</label>
                    <input type="password" name="smtp_password" placeholder="Optional">
                </div>
                <button type="submit" class="btn">Configure API Keys</button>
                <button type="button" class="btn" onclick="skipAPIKeys()">Skip for Now</button>
            </form>
        {{end}}
    </div>

    <!-- Step 4: Complete Setup -->
    {{if .CanCompleteSetup}}
    <div class="setup-step step-pending">
        <h3>Step 4: Complete Setup 🚀</h3>
        <p>Great! All required configuration is complete. Click below to finish setup and enter test mode.</p>
        <form action="/setup/complete" method="POST">
            <button type="submit" class="btn btn-success">Complete Setup & Enter Test Mode</button>
        </form>
    </div>
    {{end}}

    <script>
        function skipAPIKeys() {
            fetch('/setup/apikeys/skip', {method: 'POST'})
                .then(() => location.reload());
        }
    </script>
</body>
</html>`

	t, err := template.New("setup").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError )
		return
	}

	data := struct {
		DatabaseConfigured bool
		AdminUserCreated   bool
		APIKeysConfigured  bool
		CanCompleteSetup   bool
		ProgressPercent    int
	}{
		DatabaseConfigured: progress["database_configured"],
		AdminUserCreated:   progress["admin_user_created"],
		APIKeysConfigured:  progress["api_keys_configured"],
		CanCompleteSetup:   h.setupService.CanCompleteSetup(),
		ProgressPercent:    h.calculateProgress(progress),
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// ConfigureDatabase handles database configuration
func (h *SetupHandlers) ConfigureDatabase(w http.ResponseWriter, r *http.Request ) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed )
		return
	}

	databaseURL := r.FormValue("database_url")
	redisURL := r.FormValue("redis_url")

	// Test database connection
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Database connection failed: %v", err ), http.StatusBadRequest )
		return
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.Property{},
		&models.AdminUser{},
		&models.Contact{},
		&models.Booking{},
		&models.SetupStatus{},
		&models.SetupConfiguration{},
		&models.AdminSetup{},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database migration failed: %v", err ), http.StatusBadRequest )
		return
	}

	// Save configuration to environment file
	h.saveToEnvFile("DATABASE_URL", databaseURL)
	if redisURL != "" {
		h.saveToEnvFile("REDIS_URL", redisURL)
	}

	// Mark database as configured
	h.setupService.MarkDatabaseConfigured()

	log.Println("✅ Database configured successfully")
	http.Redirect(w, r, "/setup", http.StatusFound )
}

// CreateAdminUser handles admin user creation
func (h *SetupHandlers) CreateAdminUser(w http.ResponseWriter, r *http.Request ) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed )
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	// email, firstName and lastName not used in current AdminUser model

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Password hashing failed", http.StatusInternalServerError )
		return
	}

	// Connect to database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		http.Error(w, "Database not configured", http.StatusBadRequest )
		return
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError )
		return
	}

	// Clear any existing admin users to avoid conflicts
	if err := db.Exec("DELETE FROM admin_users").Error; err != nil {
		log.Printf("Warning: Could not clear existing admin users: %v", err)
	}

	// Create admin user
	adminUser := models.AdminUser{
		Username: username,
		PasswordHash: string(hashedPassword),
	}

	if err := db.Create(&adminUser).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to create admin user: %v", err ), http.StatusInternalServerError )
		return
	}

	log.Printf("✅ Created admin user: %s", username)

	// Mark admin user as created
	h.setupService.MarkAdminUserCreated()

	log.Println("✅ Admin user created successfully")
	http.Redirect(w, r, "/setup", http.StatusFound )
}

// ConfigureAPIKeys handles API key configuration
func (h *SetupHandlers) ConfigureAPIKeys(w http.ResponseWriter, r *http.Request ) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed )
		return
	}

	// Save API keys to environment
	if fubKey := r.FormValue("fub_api_key"); fubKey != "" {
		h.saveToEnvFile("FUB_API_KEY", fubKey)
	}
	if smtpHost := r.FormValue("smtp_host"); smtpHost != "" {
		h.saveToEnvFile("SMTP_HOST", smtpHost)
	}
	if smtpPort := r.FormValue("smtp_port"); smtpPort != "" {
		h.saveToEnvFile("SMTP_PORT", smtpPort)
	}
	if smtpUser := r.FormValue("smtp_username"); smtpUser != "" {
		h.saveToEnvFile("SMTP_USERNAME", smtpUser)
	}
	if smtpPass := r.FormValue("smtp_password"); smtpPass != "" {
		h.saveToEnvFile("SMTP_PASSWORD", smtpPass)
	}

	// Mark API keys as configured
	h.setupService.MarkAPIKeysConfigured()

	log.Println("✅ API keys configured successfully")
	http.Redirect(w, r, "/setup", http.StatusFound )
}

// SkipAPIKeys handles skipping API key configuration
func (h *SetupHandlers) SkipAPIKeys(w http.ResponseWriter, r *http.Request ) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed )
		return
	}

	// Mark API keys as configured (even though skipped)
	h.setupService.MarkAPIKeysConfigured()

	w.WriteHeader(http.StatusOK )
	json.NewEncoder(w).Encode(map[string]string{"status": "skipped"})
}

// CompleteSetup handles setup completion
func (h *SetupHandlers) CompleteSetup(w http.ResponseWriter, r *http.Request ) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed )
		return
	}

	if !h.setupService.CanCompleteSetup() {
		http.Error(w, "Setup requirements not met", http.StatusBadRequest )
		return
	}

	// Complete setup and enter test mode
	if err := h.setupService.CompleteSetup(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to complete setup: %v", err ), http.StatusInternalServerError )
		return
	}

	// Security: Remove setup files after completion
	h.cleanupSetupFiles()

	log.Println("🎉 Setup completed successfully - entering test mode")
	log.Println("🔒 Setup wizard files removed for security")
	
	// Redirect to a completion page that will restart the app
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK )
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Setup Complete</title>
    <meta http-equiv="refresh" content="3;url=/">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #28a745; font-size: 24px; margin: 20px 0; }
    </style>
</head>
<body>
    <h1>🎉 Setup Complete!</h1>
    <div class="success">PropertyHub is now configured and running in TEST MODE</div>
    <p>The setup wizard has been permanently removed for security.</p>
    <p>Redirecting to your PropertyHub dashboard...</p>
    <p><a href="/">Continue to PropertyHub</a></p>
</body>
</html>
	` ))
}

// cleanupSetupFiles removes setup-related files for security
func (h *SetupHandlers) cleanupSetupFiles() {
	filesToRemove := []string{
		"internal/handlers/setup_handlers.go",
		"internal/services/setup_service.go",
		"internal/models/setup.go",
	}

	for _, file := range filesToRemove {
		if err := os.Remove(file); err != nil {
			log.Printf("Warning: Could not remove setup file %s: %v", file, err)
		} else {
			log.Printf("🗑️ Removed setup file: %s", file)
		}
	}

	// Also remove the setup status file since it's no longer needed
	if err := os.Remove("config/setup_status.json"); err != nil {
		log.Printf("Warning: Could not remove setup status file: %v", err)
	} else {
		log.Println("🗑️ Removed setup status file")
	}

	log.Println("🔒 Setup wizard permanently disabled - files removed")
}

// calculateProgress calculates setup progress percentage
func (h *SetupHandlers) calculateProgress(progress map[string]bool) int {
	total := 3 // database, admin user, api keys
	completed := 0

	if progress["database_configured"] {
		completed++
	}
	if progress["admin_user_created"] {
		completed++
	}
	if progress["api_keys_configured"] {
		completed++
	}

	return (completed * 100) / total
}

// saveToEnvFile saves environment variables to a file
func (h *SetupHandlers) saveToEnvFile(key, value string) {
	envFile := ".env"
	
	// Simple append approach for now
	content := fmt.Sprintf("%s=%s\n", key, value)
	
	// Append to file
	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: Could not save to env file: %v", err)
		return
	}
	defer f.Close()
	
	f.WriteString(content)
	log.Printf("💾 Saved %s to environment", key)
}

// RegisterSetupRoutes registers all setup wizard routes
func RegisterSetupRoutes(mux *http.ServeMux, setupService *services.SetupService ) {
	setupHandlers := NewSetupHandlers(setupService)
	
	mux.HandleFunc("/setup", setupHandlers.SetupWizardPage)
	mux.HandleFunc("/setup/database", setupHandlers.ConfigureDatabase)
	mux.HandleFunc("/setup/admin", setupHandlers.CreateAdminUser)
	mux.HandleFunc("/setup/apikeys", setupHandlers.ConfigureAPIKeys)
	mux.HandleFunc("/setup/apikeys/skip", setupHandlers.SkipAPIKeys)
	mux.HandleFunc("/setup/complete", setupHandlers.CompleteSetup)
}

