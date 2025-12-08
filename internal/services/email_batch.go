package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"chrisgross-ctrl-project/internal/utils"
	"github.com/redis/go-redis/v9"
)

// EmailBatchService handles batched email processing
type EmailBatchService struct {
	redis         *redis.Client
	batchSize     int
	flushInterval time.Duration
	maxRetries    int

	// Queues
	emailQueue  chan EmailJob
	batchQueue  chan EmailBatch
	resultQueue chan EmailResult

	// Workers
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc

	// Statistics
	mutex               sync.RWMutex
	totalEmails         int64
	sentEmails          int64
	failedEmails        int64
	batchesSent         int64
	averageDeliveryTime time.Duration

	// Configuration
	rateLimitPerSecond int
	rateLimitBurst     int

	// Email provider settings
	smtpConfig SMTPConfig
}

// EmailJob represents a single email to be sent
type EmailJob struct {
	ID          string                 `json:"id"`
	To          []string               `json:"to"`
	CC          []string               `json:"cc,omitempty"`
	BCC         []string               `json:"bcc,omitempty"`
	Subject     string                 `json:"subject"`
	Body        string                 `json:"body"`
	HTMLBody    string                 `json:"html_body,omitempty"`
	Attachments []EmailAttachment      `json:"attachments,omitempty"`
	Priority    int                    `json:"priority"` // 1=highest, 5=lowest
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Retries     int                    `json:"retries"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
	Size        int64  `json:"size"`
}

// EmailBatch represents a batch of emails to be sent together
type EmailBatch struct {
	ID        string     `json:"id"`
	Emails    []EmailJob `json:"emails"`
	CreatedAt time.Time  `json:"created_at"`
	Priority  int        `json:"priority"`
}

// EmailResult represents the result of sending an email
type EmailResult struct {
	EmailID      string        `json:"email_id"`
	BatchID      string        `json:"batch_id"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	DeliveryTime time.Duration `json:"delivery_time"`
	Timestamp    time.Time     `json:"timestamp"`
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseTLS   bool   `json:"use_tls"`
	From     string `json:"from"`
}

// NewEmailBatchService creates a new email batch service
func NewEmailBatchService(redis *redis.Client, smtpConfig SMTPConfig) *EmailBatchService {
	ctx, cancel := context.WithCancel(context.Background())

	return &EmailBatchService{
		redis:         redis,
		batchSize:     10,               // Send up to 10 emails per batch
		flushInterval: 30 * time.Second, // Force batch send every 30 seconds
		maxRetries:    3,
		workerCount:   2, // Conservative worker count for email sending

		emailQueue:  make(chan EmailJob, 1000),
		batchQueue:  make(chan EmailBatch, 100),
		resultQueue: make(chan EmailResult, 1000),

		ctx:    ctx,
		cancel: cancel,

		rateLimitPerSecond: 5,  // 5 emails per second max
		rateLimitBurst:     10, // Allow bursts up to 10

		smtpConfig: smtpConfig,
	}
}

// Start initializes and starts the email batch service
func (e *EmailBatchService) Start() error {
	log.Println("📧 Starting email batch service...")

	// Start batch creator
	go e.batchCreator()

	// Start email workers
	for i := 0; i < e.workerCount; i++ {
		e.wg.Add(1)
		go e.emailWorker(i)
	}

	// Start result processor
	go e.resultProcessor()

	// Start periodic batch flusher
	go e.periodicFlusher()

	// Start queue processor for Redis persistence
	go e.queueProcessor()

	log.Printf("📧 Email batch service started with %d workers", e.workerCount)
	log.Printf("📧 Batch configuration: size=%d, interval=%v, rate_limit=%d/sec",
		e.batchSize, e.flushInterval, e.rateLimitPerSecond)

	return nil
}

// Stop gracefully shuts down the email batch service
func (e *EmailBatchService) Stop() error {
	log.Println("🛑 Stopping email batch service...")

	e.cancel()
	close(e.emailQueue)

	// Wait for workers to finish
	e.wg.Wait()

	close(e.batchQueue)
	close(e.resultQueue)

	log.Println("✅ Email batch service stopped")
	return nil
}

