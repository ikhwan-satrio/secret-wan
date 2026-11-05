# `cmd/set.go`

> Source: `cmd/set.go` · Package: `cmd`

## Purpose

Implements `secwan set <KEY> [VALUE]`, the non-interactive path for creating or
updating one secret. It can read a value from a flag argument, prompt for it
without echo, or generate it.

## Responsibilities

- Validate the key and argument count (1 or 2).
- Reject an empty key.
- Reject `VALUE` together with `--generate`.
- Require the key to be declared in the profile unless `--force` is used.
- Create the master key and storage if needed.
- Read the value from the argument, a hidden terminal prompt, a pipe, or the
  random generator.
- Reject an empty value.
- Resolve all configured recipients and refuse accidental self-lockout.
- Save the updated secret map encrypted for every recipient.
- Report whether the key was saved or updated.

## Execution flow

1. Trim the key; fail if empty.
2. Read `--profile`, `--generate`, and `--force`.
3. If both a positional value and `--generate` are present, fail.
4. Build a session, parse the manifest, resolve the profile.
5. `requireIdentity` loads or creates the master key.
6. Check `keyInProfile`:
   - declared → continue
   - not declared and no `--force` → fail with guidance
   - not declared and `--force` → print a warning to stderr and continue
7. Print the header and profile.
8. Load secrets with `loadSecretsAllowMissing`, which tolerates a first write.
9. Resolve the value with `resolveSecretValue`.
10. Reject an empty value and suggest `unset`.
11. Store the value and call `saveSecrets(stored, false)`, which resolves every
    configured recipient and encrypts the full map to all of them.
12. Print `saved` or `updated`, plus a hint for generated values.

## Key code

### `resolveSecretValue`

```go
func resolveSecretValue(session *appSession, key string, args []string, generate bool) (string, error) {
    if generate {
        return session.Crypto.GenerateSecurePassword(session.Config.SecretLength)
    }
    if len(args) == 2 {
        return args[1], nil
    }
    module := secrets.SecretModule{AppConfig: session.Config}
    return module.PromptUserInput(key, true)
}
```

Priority:

1. `--generate` → cryptographically random value.
2. Positional `VALUE` → used verbatim.
3. Otherwise → hidden prompt (or piped stdin when not a terminal).

### Conflict detection

`set KEY VALUE --generate` is ambiguous, so the command rejects it instead of
silently choosing one source.

### First write

`loadSecretsAllowMissing` returns an empty map when no storage exists, allowing
`set` to create `env.enc` (or the keyring entry) without requiring `sync` first.

### Recipient handling

```go
if err := session.saveSecrets(stored, false); err != nil {
    failErr("Failed to save secrets", err)
}
```

`saveSecrets` loads `secret-wan.recipients`. A missing file means the current key
is the only recipient. A non-empty list without the current key is an error, so
`set` cannot silently write a payload that the current user cannot read.

### Empty values

An empty argument or an empty prompt result is rejected:

```text
Empty secret value
Use 'secwan unset KEY' to remove a secret instead.
```

This keeps "missing" and "empty" from becoming two different states.

## Dependencies

- `internal/secrets` for terminal/piped input.
- `internal/ui` for output and warnings.
- Shared helpers: `newSession`, `requireManifest`, `requireProfile`,
  `requireIdentity`, `loadSecretsAllowMissing`, `keyInProfile`,
  `profilesContaining`, `saveSecrets`.
- `internal/recipients` indirectly, through the shared save path.
- `github.com/spf13/cobra`.

## Edge cases and notes

- `set` can bootstrap the project: it creates the master key and an empty
  storage map on first use.
- With `--force`, the key is stored but remains undeclared; `check --strict` and
  `doctor` will report it as unmanaged until the manifest is updated.
- If the key is declared in another profile, the warning names those profiles,
  but the value is still stored in the shared secret map.
- A value supplied as an argument may be visible in shell history and process
  listings. Prefer the hidden prompt for real secrets.
- A generated value is printed only as a hint; use `secwan get KEY --reveal` to
  see it.
- The generated length comes from `AppConfig.SecretLength` (100 characters by
  default).
- The command saves the whole map, so other profiles and unmanaged keys are
  preserved.
- The rewritten payload is encrypted to every key in `secret-wan.recipients`,
  not just the current key.
- If the recipient file excludes the current key, `set` fails before writing and
  suggests `secwan recipient add`.
- When stdin is not a terminal, a secret prompt reads all piped input and trims
  trailing newlines; this supports automation.

## Related files

- [`cmd/unset.md`](unset.md) — removal counterpart.
- [`internal/secrets/secrets.md`](../internal/secrets/secrets.md) — prompt behavior.
- [`internal/crypto/crypto.md`](../internal/crypto/crypto.md) — generation and save.
- [`internal/recipients/recipients.md`](../internal/recipients/recipients.md) —
  recipient list semantics.
- [`cmd/recipient.md`](recipient.md) — adding the current key back to the list.
- [`cmd/sync.md`](sync.md) — bulk interactive workflow.
