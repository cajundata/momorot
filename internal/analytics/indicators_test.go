package analytics

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateReturns_Simple(t *testing.T) {
	// Create simple price series: 100, 105, 110, 115, 120
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AdjClose: 100},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), AdjClose: 105},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), AdjClose: 110},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), AdjClose: 115},
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), AdjClose: 120},
	}

	lookbacks := map[string]int{
		"r1m":  1, // 1 day lookback
		"r3m":  2, // 2 days lookback
		"r6m":  3, // 3 days lookback
		"r12m": 4, // 4 days lookback
	}

	r1m, r3m, r6m, r12m, err := CalculateReturns(prices, lookbacks)

	require.NoError(t, err)
	// r1m: (120 / 115) - 1 = 0.0435 (4.35%)
	assert.InDelta(t, 0.0435, r1m, 0.0001)
	// r3m: (120 / 110) - 1 = 0.0909 (9.09%)
	assert.InDelta(t, 0.0909, r3m, 0.0001)
	// r6m: (120 / 105) - 1 = 0.1429 (14.29%)
	assert.InDelta(t, 0.1429, r6m, 0.0001)
	// r12m: (120 / 100) - 1 = 0.20 (20%)
	assert.InDelta(t, 0.20, r12m, 0.0001)
}

func TestCalculateReturns_InsufficientData(t *testing.T) {
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AdjClose: 100},
	}

	lookbacks := map[string]int{
		"r1m":  1,
		"r3m":  3,
		"r6m":  6,
		"r12m": 12,
	}

	_, _, _, _, err := CalculateReturns(prices, lookbacks)
	assert.Error(t, err)
}

func TestCalculateReturns_UnsortedData(t *testing.T) {
	// Provide unsorted data - function should handle it
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), AdjClose: 120},
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AdjClose: 100},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), AdjClose: 110},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), AdjClose: 105},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), AdjClose: 115},
	}

	lookbacks := map[string]int{
		"r1m":  1,
		"r3m":  2,
		"r6m":  3,
		"r12m": 4,
	}

	r1m, r3m, r6m, r12m, err := CalculateReturns(prices, lookbacks)

	require.NoError(t, err)
	assert.InDelta(t, 0.0435, r1m, 0.0001)
	assert.InDelta(t, 0.0909, r3m, 0.0001)
	assert.InDelta(t, 0.1429, r6m, 0.0001)
	assert.InDelta(t, 0.20, r12m, 0.0001)
}

func TestCalculateVolatility_Simple(t *testing.T) {
	// Create price series with known volatility
	// Using constant 1% daily return: 100, 101, 102.01, 103.0301, etc.
	// This should give very low volatility
	prices := make([]PriceBar, 65)
	price := 100.0
	for i := 0; i < 65; i++ {
		prices[i] = PriceBar{
			Date:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			AdjClose: price,
		}
		price *= 1.01 // 1% daily increase
	}

	vol, err := CalculateVolatility(prices, 63)

	require.NoError(t, err)
	// With constant returns, volatility should be near zero
	assert.Less(t, vol, 0.01, "Constant returns should have low volatility")
}

func TestCalculateVolatility_HighVolatility(t *testing.T) {
	// Create price series with alternating +5% and -5%
	prices := make([]PriceBar, 65)
	price := 100.0
	for i := 0; i < 65; i++ {
		prices[i] = PriceBar{
			Date:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			AdjClose: price,
		}
		if i%2 == 0 {
			price *= 1.05
		} else {
			price *= 0.95
		}
	}

	vol, err := CalculateVolatility(prices, 63)

	require.NoError(t, err)
	// With alternating returns, volatility should be higher
	assert.Greater(t, vol, 0.3, "Alternating returns should have high volatility")
}

func TestCalculateVolatility_InsufficientData(t *testing.T) {
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AdjClose: 100},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), AdjClose: 105},
	}

	_, err := CalculateVolatility(prices, 63)
	assert.Error(t, err)
}

