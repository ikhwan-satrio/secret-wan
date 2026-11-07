# `cmd/recipient_list.go`

> Source: `cmd/recipient_list.go` · Package: `cmd`

## Purpose

Implements `secwan recipient list` (alias `ls`), which displays the configured
public age recipients and marks the current user's key.

## Responsibilities

- Read the recipient list through the service.
- Attempt to load the current identity without making a missing key fatal.
- Show the file path.
- Show single-recipient fallback mode when the file is empty.
- Mark the current key with `(you)`.
- Warn when a non-empty list excludes the current key.

## Execution flow

1. Create a session and call `ListRecipients`.
2. Load errors fail the command.
3. Print the header and recipient file path.
4. If there are no keys:
   - print the single-recipient explanation;
   - print the implicit current key when known, or a hint to run `init`.
5. Otherwise print each key, marking the current one.
6. Print the configured count.
7. Warn to stderr if the current key is missing from a non-empty list.

## Key code

`ListRecipients` returns a `RecipientList` with `Path`, `Keys`, `Self`, and
`SelfKnown`. `SelfListed()` performs the membership check, so the command does
not need to import the recipient package.

## Edge cases and notes

- A missing master key is not fatal; the command can still display public keys.
- A missing recipient file is healthy single-recipient mode, not an error.
- The warning is informational; the command does not modify the file.
- Stdout contains only the listing; warnings go to stderr.

## Dependencies

- [`session.md`](session.md) and [`shared.md`](shared.md).
- [`../internal/app/recipients.md`](../internal/app/recipients.md).
- `github.com/spf13/cobra`.

## Related files

- [`recipient.md`](recipient.md) — family overview.
- [`recipient_me.md`](recipient_me.md) — prints only the public key.
- [`recipient_add.md`](recipient_add.md) — repairs a missing self key.
