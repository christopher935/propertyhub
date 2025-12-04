package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/repositories"
)

// AnalyticsCacheService provides high-performance analytics caching
type AnalyticsCacheService struct {
	redis        *redis.Client
	propertyRepo repositories.PropertyRepository
	bookingRepo  repositories.BookingRepository
	adminRepo    repositories.AdminRepository
	defaultTTL   time.Duration
	shortTTL     time.Duration
	longTTL      time.Duration
	cacheStats   *CacheStatistics
}

// CacheStatistics tracks cache performance
type CacheStatistics struct {
	Hits            int64 `json:"hits"`
	Misses          int64 `json:"misses"`
	Errors          int64 `json:"errors"`
	EvictionCount   int64 `json:"eviction_count"`
	TotalOperations int64 `json:"total_operations"`
}

// NewAnalyticsCacheService creates a new analytics cache service
func NewAnalyticsCacheService(
	redisClient *redis.Client,
	propertyRepo repositories.PropertyRepository,
	bookingRepo repositories.BookingRepository,
	adminRepo repositories.AdminRepository,
) *AnalyticsCacheService {
	return &AnalyticsCacheService{
		redis:        redisClient,
		propertyRepo: propertyRepo,
		bookingRepo:  bookingRepo,
		adminRepo:    adminRepo,
		defaultTTL:   5 * time.Minute,  // Default cache TTL
		shortTTL:     2 * time.Minute,  // For frequently changing data
		longTTL:      15 * time.Minute, // For stable data like cities
		cacheStats:   &CacheStatistics{},
	}
}

// Property Analytics Caching

// GetPropertyStatistics returns cached property statistics
func (acs *AnalyticsCacheService) GetPropertyStatistics(ctx context.Context) (*repositories.PropertyStatistics, error) {
	cacheKey := "analytics:property:statistics"

	// Try cache first
	if cachedStats := acs.getCachedPropertyStatistics(ctx, cacheKey); cachedStats != nil {
		acs.recordCacheHit()
		log.Printf("🚀 Cache HIT for property statistics")
		return cachedStats, nil
	}

	acs.recordCacheMiss()
	log.Printf("💾 Cache MISS for property statistics")

	// Fetch from repository
	stats, err := acs.propertyRepo.GetStatistics(ctx)
	if err != nil {
		acs.recordCacheError()
		return nil, fmt.Errorf("failed to get property statistics: %v", err)
	}

	// Cache the result
	if err := acs.cachePropertyStatistics(ctx, cacheKey, stats, acs.defaultTTL); err != nil {
		log.Printf("⚠️ Failed to cache property statistics: %v", err)
	}

	return stats, nil
}

// GetCities returns cached list of cities
func (acs *AnalyticsCacheService) GetCities(ctx context.Context) ([]string, error) {
	cacheKey := "analytics:property:cities"

	// Try cache first
	if cachedCities := acs.getCachedStringSlice(ctx, cacheKey); cachedCities != nil {
		acs.recordCacheHit()
		log.Printf("🚀 Cache HIT for cities list")
		return cachedCities, nil
	}

	acs.recordCacheMiss()
	log.Printf("💾 Cache MISS for cities list")

	// Fetch from repository
	cities, err := acs.propertyRepo.GetCities(ctx)
	if err != nil {
		acs.recordCacheError()
		return nil, fmt.Errorf("failed to get cities: %v", err)
	}

	// Cache the result with long TTL (cities don't change frequently)
	if err := acs.cacheStringSlice(ctx, cacheKey, cities, acs.longTTL); err != nil {
		log.Printf("⚠️ Failed to cache cities: %v", err)
	}

	return cities, nil
}

