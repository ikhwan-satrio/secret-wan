# `cmd/exec.go`

> Source: `cmd/exec.go` · Package: `cmd`

## Purpose

Implements `secwan exec -- <command> [args...]`. It injects a profile's secrets
into the environment of a child process without writing them to shell history.

## Responsibilities

- Require at least one argument (the command to run).
- Read the `--profile` flag.
- Load or create the master key.
- Parse the manifest and resolve the requested profile.
- Load and decrypt stored secrets.
- Filter secrets to keys declared by the selected profile.
- Add the filtered values to the current process environment.
- Run the child process with inherited stdin, stdout, stderr, and environment.
- Propagate the child's exit code.

## Execution flow

1. `cobra.MinimumNArgs(1)` ensures a command is present.
2. The profile defaults to `default`.
3. The master key is loaded or generated.
4. The manifest is parsed; an unknown profile prints the available profiles and
   exits.
5. Secrets are loaded. A missing store suggests
   `secwan sync -p <profile>`; other errors report decryption failure.
6. The allowed key set is the union of the profile's `Required` and `AutoGen`
   keys.
7. Only stored keys in that set are copied into `filteredSecrets`.
8. `os.Setenv` writes each value into the current process.
9. The command line is printed (the values are not).
10. `exec.Command` runs the target with `os.Environ()`, which now includes the
    injected variables.
11. If the child exits non-zero, its exit code is reused.

## Key code

### Filtering

```go
allowedKeys := make(map[string]bool)
for _, k := range targetProfile.Required { allowedKeys[k] = true }
for _, k := range targetProfile.AutoGen  { allowedKeys[k] = true }

filteredSecrets := make(map[string]string)
for k, v := range secrets {
    if allowedKeys[k] { filteredSecrets[k] = v }
}
```

Stored secrets that belong to other profiles or are unmanaged are not exposed to
the child process.

### Exit code propagation

```go
if exitError, ok := err.(*exec.ExitError); ok {
    os.Exit(exitError.ExitCode())
}
```

A failed command such as `sh -c 'exit 7'` makes `secwan exec` exit with `7`.
Other execution failures (for example, binary not found) exit with `1`.

### Argument forwarding

Cobra treats everything after the first non-flag argument as the command and its
arguments. The examples use `secwan exec -- npm start` to make the boundary
explicit.

## Dependencies

- `internal/config`, `internal/crypto`, `internal/manifest`,
  `internal/provider`, `internal/ui`.
- `os/exec` for the child process.
- `github.com/spf13/cobra` for argument validation and the `--profile` flag.

## Edge cases and notes

- `exec` calls `LoadOrGenerateKey`, so running it before `init` can create a key.
  Use `doctor` if you want to verify setup first.
- Secrets are injected into the **current** process environment before the child
  starts. The parent CLI exits shortly afterward, so this does not leak into the
  user's shell.
- The command line is printed to stdout; values are never printed.
- Values containing newlines or shell metacharacters are passed through the
  environment verbatim, not through a shell.
- A profile with no keys runs the child without injecting anything.
- `os.Environ()` includes the parent environment, so existing variables can be
  overridden by secrets with the same names.
- On Windows, environment variable names are case-insensitive at the OS level;
  the map and the manifest are case-sensitive on the Go side.

## Related files

- [`cmd/printEnv.md`](printEnv.md) — exports values instead of running a child.
- [`cmd/set.md`](set.md) — create values before running.
- [`internal/provider/provider.md`](../internal/provider/provider.md)
