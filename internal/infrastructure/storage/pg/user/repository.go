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

func (ur *userRepository) Create(ctx context.Context, firstName, lastName, email, passwordSalt string) (*entity.User, error) {
	userModel := &model.User{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordSalt: passwordSalt,
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

func (ur *userRepository) Update(ctx context.Context, id int64, firstName, lastName, email, passwordSalt string) (*entity.User, error) {
	userModel := &model.User{
		ID:           id,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordSalt: passwordSalt,
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

func (ur *userRepository) UpdateEmail(ctx context.Context, oldEmail, newEmail string) (*entity.User, error) {
	userModel := &model.User{}

	err := ur.conn.QueryRowxContext(ctx, UpdateUserEmailTemplate, oldEmail, newEmail).Scan(&userModel.ID)
	if err != nil {
		return nil, err
	}
	userModel.Email = newEmail
	return userModel.ModelToEntity(), nil
}

func (ur *userRepository) Delete(ctx context.Context, email string) (*entity.User, error) {
	userModel := &model.User{}

	err := ur.conn.QueryRowxContext(ctx, DeleteUserTemplate, email).Scan(&userModel.ID)
	if err != nil {
		return nil, err
	}
	userModel.Email = email
	return userModel.ModelToEntity(), nil
}

func (ur *userRepository) CheckPassword(ctx context.Context, email, saldPassword string) (bool, error) {
	check := false

	err := ur.conn.QueryRowxContext(ctx, CheckPassword, email, saldPassword).Scan(&check)
	if err != nil {
		return false, err
	}
	return check, nil
}
