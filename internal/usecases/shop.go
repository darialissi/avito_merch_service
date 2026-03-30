package usecases

import (
	"context"
	"errors"
	"github.com/darialissi/avito_merch_service/internal/models"
	"github.com/darialissi/avito_merch_service/internal/repositories/shop"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/google/uuid"
)

type ShopUsecase struct {
	repo ShopRepository
	tm   TransactionManager
}

func NewShopUsecase(repo ShopRepository, tm TransactionManager) *ShopUsecase {
	return &ShopUsecase{
		repo: repo,
		tm:   tm,
	}
}

//go:generate mockgen -source=shop.go -destination=../mocks/shop_mock.go -package=mocks
type ShopRepository interface {
	// Получить монеты пользователей по username с опциональной блокировкой строк для обновления
	GetUsersCoinsByUsernames(ctx context.Context, usernames []string, forUpdate bool) ([]models.User, error)
	// Обновить количество монет пользователей по username
	UpdateUserCoinsByUsername(ctx context.Context, userCoins *dto.UserCoins) (*models.User, error)
	// Получить товар по наименованию
	GetItemByName(ctx context.Context, name string) (*models.Item, error)
	// Обновить или создать запись инвентаря пользователя
	UpsertUserItemQuantity(ctx context.Context, data *dto.UserItemData) (*models.UserItem, error)
	// Получить инвентарь пользователя по userID с наименованиями товаров
	GetUserItemsByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserItemExtended, error)
	// Получить транзакции пользователя по userID
	GetTransactionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	// Сохранить транзакцию
	SaveTransaction(ctx context.Context, data *dto.TransactionFullData) (*models.Transaction, error)
}

type TransactionManager interface {
	RunReadCommitted(ctx context.Context, f func(txCtx context.Context) error) error
	RunRepeatableRead(ctx context.Context, f func(txCtx context.Context) error) error
	RunSerializable(ctx context.Context, f func(txCtx context.Context) error) error
}

type ShopUsecases interface {
	// Отправка монет от одного пользователя другому
	SendCoin(ctx context.Context, username string, data *dto.TransactionData) error
	// Покупка товара на монеты
	BuyItem(ctx context.Context, username string, data *dto.BuyItemData) error
	// Получение агрегированной информации о пользователе (монеты, инвентарь, история транзакций)
	Info(ctx context.Context, username string) (*dto.AggregatedInfo, error)
}

// Проверка реализации всех методов интерфейса при компиляции
var _ ShopUsecases = (*ShopUsecase)(nil)

func (sc *ShopUsecase) SendCoin(ctx context.Context, username string, data *dto.TransactionData) error {

	// TRANSACTION SCOPE
	err := sc.tm.RunRepeatableRead(ctx, func(txCtx context.Context) error {

		// 1. Получить данные отправителя и получателя с блокировкой строк для обновления монет
		users, err := sc.repo.GetUsersCoinsByUsernames(txCtx, []string{username, data.ToUser}, true)
		if err != nil {
			return err
		}

		if len(users) < 2 {
			return ErrUserNotFound
		}

		// Определить, кто из полученных пользователей является отправителем, а кто получателем
		sender, receiver := users[0], users[1]
		if sender.Username != username {
			sender, receiver = receiver, sender
		}

		// 2. Проверить, что у отправителя достаточно монет
		if sender.Coins < data.Amount {
			return ErrNotEnoughCoins
		}

		// 4. Списать монеты у отправителя и добавить монеты получателю
		sender.Coins -= data.Amount
		receiver.Coins += data.Amount

		// 5. Сохранить изменения в БД
		if _, err := sc.repo.UpdateUserCoinsByUsername(txCtx, &dto.UserCoins{Username: sender.Username, Coins: sender.Coins}); err != nil {
			return err
		}
		if _, err := sc.repo.UpdateUserCoinsByUsername(txCtx, &dto.UserCoins{Username: receiver.Username, Coins: receiver.Coins}); err != nil {
			return err
		}

		if _, err := sc.repo.SaveTransaction(txCtx, &dto.TransactionFullData{
			FromUserID: sender.ID,
			ToUserID:   receiver.ID,
			Amount:     data.Amount,
		}); err != nil {
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
		users, err := sc.repo.GetUsersCoinsByUsernames(txCtx, []string{username}, true)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			return ErrUserNotFound
		}

		user := users[0]

		// Проверить, что у пользователя достаточно монет
		totalPrice := item.Price * float64(data.Quantity)
		if user.Coins < totalPrice {
			return ErrNotEnoughCoins
		}

		// Списать монеты у пользователя
		user.Coins -= totalPrice

		// Сохранить изменения в БД
		if _, err := sc.repo.UpdateUserCoinsByUsername(txCtx, &dto.UserCoins{Username: user.Username, Coins: user.Coins}); err != nil {
			return err
		}

		record := &dto.UserItemData{
			UserID:   user.ID,
			ItemID:   item.ID,
			Quantity: data.Quantity,
		}

		// Обновить или создать запись инвентаря пользователя
		if _, err := sc.repo.UpsertUserItemQuantity(txCtx, record); err != nil {
			return err
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
	users, err := sc.repo.GetUsersCoinsByUsernames(ctx, []string{username}, false)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	user := users[0]

	// Получить данные инвентаря пользователя
	items, err := sc.repo.GetUserItemsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Получить транзакции пользователя
	transactions, err := sc.repo.GetTransactionsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	sent, received := make([]models.Transaction, 0), make([]models.Transaction, 0)
	for _, t := range transactions {
		if t.FromUser == user.ID {
			sent = append(sent, t)
		} else {
			received = append(received, t)
		}
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
			ToUser: t.ToUser.String(),
			Amount: t.Coins,
		}
	}
	receivedTransactions := make([]dto.ReceivedTransaction, len(received))
	for i, t := range received {
		receivedTransactions[i] = dto.ReceivedTransaction{
			FromUser: t.FromUser.String(),
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
