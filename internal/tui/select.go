package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SelectModel struct {
	Title   string
	Choices []string
	Cursor  int
	Chosen  bool
	Cancel  bool
	Result  string
}

func (m SelectModel) Init() tea.Cmd { return nil }

func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.Cancel = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter":
			m.Result = m.Choices[m.Cursor]
			m.Chosen = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m SelectModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.Title))
	b.WriteString("\n\n")

	for i, choice := range m.Choices {
		if i == m.Cursor {
			b.WriteString(cursorStyle.Render("❯ " + choice))
		} else {
			b.WriteString("  " + mutedStyle.Render(choice))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("↑/↓ navigate  •  enter select  •  esc cancel"))

	return boxStyle.Render(b.String()) + "\n"
}
