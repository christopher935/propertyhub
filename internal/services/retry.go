package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableStatus []int
}

var DefaultRetryConfig = RetryConfig{
	MaxRetries:      3,
	InitialDelay:    time.Second,
	MaxDelay:        30 * time.Second,
	BackoffFactor:   2.0,
	RetryableStatus: []int{429, 500, 502, 503, 504},
}

func WithRetry(ctx context.Context, config RetryConfig, operation func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		resp, lastErr = operation()

		if lastErr == nil && resp != nil {
			if !isRetryableStatus(resp.StatusCode, config.RetryableStatus) {
				return resp, nil
			}

			if resp.StatusCode == 429 {
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, err := strconv.Atoi(retryAfter); err == nil {
						time.Sleep(time.Duration(seconds) * time.Second)
						continue
					}
				}
			}
		}

		if attempt < config.MaxRetries {
			delay := calculateBackoff(attempt, config)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return resp, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt))
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	jitter := delay * 0.1 * (rand.Float64()*2 - 1)
	return time.Duration(delay + jitter)
}

func isRetryableStatus(status int, retryable []int) bool {
	for _, s := range retryable {
		if s == status {
			return true
		}
	}
	return false
}
