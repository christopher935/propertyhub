package services

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// PerformanceMonitoringService tracks system performance and optimization metrics
type PerformanceMonitoringService struct {
	redis  *redis.Client
	ctx    context.Context
	cancel context.CancelFunc

	// Service references for collecting metrics
	analyticsCacheService  *AnalyticsCacheService
	fubBatchService        *FUBBatchService
	emailBatchService      *EmailBatchService
	photoProcessingService *PhotoProcessingService

	// Metrics collection
	metrics      map[string]interface{}
	metricsMutex sync.RWMutex

	// Configuration
	collectionInterval time.Duration
	retentionPeriod    time.Duration

	// Performance tracking
	requestMetrics map[string]*RequestMetrics
	requestMutex   sync.RWMutex

	// System monitoring
	systemMetrics *SystemMetrics
	systemMutex   sync.RWMutex
}

// RequestMetrics tracks HTTP request performance
type RequestMetrics struct {
	Path            string        `json:"path"`
	Method          string        `json:"method"`
	Count           int64         `json:"count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	ErrorCount      int64         `json:"error_count"`
	LastAccessed    time.Time     `json:"last_accessed"`
	StatusCodes     map[int]int64 `json:"status_codes"`
}

// SystemMetrics tracks system resource usage
type SystemMetrics struct {
	Timestamp           time.Time     `json:"timestamp"`
	CPUUsage            float64       `json:"cpu_usage"`
	MemoryUsage         uint64        `json:"memory_usage"`
	MemoryPercent       float64       `json:"memory_percent"`
	GoroutineCount      int           `json:"goroutine_count"`
	GCStats             debug.GCStats `json:"gc_stats"`
	DatabaseConnections int           `json:"database_connections"`
	RedisConnections    int           `json:"redis_connections"`
}

// PerformanceReport represents a comprehensive performance report
type PerformanceReport struct {
	GeneratedAt       time.Time                  `json:"generated_at"`
	SystemMetrics     *SystemMetrics             `json:"system_metrics"`
	RequestMetrics    map[string]*RequestMetrics `json:"request_metrics"`
	ServiceMetrics    map[string]interface{}     `json:"service_metrics"`
	OptimizationStats *OptimizationStats         `json:"optimization_stats"`
	Recommendations   []string                   `json:"recommendations"`
	Alerts            []Alert                    `json:"alerts"`
}

// OptimizationStats tracks optimization effectiveness
type OptimizationStats struct {
	CacheHitRatio      float64 `json:"cache_hit_ratio"`
	CompressionRatio   float64 `json:"compression_ratio"`
	BatchingEfficiency float64 `json:"batching_efficiency"`
	ImageOptimization  float64 `json:"image_optimization"`
	EmailDeliveryRate  float64 `json:"email_delivery_rate"`
	OverallPerformance float64 `json:"overall_performance"`
}

// Alert represents a performance alert
type Alert struct {
	Level     string    `json:"level"` // "info", "warning", "critical"
	Component string    `json:"component"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// NewPerformanceMonitoringService creates a new performance monitoring service
func NewPerformanceMonitoringService(redis *redis.Client) *PerformanceMonitoringService {
	ctx, cancel := context.WithCancel(context.Background())

	return &PerformanceMonitoringService{
		redis:              redis,
		ctx:                ctx,
		cancel:             cancel,
		metrics:            make(map[string]interface{}),
		requestMetrics:     make(map[string]*RequestMetrics),
		collectionInterval: 30 * time.Second,
		retentionPeriod:    7 * 24 * time.Hour, // 7 days
		systemMetrics:      &SystemMetrics{},
	}
}

// RegisterServices registers services for metric collection
func (p *PerformanceMonitoringService) RegisterServices(
	analyticsCache *AnalyticsCacheService,
	fubBatch *FUBBatchService,
	emailBatch *EmailBatchService,
	photoProcessing *PhotoProcessingService,
) {
	p.analyticsCacheService = analyticsCache
	p.fubBatchService = fubBatch
	p.emailBatchService = emailBatch
	p.photoProcessingService = photoProcessing
}

// Start begins performance monitoring
func (p *PerformanceMonitoringService) Start() {
	fmt.Println("📊 Starting performance monitoring service...")

	// Start metric collection
	go p.collectMetrics()

	// Start system monitoring
	go p.monitorSystem()

	// Start alert processing
	go p.processAlerts()

	fmt.Printf("📊 Performance monitoring active (collection interval: %v)\n", p.collectionInterval)
}

// Stop gracefully shuts down the monitoring service
func (p *PerformanceMonitoringService) Stop() {
	fmt.Println("🛑 Stopping performance monitoring service...")
	p.cancel()
	fmt.Println("✅ Performance monitoring stopped")
}

// TrackRequest records metrics for an HTTP request
func (p *PerformanceMonitoringService) TrackRequest(method, path string, duration time.Duration, statusCode int, isError bool) {
	key := fmt.Sprintf("%s:%s", method, path)

	p.requestMutex.Lock()
	defer p.requestMutex.Unlock()

	metric, exists := p.requestMetrics[key]
	if !exists {
		metric = &RequestMetrics{
			Path:        path,
			Method:      method,
			Count:       0,
			MinDuration: duration,
			MaxDuration: duration,
			StatusCodes: make(map[int]int64),
		}
		p.requestMetrics[key] = metric
	}

	// Update metrics
	metric.Count++
	metric.TotalDuration += duration
	metric.AverageDuration = metric.TotalDuration / time.Duration(metric.Count)
	metric.LastAccessed = time.Now()

	if duration < metric.MinDuration {
		metric.MinDuration = duration
	}
	if duration > metric.MaxDuration {
		metric.MaxDuration = duration
	}

	if isError {
		metric.ErrorCount++
	}

	metric.StatusCodes[statusCode]++
}

// collectMetrics periodically collects metrics from all services
func (p *PerformanceMonitoringService) collectMetrics() {
	ticker := time.NewTicker(p.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.gatherServiceMetrics()
			p.storeMetrics()
		case <-p.ctx.Done():
			return
		}
	}
}

