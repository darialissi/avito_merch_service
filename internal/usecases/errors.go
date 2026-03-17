package usecases

import (
	"errors"
)

// Ошибки бизнес логики
var (
	// Некорректный Password
	ErrIncorrectPassword = errors.New("Incorrect password")
	// Существующий Username
	ErrUsernameAlreadyExists = errors.New("Username already exists")
	// Несуществующий Username
	ErrNotExistedUser = errors.New("User does not exist")
)
