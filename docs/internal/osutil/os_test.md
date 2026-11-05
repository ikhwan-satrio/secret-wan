# `internal/osutil/os_test.go`

> Source: `internal/osutil/os_test.go` · Package: `osutil`

## Purpose

Unit tests for the OS utility helpers. They are designed to pass on any platform
by computing the expected value from the current OS.

## Responsibilities

- Verify `GetByOS` returns native values and falls back correctly.
- Verify `ConfigPath` ends with the requested file and app directory.
- Verify validation rejects empty inputs.
- Verify `DockerSecretsDir` matches the platform expectation.

## Tests

### `TestGetByOS`

Two subtests:

1. A map containing the current OS returns `"native"`.
2. A map containing only an unknown OS returns `"fallback"`.

### `TestConfigPath`

Calls `ConfigPath("secret-wan", "keys.txt")` and asserts:

- the base name is `keys.txt`
- the parent directory's base name is `secret-wan`

This checks path construction without depending on an absolute home directory.

### `TestConfigPathValidation`

- Empty app name must return an error.
- Empty file name must return an error.

### `TestDockerSecretsDir`

Computes the expected path:

```go
want := "/run/secrets"
if IsWindows() {
    want = `C:\ProgramData\Docker\secrets`
}
```

Then compares it with `DockerSecretsDir()`.

## Dependencies

- Standard library `path/filepath`, `testing`.
- The package under test itself.

## Edge cases and notes

- The tests do not create files or directories.
- `ConfigPath` may consult environment variables such as `XDG_CONFIG_HOME`, but
  the assertions only check the final path components.
- Because expectations are derived from `CurrentOS()`, the same test file works
  on Linux, macOS, and Windows CI runners.

## Related files

- [`internal/osutil/os.md`](os.md)
