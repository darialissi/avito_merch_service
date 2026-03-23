package models

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	ID        uuid.UUID `db:"id"`
	FromUser  string    `db:"from_user_id"`
	ToUser    string    `db:"to_user_id"`
	Coins     float64   `db:"coins"`
	CreatedAt time.Time `db:"created_at"`
}
