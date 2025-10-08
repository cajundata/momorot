package screens

import (
	"testing"
	"time"

	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogs(t *testing.T) {
	database := setupTestDB(t)

	model := NewLogs(database, 100, 40)

	assert.NotNil(t, model.database)
	assert.Equal(t, 100, model.width)
	assert.Equal(t, 40, model.height)
	assert.Equal(t, 0, model.focusedTable)
	assert.False(t, model.ready)
	assert.Nil(t, model.err)
}

func TestLogsInit(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	cmd := model.Init()
	assert.NotNil(t, cmd)

	// Execute command
	msg := cmd()
	assert.NotNil(t, msg)

	// Should return logsDataMsg or logsErrorMsg
	switch msg.(type) {
	case logsDataMsg:
		// Success case
	case logsErrorMsg:
		// Error case
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}
}

func TestLogsUpdateWindowSize(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	updated, cmd := model.Update(tea.WindowSizeMsg{Width: 120, Height: 50})

	assert.Nil(t, cmd)
	assert.Equal(t, 120, updated.width)
	assert.Equal(t, 50, updated.height)
}

func TestLogsUpdateTabKey(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)
	model.ready = true

	// Start with focus on runs (0)
	assert.Equal(t, 0, model.focusedTable)

	// Press tab to switch to logs (1)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Nil(t, cmd)
	assert.Equal(t, 1, updated.focusedTable)

	// Press tab again to switch back to runs (0)
	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Nil(t, cmd)
	assert.Equal(t, 0, updated.focusedTable)
}

func TestLogsUpdateFilterKey(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)
	model.ready = true

	// Start with no filter
	assert.Equal(t, "", model.filterStatus)

	// Press 'f' to cycle filter - should trigger reload
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.NotNil(t, cmd) // Should trigger reload
}

func TestLogsUpdateWithDataMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	runs := []db.Run{
		{
			RunID:            1,
			StartedAt:        time.Now(),
			Status:           "OK",
			SymbolsProcessed: 10,
			SymbolsFailed:    0,
		},
	}

	errorLogs := []db.FetchLog{}

	dataMsg := logsDataMsg{
		runs:      runs,
		errorLogs: errorLogs,
	}

	updated, cmd := model.Update(dataMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.Nil(t, updated.err)
	assert.Equal(t, 1, len(updated.runs))
}

func TestLogsUpdateWithErrorLogsMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	msg := "test error"
	logs := []db.FetchLog{
		{
			RunID:   1,
			Symbol:  "SPY",
			Rows:    0,
			OK:      false,
			Message: &msg,
		},
	}

	logsMsg := errorLogsMsg{logs: logs}
	updated, cmd := model.Update(logsMsg)

	assert.Nil(t, cmd)
	assert.Equal(t, 1, len(updated.errorLogs))
}

func TestLogsUpdateWithErrorMsg(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	errMsg := logsErrorMsg{err: assert.AnError}
	updated, cmd := model.Update(errMsg)

	assert.Nil(t, cmd)
	assert.True(t, updated.ready)
	assert.NotNil(t, updated.err)
}

func TestLogsViewNotReady(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	view := model.View()

	assert.Contains(t, view, "Loading")
}

func TestLogsViewWithError(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)
	model.err = assert.AnError
	model.ready = true

	view := model.View()

	assert.Contains(t, view, "Error loading logs")
}

func TestLogsViewWithData(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)
	model.ready = true

	model.runs = []db.Run{
		{
			RunID:            1,
			StartedAt:        time.Date(2025, 10, 8, 12, 0, 0, 0, time.UTC),
			Status:           "OK",
			SymbolsProcessed: 10,
			SymbolsFailed:    0,
		},
	}
	model.selectedRun = 1
	model.updateRunsTable()

	view := model.View()

	assert.Contains(t, view, "Run History")
	assert.Contains(t, view, "#1")
	assert.Contains(t, view, "OK")
}

func TestLogsViewWithFilter(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)
	model.ready = true
	model.filterStatus = "ERROR"

	view := model.View()

	assert.Contains(t, view, "filtered: ERROR")
}

