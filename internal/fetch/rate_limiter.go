package fetch

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter manages API request quotas with daily limits.
type RateLimiter struct {
	mu            sync.Mutex
	dailyLimit    int
	requestCount  int
	lastResetTime time.Time
}

// RateLimiterStatus provides current status of the rate limiter.
type RateLimiterStatus struct {
	DailyLimit    int
	RequestsUsed  int
	RequestsLeft  int
	LastResetTime time.Time
	NextResetTime time.Time
}

// NewRateLimiter creates a new rate limiter with the specified daily limit.
func NewRateLimiter(dailyLimit int) *RateLimiter {
	return &RateLimiter{
		dailyLimit:    dailyLimit,
		requestCount:  0,
		lastResetTime: time.Now(),
	}
}

// Wait blocks if the rate limit has been exceeded, returning an error with a friendly message.
func (rl *RateLimiter) Wait() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if we need to reset (new day)
	if time.Since(rl.lastResetTime) >= 24*time.Hour {
		rl.reset()
	}

	// Check if we've exceeded the limit
	if rl.requestCount >= rl.dailyLimit {
		timeUntilReset := 24*time.Hour - time.Since(rl.lastResetTime)
		hoursUntilReset := int(timeUntilReset.Hours())
		minutesUntilReset := int(timeUntilReset.Minutes()) % 60

		return fmt.Errorf(
			"daily API quota exceeded (%d/%d requests used). Quota resets in %dh%dm. "+
				"Consider: (1) waiting for reset, (2) using CSV import for historical data, "+
				"or (3) upgrading to a paid API plan",
			rl.requestCount,
			rl.dailyLimit,
			hoursUntilReset,
			minutesUntilReset,
		)
	}

	return nil
}

// RecordRequest records that a request was made.
func (rl *RateLimiter) RecordRequest() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.requestCount++
}

// GetStatus returns the current status of the rate limiter.
func (rl *RateLimiter) GetStatus() *RateLimiterStatus {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	requestsLeft := rl.dailyLimit - rl.requestCount
	if requestsLeft < 0 {
		requestsLeft = 0
	}

	nextReset := rl.lastResetTime.Add(24 * time.Hour)

	return &RateLimiterStatus{
		DailyLimit:    rl.dailyLimit,
		RequestsUsed:  rl.requestCount,
		RequestsLeft:  requestsLeft,
		LastResetTime: rl.lastResetTime,
		NextResetTime: nextReset,
	}
}

// Reset manually resets the rate limiter.
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.reset()
}

// reset is the internal reset implementation (must be called with lock held).
func (rl *RateLimiter) reset() {
	rl.requestCount = 0
	rl.lastResetTime = time.Now()
}
