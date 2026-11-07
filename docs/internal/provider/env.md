# `internal/provider/env.go`

> Source: `internal/provider/env.go` · Package: `provider`

## Purpose

Implements a read-only provider that reads the encrypted payload from an
environment variable. It is intended for CI pipelines and containers where the
payload is injected by the platform and cannot be written to disk.

## Responsibilities

- Read `SECWAN_ENV_ENC` (or the configured variable name).
- Decode base64 into raw ciphertext bytes.
- Accept standard, raw, URL-safe, and raw URL-safe base64 alphabets.
- Refuse to save, with an actionable message.
- Validate presence and decodability.

## Key type and constant

```go
const DefaultEnvVar = "SECWAN_ENV_ENC"

type EnvProvider struct {
    VarName string
}
```

An empty `VarName` falls back to `DefaultEnvVar`. The registry factory sets the
default value.

## Methods

| Method | Behavior |
| :--- | :--- |
| `Save` | Always fails with a message pointing at `SECWAN_ENV_ENC` |
| `Load` | Reads and trims the variable, decodes base64, returns bytes |
| `Check` | Calls `Load` and reports the same error |

## Base64 handling

Environment variables cannot safely carry raw binary, so the payload is stored
as base64. `decodeBase64` tries each encoding in order:

1. `base64.StdEncoding`
2. `base64.RawStdEncoding`
3. `base64.URLEncoding`
4. `base64.RawURLEncoding`

CI systems often add or strip padding, so accepting all four forms avoids a class
of avoidable setup failures. The decoded bytes are the original age ciphertext.

## Error behavior

| Condition | Result |
| :--- | :--- |
| Variable unset or whitespace-only | wrapped `ErrNotFound` (`IsNotFound` is true) |
| Variable set but not valid base64 | decode error (not `ErrNotFound`) |
| `Save` called | unsupported error |

## Dependencies

- The standard library (`encoding/base64`, `fmt`, `os`, `strings`).
- [`internal/config`](../config/config.md) only for the factory signature.

## Edge cases and notes

- The provider does not distinguish "not set" from "set to empty"; both are
  treated as not found.
- Values are trimmed before decoding, which tolerates stray newlines in CI
  variables.
- The provider is read-only and is not offered by `init`.
- Because it reads the process environment on every call, changing the variable
  between calls is visible immediately.

## Related files

- [`provider.md`](provider.md) — shared interface and `ErrNotFound`.
- [`registry.md`](registry.md) — registers the provider as read-only.
- [`env_test.md`](env_test.md) — tests for decoding and error behavior.
