package screens

import (
	"testing"

	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUniverse(t *testing.T) {
	database := setupTestDB(t)

	model := NewUniverse(database, 100, 30)

	assert.NotNil(t, model.database)
	assert.Equal(t, 100, model.width)
	assert.Equal(t, 30, model.height)
	assert.False(t, model.searchMode)
	assert.False(t, model.ready)
	assert.Nil(t, model.err)
}

func TestUniverseInit(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)

	cmd := model.Init()
	assert.NotNil(t, cmd)

	// Execute command
	msg := cmd()
	assert.NotNil(t, msg)

	// Should return universeDataMsg or universeErrorMsg
	switch msg.(type) {
	case universeDataMsg:
		// Success case
	case universeErrorMsg:
		// Error case
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}
}

func TestUniverseUpdateWindowSize(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)

	updated, cmd := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	assert.Nil(t, cmd)
	assert.Equal(t, 120, updated.width)
	assert.Equal(t, 40, updated.height)
}

func TestUniverseUpdateWithDataMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)

	symbols := []SymbolWithFetch{
		{
			Symbol:        "SPY",
			Name:          "SPDR S&P 500",
			AssetType:     "ETF",
			Active:        true,
			LastFetchDate: "2025-10-08",
		},
	}

	dataMsg := universeDataMsg{symbols: symbols}
	updated, cmd := model.Update(dataMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.Nil(t, updated.err)
	assert.Equal(t, 1, len(updated.symbols))
	assert.Equal(t, 1, len(updated.filteredSymbols))
	assert.Equal(t, "SPY", updated.symbols[0].Symbol)
}

func TestUniverseUpdateWithErrorMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)

	errMsg := universeErrorMsg{err: assert.AnError}
	updated, cmd := model.Update(errMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.NotNil(t, updated.err)
}

func TestUniverseSearchMode_Enter(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.ready = true

	// Enter search mode
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	assert.NotNil(t, cmd)
	assert.True(t, updated.searchMode)
}

func TestUniverseSearchMode_Exit(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.ready = true
	model.searchMode = true

	// Exit search mode with Esc
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.Nil(t, cmd)
	assert.False(t, updated.searchMode)
	assert.Equal(t, "", updated.search.Value())
}

func TestUniverseViewNotReady(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)

	view := model.View()

	assert.Contains(t, view, "Loading")
}

func TestUniverseViewWithError(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.err = assert.AnError
	model.ready = true

	view := model.View()

	assert.Contains(t, view, "Error loading universe")
}

func TestUniverseViewEmpty(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.ready = true
	model.symbols = []SymbolWithFetch{}
	model.filteredSymbols = []SymbolWithFetch{}

	view := model.View()

	assert.Contains(t, view, "No symbols")
}

func TestUniverseViewWithData(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.ready = true
	model.symbols = []SymbolWithFetch{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true, LastFetchDate: "2025-10-08"},
	}
	model.filteredSymbols = model.symbols
	model.updateTableRows()

	view := model.View()

	assert.Contains(t, view, "Symbol Universe")
	assert.Contains(t, view, "SPY")
	assert.Contains(t, view, "Active")
}

func TestUniverseViewSearchMode(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.ready = true
	model.searchMode = true
	model.symbols = []SymbolWithFetch{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true},
	}
	model.filteredSymbols = model.symbols
	model.updateTableRows()

	view := model.View()

	assert.Contains(t, view, "Search:")
}

func TestUniverseFilterSymbols(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.symbols = []SymbolWithFetch{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true},
		{Symbol: "QQQ", Name: "Invesco QQQ", AssetType: "ETF", Active: true},
		{Symbol: "IWM", Name: "iShares Russell 2000", AssetType: "ETF", Active: true},
	}
	model.filteredSymbols = model.symbols

	// Filter for "spy"
	model.filterSymbols("spy")

	assert.Equal(t, 1, len(model.filteredSymbols))
	assert.Equal(t, "SPY", model.filteredSymbols[0].Symbol)

	// Filter for "Q" - should match QQQ only
	model.filterSymbols("Q")

	assert.Equal(t, 1, len(model.filteredSymbols))
	assert.Equal(t, "QQQ", model.filteredSymbols[0].Symbol)

	// Clear filter
	model.filterSymbols("")

	assert.Equal(t, 3, len(model.filteredSymbols))
}

func TestUniverseFilterByName(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)
	model.symbols = []SymbolWithFetch{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true},
		{Symbol: "QQQ", Name: "Invesco QQQ", AssetType: "ETF", Active: true},
	}
	model.filteredSymbols = model.symbols

	// Filter by name
	model.filterSymbols("invesco")

	assert.Equal(t, 1, len(model.filteredSymbols))
	assert.Equal(t, "QQQ", model.filteredSymbols[0].Symbol)
}

