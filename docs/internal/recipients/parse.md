# `internal/recipients/parse.go`

> Source: `internal/recipients/parse.go` · Package: `recipients`

## Purpose

Parses, validates, canonicalizes, and converts public recipient strings. Keeping
this separate from storage makes the file format and the key rules independently
readable.

## Responsibilities

- Parse file lines with comments and blank lines.
- Validate a single command-line recipient.
- Normalize lists by trimming, de-duplicating, and sorting.
- Convert strings into `age.Recipient` values.

## Functions

### `ParseLines(lines []string) ([]string, error)`

Ignores blank lines and lines starting with `#`. Everything after an inline `#`
is treated as a comment, so this is accepted:

```text
age1... # teammate laptop
```

Each remaining token must parse as an X25519 recipient. An invalid line produces
an error naming the one-based line number. The valid entries are passed through
`Normalize`.

### `ParseRecipient(raw) (string, error)`

Trims the input, rejects an empty value with `recipient must not be empty`,
parses it as an X25519 recipient, and returns `recipient.String()`. This
canonical form is what `recipient add/remove` store and compare.

### `Normalize(keys []string) ([]string, error)`

Trims each key, skips empty strings, parses the key as an X25519 recipient, and
uses the canonical string form. Duplicates are removed and the result is sorted
alphabetically, keeping the file stable across machines.

### `Parse(keys []string) ([]age.Recipient, error)`

Converts a normalized list into `[]age.Recipient` for encryption. It re-validates
every entry so callers cannot accidentally pass a stale or malformed key.

## Edge cases and notes

- Empty strings are ignored by `Normalize`; `ParseRecipient` rejects them because
  a command-line argument must be non-empty.
- Inline comments are supported, but `#` cannot appear inside an X25519 key.
- Invalid entries are hard errors rather than warnings. A typo should not
  silently lock a teammate out.
- Sorting is byte-wise alphabetical, which is deterministic across runs.
- The package only supports X25519 recipients in this version.

## Dependencies

- `filippo.io/age` for X25519 parsing.
- Standard library `fmt`, `sort`, and `strings`.

## Related files

- [`recipients.md`](recipients.md) — store and format.
- [`keys.md`](keys.md) — membership helpers.
- [`../../internal/app/recipients.md`](../../internal/app/recipients.md) — uses
  `ParseRecipient` and `Parse`.
