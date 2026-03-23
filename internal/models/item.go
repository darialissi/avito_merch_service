package models

import (
	"github.com/google/uuid"
)

type Item struct {
	ID    uuid.UUID `db:"id"`
	Name  string    `db:"name"`
	Price float64   `db:"price"`
}

type UserItem struct {
	UserID   uuid.UUID `db:"user_id"`
	ItemID   uuid.UUID `db:"item_id"`
	Quantity int       `db:"quantity"`
}

type UserItemExtended struct {
	UserID   uuid.UUID `db:"user_id"`
	ItemID   uuid.UUID `db:"item_id"`
	Quantity int       `db:"quantity"`
	ItemName string    `db:"name"`
}
