# `internal/manifest/profile_test.go`

> Source: `internal/manifest/profile_test.go` · Package: `manifest`

## Purpose

Unit tests for the profile and manifest query methods that were extracted from
the command package.

## Responsibilities

- Verify key union, de-duplication, and sorting.
- Verify generated keys are excluded from prompt lists.
- Verify declared-key collection across profiles.
- Verify profile membership checks cover rules and key lists.
- Verify profiles-containing lookup is sorted.

## Tests

### `TestProfileKeysDeduplicatesAndSorts`

Input:

```go
Required: []string{"B", "A"}
AutoGen:  []string{"A", "C"}
```

Expected output: `[]string{"A", "B", "C"}`. This protects the union used by
`check`, `list`, `print-env`, and `doctor`.

### `TestProfilePromptKeysSkipsGenerated`

`A` and `B` are required, `B` and `C` generated. Expected prompt list is only
`A`. This prevents a key from being generated and then prompted again.

### `TestManifestDeclaredKeysIncludesEveryProfile`

A manifest with `default` declaring `A` and `B`, and `staging` declaring `C`.
Asserts that all three are present and that `D` is absent.

### `TestProfileHasKeyUsesRulesAndLists`

A profile with one required key, one generated key, and one rule-only key.
All three must be recognized; an unrelated key must not.

### `TestManifestProfilesContaining`

Two profiles declare `A`; expects sorted names `[default staging]`. The
`default` profile uses a `Rules` entry, while `staging` uses `Required`, so both
membership paths are exercised.

## Edge cases and notes

- Tests construct profiles directly, including the defensive fallback fields.
- No filesystem or TOML parsing is involved; the parser has its own coverage
  through integration use.
- The tests moved from `cmd/shared_test.go` when the helpers became domain
  methods.

## Related files

- [`profile.md`](profile.md) — implementation under test.
- [`parser.md`](parser.md) — types and parsing.
- [`../../cmd/shared_test.md`](../../cmd/shared_test.md) — remaining command-layer
  tests.
