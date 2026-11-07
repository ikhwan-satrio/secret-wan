# `internal/tui/base.go`

> Source: `internal/tui/base.go` · Package: `tui`

## Purpose

Implements the Bubble Tea model used by `secwan sync` to prompt for required
secrets one key at a time, with a progress bar and cancel/back navigation.

## Responsibilities

- Maintain the list of keys still needing input.
- Move a cursor through the remaining keys.
- Switch between selection and input states.
- Collect entered values into a map.
- Handle Ctrl+C and Esc cancellation.
- Render progress, the key list, the input field, and status messages.

## Key types and constants

```go
const (
    StateSelect = iota
    StateInput
)
```

The model starts in `StateSelect`. Pressing Enter on a key moves to `StateInput`.

### `Model`

| Field | Meaning |
| :--- | :--- |
| `State` | Current screen (`StateSelect` or `StateInput`) |
| `Keys` | Keys still waiting for input |
| `TotalKeys` | Original number of keys, used for progress |
| `Cursor` | Selected index in `Keys` |
| `Choices` | Collected key → value map |
| `Input` | `textinput.Model` for the current value |
| `CurrentKey` | Key currently being edited |
| `Quitting` | Whether the program should terminate |
| `Cancelled` | Whether the user cancelled instead of completing |

## Key functions

### `NewModel(keys []string) Model`

Creates a text input with:

- `CharLimit = 256`
- `Width = 40`
- a styled `❯ ` prompt
- the shared input text style

It initializes `Choices` and `TotalKeys`, and starts in `StateSelect`.

### `Init`

Returns `textinput.Blink`, enabling the cursor blink.

### `Update`

Handles key messages:

| Key | State | Effect |
| :--- | :--- | :--- |
| `ctrl+c` | any | Mark cancelled, quit |
| `up` / `k` | select | Move cursor up |
| `down` / `j` | select | Move cursor down |
| `enter` | select | Enter input for the selected key |
| `enter` | input | Save value, remove key, continue or quit |
| `esc` | input | Return to selection without saving |
| `esc` | select | Cancel and quit |

After the key switch, input-state messages are forwarded to the text input model
so editing, paste, and cursor movement work normally.

When a value is submitted:

1. `Choices[CurrentKey] = Input.Value()`.
2. The key is removed from `Keys`.
3. The cursor is clamped if it is now past the end.
4. If no keys remain, the program quits.
5. Otherwise it returns to `StateSelect` and blurs the input.

### `View`

While quitting:

- `Cancelled` → a box saying no secrets were saved.
- completed → a success box.

While active:

1. Title and `completed/total`.
2. A 30-column progress bar.
3. Either the key list (select) or the key name and input (input).
4. A context-sensitive key hint.

### `progressBar(completed, total, width int)`

- Returns an empty string when `total <= 0`.
- Clamps `completed` to `[0, total]`.
- Computes `filled := completed * width / total`.
- Renders filled blocks with the success style and empty blocks with the dim
  style.

## Dependencies

- `github.com/charmbracelet/bubbles/textinput`.
- `github.com/charmbracelet/bubbletea`.
- `internal/ui` for the success icon.
- Styles from [`internal/tui/style.md`](style.md).

## Edge cases and notes

- Submitting an empty value is allowed by the model. `sync` saves it, and
  `check`/`doctor` treat it as missing.
- Esc in input mode discards the current value and returns to the list.
- Ctrl+C always cancels, even while typing.
- `Keys` is mutated by reslicing; the caller's original slice is not preserved,
  but `sync` does not need it afterward.
- The input is a password field (`EchoPassword`), so values are hidden.
- Progress is based on the number of keys, not on the number of non-empty values.
- If `NewModel` is called with zero keys, the view is not normally shown;
  `sync` only launches the TUI when `Profile.PromptKeys()` is non-empty.

## Related files

- [`cmd/sync.md`](../../cmd/sync.md) — launches this model.
- [`internal/tui/select.md`](select.md) — provider selector.
- [`internal/tui/style.md`](style.md) — styles used here.
