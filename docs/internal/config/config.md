# `internal/config/config.go`

> Source: `internal/config/config.go` · Package: `config`

## Purpose

Defines the application's fixed configuration: where the project manifest and
encrypted file live, where the global master key lives, the app name, and the
default generated-secret length.

## Responsibilities

- Provide the `AppConfig` value consumed by the crypto, manifest, provider, and
  command layers.
- Resolve the OS-specific master key path.
- Fail fast when the user config directory cannot be resolved.

## Key types

### `AppConfig`

| Field | Value from `UseApp` | Meaning |
| :--- | :--- | :--- |
| `SpecFile` | `secret-wan.toml` | Project manifest in the current directory |
| `EncFile` | `env.enc` | Encrypted payload for the `dotfiles` provider |
| `GlobalMasterKey` | OS config path for `keys.txt` | Informational master-key path |
| `AppName` | `secret-wan` | Used for config directory lookup and display |
| `SecretLength` | `100` | Length of generated passwords |

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
        GlobalMasterKey: configPath,
        AppName:         "secret-wan",
        SecretLength:    100,
    }
}
```

The configuration is intentionally not user-configurable yet. All relative paths
are resolved against the current working directory.

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
- The manifest and encrypted file names are constants; changing them in code
  changes the CLI's expectations.
- `SecretLength` is used by `GenerateSecurePassword` in `set --generate` and
  `sync`.

## Related files

- [`internal/osutil/os.md`](../osutil/os.md) — config directory rules.
- [`internal/crypto/crypto.md`](../crypto/crypto.md) — consumes the app name.
- [`internal/manifest/parser.md`](../manifest/parser.md) — consumes `SpecFile`.
- [`internal/provider/provider.md`](../provider/provider.md) — consumes `EncFile`.