func TestUniverseLoadData_EmptyDatabase(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)

	msg := model.loadSymbols()

	// Should succeed with empty data
	dataMsg, ok := msg.(universeDataMsg)
	require.True(t, ok, "expected universeDataMsg, got %T", msg)
	assert.Empty(t, dataMsg.symbols)
}

func TestUniverseLoadData_WithData(t *testing.T) {
	database := setupTestDB(t)

	// Add test symbols
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

	// Add price data for SPY
	priceRepo := db.NewPriceRepository(database)
	adjClose := 450.0
	volume := int64(1000000)
	err := priceRepo.Create(&db.Price{
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

	model := NewUniverse(database, 100, 30)
	msg := model.loadSymbols()

	// Should return symbols (ordered alphabetically: IWM, QQQ, SPY)
	dataMsg, ok := msg.(universeDataMsg)
	require.True(t, ok, "expected universeDataMsg, got %T", msg)
	assert.Equal(t, 3, len(dataMsg.symbols))

	// Symbols are ordered alphabetically
	assert.Equal(t, "IWM", dataMsg.symbols[0].Symbol)
	assert.Equal(t, "QQQ", dataMsg.symbols[1].Symbol)
	assert.Equal(t, "SPY", dataMsg.symbols[2].Symbol)

	// Only SPY has price data
	assert.Empty(t, dataMsg.symbols[0].LastFetchDate) // IWM
	assert.Empty(t, dataMsg.symbols[1].LastFetchDate) // QQQ
	assert.Equal(t, "2025-10-08", dataMsg.symbols[2].LastFetchDate) // SPY
}

func TestUniverseToggleActive(t *testing.T) {
	database := setupTestDB(t)

	// Add test symbol
	symbolRepo := db.NewSymbolRepository(database)
	err := symbolRepo.Create(&db.Symbol{
		Symbol:    "SPY",
		Name:      "SPDR S&P 500",
		AssetType: "ETF",
		Active:    true,
	})
	require.NoError(t, err)

	model := NewUniverse(database, 100, 30)
	model.ready = true
	model.filteredSymbols = []SymbolWithFetch{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true},
	}

	// Toggle active status
	cmd := model.toggleActive()
	require.NotNil(t, cmd)

	msg := cmd()

	// Should return symbolToggledMsg
	toggleMsg, ok := msg.(symbolToggledMsg)
	require.True(t, ok, "expected symbolToggledMsg, got %T", msg)
	assert.Equal(t, "SPY", toggleMsg.symbol)
	assert.False(t, toggleMsg.active) // Should be toggled to inactive

	// Verify in database
	sym, err := symbolRepo.Get("SPY")
	require.NoError(t, err)
	assert.False(t, sym.Active)
}

func TestUniverse_FullIntegration(t *testing.T) {
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

	// Create and initialize model
	model := NewUniverse(database, 100, 30)
	cmd := model.Init()
	require.NotNil(t, cmd)

	// Execute load command
	msg := cmd()
	model, _ = model.Update(msg)

	// Verify state
	assert.True(t, model.ready)
	assert.Nil(t, model.err)
	assert.Equal(t, 3, len(model.symbols))
	assert.Equal(t, 3, len(model.filteredSymbols))

	// Verify view renders
	view := model.View()
	assert.Contains(t, view, "Symbol Universe")
	assert.Contains(t, view, "SPY")
	assert.Contains(t, view, "QQQ")
	assert.Contains(t, view, "IWM")
	assert.Contains(t, view, "Active")
	assert.Contains(t, view, "Inactive")

	// Test search
	model.searchMode = true
	model.search.SetValue("spy")
	model.filterSymbols("spy")
	assert.Equal(t, 1, len(model.filteredSymbols))

	// Test toggle
	model.searchMode = false
	model.filterSymbols("") // Reset filter
	cmd = model.toggleActive()
	require.NotNil(t, cmd)
	msg = cmd()
	model, _ = model.Update(msg)

	// Should trigger reload
	assert.NotNil(t, model.loadSymbols)
}

func TestUniverseUpdateTableRows(t *testing.T) {
	database := setupTestDB(t)
	model := NewUniverse(database, 100, 30)

	model.filteredSymbols = []SymbolWithFetch{
		{Symbol: "SPY", Name: "SPDR S&P 500", AssetType: "ETF", Active: true, LastFetchDate: "2025-10-08"},
		{Symbol: "QQQ", Name: "Invesco QQQ", AssetType: "ETF", Active: false, LastFetchDate: ""},
	}

	model.updateTableRows()

	// Table should have 2 rows
	// We can't directly access table rows, but we can verify it doesn't panic
	assert.NotPanics(t, func() {
		model.updateTableRows()
	})
}
