# `cmd/rekey.go`

> Source: `cmd/rekey.go` · Package: `cmd`

## Purpose

Implements `secwan rekey`, which decrypts the current secret map and re-encrypts
it for every recipient listed in `secret-wan.recipients`. It is the manual
counterpart to the automatic rekey performed by `recipient add` and
`recipient remove`.

Use it after editing the recipient file by hand or after changing the list on
another machine.

## Responsibilities

- Require the manifest and an existing master key.
- Delegate recipient resolution, rekey, and missing-store handling to
  `app.Service.Rekey`.
- Fall back to the current key when no recipient file exists.
- Refuse to lock the current key out unless `--force` is used.
- Treat missing storage as "nothing to rekey" rather than an error.
- Report the secret and recipient counts.
- Warn when the current key was intentionally excluded.

## Execution flow

```text
rekey
 ├── read --force
 ├── session.requireProject
 ├── session.requireExistingIdentity
 ├── app.Service.Rekey(force)
 │    ├── load recipients
 │    │    ├── empty → use current key, mark UsedImplicit
 │    │    └── present → use list
 │    ├── current key missing and no --force → UserError
 │    ├── load and decrypt secrets
 │    │    ├── not found → Missing=true
 │    │    └── error     → fail
 │    ├── parse recipients into age recipients
 │    └── save the same secret map encrypted to every recipient
 ├── print fallback notice when UsedImplicit
 ├── print "No stored secrets to re-encrypt." when Missing
 └── print counts; warn if the current key is excluded
```

## Key code

### Recipient resolution

The command calls:

```go
result, err := session.Rekey(force)
if err != nil {
    failApp("Failed to re-encrypt secrets", err)
}
```

The service in [`../internal/app/recipients.md`](../internal/app/recipients.md):

- loads `secret-wan.recipients`;
- falls back to the current key when the file is empty;
- enforces self-inclusion unless `force` is true;
- returns a `RekeyResult` with counts and status flags.

The empty-list fallback is the compatibility path for projects created before
multi-recipient support.

### Why a separate command exists

`recipient add` and `recipient remove` already rekey. A standalone command is
still useful when:

- the recipient file was edited directly;
- a recipient was added on another machine and the ciphertext was not updated;
- the user wants to confirm that all keys can still decrypt the payload;
- a key rotation workflow wants an explicit, auditable rekey step.

### No-op without storage

A missing storage is not a failure. A fresh project may have a manifest and a
recipient list but no secrets yet. `rekey` reports that there is nothing to do
and exits successfully.

## Flags

| Flag | Effect |
| :--- | :--- |
| `--force` | Re-encrypt even when the current key is not in the recipient list |

`--force` is intentionally required for self-exclusion. Without it, `rekey`
fails with a hint to run `secwan recipient add <current-key>` first.

## Dependencies

- `internal/ui` for output.
- Shared helpers: `newSession`, `requireProject`, `requireExistingIdentity`,
  `Rekey`, `failApp`, `warn`, `printHeader`.
- [`../internal/app/recipients.md`](../internal/app/recipients.md) — recipient
  policy and rekey.
- `github.com/spf13/cobra`.

## Edge cases and notes

- `rekey` does not modify the recipient file; it only reads it.
- The encrypted payload is fully replaced, so a successful rekey is the point at
  which removed keys lose access to the stored secrets.
- If decryption fails, rekey stops before writing anything. This matters when
  the current key is not actually able to read the existing payload.
- Empty secret maps are rewritten as `{}` rather than treated as missing.
- The command prints a stderr warning when the current key is not a recipient,
  even when `--force` was supplied.
- Recipient parsing re-validates the file, so a malformed line fails the command
  instead of producing a partially encrypted result.

## Related files

- [`recipient.md`](recipient.md) — add/remove commands that rekey automatically.
- [`session.md`](session.md) — CLI service wrapper.
- [`../internal/app/recipients.md`](../internal/app/recipients.md) — rekey
  policy and results.
- [`../internal/recipients/recipients.md`](../internal/recipients/recipients.md)
  — recipient storage.
- [`../internal/crypto/secrets.md`](../internal/crypto/secrets.md) — variadic
  encryption.
