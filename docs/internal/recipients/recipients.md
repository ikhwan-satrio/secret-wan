# `internal/recipients/recipients.go`

> Source: `internal/recipients/recipients.go` · Package: `recipients`

## Purpose

Manages the set of public age recipients that a project's encrypted secrets are
addressed to. It owns the on-disk recipient list format, validation, and the
small query helpers used by the CLI.

The recipient list is public key material. It is safe to commit and contains no
private keys.

## Responsibilities

- Provide the default recipient file name (`secret-wan.recipients`).
- Read the recipient file while treating a missing file as an empty list.
- Parse comment lines, blank lines, and inline comments.
- Validate every entry as an X25519 age recipient.
- Normalize entries by trimming, de-duplicating, and sorting them.
- Write a stable, commented recipient file with mode `0644`.
- Convert recipient strings into `age.Recipient` values.
- Offer `Contains` and `Remove` helpers for command logic.

## Key types

### `Store`

```go
type Store struct {
    Path string
}
```

`Store` is a thin filesystem wrapper. It deliberately has no caching: every
`Load` re-reads the file so commands always observe manual edits.

### `New(path)`

Returns a `Store` for `path`. When `path` is blank it falls back to
`DefaultFile`, so callers that construct a configuration manually still get the
standard file name.

## Functions

### `Store.Exists() (bool, error)`

Uses `os.Stat`. A missing file returns `(false, nil)`; other filesystem errors
are returned to the caller so permission problems are not silently hidden.

### `Store.Load() ([]string, error)`

1. Reads the file.
2. Treats `os.IsNotExist` as "no recipient list yet" and returns `nil, nil`.
3. Splits the content on newlines and delegates to `ParseLines`.
4. Returns the normalized, de-duplicated, sorted list.

A blank result is the signal for **single-recipient mode**: the current master
key is used as the only recipient on the next write.

### `Store.Save(keys []string) error`

1. Normalizes the supplied keys.
2. Writes a short three-line header explaining the file.
3. Writes one recipient per line, sorted.
4. Uses mode `0644` because the contents are public.

Saving an empty list is allowed and produces only the header. The CLI normally
prevents this state because removing the last recipient is rejected.

### `ParseLines(lines []string) ([]string, error)`

Ignores blank lines and lines starting with `#`. Everything after an inline `#`
is treated as a comment, so a line such as:

```text
age1... # teammate laptop
```

is accepted. Each remaining token must parse as an X25519 recipient. An invalid
line produces an error that names the one-based line number.

### `Normalize(keys []string) ([]string, error)`

Trims each key, skips empty strings, parses the key as an X25519 recipient, and
uses the canonical `recipient.String()` form. Duplicates are removed and the
result is sorted alphabetically, which keeps the file stable across machines.

### `Parse(keys []string) ([]age.Recipient, error)`

Converts a normalized list into `[]age.Recipient` for encryption. It re-validates
every entry so callers cannot accidentally pass a stale or malformed key.

### `Contains(keys []string, key string) bool`

Reports exact membership. Inputs are expected to be canonical strings from
`Load` or `Normalize`.

### `Remove(keys []string, key string) []string`

Returns a new slice without the matching key. It does not modify the input
slice.

## Recipient file format

```text
# secret-wan recipients
# Public age keys, one per line. This file is safe to commit.
# Add a teammate with: secwan recipient add <age1...>

age1...
age1...
```

Only X25519 recipients (`age1...`) are accepted in this version. SSH
recipients and post-quantum age keys are rejected by `age.ParseX25519Recipient`.

## Execution flow

```text
Load
 ├── os.ReadFile
 │    ├── not found → nil, nil (single-recipient mode)
 │    └── other     → error
 ├── split lines
 ├── ParseLines
 │    ├── skip blank/comment lines
 │    ├── validate age1... key
 │    └── Normalize (dedupe + sort)
 └── return []string
```

## Dependencies

- `filippo.io/age` for X25519 recipient parsing.
- Standard library `fmt`, `os`, `sort`, and `strings`.

## Edge cases and notes

- A missing file is not an error; it preserves backward compatibility with
  projects created before multi-recipient support.
- Invalid entries are hard errors rather than warnings. This prevents a typo
  from silently locking a teammate out.
- Inline comments are supported, but the `#` character must not appear inside a
  key (X25519 keys do not contain it).
- `Store.Save` rewrites the whole file, so manual ordering is not preserved.
- `Store.Save` does not create parent directories. The recipient file lives in
  the project root, which already exists.
- The file is public, but it should still be reviewed like any other committed
  configuration change.

## Related files

- [`recipients_test.md`](recipients_test.md) — unit tests for this package.
- [`cmd/recipient.md`](../../cmd/recipient.md) — user-facing management commands.
- [`recipients-file.md`](../../recipients-file.md) — project-level file guide.
- [`internal/crypto/crypto.md`](../crypto/crypto.md) — consumes `age.Recipient`
  values.
