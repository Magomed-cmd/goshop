package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	BcryptCost = 14
)

func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func ValidatePassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}
