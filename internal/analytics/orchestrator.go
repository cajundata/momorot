package analytics

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cajundata/momorot/internal/db"
)

// Orchestrator coordinates indicator computation and ranking across the universe.
type Orchestrator struct {
	database     *db.DB
	symbolRepo   *db.SymbolRepository
	priceRepo    *db.PriceRepository
	indicatorRepo *db.IndicatorRepository
	calculator   *IndicatorCalculator
	scorer       *Scorer
}

// NewOrchestrator creates a new analytics orchestrator.
func NewOrchestrator(database *db.DB, lookbacks, volWindows map[string]int, scoringConfig ScoringConfig) *Orchestrator {
	return &Orchestrator{
		database:      database,
		symbolRepo:    db.NewSymbolRepository(database),
		priceRepo:     db.NewPriceRepository(database),
		indicatorRepo: db.NewIndicatorRepository(database),
		calculator:    NewIndicatorCalculator(lookbacks, volWindows),
		scorer:        NewScorer(scoringConfig),
	}
}

// ComputeAllIndicators computes indicators for all active symbols and stores them in the database.
// Returns the number of symbols processed and any error encountered.
func (o *Orchestrator) ComputeAllIndicators(asOfDate time.Time) (int, error) {
	// Get list of active symbols
	symbolRecords, err := o.symbolRepo.ListActive()
	if err != nil {
		return 0, fmt.Errorf("failed to list active symbols: %w", err)
	}

	// Extract symbol strings
	symbols := make([]string, len(symbolRecords))
	for i, sr := range symbolRecords {
		symbols[i] = sr.Symbol
	}

	if len(symbols) == 0 {
		return 0, fmt.Errorf("no active symbols found")
	}

	indicatorsList := make([]*Indicators, 0, len(symbols))
	processedCount := 0

	// Compute indicators for each symbol
	for _, symbol := range symbols {
		// Fetch price data for this symbol
		prices, err := o.fetchPricesForSymbol(symbol)
		if err != nil {
			// Log error but continue with other symbols
			continue
		}

		// Compute indicators
		indicators, err := o.calculator.ComputeIndicators(symbol, prices)
		if err != nil {
			// Log error but continue with other symbols
			continue
		}

		indicatorsList = append(indicatorsList, indicators)
		processedCount++
	}

	if len(indicatorsList) == 0 {
		return 0, fmt.Errorf("no indicators could be computed")
	}

	// Score and rank all symbols
	rankedSymbols, err := o.scorer.ScoreAndRank(indicatorsList)
	if err != nil {
		return processedCount, fmt.Errorf("failed to score and rank: %w", err)
	}

	// Update indicators with scores and ranks
	for _, rs := range rankedSymbols {
		rs.Indicators.Score = rs.Score
		rs.Indicators.Rank = rs.Indicators.Rank
	}

	// Persist indicators to database
	indicatorsToSave := make([]db.Indicator, 0, len(rankedSymbols))
	for _, rs := range rankedSymbols {
		r1m := rs.Indicators.R1M
		r3m := rs.Indicators.R3M
		r6m := rs.Indicators.R6M
		r12m := rs.Indicators.R12M
		vol3m := rs.Indicators.Vol3M
		vol6m := rs.Indicators.Vol6M
		adv := rs.Indicators.ADV
		score := rs.Score
		rank := rs.Indicators.Rank

		indicatorsToSave = append(indicatorsToSave, db.Indicator{
			Symbol: rs.Symbol,
			Date:   rs.Indicators.Date.Format("2006-01-02"),
			R1M:    &r1m,
			R3M:    &r3m,
			R6M:    &r6m,
			R12M:   &r12m,
			Vol3M:  &vol3m,
			Vol6M:  &vol6m,
			ADV:    &adv,
			Score:  &score,
			Rank:   &rank,
		})
	}

	err = o.indicatorRepo.UpsertBatch(indicatorsToSave)
	if err != nil {
		return processedCount, fmt.Errorf("failed to save indicators: %w", err)
	}

	return processedCount, nil
}

// fetchPricesForSymbol retrieves price data for a symbol from the database.
func (o *Orchestrator) fetchPricesForSymbol(symbol string) ([]PriceBar, error) {
	// Query prices from database
	query := `
		SELECT date, open, high, low, close, adj_close, volume
		FROM prices
		WHERE symbol = ?
		ORDER BY date ASC
	`

	rows, err := o.database.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to query prices for %s: %w", symbol, err)
	}
	defer rows.Close()

	var prices []PriceBar
	for rows.Next() {
		var dateStr string
		var adjCloseNullable sql.NullFloat64
		var volumeNullable sql.NullInt64

		pb := PriceBar{}
		err := rows.Scan(
			&dateStr,
			&pb.Open,
			&pb.High,
			&pb.Low,
			&pb.Close,
			&adjCloseNullable,
			&volumeNullable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price row: %w", err)
		}

		// Parse date
		pb.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
		}

		// Handle nullable fields
		if adjCloseNullable.Valid {
			pb.AdjClose = adjCloseNullable.Float64
		} else {
			// If adj_close is null, use close
			pb.AdjClose = pb.Close
		}

		if volumeNullable.Valid {
			pb.Volume = float64(volumeNullable.Int64)
		}

		prices = append(prices, pb)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no price data found for %s", symbol)
	}

	return prices, nil
}

// GetTopRanked retrieves the top N ranked symbols from the database for a given date.
func (o *Orchestrator) GetTopRanked(asOfDate time.Time, n int) ([]*SymbolScore, error) {
	dateStr := asOfDate.Format("2006-01-02")

	query := `
		SELECT symbol, r_1m, r_3m, r_6m, r_12m, vol_3m, vol_6m, adv, score, rank
		FROM indicators
		WHERE date = ?
		ORDER BY rank ASC
		LIMIT ?
	`

	rows, err := o.database.Query(query, dateStr, n)
	if err != nil {
		return nil, fmt.Errorf("failed to query top ranked symbols: %w", err)
	}
	defer rows.Close()

	var results []*SymbolScore
	for rows.Next() {
		var symbol string
		var r1m, r3m, r6m, r12m, vol3m, vol6m, adv, score sql.NullFloat64
		var rank sql.NullInt64

		err := rows.Scan(&symbol, &r1m, &r3m, &r6m, &r12m, &vol3m, &vol6m, &adv, &score, &rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan indicator row: %w", err)
		}

		ss := &SymbolScore{
			Symbol:     symbol,
			Score:      getFloat(score),
			Volatility: getFloat(vol6m),
			Liquidity:  getFloat(adv),
			Indicators: Indicators{
				Symbol: symbol,
				Date:   asOfDate,
				R1M:    getFloat(r1m),
				R3M:    getFloat(r3m),
				R6M:    getFloat(r6m),
				R12M:   getFloat(r12m),
				Vol3M:  getFloat(vol3m),
				Vol6M:  getFloat(vol6m),
				ADV:    getFloat(adv),
				Score:  getFloat(score),
				Rank:   int(getInt(rank)),
			},
		}

		results = append(results, ss)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating indicator rows: %w", err)
	}

	return results, nil
}

// Helper functions to extract values from sql.Null types
func getFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0
}

func getInt(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}
