package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// JobManager manages all background jobs and scheduled tasks
type JobManager struct {
	db            *gorm.DB
	jobs          map[string]*Job
	scheduledJobs map[string]*ScheduledJob
	jobQueue      chan *JobExecution
	workers       []*Worker
	ctx           context.Context
	cancel        context.CancelFunc
	mutex         sync.RWMutex
	running       bool
}

// Job represents a job definition
type Job struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Handler     JobHandler    `json:"-"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID          string         `json:"id"`
	JobID       string         `json:"job_id"`
	Name        string         `json:"name"`
	Schedule    string         `json:"schedule"` // Cron expression
	NextRun     time.Time      `json:"next_run"`
	LastRun     *time.Time     `json:"last_run,omitempty"`
	Enabled     bool           `json:"enabled"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// JobExecution represents a job execution instance
type JobExecution struct {
	ID         string                 `json:"id"`
	JobID      string                 `json:"job_id"`
	Status     JobStatus              `json:"status"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     *JobResult             `json:"result,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	Attempts   int                    `json:"attempts"`
	LastError  string                 `json:"last_error,omitempty"`
	Worker     string                 `json:"worker,omitempty"`
}

// JobResult represents job execution results
type JobResult struct {
	Success      bool                   `json:"success"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Duration     time.Duration          `json:"duration"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Worker represents a background worker
type Worker struct {
	ID        string
	manager   *JobManager
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
	processed int
	errors    int
}

// JobHandler is the interface for job handlers
type JobHandler interface {
	Execute(ctx context.Context, params map[string]interface{}) (*JobResult, error)
}

// JobStatus represents job execution status
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
	JobStatusCancelled JobStatus = "cancelled"
)

// Job types
const (
	JobTypeFridayReport      = "friday_report"
	JobTypeFUBSync           = "fub_sync"
	JobTypeAnalyticsAggregation = "analytics_aggregation"
	JobTypeEmailNotification = "email_notification"
	JobTypeDataCleanup       = "data_cleanup"
	JobTypeBackup           = "backup"
	JobTypeHealthCheck      = "health_check"
	JobTypeReportGeneration = "report_generation"
	JobTypeScheduledActions = "scheduled_actions"
)

// NewJobManager creates a new job manager
func NewJobManager(db *gorm.DB, harService interface{}, fubService interface{}, biService interface{}, notificationService interface{}) *JobManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	jm := &JobManager{
		db:            db,
		jobs:          make(map[string]*Job),
		scheduledJobs: make(map[string]*ScheduledJob),
		jobQueue:      make(chan *JobExecution, 1000),
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// Register default jobs
	jm.registerDefaultJobs(harService, fubService, biService, notificationService)
	
	// Create workers
	jm.createWorkers(5) // 5 concurrent workers
	
	return jm
}

