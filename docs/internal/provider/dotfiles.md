# `internal/provider/dotfiles.go`

> Source: `internal/provider/dotfiles.go` · Package: `provider`

## Purpose

Implements the default storage provider: an age-encrypted file in the project
directory whose name comes from `AppConfig.EncFile` (default `env.enc`). It is
the easiest provider to commit and share.

## Responsibilities

- Write the encrypted payload to `FilePath`.
- Read the encrypted payload back.
- Validate that the file exists, is a regular file, and is non-empty.

## Key type

```go
type DotfilesProvider struct {
    FilePath string
}
```

`FilePath` comes from `AppConfig.EncFile` when the provider is built by the
registry. The global configuration owns the default (`env.enc`); the provider
does not hardcode a filename.

## Methods

| Method | Behavior |
| :--- | :--- |
| `Save` | `os.WriteFile(FilePath, data, 0644)` |
| `Load` | `os.ReadFile(FilePath)` |
| `Check` | `os.Stat`; error if missing, a directory, or empty |

`Check` returns a descriptive error for the directory and empty cases so
`doctor` can explain the problem instead of only reporting "not found".

## Execution flow

- `Save` truncates and replaces the file. The payload is already encrypted.
- `Load` returns the raw ciphertext; the crypto layer decrypts it.
- `Check` does not parse the ciphertext. A syntactically valid but corrupt file
  passes `Check` and fails later during decryption.

## Dependencies

- The standard library (`fmt`, `os`).
- No other project packages.

## Edge cases and notes

- The file mode is `0644`. The payload is encrypted, but use repository policy
  or stricter permissions if metadata exposure matters.
- `Load` of a missing file returns `os.ErrNotExist`, which `IsNotFound` accepts.
- `Check` treats a zero-byte file as broken rather than "not initialized".
- The provider does not create parent directories; the project directory must
  exist.
- Concurrent writers are not synchronized; the last write wins.

## Related files

- [`provider.md`](provider.md) — shared interface and not-found handling.
- [`registry.md`](registry.md) — builds this provider from `AppConfig.EncFile`.
- [`internal/crypto/secrets.md`](../crypto/secrets.md) — encrypt/decrypt layer.
