# `cmd/session.go`

> Source: `cmd/session.go` · Package: `cmd`

## Purpose

Wraps `app.Service` for the command layer. The wrapper keeps repeated
"load or exit with a styled error" logic in one place, so command files stay
focused on flags and rendering.

## Responsibilities

- Create one application service per command invocation.
- Load the project, identity, profile, and secrets with consistent error output.
- Translate service `UserError` values into the correct heading and detail line.
- Provide a save wrapper that reports recipient counts.
- Provide a rekey wrapper for commands that do not need custom rollback output.

## Key type

```go
type cliSession struct {
    *app.Service
}
```

Embedding makes fields such as `Identity`, `Manifest`, `Config`, and `Storage`
available directly while still routing common operations through the helpers.

## Methods

| Method | Behavior |
| :--- | :--- |
| `requireProject` | `EnsureProject`; exits with `Failed to parse manifest` on failure |
| `requireIdentity` | `EnsureIdentity`; may create a key |
| `requireExistingIdentity` | `LoadIdentity`; never creates a key |
| `requireProfile` | `Profile`; exits with `Invalid profile` plus alternatives |
| `requireSecrets` | `LoadSecrets`; exits with a setup hint when storage is missing |
| `loadSecretsAllowMissing` | `LoadSecretsAllowMissing`; warns when a new store will be created |
| `saveSecrets` | `SaveSecrets`; returns the recipient count |
| `rekey` | `Rekey`; exits on failure and returns the result |

`failApp` handles both `UserError` and plain errors:

- `UserError` → use its `Short` heading and `Detail` line;
- anything else → use the command's default heading and `err.Error()`.

## Why a wrapper instead of calling the service directly

- Error presentation is a CLI concern, so commands should not repeat it.
- The wrapper makes the read-only versus write identity distinction explicit.
- It keeps command files short enough to read as command specifications.
- Special flows such as recipient rollback can still call the service directly
  and use their own renderer.

## Edge cases and notes

- `requireProject` must run before `requireProfile`; the profile lookup also
  calls `EnsureProject`, so either order works.
- `requireIdentity` is for write commands. Read-only commands must use
  `requireExistingIdentity`.
- `loadSecretsAllowMissing` prints its warning to stderr so stdout remains clean.
- `saveSecrets` discards the count for callers that do not need it; `sync` uses
  the return value in its success message.
- The wrapper does not catch `os.Exit`; helpers terminate the process when a
  required precondition fails, matching the existing CLI behavior.

## Dependencies

- `internal/app` for the service.
- `internal/manifest` for the profile return type.
- [`shared.md`](shared.md) for `fail`, `failApp`, `warn`, and `printHeader`.

## Related files

- [`shared.md`](shared.md) — error and output helpers.
- [`../internal/app/service.md`](../internal/app/service.md) — service type.
- [`../internal/app/project.md`](../internal/app/project.md) — loading methods.
- [`../internal/app/secrets.md`](../internal/app/secrets.md) — secret methods.
