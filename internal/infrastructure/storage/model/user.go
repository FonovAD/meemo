package model

import (
	"errors"
	"meemo/internal/domain/entity"
)

type User struct {
	ID           int64  `db:"id"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Email        string `db:"email"`
	PasswordSalt string `db:"password_salt"`
}

func (m *User) ModelToEntity() *entity.User {
	return &entity.User{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		Email:        m.Email,
		PasswordSalt: m.PasswordSalt,
	}
}

func (m *User) EntityToModel(entity *entity.User) error {
	if entity == nil {
		return errors.New("entity is nil")
	}
	m.ID = entity.ID
	m.FirstName = entity.FirstName
	m.LastName = entity.LastName
	m.Email = entity.Email
	m.PasswordSalt = entity.PasswordSalt
	return nil
}
