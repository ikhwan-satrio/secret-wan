# `cmd/shared.go`

> Source: `cmd/shared.go` · Package: `cmd`

## Purpose

Contains the plumbing shared by all commands: a session object that bundles
configuration, crypto, manifest parsing, identity, and storage; consistent error
output; and helpers for working with profile keys.

## Responsibilities

- Build an `appSession` from the fixed application configuration.
- Print styled errors to **stderr** and exit with status `1`.
- Parse the manifest and resolve the storage provider exactly once per command.
- Load the age identity either with or without creating a new key.
- Load secrets, with variants for "must exist" and "may be missing".
- Normalize, de-duplicate, and sort profile keys.
- Answer questions such as "is this key declared?", "which rule applies?", and
  "which profiles declare this key?".

## Key types

### `appSession`

```go
type appSession struct {
    Config   config.AppConfig
    Crypto   crypto.CryptoModule
    Parser   manifest.ParserModule
    Manifest *manifest.Manifest
    Identity *age.X25519Identity
    Storage  provider.StorageProvider
}
```

The fields are filled lazily by the `require*` methods. A typical read-only
command uses this order:

```go
session := newSession()
session.requireManifest()          // sets Manifest and Storage
profile := session.requireProfile(profileName)
session.requireExistingIdentity()  // sets Identity, never creates a key
stored := session.requireSecrets() // sets nothing else, returns map
```

### `newSession`

Creates the session with `config.UseApp()`, and initializes the crypto and
parser modules with that configuration. `Manifest`, `Identity`, and `Storage`
start empty.

### `fail` and `failErr`

`fail(shortMsg, detailMsg)` writes `ui.FormatError(...)` to stderr and calls
`os.Exit(1)`. `failErr` is a convenience wrapper that extracts `err.Error()`.

Using stderr keeps stdout usable for values, notably in `print-env` and `get`.

## Key methods

| Method | Behavior |
| :--- | :--- |
| `requireManifest` | Parses `secret-wan.toml`; resolves `Storage` from `project.provider` and `project.name` |
| `requireIdentity` | Calls `LoadOrGenerateKey`; may create the master key and config directory |
| `requireExistingIdentity` | Calls `LoadKey`; reports a friendly "run init" error when the key is missing |
| `requireProfile` | Looks up a profile and exits with the available profile list when missing |
| `requireSecrets` | Loads and decrypts storage; exits with a setup hint when storage does not exist |
| `loadSecretsAllowMissing` | Same as `requireSecrets`, but returns an empty map when storage is absent |

The distinction between `requireIdentity` and `requireExistingIdentity` is
important: read-only commands must not silently generate a new key that cannot
decrypt existing data.

## Key helpers

### `profileKeys`

Returns the union of `Required` and `AutoGen`, sorted and de-duplicated. A key
declared as both required and generated appears once.

### `requiredKeysForPrompt`

Returns required keys that are **not** auto-generated. A key with both rules
should be generated, not prompted.

### `keyInProfile`

Returns true when the key appears in `Rules` or in either `Required` or
`AutoGen`. The slice fallbacks make the helper safe for profiles constructed in
tests without a `Rules` map.

### `ruleFor`

Returns the key's `manifest.SecretRule`, or a zero rule when `Rules` is nil or
the key is absent. This prevents nil-map panics in checks.

### `allDeclaredKeys`

Collects every key from every profile's `Declared` list, then also from
`Required` and `AutoGen` as a defensive fallback. Stored keys outside this set
are "unmanaged".

### `profilesContaining`

Returns sorted profile names that declare a key. Used by `get`, `set`, and
`unset` to explain that a key belongs to another profile.

## Edge cases and notes

- `requireManifest` does not validate that at least one profile exists; the
  parser already rejects manifests with no profiles.
- `requireSecrets` normalizes a nil map to an empty map, so callers can write to
  it safely.
- Error messages are written in English and use the shared UI theme.
- `requireIdentity` creates `~/.config/secret-wan` with mode `0700` as a side
  effect (through the crypto module), so commands that write secrets are
  allowed to bootstrap the key.

## Related files

- [`internal/config/config.md`](../internal/config/config.md)
- [`internal/crypto/crypto.md`](../internal/crypto/crypto.md)
- [`internal/manifest/parser.md`](../internal/manifest/parser.md)
- [`internal/provider/provider.md`](../internal/provider/provider.md)
- [`cmd/shared_test.md`](shared_test.md) — tests for these helpers.
