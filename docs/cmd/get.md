# `cmd/get.go`

> Source: `cmd/get.go` · Package: `cmd`

## Purpose

Implements `secwan get <KEY>`, which prints a single stored secret. The value is
masked unless `--reveal` is passed.

## Responsibilities

- Require exactly one argument.
- Read `--profile` and `--reveal`.
- Require an existing master key and existing storage.
- Fail when the key is not stored.
- Warn when the key is stored but not declared in the selected profile.
- Print the plain value with `--reveal`, or a fixed mask otherwise.

## Execution flow

1. `cobra.ExactArgs(1)` validates the key argument.
2. Build a session, parse the manifest, resolve the profile.
3. Require an existing identity and load secrets.
4. Look up the key in the stored map.
5. If absent, exit with a hint to run `secwan set <KEY>`.
6. If the key is not declared in the profile, print a warning to stderr. The
   warning includes the profiles that do declare it when known.
7. Print either the value or the mask.

## Key code

### Message for undeclared keys

```go
if !keyInProfile(profile, key) {
    detail := fmt.Sprintf("%s is not declared in profile '%s'.", key, profile.Name)
    if others := profilesContaining(session.Manifest, key); len(others) > 0 {
        detail = fmt.Sprintf("%s is declared in profile(s) %s, not '%s'.",
            key, strings.Join(others, ", "), profile.Name)
    }
    fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning, ui.Warning.Render(detail))
}
```

This is a warning, not a failure: the command still prints the value if it is
stored. This is useful when diagnosing a manifest that has not been updated yet.

### Masking

`maskedValue` is defined in [`cmd/list.md`](list.md) as `••••••••`. `get` reuses
the constant so both commands display the same mask.

## Dependencies

- `internal/ui` for the warning and mask styling.
- Shared helpers: `newSession`, `requireManifest`, `requireProfile`,
  `requireExistingIdentity`, `requireSecrets`, `keyInProfile`,
  `profilesContaining`.
- `github.com/spf13/cobra`.

## Edge cases and notes

- `get` never creates a key or storage; a missing setup is an error.
- The mask is intentionally not based on the value length, so it does not leak
  length information.
- `--reveal` prints the raw value followed by a newline. Values with embedded
  newlines therefore span multiple lines.
- The key lookup is case-sensitive.
- The command does not check whether the value is empty. An empty stored value is
  printed as an empty line with `--reveal`.
- Warnings go to stderr so `value="$(secwan get KEY --reveal)"` stays clean.

## Related files

- [`cmd/list.md`](list.md) — list multiple keys and their types.
- [`cmd/set.md`](set.md) — create or update a value.
- [`cmd/printEnv.md`](printEnv.md) — export all values.
