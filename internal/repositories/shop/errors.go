package shop

import (
	"errors"
)

var (
	ErrUniqueConflict = errors.New("Unique key conflict")
	ErrNotFound       = errors.New("Not found")
)
