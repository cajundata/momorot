package screens

import (
	"fmt"

	"github.com/cajundata/momorot/internal/db"
	"github.com/cajundata/momorot/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SymbolModel represents the symbol detail screen state.
type SymbolModel struct {
	database  *db.DB
	sparkline components.SparklineModel
	theme     SymbolTheme

	// Screen data
	symbol     string
	symbolInfo *db.Symbol
	prices     []db.Price
	indicators *db.Indicator
	rank       int

	// UI state
	width  int
	height int
	ready  bool
	err    error
}

// SymbolTheme contains styling for the symbol detail screen.
type SymbolTheme struct {
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	SectionTitle lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	Positive    lipgloss.Style
	Negative    lipgloss.Style
	Neutral     lipgloss.Style
	Card        lipgloss.Style
	EmptyMsg    lipgloss.Style
}

// NewSymbol creates a new symbol detail model.
func NewSymbol(database *db.DB, symbol string, width, height int) SymbolModel {
	// Create sparkline (will be populated with data later)
	sparkline := components.NewSparkline([]float64{}, 60, 5)

	return SymbolModel{
		database:  database,
		sparkline: sparkline,
		theme:     defaultSymbolTheme(),
		symbol:    symbol,
		width:     width,
		height:    height,
		ready:     false,
	}
}

// defaultSymbolTheme returns the default symbol theme.
func defaultSymbolTheme() SymbolTheme {
	return SymbolTheme{
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
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		Value: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")),
		Positive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")),
		Negative: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9")),
		Neutral: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 2).
			MarginRight(2),
		EmptyMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Padding(2, 4),
	}
}

// Init initializes the symbol detail screen and loads data.
func (m SymbolModel) Init() tea.Cmd {
	return m.loadSymbolData
}

// Update handles messages for the symbol detail screen.
func (m SymbolModel) Update(msg tea.Msg) (SymbolModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case symbolDataMsg:
		m.symbolInfo = msg.symbolInfo
		m.prices = msg.prices
		m.indicators = msg.indicators
		m.rank = msg.rank
		m.ready = true
		m.err = nil

		// Update sparkline with price data
		if len(m.prices) > 0 {
			priceData := make([]float64, len(m.prices))
			for i, p := range m.prices {
				if p.AdjClose != nil {
					priceData[i] = *p.AdjClose
				} else {
					priceData[i] = p.Close
				}
			}
			m.sparkline.SetData(priceData)
		}

		return m, nil

	case symbolErrorMsg:
		m.err = msg.err
		m.ready = true
		return m, nil
	}

	return m, nil
}

// View renders the symbol detail screen.
func (m SymbolModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading symbol: %v", m.err))
	}

	if !m.ready {
		return m.theme.EmptyMsg.Render(fmt.Sprintf("Loading %s details...", m.symbol))
	}

	if m.symbolInfo == nil {
		return m.theme.EmptyMsg.Render(fmt.Sprintf("Symbol %s not found.", m.symbol))
	}

	// Header
	title := m.theme.Title.Render(fmt.Sprintf("📈 %s - %s", m.symbol, m.symbolInfo.Name))
	subtitle := m.theme.Subtitle.Render(fmt.Sprintf("%s | Rank: %s",
		m.symbolInfo.AssetType,
		m.formatRank()))

	// Price chart section
	chartSection := m.renderChartSection()

	// Metrics section
	metricsSection := m.renderMetricsSection()

	// Volatility section
	volSection := m.renderVolatilitySection()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		chartSection,
		"",
		metricsSection,
		"",
		volSection,
	)
}

// renderChartSection renders the price chart section.
func (m SymbolModel) renderChartSection() string {
	if len(m.prices) == 0 {
		return m.theme.EmptyMsg.Render("No price data available")
	}

	sectionTitle := m.theme.SectionTitle.Render("📊 Price Chart (90 days)")

	// Get stats from sparkline
	stats := m.sparkline.GetStats()

	// Format stats
	statsText := fmt.Sprintf(
		"%s: $%.2f  %s: $%.2f  %s: $%.2f  %s: %s",
		m.theme.Label.Render("First"),
		stats.First,
		m.theme.Label.Render("Last"),
		stats.Last,
		m.theme.Label.Render("Range"),
		stats.Max-stats.Min,
		m.theme.Label.Render("Change"),
		m.formatChange(stats.Change, stats.PctChange),
	)

	// Sparkline chart
	chart := m.sparkline.ViewWithBorder()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		sectionTitle,
		chart,
		statsText,
	)
}

