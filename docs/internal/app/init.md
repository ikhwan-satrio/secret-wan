# `internal/app/init.go`

> Source: `internal/app/init.go` · Package: `app`

## Purpose

Provides the two bootstrap operations used by `secwan init`: creating the starter
manifest and seeding the recipient list for a brand-new project.

## Responsibilities

- Write the provider-specific manifest template only when the spec file is
  absent.
- Leave an existing manifest untouched.
- Create the recipient file only when it does not exist.
- Seed a new recipient file with the current public key.
- Keep recipient creation failures non-fatal for the command layer.

## Methods

### `CreateManifest(providerName) (bool, error)`

1. Checks whether `AppConfig.SpecFile` exists.
2. Returns `false, nil` when it does.
3. Otherwise asks `manifest.GetTemplate(providerName)` for the template and
   writes it with mode `0644`.
4. Returns `true, nil` when a file was created.

The provider name is already mapped by the command from the TUI selection.

### `CreateRecipientList() (bool, error)`

1. Requires an identity to be loaded.
2. Checks whether the recipient file exists.
3. Returns `false, nil` when it does.
4. Otherwise saves a one-key list containing `identity.Recipient()`.
5. Returns `true, nil` when a file was created.

A recipient creation error is returned to the CLI, which prints it as a warning.
The project can still operate in single-recipient mode.

## Why recipient creation is conditional

`init` creates the recipient list only for a newly created manifest. Existing
projects keep their original single-recipient behavior until the user opts in
with `secwan recipient add`. This avoids changing the encryption policy of a
project merely because a newer binary ran `init`.

## Edge cases and notes

- An existing `secret-wan.toml` is never overwritten.
- An existing `secret-wan.recipients` is never overwritten, even in a new
  project.
- The manifest is written before the recipient list, matching the observable
  order of the original command.
- Recipient creation is best-effort; the command treats it as a warning.
- The master key is created by `EnsureIdentity` before these methods are called;
  `CreateRecipientList` itself does not create or load a key.
- The recipient file is public key material and is written with mode `0644`.

## Dependencies

- `internal/manifest` for the starter template.
- `internal/recipients` through the service's `Store`.
- Standard library `os` for file creation.

## Related files

- [`service.md`](service.md) — provides `Config`, `Identity`, and `Recipients`.
- [`../../cmd/init.md`](../../cmd/init.md) — TUI and output layer.
- [`../../internal/manifest/template.md`](../../internal/manifest/template.md) —
  template contents.
