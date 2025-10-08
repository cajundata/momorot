package screens

import (
	"fmt"
	"testing"

	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSymbol(t *testing.T) {
	database := setupTestDB(t)

	model := NewSymbol(database, "SPY", 100, 30)

	assert.NotNil(t, model.database)
	assert.Equal(t, "SPY", model.symbol)
	assert.Equal(t, 100, model.width)
	assert.Equal(t, 30, model.height)
	assert.False(t, model.ready)
	assert.Nil(t, model.err)
}

func TestSymbolInit(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	cmd := model.Init()
	assert.NotNil(t, cmd)

	// Execute command
	msg := cmd()
	assert.NotNil(t, msg)

	// Should return symbolDataMsg or symbolErrorMsg
	switch msg.(type) {
	case symbolDataMsg:
		// Success case
	case symbolErrorMsg:
		// Error case
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}
}

func TestSymbolUpdateWindowSize(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	updated, cmd := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	assert.Nil(t, cmd)
	assert.Equal(t, 120, updated.width)
	assert.Equal(t, 40, updated.height)
}

func TestSymbolUpdateWithDataMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	// Create test data
	adjClose := 450.0
	score := 1.5
	r1m := 0.15

	symbolInfo := &db.Symbol{
		Symbol:    "SPY",
		Name:      "SPDR S&P 500",
		AssetType: "ETF",
		Active:    true,
	}

	prices := []db.Price{
		{Symbol: "SPY", Date: "2025-10-07", Close: 445.0, AdjClose: &adjClose},
		{Symbol: "SPY", Date: "2025-10-08", Close: 450.0, AdjClose: &adjClose},
	}

	indicators := &db.Indicator{
		Symbol: "SPY",
		Date:   "2025-10-08",
		Score:  &score,
		R1M:    &r1m,
	}

	dataMsg := symbolDataMsg{
		symbolInfo: symbolInfo,
		prices:     prices,
		indicators: indicators,
		rank:       1,
	}

	updated, cmd := model.Update(dataMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.Nil(t, updated.err)
	assert.NotNil(t, updated.symbolInfo)
	assert.Equal(t, "SPY", updated.symbolInfo.Symbol)
	assert.Equal(t, 2, len(updated.prices))
	assert.NotNil(t, updated.indicators)
	assert.Equal(t, 1, updated.rank)
}

func TestSymbolUpdateWithErrorMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	errMsg := symbolErrorMsg{err: assert.AnError}
	updated, cmd := model.Update(errMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.NotNil(t, updated.err)
}

func TestSymbolViewNotReady(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	view := model.View()

	assert.Contains(t, view, "Loading SPY")
}

func TestSymbolViewWithError(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)
	model.err = assert.AnError
	model.ready = true

	view := model.View()

	assert.Contains(t, view, "Error loading symbol")
}

func TestSymbolViewNotFound(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)
	model.ready = true
	model.symbolInfo = nil

	view := model.View()

	assert.Contains(t, view, "not found")
}

func TestSymbolViewWithData(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)
	model.ready = true

	adjClose := 450.0
	score := 1.5
	r1m := 0.15
	r3m := 0.25

	model.symbolInfo = &db.Symbol{
		Symbol:    "SPY",
		Name:      "SPDR S&P 500",
		AssetType: "ETF",
		Active:    true,
	}

	model.prices = []db.Price{
		{Symbol: "SPY", Date: "2025-10-08", Close: 450.0, AdjClose: &adjClose},
	}

	model.indicators = &db.Indicator{
		Symbol: "SPY",
		Date:   "2025-10-08",
		Score:  &score,
		R1M:    &r1m,
		R3M:    &r3m,
	}

	model.rank = 1

	// Update sparkline
	model.sparkline.SetData([]float64{445.0, 450.0})

	view := model.View()

	assert.Contains(t, view, "SPY")
	assert.Contains(t, view, "SPDR S&P 500")
	assert.Contains(t, view, "ETF")
	assert.Contains(t, view, "Rank:")
	assert.Contains(t, view, "#1")
}

func TestSymbolFormatRank(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	// Test with rank
	model.rank = 5
	result := model.formatRank()
	assert.Contains(t, result, "#5")

	// Test without rank
	model.rank = 0
	result = model.formatRank()
	assert.Contains(t, result, "Unranked")
}

func TestSymbolFormatChange(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	tests := []struct {
		name      string
		change    float64
		pctChange float64
		contains  string
	}{
		{"positive", 5.0, 1.5, "+5.00"},
		{"negative", -3.0, -0.8, "-3.00"},
		{"zero", 0.0, 0.0, "+0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.formatChange(tt.change, tt.pctChange)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestSymbolFormatLargeNumber(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	tests := []struct {
		name     string
		value    float64
		contains string
	}{
		{"billions", 2.5e9, "$2.50B"},
		{"millions", 5.3e6, "$5.30M"},
		{"thousands", 1.2e3, "$1.20K"},
		{"small", 100.0, "$100.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.formatLargeNumber(tt.value)
			assert.Equal(t, tt.contains, result)
		})
	}
}

func TestSymbolLoadData_NotFound(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "NOTFOUND", 100, 30)

	msg := model.loadSymbolData()

	// Should return error
	errMsg, ok := msg.(symbolErrorMsg)
	require.True(t, ok, "expected symbolErrorMsg, got %T", msg)
	assert.NotNil(t, errMsg.err)
}