// GetFeaturedProperties returns cached featured properties
func (acs *AnalyticsCacheService) GetFeaturedProperties(ctx context.Context, limit int) ([]*models.Property, error) {
	cacheKey := fmt.Sprintf("analytics:property:featured:%d", limit)

	// Try cache first
	if cachedProperties := acs.getCachedProperties(ctx, cacheKey); cachedProperties != nil {
		acs.recordCacheHit()
		log.Printf("🚀 Cache HIT for featured properties (limit: %d)", limit)
		return cachedProperties, nil
	}

	acs.recordCacheMiss()
	log.Printf("💾 Cache MISS for featured properties (limit: %d)", limit)

	// Fetch from repository
	properties, err := acs.propertyRepo.GetFeaturedProperties(ctx, limit)
	if err != nil {
		acs.recordCacheError()
		return nil, fmt.Errorf("failed to get featured properties: %v", err)
	}

	// Cache the result with short TTL (featured properties change frequently)
	if err := acs.cacheProperties(ctx, cacheKey, properties, acs.shortTTL); err != nil {
		log.Printf("⚠️ Failed to cache featured properties: %v", err)
	}

	return properties, nil
}

// GetRecentProperties returns cached recent properties
func (acs *AnalyticsCacheService) GetRecentProperties(ctx context.Context, limit int) ([]*models.Property, error) {
	cacheKey := fmt.Sprintf("analytics:property:recent:%d", limit)

	// Try cache first
	if cachedProperties := acs.getCachedProperties(ctx, cacheKey); cachedProperties != nil {
		acs.recordCacheHit()
		log.Printf("🚀 Cache HIT for recent properties (limit: %d)", limit)
		return cachedProperties, nil
	}

	acs.recordCacheMiss()
	log.Printf("💾 Cache MISS for recent properties (limit: %d)", limit)

	// Fetch from repository
	properties, err := acs.propertyRepo.GetRecentProperties(ctx, limit)
	if err != nil {
		acs.recordCacheError()
		return nil, fmt.Errorf("failed to get recent properties: %v", err)
	}

	// Cache the result with short TTL (recent properties change frequently)
	if err := acs.cacheProperties(ctx, cacheKey, properties, acs.shortTTL); err != nil {
		log.Printf("⚠️ Failed to cache recent properties: %v", err)
	}

	return properties, nil
}

// Booking Analytics Caching

// GetBookingStatistics returns cached booking statistics
func (acs *AnalyticsCacheService) GetBookingStatistics(ctx context.Context) (*repositories.BookingStatistics, error) {
	cacheKey := "analytics:booking:statistics"

	// Try cache first
	if cachedStats := acs.getCachedBookingStatistics(ctx, cacheKey); cachedStats != nil {
		acs.recordCacheHit()
		log.Printf("🚀 Cache HIT for booking statistics")
		return cachedStats, nil
	}

	acs.recordCacheMiss()
	log.Printf("💾 Cache MISS for booking statistics")

	// Fetch from repository
	stats, err := acs.bookingRepo.GetStatistics(ctx)
	if err != nil {
		acs.recordCacheError()
		return nil, fmt.Errorf("failed to get booking statistics: %v", err)
	}

	// Cache the result
	if err := acs.cacheBookingStatistics(ctx, cacheKey, stats, acs.defaultTTL); err != nil {
		log.Printf("⚠️ Failed to cache booking statistics: %v", err)
	}

	return stats, nil
}

// GetBookingTrends returns cached booking trends
func (acs *AnalyticsCacheService) GetBookingTrends(ctx context.Context, days int) (map[string]int64, error) {
	cacheKey := fmt.Sprintf("analytics:booking:trends:%d", days)

	// Try cache first
	if cachedTrends := acs.getCachedTrends(ctx, cacheKey); cachedTrends != nil {
		acs.recordCacheHit()
		log.Printf("🚀 Cache HIT for booking trends (%d days)", days)
		return cachedTrends, nil
	}

	acs.recordCacheMiss()
	log.Printf("💾 Cache MISS for booking trends (%d days)", days)

	// Fetch from repository
	trends, err := acs.bookingRepo.GetBookingTrends(ctx, days)
	if err != nil {
		acs.recordCacheError()
		return nil, fmt.Errorf("failed to get booking trends: %v", err)
	}

	// Cache the result with default TTL
	if err := acs.cacheTrends(ctx, cacheKey, trends, acs.defaultTTL); err != nil {
		log.Printf("⚠️ Failed to cache booking trends: %v", err)
	}

	return trends, nil
}

