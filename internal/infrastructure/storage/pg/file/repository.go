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

func (fr *fileRepository) Save(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error) {
	fileModel := &model.File{}
	if err := fileModel.EntityToModel(file); err != nil {
		return nil, err
	}
	fileModel.UserID = user.ID

	file.UserID = user.ID
	rows, err := fr.conn.NamedQueryContext(ctx, SaveFileTemplate, fileModel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&fileModel.ID); err != nil {
			return nil, err
		}
		file.ID = fileModel.ID
		return file, nil
	}
	return nil, sql.ErrNoRows
}

func (fr *fileRepository) Delete(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error) {
	fileModel := &model.File{}
	if err := fileModel.EntityToModel(file); err != nil {
		return nil, err
	}
	fileModel.UserID = user.ID
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	err := fr.conn.QueryRowxContext(ctx, DeleteFileTemplate, userModel.Email, fileModel.OriginalName).Scan(&fileModel.ID)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) Get(ctx context.Context, user *entity.User, file *entity.File) (*entity.File, error) {
	fileModel := &model.File{}
	if err := fileModel.EntityToModel(file); err != nil {
		return nil, err
	}
	fileModel.UserID = user.ID
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	err := fr.conn.QueryRowxContext(ctx, GetFileTemplate, userModel.Email, fileModel.OriginalName).StructScan(fileModel)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) ChangeVisibility(ctx context.Context, user *entity.User, file *entity.File, isPublic bool) (*entity.File, error) {
	fileModel := &model.File{}
	if err := fileModel.EntityToModel(file); err != nil {
		return nil, err
	}
	fileModel.UserID = user.ID
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	err := fr.conn.QueryRowxContext(ctx, ChangeVisibilityTemplate, isPublic, userModel.Email, fileModel.OriginalName).Scan(&fileModel.ID, &fileModel.IsPublic)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) SetStatus(ctx context.Context, user *entity.User, file *entity.File, status int) (*entity.File, error) {
	fileModel := &model.File{}
	if err := fileModel.EntityToModel(file); err != nil {
		return nil, err
	}
	fileModel.UserID = user.ID
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	err := fr.conn.QueryRowxContext(ctx, SetStatusTemplate, status, userModel.Email, fileModel.OriginalName).Scan(&fileModel.ID, &fileModel.Status)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}

func (fr *fileRepository) Rename(ctx context.Context, user *entity.User, file *entity.File, newName string) (*entity.File, error) {
	fileModel := &model.File{}
	if err := fileModel.EntityToModel(file); err != nil {
		return nil, err
	}
	fileModel.UserID = user.ID
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	err := fr.conn.QueryRowxContext(ctx, RenameFileTemplate, newName, userModel.Email, fileModel.OriginalName).Scan(&fileModel.ID, &fileModel.OriginalName)
	if err != nil {
		return nil, err
	}
	return fileModel.ModelToEntity(), nil
}
