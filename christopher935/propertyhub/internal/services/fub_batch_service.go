package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/models"
)

// FUBOperation represents a queued FUB API operation
type FUBOperation struct {
	ID          string                 `json:"id"`
	Type        FUBOperationType       `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Priority    int                    `json:"priority"`
	ContactID   uint                   `json:"contact_id,omitempty"`
	BookingID   uint                   `json:"booking_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt time.Time              `json:"scheduled_at"`
}

// FUBOperationType defines the type of FUB operation
type FUBOperationType string

const (
	FUBCreateLead       FUBOperationType = "create_lead"
	FUBUpdateLead       FUBOperationType = "update_lead"
	FUBCreateNote       FUBOperationType = "create_note"
	FUBCreateTask       FUBOperationType = "create_task"
	FUBUpdateContact    FUBOperationType = "update_contact"
	FUBBulkSync         FUBOperationType = "bulk_sync"
	FUBWebhookProcessor FUBOperationType = "webhook_processor"
)

// FUBBatchResult represents the result of a batch operation
type FUBBatchResult struct {
	OperationID  string `json:"operation_id"`
	Success      bool   `json:"success"`
	FUBLeadID    string `json:"fub_lead_id,omitempty"`
	Error        string `json:"error,omitempty"`
	HTTPStatus   int    `json:"http_status"`
	ResponseBody string `json:"response_body,omitempty"`
}

// FUBBatchStats tracks batch processing statistics
type FUBBatchStats struct {
	TotalOperations    int64     `json:"total_operations"`
	SuccessfulOps      int64     `json:"successful_ops"`
	FailedOps          int64     `json:"failed_ops"`
	RetryOps           int64     `json:"retry_ops"`
	AverageProcessTime int64     `json:"average_process_time_ms"`
	RateLimitHits      int64     `json:"rate_limit_hits"`
	LastBatchProcessed time.Time `json:"last_batch_processed"`
}

// FUBBatchService handles batched FUB API operations for improved performance
type FUBBatchService struct {
	db           *gorm.DB
	redis        *redis.Client
	apiKey       string
	apiURL       string
	batchSize    int
	rateLimitPer time.Duration
	maxRetries   int

	// Batch processing
	operationQueue chan *FUBOperation
	batchTicker    *time.Ticker
	processingMux  sync.RWMutex
	isProcessing   bool
	stats          *FUBBatchStats

	// HTTP client with timeouts
	httpClient *http.Client
}

