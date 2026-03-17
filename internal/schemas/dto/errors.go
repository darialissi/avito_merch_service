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
)
