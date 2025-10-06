package analytics

import (
	"fmt"
	"math"
	"sort"
)

// IndicatorCalculator computes momentum indicators from price data.
type IndicatorCalculator struct {
	lookbacks  map[string]int // r1m, r3m, r6m, r12m in trading days
	volWindows map[string]int // short, long volatility windows
}

// NewIndicatorCalculator creates a new indicator calculator with specified lookback periods.
func NewIndicatorCalculator(lookbacks, volWindows map[string]int) *IndicatorCalculator {
	return &IndicatorCalculator{
		lookbacks:  lookbacks,
		volWindows: volWindows,
	}
}

// CalculateReturns computes multi-horizon total returns using adjusted close.
// Returns are calculated as: (price_t / price_t-n) - 1
func CalculateReturns(prices []PriceBar, lookbacks map[string]int) (r1m, r3m, r6m, r12m float64, err error) {
	if len(prices) == 0 {
		return 0, 0, 0, 0, fmt.Errorf("no price data provided")
	}

	// Sort prices by date (ascending) to ensure correct ordering
	sortedPrices := make([]PriceBar, len(prices))
	copy(sortedPrices, prices)
	sort.Slice(sortedPrices, func(i, j int) bool {
		return sortedPrices[i].Date.Before(sortedPrices[j].Date)
	})

	// Current price is the last (most recent) price
	currentPrice := sortedPrices[len(sortedPrices)-1].AdjClose

	// Calculate returns for each lookback period
	r1m, err = calculateReturn(sortedPrices, currentPrice, lookbacks["r1m"])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to calculate R1M: %w", err)
	}

	r3m, err = calculateReturn(sortedPrices, currentPrice, lookbacks["r3m"])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to calculate R3M: %w", err)
	}

	r6m, err = calculateReturn(sortedPrices, currentPrice, lookbacks["r6m"])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to calculate R6M: %w", err)
	}

	r12m, err = calculateReturn(sortedPrices, currentPrice, lookbacks["r12m"])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to calculate R12M: %w", err)
	}

	return r1m, r3m, r6m, r12m, nil
}

// calculateReturn computes the return over a specific lookback period.
func calculateReturn(prices []PriceBar, currentPrice float64, lookback int) (float64, error) {
	if len(prices) < lookback+1 {
		return 0, fmt.Errorf("insufficient data: need %d bars, have %d", lookback+1, len(prices))
	}

	// Get price from 'lookback' periods ago
	idx := len(prices) - 1 - lookback
	if idx < 0 {
		return 0, fmt.Errorf("invalid lookback index")
	}

	pastPrice := prices[idx].AdjClose
	if pastPrice == 0 {
		return 0, fmt.Errorf("past price is zero, cannot calculate return")
	}

	// Total return: (current / past) - 1
	return (currentPrice / pastPrice) - 1, nil
}

// CalculateVolatility computes annualized rolling volatility of daily log returns.
// Formula: σ_annual = σ_daily * sqrt(252)
func CalculateVolatility(prices []PriceBar, window int) (float64, error) {
	if len(prices) < window+1 {
		return 0, fmt.Errorf("insufficient data: need %d bars, have %d", window+1, len(prices))
	}

	// Sort prices by date (ascending)
	sortedPrices := make([]PriceBar, len(prices))
	copy(sortedPrices, prices)
	sort.Slice(sortedPrices, func(i, j int) bool {
		return sortedPrices[i].Date.Before(sortedPrices[j].Date)
	})

	// Calculate daily log returns over the window
	// We need 'window' returns, which requires 'window + 1' prices
	startIdx := len(sortedPrices) - window - 1
	logReturns := make([]float64, 0, window)

	for i := startIdx + 1; i < len(sortedPrices); i++ {
		if sortedPrices[i-1].AdjClose == 0 || sortedPrices[i].AdjClose == 0 {
			return 0, fmt.Errorf("zero price encountered, cannot calculate log return")
		}

		logReturn := math.Log(sortedPrices[i].AdjClose / sortedPrices[i-1].AdjClose)
		logReturns = append(logReturns, logReturn)
	}

	// Calculate standard deviation of log returns
	mean := 0.0
	for _, lr := range logReturns {
		mean += lr
	}
	mean /= float64(len(logReturns))

	variance := 0.0
	for _, lr := range logReturns {
		diff := lr - mean
		variance += diff * diff
	}
	variance /= float64(len(logReturns))

	dailyVol := math.Sqrt(variance)

	// Annualize: multiply by sqrt(252 trading days)
	annualizedVol := dailyVol * math.Sqrt(252)

	return annualizedVol, nil
}

