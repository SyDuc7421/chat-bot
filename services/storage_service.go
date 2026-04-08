package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"hsduc.com/rag/config"
	"hsduc.com/rag/database"
)

func UploadFile(ctx context.Context, objectKey, contentType string, reader io.Reader, size int64) error {
	_, err := database.Minio.PutObject(ctx, config.App.MinioBucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func DeleteFile(ctx context.Context, objectKey string) error {
	return database.Minio.RemoveObject(ctx, config.App.MinioBucket, objectKey, minio.RemoveObjectOptions{})
}

// GetPresignedURL returns a temporary download URL valid for the given duration.
func GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	u, err := database.Minio.PresignedGetObject(ctx, config.App.MinioBucket, objectKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// BuildObjectKey generates a unique object key for a document file.
func BuildObjectKey(documentID uint, fileName string) string {
	return fmt.Sprintf("documents/%d/%d_%s", documentID, time.Now().UnixNano(), fileName)
}
