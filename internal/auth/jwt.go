package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zrp9/launchl/internal/config"
)

var (
	authToken []byte
	authErr   error
	once      sync.Once
)

var Expirey = time.Now().Add((24 * time.Hour) * 365)

func GetJwtKey() ([]byte, error) {
	return getKey()
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type ErrExpiredToken struct {
	ExpireyDate string
}

func (e ErrExpiredToken) Error() string {
	return fmt.Sprintf("invalid token expired at %v", e.ExpireyDate)
}

func GenerateToken(id string, username string, role string) (string, error) {
	expirationTime := time.Now().Add(2 * time.Hour)
	//devTime := time.Now().Add((24 * time.Hour) * 365)

	claims := &UserClaims{
		ID:       id,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	key, err := getKey()

	if err != nil {
		return "", err

	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func NewRefreshToken(claims jwt.RegisteredClaims) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return refreshToken.SignedString([]byte(key))
}

func ParseAuthToken(accessToken string) *UserClaims {
	parsedAccessToken, _ := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		key, err := getKey()
		if err != nil {
			return nil, err
		}
		return []byte(key), nil
	})

	return parsedAccessToken.Claims.(*UserClaims)
}

func ParseRefreshToken(refreshToken string) *jwt.RegisteredClaims {
	parsedRefreshToken, _ := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		key, err := getKey()
		if err != nil {
			return nil, err
		}

		return []byte(key), nil
	})

	return parsedRefreshToken.Claims.(*jwt.RegisteredClaims)
}

func getKey() ([]byte, error) {
	once.Do(func() {
		authToken, authErr = config.GetAuthToken()
	})
	return authToken, authErr
}
