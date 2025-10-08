package ui

import (
	"testing"

	"github.com/cajundata/momorot/internal/config"
	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestModel(t *testing.T) (Model, *db.DB) {
	// Create in-memory test database
	database, err := db.New(db.Config{Path: ":memory:"})
	require.NoError(t, err)

	// Run migrations
	err = database.Migrate()
	require.NoError(t, err)

	// Create test config
	cfg := &config.Config{
		Lookbacks: config.LookbacksConfig{
			R1M:  21,
			R3M:  63,
			R6M:  126,
			R12M: 252,
		},
		VolWindows: config.VolWindowsConfig{
			Short: 63,
			Long:  126,
		},
		Scoring: config.ScoringConfig{
			PenaltyLambda:         0.35,
			MinADVUSD:             5000000,
			BreadthMinPositive:    2,
			BreadthTotalLookbacks: 4,
		},
	}

	model := New(database, cfg)

	return model, database
}

func TestNew(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	assert.NotNil(t, model.db)
	assert.NotNil(t, model.orchestrator)
	assert.NotNil(t, model.config)
	assert.Equal(t, ScreenDashboard, model.currentScreen)
	assert.Len(t, model.screenHistory, 0)
	assert.False(t, model.loading)
}

func TestInit(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	cmd := model.Init()
	// Init should return a batch command that initializes all screens
	assert.NotNil(t, cmd)
}

func TestNavigateTo(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	// Navigate to leaders screen
	model.NavigateTo(ScreenLeaders)

	assert.Equal(t, ScreenLeaders, model.currentScreen)
	assert.Len(t, model.screenHistory, 1)
	assert.Equal(t, ScreenDashboard, model.screenHistory[0])
}

func TestNavigateBack(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	// Navigate forward
	model.NavigateTo(ScreenLeaders)
	model.NavigateTo(ScreenUniverse)

	// Navigate back
	model.NavigateBack()

	assert.Equal(t, ScreenLeaders, model.currentScreen)
	assert.Len(t, model.screenHistory, 1)
}

func TestNavigateNext(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	tests := []struct {
		from Screen
		to   Screen
	}{
		{ScreenDashboard, ScreenLeaders},
		{ScreenLeaders, ScreenUniverse},
		{ScreenUniverse, ScreenSymbol},
		{ScreenSymbol, ScreenLogs},
		{ScreenLogs, ScreenDashboard},
	}

	for _, tt := range tests {
		model.currentScreen = tt.from
		model = model.navigateNext()
		assert.Equal(t, tt.to, model.currentScreen)
	}
}

func TestNavigatePrev(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	tests := []struct {
		from Screen
		to   Screen
	}{
		{ScreenDashboard, ScreenLogs},
		{ScreenLeaders, ScreenDashboard},
		{ScreenUniverse, ScreenLeaders},
		{ScreenSymbol, ScreenUniverse},
		{ScreenLogs, ScreenSymbol},
	}

	for _, tt := range tests {
		model.currentScreen = tt.from
		model = model.navigatePrev()
		assert.Equal(t, tt.to, model.currentScreen)
	}
}

func TestSetLoading(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	model.SetLoading(true, "Fetching data")

	assert.True(t, model.loading)
	assert.Equal(t, "Fetching data", model.loadingMsg)

	model.SetLoading(false, "")

	assert.False(t, model.loading)
	assert.Equal(t, "", model.loadingMsg)
}

func TestSetError(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	model.SetError("Test error")
	assert.Equal(t, "Test error", model.errorMsg)

	model.ClearError()
	assert.Equal(t, "", model.errorMsg)
}

func TestSetStatus(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	model.SetStatus("Test status")
	assert.Equal(t, "Test status", model.statusBarMsg)
}

func TestUpdate_WindowSize(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(Model)
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestUpdate_Quit(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)

	assert.NotNil(t, cmd)
}

func TestView(t *testing.T) {
	model, database := setupTestModel(t)
	defer database.Close()

	view := model.View()

	assert.Contains(t, view, "Momentum Screener")
	assert.Contains(t, view, "Dashboard")
	assert.Contains(t, view, "Leaders")
}

func TestDefaultKeyBindings(t *testing.T) {
	keys := DefaultKeyBindings()

	assert.NotNil(t, keys.Quit)
	assert.NotNil(t, keys.Refresh)
	assert.NotNil(t, keys.Search)
	assert.NotNil(t, keys.Export)
	assert.NotNil(t, keys.NextTab)
	assert.NotNil(t, keys.PrevTab)

	// Test help text
	shortHelp := keys.ShortHelp()
	assert.Len(t, shortHelp, 6)

	fullHelp := keys.FullHelp()
	assert.Len(t, fullHelp, 4)
}
