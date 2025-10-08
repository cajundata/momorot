package ui

import (
	"github.com/cajundata/momorot/internal/analytics"
	"github.com/cajundata/momorot/internal/config"
	"github.com/cajundata/momorot/internal/db"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents the different screens in the application.
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenLeaders
	ScreenUniverse
	ScreenSymbol
	ScreenLogs
)

// Model represents the main application state.
type Model struct {
	// Core dependencies
	db           *db.DB
	orchestrator *analytics.Orchestrator
	config       *config.Config

	// Navigation state
	currentScreen Screen
	screenHistory []Screen // For back navigation

	// Screen-specific models
	dashboard interface{} // Will be dashboardModel
	leaders   interface{} // Will be leadersModel
	universe  interface{} // Will be universeModel
	symbol    interface{} // Will be symbolModel
	logs      interface{} // Will be logsModel

	// Global UI state
	loading      bool
	loadingMsg   string
	errorMsg     string
	width        int
	height       int
	statusBarMsg string

	// Key bindings and theme
	keys  KeyBindings
	theme Theme
}


// New creates a new Model with the given dependencies.
func New(database *db.DB, cfg *config.Config) Model {
	// Create orchestrator
	orchestrator := analytics.NewOrchestrator(
		database,
		map[string]int{
			"r1m":  cfg.Lookbacks.R1M,
			"r3m":  cfg.Lookbacks.R3M,
			"r6m":  cfg.Lookbacks.R6M,
			"r12m": cfg.Lookbacks.R12M,
		},
		map[string]int{
			"short": cfg.VolWindows.Short,
			"long":  cfg.VolWindows.Long,
		},
		analytics.ScoringConfig{
			PenaltyLambda:      cfg.Scoring.PenaltyLambda,
			MinADV:             cfg.Scoring.MinADVUSD,
			BreadthMinPositive: cfg.Scoring.BreadthMinPositive,
			BreadthTotal:       cfg.Scoring.BreadthTotalLookbacks,
		},
	)

	return Model{
		db:            database,
		orchestrator:  orchestrator,
		config:        cfg,
		currentScreen: ScreenDashboard,
		screenHistory: []Screen{},
		keys:          DefaultKeyBindings(),
		theme:         DefaultTheme(),
		loading:       false,
		width:         80,
		height:        24,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// NavigateTo changes the current screen.
func (m *Model) NavigateTo(screen Screen) {
	m.screenHistory = append(m.screenHistory, m.currentScreen)
	m.currentScreen = screen
}

// NavigateBack returns to the previous screen.
func (m *Model) NavigateBack() {
	if len(m.screenHistory) > 0 {
		m.currentScreen = m.screenHistory[len(m.screenHistory)-1]
		m.screenHistory = m.screenHistory[:len(m.screenHistory)-1]
	}
}

// SetLoading sets the loading state with an optional message.
func (m *Model) SetLoading(loading bool, msg string) {
	m.loading = loading
	m.loadingMsg = msg
}

// SetError sets an error message to display.
func (m *Model) SetError(msg string) {
	m.errorMsg = msg
}

// ClearError clears the error message.
func (m *Model) ClearError() {
	m.errorMsg = ""
}

// SetStatus sets the status bar message.
func (m *Model) SetStatus(msg string) {
	m.statusBarMsg = msg
}
