# `internal/osutil/os.go`

> Source: `internal/osutil/os.go` · Package: `osutil`

## Purpose

Small cross-platform utility layer for operating-system detection and standard
config/Docker paths.

## Responsibilities

- Define a typed `OS` value with constants for supported platforms.
- Report the current OS and common predicates.
- Provide a generic map lookup with fallback.
- Resolve the user's home directory and OS config directory.
- Build app config paths.
- Resolve the default Docker secrets mount directory.

## Key types and constants

```go
type OS string

const (
    Linux   OS = "linux"
    Windows OS = "windows"
    MacOS   OS = "darwin"
    FreeBSD OS = "freebsd"
    OpenBSD OS = "openbsd"
    NetBSD  OS = "netbsd"
    Android OS = "android"
    IOS     OS = "ios"
)
```

The values match `runtime.GOOS`. Using constants avoids remembering that macOS
is `darwin` in Go.

## Key functions

### `CurrentOS`, `IsWindows`, `IsLinux`, `IsMacOS`

Thin wrappers around `runtime.GOOS`. The predicates are used in tests and could
be used for platform-specific branches.

### `GetByOS`

Returns the value registered for the current OS, or `fallback` when the OS is
not in the map. Generics keep the helper type-safe for strings and other values.

### `HomeDir`

Wraps `os.UserHomeDir`.

### `ConfigDir(appName string) (string, error)`

1. Rejects an empty app name.
2. Calls `os.UserConfigDir`.
3. Joins the app name onto the base directory.

Resulting locations by OS:

| OS | Location |
| :--- | :--- |
| Windows | `%AppData%\<appName>` |
| macOS | `~/Library/Application Support/<appName>` |
| Linux/Unix | `$XDG_CONFIG_HOME/<appName>` or `~/.config/<appName>` |

### `ConfigPath(appName, fileName string) (string, error)`

Rejects an empty file name, resolves `ConfigDir`, and joins the file name. This
is used by `config.UseApp` to compute the master key path.

### `DockerSecretsDir`

Returns:

| OS | Directory |
| :--- | :--- |
| Windows | `C:\ProgramData\Docker\secrets` |
| everything else | `/run/secrets` |

## Dependencies

- Standard library `errors`, `os`, `path/filepath`, `runtime`.

## Edge cases and notes

- `os.UserConfigDir` can fail when no home directory is available; the error is
  passed through.
- `ConfigDir` and `ConfigPath` perform validation but do not create directories.
  Directory creation belongs to the crypto module.
- `GetByOS` with a nil map simply returns the fallback.
- `DockerSecretsDir` is a default only; Docker can be configured with a
  different mount point, but the provider does not read environment variables
  for it.
- `IsMacOS` compares against `MacOS`, whose underlying value is `"darwin"`.

## Related files

- [`internal/osutil/os_test.md`](os_test.md) — tests for these helpers.
- [`internal/config/config.md`](../config/config.md) — uses `ConfigPath`.
- [`internal/crypto/crypto.md`](../crypto/crypto.md) — uses `ConfigDir`.
- [`internal/provider/provider.md`](../provider/provider.md) — uses
  `DockerSecretsDir`.
