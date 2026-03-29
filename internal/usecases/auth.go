package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/darialissi/avito_merch_service/internal/models"
	"github.com/darialissi/avito_merch_service/internal/repositories/auth"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/google/uuid"
)

type AuthUsecase struct {
	repo       AuthRepository
	storage    TokenStorage
	jwtHelper  JWTHelper
	passHelper PasswordHelper
}

func NewAuthUsecase(repo AuthRepository, storage TokenStorage, jwtHelper JWTHelper, passHelper PasswordHelper) *AuthUsecase {
	return &AuthUsecase{
		repo:       repo,
		storage:    storage,
		jwtHelper:  jwtHelper,
		passHelper: passHelper,
	}
}

//go:generate mockgen -source=auth.go -destination=../mocks/auth_mock.go -package=mocks
type AuthRepository interface {
	// Сохранить нового пользователя
	SaveUser(ctx context.Context, data *dto.UserForm) (*models.User, error)
	// Получить пользователя по ID
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	// Получить пользователя по username
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type TokenStorage interface {
	// Сохранить токен
	SetToken(ctx context.Context, data *dto.TokenStore) error
	// Получить токен
	GetToken(ctx context.Context, username string) (string, error)
}

type JWTHelper interface {
	// Сгенерировать Access токен
	CreateAccessToken(username, userID string) (string, error)
	// Сгенерировать Refresh токен
	CreateRefreshToken(username, userID string) (string, error)
}

type PasswordHelper interface {
	// Захешировать пароль
	HashPassword(password string) (string, error)
	// Сравнить пароль и его хеш
	VerifyPassword(password, hashed string) (bool, error)
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
	hashed, err := ac.passHelper.HashPassword(form.Password)
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

	if err != nil {
		return nil, fmt.Errorf("GetUserByUsername error: %w", err)
	}

	if user == nil {
		return nil, ErrNotExistedUser
	}

	// 2. Проверить корректность пароля.
	isValid, err := ac.passHelper.VerifyPassword(form.Password, user.HashedPassword)
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
