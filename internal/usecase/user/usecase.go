package user

import (
	"context"
	"errors"
	"meemo/internal/domain/entity"
	jwtService "meemo/internal/domain/token/service"
	"meemo/internal/domain/user/repository"
	"meemo/internal/domain/user/service"
)

type UseCase interface {
	CreateUser(ctx context.Context, in *CreateUserDtoIn) (*CreateUserDtoOut, error)
	GetUserInfo(ctx context.Context, in *GetUserInfoDtoIn) (*GetUserInfoOut, error)
	AuthUser(ctx context.Context, in *UserDtoIn) (*UserDtoOut, error)
	UpdateToken(ctx context.Context, in *UpdateTokenDtoIn) (*UpdateTokenDtoOut, error)
	Logout(ctx context.Context, in *LogoutDtoIn) (*LogoutDtoOut, error)
}

type useCase struct {
	repository repository.UserRepository
	service    service.UserService
	jwtService jwtService.JWTTokenService
}

func newUseCase(repository repository.UserRepository, service service.UserService) UseCase {
	return &useCase{
		repository: repository,
		service:    service,
	}
}

func (u *useCase) CreateUser(ctx context.Context, in *CreateUserDtoIn) (*CreateUserDtoOut, error) {
	user := &entity.User{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
	}
	u.service.HashPassword(user, in.Password)

	user, err := u.repository.Create(ctx, user)
	if err != nil {
		// TODO: Добавить исключение
		return nil, err
	}
	token, err := u.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}
	out := &CreateUserDtoOut{
		token.AccessToken,
		token.RefreshToken,
		token.ExpiresAt.Unix(),
	}
	return out, nil
}

func (u *useCase) GetUserInfo(ctx context.Context, in *GetUserInfoDtoIn) (*GetUserInfoOut, error) {
	userClaims, err := u.jwtService.ParseAccessToken(in.AccessToken)
	if err != nil {
		return nil, err
	}
	exp, err := userClaims.GetExpirationTime()
	if err != nil {
		return nil, err
	}
	if exp == nil {
		// TODO: Вынести в исключения
		return nil, errors.New("token expired")
	}
	user, err := u.repository.GetByEmail(ctx, userClaims.Email)
	if err != nil {
		return nil, err
	}
	out := &GetUserInfoOut{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	return out, nil
}

func (u *useCase) AuthUser(ctx context.Context, in *UserDtoIn) (*UserDtoOut, error) {
	user := &entity.User{
		Email: in.Email,
	}
	err := u.service.HashPassword(user, in.Password)
	if err != nil {
		return nil, err
	}
	check, err := u.repository.CheckPassword(ctx, user, user.PasswordSalt)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, errors.New("wrong password")
	}
	user, err = u.repository.GetByEmail(ctx, user.Email)
	token, err := u.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}
	out := &UserDtoOut{
		token.AccessToken,
		token.RefreshToken,
		token.ExpiresAt.Unix(),
	}
	return out, nil
}

func (u *useCase) UpdateToken(ctx context.Context, in *UpdateTokenDtoIn) (*UpdateTokenDtoOut, error) {
	return nil, nil
}

func (u *useCase) Logout(ctx context.Context, in *LogoutDtoIn) (*LogoutDtoOut, error) {
	return nil, nil
}
