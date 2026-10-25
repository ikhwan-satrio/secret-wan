package tui

import "github.com/charmbracelet/lipgloss"

var (
	StyleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	StyleAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	StyleCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	StyleError  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)
