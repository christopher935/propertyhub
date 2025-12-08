package services

import (
	"log"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// EventProcessor monitors event tables and triggers orchestration
// This is the "automation loop" that keeps the symphony playing
type EventProcessor struct {
	db           *gorm.DB
	orchestrator *EventCampaignOrchestrator
	stopChan     chan bool
	running      bool
}

// NewEventProcessor creates the event processing service
func NewEventProcessor(db *gorm.DB, orchestrator *EventCampaignOrchestrator) *EventProcessor {
	return &EventProcessor{
		db:           db,
		orchestrator: orchestrator,
		stopChan:     make(chan bool),
		running:      false,
	}
}

// Start begins processing events in background
func (ep *EventProcessor) Start() {
	if ep.running {
		log.Println("⚠️  Event processor already running")
		return
	}

	ep.running = true
	log.Println("🚀 Event Processor started - monitoring for automation triggers")

	// Start background goroutine for each event type
	go ep.processPriceChangeEvents()
	go ep.processBehavioralEvents()
	go ep.processBookingEvents()
}

// Stop halts event processing
func (ep *EventProcessor) Stop() {
	if !ep.running {
		return
	}

	log.Println("🛑 Stopping Event Processor...")
	ep.running = false
	close(ep.stopChan)
}

// processPriceChangeEvents monitors price_change_events table
func (ep *EventProcessor) processPriceChangeEvents() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ep.stopChan:
			return
		case <-ticker.C:
			ep.processUnprocessedPriceChanges()
		}
	}
}

func (ep *EventProcessor) processUnprocessedPriceChanges() {
	// Find price change events that haven't been processed
	var events []models.PriceChangeEvent
	err := ep.db.Where("processed_at IS NULL AND campaign_sent = ?", false).
		Order("created_at ASC").
		Limit(50). // Process in batches
		Find(&events).Error

	if err != nil {
		log.Printf("❌ Failed to fetch price change events: %v", err)
		return
	}

	if len(events) == 0 {
		return // No events to process
	}

	log.Printf("📊 Processing %d price change events", len(events))

	for _, event := range events {
		// Build event data for orchestrator
		eventData := map[string]interface{}{
			"property_id":      float64(event.PropertyID),
			"property_address": event.PropertyAddress,
			"old_price":        event.OldPrice,
			"new_price":        event.NewPrice,
			"change_amount":    event.ChangeAmount,
			"change_percent":   event.ChangePercent,
		}

		// Only trigger if it's a price REDUCTION (people love discounts!)
		if event.ChangeAmount < 0 {
			// Trigger orchestration
			if err := ep.orchestrator.ProcessEvent("price_changed", eventData); err != nil {
				log.Printf("❌ Failed to orchestrate price change event %d: %v", event.ID, err)
				continue
			}
		}

		// Mark as processed
		now := time.Now()
		ep.db.Model(&models.PriceChangeEvent{}).
			Where("id = ?", event.ID).
			Updates(map[string]interface{}{
				"processed_at":  &now,
				"campaign_sent": true,
			})

		log.Printf("✅ Processed price change event %d: %s", event.ID, event.PropertyAddress)
	}
}

// processBehavioralEvents monitors behavioral_events for hot lead signals
func (ep *EventProcessor) processBehavioralEvents() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ep.stopChan:
			return
		case <-ticker.C:
			ep.detectHotLeadBehavior()
		}
	}
}

func (ep *EventProcessor) detectHotLeadBehavior() {
	// Find leads with high activity in last hour (3+ property views)
	var leadIDs []int64
	err := ep.db.Raw(`
		SELECT lead_id, COUNT(*) as view_count
		FROM behavioral_events
		WHERE event_type = 'viewed'
		AND created_at > NOW() - INTERVAL '1 hour'
		AND lead_id IS NOT NULL
		GROUP BY lead_id
		HAVING COUNT(*) >= 3
	`).Scan(&leadIDs).Error

	if err != nil {
		return
	}

	for _, leadID := range leadIDs {
		// Check if we've already notified about this lead recently
		var recentNotification int64
		ep.db.Raw(`
			SELECT COUNT(*) 
			FROM campaign_execution_logs 
			WHERE event_type = 'lead_scored_hot' 
			AND event_data->>'lead_id' = ? 
			AND executed_at > NOW() - INTERVAL '48 hours'
		`, leadID).Scan(&recentNotification)

		if recentNotification > 0 {
			continue // Already notified recently
		}

		// Get behavioral score if available
		var score models.BehavioralScore
		if err := ep.db.Where("lead_id = ?", leadID).First(&score).Error; err != nil {
			continue
		}

		// Only trigger if score is high enough (70+)
		if score.CompositeScore >= 70 {
			eventData := map[string]interface{}{
				"lead_id":          float64(leadID),
				"overall_score":    float64(score.CompositeScore),
				"urgency_score":    score.UrgencyScore,
				"engagement_score": score.EngagementScore,
			}

			ep.orchestrator.ProcessEvent("lead_scored_hot", eventData)
			log.Printf("🔥 Hot lead detected: Lead ID %d (Score: %d)", leadID, score.CompositeScore)
		}
	}
}

// processBookingEvents monitors bookings table for showing follow-ups
func (ep *EventProcessor) processBookingEvents() {
	ticker := time.NewTicker(10 * time.Minute) // Check every 10 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ep.stopChan:
			return
		case <-ticker.C:
			ep.triggerShowingFollowUps()
		}
	}
}

func (ep *EventProcessor) triggerShowingFollowUps() {
	// Find bookings that completed 2 hours ago and haven't had follow-up
	var bookings []struct {
		ID              uint
		ContactID       string
		PropertyAddress string
		ShowingDate     time.Time
	}

	err := ep.db.Raw(`
		SELECT id, contact_id, property_address, showing_date
		FROM bookings
		WHERE showing_date < NOW() - INTERVAL '2 hours'
		AND showing_date > NOW() - INTERVAL '4 hours'
		AND status = 'confirmed'
		AND contact_id IS NOT NULL
		AND id NOT IN (
			SELECT CAST(event_data->>'booking_id' AS INTEGER)
			FROM campaign_execution_logs
			WHERE event_type = 'showing_completed'
			AND event_data->>'booking_id' IS NOT NULL
		)
		LIMIT 20
	`).Scan(&bookings).Error

	if err != nil {
		return
	}

	for _, booking := range bookings {
		eventData := map[string]interface{}{
			"booking_id":       float64(booking.ID),
			"contact_id":       booking.ContactID,
			"property_address": booking.PropertyAddress,
			"showing_date":     booking.ShowingDate.Format("2006-01-02 15:04"),
		}

		ep.orchestrator.ProcessEvent("showing_completed", eventData)
		log.Printf("📧 Triggered showing follow-up for booking %d", booking.ID)
	}
}

// ProcessEventImmediately allows manual triggering of events (for testing or API calls)
func (ep *EventProcessor) ProcessEventImmediately(eventType string, eventData map[string]interface{}) error {
	return ep.orchestrator.ProcessEvent(eventType, eventData)
}