// QueueEmail adds an email to the processing queue
func (e *EmailBatchService) QueueEmail(email EmailJob) error {
	email.ID = fmt.Sprintf("email_%d_%s", time.Now().UnixNano(), email.Subject[:utils.Min(10, len(email.Subject))])
	email.CreatedAt = time.Now()

	// Persist to Redis for durability
	emailData, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email: %v", err)
	}

	// Store in Redis with expiration
	key := fmt.Sprintf("email_queue:%d:%s", email.Priority, email.ID)
	if err := e.redis.Set(e.ctx, key, emailData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to persist email to Redis: %v", err)
	}

	// Add to processing queue
	select {
	case e.emailQueue <- email:
		e.mutex.Lock()
		e.totalEmails++
		e.mutex.Unlock()
		log.Printf("📧 Queued email %s (priority: %d)", email.ID, email.Priority)
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("email queue timeout")
	}
}

// batchCreator groups emails into batches
func (e *EmailBatchService) batchCreator() {
	currentBatch := EmailBatch{
		ID:        fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		Emails:    make([]EmailJob, 0, e.batchSize),
		CreatedAt: time.Now(),
		Priority:  5, // Start with lowest priority
	}

	flushTicker := time.NewTicker(e.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case email, ok := <-e.emailQueue:
			if !ok {
				// Queue closed, flush remaining emails
				if len(currentBatch.Emails) > 0 {
					e.flushBatch(currentBatch)
				}
				return
			}

			// Add email to current batch
			currentBatch.Emails = append(currentBatch.Emails, email)

			// Update batch priority to highest priority email
			if email.Priority < currentBatch.Priority {
				currentBatch.Priority = email.Priority
			}

			// Flush batch if full
			if len(currentBatch.Emails) >= e.batchSize {
				e.flushBatch(currentBatch)
				currentBatch = e.createNewBatch()
			}

		case <-flushTicker.C:
			// Periodic flush of partial batches
			if len(currentBatch.Emails) > 0 {
				log.Printf("📧 Periodic flush: sending batch with %d emails", len(currentBatch.Emails))
				e.flushBatch(currentBatch)
				currentBatch = e.createNewBatch()
			}

		case <-e.ctx.Done():
			return
		}
	}
}

// flushBatch sends a batch to be processed
func (e *EmailBatchService) flushBatch(batch EmailBatch) {
	select {
	case e.batchQueue <- batch:
		log.Printf("📧 Flushed batch %s with %d emails (priority: %d)",
			batch.ID, len(batch.Emails), batch.Priority)
	case <-time.After(5 * time.Second):
		log.Printf("⚠️  Batch queue timeout for batch %s", batch.ID)
	}
}

// createNewBatch creates a new empty batch
func (e *EmailBatchService) createNewBatch() EmailBatch {
	return EmailBatch{
		ID:        fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		Emails:    make([]EmailJob, 0, e.batchSize),
		CreatedAt: time.Now(),
		Priority:  5,
	}
}

// emailWorker processes email batches
func (e *EmailBatchService) emailWorker(id int) {
	defer e.wg.Done()

	rateLimiter := time.NewTicker(time.Second / time.Duration(e.rateLimitPerSecond))
	defer rateLimiter.Stop()

	for batch := range e.batchQueue {
		log.Printf("📧 Worker %d processing batch %s with %d emails", id, batch.ID, len(batch.Emails))

		for _, email := range batch.Emails {
			// Rate limiting
			<-rateLimiter.C

			startTime := time.Now()
			success, err := e.sendEmail(email)
			deliveryTime := time.Since(startTime)

			result := EmailResult{
				EmailID:      email.ID,
				BatchID:      batch.ID,
				Success:      success,
				DeliveryTime: deliveryTime,
				Timestamp:    time.Now(),
			}

			if err != nil {
				result.Error = err.Error()
			}

			// Send result to processor
			select {
			case e.resultQueue <- result:
			case <-time.After(5 * time.Second):
				log.Printf("⚠️  Result queue timeout for email %s", email.ID)
			}

			// Handle retries for failed emails
			if !success && email.Retries < e.maxRetries {
				email.Retries++
				log.Printf("🔄 Retrying email %s (attempt %d/%d)", email.ID, email.Retries, e.maxRetries)

				// Add back to queue with delay
				go func(retryEmail EmailJob) {
					time.Sleep(time.Duration(retryEmail.Retries) * time.Minute)
					select {
					case e.emailQueue <- retryEmail:
					case <-time.After(5 * time.Second):
						log.Printf("⚠️  Failed to requeue email %s for retry", retryEmail.ID)
					}
				}(email)
			}
		}

		e.mutex.Lock()
		e.batchesSent++
		e.mutex.Unlock()
	}
}

