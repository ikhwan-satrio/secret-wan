# secret-wan Documentation

This directory contains one Markdown document for every file in the repository.
Each document explains what the file is responsible for, how its code works,
which other files it depends on, and which edge cases matter.

The documents mirror the repository layout, so a file at `cmd/sync.go` is
documented at [`cmd/sync.md`](cmd/sync.md).

---

## Start here

| Area | Document | What it explains |
| :--- | :--- | :--- |
| Entry point | [`main.md`](main.md) | Process startup and build-time version injection |
| CLI commands | [`cmd/`](#cli-layer-cmd) | Cobra command definitions and shared helpers |
| Domain logic | [`internal/`](#internal-packages) | Config, crypto, manifest, recipients, providers, prompts, TUI, UI |
| Team sharing | [`recipients-file.md`](recipients-file.md) | Public recipient list format and workflows |
| Automation | [`workflows/`](#github-workflows) | CI and release workflows |
| Build config | [`goreleaser.md`](goreleaser.md) | GoReleaser build, archives, changelog, release settings |
| Release config | [`release-please-config.md`](release-please-config.md) | Release Please versioning and changelog settings |
| Ancillary files | [Ancillary files](#ancillary-files) | Generated files, build artifacts, license, README |

---

## Architecture overview

`secret-wan` is a Go CLI built with Cobra, Bubble Tea, age, and TOML. The code
is split into four layers:

1. **Process entry** — [`main.go`](main.md) injects version metadata and calls
   `cmd.Execute()`.
2. **CLI layer** — the `cmd` package defines one Cobra command per file. Commands
   validate flags and arguments, then coordinate the domain packages.
3. **Domain layer** — `internal/config`, `internal/crypto`,
   `internal/manifest`, `internal/recipients`, `internal/provider`, and
   `internal/secrets` implement configuration, key handling, manifest parsing,
   recipient management, encrypted storage, and input.
4. **Presentation layer** — `internal/ui` holds the shared color theme and
   formatting helpers; `internal/tui` holds the Bubble Tea models used by
   `init` and `sync`.

### Core data flow

```text
secret-wan.toml                  secret-wan.recipients
      │  manifest.ParserModule          │  recipients.Store.Load
      ▼                                 ▼
  Manifest + Profile            []age.Recipient
      │                                 │
      ├── provider.GetStorageProvider ──┤
      │        dotfiles / keyring / docker
      │                                 │
      └── crypto.CryptoModule.LoadSecrets
                │  age.Decrypt
                ▼
        map[string]string of secrets
                │
      ┌─────────┼──────────┬───────────┐
      ▼         ▼          ▼           ▼
    check     list/get    exec      print-env
```

The master age identity lives outside the project, normally at
`~/.config/secret-wan/keys.txt` (or the OS-specific config directory). The
encrypted payload is either `env.enc` in the project root, an OS keyring entry,
or a mounted Docker secret. The optional `secret-wan.recipients` sidecar lists
the public keys that new writes are encrypted to; without it, the current key
remains the only recipient.

---

## CLI layer (`cmd/`)

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `cmd/root.go` | [`cmd/root.md`](cmd/root.md) | Root Cobra command and process exit handling |
| `cmd/shared.go` | [`cmd/shared.md`](cmd/shared.md) | Shared session, error, recipient, save, and profile helpers |
| `cmd/init.go` | [`cmd/init.md`](cmd/init.md) | Choose provider, create master key, manifest template, and recipient list |
| `cmd/sync.go` | [`cmd/sync.md`](cmd/sync.md) | Generate and prompt for secrets, then save them |
| `cmd/exec.go` | [`cmd/exec.md`](cmd/exec.md) | Run a child process with profile secrets in its environment |
| `cmd/check.go` | [`cmd/check.md`](cmd/check.md) | Report missing secrets and optional strict inconsistencies |
| `cmd/doctor.go` | [`cmd/doctor.md`](cmd/doctor.md) | Diagnose environment, manifest, key, recipients, storage, and profiles |
| `cmd/list.go` | [`cmd/list.md`](cmd/list.md) | List declared or all stored keys with masked values |
| `cmd/get.go` | [`cmd/get.md`](cmd/get.md) | Print one secret, masked by default |
| `cmd/set.go` | [`cmd/set.md`](cmd/set.md) | Create or update one secret |
| `cmd/unset.go` | [`cmd/unset.md`](cmd/unset.md) | Remove one secret from storage |
| `cmd/recipient.go` | [`cmd/recipient.md`](cmd/recipient.md) | List, add, remove, and identify age recipients |
| `cmd/rekey.go` | [`cmd/rekey.md`](cmd/rekey.md) | Re-encrypt secrets for the configured recipients |
| `cmd/printEnv.go` | [`cmd/printEnv.md`](cmd/printEnv.md) | Emit shell-safe or JSON environment output |
| `cmd/version.go` | [`cmd/version.md`](cmd/version.md) | Build metadata and the `version` command |
| `cmd/shared_test.go` | [`cmd/shared_test.md`](cmd/shared_test.md) | Unit tests for shared helpers and strict checks |

---

## Internal packages

### `internal/config`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/config/config.go` | [`internal/config/config.md`](internal/config/config.md) | Application paths and default settings |

### `internal/crypto`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/crypto/crypto.go` | [`internal/crypto/crypto.md`](internal/crypto/crypto.md) | age key loading/generation, password generation, encrypt/decrypt |
| `internal/crypto/crypto_test.go` | [`internal/crypto/crypto_test.md`](internal/crypto/crypto_test.md) | Tests for multi-recipient encryption and password generation |

### `internal/manifest`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/manifest/parser.go` | [`internal/manifest/parser.md`](internal/manifest/parser.md) | Parse `secret-wan.toml` into typed profiles and rules |
| `internal/manifest/helper.go` | [`internal/manifest/helper.md`](internal/manifest/helper.md) | Sorted profile-name helper |
| `internal/manifest/template.go` | [`internal/manifest/template.md`](internal/manifest/template.md) | Generate the initial manifest template |

### `internal/osutil`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/osutil/os.go` | [`internal/osutil/os.md`](internal/osutil/os.md) | OS detection and config/Docker path helpers |
| `internal/osutil/os_test.go` | [`internal/osutil/os_test.md`](internal/osutil/os_test.md) | Tests for OS helpers |

### `internal/provider`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/provider/provider.go` | [`internal/provider/provider.md`](internal/provider/provider.md) | Dotfiles, keyring, and Docker storage providers |

### `internal/recipients`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/recipients/recipients.go` | [`internal/recipients/recipients.md`](internal/recipients/recipients.md) | Public recipient file parsing, validation, storage, and helpers |
| `internal/recipients/recipients_test.go` | [`internal/recipients/recipients_test.md`](internal/recipients/recipients_test.md) | Tests for recipient parsing and storage |

### `internal/secrets`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/secrets/secrets.go` | [`internal/secrets/secrets.md`](internal/secrets/secrets.md) | Random secret generation and terminal/piped prompts |

### `internal/tui`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/tui/base.go` | [`internal/tui/base.md`](internal/tui/base.md) | Secret input wizard and progress display |
| `internal/tui/select.go` | [`internal/tui/select.md`](internal/tui/select.md) | Generic single-choice selector |
| `internal/tui/style.go` | [`internal/tui/style.md`](internal/tui/style.md) | Bubble Tea style aliases and box/progress styles |

### `internal/ui`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/ui/theme.go` | [`internal/ui/theme.md`](internal/ui/theme.md) | Adaptive colors, shared styles, icons, error formatting |

---

## GitHub workflows

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `.github/workflows/ci.yaml` | [`workflows/ci.md`](workflows/ci.md) | Format, vet, test, and cross-build on pushes and PRs |
| `.github/workflows/release-build.yaml` | [`workflows/release-build.md`](workflows/release-build.md) | Reusable GoReleaser build plus Windows `.exe` assets |
| `.github/workflows/release-please.yaml` | [`workflows/release-please.md`](workflows/release-please.md) | Open release PRs and trigger release builds |
| `.github/workflows/release.yaml` | [`workflows/release.md`](workflows/release.md) | Manual tag-triggered release builds |

---

## Build and project configuration

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `go.mod` | [`go-mod.md`](go-mod.md) | Module path, Go version, direct and indirect dependencies |
| `go.sum` | [`go-sum.md`](go-sum.md) | Go dependency checksums (generated) |
| `.goreleaser.yaml` | [`goreleaser.md`](goreleaser.md) | Cross-platform release builds, archives, changelog, GitHub release |
| `release-please-config.json` | [`release-please-config.md`](release-please-config.md) | Release Please package and version settings |
| `.release-please-manifest.json` | [`release-please-manifest.md`](release-please-manifest.md) | Last released version tracked by Release Please |
| `.gitignore` | [`gitignore.md`](gitignore.md) | Ignored build output directory |

---

## Ancillary files

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `README.md` | [`project-readme.md`](project-readme.md) | User-facing overview, installation, commands, release flow |
| `CHANGELOG.md` | [`changelog.md`](changelog.md) | Generated release history |
| `LICENSE` | [`license.md`](license.md) | MIT License text and copyright notice |
| `target/secwan` | [`target-secwan.md`](target-secwan.md) | Local build artifact (ignored by Git) |

## Generated project files

These files are created inside a user's project rather than in this repository,
but their behavior is part of the product:

| Generated file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `secret-wan.recipients` | [`recipients-file.md`](recipients-file.md) | Public age recipient list for multi-recipient encryption |

---

## Documentation conventions

Each per-file document uses the same structure:

- **Purpose** — why the file exists.
- **Responsibilities** — the behavior it owns.
- **Key code** — important types, functions, or fields.
- **Execution flow** — how the code runs, when relevant.
- **Dependencies** — packages and files it relies on.
- **Edge cases and notes** — behavior that is easy to miss.
- **Related files** — nearby code worth reading next.

All documentation is written in English. User-facing CLI messages are also in
English.