func TestSymbolLoadData_WithFullData(t *testing.T) {
	database := setupTestDB(t)

	// Set up test data
	setupFullSymbolData(t, database, "SPY")

	model := NewSymbol(database, "SPY", 100, 30)
	msg := model.loadSymbolData()

	// Should return data
	dataMsg, ok := msg.(symbolDataMsg)
	require.True(t, ok, "expected symbolDataMsg, got %T", msg)

	assert.NotNil(t, dataMsg.symbolInfo)
	assert.Equal(t, "SPY", dataMsg.symbolInfo.Symbol)
	assert.NotEmpty(t, dataMsg.prices)
	assert.NotNil(t, dataMsg.indicators)
	assert.Equal(t, 1, dataMsg.rank)
}

func TestSymbolLoadData_NoPrices(t *testing.T) {
	database := setupTestDB(t)

	// Add only symbol
	symbolRepo := db.NewSymbolRepository(database)
	err := symbolRepo.Create(&db.Symbol{
		Symbol:    "SPY",
		Name:      "SPDR S&P 500",
		AssetType: "ETF",
		Active:    true,
	})
	require.NoError(t, err)

	model := NewSymbol(database, "SPY", 100, 30)
	msg := model.loadSymbolData()

	// Should return data with no prices/indicators
	dataMsg, ok := msg.(symbolDataMsg)
	require.True(t, ok, "expected symbolDataMsg, got %T", msg)

	assert.NotNil(t, dataMsg.symbolInfo)
	assert.Empty(t, dataMsg.prices)
	assert.Nil(t, dataMsg.indicators)
	assert.Equal(t, 0, dataMsg.rank)
}

func TestSymbol_FullIntegration(t *testing.T) {
	database := setupTestDB(t)

	// Set up test data
	setupFullSymbolData(t, database, "SPY")

	// Create and initialize model
	model := NewSymbol(database, "SPY", 100, 30)
	cmd := model.Init()
	require.NotNil(t, cmd)

	// Execute load command
	msg := cmd()
	model, _ = model.Update(msg)

	// Verify state
	assert.True(t, model.ready)
	assert.Nil(t, model.err)
	assert.NotNil(t, model.symbolInfo)
	assert.NotEmpty(t, model.prices)
	assert.NotNil(t, model.indicators)

	// Verify view renders
	view := model.View()
	assert.Contains(t, view, "SPY")
	assert.Contains(t, view, "SPDR S&P 500")
	assert.Contains(t, view, "Price Chart")
	assert.Contains(t, view, "Return Metrics")
	assert.Contains(t, view, "Volatility")
}

func TestSymbolRenderMetricCard(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	// Test with value
	value := 0.15
	card := model.renderMetricCard("Test", &value)
	assert.Contains(t, card, "Test")
	assert.Contains(t, card, "15.00%")

	// Test with nil
	card = model.renderMetricCard("Test", nil)
	assert.Contains(t, card, "N/A")
}

func TestSymbolRenderVolCard(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	// Test with value
	value := 0.12
	card := model.renderVolCard("Vol", &value)
	assert.Contains(t, card, "Vol")
	assert.Contains(t, card, "12.00%")

	// Test with nil
	card = model.renderVolCard("Vol", nil)
	assert.Contains(t, card, "N/A")
}

func TestSymbolRenderADVCard(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	// Test with value
	value := 5000000.0
	card := model.renderADVCard("ADV", &value)
	assert.Contains(t, card, "ADV")
	assert.Contains(t, card, "$5.00M")

	// Test with nil
	card = model.renderADVCard("ADV", nil)
	assert.Contains(t, card, "N/A")
}

func TestSymbolRenderScoreCard(t *testing.T) {
	database := setupTestDB(t)
	model := NewSymbol(database, "SPY", 100, 30)

	// Test with positive value
	value := 1.5
	card := model.renderScoreCard("Score", &value)
	assert.Contains(t, card, "Score")
	assert.Contains(t, card, "1.500")

	// Test with nil
	card = model.renderScoreCard("Score", nil)
	assert.Contains(t, card, "N/A")
}

// Helper function to set up full symbol data for testing
func setupFullSymbolData(t *testing.T, database *db.DB, symbol string) {
	t.Helper()

	// Add symbol
	symbolRepo := db.NewSymbolRepository(database)
	err := symbolRepo.Create(&db.Symbol{
		Symbol:    symbol,
		Name:      "SPDR S&P 500",
		AssetType: "ETF",
		Active:    true,
	})
	require.NoError(t, err)

	// Add price data
	priceRepo := db.NewPriceRepository(database)
	for i := 0; i < 5; i++ {
		adjClose := 450.0 + float64(i)
		volume := int64(1000000)
		date := fmt.Sprintf("2025-10-%02d", 8-i)
		err := priceRepo.Create(&db.Price{
			Symbol:   symbol,
			Date:     date,
			Open:     445.0,
			High:     452.0,
			Low:      444.0,
			Close:    450.0 + float64(i),
			AdjClose: &adjClose,
			Volume:   &volume,
		})
		require.NoError(t, err)
	}

	// Add indicator data
	score := 1.5
	r1m := 0.15
	r3m := 0.25
	r6m := 0.35
	r12m := 0.45
	vol3m := 0.12
	vol6m := 0.15
	adv := 5000000.0
	rank := 1

	indicators := []db.Indicator{
		{
			Symbol: symbol,
			Date:   "2025-10-08",
			R1M:    &r1m,
			R3M:    &r3m,
			R6M:    &r6m,
			R12M:   &r12m,
			Vol3M:  &vol3m,
			Vol6M:  &vol6m,
			ADV:    &adv,
			Score:  &score,
			Rank:   &rank,
		},
	}

	indicatorRepo := db.NewIndicatorRepository(database)
	err = indicatorRepo.UpsertBatch(indicators)
	require.NoError(t, err)
}
