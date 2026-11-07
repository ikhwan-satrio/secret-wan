# `internal/app/secrets.go`

> Source: `internal/app/secrets.go` · Package: `app`

## Purpose

Owns secret-map loading, first-write tolerance, saving, and random value
generation. It keeps the "empty store means first write" policy out of the
commands.

## Responsibilities

- Load and decrypt the stored map.
- Distinguish a missing store from a corrupt or unreadable one.
- Return an empty map when a write command is allowed to create storage.
- Save the full map through the configured recipient policy.
- Generate values with the configured length.
- Normalize nil maps so callers can write safely.

## Methods

### `LoadSecrets() (map[string]string, error)`

Loads secrets and reports a missing store as a `UserError`:

```text
Secrets storage not found
Run 'secwan sync' or 'secwan set <KEY>' to create it first.
```

Read-only commands use this variant.

### `LoadSecretsAllowMissing() (map[string]string, bool, error)`

Returns `(map, missing, error)`. When storage does not exist it returns an empty
map and `missing=true` instead of an error. `set` uses the flag to print a
warning; `sync` ignores it because a first sync is expected to start empty.

### `SaveSecrets(secrets, allowSelfExclusion) (int, error)`

1. Resolves recipients through `EncryptionRecipients`.
2. Encrypts and saves the full map.
3. Returns the number of recipients used.

The self-exclusion flag is false for normal writes and true only for operations
that intentionally remove the current key.

### `GenerateSecret() (string, error)`

Delegates to `Crypto.GenerateSecurePassword` with `AppConfig.SecretLength`.

### `isNotFound(err) bool`

Thin wrapper around `provider.IsNotFound`, used by the recipient and rekey code.

## Missing-store policy

| Caller | Method | Missing store |
| :--- | :--- | :--- |
| `list`, `get`, `check`, `unset`, `print-env`, `exec` | `LoadSecrets` | setup error |
| `set` | `LoadSecretsAllowMissing` | empty map plus warning |
| `sync` | `LoadSecretsAllowMissing` | empty map, silent |
| `rekey` | `Rekey` | "nothing to rekey" result |
| `recipient add/remove` | `Rekey` | "nothing to rekey" result |

## Edge cases and notes

- `LoadSecrets` never creates storage. Only a successful `SaveSecrets` does.
- A nil map from a valid but empty JSON payload is normalized to an empty map.
- `SaveSecrets` saves the whole map, not a delta. Callers merge first so other
  profiles and unmanaged keys survive.
- The returned recipient count is the count of age recipients actually used,
  which may differ from the raw file length only if parsing fails first.
- Decryption and JSON errors are returned unchanged from the crypto module; the
  CLI renders them under a generic heading.
- `GenerateSecret` uses the same charset and bias-free random source as `sync`.

## Dependencies

- `internal/provider` for not-found detection.
- `internal/crypto` for encrypt/decrypt and password generation.
- [`recipients.md`](recipients.md) for recipient policy.

## Related files

- [`project.md`](project.md) — loads the identity and storage used here.
- [`../../internal/crypto/secrets.md`](../../internal/crypto/secrets.md) —
  encryption implementation.
- [`../../cmd/set.md`](../../cmd/set.md) and
  [`../../cmd/sync.md`](../../cmd/sync.md) — primary callers.
