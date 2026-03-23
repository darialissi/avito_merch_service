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
	// Недостаточно монет для транзакции
	ErrNotEnoughCoins = errors.New("Not enough coins")
	// Товар не найден
	ErrItemNotFound = errors.New("Item not found")
	// Пользователь не найден
	ErrUserNotFound = errors.New("User not found")
)
