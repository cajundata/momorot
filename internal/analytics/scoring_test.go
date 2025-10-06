package analytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeScore(t *testing.T) {
	ind := &Indicators{
		R1M:   0.05,  // 5%
		R3M:   0.10,  // 10%
		R6M:   0.15,  // 15%
		R12M:  0.20,  // 20%
		Vol6M: 0.25,  // 25% volatility
	}

	score := ComputeScore(ind, 0.35)

	// Average return: (0.05 + 0.10 + 0.15 + 0.20) / 4 = 0.125
	// Score: 0.125 - (0.35 * 0.25) = 0.125 - 0.0875 = 0.0375
	expected := 0.0375
	assert.InDelta(t, expected, score, 0.0001)
}

func TestComputeScore_HighVolatility(t *testing.T) {
	ind := &Indicators{
		R1M:   0.10,
		R3M:   0.10,
		R6M:   0.10,
		R12M:  0.10,
		Vol6M: 0.50, // High volatility
	}

	score := ComputeScore(ind, 0.35)

	// Average return: 0.10
	// Score: 0.10 - (0.35 * 0.50) = 0.10 - 0.175 = -0.075
	expected := -0.075
	assert.InDelta(t, expected, score, 0.0001)
}

func TestZScoreNormalize_Simple(t *testing.T) {
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	normalized, err := ZScoreNormalize(values)

	require.NoError(t, err)
	assert.Len(t, normalized, 5)

	// Mean should be 3.0, stddev should be sqrt(2) ≈ 1.414
	// Normalized values should be approximately: -1.414, -0.707, 0, 0.707, 1.414
	assert.InDelta(t, -1.414, normalized[0], 0.01)
	assert.InDelta(t, -0.707, normalized[1], 0.01)
	assert.InDelta(t, 0.0, normalized[2], 0.01)
	assert.InDelta(t, 0.707, normalized[3], 0.01)
	assert.InDelta(t, 1.414, normalized[4], 0.01)
}

func TestZScoreNormalize_AllSame(t *testing.T) {
	values := []float64{5.0, 5.0, 5.0, 5.0}

	normalized, err := ZScoreNormalize(values)

	require.NoError(t, err)
	// When all values are identical, should return zeros
	for _, v := range normalized {
		assert.Equal(t, 0.0, v)
	}
}

func TestZScoreNormalize_Single(t *testing.T) {
	values := []float64{5.0}

	normalized, err := ZScoreNormalize(values)

	require.NoError(t, err)
	assert.Len(t, normalized, 1)
	assert.Equal(t, 0.0, normalized[0])
}

func TestZScoreNormalize_Empty(t *testing.T) {
	values := []float64{}

	_, err := ZScoreNormalize(values)

	assert.Error(t, err)
}

func TestNewScorer(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             5000000,
		BreadthMinPositive: 3,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	assert.NotNil(t, scorer)
	assert.Equal(t, config, scorer.config)
}

func TestScoreAndRank_DeterministicOrdering(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             1000000,
		BreadthMinPositive: 2,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	// Create three symbols with different characteristics
	indicators := []*Indicators{
		{
			Symbol: "HIGH_SCORE",
			R1M:    0.15,
			R3M:    0.20,
			R6M:    0.25,
			R12M:   0.30,
			Vol6M:  0.10,
			ADV:    10000000,
		},
		{
			Symbol: "MED_SCORE",
			R1M:    0.10,
			R3M:    0.12,
			R6M:    0.15,
			R12M:   0.18,
			Vol6M:  0.15,
			ADV:    8000000,
		},
		{
			Symbol: "LOW_SCORE",
			R1M:    0.05,
			R3M:    0.08,
			R6M:    0.10,
			R12M:   0.12,
			Vol6M:  0.20,
			ADV:    5000000,
		},
	}

	ranked, err := scorer.ScoreAndRank(indicators)

	require.NoError(t, err)
	assert.Len(t, ranked, 3)

	// Verify ordering: HIGH_SCORE should be rank 1
	assert.Equal(t, "HIGH_SCORE", ranked[0].Symbol)
	assert.Equal(t, 1, ranked[0].Indicators.Rank)

	assert.Equal(t, "MED_SCORE", ranked[1].Symbol)
	assert.Equal(t, 2, ranked[1].Indicators.Rank)

	assert.Equal(t, "LOW_SCORE", ranked[2].Symbol)
	assert.Equal(t, 3, ranked[2].Indicators.Rank)
}

