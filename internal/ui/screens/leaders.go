package screens

import (
	"fmt"

	"github.com/cajundata/momorot/internal/db"
	"github.com/cajundata/momorot/internal/ui/components"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LeadersModel represents the leaders screen state.
type LeadersModel struct {
	database *db.DB
	table    components.TableModel
	theme    LeadersTheme

	// Screen data
	leaders        []db.Indicator
	topN           int
	selectedSymbol string

	// UI state
	width  int
	height int
	ready  bool
	err    error
}

// LeadersTheme contains styling for the leaders screen.
type LeadersTheme struct {
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Positive  lipgloss.Style
	Negative  lipgloss.Style
	Neutral   lipgloss.Style
	EmptyMsg  lipgloss.Style
}

// NewLeaders creates a new leaders model.
func NewLeaders(database *db.DB, width, height int) LeadersModel {
	// Create empty table initially
	columns := []table.Column{
		{Title: "Rank", Width: 6},
		{Title: "Symbol", Width: 10},
		{Title: "Score", Width: 10},
		{Title: "R1M", Width: 10},
		{Title: "R3M", Width: 10},
		{Title: "R6M", Width: 10},
		{Title: "Vol", Width: 10},
		{Title: "ADV", Width: 12},
	}

	tableModel := components.NewTable(columns, []table.Row{}, width-4, height-8)

	return LeadersModel{
		database: database,
		table:    tableModel,
		theme:    defaultLeadersTheme(),
		topN:     10, // Default to top 10
		width:    width,
		height:   height,
		ready:    false,
	}
}

// defaultLeadersTheme returns the default leaders theme.
func defaultLeadersTheme() LeadersTheme {
	return LeadersTheme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			MarginBottom(1),
		Positive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),
		Negative: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),
		Neutral: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		EmptyMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Padding(2, 4),
	}
}

// Init initializes the leaders screen and loads data.
func (m LeadersModel) Init() tea.Cmd {
	return m.loadLeaders
}

// Update handles messages for the leaders screen.
func (m LeadersModel) Update(msg tea.Msg) (LeadersModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 8)
		return m, nil

	case tea.KeyMsg:
		// Handle key presses
		switch msg.String() {
		case "enter":
			// Navigate to symbol detail for selected row
			if m.ready && len(m.leaders) > 0 {
				selected := m.table.SelectedRow()
				if len(selected) > 1 {
					m.selectedSymbol = selected[1] // Symbol is column 1
					return m, func() tea.Msg {
						return navigateToSymbolMsg{symbol: m.selectedSymbol}
					}
				}
			}
		}

	case leadersDataMsg:
		m.leaders = msg.leaders
		m.ready = true
		m.err = nil

		// Update table with data
		m.updateTableRows()
		return m, nil

	case leadersErrorMsg:
		m.err = msg.err
		m.ready = true
		return m, nil
	}

	// Pass through to table for navigation
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the leaders screen.
func (m LeadersModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading leaders: %v", m.err))
	}

	if !m.ready {
		return m.theme.EmptyMsg.Render("Loading top performers...")
	}

	if len(m.leaders) == 0 {
		return m.theme.EmptyMsg.Render("No ranking data available.\nRun a refresh to compute momentum indicators.")
	}

	// Header
	title := m.theme.Title.Render("🏆 Top Leaders")
	subtitle := m.theme.Subtitle.Render(fmt.Sprintf("Showing top %d momentum leaders", len(m.leaders)))

	// Table
	tableView := m.table.View()

	// Help text
	help := m.theme.Neutral.Render("↑/↓: Navigate | Enter: View Details | r: Refresh")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		tableView,
		"",
		help,
	)
}

// updateTableRows updates the table with leader data.
func (m *LeadersModel) updateTableRows() {
	rows := make([]table.Row, 0, len(m.leaders))

	for _, leader := range m.leaders {
		// Format values
		rank := fmt.Sprintf("#%d", valueOrZero(leader.Rank))
		symbol := leader.Symbol
		score := m.formatValue(leader.Score, true)
		r1m := m.formatPercent(leader.R1M)
		r3m := m.formatPercent(leader.R3M)
		r6m := m.formatPercent(leader.R6M)
		vol := m.formatPercent(leader.Vol3M)
		adv := m.formatLargeNumber(leader.ADV)

		rows = append(rows, table.Row{
			rank,
			symbol,
			score,
			r1m,
			r3m,
			r6m,
			vol,
			adv,
		})
	}

	m.table.SetRows(rows)
}

// formatValue formats a float pointer with color coding.
func (m LeadersModel) formatValue(val *float64, colorize bool) string {
	if val == nil {
		return m.theme.Neutral.Render("N/A")
	}

	formatted := fmt.Sprintf("%.3f", *val)

	if !colorize {
		return formatted
	}

	if *val > 0 {
		return m.theme.Positive.Render(formatted)
	} else if *val < 0 {
		return m.theme.Negative.Render(formatted)
	}
	return formatted
}

// formatPercent formats a percentage value with color coding.
func (m LeadersModel) formatPercent(val *float64) string {
	if val == nil {
		return m.theme.Neutral.Render("N/A")
	}

	formatted := fmt.Sprintf("%.2f%%", *val*100)

	if *val > 0 {
		return m.theme.Positive.Render(formatted)
	} else if *val < 0 {
		return m.theme.Negative.Render(formatted)
	}
	return formatted
}

// formatLargeNumber formats large numbers with K/M/B suffixes.
func (m LeadersModel) formatLargeNumber(val *float64) string {
	if val == nil {
		return m.theme.Neutral.Render("N/A")
	}

	v := *val
	switch {
	case v >= 1e9:
		return fmt.Sprintf("$%.2fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("$%.2fM", v/1e6)
	case v >= 1e3:
		return fmt.Sprintf("$%.2fK", v/1e3)
	default:
		return fmt.Sprintf("$%.2f", v)
	}
}

// valueOrZero returns the value or 0 if nil.
func valueOrZero(val *int) int {
	if val == nil {
		return 0
	}
	return *val
}

// loadLeaders loads the top N leaders from the database.
func (m LeadersModel) loadLeaders() tea.Msg {
	// Get the latest date with indicator data
	var latestDate string
	query := `
		SELECT COALESCE(MAX(date), '')
		FROM indicators
	`
	err := m.database.QueryRow(query).Scan(&latestDate)
	if err != nil {
		return leadersErrorMsg{err: fmt.Errorf("failed to find latest indicator date: %w", err)}
	}

	if latestDate == "" {
		// No indicator data yet
		return leadersDataMsg{leaders: []db.Indicator{}}
	}

	// Get top N indicators for that date
	indicatorRepo := db.NewIndicatorRepository(m.database)
	leaders, err := indicatorRepo.GetTopN(latestDate, m.topN)
	if err != nil {
		return leadersErrorMsg{err: fmt.Errorf("failed to get top leaders: %w", err)}
	}

	return leadersDataMsg{leaders: leaders}
}

// leadersDataMsg carries loaded leaders data.
type leadersDataMsg struct {
	leaders []db.Indicator
}

// leadersErrorMsg carries an error from data loading.
type leadersErrorMsg struct {
	err error
}

// navigateToSymbolMsg signals navigation to symbol detail screen.
type navigateToSymbolMsg struct {
	symbol string
}
