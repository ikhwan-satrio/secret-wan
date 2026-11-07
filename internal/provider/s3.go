package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
)

const (
	defaultS3Timeout = 30 * time.Second
	s3ConfigTimeout  = 15 * time.Second

	s3BucketEnv    = "SECWAN_S3_BUCKET"
	s3KeyEnv       = "SECWAN_S3_KEY"
	s3RegionEnv    = "SECWAN_S3_REGION"
	s3EndpointEnv  = "SECWAN_S3_ENDPOINT"
	s3PathStyleEnv = "SECWAN_S3_PATH_STYLE"
)

// s3API is the subset of the S3 client used by the provider. Keeping it small
// makes the provider testable without a network connection.
type s3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// S3Provider stores the encrypted payload as one object in an S3-compatible
// bucket (AWS S3, MinIO, Cloudflare R2, Backblaze B2, and similar).
//
// All configuration comes from the environment so credentials and bucket
// coordinates never need to be written into the project manifest.
type S3Provider struct {
	Bucket    string
	Key       string
	Region    string
	Endpoint  string
	PathStyle bool
	Timeout   time.Duration

	client  s3API
	once    sync.Once
	initErr error
}

func newS3Provider(appConfig config.AppConfig, projectName string) StorageProvider {
	key := strings.TrimSpace(os.Getenv(s3KeyEnv))
	if key == "" {
		object := strings.TrimSpace(appConfig.EncFile)
		if object != "" {
			prefix := strings.Trim(strings.TrimSpace(projectName), "/")
			if prefix == "" {
				key = object
			} else {
				key = prefix + "/" + object
			}
		}
	}

	pathStyle, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv(s3PathStyleEnv)))

	return &S3Provider{
		Bucket:    strings.TrimSpace(os.Getenv(s3BucketEnv)),
		Key:       key,
		Region:    strings.TrimSpace(os.Getenv(s3RegionEnv)),
		Endpoint:  strings.TrimSpace(os.Getenv(s3EndpointEnv)),
		PathStyle: pathStyle,
	}
}

// Save writes the ciphertext to the configured object.
func (p *S3Provider) Save(data []byte) error {
	api, err := p.api()
	if err != nil {
		return err
	}

	ctx, cancel := p.context()
	defer cancel()

	_, err = api.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.Bucket),
		Key:    aws.String(p.Key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("write s3://%s/%s: %w", p.Bucket, p.Key, err)
	}

	return nil
}

// Load reads the ciphertext from the configured object.
func (p *S3Provider) Load() ([]byte, error) {
	api, err := p.api()
	if err != nil {
		return nil, err
	}

	ctx, cancel := p.context()
	defer cancel()

	output, err := api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.Bucket),
		Key:    aws.String(p.Key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, fmt.Errorf("%w: s3://%s/%s", ErrNotFound, p.Bucket, p.Key)
		}
		return nil, fmt.Errorf("read s3://%s/%s: %w", p.Bucket, p.Key, err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("read s3://%s/%s body: %w", p.Bucket, p.Key, err)
	}

	return data, nil
}

// Check verifies that the object exists and is reachable.
func (p *S3Provider) Check() error {
	api, err := p.api()
	if err != nil {
		return err
	}

	ctx, cancel := p.context()
	defer cancel()

	_, err = api.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.Bucket),
		Key:    aws.String(p.Key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return fmt.Errorf("%w: s3://%s/%s", ErrNotFound, p.Bucket, p.Key)
		}
		return fmt.Errorf("check s3://%s/%s: %w", p.Bucket, p.Key, err)
	}

	return nil
}

func (p *S3Provider) api() (s3API, error) {
	if p.client != nil {
		return p.client, nil
	}

	p.once.Do(func() {
		if err := p.validate(); err != nil {
			p.initErr = err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), s3ConfigTimeout)
		defer cancel()

		var options []func(*awsconfig.LoadOptions) error
		if p.Region != "" {
			options = append(options, awsconfig.WithRegion(p.Region))
		}

		cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
		if err != nil {
			p.initErr = fmt.Errorf("load AWS configuration: %w", err)
			return
		}

		p.client = s3.NewFromConfig(cfg, func(options *s3.Options) {
			if p.Endpoint != "" {
				options.BaseEndpoint = aws.String(p.Endpoint)
			}
			options.UsePathStyle = p.PathStyle
		})
	})

	return p.client, p.initErr
}

func (p *S3Provider) validate() error {
	if p.Bucket == "" {
		return fmt.Errorf("%s is not set", s3BucketEnv)
	}
	if p.Key == "" {
		return fmt.Errorf("%s is not set and the encrypted file name is empty", s3KeyEnv)
	}

	return nil
}

func (p *S3Provider) context() (context.Context, context.CancelFunc) {
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = defaultS3Timeout
	}

	return context.WithTimeout(context.Background(), timeout)
}

// isS3NotFound recognizes the error shapes produced for a missing object by
// both AWS S3 and S3-compatible services.
func isS3NotFound(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	switch apiErr.ErrorCode() {
	case "NoSuchKey", "NotFound", "404":
		return true
	default:
		return false
	}
}
