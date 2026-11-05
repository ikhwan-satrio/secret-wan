# `internal/manifest/helper.go`

> Source: `internal/manifest/helper.go` · Package: `manifest`

## Purpose

Provides a tiny helper for listing profile names in a deterministic order.

## Responsibilities

- Extract all keys from the `Profiles` map.
- Sort them alphabetically.
- Return a slice suitable for error messages and reports.

## Key code

```go
func GetProfileNames(profiles map[string]*Profile) []string {
    names := make([]string, 0, len(profiles))
    for k := range profiles {
        names = append(names, k)
    }
    sort.Strings(names)
    return names
}
```

## Why it exists

Go map iteration order is randomized. Several commands print "Available
profiles: [...]" or list profiles in reports. Sorting makes that output stable
across runs and easier to test.

## Used by

- `cmd/shared.go` — invalid profile errors.
- `cmd/check.go` — strict checks iterate profiles in sorted order.
- `cmd/doctor.go` — manifest summary and profile checks.
- `cmd/sync.go` and `cmd/exec.go` — unknown-profile messages.

## Dependencies

- Standard library `sort`.

## Edge cases and notes

- A nil or empty map returns an empty, non-nil slice.
- The helper only sorts names; it does not validate profile contents.
- Profile names are case-sensitive, matching TOML table names.

## Related files

- [`internal/manifest/parser.md`](parser.md) — produces the map.
- [`cmd/shared.md`](../../cmd/shared.md) — consumes the names in error messages.
