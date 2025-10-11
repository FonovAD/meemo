package repository

import (
	"context"
	"meemo/internal/domain/model"
)

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, user *model.User) error
}
