package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/config"
)

// S3Client wraps the AWS S3 client for file operations.
type S3Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
	log       *zap.Logger
}

// NewS3Client creates a new S3 client connected to the configured endpoint.
func NewS3Client(cfg config.S3Config, log *zap.Logger) (*S3Client, error) {
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.Endpoint),
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	})

	// Verify connectivity by listing objects (max 1) in the bucket
	ctx := context.Background()
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 head bucket %q: %w", cfg.Bucket, err)
	}

	log.Info("s3 connected",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.Bucket),
	)

	return &S3Client{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
		log:       log,
	}, nil
}

// Upload uploads data to S3 at the given key with the specified content type.
func (s *S3Client) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 put object %q: %w", key, err)
	}

	url := s.publicURL + "/" + key
	s.log.Debug("s3 uploaded", zap.String("key", key), zap.Int("size", len(data)))
	return url, nil
}

// Delete removes a single object from S3.
func (s *S3Client) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete object %q: %w", key, err)
	}

	s.log.Debug("s3 deleted", zap.String("key", key))
	return nil
}

// DeletePrefix removes all objects with the given prefix (e.g. "products/1/uuid").
func (s *S3Client) DeletePrefix(ctx context.Context, prefix string) error {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("s3 list objects prefix %q: %w", prefix, err)
		}
		for _, obj := range page.Contents {
			if err := s.Delete(ctx, *obj.Key); err != nil {
				return err
			}
		}
	}
	return nil
}

// UploadFromReader uploads data from an io.Reader to S3.
func (s *S3Client) UploadFromReader(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read upload data: %w", err)
	}
	return s.Upload(ctx, key, data, contentType)
}

// PublicURL returns the full public URL for a given S3 key.
func (s *S3Client) PublicURL(key string) string {
	return s.publicURL + "/" + key
}

// KeyFromPath constructs a key from path parts, e.g. ("products", "1", "abc.jpg") → "products/1/abc.jpg".
func KeyFromPath(parts ...string) string {
	return path.Join(parts...)
}
