package vo

import domainerrors "goshop/internal/core/domain/errors"

type UserID int64
type AddressID int64

func NewUserID(id int64) (UserID, error) {
	if id <= 0 {
		return 0, domainerrors.ErrInvalidInput
	}
	return UserID(id), nil
}

func NewAddressID(id int64) (AddressID, error) {
	if id <= 0 {
		return 0, domainerrors.ErrInvalidInput
	}
	return AddressID(id), nil
}

func (id UserID) Int64() int64 {
	return int64(id)
}

func (id AddressID) Int64() int64 {
	return int64(id)
}
