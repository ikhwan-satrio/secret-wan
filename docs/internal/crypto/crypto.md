# `internal/crypto/crypto.go`

> Source: `internal/crypto/crypto.go` · Package: `crypto`

## Purpose

Defines the `crypto` package and its `CryptoModule` type. The implementation is
split by responsibility:

| File | Responsibility |
| :--- | :--- |
| `key.go` | Key paths, identity loading, identity generation |
| `password.go` | Random password generation |
| `secrets.go` | Encrypt/decrypt the secret map |

## Key type

```go
type CryptoModule struct {
    AppConfig config.AppConfig
}
```

The module carries configuration for:

- the application name used in the config directory path;
- the generated secret length (read by the service, not by the crypto module).

Methods are defined on `*CryptoModule` in the split files. `SaveSecrets` and
`LoadSecrets` do not read configuration; `KeyPath`, `LoadKey`, and
`LoadOrGenerateKey` use `AppConfig.AppName`.

## Design notes

- The package owns the cryptographic boundary and knows nothing about CLI
  rendering or recipient policy.
- Storage is abstracted through `provider.StorageProvider`, so the same crypto
  code works with every supported backend (dotfiles, keyring, Docker, env, S3).
- Splitting the original file keeps key handling, password generation, and
  payload encryption independently readable and testable.

## Dependencies

- `internal/config` for `AppConfig`.
- [`key.md`](key.md), [`password.md`](password.md), [`secrets.md`](secrets.md).

## Related files

- [`key.md`](key.md) — key lifecycle.
- [`password.md`](password.md) — random values.
- [`secrets.md`](secrets.md) — payload encryption.
- [`crypto_test.md`](crypto_test.md) — tests covering the split files.
- [`../provider/provider.md`](../provider/provider.md) — storage interface.
