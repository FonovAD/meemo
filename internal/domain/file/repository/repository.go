package repository

import (
	"context"
	"meemo/internal/domain/model"
)

type FileRepository interface {
	SaveFile(ctx context.Context, user *model.User, file *model.File) (*model.File, error)
	DeleteFile(ctx context.Context, user *model.User, file *model.File) error
	GetFile(ctx context.Context, user *model.User, id int64) (*model.File, error)
}
