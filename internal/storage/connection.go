package storage

import (
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

func NewS3Connection(endpoint, accessKey, secretKey string, useSSL bool, logger *zap.Logger) (*minio.Client, error) {

	logger.Info("Connecting to S3", zap.String("endpoint", endpoint), zap.Bool("useSSL", useSSL), zap.String("accessKey", accessKey), zap.String("secretKey", secretKey))
	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logger.Error("Failed to create S3 client", zap.Error(err))
		return nil, err
	}

	s3Client.TraceOn(os.Stdout)

	return s3Client, nil
}
