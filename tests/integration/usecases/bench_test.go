package usecases

import (
	"fmt"
	"testing"
	"time"

	authRepo "github.com/darialissi/avito_merch_service/internal/repositories/auth"
	shopRepo "github.com/darialissi/avito_merch_service/internal/repositories/shop"
	tokenStorage "github.com/darialissi/avito_merch_service/internal/repositories/token"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	uc "github.com/darialissi/avito_merch_service/internal/usecases"
	authUtils "github.com/darialissi/avito_merch_service/internal/utils/auth"
	"github.com/darialissi/avito_merch_service/lib/postgres"
	"github.com/darialissi/avito_merch_service/tests/testutil"
	"github.com/stretchr/testify/require"
)

func setupUsecases(b *testing.B) (*testutil.Env, *uc.AuthUsecase, *uc.ShopUsecase) {
	b.Helper()
	testutil.SilenceStandardLogger(b)

	env := testutil.SetupEnv(b)
	tm := postgres.New(env.Pool)

	accessTTL, err := time.ParseDuration(env.Cfg.App.JWTConfig.AccessToken.Exp)
	require.NoError(b, err)

	refreshTTL, err := time.ParseDuration(env.Cfg.App.JWTConfig.RefreshToken.Exp)
	require.NoError(b, err)

	tokenHelper := authUtils.NewJWTHelper(
		env.Cfg.App.JWTConfig.AccessToken.Secret,
		env.Cfg.App.JWTConfig.RefreshToken.Secret,
		accessTTL,
		refreshTTL,
	)
	passwordHelper := authUtils.NewPasswordHelper()

	authRepository := authRepo.NewAuthRepository(tm)
	storage := tokenStorage.NewTokenStorage(env.RDB, refreshTTL)
	authUsecase := uc.NewAuthUsecase(authRepository, storage, tokenHelper, passwordHelper)

	shopRepository := shopRepo.NewShopRepository(tm)
	shopUsecase := uc.NewShopUsecase(shopRepository, tm)

	return env, authUsecase, shopUsecase
}

func BenchmarkAuthSignIn(b *testing.B) {
	env, authUsecase, _ := setupUsecases(b)

	forms := make([]dto.AuthForm, b.N)
	for i := 0; i < b.N; i++ {
		forms[i] = dto.AuthForm{
			Username: fmt.Sprintf("bench_signup_%d", i),
			Password: "StrongPass01",
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authUsecase.SignIn(env.Ctx, &forms[i])
		require.NoError(b, err)
	}
}

func BenchmarkAuthLogIn(b *testing.B) {
	env, authUsecase, _ := setupUsecases(b)

	form := dto.AuthForm{Username: "bench_login", Password: "StrongPass01"}
	_, err := authUsecase.SignIn(env.Ctx, &form)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authUsecase.LogIn(env.Ctx, &form)
		require.NoError(b, err)
	}
}

func BenchmarkShopBuyItem(b *testing.B) {
	env, _, shopUsecase := setupUsecases(b)

	_, err := env.Pool.Exec(env.Ctx, `
		INSERT INTO users (username, hashed_password)
		SELECT 'bench_buyer_' || g::text, 'hash'
		FROM generate_series(1, $1) AS g
	`, b.N)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := fmt.Sprintf("bench_buyer_%d", i+1)
		err := shopUsecase.BuyItem(env.Ctx, username, &dto.BuyItemData{ItemName: "pen", Quantity: 1})
		require.NoError(b, err)
	}
}

func BenchmarkShopSendCoin(b *testing.B) {
	env, _, shopUsecase := setupUsecases(b)

	_, err := env.Pool.Exec(env.Ctx, `
		INSERT INTO users (username, hashed_password)
		SELECT 'bench_sender_' || g::text, 'hash'
		FROM generate_series(1, $1) AS g
	`, b.N)
	require.NoError(b, err)

	_, err = env.Pool.Exec(env.Ctx, `
		INSERT INTO users (username, hashed_password)
		SELECT 'bench_receiver_' || g::text, 'hash'
		FROM generate_series(1, $1) AS g
	`, b.N)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := shopUsecase.SendCoin(env.Ctx, fmt.Sprintf("bench_sender_%d", i+1), &dto.TransactionData{
			ToUser: fmt.Sprintf("bench_receiver_%d", i+1),
			Amount: 100,
		})
		require.NoError(b, err)
	}
}

func BenchmarkShopInfo(b *testing.B) {
	env, _, shopUsecase := setupUsecases(b)

	_, err := env.Pool.Exec(env.Ctx, `
		INSERT INTO users (username, hashed_password)
		VALUES ('bench_info_user', 'hash'), ('bench_info_other', 'hash')
	`)
	require.NoError(b, err)

	_, err = env.Pool.Exec(env.Ctx, `
		INSERT INTO user_items (user_id, item_id, quantity)
		SELECT u.id, i.id, 3
		FROM users u
		JOIN items i ON i.name IN ('pen', 'book', 'cup')
		WHERE u.username = 'bench_info_user'
	`)
	require.NoError(b, err)

	_, err = env.Pool.Exec(env.Ctx, `
		INSERT INTO transactions (from_user_id, to_user_id, coins)
		SELECT u1.id, u2.id, 5
		FROM users u1
		JOIN users u2 ON u1.username = 'bench_info_user' AND u2.username = 'bench_info_other'
		CROSS JOIN generate_series(1, 50)
	`)
	require.NoError(b, err)

	_, err = env.Pool.Exec(env.Ctx, `
		INSERT INTO transactions (from_user_id, to_user_id, coins)
		SELECT u2.id, u1.id, 7
		FROM users u1
		JOIN users u2 ON u1.username = 'bench_info_user' AND u2.username = 'bench_info_other'
		CROSS JOIN generate_series(1, 50)
	`)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := shopUsecase.Info(env.Ctx, "bench_info_user")
		require.NoError(b, err)
	}
}
