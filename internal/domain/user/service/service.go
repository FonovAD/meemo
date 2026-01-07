package service

import (
	"meemo/internal/domain/entity"

	"golang.org/x/crypto/bcrypt"
)

const secret = "your-256-bit-secret"

type UserService interface {
	HashPassword(user *entity.User, password string) error
}

type userService struct {
	secret []byte
}

func NewUserService() UserService {
	return &userService{secret: []byte(secret)}
}

func (us *userService) HashPassword(user *entity.User, password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordSalt = string(bytes)
	return nil
}
