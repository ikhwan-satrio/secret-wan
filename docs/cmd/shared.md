# `cmd/shared.go`

> Source: `cmd/shared.go` · Package: `cmd`

## Purpose

Contains the small cross-command helpers that are purely about CLI presentation:
styled error output, warning output, and the standard command header.

Business logic and session loading have moved to [`cmd/session.go`](session.md)
and [`internal/app`](../internal/app/service.md).

## Responsibilities

- Print styled errors to **stderr** and exit with status `1`.
- Render `app.UserError` values with their own heading and detail.
- Print styled warnings to stderr.
- Print the shared command header.

## Key functions

### `fail(shortMsg, detailMsg)`

Writes `ui.FormatError(shortMsg, detailMsg)` to stderr and calls `os.Exit(1)`.
Using stderr keeps stdout usable for values, notably in `print-env` and `get`.

### `failErr(shortMsg, err)`

Convenience wrapper that uses `err.Error()` as the detail line. A nil error
produces an empty detail.

### `failApp(defaultShort, err)`

Checks for `*app.UserError` with `errors.As`:

- when found, uses `UserError.Short` and `UserError.Detail`;
- otherwise, uses `defaultShort` and `err.Error()`.

This is why service errors such as `Invalid profile` and
`Current key is not a recipient` keep a specific heading after the refactor.

### `warn(message)`

Writes an icon and warning-styled message to stderr. Commands use it for
non-fatal issues:

- undeclared keys;
- keys still referenced by a profile after `unset`;
- missing self key in a recipient list;
- recipient-list creation failures during `init`.

### `printHeader(action)`

Prints:

```text
secret-wan  •  <action>
────────────────────────────────────────────
```

The divider width is 44.

## Edge cases and notes

- Error messages are written in English and use the shared UI theme.
- `fail` and `failErr` terminate the process; they do not return an error.
- `warn` never changes the exit status.
- Commands that need custom rollback rendering use `failRecipientChange` in
  [`recipient.md`](recipient.md), which falls back to `failApp`.

## Dependencies

- `internal/app` for `UserError`.
- `internal/ui` for styles and icons.
- Standard library `errors`, `fmt`, and `os`.

## Related files

- [`session.md`](session.md) — service loading wrappers.
- [`recipient.md`](recipient.md) — recipient-change error rendering.
- [`../internal/app/errors.md`](../internal/app/errors.md) — error types.
