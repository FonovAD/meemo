package repository

import (
	"context"
	"meemo/internal/domain/entity"
)

type FileRepository interface {
	Save(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error)
	Delete(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error)
	Get(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error)
	Rename(ctx context.Context, user *entity.User, file *entity.File, newName string) (*entity.File, error)
	ChangeVisibility(ctx context.Context, user *entity.User, file *entity.File, isPublic bool) (*entity.File, error)
	SetStatus(ctx context.Context, user *entity.User, file *entity.File, status int) (*entity.File, error)
}
