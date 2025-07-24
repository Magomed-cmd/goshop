package domain_errors

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("access forbidden")
	ErrDuplicate           = errors.New("duplicate entry")
	ErrInvalidInput        = errors.New("invalid input data")
	ErrInvalidAddressData  = errors.New("invalid address data")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailExists         = errors.New("email already exists")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrAddressNotFound     = errors.New("address not found")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryNameExists  = errors.New("category name already exists")
	ErrInvalidCategoryData = errors.New("invalid category data")
	ErrProductNotFound     = errors.New("product not found")
	ErrProductNameExists   = errors.New("product name already exists")
	ErrInvalidProductData  = errors.New("invalid product data")
	ErrInvalidPrice        = errors.New("invalid price value")
	ErrInvalidStock        = errors.New("invalid stock value")
)