// NewFUBBatchService creates a new FUB batch processing service
func NewFUBBatchService(db *gorm.DB, redis *redis.Client) *FUBBatchService {
	apiKey := os.Getenv("FUB_API_TOKEN")
	if apiKey == "" {
		log.Println("⚠️ FUB_API_TOKEN not configured - FUB batching disabled")
	}

	service := &FUBBatchService{
		db:             db,
		redis:          redis,
		apiKey:         apiKey,
		apiURL:         "https://api.followupboss.com/v1",
		batchSize:      10,          // Process in batches of 10 operations
		rateLimitPer:   time.Second, // Rate limit: 1 request per second
		maxRetries:     3,
		operationQueue: make(chan *FUBOperation, 1000), // Buffer up to 1000 operations
		stats:          &FUBBatchStats{},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Start batch processing
	if apiKey != "" {
		service.startBatchProcessor()
		log.Println("🔄 FUB batch processing service started")
	}

	return service
}

// QueueCreateLead queues a lead creation operation
func (fbs *FUBBatchService) QueueCreateLead(contactID uint, leadData map[string]interface{}) error {
	if fbs.apiKey == "" {
		log.Println("FUB API not configured, skipping lead creation")
		return nil
	}

	operation := &FUBOperation{
		ID:          fmt.Sprintf("create_lead_%d_%d", contactID, time.Now().Unix()),
		Type:        FUBCreateLead,
		Payload:     leadData,
		ContactID:   contactID,
		MaxRetries:  fbs.maxRetries,
		Priority:    1, // High priority for new leads
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
	}

	select {
	case fbs.operationQueue <- operation:
		log.Printf("🔄 Queued FUB lead creation for contact %d", contactID)
		return nil
	default:
		return fmt.Errorf("FUB operation queue is full")
	}
}

// QueueUpdateLead queues a lead update operation
func (fbs *FUBBatchService) QueueUpdateLead(fubLeadID string, updateData map[string]interface{}) error {
	if fbs.apiKey == "" {
		return nil
	}

	operation := &FUBOperation{
		ID:          fmt.Sprintf("update_lead_%s_%d", fubLeadID, time.Now().Unix()),
		Type:        FUBUpdateLead,
		Payload:     updateData,
		MaxRetries:  fbs.maxRetries,
		Priority:    2, // Medium priority for updates
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
	}

	select {
	case fbs.operationQueue <- operation:
		log.Printf("🔄 Queued FUB lead update for %s", fubLeadID)
		return nil
	default:
		return fmt.Errorf("FUB operation queue is full")
	}
}

// QueueCreateNote queues a note creation operation
func (fbs *FUBBatchService) QueueCreateNote(fubLeadID string, noteData map[string]interface{}) error {
	if fbs.apiKey == "" {
		return nil
	}

	operation := &FUBOperation{
		ID:          fmt.Sprintf("create_note_%s_%d", fubLeadID, time.Now().Unix()),
		Type:        FUBCreateNote,
		Payload:     noteData,
		MaxRetries:  fbs.maxRetries,
		Priority:    3, // Lower priority for notes
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
	}

	select {
	case fbs.operationQueue <- operation:
		log.Printf("🔄 Queued FUB note creation for %s", fubLeadID)
		return nil
	default:
		return fmt.Errorf("FUB operation queue is full")
	}
}

// QueueBulkSync queues a bulk synchronization operation
func (fbs *FUBBatchService) QueueBulkSync(contactIDs []uint) error {
	if fbs.apiKey == "" {
		return nil
	}

	// Split large bulk operations into manageable chunks
	chunkSize := 50
	for i := 0; i < len(contactIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(contactIDs) {
			end = len(contactIDs)
		}

		chunk := contactIDs[i:end]
		operation := &FUBOperation{
			ID:          fmt.Sprintf("bulk_sync_%d_%d", i, time.Now().Unix()),
			Type:        FUBBulkSync,
			Payload:     map[string]interface{}{"contact_ids": chunk},
			MaxRetries:  2, // Fewer retries for bulk operations
			Priority:    4, // Lowest priority for bulk operations
			CreatedAt:   time.Now(),
			ScheduledAt: time.Now().Add(time.Duration(i/chunkSize) * 5 * time.Second), // Spread out bulk operations
		}

		select {
		case fbs.operationQueue <- operation:
			log.Printf("🔄 Queued FUB bulk sync for %d contacts (chunk %d)", len(chunk), i/chunkSize+1)
		default:
			return fmt.Errorf("FUB operation queue is full")
		}
	}

	return nil
}

// startBatchProcessor starts the background batch processing
func (fbs *FUBBatchService) startBatchProcessor() {
	// Process operations immediately as they come in, with rate limiting
	go func() {
		ticker := time.NewTicker(fbs.rateLimitPer)
		defer ticker.Stop()

		batch := make([]*FUBOperation, 0, fbs.batchSize)

		for {
			select {
			case operation := <-fbs.operationQueue:
				// Check if operation should be processed now
				if time.Now().Before(operation.ScheduledAt) {
					// Re-queue for later processing
					time.AfterFunc(operation.ScheduledAt.Sub(time.Now()), func() {
						select {
						case fbs.operationQueue <- operation:
						default:
							log.Printf("⚠️ Failed to re-queue delayed operation %s", operation.ID)
						}
					})
					continue
				}

				batch = append(batch, operation)

				// Process batch when it reaches the desired size or after a timeout
				if len(batch) >= fbs.batchSize {
					fbs.processBatch(batch)
					batch = make([]*FUBOperation, 0, fbs.batchSize)
				}

			case <-ticker.C:
				// Process any remaining operations in the batch
				if len(batch) > 0 {
					fbs.processBatch(batch)
					batch = make([]*FUBOperation, 0, fbs.batchSize)
				}
			}
		}
	}()

	// Periodic bulk sync processor
	go func() {
		bulkTicker := time.NewTicker(30 * time.Minute) // Run bulk sync every 30 minutes
		defer bulkTicker.Stop()

		for range bulkTicker.C {
			fbs.processScheduledBulkSync()
		}
	}()
}

// processBatch processes a batch of FUB operations
func (fbs *FUBBatchService) processBatch(batch []*FUBOperation) {
	fbs.processingMux.Lock()
	defer fbs.processingMux.Unlock()

	if fbs.isProcessing {
		// Already processing, re-queue operations
		for _, op := range batch {
			select {
			case fbs.operationQueue <- op:
			default:
				log.Printf("⚠️ Failed to re-queue operation %s", op.ID)
			}
		}
		return
	}

	fbs.isProcessing = true
	defer func() {
		fbs.isProcessing = false
	}()

	startTime := time.Now()
	results := make([]*FUBBatchResult, 0, len(batch))

	// Sort batch by priority (lower number = higher priority)
	fbs.sortBatchByPriority(batch)

	log.Printf("🔄 Processing FUB batch of %d operations", len(batch))

	for _, operation := range batch {
		result := fbs.processOperation(operation)
		results = append(results, result)

		// Update stats
		fbs.stats.TotalOperations++
		if result.Success {
			fbs.stats.SuccessfulOps++
		} else {
			fbs.stats.FailedOps++

			// Retry failed operations
			if operation.RetryCount < operation.MaxRetries {
				operation.RetryCount++
				operation.ScheduledAt = time.Now().Add(time.Duration(operation.RetryCount) * 30 * time.Second)

				select {
				case fbs.operationQueue <- operation:
					fbs.stats.RetryOps++
					log.Printf("🔄 Retrying FUB operation %s (attempt %d/%d)",
						operation.ID, operation.RetryCount, operation.MaxRetries)
				default:
					log.Printf("⚠️ Failed to queue retry for operation %s", operation.ID)
				}
			}
		}

		// Rate limiting
		time.Sleep(fbs.rateLimitPer)
	}

	processingTime := time.Since(startTime)
	fbs.stats.AverageProcessTime = processingTime.Milliseconds() / int64(len(batch))
	fbs.stats.LastBatchProcessed = time.Now()

	log.Printf("✅ FUB batch processed: %d operations in %v (avg: %dms per op)",
		len(batch), processingTime, fbs.stats.AverageProcessTime)

	// Cache results for debugging
	fbs.cacheResults(results)
}

// processOperation processes a single FUB operation
func (fbs *FUBBatchService) processOperation(operation *FUBOperation) *FUBBatchResult {
	switch operation.Type {
	case FUBCreateLead:
		return fbs.processCreateLead(operation)
	case FUBUpdateLead:
		return fbs.processUpdateLead(operation)
	case FUBCreateNote:
		return fbs.processCreateNote(operation)
	case FUBBulkSync:
		return fbs.processBulkSync(operation)
	default:
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("Unknown operation type: %s", operation.Type),
		}
	}
}

