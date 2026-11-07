# `cmd/doctor.go`

> Source: `cmd/doctor.go` · Package: `cmd`

## Purpose

Defines the `secwan doctor` command. The diagnostic logic is split so the command
file stays small and each concern has its own document:

| File | Responsibility |
| :--- | :--- |
| `cmd/doctor.go` | Cobra command registration |
| `cmd/doctor_report.go` | Report model, rendering, summary, exit status |
| `cmd/doctor_checks.go` | Environment, manifest, key, recipients, storage, profile checks |

## Responsibilities

- Register the `doctor` command with `cobra.NoArgs`.
- Delegate to `runDoctor`.

## Key code

```go
var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Diagnose configuration, keys, storage, and secrets",
    Args:  cobra.NoArgs,
    Run: func(cmd *cobra.Command, args []string) {
        runDoctor()
    },
}
```

## Edge cases and notes

- `doctor` is read-oriented and never creates a master key or storage.
- Failures exit `1`; warnings alone exit `0`.
- The command is intentionally thin so the report and checks can be read without
  Cobra noise.

## Dependencies

- [`doctor_report.md`](doctor_report.md)
- [`doctor_checks.md`](doctor_checks.md)
- `github.com/spf13/cobra`.

## Related files

- [`check.md`](check.md) — narrower completeness check.
- [`../internal/crypto/key.md`](../internal/crypto/key.md) — key path helper.
