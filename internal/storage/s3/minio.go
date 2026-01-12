package s3

import (
	"context"
	"io"
	"time"

	"github.com/Sesame2/go-admin/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	*minio.Client
	config *config.Config
}

func NewMinioClient(config *config.Config) (*MinioClient, error) {
	endPoint := config.S3.Endpoint
	accessKey := config.S3.AccessKey
	secretKey := config.S3.SecretKey

	minioClient, err := minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{
		minioClient,
		config,
	}, nil
}

func (c *MinioClient) SignURL(ctx context.Context, bucket, objectName string, expire time.Duration) (string, error) {
	url, err := c.Client.PresignedGetObject(ctx, bucket, objectName, expire, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (c *MinioClient) Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64) error {
	_, err := c.Client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{})
	return err
}
