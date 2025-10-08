package screens

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.New(db.Config{Path: dbPath})
	require.NoError(t, err)

	// Run migrations
	err = database.Migrate()
	require.NoError(t, err)

	t.Cleanup(func() {
		database.Close()
	})

	return database
}

func TestNewDashboard(t *testing.T) {
	database := setupTestDB(t)

	model := NewDashboard(database, 80, 24)

	assert.NotNil(t, model.database)
	assert.Equal(t, 80, model.width)
	assert.Equal(t, 24, model.height)
	assert.Equal(t, 25, model.apiQuotaLimit)
	assert.False(t, model.ready)
	assert.Nil(t, model.err)
}

func TestDashboardInit(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)

	cmd := model.Init()
	assert.NotNil(t, cmd)

	// Execute the command to verify it returns a message
	msg := cmd()
	assert.NotNil(t, msg)

	// Should return either dashboardDataMsg or dashboardErrorMsg
	switch msg.(type) {
	case dashboardDataMsg:
		// Success case
	case dashboardErrorMsg:
		// Error case (expected if DB is empty)
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}
}

func TestDashboardUpdateWindowSize(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)

	updated, cmd := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	assert.Nil(t, cmd)
	assert.Equal(t, 100, updated.width)
	assert.Equal(t, 30, updated.height)
}

func TestDashboardUpdateWithDataMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)

	// Create test data
	testRun := &db.Run{
		RunID:            1,
		StartedAt:        time.Now(),
		FinishedAt:       nil,
		Status:           "OK",
		SymbolsProcessed: 10,
		SymbolsFailed:    2,
	}

	dataMsg := dashboardDataMsg{
		run:            testRun,
		totalSymbols:   25,
		activeSymbols:  20,
		lastFetchDate:  "2025-10-08",
		apiQuotaUsed:   10,
		nextResetTime:  time.Now().Add(24 * time.Hour),
	}

	updated, cmd := model.Update(dataMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.Nil(t, updated.err)
	assert.Equal(t, testRun, updated.latestRun)
	assert.Equal(t, 25, updated.totalSymbols)
	assert.Equal(t, 20, updated.activeSymbols)
	assert.Equal(t, "2025-10-08", updated.lastFetchDate)
	assert.Equal(t, 10, updated.apiQuotaUsed)
}

func TestDashboardUpdateWithErrorMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)

	errMsg := dashboardErrorMsg{err: fmt.Errorf("test error")}

	updated, cmd := model.Update(errMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.NotNil(t, updated.err)
	assert.Equal(t, "test error", updated.err.Error())
}

func TestDashboardViewNotReady(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)

	view := model.View()

	assert.Contains(t, view, "Loading dashboard")
}

func TestDashboardViewWithError(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.err = fmt.Errorf("test error")
	model.ready = true

	view := model.View()

	assert.Contains(t, view, "Error loading dashboard")
	assert.Contains(t, view, "test error")
}

func TestDashboardViewWithData(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)

	// Set up test data
	finishedTime := time.Now()
	model.latestRun = &db.Run{
		RunID:            1,
		StartedAt:        time.Now().Add(-1 * time.Hour),
		FinishedAt:       &finishedTime,
		Status:           "OK",
		SymbolsProcessed: 10,
		SymbolsFailed:    2,
	}
	model.totalSymbols = 25
	model.activeSymbols = 20
	model.lastFetchDate = "2025-10-08"
	model.apiQuotaUsed = 10
	model.ready = true

	view := model.View()

	// Check that all cards are rendered
	assert.Contains(t, view, "Last Run")
	assert.Contains(t, view, "Universe")
	assert.Contains(t, view, "Cache Health")
	assert.Contains(t, view, "API Quota")

	// Check specific values
	assert.Contains(t, view, "OK")
	assert.Contains(t, view, "20") // active symbols
	assert.Contains(t, view, "25") // total symbols
	assert.Contains(t, view, "2025-10-08")
	assert.Contains(t, view, "10/25") // quota
}

func TestDashboardRenderLastRunCard_NoRun(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.ready = true
	model.latestRun = nil

	card := model.renderLastRunCard()

	assert.Contains(t, card, "Last Run")
	assert.Contains(t, card, "No runs yet")
}

func TestDashboardRenderLastRunCard_OKStatus(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.ready = true

	finishedTime := time.Now()
	model.latestRun = &db.Run{
		RunID:            5,
		StartedAt:        time.Date(2025, 10, 8, 12, 30, 0, 0, time.UTC),
		FinishedAt:       &finishedTime,
		Status:           "OK",
		SymbolsProcessed: 20,
		SymbolsFailed:    0,
	}

	card := model.renderLastRunCard()

	assert.Contains(t, card, "Last Run")
	assert.Contains(t, card, "OK")
	assert.Contains(t, card, "Run #5")
	assert.Contains(t, card, "20/20")
}

func TestDashboardRenderLastRunCard_ErrorStatus(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.ready = true

	model.latestRun = &db.Run{
		RunID:            3,
		StartedAt:        time.Now(),
		Status:           "ERROR",
		SymbolsProcessed: 10,
		SymbolsFailed:    5,
	}

	card := model.renderLastRunCard()

	assert.Contains(t, card, "ERROR")
	assert.Contains(t, card, "10/15")
}

func TestDashboardRenderSymbolsCard(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.ready = true
	model.totalSymbols = 30
	model.activeSymbols = 25

	card := model.renderSymbolsCard()

	assert.Contains(t, card, "Universe")
	assert.Contains(t, card, "25")
	assert.Contains(t, card, "30")
	assert.Contains(t, card, "5 inactive")
}

