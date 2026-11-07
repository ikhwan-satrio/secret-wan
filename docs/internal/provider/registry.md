# `internal/provider/registry.go`

> Source: `internal/provider/registry.go` · Package: `provider`

## Purpose

Centralizes provider metadata and construction. Every backend is registered once
so commands, the factory, and configuration checks all agree on the available
provider names.

## Responsibilities

- Map provider names to their factories.
- Record whether a provider is read-only.
- Normalize user input (trim, lowercase, default to `dotfiles`).
- Expose sorted names for help text and error messages.
- Answer `IsKnown`/`IsReadOnly` without constructing a provider.
- Build a provider with `New` or fall back with `GetStorageProvider`.

## Key types

```go
type Factory func(config.AppConfig, string) StorageProvider

type Descriptor struct {
    Name     string
    ReadOnly bool
    New      Factory
}
```

`New` receives the application configuration and the project name. Factories that
do not need an argument ignore it with `_`.

The package-level registry is a plain map:

```go
var registry = map[string]Descriptor{
    "dotfiles": {Name: "dotfiles", New: ...},
    "keyring":  {Name: "keyring",  New: ...},
    "docker":   {Name: "docker",   ReadOnly: true, New: ...},
    "env":      {Name: "env",      ReadOnly: true, New: ...},
    "s3":       {Name: "s3",       New: ...},
}
```

## Functions

| Function | Behavior |
| :--- | :--- |
| `Normalize(name)` | Trims and lowercases; empty becomes `dotfiles` |
| `Names()` | Returns every registered name in sorted order |
| `IsKnown(name)` | `true` when the normalized name is registered |
| `IsReadOnly(name)` | `true` for `docker` and `env` |
| `New(name, cfg, project)` | Builds the provider or returns an unknown-provider error |
| `GetStorageProvider(cfg, name, project)` | `New`, falling back to `dotfiles` on error |

`Names()` currently returns `docker`, `dotfiles`, `env`, `keyring`, `s3`, in that
order.

### `New`

```go
func New(name string, appConfig config.AppConfig, projectName string) (StorageProvider, error) {
    normalized := Normalize(name)
    descriptor, exists := registry[normalized]
    if !exists {
        return nil, fmt.Errorf("unknown provider %q (expected one of: %s)",
            name, strings.Join(Names(), ", "))
    }

    return descriptor.New(appConfig, projectName), nil
}
```

The error lists every valid name, which keeps CLI messages accurate as backends
are added.

### `GetStorageProvider`

```go
func GetStorageProvider(appConfig config.AppConfig, providerName string, projectName string) StorageProvider
```

`GetStorageProvider` is the backward-compatible entry point used by the
application service. Unknown names fall back to `DotfilesProvider` so existing
manifests keep working. Validation is deliberately not performed here;
`check --strict` and `doctor` call `IsKnown` and report the misconfiguration.

## Dependencies

- [`internal/config`](../config/config.md) for `AppConfig.EncFile`.
- The provider implementations in the same package.

## Edge cases and notes

- `Normalize("")` returns `dotfiles`, so `IsKnown("")` is `true`.
- Names are case-insensitive and surrounding whitespace is ignored.
- `IsReadOnly` returns `false` for unknown names; check `IsKnown` first when the
  distinction matters.
- Factories do not perform I/O. A provider such as `s3` validates its
  configuration lazily on the first `Save`/`Load`/`Check`.
- `Descriptor.Name` mirrors its map key; the key is what `Normalize` looks up.

## Related files

- [`provider.md`](provider.md) — interface and not-found handling.
- [`dotfiles.md`](dotfiles.md), [`keyring.md`](keyring.md),
  [`docker.md`](docker.md), [`env.md`](env.md), [`s3.md`](s3.md) — factories.
- [`registry_test.md`](registry_test.md) — tests for names, normalization, and
  fallback behavior.
