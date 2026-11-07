# `README.md`

> Source: `README.md` · User-facing project documentation

## Purpose

The repository's primary entry point for users. It explains the product, how to
install it, how to initialize a project, the command reference, storage
providers, and the release workflow.

## Contents

### Introduction and badges

Describes `secret-wan` as a secure, interactive CLI secret manager and lists Go,
license, and release badges.

### Features

Highlights:

- interactive Bubble Tea TUI
- TOML manifests
- multiple profiles
- auto-generated secrets
- `dotfiles`, `keyring`, `docker`, `env`, and `s3` storage
- runtime injection with `exec`
- per-key management with `list`, `get`, `set`, and `unset`
- multi-recipient age encryption for team sharing
- recipient management with `recipient` and `rekey`
- diagnostics with `doctor`
- script-friendly `print-env` and `check --strict`
- age encryption

### Documentation

Links to the [`docs/`](README.md) index, which provides an architecture overview
and one guide per source, workflow, and configuration file.

### Installation

- `go install github.com/ikhwan-satrio/secret-wan@latest`
- Pre-built binaries from GitHub Releases
- Standalone Windows `.exe` assets
- Linux keyring note about `libsecret` and a Secret Service daemon

### Quick start

1. `secret-wan init`
2. Edit `secret-wan.toml`
3. `secwan sync`
4. `secwan exec -- <command>`

### Command reference

A table covering `init`, `sync`, `check`, `doctor`, `list`, `get`, `set`,
`unset`, `recipient`, `rekey`, `print-env`, `exec`, and `version`.

It also documents common flags (`-p/--profile`, `--version`) and scripting
examples:

```bash
eval "$(secwan print-env)"
secwan print-env --json
secwan print-env -p staging --json | jq -r '.API_KEY'
```

### Daily secret management

Examples for listing, revealing, setting, generating, removing, strict checking,
and diagnosing.

### Team sharing

Explains how to share a project with multi-recipient age:

- publish a public key with `secwan recipient me`;
- add a teammate with `secwan recipient add <age1...>`;
- re-encrypt after manual edits with `secwan rekey`;
- inspect and remove recipients;
- the `--force` escape hatch for intentional self-exclusion;
- commit `secret-wan.recipients` because it contains only public keys.

### Storage providers

Explains the trade-offs of:

- **Dotfiles** — age-encrypted configured file (default `env.enc`), easy to
  commit and share; recipient management is most useful here.
- **Keyring** — OS-native credential storage, harder to share, Linux
  dependencies; the recipient list still documents intended access.
- **Docker secret** — read-only payload from a mounted secret, suitable for
  containers.
- **Environment variable** — read-only base64 payload from `SECWAN_ENV_ENC`,
  suitable for CI.
- **S3** — one encrypted object in an S3-compatible bucket, configured through
  `SECWAN_S3_*` environment variables.

Only `dotfiles` and `keyring` are offered by `init`; the others are selected by
editing `project.provider`.

### Development and building

- Project structure tree, including the `internal/app` service layer.
- Build command with `CGO_ENABLED=0`, `-trimpath`, `-ldflags="-s -w"`.
- Local GoReleaser snapshot command.
- Automated release flow via Release Please.
- Conventional Commit examples.
- Manual tag release example.

### Tech stack and license

Lists Cobra, Bubble Tea, age, BurntSushi/toml, 99designs/keyring, the AWS SDK for
Go v2, and GoReleaser, then links to the license file.

## Accuracy notes

- The Go badge says `1.22+`, but `go.mod` requires `1.26.9`. The `go.mod`
  directive is authoritative.
- The README documents the behavior of the current command set, including the
  newer `doctor`, `list`, `get`, `set`, `unset`, `recipient`, `rekey`,
  `version`, `--strict`, and `--json` features.
- The project tree lists `internal/app` as the application service layer between
  `cmd` and the domain packages; see
  [`internal/app/service.md`](internal/app/service.md).
- The README documents `secret-wan.recipients` as public key material and warns
  that removing a recipient does not scrub old Git history.
- The provider descriptions match the provider package documentation, with the
  caveats that `docker` and `env` are read-only and `init` only offers
  `dotfiles` and `keyring`. See
  [`internal/provider/provider.md`](internal/provider/provider.md).
- The `LICENSE` file contains the MIT License; see
  [`license.md`](license.md).

## Maintenance

When adding a command or changing behavior:

1. Update the command reference table.
2. Add a concrete example for the new workflow.
3. Keep the scripting section accurate if stdout/stderr behavior changes.
4. Document any new generated project files alongside the command that creates
   them.
5. Update this documentation index if new source files are added.

## Related files

- [`README.md`](README.md) — documentation index.
- [`cmd/root.md`](cmd/root.md) — command registration.
- [`internal/app/service.md`](internal/app/service.md) — application service
  described in the project tree.
- [`recipients-file.md`](recipients-file.md) — the sharing file described in the
  README.
- [`goreleaser.md`](goreleaser.md) — release process.
