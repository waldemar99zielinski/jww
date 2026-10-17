// Package s3test starts a disposable MinIO container for tests that need a
// real S3-compatible object store — internal/adapter/storage's integration
// tests and server/tests/e2e's full-stack journeys. It is test-only: nothing
// outside a test imports it, and it never ships in a production binary.
//
// It is deliberately independent of internal/tests/pgtest: the two containers
// have different lifecycles and a test that needs Postgres does not
// necessarily need object storage.
package s3test

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go/modules/minio"

	"jww/internal/adapter/storage"
)

// defaultImage is pinned so local and CI runs are reproducible. MinIO deleted
// its own Docker Hub images (minio/minio) in September 2026, so the default
// points at the community-maintained pgsty/minio rebuild of the same binary.
// Set JWW_TEST_MINIO_IMAGE to override it (e.g. an internal mirror).
const defaultImage = "pgsty/minio:RELEASE.2026-08-04T00-00-00Z"

func image() string {
	if v := os.Getenv("JWW_TEST_MINIO_IMAGE"); v != "" {
		return v
	}
	return defaultImage
}

// Bucket is a running MinIO instance with one pre-created private bucket.
type Bucket struct {
	// Config wires the storage adapter straight at this instance.
	Config storage.Config
	// Endpoint is the http(s) base URL, exposed for diagnostics.
	Endpoint  string
	container *minio.MinioContainer
}

// Start launches MinIO, creates bucketName, and returns a storage.Config that
// points a real storage.S3Store at it. Call once per test binary.
func Start(ctx context.Context, bucketName string) (*Bucket, error) {
	container, err := minio.Run(ctx, image())
	if err != nil {
		return nil, fmt.Errorf("s3test: starting MinIO: %w", err)
	}
	hostPort, err := container.ConnectionString(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("s3test: connection string: %w", err)
	}
	endpoint := "http://" + hostPort

	cfg := storage.Config{
		Bucket:          bucketName,
		Region:          "us-east-1",
		Endpoint:        endpoint,
		ForcePathStyle:  true,
		AccessKeyID:     container.Username,
		SecretAccessKey: container.Password,
	}

	if err := createBucket(ctx, cfg); err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}
	return &Bucket{Config: cfg, Endpoint: endpoint, container: container}, nil
}

// Close terminates the container. Call from TestMain after m.Run().
func (b *Bucket) Close(ctx context.Context) {
	_ = b.container.Terminate(ctx)
}

func createBucket(ctx context.Context, cfg storage.Config) error {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return fmt.Errorf("s3test: loading AWS config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})
	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(cfg.Bucket)}); err != nil {
		return fmt.Errorf("s3test: creating bucket %q: %w", cfg.Bucket, err)
	}
	return nil
}
