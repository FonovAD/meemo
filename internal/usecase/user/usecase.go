package user

import (
	"context"
	"errors"
	"meemo/internal/domain/entity"
	jwtService "meemo/internal/domain/token/service"
	"meemo/internal/domain/user/repository"
	"meemo/internal/domain/user/service"

	"golang.org/x/crypto/bcrypt"
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
	jwtService jwtService.TokenService
}

func NewUseCase(repository repository.UserRepository, service service.UserService, jwtService jwtService.TokenService) UseCase {
	return &useCase{
		repository: repository,
		service:    service,
		jwtService: jwtService,
	}
}

func (u *useCase) CreateUser(ctx context.Context, in *CreateUserDtoIn) (*CreateUserDtoOut, error) {
	user := &entity.User{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
	}
	if err := u.service.HashPassword(user, in.Password); err != nil {
		return nil, err
	}

	user, err := u.repository.Create(ctx, user.FirstName, user.LastName, user.Email, user.PasswordSalt)
	if err != nil {
		// TODO: Добавить исключение
		return nil, err
	}
	token, err := u.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}
	out := &CreateUserDtoOut{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresAt.Unix(),
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
	// Получаем пользователя по email
	user, err := u.repository.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Проверяем пароль с помощью bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordSalt), []byte(in.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Генерируем токены
	token, err := u.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}
	out := &UserDtoOut{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresAt.Unix(),
	}
	return out, nil
}

func (u *useCase) UpdateToken(ctx context.Context, in *UpdateTokenDtoIn) (*UpdateTokenDtoOut, error) {
	return nil, nil
}

func (u *useCase) Logout(ctx context.Context, in *LogoutDtoIn) (*LogoutDtoOut, error) {
	return nil, nil
}