// gatherServiceMetrics collects metrics from all registered services
func (p *PerformanceMonitoringService) gatherServiceMetrics() {
	p.metricsMutex.Lock()
	defer p.metricsMutex.Unlock()

	timestamp := time.Now()
	p.metrics["timestamp"] = timestamp

	// Analytics cache metrics
	if p.analyticsCacheService != nil {
		p.metrics["analytics_cache"] = p.analyticsCacheService.GetCacheStatistics()
	}

	// FUB batch service metrics
	if p.fubBatchService != nil {
		p.metrics["fub_batch"] = p.fubBatchService.GetBatchStats()
	}

	// Email batch service metrics
	if p.emailBatchService != nil {
		p.metrics["email_batch"] = p.emailBatchService.GetStats()
	}

	// Photo processing service metrics
	if p.photoProcessingService != nil {
		p.metrics["photo_processing"] = p.photoProcessingService.GetStats()
	}

	// Request metrics
	p.requestMutex.RLock()
	requestMetricsCopy := make(map[string]*RequestMetrics)
	for k, v := range p.requestMetrics {
		requestMetricsCopy[k] = v
	}
	p.requestMutex.RUnlock()
	p.metrics["requests"] = requestMetricsCopy
}

// monitorSystem monitors system resource usage
func (p *PerformanceMonitoringService) monitorSystem() {
	ticker := time.NewTicker(10 * time.Second) // More frequent system monitoring
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.updateSystemMetrics()
		case <-p.ctx.Done():
			return
		}
	}
}

// updateSystemMetrics updates system resource metrics
func (p *PerformanceMonitoringService) updateSystemMetrics() {
	p.systemMutex.Lock()
	defer p.systemMutex.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	p.systemMetrics.Timestamp = time.Now()
	p.systemMetrics.MemoryUsage = memStats.Alloc
	p.systemMetrics.GoroutineCount = runtime.NumGoroutine()

	// Calculate memory percentage (simplified)
	p.systemMetrics.MemoryPercent = float64(memStats.Alloc) / float64(memStats.Sys) * 100

	// GC stats
	debug.ReadGCStats(&p.systemMetrics.GCStats)
}

