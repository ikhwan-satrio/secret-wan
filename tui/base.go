package tui;

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	StateSelect = iota
	StateInput
)

// Style Minimalis
var (
	styleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // Biru muda
	styleCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Pink untuk kursor
	styleInput  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
)

type Model struct {
	State      int
	Keys       []string
	TotalKeys  int // Untuk menampilkan progres (misal: 1/3)
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
	ti.Prompt = styleAccent.Render("❯ ")
	ti.TextStyle = styleInput

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
			} else if m.State == StateInput {
				m.Choices[m.CurrentKey] = m.Input.Value()
				
				// Hapus dari daftar
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
			}
		case "esc":
			if m.State == StateInput {
				m.State = StateSelect
				m.Input.Blur()
			}
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
			return styleDim.Render("\n  Cancelled. No secrets were saved.\n")
		}
		return styleAccent.Render("\n  ✓ Secrets configured successfully.\n")
	}

	var b strings.Builder
	
	// Header Minimalis dengan Progres
	remaining := len(m.Keys)
	completed := m.TotalKeys - remaining
	b.WriteString(styleDim.Render(fmt.Sprintf("  Configure secrets (%d/%d)\n", completed, m.TotalKeys)))
	b.WriteString(styleDim.Render("  ────────────────────────────────────────\n"))

	if m.State == StateSelect {
		for i, choice := range m.Keys {
			if m.Cursor == i {
				// Item yang dipilih
				b.WriteString(fmt.Sprintf("  %s %s\n", 
					styleCursor.Render("❯"), 
					styleAccent.Render(choice)))
			} else {
				// Item biasa
				b.WriteString(fmt.Sprintf("    %s\n", styleDim.Render(choice)))
			}
		}
		b.WriteString("\n" + styleDim.Render("  ↑/↓ to navigate • Enter to select • ctrl+c to quit"))
		
	} else if m.State == StateInput {
		b.WriteString(fmt.Sprintf("\n  %s\n", styleAccent.Render(m.CurrentKey)))
		b.WriteString(fmt.Sprintf("  %s\n", m.Input.View()))
		b.WriteString("\n" + styleDim.Render("  Enter to save • Esc to cancel • ctrl+c to quit"))
	}

	return "\n" + b.String() + "\n"
}
