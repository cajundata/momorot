package fetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(25)

	assert.NotNil(t, rl)
	assert.Equal(t, 25, rl.dailyLimit)
	assert.Equal(t, 0, rl.requestCount)
}

func TestRateLimiter_Wait_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(25)

	// Should not block when under limit
	err := rl.Wait()
	assert.NoError(t, err)
}

func TestRateLimiter_Wait_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(5)

	// Use up all requests
	for i := 0; i < 5; i++ {
		err := rl.Wait()
		require.NoError(t, err)
		rl.RecordRequest()
	}

	// Next request should fail
	err := rl.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "daily API quota exceeded")
	assert.Contains(t, err.Error(), "5/5 requests used")
}

func TestRateLimiter_RecordRequest(t *testing.T) {
	rl := NewRateLimiter(25)

	status := rl.GetStatus()
	assert.Equal(t, 0, status.RequestsUsed)
	assert.Equal(t, 25, status.RequestsLeft)

	rl.RecordRequest()

	status = rl.GetStatus()
	assert.Equal(t, 1, status.RequestsUsed)
	assert.Equal(t, 24, status.RequestsLeft)
}

func TestRateLimiter_GetStatus(t *testing.T) {
	rl := NewRateLimiter(25)

	// Record some requests
	rl.RecordRequest()
	rl.RecordRequest()
	rl.RecordRequest()

	status := rl.GetStatus()
	assert.Equal(t, 25, status.DailyLimit)
	assert.Equal(t, 3, status.RequestsUsed)
	assert.Equal(t, 22, status.RequestsLeft)
	assert.False(t, status.LastResetTime.IsZero())
	assert.False(t, status.NextResetTime.IsZero())
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(25)

	// Use some quota
	for i := 0; i < 10; i++ {
		rl.RecordRequest()
	}

	status := rl.GetStatus()
	assert.Equal(t, 10, status.RequestsUsed)

	// Reset
	rl.Reset()

	status = rl.GetStatus()
	assert.Equal(t, 0, status.RequestsUsed)
	assert.Equal(t, 25, status.RequestsLeft)
}

func TestRateLimiter_AutoReset(t *testing.T) {
	rl := NewRateLimiter(5)

	// Set last reset time to 25 hours ago
	rl.mu.Lock()
	rl.lastResetTime = time.Now().Add(-25 * time.Hour)
	rl.requestCount = 5
	rl.mu.Unlock()

	// Wait should trigger auto-reset
	err := rl.Wait()
	assert.NoError(t, err)

	status := rl.GetStatus()
	assert.Equal(t, 0, status.RequestsUsed)
}

func TestRateLimiter_FriendlyErrorMessage(t *testing.T) {
	rl := NewRateLimiter(3)

	// Exhaust quota
	for i := 0; i < 3; i++ {
		rl.Wait()
		rl.RecordRequest()
	}

	err := rl.Wait()
	require.Error(t, err)

	// Check that error message contains helpful suggestions
	assert.Contains(t, err.Error(), "daily API quota exceeded")
	assert.Contains(t, err.Error(), "3/3 requests used")
	assert.Contains(t, err.Error(), "Quota resets in")
	assert.Contains(t, err.Error(), "waiting for reset")
	assert.Contains(t, err.Error(), "CSV import")
	assert.Contains(t, err.Error(), "paid API plan")
}
