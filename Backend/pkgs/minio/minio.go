package minio

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/pkgs/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type IUploadService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	UploadFileFromPath(ctx context.Context, localPath, folder, contentType string) (string, error)
	GetPresignedURL(ctx context.Context, fileURL string, expiry time.Duration) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}

type MinioClient struct {
	Client        *minio.Client
	Bucket        string
	BaseURL       string
	PublicBaseURL string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucket, baseURL, publicBaseURL string, useSSL bool) (*MinioClient, error) {
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
		Client:        client,
		Bucket:        bucket,
		BaseURL:       baseURL,
		PublicBaseURL: publicBaseURL,
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

// UploadFileFromPath upload file từ đường dẫn local lên MinIO
func (m *MinioClient) UploadFileFromPath(ctx context.Context, localPath, folder, contentType string) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", err
	}

	if folder == "" {
		folder = "uploads"
	}
	objectName := fmt.Sprintf("%s/%d-%s", folder, time.Now().UnixNano(), filepath.Base(localPath))

	_, err = m.Client.PutObject(ctx, m.Bucket, objectName, f, stat.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", m.BaseURL, m.Bucket, objectName), nil
}

// GetPresignedURL tạo URL tạm thời để download file trực tiếp từ MinIO
func (m *MinioClient) GetPresignedURL(ctx context.Context, fileURL string, expiry time.Duration) (string, error) {
	objectName := extractFilePath(fileURL, m.BaseURL, m.Bucket)
	presignedURL, err := m.Client.PresignedGetObject(ctx, m.Bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	
	// Thay thế internal endpoint bằng public endpoint cho client truy cập
	urlStr := presignedURL.String()
	if m.PublicBaseURL != "" && m.PublicBaseURL != m.BaseURL {
		urlStr = strings.Replace(urlStr, m.BaseURL, m.PublicBaseURL, 1)
	}
	
	return urlStr, nil
}

func (m *MinioClient) DeleteFile(ctx context.Context, fileURL string) error {
	objectName := extractFilePath(fileURL, m.BaseURL, m.Bucket)
	return m.Client.RemoveObject(ctx, m.Bucket, objectName, minio.RemoveObjectOptions{})
}

func extractFilePath(fileURL, baseURL, bucket string) string {
	return strings.TrimPrefix(fileURL, fmt.Sprintf("%s/%s/", baseURL, bucket))
}