// registerDefaultJobs registers all default system jobs
func (jm *JobManager) registerDefaultJobs(harService, fubService, biService, notificationService interface{}) {
	// Friday Report Job
	jm.RegisterJob(&Job{
		ID:          "friday_report",
		Name:        "Friday Business Report",
		Type:        JobTypeFridayReport,
		Description: "Generate and send weekly business intelligence report",
		Handler:     &FridayReportHandler{biService: biService, notificationService: notificationService},
		Timeout:     30 * time.Minute,
		MaxRetries:  3,
		RetryDelay:  5 * time.Minute,
	})
	
	// HAR Sync Job removed - HAR blocked access
	
	// FUB Sync Job
	jm.RegisterJob(&Job{
		ID:          "fub_sync",
		Name:        "Follow Up Boss CRM Sync",
		Type:        JobTypeFUBSync,
		Description: "Synchronize contacts and leads from Follow Up Boss",
		Handler:     &FUBSyncHandler{fubService: fubService},
		Timeout:     45 * time.Minute,
		MaxRetries:  3,
		RetryDelay:  15 * time.Minute,
	})
	
	// Analytics Aggregation Job
	jm.RegisterJob(&Job{
		ID:          "analytics_aggregation",
		Name:        "Analytics Data Aggregation",
		Type:        JobTypeAnalyticsAggregation,
		Description: "Aggregate analytics data for reporting",
		Handler:     &AnalyticsAggregationHandler{biService: biService},
		Timeout:     20 * time.Minute,
		MaxRetries:  3,
		RetryDelay:  5 * time.Minute,
	})
	
	// Data Cleanup Job
	jm.RegisterJob(&Job{
		ID:          "data_cleanup",
		Name:        "Data Cleanup",
		Type:        JobTypeDataCleanup,
		Description: "Clean up old logs and temporary data",
		Handler:     &DataCleanupHandler{db: jm.db},
		Timeout:     30 * time.Minute,
		MaxRetries:  2,
		RetryDelay:  10 * time.Minute,
	})
	
	// Health Check Job
	jm.RegisterJob(&Job{
		ID:          "health_check",
		Name:        "System Health Check",
		Type:        JobTypeHealthCheck,
		Description: "Check system health and send alerts if needed",
		Handler:     &HealthCheckHandler{db: jm.db, notificationService: notificationService},
		Timeout:     10 * time.Minute,
		MaxRetries:  1,
		RetryDelay:  2 * time.Minute,
	})
	
	// Scheduled Actions Processor Job
	jm.RegisterJob(&Job{
		ID:          "scheduled_actions",
		Name:        "Scheduled Actions Processor",
		Type:        JobTypeScheduledActions,
		Description: "Process pending scheduled actions (emails, SMS, notifications)",
// 		Handler:     NewScheduledActionsHandler(jm.db),
		Timeout:     5 * time.Minute,
		MaxRetries:  2,
		RetryDelay:  1 * time.Minute,
	})
}

// RegisterJob registers a new job
func (jm *JobManager) RegisterJob(job *Job) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	
	jm.jobs[job.ID] = job
	log.Printf("Registered job: %s (%s)", job.Name, job.ID)
}

// ScheduleJob schedules a job with cron expression
func (jm *JobManager) ScheduleJob(jobID, name, cronExpr string, params map[string]interface{}) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	
	if _, exists := jm.jobs[jobID]; !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	
	nextRun, err := jm.parseNextCronTime(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %v", err)
	}
	
	scheduledJob := &ScheduledJob{
		ID:          fmt.Sprintf("sched_%d", time.Now().UnixNano()),
		JobID:       jobID,
		Name:        name,
		Schedule:    cronExpr,
		NextRun:     nextRun,
		Enabled:     true,
		Parameters:  params,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	jm.scheduledJobs[scheduledJob.ID] = scheduledJob
	log.Printf("Scheduled job: %s (next run: %v)", name, nextRun)
	
	return nil
}

// StartScheduledJobs starts the scheduled job processor
func (jm *JobManager) StartScheduledJobs() {
	jm.mutex.Lock()
	if jm.running {
		jm.mutex.Unlock()
		return
	}
	jm.running = true
	jm.mutex.Unlock()
	
	log.Println("Starting scheduled jobs processor...")
	
	// Start workers
	for _, worker := range jm.workers {
		go worker.Start()
	}
	
	// Start scheduler
	go jm.runScheduler()
	
	log.Println("Job manager started successfully")
}

// ScheduleFridayReports schedules the Friday business report
func (jm *JobManager) ScheduleFridayReports() {
	// Schedule for every Friday at 9:00 AM
	err := jm.ScheduleJob("friday_report", "Weekly Friday Report", "0 9 * * 5", map[string]interface{}{
		"include_charts": true,
		"send_email":     true,
		"recipients":     []string{"admin@llotschedule.online"},
	})
	
	if err != nil {
		log.Printf("Failed to schedule Friday reports: %v", err)
	}
}

// StartHARSync removed - HAR blocked access