// processCreateLead handles lead creation via FUB API
func (fbs *FUBBatchService) processCreateLead(operation *FUBOperation) *FUBBatchResult {
	jsonPayload, err := json.Marshal(operation.Payload)
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("JSON marshal error: %v", err),
		}
	}

	req, err := http.NewRequest("POST", fbs.apiURL+"/people", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("Request creation error: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+fbs.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := fbs.httpClient.Do(req)
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("HTTP request error: %v", err),
		}
	}
	defer resp.Body.Close()

	result := &FUBBatchResult{
		OperationID: operation.ID,
		HTTPStatus:  resp.StatusCode,
		Success:     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	// Handle rate limiting
	if resp.StatusCode == 429 {
		fbs.stats.RateLimitHits++
		result.Error = "Rate limited by FUB API"
		return result
	}

	if result.Success {
		var fubResponse struct {
			ID string `json:"id"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&fubResponse); err == nil {
			result.FUBLeadID = fubResponse.ID

			// Update contact record with FUB lead ID
			if operation.ContactID > 0 {
				fbs.db.Model(&models.Contact{}).Where("id = ?", operation.ContactID).Updates(map[string]interface{}{
					"fub_lead_id": fubResponse.ID,
					"fub_synced":  true,
				})
				log.Printf("✅ Contact %d synced with FUB lead %s", operation.ContactID, fubResponse.ID)
			}
		} else {
			result.Error = fmt.Sprintf("Response parsing error: %v", err)
		}
	} else {
		bodyBytes, _ := json.Marshal(resp.Body)
		result.ResponseBody = string(bodyBytes)
		result.Error = fmt.Sprintf("FUB API error: %d", resp.StatusCode)
	}

	return result
}

// processUpdateLead handles lead updates via FUB API
func (fbs *FUBBatchService) processUpdateLead(operation *FUBOperation) *FUBBatchResult {
	// Extract FUB lead ID from payload
	fubLeadID, ok := operation.Payload["fub_lead_id"].(string)
	if !ok || fubLeadID == "" {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       "Missing or invalid FUB lead ID",
		}
	}

	// Remove FUB lead ID from payload (not needed in the update)
	updatePayload := make(map[string]interface{})
	for k, v := range operation.Payload {
		if k != "fub_lead_id" {
			updatePayload[k] = v
		}
	}

	jsonPayload, err := json.Marshal(updatePayload)
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("JSON marshal error: %v", err),
		}
	}

	req, err := http.NewRequest("PATCH", fbs.apiURL+"/people/"+fubLeadID, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("Request creation error: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+fbs.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := fbs.httpClient.Do(req)
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("HTTP request error: %v", err),
		}
	}
	defer resp.Body.Close()

	result := &FUBBatchResult{
		OperationID: operation.ID,
		HTTPStatus:  resp.StatusCode,
		Success:     resp.StatusCode >= 200 && resp.StatusCode < 300,
		FUBLeadID:   fubLeadID,
	}

	if resp.StatusCode == 429 {
		fbs.stats.RateLimitHits++
		result.Error = "Rate limited by FUB API"
	} else if !result.Success {
		bodyBytes, _ := json.Marshal(resp.Body)
		result.ResponseBody = string(bodyBytes)
		result.Error = fmt.Sprintf("FUB API error: %d", resp.StatusCode)
	}

	return result
}

// processCreateNote handles note creation via FUB API
func (fbs *FUBBatchService) processCreateNote(operation *FUBOperation) *FUBBatchResult {
	fubLeadID, ok := operation.Payload["fub_lead_id"].(string)
	if !ok || fubLeadID == "" {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       "Missing or invalid FUB lead ID",
		}
	}

	// Build note payload
	notePayload := map[string]interface{}{
		"body":     operation.Payload["content"],
		"personId": fubLeadID,
	}

	jsonPayload, err := json.Marshal(notePayload)
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("JSON marshal error: %v", err),
		}
	}

	req, err := http.NewRequest("POST", fbs.apiURL+"/notes", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("Request creation error: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+fbs.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := fbs.httpClient.Do(req)
	if err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("HTTP request error: %v", err),
		}
	}
	defer resp.Body.Close()

	result := &FUBBatchResult{
		OperationID: operation.ID,
		HTTPStatus:  resp.StatusCode,
		Success:     resp.StatusCode >= 200 && resp.StatusCode < 300,
		FUBLeadID:   fubLeadID,
	}

	if resp.StatusCode == 429 {
		fbs.stats.RateLimitHits++
		result.Error = "Rate limited by FUB API"
	} else if !result.Success {
		bodyBytes, _ := json.Marshal(resp.Body)
		result.ResponseBody = string(bodyBytes)
		result.Error = fmt.Sprintf("FUB API error: %d", resp.StatusCode)
	}

	return result
}

// processBulkSync handles bulk synchronization operations
func (fbs *FUBBatchService) processBulkSync(operation *FUBOperation) *FUBBatchResult {
	contactIDs, ok := operation.Payload["contact_ids"].([]uint)
	if !ok {
		// Try to convert from []interface{}
		if contactIDsInterface, ok := operation.Payload["contact_ids"].([]interface{}); ok {
			contactIDs = make([]uint, len(contactIDsInterface))
			for i, id := range contactIDsInterface {
				if intID, ok := id.(float64); ok {
					contactIDs[i] = uint(intID)
				}
			}
		} else {
			return &FUBBatchResult{
				OperationID: operation.ID,
				Success:     false,
				Error:       "Invalid contact IDs in bulk sync operation",
			}
		}
	}

	successCount := 0
	errorCount := 0

	// Process each contact in the bulk operation
	var contacts []models.Contact
	if err := fbs.db.Where("id IN ? AND (fub_lead_id = '' OR fub_lead_id IS NULL)", contactIDs).Find(&contacts).Error; err != nil {
		return &FUBBatchResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       fmt.Sprintf("Database error: %v", err),
		}
	}

	for _, contact := range contacts {
		// Create FUB payload for each contact
		leadData := map[string]interface{}{
			"firstName": contact.Name,
			"phones": []map[string]interface{}{
				{"number": contact.Phone, "type": "Mobile", "primary": true},
			},
			"source": contact.Source,
			"tags":   []string{"bulk-sync", "website-lead"},
		}

		if contact.Email != "" {
			leadData["emails"] = []map[string]interface{}{
				{"address": contact.Email, "type": "Personal", "primary": true},
			}
		}

		// Create individual operation for this contact
		contactOp := &FUBOperation{
			ID:          fmt.Sprintf("bulk_contact_%d_%d", contact.ID, time.Now().Unix()),
			Type:        FUBCreateLead,
			Payload:     leadData,
			ContactID:   contact.ID,
			MaxRetries:  1, // Limited retries for bulk operations
			Priority:    5, // Lowest priority
			CreatedAt:   time.Now(),
			ScheduledAt: time.Now(),
		}

		result := fbs.processCreateLead(contactOp)
		if result.Success {
			successCount++
		} else {
			errorCount++
			log.Printf("⚠️ Bulk sync failed for contact %d: %s", contact.ID, result.Error)
		}

		// Rate limiting within bulk operations
		time.Sleep(fbs.rateLimitPer)
	}

	return &FUBBatchResult{
		OperationID: operation.ID,
		Success:     errorCount == 0,
		Error:       fmt.Sprintf("Processed %d contacts: %d successful, %d failed", len(contacts), successCount, errorCount),
	}
}

// processScheduledBulkSync processes unsynced contacts
func (fbs *FUBBatchService) processScheduledBulkSync() {
	if fbs.apiKey == "" {
		return
	}

	var unsyncedContactIDs []uint
	if err := fbs.db.Model(&models.Contact{}).
		Where("fub_synced = false AND (fub_lead_id = '' OR fub_lead_id IS NULL)").
		Limit(100). // Limit bulk sync to prevent overwhelming
		Pluck("id", &unsyncedContactIDs).Error; err != nil {
		log.Printf("⚠️ Failed to fetch unsynced contacts: %v", err)
		return
	}

	if len(unsyncedContactIDs) > 0 {
		log.Printf("🔄 Starting scheduled bulk sync for %d unsynced contacts", len(unsyncedContactIDs))
		fbs.QueueBulkSync(unsyncedContactIDs)
	}
}

// sortBatchByPriority sorts operations by priority (lower number = higher priority)
func (fbs *FUBBatchService) sortBatchByPriority(batch []*FUBOperation) {
	for i := 0; i < len(batch)-1; i++ {
		for j := i + 1; j < len(batch); j++ {
			if batch[j].Priority < batch[i].Priority {
				batch[i], batch[j] = batch[j], batch[i]
			}
		}
	}
}

// cacheResults caches batch results for debugging and monitoring
func (fbs *FUBBatchService) cacheResults(results []*FUBBatchResult) {
	if fbs.redis == nil {
		return
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("fub:batch_results:%d", time.Now().Unix())

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		log.Printf("⚠️ Failed to marshal batch results: %v", err)
		return
	}

	// Cache results for 1 hour
	fbs.redis.SetEx(ctx, cacheKey, resultsJSON, time.Hour).Err()
}

// GetBatchStats returns current batch processing statistics
func (fbs *FUBBatchService) GetBatchStats() *FUBBatchStats {
	return &FUBBatchStats{
		TotalOperations:    fbs.stats.TotalOperations,
		SuccessfulOps:      fbs.stats.SuccessfulOps,
		FailedOps:          fbs.stats.FailedOps,
		RetryOps:           fbs.stats.RetryOps,
		AverageProcessTime: fbs.stats.AverageProcessTime,
		RateLimitHits:      fbs.stats.RateLimitHits,
		LastBatchProcessed: fbs.stats.LastBatchProcessed,
	}
}

// GetQueueSize returns the current size of the operation queue
func (fbs *FUBBatchService) GetQueueSize() int {
	return len(fbs.operationQueue)
}

// Close gracefully shuts down the batch service
func (fbs *FUBBatchService) Close() error {
	log.Println("🛑 Shutting down FUB batch service...")

	// Process remaining operations in queue
	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-timeout:
			log.Printf("⚠️ FUB batch service shutdown timeout - %d operations remain", len(fbs.operationQueue))
			return nil
		default:
			if len(fbs.operationQueue) == 0 {
				log.Println("✅ FUB batch service shutdown complete")
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
