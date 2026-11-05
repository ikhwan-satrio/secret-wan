# `cmd/doctor.go`

> Source: `cmd/doctor.go` · Package: `cmd`

## Purpose

Implements `secwan doctor`, a read-oriented diagnostic command that inspects the
environment, manifest, master key, storage provider, and profile consistency in
one pass.

## Responsibilities

- Build a structured report of sections, checks, warnings, and failures.
- Never create a master key or config directory.
- Validate the manifest and project metadata.
- Inspect key location, directory permissions, file type, file permissions, and
  parseability.
- Ask the storage provider to verify availability.
- Decrypt stored secrets when possible.
- Check each profile for missing values, missing rules, and conflicting rules.
- Report unmanaged and empty stored secrets.
- Render a summary and exit `1` when there are failures.

## Report framework

### `checkLevel`

```go
const (
    checkOK checkLevel = iota
    checkWarn
    checkFail
)
```

### `doctorCheck` and `doctorSection`

A check has a level, label, and detail. A section groups related checks under a
title.

### `doctorReport`

Tracks the sections plus warning and failure counts. Helper methods:

| Method | Effect |
| :--- | :--- |
| `section(title)` | Appends and returns a new section |
| `add(section, level, label, detail)` | Appends a check and increments counters |
| `ok` / `warn` / `fail` | Level-specific wrappers |
| `render()` | Prints every non-empty section with icons and colors |
| `finish()` | Renders, prints the summary, and exits `1` on failures |

`printDoctorSummary` reports one of three outcomes:

- `Everything looks good.`
- `<n> warning(s), no errors.`
- `<n> error(s), <n> warning(s).`

## Execution flow

### 1. Environment

Reports platform (`GOOS/GOARCH`), Go runtime, and the CLI version with commit and
build date.

### 2. Manifest

Parses `secret-wan.toml`.

- Parse failure is a hard failure; the command finishes immediately because the
  remaining checks depend on the manifest.
- Otherwise checks the file, project name, provider name, and number of
  profiles.
- An empty provider is displayed as `dotfiles`; unknown provider names fail.

### 3. Master key

`crypto.KeyPath(appName)` computes the expected location without creating
directories.

- On non-Windows, a config directory writable by group/others produces a
  permission warning (recommended `0700`).
- Missing file → failure with a hint to run `init` or `set`.
- Directory instead of file → failure.
- On non-Windows, file permissions with group/other bits produce a warning
  (recommended `0600`).
- `readDoctorIdentity` opens the file, parses age identities, rejects empty
  files, rejects non-X25519 keys, and reports the public recipient.

### 4. Storage

The provider is created from the manifest, then checked through an optional
interface:

```go
if checker, ok := storage.(interface{ Check() error }); ok {
    ...
}
```

All three providers implement `Check`. A "not found" result is a warning; other
errors are failures. This lets `doctor` distinguish "not initialized yet" from
"broken".

### 5. Secrets

Secrets are loaded only when both the identity and storage are usable.

- not found → warning "no secrets stored yet"
- decrypt error → failure
- success → report the count

### 6. Profiles and consistency

`checkDoctorProfiles` covers:

| Check | Level | Condition |
| :--- | :--- | :--- |
| Profile secrets | warning | Profile has no keys |
| Profile secrets | warning | Secrets unavailable |
| Profile secrets | failure | One or more expected values missing/blank |
| Rule quality | warning | Declared key has no rule |
| Rule quality | warning | Key is both required and generated |
| Unmanaged secrets | warning | Stored keys not declared anywhere |
| Empty values | warning | Stored values that are blank |

Expected keys use the shared `profileKeys` union. Missing keys are listed in the
detail string.

## Dependencies

- `filippo.io/age` for identity parsing.
- `internal/config`, `internal/crypto`, `internal/manifest`,
  `internal/provider`, `internal/ui`.
- Shared helpers: `profileKeys`, `allDeclaredKeys`, `ruleFor`,
  `displayVersion`, `currentBuild`.
- `github.com/spf13/cobra`.

## Edge cases and notes

- `doctor` does not run `requireManifest`; it calls the parser directly so it can
  render a report even when parsing fails.
- Failures exit `1`; warnings alone still exit `0`.
- The provider display name defaults to `dotfiles` when the manifest omits it,
  matching `GetStorageProvider`.
- On Windows, permission checks are skipped because Unix mode bits are not
  meaningful.
- Key and storage checks can partially succeed: for example, the file can be
  readable but fail to parse.
- `doctor` never saves or modifies secrets.

## Related files

- [`cmd/check.md`](check.md) — a narrower completeness check.
- [`internal/crypto/crypto.md`](../internal/crypto/crypto.md) — `KeyPath`,
  `LoadSecrets`.
- [`internal/provider/provider.md`](../internal/provider/provider.md) — `Check`.
- [`cmd/shared.md`](shared.md) — profile helpers.
