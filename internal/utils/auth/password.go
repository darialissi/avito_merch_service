package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHelper struct{}

func NewPasswordHelper() *PasswordHelper {
	return &PasswordHelper{}
}

func (ph *PasswordHelper) HashPassword(password string) (string, error) {
	const cost = 12

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (ph *PasswordHelper) VerifyPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, fmt.Errorf("bcrypt compare failed: %w", err)
}
