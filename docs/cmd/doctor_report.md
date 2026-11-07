# `cmd/doctor_report.go`

> Source: `cmd/doctor_report.go` · Package: `cmd`

## Purpose

Provides the small reporting framework used by `secwan doctor`: check levels,
sections, counters, rendering, summary text, and failure exit status.

## Responsibilities

- Group checks under titled sections.
- Track warning and failure counts.
- Render icons and colors for each level.
- Print a final summary.
- Exit `1` when there are failures.

## Key types

### `checkLevel`

```go
const (
    checkOK checkLevel = iota
    checkWarn
    checkFail
)
```

### `doctorCheck` and `doctorSection`

A check has a level, label, and detail. A section groups related checks under a
title.

### `doctorReport`

Tracks sections plus warning and failure counts.

| Method | Effect |
| :--- | :--- |
| `section(title)` | Appends and returns a new section |
| `add(section, level, label, detail)` | Appends a check and increments counters |
| `ok` / `warn` / `fail` | Level-specific wrappers |
| `render()` | Prints every non-empty section with icons and colors |
| `finish()` | Renders, prints the summary, exits `1` on failures |

## Summary output

`printDoctorSummary` reports one of:

- `Everything looks good.`
- `<n> warning(s), no errors.`
- `<n> error(s), <n> warning(s).`

When there are failures it also prints:

```text
Fix the errors above, then run 'secwan doctor' again.
```

## Rendering details

- Sections with no checks are skipped.
- Labels are left-padded to 26 characters for alignment.
- OK checks use the success icon and muted detail; warnings use the warning icon
  and warning color; failures use the error icon and error color.
- The divider length is 44, matching other command headers.

## Edge cases and notes

- `finish` may terminate the process, so checks must add all important findings
  before calling it.
- Counters are incremented in `add`, so callers cannot create a check without
  updating the totals.
- The report object is not thread-safe; `doctor` is single-threaded.

## Dependencies

- `internal/ui` for styles, icons, and dividers.
- Standard library `fmt` and `os`.

## Related files

- [`doctor.md`](doctor.md) — command entry point.
- [`doctor_checks.md`](doctor_checks.md) — fills the report.
