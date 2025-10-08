package screens

import (
	"testing"

	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLeaders(t *testing.T) {
	database := setupTestDB(t)

	model := NewLeaders(database, 100, 30)

	assert.NotNil(t, model.database)
	assert.Equal(t, 100, model.width)
	assert.Equal(t, 30, model.height)
	assert.Equal(t, 10, model.topN)
	assert.False(t, model.ready)
	assert.Nil(t, model.err)
}

func TestLeadersInit(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	cmd := model.Init()
	assert.NotNil(t, cmd)

	// Execute command
	msg := cmd()
	assert.NotNil(t, msg)

	// Should return either leadersDataMsg or leadersErrorMsg
	switch msg.(type) {
	case leadersDataMsg:
		// Success case
	case leadersErrorMsg:
		// Error case (expected if DB is empty)
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}
}

func TestLeadersUpdateWindowSize(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	updated, cmd := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	assert.Nil(t, cmd)
	assert.Equal(t, 120, updated.width)
	assert.Equal(t, 40, updated.height)
}

func TestLeadersUpdateWithDataMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	// Create test data
	score := 1.5
	r1m := 0.15
	r3m := 0.25
	r6m := 0.35
	vol3m := 0.12
	adv := 5000000.0
	rank := 1

	leaders := []db.Indicator{
		{
			Symbol: "SPY",
			Date:   "2025-10-08",
			Score:  &score,
			R1M:    &r1m,
			R3M:    &r3m,
			R6M:    &r6m,
			Vol3M:  &vol3m,
			ADV:    &adv,
			Rank:   &rank,
		},
	}

	dataMsg := leadersDataMsg{leaders: leaders}
	updated, cmd := model.Update(dataMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.Nil(t, updated.err)
	assert.Equal(t, 1, len(updated.leaders))
	assert.Equal(t, "SPY", updated.leaders[0].Symbol)
}

func TestLeadersUpdateWithErrorMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	errMsg := leadersErrorMsg{err: assert.AnError}
	updated, cmd := model.Update(errMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.NotNil(t, updated.err)
}

func TestLeadersViewNotReady(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	view := model.View()

	assert.Contains(t, view, "Loading")
}

func TestLeadersViewWithError(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)
	model.err = assert.AnError
	model.ready = true

	view := model.View()

	assert.Contains(t, view, "Error loading leaders")
}

func TestLeadersViewEmpty(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)
	model.ready = true
	model.leaders = []db.Indicator{}

	view := model.View()

	assert.Contains(t, view, "No ranking data")
}

func TestLeadersViewWithData(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	// Set up test data
	score := 1.5
	r1m := 0.15
	rank := 1
	model.leaders = []db.Indicator{
		{Symbol: "SPY", Score: &score, R1M: &r1m, Rank: &rank},
	}
	model.ready = true
	model.updateTableRows()

	view := model.View()

	assert.Contains(t, view, "Top Leaders")
	assert.Contains(t, view, "SPY")
}

func TestLeadersFormatValue(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	tests := []struct {
		name     string
		value    *float64
		colorize bool
		contains string
	}{
		{"nil value", nil, false, "N/A"},
		{"positive value", ptr(1.5), true, "1.500"},
		{"negative value", ptr(-0.5), true, "0.500"},
		{"zero value", ptr(0.0), false, "0.000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.formatValue(tt.value, tt.colorize)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestLeadersFormatPercent(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	tests := []struct {
		name     string
		value    *float64
		contains string
	}{
		{"nil value", nil, "N/A"},
		{"positive percent", ptr(0.15), "15.00%"},
		{"negative percent", ptr(-0.05), "5.00%"},
		{"zero percent", ptr(0.0), "0.00%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.formatPercent(tt.value)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestLeadersFormatLargeNumber(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	tests := []struct {
		name     string
		value    *float64
		contains string
	}{
		{"nil value", nil, "N/A"},
		{"billions", ptr(2.5e9), "$2.50B"},
		{"millions", ptr(5.3e6), "$5.30M"},
		{"thousands", ptr(1.2e3), "$1.20K"},
		{"small", ptr(100.0), "$100.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.formatLargeNumber(tt.value)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestLeadersLoadData_EmptyDatabase(t *testing.T) {
	database := setupTestDB(t)
	model := NewLeaders(database, 100, 30)

	msg := model.loadLeaders()

	// Should succeed with empty data
	dataMsg, ok := msg.(leadersDataMsg)
	require.True(t, ok, "expected leadersDataMsg, got %T", msg)
	assert.Empty(t, dataMsg.leaders)
}

func TestLeadersLoadData_WithData(t *testing.T) {
	database := setupTestDB(t)

	// Set up test data
	setupTestIndicators(t, database)

	model := NewLeaders(database, 100, 30)
	msg := model.loadLeaders()

	// Should return leaders
	dataMsg, ok := msg.(leadersDataMsg)
	require.True(t, ok, "expected leadersDataMsg, got %T", msg)
	assert.NotEmpty(t, dataMsg.leaders)
	assert.Equal(t, "SPY", dataMsg.leaders[0].Symbol)
}

func TestLeaders_FullIntegration(t *testing.T) {
	database := setupTestDB(t)

	// Set up test data
	setupTestIndicators(t, database)

	// Create and initialize model
	model := NewLeaders(database, 100, 30)
	cmd := model.Init()
	require.NotNil(t, cmd)

	// Execute load command
	msg := cmd()
	model, _ = model.Update(msg)

	// Verify state
	assert.True(t, model.ready)
	assert.Nil(t, model.err)
	assert.NotEmpty(t, model.leaders)

	// Verify view renders
	view := model.View()
	assert.Contains(t, view, "Top Leaders")
	assert.Contains(t, view, "SPY")
}

// Helper functions

func ptr(f float64) *float64 {
	return &f
}

func setupTestIndicators(t *testing.T, database *db.DB) {
	t.Helper()

	// Add symbol
	symbolRepo := db.NewSymbolRepository(database)
	err := symbolRepo.Create(&db.Symbol{
		Symbol:    "SPY",
		Name:      "SPDR S&P 500",
		AssetType: "ETF",
		Active:    true,
	})
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
			Symbol: "SPY",
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
