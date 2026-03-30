package dto

import (
	"github.com/google/uuid"
)

type TransactionData struct {
	ToUser string  `json:"toUser"`
	Amount float64 `json:"amount"`
}

func (t *TransactionData) Validate(fromUser string) error {
	if t.ToUser == fromUser {
		return ErrTransactionYourself
	}
	if t.Amount <= 0 {
		return ErrInvalidAmount
	}

	return nil
}

type TransactionFullData struct {
	FromUser string  `json:"fromUser"`
	ToUser   string  `json:"toUser"`
	Amount   float64 `json:"amount"`
}

type BuyItemRequest struct {
	Quantity int `json:"quantity"`
}

type BuyItemData struct {
	ItemName string `json:"item"`
	Quantity int    `json:"quantity"`
}

func (t *BuyItemData) Validate() error {
	if t.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	return nil
}

type InventoryUnit struct {
	ItemName string `json:"type"`
	Quantity int    `json:"quantity"`
}

type ReceivedTransaction struct {
	FromUser string  `json:"fromUser"`
	Amount   float64 `json:"amount"`
}

type SentTransaction struct {
	ToUser string  `json:"toUser"`
	Amount float64 `json:"amount"`
}

type CoinHistory struct {
	Sent     []SentTransaction     `json:"sent"`
	Received []ReceivedTransaction `json:"received"`
}

type AggregatedInfo struct {
	Coins       float64         `json:"coins"`
	Inventory   []InventoryUnit `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type UserItemData struct {
	UserID   uuid.UUID
	ItemID   uuid.UUID
	Quantity int
}

type UserCoins struct {
	Username string
	Coins    float64
}
