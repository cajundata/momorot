package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(Config{Path: dbPath})
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)

	return db
}

func TestSymbolRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSymbolRepository(db)

	symbol := &Symbol{
		Symbol:    "SPY",
		Name:      "SPDR S&P 500 ETF",
		AssetType: "ETF",
		Active:    true,
	}

	err := repo.Create(symbol)
	require.NoError(t, err)

	// Verify it was created
	retrieved, err := repo.Get("SPY")
	require.NoError(t, err)
	assert.Equal(t, "SPY", retrieved.Symbol)
	assert.Equal(t, "SPDR S&P 500 ETF", retrieved.Name)
	assert.Equal(t, "ETF", retrieved.AssetType)
	assert.True(t, retrieved.Active)
}

func TestSymbolRepository_ListActive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSymbolRepository(db)

	// Create active and inactive symbols
	symbols := []*Symbol{
		{Symbol: "SPY", Name: "S&P 500", AssetType: "ETF", Active: true},
		{Symbol: "QQQ", Name: "Nasdaq 100", AssetType: "ETF", Active: true},
		{Symbol: "IWM", Name: "Russell 2000", AssetType: "ETF", Active: false},
	}

	for _, s := range symbols {
		require.NoError(t, repo.Create(s))
	}

	// Get active symbols
	active, err := repo.ListActive()
	require.NoError(t, err)
	assert.Len(t, active, 2)

	// Verify only active symbols are returned
	for _, s := range active {
		assert.True(t, s.Active)
	}
}

func TestSymbolRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSymbolRepository(db)

	symbol := &Symbol{
		Symbol:    "SPY",
		Name:      "S&P 500",
		AssetType: "ETF",
		Active:    true,
	}

	require.NoError(t, repo.Create(symbol))

	// Update the symbol
	symbol.Name = "SPDR S&P 500 ETF Trust"
	symbol.Active = false
	err := repo.Update(symbol)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get("SPY")
	require.NoError(t, err)
	assert.Equal(t, "SPDR S&P 500 ETF Trust", retrieved.Name)
	assert.False(t, retrieved.Active)
}

func TestPriceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create symbol first
	symRepo := NewSymbolRepository(db)
	require.NoError(t, symRepo.Create(&Symbol{
		Symbol:    "SPY",
		Name:      "S&P 500",
		AssetType: "ETF",
		Active:    true,
	}))

	priceRepo := NewPriceRepository(db)

	adjClose := 450.50
	volume := int64(50000000)
	price := &Price{
		Symbol:   "SPY",
		Date:     "2025-10-04",
		Open:     448.00,
		High:     451.00,
		Low:      447.50,
		Close:    450.00,
		AdjClose: &adjClose,
		Volume:   &volume,
	}

	err := priceRepo.Create(price)
	require.NoError(t, err)
}

func TestPriceRepository_UpsertBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create symbol first
	symRepo := NewSymbolRepository(db)
	require.NoError(t, symRepo.Create(&Symbol{
		Symbol:    "SPY",
		Name:      "S&P 500",
		AssetType: "ETF",
		Active:    true,
	}))

	priceRepo := NewPriceRepository(db)

	adjClose1 := 450.50
	volume1 := int64(50000000)
	adjClose2 := 451.75
	volume2 := int64(48000000)

	prices := []Price{
		{
			Symbol:   "SPY",
			Date:     "2025-10-04",
			Open:     448.00,
			High:     451.00,
			Low:      447.50,
			Close:    450.00,
			AdjClose: &adjClose1,
			Volume:   &volume1,
		},
		{
			Symbol:   "SPY",
			Date:     "2025-10-03",
			Open:     450.00,
			High:     452.00,
			Low:      449.50,
			Close:    451.50,
			AdjClose: &adjClose2,
			Volume:   &volume2,
		},
	}

	err := priceRepo.UpsertBatch(prices)
	require.NoError(t, err)

	// Verify both prices were inserted
	retrieved, err := priceRepo.GetRange("SPY", "2025-10-01", "2025-10-10")
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestPriceRepository_GetLatestDate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create symbol first
	symRepo := NewSymbolRepository(db)
	require.NoError(t, symRepo.Create(&Symbol{
		Symbol:    "SPY",
		Name:      "S&P 500",
		AssetType: "ETF",
		Active:    true,
	}))

	priceRepo := NewPriceRepository(db)

	// Insert prices with different dates
	prices := []Price{
		{Symbol: "SPY", Date: "2025-10-01", Open: 100, High: 101, Low: 99, Close: 100},
		{Symbol: "SPY", Date: "2025-10-02", Open: 100, High: 101, Low: 99, Close: 100},
		{Symbol: "SPY", Date: "2025-10-03", Open: 100, High: 101, Low: 99, Close: 100},
	}

	err := priceRepo.UpsertBatch(prices)
	require.NoError(t, err)

	// Get latest date
	latestDate, err := priceRepo.GetLatestDate("SPY")
	require.NoError(t, err)
	assert.Equal(t, "2025-10-03", latestDate)

	// Test for symbol with no data
	latestDate, err = priceRepo.GetLatestDate("QQQ")
	require.NoError(t, err)
	assert.Empty(t, latestDate)
}

