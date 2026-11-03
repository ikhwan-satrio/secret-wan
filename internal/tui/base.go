package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

const (
	StateSelect = iota
	StateInput
)

type Model struct {
	State      int
	Keys       []string
	TotalKeys  int
	Cursor     int
	Choices    map[string]string
	Input      textinput.Model
	CurrentKey string
	Quitting   bool
	Cancelled  bool
}

func NewModel(keys []string) Model {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 40
	ti.Prompt = accentStyle.Render("❯ ")
	ti.TextStyle = inputStyle

	return Model{
		State:     StateSelect,
		Keys:      keys,
		TotalKeys: len(keys),
		Choices:   make(map[string]string),
		Input:     ti,
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.Cancelled = true
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.State == StateSelect && m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.State == StateSelect && m.Cursor < len(m.Keys)-1 {
				m.Cursor++
			}
		case "enter":
			if m.State == StateSelect {
				m.State = StateInput
				m.CurrentKey = m.Keys[m.Cursor]
				m.Input.SetValue("")
				m.Input.Placeholder = fmt.Sprintf("Enter value for %s", m.CurrentKey)
				m.Input.EchoMode = textinput.EchoPassword
				m.Input.Focus()
				return m, textinput.Blink
			}

			m.Choices[m.CurrentKey] = m.Input.Value()
			m.Keys = append(m.Keys[:m.Cursor], m.Keys[m.Cursor+1:]...)

			if m.Cursor >= len(m.Keys) && len(m.Keys) > 0 {
				m.Cursor = len(m.Keys) - 1
			}

			if len(m.Keys) == 0 {
				m.Quitting = true
				return m, tea.Quit
			}

			m.State = StateSelect
			m.Input.Blur()
		case "esc":
			if m.State == StateInput {
				m.State = StateSelect
				m.Input.Blur()
				return m, nil
			}

			m.Cancelled = true
			m.Quitting = true
			return m, tea.Quit
		}
	}

	if m.State == StateInput {
		var cmd tea.Cmd
		m.Input, cmd = m.Input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.Quitting {
		if m.Cancelled {
			return boxStyle.Render(mutedStyle.Render("Cancelled. No secrets were saved.")) + "\n"
		}

		return boxStyle.Render(
			ui.IconSuccess+" "+successStyle.Render("Secrets configured successfully."),
		) + "\n"
	}

	remaining := len(m.Keys)
	completed := m.TotalKeys - remaining
	if completed < 0 {
		completed = 0
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Configure secrets"))
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  %d/%d", completed, m.TotalKeys)))
	b.WriteString("\n")
	b.WriteString(progressBar(completed, m.TotalKeys, 30))
	b.WriteString("\n\n")

	switch m.State {
	case StateSelect:
		for i, choice := range m.Keys {
			if m.Cursor == i {
				b.WriteString(cursorStyle.Render("❯ " + choice))
			} else {
				b.WriteString("  " + mutedStyle.Render(choice))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("↑/↓ navigate  •  enter select  •  esc cancel"))
	case StateInput:
		b.WriteString(accentStyle.Render(m.CurrentKey))
		b.WriteString("\n")
		b.WriteString(m.Input.View())
		b.WriteString("\n\n")
		b.WriteString(mutedStyle.Render("enter save  •  esc back  •  ctrl+c quit"))
	}

	return "\n" + boxStyle.Render(b.String()) + "\n"
}

func progressBar(completed, total, width int) string {
	if total <= 0 {
		return ""
	}

	if completed < 0 {
		completed = 0
	}
	if completed > total {
		completed = total
	}

	filled := completed * width / total

	return progressFull.Render(strings.Repeat("█", filled)) +
		progressEmpty.Render(strings.Repeat("░", width-filled))
}