func TestDashboardRenderCacheCard_NoData(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.ready = true
	model.lastFetchDate = ""

	card := model.renderCacheCard()

	assert.Contains(t, card, "Cache Health")
	assert.Contains(t, card, "No data cached")
}

func TestDashboardRenderCacheCard_WithData(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.ready = true
	model.lastFetchDate = "2025-10-08"

	card := model.renderCacheCard()

	assert.Contains(t, card, "Cache Health")
	assert.Contains(t, card, "2025-10-08")
	assert.Contains(t, card, "days old")
}

func TestDashboardRenderQuotaCard(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)
	model.ready = true
	model.apiQuotaUsed = 15
	model.apiQuotaLimit = 25

	card := model.renderQuotaCard()

	assert.Contains(t, card, "API Quota")
	assert.Contains(t, card, "15/25")
	assert.Contains(t, card, "Resets daily")
}

func TestDashboardLoadData_EmptyDatabase(t *testing.T) {
	database := setupTestDB(t)
	model := NewDashboard(database, 80, 24)

	msg := model.loadData()

	// Should succeed with empty data
	dataMsg, ok := msg.(dashboardDataMsg)
	require.True(t, ok, "expected dashboardDataMsg, got %T", msg)

	assert.Nil(t, dataMsg.run)
	assert.Equal(t, 0, dataMsg.totalSymbols)
	assert.Equal(t, 0, dataMsg.activeSymbols)
	assert.Empty(t, dataMsg.lastFetchDate)
}

func TestDashboardLoadData_WithData(t *testing.T) {
	database := setupTestDB(t)

	// Insert test symbol
	symbolRepo := db.NewSymbolRepository(database)
	err := symbolRepo.Create(&db.Symbol{
		Symbol:    "SPY",
		Name:      "SPDR S&P 500",
		AssetType: "ETF",
		Active:    true,
	})
	require.NoError(t, err)

	// Insert test run
	runRepo := db.NewRunRepository(database)
	runID, err := runRepo.Create("test run")
	require.NoError(t, err)

	err = runRepo.Finish(runID, "OK", 1, 0)
	require.NoError(t, err)

	// Insert test price data
	priceRepo := db.NewPriceRepository(database)
	adjClose := 450.0
	volume := int64(1000000)
	err = priceRepo.Create(&db.Price{
		Symbol:   "SPY",
		Date:     "2025-10-08",
		Open:     445.0,
		High:     452.0,
		Low:      444.0,
		Close:    450.0,
		AdjClose: &adjClose,
		Volume:   &volume,
	})
	require.NoError(t, err)

	model := NewDashboard(database, 80, 24)
	msg := model.loadData()

	// Should return dashboardDataMsg
	dataMsg, ok := msg.(dashboardDataMsg)
	require.True(t, ok, "expected dashboardDataMsg, got %T", msg)

	assert.NotNil(t, dataMsg.run)
	assert.Equal(t, "OK", dataMsg.run.Status)
	assert.Equal(t, 1, dataMsg.totalSymbols)
	assert.Equal(t, 1, dataMsg.activeSymbols)
	assert.Equal(t, "2025-10-08", dataMsg.lastFetchDate)
	assert.Equal(t, 1, dataMsg.apiQuotaUsed) // symbols processed
}

func TestDashboard_FullIntegration(t *testing.T) {
	database := setupTestDB(t)

	// Set up test data
	symbolRepo := db.NewSymbolRepository(database)
	symbols := []db.Symbol{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true},
		{Symbol: "QQQ", Name: "Invesco QQQ", AssetType: "ETF", Active: true},
		{Symbol: "IWM", Name: "iShares Russell 2000", AssetType: "ETF", Active: false},
	}
	for _, sym := range symbols {
		err := symbolRepo.Create(&sym)
		require.NoError(t, err)
	}

	// Create a run
	runRepo := db.NewRunRepository(database)
	runID, err := runRepo.Create("test integration run")
	require.NoError(t, err)
	err = runRepo.Finish(runID, "OK", 2, 0)
	require.NoError(t, err)

	// Add price data
	priceRepo := db.NewPriceRepository(database)
	adjClose := 450.0
	volume := int64(1000000)
	err = priceRepo.Create(&db.Price{
		Symbol:   "SPY",
		Date:     "2025-10-08",
		Open:     445.0,
		High:     452.0,
		Low:      444.0,
		Close:    450.0,
		AdjClose: &adjClose,
		Volume:   &volume,
	})
	require.NoError(t, err)

	// Create model and initialize
	model := NewDashboard(database, 80, 24)
	cmd := model.Init()
	require.NotNil(t, cmd)

	// Execute load command
	msg := cmd()
	model, _ = model.Update(msg)

	// Verify state
	assert.True(t, model.ready)
	assert.Nil(t, model.err)
	assert.NotNil(t, model.latestRun)
	assert.Equal(t, "OK", model.latestRun.Status)
	assert.Equal(t, 3, model.totalSymbols)
	assert.Equal(t, 2, model.activeSymbols)
	assert.Equal(t, "2025-10-08", model.lastFetchDate)

	// Verify view renders
	view := model.View()
	assert.Contains(t, view, "Last Run")
	assert.Contains(t, view, "OK")
	assert.Contains(t, view, "Universe")
	assert.Contains(t, view, "2")  // active
	assert.Contains(t, view, "3")  // total
	assert.Contains(t, view, "Cache Health")
	assert.Contains(t, view, "2025-10-08")
	assert.Contains(t, view, "API Quota")
}
