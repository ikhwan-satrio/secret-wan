# `cmd/sync.go`

> Source: `cmd/sync.go` · Package: `cmd`

## Purpose

Implements `secwan sync`, the interactive path for filling a profile's secrets.
It generates values for `generate = true` keys, prompts for `required = true`
keys, merges the results into existing storage, and saves the encrypted result.

## Responsibilities

- Read and validate the `--profile` flag.
- Load or create the master key.
- Parse the manifest and find the requested profile.
- Load existing secrets while tolerating a missing store on first use.
- Preserve secrets belonging to other profiles.
- Keep existing non-empty generated values instead of regenerating them.
- Prompt only for required keys that are not generated.
- Resolve the configured public recipients before saving.
- Refuse to save when the current key is not a recipient.
- Save only when something changed.
- Report cancellation without saving.

## Execution flow

```text
sync
 ├── read --profile (default: "default")
 ├── LoadOrGenerateKey
 ├── ParseManifest
 ├── require profile
 ├── GetStorageProvider
 ├── LoadSecrets
 │     ├── success  → reuse map (preserves other profiles)
 │     ├── not found → empty map (first sync)
 │     └── other     → fail
 ├── for each AutoGen key:
 │     ├── existing non-empty → print "kept"
 │     └── otherwise         → GenerateSecurePassword, print "generated"
 ├── requiredKeysForPrompt(profile) → Bubble Tea input model
 │     ├── cancelled → return without saving
 │     └── completed → copy choices into map
 └── if changed
      ├── encryptionRecipients(appConfig, identity, false)
      │     ├── no recipient file → current key
      │     └── current key absent → fail
      └── SaveSecrets(storage, stored, recipients...)
     else → "No changes to save."
```

## Key code

### Merge, don't replace

The command loads the complete secret map and mutates only the target profile's
keys. This is why syncing `staging` does not erase secrets used by `default`.

### Preserve generated values

```go
if existing, ok := stored[key]; ok && existing != "" {
    fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Muted.Render("kept"), ui.Key.Render(key))
    continue
}
```

Running `sync` twice is safe: generated passwords are not rotated unless the
user explicitly sets a new value.

### Prompt selection

`requiredKeysForPrompt(profile)` excludes keys that are also generated. This
prevents a key declared with both rules from being generated and prompted.

### Save condition

`changed` becomes true when a value is generated or when the TUI returns at
least one choice. If nothing changed, the storage is not rewritten.

### Recipient resolution

```go
recips, err := encryptionRecipients(appConfig, identity, false)
if err != nil {
    // fail
}
err = cryptoModule.SaveSecrets(storage, stored, recips...)
```

`sync` uses the package-level helper because it does not build an `appSession`.
The helper reads `secret-wan.recipients`, falls back to the current key when the
file is absent, and rejects a non-empty list that does not contain the current
key.

The success message includes the recipient count:

```text
✓ Secrets encrypted and saved via dotfiles (2 recipient(s))
```

## Dependencies

- `internal/config`, `internal/crypto`, `internal/manifest`,
  `internal/provider`, `internal/tui`, `internal/ui`.
- Shared helper `encryptionRecipients` from `cmd/shared.go`.
- `github.com/charmbracelet/bubbletea` for the input wizard.
- `github.com/spf13/cobra` for the command and `--profile` flag.

## Edge cases and notes

- **Cancellation after generation:** generated values are placed in the
  in-memory map but the command returns before `SaveSecrets`, so nothing is
  persisted. The next `sync` regenerates those values.
- **Profile not found:** the command exits with the sorted list of available
  profiles.
- **Corrupt storage:** a decrypt error is a hard failure; only "not found" is
  treated as a first sync.
- **Unmanaged keys:** keys in storage that the profile does not mention are
  preserved.
- **Empty prompt value:** the TUI stores whatever the user submits. An empty
  value can therefore be saved; `check` and `doctor` treat it as missing.
- **No master key yet:** `sync` creates one, making it a valid bootstrap command
  after or instead of `init`.
- **No recipient file:** the current key is used, preserving the original
  single-recipient behavior.
- **Current key missing from the list:** `sync` fails before saving and tells
  the user to run `secwan recipient add`.
- **Partial rekey ownership:** `sync` only encrypts the full map it just saved;
  it does not modify `secret-wan.recipients`.
- **Docker provider:** saving to Docker secrets is not supported; `sync` will
  fail when it attempts to save through `DockerSecretProvider`.

## Related files

- [`cmd/shared.md`](shared.md) — `requiredKeysForPrompt`, `encryptionRecipients`,
  and session helpers.
- [`internal/tui/base.md`](../internal/tui/base.md) — the input wizard.
- [`internal/crypto/crypto.md`](../internal/crypto/crypto.md) — generation and save.
- [`internal/recipients/recipients.md`](../internal/recipients/recipients.md) —
  recipient list semantics.
- [`cmd/set.md`](set.md) — non-interactive single-key alternative.
