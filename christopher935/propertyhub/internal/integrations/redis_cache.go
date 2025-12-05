package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager handles Redis caching for dashboard data
type CacheManager struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheManager creates a new cache manager
func NewCacheManager(redisURL string) *CacheManager {
	// Parse Redis URL or use default settings
	opts := &redis.Options{
		Addr:     "localhost:6379", // Default Redis address
		Password: "",               // No password
		DB:       0,                // Default DB
	}

	// If Redis URL is provided, parse it
	if redisURL != "" {
		parsedOpts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Error parsing Redis URL, using defaults: %v", err)
		} else {
			opts = parsedOpts
		}
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	// Test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis connection failed, caching disabled: %v", err)
		return &CacheManager{client: nil, ctx: ctx}
	}

	log.Println("Redis cache connected successfully")
	return &CacheManager{client: client, ctx: ctx}
}

// Cache key constants
const (
	DashboardMetricsKey       = "dashboard:metrics"
	LeadAnalyticsKey          = "analytics:leads"
	CommunicationAnalyticsKey = "analytics:communication"
	AgentPerformanceKey       = "analytics:agents"
	ConversionFunnelKey       = "analytics:funnel"
	RevenueMetricsKey         = "analytics:revenue"

	// Cache TTL durations
	ShortTTL  = 5 * time.Minute  // For frequently changing data
	MediumTTL = 15 * time.Minute // For moderately changing data
	LongTTL   = 1 * time.Hour    // For slowly changing data
)

// CacheData represents cached data with metadata
type CacheData struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       int64       `json:"ttl"`
}

// Set stores data in cache with TTL
func (cm *CacheManager) Set(key string, data interface{}, ttl time.Duration) error {
	if cm.client == nil {
		return fmt.Errorf("cache not available")
	}

	cacheData := CacheData{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       int64(ttl.Seconds()),
	}

	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("error marshaling cache data: %w", err)
	}

	err = cm.client.Set(cm.ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("error setting cache: %w", err)
	}

	log.Printf("Cached data for key: %s (TTL: %v)", key, ttl)
	return nil
}

// Get retrieves data from cache
func (cm *CacheManager) Get(key string, dest interface{}) (bool, error) {
	if cm.client == nil {
		return false, fmt.Errorf("cache not available")
	}

	jsonData, err := cm.client.Get(cm.ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Cache miss
	}
	if err != nil {
		return false, fmt.Errorf("error getting cache: %w", err)
	}

	var cacheData CacheData
	err = json.Unmarshal([]byte(jsonData), &cacheData)
	if err != nil {
		return false, fmt.Errorf("error unmarshaling cache data: %w", err)
	}

	// Convert the data back to the destination type
	dataBytes, err := json.Marshal(cacheData.Data)
	if err != nil {
		return false, fmt.Errorf("error marshaling cached data: %w", err)
	}

	err = json.Unmarshal(dataBytes, dest)
	if err != nil {
		return false, fmt.Errorf("error unmarshaling to destination: %w", err)
	}

	log.Printf("Cache hit for key: %s (age: %v)", key, time.Since(cacheData.Timestamp))
	return true, nil
}

// Delete removes data from cache
func (cm *CacheManager) Delete(key string) error {
	if cm.client == nil {
		return fmt.Errorf("cache not available")
	}

	err := cm.client.Del(cm.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error deleting cache: %w", err)
	}

	log.Printf("Deleted cache for key: %s", key)
	return nil
}

// Exists checks if a key exists in cache
func (cm *CacheManager) Exists(key string) (bool, error) {
	if cm.client == nil {
		return false, fmt.Errorf("cache not available")
	}

	count, err := cm.client.Exists(cm.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("error checking cache existence: %w", err)
	}

	return count > 0, nil
}

// InvalidatePattern removes all keys matching a pattern
func (cm *CacheManager) InvalidatePattern(pattern string) error {
	if cm.client == nil {
		return fmt.Errorf("cache not available")
	}

	keys, err := cm.client.Keys(cm.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("error getting keys for pattern: %w", err)
	}

	if len(keys) > 0 {
		err = cm.client.Del(cm.ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("error deleting keys: %w", err)
		}
		log.Printf("Invalidated %d keys matching pattern: %s", len(keys), pattern)
	}

	return nil
}

// GetOrSet retrieves data from cache, or sets it if not found
func (cm *CacheManager) GetOrSet(key string, dest interface{}, ttl time.Duration, fetchFunc func() (interface{}, error)) error {
	// Try to get from cache first
	found, err := cm.Get(key, dest)
	if err != nil {
		log.Printf("Cache get error for key %s: %v", key, err)
	}

	if found {
		return nil // Cache hit
	}

	// Cache miss, fetch data
	log.Printf("Cache miss for key: %s, fetching data", key)
	data, err := fetchFunc()
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}

	// Store in cache
	err = cm.Set(key, data, ttl)
	if err != nil {
		log.Printf("Cache set error for key %s: %v", key, err)
		// Continue even if caching fails
	}

	// Convert data to destination
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling fetched data: %w", err)
	}

	err = json.Unmarshal(dataBytes, dest)
	if err != nil {
		return fmt.Errorf("error unmarshaling to destination: %w", err)
	}

	return nil
}

