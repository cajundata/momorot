package analytics

import (
	"fmt"
	"math"
	"sort"
)

// ScoringConfig contains parameters for momentum scoring.
type ScoringConfig struct {
	PenaltyLambda      float64 // Volatility penalty factor (default: 0.35)
	MinADV             float64 // Minimum average dollar volume threshold
	BreadthMinPositive int     // Minimum number of positive lookbacks required
	BreadthTotal       int     // Total number of lookbacks to check
}

// Scorer computes composite momentum scores and rankings.
type Scorer struct {
	config ScoringConfig
}

// NewScorer creates a new momentum scorer with the given configuration.
func NewScorer(config ScoringConfig) *Scorer {
	return &Scorer{
		config: config,
	}
}

// ComputeScore calculates the composite momentum score for a symbol.
// Formula: score = normalized_momentum - λ·volatility
// Where normalized_momentum is the z-score normalized average of multi-horizon returns.
func ComputeScore(indicators *Indicators, penaltyLambda float64) float64 {
	// Calculate average return across all horizons
	avgReturn := (indicators.R1M + indicators.R3M + indicators.R6M + indicators.R12M) / 4.0

	// Apply volatility penalty (using 6M volatility as primary metric)
	score := avgReturn - (penaltyLambda * indicators.Vol6M)

	return score
}

// ZScoreNormalize normalizes a slice of values using z-score normalization.
// Returns the normalized values: (x - mean) / stddev
func ZScoreNormalize(values []float64) ([]float64, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("cannot normalize empty slice")
	}

	if len(values) == 1 {
		return []float64{0.0}, nil
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	stdDev := math.Sqrt(variance)

	// Avoid division by zero
	if stdDev == 0 {
		// All values are identical, return zeros
		result := make([]float64, len(values))
		return result, nil
	}

	// Normalize
	normalized := make([]float64, len(values))
	for i, v := range values {
		normalized[i] = (v - mean) / stdDev
	}

	return normalized, nil
}

// ScoreAndRank computes scores and ranks for all symbols in the universe.
// Returns ranked symbols with deterministic tie-breaking.
func (s *Scorer) ScoreAndRank(indicatorsList []*Indicators) ([]*SymbolScore, error) {
	if len(indicatorsList) == 0 {
		return nil, fmt.Errorf("no indicators provided")
	}

	// Filter by breadth and liquidity requirements
	filtered := make([]*Indicators, 0, len(indicatorsList))
	for _, ind := range indicatorsList {
		// Check breadth filter
		returns := []float64{ind.R1M, ind.R3M, ind.R6M, ind.R12M}
		if !CheckBreadthFilter(returns, s.config.BreadthMinPositive) {
			continue
		}

		// Check minimum ADV
		if ind.ADV < s.config.MinADV {
			continue
		}

		filtered = append(filtered, ind)
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no symbols passed filtering criteria")
	}

	// Calculate raw scores for all symbols
	scores := make([]float64, len(filtered))
	for i, ind := range filtered {
		scores[i] = ComputeScore(ind, s.config.PenaltyLambda)
	}

	// Z-score normalize the scores across the universe
	normalizedScores, err := ZScoreNormalize(scores)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize scores: %w", err)
	}

	// Create symbol scores
	symbolScores := make([]*SymbolScore, len(filtered))
	for i, ind := range filtered {
		symbolScores[i] = &SymbolScore{
			Symbol:     ind.Symbol,
			Score:      normalizedScores[i],
			Volatility: ind.Vol6M,
			Liquidity:  ind.ADV,
			Indicators: *ind,
		}
	}

	// Sort with deterministic tie-breaking
	sort.SliceStable(symbolScores, func(i, j int) bool {
		return compareSymbolScores(symbolScores[i], symbolScores[j])
	})

	// Assign ranks (1-indexed, 1 = best)
	for i, ss := range symbolScores {
		symbolScores[i].Indicators.Rank = i + 1
		symbolScores[i].Indicators.Score = ss.Score
	}

	return symbolScores, nil
}

// compareSymbolScores implements deterministic tie-breaking.
// Primary: Higher score wins
// Tie-break 1: Lower volatility wins
// Tie-break 2: Higher liquidity (ADV) wins
// Tie-break 3: Alphabetical by symbol (for absolute determinism)
func compareSymbolScores(a, b *SymbolScore) bool {
	// Primary: Higher score wins (return true if a > b)
	if math.Abs(a.Score-b.Score) > 1e-10 {
		return a.Score > b.Score
	}

	// Tie-break 1: Lower volatility wins (return true if a < b)
	if math.Abs(a.Volatility-b.Volatility) > 1e-10 {
		return a.Volatility < b.Volatility
	}

	// Tie-break 2: Higher liquidity wins (return true if a > b)
	if math.Abs(a.Liquidity-b.Liquidity) > 1e-10 {
		return a.Liquidity > b.Liquidity
	}

	// Tie-break 3: Alphabetical by symbol (for absolute determinism)
	return a.Symbol < b.Symbol
}

// ApplyFilters filters symbols based on breadth and liquidity requirements.
func (s *Scorer) ApplyFilters(indicatorsList []*Indicators) []*Indicators {
	filtered := make([]*Indicators, 0, len(indicatorsList))

	for _, ind := range indicatorsList {
		// Check breadth filter
		returns := []float64{ind.R1M, ind.R3M, ind.R6M, ind.R12M}
		if !CheckBreadthFilter(returns, s.config.BreadthMinPositive) {
			continue
		}

		// Check minimum ADV
		if ind.ADV < s.config.MinADV {
			continue
		}

		filtered = append(filtered, ind)
	}

	return filtered
}

// GetTopN returns the top N symbols from the ranked list.
func GetTopN(rankedSymbols []*SymbolScore, n int) []*SymbolScore {
	if n > len(rankedSymbols) {
		n = len(rankedSymbols)
	}
	return rankedSymbols[:n]
}
