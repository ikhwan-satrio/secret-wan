# `cmd/recipient.go`

> Source: `cmd/recipient.go` · Package: `cmd`

## Purpose

Defines the `recipient` command family and the shared output/error helpers used
by its subcommands. The actual commands live in focused files:

| File | Command |
| :--- | :--- |
| `cmd/recipient_list.go` | `recipient list` / `ls` |
| `cmd/recipient_me.go` | `recipient me` / `whoami` |
| `cmd/recipient_add.go` | `recipient add <age1...>` |
| `cmd/recipient_remove.go` | `recipient remove <age1...>` / `rm` |

## Responsibilities

- Register the parent command and its subcommands.
- Register the `--force` flag on `remove`.
- Render the shared rekey outcome.
- Render recipient-change failures, including rollback states.

## Key functions

### `printRekeyResult(result *app.RekeyResult)`

When `Secrets == 0`, prints:

```text
No stored secrets to re-encrypt yet.
Future saves will be encrypted to all configured recipients.
```

Otherwise prints:

```text
✓ Re-encrypted 3 secret(s) for 2 recipient(s).
```

### `failRecipientChange(err)`

Handles errors from `app.AddRecipient` and `app.RemoveRecipient`:

- `*app.RecipientChangeError` with `Restored=true` → heading
  `Failed to re-encrypt secrets; recipient list was restored`;
- `*app.RecipientChangeError` with `Restored=false` → heading
  `Failed to re-encrypt secrets and could not restore recipients`, including the
  rollback error when present;
- any other error → `failApp`.

## Registration

`recipientCmd` uses aliases `recipients`. The `init` function wires the four
subcommands and the `--force` flag, then attaches the family to `rootCmd`.

## Edge cases and notes

- The commands call the application service; they do not manipulate the
  recipient file directly.
- Output is printed only after the service has completed the change and rekey,
  so a rolled-back operation does not claim success.
- `printRekeyResult` is also reused by the standalone `rekey` command.

## Dependencies

- `internal/app` for `RekeyResult` and `RecipientChangeError`.
- `internal/ui` for icons and styles.
- [`session.md`](session.md) and [`shared.md`](shared.md) for CLI helpers.

## Related files

- [`recipient_list.md`](recipient_list.md)
- [`recipient_me.md`](recipient_me.md)
- [`recipient_add.md`](recipient_add.md)
- [`recipient_remove.md`](recipient_remove.md)
- [`rekey.md`](rekey.md)
- [`../internal/app/recipients.md`](../internal/app/recipients.md)