// SetMultiple stores multiple key-value pairs
func (cm *CacheManager) SetMultiple(data map[string]interface{}, ttl time.Duration) error {
	if cm.client == nil {
		return fmt.Errorf("cache not available")
	}

	pipe := cm.client.Pipeline()

	for key, value := range data {
		cacheData := CacheData{
			Data:      value,
			Timestamp: time.Now(),
			TTL:       int64(ttl.Seconds()),
		}

		jsonData, err := json.Marshal(cacheData)
		if err != nil {
			return fmt.Errorf("error marshaling cache data for key %s: %w", key, err)
		}

		pipe.Set(cm.ctx, key, jsonData, ttl)
	}

	_, err := pipe.Exec(cm.ctx)
	if err != nil {
		return fmt.Errorf("error executing pipeline: %w", err)
	}

	log.Printf("Cached %d keys with TTL: %v", len(data), ttl)
	return nil
}

// GetMultiple retrieves multiple keys
func (cm *CacheManager) GetMultiple(keys []string) (map[string]interface{}, error) {
	if cm.client == nil {
		return nil, fmt.Errorf("cache not available")
	}

	pipe := cm.client.Pipeline()

	for _, key := range keys {
		pipe.Get(cm.ctx, key)
	}

	results, err := pipe.Exec(cm.ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("error executing pipeline: %w", err)
	}

	data := make(map[string]interface{})

	for i, result := range results {
		cmd := result.(*redis.StringCmd)
		jsonData, err := cmd.Result()

		if err == redis.Nil {
			continue // Skip cache misses
		}
		if err != nil {
			log.Printf("Error getting key %s: %v", keys[i], err)
			continue
		}

		var cacheData CacheData
		err = json.Unmarshal([]byte(jsonData), &cacheData)
		if err != nil {
			log.Printf("Error unmarshaling cache data for key %s: %v", keys[i], err)
			continue
		}

		data[keys[i]] = cacheData.Data
	}

	log.Printf("Retrieved %d/%d keys from cache", len(data), len(keys))
	return data, nil
}

// GetCacheStats returns cache statistics
func (cm *CacheManager) GetCacheStats() (map[string]interface{}, error) {
	if cm.client == nil {
		return map[string]interface{}{
			"status": "disabled",
		}, nil
	}

	info, err := cm.client.Info(cm.ctx, "memory", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("error getting cache stats: %w", err)
	}

	// Get key count
	dbSize, err := cm.client.DBSize(cm.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("error getting DB size: %w", err)
	}

	return map[string]interface{}{
		"status":  "connected",
		"db_size": dbSize,
		"info":    info,
	}, nil
}

// InvalidateWebhookCache invalidates cache based on webhook events
func (cm *CacheManager) InvalidateWebhookCache(eventType string, resourceIDs []int) error {
	var patterns []string

	switch eventType {
	case "peopleCreated", "peopleUpdated", "peopleDeleted":
		patterns = append(patterns, DashboardMetricsKey, LeadAnalyticsKey, AgentPerformanceKey)
	case "peopleStageUpdated":
		patterns = append(patterns, ConversionFunnelKey, LeadAnalyticsKey, DashboardMetricsKey)
	case "emailsCreated", "textMessagesCreated", "callsCreated":
		patterns = append(patterns, CommunicationAnalyticsKey, DashboardMetricsKey)
	case "dealsCreated", "dealsUpdated":
		patterns = append(patterns, RevenueMetricsKey, DashboardMetricsKey, AgentPerformanceKey)
	case "tasksCreated", "appointmentsCreated":
		patterns = append(patterns, DashboardMetricsKey, AgentPerformanceKey)
	default:
		// For unknown events, invalidate dashboard metrics
		patterns = append(patterns, DashboardMetricsKey)
	}

	for _, pattern := range patterns {
		err := cm.Delete(pattern)
		if err != nil {
			log.Printf("Error invalidating cache for pattern %s: %v", pattern, err)
		}
	}

	log.Printf("Invalidated cache for event: %s (resources: %v)", eventType, resourceIDs)
	return nil
}

// WarmupCache pre-loads frequently accessed data
func (cm *CacheManager) WarmupCache(fetchFunctions map[string]func() (interface{}, error)) error {
	log.Println("Starting cache warmup...")

	for key, fetchFunc := range fetchFunctions {
		data, err := fetchFunc()
		if err != nil {
			log.Printf("Error warming up cache for key %s: %v", key, err)
			continue
		}

		var ttl time.Duration
		switch key {
		case DashboardMetricsKey:
			ttl = ShortTTL
		case LeadAnalyticsKey, CommunicationAnalyticsKey:
			ttl = MediumTTL
		default:
			ttl = LongTTL
		}

		err = cm.Set(key, data, ttl)
		if err != nil {
			log.Printf("Error setting cache during warmup for key %s: %v", key, err)
		}
	}

	log.Println("Cache warmup completed")
	return nil
}

// Close closes the Redis connection
func (cm *CacheManager) Close() error {
	if cm.client == nil {
		return nil
	}

	err := cm.client.Close()
	if err != nil {
		return fmt.Errorf("error closing cache connection: %w", err)
	}

	log.Println("Cache connection closed")
	return nil
}
