# `internal/provider/provider.go`

> Source: `internal/provider/provider.go` · Package: `provider`

## Purpose

Defines the storage abstraction and three implementations: local encrypted
file (`dotfiles`), OS keyring (`keyring`), and mounted Docker secret (`docker`).

## Responsibilities

- Define the `StorageProvider` interface (`Save` and `Load` byte slices).
- Distinguish "nothing stored yet" from real failures with `IsNotFound`.
- Read and write `env.enc` for the dotfiles provider.
- Read and write OS keychain entries for the keyring provider.
- Read mounted Docker secrets (saving is unsupported).
- Validate provider availability with optional `Check()` methods.
- Select the correct provider from the manifest's `project.provider`.

## Key interface

```go
type StorageProvider interface {
    Save(data []byte) error
    Load() ([]byte, error)
}
```

Providers move opaque bytes only. Encryption and decryption happen in the crypto
module, so providers never see plaintext (except Docker, whose input is already
an encrypted payload produced by the same pipeline).

## `IsNotFound`

```go
func IsNotFound(err error) bool {
    return errors.Is(err, os.ErrNotExist) || errors.Is(err, keyring.ErrKeyNotFound)
}
```

Commands use this to treat a missing store as a normal first-run condition. A
corrupt store, permission error, or decrypt failure is not "not found".

## `DotfilesProvider`

```go
type DotfilesProvider struct {
    FilePath string
}
```

| Method | Behavior |
| :--- | :--- |
| `Save` | `os.WriteFile(path, data, 0644)` |
| `Load` | `os.ReadFile(path)` |
| `Check` | Stat the file; error if missing, a directory, or empty |

The path is normally `env.enc` from `AppConfig.EncFile`. The ciphertext is safe
to commit, although the file mode is world-readable.

## `KeyringProvider`

```go
type KeyringProvider struct {
    ServiceName string
    Key         string
}
```

| Method | Behavior |
| :--- | :--- |
| `Save` | Opens the keyring with the service name, then `Set(keyring.Item{Key, Data})` |
| `Load` | Opens the keyring, fetches the item by key, returns `item.Data` |
| `Check` | Attempts to open the keyring backend |

The provider factory sets `ServiceName` to `secret-wan` and `Key` to the project
name. On Linux, a Secret Service daemon (GNOME Keyring, KeePassXC, etc.) and
`libsecret` are required; otherwise `keyring.Open` fails.

## `DockerSecretProvider`

```go
type DockerSecretProvider struct {
    SecretName string
}
```

| Method | Behavior |
| :--- | :--- |
| `Save` | Always returns `saving to docker secret is not supported directly` |
| `Load` | Reads `<DockerSecretsDir>/<SecretName>`, trims surrounding whitespace, returns bytes |
| `Check` | Stat the file; error if missing or a directory |

Docker secrets are mounted read-only, so they are a consumption-only provider.
The stored content is expected to be an age-encrypted JSON payload (for example,
a secret created elsewhere and mounted into the container).

## `GetStorageProvider`

```go
func GetStorageProvider(appConfig config.AppConfig, providerName string, projectName string) StorageProvider
```

| `providerName` | Result |
| :--- | :--- |
| `"keyring"` | `KeyringProvider{ServiceName: "secret-wan", Key: projectName}` |
| `"docker"` | `DockerSecretProvider{SecretName: projectName}` |
| anything else (including `""`) | `DotfilesProvider{FilePath: appConfig.EncFile}` |

The default branch makes an omitted provider behave like `dotfiles`. Unknown
provider names also fall back here; `check --strict` and `doctor` are responsible
for flagging them.

## Dependencies

- `github.com/99designs/keyring`.
- `internal/config` for the encrypted file path.
- `internal/osutil` for the Docker secrets directory.

## Edge cases and notes

- Providers do not validate that the payload is valid age data; decryption
  reports that.
- `DotfilesProvider.Save` uses `0644`. Use filesystem permissions or repository
  policy if stricter access is required.
- `DockerSecretProvider.Load` trims whitespace, which can matter if the mounted
  file has a trailing newline.
- `Check` is not part of the `StorageProvider` interface; callers use a type
  assertion (`interface{ Check() error }`) so the interface stays minimal.
- The keyring service name is hardcoded in the factory rather than taken from
  `AppConfig.AppName`.
- Docker secrets cannot be listed or created by this tool.

## Related files

- [`internal/crypto/crypto.md`](../crypto/crypto.md) — encrypts before `Save`,
  decrypts after `Load`.
- [`cmd/doctor.md`](../../cmd/doctor.md) — uses `Check` and `IsNotFound`.
- [`internal/osutil/os.md`](../osutil/os.md) — Docker secrets directory.
