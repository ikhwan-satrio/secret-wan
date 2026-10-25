package cmd

import "github.com/charmbracelet/lipgloss"

var (
	IconSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓")
	IconError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("×")
	IconArrow   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("╰─▶")
	IconInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("󰘧") // Ikon info/lightbulb

	// Style Teks
	StyleAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	StyleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	StyleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("242")) // Abu-abu muted
	StyleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	
	StyleErrorBlock = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("240")).
				PaddingLeft(1)
)

func FormatError(shortMsg, detailMsg string) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, 
			lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error:"),
			"   ",
			IconError,
			" ",
			lipgloss.NewStyle().Bold(true).Render(shortMsg),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			"  ",
			IconArrow,
			" ",
			detailMsg,
		),
	)
}
