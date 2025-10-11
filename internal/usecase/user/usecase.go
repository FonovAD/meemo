package user

import "context"

type UserUseCase interface {
	CreateUser(ctx context.Context, in *CreateUserDtoIn) (*CreateUserDtoOut, error)
	GetUserInfo(ctx context.Context, in *GetUserInfoDtoIn) (*GetUserInfoOut, error)
	AuthUser(ctx context.Context, in *UserDtoIn) (*UserDtoOut, error)
	UpdateToken(ctx context.Context, in *UpdateTokenDtoIn) (*UpdateTokenDtoOut, error)
	Logout(ctx context.Context, in *LogoutDtoIn) (*LogoutDtoOut, error)
}
