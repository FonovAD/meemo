package repository

import (
	"context"
	"meemo/internal/domain/entity"
)

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *entity.RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, user *entity.User) error
}
