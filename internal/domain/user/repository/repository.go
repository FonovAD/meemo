package repository

import (
	"context"
	"meemo/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdateEmail(ctx context.Context, user *entity.User, email string) (*entity.User, error)
	Delete(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}
