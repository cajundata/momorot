package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update all screens with new dimensions
		m.dashboard, _ = m.dashboard.Update(msg)
		m.leaders, _ = m.leaders.Update(msg)
		m.universe, _ = m.universe.Update(msg)
		m.symbol, _ = m.symbol.Update(msg)
		m.logs, _ = m.logs.Update(msg)

		return m, nil

	case refreshCompleteMsg:
		m.SetLoading(false, "")
		m.SetStatus("Refresh complete")
		return m, nil

	case refreshErrorMsg:
		m.SetLoading(false, "")
		m.SetError(string(msg))
		return m, nil

	case NavigateToSymbolMsg:
		m.NavigateToSymbol(msg.Symbol)
		return m, m.symbol.Init()
	}

	// Pass all other messages to the active screen
	switch m.currentScreen {
	case ScreenDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case ScreenLeaders:
		m.leaders, cmd = m.leaders.Update(msg)
	case ScreenUniverse:
		m.universe, cmd = m.universe.Update(msg)
	case ScreenSymbol:
		m.symbol, cmd = m.symbol.Update(msg)
	case ScreenLogs:
		m.logs, cmd = m.logs.Update(msg)
	}

	return m, cmd
}

// handleKeyPress processes keyboard input.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global key bindings
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Refresh):
		return m, m.triggerRefresh()

	case key.Matches(msg, m.keys.NextTab):
		return m.navigateNext(), nil

	case key.Matches(msg, m.keys.PrevTab):
		return m.navigatePrev(), nil

	case key.Matches(msg, m.keys.Back):
		m.NavigateBack()
		return m, nil
	}

	// Screen-specific key handling
	switch m.currentScreen {
	case ScreenDashboard:
		return m.updateDashboard(msg)
	case ScreenLeaders:
		return m.updateLeaders(msg)
	case ScreenUniverse:
		return m.updateUniverse(msg)
	case ScreenSymbol:
		return m.updateSymbol(msg)
	case ScreenLogs:
		return m.updateLogs(msg)
	}

	return m, nil
}

// navigateNext moves to the next screen in the tab order.
func (m Model) navigateNext() Model {
	switch m.currentScreen {
	case ScreenDashboard:
		m.currentScreen = ScreenLeaders
	case ScreenLeaders:
		m.currentScreen = ScreenUniverse
	case ScreenUniverse:
		m.currentScreen = ScreenSymbol
	case ScreenSymbol:
		m.currentScreen = ScreenLogs
	case ScreenLogs:
		m.currentScreen = ScreenDashboard
	}
	return m
}

// navigatePrev moves to the previous screen in the tab order.
func (m Model) navigatePrev() Model {
	switch m.currentScreen {
	case ScreenDashboard:
		m.currentScreen = ScreenLogs
	case ScreenLeaders:
		m.currentScreen = ScreenDashboard
	case ScreenUniverse:
		m.currentScreen = ScreenLeaders
	case ScreenSymbol:
		m.currentScreen = ScreenUniverse
	case ScreenLogs:
		m.currentScreen = ScreenSymbol
	}
	return m
}

// Screen-specific update functions that delegate to screen models.

func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.dashboard, cmd = m.dashboard.Update(msg)
	return m, cmd
}

func (m Model) updateLeaders(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.leaders, cmd = m.leaders.Update(msg)
	return m, cmd
}

func (m Model) updateUniverse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.universe, cmd = m.universe.Update(msg)
	return m, cmd
}

func (m Model) updateSymbol(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.symbol, cmd = m.symbol.Update(msg)
	return m, cmd
}

func (m Model) updateLogs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.logs, cmd = m.logs.Update(msg)
	return m, cmd
}

// Messages for async operations.

type refreshCompleteMsg struct{}

type refreshErrorMsg string

// triggerRefresh initiates a background refresh operation.
func (m Model) triggerRefresh() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement actual refresh logic
		// This is a placeholder
		return refreshCompleteMsg{}
	}
}
