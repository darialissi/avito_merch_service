package dto

import (
	"errors"
)

// Ошибки валидации
var (
	// Невалидный Password
	ErrInvalidPassword = errors.New("Password must contain at least 10 symbols with at least 1 lower && 1 upper && 1 digit")
	// Невалидный Username
	ErrInvalidUsername = errors.New("Username can contain lower|upper|digit|undescore and the length should be 3-15 symbols")
	// Невалидный Amount
	ErrInvalidAmount = errors.New("Amount must be greater than 0")
	// Невалидное Quantity
	ErrInvalidQuantity = errors.New("Quantity must be greater than 0")
	// Невозможно отправить монеты самому себе
	ErrTransactionYourself = errors.New("You cannot send coins to yourself")
)