func TestRunRepository_CreateAndFinish(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRunRepository(db)

	// Create a new run
	runID, err := repo.Create("Test run")
	require.NoError(t, err)
	assert.Greater(t, runID, int64(0))

	// Finish the run
	err = repo.Finish(runID, "OK", 10, 0)
	require.NoError(t, err)

	// Get the latest run
	run, err := repo.GetLatest()
	require.NoError(t, err)
	assert.Equal(t, runID, run.RunID)
	assert.Equal(t, "OK", run.Status)
	assert.Equal(t, 10, run.SymbolsProcessed)
	assert.Equal(t, 0, run.SymbolsFailed)
	assert.NotNil(t, run.FinishedAt)
}

func TestFetchLogRepository_Log(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	runRepo := NewRunRepository(db)
	logRepo := NewFetchLogRepository(db)

	// Create a run
	runID, err := runRepo.Create("Test run")
	require.NoError(t, err)

	// Log a successful fetch
	fromDate := "2025-01-01"
	toDate := "2025-10-04"
	entry := &FetchLog{
		RunID:    runID,
		Symbol:   "SPY",
		FromDate: &fromDate,
		ToDate:   &toDate,
		Rows:     250,
		OK:       true,
	}

	err = logRepo.Log(entry)
	require.NoError(t, err)

	// Log a failed fetch
	errMsg := "API rate limit exceeded"
	failedEntry := &FetchLog{
		RunID:   runID,
		Symbol:  "QQQ",
		Rows:    0,
		OK:      false,
		Message: &errMsg,
	}

	err = logRepo.Log(failedEntry)
	require.NoError(t, err)

	// Get failures
	failures, err := logRepo.GetFailures(runID)
	require.NoError(t, err)
	assert.Len(t, failures, 1)
	assert.Equal(t, "QQQ", failures[0].Symbol)
	assert.False(t, failures[0].OK)
}

func TestIndicatorRepository_UpsertBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create symbol and price first
	symRepo := NewSymbolRepository(db)
	require.NoError(t, symRepo.Create(&Symbol{
		Symbol:    "SPY",
		Name:      "S&P 500",
		AssetType: "ETF",
		Active:    true,
	}))

	priceRepo := NewPriceRepository(db)
	require.NoError(t, priceRepo.Create(&Price{
		Symbol: "SPY",
		Date:   "2025-10-04",
		Open:   100,
		High:   101,
		Low:    99,
		Close:  100,
	}))

	indRepo := NewIndicatorRepository(db)

	r1m := 0.05
	r3m := 0.12
	score := 0.75
	rank := 1

	indicators := []Indicator{
		{
			Symbol: "SPY",
			Date:   "2025-10-04",
			R1M:    &r1m,
			R3M:    &r3m,
			Score:  &score,
			Rank:   &rank,
		},
	}

	err := indRepo.UpsertBatch(indicators)
	require.NoError(t, err)

	// Verify insert
	top, err := indRepo.GetTopN("2025-10-04", 5)
	require.NoError(t, err)
	assert.Len(t, top, 1)
	assert.Equal(t, "SPY", top[0].Symbol)
	assert.NotNil(t, top[0].R1M)
	assert.Equal(t, 0.05, *top[0].R1M)
}
