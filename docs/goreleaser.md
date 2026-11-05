# `.goreleaser.yaml`

> Source: `.goreleaser.yaml` · GoReleaser v2 configuration

## Purpose

Defines how release archives are built, named, checksummed, and published to
GitHub.

## Top-level settings

```yaml
version: 2
project_name: secret-wan
```

`version: 2` selects the GoReleaser configuration schema.

## Builds

```yaml
builds:
  - id: secret-wan
    main: .
    binary: secwan
```

| Setting | Value | Effect |
| :--- | :--- | :--- |
| `main` | `.` | Build the root package (`main.go`) |
| `binary` | `secwan` | Binary name inside archives |
| `env` | `CGO_ENABLED=0` | Static, portable binaries |
| `goos` | linux, darwin, windows | Target operating systems |
| `goarch` | amd64, arm64 | Target architectures |

### Linker flags

```yaml
ldflags:
  - -s -w
  - -X main.version={{.Version}}
  - -X main.commit={{.Commit}}
  - -X main.date={{.Date}}
```

- `-s -w` strips the symbol table and DWARF debug data.
- The `-X` flags inject build metadata into `main.version`, `main.commit`, and
  `main.date`.
- `-trimpath` removes local source paths, improving reproducibility and avoiding
  leaking build-machine paths.

The result is six binaries (3 OSes × 2 architectures).

## Archives

```yaml
archives:
  - id: default
    name_template: >-
      {{ .ProjectName }}_{{ title .Os }}_{{ if eq .Arch "amd64" }}x86_64 ...
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE
      - CHANGELOG.md
```

- Archive names use `secret-wan_<OS>_<arch>` with `amd64` renamed to `x86_64`
  and `386` to `i386` (the latter is not a configured target but is handled).
- Windows archives use `.zip`; other platforms use GoReleaser's default
  `.tar.gz`.
- Each archive includes `README.md`, `LICENSE`, and `CHANGELOG.md`.

## Checksums

```yaml
checksum:
  name_template: 'checksums.txt'
  algorithm: sha256
```

A single `checksums.txt` covers all archives. The Windows-exe job later appends
the standalone `.exe` hashes.

## Changelog

```yaml
changelog:
  sort: asc
  use: github
  filters:
    exclude: ['^docs:', '^test:', '^chore:', 'merge conflict', 'Merge pull request', 'Merge branch']
  groups:
    - title: '🚀 Features'      # feat
    - title: '🐛 Bug Fixes'     # fix
    - title: '🔧 Enhancements'  # refactor
    - title: '📦 Dependencies'  # deps
    - title: 'Other Work'       # everything else
```

- Commits are sorted ascending.
- GitHub metadata is used for links/authors.
- Documentation, test, and chore commits are excluded from the generated
  section.
- Merge commits are excluded.
- Entries are grouped by conventional-commit prefix.

Note: Release Please writes the primary changelog into the release body;
GoReleaser is configured to keep that existing body.

## Release

```yaml
release:
  github:
    owner: ikhwan-satrio
    name: secret-wan
  mode: keep-existing
  draft: false
  prerelease: auto
  name_template: "🚀 {{.ProjectName}} v{{.Version}}"
```

| Setting | Effect |
| :--- | :--- |
| `owner` / `name` | Target GitHub repository |
| `mode: keep-existing` | Preserve the Release Please changelog body |
| `draft: false` | Publish immediately |
| `prerelease: auto` | Mark pre-release versions automatically |
| `name_template` | Release title, e.g. `🚀 secret-wan v1.2.0` |

## Dependencies

- Invoked by [`workflows/release-build.md`](workflows/release-build.md).
- Produces assets consumed by users and by the Windows-exe job.
- Version metadata is displayed by [`cmd/version.md`](cmd/version.md).

## Edge cases and notes

- `CGO_ENABLED=0` means no cgo-based keyring backend can be linked; the keyring
  library uses pure-Go/platform-specific implementations.
- Archives include the MIT-licensed `LICENSE` file.
- The changelog configuration affects GoReleaser's generated notes, but the
  release body is kept from Release Please.
- The Windows-exe job builds additional assets outside this configuration.
- A local snapshot can be produced with `goreleaser release --snapshot --clean`.

## Related files

- [`workflows/release-build.md`](workflows/release-build.md)
- [`release-please-config.md`](release-please-config.md)
- [`main.md`](main.md) — linker targets.
