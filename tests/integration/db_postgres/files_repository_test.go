package db_postgres

import (
	"context"
	"github.com/jmoiron/sqlx"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/user/service"
	"meemo/internal/infrastructure/storage/pg/file"
	"meemo/internal/infrastructure/storage/pg/user"
	"testing"
)

func TestSaveFile_Success(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	testUser := createTestUser(t, db, "fileuser@test.com")

	fr := file.NewFileRepository(db)

	testFile := &entity.File{
		OriginalName: "test_document.pdf",
		MimeType:     "application/pdf",
		SizeInBytes:  1024,
		S3Bucket:     "test-bucket",
		S3Key:        "files/12345.pdf",
		Status:       0,
		IsPublic:     false,
	}

	savedFile, err := fr.SaveFile(t.Context(), testUser, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	if savedFile.ID == 0 {
		t.Error("Expected file ID to be set")
	}
	if savedFile.OriginalName != testFile.OriginalName {
		t.Errorf("Expected original name %s, got %s", testFile.OriginalName, savedFile.OriginalName)
	}
	if savedFile.UserID != testUser.ID {
		t.Errorf("Expected user ID %d, got %d", testUser.ID, savedFile.UserID)
	}
}

func TestSaveFile_DuplicateName(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user := createTestUser(t, db, "duplicate@test.com")
	fr := file.NewFileRepository(db)

	testFile := &entity.File{
		OriginalName: "duplicate.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/1.txt",
		Status:       0,
		IsPublic:     false,
	}

	_, err := fr.SaveFile(context.Background(), user, testFile)
	if err != nil {
		t.Fatalf("Failed to save first file: %v", err)
	}

	duplicateFile := &entity.File{
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

	testFile := &entity.File{
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

	searchFile := &entity.File{OriginalName: "get_test.jpg"}
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

	searchFile := &entity.File{OriginalName: "nonexistent.txt"}
	_, err := fr.GetFile(context.Background(), user, searchFile)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestGetFile_WrongUser(t *testing.T) {
	db, teardown := initDB(t)
	defer teardown()

	user1 := createTestUser(t, db, "user1@test.com")
	user2 := createTestUser(t, db, "user2@test.com")
	fr := file.NewFileRepository(db)

	testFile := &entity.File{
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

	searchFile := &entity.File{OriginalName: "private.txt"}
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

	testFile := &entity.File{
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

	deletedFile, err := fr.DeleteFile(context.Background(), user, savedFile)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	if deletedFile.ID != savedFile.ID {
		t.Errorf("Expected deleted file ID %d, got %d", savedFile.ID, deletedFile.ID)
	}

	searchFile := &entity.File{OriginalName: "to_delete.pdf"}
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

	nonExistentFile := &entity.File{OriginalName: "nonexistent.txt"}
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

	testFile := &entity.File{
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

	updatedFile, err := fr.ChangeVisibility(context.Background(), user, savedFile, true)
	if err != nil {
		t.Fatalf("Failed to change visibility: %v", err)
	}

	if !updatedFile.IsPublic {
		t.Error("Expected file to be public after visibility change")
	}

	searchFile := &entity.File{OriginalName: "visibility_test.txt"}
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

	testFile := &entity.File{
		OriginalName: "status_test.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/status.txt",
		Status:       0,
		IsPublic:     false,
	}

	savedFile, err := fr.SaveFile(context.Background(), user, testFile)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	newStatus := 1
	updatedFile, err := fr.SetStatus(context.Background(), user, savedFile, newStatus)
	if err != nil {
		t.Fatalf("Failed to set status: %v", err)
	}

	if updatedFile.Status != newStatus {
		t.Errorf("Expected status %d, got %d", newStatus, updatedFile.Status)
	}

	searchFile := &entity.File{OriginalName: "status_test.txt"}
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

	user1 := createTestUser(t, db, "multi1@test.com")
	user2 := createTestUser(t, db, "multi2@test.com")
	fr := file.NewFileRepository(db)

	file1 := &entity.File{
		OriginalName: "same_name.txt",
		MimeType:     "text/plain",
		SizeInBytes:  100,
		S3Bucket:     "test-bucket",
		S3Key:        "files/user1_same.txt",
		Status:       0,
		IsPublic:     false,
	}

	file2 := &entity.File{
		OriginalName: "same_name.txt",
		MimeType:     "text/plain",
		SizeInBytes:  200,
		S3Bucket:     "test-bucket",
		S3Key:        "files/user2_same.txt",
		Status:       0,
		IsPublic:     true,
	}

	saved1, err := fr.SaveFile(context.Background(), user1, file1)
	if err != nil {
		t.Fatalf("Failed to save file for user1: %v", err)
	}

	saved2, err := fr.SaveFile(context.Background(), user2, file2)
	if err != nil {
		t.Fatalf("Failed to save file for user2: %v", err)
	}

	if saved1.ID == saved2.ID {
		t.Error("Expected different file IDs for different users")
	}

	searchFile1 := &entity.File{OriginalName: "same_name.txt"}
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

func createTestUser(t *testing.T, db *sqlx.DB, email string) *entity.User {
	ur := user.NewUserRepository(db)
	us := service.NewUserService()

	testUser := &entity.User{
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
