# `internal/config/config.go`

> Source: `internal/config/config.go` · Package: `config`

## Purpose

Defines the application's fixed configuration: where the project manifest,
encrypted file, and public recipient list live, where the global master key
lives, the app name, and the default generated-secret length.

## Responsibilities

- Provide the `AppConfig` value consumed by the crypto, manifest, provider,
  recipients, and command layers.
- Resolve the recipients path with a safe default.
- Resolve the OS-specific master key path.
- Fail fast when the user config directory cannot be resolved.

## Key types

### `AppConfig`

| Field | Value from `UseApp` | Meaning |
| :--- | :--- | :--- |
| `SpecFile` | `secret-wan.toml` | Project manifest in the current directory |
| `EncFile` | `env.enc` | Encrypted payload for the `dotfiles` provider |
| `RecipientsFile` | `secret-wan.recipients` | Public age recipient list |
| `GlobalMasterKey` | OS config path for `keys.txt` | Informational master-key path |
| `AppName` | `secret-wan` | Used for config directory lookup and display |
| `SecretLength` | `100` | Length of generated passwords |

`EncFile` is the single source of truth for the encrypted filename. Providers
read `AppConfig.EncFile` rather than hardcoding a name, so changing this one
value in `UseApp` moves the payload for every backend that derives its location
from it (the dotfiles file and the default S3 object key).

### `UseApp()`

```go
func UseApp() AppConfig {
    configPath, err := osutil.ConfigPath("secret-wan", "keys.txt")
    if err != nil {
        log.Fatalln(err)
    }

    return AppConfig{
        SpecFile:        "secret-wan.toml",
        EncFile:         "env.enc",
        RecipientsFile:  "secret-wan.recipients",
        GlobalMasterKey: configPath,
        AppName:         "secret-wan",
        SecretLength:    100,
    }
}
```

The configuration is intentionally not user-configurable yet. All relative paths
are resolved against the current working directory.

### `RecipientsPath()`

```go
func (c AppConfig) RecipientsPath() string {
    if c.RecipientsFile != "" {
        return c.RecipientsFile
    }
    return "secret-wan.recipients"
}
```

Commands and tests sometimes construct `AppConfig` directly without calling
`UseApp`. The method guarantees that those callers still use the standard sidecar
file instead of an empty path.

## Dependencies

- `internal/osutil` for the OS-specific config path.
- Standard library `log`.

## Edge cases and notes

- `UseApp` calls `log.Fatalln` when `os.UserConfigDir()` fails. This is a rare
  environment problem (for example, no home directory) and terminates the
  process.
- `GlobalMasterKey` is computed but the crypto module resolves the key path
  independently through `osutil.ConfigDir`; both resolve to the same
  `<config-dir>/secret-wan/keys.txt`.
- The manifest, encrypted file, and recipient file names are constants; changing
  them in code changes the CLI's expectations.
- `SecretLength` is used by `GenerateSecurePassword` in `set --generate` and
  `sync`.
- `RecipientsPath` keeps its default literal in this package so `config` owns
  every default project filename and does not depend on the recipient
  implementation. `recipients.DefaultFile` mirrors the same value for callers
  that use that package directly.

## Related files

- [`internal/osutil/os.md`](../osutil/os.md) — config directory rules.
- [`internal/crypto/crypto.md`](../crypto/crypto.md) — consumes the app name.
- [`internal/manifest/parser.md`](../manifest/parser.md) — consumes `SpecFile`.
- [`internal/provider/provider.md`](../provider/provider.md) — consumes `EncFile`.
- [`internal/recipients/recipients.md`](../recipients/recipients.md) — consumes
  `RecipientsFile` through `RecipientsPath`.
