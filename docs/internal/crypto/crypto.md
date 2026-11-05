# `internal/crypto/crypto.go`

> Source: `internal/crypto/crypto.go` · Package: `crypto`

## Purpose

Owns the cryptography boundary: locating and loading the age master key,
generating a new key when allowed, generating random passwords, and
encrypting/decrypting the secret map through a storage provider.

## Responsibilities

- Expose the expected key path without creating directories (`KeyPath`).
- Create the config directory with strict permissions when needed
  (`GetKeyPath`).
- Generate unbiased random passwords.
- Load an existing X25519 identity without creating one (`LoadKey`).
- Load or generate the master key (`LoadOrGenerateKey`).
- Serialize secrets to JSON, encrypt with age for one or more recipients, and
  hand bytes to storage.
- Load bytes from storage, decrypt with age, and parse JSON.

## Key type

### `CryptoModule`

```go
type CryptoModule struct {
    AppConfig config.AppConfig
}
```

All methods use the embedded app configuration for the app name and secret
length.

## Key functions

### `KeyPath(appName string) (string, error)`

1. Resolves the OS config directory via `osutil.ConfigDir`.
2. Joins it with `keys.txt`.
3. Does **not** create anything.

This is used by `doctor` so diagnostics have no side effects. It is also the
path used by `LoadKey`.

### `GetKeyPath(appName string) (string, error)`

Same as `KeyPath`, but calls `os.MkdirAll(dir, 0700)` first. This is used by
`LoadOrGenerateKey` before writing a new key.

### `GenerateSecurePassword(length int) (string, error)`

- Rejects `length <= 0`.
- Builds a byte slice of the requested length.
- Uses `crypto/rand.Int` with the charset length, which avoids modulo bias.
- Maps each random index into `passwordCharset`:

```text
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?
```

### `LoadKey() (*age.X25519Identity, error)`

1. Resolves the key path with `KeyPath` (no directory creation).
2. Opens the file; a missing file returns the raw `os.ErrNotExist` so callers
   can detect it with `os.IsNotExist` / `errors.Is`.
3. Parses all age identities with `age.ParseIdentities`.
4. Rejects an empty identity list.
5. Accepts only the first identity if it is `*age.X25519Identity`.
6. Rejects other identity types with `unsupported key type`.

Read-only commands use this method so they never generate a key that cannot
decrypt existing storage.

### `LoadOrGenerateKey() (*age.X25519Identity, error)`

1. Ensures the config directory exists (`GetKeyPath`, mode `0700`).
2. If `keys.txt` exists, delegates to `LoadKey`.
3. If `os.Stat` returns an error other than "not exist", returns it.
4. Otherwise generates an X25519 identity with `age.GenerateX25519Identity`.
5. Creates the file with `0600` and writes:
   - the private identity line
   - a comment line containing the public recipient for sharing

### `SaveSecrets(storage, secrets, recipients...) error`

```go
func (c *CryptoModule) SaveSecrets(
    storage provider.StorageProvider,
    secrets map[string]string,
    recipients ...age.Recipient,
) error
```

1. Rejects an empty recipient list.
2. Creates an in-memory buffer.
3. Starts an age writer with every recipient.
4. Marshals the map to indented JSON.
5. Writes the JSON and closes the age writer (closing finalizes the stream).
6. Calls `storage.Save(buf.Bytes())`.

The provider receives only ciphertext. The payload contains one age recipient
stanza per key, so any matching private key can decrypt it.

The variadic signature keeps the old single-recipient call style valid while
enabling multi-recipient encryption. Passing zero recipients is a programming
error and returns `at least one recipient is required`.

### `LoadSecrets(storage, identity) (map[string]string, error)`

1. Calls `storage.Load()`.
2. Creates an age reader with the identity.
3. Decodes JSON from the decrypted stream.
4. Returns the map.

Storage errors are returned unchanged so callers can use
`provider.IsNotFound`. Decrypt and JSON errors are wrapped with context.

## Dependencies

- `filippo.io/age` for X25519 identities and encryption.
- `internal/config` for app settings.
- `internal/osutil` for config directories.
- `internal/provider` for the storage interface.
- `crypto/rand`, `math/big`, `encoding/json`.

## Edge cases and notes

- The key file contains the **private** identity. Treat it as sensitive; the
  permissions are `0600` and the directory is `0700`.
- The public recipient comment is informational. The recipient is derived from
  the identity at encryption time (`identity.Recipient()`), not read from the
  comment.
- `LoadKey` wraps parse errors with `%w`, so callers can unwrap the underlying
  age error while still detecting a missing file through `os.IsNotExist`.
- `LoadSecrets` may return a nil map if the encrypted JSON is `null`. Callers
  normalize nil maps where needed.
- `LoadSecrets` wraps decrypt and JSON errors with `%w`; this lets tests and
  callers check `age.ErrIncorrectIdentity` with `errors.Is`.
- `GenerateSecurePassword` returns an error if the OS random source fails.
- `SaveSecrets` writes the full map; callers are responsible for merging
  existing values before saving.
- Adding or removing a recipient does not change the plaintext; it only changes
  the recipient stanzas produced by the next `SaveSecrets`.
- `Set` and `sync` call `LoadOrGenerateKey`; read-only commands call `LoadKey`.
- The caller decides which recipients to use. The crypto module does not read
  `secret-wan.recipients`; that policy lives in `cmd/shared.go` and
  `internal/recipients`.

## Related files

- [`crypto_test.md`](crypto_test.md) — multi-recipient encryption tests.
- [`internal/config/config.md`](../config/config.md)
- [`internal/osutil/os.md`](../osutil/os.md)
- [`internal/provider/provider.md`](../provider/provider.md)
- [`internal/recipients/recipients.md`](../recipients/recipients.md) — builds
  the recipient values passed to `SaveSecrets`.
- [`cmd/shared.md`](../../cmd/shared.md) — applies recipient policy.
- [`cmd/doctor.md`](../../cmd/doctor.md) — uses `KeyPath`.
