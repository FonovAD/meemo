package file

import (
	"context"
	"github.com/jmoiron/sqlx"
	"meemo/internal/domain/file/repository"
	"meemo/internal/domain/model"
)

type fileRepository struct {
	conn *sqlx.DB
}

func NewFileRepository(conn *sqlx.DB) repository.FileRepository {
	return &fileRepository{}
}

//DeleteFile(ctx context.Context, user *model.User, file *model.File) error
//GetFile(ctx context.Context, user *model.User, id int64) (*model.File, error)
//ChangeVisibility(ctx context.Context, user *model.User, file *model.File) error
//SetStatus(ctx context.Context, user *model.User, file *model.File) error

func (fr *fileRepository) SaveFile(ctx context.Context, user *model.User, file *model.File) (*model.File, error) {
	err := fr.conn.QueryRowxContext(ctx, SaveFileTemplate, user, file).Scan(&file.ID)
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (fr *fileRepository) DeleteFile(ctx context.Context, user *model.User, file *model.File) (*model.File, error) {
	err := fr.conn.QueryRowxContext(ctx, DeleteFileTemplate, user.Email, file.OriginalName).Scan(&file.ID)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) GetFile(ctx context.Context, user *model.User, file *model.File) (*model.File, error) {
	err := fr.conn.QueryRowxContext(ctx, GetFileTemplate, user.Email, file.OriginalName).Scan(&file.ID)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) ChangeVisibility(ctx context.Context, user *model.User, file *model.File, isPublic bool) (*model.File, error) {
	err := fr.conn.QueryRowxContext(ctx, ChangeVisibilityTemplate, isPublic, user.Email, file.OriginalName).Scan(&file.ID)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) SetStatus(ctx context.Context, user *model.User, file *model.File, status int) (*model.File, error) {
	err := fr.conn.QueryRowxContext(ctx, ChangeVisibilityTemplate, status, user.Email, file.OriginalName).Scan(&file.ID)
	if err != nil {
		return nil, err
	}
	return file, nil
}
