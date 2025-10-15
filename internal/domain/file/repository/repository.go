package repository

import (
	"context"
	"meemo/internal/domain/entity"
)

type FileRepository interface {
	SaveFile(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error)
	DeleteFile(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error)
	GetFile(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error)
	ChangeVisibility(ctx context.Context, user *entity.User, file *entity.File, isPublic bool) (*entity.File, error)
	SetStatus(ctx context.Context, user *entity.User, file *entity.File, status int) (*entity.File, error)
}
