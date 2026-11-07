# `internal/crypto/key.go`

> Source: `internal/crypto/key.go` · Package: `crypto`

## Purpose

Owns the master-key lifecycle: locating the key file, loading an existing
identity, and generating a new one only when a command is allowed to bootstrap.

## Responsibilities

- Compute the key path with and without filesystem side effects.
- Create the config directory with owner-only permissions before writing.
- Parse an existing age identity file.
- Generate and persist a new X25519 identity.

## Functions

### `KeyPath(appName) (string, error)`

1. Resolves the OS config directory through `osutil.ConfigDir`.
2. Joins it with `keys.txt`.
3. Does **not** create anything.

`doctor` uses this so diagnostics have no side effects. `LoadKey` uses it too.

### `GetKeyPath(appName) (string, error)`

Same path, but calls `os.MkdirAll(dir, 0700)` first. `LoadOrGenerateKey` uses it
before writing a new key.

### `LoadKey() (*age.X25519Identity, error)`

1. Resolves the path with `KeyPath`.
2. Opens the file. A missing file returns the raw `os.ErrNotExist` so callers
   can detect it with `os.IsNotExist` or `errors.Is`.
3. Parses identities with `age.ParseIdentities`.
4. Rejects an empty identity list.
5. Accepts only the first identity when it is `*age.X25519Identity`.
6. Rejects other types with `unsupported key type`.

Read-only commands use this method so they never generate a key that cannot
decrypt existing storage.

### `LoadOrGenerateKey() (*age.X25519Identity, error)`

1. Ensures the config directory exists (`GetKeyPath`, mode `0700`).
2. If `keys.txt` exists, delegates to `LoadKey`.
3. If `os.Stat` fails with something other than "not exist", returns the error.
4. Generates an X25519 identity with `age.GenerateX25519Identity`.
5. Creates the file with mode `0600`.
6. Writes the private identity and a comment containing the public recipient.

## Edge cases and notes

- The key file contains the **private** identity. Treat it as sensitive.
- The public recipient comment is informational; encryption derives the
  recipient from the identity, not from the comment.
- `LoadKey` wraps parse errors with `%w` but does not wrap the initial open
  error, preserving missing-file detection.
- `LoadOrGenerateKey` truncates an existing file only after confirming it does
  not exist; it does not overwrite a readable key.
- On Windows, Unix mode bits are less meaningful, but the same calls are used.

## Dependencies

- `filippo.io/age` for identity parsing and generation.
- `internal/osutil` for config directories.
- Standard library `os`, `path/filepath`, and `fmt`.

## Related files

- [`crypto.md`](crypto.md) — module type.
- [`secrets.md`](secrets.md) — uses the identity to decrypt.
- [`../../cmd/doctor.md`](../../cmd/doctor.md) — uses `KeyPath`.
- [`../../internal/app/project.md`](../../internal/app/project.md) — wraps these
  methods.
