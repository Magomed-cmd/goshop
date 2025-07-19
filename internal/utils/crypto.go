package utils

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

const (
	BcryptCost = 14
)

// HashPassword хэширует пароль с помощью bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

// ValidatePassword проверяет соответствие пароля хэшу
func ValidatePassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}
