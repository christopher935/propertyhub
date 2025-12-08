package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type IntelligenceCacheService struct {
	redis *redis.Client
	ctx   context.Context
}

func NewIntelligenceCacheService(redisClient *redis.Client) *IntelligenceCacheService {
	return &IntelligenceCacheService{
		redis: redisClient,
		ctx:   context.Background(),
	}
}

const (
	KeyDashboardHot      = "intelligence:dashboard:hot:v1"
	KeyDashboardWarm     = "intelligence:dashboard:warm:v1"
	KeyDashboardDaily    = "intelligence:dashboard:daily:v1"
	KeyLeadsHot          = "intelligence:leads:hot:v1"
	KeyLeadsWarm         = "intelligence:leads:warm:v1"
	KeyPropertiesHot     = "intelligence:properties:hot:v1"
	KeyPropertiesWarm    = "intelligence:properties:warm:v1"
	KeyCommunicationsHot = "intelligence:communications:hot:v1"
	KeyWorkflowHot       = "intelligence:workflow:hot:v1"
	KeyTeamHot           = "intelligence:team:hot:v1"
	KeySystemHot         = "intelligence:system:hot:v1"
	TTLHot               = 5 * time.Minute
	TTLWarm              = 1 * time.Hour
	TTLDaily             = 24 * time.Hour
)

func (ics *IntelligenceCacheService) GetDashboardHot() (map[string]interface{}, error) {
	return ics.get(KeyDashboardHot)
}

func (ics *IntelligenceCacheService) SetDashboardHot(data map[string]interface{}) error {
	return ics.set(KeyDashboardHot, data, TTLHot)
}

func (ics *IntelligenceCacheService) GetDashboardWarm() (map[string]interface{}, error) {
	return ics.get(KeyDashboardWarm)
}

func (ics *IntelligenceCacheService) SetDashboardWarm(data map[string]interface{}) error {
	return ics.set(KeyDashboardWarm, data, TTLWarm)
}

func (ics *IntelligenceCacheService) GetDashboardDaily() (map[string]interface{}, error) {
	return ics.get(KeyDashboardDaily)
}

func (ics *IntelligenceCacheService) SetDashboardDaily(data map[string]interface{}) error {
	return ics.set(KeyDashboardDaily, data, TTLDaily)
}

func (ics *IntelligenceCacheService) GetLeadsHot() (map[string]interface{}, error) {
	return ics.get(KeyLeadsHot)
}

func (ics *IntelligenceCacheService) SetLeadsHot(data map[string]interface{}) error {
	return ics.set(KeyLeadsHot, data, TTLHot)
}

func (ics *IntelligenceCacheService) GetLeadsWarm() (map[string]interface{}, error) {
	return ics.get(KeyLeadsWarm)
}

func (ics *IntelligenceCacheService) SetLeadsWarm(data map[string]interface{}) error {
	return ics.set(KeyLeadsWarm, data, TTLWarm)
}

func (ics *IntelligenceCacheService) GetPropertiesHot() (map[string]interface{}, error) {
	return ics.get(KeyPropertiesHot)
}

func (ics *IntelligenceCacheService) SetPropertiesHot(data map[string]interface{}) error {
	return ics.set(KeyPropertiesHot, data, TTLHot)
}

func (ics *IntelligenceCacheService) GetPropertiesWarm() (map[string]interface{}, error) {
	return ics.get(KeyPropertiesWarm)
}

func (ics *IntelligenceCacheService) SetPropertiesWarm(data map[string]interface{}) error {
	return ics.set(KeyPropertiesWarm, data, TTLWarm)
}

func (ics *IntelligenceCacheService) InvalidateAll() error {
	keys := []string{
		KeyDashboardHot, KeyDashboardWarm, KeyDashboardDaily,
		KeyLeadsHot, KeyLeadsWarm, KeyPropertiesHot, KeyPropertiesWarm,
		KeyCommunicationsHot, KeyWorkflowHot, KeyTeamHot, KeySystemHot,
	}

	for _, key := range keys {
		if err := ics.redis.Del(ics.ctx, key).Err(); err != nil {
			log.Printf("⚠️ Failed to delete cache key %s: %v", key, err)
		}
	}

	log.Println("🗑️ Intelligence cache invalidated")
	return nil
}

func (ics *IntelligenceCacheService) get(key string) (map[string]interface{}, error) {
	if ics.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	val, err := ics.redis.Get(ics.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("cache miss")
	}
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (ics *IntelligenceCacheService) set(key string, data map[string]interface{}, ttl time.Duration) error {
	if ics.redis == nil {
		return fmt.Errorf("redis not available")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ics.redis.Set(ics.ctx, key, jsonData, ttl).Err()
}

func (ics *IntelligenceCacheService) GetOrCompute(
	key string,
	ttl time.Duration,
	compute func() (map[string]interface{}, error),
) (map[string]interface{}, error) {
	cached, err := ics.get(key)
	if err == nil {
		log.Printf("🎯 Cache HIT: %s", key)
		return cached, nil
	}

	log.Printf("❌ Cache MISS: %s - Computing...", key)

	data, err := compute()
	if err != nil {
		return nil, err
	}

	if err := ics.set(key, data, ttl); err != nil {
		log.Printf("⚠️ Failed to cache %s: %v", key, err)
	} else {
		log.Printf("✅ Cached: %s (TTL: %v)", key, ttl)
	}

	return data, nil
}

func (ics *IntelligenceCacheService) IsAvailable() bool {
	if ics.redis == nil {
		return false
	}

	err := ics.redis.Ping(ics.ctx).Err()
	return err == nil
}
