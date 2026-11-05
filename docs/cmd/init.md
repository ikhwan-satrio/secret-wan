# `cmd/init.go`

> Source: `cmd/init.go` · Package: `cmd`

## Purpose

Implements `secwan init`. It bootstraps a project by letting the user choose a
storage provider, creating the age master key when necessary, and writing a
starter `secret-wan.toml` manifest.

## Responsibilities

- Show a Bubble Tea selector for `dotfiles` or `keyring`.
- Handle cancellation and TUI errors gracefully.
- Load an existing master key or generate a new one.
- Generate the provider-specific manifest template.
- Create `secret-wan.toml` only when it does not already exist.
- Print the next steps for a new user.

## Execution flow

1. `config.UseApp()` builds the application configuration.
2. A header is printed with the shared UI theme.
3. `tui.SelectModel` asks the user to choose a storage provider.
4. On cancellation, the command prints "Initialization cancelled." and returns
   successfully.
5. The selected label is mapped to a provider name:
   - labels beginning with `keyring` → `keyring`
   - everything else → `dotfiles`
6. `cryptoModule.LoadOrGenerateKey()` ensures the master key exists.
7. `manifest.GetTemplate(providerName)` returns TOML text for the chosen
   provider.
8. `os.Stat` checks whether the manifest already exists:
   - present → print an informational "already exists, skipped creation" message
   - absent → write it with mode `0644`
9. The command prints the next steps (`sync`, `check`, `exec`).

## Key code

### Provider selection

```go
providerName := "dotfiles"
if strings.HasPrefix(selModel.Result, "keyring") {
    providerName = "keyring"
}
```

The selector labels are human-readable, so the command maps them to the internal
provider names accepted by the manifest.

### Manifest creation guard

```go
if _, err := os.Stat(appConfig.SpecFile); err == nil {
    // skip creation
} else {
    os.WriteFile(appConfig.SpecFile, []byte(specTemplate), 0644)
}
```

`init` is intentionally idempotent for the manifest: it will never overwrite an
existing configuration. It is less guarded for the key, because
`LoadOrGenerateKey` reuses any existing key.

## Dependencies

- `internal/config` for the spec file name and app settings.
- `internal/crypto` for master-key creation.
- `internal/manifest` for the template.
- `internal/tui` for the selector.
- `internal/ui` for output.
- `github.com/charmbracelet/bubbletea` to run the TUI.
- `github.com/spf13/cobra` for the command definition.

## Edge cases and notes

- If `secret-wan.toml` exists, the template is not written, but the "Created
  master key and ..." success line is also not printed. The command still prints
  the next steps.
- The manifest mode is `0644`; the master key itself is created with `0600` by
  the crypto module.
- The command does not create `env.enc`; storage appears on the first
  `sync` or `set` that saves data.
- Choosing `keyring` does not verify that a keyring backend is available. Use
  `secwan doctor` to check that.
- The template uses the current directory name as `project.name`; see
  [`internal/manifest/template.md`](../internal/manifest/template.md).
- The default profile template includes one generated key (`DB_PASS`) and one
  required key (`DB_URL`).

## Related files

- [`cmd/sync.md`](sync.md) — first command that fills the template.
- [`cmd/doctor.md`](doctor.md) — verifies storage and key health.
- [`internal/manifest/template.md`](../internal/manifest/template.md)
- [`internal/tui/select.md`](../internal/tui/select.md)
