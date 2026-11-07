# `internal/provider/docker.go`

> Source: `internal/provider/docker.go` · Package: `provider`

## Purpose

Implements a read-only provider that loads an encrypted payload from a mounted
Docker secret. It is meant for containers that receive a `secret-wan` payload at
runtime.

## Responsibilities

- Read `<DockerSecretsDir>/<SecretName>`.
- Trim surrounding whitespace from the mounted file.
- Refuse to save, because Docker secrets are read-only mounts.
- Validate that the secret file exists and is not a directory.

## Key type

```go
type DockerSecretProvider struct {
    SecretName string
}
```

The registry sets `SecretName` to the project name.

## Methods

| Method | Behavior |
| :--- | :--- |
| `Save` | Always returns `saving to docker secret is not supported directly` |
| `Load` | Reads the file, trims whitespace, returns the bytes |
| `Check` | `os.Stat`; error if missing or a directory |

## Execution flow

1. `osutil.DockerSecretsDir()` resolves `/run/secrets` on Unix-like systems and
   the platform-appropriate location elsewhere.
2. The secret name is joined to that directory.
3. `Load` reads the whole file, applies `strings.TrimSpace`, and returns the
   ciphertext.
4. The crypto layer decrypts the returned bytes.

## Dependencies

- [`internal/osutil`](../osutil/os.md) for `DockerSecretsDir`.
- The standard library (`fmt`, `os`, `path/filepath`, `strings`).

## Edge cases and notes

- Trimming is important because Docker secrets created from files often end with
  a newline. It also means leading/trailing whitespace in the ciphertext is
  discarded, which is safe for base64 or age payloads.
- The provider cannot create or update the mounted secret. Attempting to `sync`
  or `set` with this provider fails.
- `Check` does not verify the payload is valid age data.
- `Load` of a missing file returns `os.ErrNotExist`, accepted by `IsNotFound`.
- The provider is not offered by `init`; set `project.provider = "docker"`
  manually.

## Related files

- [`provider.md`](provider.md) — shared interface and not-found handling.
- [`registry.md`](registry.md) — registers the provider as read-only.
- [`internal/osutil/os.md`](../osutil/os.md) — Docker secrets directory helper.
