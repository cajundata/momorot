package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Symbol represents a ticker symbol in the universe
type Symbol struct {
	Symbol    string
	Name      string
	AssetType string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Price represents a single day's OHLCV data
type Price struct {
	Symbol    string
	Date      string // ISO format: yyyy-mm-dd
	Open      float64
	High      float64
	Low       float64
	Close     float64
	AdjClose  *float64 // Nullable
	Volume    *int64   // Nullable
	CreatedAt time.Time
}

// Indicator represents calculated momentum metrics for a symbol on a date
type Indicator struct {
	Symbol    string
	Date      string
	R1M       *float64 // 1-month return
	R3M       *float64 // 3-month return
	R6M       *float64 // 6-month return
	R12M      *float64 // 12-month return
	Vol3M     *float64 // 3-month volatility
	Vol6M     *float64 // 6-month volatility
	ADV       *float64 // Average dollar volume
	Score     *float64 // Composite momentum score
	Rank      *int     // Rank within universe
	CreatedAt time.Time
}

// Run represents a data refresh/computation run
type Run struct {
	RunID            int64
	StartedAt        time.Time
	FinishedAt       *time.Time
	Status           string // RUNNING, OK, ERROR
	SymbolsProcessed int
	SymbolsFailed    int
	Notes            *string
}

// FetchLog represents a log entry for a symbol fetch
type FetchLog struct {
	RunID     int64
	Symbol    string
	FromDate  *string
	ToDate    *string
	Rows      int
	OK        bool
	Message   *string
	FetchedAt time.Time
}

// SymbolRepository provides data access for symbols
type SymbolRepository struct {
	db *DB
}

// NewSymbolRepository creates a new symbol repository
func NewSymbolRepository(db *DB) *SymbolRepository {
	return &SymbolRepository{db: db}
}

// Create inserts a new symbol
func (r *SymbolRepository) Create(s *Symbol) error {
	query := `
		INSERT INTO symbols (symbol, name, asset_type, active)
		VALUES (?, ?, ?, ?)
	`
	active := 0
	if s.Active {
		active = 1
	}
	_, err := r.db.Exec(query, s.Symbol, s.Name, s.AssetType, active)
	return err
}

// Get retrieves a symbol by its ticker
func (r *SymbolRepository) Get(symbol string) (*Symbol, error) {
	query := `
		SELECT symbol, name, asset_type, active, created_at, updated_at
		FROM symbols
		WHERE symbol = ?
	`
	var s Symbol
	var active int
	var createdAt, updatedAt string
	err := r.db.QueryRow(query, symbol).Scan(
		&s.Symbol, &s.Name, &s.AssetType, &active, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.Active = active == 1
	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &s, nil
}

// ListActive returns all active symbols
func (r *SymbolRepository) ListActive() ([]Symbol, error) {
	query := `
		SELECT symbol, name, asset_type, active, created_at, updated_at
		FROM symbols
		WHERE active = 1
		ORDER BY symbol
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []Symbol
	for rows.Next() {
		var s Symbol
		var active int
		var createdAt, updatedAt string
		if err := rows.Scan(&s.Symbol, &s.Name, &s.AssetType, &active, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		s.Active = active == 1
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		symbols = append(symbols, s)
	}
	return symbols, rows.Err()
}

// Update updates a symbol's information
func (r *SymbolRepository) Update(s *Symbol) error {
	query := `
		UPDATE symbols
		SET name = ?, asset_type = ?, active = ?, updated_at = datetime('now')
		WHERE symbol = ?
	`
	active := 0
	if s.Active {
		active = 1
	}
	_, err := r.db.Exec(query, s.Name, s.AssetType, active, s.Symbol)
	return err
}

// PriceRepository provides data access for prices
type PriceRepository struct {
	db *DB
}

// NewPriceRepository creates a new price repository
func NewPriceRepository(db *DB) *PriceRepository {
	return &PriceRepository{db: db}
}

// Create inserts a new price record
func (r *PriceRepository) Create(p *Price) error {
	query := `
		INSERT INTO prices (symbol, date, open, high, low, close, adj_close, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, p.Symbol, p.Date, p.Open, p.High, p.Low, p.Close, p.AdjClose, p.Volume)
	return err
}

// UpsertBatch efficiently inserts or replaces multiple price records
func (r *PriceRepository) UpsertBatch(prices []Price) error {
	if len(prices) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO prices (symbol, date, open, high, low, close, adj_close, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, date) DO UPDATE SET
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			adj_close = excluded.adj_close,
			volume = excluded.volume
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, p := range prices {
		if _, err := stmt.Exec(p.Symbol, p.Date, p.Open, p.High, p.Low, p.Close, p.AdjClose, p.Volume); err != nil {
			return fmt.Errorf("failed to insert price for %s on %s: %w", p.Symbol, p.Date, err)
		}
	}

	return tx.Commit()
}

// GetLatestDate returns the most recent date for which we have price data for a symbol
func (r *PriceRepository) GetLatestDate(symbol string) (string, error) {
	query := `SELECT MAX(date) FROM prices WHERE symbol = ?`
	var date sql.NullString
	err := r.db.QueryRow(query, symbol).Scan(&date)
	if err != nil {
		return "", err
	}
	if !date.Valid {
		return "", nil
	}
	return date.String, nil
}

// GetRange retrieves price data for a symbol within a date range
func (r *PriceRepository) GetRange(symbol, startDate, endDate string) ([]Price, error) {
	query := `
		SELECT symbol, date, open, high, low, close, adj_close, volume, created_at
		FROM prices
		WHERE symbol = ? AND date >= ? AND date <= ?
		ORDER BY date ASC
	`
	rows, err := r.db.Query(query, symbol, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []Price
	for rows.Next() {
		var p Price
		var createdAt string
		if err := rows.Scan(&p.Symbol, &p.Date, &p.Open, &p.High, &p.Low, &p.Close, &p.AdjClose, &p.Volume, &createdAt); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

// IndicatorRepository provides data access for indicators
type IndicatorRepository struct {
	db *DB
}

// NewIndicatorRepository creates a new indicator repository
func NewIndicatorRepository(db *DB) *IndicatorRepository {
	return &IndicatorRepository{db: db}
}

// UpsertBatch efficiently inserts or replaces multiple indicator records
func (r *IndicatorRepository) UpsertBatch(indicators []Indicator) error {
	if len(indicators) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO indicators (symbol, date, r_1m, r_3m, r_6m, r_12m, vol_3m, vol_6m, adv, score, rank)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, date) DO UPDATE SET
			r_1m = excluded.r_1m,
			r_3m = excluded.r_3m,
			r_6m = excluded.r_6m,
			r_12m = excluded.r_12m,
			vol_3m = excluded.vol_3m,
			vol_6m = excluded.vol_6m,
			adv = excluded.adv,
			score = excluded.score,
			rank = excluded.rank
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, ind := range indicators {
		if _, err := stmt.Exec(ind.Symbol, ind.Date, ind.R1M, ind.R3M, ind.R6M, ind.R12M,
			ind.Vol3M, ind.Vol6M, ind.ADV, ind.Score, ind.Rank); err != nil {
			return fmt.Errorf("failed to insert indicator for %s on %s: %w", ind.Symbol, ind.Date, err)
		}
	}

	return tx.Commit()
}

// GetTopN returns the top N ranked symbols for a given date
func (r *IndicatorRepository) GetTopN(date string, n int) ([]Indicator, error) {
	query := `
		SELECT symbol, date, r_1m, r_3m, r_6m, r_12m, vol_3m, vol_6m, adv, score, rank, created_at
		FROM indicators
		WHERE date = ? AND rank IS NOT NULL
		ORDER BY rank ASC
		LIMIT ?
	`
	rows, err := r.db.Query(query, date, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanIndicators(rows)
}

// scanIndicators is a helper to scan indicator rows
func (r *IndicatorRepository) scanIndicators(rows *sql.Rows) ([]Indicator, error) {
	var indicators []Indicator
	for rows.Next() {
		var ind Indicator
		var createdAt string
		if err := rows.Scan(&ind.Symbol, &ind.Date, &ind.R1M, &ind.R3M, &ind.R6M, &ind.R12M,
			&ind.Vol3M, &ind.Vol6M, &ind.ADV, &ind.Score, &ind.Rank, &createdAt); err != nil {
			return nil, err
		}
		ind.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		indicators = append(indicators, ind)
	}
	return indicators, rows.Err()
}

// RunRepository provides data access for runs
type RunRepository struct {
	db *DB
}

// NewRunRepository creates a new run repository
func NewRunRepository(db *DB) *RunRepository {
	return &RunRepository{db: db}
}

// Create starts a new run and returns its ID
func (r *RunRepository) Create(notes string) (int64, error) {
	query := `INSERT INTO runs (status, notes) VALUES ('RUNNING', ?)`
	result, err := r.db.Exec(query, sql.NullString{String: notes, Valid: notes != ""})
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Finish marks a run as complete
func (r *RunRepository) Finish(runID int64, status string, symbolsProcessed, symbolsFailed int) error {
	query := `
		UPDATE runs
		SET finished_at = datetime('now'),
			status = ?,
			symbols_processed = ?,
			symbols_failed = ?
		WHERE run_id = ?
	`
	_, err := r.db.Exec(query, status, symbolsProcessed, symbolsFailed, runID)
	return err
}

// GetLatest returns the most recent run
func (r *RunRepository) GetLatest() (*Run, error) {
	query := `
		SELECT run_id, started_at, finished_at, status, symbols_processed, symbols_failed, notes
		FROM runs
		ORDER BY run_id DESC
		LIMIT 1
	`
	var run Run
	var startedAt, finishedAt, notes sql.NullString
	err := r.db.QueryRow(query).Scan(
		&run.RunID, &startedAt, &finishedAt, &run.Status,
		&run.SymbolsProcessed, &run.SymbolsFailed, &notes,
	)
	if err != nil {
		return nil, err
	}

	run.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt.String)
	if finishedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", finishedAt.String)
		run.FinishedAt = &t
	}
	if notes.Valid {
		run.Notes = &notes.String
	}

	return &run, nil
}

// FetchLogRepository provides data access for fetch logs
type FetchLogRepository struct {
	db *DB
}

// NewFetchLogRepository creates a new fetch log repository
func NewFetchLogRepository(db *DB) *FetchLogRepository {
	return &FetchLogRepository{db: db}
}

// Log records a fetch attempt
func (r *FetchLogRepository) Log(entry *FetchLog) error {
	query := `
		INSERT INTO fetch_log (run_id, symbol, from_dt, to_dt, rows, ok, msg)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	ok := 0
	if entry.OK {
		ok = 1
	}
	_, err := r.db.Exec(query, entry.RunID, entry.Symbol, entry.FromDate, entry.ToDate, entry.Rows, ok, entry.Message)
	return err
}

// GetFailures returns all failed fetch attempts for a run
func (r *FetchLogRepository) GetFailures(runID int64) ([]FetchLog, error) {
	query := `
		SELECT run_id, symbol, from_dt, to_dt, rows, ok, msg, fetched_at
		FROM fetch_log
		WHERE run_id = ? AND ok = 0
		ORDER BY symbol
	`
	rows, err := r.db.Query(query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []FetchLog
	for rows.Next() {
		var log FetchLog
		var ok int
		var fetchedAt string
		if err := rows.Scan(&log.RunID, &log.Symbol, &log.FromDate, &log.ToDate, &log.Rows, &ok, &log.Message, &fetchedAt); err != nil {
			return nil, err
		}
		log.OK = ok == 1
		log.FetchedAt, _ = time.Parse("2006-01-02 15:04:05", fetchedAt)
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
