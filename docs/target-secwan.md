# `target/secwan`

> Source: `target/secwan` · Local build artifact

## Purpose

`target/secwan` is a locally compiled binary of the CLI. It is not source code
and is not tracked by Git.

## Why it exists

Developers build to `target/` for quick local testing, for example:

```bash
go build -o target/secwan .
```

The exact build command used to produce the current file is not recorded, so its
version metadata may be `dev` or a locally injected value. Run
`target/secwan version` to inspect it.

## Git status

`.gitignore` contains `/target`, so the entire directory is ignored:

```gitignore
/target
```

This prevents a large executable (currently several megabytes) from entering
repository history.

## How to rebuild

```bash
# Simple debug build
go build -o target/secwan .

# Smaller release-like build with version metadata
go build -trimpath -ldflags="-s -w \
  -X main.version=1.0.0 \
  -X main.commit=$(git rev-parse --short HEAD) \
  -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o target/secwan .
```

## Edge cases and notes

- The file is platform-specific. A Linux binary will not run on macOS or
  Windows.
- Never commit build artifacts; use GitHub Releases for distribution.
- Running the binary from `target/` still reads `secret-wan.toml` and the
  configured encrypted file (default `env.enc`) from the **current working
  directory**, not from the binary's directory.
- The binary is safe to delete and recreate.

## Related files

- [`gitignore.md`](gitignore.md) — ignore rule.
- [`main.md`](main.md) — entry point and version variables.
- [`goreleaser.md`](goreleaser.md) — production build configuration.
