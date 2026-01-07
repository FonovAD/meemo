package repository

import (
	"context"
	"meemo/internal/domain/entity"
)

type FileRepository interface {
	Save(ctx context.Context, userID int64, originalName, mimeType, s3Bucket, s3Key string, sizeInBytes int64, isPublic bool) (*entity.File, error)
	Delete(ctx context.Context, userEmail, originalName string) (*entity.File, error)
	Get(ctx context.Context, fileID int64) (*entity.File, error)
	GetByOriginalNameAndUserEmail(ctx context.Context, userEmail, originalName string) (*entity.File, error)
	Rename(ctx context.Context, userEmail, originalName, newName string) (*entity.File, error)
	ChangeVisibility(ctx context.Context, userEmail, originalName string, isPublic bool) (*entity.File, error)
	SetStatus(ctx context.Context, userEmail, originalName string, status int) (*entity.File, error)
	List(ctx context.Context, userEmail string) ([]*entity.File, error)
	GetTotalUsedSpace(ctx context.Context, userEmail string) (int64, error)
}
