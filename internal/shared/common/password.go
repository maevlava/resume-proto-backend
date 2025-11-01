package common

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate bcrypt hash %w", err)
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) error {
	userPassword := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if userPassword != nil {
		return errors.New("password does not match")
	}
	return nil
}
