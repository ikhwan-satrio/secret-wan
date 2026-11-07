# `internal/crypto/secrets.go`

> Source: `internal/crypto/secrets.go` · Package: `crypto`

## Purpose

Encrypts and decrypts the secret map through a `provider.StorageProvider`. The
payload is a JSON object encrypted with age to one or more recipients.

## Responsibilities

- Reject saves with no recipient.
- Marshal the full map to indented JSON.
- Encrypt with every supplied age recipient.
- Hand only ciphertext to storage.
- Read ciphertext, decrypt with an identity, and parse JSON.

## Functions

### `SaveSecrets(storage, secrets, recipients...) error`

1. Rejects an empty recipient list with `at least one recipient is required`.
2. Creates an in-memory buffer.
3. Starts an age writer with every recipient.
4. Marshals the map to indented JSON.
5. Writes the JSON and closes the age writer (closing finalizes the stream).
6. Calls `storage.Save(buf.Bytes())`.

The payload contains one age recipient stanza per key, so any matching private
key can decrypt it. Passing zero recipients is a programming error.

### `LoadSecrets(storage, identity) (map[string]string, error)`

1. Calls `storage.Load()`.
2. Creates an age reader with the identity.
3. Decodes JSON from the decrypted stream.
4. Returns the map.

Storage errors are returned unchanged so callers can use
`provider.IsNotFound`. Decrypt and JSON errors are wrapped with context and `%w`,
which lets tests check `age.ErrIncorrectIdentity` with `errors.Is`.

## Edge cases and notes

- `SaveSecrets` writes the full map; callers merge existing values first.
- A valid encrypted `null` payload can produce a nil map. The application
  service normalizes nil maps.
- Adding or removing a recipient does not change plaintext; it changes the
  recipient stanzas produced by the next save.
- The crypto module does not read `secret-wan.recipients`; recipient policy
  lives in `internal/app`.
- Docker storage is read-only, so a save through that provider fails in the
  provider, not here.

## Dependencies

- `filippo.io/age` for encryption and decryption.
- `internal/provider` for the storage interface.
- Standard library `bytes`, `encoding/json`, and `fmt`.

## Related files

- [`crypto.md`](crypto.md) — module type.
- [`key.md`](key.md) — identity loading.
- [`../../internal/app/secrets.md`](../../internal/app/secrets.md) — recipient and
  missing-store policy.
- [`../../internal/provider/provider.md`](../../internal/provider/provider.md) —
  storage implementations.
