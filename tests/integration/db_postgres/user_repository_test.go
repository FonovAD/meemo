package db_postgres

import (
	"context"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/user/service"
	"meemo/internal/infrastructure/storage/pg/user"
	"testing"
)

func hashPassword(t *testing.T, password string) string {
	us := service.NewUserService()
	u := &entity.User{}
	if err := us.HashPassword(u, password); err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	return u.PasswordSalt
}

func TestCreateUser(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "test")

	ur := user.NewUserRepository(db)

	createdUser, err := ur.Create(context.Background(), "Тест", "Тестов", "test@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if createdUser.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if createdUser.FirstName != "Тест" {
		t.Errorf("Expected first name %s, got %s", "Тест", createdUser.FirstName)
	}
	if createdUser.Email != "test@test.com" {
		t.Errorf("Expected email %s, got %s", "test@test.com", createdUser.Email)
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "password123")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), "Иван", "Иванов", "ivan@test.com", passwordHash)
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
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	_, err := ur.GetByEmail(context.Background(), "nonexistent@test.com")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

func TestUpdateUser(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "oldpassword")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), "Петр", "Петров", "petr@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	newPasswordHash := hashPassword(t, "newpassword")
	updatedUser, err := ur.Update(context.Background(), createdUser.ID, "Петр Updated", "Петров Updated", createdUser.Email, newPasswordHash)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if updatedUser.FirstName != "Петр Updated" {
		t.Errorf("Expected first name %s, got %s", "Петр Updated", updatedUser.FirstName)
	}
}

func TestUpdateUserEmail_DuplicateEmail(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	passwordHash1 := hashPassword(t, "pass1")
	_, err := ur.Create(context.Background(), "Первый", "Пользователь", "user1@test.com", passwordHash1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	passwordHash2 := hashPassword(t, "pass2")
	_, err = ur.Create(context.Background(), "Второй", "Пользователь", "user2@test.com", passwordHash2)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	_, err = ur.UpdateEmail(context.Background(), "user2@test.com", "user1@test.com")
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestUpdateUserEmail_SameEmail(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "password")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), "Мария", "Иванова", "maria@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	updatedUser, err := ur.UpdateEmail(context.Background(), createdUser.Email, "maria@test.com")
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
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "password")

	ur := user.NewUserRepository(db)
	_, err := ur.Create(context.Background(), "Дмитрий", "Сидоров", "dmitry@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = ur.UpdateEmail(context.Background(), "dmitry@test.com", "")
	if err == nil {
		t.Error("Expected error for empty email, got nil")
	}
}

func TestDeleteUser(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "password")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), "Удаляемый", "Пользователь", "delete@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	deletedUser, err := ur.Delete(context.Background(), createdUser.Email)
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
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "test")

	ur := user.NewUserRepository(db)

	_, err := ur.Create(context.Background(), "Тест", "Тестов", "duplicate@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	passwordHash2 := hashPassword(t, "test2")
	_, err = ur.Create(context.Background(), "Другой", "Пользователь", "duplicate@test.com", passwordHash2)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestCheckPassword_Success(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "correctpassword123")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), "Проверка", "Пароля", "checkpass@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Проверяем пароль с правильным хешем
	check, err := ur.CheckPassword(context.Background(), createdUser.Email, createdUser.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password: %v", err)
	}

	if !check {
		t.Error("Expected password check to return true for correct password hash, got false")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	passwordHash := hashPassword(t, "correctpassword")

	ur := user.NewUserRepository(db)
	createdUser, err := ur.Create(context.Background(), "Неправильный", "Пароль", "wrongpass@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Проверяем с неправильным хешем
	check, err := ur.CheckPassword(context.Background(), createdUser.Email, "wrong_hash_that_does_not_match")
	if err != nil {
		t.Fatalf("Failed to check password: %v", err)
	}

	if check {
		t.Error("Expected password check to return false for wrong password hash, got true")
	}
}

func TestCheckPassword_NonExistentUser(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	check, err := ur.CheckPassword(context.Background(), "nonexistent@test.com", "some_hash")
	if err != nil {
		t.Fatalf("Unexpected error when checking password for non-existent user: %v", err)
	}

	if check {
		t.Error("Expected password check to return false for non-existent user, got true")
	}
}

func TestCheckPassword_EmptyEmail(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	check, err := ur.CheckPassword(context.Background(), "", "some_hash")
	if err != nil {
		t.Fatalf("Unexpected error when checking password with empty email: %v", err)
	}

	if check {
		t.Error("Expected password check to return false for empty email, got true")
	}
}

func TestCheckPassword_DifferentUsersSamePassword(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)

	// Создаем первого пользователя
	passwordHash1 := hashPassword(t, "samepassword")
	createdUser1, err := ur.Create(context.Background(), "Первый", "Пользователь", "user1@test.com", passwordHash1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Создаем второго пользователя с тем же паролем
	passwordHash2 := hashPassword(t, "samepassword")
	createdUser2, err := ur.Create(context.Background(), "Второй", "Пользователь", "user2@test.com", passwordHash2)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	// Проверяем, что хеши разные (bcrypt генерирует разные хеши)
	if createdUser1.PasswordSalt == createdUser2.PasswordSalt {
		t.Error("Expected different password hashes for same password (bcrypt uses random salt), got same hash")
	}

	// Проверяем пароль первого пользователя
	check1, err := ur.CheckPassword(context.Background(), createdUser1.Email, createdUser1.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password for user1: %v", err)
	}
	if !check1 {
		t.Error("Expected password check to return true for user1, got false")
	}

	// Проверяем пароль второго пользователя
	check2, err := ur.CheckPassword(context.Background(), createdUser2.Email, createdUser2.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password for user2: %v", err)
	}
	if !check2 {
		t.Error("Expected password check to return true for user2, got false")
	}

	// Проверяем, что нельзя использовать хеш одного пользователя для другого
	wrongCheck, err := ur.CheckPassword(context.Background(), createdUser1.Email, createdUser2.PasswordSalt)
	if err != nil {
		t.Fatalf("Failed to check password: %v", err)
	}
	if wrongCheck {
		t.Error("Expected password check to return false when using wrong user's hash, got true")
	}
}
