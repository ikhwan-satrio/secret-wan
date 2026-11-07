# `cmd/check.go`

> Source: `cmd/check.go` · Package: `cmd`

## Purpose

Implements `secwan check` and its `--strict` mode. The command verifies that a
profile's expected secrets are present and non-empty, and optionally reports
manifest/storage inconsistencies.

## Responsibilities

- Validate the profile and require at least one declared secret.
- Load the existing master key without creating a new one.
- Load stored secrets.
- Print one status line per expected key.
- In strict mode, collect and print structural issues.
- Exit with code `1` when secrets are missing or strict issues exist.

## Execution flow

1. Read `--profile` and `--strict` flags.
2. Build a session, load the project, and resolve the profile.
3. If the profile has no keys, fail with "Profile '<name>' must define at least
   one secret".
4. Require an existing master key and load secrets.
5. Print a header and the profile name.
6. For each sorted expected key:
   - present and non-blank → `configured`
   - missing or blank → `missing`
7. If `--strict`, call `collectStrictIssues` and `printStrictIssues`.
8. Print a summary and exit `1` if anything failed.

## Key types

### `strictIssue`

```go
type strictIssue struct {
    label  string
    detail string
}
```

The label is the key or manifest field; the detail explains the problem.

## Strict checks

`collectStrictIssues(manifest, stored)` reports:

| Condition | Label | Detail |
| :--- | :--- | :--- |
| Empty project name | `project.name` | `project name is empty` |
| Unknown provider | `project.provider` | `unknown provider '<name>' (expected docker, dotfiles, env, keyring, s3)` |
| Stored key not declared in any profile | key | `stored but not declared in any profile` |
| Profile with no keys | profile name | `profile has no secrets defined` |
| Declared key with no rule | key | `profile '<name>' has no rule (use required or generate)` |
| Key with both rules | key | `profile '<name>' marks the key as both required and generate` |

Notes:

- An empty provider string is accepted because the provider registry defaults to
  `dotfiles`.
- The accepted provider names come from `provider.Names()`, so the message stays
  correct as backends are added.
- Unmanaged keys are sorted before being reported.
- The same key can produce issues in multiple profiles.
- Strict issues are reported after the normal per-key status table.

## Rendering

`printStrictIssues` prints a `Strict checks` heading and either
`No inconsistencies found.` or one warning line per issue. The final summary
distinguishes:

- all secrets present and no strict issues → success
- missing secrets → failure and a sync hint
- strict issues → failure with the issue count

The command exits `1` if either the normal check or the strict check failed.

## Dependencies

- `internal/manifest`, `internal/provider`, `internal/ui`.
- Shared session helpers from [`cmd/session.md`](session.md): `newSession`,
  `requireProject`, `requireProfile`, `requireExistingIdentity`,
  `requireSecrets`.
- [`internal/manifest/profile.md`](../internal/manifest/profile.md) — `Keys`,
  `Rule`, `DeclaredKeys`.
- [`internal/provider/registry.md`](../internal/provider/registry.md) —
  `IsKnown` and `Names`.
- `github.com/spf13/cobra`.

## Edge cases and notes

- `check` is read-only with respect to the master key: it uses
  `requireExistingIdentity`.
- A value consisting only of whitespace is treated as missing.
- Strict mode does not modify storage or the manifest.
- A missing storage is reported as a setup error rather than as individual
  missing keys.
- `required + generate` is flagged for consistency, but `sync` handles it by
  generating rather than prompting.

## Related files

- [`cmd/doctor.md`](doctor.md) — broader diagnostics with warnings and sections.
- [`cmd/shared_test.md`](shared_test.md) — tests for strict issue collection.
- [`internal/manifest/parser.md`](../internal/manifest/parser.md) — `Declared` and
  `Rules` metadata.
- [`internal/manifest/profile.md`](../internal/manifest/profile.md) — manifest
  query methods used by the checks.
