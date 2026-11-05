# `cmd/root.go`

> Source: `cmd/root.go` · Package: `cmd`

## Purpose

Defines the top-level Cobra command for `secret-wan` (binary name `secwan`).
Every subcommand is attached to this command by its own file's `init`
function.

## Responsibilities

- Define the command name, short description, long description, and examples.
- Keep Cobra's usage output quiet when a command returns an error
  (`SilenceUsage: true`).
- Provide `Execute`, the single process-wide entry point for running the CLI.
- Exit with code `1` when Cobra returns an error.

## Key code

### `rootCmd`

```go
var rootCmd = &cobra.Command{
    Use:   "secret-wan",
    Short: "Secure secret manager for your projects",
    ...
    SilenceUsage: true,
}
```

`Use` is the canonical name printed in help. The project's distributed binary is
named `secwan`, so the examples use that shorter name.

`SilenceUsage` prevents Cobra from printing the full help text every time a
runtime error occurs. Command-specific validation messages are already
formatted by the command itself.

### `Execute`

```go
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

Cobra prints the error. `Execute` only converts it into a non-zero process exit
status.

### `init`

The `init` function contains a commented-out persistent `--config` flag. It is
currently inactive; global configuration is fixed by
`internal/config.UseApp()`.

## How commands are registered

Each command file ends with an `init` function that calls
`rootCmd.AddCommand(...)`. This is why importing the `cmd` package registers
`init`, `sync`, `exec`, `check`, `doctor`, `list`, `get`, `set`, `unset`,
`print-env`, and `version` without a central registry.

## Dependencies

- `github.com/spf13/cobra` for command parsing.
- `os` for exit handling.

## Edge cases and notes

- Unknown commands and bad flags are handled by Cobra and return an error, which
  becomes exit code `1`.
- Commands that use `fail` or `os.Exit` terminate before `Execute` returns, so
  their chosen exit code wins.
- No secrets are read in this file.

## Related files

- [`cmd/shared.md`](shared.md) — the shared `fail` helper used by commands.
- [`cmd/version.md`](version.md) — configures `--version` on `rootCmd`.
- [`main.md`](../main.md) — calls `Execute`.
