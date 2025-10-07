package user

import (
	"context"
	"github.com/jmoiron/sqlx"
	"meemo/internal/domain/model"
	"meemo/internal/domain/user/repository"
)

type userRepository struct {
	conn *sqlx.DB
}

func NewUserRepository(conn *sqlx.DB) repository.UserRepository {
	return &userRepository{conn}
}

func (ur *userRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	err := ur.conn.QueryRowxContext(ctx, CreateUserTemplate, user.FirstName, user.LastName, user.Email, user.PasswordSalt).Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := ur.conn.QueryRowxContext(ctx, GetUserByEmailTemplate, email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordSalt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) UpdateUser(ctx context.Context, user *model.User) (*model.User, error) {
	err := ur.conn.QueryRowxContext(ctx, UpdateUserTemplate, user.Email, user.FirstName, user.LastName, user.PasswordSalt).Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) UpdateUserEmail(ctx context.Context, user *model.User, email string) (*model.User, error) {
	err := ur.conn.QueryRowxContext(ctx, UpdateUserEmailTemplate, user.Email, email).Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) DeleteUser(ctx context.Context, user *model.User) (*model.User, error) {
	err := ur.conn.QueryRowxContext(ctx, DeleteUserTemplate, user.Email).Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
