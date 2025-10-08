// Package ui provides the terminal user interface for the momentum screener.
package ui

// Import dependencies to ensure they're included in go.mod
import (
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/bubbles/table"
	_ "github.com/charmbracelet/bubbles/spinner"
	_ "github.com/charmbracelet/bubbles/textinput"
	_ "github.com/charmbracelet/lipgloss"
)
