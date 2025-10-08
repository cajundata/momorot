package screens

import (
	"fmt"
	"strings"

	"github.com/cajundata/momorot/internal/db"
	"github.com/cajundata/momorot/internal/ui/components"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UniverseModel represents the universe screen state.
type UniverseModel struct {
	database *db.DB
	table    components.TableModel
	search   components.SearchModel
	theme    UniverseTheme

	// Screen data
	symbols        []SymbolWithFetch
	filteredSymbols []SymbolWithFetch

	// UI state
	searchMode     bool
	width          int
	height         int
	ready          bool
	err            error
}

// SymbolWithFetch combines symbol info with last fetch date.
type SymbolWithFetch struct {
	Symbol       string
	Name         string
	AssetType    string
	Active       bool
	LastFetchDate string
}

// UniverseTheme contains styling for the universe screen.
type UniverseTheme struct {
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	ActiveBadge  lipgloss.Style
	InactiveBadge lipgloss.Style
	SearchPrompt lipgloss.Style
	Help         lipgloss.Style
	EmptyMsg     lipgloss.Style
}

// NewUniverse creates a new universe model.
func NewUniverse(database *db.DB, width, height int) UniverseModel {
	// Create table
	columns := []table.Column{
		{Title: "Symbol", Width: 12},
		{Title: "Name", Width: 30},
		{Title: "Type", Width: 10},
		{Title: "Status", Width: 12},
		{Title: "Last Fetch", Width: 15},
	}

	tableModel := components.NewTable(columns, []table.Row{}, width-4, height-10)

	// Create search input
	searchModel := components.NewSearch("Search symbols...")

	return UniverseModel{
		database:   database,
		table:      tableModel,
		search:     searchModel,
		theme:      defaultUniverseTheme(),
		searchMode: false,
		width:      width,
		height:     height,
		ready:      false,
	}
}

// defaultUniverseTheme returns the default universe theme.
func defaultUniverseTheme() UniverseTheme {
	return UniverseTheme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			MarginBottom(1),
		ActiveBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true),
		InactiveBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
		SearchPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		EmptyMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Padding(2, 4),
	}
}

// Init initializes the universe screen and loads data.
func (m UniverseModel) Init() tea.Cmd {
	return m.loadSymbols
}

