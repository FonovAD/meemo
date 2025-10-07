package file

import (
	"context"
	"github.com/jmoiron/sqlx"
	"log"
	"meemo/internal/domain/file/repository"
	"meemo/internal/domain/model"
)

// TODO: добавить запросы по именам атрибутов вместо позиционной
type fileRepository struct {
	conn *sqlx.DB
}

func NewFileRepository(conn *sqlx.DB) repository.FileRepository {
	return &fileRepository{
		conn: conn,
	}
}

func (fr *fileRepository) SaveFile(ctx context.Context, user *model.User, file *model.File) (*model.File, error) {
	log.Print(user)
	log.Print(file)
	file.UserID = user.ID
	err := fr.conn.QueryRowxContext(ctx, SaveFileTemplate, user.ID,
		file.OriginalName,
		file.MimeType,
		file.SizeInBytes,
		file.S3Bucket,
		file.S3Key,
		file.Status,
		file.CreatedAt,
		file.UpdatedAt,
		file.IsPublic).Scan(&file.ID)
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
	err := fr.conn.QueryRowxContext(ctx, GetFileTemplate, user.Email, file.OriginalName).Scan(
		&file.ID,
		&file.UserID,
		&file.OriginalName,
		&file.MimeType,
		&file.SizeInBytes,
		&file.S3Bucket,
		&file.S3Key,
		&file.Status,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.IsPublic)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) ChangeVisibility(ctx context.Context, user *model.User, file *model.File, isPublic bool) (*model.File, error) {
	err := fr.conn.QueryRowxContext(ctx, ChangeVisibilityTemplate, isPublic, user.Email, file.OriginalName).Scan(&file.ID, &file.IsPublic)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) SetStatus(ctx context.Context, user *model.User, file *model.File, status int) (*model.File, error) {
	err := fr.conn.QueryRowxContext(ctx, SetStatusTemplate, status, user.Email, file.OriginalName).Scan(&file.ID, &file.Status) // Исправлено
	if err != nil {
		return nil, err
	}
	return file, nil
}
