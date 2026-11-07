# `internal/provider/env_test.go`

> Source: `internal/provider/env_test.go` · Package: `provider`

## Purpose

Unit tests for the environment provider. They cover base64 decoding, missing
variables, invalid input, and the read-only contract.

## Test cases

| Test | What it proves |
| :--- | :--- |
| `TestEnvProviderLoadsBase64` | Standard padded base64 decodes to the original bytes; `Check` succeeds |
| `TestEnvProviderMissingVariableIsNotFound` | An unset variable produces an error matched by `IsNotFound` |
| `TestEnvProviderRejectsInvalidBase64` | Garbage input returns a decode error that is *not* `ErrNotFound` |
| `TestEnvProviderSaveIsUnsupported` | `Save` always returns an error |
| `TestEnvProviderAcceptsURISafeAndRawBase64` | URL-safe and unpadded alphabets decode correctly |

## Key code

```go
t.Setenv(DefaultEnvVar, base64.StdEncoding.EncodeToString(payload))

got, err := (&EnvProvider{}).Load()
```

`t.Setenv` scopes the variable to the test and restores the previous value
automatically. An empty `VarName` exercises the default-variable fallback.

The invalid-input test distinguishes decode failures from "not found":

```go
if errors.Is(err, ErrNotFound) {
    t.Fatalf("Load() error = %v, want a decode error, not ErrNotFound", err)
}
```

## Dependencies

- `encoding/base64`, `errors`, and `testing` from the standard library.
- [`env.md`](env.md) — the code under test.

## Edge cases and notes

- Tests run in the same package (`package provider`) so they can construct
  unexported fields and call `decodeBase64` indirectly.
- The URL-safe test payload deliberately contains bytes that differ between the
  standard and URL-safe alphabets, so a wrong decoder would fail.
- No network or filesystem access is required.

## Related files

- [`env.md`](env.md) — environment provider implementation.
- [`provider.md`](provider.md) — `ErrNotFound` and `IsNotFound`.
