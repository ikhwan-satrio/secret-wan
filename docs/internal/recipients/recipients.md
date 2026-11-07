# `internal/recipients/recipients.go`

> Source: `internal/recipients/recipients.go` · Package: `recipients`

## Purpose

Defines the public recipient sidecar and its filesystem store. The package is
split into:

| File | Responsibility |
| :--- | :--- |
| `recipients.go` | `Store`, file existence, load, save |
| `parse.go` | Line parsing, canonicalization, age conversion |
| `keys.go` | Membership and removal helpers |

The recipient list is public key material. It is safe to commit and contains no
private keys.

## Responsibilities

- Provide the default file name (`secret-wan.recipients`).
- Read the file while treating a missing file as an empty list.
- Write a stable, commented file with mode `0644`.
- Expose a small store type so callers do not duplicate path handling.

## Key type

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
`DefaultFile`, so callers that construct configuration manually still get the
standard file name.

## Methods

### `Store.Exists() (bool, error)`

Uses `os.Stat`. A missing file returns `(false, nil)`; other filesystem errors
are returned so permission problems are not silently hidden.

### `Store.Load() ([]string, error)`

1. Reads the file.
2. Treats `os.IsNotExist` as "no recipient list yet" and returns `nil, nil`.
3. Splits the content on newlines and delegates to `ParseLines`.
4. Returns the normalized, de-duplicated, sorted list.

A blank result is the signal for **single-recipient mode**: the current master
key is used as the only recipient on the next write.

### `Store.Save(keys []string) error`

1. Normalizes the supplied keys with `Normalize`.
2. Writes a short three-line header explaining the file.
3. Writes one recipient per line, sorted.
4. Uses mode `0644` because the contents are public.

Saving an empty list is allowed and produces only the header. The CLI normally
prevents that state because removing the last recipient is rejected.

## Recipient file format

```text
# secret-wan recipients
# Public age keys, one per line. This file is safe to commit.
# Add a teammate with: secwan recipient add <age1...>

age1...
age1...
```

Only X25519 recipients (`age1...`) are accepted in this version. SSH recipients
and post-quantum age keys are rejected by `age.ParseX25519Recipient`.

## Edge cases and notes

- A missing file is not an error; it preserves backward compatibility with
  projects created before multi-recipient support.
- `Store.Save` rewrites the whole file, so manual ordering is not preserved.
- `Store.Save` does not create parent directories. The recipient file lives in
  the project root, which already exists.
- The file is public, but it should still be reviewed like any other committed
  configuration change.

## Dependencies

- `filippo.io/age` indirectly through `parse.go`.
- Standard library `os` and `strings`.

## Related files

- [`parse.md`](parse.md) — parsing and canonicalization.
- [`keys.md`](keys.md) — membership helpers.
- [`recipients_test.md`](recipients_test.md) — unit tests.
- [`../../internal/app/recipients.md`](../../internal/app/recipients.md) — policy
  layer.
- [`../../recipients-file.md`](../../recipients-file.md) — project-level guide.
