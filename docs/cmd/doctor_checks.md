# `cmd/doctor_checks.go`

> Source: `cmd/doctor_checks.go` · Package: `cmd`

## Purpose

Implements the diagnostic checks behind `secwan doctor`. It inspects the
environment, manifest, master key, recipient list, storage provider, and profile
consistency, then fills a `doctorReport`.

## Responsibilities

- Report platform, runtime, and version.
- Validate the manifest and project metadata.
- Inspect key location, directory permissions, file type, permissions, and
  parseability.
- Parse the public recipient list and check whether the current key is included.
- Ask the storage provider to verify availability.
- Decrypt stored secrets when possible.
- Check profiles for missing values and rule problems.
- Report unmanaged and empty stored secrets.
- Finish the report, which exits `1` on failures.

## Execution flow

### 1. Environment

Reports `GOOS/GOARCH`, Go runtime version, and CLI version with commit and build
date.

### 2. Manifest

Parses `secret-wan.toml`.

- Parse failure is a hard failure; the command finishes immediately because the
  remaining checks depend on the manifest.
- Otherwise checks the file, project name, provider name, and profile count.
- An empty provider is displayed as `dotfiles`; unknown provider names fail.
- Known read-only providers (`docker`, `env`) also report a `Provider mode:
  read-only` line.

### 3. Master key

`crypto.KeyPath(appName)` computes the expected path without creating
directories.

- On non-Windows, a config directory writable by group/others warns
  (recommended `0700`).
- Missing file fails with a hint to run `init` or `set`.
- A directory instead of a file fails.
- On non-Windows, file permissions with group/other bits warn (recommended
  `0600`).
- `readDoctorIdentity` opens the file, parses age identities, rejects empty
  files, rejects non-X25519 keys, and reports the public recipient.

### 4. Recipients

Loads `secret-wan.recipients` through `recipients.Store.Load`:

| Check | Level | Condition |
| :--- | :--- | :--- |
| File | failure | The file exists but cannot be parsed or read |
| Recipients | ok | No recipients configured: single-recipient mode |
| Recipients | ok | A list is present; reports the count |
| Your key | ok | The current key appears in the list |
| Your key | warning | The list exists but excludes the current key |

The "Your key" check is skipped when the master key could not be loaded.

### 5. Storage

The provider is created from the manifest, then checked through an optional
interface:

```go
if checker, ok := storage.(interface{ Check() error }); ok {
    ...
}
```

All five providers implement `Check`. A "not found" result is a warning; other
errors are failures. This distinguishes "not initialized yet" from "broken".

### 6. Secrets

Secrets are loaded only when both the identity and storage are usable.

- not found → warning `no secrets stored yet`
- decrypt error → failure
- success → report the count

### 7. Profiles and consistency

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

Expected keys use `Profile.Keys()`. Rule checks use `Profile.Rule()`. Unmanaged
keys use `Manifest.DeclaredKeys()`.

## Edge cases and notes

- `doctor` does not run `requireProject`; it calls the parser directly so it can
  render a report even when parsing fails.
- Failures exit `1`; warnings alone still exit `0`.
- The provider display name defaults to `dotfiles` when the manifest omits it,
  matching `GetStorageProvider`.
- Provider validation uses `provider.IsKnown`/`provider.Names` rather than a
  hardcoded list; read-only mode comes from `provider.IsReadOnly`.
- On Windows, permission checks are skipped because Unix mode bits are not
  meaningful.
- Key and storage checks can partially succeed: the file can be readable but fail
  to parse.
- Recipient checks are independent of storage checks.
- An empty or missing recipient file is reported as healthy single-recipient
  mode, not as an error.
- A recipient file that excludes the current key is only a warning; `doctor` does
  not rewrite or rekey anything.
- `doctor` never saves or modifies secrets.

## Dependencies

- `filippo.io/age` for identity parsing.
- `internal/config`, `internal/crypto`, `internal/manifest`,
  `internal/provider`, `internal/recipients`, `internal/ui`.
- `displayVersion` and `currentBuild` from `cmd/version.go`.
- [`doctor_report.md`](doctor_report.md) for the reporting framework.

## Related files

- [`doctor.md`](doctor.md) — command entry point.
- [`check.md`](check.md) — narrower completeness check.
- [`../internal/crypto/key.md`](../internal/crypto/key.md) — `KeyPath`.
- [`../internal/provider/provider.md`](../internal/provider/provider.md) —
  interface and `Check`.
- [`../internal/provider/registry.md`](../internal/provider/registry.md) —
  provider names and read-only metadata.
- [`../internal/manifest/profile.md`](../internal/manifest/profile.md) — key and
  rule queries.
