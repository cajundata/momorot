package screens

import (
	"fmt"
	"time"

	"github.com/cajundata/momorot/internal/db"
	"github.com/cajundata/momorot/internal/ui/components"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogsModel represents the runs/logs screen state.
type LogsModel struct {
	database   *db.DB
	runsTable  components.TableModel
	logsTable  components.TableModel
	theme      LogsTheme

	// Screen data
	runs         []db.Run
	errorLogs    []db.FetchLog
	selectedRun  int64
	filterStatus string // "", "OK", "ERROR", "RUNNING"

	// UI state
	focusedTable int  // 0 = runs, 1 = logs
	width        int
	height       int
	ready        bool
	err          error
}

// LogsTheme contains styling for the logs screen.
type LogsTheme struct {
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	SectionTitle lipgloss.Style
	StatusOK     lipgloss.Style
	StatusError  lipgloss.Style
	StatusRunning lipgloss.Style
	Help         lipgloss.Style
	EmptyMsg     lipgloss.Style
}

// NewLogs creates a new logs model.
func NewLogs(database *db.DB, width, height int) LogsModel {
	// Calculate table heights (split screen)
	runsHeight := (height - 12) / 2
	logsHeight := (height - 12) / 2

	// Create runs table
	runsColumns := []table.Column{
		{Title: "Run ID", Width: 8},
		{Title: "Started", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Processed", Width: 12},
		{Title: "Failed", Width: 10},
	}
	runsTable := components.NewTable(runsColumns, []table.Row{}, width-4, runsHeight)

	// Create logs table
	logsColumns := []table.Column{
		{Title: "Symbol", Width: 10},
		{Title: "Date Range", Width: 25},
		{Title: "Rows", Width: 8},
		{Title: "Error", Width: 50},
	}
	logsTable := components.NewTable(logsColumns, []table.Row{}, width-4, logsHeight)

	return LogsModel{
		database:     database,
		runsTable:    runsTable,
		logsTable:    logsTable,
		theme:        defaultLogsTheme(),
		focusedTable: 0,
		width:        width,
		height:       height,
		ready:        false,
	}
}

// defaultLogsTheme returns the default logs theme.
func defaultLogsTheme() LogsTheme {
	return LogsTheme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			MarginBottom(1),
		SectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")).
			MarginTop(1).
			MarginBottom(1),
		StatusOK: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true),
		StatusRunning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		EmptyMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Padding(1, 2),
	}
}

// Init initializes the logs screen and loads data.
func (m LogsModel) Init() tea.Cmd {
	return m.loadLogs
}

// Update handles messages for the logs screen.
func (m LogsModel) Update(msg tea.Msg) (LogsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		runsHeight := (msg.Height - 12) / 2
		logsHeight := (msg.Height - 12) / 2

		m.runsTable.SetWidth(msg.Width - 4)
		m.runsTable.SetHeight(runsHeight)
		m.logsTable.SetWidth(msg.Width - 4)
		m.logsTable.SetHeight(logsHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Switch focus between tables
			m.focusedTable = (m.focusedTable + 1) % 2
			if m.focusedTable == 0 {
				m.runsTable.Focus()
				m.logsTable.Blur()
			} else {
				m.runsTable.Blur()
				m.logsTable.Focus()
			}
			return m, nil

		case "f":
			// Cycle through status filters
			return m, m.cycleFilter()

		case "enter":
			// Load error logs for selected run
			if m.focusedTable == 0 && m.ready && len(m.runs) > 0 {
				cursor := m.runsTable.Cursor()
				if cursor < len(m.runs) {
					m.selectedRun = m.runs[cursor].RunID
					return m, m.loadErrorLogs
				}
			}
		}

	case logsDataMsg:
		m.runs = msg.runs
		m.errorLogs = msg.errorLogs
		m.ready = true
		m.err = nil
		// Set selectedRun to the first run if available
		if len(m.runs) > 0 {
			m.selectedRun = m.runs[0].RunID
		}
		m.updateRunsTable()
		m.updateLogsTable()
		return m, nil

	case errorLogsMsg:
		m.errorLogs = msg.logs
		m.updateLogsTable()
		return m, nil

	case logsErrorMsg:
		m.err = msg.err
		m.ready = true
		return m, nil
	}

	// Pass through to focused table
	if m.focusedTable == 0 {
		m.runsTable, cmd = m.runsTable.Update(msg)
	} else {
		m.logsTable, cmd = m.logsTable.Update(msg)
	}

	return m, cmd
}

// View renders the logs screen.
func (m LogsModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading logs: %v", m.err))
	}

	if !m.ready {
		return m.theme.EmptyMsg.Render("Loading run history...")
	}

	// Header
	title := m.theme.Title.Render("📝 Run History & Error Logs")

	filterText := ""
	if m.filterStatus != "" {
		filterText = fmt.Sprintf(" (filtered: %s)", m.filterStatus)
	}
	subtitle := m.theme.Subtitle.Render(fmt.Sprintf(
		"Showing %d runs%s",
		len(m.runs),
		filterText,
	))

	// Runs section
	runsTitle := m.theme.SectionTitle.Render("Run History")
	runsView := m.runsTable.View()

	// Logs section
	logsTitle := m.theme.SectionTitle.Render(fmt.Sprintf(
		"Error Logs (Run #%d)",
		m.selectedRun,
	))
	logsView := m.logsTable.View()

	// Help text
	help := m.theme.Help.Render("Tab: Switch Tables | ↑/↓: Navigate | Enter: Load Errors | f: Filter Status")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		runsTitle,
		runsView,
		"",
		logsTitle,
		logsView,
		"",
		help,
	)
}

