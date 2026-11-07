# `internal/app/errors.go`

> Source: `internal/app/errors.go` · Package: `app`

## Purpose

Defines errors that carry enough context for the command layer to print a
specific heading and detail line without reimplementing business logic.

## Key types

### `UserError`

```go
type UserError struct {
    Short  string
    Detail string
}
```

`Error()` returns `Detail`, so the error still works with normal Go error
handling. The CLI checks for `*UserError` with `errors.As` and uses `Short` as
the heading.

Examples:

| Short | Detail |
| :--- | :--- |
| `Master key not found` | `Run 'secwan init' or 'secwan set' first.` |
| `Invalid profile` | `Profile 'staging' not found. Available profiles: [default]` |
| `Current key is not a recipient` | `Add it with 'secwan recipient add age1...' or use --force...` |

### `RecipientChangeError`

```go
type RecipientChangeError struct {
    Err         error
    Restored    bool
    RollbackErr error
}
```

This wraps a failed rekey after the recipient file was changed. `Restored` says
whether writing the old list back succeeded. `RollbackErr` keeps the rollback
failure for diagnostics when it did not.

`Unwrap()` returns `Err`, so callers can inspect the original rekey error with
`errors.Is` or `errors.As`.

## Design notes

- Only expected, user-actionable failures become `UserError`. Storage or
  cryptography failures stay as ordinary errors.
- `newUserError` is private because it always returns `error`; package callers
  should not depend on the concrete type when constructing values.
- The CLI helper `failApp` renders `UserError` with its own heading and falls
  back to a command-specific heading for other errors.

## Related files

- [`../../cmd/shared.md`](../../cmd/shared.md) — `failApp` rendering.
- [`recipients.md`](recipients.md) — uses `RecipientChangeError` for rollback
  failures.
- [`project.md`](project.md) and [`secrets.md`](secrets.md) — common
  `UserError` sources.
