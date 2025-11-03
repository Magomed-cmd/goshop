package errors

import "errors"

var (
	ErrProductNotFound        = errors.New("product not found")
	ErrProductNameExists      = errors.New("product name already exists")
	ErrInvalidProductData     = errors.New("invalid product data")
	ErrInvalidPrice           = errors.New("price must be greater than 0")
	ErrInvalidStock           = errors.New("stock cannot be negative")
	ErrInvalidProductID       = errors.New("invalid product id")
	ErrProductImageNotFound   = errors.New("product image not found")
	ErrProductImageUploadFail = errors.New("failed to upload product image")
	ErrProductImageDeleteFail = errors.New("failed to delete product image")
)
