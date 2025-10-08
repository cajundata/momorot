package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case refreshCompleteMsg:
		m.SetLoading(false, "")
		m.SetStatus("Refresh complete")
		return m, nil

	case refreshErrorMsg:
		m.SetLoading(false, "")
		m.SetError(string(msg))
		return m, nil
	}

	return m, nil
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

// Screen-specific update functions (stubs for now).

func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement dashboard-specific key handling
	return m, nil
}

func (m Model) updateLeaders(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement leaders-specific key handling
	return m, nil
}

func (m Model) updateUniverse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement universe-specific key handling
	return m, nil
}

func (m Model) updateSymbol(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement symbol-specific key handling
	return m, nil
}

func (m Model) updateLogs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement logs-specific key handling
	return m, nil
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
