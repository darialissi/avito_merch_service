package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	ctrl "github.com/darialissi/avito_merch_service/internal/controllers/http"
	mw "github.com/darialissi/avito_merch_service/internal/middleware"
	auth_repo "github.com/darialissi/avito_merch_service/internal/repositories/auth"
	shop_repo "github.com/darialissi/avito_merch_service/internal/repositories/shop"
	token_storage "github.com/darialissi/avito_merch_service/internal/repositories/token"
	uc "github.com/darialissi/avito_merch_service/internal/usecases"
	auth_utils "github.com/darialissi/avito_merch_service/internal/utils/auth"

	"github.com/darialissi/avito_merch_service/lib/config"
	"github.com/darialissi/avito_merch_service/lib/postgres"
)

func main() {
	ctx := context.Background()

	// Get environment config
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("No defined CONFIG_PATH")
	}

	cfg, err := config.GetConfig(configPath)
	if err != nil {
		log.Fatalf("Error in loading config %v", err)
	}

	// Set db pool connections
	pool, err := cfg.DB.CreatePool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	txMngr := postgres.New(pool)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0, // use default DB
	})
	defer rdb.Close()

	// Init services
	accessTTL, _ := time.ParseDuration(cfg.App.JWTConfig.AccessToken.Exp)
	refreshTTL, _ := time.ParseDuration(cfg.App.JWTConfig.RefreshToken.Exp)
	TokenHelper := auth_utils.NewJWTHelper(cfg.App.JWTConfig.AccessToken.Secret, cfg.App.JWTConfig.RefreshToken.Secret, accessTTL, refreshTTL)

	AuthRepository := auth_repo.NewAuthRepository(txMngr)
	TokenStorage := token_storage.NewTokenStorage(rdb, refreshTTL)
	AuthUsecase := uc.NewAuthUsecase(AuthRepository, TokenStorage, TokenHelper)
	AuthHandler := ctrl.NewAuthHandler(AuthUsecase)

	ShopRepository := shop_repo.NewShopRepository(txMngr)
	ShopUsecase := uc.NewShopUsecase(ShopRepository, txMngr)
	ShopHandler := ctrl.NewShopHandler(ShopUsecase)

	// Register handlers
	r := chi.NewRouter()

	// Define middleware
	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	authMw := mw.NewAuthMiddleware(TokenHelper, TokenStorage)

	requestTimeout, _ := time.ParseDuration(cfg.App.Timeout.Request)
	r.Use(middleware.Timeout(requestTimeout))

	r.Route("/api", func(r chi.Router) {
		r.Post("/register", AuthHandler.RegisterUser)
		r.Post("/auth", AuthHandler.AuthUser)

		r.Group(func(r chi.Router) { // группа защищенных роутов
			r.Use(authMw.RequireAuth)

			r.Get("/info", ShopHandler.Info)
			r.Post("/buy/{item}", ShopHandler.BuyItem)
			r.Post("/sendCoin", ShopHandler.SendCoin)
		})
	})

	// Run server and shutdown gracefully
	port := cfg.App.Port
	server := http.Server{
		Handler:           r,
		Addr:              port,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server started at %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Error: %v", err)
		}
	}()

	sig := <-signalCh
	log.Printf("Received signal: %s\nStarted graceful shutdown...", sig)

	gracefulTimeout, _ := time.ParseDuration(cfg.App.Timeout.Graceful)
	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %s", err.Error())
	}
	log.Println("Server stopped")
}
