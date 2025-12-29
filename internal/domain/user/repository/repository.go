package repository

import (
	"context"
	"meemo/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, firstName, lastName, email, passwordSalt string) (*entity.User, error)
	Update(ctx context.Context, id int64, firstName, lastName, email, passwordSalt string) (*entity.User, error)
	UpdateEmail(ctx context.Context, oldEmail, newEmail string) (*entity.User, error)
	Delete(ctx context.Context, email string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	CheckPassword(ctx context.Context, email, saldPassword string) (bool, error)
}
