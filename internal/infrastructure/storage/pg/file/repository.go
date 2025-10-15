package file

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/file/repository"
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

func (fr *fileRepository) SaveFile(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error) {
	file.UserID = user.ID
	rows, err := fr.conn.NamedQueryContext(ctx, SaveFileTemplate, file)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&file.ID); err != nil {
			return nil, err
		}
		return file, nil
	}
	return nil, sql.ErrNoRows
}

func (fr *fileRepository) DeleteFile(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error) {
	err := fr.conn.QueryRowxContext(ctx, DeleteFileTemplate, user.Email, file.OriginalName).Scan(&file.ID)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) GetFile(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error) {
	err := fr.conn.QueryRowxContext(ctx, GetFileTemplate, user.Email, file.OriginalName).StructScan(file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) ChangeVisibility(ctx context.Context, user *entity.User, file *entity.File, isPublic bool) (*entity.File, error) {
	err := fr.conn.QueryRowxContext(ctx, ChangeVisibilityTemplate, isPublic, user.Email, file.OriginalName).Scan(&file.ID, &file.IsPublic)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fr *fileRepository) SetStatus(ctx context.Context, user *entity.User, file *entity.File, status int) (*entity.File, error) {
	err := fr.conn.QueryRowxContext(ctx, SetStatusTemplate, status, user.Email, file.OriginalName).Scan(&file.ID, &file.Status) // Исправлено
	if err != nil {
		return nil, err
	}
	return file, nil
}
