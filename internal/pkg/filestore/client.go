package filestore

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client is a minimal wrapper around the S3 SDK that is simple to require in interfaces (doesn't return s3 structs).
// It is safe for concurrent use by multiple Go routines.
type S3Client struct {
	s3Client *s3.Client
	bucket   string
}

// NewS3Client creates an S3 client and validates that the bucket exists.
func NewS3Client(ctx context.Context, bucket string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed loading default config for s3: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := os.Getenv("AWS_S3_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true // Required for MinIO
	})

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed head bucket '%s': %w", bucket, err)
	}

	return &S3Client{
		s3Client: client,
		bucket:   bucket,
	}, nil
}

func (c *S3Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return out.Body, nil
}

func (c *S3Client) PutObject(ctx context.Context, key string, data io.Reader) (string, error) {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get object: %w", err)
	}
	return key, nil
}
