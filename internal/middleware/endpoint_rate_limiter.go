package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

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

// CheckOnly checks rate limit status without recording a new request
func (erl *EndpointRateLimiter) CheckOnly(clientIP string) (blocked bool, retryAfter int64) {
	erl.mutex.RLock()
	defer erl.mutex.RUnlock()

	now := time.Now()
	client, exists := erl.clients[clientIP]
	if !exists {
		return false, 0
	}

	if client.blocked && now.Before(client.blockUntil) {
		return true, client.blockUntil.Unix() - now.Unix()
	}

	return false, 0
}

// RecordRequest records a request for rate limiting purposes
func (erl *EndpointRateLimiter) RecordRequest(clientIP string) (blocked bool, retryAfter int64) {
	return erl.checkRateLimit(clientIP)
}

// FormatRetryTime returns a human-readable string for the retry duration
func FormatRetryTime(seconds int64) string {
	if seconds >= 3600 {
		hours := seconds / 3600
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if seconds >= 60 {
		minutes := seconds / 60
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	if seconds == 1 {
		return "1 second"
	}
	return fmt.Sprintf("%d seconds", seconds)
}

// Predefined rate limiters for common use cases
var (
	// For booking and contact form submissions
	BookingRateLimiter = NewEndpointRateLimiter(5, 20, 15*time.Minute) // 5 per minute, 20 per hour, 15min block

	// For admin login attempts
	AdminLoginRateLimiter = NewEndpointRateLimiter(3, 10, 30*time.Minute) // 3 per minute, 10 per hour, 30min block

	// For public API endpoints
	PublicAPIRateLimiter = NewEndpointRateLimiter(10, 50, 5*time.Minute) // 10 per minute, 50 per hour, 5min block
)
