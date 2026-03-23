package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/darialissi/avito_merch_service/internal/repositories/auth"
	"github.com/darialissi/avito_merch_service/internal/repositories/token"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	utils "github.com/darialissi/avito_merch_service/internal/utils/auth"
)

type AuthUsecase struct {
	repo      *auth.AuthRepository
	storage   *token.TokenStorage
	jwtHelper *utils.JWTHelper
}

func NewAuthUsecase(repo *auth.AuthRepository, storage *token.TokenStorage, jwtHelper *utils.JWTHelper) *AuthUsecase {
	return &AuthUsecase{
		repo:      repo,
		storage:   storage,
		jwtHelper: jwtHelper,
	}
}

type AuthUsecases interface {
	// Регистрация пользователя
	SignIn(ctx context.Context, form *dto.AuthForm) (*dto.UserResponse, error)
	// Аутентификация пользователя и выдача токенов
	LogIn(ctx context.Context, form *dto.AuthForm) (*dto.AuthResponse, error)
}

// Проверка реализации всех методов интерфейса при компиляции
var _ AuthUsecases = (*AuthUsecase)(nil)

func (ac *AuthUsecase) SignIn(ctx context.Context, form *dto.AuthForm) (*dto.UserResponse, error) {

	// 0. Хешировать пароль.
	hashed, err := utils.HashPassword(form.Password)
	if err != nil {
		return nil, fmt.Errorf("HashPassword error: %w", err)
	}

	// 1. Сохранить пользователя.
	data := &dto.UserForm{
		Username:       form.Username,
		HashedPassword: hashed,
	}

	saved, err := ac.repo.SaveUser(ctx, data)
	if err != nil {
		if errors.Is(err, auth.ErrUniqueConflict) {
			return nil, ErrUsernameAlreadyExists
		}
		return nil, fmt.Errorf("SaveUser error: %w", err)
	}

	// 2. Сформировать ответ.
	response := &dto.UserResponse{
		ID:        saved.ID,
		Username:  saved.Username,
		Coins:     saved.Coins,
		CreatedAt: saved.CreatedAt,
	}

	return response, nil
}

func (ac *AuthUsecase) LogIn(ctx context.Context, form *dto.AuthForm) (*dto.AuthResponse, error) {

	// 1. Проверить наличие пользователя в БД.
	user, err := ac.repo.GetUserByUsername(ctx, form.Username)

	if user == nil {
		return nil, ErrNotExistedUser
	}

	// 2. Проверить корректность пароля.
	isValid, err := utils.VerifyPassword(form.Password, user.HashedPassword)
	if err != nil {
		return nil, fmt.Errorf("VerifyPassword error: %w", err)
	}

	if !isValid {
		return nil, ErrIncorrectPassword
	}

	// 3. Сгенерировать токены авторизации.
	access, err := ac.jwtHelper.CreateAccessToken(form.Username, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("CreateAccessToken error: %w", err)
	}

	refresh, err := ac.jwtHelper.CreateRefreshToken(form.Username, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("CreateRefreshToken error: %w", err)
	}

	// 3.1. Refresh токен сохранить в хранилище.
	tokens := &dto.TokenStore{
		Username: form.Username,
		Refresh:  refresh,
	}
	if err := ac.storage.SetToken(ctx, tokens); err != nil {
		return nil, fmt.Errorf("SetToken error: %w", err)
	}

	// 4. Сформировать ответ.
	response := &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}

	return response, nil
}
