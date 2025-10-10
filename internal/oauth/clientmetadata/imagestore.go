package clientmetadata

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ImageStore struct {
	s3Client *s3.Client
}

func NewImageStore(s3Client *s3.Client) *ImageStore {
	return &ImageStore{
		s3Client: s3Client,
	}
}

func (i *ImageStore) Store(ctx context.Context, clientId, url string) (string, error) {
	return "", fmt.Errorf("todo: implement") // TODO
}
