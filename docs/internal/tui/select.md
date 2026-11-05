# `internal/tui/select.go`

> Source: `internal/tui/select.go` · Package: `tui`

## Purpose

Implements a small Bubble Tea model for choosing one item from a list. It is
used by `secwan init` to choose the storage provider.

## Responsibilities

- Track the cursor and the list of choices.
- Support vim-style `j`/`k` and arrow-key navigation.
- Confirm with Enter and cancel with Esc or Ctrl+C.
- Render a titled, bordered list with a hint line.

## Key type

### `SelectModel`

| Field | Meaning |
| :--- | :--- |
| `Title` | Heading shown above the list |
| `Choices` | Human-readable options |
| `Cursor` | Selected index |
| `Chosen` | True when the user pressed Enter |
| `Cancel` | True when the user cancelled |
| `Result` | The selected choice string |

## Key functions

### `Init`

Returns `nil`; no initial command is needed.

### `Update`

| Key | Effect |
| :--- | :--- |
| `ctrl+c`, `esc` | `Cancel = true`, quit |
| `up`, `k` | Move cursor up when possible |
| `down`, `j` | Move cursor down when possible |
| `enter` | `Result = Choices[Cursor]`, `Chosen = true`, quit |

The cursor bounds are checked against `len(m.Choices)-1`, so an empty choice
list does not panic on Enter because no Enter handling can index a valid item
unless the list is non-empty. In practice, `init` always supplies two choices.

### `View`

1. Renders the title with the title style.
2. Renders each choice, highlighting the cursor with `❯ `.
3. Adds the navigation hint.
4. Wraps everything in the shared rounded box style.

## Dependencies

- `github.com/charmbracelet/bubbletea`.
- Styles from [`internal/tui/style.md`](style.md).

## Edge cases and notes

- Cancellation does not set `Result`, so callers should check `Cancel` before
  using `Result`.
- `Chosen` and `Cancel` are separate flags; normally only one is true.
- The model does not validate the selected string. The caller maps labels to
  internal provider names.
- The choice list is rendered in the provided order; it is not sorted.

## Related files

- [`cmd/init.md`](../../cmd/init.md) — consumes `SelectModel`.
- [`internal/tui/base.md`](base.md) — multi-key input model.
- [`internal/tui/style.md`](style.md)
