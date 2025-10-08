package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyBindings defines all key bindings for the application.
type KeyBindings struct {
	Quit     key.Binding
	Refresh  key.Binding
	Search   key.Binding
	Export   key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding
	Enter    key.Binding
	Back     key.Binding
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
}

// DefaultKeyBindings returns the default key bindings.
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("right", "tab"),
			key.WithHelp("→/tab", "next screen"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("left", "shift+tab"),
			key.WithHelp("←/shift+tab", "prev screen"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "bottom"),
		),
	}
}

// ShortHelp returns a short help text for the key bindings.
func (k KeyBindings) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Refresh,
		k.NextTab,
		k.PrevTab,
		k.Search,
		k.Export,
		k.Quit,
	}
}

// FullHelp returns the full help text for the key bindings.
func (k KeyBindings) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Home, k.End, k.Enter, k.Back},
		{k.NextTab, k.PrevTab, k.Refresh, k.Search},
		{k.Export, k.Quit},
	}
}
