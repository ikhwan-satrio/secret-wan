# `internal/manifest/parser.go`

> Source: `internal/manifest/parser.go` · Package: `manifest`

## Purpose

Parses `secret-wan.toml` into a typed `Manifest`, expanding TOML rule tables
into sorted key lists and preserving the raw rules for strict validation.

## Responsibilities

- Read the manifest file configured by `AppConfig.SpecFile`.
- Decode TOML with `github.com/BurntSushi/toml`.
- Convert each `[profiles.<name>]` table into a `Profile`.
- Build `Required`, `AutoGen`, `Declared`, and `Rules`.
- Sort key lists for deterministic output.
- Reject manifests with no profiles.
- Return contextual errors for read, format, and structure problems.

## Key types

### TOML decoding types

```go
type tomlManifest struct {
    Project  ProjectInfo            `toml:"project"`
    Profiles map[string]tomlProfile `toml:"profiles"`
}

type tomlProfile map[string]SecretRule
```

A profile is a map from secret name to rule. This matches the TOML form:

```toml
[profiles.default]
DB_PASS = { generate = true }
DB_URL  = { required = true }
```

### `SecretRule`

```go
type SecretRule struct {
    Generate bool `toml:"generate"`
    Required bool `toml:"required"`
}
```

A key may set neither, one, or both flags. The parser records the rule as-is;
the commands decide whether the combination is valid.

### `ProjectInfo`

```go
type ProjectInfo struct {
    Name     string `toml:"name"`
    Provider string `toml:"provider"`
}
```

### `Profile`

| Field | Content |
| :--- | :--- |
| `Name` | Profile name from the TOML table |
| `Required` | Sorted keys with `required = true` |
| `AutoGen` | Sorted keys with `generate = true` |
| `Declared` | Sorted keys mentioned in the profile, regardless of rule |
| `Rules` | Raw map of key → `SecretRule` |

`Declared` and `Rules` exist for strict validation: the parser must remember keys
whose rule has no action, which would otherwise disappear from `Required` and
`AutoGen`.

### `Manifest`

```go
type Manifest struct {
    Project  ProjectInfo
    Profiles map[string]*Profile
}
```

### `ParserModule`

```go
type ParserModule struct {
    AppConfig config.AppConfig
}
```

The only input is the manifest path.

## Execution flow

1. `os.ReadFile(AppConfig.SpecFile)`.
2. `toml.Decode` into `tomlManifest`.
3. Create an empty `Manifest`.
4. For each profile:
   - create a `Profile` with an allocated `Rules` map;
   - for each secret/rule pair:
     - append the name to `Declared`;
     - store the rule in `Rules`;
     - append to `AutoGen` when `Generate`;
     - append to `Required` when `Required`;
   - sort `Declared`, `AutoGen`, and `Required`;
   - store the profile.
5. If no profiles were decoded, return
   `no profiles defined inside 'profiles'`.

## Error messages

| Condition | Error |
| :--- | :--- |
| File read failure | `failed to read file: ...` |
| Invalid TOML | `invalid TOML format: ...` |
| No profiles | `no profiles defined inside 'profiles'` |

## Dependencies

- `github.com/BurntSushi/toml`.
- `internal/config` for `SpecFile`.

## Edge cases and notes

- A profile with no keys is valid at parse time; `check` and `doctor` flag it.
- A key with both flags appears in both `Required` and `AutoGen`; shared helpers
  de-duplicate it, and strict checks warn about the combination.
- The parser does not validate the provider name. `check --strict` and `doctor`
  do, while `provider.GetStorageProvider` falls back to `dotfiles`.
- Duplicate TOML keys are rejected by the TOML decoder before this code sees
  them.
- TOML is case-sensitive, so `API_KEY` and `api_key` are different secrets.
- The project name is not validated here.
- Sorted lists make command output stable and improve test predictability.

## Related files

- [`internal/manifest/helper.md`](helper.md) — sorted profile names.
- [`internal/manifest/template.md`](template.md) — generates the manifest.
- [`cmd/check.md`](../../cmd/check.md) — consumes `Declared` and `Rules`.
- [`cmd/shared.md`](../../cmd/shared.md) — de-duplicates `Required`/`AutoGen`.