func TestScoreAndRank_TieBreaking_Volatility(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.0, // No penalty, so scores will be identical
		MinADV:             1000000,
		BreadthMinPositive: 2,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	// Create symbols with identical returns but different volatility
	indicators := []*Indicators{
		{
			Symbol: "HIGH_VOL",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.30, // Higher volatility
			ADV:    5000000,
		},
		{
			Symbol: "LOW_VOL",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.10, // Lower volatility
			ADV:    5000000,
		},
	}

	ranked, err := scorer.ScoreAndRank(indicators)

	require.NoError(t, err)
	// LOW_VOL should rank first (lower volatility wins tie)
	assert.Equal(t, "LOW_VOL", ranked[0].Symbol)
	assert.Equal(t, "HIGH_VOL", ranked[1].Symbol)
}

func TestScoreAndRank_TieBreaking_Liquidity(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.0, // No penalty
		MinADV:             1000000,
		BreadthMinPositive: 2,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	// Create symbols with identical returns and volatility but different ADV
	indicators := []*Indicators{
		{
			Symbol: "LOW_LIQ",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.20,
			ADV:    3000000, // Lower liquidity
		},
		{
			Symbol: "HIGH_LIQ",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.20,
			ADV:    10000000, // Higher liquidity
		},
	}

	ranked, err := scorer.ScoreAndRank(indicators)

	require.NoError(t, err)
	// HIGH_LIQ should rank first (higher liquidity wins tie)
	assert.Equal(t, "HIGH_LIQ", ranked[0].Symbol)
	assert.Equal(t, "LOW_LIQ", ranked[1].Symbol)
}

func TestScoreAndRank_TieBreaking_Alphabetical(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.0,
		MinADV:             1000000,
		BreadthMinPositive: 2,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	// Create symbols with identical everything except symbol name
	indicators := []*Indicators{
		{
			Symbol: "ZZZ",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.20,
			ADV:    5000000,
		},
		{
			Symbol: "AAA",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.20,
			ADV:    5000000,
		},
	}

	ranked, err := scorer.ScoreAndRank(indicators)

	require.NoError(t, err)
	// AAA should rank first (alphabetical tie-breaking)
	assert.Equal(t, "AAA", ranked[0].Symbol)
	assert.Equal(t, "ZZZ", ranked[1].Symbol)
}

func TestScoreAndRank_BreadthFilter(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             1000000,
		BreadthMinPositive: 3, // Require 3 of 4 positive
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	indicators := []*Indicators{
		{
			Symbol: "PASS",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   -0.05, // Only 1 negative, passes 3-of-4
			Vol6M:  0.20,
			ADV:    5000000,
		},
		{
			Symbol: "FAIL",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    -0.05,
			R12M:   -0.05, // 2 negatives, fails 3-of-4
			Vol6M:  0.20,
			ADV:    5000000,
		},
	}

	ranked, err := scorer.ScoreAndRank(indicators)

	require.NoError(t, err)
	assert.Len(t, ranked, 1, "Only 1 symbol should pass breadth filter")
	assert.Equal(t, "PASS", ranked[0].Symbol)
}

func TestScoreAndRank_MinADVFilter(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             5000000, // $5M minimum
		BreadthMinPositive: 2,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	indicators := []*Indicators{
		{
			Symbol: "PASS",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.20,
			ADV:    6000000, // Passes
		},
		{
			Symbol: "FAIL",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.20,
			ADV:    3000000, // Fails
		},
	}

	ranked, err := scorer.ScoreAndRank(indicators)

	require.NoError(t, err)
	assert.Len(t, ranked, 1, "Only 1 symbol should pass ADV filter")
	assert.Equal(t, "PASS", ranked[0].Symbol)
}

func TestScoreAndRank_EmptyInput(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             5000000,
		BreadthMinPositive: 3,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	_, err := scorer.ScoreAndRank([]*Indicators{})

	assert.Error(t, err)
}

func TestScoreAndRank_AllFiltered(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             100000000, // Unreasonably high
		BreadthMinPositive: 3,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	indicators := []*Indicators{
		{
			Symbol: "TEST",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			Vol6M:  0.20,
			ADV:    5000000, // Too low
		},
	}

	_, err := scorer.ScoreAndRank(indicators)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no symbols passed filtering")
}