func TestCalculateVolatility_ZeroPrice(t *testing.T) {
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AdjClose: 100},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), AdjClose: 0}, // Zero price
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), AdjClose: 105},
	}

	_, err := CalculateVolatility(prices, 2)
	assert.Error(t, err)
}

func TestCalculateADV_Simple(t *testing.T) {
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Close: 100, Volume: 1000000},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Close: 105, Volume: 1200000},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Close: 110, Volume: 1500000},
	}

	adv, err := CalculateADV(prices, 3)

	require.NoError(t, err)
	// ADV = average of (100*1M, 105*1.2M, 110*1.5M)
	// = average of (100M, 126M, 165M) = 130.33M
	expected := (100*1000000 + 105*1200000 + 110*1500000) / 3.0
	assert.InDelta(t, expected, adv, 0.01)
}

func TestCalculateADV_InsufficientData(t *testing.T) {
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Close: 100, Volume: 1000000},
	}

	_, err := CalculateADV(prices, 63)
	assert.Error(t, err)
}

func TestCheckBreadthFilter_AllPositive(t *testing.T) {
	returns := []float64{0.05, 0.10, 0.15, 0.20}
	result := CheckBreadthFilter(returns, 3)
	assert.True(t, result, "All 4 positive should pass 3-of-4 filter")
}

func TestCheckBreadthFilter_ExactlyMinPositive(t *testing.T) {
	returns := []float64{0.05, 0.10, 0.15, -0.05}
	result := CheckBreadthFilter(returns, 3)
	assert.True(t, result, "Exactly 3 positive should pass 3-of-4 filter")
}

func TestCheckBreadthFilter_BelowMinPositive(t *testing.T) {
	returns := []float64{0.05, 0.10, -0.05, -0.10}
	result := CheckBreadthFilter(returns, 3)
	assert.False(t, result, "Only 2 positive should fail 3-of-4 filter")
}

func TestCheckBreadthFilter_AllNegative(t *testing.T) {
	returns := []float64{-0.05, -0.10, -0.15, -0.20}
	result := CheckBreadthFilter(returns, 1)
	assert.False(t, result, "All negative should fail any positive requirement")
}

func TestCheckBreadthFilter_Empty(t *testing.T) {
	returns := []float64{}
	result := CheckBreadthFilter(returns, 1)
	assert.False(t, result, "Empty returns should fail")
}

func TestNewIndicatorCalculator(t *testing.T) {
	lookbacks := map[string]int{"r1m": 21, "r3m": 63, "r6m": 126, "r12m": 252}
	volWindows := map[string]int{"short": 63, "long": 126}

	calc := NewIndicatorCalculator(lookbacks, volWindows)

	assert.NotNil(t, calc)
	assert.Equal(t, lookbacks, calc.lookbacks)
	assert.Equal(t, volWindows, calc.volWindows)
}

func TestComputeIndicators_Success(t *testing.T) {
	lookbacks := map[string]int{"r1m": 5, "r3m": 10, "r6m": 15, "r12m": 20}
	volWindows := map[string]int{"short": 10, "long": 15}

	calc := NewIndicatorCalculator(lookbacks, volWindows)

	// Generate 30 days of price data with steady 1% daily growth
	prices := make([]PriceBar, 30)
	price := 100.0
	for i := 0; i < 30; i++ {
		prices[i] = PriceBar{
			Date:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			Open:     price,
			High:     price * 1.02,
			Low:      price * 0.98,
			Close:    price * 1.01,
			AdjClose: price,
			Volume:   1000000,
		}
		price *= 1.01
	}

	indicators, err := calc.ComputeIndicators("TEST", prices)

	require.NoError(t, err)
	assert.NotNil(t, indicators)
	assert.Equal(t, "TEST", indicators.Symbol)
	assert.False(t, indicators.Date.IsZero())
	// Returns should be positive (growth trend)
	assert.Greater(t, indicators.R1M, 0.0)
	assert.Greater(t, indicators.R3M, 0.0)
	assert.Greater(t, indicators.R6M, 0.0)
	assert.Greater(t, indicators.R12M, 0.0)
	// Volatility should be low (steady growth)
	assert.Greater(t, indicators.Vol3M, 0.0)
	assert.Greater(t, indicators.Vol6M, 0.0)
	// ADV should be positive
	assert.Greater(t, indicators.ADV, 0.0)
}