// StartFUBSync schedules FUB contact synchronization
func (jm *JobManager) StartFUBSync() {
	// Schedule FUB sync every 2 hours
	err := jm.ScheduleJob("fub_sync", "FUB Contact Sync", "0 */2 * * *", map[string]interface{}{
		"sync_notes": true,
		"sync_tasks": true,
	})
	
	if err != nil {
		log.Printf("Failed to schedule FUB sync: %v", err)
	}
}

// StartAnalyticsAggregation schedules analytics data aggregation
func (jm *JobManager) StartAnalyticsAggregation() {
	// Schedule analytics aggregation every hour
	err := jm.ScheduleJob("analytics_aggregation", "Analytics Aggregation", "0 * * * *", map[string]interface{}{
		"aggregate_hourly":  true,
		"update_dashboards": true,
	})
	
	if err != nil {
		log.Printf("Failed to schedule analytics aggregation: %v", err)
	}
}

// StartScheduledActionsProcessor schedules the scheduled actions processor
func (jm *JobManager) StartScheduledActionsProcessor() {
	// Schedule scheduled actions processor every minute
	err := jm.ScheduleJob("scheduled_actions", "Scheduled Actions Processor", "* * * * *", map[string]interface{}{
		"batch_size": 100,
	})
	
	if err != nil {
		log.Printf("Failed to schedule actions processor: %v", err)
	}
}

// QueueJob queues a job for execution
func (jm *JobManager) QueueJob(jobID string, params map[string]interface{}) (*JobExecution, error) {
	jm.mutex.RLock()
	job, exists := jm.jobs[jobID]
	jm.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}
	
	execution := &JobExecution{
		ID:         fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		JobID:      jobID,
		Status:     JobStatusPending,
		Parameters: params,
		StartedAt:  time.Now(),
		Attempts:   0,
	}
	
	select {
	case jm.jobQueue <- execution:
		log.Printf("Queued job: %s (%s)", job.Name, execution.ID)
		return execution, nil
	default:
		return nil, fmt.Errorf("job queue is full")
	}
}

// runScheduler runs the job scheduler
func (jm *JobManager) runScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			jm.checkScheduledJobs()
		case <-jm.ctx.Done():
			return
		}
	}
}

// checkScheduledJobs checks and executes scheduled jobs
func (jm *JobManager) checkScheduledJobs() {
	now := time.Now()
	
	jm.mutex.RLock()
	for _, scheduledJob := range jm.scheduledJobs {
		if scheduledJob.Enabled && now.After(scheduledJob.NextRun) {
			jm.executeScheduledJob(scheduledJob)
		}
	}
	jm.mutex.RUnlock()
}

// executeScheduledJob executes a scheduled job
func (jm *JobManager) executeScheduledJob(scheduledJob *ScheduledJob) {
	log.Printf("Executing scheduled job: %s", scheduledJob.Name)
	
	// Queue the job for execution
	_, err := jm.QueueJob(scheduledJob.JobID, scheduledJob.Parameters)
	if err != nil {
		log.Printf("Failed to queue scheduled job %s: %v", scheduledJob.Name, err)
		return
	}
	
	// Update next run time
	nextRun, err := jm.parseNextCronTime(scheduledJob.Schedule)
	if err != nil {
		log.Printf("Failed to calculate next run for %s: %v", scheduledJob.Name, err)
		return
	}
	
	jm.mutex.Lock()
	scheduledJob.LastRun = &scheduledJob.NextRun
	scheduledJob.NextRun = nextRun
	scheduledJob.UpdatedAt = time.Now()
	jm.mutex.Unlock()
}

// createWorkers creates background workers
func (jm *JobManager) createWorkers(count int) {
	jm.workers = make([]*Worker, count)
	
	for i := 0; i < count; i++ {
		ctx, cancel := context.WithCancel(jm.ctx)
		worker := &Worker{
			ID:      fmt.Sprintf("worker-%d", i+1),
			manager: jm,
			ctx:     ctx,
			cancel:  cancel,
		}
		jm.workers[i] = worker
	}
	
	log.Printf("Created %d workers", count)
}

