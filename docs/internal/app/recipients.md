# `internal/app/recipients.go`

> Source: `internal/app/recipients.go` · Package: `app`

## Purpose

Applies project-level recipient policy: which public keys new writes are
encrypted to, when the current key must be included, how `recipient add/remove`
pair a list change with a rekey, and how a failed rekey is rolled back.

The low-level file format and key parsing live in `internal/recipients`. This
file is the policy layer above them.

## Responsibilities

- Load the public recipient list through the service's `Store`.
- Identify the current user's recipient when a key is available.
- Resolve strings into `[]age.Recipient`.
- Fall back to the current key in single-recipient mode.
- Refuse normal writes that would exclude the current key.
- Rekey the stored map and report what happened.
- Add or remove a recipient and roll back the file on rekey failure.
- Allow adding the current key back to a list that omits it.

## Key types

### `RecipientList`

```go
type RecipientList struct {
    Path      string
    Keys      []string
    Self      string
    SelfKnown bool
}
```

`SelfListed()` reports whether the current key appears in a non-empty list. It
is used by `recipient list`, `recipient me`, and `rekey` warnings.

### `RekeyResult`

```go
type RekeyResult struct {
    Secrets      int
    Recipients   int
    UsedImplicit bool
    Missing      bool
    Self         string
    SelfListed   bool
}
```

| Field | Meaning |
| :--- | :--- |
| `Secrets` | Number of stored secrets written |
| `Recipients` | Number of age recipients used |
| `UsedImplicit` | No file existed; the current key was used |
| `Missing` | Storage does not exist yet, so there was nothing to rekey |
| `Self` | Current public recipient |
| `SelfListed` | Current key appears in the effective list |

### `RecipientChange`

Reports the outcome of `AddRecipient` or `RemoveRecipient`: the canonical key,
whether it was already present, whether the current key was removed, and the
rekey counts.

## Methods

### `RecipientKeys() ([]string, error)`

Loads the normalized public strings. A missing file returns an empty slice.

### `ListRecipients() (RecipientList, error)`

Loads the list and tries to load the current identity. A missing master key is
not fatal because `recipient list` can still display public keys.

### `EncryptionRecipients(allowSelfExclusion) ([]age.Recipient, error)`

1. Rejects a nil identity.
2. Loads the recipient file.
3. Falls back to the current key when the file is empty.
4. Unless `allowSelfExclusion` is set, rejects a list without the current key:

```text
current key age1... is not a recipient; run 'secwan recipient add age1...' first
```

5. Parses every string into an age recipient.

This is the normal write path used by `set`, `unset`, and `sync`.

### `Rekey(allowSelfExclusion) (*RekeyResult, error)`

1. Requires an identity and storage.
2. Loads the recipient list.
3. Uses the current key when the list is empty and records `UsedImplicit`.
4. Enforces the self-inclusion rule unless allowed.
5. Decrypts the stored map.
6. Treats a missing store as `Missing=true` with no error.
7. Re-encrypts the same map for every recipient and returns the counts.

### `AddRecipient(raw) (*RecipientChange, error)`

1. Canonicalizes the argument with `recipients.ParseRecipient`.
2. Ensures the project and identity exist.
3. Loads the list and seeds it with the current key when empty.
4. Allows adding the current key itself to repair a list that omits it; adding a
   different key while the current key is absent is rejected.
5. Returns `AlreadyPresent` without writing when the key is already listed.
6. Saves the new list, then calls `Rekey(false)`.
7. On rekey failure, writes the old list back and returns
   `RecipientChangeError`.

### `RemoveRecipient(raw, force) (*RecipientChange, error)`

1. Canonicalizes the target and requires an existing identity.
2. Loads the list and seeds it with the current key when empty.
3. Rejects a missing target, the last remaining recipient, and self-exclusion
   without `--force`.
4. Saves the reduced list, then calls `Rekey(true)`.
5. On rekey failure, writes the old list back and returns
   `RecipientChangeError`.
6. Reports `SelfRemoved` so the command can warn about future writes.

## Safety model

| Invariant | Where enforced |
| :--- | :--- |
| At least one recipient remains | `RemoveRecipient` |
| A normal write includes the current key | `EncryptionRecipients`, `Rekey` |
| A recipient change is paired with a rekey | `AddRecipient`, `RemoveRecipient` |
| A failed rekey restores the old list | `RecipientChangeError` rollback |
| Adding the current key back is allowed | `AddRecipient` recovery path |

`--force` is the explicit escape hatch for intentional self-exclusion. It never
bypasses the "last recipient" rule.

## Rollback behavior

The list is written first, then the ciphertext. If rekeying fails, the previous
list is restored so the public file never claims a change that the ciphertext did
not receive. If the rollback also fails, `RecipientChangeError.RollbackErr`
carries the second failure and the CLI says the list could not be restored.

## Edge cases and notes

- Empty recipient files are not an error; they preserve compatibility with
  single-recipient projects.
- A non-empty list without the current key is an error for normal writes even if
  another recipient could still decrypt the data.
- Adding your own key is allowed when the list omits it, because that is the
  documented recovery path.
- Removing your own key is allowed only with `--force`; the command warns that
  future writes cannot be decrypted by it.
- The service never edits the recipient file during a plain `rekey`.
- Only X25519 `age1...` keys are accepted.

## Dependencies

- `filippo.io/age` for recipient values.
- `internal/recipients` for parsing, storage, membership, and removal.
- [`secrets.md`](secrets.md) for `isNotFound`.
- [`errors.md`](errors.md) for `UserError` and `RecipientChangeError`.

## Related files

- [`../../cmd/recipient.md`](../../cmd/recipient.md) — CLI commands.
- [`../../cmd/rekey.md`](../../cmd/rekey.md) — standalone rekey command.
- [`../../internal/recipients/recipients.md`](../../internal/recipients/recipients.md)
  — file format and store.
- [`../../internal/recipients/parse.md`](../../internal/recipients/parse.md) —
  parsing and canonicalization.
