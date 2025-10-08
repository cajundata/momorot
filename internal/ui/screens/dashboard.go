package screens

import (
	"fmt"
	"time"

	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DashboardModel represents the dashboard screen state.
type DashboardModel struct {
	database *db.DB
	theme    DashboardTheme

	// Dashboard data
	latestRun      *db.Run
	totalSymbols   int
	activeSymbols  int
	lastFetchDate  string
	apiQuotaUsed   int
	apiQuotaLimit  int
	nextResetTime  time.Time

	// UI state
	width  int
	height int
	ready  bool
	err    error
}

// DashboardTheme contains styling for the dashboard.
type DashboardTheme struct {
	CardStyle      lipgloss.Style
	CardTitle      lipgloss.Style
	CardValue      lipgloss.Style
	CardLabel      lipgloss.Style
	StatusOK       lipgloss.Style
	StatusError    lipgloss.Style
	StatusRunning  lipgloss.Style
	StatusNA       lipgloss.Style
}

// NewDashboard creates a new dashboard model.
func NewDashboard(database *db.DB, width, height int) DashboardModel {
	return DashboardModel{
		database:      database,
		width:         width,
		height:        height,
		theme:         defaultDashboardTheme(),
		apiQuotaLimit: 25, // Alpha Vantage free tier
		ready:         false,
	}
}

// defaultDashboardTheme returns the default dashboard theme.
func defaultDashboardTheme() DashboardTheme {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1, 2).
		Width(30).
		Height(6)

	return DashboardTheme{
		CardStyle: cardStyle,
		CardTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1),
		CardValue: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			MarginBottom(0),
		CardLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
		StatusOK: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true),
		StatusRunning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true),
		StatusNA: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
	}
}

// Init initializes the dashboard and loads data.
func (m DashboardModel) Init() tea.Cmd {
	return m.loadData
}

// Update handles messages for the dashboard.
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case dashboardDataMsg:
		m.latestRun = msg.run
		m.totalSymbols = msg.totalSymbols
		m.activeSymbols = msg.activeSymbols
		m.lastFetchDate = msg.lastFetchDate
		m.apiQuotaUsed = msg.apiQuotaUsed
		m.nextResetTime = msg.nextResetTime
		m.ready = true
		m.err = nil
		return m, nil

	case dashboardErrorMsg:
		m.err = msg.err
		m.ready = true
		return m, nil
	}

	return m, nil
}

// View renders the dashboard.
func (m DashboardModel) View() string {
	if m.err != nil {
		return m.theme.StatusError.Render(fmt.Sprintf("Error loading dashboard: %v", m.err))
	}

	if !m.ready {
		return m.theme.StatusNA.Render("Loading dashboard...")
	}

	// Build dashboard cards
	cards := []string{
		m.renderLastRunCard(),
		m.renderSymbolsCard(),
		m.renderCacheCard(),
		m.renderQuotaCard(),
	}

	// Arrange cards in 2x2 grid
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, cards[0], " ", cards[1])
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, cards[2], " ", cards[3])

	content := lipgloss.JoinVertical(lipgloss.Left, row1, "", row2)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