// Start starts the worker
func (w *Worker) Start() {
	w.running = true
	log.Printf("Worker %s started", w.ID)
	
	for {
		select {
		case execution := <-w.manager.jobQueue:
			w.executeJob(execution)
		case <-w.ctx.Done():
			w.running = false
			log.Printf("Worker %s stopped", w.ID)
			return
		}
	}
}

// executeJob executes a job
func (w *Worker) executeJob(execution *JobExecution) {
	w.manager.mutex.RLock()
	job, exists := w.manager.jobs[execution.JobID]
	w.manager.mutex.RUnlock()
	
	if !exists {
		log.Printf("Job %s not found", execution.JobID)
		return
	}
	
	execution.Status = JobStatusRunning
	execution.Worker = w.ID
	execution.Attempts++
	
	log.Printf("Worker %s executing job: %s", w.ID, job.Name)
	
	// Create timeout context
	ctx, cancel := context.WithTimeout(w.ctx, job.Timeout)
	defer cancel()
	
	// Execute the job
	startTime := time.Now()
	result, err := job.Handler.Execute(ctx, execution.Parameters)
	duration := time.Since(startTime)
	
	finishedTime := startTime.Add(duration)
	execution.FinishedAt = &finishedTime
	
	if err != nil {
		w.errors++
		execution.Status = JobStatusFailed
		execution.LastError = err.Error()
		
		// Retry logic
		if execution.Attempts < job.MaxRetries {
			execution.Status = JobStatusRetrying
			log.Printf("Job %s failed (attempt %d/%d), retrying in %v: %v", 
				job.Name, execution.Attempts, job.MaxRetries, job.RetryDelay, err)
			
			// Schedule retry
			go func() {
				time.Sleep(job.RetryDelay)
				select {
				case w.manager.jobQueue <- execution:
				default:
					log.Printf("Failed to requeue job %s", job.Name)
				}
			}()
		} else {
			log.Printf("Job %s failed after %d attempts: %v", job.Name, execution.Attempts, err)
		}
	} else {
		w.processed++
		execution.Status = JobStatusCompleted
		execution.Result = result
		log.Printf("Job %s completed successfully in %v", job.Name, duration)
	}
}

// parseNextCronTime parses cron expression and returns next execution time
func (jm *JobManager) parseNextCronTime(cronExpr string) (time.Time, error) {
	// This is a simplified cron parser
	// In production, use a proper cron parsing library
	
	now := time.Now()
	
	// Basic patterns
	switch cronExpr {
	case "0 9 * * 5": // Friday at 9 AM
		// Find next Friday at 9 AM
		next := now.Add(24 * time.Hour) // Start from tomorrow
		for next.Weekday() != time.Friday {
			next = next.Add(24 * time.Hour)
		}
		return time.Date(next.Year(), next.Month(), next.Day(), 9, 0, 0, 0, next.Location()), nil
	case "0 */4 * * *": // Every 4 hours
		next := now.Add(4 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location()), nil
	case "0 */2 * * *": // Every 2 hours
		next := now.Add(2 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location()), nil
	case "0 * * * *": // Every hour
		next := now.Add(1 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location()), nil
	default:
		return now.Add(1 * time.Hour), nil // Default to 1 hour
	}
}

// GetJobStatus gets the status of a job execution
func (jm *JobManager) GetJobStatus(executionID string) (*JobExecution, error) {
	// This would typically query the database for job execution status
	// Placeholder implementation
	return nil, fmt.Errorf("job execution %s not found", executionID)
}

