package vo

import (
	"net/mail"
	"strings"

	domainerrors "goshop/internal/core/domain/errors"
)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return Email{}, domainerrors.ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized {
		return Email{}, domainerrors.ErrInvalidEmail
	}

	return Email{value: normalized}, nil
}

func (e Email) String() string {
	return e.value
}
