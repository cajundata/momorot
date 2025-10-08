package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (m Model) View() string {
	// Build the screen content
	var content string

	switch m.currentScreen {
	case ScreenDashboard:
		content = m.viewDashboard()
	case ScreenLeaders:
		content = m.viewLeaders()
	case ScreenUniverse:
		content = m.viewUniverse()
	case ScreenSymbol:
		content = m.viewSymbol()
	case ScreenLogs:
		content = m.viewLogs()
	default:
		content = "Unknown screen"
	}

	// Add header
	header := m.renderHeader()

	// Add status bar
	statusBar := m.renderStatusBar()

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
	)
}

// renderHeader renders the application header with navigation tabs.
func (m Model) renderHeader() string {
	tabs := []string{
		m.renderTab("Dashboard", ScreenDashboard),
		m.renderTab("Leaders", ScreenLeaders),
		m.renderTab("Universe", ScreenUniverse),
		m.renderTab("Symbol", ScreenSymbol),
		m.renderTab("Logs", ScreenLogs),
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	title := m.theme.Title.Render("Momentum Screener")

	separator := lipgloss.NewStyle().
		Foreground(m.theme.BorderColor).
		Render(strings.Repeat("─", m.width))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		tabBar,
		separator,
	)
}

// renderTab renders a single tab with active/inactive styling.
func (m Model) renderTab(label string, screen Screen) string {
	if m.currentScreen == screen {
		return m.theme.TabActive.Render(label)
	}
	return m.theme.TabInactive.Render(label)
}

// renderStatusBar renders the bottom status bar with key bindings and messages.
func (m Model) renderStatusBar() string {
	// Loading indicator
	if m.loading {
		msg := "Loading"
		if m.loadingMsg != "" {
			msg = m.loadingMsg
		}
		return m.theme.Loading.Render(fmt.Sprintf("⏳ %s...", msg))
	}

	// Error message
	if m.errorMsg != "" {
		return m.theme.StatusBarErr.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
	}

	// Status message
	if m.statusBarMsg != "" {
		return m.theme.StatusBarInfo.Render(m.statusBarMsg)
	}

	// Default help text with current screen name
	screenName := m.getScreenName()
	helpText := fmt.Sprintf("[%s] ", screenName)

	for i, binding := range m.keys.ShortHelp() {
		if i > 0 {
			helpText += " | "
		}
		helpText += binding.Help().Key + ": " + binding.Help().Desc
	}

	return m.theme.Help.Render(helpText)
}

// getScreenName returns the name of the current screen.
func (m Model) getScreenName() string {
	switch m.currentScreen {
	case ScreenDashboard:
		return "Dashboard"
	case ScreenLeaders:
		return "Leaders"
	case ScreenUniverse:
		return "Universe"
	case ScreenSymbol:
		if m.selectedSymbol != "" {
			return fmt.Sprintf("Symbol: %s", m.selectedSymbol)
		}
		return "Symbol"
	case ScreenLogs:
		return "Logs"
	default:
		return "Unknown"
	}
}

// Screen-specific view functions that delegate to screen models.

func (m Model) viewDashboard() string {
	return m.dashboard.View()
}

func (m Model) viewLeaders() string {
	return m.leaders.View()
}

func (m Model) viewUniverse() string {
	return m.universe.View()
}

func (m Model) viewSymbol() string {
	return m.symbol.View()
}

func (m Model) viewLogs() string {
	return m.logs.View()
}
