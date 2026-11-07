# `go.mod`

> Source: `go.mod` · Go module definition

## Purpose

Declares the Go module path, the required Go language/toolchain version, and all
direct and indirect dependencies.

## Module identity

```go
module github.com/ikhwan-satrio/secret-wan
```

This path is used by:

- `go install github.com/ikhwan-satrio/secret-wan@latest`
- internal imports such as `github.com/ikhwan-satrio/secret-wan/cmd`
- module-aware tooling and the Go proxy

## Go version

```go
go 1.26.9
```

The `go` directive is the minimum language/toolchain version. CI reads this value
through `go-version-file: go.mod`, so CI and local builds use the same target.

> Note: the README badge advertises Go 1.22+, which is older than the directive.
> The directive is the authoritative requirement.

## Direct dependencies

| Module | Version | Role |
| :--- | :--- | :--- |
| `filippo.io/age` | v1.3.2 | X25519 identities and file/stream encryption |
| `github.com/99designs/keyring` | v1.2.2 | OS keychain/keyring access |
| `github.com/BurntSushi/toml` | v1.6.0 | Manifest parsing |
| `github.com/aws/aws-sdk-go-v2` | v1.47.8 | AWS SDK core types |
| `github.com/aws/aws-sdk-go-v2/config` | v1.33.16 | AWS credential and region loading |
| `github.com/aws/aws-sdk-go-v2/service/s3` | v1.114.9 | S3-compatible object storage |
| `github.com/aws/smithy-go` | v1.28.8 | AWS API error types and transport |
| `github.com/charmbracelet/bubbles` | v1.0.0 | Text input component for Bubble Tea |
| `github.com/charmbracelet/bubbletea` | v1.3.10 | Interactive TUI runtime |
| `github.com/charmbracelet/lipgloss` | v1.1.0 | Terminal styling and layout |
| `github.com/spf13/cobra` | v1.10.2 | CLI command and flag framework |
| `golang.org/x/term` | v0.47.0 | Hidden password input and terminal detection |

## Indirect dependencies

The second `require` block lists modules pulled in by the direct dependencies,
including:

- age's HPKE primitives
- AWS SDK internals (`credentials`, `feature/ec2/imds`, `service/internal/*`,
  `service/sso`, `service/ssooidc`, `service/sts`, `service/signin`)
- keychain/keyring platform backends (`go-keychain`, `wincred`, `dbus`,
  `go-libsecret`, `jose2go`)
- Bubble Tea's terminal and width libraries (`go-isatty`, `termenv`,
  `go-runewidth`, `uniseg`, `cancelreader`, `coninput`)
- Cobra's `pflag` and `mousetrap`
- `golang.org/x/crypto`, `x/sys`, `x/text`
- `gopkg.in/yaml.v3` (used by tooling/dependency chains)

Indirect modules are managed by `go mod tidy`; they should not be edited by
hand.

## Dependencies by feature

| Feature | Modules |
| :--- | :--- |
| CLI | Cobra |
| TUI | Bubble Tea, Bubbles, Lip Gloss |
| Encryption | age |
| Manifest | BurntSushi/toml |
| Keyring | 99designs/keyring |
| S3 storage | AWS SDK for Go v2 (core, config, S3, smithy-go) |
| Prompting | x/term |

## Edge cases and notes

- Direct dependencies are those imported by the project's own packages.
- The module requires a very recent Go toolchain (`1.26.9`); older toolchains
  refuse to build it.
- `go.sum` must stay consistent with this file; run `go mod tidy` after changing
  dependencies.
- Cross-compilation is supported with `CGO_ENABLED=0`, matching the release
  configuration.

## Related files

- [`go-sum.md`](go-sum.md) — dependency checksums.
- [`workflows/ci.md`](workflows/ci.md) — uses `go-version-file`.
- [`goreleaser.md`](goreleaser.md) — builds with `CGO_ENABLED=0`.