// storeMetrics stores current metrics in Redis
func (p *PerformanceMonitoringService) storeMetrics() {
	key := fmt.Sprintf("performance_metrics:%d", time.Now().Unix())

	data, err := json.Marshal(p.metrics)
	if err != nil {
		fmt.Printf("⚠️  Failed to marshal metrics: %v\n", err)
		return
	}

	// Store with expiration
	if err := p.redis.Set(p.ctx, key, data, p.retentionPeriod).Err(); err != nil {
		fmt.Printf("⚠️  Failed to store metrics: %v\n", err)
		return
	}

	// Maintain a list of metric keys for easier retrieval
	listKey := "performance_metrics_list"
	p.redis.LPush(p.ctx, listKey, key)
	p.redis.LTrim(p.ctx, listKey, 0, 1000) // Keep last 1000 entries
	p.redis.Expire(p.ctx, listKey, p.retentionPeriod)
}

// GenerateReport generates a comprehensive performance report
func (p *PerformanceMonitoringService) GenerateReport() (*PerformanceReport, error) {
	p.metricsMutex.RLock()
	p.requestMutex.RLock()
	p.systemMutex.RLock()

	defer p.metricsMutex.RUnlock()
	defer p.requestMutex.RUnlock()
	defer p.systemMutex.RUnlock()

	// Create report
	report := &PerformanceReport{
		GeneratedAt:       time.Now(),
		SystemMetrics:     p.systemMetrics,
		RequestMetrics:    p.requestMetrics,
		ServiceMetrics:    p.metrics,
		OptimizationStats: p.calculateOptimizationStats(),
		Recommendations:   p.generateRecommendations(),
		Alerts:            p.generateAlerts(),
	}

	return report, nil
}

// calculateOptimizationStats calculates optimization effectiveness
func (p *PerformanceMonitoringService) calculateOptimizationStats() *OptimizationStats {
	stats := &OptimizationStats{}

	// Analytics cache hit ratio
	if analyticsStats, ok := p.metrics["analytics_cache"].(map[string]interface{}); ok {
		if hitRatio, ok := analyticsStats["hit_ratio"].(float64); ok {
			stats.CacheHitRatio = hitRatio
		}
	}

	// Email delivery rate
	if emailStats, ok := p.metrics["email_batch"].(map[string]interface{}); ok {
		if deliveryRate, ok := emailStats["delivery_rate"].(float64); ok {
			stats.EmailDeliveryRate = deliveryRate
		}
	}

	// Photo optimization (compression ratio)
	if photoStats, ok := p.metrics["photo_processing"].(map[string]interface{}); ok {
		if compressionRatio, ok := photoStats["compression_ratio"].(float64); ok {
			stats.ImageOptimization = compressionRatio
		}
	}

	// FUB batching efficiency
	if fubStats, ok := p.metrics["fub_batch"].(map[string]interface{}); ok {
		if successRate, ok := fubStats["success_rate"].(float64); ok {
			stats.BatchingEfficiency = successRate
		}
	}

	// Calculate overall performance score
	scores := []float64{stats.CacheHitRatio, stats.EmailDeliveryRate, stats.ImageOptimization, stats.BatchingEfficiency}
	total := 0.0
	count := 0
	for _, score := range scores {
		if score > 0 {
			total += score
			count++
		}
	}
	if count > 0 {
		stats.OverallPerformance = total / float64(count)
	}

	return stats
}

