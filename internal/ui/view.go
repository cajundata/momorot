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

	// Default help text from key bindings
	helpText := ""
	for i, binding := range m.keys.ShortHelp() {
		if i > 0 {
			helpText += " | "
		}
		helpText += binding.Help().Key + ": " + binding.Help().Desc
	}

	return m.theme.Help.Render(helpText)
}

// Screen-specific view functions (stubs for now).

func (m Model) viewDashboard() string {
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render("📊 Dashboard\n\nLast Run: N/A\nCache Health: N/A\nAPI Quota: N/A")
}

func (m Model) viewLeaders() string {
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render("🏆 Top Leaders\n\n[Table will be rendered here]")
}

func (m Model) viewUniverse() string {
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render("🌐 Symbol Universe\n\n[Symbol list will be rendered here]")
}

func (m Model) viewSymbol() string {
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render("📈 Symbol Detail\n\n[Symbol details will be rendered here]")
}

func (m Model) viewLogs() string {
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render("📝 Runs & Logs\n\n[Log history will be rendered here]")
}