// renderMetricsSection renders the return metrics section.
func (m SymbolModel) renderMetricsSection() string {
	sectionTitle := m.theme.SectionTitle.Render("📈 Return Metrics")

	if m.indicators == nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			sectionTitle,
			m.theme.EmptyMsg.Render("No indicator data available"),
		)
	}

	// Build metrics cards
	cards := []string{
		m.renderMetricCard("1M Return", m.indicators.R1M),
		m.renderMetricCard("3M Return", m.indicators.R3M),
		m.renderMetricCard("6M Return", m.indicators.R6M),
		m.renderMetricCard("12M Return", m.indicators.R12M),
	}

	cardsRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		sectionTitle,
		cardsRow,
	)
}

// renderVolatilitySection renders the volatility metrics section.
func (m SymbolModel) renderVolatilitySection() string {
	sectionTitle := m.theme.SectionTitle.Render("📉 Volatility & Liquidity")

	if m.indicators == nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			sectionTitle,
			m.theme.EmptyMsg.Render("No indicator data available"),
		)
	}

	cards := []string{
		m.renderVolCard("3M Volatility", m.indicators.Vol3M),
		m.renderVolCard("6M Volatility", m.indicators.Vol6M),
		m.renderADVCard("Avg Daily Vol", m.indicators.ADV),
		m.renderScoreCard("Momentum Score", m.indicators.Score),
	}

	cardsRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		sectionTitle,
		cardsRow,
	)
}

// renderMetricCard renders a return metric card.
func (m SymbolModel) renderMetricCard(label string, value *float64) string {
	if value == nil {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.theme.Label.Render(label),
			m.theme.Neutral.Render("N/A"),
		)
		return m.theme.Card.Render(content)
	}

	pct := *value * 100
	var valueStyle lipgloss.Style
	if *value > 0 {
		valueStyle = m.theme.Positive
	} else if *value < 0 {
		valueStyle = m.theme.Negative
	} else {
		valueStyle = m.theme.Neutral
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Label.Render(label),
		valueStyle.Render(fmt.Sprintf("%+.2f%%", pct)),
	)

	return m.theme.Card.Render(content)
}

// renderVolCard renders a volatility metric card.
func (m SymbolModel) renderVolCard(label string, value *float64) string {
	if value == nil {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.theme.Label.Render(label),
			m.theme.Neutral.Render("N/A"),
		)
		return m.theme.Card.Render(content)
	}

	pct := *value * 100
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Label.Render(label),
		m.theme.Value.Render(fmt.Sprintf("%.2f%%", pct)),
	)

	return m.theme.Card.Render(content)
}

// renderADVCard renders the average daily volume card.
func (m SymbolModel) renderADVCard(label string, value *float64) string {
	if value == nil {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.theme.Label.Render(label),
			m.theme.Neutral.Render("N/A"),
		)
		return m.theme.Card.Render(content)
	}

	formatted := m.formatLargeNumber(*value)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Label.Render(label),
		m.theme.Value.Render(formatted),
	)

	return m.theme.Card.Render(content)
}

// renderScoreCard renders the momentum score card.
func (m SymbolModel) renderScoreCard(label string, value *float64) string {
	if value == nil {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.theme.Label.Render(label),
			m.theme.Neutral.Render("N/A"),
		)
		return m.theme.Card.Render(content)
	}

	var valueStyle lipgloss.Style
	if *value > 0 {
		valueStyle = m.theme.Positive
	} else if *value < 0 {
		valueStyle = m.theme.Negative
	} else {
		valueStyle = m.theme.Neutral
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Label.Render(label),
		valueStyle.Render(fmt.Sprintf("%.3f", *value)),
	)

	return m.theme.Card.Render(content)
}

// formatRank formats the rank display.
func (m SymbolModel) formatRank() string {
	if m.rank <= 0 {
		return m.theme.Neutral.Render("Unranked")
	}
	return m.theme.Value.Render(fmt.Sprintf("#%d", m.rank))
}

