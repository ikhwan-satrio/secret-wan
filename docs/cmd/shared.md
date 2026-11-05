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
- Resolve the project's public recipient list into age recipients.
- Encrypt and save the whole secret map for every configured recipient.
- Re-encrypt a stored map without changing its contents.
- Refuse writes that would lock the current key out.
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
| `recipientsStore` | Returns the `recipients.Store` for the project's sidecar file |
| `recipientKeys` | Loads the normalized public recipient strings |
| `encryptionRecipients` | Resolves strings into `[]age.Recipient` and enforces self-inclusion |
| `saveSecrets` | Encrypts and saves the map for every configured recipient |
| `rekey` | Decrypts and re-encrypts the current map |

The distinction between `requireIdentity` and `requireExistingIdentity` is
important: read-only commands must not silently generate a new key that cannot
decrypt existing data.

## Recipient helpers

### `recipientsStore` and `recipientKeys`

`recipientsStore` builds a store for `Config.RecipientsPath()`.
`recipientKeys` calls `Load`. A missing file returns an empty slice, which means
"single-recipient mode".

### `encryptionRecipients(appConfig, identity, allowSelfExclusion)`

This package-level function is used both by the session method and by `sync`,
which does not build an `appSession`.

1. Rejects a nil identity.
2. Loads the recipient list.
3. Falls back to `[]string{identity.Recipient().String()}` when the list is
   empty.
4. Unless `allowSelfExclusion` is true, fails when the current key is absent:

```text
current key age1... is not a recipient; run 'secwan recipient add age1...' first
```

5. Parses the strings into `[]age.Recipient`.

The self-inclusion guard is what prevents an accidental lockout during `set`,
`unset`, and `sync`.

### `saveSecrets`

Calls `encryptionRecipients(false)`, then `Crypto.SaveSecrets` with every
resolved recipient. `set` and `unset` use this instead of passing the current
key directly.

### `rekeySecrets`

1. Loads and decrypts the stored map.
2. Treats `provider.IsNotFound` as "nothing to rekey" and returns `0, nil`.
3. Resolves recipients with `encryptionRecipients`.
4. Saves the same map encrypted to every recipient.
5. Returns the number of secrets.

The `allowSelfExclusion` parameter is set by `recipient remove` and
`rekey --force` when the current key is intentionally excluded.

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
- An empty recipient list is not an error; it keeps projects created before
  multi-recipient support working with the current key.
- A non-empty recipient list without the current key is an error for normal
  writes, even if another recipient could still decrypt the data.
- `rekeySecrets` never modifies the recipient file. It only reads it.
- `sync` calls the package-level `encryptionRecipients` directly because it does
  not construct an `appSession`.

## Related files

- [`internal/config/config.md`](../internal/config/config.md)
- [`internal/crypto/crypto.md`](../internal/crypto/crypto.md)
- [`internal/manifest/parser.md`](../internal/manifest/parser.md)
- [`internal/provider/provider.md`](../internal/provider/provider.md)
- [`internal/recipients/recipients.md`](../internal/recipients/recipients.md)
- [`cmd/recipient.md`](recipient.md) — recipient commands.
- [`cmd/rekey.md`](rekey.md) — standalone re-encryption.
- [`cmd/shared_test.md`](shared_test.md) — tests for these helpers.
