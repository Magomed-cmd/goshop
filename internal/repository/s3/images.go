package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type ImgStorage struct {
	client *minio.Client
	bucket string
	region string
}

func NewImgStorage(client *minio.Client, bucket string, region string) *ImgStorage {
	return &ImgStorage{
		client: client,
		bucket: bucket,
		region: region,
	}
}

func (st *ImgStorage) UploadImage(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (*string, error) {
	_, err := st.client.PutObject(ctx, st.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}
	url := fmt.Sprintf("https://%s/%s/%s", st.client.EndpointURL().Host, st.bucket, objectName)
	return &url, nil
}

func (st *ImgStorage) DeleteImage(ctx context.Context, objectName string) error {
	return st.client.RemoveObject(ctx, st.bucket, objectName, minio.RemoveObjectOptions{})
}

func (st *ImgStorage) GetImageURL(objectName string) string {
	return fmt.Sprintf("https://%s/%s/%s", st.client.EndpointURL().Host, st.bucket, objectName)
}

func (st *ImgStorage) DownloadImage(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := st.client.GetObject(ctx, st.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