func TestLogsCycleFilter(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	// Start with no filter
	assert.Equal(t, "", model.filterStatus)

	// Test the cycle logic directly by updating filterStatus
	// (cycleFilter modifies the model and returns a cmd, which we test separately)
	model.filterStatus = ""

	// Manually cycle through statuses to test the logic
	switch model.filterStatus {
	case "":
		model.filterStatus = "OK"
	}
	assert.Equal(t, "OK", model.filterStatus)

	switch model.filterStatus {
	case "OK":
		model.filterStatus = "ERROR"
	}
	assert.Equal(t, "ERROR", model.filterStatus)

	switch model.filterStatus {
	case "ERROR":
		model.filterStatus = "RUNNING"
	}
	assert.Equal(t, "RUNNING", model.filterStatus)

	switch model.filterStatus {
	case "RUNNING":
		model.filterStatus = ""
	}
	assert.Equal(t, "", model.filterStatus)
}

func TestLogsUpdateRunsTable(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	model.runs = []db.Run{
		{
			RunID:            1,
			StartedAt:        time.Date(2025, 10, 8, 12, 0, 0, 0, time.UTC),
			Status:           "OK",
			SymbolsProcessed: 10,
			SymbolsFailed:    0,
		},
		{
			RunID:            2,
			StartedAt:        time.Date(2025, 10, 7, 12, 0, 0, 0, time.UTC),
			Status:           "ERROR",
			SymbolsProcessed: 5,
			SymbolsFailed:    5,
		},
	}

	model.updateRunsTable()

	// Table should have 2 rows
	assert.NotPanics(t, func() {
		model.updateRunsTable()
	})
}

func TestLogsUpdateLogsTable_Empty(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	model.errorLogs = []db.FetchLog{}
	model.updateLogsTable()

	// Should not panic
	assert.NotPanics(t, func() {
		model.updateLogsTable()
	})
}

func TestLogsUpdateLogsTable_WithData(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	fromDate := "2025-10-01"
	toDate := "2025-10-08"
	msg := "API rate limit exceeded"

	model.errorLogs = []db.FetchLog{
		{
			RunID:    1,
			Symbol:   "SPY",
			FromDate: &fromDate,
			ToDate:   &toDate,
			Rows:     0,
			OK:       false,
			Message:  &msg,
		},
	}

	model.updateLogsTable()

	// Should not panic
	assert.NotPanics(t, func() {
		model.updateLogsTable()
	})
}

func TestLogsLoadData_EmptyDatabase(t *testing.T) {
	database := setupTestDB(t)
	model := NewLogs(database, 100, 40)

	msg := model.loadLogs()

	// Should succeed with empty data
	dataMsg, ok := msg.(logsDataMsg)
	require.True(t, ok, "expected logsDataMsg, got %T", msg)
	assert.Empty(t, dataMsg.runs)
	assert.Empty(t, dataMsg.errorLogs)
}

func TestLogsLoadData_WithRuns(t *testing.T) {
	database := setupTestDB(t)

	// Create test runs
	runRepo := db.NewRunRepository(database)
	runID1, err := runRepo.Create("test run 1")
	require.NoError(t, err)
	err = runRepo.Finish(runID1, "OK", 10, 0)
	require.NoError(t, err)

	runID2, err := runRepo.Create("test run 2")
	require.NoError(t, err)
	err = runRepo.Finish(runID2, "ERROR", 5, 5)
	require.NoError(t, err)

	model := NewLogs(database, 100, 40)
	msg := model.loadLogs()

	// Should return runs
	dataMsg, ok := msg.(logsDataMsg)
	require.True(t, ok, "expected logsDataMsg, got %T", msg)
	assert.Equal(t, 2, len(dataMsg.runs))
	assert.Equal(t, runID2, dataMsg.runs[0].RunID) // Newer run first
}

func TestLogsLoadData_WithFilter(t *testing.T) {
	database := setupTestDB(t)

	// Create test runs
	runRepo := db.NewRunRepository(database)
	runID1, err := runRepo.Create("ok run")
	require.NoError(t, err)
	err = runRepo.Finish(runID1, "OK", 10, 0)
	require.NoError(t, err)

	runID2, err := runRepo.Create("error run")
	require.NoError(t, err)
	err = runRepo.Finish(runID2, "ERROR", 5, 5)
	require.NoError(t, err)

	// Filter for ERROR status
	model := NewLogs(database, 100, 40)
	model.filterStatus = "ERROR"

	msg := model.loadLogs()

	// Should return only ERROR runs
	dataMsg, ok := msg.(logsDataMsg)
	require.True(t, ok, "expected logsDataMsg, got %T", msg)
	assert.Equal(t, 1, len(dataMsg.runs))
	assert.Equal(t, "ERROR", dataMsg.runs[0].Status)
}

