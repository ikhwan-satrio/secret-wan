# `internal/provider/provider.go`

> Source: `internal/provider/provider.go` · Package: `provider`

## Purpose

Defines the storage abstraction shared by every provider and the sentinel error
used to distinguish an empty store from a broken one. The implementations are
split by backend:

## Package layout

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/provider/provider.go` | [`provider.md`](provider.md) | Interface, `ErrNotFound`, `IsNotFound` |
| `internal/provider/registry.go` | [`registry.md`](registry.md) | Provider metadata, factory, `GetStorageProvider` |
| `internal/provider/dotfiles.go` | [`dotfiles.md`](dotfiles.md) | Local encrypted file backend |
| `internal/provider/keyring.go` | [`keyring.md`](keyring.md) | OS keyring backend |
| `internal/provider/docker.go` | [`docker.md`](docker.md) | Mounted Docker secret backend (read-only) |
| `internal/provider/env.go` | [`env.md`](env.md) | Base64 environment variable backend (read-only) |
| `internal/provider/s3.go` | [`s3.md`](s3.md) | S3-compatible object storage backend |
| `internal/provider/env_test.go` | [`env_test.md`](env_test.md) | Tests for the environment provider |
| `internal/provider/registry_test.go` | [`registry_test.md`](registry_test.md) | Tests for the provider registry |
| `internal/provider/s3_test.go` | [`s3_test.md`](s3_test.md) | Tests for the S3 provider |

## Core interface

```go
type StorageProvider interface {
    Save(data []byte) error
    Load() ([]byte, error)
}
```

`Save` and `Load` receive already-encrypted bytes. The interface deliberately
stays minimal so a backend only has to move bytes. Backends that can validate
availability expose an additional `Check() error` method, which callers use
through a type assertion:

```go
if checker, ok := storage.(interface{ Check() error }); ok {
    err := checker.Check()
}
```

`Check` is not part of `StorageProvider`; adding it would force every backend
(including future ones) to implement validation logic it may not need.

## Not-found handling

```go
var ErrNotFound = errors.New("secret not found")

func IsNotFound(err error) bool {
    return errors.Is(err, ErrNotFound) ||
        errors.Is(err, os.ErrNotExist) ||
        errors.Is(err, keyring.ErrKeyNotFound)
}
```

`IsNotFound` lets commands treat "nothing has been stored yet" as a normal
first-run condition. A corrupt payload, permission error, or decryption failure
is not "not found".

## Registered providers

| Name | Read-only | Backing store | Selected by `init` |
| :--- | :--- | :--- | :--- |
| `dotfiles` | no | Configured `EncFile` (default `env.enc`) | yes (default) |
| `keyring` | no | OS credential store | yes |
| `docker` | yes | `/run/secrets/<name>` (or OS equivalent) | no |
| `env` | yes | `SECWAN_ENV_ENC` environment variable | no |
| `s3` | no | One object in an S3-compatible bucket | no |

`init` still offers only `dotfiles` and `keyring`. The other providers are chosen
by editing `project.provider` in `secret-wan.toml`.

## Selection flow

```text
manifest project.provider
        │
        ▼
provider.GetStorageProvider(appConfig, name, projectName)
        │
        ├─ known name ──► registry factory ──► concrete provider
        │
        └─ unknown name ─► DotfilesProvider (backward-compatible fallback)
```

`check --strict` and `doctor` consult the registry instead of hardcoding names,
so they report the same set of providers the factory understands. Unknown names
are flagged there rather than silently falling back.

## Dependencies

- `github.com/99designs/keyring` for OS credential storage.
- `github.com/aws/aws-sdk-go-v2` (config, S3, smithy-go) for the S3 provider.
- [`internal/config`](../config/config.md) for the encrypted file path.
- [`internal/osutil`](../osutil/os.md) for the Docker secrets directory.

## Edge cases and notes

- Providers do not validate that the payload is valid age data; decryption
  reports that.
- `DotfilesProvider.Save` uses `0644`; use filesystem policy if stricter access
  is required.
- The `env` and `docker` providers cannot save. Commands that write secrets fail
  with an actionable message when they are selected.
- The S3 provider reads all of its configuration from `SECWAN_S3_*` environment
  variables; nothing bucket-related is written to the manifest.
- `GetStorageProvider` still falls back to `dotfiles` for unknown names to keep
  existing manifests working. Validation lives in `check` and `doctor`.

## Related files

- [`internal/app/service.md`](../app/service.md) — owns the selected provider.
- [`cmd/doctor_checks.md`](../../cmd/doctor_checks.md) — calls `Check` and reports
  read-only mode.
- [`cmd/check.md`](../../cmd/check.md) — validates provider names in strict mode.
- [`internal/crypto/crypto.md`](../crypto/crypto.md) — encrypts before `Save` and
  decrypts after `Load`.
