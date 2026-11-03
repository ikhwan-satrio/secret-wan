package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette adapts to both light and dark terminal themes.
var (
	ColorPrimary = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}
	ColorAccent  = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#C084FC"}
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#4ADE80"}
	ColorError   = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	ColorWarning = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}
	ColorMuted   = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}
	ColorBorder  = lipgloss.AdaptiveColor{Light: "#CBD5E1", Dark: "#334155"}
	ColorText    = lipgloss.AdaptiveColor{Light: "#0F172A", Dark: "#E2E8F0"}
)

// Shared text styles.
var (
	Title   = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	Accent  = lipgloss.NewStyle().Foreground(ColorAccent)
	Success = lipgloss.NewStyle().Foreground(ColorSuccess)
	Error   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(ColorWarning)
	Muted   = lipgloss.NewStyle().Foreground(ColorMuted)
	Dim     = lipgloss.NewStyle().Foreground(ColorMuted)
	Text    = lipgloss.NewStyle().Foreground(ColorText)
	Bold    = lipgloss.NewStyle().Bold(true)
	Key     = lipgloss.NewStyle().Foreground(ColorText).Bold(true)
)

// Icons are pre-rendered so they can be interpolated directly into output.
var (
	IconSuccess = Success.Render("✓")
	IconError   = Error.Render("✗")
	IconInfo    = Accent.Render("ℹ")
	IconWarning = Warning.Render("⚠")
	IconArrow   = Muted.Render("→")
)

// Divider returns a horizontal rule suitable for terminal output.
func Divider(width int) string {
	if width <= 0 {
		width = 44
	}

	return Muted.Render(strings.Repeat("─", width))
}

// Step formats a numbered next-step line.
func Step(number int, command, description string) string {
	return fmt.Sprintf(
		"  %s %s %s",
		Muted.Render(fmt.Sprintf("%d.", number)),
		Accent.Render(command),
		Muted.Render(description),
	)
}

// FormatError renders a compact error block with a colored left border.
func FormatError(shortMsg, detailMsg string) string {
	lines := []string{IconError + " " + Error.Render(shortMsg)}

	if detailMsg != "" {
		lines = append(lines, Muted.Render("  "+"→ "+detailMsg))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ColorError).
		PaddingLeft(1).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
