# `internal/app/project.go`

> Source: `internal/app/project.go` · Package: `app`

## Purpose

Loads and caches the project pieces that most commands need: the parsed manifest,
the storage provider, the age identity, and a named profile.

## Responsibilities

- Parse `secret-wan.toml` and resolve the provider exactly once.
- Load an existing identity or create one when a command is allowed to write.
- Load an existing identity without generating a replacement.
- Resolve a profile and produce a helpful list of alternatives when it is
  missing.

## Methods

### `LoadProject() error`

1. Calls `Parser.ParseManifest()`.
2. Stores the result in `Service.Manifest`.
3. Creates the storage provider from `project.provider` and `project.name`.
4. Stores it in `Service.Storage`.

Parse errors are returned unchanged so the CLI can keep the
`Failed to parse manifest` heading.

### `EnsureProject() error`

Returns immediately when both the manifest and storage are already loaded.
Otherwise it calls `LoadProject`. Commands can call it more than once safely.

### `EnsureIdentity() error`

Returns immediately when `Service.Identity` is set. Otherwise it calls
`Crypto.LoadOrGenerateKey()` and stores the identity.

Use this for write commands such as `set`, `sync`, `init`, and `recipient add`,
where bootstrapping a key is acceptable.

### `LoadIdentity() error`

Returns immediately when an identity is already loaded. Otherwise it calls
`Crypto.LoadKey()` and stores it.

A missing key becomes a `UserError`:

```text
Master key not found
Run 'secwan init' or 'secwan set' first.
```

Read-only commands use this so a missing key never silently creates a new key
that cannot decrypt existing data.

### `Profile(name) (*manifest.Profile, error)`

Calls `EnsureProject`, looks up the profile, and returns a `UserError` when it is
missing. The detail lists the sorted available profiles from
`manifest.GetProfileNames`.

## Edge cases and notes

- `EnsureProject` checks both manifest and storage. A partially initialized
  service reloads both rather than mixing old and new state.
- `LoadIdentity` is idempotent after the identity is loaded, which matters for
  operations such as `recipient remove` that reuse an already loaded key.
- The service does not create config directories in `LoadProject`; key creation
  side effects belong to the crypto module.
- Profile names are case-sensitive, matching TOML table names.

## Dependencies

- `internal/manifest` for parsing and profile types.
- `internal/provider` for storage selection.
- `filippo.io/age` for the identity type.
- Standard library `os` for missing-file detection.

## Related files

- [`service.md`](service.md) — service state and constructors.
- [`secrets.md`](secrets.md) — uses the loaded identity and storage.
- [`../../internal/manifest/profile.md`](../../internal/manifest/profile.md) —
  profile queries.
- [`../../cmd/session.md`](../../cmd/session.md) — CLI wrappers around these
  methods.
