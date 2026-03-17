package auth

import (
	"errors"
)

var (
	ErrUniqueConflict = errors.New("Unique key conflict")
)
