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
| CLI commands | [`cmd/`](#cli-layer-cmd) | Cobra command definitions, session wrapper, and output helpers |
| Application service | [`internal/app/`](#application-layer-internalapp) | Project loading, persistence, recipient policy, rekey, and rollback |
| Domain logic | [`internal/`](#internal-packages) | Config, crypto, manifest, recipients, providers, prompts, TUI, UI |
| Team sharing | [`recipients-file.md`](recipients-file.md) | Public recipient list format and workflows |
| Automation | [`workflows/`](#github-workflows) | CI and release workflows |
| Build config | [`goreleaser.md`](goreleaser.md) | GoReleaser build, archives, changelog, release settings |
| Release config | [`release-please-config.md`](release-please-config.md) | Release Please versioning and changelog settings |
| Ancillary files | [Ancillary files](#ancillary-files) | Generated files, build artifacts, license, README |

---

## Architecture overview

`secret-wan` is a Go CLI built with Cobra, Bubble Tea, age, and TOML. The code
is split into five layers:

1. **Process entry** — [`main.go`](main.md) injects version metadata and calls
   `cmd.Execute()`.
2. **CLI layer** — the `cmd` package defines one Cobra command per file. Commands
   validate flags, call the application service, and render results.
3. **Application layer** — `internal/app` owns one CLI operation's state and
   business rules: project loading, secret persistence, recipient policy,
   rekey, rollback, and bootstrap. It returns typed user-facing errors.
4. **Domain layer** — `internal/config`, `internal/crypto`,
   `internal/manifest`, `internal/recipients`, `internal/provider`, and
   `internal/secrets` implement configuration, key handling, manifest parsing,
   recipient management, encrypted storage, and input.
5. **Presentation layer** — `internal/ui` holds the shared color theme and
   formatting helpers; `internal/tui` holds the Bubble Tea models used by
   `init` and `sync`.

### Core data flow

```text
secret-wan.toml                  secret-wan.recipients
      │  manifest.ParserModule          │  recipients.Store.Load
      ▼                                 ▼
  Manifest + Profile            []age.Recipient
      │                                 │
      └──────────── internal/app ───────┘
                       │
       ├── provider.GetStorageProvider ──┤
       │   dotfiles / keyring / docker   │
       │   env / s3                      │
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
encrypted payload is either the configured encrypted file (default `env.enc`) in
the project root, an OS keyring entry, a mounted Docker secret, a base64
`SECWAN_ENV_ENC` environment variable, or one object in an S3-compatible bucket.
The optional `secret-wan.recipients` sidecar lists the public keys that new
writes are encrypted to; without it, the current key remains the only recipient.

---

## CLI layer (`cmd/`)

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `cmd/root.go` | [`cmd/root.md`](cmd/root.md) | Root Cobra command and process exit handling |
| `cmd/session.go` | [`cmd/session.md`](cmd/session.md) | CLI wrapper around `app.Service` with styled failure helpers |
| `cmd/shared.go` | [`cmd/shared.md`](cmd/shared.md) | Error, warning, and header output helpers |
| `cmd/init.go` | [`cmd/init.md`](cmd/init.md) | Choose provider, create master key, manifest template, and recipient list |
| `cmd/sync.go` | [`cmd/sync.md`](cmd/sync.md) | Generate and prompt for secrets, then save them |
| `cmd/exec.go` | [`cmd/exec.md`](cmd/exec.md) | Run a child process with profile secrets in its environment |
| `cmd/check.go` | [`cmd/check.md`](cmd/check.md) | Report missing secrets and optional strict inconsistencies |
| `cmd/doctor.go` | [`cmd/doctor.md`](cmd/doctor.md) | Register the `doctor` command |
| `cmd/doctor_checks.go` | [`cmd/doctor_checks.md`](cmd/doctor_checks.md) | Environment, manifest, key, recipient, storage, and profile checks |
| `cmd/doctor_report.go` | [`cmd/doctor_report.md`](cmd/doctor_report.md) | Report model, rendering, summary, and exit status |
| `cmd/list.go` | [`cmd/list.md`](cmd/list.md) | List declared or all stored keys with masked values |
| `cmd/get.go` | [`cmd/get.md`](cmd/get.md) | Print one secret, masked by default |
| `cmd/set.go` | [`cmd/set.md`](cmd/set.md) | Create or update one secret |
| `cmd/unset.go` | [`cmd/unset.md`](cmd/unset.md) | Remove one secret from storage |
| `cmd/recipient.go` | [`cmd/recipient.md`](cmd/recipient.md) | Register the recipient family and render shared results |
| `cmd/recipient_list.go` | [`cmd/recipient_list.md`](cmd/recipient_list.md) | List configured recipients and mark the current key |
| `cmd/recipient_me.go` | [`cmd/recipient_me.md`](cmd/recipient_me.md) | Print the current public recipient |
| `cmd/recipient_add.go` | [`cmd/recipient_add.md`](cmd/recipient_add.md) | Add a recipient and rekey |
| `cmd/recipient_remove.go` | [`cmd/recipient_remove.md`](cmd/recipient_remove.md) | Remove a recipient and rekey |
| `cmd/rekey.go` | [`cmd/rekey.md`](cmd/rekey.md) | Re-encrypt secrets for the configured recipients |
| `cmd/printEnv.go` | [`cmd/printEnv.md`](cmd/printEnv.md) | Emit shell-safe or JSON environment output |
| `cmd/version.go` | [`cmd/version.md`](cmd/version.md) | Build metadata and the `version` command |
| `cmd/shared_test.go` | [`cmd/shared_test.md`](cmd/shared_test.md) | Unit tests for quoting, strict checks, and secret types |

---

## Application layer (`internal/app`)

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/app/service.go` | [`internal/app/service.md`](internal/app/service.md) | Service state and constructors |
| `internal/app/project.go` | [`internal/app/project.md`](internal/app/project.md) | Project, identity, and profile loading |
| `internal/app/secrets.go` | [`internal/app/secrets.md`](internal/app/secrets.md) | Secret loading, saving, missing-store policy, generation |
| `internal/app/recipients.go` | [`internal/app/recipients.md`](internal/app/recipients.md) | Recipient policy, rekey, add/remove, rollback |
| `internal/app/init.go` | [`internal/app/init.md`](internal/app/init.md) | Manifest and recipient bootstrap for `init` |
| `internal/app/errors.go` | [`internal/app/errors.md`](internal/app/errors.md) | `UserError` and `RecipientChangeError` |
| `internal/app/recipients_test.go` | [`internal/app/recipients_test.md`](internal/app/recipients_test.md) | Recipient policy, rekey, and rollback tests |

---

## Internal packages

### `internal/config`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/config/config.go` | [`internal/config/config.md`](internal/config/config.md) | Application paths and default settings |

### `internal/crypto`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/crypto/crypto.go` | [`internal/crypto/crypto.md`](internal/crypto/crypto.md) | Package type and responsibility split |
| `internal/crypto/key.go` | [`internal/crypto/key.md`](internal/crypto/key.md) | Key paths, identity loading and generation |
| `internal/crypto/password.go` | [`internal/crypto/password.md`](internal/crypto/password.md) | Cryptographically secure random values |
| `internal/crypto/secrets.go` | [`internal/crypto/secrets.md`](internal/crypto/secrets.md) | Encrypt/decrypt the secret map |
| `internal/crypto/crypto_test.go` | [`internal/crypto/crypto_test.md`](internal/crypto/crypto_test.md) | Tests for encryption, recipients, and passwords |

### `internal/manifest`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/manifest/parser.go` | [`internal/manifest/parser.md`](internal/manifest/parser.md) | Parse `secret-wan.toml` into typed profiles and rules |
| `internal/manifest/helper.go` | [`internal/manifest/helper.md`](internal/manifest/helper.md) | Sorted profile-name helper |
| `internal/manifest/profile.go` | [`internal/manifest/profile.md`](internal/manifest/profile.md) | Profile and manifest key queries |
| `internal/manifest/template.go` | [`internal/manifest/template.md`](internal/manifest/template.md) | Generate the initial manifest template |
| `internal/manifest/profile_test.go` | [`internal/manifest/profile_test.md`](internal/manifest/profile_test.md) | Tests for profile and manifest queries |

### `internal/osutil`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/osutil/os.go` | [`internal/osutil/os.md`](internal/osutil/os.md) | OS detection and config/Docker path helpers |
| `internal/osutil/os_test.go` | [`internal/osutil/os_test.md`](internal/osutil/os_test.md) | Tests for OS helpers |

### `internal/provider`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/provider/provider.go` | [`internal/provider/provider.md`](internal/provider/provider.md) | Interface, `ErrNotFound`, and package overview |
| `internal/provider/registry.go` | [`internal/provider/registry.md`](internal/provider/registry.md) | Provider metadata, factory, normalization, fallback |
| `internal/provider/dotfiles.go` | [`internal/provider/dotfiles.md`](internal/provider/dotfiles.md) | Local encrypted file storage |
| `internal/provider/keyring.go` | [`internal/provider/keyring.md`](internal/provider/keyring.md) | OS keyring storage |
| `internal/provider/docker.go` | [`internal/provider/docker.md`](internal/provider/docker.md) | Mounted Docker secret (read-only) |
| `internal/provider/env.go` | [`internal/provider/env.md`](internal/provider/env.md) | Base64 environment variable (read-only) |
| `internal/provider/s3.go` | [`internal/provider/s3.md`](internal/provider/s3.md) | S3-compatible object storage |
| `internal/provider/env_test.go` | [`internal/provider/env_test.md`](internal/provider/env_test.md) | Environment provider tests |
| `internal/provider/registry_test.go` | [`internal/provider/registry_test.md`](internal/provider/registry_test.md) | Registry tests |
| `internal/provider/s3_test.go` | [`internal/provider/s3_test.md`](internal/provider/s3_test.md) | S3 provider tests |

### `internal/recipients`

| Source file | Documentation | Responsibility |
| :--- | :--- | :--- |
| `internal/recipients/recipients.go` | [`internal/recipients/recipients.md`](internal/recipients/recipients.md) | Recipient file existence, load, and save |
| `internal/recipients/parse.go` | [`internal/recipients/parse.md`](internal/recipients/parse.md) | Line parsing, canonicalization, age conversion |
| `internal/recipients/keys.go` | [`internal/recipients/keys.md`](internal/recipients/keys.md) | Membership and removal helpers |
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
