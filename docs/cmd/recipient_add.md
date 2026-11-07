# `cmd/recipient_add.go`

> Source: `cmd/recipient_add.go` · Package: `cmd`

## Purpose

Implements `secwan recipient add <age1...>`. It adds a public key to the project
and re-encrypts existing secrets for every configured recipient.

## Responsibilities

- Delegate validation, list changes, rekeying, and rollback to the service.
- Report an already-configured key without rewriting anything.
- Print the added key and the rekey outcome.
- Render rollback-specific errors.

## Execution flow

1. Create a session.
2. Call `AddRecipient`.
3. On error, use `failRecipientChange`.
4. Print the header.
5. If the key was already present, print `Recipient is already configured.` and
   return.
6. Print `Added <key>`.
7. Print the rekey result with `printRekeyResult`.

## Service behavior

`AddRecipient`:

- canonicalizes the argument;
- ensures the project and identity exist;
- seeds an empty list with the current key;
- allows adding the current key back when a list omits it;
- refuses adding a different key while the current key is absent;
- saves the new list, then calls `Rekey(false)`;
- restores the old list if rekeying fails.

## Edge cases and notes

- Adding a duplicate is a no-op, which makes scripts idempotent.
- The command can create the master key because it is a write command.
- If no secrets are stored yet, the rekey result explains that future saves will
  use all recipients.
- A rollback failure is reported distinctly so the user knows manual inspection
  is required.
- Output is printed after the service succeeds, so a rolled-back add is not
  reported as successful.

## Dependencies

- [`session.md`](session.md), [`shared.md`](shared.md), and
  [`recipient.md`](recipient.md).
- [`../internal/app/recipients.md`](../internal/app/recipients.md).
- `github.com/spf13/cobra`.

## Related files

- [`recipient_remove.md`](recipient_remove.md) — removal counterpart.
- [`recipient_list.md`](recipient_list.md) — verify the result.
- [`rekey.md`](rekey.md) — manual rekey after file edits.
