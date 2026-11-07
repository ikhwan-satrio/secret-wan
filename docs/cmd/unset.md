# `cmd/unset.go`

> Source: `cmd/unset.go` · Package: `cmd`

## Purpose

Implements `secwan unset <KEY>` (aliases `remove` and `rm`). It deletes one key
from the stored secret map and saves the result.

## Responsibilities

- Require exactly one non-empty key.
- Require an existing master key and storage.
- Fail when the key is not stored.
- Warn when the key is still declared in one or more profiles.
- Resolve all configured recipients and refuse accidental self-lockout.
- Delete the key and persist the updated map encrypted for every recipient.

## Execution flow

1. `cobra.ExactArgs(1)` validates the argument count.
2. Trim the key; fail if empty.
3. Build a session, load the project, and require an existing identity.
4. Load secrets with `requireSecrets`.
5. If the key is absent, fail with "Secret not found".
6. Print the header.
7. If any profile still declares the key, print a warning to stderr listing
   those profiles.
8. Delete the key from the map.
9. Save the whole map through `saveSecrets(stored, false)`, which encrypts to
   every configured recipient.
10. Print `removed <KEY>`.

## Key code

### Removal is global

`unset` has no `--profile` flag. Secrets are stored as one map per project, so
removing a key removes it for every profile that references it. This is why the
command warns when the key is still declared:

```go
declaredIn := session.Manifest.ProfilesContaining(key)
if len(declaredIn) > 0 {
    warn(fmt.Sprintf("%s is still declared in profile(s) %s.",
        key, strings.Join(declaredIn, ", ")))
}
```

The warning is informational; the removal proceeds.

### Always saves

Unlike `set`, `unset` does not track a `changed` flag. It only reaches the save
step after confirming the key existed, so a save is always meaningful.

## Dependencies

- `internal/ui` for output.
- Shared session helpers: `newSession`, `requireProject`,
  `requireExistingIdentity`, `requireSecrets`, `saveSecrets`, and `warn`.
- [`internal/manifest/profile.md`](../internal/manifest/profile.md) —
  `ProfilesContaining`.
- `github.com/spf13/cobra`.

## Edge cases and notes

- A missing master key or storage is a setup error, not a "not found" result for
  the key.
- The key lookup is case-sensitive and exact; no profile is consulted.
- Removing the last key leaves an encrypted empty JSON object (`{}`) in storage.
  The file/keyring entry is not deleted.
- If the manifest still declares the key, `check` will report it as missing
  until it is set again or removed from the manifest.
- The rewritten payload is encrypted to every configured recipient, preserving
  team access after the removal.
- If the recipient file excludes the current key, `unset` fails before writing.
- Aliases make `secwan rm KEY` and `secwan remove KEY` equivalent to
  `secwan unset KEY`.

## Related files

- [`cmd/set.md`](set.md) — create or update values.
- [`cmd/check.md`](check.md) — reports keys that are still declared but missing.
- [`cmd/shared.md`](shared.md) — warning and error helpers.
- [`internal/manifest/profile.md`](../internal/manifest/profile.md) —
  `ProfilesContaining`.
- [`internal/recipients/recipients.md`](../internal/recipients/recipients.md) —
  recipient list semantics.
