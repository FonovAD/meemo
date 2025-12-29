package repository

import (
	"context"
	"meemo/internal/domain/entity"
	"time"
)

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, id string, userID int, expiresAt, createdAt time.Time, revoked bool) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID int) error
}
