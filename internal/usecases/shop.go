package usecases

import (
	"context"
	"errors"
	"github.com/darialissi/avito_merch_service/internal/repositories/shop"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/darialissi/avito_merch_service/lib/postgres"
)

type ShopUsecase struct {
	repo *shop.ShopRepository
	tm   *postgres.TransactionManager
}

func NewShopUsecase(
	repo *shop.ShopRepository,
	tm *postgres.TransactionManager,
) *ShopUsecase {
	return &ShopUsecase{
		repo: repo,
		tm:   tm,
	}
}

type ShopUsecases interface {
	// Отправка монет от одного пользователя другому
	SendCoin(ctx context.Context, username string, data *dto.TransactionData) error
	// Покупка товара на монеты
	BuyItem(ctx context.Context, username string, data *dto.BuyItemData) error
	Info(ctx context.Context, username string) (*dto.AggregatedInfo, error)
}

// Проверка реализации всех методов интерфейса при компиляции
var _ ShopUsecases = (*ShopUsecase)(nil)

func (sc *ShopUsecase) SendCoin(ctx context.Context, username string, data *dto.TransactionData) error {

	// TRANSACTION SCOPE
	err := sc.tm.RunRepeatableRead(ctx, func(txCtx context.Context) error {

		// 1. Получить данные отправителя и получателя с блокировкой строк для обновления монет
		users, err := sc.repo.GetSenderReceiverCoins(ctx, username, data.ToUser, true)
		if err != nil {
			return err
		}

		if len(users) < 2 {
			return ErrUserNotFound
		}

		// Определить, кто из полученных пользователей является отправителем, а кто получателем
		sender, receiver := users[0], users[1]
		if sender.Username != username {
			sender = receiver
			receiver = sender
		}

		// 2. Проверить, что у отправителя достаточно монет
		if sender.Coins < data.Amount {
			return ErrNotEnoughCoins
		}

		// 4. Списать монеты у отправителя и добавить монеты получателю
		sender.Coins -= data.Amount
		receiver.Coins += data.Amount

		// 5. Сохранить изменения в БД
		if _, err := sc.repo.UpdateUserCoinsByUsername(txCtx, sender.Username, sender.Coins); err != nil {
			return err
		}
		if _, err := sc.repo.UpdateUserCoinsByUsername(txCtx, receiver.Username, receiver.Coins); err != nil {
			return err
		}

		if _, err := sc.repo.SaveTransaction(txCtx, sender.ID, receiver.ID, data.Amount); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (sc *ShopUsecase) BuyItem(ctx context.Context, username string, data *dto.BuyItemData) error {

	// TRANSACTION SCOPE
	err := sc.tm.RunRepeatableRead(ctx, func(txCtx context.Context) error {

		// Получить данные товара (в транзакции для дальнейшего ограничения по количеству)
		item, err := sc.repo.GetItemByName(txCtx, data.ItemName)
		if err != nil {
			if errors.Is(err, shop.ErrNotFound) {
				return ErrItemNotFound
			}
			return err
		}

		// Получить данные пользователя с блокировкой строки
		user, err := sc.repo.GetUserCoinsByUsername(txCtx, username, true)
		if err != nil {
			return err
		}

		// Проверить, что у пользователя достаточно монет
		totalPrice := item.Price * float64(data.Quantity)
		if user.Coins < totalPrice {
			return ErrNotEnoughCoins
		}

		// Списать монеты у пользователя
		user.Coins -= totalPrice

		// Сохранить изменения в БД
		if _, err := sc.repo.UpdateUserCoinsByUsername(txCtx, user.Username, user.Coins); err != nil {
			return err
		}

		record, err := sc.repo.GetUserItem(txCtx, user.ID, item.ID, true)

		if err != nil {
			return err
		}

		if record == nil {
			// Если у пользователя нет этого товара, создать запись
			if _, err := sc.repo.SaveUserItem(txCtx, user.ID, item.ID, data.Quantity); err != nil {
				return err
			}
		} else {
			// Если товар уже есть, обновить количество
			newQuantity := record.Quantity + data.Quantity
			if _, err := sc.repo.UpdateUserItemQuantity(txCtx, user.ID, item.ID, newQuantity); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (sc *ShopUsecase) Info(ctx context.Context, username string) (*dto.AggregatedInfo, error) {

	// Получить данные пользователя
	user, err := sc.repo.GetUserCoinsByUsername(ctx, username, false)
	if err != nil {
		return nil, err
	}

	// Получить данные инвентаря пользователя
	items, err := sc.repo.GetUserItemsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Получить отправленные транзакции пользователя
	sent, err := sc.repo.GetSentTransactionsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Получить полученные транзакции пользователя
	received, err := sc.repo.GetReceivedTransactionsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Сформировать ответ
	inventory := make([]dto.InventoryUnit, len(items))
	for i, item := range items {
		inventory[i] = dto.InventoryUnit{
			ItemName: item.ItemName,
			Quantity: item.Quantity,
		}
	}
	sentTransactions := make([]dto.SentTransaction, len(sent))
	for i, t := range sent {
		sentTransactions[i] = dto.SentTransaction{
			ToUser: t.ToUser,
			Amount: t.Coins,
		}
	}
	receivedTransactions := make([]dto.ReceivedTransaction, len(received))
	for i, t := range received {
		receivedTransactions[i] = dto.ReceivedTransaction{
			FromUser: t.FromUser,
			Amount:   t.Coins,
		}
	}
	response := &dto.AggregatedInfo{
		Coins:     user.Coins,
		Inventory: inventory,
		CoinHistory: dto.CoinHistory{
			Sent:     sentTransactions,
			Received: receivedTransactions,
		},
	}

	return response, nil
}
