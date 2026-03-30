package usecases

import (
	"testing"

	"github.com/stretchr/testify/require"

	shopRepo "github.com/darialissi/avito_merch_service/internal/repositories/shop"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	uc "github.com/darialissi/avito_merch_service/internal/usecases"
	"github.com/darialissi/avito_merch_service/lib/postgres"
	"github.com/darialissi/avito_merch_service/tests/testutil"
)

func TestShopUsecase_BuyItem_UpdatesCoinsAndInventory(t *testing.T) {
	env := testutil.SetupEnv(t)

	tm := postgres.New(env.Pool)
	repo := shopRepo.NewShopRepository(tm)
	usecase := uc.NewShopUsecase(repo, tm)

	username := "buyer_user"
	_, err := env.Pool.Exec(env.Ctx, `INSERT INTO users (username, hashed_password) VALUES ($1, $2)`, username, "hash")
	require.NoError(t, err)

	err = usecase.BuyItem(env.Ctx, username, &dto.BuyItemData{
		ItemName: "book",
		Quantity: 2,
	})
	require.NoError(t, err)

	var coins float64
	err = env.Pool.QueryRow(env.Ctx, `SELECT coins FROM users WHERE username = $1`, username).Scan(&coins)
	require.NoError(t, err)
	require.Equal(t, 900.0, coins)

	var quantity int
	err = env.Pool.QueryRow(env.Ctx, `
		SELECT ui.quantity
		FROM user_items ui
		JOIN users u ON u.id = ui.user_id
		JOIN items i ON i.id = ui.item_id
		WHERE u.username = $1 AND i.name = $2
	`, username, "book").Scan(&quantity)
	require.NoError(t, err)
	require.Equal(t, 2, quantity)
}

func TestShopUsecase_SendCoin_TransfersBalanceAndStoresTransaction(t *testing.T) {
	env := testutil.SetupEnv(t)

	tm := postgres.New(env.Pool)
	repo := shopRepo.NewShopRepository(tm)
	usecase := uc.NewShopUsecase(repo, tm)

	from := "sender_user"
	to := "receiver_user"

	_, err := env.Pool.Exec(env.Ctx, `
		INSERT INTO users (username, hashed_password)
		VALUES ($1, $3), ($2, $3)
	`, from, to, "hash")
	require.NoError(t, err)

	err = usecase.SendCoin(env.Ctx, from, &dto.TransactionData{
		ToUser: to,
		Amount: 250,
	})
	require.NoError(t, err)

	var senderCoins float64
	err = env.Pool.QueryRow(env.Ctx, `SELECT coins FROM users WHERE username = $1`, from).Scan(&senderCoins)
	require.NoError(t, err)
	require.Equal(t, 750.0, senderCoins)

	var receiverCoins float64
	err = env.Pool.QueryRow(env.Ctx, `SELECT coins FROM users WHERE username = $1`, to).Scan(&receiverCoins)
	require.NoError(t, err)
	require.Equal(t, 1250.0, receiverCoins)

	var txCount int
	err = env.Pool.QueryRow(env.Ctx, `
		SELECT COUNT(*)
		FROM transactions t
		JOIN users u1 ON u1.id = t.from_user_id
		JOIN users u2 ON u2.id = t.to_user_id
		WHERE u1.username = $1 AND u2.username = $2 AND t.coins = $3
	`, from, to, 250.0).Scan(&txCount)
	require.NoError(t, err)
	require.Equal(t, 1, txCount)
}
