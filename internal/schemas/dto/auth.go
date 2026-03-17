package dto

import (
	"github.com/google/uuid"
	"regexp"
	"time"
	"unicode"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,15}$`)

type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserForm struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
}

type TokenStore struct {
	Username string
	Refresh  string
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *AuthForm) Validate() error {
	if !usernameRegex.MatchString(a.Username) {
		return ErrInvalidUsername
	}
	if len(a.Password) < 10 {
		return ErrInvalidPassword
	}

	var lower, upper, digit bool

	for _, r := range []rune(a.Password) {
		if unicode.IsLower(r) {
			lower = true
		} else if unicode.IsUpper(r) {
			upper = true
		} else if unicode.IsDigit(r) {
			digit = true
		} else {
			return ErrInvalidPassword
		}
	}

	if !(lower && upper && digit) {
		return ErrInvalidPassword
	}

	return nil
}