// sendEmail sends a single email (placeholder implementation)
func (e *EmailBatchService) sendEmail(email EmailJob) (bool, error) {
	// In production, this would use a real SMTP client or email service API
	// For now, simulate email sending

	log.Printf("📨 Sending email %s to %v: %s", email.ID, email.To, email.Subject)

	// Simulate sending time
	time.Sleep(time.Millisecond * 100)

	// Simulate occasional failures (5% failure rate)
	if time.Now().UnixNano()%20 == 0 {
		return false, fmt.Errorf("simulated email delivery failure")
	}

	return true, nil
}

// resultProcessor handles email sending results
func (e *EmailBatchService) resultProcessor() {
	for result := range e.resultQueue {
		e.mutex.Lock()
		if result.Success {
			e.sentEmails++
			// Update average delivery time
			if e.averageDeliveryTime == 0 {
				e.averageDeliveryTime = result.DeliveryTime
			} else {
				e.averageDeliveryTime = (e.averageDeliveryTime + result.DeliveryTime) / 2
			}
		} else {
			e.failedEmails++
		}
		e.mutex.Unlock()

		// Store result in Redis for analytics
		resultKey := fmt.Sprintf("email_result:%s", result.EmailID)
		resultData, _ := json.Marshal(result)
		e.redis.Set(e.ctx, resultKey, resultData, 7*24*time.Hour) // Keep for 7 days

		if result.Success {
			// Remove from queue
			queueKey := fmt.Sprintf("email_queue:*:%s", result.EmailID)
			keys, err := e.redis.Keys(e.ctx, queueKey).Result()
			if err == nil {
				for _, key := range keys {
					e.redis.Del(e.ctx, key)
				}
			}
		}
	}
}

// periodicFlusher forces periodic batch flushes
func (e *EmailBatchService) periodicFlusher() {
	ticker := time.NewTicker(e.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// This is handled by batchCreator
		case <-e.ctx.Done():
			return
		}
	}
}

// queueProcessor processes persistent queues from Redis
func (e *EmailBatchService) queueProcessor() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.processPersistedEmails()
		case <-e.ctx.Done():
			return
		}
	}
}

// processPersistedEmails processes emails persisted in Redis
func (e *EmailBatchService) processPersistedEmails() {
	// Get all persisted emails
	keys, err := e.redis.Keys(e.ctx, "email_queue:*").Result()
	if err != nil {
		log.Printf("⚠️  Failed to get persisted emails: %v", err)
		return
	}

	if len(keys) == 0 {
		return
	}

	log.Printf("📧 Processing %d persisted emails", len(keys))

	for _, key := range keys {
		emailData, err := e.redis.Get(e.ctx, key).Result()
		if err != nil {
			continue
		}

		var email EmailJob
		if err := json.Unmarshal([]byte(emailData), &email); err != nil {
			continue
		}

		// Check if email is scheduled for future
		if email.ScheduledAt != nil && email.ScheduledAt.After(time.Now()) {
			continue
		}

		// Add to processing queue
		select {
		case e.emailQueue <- email:
		case <-time.After(time.Second):
			// Queue full, try again later
		}
	}
}

// GetStats returns email processing statistics
func (e *EmailBatchService) GetStats() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	deliveryRate := float64(0)
	if e.totalEmails > 0 {
		deliveryRate = (float64(e.sentEmails) / float64(e.totalEmails)) * 100
	}

	return map[string]interface{}{
		"total_emails":          e.totalEmails,
		"sent_emails":           e.sentEmails,
		"failed_emails":         e.failedEmails,
		"pending_emails":        e.totalEmails - e.sentEmails - e.failedEmails,
		"batches_sent":          e.batchesSent,
		"delivery_rate":         deliveryRate,
		"average_delivery_time": e.averageDeliveryTime.String(),
		"queue_size":            len(e.emailQueue),
		"batch_queue_size":      len(e.batchQueue),
		"worker_count":          e.workerCount,
		"configuration": map[string]interface{}{
			"batch_size":            e.batchSize,
			"flush_interval":        e.flushInterval.String(),
			"max_retries":           e.maxRetries,
			"rate_limit_per_second": e.rateLimitPerSecond,
		},
	}
}
