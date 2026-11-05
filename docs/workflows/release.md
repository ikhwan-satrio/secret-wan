# `.github/workflows/release.yaml`

> Source: `.github/workflows/release.yaml` · Workflow: `Release`

## Purpose

Provides a manual release path: pushing a tag that starts with `v` triggers the
same reusable build used by Release Please.

## Triggers

```yaml
on:
  push:
    tags:
      - 'v*'
```

Examples: `v1.0.0`, `v1.2.3`, `v2.0.0-rc.1`.

## Permissions

```yaml
permissions:
  contents: write
```

## Job

### `release`

```yaml
release:
  uses: ./.github/workflows/release-build.yaml
  with:
    tag: ${{ github.ref_name }}
```

It delegates entirely to [`release-build.yaml`](release-build.md), passing the
pushed tag name.

## When to use this

- You need to release without going through a Release Please PR.
- You are creating a hotfix tag directly.
- You want to rebuild assets for an existing tag.

For the normal flow, use Release Please as described in
[`release-please.md`](release-please.md).

## Dependencies

- `.github/workflows/release-build.yaml`
- A tag that matches `v*`
- GoReleaser configuration in [`.goreleaser.yaml`](../goreleaser.md)

## Edge cases and notes

- The workflow uses `github.ref_name`, which is the short tag name (for example
  `v1.0.1`).
- Pushing a tag does not update `.release-please-manifest.json`. If you release
  manually, the manifest may lag behind the actual latest release until Release
  Please runs again.
- The tag must already point at the commit you want to build; the reusable
  workflow checks out that ref.
- Re-running the workflow for the same tag republishes assets; the Windows job
  uses `--clobber`.

## Related files

- [`release-build.md`](release-build.md)
- [`release-please.md`](release-please.md)
- [`goreleaser.md`](../goreleaser.md)
