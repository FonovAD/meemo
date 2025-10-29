package user

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/user/repository"
	"meemo/internal/infrastructure/storage/model"
)

type userRepository struct {
	conn *sqlx.DB
}

func NewUserRepository(conn *sqlx.DB) repository.UserRepository {
	return &userRepository{conn}
}

func (ur *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	rows, err := ur.conn.NamedQueryContext(ctx, CreateUserTemplate, &userModel)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		if err := rows.Scan(&userModel.ID); err != nil {
			return nil, err
		}
		return userModel.ModelToEntity(), nil
	}
	return nil, sql.ErrNoRows
}

func (ur *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	userModel := &model.User{}

	err := ur.conn.QueryRowxContext(ctx, GetUserByEmailTemplate, email).StructScan(userModel)
	if err != nil {
		return nil, err
	}
	return userModel.ModelToEntity(), nil
}

func (ur *userRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	rows, err := ur.conn.NamedQueryContext(ctx, UpdateUserTemplate, &userModel)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		if err := rows.Scan(&userModel.ID); err != nil {
			return nil, err
		}
		return userModel.ModelToEntity(), nil
	}
	return nil, sql.ErrNoRows
}

func (ur *userRepository) UpdateEmail(ctx context.Context, user *entity.User, email string) (*entity.User, error) {
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	err := ur.conn.QueryRowxContext(ctx, UpdateUserEmailTemplate, userModel.Email, email).Scan(&userModel.ID)
	if err != nil {
		return nil, err
	}
	return userModel.ModelToEntity(), nil
}

func (ur *userRepository) Delete(ctx context.Context, user *entity.User) (*entity.User, error) {
	userModel := &model.User{}
	if err := userModel.EntityToModel(user); err != nil {
		return nil, err
	}

	err := ur.conn.QueryRowxContext(ctx, DeleteUserTemplate, userModel.Email).Scan(&userModel.ID)
	if err != nil {
		return nil, err
	}
	return userModel.ModelToEntity(), nil
}
