# `cmd/init.go`

> Source: `cmd/init.go` · Package: `cmd`

## Purpose

Implements `secwan init`. It bootstraps a project by letting the user choose a
storage provider, creating the age master key when necessary, and asking the
application service to write a starter manifest and recipient list.

## Responsibilities

- Show a Bubble Tea selector for `dotfiles` or `keyring`.
- Handle cancellation and TUI errors gracefully.
- Load an existing master key or generate a new one.
- Delegate manifest creation to `app.Service.CreateManifest`.
- Delegate new-project recipient creation to `app.Service.CreateRecipientList`.
- Report skipped manifest creation without overwriting existing files.
- Treat recipient creation failures as warnings.
- Print the next steps for a new user.

## Execution flow

1. `newSession()` builds the service with the standard configuration.
2. A header is printed with the shared UI theme.
3. `tui.SelectModel` asks the user to choose a storage provider.
4. On cancellation, the command prints "Initialization cancelled." and returns
   successfully.
5. The selected label is mapped to a provider name:
   - labels beginning with `keyring` → `keyring`
   - everything else → `dotfiles`
6. `session.requireIdentity()` ensures the master key exists.
7. `session.CreateManifest(providerName)` writes the template only when the spec
   file is absent.
8. If a manifest already existed, the command prints an informational
   "already exists, skipped creation" message.
9. If a manifest was created, the command prints the success line and calls
   `session.CreateRecipientList()`.
10. A recipient list error is a warning; a created list prints its own success
    line.
11. The command prints the next steps (`sync`, `check`, `exec`).

## Key code

### Provider selection

```go
providerName := "dotfiles"
if strings.HasPrefix(selModel.Result, "keyring") {
    providerName = "keyring"
}
```

The selector labels are human-readable, so the command maps them to internal
provider names accepted by the manifest.

### Manifest and recipient creation

```go
created, err := session.CreateManifest(providerName)
if err != nil {
    failErr("Failed to create config file", err)
}

if created {
    createdRecipients, recipientErr := session.CreateRecipientList()
    switch {
    case recipientErr != nil:
        warn(...)
    case createdRecipients:
        // print success
    }
}
```

`CreateManifest` is idempotent for the manifest: it never overwrites existing
configuration. `CreateRecipientList` seeds the file only when it is absent.

## Edge cases and notes

- If `secret-wan.toml` exists, the template is not written, the "Created master
  key and ..." line is not printed, and the recipient list is untouched.
- If `secret-wan.recipients` already exists in a new project, it is not
  overwritten.
- Recipient-list creation failures are warnings because the project can still
  operate in single-recipient mode.
- The manifest mode is `0644`; the recipient file is also `0644`; the master key
  is created with `0600` by the crypto module.
- The command does not create the configured encrypted file (default `env.enc`);
  storage appears on the first `sync` or `set` that saves data.
- Choosing `keyring` does not verify that a backend is available. Use
  `secwan doctor` to check that.
- `init` offers only `dotfiles` and `keyring`. To use `docker`, `env`, or `s3`,
  set `project.provider` in `secret-wan.toml` after initialization; S3 also needs
  the `SECWAN_S3_*` environment variables.
- The template uses the current directory name as `project.name`; see
  [`../internal/manifest/template.md`](../internal/manifest/template.md).

## Dependencies

- [`session.md`](session.md) and [`shared.md`](shared.md).
- [`../internal/app/init.md`](../internal/app/init.md) — manifest and recipient
  bootstrap.
- `internal/tui` for the selector.
- `internal/ui` for output.
- `github.com/charmbracelet/bubbletea` to run the TUI.
- `github.com/spf13/cobra` for the command definition.

## Related files

- [`sync.md`](sync.md) — first command that fills the template.
- [`doctor.md`](doctor.md) — verifies storage, key, and recipient health.
- [`recipient.md`](recipient.md) — adds teammates after initialization.
- [`../internal/manifest/template.md`](../internal/manifest/template.md)
- [`../internal/recipients/recipients.md`](../internal/recipients/recipients.md)
- [`../internal/tui/select.md`](../internal/tui/select.md)
