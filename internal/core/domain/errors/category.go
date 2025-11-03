package errors

import "errors"

var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryNameExists  = errors.New("category name already exists")
	ErrInvalidCategoryData = errors.New("invalid category data")
)
