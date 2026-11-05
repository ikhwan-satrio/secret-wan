# `cmd/version.go`

> Source: `cmd/version.go` · Package: `cmd`

## Purpose

Stores build metadata and implements both the `--version` root flag and the
`version` subcommand.

## Responsibilities

- Hold the normalized build information in `currentBuild`.
- Accept linker-injected values through `SetVersionInfo`.
- Normalize missing values and ensure the displayed version has a leading `v`
  when appropriate.
- Configure Cobra's version flag and template.
- Print a multi-line version report for the `version` command.

## Key types

### `buildInfo`

```go
type buildInfo struct {
    Version string
    Commit  string
    Date    string
}
```

### `currentBuild`

Defaults to:

| Field | Default |
| :--- | :--- |
| `Version` | `dev` |
| `Commit` | `none` |
| `Date` | `unknown` |

These defaults make local builds identifiable without pretending to be a
release.

## Key functions

### `SetVersionInfo(version, commit, date string)`

1. Replaces empty strings with the defaults above.
2. Stores the values in `currentBuild`.
3. Sets `rootCmd.Version = displayVersion()`.
4. Sets the root command's version template to `versionTemplate()`.

`main.main` calls this with values injected by GoReleaser or the release
workflow link flags.

### `displayVersion`

```go
if currentBuild.Version == "dev" || len(currentBuild.Version) == 0 {
    return currentBuild.Version
}
if currentBuild.Version[0] == 'v' {
    return currentBuild.Version
}
return "v" + currentBuild.Version
```

- `dev` stays `dev`.
- `1.2.3` becomes `v1.2.3`.
- `v1.2.3` stays unchanged.

GoReleaser supplies `.Version` without a leading `v`, so the normalization
produces the conventional display form.

### `versionTemplate`

Prints:

```text
secret-wan <version>
  commit:   <commit>
  built:    <date>
  go:       <runtime.Version()>
  platform: <GOOS>/<GOARCH>
```

The template is used both by `secret-wan --version` and the `version` command,
so the output is identical.

### `versionCmd`

A Cobra command with `Use: "version"` and `Args: cobra.NoArgs`. It calls
`fmt.Print(versionTemplate())`. The template already ends with a newline.

## Edge cases and notes

- The command does not read the manifest, key, or storage, so it works even
  before the project is initialized.
- `runtime.Version()` reports the Go toolchain used to build the binary, not the
  version installed on the user's machine.
- The root command's `--version` flag is only configured when
  `SetVersionInfo` runs, which always happens in `main`.
- The `go` directive in `go.mod` and the actual build toolchain can differ; the
  displayed value reflects the build toolchain.

## Related files

- [`main.md`](../main.md) — supplies the linker-injected values.
- [`goreleaser.md`](../goreleaser.md) — defines `-X main.version`, `-X main.commit`,
  and `-X main.date`.
- [`workflows/release-build.md`](../workflows/release-build.md) — injects the same
  values for the standalone Windows executables.
