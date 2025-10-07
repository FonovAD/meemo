package integration

import (
	"context"
	"github.com/jmoiron/sqlx"
	"log"
	"meemo/internal/domain/model"
	"meemo/internal/domain/user/service"
	"meemo/internal/infradtructure/storage/pg/file"
	"meemo/internal/infradtructure/storage/pg/user"
	"testing"
)

func TestSaveFile_Success(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	// Создаем пользователя
	test_user := createTestUser(t, db, "fileuser@test.com")

	fr := file.NewFileRepository(db)

	log.Print("1")

	testFile := &model.File{
		OriginalName: "test_document.pdf",
		MimeType:     "application/pdf",
		SizeInBytes:  1024,
		S3Bucket:     "test-bucket",
		S3Key:        "files/12345.pdf",
		Status:       0,
		IsPublic:     false,
	}

	savedFile, err := fr.SaveFile(t.Context(), test_user, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	if savedFile.ID == 0 {
		t.Error("Expected file ID to be set")
	}
	if savedFile.OriginalName != testFile.OriginalName {
		t.Errorf("Expected original name %s, got %s", testFile.OriginalName, savedFile.OriginalName)
	}
	if savedFile.UserID != test_user.ID {
		t.Errorf("Expected user ID %d, got %d", test_user.ID, savedFile.UserID)
	}
}

func TestSaveFile_DuplicateName(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "duplicate@test.com")
	fr := file.NewFileRepository(db)

	testFile := &model.File{
		OriginalName: "duplicate.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/1.txt",
		Status:       0,
		IsPublic:     false,
	}

	// Первое сохранение
	_, err := fr.SaveFile(context.Background(), user, testFile)
	if err != nil {
		t.Fatalf("Failed to save first file: %v", err)
	}

	// Второе сохранение с тем же именем для того же пользователя
	duplicateFile := &model.File{
		OriginalName: "duplicate.txt", // То же имя
		MimeType:     "text/plain",
		SizeInBytes:  200,
		S3Bucket:     "test-bucket",
		S3Key:        "files/2.txt",
		Status:       0,
		IsPublic:     true,
	}

	_, err = fr.SaveFile(context.Background(), user, duplicateFile)
	if err == nil {
		t.Error("Expected error for duplicate file name, got nil")
	}
}

func TestGetFile_Success(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "getfile@test.com")
	fr := file.NewFileRepository(db)

	// Сначала сохраняем файл
	testFile := &model.File{
		OriginalName: "get_test.jpg",
		MimeType:     "image/jpeg",
		SizeInBytes:  2048,
		S3Bucket:     "test-bucket",
		S3Key:        "files/get_test.jpg",
		Status:       1,
		IsPublic:     true,
	}

	savedFile, err := fr.SaveFile(context.Background(), user, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Теперь получаем его
	searchFile := &model.File{OriginalName: "get_test.jpg"}
	foundFile, err := fr.GetFile(context.Background(), user, searchFile)
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}

	if foundFile.ID != savedFile.ID {
		t.Errorf("Expected file ID %d, got %d", savedFile.ID, foundFile.ID)
	}
	if foundFile.OriginalName != "get_test.jpg" {
		t.Errorf("Expected original name %s, got %s", "get_test.jpg", foundFile.OriginalName)
	}
	if foundFile.MimeType != "image/jpeg" {
		t.Errorf("Expected mime type %s, got %s", "image/jpeg", foundFile.MimeType)
	}
}

func TestGetFile_NotFound(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "notfound@test.com")
	fr := file.NewFileRepository(db)

	searchFile := &model.File{OriginalName: "nonexistent.txt"}
	_, err := fr.GetFile(context.Background(), user, searchFile)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestGetFile_WrongUser(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	// Создаем двух пользователей
	user1 := createTestUser(t, db, "user1@test.com")
	user2 := createTestUser(t, db, "user2@test.com")
	fr := file.NewFileRepository(db)

	// Сохраняем файл для первого пользователя
	testFile := &model.File{
		OriginalName: "private.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/private.txt",
		Status:       0,
		IsPublic:     false,
	}

	_, err := fr.SaveFile(context.Background(), user1, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Пытаемся получить файл второго пользователем
	searchFile := &model.File{OriginalName: "private.txt"}
	_, err = fr.GetFile(context.Background(), user2, searchFile)
	if err == nil {
		t.Error("Expected error when accessing other user's file, got nil")
	}
}

func TestDeleteFile_Success(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "delete@test.com")
	fr := file.NewFileRepository(db)

	// Сначала сохраняем файл
	testFile := &model.File{
		OriginalName: "to_delete.pdf",
		MimeType:     "application/pdf",
		SizeInBytes:  512,
		S3Bucket:     "test-bucket",
		S3Key:        "files/to_delete.pdf",
		Status:       0,
		IsPublic:     false,
	}

	savedFile, err := fr.SaveFile(context.Background(), user, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Удаляем файл
	deletedFile, err := fr.DeleteFile(context.Background(), user, savedFile)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	if deletedFile.ID != savedFile.ID {
		t.Errorf("Expected deleted file ID %d, got %d", savedFile.ID, deletedFile.ID)
	}

	// Проверяем, что файл действительно удален
	searchFile := &model.File{OriginalName: "to_delete.pdf"}
	_, err = fr.GetFile(context.Background(), user, searchFile)
	if err == nil {
		t.Error("Expected error when getting deleted file, got nil")
	}
}

func TestDeleteFile_NotFound(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "deletenotfound@test.com")
	fr := file.NewFileRepository(db)

	nonExistentFile := &model.File{OriginalName: "nonexistent.txt"}
	_, err := fr.DeleteFile(context.Background(), user, nonExistentFile)
	if err == nil {
		t.Error("Expected error when deleting non-existent file, got nil")
	}
}

