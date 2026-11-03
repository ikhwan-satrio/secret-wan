package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

var (
	titleStyle   = ui.Title
	mutedStyle   = ui.Muted
	accentStyle  = ui.Accent
	cursorStyle  = ui.Accent.Bold(true)
	successStyle = ui.Success
	inputStyle   = ui.Text

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorBorder).
			Padding(1, 2)

	progressFull  = ui.Success
	progressEmpty = ui.Dim
)
