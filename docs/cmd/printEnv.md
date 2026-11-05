# `cmd/printEnv.go`

> Source: `cmd/printEnv.go` · Package: `cmd`

## Purpose

Implements `secwan print-env`. It writes decrypted secrets to **stdout** in a
script-friendly format while sending every header, warning, and error to
**stderr**. It also supports `--json`.

## Responsibilities

- Read `--profile` and `--json`.
- Require an existing master key (never create one).
- Load stored secrets and keep only keys declared by the profile.
- Sort keys alphabetically for deterministic output.
- Print human-facing context to stderr.
- Print shell-safe `KEY="VALUE"` assignments or a JSON object to stdout.
- Handle an empty profile without polluting stdout.

## Execution flow

1. Build a session, parse the manifest, resolve the profile.
2. Require an existing identity and load secrets.
3. Build `active` from `profileKeys(profile)` and the stored map.
4. Sort the active keys.
5. In text mode, print a comment header and a warning to stderr.
6. If there are no keys:
   - warn on stderr
   - print `{}` to stdout in JSON mode
   - print nothing else in text mode
   - return successfully
7. In JSON mode, encode the map with indentation and HTML escaping disabled.
8. In text mode, print one `KEY="escaped value"` line per key.

## Key code

### stderr/stdout separation

```go
fmt.Fprintf(os.Stderr, "%s %s\n", ...)
fmt.Fprintln(os.Stderr, ui.Warning.Render(...))
fmt.Printf("%s=%s\n", key, quoteEnvValue(active[key]))
```

The comment header starts with `#`, which is harmless if a user redirects
stderr to stdout, but keeping it on stderr is what makes
`eval "$(secwan print-env)"` work.

### `quoteEnvValue`

```go
escaped := strings.NewReplacer(
    `\`, `\\`,
    `"`, `\"`,
    `$`, `\$`,
    "`", "\\`",
).Replace(value)

return `"` + escaped + `"`
```

Every value is double-quoted and the following characters are escaped:

| Character | Escaped as | Why |
| :--- | :--- | :--- |
| `\` | `\\` | Prevent escape sequences from being reinterpreted |
| `"` | `\"` | Keep the quoted string intact |
| `$` | `\$` | Prevent variable expansion |
| `` ` `` | `` \` `` | Prevent command substitution |

This produces output that is safe for `eval` in POSIX-style shells.

### JSON output

```go
encoder := json.NewEncoder(os.Stdout)
encoder.SetEscapeHTML(false)
encoder.SetIndent("", "  ")
encoder.Encode(active)
```

`SetEscapeHTML(false)` keeps characters like `<`, `>`, and `&` readable. Keys
are encoded by `encoding/json`, which sorts map keys deterministically.

## Dependencies

- `internal/ui` for warnings and colors.
- Shared session helpers from [`cmd/shared.md`](shared.md).
- `encoding/json`, `os`, `sort`, `strings`.
- `github.com/spf13/cobra`.

## Edge cases and notes

- Values are only included when the key is declared by the selected profile.
  Unmanaged keys are excluded.
- A value that is present but empty is still printed as `KEY=""`, because the
  command filters by key presence, not by value. Use `check` to detect blanks.
- In JSON mode, the header/warning is suppressed entirely; only JSON is written
  to stdout.
- The JSON object is the raw map, not a wrapper with metadata.
- If JSON encoding fails (rare for a `map[string]string`), the command exits with
  a formatted error.
- The command uses `requireExistingIdentity`, so a missing key is an error, not
  a reason to generate one.
- On Windows, the quoting syntax targets POSIX shells. Use `--json` for
  PowerShell or other consumers.

## Related files

- [`cmd/exec.md`](exec.md) — injects values directly instead of printing them.
- [`cmd/check.md`](check.md) — detects empty or missing values.
- [`cmd/shared_test.md`](shared_test.md) — includes `quoteEnvValue` cases.
