package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	MaxLoginAttempts      = 5
	LockoutDuration       = 15 * time.Minute
	AttemptWindowDuration = 15 * time.Minute
)

type BruteForceProtection struct {
	db          *gorm.DB
	redis       *redis.Client
	auditLogger *AuditLogger
}

func NewBruteForceProtection(db *gorm.DB, redisClient *redis.Client) *BruteForceProtection {
	return &BruteForceProtection{
		db:          db,
		redis:       redisClient,
		auditLogger: NewAuditLogger(db),
	}
}

func (bfp *BruteForceProtection) CheckLoginAttempt(identifier, ipAddress string) (allowed bool, remainingAttempts int, retryAfter time.Duration, err error) {
	ctx := context.Background()

	if bfp.redis != nil {
		return bfp.checkWithRedis(ctx, identifier, ipAddress)
	}

	return bfp.checkWithDatabase(identifier, ipAddress)
}

func (bfp *BruteForceProtection) checkWithRedis(ctx context.Context, identifier, ipAddress string) (bool, int, time.Duration, error) {
	lockoutKey := fmt.Sprintf("lockout:%s", identifier)
	attemptKey := fmt.Sprintf("login_attempts:%s", identifier)

	isLocked, err := bfp.redis.Exists(ctx, lockoutKey).Result()
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to check lockout: %w", err)
	}

	if isLocked > 0 {
		ttl, _ := bfp.redis.TTL(ctx, lockoutKey).Result()
		bfp.auditLogger.LogSecurityEvent(
			"brute_force_blocked",
			nil,
			ipAddress,
			"",
			fmt.Sprintf("Login attempt blocked due to account lockout: %s", identifier),
			map[string]interface{}{
				"identifier": identifier,
				"ttl":        ttl.String(),
			},
			90,
		)
		return false, 0, ttl, nil
	}

	attempts, err := bfp.redis.Get(ctx, attemptKey).Int()
	if err != nil && err != redis.Nil {
		return false, 0, 0, fmt.Errorf("failed to get attempts: %w", err)
	}

	remaining := MaxLoginAttempts - attempts
	if remaining <= 0 {
		bfp.redis.Set(ctx, lockoutKey, "1", LockoutDuration)
		bfp.redis.Del(ctx, attemptKey)

		bfp.auditLogger.LogSecurityEvent(
			"brute_force_lockout",
			nil,
			ipAddress,
			"",
			fmt.Sprintf("Account locked due to too many failed login attempts: %s", identifier),
			map[string]interface{}{
				"identifier":      identifier,
				"attempts":        attempts,
				"lockout_duration": LockoutDuration.String(),
			},
			95,
		)

		return false, 0, LockoutDuration, nil
	}

	return true, remaining, 0, nil
}

func (bfp *BruteForceProtection) checkWithDatabase(identifier, ipAddress string) (bool, int, time.Duration, error) {
	var failedAttempts int64
	windowStart := time.Now().Add(-AttemptWindowDuration)

	bfp.db.Model(&SecurityEvent{}).
		Where("event_type = ? AND (user_id IN (SELECT id FROM admin_users WHERE username = ? OR email = ?) OR details->>'identifier' = ?) AND created_at > ?",
			"login_failure", identifier, identifier, identifier, windowStart).
		Count(&failedAttempts)

	remaining := MaxLoginAttempts - int(failedAttempts)
	if remaining <= 0 {
		bfp.auditLogger.LogSecurityEvent(
			"brute_force_lockout",
			nil,
			ipAddress,
			"",
			fmt.Sprintf("Account locked due to too many failed login attempts: %s", identifier),
			map[string]interface{}{
				"identifier":      identifier,
				"attempts":        failedAttempts,
				"lockout_duration": LockoutDuration.String(),
			},
			95,
		)

		return false, 0, LockoutDuration, nil
	}

	return true, remaining, 0, nil
}

func (bfp *BruteForceProtection) RecordFailedAttempt(identifier, ipAddress, userAgent string) error {
	ctx := context.Background()

	if bfp.redis != nil {
		attemptKey := fmt.Sprintf("login_attempts:%s", identifier)
		pipe := bfp.redis.Pipeline()
		pipe.Incr(ctx, attemptKey)
		pipe.Expire(ctx, attemptKey, AttemptWindowDuration)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to record attempt in Redis: %w", err)
		}
	}

	bfp.auditLogger.LogSecurityEvent(
		"login_failure",
		nil,
		ipAddress,
		userAgent,
		fmt.Sprintf("Failed login attempt for: %s", identifier),
		map[string]interface{}{
			"identifier": identifier,
		},
		60,
	)

	return nil
}

func (bfp *BruteForceProtection) RecordSuccessfulLogin(identifier string) error {
	ctx := context.Background()

	if bfp.redis != nil {
		attemptKey := fmt.Sprintf("login_attempts:%s", identifier)
		lockoutKey := fmt.Sprintf("lockout:%s", identifier)

		pipe := bfp.redis.Pipeline()
		pipe.Del(ctx, attemptKey)
		pipe.Del(ctx, lockoutKey)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to clear attempts in Redis: %w", err)
		}
	}

	return nil
}

func (bfp *BruteForceProtection) IsIPSuspicious(ipAddress string) (bool, error) {
	ctx := context.Background()

	if bfp.redis != nil {
		key := fmt.Sprintf("suspicious_ip:%s", ipAddress)
		exists, err := bfp.redis.Exists(ctx, key).Result()
		if err != nil {
			return false, fmt.Errorf("failed to check suspicious IP: %w", err)
		}
		return exists > 0, nil
	}

	var suspiciousEvents int64
	windowStart := time.Now().Add(-1 * time.Hour)

	bfp.db.Model(&SecurityEvent{}).
		Where("ip_address = ? AND risk_score >= 80 AND created_at > ?", ipAddress, windowStart).
		Count(&suspiciousEvents)

	return suspiciousEvents >= 3, nil
}

func (bfp *BruteForceProtection) MarkIPSuspicious(ipAddress string, duration time.Duration) error {
	ctx := context.Background()

	if bfp.redis != nil {
		key := fmt.Sprintf("suspicious_ip:%s", ipAddress)
		return bfp.redis.Set(ctx, key, "1", duration).Err()
	}

	bfp.auditLogger.LogSecurityEvent(
		"suspicious_ip_detected",
		nil,
		ipAddress,
		"",
		"IP address marked as suspicious",
		map[string]interface{}{
			"duration": duration.String(),
		},
		85,
	)

	return nil
}

func (bfp *BruteForceProtection) GetAttemptStats(identifier string) (attempts int, windowEnd time.Time, err error) {
	ctx := context.Background()

	if bfp.redis != nil {
		attemptKey := fmt.Sprintf("login_attempts:%s", identifier)
		attempts, err := bfp.redis.Get(ctx, attemptKey).Int()
		if err == redis.Nil {
			return 0, time.Time{}, nil
		}
		if err != nil {
			return 0, time.Time{}, fmt.Errorf("failed to get attempt stats: %w", err)
		}

		ttl, err := bfp.redis.TTL(ctx, attemptKey).Result()
		if err != nil {
			return attempts, time.Time{}, nil
		}

		return attempts, time.Now().Add(ttl), nil
	}

	var count int64
	windowStart := time.Now().Add(-AttemptWindowDuration)
	bfp.db.Model(&SecurityEvent{}).
		Where("event_type = ? AND details->>'identifier' = ? AND created_at > ?",
			"login_failure", identifier, windowStart).
		Count(&count)

	return int(count), time.Now().Add(AttemptWindowDuration), nil
}
