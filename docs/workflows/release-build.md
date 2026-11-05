# `.github/workflows/release-build.yaml`

> Source: `.github/workflows/release-build.yaml` · Workflow: `Release Build`

## Purpose

A reusable workflow (`workflow_call`) that publishes a GitHub release for a
given tag and then adds standalone versioned Windows executables to that
release.

## Inputs

| Input | Required | Type | Meaning |
| :--- | :--- | :--- | :--- |
| `tag` | yes | string | Git tag to check out, build, and publish |

## Permissions

```yaml
permissions:
  contents: write
```

Writing is required to create/update the GitHub release and upload assets.

## Jobs

### `goreleaser`

Runs on `ubuntu-latest`:

1. **Checkout** with `fetch-depth: 0` and `ref: ${{ inputs.tag }}`.
   Full history is needed for changelog generation; the explicit ref ensures the
   tag's commit is built.
2. **Set up Go** from `go.mod` with caching.
3. **Run GoReleaser** with `goreleaser/goreleaser-action@v6`:
   - distribution: `goreleaser`
   - version: `latest`
   - args: `release --clean`
   - `GITHUB_TOKEN` from `${{ github.token }}`

GoReleaser builds all targets, creates archives and checksums, and publishes the
release. Its behavior is defined in [`.goreleaser.yaml`](../goreleaser.md).

### `windows-exe`

Depends on `goreleaser`, so the release and `checksums.txt` already exist.

1. **Checkout** the tag.
2. **Set up Go** from `go.mod`.
3. **Download existing checksums**:

   ```bash
   gh release download "$TAG" --pattern checksums.txt --dir dist
   ```

4. **Build standalone Windows executables**:

   ```bash
   BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
   LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"
   GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS" -o dist/secwan-windows-amd64.exe .
   GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="$LDFLAGS" -o dist/secwan-windows-arm64.exe .
   ```

   `VERSION` is the tag, `COMMIT` is the workflow's `github.sha`, and the build
   date is UTC ISO 8601. This is the same `main.version`, `main.commit`, and
   `main.date` wiring used by GoReleaser, so `secwan version` works identically.

5. **Add checksums**:

   ```bash
   cd dist
   sha256sum *.exe >> checksums.txt
   ```

   The downloaded GoReleaser checksum file is extended with the new executables.

6. **Upload assets**:

   ```bash
   gh release upload "$TAG" \
     dist/secwan-windows-amd64.exe \
     dist/secwan-windows-arm64.exe \
     dist/checksums.txt \
     --clobber
   ```

   `--clobber` makes reruns idempotent.

## Why a separate Windows job

GoReleaser already produces Windows archives, but the project also publishes
bare `.exe` files for users who do not want to extract a zip. Building them in a
second job keeps the GoReleaser configuration focused and allows the checksum
file to be updated consistently.

## Dependencies

- [`release-please.yaml`](release-please.md) calls this workflow after a release
  is created.
- [`release.yaml`](release.md) calls it for manually pushed tags.
- `gh` CLI is preinstalled on GitHub-hosted Ubuntu runners.
- The `GITHUB_TOKEN`/`GH_TOKEN` must have `contents: write`.

## Edge cases and notes

- `needs: goreleaser` also ensures the release exists before downloading
  checksums.
- If `checksums.txt` is missing from the release, the download step fails.
- The standalone executables are not included in the GoReleaser archive
  checksums; they are appended manually.
- Both Windows architectures are built even if one already exists; `--clobber`
  replaces previous assets.
- The GoReleaser job uses `mode: keep-existing`, so it does not overwrite the
  Release Please changelog body.

## Related files

- [`goreleaser.md`](../goreleaser.md)
- [`release-please.md`](release-please.md)
- [`release.md`](release.md)
- [`cmd/version.md`](../cmd/version.md) — consumes the injected values.