func TestComputeIndicators_InsufficientData(t *testing.T) {
	lookbacks := map[string]int{"r1m": 21, "r3m": 63, "r6m": 126, "r12m": 252}
	volWindows := map[string]int{"short": 63, "long": 126}

	calc := NewIndicatorCalculator(lookbacks, volWindows)

	// Only 10 days of data (insufficient)
	prices := make([]PriceBar, 10)
	for i := 0; i < 10; i++ {
		prices[i] = PriceBar{
			Date:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			AdjClose: 100.0,
		}
	}

	_, err := calc.ComputeIndicators("TEST", prices)
	assert.Error(t, err)
}

func TestComputeIndicators_EmptyData(t *testing.T) {
	lookbacks := map[string]int{"r1m": 21, "r3m": 63, "r6m": 126, "r12m": 252}
	volWindows := map[string]int{"short": 63, "long": 126}

	calc := NewIndicatorCalculator(lookbacks, volWindows)

	_, err := calc.ComputeIndicators("TEST", []PriceBar{})
	assert.Error(t, err)
}

// Golden vector test with known inputs and outputs
func TestCalculateReturns_GoldenVector(t *testing.T) {
	// Precise test with golden vector
	prices := []PriceBar{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AdjClose: 100.00},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), AdjClose: 102.00},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), AdjClose: 104.04},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), AdjClose: 106.1208},
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), AdjClose: 108.243216},
	}

	lookbacks := map[string]int{
		"r1m":  1,
		"r3m":  2,
		"r6m":  3,
		"r12m": 4,
	}

	r1m, r3m, r6m, r12m, err := CalculateReturns(prices, lookbacks)

	require.NoError(t, err)
	// r1m: (108.243216 / 106.1208) - 1 = 0.02 (2%)
	assert.InDelta(t, 0.02, r1m, 0.0001)
	// r3m: (108.243216 / 104.04) - 1 = 0.0404 (4.04%)
	assert.InDelta(t, 0.0404, r3m, 0.0001)
	// r6m: (108.243216 / 102.00) - 1 = 0.0612 (6.12%)
	assert.InDelta(t, 0.0612, r6m, 0.0001)
	// r12m: (108.243216 / 100.00) - 1 = 0.08243216 (8.24%)
	assert.InDelta(t, 0.08243216, r12m, 0.0001)
}

// Benchmark tests
func BenchmarkCalculateReturns(b *testing.B) {
	prices := make([]PriceBar, 252)
	for i := 0; i < 252; i++ {
		prices[i] = PriceBar{
			Date:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			AdjClose: 100.0 * math.Pow(1.001, float64(i)),
		}
	}

	lookbacks := map[string]int{"r1m": 21, "r3m": 63, "r6m": 126, "r12m": 252}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateReturns(prices, lookbacks)
	}
}

func BenchmarkCalculateVolatility(b *testing.B) {
	prices := make([]PriceBar, 127)
	for i := 0; i < 127; i++ {
		prices[i] = PriceBar{
			Date:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			AdjClose: 100.0 * math.Pow(1.001, float64(i)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateVolatility(prices, 126)
	}
}

func BenchmarkCalculateADV(b *testing.B) {
	prices := make([]PriceBar, 63)
	for i := 0; i < 63; i++ {
		prices[i] = PriceBar{
			Date:   time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			Close:  100.0,
			Volume: 1000000,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateADV(prices, 63)
	}
}
