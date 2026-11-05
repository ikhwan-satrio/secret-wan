# `.github/workflows/release-please.yaml`

> Source: `.github/workflows/release-please.yaml` · Workflow: `Release Please`

## Purpose

Automates semantic versioning and release creation. On every push to `main`, it
runs Release Please. When Release Please creates a release, this workflow calls
the reusable release build.

## Triggers

```yaml
on:
  push:
    branches: [main]
```

## Permissions

```yaml
permissions:
  contents: write
  pull-requests: write
  issues: write
```

Release Please needs these permissions to create release PRs, labels, and
releases.

## Jobs

### `release-please`

Runs `googleapis/release-please-action@v4` with:

- `token: ${{ secrets.GITHUB_TOKEN }}`
- `config-file: release-please-config.json`
- `manifest-file: .release-please-manifest.json`

Outputs:

| Output | Meaning |
| :--- | :--- |
| `release_created` | `"true"` when a release was just created |
| `tag_name` | The created tag, for example `v1.2.0` |

### `release-build`

```yaml
needs: release-please
if: ${{ needs.release-please.outputs.release_created == 'true' }}
permissions:
  contents: write
uses: ./.github/workflows/release-build.yaml
with:
  tag: ${{ needs.release-please.outputs.tag_name }}
```

The job is skipped unless a release was created. It delegates all building and
asset upload to [`release-build.yaml`](release-build.md).

## Release flow

1. A conventional commit is merged into `main`.
2. Release Please opens or updates a release PR with:
   - the next version
   - an updated `CHANGELOG.md`
   - an updated `.release-please-manifest.json`
3. When the release PR is merged, Release Please creates the tag and GitHub
   Release.
4. `release_created` becomes `"true"`, so `release-build` runs.
5. GoReleaser publishes archives/checksums, then the Windows job adds standalone
   `.exe` assets.

## Versioning rules

Versioning behavior comes from
[`release-please-config.json`](../release-please-config.md):

- `release-type: go`
- `bump-minor-pre-major: true`, so `feat:` bumps the minor version while the
  major version is `0`.
- The package name is `secret-wan`.
- The component is not included in the tag (`v1.2.3`, not
  `secret-wan-v1.2.3`).

The current tracked version is in
[`.release-please-manifest.json`](../release-please-manifest.md).

## Dependencies

- `release-please-config.json`
- `.release-please-manifest.json`
- `.github/workflows/release-build.yaml`
- Conventional Commit messages for correct changelog and semver output.

## Edge cases and notes

- The workflow only runs on `main`; release PRs targeting `main` are not release
  triggers until merged.
- `release_created` is a string output, so the condition compares it to
  `'true'`.
- The release build receives the tag produced by Release Please, not the branch
  SHA.
- If the release build fails, the GitHub Release may exist with partial or no
  assets; rerunning the failed job is safe because uploads use `--clobber`.
- Manual tags can bypass this workflow and use [`release.yaml`](release.md).

## Related files

- [`release-please-config.md`](../release-please-config.md)
- [`release-please-manifest.md`](../release-please-manifest.md)
- [`release-build.md`](release-build.md)
- [`changelog.md`](../changelog.md)
