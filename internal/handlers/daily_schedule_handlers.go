package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// DailyScheduleHandlers handles daily schedule API endpoints
type DailyScheduleHandlers struct {
	dailyScheduleService *services.DailyScheduleService
}

// NewDailyScheduleHandlers creates new daily schedule handlers
func NewDailyScheduleHandlers(db *gorm.DB) *DailyScheduleHandlers {
	return &DailyScheduleHandlers{
		dailyScheduleService: services.NewDailyScheduleService(db),
	}
}

// RegisterRoutes registers all daily schedule routes
func (dsh *DailyScheduleHandlers) RegisterRoutes(mux *http.ServeMux) {
	// Daily schedule routes
	mux.HandleFunc("/api/v1/schedule/daily", dsh.GetDailySchedule)
	mux.HandleFunc("/api/v1/schedule/today", dsh.GetTodaySchedule)
	mux.HandleFunc("/api/v1/schedule/mobile", dsh.GetMobileSchedule)
}

// GetDailySchedule returns the daily schedule for today
func (dsh *DailyScheduleHandlers) GetDailySchedule(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get agent name from query parameters (default to "Agent")
	agentName := r.URL.Query().Get("agent")
	if agentName == "" {
		agentName = "Agent"
	}

	// Get today's schedule
	today := time.Now()
	schedule, err := dsh.dailyScheduleService.GetDailySchedule(today)
	if err != nil {
		log.Printf("Error getting daily schedule: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get daily schedule: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// GetDailyScheduleByDate returns the daily schedule for a specific date
func (dsh *DailyScheduleHandlers) GetDailyScheduleByDate(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get date from URL parameters
	vars := mux.Vars(r)
	dateStr := vars["date"]

	// Parse date (expected format: YYYY-MM-DD)
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Get agent name from query parameters
	agentName := r.URL.Query().Get("agent")
	if agentName == "" {
		agentName = "Agent"
	}

	// Get schedule for the specified date
	schedule, err := dsh.dailyScheduleService.GetDailySchedule(date)
	if err != nil {
		log.Printf("Error getting daily schedule for %s: %v", dateStr, err)
		http.Error(w, fmt.Sprintf("Failed to get daily schedule: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// GetTodaySchedule returns a simplified version for quick access
func (dsh *DailyScheduleHandlers) GetTodaySchedule(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	agentName := r.URL.Query().Get("agent")
	if agentName == "" {
		agentName = "Agent"
	}

	today := time.Now()
	schedule, err := dsh.dailyScheduleService.GetDailySchedule(today)
	if err != nil {
		log.Printf("Error getting today's schedule: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get today's schedule: %v", err), http.StatusInternalServerError)
		return
	}

	// Create simplified response for quick access
	simplifiedResponse := map[string]interface{}{
		"greeting":           fmt.Sprintf("Good %s, %s!", dsh.getGreeting(), agentName),
		"agent_name":         agentName,
		"formatted_date":     today.Format("Monday, January 2, 2006"),
		"day_name":           today.Format("Monday"),
		"total_showings":     dsh.countItemsByType(schedule, "showing"),
		"confirmed_showings": dsh.countConfirmedShowings(schedule),
		"pending_showings":   dsh.countPendingShowings(schedule),
		"urgent_todos":       dsh.countUrgentItems(schedule),
		"overdue_items":      dsh.countOverdueItems(schedule),
		"next_appointment":   dsh.getNextAppointment(schedule),
		"top_priority_todo":  dsh.getTopPriorityItem(schedule),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(simplifiedResponse)
}

// GetMobileSchedule returns optimized data for mobile interface
func (dsh *DailyScheduleHandlers) GetMobileSchedule(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	agentName := r.URL.Query().Get("agent")
	if agentName == "" {
		agentName = "Agent"
	}

	today := time.Now()
	schedule, err := dsh.dailyScheduleService.GetDailySchedule(today)
	if err != nil {
		log.Printf("Error getting mobile schedule: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get mobile schedule: %v", err), http.StatusInternalServerError)
		return
	}

	// Optimize for mobile - limit items and focus on most important
	mobileResponse := map[string]interface{}{
		"header": map[string]interface{}{
			"greeting":   fmt.Sprintf("Good %s, %s!", dsh.getGreeting(), agentName),
			"agent_name": agentName,
			"date":       today.Format("Monday, January 2, 2006"),
			"day_name":   today.Format("Monday"),
		},
		"summary": map[string]interface{}{
			"showings_today": dsh.countItemsByType(schedule, "showing"),
			"confirmed":      dsh.countConfirmedShowings(schedule),
			"pending":        dsh.countPendingShowings(schedule),
			"urgent_todos":   dsh.countUrgentItems(schedule),
			"overdue_items":  dsh.countOverdueItems(schedule),
		},
		"next_items":  dsh.getNextItems(schedule, 5),        // Next 5 schedule items
		"top_todos":   dsh.getTopPriorityItems(schedule, 5), // Top 5 priority todos
		"quick_stats": dsh.getQuickStats(schedule),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mobileResponse)
}

// CompleteScheduleItem marks a schedule item as completed
func (dsh *DailyScheduleHandlers) CompleteScheduleItem(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	itemID := vars["id"]

	// Parse request body for completion details
	var completionData struct {
		Notes  string `json:"notes"`
		Rating int    `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&completionData); err != nil {
		// If no body provided, that's OK - just mark as complete
		log.Printf("No completion data provided for item %s", itemID)
	}

	itemType := r.URL.Query().Get("type")
	if itemType == "" {
		http.Error(w, "Missing item type parameter", http.StatusBadRequest)
		return
	}

	completedAt := time.Now()

	switch itemType {
	case "showing":
		if err := dsh.dailyScheduleService.db.Model(&models.Booking{}).Where("id = ?", itemID).
			Updates(map[string]interface{}{"status": "completed", "completed_at": &completedAt}).Error; err != nil {
			log.Printf("Failed to complete showing %s: %v", itemID, err)
			http.Error(w, "Failed to complete showing", http.StatusInternalServerError)
			return
		}
		log.Printf("✓ Completed showing: %s", itemID)
	case "followup":
		if err := dsh.dailyScheduleService.db.Model(&services.CalendarEvent{}).Where("id = ?", itemID).
			Update("status", "completed").Error; err != nil {
			log.Printf("Failed to complete followup %s: %v", itemID, err)
			http.Error(w, "Failed to complete followup", http.StatusInternalServerError)
			return
		}
		log.Printf("✓ Completed followup: %s", itemID)
	case "task":
		http.Error(w, "Task type not yet implemented", http.StatusNotImplemented)
		return
	default:
		http.Error(w, "Invalid item type. Must be: showing, followup, or task", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"message":      fmt.Sprintf("Schedule item %s marked as completed", itemID),
		"item_id":      itemID,
		"item_type":    itemType,
		"completed_at": completedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SnoozeScheduleItem postpones a schedule item
func (dsh *DailyScheduleHandlers) SnoozeScheduleItem(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	itemID := vars["id"]

	// Parse request body for snooze details
	var snoozeData struct {
		Duration string `json:"duration"` // "1h", "2h", "1d", etc.
		Reason   string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&snoozeData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse duration
	duration, err := time.ParseDuration(snoozeData.Duration)
	if err != nil {
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	itemType := r.URL.Query().Get("type")
	if itemType == "" {
		http.Error(w, "Missing item type parameter", http.StatusBadRequest)
		return
	}

	newTime := time.Now().Add(duration)

	switch itemType {
	case "showing":
		if err := dsh.dailyScheduleService.db.Model(&models.Booking{}).Where("id = ?", itemID).
			Update("showing_date", newTime).Error; err != nil {
			log.Printf("Failed to snooze showing %s: %v", itemID, err)
			http.Error(w, "Failed to snooze showing", http.StatusInternalServerError)
			return
		}
		log.Printf("⏰ Snoozed showing %s until %s", itemID, newTime.Format("3:04 PM"))
	case "followup":
		if err := dsh.dailyScheduleService.db.Model(&services.CalendarEvent{}).Where("id = ?", itemID).
			Update("start_time", newTime).Error; err != nil {
			log.Printf("Failed to snooze followup %s: %v", itemID, err)
			http.Error(w, "Failed to snooze followup", http.StatusInternalServerError)
			return
		}
		log.Printf("⏰ Snoozed followup %s until %s", itemID, newTime.Format("3:04 PM"))
	case "task":
		http.Error(w, "Task type not yet implemented", http.StatusNotImplemented)
		return
	default:
		http.Error(w, "Invalid item type. Must be: showing, followup, or task", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Schedule item %s snoozed until %s", itemID, newTime.Format("3:04 PM")),
		"item_id":       itemID,
		"item_type":     itemType,
		"snoozed_until": newTime,
		"reason":        snoozeData.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods for mobile optimization

func (dsh *DailyScheduleHandlers) getNextAppointment(schedule []services.ScheduleItem) interface{} {
	now := time.Now()

	for _, item := range schedule {
		// Check if item starts in the future
		if item.StartTime.After(now) {
			return map[string]interface{}{
				"time":   item.StartTime.Format("3:04 PM"),
				"title":  item.Title,
				"type":   item.Type,
				"status": item.Status,
			}
		}
	}

	return nil
}

func (dsh *DailyScheduleHandlers) getTopPriorityTodo(todos []services.TodoItem) interface{} {
	if len(todos) == 0 {
		return nil
	}

	// Return the first todo (already sorted by priority)
	todo := todos[0]
	isOverdue := false
	if todo.DueDate != nil && todo.DueDate.Before(time.Now()) {
		isOverdue = true
	}
	return map[string]interface{}{
		"title":       todo.Title,
		"description": todo.Description,
		"priority":    todo.Priority,
		"category":    todo.Category,
		"overdue":     isOverdue,
	}
}

func (dsh *DailyScheduleHandlers) getNextItems(schedule []services.ScheduleItem, limit int) []map[string]interface{} {
	now := time.Now()

	var nextItems []map[string]interface{}
	count := 0

	for _, item := range schedule {
		if count >= limit {
			break
		}

		// Check if item starts in the future
		if item.StartTime.After(now) {
			nextItems = append(nextItems, map[string]interface{}{
				"time":        item.StartTime.Format("3:04 PM"),
				"type":        item.Type,
				"title":       item.Title,
				"description": item.Description,
				"status":      item.Status,
				"priority":    item.Priority,
				"id":          item.ID,
			})
			count++
		}
	}

	return nextItems
}

func (dsh *DailyScheduleHandlers) getTopTodos(todos []services.TodoItem, limit int) []map[string]interface{} {
	var topTodos []map[string]interface{}

	maxItems := limit
	if len(todos) < maxItems {
		maxItems = len(todos)
	}

	for i := 0; i < maxItems; i++ {
		todo := todos[i]
		isOverdue := false
		daysOverdue := 0
		if todo.DueDate != nil && todo.DueDate.Before(time.Now()) {
			isOverdue = true
			daysOverdue = int(time.Since(*todo.DueDate).Hours() / 24)
		}
		topTodos = append(topTodos, map[string]interface{}{
			"id":           todo.ID,
			"category":     todo.Category,
			"title":        todo.Title,
			"description":  todo.Description,
			"priority":     todo.Priority,
			"status":       todo.Status,
			"overdue":      isOverdue,
			"days_overdue": daysOverdue,
			"due_date":     todo.DueDate,
		})
	}

	return topTodos
}

// Additional helper methods for simplified response

func (dsh *DailyScheduleHandlers) getGreeting() string {
	hour := time.Now().Hour()
	if hour < 12 {
		return "morning"
	} else if hour < 17 {
		return "afternoon"
	}
	return "evening"
}

func (dsh *DailyScheduleHandlers) countItemsByType(schedule []services.ScheduleItem, itemType string) int {
	count := 0
	for _, item := range schedule {
		if item.Type == itemType {
			count++
		}
	}
	return count
}

func (dsh *DailyScheduleHandlers) countConfirmedShowings(schedule []services.ScheduleItem) int {
	count := 0
	for _, item := range schedule {
		if item.Type == "showing" && item.Status == "confirmed" {
			count++
		}
	}
	return count
}

func (dsh *DailyScheduleHandlers) countPendingShowings(schedule []services.ScheduleItem) int {
	count := 0
	for _, item := range schedule {
		if item.Type == "showing" && item.Status == "pending" {
			count++
		}
	}
	return count
}

func (dsh *DailyScheduleHandlers) countUrgentItems(schedule []services.ScheduleItem) int {
	count := 0
	for _, item := range schedule {
		if item.Priority == "high" || item.Priority == "urgent" {
			count++
		}
	}
	return count
}

func (dsh *DailyScheduleHandlers) countOverdueItems(schedule []services.ScheduleItem) int {
	count := 0
	now := time.Now()
	for _, item := range schedule {
		if item.EndTime.Before(now) && item.Status != "completed" {
			count++
		}
	}
	return count
}

func (dsh *DailyScheduleHandlers) getTopPriorityItem(schedule []services.ScheduleItem) interface{} {
	for _, item := range schedule {
		if item.Priority == "high" || item.Priority == "urgent" {
			return map[string]interface{}{
				"title":       item.Title,
				"description": item.Description,
				"priority":    item.Priority,
				"status":      item.Status,
			}
		}
	}
	return nil
}

func (dsh *DailyScheduleHandlers) getTopPriorityItems(schedule []services.ScheduleItem, limit int) []map[string]interface{} {
	var priorityItems []map[string]interface{}
	count := 0

	for _, item := range schedule {
		if count >= limit {
			break
		}
		if item.Priority == "high" || item.Priority == "urgent" {
			priorityItems = append(priorityItems, map[string]interface{}{
				"title":       item.Title,
				"description": item.Description,
				"priority":    item.Priority,
				"status":      item.Status,
				"type":        item.Type,
			})
			count++
		}
	}

	return priorityItems
}

func (dsh *DailyScheduleHandlers) getQuickStats(schedule []services.ScheduleItem) map[string]interface{} {
	totalItems := len(schedule)
	completedItems := 0
	pendingItems := 0

	for _, item := range schedule {
		if item.Status == "completed" {
			completedItems++
		} else if item.Status == "pending" {
			pendingItems++
		}
	}

	return map[string]interface{}{
		"total_items":     totalItems,
		"completed_items": completedItems,
		"pending_items":   pendingItems,
		"completion_rate": float64(completedItems) / float64(totalItems) * 100,
	}
}
