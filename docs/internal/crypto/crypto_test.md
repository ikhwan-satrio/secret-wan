# `internal/crypto/crypto_test.go`

> Source: `internal/crypto/crypto_test.go` · Package: `crypto`

## Purpose

Unit tests for multi-recipient encryption, recipient validation, and password
generation. The tests use an in-memory storage provider, so they never touch the
user's project files or keyring.

## Responsibilities

- Verify that one ciphertext can be decrypted by every configured recipient.
- Verify that an unrelated identity cannot decrypt it.
- Verify that saving without a recipient is rejected.
- Verify password length and the invalid-length error.

## Test helper

### `memoryStorage`

```go
type memoryStorage struct {
    data []byte
}
```

Implements `provider.StorageProvider`:

| Method | Behavior |
| :--- | :--- |
| `Save(data)` | Copies the ciphertext into memory |
| `Load()` | Returns the bytes, or `os.ErrNotExist` when nothing was saved |

The copied slice prevents accidental aliasing between the writer and reader.

## Tests

### `TestSaveSecretsEncryptsForEveryRecipient`

1. Generates three X25519 identities: `first`, `second`, and `outsider`.
2. Saves a two-key secret map with `first.Recipient()` and
   `second.Recipient()`.
3. Decrypts with `first` and `second`; both must return the original map.
4. Decrypts with `outsider` and expects `errors.Is(err, age.ErrIncorrectIdentity)`.

This is the core regression test for multi-recipient behavior. The
`errors.Is` check also verifies that `LoadSecrets` wraps the age error with
`%w` rather than flattening it to a string.

### `TestSaveSecretsRequiresRecipient`

Calls `SaveSecrets` with no recipients and expects an error. This guards the
invariant that a payload always has at least one intended reader.

### `TestGenerateSecurePassword`

- Requests 32 characters and checks the length.
- Requests 0 characters and expects an error.

The test does not assert character classes, because the generator is random.

## Edge cases and notes

- The test module has no `AppConfig`; `CryptoModule{}` is sufficient because
  `SaveSecrets` and `LoadSecrets` do not read configuration.
- The outsider failure message currently includes age's
  `incorrect identity for recipient block` text, but the assertion is based on
  the typed error, not the message.
- Storage errors are not simulated here; provider-specific tests cover them.
- The test does not verify ciphertext format, only observable decryptability.

## Related files

- [`crypto.md`](crypto.md) — implementation under test.
- [`../../internal/provider/provider.md`](../provider/provider.md) — storage
  interface used by `memoryStorage`.
- [`../../cmd/recipient.md`](../../cmd/recipient.md) — command layer that builds
  multi-recipient lists.
