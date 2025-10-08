package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme and styling for the application.
type Theme struct {
	// Colors
	Primary      lipgloss.Color
	Secondary    lipgloss.Color
	Success      lipgloss.Color
	Warning      lipgloss.Color
	Error        lipgloss.Color
	Info         lipgloss.Color
	Muted        lipgloss.Color
	Background   lipgloss.Color
	Foreground   lipgloss.Color
	BorderColor  lipgloss.Color
	HighlightBg  lipgloss.Color
	HighlightFg  lipgloss.Color

	// Styles
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	TabActive     lipgloss.Style
	TabInactive   lipgloss.Style
	StatusBar     lipgloss.Style
	StatusBarInfo lipgloss.Style
	StatusBarWarn lipgloss.Style
	StatusBarErr  lipgloss.Style
	Border        lipgloss.Style
	TableHeader   lipgloss.Style
	TableRow      lipgloss.Style
	TableRowAlt   lipgloss.Style
	Highlight     lipgloss.Style
	Positive      lipgloss.Style
	Negative      lipgloss.Style
	Neutral       lipgloss.Style
	Loading       lipgloss.Style
	Help          lipgloss.Style
}

// DefaultTheme returns the default color theme.
func DefaultTheme() Theme {
	t := Theme{
		// Color palette
		Primary:      lipgloss.Color("12"),  // Bright blue
		Secondary:    lipgloss.Color("14"),  // Bright cyan
		Success:      lipgloss.Color("10"),  // Bright green
		Warning:      lipgloss.Color("11"),  // Bright yellow
		Error:        lipgloss.Color("9"),   // Bright red
		Info:         lipgloss.Color("12"),  // Bright blue
		Muted:        lipgloss.Color("8"),   // Bright black (gray)
		Background:   lipgloss.Color("0"),   // Black
		Foreground:   lipgloss.Color("15"),  // White
		BorderColor:  lipgloss.Color("8"),   // Bright black (gray)
		HighlightBg:  lipgloss.Color("8"),   // Bright black (gray)
		HighlightFg:  lipgloss.Color("12"),  // Bright blue
	}

	// Build styles
	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		MarginBottom(1)

	t.Subtitle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Italic(true)

	t.TabActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.HighlightFg).
		Background(t.HighlightBg).
		Padding(0, 2)

	t.TabInactive = lipgloss.NewStyle().
		Foreground(t.Muted).
		Padding(0, 2)

	t.StatusBar = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Padding(0, 1)

	t.StatusBarInfo = lipgloss.NewStyle().
		Foreground(t.Info)

	t.StatusBarWarn = lipgloss.NewStyle().
		Foreground(t.Warning)

	t.StatusBarErr = lipgloss.NewStyle().
		Foreground(t.Error)

	t.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderColor).
		Padding(1, 2)

	t.TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.BorderColor).
		BorderBottom(true).
		Padding(0, 1)

	t.TableRow = lipgloss.NewStyle().
		Padding(0, 1)

	t.TableRowAlt = lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	t.Highlight = lipgloss.NewStyle().
		Background(t.HighlightBg).
		Foreground(t.HighlightFg).
		Bold(true)

	t.Positive = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)

	t.Negative = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true)

	t.Neutral = lipgloss.NewStyle().
		Foreground(t.Muted)

	t.Loading = lipgloss.NewStyle().
		Foreground(t.Warning).
		Italic(true)

	t.Help = lipgloss.NewStyle().
		Foreground(t.Muted).
		Italic(true)

	return t
}

// Colorize returns a styled string based on value (positive/negative/neutral).
func (t Theme) Colorize(value float64) lipgloss.Style {
	if value > 0.001 {
		return t.Positive
	} else if value < -0.001 {
		return t.Negative
	}
	return t.Neutral
}

// FormatPercentage formats a float as a percentage with appropriate color.
func (t Theme) FormatPercentage(value float64) string {
	style := t.Colorize(value)
	sign := ""
	if value > 0 {
		sign = "+"
	}
	return style.Render(sign + lipgloss.NewStyle().Width(7).Align(lipgloss.Right).Render(formatFloat(value*100, 2)) + "%")
}

// FormatFloat formats a float with appropriate color based on value.
func (t Theme) FormatFloat(value float64, decimals int) string {
	style := t.Colorize(value)
	return style.Render(formatFloat(value, decimals))
}

// Helper function to format floats with fixed decimal places.
func formatFloat(value float64, decimals int) string {
	var result string
	switch decimals {
	case 0:
		result = lipgloss.NewStyle().Render(fmt.Sprintf("%.0f", value))
	case 1:
		result = lipgloss.NewStyle().Render(fmt.Sprintf("%.1f", value))
	case 2:
		result = lipgloss.NewStyle().Render(fmt.Sprintf("%.2f", value))
	case 3:
		result = lipgloss.NewStyle().Render(fmt.Sprintf("%.3f", value))
	default:
		result = lipgloss.NewStyle().Render(fmt.Sprintf("%f", value))
	}
	return result
}
