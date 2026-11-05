# `internal/ui/theme.go`

> Source: `internal/ui/theme.go` · Package: `ui`

## Purpose

Defines the shared terminal theme: adaptive colors for light and dark
backgrounds, reusable text styles, pre-rendered icons, and formatting helpers
for dividers, steps, and errors.

## Responsibilities

- Provide a color palette that adapts to the terminal background.
- Export styles used by commands and TUI models.
- Provide pre-rendered status icons.
- Render horizontal dividers.
- Format numbered next-step lines.
- Format compact multi-line error blocks.

## Color palette

Each color is a `lipgloss.AdaptiveColor` with a light and dark variant:

| Variable | Light | Dark | Used for |
| :--- | :--- | :--- | :--- |
| `ColorPrimary` | `#2563EB` | `#60A5FA` | Titles |
| `ColorAccent` | `#7C3AED` | `#C084FC` | Highlights, prompts |
| `ColorSuccess` | `#15803D` | `#4ADE80` | Success values |
| `ColorError` | `#DC2626` | `#F87171` | Errors |
| `ColorWarning` | `#D97706` | `#FBBF24` | Warnings |
| `ColorMuted` | `#64748B` | `#94A3B8` | Secondary text |
| `ColorBorder` | `#CBD5E1` | `#334155` | Box borders |
| `ColorText` | `#0F172A` | `#E2E8F0` | Primary text |

Lip Gloss chooses the variant based on the detected terminal background.

## Styles

| Style | Definition |
| :--- | :--- |
| `Title` | Bold, primary color |
| `Accent` | Accent color |
| `Success` | Success color |
| `Error` | Error color, bold |
| `Warning` | Warning color |
| `Muted` / `Dim` | Muted color |
| `Text` | Primary text color |
| `Bold` | Bold only |
| `Key` | Text color, bold |

Commands use these to keep output consistent across every subcommand.

## Icons

```go
var (
    IconSuccess = Success.Render("✓")
    IconError   = Error.Render("✗")
    IconInfo    = Accent.Render("ℹ")
    IconWarning = Warning.Render("⚠")
    IconArrow   = Muted.Render("→")
)
```

The icons are rendered once at package initialization, so they already contain
ANSI escape codes when interpolated into output. This is why code writes
`fmt.Printf("%s ...", ui.IconSuccess)`.

## Helper functions

### `Divider(width int) string`

- Uses `44` when `width <= 0`.
- Returns a muted string of `─` characters repeated `width` times.

### `Step(number int, command, description string) string`

Formats:

```text
  <n>. <command> <description>
```

using muted numbers, accent commands, and muted descriptions. Used by `init`.

### `FormatError(shortMsg, detailMsg string) string`

Builds a vertical stack:

1. `✗ <short message>` in error style.
2. An optional muted `→ <detail>` line.

The block has a left border in the error color and one column of padding. This
is the format used by `cmd.fail` and direct error paths.

## Dependencies

- `github.com/charmbracelet/lipgloss`.
- Standard library `fmt`, `strings`.

## Edge cases and notes

- Colors may be disabled or stripped when output is not a TTY, depending on Lip
  Gloss's renderer behavior.
- Icons contain ANSI sequences; comparing rendered strings for equality should
  account for that.
- `Dim` and `Muted` are currently identical styles.
- `FormatError` returns a block without a trailing newline; callers add one.
- The theme does not define success icons for all platforms; it assumes Unicode
  support.

## Related files

- [`internal/tui/style.md`](../tui/style.md) — TUI-specific aliases.
- [`cmd/shared.md`](../../cmd/shared.md) — `fail` uses `FormatError`.
- [`cmd/doctor.md`](../../cmd/doctor.md) — heavy user of icons and styles.
