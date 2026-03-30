package testutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	ctrl "github.com/darialissi/avito_merch_service/internal/controllers/http"
	mw "github.com/darialissi/avito_merch_service/internal/middleware"
	authRepo "github.com/darialissi/avito_merch_service/internal/repositories/auth"
	shopRepo "github.com/darialissi/avito_merch_service/internal/repositories/shop"
	tokenStorage "github.com/darialissi/avito_merch_service/internal/repositories/token"
	uc "github.com/darialissi/avito_merch_service/internal/usecases"
	authUtils "github.com/darialissi/avito_merch_service/internal/utils/auth"
	"github.com/darialissi/avito_merch_service/lib/config"
	"github.com/darialissi/avito_merch_service/lib/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Env struct {
	Ctx  context.Context
	Cfg  config.Config
	Pool *pgxpool.Pool
	RDB  *redis.Client
}

func setupConfigPath(tb testing.TB) string {
	tb.Helper()

	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	_, filename, _, ok := runtime.Caller(0)
	require.True(tb, ok)

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	cfgPath := filepath.Join(repoRoot, "configs", ".config.test.yml")

	if _, err := os.Stat(cfgPath); err != nil {
		require.NoError(tb, err, "test config not found")
	}

	return cfgPath
}

func SetupEnv(tb testing.TB) *Env {
	tb.Helper()

	ctx := context.Background()
	cfgPath := setupConfigPath(tb)

	cfg, err := config.GetConfig(cfgPath)
	require.NoError(tb, err)

	pool, err := cfg.DB.CreatePool(ctx)
	require.NoError(tb, err)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})
	require.NoError(tb, rdb.Ping(ctx).Err())

	env := &Env{Ctx: ctx, Cfg: cfg, Pool: pool, RDB: rdb}

	ResetState(tb, env)
	tb.Cleanup(func() {
		ResetState(tb, env)
		rdb.Close()
		pool.Close()
	})

	return env
}

func ResetState(tb testing.TB, env *Env) {
	tb.Helper()

	_, err := env.Pool.Exec(env.Ctx, `
		TRUNCATE TABLE transactions, user_items, users RESTART IDENTITY CASCADE;
	`)
	require.NoError(tb, err)

	require.NoError(tb, env.RDB.FlushDB(env.Ctx).Err())
}

func BuildHTTPHandler(tb testing.TB, env *Env) http.Handler {
	tb.Helper()

	txMngr := postgres.New(env.Pool)

	accessTTL, err := time.ParseDuration(env.Cfg.App.JWTConfig.AccessToken.Exp)
	require.NoError(tb, err)

	refreshTTL, err := time.ParseDuration(env.Cfg.App.JWTConfig.RefreshToken.Exp)
	require.NoError(tb, err)

	tokenHelper := authUtils.NewJWTHelper(
		env.Cfg.App.JWTConfig.AccessToken.Secret,
		env.Cfg.App.JWTConfig.RefreshToken.Secret,
		accessTTL,
		refreshTTL,
	)
	passwordHelper := authUtils.NewPasswordHelper()

	authRepository := authRepo.NewAuthRepository(txMngr)
	storage := tokenStorage.NewTokenStorage(env.RDB, refreshTTL)
	authUsecase := uc.NewAuthUsecase(authRepository, storage, tokenHelper, passwordHelper)
	authHandler := ctrl.NewAuthHandler(authUsecase)

	shopRepository := shopRepo.NewShopRepository(txMngr)
	shopUsecase := uc.NewShopUsecase(shopRepository, txMngr)
	shopHandler := ctrl.NewShopHandler(shopUsecase)

	r := chi.NewRouter()
	r.Use(chimw.CleanPath)
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)

	authMw := mw.NewAuthMiddleware(tokenHelper, storage)
	requestTimeout, err := time.ParseDuration(env.Cfg.App.Timeout.Request)
	require.NoError(tb, err)
	r.Use(chimw.Timeout(requestTimeout))

	r.Route("/api", func(r chi.Router) {
		r.Post("/register", authHandler.RegisterUser)
		r.Post("/auth", authHandler.AuthUser)
		r.Group(func(r chi.Router) {
			r.Use(authMw.RequireAuth)
			r.Get("/info", shopHandler.Info)
			r.Post("/buy/{item}", shopHandler.BuyItem)
			r.Post("/sendCoin", shopHandler.SendCoin)
		})
	})

	return r
}
