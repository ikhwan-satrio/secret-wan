# `secret-wan.recipients`

> Generated project file · Public age recipient list

## Purpose

`secret-wan.recipients` is a plain-text sidecar file that lists the public age
recipients allowed to decrypt a project's secrets. It lives next to
`secret-wan.toml` and is normally committed to the repository.

The file contains **public** keys only. Private keys stay in
`~/.config/secret-wan/keys.txt` (or the OS-specific config directory) and must
never be committed.

## Why a sidecar file

The recipient list could live inside `secret-wan.toml`, but the sidecar keeps
public key material separate from the manifest:

- TOML comments and formatting are preserved when recipients change.
- The file is easy to review in a pull request.
- It can be shared without exposing the rest of the project configuration.
- Existing projects can add it without rewriting their manifest.

## Format

```text
# secret-wan recipients
# Public age keys, one per line. This file is safe to commit.
# Add a teammate with: secwan recipient add <age1...>

age1...
age1...
```

Rules:

- One X25519 `age1...` recipient per line.
- Blank lines are ignored.
- Lines starting with `#` are ignored.
- Text after `#` on a normal line is treated as a comment.
- Duplicate keys are removed.
- The list is written in sorted order.
- File mode is `0644`.

Only X25519 recipients are supported in this version.

## Single-recipient fallback

If the file does not exist, `secret-wan` keeps the pre-multi-recipient
behavior: the current master key is used as the only recipient for new writes.
This makes the upgrade backward compatible.

When the file exists but is empty, the same fallback applies for reads and
writes, but the CLI normally prevents an empty list through
`recipient remove`.

## Common workflows

### Add a teammate

```bash
# The teammate sends you their public key:
secwan recipient me

# You add it; existing secrets are re-encrypted automatically:
secwan recipient add age1...
```

Commit the updated recipient file.

### Remove a teammate

```bash
secwan recipient remove age1...
```

Existing secrets are re-encrypted for the remaining recipients. The removed key
can still read old commits that contain previous ciphertext, so rotate any
secret that the teammate should no longer know.

### Rotate your own key

```bash
secwan recipient add age1...new-key
secwan recipient remove age1...old-key --force
```

`--force` is required because removing your current key means future writes
cannot be decrypted by it.

### Apply manual edits

```bash
# Edit secret-wan.recipients by hand, then:
secwan rekey
```

### Inspect the list

```bash
secwan recipient list
secwan recipient me
secwan doctor
```

## Safety rules

| Rule | Behavior |
| :--- | :--- |
| Never write zero recipients | `SaveSecrets` rejects an empty recipient slice |
| Keep the current key on normal writes | `set`, `unset`, and `sync` fail if the current key is missing |
| Rekey after changes | `recipient add` and `recipient remove` reencrypt automatically |
| Allow explicit self-exclusion | `remove --force` and `rekey --force` |
| Keep at least one recipient | `recipient remove` rejects removing the last key |
| Roll back on failure | Recipient changes are restored if re-encryption fails |

## Git and sharing notes

- Commit `secret-wan.recipients`; it is public key material.
- Do not commit the configured encrypted file (default `env.enc`) unless the
  project intentionally uses the `dotfiles` provider for GitOps.
- Removed recipients can still decrypt old Git history; rotate exposed secrets.
- `keyring` storage is per machine, so this file mainly helps with `dotfiles`
  and documents the intended access list for other providers.

## How it relates to the manifest

`secret-wan.toml` still controls project name, provider, profiles, and secret
rules. `secret-wan.recipients` only controls who can decrypt the stored payload.
Changing the recipient file does not change the manifest, and changing the
manifest does not change the recipient list.

## Related files

- [`cmd/recipient.md`](cmd/recipient.md) — management commands.
- [`cmd/rekey.md`](cmd/rekey.md) — manual re-encryption.
- [`internal/recipients/recipients.md`](internal/recipients/recipients.md) —
  store and file format.
- [`internal/app/recipients.md`](internal/app/recipients.md) — recipient policy
  and rekey.
- [`internal/crypto/secrets.md`](internal/crypto/secrets.md) — payload
  encryption.
