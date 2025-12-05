package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitTier defines different rate limit tiers for different endpoint types
type RateLimitTier string

const (
	TierPublic    RateLimitTier = "public"    // Public endpoints (property listings, search)
	TierAuth      RateLimitTier = "auth"       // Authenticated user endpoints
	TierSensitive RateLimitTier = "sensitive"  // Login, password reset, MFA
	TierAdmin     RateLimitTier = "admin"      // Admin operations
	TierAPI       RateLimitTier = "api"        // External API endpoints
)

// RateLimitConfig defines tiered rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	BurstSize         int
	BlockDuration     time.Duration
}

// EndpointRateLimiter provides per-endpoint rate limiting
type EndpointRateLimiter struct {
	clients map[string]*ClientRateLimit
	mutex   sync.RWMutex

	// Configuration per endpoint
	requestsPerMinute int
	requestsPerHour   int
	blockDuration     time.Duration
}

// ClientRateLimit tracks rate limiting for a specific client
type ClientRateLimit struct {
	minuteRequests []time.Time
	hourRequests   []time.Time
	blocked        bool
	blockUntil     time.Time
	lastRequest    time.Time
}

// NewEndpointRateLimiter creates a rate limiter for specific endpoints
func NewEndpointRateLimiter(requestsPerMinute, requestsPerHour int, blockDuration time.Duration) *EndpointRateLimiter {
	limiter := &EndpointRateLimiter{
		clients:           make(map[string]*ClientRateLimit),
		requestsPerMinute: requestsPerMinute,
		requestsPerHour:   requestsPerHour,
		blockDuration:     blockDuration,
	}

	// Start cleanup routine
	go limiter.cleanup()

	return limiter
}

// RateLimit returns a Gin middleware function for rate limiting
func (erl *EndpointRateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if blocked, remaining := erl.checkRateLimit(clientIP); blocked {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d per minute, %d per hour", erl.requestsPerMinute, erl.requestsPerHour))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", remaining))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": remaining,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit validates if a client can make a request
func (erl *EndpointRateLimiter) checkRateLimit(clientIP string) (blocked bool, retryAfter int64) {
	erl.mutex.Lock()
	defer erl.mutex.Unlock()

	now := time.Now()

	// Get or create client rate limit
	client, exists := erl.clients[clientIP]
	if !exists {
		client = &ClientRateLimit{
			minuteRequests: []time.Time{},
			hourRequests:   []time.Time{},
			blocked:        false,
		}
		erl.clients[clientIP] = client
	}

	// Check if client is still blocked
	if client.blocked && now.Before(client.blockUntil) {
		return true, client.blockUntil.Unix() - now.Unix()
	}

	// Unblock if block period has passed
	if client.blocked && now.After(client.blockUntil) {
		client.blocked = false
		client.minuteRequests = []time.Time{}
		client.hourRequests = []time.Time{}
	}

	// Clean old requests
	client.minuteRequests = erl.filterRecentRequests(client.minuteRequests, now.Add(-time.Minute))
	client.hourRequests = erl.filterRecentRequests(client.hourRequests, now.Add(-time.Hour))

	// Check minute limit
	if len(client.minuteRequests) >= erl.requestsPerMinute {
		client.blocked = true
		client.blockUntil = now.Add(erl.blockDuration)
		return true, client.blockUntil.Unix() - now.Unix()
	}

	// Check hour limit
	if len(client.hourRequests) >= erl.requestsPerHour {
		client.blocked = true
		client.blockUntil = now.Add(erl.blockDuration)
		return true, client.blockUntil.Unix() - now.Unix()
	}

	// Record this request
	client.minuteRequests = append(client.minuteRequests, now)
	client.hourRequests = append(client.hourRequests, now)
	client.lastRequest = now

	return false, 0
}

// filterRecentRequests removes requests older than the cutoff time
func (erl *EndpointRateLimiter) filterRecentRequests(requests []time.Time, cutoff time.Time) []time.Time {
	filtered := []time.Time{}
	for _, req := range requests {
		if req.After(cutoff) {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// cleanup removes old client records to prevent memory leaks
func (erl *EndpointRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		erl.mutex.Lock()

		cutoff := time.Now().Add(-2 * time.Hour)
		for clientIP, client := range erl.clients {
			if client.lastRequest.Before(cutoff) {
				delete(erl.clients, clientIP)
			}
		}

		erl.mutex.Unlock()
	}
}

// GetTierConfig returns rate limit configuration for a specific tier
func GetTierConfig(tier RateLimitTier) RateLimitConfig {
	switch tier {
	case TierPublic:
		// Public endpoints - generous limits for anonymous users
		return RateLimitConfig{
			RequestsPerMinute: 100,
			RequestsPerHour:   1000,
			BurstSize:         20,
			BlockDuration:     5 * time.Minute,
		}
	case TierAuth:
		// Authenticated users - higher limits
		return RateLimitConfig{
			RequestsPerMinute: 300,
			RequestsPerHour:   3000,
			BurstSize:         50,
			BlockDuration:     5 * time.Minute,
		}
	case TierSensitive:
		// Login, password reset, MFA - very strict
		return RateLimitConfig{
			RequestsPerMinute: 5,
			RequestsPerHour:   20,
			BurstSize:         2,
			BlockDuration:     30 * time.Minute,
		}
	case TierAdmin:
		// Admin operations - strict but usable
		return RateLimitConfig{
			RequestsPerMinute: 50,
			RequestsPerHour:   500,
			BurstSize:         10,
			BlockDuration:     15 * time.Minute,
		}
	case TierAPI:
		// External API endpoints - moderate limits
		return RateLimitConfig{
			RequestsPerMinute: 60,
			RequestsPerHour:   600,
			BurstSize:         15,
			BlockDuration:     10 * time.Minute,
		}
	default:
		// Default to public tier
		return GetTierConfig(TierPublic)
	}
}

// NewEndpointRateLimiterWithTier creates a rate limiter with a specific tier
func NewEndpointRateLimiterWithTier(tier RateLimitTier) *EndpointRateLimiter {
	config := GetTierConfig(tier)
	return NewEndpointRateLimiter(config.RequestsPerMinute, config.RequestsPerHour, config.BlockDuration)
}

// Predefined rate limiters for common use cases (legacy compatibility)
var (
	// For booking and contact form submissions
	BookingRateLimiter = NewEndpointRateLimiterWithTier(TierSensitive)

	// For admin login attempts
	AdminLoginRateLimiter = NewEndpointRateLimiterWithTier(TierSensitive)

	// For public API endpoints
	PublicAPIRateLimiter = NewEndpointRateLimiterWithTier(TierPublic)

	// For authenticated user endpoints
	AuthenticatedRateLimiter = NewEndpointRateLimiterWithTier(TierAuth)

	// For admin dashboard operations
	AdminRateLimiter = NewEndpointRateLimiterWithTier(TierAdmin)

	// For external API integrations
	ExternalAPIRateLimiter = NewEndpointRateLimiterWithTier(TierAPI)
)
