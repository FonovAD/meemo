package file

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/file/repository"
	"meemo/internal/infrastructure/storage/model"
)

type fileRepository struct {
	conn *sqlx.DB
}

func NewFileRepository(conn *sqlx.DB) repository.FileRepository {
	return &fileRepository{
		conn: conn,
	}
}

func (fr *fileRepository) Save(ctx context.Context, userID int64, originalName, mimeType, s3Bucket, s3Key string, sizeInBytes int64, isPublic bool) (*entity.File, error) {
	fileModel := &model.File{
		UserID:       userID,
		OriginalName: originalName,
		MimeType:     mimeType,
		S3Bucket:     s3Bucket,
		S3Key:        s3Key,
		SizeInBytes:  sizeInBytes,
		IsPublic:     isPublic,
	}

	rows, err := fr.conn.NamedQueryContext(ctx, SaveFileTemplate, fileModel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&fileModel.ID); err != nil {
			return nil, err
		}
		return fileModel.ModelToEntity(), nil
	}
	return nil, sql.ErrNoRows
}

func (fr *fileRepository) Delete(ctx context.Context, userEmail, originalName string) (*entity.File, error) {
	fileModel := &model.File{}

	err := fr.conn.QueryRowxContext(ctx, DeleteFileTemplate, userEmail, originalName).Scan(&fileModel.ID)
	if err != nil {
		return nil, err
	}
	fileModel.OriginalName = originalName
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) Get(ctx context.Context, fileID int64) (*entity.File, error) {
	fileModel := &model.File{}

	err := fr.conn.QueryRowxContext(ctx, GetFileTemplate, fileID).StructScan(fileModel)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) GetByOriginalNameAndUserEmail(ctx context.Context, userEmail, originalName string) (*entity.File, error) {
	fileModel := &model.File{}

	err := fr.conn.QueryRowxContext(ctx, GetFileByOriginalNameAndUserEmailTemplate, userEmail, originalName).StructScan(fileModel)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) ChangeVisibility(ctx context.Context, userEmail, originalName string, isPublic bool) (*entity.File, error) {
	fileModel := &model.File{}

	err := fr.conn.QueryRowxContext(ctx, ChangeVisibilityTemplate, isPublic, userEmail, originalName).Scan(&fileModel.ID, &fileModel.IsPublic, &fileModel.UpdatedAt)
	if err != nil {
		return nil, err
	}
	fileModel.OriginalName = originalName
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) SetStatus(ctx context.Context, userEmail, originalName string, status int) (*entity.File, error) {
	fileModel := &model.File{}

	err := fr.conn.QueryRowxContext(ctx, SetStatusTemplate, status, userEmail, originalName).Scan(&fileModel.ID, &fileModel.Status, &fileModel.UpdatedAt)
	if err != nil {
		return nil, err
	}
	fileModel.OriginalName = originalName
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) Rename(ctx context.Context, userEmail, originalName, newName string) (*entity.File, error) {
	fileModel := &model.File{}

	err := fr.conn.QueryRowxContext(ctx, RenameFileTemplate, newName, userEmail, originalName).Scan(&fileModel.ID, &fileModel.OriginalName, &fileModel.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) List(ctx context.Context, userEmail string) ([]*entity.File, error) {
	rows, err := fr.conn.QueryxContext(ctx, ListUserFilesTemplate, userEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*entity.File
	for rows.Next() {
		fileModel := &model.File{}
		if err := rows.StructScan(fileModel); err != nil {
			return nil, err
		}
		files = append(files, fileModel.ModelToEntity())
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
