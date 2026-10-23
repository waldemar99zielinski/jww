package storage

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"jww/internal/domain"
)

// Check verifies connectivity, credentials, and access to the configured
// bucket without reading or writing an object.
func (s *S3Store) Check(ctx domain.AppContext) error {
	_, err := s.client.HeadBucket(ctx.Context(), &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return fmt.Errorf("storage: checking bucket availability: %w", err)
	}
	return nil
}
