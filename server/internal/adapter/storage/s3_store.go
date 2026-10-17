// Package storage implements the shared object_storage.Store port against a
// private S3-compatible object store using the AWS SDK for Go v2. It is
// deliberately purpose-agnostic: it adds no namespace and knows nothing about
// attachments (or any other caller), so a single client can back any number of
// features that pass their own object keys. It is the only place AWS SDK types
// appear; everything inward of it deals in io.Reader/io.ReadCloser and the
// object_storage result structs.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"jww/internal/domain"
	"jww/internal/service/object_storage"
)

// uploadConcurrency keeps a single upload to a small number of parts in
// flight; memory use is roughly part size × concurrency, and this workload is
// personal/small scale.
const uploadConcurrency = 2

// S3Store implements object_storage.Store against one bucket. Keys are used
// verbatim; callers own any folder/namespace convention.
type S3Store struct {
	client   *s3.Client
	uploader *transfermanager.Client
	bucket   string
}

// NewS3Store builds the adapter from cfg. It loads the AWS default credential
// chain unless a static credential pair is configured, and never creates the
// bucket — deployment (or local init) owns bucket creation and the policy.
func NewS3Store(ctx context.Context, cfg Config) (*S3Store, error) {
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKeyID != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("storage: loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
		// Ask only when the operation requires a checksum so third-party
		// S3-compatible stores (MinIO and friends) are not sent algorithms
		// they may not implement.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	uploader := transfermanager.New(client, func(o *transfermanager.Options) {
		o.Concurrency = uploadConcurrency
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})

	return &S3Store{client: client, uploader: uploader, bucket: cfg.Bucket}, nil
}

// Put streams body to ref.Key. The stream's length is unknown, so the transfer
// manager chooses a single PutObject or a multipart upload; either way the whole
// file is never held in memory here.
func (s *S3Store) Put(ctx domain.AppContext, ref object_storage.Ref, body io.Reader, mediaType string) (object_storage.Stored, error) {
	out, err := s.uploader.UploadObject(ctx.Context(), &transfermanager.UploadObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(ref.Key),
		Body:        body,
		ContentType: aws.String(mediaType),
	})
	if err != nil {
		return object_storage.Stored{}, err
	}
	return object_storage.Stored{
		Version: aws.ToString(out.VersionID),
		ETag:    aws.ToString(out.ETag),
	}, nil
}

// Open returns a streaming reader for ref, at the exact stored version when one
// is known.
func (s *S3Store) Open(ctx domain.AppContext, ref object_storage.Ref) (object_storage.Stream, error) {
	in := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(ref.Key),
	}
	if ref.Version != "" {
		in.VersionId = aws.String(ref.Version)
	}
	out, err := s.client.GetObject(ctx.Context(), in)
	if err != nil {
		if isObjectNotFound(err) {
			return object_storage.Stream{}, fmt.Errorf("%w: %v", object_storage.ErrNotFound, err)
		}
		return object_storage.Stream{}, err
	}
	return object_storage.Stream{
		Body:     out.Body,
		ByteSize: aws.ToInt64(out.ContentLength),
	}, nil
}

// Delete removes the exact version (or the current object when no version is
// recorded). S3 delete is idempotent: deleting an absent object or version
// succeeds, so callers can retry a partially failed cleanup.
func (s *S3Store) Delete(ctx domain.AppContext, ref object_storage.Ref) error {
	in := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(ref.Key),
	}
	if ref.Version != "" {
		in.VersionId = aws.String(ref.Version)
	}
	_, err := s.client.DeleteObject(ctx.Context(), in)
	return err
}

func isObjectNotFound(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound":
			return true
		}
	}
	var noSuchKey *s3types.NoSuchKey
	return errors.As(err, &noSuchKey)
}

// compile-time assertion that the adapter satisfies the port.
var _ object_storage.Store = (*S3Store)(nil)
