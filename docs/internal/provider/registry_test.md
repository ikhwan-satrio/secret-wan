# `internal/provider/registry_test.go`

> Source: `internal/provider/registry_test.go` · Package: `provider`

## Purpose

Unit tests for the provider registry. They lock down the registered name set,
normalization rules, read-only metadata, factory output types, and the legacy
fallback behavior.

## Test cases

| Test | What it proves |
| :--- | :--- |
| `TestNamesAreSortedAndComplete` | `Names()` returns exactly `[docker dotfiles env keyring s3]` |
| `TestIsKnownNormalizesNames` | Empty, lowercase, uppercase, and padded names are known; `vault` is not |
| `TestIsReadOnly` | `docker` and `env` are read-only; `s3` and unknown names are not |
| `TestNewBuildsKnownProviders` | Every name maps to the expected concrete type |
| `TestNewRejectsUnknownProvider` | `New("vault", ...)` returns an error |
| `TestGetStorageProviderFallsBackToDotfiles` | Unknown names produce a `DotfilesProvider` with the manifest path |

## Key code

```go
cases := map[string]any{
    "dotfiles": &DotfilesProvider{},
    "keyring":  &KeyringProvider{},
    "docker":   &DockerSecretProvider{},
    "env":      &EnvProvider{},
    "s3":       &S3Provider{},
}
```

The test compares `reflect.TypeOf(instance)` with `reflect.TypeOf(wantType)` so
it only asserts the concrete backend type, not its internal configuration.

## Dependencies

- `reflect` and `testing` from the standard library.
- [`internal/config`](../config/config.md) for constructing an `AppConfig`.
- [`registry.md`](registry.md) — the code under test.

## Edge cases and notes

- The expected name list is intentionally exact. Adding a provider requires
  updating this test and the documentation index.
- `IsKnown("")` is expected to be true because `Normalize("")` returns the
  default provider.
- The tests use `testEncFile = "secrets.enc"` as a custom `AppConfig.EncFile`,
  proving the registry propagates the configured path instead of a hardcoded
  default.
- The fallback test verifies both the type and the propagated `EncFile` path.
- Tests do not touch the keyring, filesystem, or network.

## Related files

- [`registry.md`](registry.md) — registry implementation.
- [`provider.md`](provider.md) — shared interface.