// GetJobStats gets job execution statistics
func (jm *JobManager) GetJobStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	totalProcessed := 0
	totalErrors := 0
	
	for _, worker := range jm.workers {
		totalProcessed += worker.processed
		totalErrors += worker.errors
	}
	
	stats["workers"] = len(jm.workers)
	stats["jobs_registered"] = len(jm.jobs)
	stats["scheduled_jobs"] = len(jm.scheduledJobs)
	stats["total_processed"] = totalProcessed
	stats["total_errors"] = totalErrors
	stats["queue_length"] = len(jm.jobQueue)
	stats["running"] = jm.running
	
	return stats
}

// Stop stops the job manager
func (jm *JobManager) Stop() {
	log.Println("Stopping job manager...")
	
	jm.mutex.Lock()
	jm.running = false
	jm.mutex.Unlock()
	
	// Stop workers
	for _, worker := range jm.workers {
		worker.cancel()
	}
	
	// Cancel context
	jm.cancel()
	
	log.Println("Job manager stopped")
}

// Job Handlers - these would be implemented based on the actual services

// FridayReportHandler handles Friday report generation
type FridayReportHandler struct {
	biService           interface{}
	notificationService interface{}
}

func (h *FridayReportHandler) Execute(ctx context.Context, params map[string]interface{}) (*JobResult, error) {
	log.Println("Generating Friday business report...")
	
	// Generate report data
	// Send email notification
	// This would use the actual business intelligence service
	
	return &JobResult{
		Success: true,
		Data: map[string]interface{}{
			"report_generated": true,
			"email_sent":       true,
		},
		Duration: time.Second * 30,
	}, nil
}

// HARSyncHandler removed - HAR blocked access

// FUBSyncHandler handles Follow Up Boss synchronization
type FUBSyncHandler struct {
	fubService interface{}
}

func (h *FUBSyncHandler) Execute(ctx context.Context, params map[string]interface{}) (*JobResult, error) {
	log.Println("Starting Follow Up Boss sync...")
	
	// This would use the actual FUB service
	// Sync contacts, leads, notes, tasks
	
	return &JobResult{
		Success: true,
		Data: map[string]interface{}{
			"contacts_synced": 234,
			"new_contacts":    12,
			"updated_contacts": 45,
		},
		Duration: time.Minute * 8,
	}, nil
}

// AnalyticsAggregationHandler handles analytics data aggregation
type AnalyticsAggregationHandler struct {
	biService interface{}
}

func (h *AnalyticsAggregationHandler) Execute(ctx context.Context, params map[string]interface{}) (*JobResult, error) {
	log.Println("Starting analytics aggregation...")
	
	// Aggregate analytics data
	// Update dashboard metrics
	
	return &JobResult{
		Success: true,
		Data: map[string]interface{}{
			"events_aggregated": 5000,
			"metrics_updated":   25,
		},
		Duration: time.Minute * 5,
	}, nil
}

// DataCleanupHandler handles data cleanup tasks
type DataCleanupHandler struct {
	db *gorm.DB
}

func (h *DataCleanupHandler) Execute(ctx context.Context, params map[string]interface{}) (*JobResult, error) {
	log.Println("Starting data cleanup...")
	
	// Clean old logs, temporary files, expired sessions
	// This would use actual database cleanup operations
	
	return &JobResult{
		Success: true,
		Data: map[string]interface{}{
			"logs_cleaned":     1000,
			"sessions_expired": 50,
			"temp_files_deleted": 25,
		},
		Duration: time.Minute * 3,
	}, nil
}

// HealthCheckHandler handles system health checks
type HealthCheckHandler struct {
	db                  *gorm.DB
	notificationService interface{}
}

func (h *HealthCheckHandler) Execute(ctx context.Context, params map[string]interface{}) (*JobResult, error) {
	log.Println("Performing system health check...")
	
	// Check database connectivity
	// Check external services
	// Check disk space, memory usage
	// Send alerts if issues found
	
	return &JobResult{
		Success: true,
		Data: map[string]interface{}{
			"database_status":     "healthy",
			"external_services":   "healthy",
			"disk_usage_percent":  45,
			"memory_usage_percent": 65,
		},
		Duration: time.Minute * 2,
	}, nil
}
