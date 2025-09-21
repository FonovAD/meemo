package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"meemo/internal/domain/model"
	"strconv"
)

const secret = "your-256-bit-secret"

type UserService interface {
	CreateToken(ctx context.Context, user *model.User) (string, error)
	ParseToken(ctx context.Context, token string) (*model.User, error)
}

type userService struct {
	secret []byte
}

func NewUserService() UserService {
	return &userService{secret: []byte(secret)}
}

func (us *userService) CreateToken(ctx context.Context, user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"id":         user.ID,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
	})
	return token.SignedString(us.secret)
}

func (us *userService) ParseToken(ctx context.Context, tokenString string) (*model.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return us.secret, nil
	})
	if err != nil {
		return nil, errors.Join(err, errors.New("get token failed"))
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		str, ok := claims["id"].(string)
		if !ok {
			return nil, fmt.Errorf("claim parsing id failed")
		}
		userID, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("claim parsing id failed")
		}
		firstName, ok := claims["first_name"].(string)
		if !ok {
			return nil, fmt.Errorf("claim parsing first_name failed")
		}
		lastName, ok := claims["last_name"].(string)
		if !ok {
			return nil, fmt.Errorf("claim parsing last_name failed")
		}
		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("claim parsing email failed")
		}
		return &model.User{
			ID:        userID,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
		}, nil
	}

	return nil, fmt.Errorf("claim parsing id failed")
}
