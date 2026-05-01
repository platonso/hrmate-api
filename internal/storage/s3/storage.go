package s3

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/platonso/hrmate-api/internal/config"
)

type Storage struct {
	client     *minio.Client
	bucketName string
}

func New(ctx context.Context, cfg *config.MinIOConfig) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.BucketName, err)
		}
		log.Printf("Successfully created bucket %s", cfg.BucketName)
	}

	return &Storage{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (s *Storage) UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", objectKey, err)
	}
	return nil
}

func (s *Storage) DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download object %s: %w", objectKey, err)
	}

	// Verify the object exists and is accessible
	_, err = object.Stat()
	if err != nil {
		_ = object.Close()
		return nil, fmt.Errorf("failed to stat object %s: %w", objectKey, err)
	}

	return object, nil
}

func (s *Storage) DeleteFile(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", objectKey, err)
	}
	return nil
}
