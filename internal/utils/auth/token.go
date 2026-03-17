package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTHelper struct {
	accessSecret  string
	refreshSecret string
	accessExp     time.Duration
	refreshExp    time.Duration
}

func NewJWTHelper(accessSecret, refreshSecret string, accessExp, refreshExp time.Duration) *JWTHelper {
	return &JWTHelper{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessExp:     accessExp,
		refreshExp:    refreshExp,
	}
}

func (hp *JWTHelper) CreateAccessToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(hp.accessExp).Unix(),
		"iat": time.Now().Unix(),
		"typ": "access",
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(hp.accessSecret))
}

func (hp *JWTHelper) CreateRefreshToken(username string) (string, error) {

	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(hp.refreshExp).Unix(),
		"iat": time.Now().Unix(),
		"typ": "refresh",
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(hp.refreshSecret))
	if err != nil {
		return "", err
	}

	return token, nil
}
