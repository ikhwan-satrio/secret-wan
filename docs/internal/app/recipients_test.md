# `internal/app/recipients_test.go`

> Source: `internal/app/recipients_test.go` · Package: `app`

## Purpose

Regression tests for recipient policy, rekey behavior, self-lockout prevention,
self-readd recovery, and rollback when a rekey fails after the recipient file was
changed.

## Responsibilities

- Verify single-recipient fallback when no file exists.
- Verify the current key is required for normal writes.
- Verify configured team keys are preserved.
- Verify adding the current key back repairs a list that omits it.
- Verify adding a different key while the current key is missing is refused.
- Verify the last recipient cannot be removed.
- Verify removing a recipient rekeys the payload so the removed key loses access.
- Verify a failed rekey restores the previous recipient list.

## Test helper

### `memoryStorage`

Implements `provider.StorageProvider` in memory:

| Method | Behavior |
| :--- | :--- |
| `Save(data)` | Copies the ciphertext, unless `saveErr` is set |
| `Load()` | Returns the bytes, or `os.ErrNotExist` when empty |

`saveErr` lets a test allow initial setup to succeed and then fail the rekey
write.

### `newRecipientService`

Builds a `Service` with:

- a temporary recipient file path;
- a generated identity already assigned;
- a minimal manifest and an in-memory storage provider;
- an optional pre-populated recipient file.

## Tests

### `TestEncryptionRecipientsUsesCurrentKeyWhenNoFile`

No recipient file: expects exactly one age recipient, derived from the current
key.

### `TestEncryptionRecipientsRequiresCurrentKey`

The file contains only another identity:

- normal resolution returns a `UserError`;
- self-exclusion allowed returns one recipient.

### `TestEncryptionRecipientsKeepsConfiguredKeys`

The file contains self and another key: expects both recipients.

### `TestAddRecipientAllowsSelfReadd`

Starting from a list containing only another key, adding the current key must
succeed, produce a two-key list, and report zero secrets because storage is
empty.

### `TestAddRecipientRefusesDifferentKeyWhenSelfMissing`

Starting from a list without the current key, adding a third key must fail. This
is the guard against accidentally locking the operator out.

### `TestRemoveRecipientRejectsLastRecipient`

A one-key list cannot be reduced to zero, even when the target is the current
key.

### `TestRemoveRecipientRekeysRemainingKey`

Starts with self and another key, saves a secret to both, removes the other key,
and verifies:

- the operation reports one secret for one recipient;
- the remaining key can decrypt;
- the removed key receives `age.ErrIncorrectIdentity`.

### `TestAddRecipientRollsBackOnRekeyFailure`

Saves a secret for the current key, then makes storage writes fail. Adding a
second key must:

- return `*RecipientChangeError`;
- report `Restored=true`;
- leave the recipient file containing only the original key.

## Helper

`containsAll` checks membership without depending on slice order, since the store
sorts recipients.

## Edge cases and notes

- Tests never touch the user's real config directory or project files.
- The service's manifest and storage fields are pre-set so `EnsureProject`
  returns without reading a manifest from disk.
- The identity is pre-set so `EnsureIdentity` and `LoadIdentity` do not need a
  real key file.
- Rollback is tested by failing `storage.Save`, not by failing decryption, so the
  list write and rekey order is exercised.

## Related files

- [`recipients.md`](recipients.md) — implementation under test.
- [`../../internal/recipients/recipients_test.md`](../../internal/recipients/recipients_test.md)
  — lower-level parsing and store tests.
- [`../../cmd/recipient.md`](../../cmd/recipient.md) — user-facing commands.
