package analytics

import "time"

// PriceBar represents a single day's OHLCV data.
type PriceBar struct {
	Date      time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	AdjClose  float64
	Volume    float64
}

// Indicators represents calculated momentum indicators for a symbol on a specific date.
type Indicators struct {
	Symbol string
	Date   time.Time
	R1M    float64 // 1-month return (21 days)
	R3M    float64 // 3-month return (63 days)
	R6M    float64 // 6-month return (126 days)
	R12M   float64 // 12-month return (252 days)
	Vol3M  float64 // 3-month volatility (annualized)
	Vol6M  float64 // 6-month volatility (annualized)
	ADV    float64 // Average dollar volume
	Score  float64 // Composite momentum score
	Rank   int     // Rank within universe (1 = best)
}

// SymbolScore represents a symbol's composite score and related metrics for ranking.
type SymbolScore struct {
	Symbol     string
	Score      float64
	Volatility float64
	Liquidity  float64 // ADV for tie-breaking
	Indicators Indicators
}