// CalculateADV computes average dollar volume over a rolling window.
// ADV = average of (close * volume) over the window.
func CalculateADV(prices []PriceBar, window int) (float64, error) {
	if len(prices) < window {
		return 0, fmt.Errorf("insufficient data: need %d bars, have %d", window, len(prices))
	}

	// Sort prices by date (ascending)
	sortedPrices := make([]PriceBar, len(prices))
	copy(sortedPrices, prices)
	sort.Slice(sortedPrices, func(i, j int) bool {
		return sortedPrices[i].Date.Before(sortedPrices[j].Date)
	})

	// Calculate dollar volume for the most recent 'window' days
	startIdx := len(sortedPrices) - window
	sum := 0.0

	for i := startIdx; i < len(sortedPrices); i++ {
		dollarVolume := sortedPrices[i].Close * sortedPrices[i].Volume
		sum += dollarVolume
	}

	return sum / float64(window), nil
}

// CheckBreadthFilter checks if a symbol passes the breadth filter.
// Returns true if at least minPositive out of the provided returns are positive.
func CheckBreadthFilter(returns []float64, minPositive int) bool {
	if len(returns) == 0 {
		return false
	}

	positiveCount := 0
	for _, r := range returns {
		if r > 0 {
			positiveCount++
		}
	}

	return positiveCount >= minPositive
}

// ComputeIndicators calculates all indicators for a symbol given its price history.
func (ic *IndicatorCalculator) ComputeIndicators(symbol string, prices []PriceBar) (*Indicators, error) {
	if len(prices) == 0 {
		return nil, fmt.Errorf("no price data for symbol %s", symbol)
	}

	// Sort prices to ensure latest date is used
	sortedPrices := make([]PriceBar, len(prices))
	copy(sortedPrices, prices)
	sort.Slice(sortedPrices, func(i, j int) bool {
		return sortedPrices[i].Date.Before(sortedPrices[j].Date)
	})

	latestDate := sortedPrices[len(sortedPrices)-1].Date

	// Calculate returns
	r1m, r3m, r6m, r12m, err := CalculateReturns(sortedPrices, ic.lookbacks)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate returns for %s: %w", symbol, err)
	}

	// Calculate volatility (3M and 6M)
	vol3m, err := CalculateVolatility(sortedPrices, ic.volWindows["short"])
	if err != nil {
		return nil, fmt.Errorf("failed to calculate 3M volatility for %s: %w", symbol, err)
	}

	vol6m, err := CalculateVolatility(sortedPrices, ic.volWindows["long"])
	if err != nil {
		return nil, fmt.Errorf("failed to calculate 6M volatility for %s: %w", symbol, err)
	}

	// Calculate ADV (using 63-day window by default)
	adv, err := CalculateADV(sortedPrices, ic.volWindows["short"])
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ADV for %s: %w", symbol, err)
	}

	return &Indicators{
		Symbol:   symbol,
		Date:     latestDate,
		R1M:      r1m,
		R3M:      r3m,
		R6M:      r6m,
		R12M:     r12m,
		Vol3M:    vol3m,
		Vol6M:    vol6m,
		ADV:      adv,
		Score:    0, // Will be computed by scoring module
		Rank:     0, // Will be assigned by ranking algorithm
	}, nil
}
