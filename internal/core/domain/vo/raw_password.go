package vo

import (
	"strings"

	domainerrors "goshop/internal/core/domain/errors"
)

type RawPassword struct {
	value string
}

func NewRawPassword(raw string) (RawPassword, error) {
	if strings.TrimSpace(raw) == "" {
		return RawPassword{}, domainerrors.ErrInvalidInput
	}
	return RawPassword{value: raw}, nil
}

func (p RawPassword) String() string {
	return p.value
}
