# `cmd/recipient_remove.go`

> Source: `cmd/recipient_remove.go` · Package: `cmd`

## Purpose

Implements `secwan recipient remove <age1...>` (alias `rm`). It removes a public
key and re-encrypts existing secrets for the remaining recipients.

## Responsibilities

- Read `--force`.
- Delegate validation, list changes, rekeying, and rollback to the service.
- Print the removed key and rekey outcome.
- Warn when the current key itself was removed.
- Render rollback-specific errors.

## Execution flow

1. Read `--force`.
2. Create a session.
3. Call `RemoveRecipient`.
4. On error, use `failRecipientChange`.
5. Print the header and `Removed <key>`.
6. Print the rekey result.
7. Warn when `SelfRemoved` is true.

## Safety rules enforced by the service

| Rule | Behavior |
| :--- | :--- |
| Target must exist | `Recipient not found` |
| Last recipient cannot be removed | `Cannot remove the last recipient` |
| Self-exclusion needs `--force` | `Refusing to remove your own key` |
| Already-excluded self needs `--force` | `Current key is not a recipient` |
| Rekey failure restores the list | `RecipientChangeError` rollback |

The final recipient rule applies even when `--force` is used.

## Edge cases and notes

- Removing the current key is allowed with `--force`, but the command warns that
  new writes cannot be decrypted with it.
- The removed key still decrypts old Git history or old ciphertext copies; a
  rekey only protects future reads of the current payload.
- Removing from an empty file seeds it with the current key first, so the target
  lookup is consistent.
- An existing key is required; `remove` does not create one.
- Output is printed after a successful rekey, so a rolled-back removal is not
  reported as successful.

## Dependencies

- [`session.md`](session.md), [`shared.md`](shared.md), and
  [`recipient.md`](recipient.md).
- [`../internal/app/recipients.md`](../internal/app/recipients.md).
- `github.com/spf13/cobra`.

## Related files

- [`recipient_add.md`](recipient_add.md) — addition counterpart.
- [`recipient_list.md`](recipient_list.md) — verify the result.
- [`rekey.md`](rekey.md) — manual rekey after file edits.
