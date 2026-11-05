# `release-please-config.json`

> Source: `release-please-config.json` · Release Please configuration

## Purpose

Configures Release Please to manage semantic versions, tags, and the changelog
for this Go module.

## Configuration

```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "release-type": "go",
  "packages": {
    ".": {
      "release-type": "go",
      "package-name": "secret-wan",
      "include-component-in-tag": false,
      "changelog-path": "CHANGELOG.md",
      "bump-minor-pre-major": true
    }
  }
}
```

## Field reference

| Field | Value | Meaning |
| :--- | :--- | :--- |
| `$schema` | Release Please schema URL | Enables editor validation |
| `release-type` | `go` | Use Go release conventions |
| `packages` | `"."` | One package rooted at the repository |
| `package-name` | `secret-wan` | Stable package identifier |
| `include-component-in-tag` | `false` | Tags are `vX.Y.Z`, without a component prefix |
| `changelog-path` | `CHANGELOG.md` | Changelog file maintained in the repo |
| `bump-minor-pre-major` | `true` | `feat:` bumps minor while major is 0 |

## Versioning behavior

Release Please reads Conventional Commit messages on `main`:

| Commit type | Version effect |
| :--- | :--- |
| `fix:` | Patch |
| `feat:` | Minor (or patch pre-1.0 with `bump-minor-pre-major`) |
| `feat!:` / `BREAKING CHANGE:` | Major |
| `docs:`, `test:`, `chore:` | No version bump (typically excluded) |

The current version is tracked in
[`.release-please-manifest.json`](release-please-manifest.md).

## Repository layout assumption

Because the only package is `"."`, the module root is the release unit. There
are no submodules or monorepo packages to version independently.

## Dependencies

- [`workflows/release-please.md`](workflows/release-please.md) runs the action
  with this file.
- [`release-please-manifest.md`](release-please-manifest.md) stores the last
  released version.
- [`changelog.md`](changelog.md) is the generated changelog.

## Edge cases and notes

- Manual tag releases do not update this configuration or the manifest.
- If the manifest is missing or malformed, Release Please cannot determine the
  next version.
- Changing `package-name` can affect PR titles and changelog headings.
- `include-component-in-tag: false` keeps tags simple; enabling it would produce
  tags like `secret-wan-v1.2.3`.
- The JSON file must remain valid; CI does not currently validate it directly,
  but the workflow will fail if Release Please cannot parse it.

## Related files

- [`workflows/release-please.md`](workflows/release-please.md)
- [`release-please-manifest.md`](release-please-manifest.md)
- [`changelog.md`](changelog.md)
