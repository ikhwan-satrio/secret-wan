# `internal/tui/style.go`

> Source: `internal/tui/style.go` · Package: `tui`

## Purpose

Centralizes the Lip Gloss styles used by the Bubble Tea models. The styles are
aliases of the shared theme in `internal/ui`, plus a few TUI-specific additions.

## Responsibilities

- Provide short, readable style names for TUI code.
- Define the bordered box used by both models.
- Define progress-bar colors.

## Key code

```go
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
```

## Style reference

| Style | Used for |
| :--- | :--- |
| `titleStyle` | Model headings |
| `mutedStyle` | Hints, secondary text, unselected choices |
| `accentStyle` | Input prompt and active key name |
| `cursorStyle` | Selected list item (`❯ ...`) |
| `successStyle` | Completion message |
| `inputStyle` | Text input content |
| `boxStyle` | Rounded border and padding around the whole screen |
| `progressFull` | Filled progress blocks |
| `progressEmpty` | Remaining progress blocks |

## Why aliases

Keeping style choices in one file means the TUI models read cleanly and a theme
change only touches `internal/ui/theme.go` plus this mapping. It also avoids
duplicating color definitions.

## Dependencies

- `github.com/charmbracelet/lipgloss`.
- `internal/ui` for colors and base styles.

## Edge cases and notes

- `cursorStyle` derives a new bold style from `ui.Accent`, so it does not mutate
  the shared style.
- `boxStyle` uses a rounded border and the shared border color, which adapts to
  light and dark terminals.
- Progress colors reuse success and dim styles rather than defining new colors.
- Styles are package-level values and are safe to reuse.

## Related files

- [`internal/ui/theme.md`](../ui/theme.md) — shared colors and styles.
- [`internal/tui/base.md`](base.md)
- [`internal/tui/select.md`](select.md)