// generateRecommendations generates performance improvement recommendations
func (p *PerformanceMonitoringService) generateRecommendations() []string {
	recommendations := []string{}

	// System performance recommendations
	if p.systemMetrics.MemoryPercent > 80 {
		recommendations = append(recommendations, "High memory usage detected. Consider increasing server memory or optimizing memory-intensive operations.")
	}

	if p.systemMetrics.GoroutineCount > 1000 {
		recommendations = append(recommendations, "High goroutine count detected. Review concurrent operations for potential goroutine leaks.")
	}

	// Cache performance recommendations
	if analyticsStats, ok := p.metrics["analytics_cache"].(map[string]interface{}); ok {
		if hitRatio, ok := analyticsStats["hit_ratio"].(float64); ok && hitRatio < 70 {
			recommendations = append(recommendations, "Analytics cache hit ratio is low. Consider increasing cache TTL or reviewing cache invalidation strategy.")
		}
	}

	// Request performance recommendations
	for _, metric := range p.requestMetrics {
		if metric.AverageDuration > 2*time.Second {
			recommendations = append(recommendations, fmt.Sprintf("Slow endpoint detected: %s %s (avg: %v). Consider optimization.", metric.Method, metric.Path, metric.AverageDuration))
		}

		errorRate := float64(metric.ErrorCount) / float64(metric.Count) * 100
		if errorRate > 5 {
			recommendations = append(recommendations, fmt.Sprintf("High error rate on %s %s (%.1f%%). Review error handling.", metric.Method, metric.Path, errorRate))
		}
	}

	return recommendations
}

// generateAlerts generates performance alerts
func (p *PerformanceMonitoringService) generateAlerts() []Alert {
	alerts := []Alert{}
	now := time.Now()

	// System alerts
	if p.systemMetrics.MemoryPercent > 90 {
		alerts = append(alerts, Alert{
			Level:     "critical",
			Component: "system",
			Message:   fmt.Sprintf("Critical memory usage: %.1f%%", p.systemMetrics.MemoryPercent),
			Timestamp: now,
		})
	}

	// Service alerts
	if emailStats, ok := p.metrics["email_batch"].(map[string]interface{}); ok {
		if pendingEmails, ok := emailStats["pending_emails"].(int64); ok && pendingEmails > 1000 {
			alerts = append(alerts, Alert{
				Level:     "warning",
				Component: "email_batch",
				Message:   fmt.Sprintf("High email queue: %d pending emails", pendingEmails),
				Timestamp: now,
			})
		}
	}

	return alerts
}

// processAlerts processes and stores alerts
func (p *PerformanceMonitoringService) processAlerts() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			alerts := p.generateAlerts()
			if len(alerts) > 0 {
				p.storeAlerts(alerts)
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// storeAlerts stores alerts in Redis
func (p *PerformanceMonitoringService) storeAlerts(alerts []Alert) {
	for _, alert := range alerts {
		key := fmt.Sprintf("performance_alert:%s:%d", alert.Component, time.Now().UnixNano())
		data, err := json.Marshal(alert)
		if err != nil {
			continue
		}

		p.redis.Set(p.ctx, key, data, 24*time.Hour)

		// Log critical alerts
		if alert.Level == "critical" {
			fmt.Printf("🚨 CRITICAL ALERT [%s]: %s\n", alert.Component, alert.Message)
		}
	}
}

// GetHistoricalMetrics retrieves historical performance metrics
func (p *PerformanceMonitoringService) GetHistoricalMetrics(hours int) ([]map[string]interface{}, error) {
	listKey := "performance_metrics_list"

	// Get metric keys from the last N hours
	keys, err := p.redis.LRange(p.ctx, listKey, 0, int64(hours*2)).Result() // Approximate
	if err != nil {
		return nil, fmt.Errorf("failed to get metric keys: %v", err)
	}

	var metrics []map[string]interface{}
	for _, key := range keys {
		data, err := p.redis.Get(p.ctx, key).Result()
		if err != nil {
			continue
		}

		var metric map[string]interface{}
		if err := json.Unmarshal([]byte(data), &metric); err != nil {
			continue
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetCurrentStats returns current performance statistics
func (p *PerformanceMonitoringService) GetCurrentStats() map[string]interface{} {
	p.metricsMutex.RLock()
	defer p.metricsMutex.RUnlock()

	// Create a copy of current metrics
	currentStats := make(map[string]interface{})
	for k, v := range p.metrics {
		currentStats[k] = v
	}

	// Add system metrics
	p.systemMutex.RLock()
	currentStats["system"] = p.systemMetrics
	p.systemMutex.RUnlock()

	return currentStats
}
