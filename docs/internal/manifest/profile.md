# `internal/manifest/profile.go`

> Source: `internal/manifest/profile.go` · Package: `manifest`

## Purpose

Holds domain queries about profiles and manifests: which keys a profile manages,
which keys need prompting, whether a key is declared, and which profiles contain
a key. These helpers used to live in the command package; they belong with the
manifest types because they describe manifest data, not CLI presentation.

## Responsibilities

- Return sorted, de-duplicated key sets for a profile.
- Separate required keys from generated keys.
- Answer membership questions about a profile.
- Return a key's rule without nil-map panics.
- Build the full key list used by `list --all`.
- Collect declared keys across a manifest.
- Find profiles that declare a key.

## `Profile` methods

### `Keys() []string`

Returns the union of `Required` and `AutoGen`, sorted and de-duplicated. A key
declared with both rules appears once.

This is the expected-key set for `check`, `print-env`, `list`, and `doctor`.

### `PromptKeys() []string`

Returns required keys that are **not** auto-generated. A key with both rules
should be generated, not prompted twice. `sync` uses this list for its Bubble Tea
input model.

### `HasKey(key) bool`

Returns true when the key appears in `Rules`, `Required`, or `AutoGen`. The slice
checks keep the method safe for profiles constructed without a `Rules` map.

### `Rule(key) SecretRule`

Returns the key's rule, or a zero rule when `Rules` is nil or the key is absent.
This prevents nil-map panics in strict checks and reports.

### `AllKeys(stored) []string`

Returns the union of profile-managed keys and every stored key, sorted. This
backs `list --all`, which shows unmanaged keys as well as expected ones.

## `Manifest` methods

### `DeclaredKeys() map[string]bool`

Collects every key from every profile's `Declared` list, then also from
`Required` and `AutoGen` as a defensive fallback. Stored keys outside this set are
"unmanaged".

### `ProfilesContaining(key) []string`

Returns sorted profile names whose `HasKey` reports the key. `get`, `set`, and
`unset` use this to explain that a key belongs to another profile.

## Edge cases and notes

- `Declared` alone is not enough for `HasKey`: a profile parsed from TOML also
  populates `Rules`, while tests may construct profiles with only `Required` or
  `AutoGen`.
- `Keys` de-duplicates before sorting, so output order is stable.
- `PromptKeys` preserves the order of `Required`; the parser already sorts that
  slice.
- A nil manifest or profile is not handled; callers pass parsed values or
  test-constructed values with the fields they need.
- The methods do not mutate the profile or manifest.

## Dependencies

- Standard library `sort`.

## Related files

- [`parser.md`](parser.md) — produces the fields these methods query.
- [`profile_test.md`](profile_test.md) — unit tests.
- [`../../cmd/check.md`](../../cmd/check.md),
  [`../../cmd/list.md`](../../cmd/list.md),
  [`../../cmd/doctor.md`](../../cmd/doctor.md) — consumers.
