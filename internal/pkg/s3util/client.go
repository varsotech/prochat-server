package s3util

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"
)

func NewClient(ctx context.Context, s3Bucket string) (*s3.Client, error) {
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
		Bucket: aws.String(s3Bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed testing head bucket '%s': %w", s3Bucket, err)
	}

	return client, nil
}
