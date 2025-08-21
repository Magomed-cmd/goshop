package productImages

import (
	"context"
	"goshop/internal/domain/entities"
	"io"
)

type ProductImageService struct {
	// I will append some fields here later
}

func NewProductImageService() *ProductImageService {
	return &ProductImageService{}
}

func SaveProduct(ctx context.Context, reader io.ReadCloser, size, userID int64, contentType, extension string) (string, error) {

}