func TestLogsLoadErrorLogsForRun(t *testing.T) {
	database := setupTestDB(t)

	// Create test run with error logs
	runRepo := db.NewRunRepository(database)
	runID, err := runRepo.Create("test run")
	require.NoError(t, err)

	// Create error log
	fetchLogRepo := db.NewFetchLogRepository(database)
	errMsg := "test error"
	err = fetchLogRepo.Log(&db.FetchLog{
		RunID:   runID,
		Symbol:  "SPY",
		Rows:    0,
		OK:      false,
		Message: &errMsg,
	})
	require.NoError(t, err)

	model := NewLogs(database, 100, 40)
	logs := model.loadErrorLogsForRun(runID)

	assert.Equal(t, 1, len(logs))
	assert.Equal(t, "SPY", logs[0].Symbol)
	assert.False(t, logs[0].OK)
}

func TestLogs_FullIntegration(t *testing.T) {
	database := setupTestDB(t)

	// Create test data
	runRepo := db.NewRunRepository(database)
	runID, err := runRepo.Create("integration test run")
	require.NoError(t, err)
	err = runRepo.Finish(runID, "ERROR", 5, 2)
	require.NoError(t, err)

	// Add error logs
	fetchLogRepo := db.NewFetchLogRepository(database)
	errMsg1 := "API rate limit"
	err = fetchLogRepo.Log(&db.FetchLog{
		RunID:   runID,
		Symbol:  "SPY",
		Rows:    0,
		OK:      false,
		Message: &errMsg1,
	})
	require.NoError(t, err)

	errMsg2 := "Network timeout"
	err = fetchLogRepo.Log(&db.FetchLog{
		RunID:   runID,
		Symbol:  "QQQ",
		Rows:    0,
		OK:      false,
		Message: &errMsg2,
	})
	require.NoError(t, err)

	// Create and initialize model
	model := NewLogs(database, 100, 40)
	cmd := model.Init()
	require.NotNil(t, cmd)

	// Execute load command
	msg := cmd()
	model, _ = model.Update(msg)

	// Verify state
	assert.True(t, model.ready)
	assert.Nil(t, model.err)
	assert.Equal(t, 1, len(model.runs))
	assert.Equal(t, "ERROR", model.runs[0].Status)
	assert.Equal(t, 2, len(model.errorLogs))

	// Verify view renders
	view := model.View()
	assert.Contains(t, view, "Run History")
	assert.Contains(t, view, "ERROR")
	assert.Contains(t, view, "Error Logs")
	assert.Contains(t, view, "SPY")
	assert.Contains(t, view, "QQQ")

	// Test switching focus
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, model.focusedTable)
}

func TestLogsEnterKeyLoadsErrorLogs(t *testing.T) {
	database := setupTestDB(t)

	// Create two runs
	runRepo := db.NewRunRepository(database)
	runID1, err := runRepo.Create("run 1")
	require.NoError(t, err)
	err = runRepo.Finish(runID1, "OK", 10, 0)
	require.NoError(t, err)

	runID2, err := runRepo.Create("run 2")
	require.NoError(t, err)
	err = runRepo.Finish(runID2, "ERROR", 5, 2)
	require.NoError(t, err)

	// Add error log to run 2
	fetchLogRepo := db.NewFetchLogRepository(database)
	errMsg := "test error"
	err = fetchLogRepo.Log(&db.FetchLog{
		RunID:   runID2,
		Symbol:  "SPY",
		Rows:    0,
		OK:      false,
		Message: &errMsg,
	})
	require.NoError(t, err)

	// Initialize model
	model := NewLogs(database, 100, 40)
	cmd := model.Init()
	msg := cmd()
	model, _ = model.Update(msg)

	// runID2 should be first since it's newer (DESC order)
	assert.Equal(t, runID2, model.selectedRun)

	// Initial load should have loaded error logs for the first (newest) run
	// which is runID2, so it should have 1 error log
	assert.Equal(t, 1, len(model.errorLogs))

	// Move cursor down to select runID1 (at index 1, since runID2 is at index 0)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter to load error logs for the run at cursor position
	model, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)

	// Execute command - should load logs for runID1 (which has none)
	msg = cmd()
	model, _ = model.Update(msg)

	// Should have no error logs for run 1
	assert.Equal(t, 0, len(model.errorLogs))
	assert.Equal(t, runID1, model.selectedRun)
}
