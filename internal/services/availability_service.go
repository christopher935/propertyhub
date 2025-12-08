package services

import (
	"fmt"
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// AvailabilityService manages property availability and blackout dates
type AvailabilityService struct {
	db *gorm.DB
}

// BlackoutDate represents a date range when bookings are not allowed
type BlackoutDate struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	PropertyID *uint     `json:"property_id,omitempty"` // nil for global blackouts
	MLSId      string    `json:"mls_id,omitempty"`      // specific property
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Reason     string    `json:"reason"`
	IsGlobal   bool      `json:"is_global"` // affects all properties

	// Enhanced layered blackout support
	BlackoutType  string `json:"blackout_type"`            // "one_time", "recurring_weekly", "recurring_monthly", "vacation", "maintenance"
	RecurringRule string `json:"recurring_rule,omitempty"` // "SUNDAY", "TUESDAY", "FIRST_MONDAY", etc.
	IsRecurring   bool   `json:"is_recurring"`             // true for recurring patterns
	Priority      int    `json:"priority"`                 // 1=highest (vacation), 2=medium (property rules), 3=low (preferences)

	// Vacation/personal blackouts
	IsVacation    bool   `json:"is_vacation"`              // personal time off
	VacationOwner string `json:"vacation_owner,omitempty"` // who is on vacation

	// Property-specific rules
	DaysOfWeek      string `json:"days_of_week,omitempty"`     // JSON array: ["SUNDAY", "TUESDAY"]
	TimeRestriction string `json:"time_restriction,omitempty"` // "mornings", "afternoons", "evenings"

	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AvailabilityCheck represents the result of an availability check
type AvailabilityCheck struct {
	IsAvailable      bool           `json:"is_available"`
	BlockingReasons  []string       `json:"blocking_reasons"`
	BlackoutDates    []BlackoutDate `json:"blackout_dates,omitempty"`
	AlternativeSlots []time.Time    `json:"alternative_slots,omitempty"`
	Warnings         []string       `json:"warnings,omitempty"`
}

// NewAvailabilityService creates a new availability service
func NewAvailabilityService(db *gorm.DB) *AvailabilityService {
	service := &AvailabilityService{db: db}

	// Auto-migrate blackout dates table
	if err := db.AutoMigrate(&BlackoutDate{}); err != nil {
		log.Printf("Failed to migrate BlackoutDate table: %v", err)
	}

	return service
}

// CheckAvailability checks if a property is available for booking on a specific date
func (s *AvailabilityService) CheckAvailability(mlsID string, requestedDate time.Time) (*AvailabilityCheck, error) {
	check := &AvailabilityCheck{
		IsAvailable:     true,
		BlockingReasons: []string{},
		BlackoutDates:   []BlackoutDate{},
		Warnings:        []string{},
	}

	// Layer 1: Check global blackout dates (highest priority)
	globalBlackouts, err := s.checkGlobalBlackouts(requestedDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check global blackouts: %v", err)
	}

	if len(globalBlackouts) > 0 {
		check.IsAvailable = false
		for _, blackout := range globalBlackouts {
			check.BlockingReasons = append(check.BlockingReasons,
				fmt.Sprintf("Global blackout: %s", blackout.Reason))
			check.BlackoutDates = append(check.BlackoutDates, blackout)
		}
	}

	// Layer 2: Check property-specific blackout dates
	propertyBlackouts, err := s.checkPropertyBlackouts(mlsID, requestedDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check property blackouts: %v", err)
	}

	if len(propertyBlackouts) > 0 {
		check.IsAvailable = false
		for _, blackout := range propertyBlackouts {
			check.BlockingReasons = append(check.BlockingReasons,
				fmt.Sprintf("Property blackout: %s", blackout.Reason))
			check.BlackoutDates = append(check.BlackoutDates, blackout)
		}
	}

	// Layer 3: Check recurring patterns (weekly/monthly rules)
	recurringBlackouts, err := s.checkRecurringBlackouts(mlsID, requestedDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check recurring blackouts: %v", err)
	}

	if len(recurringBlackouts) > 0 {
		check.IsAvailable = false
		for _, blackout := range recurringBlackouts {
			check.BlockingReasons = append(check.BlockingReasons,
				fmt.Sprintf("Recurring restriction: %s", blackout.Reason))
			check.BlackoutDates = append(check.BlackoutDates, blackout)
		}
	}

	// Layer 4: Check vacation blackouts (personal time off)
	vacationBlackouts, err := s.checkVacationBlackouts(requestedDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check vacation blackouts: %v", err)
	}

	if len(vacationBlackouts) > 0 {
		check.IsAvailable = false
		for _, blackout := range vacationBlackouts {
			check.BlockingReasons = append(check.BlockingReasons,
				fmt.Sprintf("Vacation blackout: %s unavailable (%s)", blackout.VacationOwner, blackout.Reason))
			check.BlackoutDates = append(check.BlackoutDates, blackout)
		}
	}

	// Check for existing bookings on the same date (capacity management)
	var existingBookings []models.Booking
	startOfDay := time.Date(requestedDate.Year(), requestedDate.Month(), requestedDate.Day(), 0, 0, 0, 0, requestedDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err = s.db.Where("property_address LIKE ? AND showing_date >= ? AND showing_date < ? AND status IN (?)",
		"%"+mlsID+"%", startOfDay, endOfDay, []string{"confirmed", "pending"}).Find(&existingBookings).Error
	if err != nil {
		log.Printf("Warning: Failed to check existing bookings: %v", err)
		check.Warnings = append(check.Warnings, "Could not verify existing bookings")
	}

	// Check if too many bookings already exist (optional limit)
	if len(existingBookings) >= 5 { // Max 5 showings per day per property
		check.IsAvailable = false
		check.BlockingReasons = append(check.BlockingReasons,
			fmt.Sprintf("Maximum daily bookings reached (%d)", len(existingBookings)))
	} else if len(existingBookings) >= 3 {
		check.Warnings = append(check.Warnings,
			fmt.Sprintf("High booking volume for this date (%d existing)", len(existingBookings)))
	}

	// Generate alternative slots if not available
	if !check.IsAvailable {
		check.AlternativeSlots = s.generateAlternativeSlots(mlsID, requestedDate)
	}

	return check, nil
}

// CreateBlackoutDate creates a new blackout date
func (s *AvailabilityService) CreateBlackoutDate(mlsID string, startDate, endDate time.Time, reason, createdBy string, isGlobal bool) (*BlackoutDate, error) {
	blackout := &BlackoutDate{
		MLSId:     mlsID,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
		IsGlobal:  isGlobal,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(blackout).Error; err != nil {
		return nil, fmt.Errorf("failed to create blackout date: %v", err)
	}

	log.Printf("Created blackout date: %s to %s for %s (Global: %v) - %s",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), mlsID, isGlobal, reason)

	return blackout, nil
}

// GetBlackoutDates retrieves blackout dates for a property or all global blackouts
func (s *AvailabilityService) GetBlackoutDates(mlsID string) ([]BlackoutDate, error) {
	var blackouts []BlackoutDate

	query := s.db.Where("is_global = ?", true)
	if mlsID != "" {
		query = query.Or("mls_id = ?", mlsID)
	}

	err := query.Order("start_date ASC").Find(&blackouts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get blackout dates: %v", err)
	}

	return blackouts, nil
}

// RemoveBlackoutDate removes a blackout date by ID
func (s *AvailabilityService) RemoveBlackoutDate(blackoutID uint) error {
	result := s.db.Delete(&BlackoutDate{}, blackoutID)
	if result.Error != nil {
		return fmt.Errorf("failed to remove blackout date: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("blackout date not found")
	}

	log.Printf("Removed blackout date ID: %d", blackoutID)
	return nil
}

// ValidateBookingDate validates if a booking can be made on a specific date
func (s *AvailabilityService) ValidateBookingDate(mlsID string, requestedDate time.Time) error {
	check, err := s.CheckAvailability(mlsID, requestedDate)
	if err != nil {
		return fmt.Errorf("availability check failed: %v", err)
	}

	if !check.IsAvailable {
		return fmt.Errorf("booking not available: %s", check.BlockingReasons[0])
	}

	return nil
}

// generateAlternativeSlots suggests alternative booking dates
func (s *AvailabilityService) generateAlternativeSlots(mlsID string, requestedDate time.Time) []time.Time {
	var alternatives []time.Time

	// Check next 14 days for available slots
	for i := 1; i <= 14; i++ {
		alternativeDate := requestedDate.AddDate(0, 0, i)

		// Skip weekends for business properties (optional logic)
		if alternativeDate.Weekday() == time.Saturday || alternativeDate.Weekday() == time.Sunday {
			continue
		}

		check, err := s.CheckAvailability(mlsID, alternativeDate)
		if err != nil {
			continue
		}

		if check.IsAvailable {
			alternatives = append(alternatives, alternativeDate)
			if len(alternatives) >= 5 { // Limit to 5 suggestions
				break
			}
		}
	}

	return alternatives
}

// GetUpcomingBlackouts returns blackout dates in the next 30 days
func (s *AvailabilityService) GetUpcomingBlackouts() ([]BlackoutDate, error) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, 30)

	var blackouts []BlackoutDate
	err := s.db.Where("start_date >= ? AND start_date <= ?", now, futureDate).
		Order("start_date ASC").Find(&blackouts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming blackouts: %v", err)
	}

	return blackouts, nil
}

// CreateGlobalBlackout creates a blackout that affects all properties
func (s *AvailabilityService) CreateGlobalBlackout(startDate, endDate time.Time, reason, createdBy string) (*BlackoutDate, error) {
	return s.CreateBlackoutDate("", startDate, endDate, reason, createdBy, true)
}

// GetAvailabilityStats returns statistics about availability
func (s *AvailabilityService) GetAvailabilityStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count total blackout dates
	var totalBlackouts int64
	s.db.Model(&BlackoutDate{}).Count(&totalBlackouts)
	stats["total_blackouts"] = totalBlackouts

	// Count global blackouts
	var globalBlackouts int64
	s.db.Model(&BlackoutDate{}).Where("is_global = ?", true).Count(&globalBlackouts)
	stats["global_blackouts"] = globalBlackouts

	// Count property-specific blackouts
	stats["property_blackouts"] = totalBlackouts - globalBlackouts

	// Count active blackouts (current date falls within range)
	now := time.Now()
	var activeBlackouts int64
	s.db.Model(&BlackoutDate{}).Where("start_date <= ? AND end_date >= ?", now, now).Count(&activeBlackouts)
	stats["active_blackouts"] = activeBlackouts

	// Count upcoming blackouts (next 7 days)
	futureDate := now.AddDate(0, 0, 7)
	var upcomingBlackouts int64
	s.db.Model(&BlackoutDate{}).Where("start_date >= ? AND start_date <= ?", now, futureDate).Count(&upcomingBlackouts)
	stats["upcoming_blackouts"] = upcomingBlackouts

	return stats, nil
}

