# `main.go`

> Source: `main.go` · Package: `main`

## Purpose

`main.go` is the process entry point. It has one job: receive build-time
version metadata, pass it to the `cmd` package, and start the Cobra CLI.

## Responsibilities

- Declare the `version`, `commit`, and `date` variables that GoReleaser and the
  release workflow populate with `-ldflags -X`.
- Provide safe defaults (`dev`, `none`, `unknown`) for normal `go build` runs.
- Call `cmd.SetVersionInfo` before executing any command.
- Call `cmd.Execute`, which handles the final process exit code.

## Key code

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    cmd.SetVersionInfo(version, commit, date)
    cmd.Execute()
}
```

### Why the variables are package-level

The Go linker can only set string variables with `-ldflags -X`. The variable
must be a package-level `string`. That is why `version`, `commit`, and `date`
are declared here and not as local values.

### `SetVersionInfo`

`cmd.SetVersionInfo` normalizes empty values and configures the root command's
`--version` output. Passing the values from `main` keeps the version package
independent from linker details.

## Execution flow

1. The operating system starts the binary.
2. Go initializes the three package-level strings.
3. `main` forwards them to `cmd.SetVersionInfo`.
4. `cmd.Execute` runs the root command and its subcommands.
5. Any command error results in exit code `1`; commands that call `os.Exit`
   directly choose their own exit code.

## Dependencies

- `github.com/ikhwan-satrio/secret-wan/cmd` for CLI setup and execution.

## Edge cases and notes

- A plain `go build` produces a binary that reports `dev` as its version.
- `SetVersionInfo` treats empty strings defensively even though this file
  supplies defaults.
- Secrets are never handled directly in `main.go`.

## Related files

- [`cmd/version.md`](cmd/version.md) — version formatting and command.
- [`cmd/root.md`](cmd/root.md) — root command and exit handling.
- [`goreleaser.md`](goreleaser.md) — linker flags used for releases.
