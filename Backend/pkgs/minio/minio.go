package minio

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"backend/pkgs/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type IUploadService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}

type MinioClient struct {
	Client  *minio.Client
	Bucket  string
	BaseURL string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucket, baseURL string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
		logger.Info("MinIO bucket created: %s", bucket)
	}

	logger.Info("MinIO connection established")

	return &MinioClient{
		Client:  client,
		Bucket:  bucket,
		BaseURL: baseURL,
	}, nil
}

func (m *MinioClient) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	if folder == "" {
		folder = "uploads"
	}

	objectName := fmt.Sprintf("%s/%d-%s", folder, time.Now().UnixNano(), file.Filename)

	_, err = m.Client.PutObject(ctx, m.Bucket, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", m.BaseURL, m.Bucket, objectName), nil
}

func (m *MinioClient) DeleteFile(ctx context.Context, fileURL string) error {
	objectName := extractFilePath(fileURL, m.BaseURL, m.Bucket)
	return m.Client.RemoveObject(ctx, m.Bucket, objectName, minio.RemoveObjectOptions{})
}

func extractFilePath(fileURL, baseURL, bucket string) string {
	return strings.TrimPrefix(fileURL, fmt.Sprintf("%s/%s/", baseURL, bucket))
}
