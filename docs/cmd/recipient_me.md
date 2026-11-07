# `cmd/recipient_me.go`

> Source: `cmd/recipient_me.go` · Package: `cmd`

## Purpose

Implements `secwan recipient me` (alias `whoami`), which prints the current
user's public age recipient. This is the key a teammate adds to their project.

## Responsibilities

- Require an existing master key.
- Print only the public `age1...` key to stdout.
- Warn to stderr when a non-empty recipient list excludes the current key.
- Warn to stderr, without failing, when the recipient file cannot be read.

## Execution flow

1. Create a session and require an existing identity.
2. Print `identity.Recipient().String()`.
3. Load the recipient list through `ListRecipients`.
4. On read error, print a warning and return successfully.
5. If the list is non-empty and does not include the current key, warn.

## Key code

```go
self := session.Identity.Recipient().String()
fmt.Println(self)
```

Printing before the recipient check keeps the command usable in shell
substitution even when the warning path is taken:

```bash
TEAMMATE_KEY="$(secwan recipient me)"
```

## Edge cases and notes

- A missing master key fails with a hint to run `init`; the command never creates
  a key.
- Warnings go to stderr, so command substitution captures only the key.
- A missing recipient file is normal and produces no warning.
- The key is printed in canonical X25519 form.

## Dependencies

- [`session.md`](session.md) and [`shared.md`](shared.md).
- [`../internal/app/recipients.md`](../internal/app/recipients.md).
- `github.com/spf13/cobra`.

## Related files

- [`recipient.md`](recipient.md) — family overview.
- [`recipient_list.md`](recipient_list.md) — full configured list.
- [`recipient_add.md`](recipient_add.md) — adds this key elsewhere.
