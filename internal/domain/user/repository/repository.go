package repository

import (
	"context"
	"meemo/internal/domain/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, id int64) (*model.User, error)
}
