# `internal/provider/keyring.go`

> Source: `internal/provider/keyring.go` · Package: `provider`

## Purpose

Implements the OS credential-store provider using
[`github.com/99designs/keyring`](https://github.com/99designs/keyring). Secrets
never touch the project directory.

## Responsibilities

- Store the encrypted payload in the OS keyring under a service name and key.
- Read the payload back.
- Validate that a keyring backend can be opened.

## Key type

```go
type KeyringProvider struct {
    ServiceName string
    Key         string
}
```

The registry sets `ServiceName` to `secret-wan` and `Key` to the project name.

## Methods

| Method | Behavior |
| :--- | :--- |
| `Save` | Opens the keyring, then `Set(keyring.Item{Key, Data})` |
| `Load` | Opens the keyring, `Get(Key)`, returns `item.Data` |
| `Check` | Attempts to open the keyring backend |

Each call opens the keyring independently; the provider does not cache a handle.

## Execution flow

1. `keyring.Open(keyring.Config{ServiceName: p.ServiceName})` selects the
   platform backend (macOS Keychain, Windows Credential Manager, Secret Service,
   etc.).
2. `Save` writes one item; `Load` reads it by key.
3. `Check` only verifies that a backend can be opened. It does not prove that the
   item exists.

## Dependencies

- `github.com/99designs/keyring`, which pulls in platform-specific backends
  (`go-keychain`, `wincred`, `dbus`, `go-libsecret`, `jose2go`).
- [`internal/config`](../config/config.md) indirectly through the registry.

## Edge cases and notes

- On Linux, `libsecret` and a Secret Service daemon (GNOME Keyring, KeePassXC,
  etc.) are required; otherwise `keyring.Open` fails.
- A missing item returns `keyring.ErrKeyNotFound`, which `IsNotFound` accepts.
- The service name is hardcoded in the registry rather than taken from
  `AppConfig.AppName`.
- Keyring contents are per-user and do not travel with the repository, so sharing
  requires the recipient workflow plus an out-of-band transfer of the payload.
- `Check` succeeding while the item is absent is expected before the first save.

## Related files

- [`provider.md`](provider.md) — shared interface and not-found handling.
- [`registry.md`](registry.md) — wires the service name and project key.
- [`internal/crypto/key.md`](../crypto/key.md) — master key location.
