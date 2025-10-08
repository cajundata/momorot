package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableModel wraps the Bubbles table component with custom styling.
type TableModel struct {
	table  table.Model
	theme  TableTheme
	width  int
	height int
}

// TableTheme contains styling for the table component.
type TableTheme struct {
	HeaderStyle  lipgloss.Style
	CellStyle    lipgloss.Style
	SelectedStyle lipgloss.Style
	BorderStyle   lipgloss.Style
}

// NewTable creates a new table with the given columns and rows.
func NewTable(columns []table.Column, rows []table.Row, width, height int) TableModel {
	theme := defaultTableTheme()

	// Create the underlying Bubbles table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	// Apply custom styles
	s := table.DefaultStyles()
	s.Header = theme.HeaderStyle
	s.Cell = theme.CellStyle
	s.Selected = theme.SelectedStyle
	t.SetStyles(s)

	return TableModel{
		table:  t,
		theme:  theme,
		width:  width,
		height: height,
	}
}

// defaultTableTheme returns the default table theme.
func defaultTableTheme() TableTheme {
	return TableTheme{
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			BorderBottom(true).
			Padding(0, 1),
		CellStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1),
		SelectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Background(lipgloss.Color("8")).
			Padding(0, 1),
		BorderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")),
	}
}

// Init initializes the table.
func (m TableModel) Init() tea.Cmd {
	return nil
}

// Update handles table updates.
func (m TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table.
func (m TableModel) View() string {
	return m.theme.BorderStyle.Render(m.table.View())
}

// SetRows updates the table rows.
func (m *TableModel) SetRows(rows []table.Row) {
	m.table.SetRows(rows)
}

// SetColumns updates the table columns.
func (m *TableModel) SetColumns(columns []table.Column) {
	m.table.SetColumns(columns)
}

// SelectedRow returns the currently selected row.
func (m TableModel) SelectedRow() table.Row {
	return m.table.SelectedRow()
}

// Cursor returns the current cursor position.
func (m TableModel) Cursor() int {
	return m.table.Cursor()
}

// SetCursor sets the cursor position.
func (m *TableModel) SetCursor(n int) {
	m.table.SetCursor(n)
}

// SetWidth sets the table width.
func (m *TableModel) SetWidth(width int) {
	m.width = width
	m.table.SetWidth(width)
}

// SetHeight sets the table height.
func (m *TableModel) SetHeight(height int) {
	m.height = height
	m.table.SetHeight(height)
}

// Focus sets the table focus state.
func (m *TableModel) Focus() {
	m.table.Focus()
}

// Blur removes focus from the table.
func (m *TableModel) Blur() {
	m.table.Blur()
}

// Focused returns whether the table is focused.
func (m TableModel) Focused() bool {
	return m.table.Focused()
}
