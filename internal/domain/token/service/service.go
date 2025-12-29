package service

import (
	"crypto/rand"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"math/big"
	"meemo/internal/domain/entity"
	"strconv"
	"time"
)

type TokenService interface {
	GenerateTokenPair(user *entity.User) (*entity.TokenPair, error)
	ParseAccessToken(tokenString string) (*UserClaims, error)
	ValidateAccessToken(tokenString string) error
}

type JWTTokenService struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTTokenService(secretKey string, accessExpiry, refreshExpiry time.Duration) *JWTTokenService {
	return &JWTTokenService{
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *JWTTokenService) GenerateTokenPair(user *entity.User) (*entity.TokenPair, error) {
	accessExpiresAt := time.Now().Add(s.accessExpiry)
	userIDStr := strconv.FormatInt(user.ID, 10)
	accessClaims := UserClaims{
		UserID: userIDStr,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userIDStr,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := generateRandomString(128)
	if err != nil {
		return nil, err
	}

	return &entity.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

func (s *JWTTokenService) ParseAccessToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *JWTTokenService) ValidateAccessToken(tokenString string) error {
	_, err := s.ParseAccessToken(tokenString)
	return err
}
