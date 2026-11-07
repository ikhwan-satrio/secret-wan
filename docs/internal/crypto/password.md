# `internal/crypto/password.go`

> Source: `internal/crypto/password.go` · Package: `crypto`

## Purpose

Generates cryptographically secure random passwords for `sync`, `set
--generate`, and any other caller that needs a generated secret value.

## Responsibilities

- Reject non-positive lengths.
- Draw unbiased random indices.
- Map indices into the project's password charset.
- Return a string of exactly the requested length.

## Key code

### `GenerateSecurePassword(length int) (string, error)`

1. Rejects `length <= 0` with `password length must be greater than 0`.
2. Allocates a byte slice of the requested length.
3. Uses `crypto/rand.Int` with the charset length, which avoids modulo bias.
4. Maps each index into:

```text
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?
```

### `passwordCharset`

Includes lowercase, uppercase, digits, and a broad set of punctuation. It does
not guarantee that every class appears in a generated value; the generator is
uniform over the charset, not over classes.

## Edge cases and notes

- A random source failure returns a wrapped error.
- The generated length comes from `AppConfig.SecretLength` (100 by default) in
  the application service.
- The value is returned to the caller; the crypto package does not write it to
  disk or print it.
- Generated values are not rotated automatically. `sync` keeps an existing
  non-empty value.

## Dependencies

- Standard library `crypto/rand`, `math/big`, and `fmt`.

## Related files

- [`crypto.md`](crypto.md) — module type.
- [`../../internal/app/secrets.md`](../../internal/app/secrets.md) — configured
  generation.
- [`../../cmd/sync.md`](../../cmd/sync.md) and
  [`../../cmd/set.md`](../../cmd/set.md) — callers.
