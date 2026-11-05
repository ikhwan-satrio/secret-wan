# `.release-please-manifest.json`

> Source: `.release-please-manifest.json` · Release Please version manifest

## Purpose

Stores the last released version for each package so Release Please can compute
the next version from Conventional Commits.

## Content

```json
{
  ".": "1.0.0"
}
```

| Key | Meaning |
| :--- | :--- |
| `"."` | The repository root package, matching `packages["."]` in the config |
| `"1.0.0"` | The last version Release Please knows about |

## How it is updated

Release Please updates this file as part of its release PR. When the PR is
merged, the new version is persisted and the tag/release is created.

Do not bump it manually unless you are repairing release state. The value must
match the latest released tag; otherwise the next release PR may propose an
incorrect version.

## Dependencies

- [`release-please-config.md`](release-please-config.md) — package definition.
- [`workflows/release-please.md`](workflows/release-please.md) — action input.

## Edge cases and notes

- Manual tags do **not** update this file, so a manual release can leave the
  manifest behind the latest tag.
- The key must match the package path in the config (`"."`).
- The version is stored without a leading `v`.
- If the file is deleted, Release Please may treat the package as new or fail,
  depending on the action version and repository history.

## Related files

- [`release-please-config.md`](release-please-config.md)
- [`workflows/release-please.md`](workflows/release-please.md)
- [`changelog.md`](changelog.md)
