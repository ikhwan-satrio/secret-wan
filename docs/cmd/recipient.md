# `cmd/recipient.go`

> Source: `cmd/recipient.go` · Package: `cmd`

## Purpose

Implements the `recipient` command family for managing the public age keys that
can decrypt a project's secrets:

```text
secwan recipient list
secwan recipient me
secwan recipient add <age1...>
secwan recipient remove <age1...> [--force]
```

`add` and `remove` re-encrypt existing secrets automatically, so the recipient
file and the encrypted payload stay consistent.

## Responsibilities

- List configured recipients and mark the current key.
- Show single-recipient fallback mode when no recipient file exists.
- Print the current user's public recipient (`me` / `whoami`).
- Validate new recipients as X25519 age public keys.
- Seed the recipient file with the current key when it is absent.
- Refuse to add a different recipient when the current key is not in the list.
- Allow adding the current key itself to repair a list that omits it.
- Re-encrypt the stored secret map after adding or removing a recipient.
- Roll the recipient file back if re-encryption fails.
- Refuse to remove the last recipient.
- Require `--force` to remove the current user's own key.

## Commands

### `recipient list` (alias `ls`)

1. Reads the recipient file.
2. Attempts to load the current master key; a missing key is not fatal.
3. Prints the file path.
4. If the list is empty, prints the single-recipient fallback and the implicit
   current key.
5. Otherwise prints every recipient and marks the current key with `(you)`.

A warning is written to stderr when the current key is missing from a non-empty
list, because future writes would not be decryptable by the current user.

### `recipient me` (alias `whoami`)

Loads the existing master key and prints only the public `age1...` recipient to
stdout. This makes it script-friendly:

```bash
TEAMMATE_KEY="$(secwan recipient me)"
```

If a recipient file exists but does not contain the current key, a warning is
written to stderr. A missing master key fails with a hint to run `init`.

### `recipient add <age1...>`

1. Normalize the argument with `age.ParseX25519Recipient`.
2. Require the manifest and an identity. The identity may be created if this is
   the first write in the project.
3. Load the recipient list.
4. If the file is empty, seed it with the current key (single-recipient mode).
5. Fail if the current key is not in the list and the key being added is a
   different key; adding somebody else while locking yourself out would be
   surprising. Adding your own key is allowed, because it repairs the list.
6. Print `Recipient is already configured.` and stop when the key is already
   present.
7. Save the new list, print the added key, then call `session.rekey(false)`.
8. If re-encryption fails, restore the previous recipient list and fail.

On success it prints the number of secrets and recipients that were re-encrypted.
If no secrets are stored yet, it explains that future saves will use every
recipient.

### `recipient remove <age1...>` (alias `rm`)

1. Normalize the argument.
2. Read `--force`.
3. Require the manifest and an existing master key.
4. Load the list, seeding it with the current key when the file is absent.
5. Fail if the target is not in the list.
6. Fail if the target is the last recipient.
7. Require `--force` when the current key is not a recipient.
8. Require `--force` when removing the current key.
9. Save the reduced list, print the removal, then re-encrypt with
   `session.rekey(true)`, which allows the current key to be excluded.
10. On re-encryption failure, restore the previous list.
11. Warn that the removed key cannot decrypt future writes when it was the
    current key.

## Key helpers

### `normalizeRecipient(raw string) (string, error)`

Trims the input, rejects an empty value, parses it as an X25519 recipient, and
returns `recipient.String()`. This canonical form is what gets stored and
compared.

### `printRekeyResult(count, recipientCount int)`

Renders either:

```text
✓ Re-encrypted 3 secret(s) for 2 recipient(s).
```

or, when there is nothing stored yet:

```text
No stored secrets to re-encrypt yet.
Future saves will be encrypted to all configured recipients.
```

## Safety model

The command layer enforces three invariants:

| Invariant | Reason |
| :--- | :--- |
| At least one recipient remains | A file with zero keys would make new writes impossible to decrypt |
| A normal write includes the current key | Prevents accidental self-lockout |
| Recipient changes are paired with a rekey | Old ciphertext must not remain readable only by removed keys |

`--force` is the explicit escape hatch for intentionally locking the current key
out of future writes. It never bypasses the "last recipient" rule.

## Rollback behavior

`add` and `remove` save the recipient file first, then re-encrypt. If the rekey
fails, they write the old list back. This avoids a state where the public list
claims a key is present but the ciphertext was never updated, or vice versa.

If both the rekey and the rollback fail, the error message says so explicitly;
the project should then be inspected manually with `secwan doctor`.

## Dependencies

- `filippo.io/age` for recipient parsing.
- `internal/config`, `internal/crypto`, `internal/recipients`, `internal/ui`.
- Shared helpers: `newSession`, `requireManifest`, `requireIdentity`,
  `requireExistingIdentity`, `recipientsStore`, `rekey`, `fail`, `failErr`,
  `printHeader`.
- `github.com/spf13/cobra`.

## Edge cases and notes

- `recipient list` works without a master key so it can still show the configured
  public keys.
- `recipient me` writes only the key to stdout; warnings go to stderr.
- `recipient add` can create the master key because it is a write command. This
  lets a fresh project bootstrap its recipient list.
- If a non-empty list omits the current key but the existing ciphertext is still
  decryptable by it (for example after a manual list edit), adding the current
  key repairs the list and rekeys. Adding a different key still fails.
- If the payload was already rekeyed without the current key, adding it back
  cannot decrypt the old payload and the recipient list is rolled back.
- `recipient remove` requires an existing key and manifest.
- The command is exact about keys: `age1...` strings are canonicalized, but there
  is no fuzzy matching.
- Only X25519 recipients are supported in this version. SSH recipients are
  rejected with the age parser's error.
- Adding a duplicate key is a no-op, not an error, which makes scripts
  idempotent.
- The recipient file is rewritten in sorted order by the internal store.

## Related files

- [`rekey.md`](rekey.md) — standalone re-encryption command.
- [`shared.md`](shared.md) — recipient resolution and save helpers.
- [`../internal/recipients/recipients.md`](../internal/recipients/recipients.md)
  — file format and validation.
- [`../recipients-file.md`](../recipients-file.md) — user-facing file guide.
