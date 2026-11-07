# `internal/app/service.go`

> Source: `internal/app/service.go` · Package: `app`

## Purpose

Defines the application service that sits between the Cobra command layer and the
domain packages. It owns the state loaded during one CLI invocation and exposes
operations such as loading a project, reading secrets, rekeying, and changing the
recipient list.

The service is where business rules live. Commands should mostly parse flags,
call the service, and render the result.

## Responsibilities

- Hold the application configuration and the long-lived domain modules.
- Hold the lazily loaded manifest, identity, storage provider, and recipient
  store.
- Provide constructors for production and for tests.
- Give each command a single object to coordinate instead of rebuilding modules
  by hand.

## Key type

```go
type Service struct {
    Config     config.AppConfig
    Crypto     crypto.CryptoModule
    Parser     manifest.ParserModule
    Recipients recipients.Store

    Manifest *manifest.Manifest
    Identity *age.X25519Identity
    Storage  provider.StorageProvider
}
```

The first group is initialized by the constructor. The second group is filled
lazily by methods in `project.go`, `secrets.go`, and `recipients.go`.

## Constructors

### `New()`

Builds a service with `config.UseApp()`. This is the production path used by the
CLI session wrapper.

### `NewWithConfig(appConfig)`

Builds a service with an explicit configuration. Tests use this to point the
master key, manifest, and recipient file at temporary locations without changing
the caller's environment.

Both constructors create:

- `crypto.CryptoModule` from the configuration;
- `manifest.ParserModule` from the configuration;
- `recipients.Store` for `AppConfig.RecipientsPath()`.

## Design notes

- The service is deliberately not a singleton. One CLI command creates one
  service, uses it, and exits.
- Loading is lazy so read-only commands can avoid creating a master key or
  touching unrelated state.
- The service returns `UserError` for expected user-facing failures and plain
  errors for infrastructure failures. The command layer decides how to render
  both.
- The service does not print anything. All rendering stays in `cmd`.

## Dependencies

- `filippo.io/age` for the identity type.
- `internal/config`, `internal/crypto`, `internal/manifest`,
  `internal/provider`, `internal/recipients`.

## Related files

- [`project.md`](project.md) — project, identity, and profile loading.
- [`secrets.md`](secrets.md) — secret loading, saving, and generation.
- [`recipients.md`](recipients.md) — recipient policy and rekey operations.
- [`init.md`](init.md) — bootstrap helpers used by `secwan init`.
- [`errors.md`](errors.md) — CLI-safe error types.
- [`../../cmd/session.md`](../../cmd/session.md) — command-layer wrapper.
