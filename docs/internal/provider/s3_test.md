# `internal/provider/s3_test.go`

> Source: `internal/provider/s3_test.go` · Package: `provider`

## Purpose

Unit tests for the S3 provider. A fake `s3API` captures requests so the tests run
without AWS credentials or a network connection.

## Test helper

```go
type fakeS3 struct {
    saved     []byte
    putInput  *s3.PutObjectInput
    getOutput *s3.GetObjectOutput
    getErr    error
    headErr   error
}
```

`fakeS3` implements `PutObject`, `GetObject`, and `HeadObject`:

- `PutObject` reads the request body and records the bytes.
- `GetObject` returns the configured body or error.
- `HeadObject` returns the configured error, or success.

## Test cases

| Test | What it proves |
| :--- | :--- |
| `TestS3ProviderSaveAndLoad` | Save writes the exact bytes; Load returns them; Check succeeds |
| `TestS3ProviderMissingObjectIsNotFound` | `NoSuchKey`/`NotFound` map to `ErrNotFound` for Load and Check |
| `TestS3ProviderRequiresBucket` | A missing bucket yields a configuration error naming `SECWAN_S3_BUCKET` |
| `TestNewS3ProviderDerivesObjectKey` | The default key is `<project>/<EncFile>` and env config is applied |
| `TestNewS3ProviderHonorsExplicitKey` | `SECWAN_S3_KEY` overrides the derived key |

## Key code

```go
provider := &S3Provider{Bucket: "bucket", Key: "demo/" + testEncFile, client: fake}
```

Injecting `client` bypasses the `sync.Once` initializer and the AWS SDK, so no
credentials are required. Missing-object errors use `smithy.GenericAPIError`,
which satisfies the `smithy.APIError` interface used by `isS3NotFound`.

## Dependencies

- `bytes`, `context`, `io`, `strings`, and `testing` from the standard library.
- `github.com/aws/aws-sdk-go-v2/service/s3` and `github.com/aws/smithy-go`.
- [`s3.md`](s3.md) — the code under test.

## Edge cases and notes

- The configuration tests use `t.Setenv` to isolate `SECWAN_S3_*` values.
- `TestNewS3ProviderDerivesObjectKey` clears `SECWAN_S3_KEY` so the default path
  is exercised.
- The tests use `testEncFile = "secrets.enc"` instead of the default name, which
  proves providers honor `AppConfig.EncFile` rather than a hardcoded filename.
- The fake does not validate request contexts or deadlines; timeout behavior is
  covered by the implementation, not these tests.
- No test creates a bucket or contacts AWS.

## Related files

- [`s3.md`](s3.md) — S3 provider implementation.
- [`provider.md`](provider.md) — `ErrNotFound` and `IsNotFound`.
- [`registry.md`](registry.md) — builds the provider from environment variables.
