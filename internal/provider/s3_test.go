package provider

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
)

type fakeS3 struct {
	saved     []byte
	putInput  *s3.PutObjectInput
	getOutput *s3.GetObjectOutput
	getErr    error
	headErr   error
}

func (f *fakeS3) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.putInput = params
	data, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}
	f.saved = data
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getOutput, nil
}

func (f *fakeS3) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if f.headErr != nil {
		return nil, f.headErr
	}
	return &s3.HeadObjectOutput{}, nil
}

func TestS3ProviderSaveAndLoad(t *testing.T) {
	fake := &fakeS3{}
	provider := &S3Provider{Bucket: "bucket", Key: "demo/" + testEncFile, client: fake}

	if err := provider.Save([]byte("ciphertext")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if string(fake.saved) != "ciphertext" {
		t.Fatalf("saved = %q, want ciphertext", fake.saved)
	}

	fake.getOutput = &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("ciphertext")))}
	got, err := provider.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(got) != "ciphertext" {
		t.Fatalf("Load() = %q, want ciphertext", got)
	}

	if err := provider.Check(); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestS3ProviderMissingObjectIsNotFound(t *testing.T) {
	provider := &S3Provider{
		Bucket: "bucket",
		Key:    "demo/" + testEncFile,
		client: &fakeS3{getErr: &smithy.GenericAPIError{Code: "NoSuchKey"}},
	}

	_, err := provider.Load()
	if !IsNotFound(err) {
		t.Fatalf("Load() error = %v, want ErrNotFound", err)
	}

	provider.client = &fakeS3{headErr: &smithy.GenericAPIError{Code: "NotFound"}}
	if err := provider.Check(); !IsNotFound(err) {
		t.Fatalf("Check() error = %v, want ErrNotFound", err)
	}
}

func TestS3ProviderRequiresBucket(t *testing.T) {
	provider := &S3Provider{Key: "demo/" + testEncFile}

	_, err := provider.Load()
	if err == nil || !strings.Contains(err.Error(), s3BucketEnv) {
		t.Fatalf("Load() error = %v, want a %s error", err, s3BucketEnv)
	}
}

func TestNewS3ProviderDerivesObjectKey(t *testing.T) {
	t.Setenv(s3KeyEnv, "")
	t.Setenv(s3BucketEnv, "team-secrets")
	t.Setenv(s3RegionEnv, "ap-southeast-1")
	t.Setenv(s3PathStyleEnv, "true")

	instance := newS3Provider(config.AppConfig{EncFile: testEncFile}, "demo").(*S3Provider)

	if instance.Key != "demo/"+testEncFile {
		t.Fatalf("Key = %q, want demo/%s", instance.Key, testEncFile)
	}
	if instance.Bucket != "team-secrets" {
		t.Fatalf("Bucket = %q, want team-secrets", instance.Bucket)
	}
	if instance.Region != "ap-southeast-1" {
		t.Fatalf("Region = %q, want ap-southeast-1", instance.Region)
	}
	if !instance.PathStyle {
		t.Fatal("PathStyle = false, want true")
	}
}

func TestNewS3ProviderHonorsExplicitKey(t *testing.T) {
	t.Setenv(s3KeyEnv, "custom/path.enc")
	t.Setenv(s3BucketEnv, "")

	instance := newS3Provider(config.AppConfig{EncFile: testEncFile}, "demo").(*S3Provider)
	if instance.Key != "custom/path.enc" {
		t.Fatalf("Key = %q, want custom/path.enc", instance.Key)
	}
}
