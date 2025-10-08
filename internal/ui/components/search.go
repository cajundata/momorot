package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchModel wraps the Bubbles textinput component for search functionality.
type SearchModel struct {
	input textinput.Model
	theme SearchTheme
}

// SearchTheme contains styling for the search component.
type SearchTheme struct {
	FocusedStyle   lipgloss.Style
	BlurredStyle   lipgloss.Style
	CursorStyle    lipgloss.Style
	PlaceholderStyle lipgloss.Style
	PromptStyle    lipgloss.Style
}

// NewSearch creates a new search input.
func NewSearch(placeholder string) SearchModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 50
	ti.Width = 30

	theme := defaultSearchTheme()

	// Apply theme
	ti.PromptStyle = theme.PromptStyle
	ti.TextStyle = theme.FocusedStyle
	ti.PlaceholderStyle = theme.PlaceholderStyle
	ti.Cursor.Style = theme.CursorStyle

	return SearchModel{
		input: ti,
		theme: theme,
	}
}

// defaultSearchTheme returns the default search theme.
func defaultSearchTheme() SearchTheme {
	return SearchTheme{
		FocusedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),
		BlurredStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		CursorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")),
		PlaceholderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
		PromptStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true),
	}
}

// Init initializes the search input.
func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles search input updates.
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the search input.
func (m SearchModel) View() string {
	return m.input.View()
}

// Focus sets focus on the input.
func (m *SearchModel) Focus() tea.Cmd {
	return m.input.Focus()
}

// Blur removes focus from the input.
func (m *SearchModel) Blur() {
	m.input.Blur()
}

// Focused returns whether the input is focused.
func (m SearchModel) Focused() bool {
	return m.input.Focused()
}

// Value returns the current input value.
func (m SearchModel) Value() string {
	return m.input.Value()
}

// SetValue sets the input value.
func (m *SearchModel) SetValue(s string) {
	m.input.SetValue(s)
}

// Reset clears the input value.
func (m *SearchModel) Reset() {
	m.input.SetValue("")
}

// SetWidth sets the input width.
func (m *SearchModel) SetWidth(width int) {
	m.input.Width = width
}

// SetPlaceholder sets the placeholder text.
func (m *SearchModel) SetPlaceholder(placeholder string) {
	m.input.Placeholder = placeholder
}
