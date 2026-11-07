# `cmd/shared_test.go`

> Source: `cmd/shared_test.go` · Package: `cmd`

## Purpose

Unit tests for the command-layer logic that remains in `cmd`: environment-value
quoting, strict-check issue collection, and the `list` secret-type classifier.

Profile-key tests moved to
[`internal/manifest/profile_test.md`](../internal/manifest/profile_test.md).
Recipient-resolution tests moved to
[`internal/app/recipients_test.md`](../internal/app/recipients_test.md).

## Responsibilities

- Verify shell-safe quoting for `print-env`.
- Verify strict-check findings for empty profiles, missing rules, and unknown
  providers.
- Verify the `list` command's secret type labels.

## Tests

### `TestQuoteEnvValue`

Table-driven cases:

| Input | Expected |
| :--- | :--- |
| `hello` | `"hello"` |
| `a"b` | `"a\"b"` |
| `$HOME` | `"\$HOME"` |
| `a\b` | `"a\\b"` |

This covers the escaping rules used for safe
`eval "$(secwan print-env)"`.

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
providers from `provider.Names()` (`docker`, `dotfiles`, `env`, `keyring`, `s3`).

### `TestSecretType`

Checks the labels used by `list`:

- `A` in its own profile → `required`
- `B`, declared only in `staging` → `other`
- `Z`, not declared anywhere → `unmanaged`

## Helper

`hasIssue` performs an exact `(label, detail)` match over the issue slice, which
makes assertions resilient to ordering.

## Edge cases and notes

- These tests do not touch the filesystem, the keyring, or the network.
- Profiles are constructed directly, including `Declared` and `Rules`, so the
  defensive fallbacks in the manifest methods are exercised.
- Strict checks now call `Profile.Keys()`, `Profile.Rule()`, and
  `Manifest.DeclaredKeys()`.

## Related files

- [`shared.md`](shared.md) — error and output helpers.
- [`check.md`](check.md) — strict checks.
- [`list.md`](list.md) — secret types.
- [`printEnv.md`](printEnv.md) — quoting.
- [`../internal/manifest/profile_test.md`](../internal/manifest/profile_test.md)
  — moved profile tests.
- [`../internal/app/recipients_test.md`](../internal/app/recipients_test.md) —
  moved recipient tests.
