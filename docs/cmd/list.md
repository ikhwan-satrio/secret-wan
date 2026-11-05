# `cmd/list.go`

> Source: `cmd/list.go` · Package: `cmd`

## Purpose

Implements `secwan list` (alias `ls`). It displays the secrets relevant to a
profile in a table, with values masked unless `--reveal` is used.

## Responsibilities

- Read `--profile`, `--reveal`, and `--all`.
- Require an existing key and storage.
- Show declared keys by default; with `--all`, also include unmanaged stored
  keys.
- Display each key's type and value status.
- Provide the shared `maskedValue` constant and secret-type classifier.

## Execution flow

1. `cobra.NoArgs` rejects positional arguments.
2. Build a session, parse the manifest, resolve the profile.
3. Require an existing identity and load secrets.
4. Choose keys:
   - `profileKeys(profile)` by default
   - `allKeys(profile, stored)` with `--all`
5. Print a header and optional "all stored secrets" note.
6. If there are no keys, print "No secrets defined for this profile." and
   return.
7. Print the `KEY TYPE VALUE` column headers.
8. Print one row per key.

## Key code

### `maskedValue`

```go
const maskedValue = "••••••••"
```

The same constant is used by `get`, so masked output is consistent.

### `allKeys`

Computes the union of the profile's required keys, generated keys, and every
stored key, then sorts the result. This is used by `--all`.

### `secretType`

Classifies a key for the `TYPE` column:

| Label | Meaning |
| :--- | :--- |
| `required` | `{ required = true }` in this profile |
| `generated` | `{ generate = true }` in this profile |
| `gen+req` | Both rules (flagged by strict checks) |
| `declared` | Mentioned in the profile with no active rule |
| `other` | Declared by a different profile |
| `unmanaged` | Not declared anywhere |

The `other` and `unmanaged` labels are relevant when `--all` reveals stored keys
outside the current profile.

### Value column

| Situation | Output |
| :--- | :--- |
| Key not in storage | `not set` (warning style) |
| `--reveal` | Raw value, with newlines rendered as `\n` |
| Default | Mask |

Rendering newlines as the literal characters `\n` keeps the table on one line
and avoids breaking terminal layout.

## Dependencies

- `internal/manifest` for profile types.
- `internal/ui` for styling.
- Shared helpers: `newSession`, `requireManifest`, `requireProfile`,
  `requireExistingIdentity`, `requireSecrets`, `profileKeys`, `keyInProfile`,
  `ruleFor`, `profilesContaining`.
- `github.com/spf13/cobra`.

## Edge cases and notes

- `list` is read-only: it uses `requireExistingIdentity`.
- Values are not printed to stderr; the table itself goes to stdout.
- A stored key with an empty value is shown as masked or as an empty value; it
  is not labeled `not set` because it exists in the map. Use `check` for blank
  detection.
- Column widths are fixed (`28` for key, `12` for type) for simple alignment;
  long keys are not truncated.
- The union in `allKeys` is sorted for deterministic output.
- `list --reveal` can expose all secrets on screen; use it carefully.

## Related files

- [`cmd/get.md`](get.md) — single-key lookup and the shared mask.
- [`cmd/shared_test.md`](shared_test.md) — `secretType` tests.
- [`cmd/check.md`](check.md) — missing/blank detection.