// formatChange formats a price change with color coding.
func (m SymbolModel) formatChange(change, pctChange float64) string {
	var style lipgloss.Style
	if change > 0 {
		style = m.theme.Positive
	} else if change < 0 {
		style = m.theme.Negative
	} else {
		style = m.theme.Neutral
	}

	return style.Render(fmt.Sprintf("$%+.2f (%+.2f%%)", change, pctChange))
}

// formatLargeNumber formats large numbers with K/M/B suffixes.
func (m SymbolModel) formatLargeNumber(value float64) string {
	switch {
	case value >= 1e9:
		return fmt.Sprintf("$%.2fB", value/1e9)
	case value >= 1e6:
		return fmt.Sprintf("$%.2fM", value/1e6)
	case value >= 1e3:
		return fmt.Sprintf("$%.2fK", value/1e3)
	default:
		return fmt.Sprintf("$%.2f", value)
	}
}

// loadSymbolData loads all data for the symbol.
func (m SymbolModel) loadSymbolData() tea.Msg {
	// Get symbol info
	symbolRepo := db.NewSymbolRepository(m.database)
	symbolInfo, err := symbolRepo.Get(m.symbol)
	if err != nil {
		return symbolErrorMsg{err: fmt.Errorf("failed to get symbol: %w", err)}
	}

	// Get latest 90 days of price data
	var latestDate string
	err = m.database.QueryRow("SELECT COALESCE(MAX(date), '') FROM prices WHERE symbol = ?", m.symbol).Scan(&latestDate)
	if err != nil {
		return symbolErrorMsg{err: fmt.Errorf("failed to get latest date: %w", err)}
	}

	var prices []db.Price
	if latestDate != "" {
		// Get last 90 trading days
		query := `
			SELECT symbol, date, open, high, low, close, adj_close, volume, created_at
			FROM prices
			WHERE symbol = ?
			ORDER BY date DESC
			LIMIT 90
		`
		rows, err := m.database.Query(query, m.symbol)
		if err != nil {
			return symbolErrorMsg{err: fmt.Errorf("failed to query prices: %w", err)}
		}
		defer rows.Close()

		for rows.Next() {
			var p db.Price
			var adjClose *float64
			var vol *int64
			var createdAt string
			err := rows.Scan(&p.Symbol, &p.Date, &p.Open, &p.High, &p.Low, &p.Close, &adjClose, &vol, &createdAt)
			if err != nil {
				return symbolErrorMsg{err: fmt.Errorf("failed to scan price: %w", err)}
			}
			p.AdjClose = adjClose
			if vol != nil {
				p.Volume = vol
			}
			prices = append(prices, p)
		}

		// Reverse to get chronological order
		for i, j := 0, len(prices)-1; i < j; i, j = i+1, j-1 {
			prices[i], prices[j] = prices[j], prices[i]
		}
	}

	// Get latest indicators
	var indicators *db.Indicator
	if latestDate != "" {
		query := `
			SELECT symbol, date, r_1m, r_3m, r_6m, r_12m, vol_3m, vol_6m, adv, score, rank, created_at
			FROM indicators
			WHERE symbol = ? AND date = ?
		`
		var ind db.Indicator
		var createdAt string
		err := m.database.QueryRow(query, m.symbol, latestDate).Scan(
			&ind.Symbol, &ind.Date, &ind.R1M, &ind.R3M, &ind.R6M, &ind.R12M,
			&ind.Vol3M, &ind.Vol6M, &ind.ADV, &ind.Score, &ind.Rank, &createdAt,
		)
		if err == nil {
			indicators = &ind
		}
		// If no indicators, that's okay - just means they haven't been computed yet
	}

	// Get rank from indicators
	rank := 0
	if indicators != nil && indicators.Rank != nil {
		rank = *indicators.Rank
	}

	return symbolDataMsg{
		symbolInfo: symbolInfo,
		prices:     prices,
		indicators: indicators,
		rank:       rank,
	}
}

// symbolDataMsg carries loaded symbol data.
type symbolDataMsg struct {
	symbolInfo *db.Symbol
	prices     []db.Price
	indicators *db.Indicator
	rank       int
}

// symbolErrorMsg carries an error from data loading.
type symbolErrorMsg struct {
	err error
}