// updateRunsTable updates the runs table with data.
func (m *LogsModel) updateRunsTable() {
	rows := make([]table.Row, 0, len(m.runs))

	for _, run := range m.runs {
		// Format status with color
		var status string
		switch run.Status {
		case "OK":
			status = m.theme.StatusOK.Render("✓ OK")
		case "ERROR":
			status = m.theme.StatusError.Render("✗ ERROR")
		case "RUNNING":
			status = m.theme.StatusRunning.Render("◌ RUNNING")
		default:
			status = run.Status
		}

		// Format started time
		started := run.StartedAt.Format("2006-01-02 15:04:05")

		rows = append(rows, table.Row{
			fmt.Sprintf("#%d", run.RunID),
			started,
			status,
			fmt.Sprintf("%d", run.SymbolsProcessed),
			fmt.Sprintf("%d", run.SymbolsFailed),
		})
	}

	m.runsTable.SetRows(rows)
}

// updateLogsTable updates the error logs table with data.
func (m *LogsModel) updateLogsTable() {
	if len(m.errorLogs) == 0 {
		m.logsTable.SetRows([]table.Row{})
		return
	}

	rows := make([]table.Row, 0, len(m.errorLogs))

	for _, log := range m.errorLogs {
		// Format date range
		dateRange := "N/A"
		if log.FromDate != nil && log.ToDate != nil {
			dateRange = fmt.Sprintf("%s to %s", *log.FromDate, *log.ToDate)
		} else if log.FromDate != nil {
			dateRange = *log.FromDate
		}

		// Format error message
		errMsg := "Unknown error"
		if log.Message != nil {
			errMsg = *log.Message
			// Truncate if too long
			if len(errMsg) > 47 {
				errMsg = errMsg[:44] + "..."
			}
		}

		rows = append(rows, table.Row{
			log.Symbol,
			dateRange,
			fmt.Sprintf("%d", log.Rows),
			errMsg,
		})
	}

	m.logsTable.SetRows(rows)
}

// cycleFilter cycles through status filters.
func (m LogsModel) cycleFilter() tea.Cmd {
	// Cycle: "" -> "OK" -> "ERROR" -> "RUNNING" -> ""
	switch m.filterStatus {
	case "":
		m.filterStatus = "OK"
	case "OK":
		m.filterStatus = "ERROR"
	case "ERROR":
		m.filterStatus = "RUNNING"
	case "RUNNING":
		m.filterStatus = ""
	}

	return m.loadLogs
}

// loadLogs loads run history from the database.
func (m LogsModel) loadLogs() tea.Msg {
	// Build query based on filter
	query := `
		SELECT run_id, started_at, finished_at, status, symbols_processed, symbols_failed, notes
		FROM runs
	`

	var args []interface{}
	if m.filterStatus != "" {
		query += " WHERE status = ?"
		args = append(args, m.filterStatus)
	}

	query += " ORDER BY started_at DESC, run_id DESC LIMIT 50"

	rows, err := m.database.Query(query, args...)
	if err != nil {
		return logsErrorMsg{err: fmt.Errorf("failed to query runs: %w", err)}
	}
	defer rows.Close()

	var runs []db.Run
	for rows.Next() {
		var run db.Run
		var finishedAt, notes *string
		var startedAt string

		err := rows.Scan(
			&run.RunID,
			&startedAt,
			&finishedAt,
			&run.Status,
			&run.SymbolsProcessed,
			&run.SymbolsFailed,
			&notes,
		)
		if err != nil {
			return logsErrorMsg{err: fmt.Errorf("failed to scan run: %w", err)}
		}

		// Parse started_at
		run.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)

		if finishedAt != nil {
			t, _ := time.Parse("2006-01-02 15:04:05", *finishedAt)
			run.FinishedAt = &t
		}

		if notes != nil {
			run.Notes = notes
		}

		runs = append(runs, run)
	}

	// Load error logs for the first run (if any)
	var errorLogs []db.FetchLog
	if len(runs) > 0 {
		m.selectedRun = runs[0].RunID
		errorLogs = m.loadErrorLogsForRun(runs[0].RunID)
	}

	return logsDataMsg{
		runs:      runs,
		errorLogs: errorLogs,
	}
}

// loadErrorLogs loads error logs for the selected run.
func (m LogsModel) loadErrorLogs() tea.Msg {
	logs := m.loadErrorLogsForRun(m.selectedRun)
	return errorLogsMsg{logs: logs}
}

// loadErrorLogsForRun loads error logs for a specific run ID.
func (m LogsModel) loadErrorLogsForRun(runID int64) []db.FetchLog {
	query := `
		SELECT run_id, symbol, from_dt, to_dt, rows, ok, msg, fetched_at
		FROM fetch_log
		WHERE run_id = ? AND ok = 0
		ORDER BY fetched_at DESC
	`

	rows, err := m.database.Query(query, runID)
	if err != nil {
		return []db.FetchLog{}
	}
	defer rows.Close()

	var logs []db.FetchLog
	for rows.Next() {
		var log db.FetchLog
		var ok int
		var fetchedAt string

		err := rows.Scan(
			&log.RunID,
			&log.Symbol,
			&log.FromDate,
			&log.ToDate,
			&log.Rows,
			&ok,
			&log.Message,
			&fetchedAt,
		)
		if err != nil {
			continue
		}

		log.OK = ok == 1
		log.FetchedAt, _ = time.Parse("2006-01-02 15:04:05", fetchedAt)

		logs = append(logs, log)
	}

	return logs
}

// logsDataMsg carries loaded logs data.
type logsDataMsg struct {
	runs      []db.Run
	errorLogs []db.FetchLog
}

// errorLogsMsg carries error logs for a specific run.
type errorLogsMsg struct {
	logs []db.FetchLog
}

// logsErrorMsg carries an error from data loading.
type logsErrorMsg struct {
	err error
}