// CleanupExpiredBlackouts removes blackout dates that have passed
func (s *AvailabilityService) CleanupExpiredBlackouts() error {
	now := time.Now()
	result := s.db.Where("end_date < ?", now).Delete(&BlackoutDate{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired blackouts: %v", result.Error)
	}

	log.Printf("Cleaned up %d expired blackout dates", result.RowsAffected)
	return nil
}

// checkGlobalBlackouts checks for global blackout dates and recurring global rules
func (s *AvailabilityService) checkGlobalBlackouts(requestedDate time.Time) ([]BlackoutDate, error) {
	var blackouts []BlackoutDate

	// Check one-time global blackouts
	err := s.db.Where("is_global = ? AND start_date <= ? AND end_date >= ?",
		true, requestedDate, requestedDate).Find(&blackouts).Error
	if err != nil {
		return nil, err
	}

	// Check recurring global blackouts (e.g., "Every Sunday")
	var recurringGlobal []BlackoutDate
	err = s.db.Where("is_global = ? AND is_recurring = ?", true, true).Find(&recurringGlobal).Error
	if err != nil {
		return nil, err
	}

	for _, recurring := range recurringGlobal {
		if s.matchesRecurringRule(recurring.RecurringRule, requestedDate) {
			blackouts = append(blackouts, recurring)
		}
	}

	return blackouts, nil
}

// checkPropertyBlackouts checks for property-specific blackout dates
func (s *AvailabilityService) checkPropertyBlackouts(mlsID string, requestedDate time.Time) ([]BlackoutDate, error) {
	var blackouts []BlackoutDate

	// Check one-time property blackouts
	err := s.db.Where("mls_id = ? AND start_date <= ? AND end_date >= ? AND is_recurring = ?",
		mlsID, requestedDate, requestedDate, false).Find(&blackouts).Error
	if err != nil {
		return nil, err
	}

	return blackouts, nil
}

// checkRecurringBlackouts checks for recurring patterns (weekly/monthly)
func (s *AvailabilityService) checkRecurringBlackouts(mlsID string, requestedDate time.Time) ([]BlackoutDate, error) {
	var blackouts []BlackoutDate
	var matchingBlackouts []BlackoutDate

	// Get all recurring rules for this property
	err := s.db.Where("mls_id = ? AND is_recurring = ?", mlsID, true).Find(&blackouts).Error
	if err != nil {
		return nil, err
	}

	// Check each recurring rule against the requested date
	for _, blackout := range blackouts {
		if s.matchesRecurringRule(blackout.RecurringRule, requestedDate) {
			matchingBlackouts = append(matchingBlackouts, blackout)
		}
	}

	return matchingBlackouts, nil
}

// checkVacationBlackouts checks for vacation/personal time off
func (s *AvailabilityService) checkVacationBlackouts(requestedDate time.Time) ([]BlackoutDate, error) {
	var blackouts []BlackoutDate

	err := s.db.Where("is_vacation = ? AND start_date <= ? AND end_date >= ?",
		true, requestedDate, requestedDate).Find(&blackouts).Error
	if err != nil {
		return nil, err
	}

	return blackouts, nil
}

// matchesRecurringRule checks if a date matches a recurring rule
func (s *AvailabilityService) matchesRecurringRule(rule string, date time.Time) bool {
	switch rule {
	case "SUNDAY":
		return date.Weekday() == time.Sunday
	case "MONDAY":
		return date.Weekday() == time.Monday
	case "TUESDAY":
		return date.Weekday() == time.Tuesday
	case "WEDNESDAY":
		return date.Weekday() == time.Wednesday
	case "THURSDAY":
		return date.Weekday() == time.Thursday
	case "FRIDAY":
		return date.Weekday() == time.Friday
	case "SATURDAY":
		return date.Weekday() == time.Saturday
	case "WEEKENDS":
		return date.Weekday() == time.Saturday || date.Weekday() == time.Sunday
	case "WEEKDAYS":
		return date.Weekday() >= time.Monday && date.Weekday() <= time.Friday
	case "FIRST_MONDAY":
		return date.Weekday() == time.Monday && date.Day() <= 7
	case "LAST_FRIDAY":
		// Check if this is the last Friday of the month
		nextWeek := date.AddDate(0, 0, 7)
		return date.Weekday() == time.Friday && nextWeek.Month() != date.Month()
	default:
		return false
	}
}

// CreateRecurringBlackout creates a recurring blackout rule
func (s *AvailabilityService) CreateRecurringBlackout(mlsID, recurringRule, reason, createdBy string, isGlobal bool, priority int) (*BlackoutDate, error) {
	blackout := &BlackoutDate{
		MLSId:         mlsID,
		Reason:        reason,
		IsGlobal:      isGlobal,
		IsRecurring:   true,
		RecurringRule: recurringRule,
		BlackoutType:  "recurring_weekly",
		Priority:      priority,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(blackout).Error; err != nil {
		return nil, fmt.Errorf("failed to create recurring blackout: %v", err)
	}

	log.Printf("Created recurring blackout: %s for %s (Global: %v) - %s",
		recurringRule, mlsID, isGlobal, reason)

	return blackout, nil
}

// CreateVacationBlackout creates a vacation blackout
func (s *AvailabilityService) CreateVacationBlackout(startDate, endDate time.Time, vacationOwner, reason, createdBy string) (*BlackoutDate, error) {
	blackout := &BlackoutDate{
		StartDate:     startDate,
		EndDate:       endDate,
		Reason:        reason,
		IsGlobal:      true, // Vacation affects all properties
		IsVacation:    true,
		VacationOwner: vacationOwner,
		BlackoutType:  "vacation",
		Priority:      1, // Highest priority
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(blackout).Error; err != nil {
		return nil, fmt.Errorf("failed to create vacation blackout: %v", err)
	}

	log.Printf("Created vacation blackout: %s to %s for %s - %s",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), vacationOwner, reason)

	return blackout, nil
}

// GetLayeredBlackouts returns all blackouts organized by layer/priority
func (s *AvailabilityService) GetLayeredBlackouts(mlsID string) (map[string][]BlackoutDate, error) {
	layers := make(map[string][]BlackoutDate)

	// Layer 1: Global one-time blackouts
	var globalOneTime []BlackoutDate
	err := s.db.Where("is_global = ? AND is_recurring = ? AND is_vacation = ?",
		true, false, false).Find(&globalOneTime).Error
	if err != nil {
		return nil, err
	}
	layers["global_one_time"] = globalOneTime

	// Layer 2: Global recurring blackouts (e.g., "Every Sunday")
	var globalRecurring []BlackoutDate
	err = s.db.Where("is_global = ? AND is_recurring = ?", true, true).Find(&globalRecurring).Error
	if err != nil {
		return nil, err
	}
	layers["global_recurring"] = globalRecurring

	// Layer 3: Vacation blackouts
	var vacation []BlackoutDate
	err = s.db.Where("is_vacation = ?", true).Find(&vacation).Error
	if err != nil {
		return nil, err
	}
	layers["vacation"] = vacation

	// Layer 4: Property-specific one-time
	var propertyOneTime []BlackoutDate
	if mlsID != "" {
		err = s.db.Where("mls_id = ? AND is_recurring = ?", mlsID, false).Find(&propertyOneTime).Error
		if err != nil {
			return nil, err
		}
	}
	layers["property_one_time"] = propertyOneTime

	// Layer 5: Property-specific recurring
	var propertyRecurring []BlackoutDate
	if mlsID != "" {
		err = s.db.Where("mls_id = ? AND is_recurring = ?", mlsID, true).Find(&propertyRecurring).Error
		if err != nil {
			return nil, err
		}
	}
	layers["property_recurring"] = propertyRecurring

	return layers, nil
}

// GetBlackoutSummary returns a human-readable summary of all blackout rules
func (s *AvailabilityService) GetBlackoutSummary(mlsID string) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	layers, err := s.GetLayeredBlackouts(mlsID)
	if err != nil {
		return nil, err
	}

	// Count by type
	summary["global_rules"] = len(layers["global_recurring"])
	summary["vacation_periods"] = len(layers["vacation"])
	summary["property_rules"] = len(layers["property_recurring"])
	summary["one_time_blackouts"] = len(layers["global_one_time"]) + len(layers["property_one_time"])

	// List active rules
	var activeRules []string
	for _, blackout := range layers["global_recurring"] {
		activeRules = append(activeRules, fmt.Sprintf("Global: %s (%s)", blackout.RecurringRule, blackout.Reason))
	}
	for _, blackout := range layers["property_recurring"] {
		activeRules = append(activeRules, fmt.Sprintf("Property: %s (%s)", blackout.RecurringRule, blackout.Reason))
	}
	summary["active_rules"] = activeRules

	// Upcoming vacation periods
	var upcomingVacations []string
	now := time.Now()
	for _, vacation := range layers["vacation"] {
		if vacation.StartDate.After(now) {
			upcomingVacations = append(upcomingVacations,
				fmt.Sprintf("%s: %s to %s", vacation.VacationOwner,
					vacation.StartDate.Format("Jan 2"), vacation.EndDate.Format("Jan 2")))
		}
	}
	summary["upcoming_vacations"] = upcomingVacations

	return summary, nil
}
