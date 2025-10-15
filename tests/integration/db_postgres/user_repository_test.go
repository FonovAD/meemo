package db_postgres

import (
	"context"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v3"
	"log"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/user/service"
	storage "meemo/internal/infrastructure/storage/pg"
	"meemo/internal/infrastructure/storage/pg/user"
	"os"
	"testing"
)

func initDB(t *testing.T) (*sqlx.DB, func()) {
	configFile, err := os.ReadFile("db_config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg storage.PGConfig
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	return SetupTestDB(t, cfg)
}

func TestCreateUser(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Тест",
		LastName:  "Тестов",
		Email:     "test@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "test")

	ur := user.NewUserRepository(db)

	createdUser, err := ur.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if createdUser.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if createdUser.FirstName != newUser.FirstName {
		t.Errorf("Expected first name %s, got %s", newUser.FirstName, createdUser.FirstName)
	}
	if createdUser.Email != newUser.Email {
		t.Errorf("Expected email %s, got %s", newUser.Email, createdUser.Email)
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Иван",
		LastName:  "Иванов",
		Email:     "ivan@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "password123")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	foundUser, err := ur.GetUserByEmail(context.Background(), "ivan@test.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if foundUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
	}
	if foundUser.Email != "ivan@test.com" {
		t.Errorf("Expected email %s, got %s", "ivan@test.com", foundUser.Email)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	_, err := ur.GetUserByEmail(context.Background(), "nonexistent@test.com")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

func TestUpdateUser(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Петр",
		LastName:  "Петров",
		Email:     "petr@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "oldpassword")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	createdUser.FirstName = "Петр Updated"
	createdUser.LastName = "Петров Updated"
	us.HashPassword(createdUser, "newpassword")

	updatedUser, err := ur.UpdateUser(context.Background(), createdUser)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if updatedUser.FirstName != "Петр Updated" {
		t.Errorf("Expected first name %s, got %s", "Петр Updated", updatedUser.FirstName)
	}
}

func TestUpdateUserEmail_DuplicateEmail(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	user1 := &entity.User{
		FirstName: "Первый",
		LastName:  "Пользователь",
		Email:     "user1@test.com",
	}
	us := service.NewUserService()
	us.HashPassword(user1, "pass1")
	_, err := ur.CreateUser(context.Background(), user1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	user2 := &entity.User{
		FirstName: "Второй",
		LastName:  "Пользователь",
		Email:     "user2@test.com",
	}
	us.HashPassword(user2, "pass2")
	createdUser2, err := ur.CreateUser(context.Background(), user2)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	_, err = ur.UpdateUserEmail(context.Background(), createdUser2, "user1@test.com")
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestUpdateUserEmail_SameEmail(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Мария",
		LastName:  "Иванова",
		Email:     "maria@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "password")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	updatedUser, err := ur.UpdateUserEmail(context.Background(), createdUser, "maria@test.com")
	if err != nil {
		t.Fatalf("Unexpected error when updating to same email: %v", err)
	}

	if updatedUser.Email != "maria@test.com" {
		t.Errorf("Expected email %s, got %s", "maria@test.com", updatedUser.Email)
	}

	foundUser, err := ur.GetUserByEmail(context.Background(), "maria@test.com")
	if err != nil {
		t.Fatalf("Failed to get user after same email update: %v", err)
	}
	if foundUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
	}
}

func TestUpdateUserEmail_EmptyEmail(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Дмитрий",
		LastName:  "Сидоров",
		Email:     "dmitry@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "password")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = ur.UpdateUserEmail(context.Background(), createdUser, "")
	if err == nil {
		t.Error("Expected error for empty email, got nil")
	}
}

func TestDeleteUser(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Удаляемый",
		LastName:  "Пользователь",
		Email:     "delete@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "password")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	deletedUser, err := ur.DeleteUser(context.Background(), createdUser)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if deletedUser.ID != createdUser.ID {
		t.Errorf("Expected deleted user ID %d, got %d", createdUser.ID, deletedUser.ID)
	}

	_, err = ur.GetUserByEmail(context.Background(), "delete@test.com")
	if err == nil {
		t.Error("Expected error when getting deleted user, got nil")
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Тест",
		LastName:  "Тестов",
		Email:     "duplicate@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "test")

	ur := user.NewUserRepository(db)

	_, err := ur.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	duplicateUser := &entity.User{
		FirstName: "Другой",
		LastName:  "Пользователь",
		Email:     "duplicate@test.com",
	}
	us.HashPassword(duplicateUser, "test2")

	_, err = ur.CreateUser(context.Background(), duplicateUser)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}
