package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("Invalid token")
	ErrWrongType    = errors.New("Wrong token type")
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

func (hp *JWTHelper) ValidateAccessToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(hp.accessSecret), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	if claims["typ"] != "access" {
		return "", ErrWrongType
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", ErrInvalidToken
	}

	return sub, nil
}

func (hp *JWTHelper) ValidateRefreshToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(hp.refreshSecret), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	if claims["typ"] != "refresh" {
		return "", ErrWrongType
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", ErrInvalidToken
	}

	return sub, nil
}
