package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/darialissi/avito_merch_service/internal/mocks"
	"github.com/darialissi/avito_merch_service/internal/models"
	"github.com/darialissi/avito_merch_service/internal/repositories/auth"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/darialissi/avito_merch_service/internal/usecases"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

type saveUserResult struct {
	user *models.User
	err  error
}

type hashPasswordResult struct {
	hashed string
	err    error
}

type verifyPasswordResult struct {
	match bool
	err   error
}

type getUserByUsernameResult struct {
	user *models.User
	err  error
}

type createTokenResult struct {
	token string
	err   error
}

type setTokenResult struct {
	err error
}

func TestAuthUsecase_SignIn(t *testing.T) {
	userID := uuid.New()
	username := "test_username"
	password := "Passw0rd!"
	hashedPassword := "hashed...."
	coins := 1000.00
	createdAt := time.Now()

	form := &dto.AuthForm{Username: username, Password: password}
	userForm := &dto.UserForm{Username: username, HashedPassword: hashedPassword}

	hashErr := errors.New("hash error")
	repoErr := errors.New("repo error")

	tests := []struct {
		name     string
		mockUser *models.User
		hashResp hashPasswordResult
		saveResp saveUserResult
		wantResp *dto.UserResponse
		wantErr  error
	}{
		{
			name: "Success",
			hashResp: hashPasswordResult{
				hashed: hashedPassword,
			},
			saveResp: saveUserResult{
				user: &models.User{
					ID:        userID,
					Username:  username,
					Coins:     coins,
					CreatedAt: createdAt,
				},
			},
			wantResp: &dto.UserResponse{
				ID:        userID,
				Username:  username,
				Coins:     coins,
				CreatedAt: createdAt,
			},
		},
		{
			name: "HashPassword error",
			hashResp: hashPasswordResult{
				err: hashErr,
			},
			wantErr: hashErr,
		},
		{
			name: "SaveUser ErrUniqueConflict error",
			hashResp: hashPasswordResult{
				hashed: hashedPassword,
			},
			saveResp: saveUserResult{
				err: auth.ErrUniqueConflict,
			},
			wantErr: usecases.ErrUsernameAlreadyExists,
		},
		{
			name: "SaveUser error",
			hashResp: hashPasswordResult{
				hashed: hashedPassword,
			},
			saveResp: saveUserResult{
				err: repoErr,
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockAuthRepository(ctrl)
			passHelper := mocks.NewMockPasswordHelper(ctrl)

			passHelper.EXPECT().
				HashPassword(form.Password).
				Return(tt.hashResp.hashed, tt.hashResp.err)

			if tt.hashResp.err == nil {
				repo.EXPECT().
					SaveUser(gomock.Any(), userForm).
					Return(tt.saveResp.user, tt.saveResp.err)
			}

			uc := usecases.NewAuthUsecase(repo, nil, nil, passHelper)

			gotResp, err := uc.SignIn(context.Background(), form)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotResp.ID != tt.wantResp.ID {
				t.Fatalf("expected ID %v, got %v", tt.wantResp.ID, gotResp.ID)
			}
			if gotResp.Username != tt.wantResp.Username {
				t.Fatalf("expected Username %q, got %q", tt.wantResp.Username, gotResp.Username)
			}
			if gotResp.Coins != tt.wantResp.Coins {
				t.Fatalf("expected Coins %f, got %f", tt.wantResp.Coins, gotResp.Coins)
			}
			if !gotResp.CreatedAt.Equal(tt.wantResp.CreatedAt) {
				t.Fatalf("expected CreatedAt %v, got %v", tt.wantResp.CreatedAt, gotResp.CreatedAt)
			}
		})
	}
}