func TestApplyFilters(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             5000000,
		BreadthMinPositive: 3,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	indicators := []*Indicators{
		{
			Symbol: "PASS",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			ADV:    10000000,
		},
		{
			Symbol: "FAIL_BREADTH",
			R1M:    -0.05,
			R3M:    -0.05,
			R6M:    0.10,
			R12M:   0.10,
			ADV:    10000000,
		},
		{
			Symbol: "FAIL_ADV",
			R1M:    0.10,
			R3M:    0.10,
			R6M:    0.10,
			R12M:   0.10,
			ADV:    1000000, // Too low
		},
	}

	filtered := scorer.ApplyFilters(indicators)

	assert.Len(t, filtered, 1)
	assert.Equal(t, "PASS", filtered[0].Symbol)
}

func TestGetTopN(t *testing.T) {
	ranked := []*SymbolScore{
		{Symbol: "RANK1", Indicators: Indicators{Rank: 1}},
		{Symbol: "RANK2", Indicators: Indicators{Rank: 2}},
		{Symbol: "RANK3", Indicators: Indicators{Rank: 3}},
		{Symbol: "RANK4", Indicators: Indicators{Rank: 4}},
		{Symbol: "RANK5", Indicators: Indicators{Rank: 5}},
	}

	top3 := GetTopN(ranked, 3)

	assert.Len(t, top3, 3)
	assert.Equal(t, "RANK1", top3[0].Symbol)
	assert.Equal(t, "RANK2", top3[1].Symbol)
	assert.Equal(t, "RANK3", top3[2].Symbol)
}

func TestGetTopN_RequestMoreThanAvailable(t *testing.T) {
	ranked := []*SymbolScore{
		{Symbol: "RANK1"},
		{Symbol: "RANK2"},
	}

	top5 := GetTopN(ranked, 5)

	// Should return all available (2)
	assert.Len(t, top5, 2)
}

// Determinism test: running twice on same data produces identical results
func TestScoreAndRank_Deterministic(t *testing.T) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             1000000,
		BreadthMinPositive: 2,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	indicators := []*Indicators{
		{Symbol: "SPY", R1M: 0.10, R3M: 0.12, R6M: 0.15, R12M: 0.20, Vol6M: 0.18, ADV: 10000000},
		{Symbol: "QQQ", R1M: 0.12, R3M: 0.15, R6M: 0.18, R12M: 0.22, Vol6M: 0.25, ADV: 12000000},
		{Symbol: "IWM", R1M: 0.08, R3M: 0.10, R6M: 0.12, R12M: 0.15, Vol6M: 0.20, ADV: 8000000},
	}

	// Run twice
	ranked1, err1 := scorer.ScoreAndRank(indicators)
	require.NoError(t, err1)

	ranked2, err2 := scorer.ScoreAndRank(indicators)
	require.NoError(t, err2)

	// Results should be identical
	require.Equal(t, len(ranked1), len(ranked2))
	for i := range ranked1 {
		assert.Equal(t, ranked1[i].Symbol, ranked2[i].Symbol, "Order should be identical")
		assert.Equal(t, ranked1[i].Indicators.Rank, ranked2[i].Indicators.Rank, "Ranks should be identical")
		assert.InDelta(t, ranked1[i].Score, ranked2[i].Score, 1e-10, "Scores should be identical")
	}
}

// Benchmark
func BenchmarkScoreAndRank(b *testing.B) {
	config := ScoringConfig{
		PenaltyLambda:      0.35,
		MinADV:             5000000,
		BreadthMinPositive: 3,
		BreadthTotal:       4,
	}

	scorer := NewScorer(config)

	// Create 25 symbols (typical universe size)
	indicators := make([]*Indicators, 25)
	for i := 0; i < 25; i++ {
		indicators[i] = &Indicators{
			Symbol: string(rune('A' + i)),
			Date:   time.Now(),
			R1M:    0.05 + float64(i)*0.01,
			R3M:    0.10 + float64(i)*0.01,
			R6M:    0.15 + float64(i)*0.01,
			R12M:   0.20 + float64(i)*0.01,
			Vol6M:  0.20 - float64(i)*0.005,
			ADV:    5000000 + float64(i)*1000000,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scorer.ScoreAndRank(indicators)
	}
}
