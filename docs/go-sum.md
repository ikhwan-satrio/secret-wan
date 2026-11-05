# `go.sum`

> Source: `go.sum` · Generated dependency checksums

## Purpose

Records cryptographic checksums for module files and module metadata. Go uses it
to verify that downloaded dependencies match the versions recorded when the
module graph was resolved.

## Responsibilities

- Provide reproducibility and supply-chain integrity for `go build`, `go test`,
  and CI.
- Allow the Go command to detect a modified or substituted module.

## How it is maintained

`go.sum` is generated and updated automatically. Do not edit it by hand.

Common commands that update it:

```bash
go mod tidy     # add/remove entries to match imports
go get <module> # add or upgrade a dependency
go mod download # populate the module cache and checksums
```

## Format

Each line has three fields:

```text
<module> <version> <hash>
```

Examples of the two hash forms:

- `h1:...` — hash of the module's file tree.
- `<version>/go.mod h1:...` — hash of the module's `go.mod` file.

Both forms are needed for normal verification.

## Relationship to `go.mod`

- `go.mod` declares requirements.
- `go.sum` pins the exact content hashes for those requirements and their
  transitive graph.
- Removing a dependency from `go.mod` may leave stale `go.sum` lines until
  `go mod tidy` runs.

## Edge cases and notes

- `go.sum` does not list checksums for the main module itself.
- It can contain multiple versions of the same module because the module graph
  may resolve older versions during graph loading.
- CI does not run `go mod verify` explicitly, but normal Go commands verify
  modules against `go.sum`.
- A checksum mismatch fails the build rather than silently using modified code.
- The file is intentionally not documented line by line; it is generated data.

## Related files

- [`go-mod.md`](go-mod.md) — declared requirements.
- [`workflows/ci.md`](workflows/ci.md) — enables Go module caching.