// GetDashboardMetrics returns cached dashboard metrics
func (acs *AnalyticsCacheService) GetDashboardMetrics(ctx context.Context) (map[string]interface{}, error) {
	cacheKey := "analytics:dashboard:metrics"

	// Try cache first
	if cachedMetrics := acs.getCachedDashboardMetrics(ctx, cacheKey); cachedMetrics != nil {
		acs.recordCacheHit()
		log.Printf("🚀 Cache HIT for dashboard metrics")
		return cachedMetrics, nil
	}

	acs.recordCacheMiss()
	log.Printf("💾 Cache MISS for dashboard metrics")

	// Fetch from repository
	metrics, err := acs.adminRepo.GetDashboardMetrics(ctx)
	if err != nil {
		acs.recordCacheError()
		return nil, fmt.Errorf("failed to get dashboard metrics: %v", err)
	}

	// Cache the result
	if err := acs.cacheDashboardMetrics(ctx, cacheKey, metrics, acs.defaultTTL); err != nil {
		log.Printf("⚠️ Failed to cache dashboard metrics: %v", err)
	}

	return metrics, nil
}

// Cache invalidation methods

// InvalidatePropertyAnalytics clears property-related analytics cache
func (acs *AnalyticsCacheService) InvalidatePropertyAnalytics(ctx context.Context) error {
	keys := []string{
		"analytics:property:statistics",
		"analytics:property:cities",
		"analytics:property:featured:*",
		"analytics:property:recent:*",
	}

	for _, key := range keys {
		if err := acs.deleteFromCache(ctx, key); err != nil {
			log.Printf("⚠️ Failed to invalidate cache key %s: %v", key, err)
		}
	}

	log.Printf("🔄 Property analytics cache invalidated")
	return nil
}

// InvalidateBookingAnalytics clears booking-related analytics cache
func (acs *AnalyticsCacheService) InvalidateBookingAnalytics(ctx context.Context) error {
	keys := []string{
		"analytics:booking:statistics",
		"analytics:booking:trends:*",
		"analytics:dashboard:metrics",
	}

	for _, key := range keys {
		if err := acs.deleteFromCache(ctx, key); err != nil {
			log.Printf("⚠️ Failed to invalidate cache key %s: %v", key, err)
		}
	}

	log.Printf("🔄 Booking analytics cache invalidated")
	return nil
}

// GetCacheStatistics returns cache performance statistics
func (acs *AnalyticsCacheService) GetCacheStatistics() *CacheStatistics {
	return &CacheStatistics{
		Hits:            acs.cacheStats.Hits,
		Misses:          acs.cacheStats.Misses,
		Errors:          acs.cacheStats.Errors,
		TotalOperations: acs.cacheStats.TotalOperations,
	}
}

// Private helper methods for caching

func (acs *AnalyticsCacheService) getCachedPropertyStatistics(ctx context.Context, key string) *repositories.PropertyStatistics {
	if acs.redis == nil {
		return nil
	}

	val, err := acs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var stats repositories.PropertyStatistics
	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		log.Printf("⚠️ Failed to unmarshal property statistics from cache: %v", err)
		return nil
	}

	return &stats
}

func (acs *AnalyticsCacheService) cachePropertyStatistics(ctx context.Context, key string, stats *repositories.PropertyStatistics, ttl time.Duration) error {
	if acs.redis == nil {
		return nil
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal property statistics: %v", err)
	}

	return acs.redis.SetEx(ctx, key, data, ttl).Err()
}

func (acs *AnalyticsCacheService) getCachedStringSlice(ctx context.Context, key string) []string {
	if acs.redis == nil {
		return nil
	}

	val, err := acs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var slice []string
	if err := json.Unmarshal([]byte(val), &slice); err != nil {
		log.Printf("⚠️ Failed to unmarshal string slice from cache: %v", err)
		return nil
	}

	return slice
}

