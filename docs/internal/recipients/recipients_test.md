# `internal/recipients/recipients_test.go`

> Source: `internal/recipients/recipients_test.go` · Package: `recipients`

## Purpose

Unit tests for recipient parsing, normalization, storage, single-recipient
command-line keys, and set helpers. The implementation is split across
`recipients.go`, `parse.go`, and `keys.go`; this file exercises all three.

## Responsibilities

- Verify comments, blank lines, inline comments, duplicates, and whitespace.
- Verify invalid recipients are rejected.
- Verify single-argument canonicalization.
- Verify the `Store` save/load round trip and missing-file behavior.
- Verify `Contains` and `Remove`.

## Tests

### `TestParseLinesNormalizesAndSorts`

Builds a list containing a comment, a blank line, two generated recipients, an
inline comment, and a duplicate. It expects exactly two canonical keys in sorted
order.

### `TestParseLinesRejectsInvalidRecipient`

Passes `not-a-recipient` and expects an error.

### `TestStoreRoundTrip`

Creates a `Store` in `t.TempDir()` and checks:

1. `Exists()` is false before saving.
2. `Save()` succeeds.
3. `Exists()` becomes true.
4. `Load()` returns the saved recipient.

### `TestStoreLoadMissingIsEmpty`

Loads a path that does not exist and expects `nil, nil`. This is the contract
that enables backward-compatible single-recipient mode.

### `TestContainsAndRemove`

Uses synthetic strings to check exact membership and confirms that `Remove`
returns a new slice containing only the remaining key.

### `TestNormalizeRejectsEmptyAndInvalid`

Confirms that an empty key is ignored but a malformed key is an error.

### `TestParseRecipientCanonicalizesAndRejectsEmpty`

Trims a valid recipient, checks the canonical string, and verifies that empty and
malformed command-line inputs are rejected. This covers the helper used by
`recipient add/remove`.

## Edge cases and notes

- The tests generate real X25519 identities with `age.GenerateX25519Identity`,
  so sample keys are always valid for the parser.
- Filesystem tests use `t.TempDir()` and do not touch the user's config
  directory.
- The tests do not exercise file permissions; `Save` always uses `0644`.
- The parser tests cover both file-line rules and single-argument rules.

## Related files

- [`recipients.md`](recipients.md) — store and file format.
- [`parse.md`](parse.md) — parsing and canonicalization.
- [`keys.md`](keys.md) — membership helpers.
- [`../../cmd/recipient.md`](../../cmd/recipient.md) — command layer.
