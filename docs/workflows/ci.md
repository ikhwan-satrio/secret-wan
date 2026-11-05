# `.github/workflows/ci.yaml`

> Source: `.github/workflows/ci.yaml` · Workflow: `CI`

## Purpose

Continuous integration for pushes to `main`, pull requests targeting `main`, and
manual runs. It verifies formatting, vetting, tests, and cross-platform builds.

## Triggers

```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:
```

## Permissions

```yaml
permissions:
  contents: read
```

CI only reads the repository; it does not publish artifacts.

## Jobs

### `test`

Runs on `ubuntu-latest`:

1. **Checkout** with `actions/checkout@v4`.
2. **Set up Go** with `actions/setup-go@v5`, using `go-version-file: go.mod` and
   dependency caching.
3. **Check formatting**:

   ```bash
   unformatted="$(gofmt -l .)"
   if [ -n "$unformatted" ]; then
     echo "These files are not formatted:"
     echo "$unformatted"
     exit 1
   fi
   ```

   Any file reported by `gofmt -l` fails the job.

4. **Vet**: `go vet ./...`.
5. **Test**: `go test ./...`.

### `build`

Depends on `test` and runs a matrix with `fail-fast: false`, so one platform
failure does not cancel the others.

| `goos` | `goarch` | Output |
| :--- | :--- | :--- |
| linux | amd64 | `secwan` |
| linux | arm64 | `secwan` |
| darwin | amd64 | `secwan` |
| darwin | arm64 | `secwan` |
| windows | amd64 | `secwan.exe` |
| windows | arm64 | `secwan.exe` |

Each matrix entry:

1. Checks out the repository.
2. Sets up Go from `go.mod` with caching.
3. Builds with:

   ```bash
   GOOS=<os> GOARCH=<arch> CGO_ENABLED=0 \
     go build -trimpath -ldflags="-s -w" -o dist/<output> .
   ```

   `CGO_ENABLED=0` produces static binaries. `-trimpath` removes local paths.
   `-s -w` strips symbols and debug information for a smaller binary.

4. Uploads the binary as the artifact `secwan-<goos>-<goarch>` with
   `if-no-files-found: error`.

CI does not inject version metadata, so these build artifacts report the default
`dev` version. Release builds inject the real values.

## Dependencies

- Go version comes from `go.mod`.
- The `test` job produces no artifacts; the `build` job creates downloadable CI
  artifacts, not GitHub release assets.

## Edge cases and notes

- Formatting is checked from the repository root and includes all Go packages.
- Tests run without `-race`; race detection is not part of this workflow.
- The build matrix covers only the six published targets.
- Pull requests from forks run with read-only permissions.
- Release Please pull requests also run CI like any other pull request.

## Related files

- [`workflows/release-build.md`](release-build.md) — production build and release.
- [`goreleaser.md`](../goreleaser.md) — release build configuration.
- [`go-mod.md`](../go-mod.md) — source of the Go version.
