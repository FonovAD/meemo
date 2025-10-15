package service

import (
	"golang.org/x/crypto/bcrypt"
	"meemo/internal/domain/entity"
)

const secret = "your-256-bit-secret"
const salt = "salt"

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
