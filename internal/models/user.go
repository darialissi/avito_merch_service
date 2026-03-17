package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID             uuid.UUID `db:"id"`
	Username       string    `db:"username"`
	HashedPassword string    `db:"hashed_password"`
	CreatedAt      time.Time `db:"created_at"`
}
