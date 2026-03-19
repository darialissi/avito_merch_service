package middleware

import (
	"context"
	"github.com/darialissi/avito_merch_service/internal/repositories/token"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	utils "github.com/darialissi/avito_merch_service/internal/utils/auth"
	"net/http"
	"strings"
)

type contextKey string

const authUserKey contextKey = "auth_username"

var (
	ErrUserNotAuthenticated = "User is not authenticated"
)

type AuthMiddleware struct {
	jwtHelper *utils.JWTHelper
	storage   *token.TokenStorage
}

func NewAuthMiddleware(jwtHelper *utils.JWTHelper, storage *token.TokenStorage) *AuthMiddleware {
	return &AuthMiddleware{jwtHelper: jwtHelper, storage: storage}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Валидируем access.
		authHeader := r.Header.Get("Authorization")
		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		if accessToken == "" {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		// Если валиден, просто пропускаем запрос дальше.
		username, err := m.jwtHelper.ValidateAccessToken(accessToken)
		if err == nil && username != "" {
			ctx := context.WithValue(r.Context(), authUserKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 2. Валидируем refresh.
		refreshToken := r.Header.Get("X-Refresh-Token")
		if refreshToken == "" {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		username, err = m.jwtHelper.ValidateRefreshToken(refreshToken)
		if err != nil || username == "" {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		storedToken, err := m.storage.GetToken(r.Context(), username)
		if err != nil {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		if storedToken != refreshToken {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		// 3. Генерируем новую пару токенов.
		newAccessToken, err := m.jwtHelper.CreateAccessToken(username)
		if err != nil {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		newRefreshToken, err := m.jwtHelper.CreateRefreshToken(username)
		if err != nil {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		// 4. Устанавливаем новый refresh токен в хранилище.
		if err := m.storage.SetToken(r.Context(), &dto.TokenStore{
			Username: username,
			Refresh:  newRefreshToken,
		}); err != nil {
			http.Error(w, ErrUserNotAuthenticated, http.StatusUnauthorized)
			return
		}

		// 5. Отдаем новую пару токенов клиенту.
		w.Header().Set("Authorization", "Bearer "+newAccessToken)
		w.Header().Set("X-Refresh-Token", newRefreshToken)

		ctx := context.WithValue(r.Context(), authUserKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
