# `cmd/shared_test.go`

> Source: `cmd/shared_test.go` · Package: `cmd`

## Purpose

Unit tests for the shared command helpers that contain the most subtle logic:
profile key normalization, strict-check issue collection, environment-value
quoting, and secret type classification.

## Responsibilities

- Verify that profile keys are de-duplicated and sorted.
- Verify that generated keys are excluded from interactive prompts.
- Verify that unmanaged-key detection sees every profile.
- Verify shell-safe quoting for `print-env`.
- Verify strict-check findings for empty profiles, missing rules, and unknown
  providers.
- Verify the `list` command's secret type labels.

## Tests

### `TestProfileKeysDeduplicatesAndSorts`

Input:

```go
Required: []string{"B", "A"}
AutoGen:  []string{"A", "C"}
```

Expected output: `[]string{"A", "B", "C"}`. This protects the union behavior in
`profileKeys`.

### `TestRequiredKeysForPromptSkipsGenerated`

Input has `A` and `B` required, and `B` and `C` generated. Expected prompt list
is only `A`. This prevents a key from being generated and then prompted again.

### `TestAllDeclaredKeysIncludesEveryProfile`

Builds a manifest with `default` declaring `A` and `B`, and `staging`
declaring `C`. Asserts that all three are present and that `D` is absent.

### `TestQuoteEnvValue`

Table-driven cases:

| Input | Expected |
| :--- | :--- |
| `hello` | `"hello"` |
| `a"b` | `"a\"b"` |
| `$HOME` | `"\$HOME"` |
| `a\b` | `"a\\b"` |

The test covers the escaping rules used for safe `eval "$(secwan print-env)"`.

### `TestCollectStrictIssues`

Manifest contains a project, a valid `default` profile with key `A`, a `B` key
with no action, and an empty profile. Stored secrets include `A` and an
undeclared `EXTRA`.

There are exactly three issues:

1. `EXTRA` is stored but not declared in any profile.
2. `empty` has no secrets defined.
3. `B` has no rule.

### `TestCollectStrictIssuesFlagsUnknownProvider`

Uses provider `vault` and expects a `project.provider` issue listing the valid
providers (`dotfiles`, `keyring`, `docker`).

### `TestSecretType`

Checks the labels used by `list`:

- `A` in its own profile → `required`
- `B`, declared only in `staging` → `other`
- `Z`, not declared anywhere → `unmanaged`

## Helper

`hasIssue` performs an exact `(label, detail)` match over the issue slice, which
makes assertions resilient to ordering.

## Edge cases and notes

- The tests do not touch the filesystem, the keyring, or the network.
- They use profiles constructed directly, including `Declared` metadata, so they
  also exercise the defensive fallbacks in the shared helpers.
- `cmd` has no other test files yet; these tests are the main regression guard
  for the daily-management features.

## Related files

- [`cmd/shared.md`](shared.md)
- [`cmd/check.md`](check.md)
- [`cmd/list.md`](list.md)
- [`cmd/printEnv.md`](printEnv.md)