// renderLastRunCard shows the last run status.
func (m DashboardModel) renderLastRunCard() string {
	title := m.theme.CardTitle.Render("📊 Last Run")

	var content string
	if m.latestRun == nil {
		content = m.theme.StatusNA.Render("No runs yet")
	} else {
		// Status with color coding
		var status string
		switch m.latestRun.Status {
		case "OK":
			status = m.theme.StatusOK.Render("✓ OK")
		case "ERROR":
			status = m.theme.StatusError.Render("✗ ERROR")
		case "RUNNING":
			status = m.theme.StatusRunning.Render("◌ RUNNING")
		default:
			status = m.theme.StatusNA.Render("N/A")
		}

		// Run details
		runID := m.theme.CardLabel.Render(fmt.Sprintf("Run #%d", m.latestRun.RunID))
		timestamp := m.theme.CardLabel.Render(m.latestRun.StartedAt.Format("2006-01-02 15:04"))
		processed := m.theme.CardValue.Render(fmt.Sprintf("%d/%d",
			m.latestRun.SymbolsProcessed,
			m.latestRun.SymbolsProcessed+m.latestRun.SymbolsFailed))

		content = lipgloss.JoinVertical(lipgloss.Left,
			status,
			runID,
			timestamp,
			processed+" symbols",
		)
	}

	return m.theme.CardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// renderSymbolsCard shows symbol count statistics.
func (m DashboardModel) renderSymbolsCard() string {
	title := m.theme.CardTitle.Render("🌐 Universe")

	active := m.theme.CardValue.Render(fmt.Sprintf("%d", m.activeSymbols))
	total := m.theme.CardValue.Render(fmt.Sprintf("%d", m.totalSymbols))
	inactive := m.totalSymbols - m.activeSymbols

	content := lipgloss.JoinVertical(lipgloss.Left,
		active+" "+m.theme.CardLabel.Render("active"),
		total+" "+m.theme.CardLabel.Render("total"),
		"",
		m.theme.CardLabel.Render(fmt.Sprintf("(%d inactive)", inactive)),
	)

	return m.theme.CardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// renderCacheCard shows cache health metrics.
func (m DashboardModel) renderCacheCard() string {
	title := m.theme.CardTitle.Render("💾 Cache Health")

	var content string
	if m.lastFetchDate == "" {
		content = m.theme.StatusNA.Render("No data cached")
	} else {
		lastFetch, err := time.Parse("2006-01-02", m.lastFetchDate)
		var age string
		if err == nil {
			days := int(time.Since(lastFetch).Hours() / 24)
			age = fmt.Sprintf("%d days old", days)

			// Color code by age
			if days <= 1 {
				age = m.theme.StatusOK.Render(age)
			} else if days <= 7 {
				age = m.theme.CardLabel.Render(age)
			} else {
				age = m.theme.StatusError.Render(age)
			}
		}

		latest := m.theme.CardValue.Render(m.lastFetchDate)

		content = lipgloss.JoinVertical(lipgloss.Left,
			latest,
			m.theme.CardLabel.Render("latest data"),
			"",
			age,
		)
	}

	return m.theme.CardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// renderQuotaCard shows API quota status.
func (m DashboardModel) renderQuotaCard() string {
	title := m.theme.CardTitle.Render("📡 API Quota")

	remaining := m.apiQuotaLimit - m.apiQuotaUsed
	quotaText := fmt.Sprintf("%d/%d", m.apiQuotaUsed, m.apiQuotaLimit)

	var quotaStatus string
	if remaining > 10 {
		quotaStatus = m.theme.StatusOK.Render(quotaText)
	} else if remaining > 0 {
		quotaStatus = m.theme.StatusRunning.Render(quotaText)
	} else {
		quotaStatus = m.theme.StatusError.Render(quotaText)
	}

	// Next reset (daily reset at midnight UTC)
	nextReset := m.theme.CardLabel.Render("Resets daily at midnight UTC")

	content := lipgloss.JoinVertical(lipgloss.Left,
		quotaStatus,
		m.theme.CardLabel.Render("requests used"),
		"",
		nextReset,
	)

	return m.theme.CardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// loadData loads dashboard data from the database.
func (m DashboardModel) loadData() tea.Msg {
	// Get latest run
	runRepo := db.NewRunRepository(m.database)
	latestRun, err := runRepo.GetLatest()
	if err != nil && err.Error() != "sql: no rows in result set" {
		return dashboardErrorMsg{err: fmt.Errorf("failed to get latest run: %w", err)}
	}

	// Get symbol counts
	symbolRepo := db.NewSymbolRepository(m.database)
	activeSymbols, err := symbolRepo.ListActive()
	if err != nil {
		return dashboardErrorMsg{err: fmt.Errorf("failed to list symbols: %w", err)}
	}

	// Count total symbols (we need to add a method for this)
	// For now, we'll query directly
	var totalCount int
	err = m.database.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&totalCount)
	if err != nil {
		return dashboardErrorMsg{err: fmt.Errorf("failed to count symbols: %w", err)}
	}

	// Get latest fetch date across all symbols
	var lastFetch string
	query := `
		SELECT COALESCE(MAX(date), '')
		FROM prices
		WHERE symbol IN (SELECT symbol FROM symbols WHERE active = 1)
	`
	err = m.database.QueryRow(query).Scan(&lastFetch)
	if err != nil {
		return dashboardErrorMsg{err: fmt.Errorf("failed to get latest fetch date: %w", err)}
	}

	// API quota (for now, hardcoded - would need to track this in DB)
	apiQuotaUsed := 0
	if latestRun != nil {
		apiQuotaUsed = latestRun.SymbolsProcessed
	}

	return dashboardDataMsg{
		run:            latestRun,
		totalSymbols:   totalCount,
		activeSymbols:  len(activeSymbols),
		lastFetchDate:  lastFetch,
		apiQuotaUsed:   apiQuotaUsed,
		nextResetTime:  time.Now().Add(24 * time.Hour), // Placeholder
	}
}

// dashboardDataMsg carries loaded dashboard data.
type dashboardDataMsg struct {
	run            *db.Run
	totalSymbols   int
	activeSymbols  int
	lastFetchDate  string
	apiQuotaUsed   int
	nextResetTime  time.Time
}

// dashboardErrorMsg carries an error from data loading.
type dashboardErrorMsg struct {
	err error
}
