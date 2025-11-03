package errors

import "errors"

var (
	ErrAddressNotFound    = errors.New("address not found")
	ErrInvalidAddressData = errors.New("invalid address data")
)
