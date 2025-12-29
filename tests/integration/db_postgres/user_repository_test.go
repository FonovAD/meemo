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

	createdUser, err := ur.Create(context.Background(), newUser)
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
	createdUser, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	foundUser, err := ur.GetByEmail(context.Background(), "ivan@test.com")
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

	_, err := ur.GetByEmail(context.Background(), "nonexistent@test.com")
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
	createdUser, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	createdUser.FirstName = "Петр Updated"
	createdUser.LastName = "Петров Updated"
	us.HashPassword(createdUser, "newpassword")

	updatedUser, err := ur.Update(context.Background(), createdUser)
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
	_, err := ur.Create(context.Background(), user1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	user2 := &entity.User{
		FirstName: "Второй",
		LastName:  "Пользователь",
		Email:     "user2@test.com",
	}
	us.HashPassword(user2, "pass2")
	createdUser2, err := ur.Create(context.Background(), user2)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	_, err = ur.UpdateEmail(context.Background(), createdUser2, "user1@test.com")
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
	createdUser, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	updatedUser, err := ur.UpdateEmail(context.Background(), createdUser, "maria@test.com")
	if err != nil {
		t.Fatalf("Unexpected error when updating to same email: %v", err)
	}

	if updatedUser.Email != "maria@test.com" {
		t.Errorf("Expected email %s, got %s", "maria@test.com", updatedUser.Email)
	}

	foundUser, err := ur.GetByEmail(context.Background(), "maria@test.com")
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
	createdUser, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = ur.UpdateEmail(context.Background(), createdUser, "")
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
	createdUser, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	deletedUser, err := ur.Delete(context.Background(), createdUser)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if deletedUser.ID != createdUser.ID {
		t.Errorf("Expected deleted user ID %d, got %d", createdUser.ID, deletedUser.ID)
	}

	_, err = ur.GetByEmail(context.Background(), "delete@test.com")
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

	_, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	duplicateUser := &entity.User{
		FirstName: "Другой",
		LastName:  "Пользователь",
		Email:     "duplicate@test.com",
	}
	us.HashPassword(duplicateUser, "test2")

	_, err = ur.Create(context.Background(), duplicateUser)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestCheckPassword_Success(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Проверка",
		LastName:  "Пароля",
		Email:     "checkpass@test.com",
	}

	us := service.NewUserService()
	password := "correctpassword123"
	us.HashPassword(newUser, password)

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Проверяем пароль с правильным хешем
	check, err := ur.CheckPassword(context.Background(), createdUser, createdUser.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password: %v", err)
	}

	if !check {
		t.Error("Expected password check to return true for correct password hash, got false")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	newUser := &entity.User{
		FirstName: "Неправильный",
		LastName:  "Пароль",
		Email:     "wrongpass@test.com",
	}

	us := service.NewUserService()
	us.HashPassword(newUser, "correctpassword")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Создаем пользователя с другим паролем для проверки
	wrongPasswordUser := &entity.User{
		Email:        createdUser.Email,
		PasswordSalt: "wrong_hash_that_does_not_match",
	}

	check, err := ur.CheckPassword(context.Background(), wrongPasswordUser, wrongPasswordUser.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password: %v", err)
	}

	if check {
		t.Error("Expected password check to return false for wrong password hash, got true")
	}
}

func TestCheckPassword_NonExistentUser(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	nonExistentUser := &entity.User{
		Email:        "nonexistent@test.com",
		PasswordSalt: "some_hash",
	}

	check, err := ur.CheckPassword(context.Background(), nonExistentUser, nonExistentUser.PasswordSalt)
	if err != nil {
		t.Fatalf("Unexpected error when checking password for non-existent user: %v", err)
	}

	if check {
		t.Error("Expected password check to return false for non-existent user, got true")
	}
}

func TestCheckPassword_EmptyEmail(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	emptyEmailUser := &entity.User{
		Email:        "",
		PasswordSalt: "some_hash",
	}

	check, err := ur.CheckPassword(context.Background(), emptyEmailUser, emptyEmailUser.PasswordSalt)
	if err != nil {
		t.Fatalf("Unexpected error when checking password with empty email: %v", err)
	}

	if check {
		t.Error("Expected password check to return false for empty email, got true")
	}
}

func TestCheckPassword_DifferentUsersSamePassword(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	us := service.NewUserService()

	// Создаем первого пользователя
	user1 := &entity.User{
		FirstName: "Первый",
		LastName:  "Пользователь",
		Email:     "user1@test.com",
	}
	us.HashPassword(user1, "samepassword")
	createdUser1, err := ur.Create(context.Background(), user1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Создаем второго пользователя с тем же паролем
	user2 := &entity.User{
		FirstName: "Второй",
		LastName:  "Пользователь",
		Email:     "user2@test.com",
	}
	us.HashPassword(user2, "samepassword")
	createdUser2, err := ur.Create(context.Background(), user2)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	// Проверяем, что хеши разные (bcrypt генерирует разные хеши)
	if createdUser1.PasswordSalt == createdUser2.PasswordSalt {
		t.Error("Expected different password hashes for same password (bcrypt uses random salt), got same hash")
	}

	// Проверяем пароль первого пользователя
	check1, err := ur.CheckPassword(context.Background(), createdUser1, createdUser1.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password for user1: %v", err)
	}
	if !check1 {
		t.Error("Expected password check to return true for user1, got false")
	}

	// Проверяем пароль второго пользователя
	check2, err := ur.CheckPassword(context.Background(), createdUser2, createdUser2.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password for user2: %v", err)
	}
	if !check2 {
		t.Error("Expected password check to return true for user2, got false")
	}

	// Проверяем, что нельзя использовать хеш одного пользователя для другого
	wrongCheck, err := ur.CheckPassword(context.Background(), createdUser1, createdUser2.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password: %v", err)
	}
	if wrongCheck {
		t.Error("Expected password check to return false when using wrong user's hash, got true")
	}
}
