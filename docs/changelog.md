# `CHANGELOG.md`

> Source: `CHANGELOG.md` · Generated release history

## Purpose

Records notable changes between releases. It is generated and maintained by
[Release Please](https://github.com/googleapis/release-please) from Conventional
Commit messages.

## Current content

The file currently contains a title, a short description, and a note that the
changelog is generated automatically. Release entries are added by Release
Please as releases are created.

## How it is updated

1. Commits merged into `main` use Conventional Commit prefixes (`feat:`, `fix:`,
   etc.).
2. Release Please opens a release PR containing the next version and a new
   changelog section.
3. Merging the PR persists the changelog entry and creates the tag/release.

GoReleaser is configured with `mode: keep-existing`, so it does not overwrite the
release body that Release Please generated.

## Dependencies

- [`release-please-config.md`](release-please-config.md) — sets
  `changelog-path` to this file.
- [`workflows/release-please.md`](workflows/release-please.md) — runs the action.
- [`goreleaser.md`](goreleaser.md) — includes this file in archives.

## Edge cases and notes

- Do not hand-edit released sections; the next Release Please run may conflict.
- Manual tag releases do not add entries here.
- Documentation, test, and chore commits are typically excluded from the
  GoReleaser changelog groups, and Release Please may also omit them depending
  on configuration.
- The file is included in every release archive, even when it has no entries for
  that version.

## Related files

- [`release-please-config.md`](release-please-config.md)
- [`release-please-manifest.md`](release-please-manifest.md)
- [`workflows/release-please.md`](workflows/release-please.md)
