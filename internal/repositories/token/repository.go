package token

import (
	"context"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/redis/go-redis/v9"
	"time"
)

type TokenStorage struct {
	db  *redis.Client
	ttl time.Duration
}

func NewTokenStorage(db *redis.Client, ttl time.Duration) *TokenStorage {
	return &TokenStorage{
		db:  db,
		ttl: ttl,
	}
}

func (st *TokenStorage) SetToken(ctx context.Context, data *dto.TokenStore) error {
	err := st.db.Set(ctx, data.Username, data.Refresh, st.ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (st *TokenStorage) GetToken(ctx context.Context, username string) (string, error) {
	refresh, err := st.db.Get(ctx, username).Result()
	if err != nil {
		return "", err
	}

	return refresh, nil
}