// Update handles messages for the universe screen.
func (m UniverseModel) Update(msg tea.Msg) (UniverseModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 10)
		return m, nil

	case tea.KeyMsg:
		// Handle key presses
		if m.searchMode {
			// In search mode
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.search.Blur()
				m.search.Reset()
				m.filterSymbols("")
				return m, nil
			case "enter":
				m.searchMode = false
				m.search.Blur()
				return m, nil
			default:
				// Pass to search input
				m.search, cmd = m.search.Update(msg)
				m.filterSymbols(m.search.Value())
				return m, cmd
			}
		} else {
			// Normal mode
			switch msg.String() {
			case "/":
				m.searchMode = true
				return m, m.search.Focus()
			case " ":
				// Toggle active/inactive for selected symbol
				return m, m.toggleActive()
			case "enter":
				// Navigate to symbol detail
				if m.ready && len(m.filteredSymbols) > 0 {
					cursor := m.table.Cursor()
					if cursor < len(m.filteredSymbols) {
						symbol := m.filteredSymbols[cursor].Symbol
						return m, func() tea.Msg {
							return navigateToSymbolMsg{symbol: symbol}
						}
					}
				}
			}
		}

	case universeDataMsg:
		m.symbols = msg.symbols
		m.filteredSymbols = msg.symbols
		m.ready = true
		m.err = nil
		m.updateTableRows()
		return m, nil

	case universeErrorMsg:
		m.err = msg.err
		m.ready = true
		return m, nil

	case symbolToggledMsg:
		// Reload symbols after toggle
		return m, m.loadSymbols
	}

	// Pass through to table for navigation (only if not in search mode)
	if !m.searchMode {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

// View renders the universe screen.
func (m UniverseModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading universe: %v", m.err))
	}

	if !m.ready {
		return m.theme.EmptyMsg.Render("Loading symbol universe...")
	}

	if len(m.symbols) == 0 {
		return m.theme.EmptyMsg.Render("No symbols in universe.\nAdd symbols to begin tracking.")
	}

	// Header
	title := m.theme.Title.Render("🌐 Symbol Universe")
	subtitle := m.theme.Subtitle.Render(fmt.Sprintf(
		"Showing %d of %d symbols",
		len(m.filteredSymbols),
		len(m.symbols),
	))

	// Search bar (if in search mode)
	var searchBar string
	if m.searchMode {
		searchBar = m.theme.SearchPrompt.Render("Search: ") + m.search.View()
	}

	// Table
	tableView := m.table.View()

	// Help text
	help := m.theme.Help.Render("↑/↓: Navigate | Space: Toggle Active | /: Search | Enter: Details")

	content := []string{title, subtitle}
	if searchBar != "" {
		content = append(content, "", searchBar)
	}
	content = append(content, "", tableView, "", help)

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// updateTableRows updates the table with symbol data.
func (m *UniverseModel) updateTableRows() {
	rows := make([]table.Row, 0, len(m.filteredSymbols))

	for _, sym := range m.filteredSymbols {
		// Format status badge
		var status string
		if sym.Active {
			status = m.theme.ActiveBadge.Render("● Active")
		} else {
			status = m.theme.InactiveBadge.Render("○ Inactive")
		}

		// Format last fetch
		lastFetch := sym.LastFetchDate
		if lastFetch == "" {
			lastFetch = m.theme.Help.Render("Never")
		}

		rows = append(rows, table.Row{
			sym.Symbol,
			sym.Name,
			sym.AssetType,
			status,
			lastFetch,
		})
	}

	m.table.SetRows(rows)
}

// filterSymbols filters the symbol list based on search query.
func (m *UniverseModel) filterSymbols(query string) {
	if query == "" {
		m.filteredSymbols = m.symbols
		m.updateTableRows()
		return
	}

	query = strings.ToLower(query)
	filtered := make([]SymbolWithFetch, 0)

	for _, sym := range m.symbols {
		if strings.Contains(strings.ToLower(sym.Symbol), query) ||
			strings.Contains(strings.ToLower(sym.Name), query) {
			filtered = append(filtered, sym)
		}
	}

	m.filteredSymbols = filtered
	m.updateTableRows()
}

// toggleActive toggles the active status of the currently selected symbol.
func (m UniverseModel) toggleActive() tea.Cmd {
	if !m.ready || len(m.filteredSymbols) == 0 {
		return nil
	}

	cursor := m.table.Cursor()
	if cursor >= len(m.filteredSymbols) {
		return nil
	}

	symbol := m.filteredSymbols[cursor]

	return func() tea.Msg {
		// Update symbol in database
		symbolRepo := db.NewSymbolRepository(m.database)
		dbSymbol, err := symbolRepo.Get(symbol.Symbol)
		if err != nil {
			return universeErrorMsg{err: fmt.Errorf("failed to get symbol: %w", err)}
		}

		// Toggle active status
		dbSymbol.Active = !dbSymbol.Active

		err = symbolRepo.Update(dbSymbol)
		if err != nil {
			return universeErrorMsg{err: fmt.Errorf("failed to update symbol: %w", err)}
		}

		return symbolToggledMsg{symbol: symbol.Symbol, active: dbSymbol.Active}
	}
}

// loadSymbols loads all symbols from the database with last fetch dates.
func (m UniverseModel) loadSymbols() tea.Msg {
	// Get all symbols (not just active)
	var symbols []db.Symbol
	query := `SELECT symbol, name, asset_type, active, created_at, updated_at FROM symbols ORDER BY symbol`
	rows, err := m.database.Query(query)
	if err != nil {
		return universeErrorMsg{err: fmt.Errorf("failed to query symbols: %w", err)}
	}
	defer rows.Close()

	for rows.Next() {
		var sym db.Symbol
		var active int
		var createdAt, updatedAt string
		err := rows.Scan(&sym.Symbol, &sym.Name, &sym.AssetType, &active, &createdAt, &updatedAt)
		if err != nil {
			return universeErrorMsg{err: fmt.Errorf("failed to scan symbol: %w", err)}
		}
		sym.Active = active == 1
		// Note: created_at and updated_at are scanned as strings but not used in display
		symbols = append(symbols, sym)
	}

	// Get last fetch date for each symbol
	result := make([]SymbolWithFetch, 0, len(symbols))
	for _, sym := range symbols {
		var lastFetch string
		query := `SELECT COALESCE(MAX(date), '') FROM prices WHERE symbol = ?`
		err := m.database.QueryRow(query, sym.Symbol).Scan(&lastFetch)
		if err != nil {
			return universeErrorMsg{err: fmt.Errorf("failed to get last fetch date: %w", err)}
		}

		result = append(result, SymbolWithFetch{
			Symbol:        sym.Symbol,
			Name:          sym.Name,
			AssetType:     sym.AssetType,
			Active:        sym.Active,
			LastFetchDate: lastFetch,
		})
	}

	return universeDataMsg{symbols: result}
}

// universeDataMsg carries loaded symbol data.
type universeDataMsg struct {
	symbols []SymbolWithFetch
}

// universeErrorMsg carries an error from data loading.
type universeErrorMsg struct {
	err error
}

// symbolToggledMsg signals that a symbol's active status was toggled.
type symbolToggledMsg struct {
	symbol string
	active bool
}