func TestChangeVisibility_Success(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "visibility@test.com")
	fr := file.NewFileRepository(db)

	// Сохраняем файл с isPublic = false
	testFile := &model.File{
		OriginalName: "visibility_test.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/visibility.txt",
		Status:       0,
		IsPublic:     false,
	}

	savedFile, err := fr.SaveFile(context.Background(), user, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Меняем видимость на true
	updatedFile, err := fr.ChangeVisibility(context.Background(), user, savedFile, true)
	if err != nil {
		t.Fatalf("Failed to change visibility: %v", err)
	}

	if !updatedFile.IsPublic {
		t.Error("Expected file to be public after visibility change")
	}

	// Проверяем, что изменение сохранилось в БД
	searchFile := &model.File{OriginalName: "visibility_test.txt"}
	foundFile, err := fr.GetFile(context.Background(), user, searchFile)
	if err != nil {
		t.Fatalf("Failed to get file after visibility change: %v", err)
	}

	if !foundFile.IsPublic {
		t.Error("Expected file to be public in database")
	}
}

func TestSetStatus_Success(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "status@test.com")
	fr := file.NewFileRepository(db)

	testFile := &model.File{
		OriginalName: "status_test.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/status.txt",
		Status:       0, // начальный статус
		IsPublic:     false,
	}

	savedFile, err := fr.SaveFile(context.Background(), user, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Меняем статус на 1
	newStatus := 1
	updatedFile, err := fr.SetStatus(context.Background(), user, savedFile, newStatus)
	if err != nil {
		t.Fatalf("Failed to set status: %v", err)
	}

	if updatedFile.Status != newStatus {
		t.Errorf("Expected status %d, got %d", newStatus, updatedFile.Status)
	}

	// Проверяем в БД
	searchFile := &model.File{OriginalName: "status_test.txt"}
	foundFile, err := fr.GetFile(context.Background(), user, searchFile)
	if err != nil {
		t.Fatalf("Failed to get file after status change: %v", err)
	}

	if foundFile.Status != newStatus {
		t.Errorf("Expected status %d in database, got %d", newStatus, foundFile.Status)
	}
}

func TestMultipleUsersSameFileName(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	// Создаем двух пользователей
	user1 := createTestUser(t, db, "multi1@test.com")
	user2 := createTestUser(t, db, "multi2@test.com")
	fr := file.NewFileRepository(db)

	// Оба пользователя сохраняют файл с одинаковым именем
	file1 := &model.File{
		OriginalName: "same_name.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/user1_same.txt",
		Status:       0,
		IsPublic:     false,
	}

	file2 := &model.File{
		OriginalName: "same_name.txt", // То же имя файла
		MimeType:     "text/plain",
		SizeInBytes:  200,
		S3Bucket:     "test-bucket",
		S3Key:        "files/user2_same.txt",
		Status:       0,
		IsPublic:     true,
	}

	// Сохраняем для первого пользователя
	saved1, err := fr.SaveFile(context.Background(), user1, file1)
	if err != nil {
		t.Fatalf("Failed to save file for user1: %v", err)
	}

	// Сохраняем для второго пользователя
	saved2, err := fr.SaveFile(context.Background(), user2, file2)
	if err != nil {
		t.Fatalf("Failed to save file for user2: %v", err)
	}

	// Проверяем, что это разные файлы
	if saved1.ID == saved2.ID {
		t.Error("Expected different file IDs for different users")
	}

	// Каждый пользователь может получить свой файл
	searchFile1 := &model.File{OriginalName: "same_name.txt"}
	found1, err := fr.GetFile(context.Background(), user1, searchFile1)
	if err != nil {
		t.Fatalf("Failed to get file for user1: %v", err)
	}
	if found1.ID != saved1.ID {
		t.Errorf("User1 got wrong file ID: expected %d, got %d", saved1.ID, found1.ID)
	}

	found2, err := fr.GetFile(context.Background(), user2, searchFile1)
	if err != nil {
		t.Fatalf("Failed to get file for user2: %v", err)
	}
	if found2.ID != saved2.ID {
		t.Errorf("User2 got wrong file ID: expected %d, got %d", saved2.ID, found2.ID)
	}
}

func createTestUser(t *testing.T, db *sqlx.DB, email string) *model.User {
	ur := user.NewUserRepository(db)
	us := service.NewUserService()

	testUser := &model.User{
		FirstName: "Test",
		LastName:  "User",
		Email:     email,
	}
	us.HashPassword(testUser, "password")

	createdUser, err := ur.CreateUser(context.Background(), testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return createdUser
}