func (acs *AnalyticsCacheService) cacheStringSlice(ctx context.Context, key string, slice []string, ttl time.Duration) error {
	if acs.redis == nil {
		return nil
	}

	data, err := json.Marshal(slice)
	if err != nil {
		return fmt.Errorf("failed to marshal string slice: %v", err)
	}

	return acs.redis.SetEx(ctx, key, data, ttl).Err()
}

func (acs *AnalyticsCacheService) getCachedProperties(ctx context.Context, key string) []*models.Property {
	if acs.redis == nil {
		return nil
	}

	val, err := acs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var properties []*models.Property
	if err := json.Unmarshal([]byte(val), &properties); err != nil {
		log.Printf("⚠️ Failed to unmarshal properties from cache: %v", err)
		return nil
	}

	return properties
}

func (acs *AnalyticsCacheService) cacheProperties(ctx context.Context, key string, properties []*models.Property, ttl time.Duration) error {
	if acs.redis == nil {
		return nil
	}

	data, err := json.Marshal(properties)
	if err != nil {
		return fmt.Errorf("failed to marshal properties: %v", err)
	}

	return acs.redis.SetEx(ctx, key, data, ttl).Err()
}

func (acs *AnalyticsCacheService) getCachedBookingStatistics(ctx context.Context, key string) *repositories.BookingStatistics {
	if acs.redis == nil {
		return nil
	}

	val, err := acs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var stats repositories.BookingStatistics
	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		log.Printf("⚠️ Failed to unmarshal booking statistics from cache: %v", err)
		return nil
	}

	return &stats
}

func (acs *AnalyticsCacheService) cacheBookingStatistics(ctx context.Context, key string, stats *repositories.BookingStatistics, ttl time.Duration) error {
	if acs.redis == nil {
		return nil
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal booking statistics: %v", err)
	}

	return acs.redis.SetEx(ctx, key, data, ttl).Err()
}

func (acs *AnalyticsCacheService) getCachedTrends(ctx context.Context, key string) map[string]int64 {
	if acs.redis == nil {
		return nil
	}

	val, err := acs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var trends map[string]int64
	if err := json.Unmarshal([]byte(val), &trends); err != nil {
		log.Printf("⚠️ Failed to unmarshal trends from cache: %v", err)
		return nil
	}

	return trends
}

func (acs *AnalyticsCacheService) cacheTrends(ctx context.Context, key string, trends map[string]int64, ttl time.Duration) error {
	if acs.redis == nil {
		return nil
	}

	data, err := json.Marshal(trends)
	if err != nil {
		return fmt.Errorf("failed to marshal trends: %v", err)
	}

	return acs.redis.SetEx(ctx, key, data, ttl).Err()
}

func (acs *AnalyticsCacheService) getCachedDashboardMetrics(ctx context.Context, key string) map[string]interface{} {
	if acs.redis == nil {
		return nil
	}

	val, err := acs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal([]byte(val), &metrics); err != nil {
		log.Printf("⚠️ Failed to unmarshal dashboard metrics from cache: %v", err)
		return nil
	}

	return metrics
}

func (acs *AnalyticsCacheService) cacheDashboardMetrics(ctx context.Context, key string, metrics map[string]interface{}, ttl time.Duration) error {
	if acs.redis == nil {
		return nil
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard metrics: %v", err)
	}

	return acs.redis.SetEx(ctx, key, data, ttl).Err()
}

func (acs *AnalyticsCacheService) deleteFromCache(ctx context.Context, pattern string) error {
	if acs.redis == nil {
		return nil
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		keys, err := acs.redis.Keys(ctx, pattern).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			return acs.redis.Del(ctx, keys...).Err()
		}
		return nil
	}

	// Delete single key
	return acs.redis.Del(ctx, pattern).Err()
}

func (acs *AnalyticsCacheService) recordCacheHit() {
	acs.cacheStats.Hits++
	acs.cacheStats.TotalOperations++
}

func (acs *AnalyticsCacheService) recordCacheMiss() {
	acs.cacheStats.Misses++
	acs.cacheStats.TotalOperations++
}

func (acs *AnalyticsCacheService) recordCacheError() {
	acs.cacheStats.Errors++
	acs.cacheStats.TotalOperations++
}
