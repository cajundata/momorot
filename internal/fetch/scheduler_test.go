package fetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrioritizeFetchOrder_NeverFetched(t *testing.T) {
	symbols := []string{"SPY", "QQQ", "IWM"}
	lastFetchTimes := map[string]time.Time{}

	ordered := PrioritizeFetchOrder(symbols, lastFetchTimes)

	// All have same priority (never fetched), order should be stable
	assert.Len(t, ordered, 3)
	assert.Contains(t, ordered, "SPY")
	assert.Contains(t, ordered, "QQQ")
	assert.Contains(t, ordered, "IWM")
}

func TestPrioritizeFetchOrder_MixedStaleness(t *testing.T) {
	now := time.Now()
	symbols := []string{"SPY", "QQQ", "IWM", "DIA"}
	lastFetchTimes := map[string]time.Time{
		"SPY": now.Add(-48 * time.Hour), // 2 days ago (most stale)
		"QQQ": now.Add(-24 * time.Hour), // 1 day ago
		"IWM": now.Add(-1 * time.Hour),  // 1 hour ago (most recent)
		// DIA never fetched
	}

	ordered := PrioritizeFetchOrder(symbols, lastFetchTimes)

	assert.Len(t, ordered, 4)
	// DIA should be first (never fetched = zero time)
	assert.Equal(t, "DIA", ordered[0])
	// SPY should be second (oldest fetch)
	assert.Equal(t, "SPY", ordered[1])
	// QQQ should be third
	assert.Equal(t, "QQQ", ordered[2])
	// IWM should be last (most recent)
	assert.Equal(t, "IWM", ordered[3])
}

func TestPrioritizeFetchOrder_AllFetchedRecently(t *testing.T) {
	now := time.Now()
	symbols := []string{"SPY", "QQQ", "IWM"}
	lastFetchTimes := map[string]time.Time{
		"SPY": now.Add(-3 * time.Hour),
		"QQQ": now.Add(-2 * time.Hour),
		"IWM": now.Add(-1 * time.Hour),
	}

	ordered := PrioritizeFetchOrder(symbols, lastFetchTimes)

	assert.Len(t, ordered, 3)
	// SPY is oldest
	assert.Equal(t, "SPY", ordered[0])
	// QQQ is middle
	assert.Equal(t, "QQQ", ordered[1])
	// IWM is newest
	assert.Equal(t, "IWM", ordered[2])
}

func TestEstimateFetchTime_NoSymbols(t *testing.T) {
	duration := EstimateFetchTime(0, 5)
	assert.Equal(t, time.Duration(0), duration)
}

func TestEstimateFetchTime_FewSymbols(t *testing.T) {
	// 3 symbols with 5 workers should take ~1 batch (~2 seconds)
	duration := EstimateFetchTime(3, 5)
	assert.Equal(t, 2*time.Second, duration)
}

func TestEstimateFetchTime_ManySymbols(t *testing.T) {
	// 25 symbols with 5 workers should take 5 batches (~10 seconds)
	duration := EstimateFetchTime(25, 5)
	assert.Equal(t, 10*time.Second, duration)
}

func TestEstimateFetchTime_SingleWorker(t *testing.T) {
	// 10 symbols with 1 worker should take 10 batches (~20 seconds)
	duration := EstimateFetchTime(10, 1)
	assert.Equal(t, 20*time.Second, duration)
}