func TestAuthUsecase_LogIn(t *testing.T) {
	username := "test_username"
	password := "Passw0rd!"
	hashedPassword := "hashed...."

	accessToken := "access_token"
	refreshToken := "refresh_token"

	form := &dto.AuthForm{Username: username, Password: password}

	repoErr := errors.New("repo error")
	verifyErr := errors.New("verify error")
	createTokenErr := errors.New("create token error")
	storageErr := errors.New("storage error")

	tests := []struct {
		name                   string
		getUserResp            getUserByUsernameResult
		verifyResp             verifyPasswordResult
		createAccessTokenResp  createTokenResult
		createRefreshTokenResp createTokenResult
		setTokenResp           setTokenResult
		wantResp               *dto.AuthResponse
		wantErr                error
	}{
		{
			name: "Success",
			getUserResp: getUserByUsernameResult{
				user: &models.User{
					Username:       username,
					HashedPassword: hashedPassword,
				},
			},
			verifyResp: verifyPasswordResult{
				match: true,
			},
			createAccessTokenResp: createTokenResult{
				token: accessToken,
			},
			createRefreshTokenResp: createTokenResult{
				token: refreshToken,
			},
			setTokenResp: setTokenResult{
				err: nil,
			},
			wantResp: &dto.AuthResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		},
		{
			name: "GetUserByUsername error",
			getUserResp: getUserByUsernameResult{
				err: repoErr,
			},
			wantErr: repoErr,
		},
		{
			name: "GetUserByUsername nil user",
			getUserResp: getUserByUsernameResult{
				user: nil,
				err:  nil,
			},
			wantErr: usecases.ErrNotExistedUser,
		},
		{
			name: "VerifyPassword error",
			getUserResp: getUserByUsernameResult{
				user: &models.User{
					Username:       username,
					HashedPassword: hashedPassword,
				},
			},
			verifyResp: verifyPasswordResult{
				err: verifyErr,
			},
			wantErr: verifyErr,
		},
		{
			name: "VerifyPassword false",
			getUserResp: getUserByUsernameResult{
				user: &models.User{
					Username:       username,
					HashedPassword: hashedPassword,
				},
			},
			verifyResp: verifyPasswordResult{
				match: false,
			},
			wantErr: usecases.ErrIncorrectPassword,
		},
		{
			name: "CreateAccessToken error",
			getUserResp: getUserByUsernameResult{
				user: &models.User{
					Username:       username,
					HashedPassword: hashedPassword,
				},
			},
			verifyResp: verifyPasswordResult{
				match: true,
			},
			createAccessTokenResp: createTokenResult{
				err: createTokenErr,
			},
			wantErr: createTokenErr,
		},
		{
			name: "CreateRefreshToken error",
			getUserResp: getUserByUsernameResult{
				user: &models.User{
					Username:       username,
					HashedPassword: hashedPassword,
				},
			},
			verifyResp: verifyPasswordResult{
				match: true,
			},
			createAccessTokenResp: createTokenResult{
				token: accessToken,
			},
			createRefreshTokenResp: createTokenResult{
				err: createTokenErr,
			},
			wantErr: createTokenErr,
		},
		{
			name: "SetToken error",
			getUserResp: getUserByUsernameResult{
				user: &models.User{
					Username:       username,
					HashedPassword: hashedPassword,
				},
			},
			verifyResp: verifyPasswordResult{
				match: true,
			},
			createAccessTokenResp: createTokenResult{
				token: accessToken,
			},
			createRefreshTokenResp: createTokenResult{
				token: refreshToken,
			},
			setTokenResp: setTokenResult{
				err: storageErr,
			},
			wantErr: storageErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockAuthRepository(ctrl)
			storage := mocks.NewMockTokenStorage(ctrl)
			jwtHelper := mocks.NewMockJWTHelper(ctrl)
			passHelper := mocks.NewMockPasswordHelper(ctrl)

			repo.EXPECT().
				GetUserByUsername(gomock.Any(), form.Username).
				Return(tt.getUserResp.user, tt.getUserResp.err)

			if tt.getUserResp.err == nil && tt.getUserResp.user != nil {
				passHelper.EXPECT().
					VerifyPassword(form.Password, tt.getUserResp.user.HashedPassword).
					Return(tt.verifyResp.match, tt.verifyResp.err)

				if tt.verifyResp.err == nil && tt.verifyResp.match {
					jwtHelper.EXPECT().
						CreateAccessToken(form.Username, tt.getUserResp.user.ID.String()).
						Return(tt.createAccessTokenResp.token, tt.createAccessTokenResp.err)

					if tt.createAccessTokenResp.err == nil {
						jwtHelper.EXPECT().
							CreateRefreshToken(form.Username, tt.getUserResp.user.ID.String()).
							Return(tt.createRefreshTokenResp.token, tt.createRefreshTokenResp.err)

						if tt.createRefreshTokenResp.err == nil {
							storage.EXPECT().
								SetToken(gomock.Any(), &dto.TokenStore{
									Username: form.Username,
									Refresh:  tt.createRefreshTokenResp.token,
								}).
								Return(tt.setTokenResp.err)
						}
					}
				}
			}

			uc := usecases.NewAuthUsecase(repo, storage, jwtHelper, passHelper)

			gotResp, err := uc.LogIn(context.Background(), form)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotResp.AccessToken != tt.wantResp.AccessToken {
				t.Fatalf("expected AccessToken %q, got %q", tt.wantResp.AccessToken, gotResp.AccessToken)
			}
			if gotResp.RefreshToken != tt.wantResp.RefreshToken {
				t.Fatalf("expected RefreshToken %q, got %q", tt.wantResp.RefreshToken, gotResp.RefreshToken)
			}
		})
	}
}
