package storage

import (
	"context"
	"io"
)

type ImgStorage interface {
	UploadImage(ctx context.Context, objectName string, reader io.ReadCloser, size int64, contentType string) (*string, error)
	DeleteImage(ctx context.Context, objectName string) error
	GetImageURL(objectName string) string
	DownloadImage(ctx context.Context, objectName string) (io.ReadCloser, error)
}
