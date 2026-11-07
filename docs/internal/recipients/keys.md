# `internal/recipients/keys.go`

> Source: `internal/recipients/keys.go` · Package: `recipients`

## Purpose

Small pure helpers for working with a normalized recipient list. They contain no
filesystem or age logic, so they are cheap to reuse and test.

## Functions

### `Contains(keys []string, key string) bool`

Reports exact membership. Inputs are expected to be canonical strings from
`Load`, `Normalize`, or `ParseRecipient`.

### `Remove(keys []string, key string) []string`

Returns a new slice without the matching key. It does not modify the input slice.

## Edge cases and notes

- Matching is exact and case-sensitive; recipient strings are lowercase
  canonical keys.
- `Remove` removes every occurrence, although normalized lists should not contain
  duplicates.
- Neither helper sorts the result; callers normally pass a normalized list or
  save through `Store.Save`, which normalizes again.

## Dependencies

- None beyond the language runtime.

## Related files

- [`parse.md`](parse.md) — produces canonical keys.
- [`recipients.md`](recipients.md) — store normalization.
- [`../../internal/app/recipients.md`](../../internal/app/recipients.md) — uses
  both helpers for self-lockout checks and removal.
