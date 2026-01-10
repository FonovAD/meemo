package db_postgres

import (
	"context"
	"meemo/internal/infrastructure/storage/pg/file"
	"meemo/internal/infrastructure/storage/pg/user"
	"testing"
)

func TestSaveFile_Success(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "fileuser@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	savedFile, err := fr.Save(context.Background(), testUser.ID, "test_document.pdf", "application/pdf", "test-bucket", "files/12345.pdf", 1024, false)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	if savedFile.ID == 0 {
		t.Error("Expected file ID to be set")
	}
	if savedFile.OriginalName != "test_document.pdf" {
		t.Errorf("Expected original name %s, got %s", "test_document.pdf", savedFile.OriginalName)
	}
	if savedFile.UserID != testUser.ID {
		t.Errorf("Expected user ID %d, got %d", testUser.ID, savedFile.UserID)
	}
}

func TestSaveFile_DuplicateName(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "duplicate@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	_, err = fr.Save(context.Background(), testUser.ID, "duplicate.txt", "text/plain", "test-bucket", "files/1.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save first file: %v", err)
	}

	_, err = fr.Save(context.Background(), testUser.ID, "duplicate.txt", "text/plain", "test-bucket", "files/2.txt", 200, true)
	if err == nil {
		t.Error("Expected error for duplicate file name, got nil")
	}
}

func TestGetFile_Success(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "getfile@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	savedFile, err := fr.Save(context.Background(), testUser.ID, "get_test.jpg", "image/jpeg", "test-bucket", "files/get_test.jpg", 2048, true)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	foundFile, err := fr.Get(context.Background(), savedFile.ID)
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

func TestGetFileByOriginalNameAndUserEmail(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "getbyname@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	savedFile, err := fr.Save(context.Background(), testUser.ID, "byname_test.jpg", "image/jpeg", "test-bucket", "files/byname_test.jpg", 2048, true)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	foundFile, err := fr.GetByOriginalNameAndUserEmail(context.Background(), testUser.Email, "byname_test.jpg")
	if err != nil {
		t.Fatalf("Failed to get file by name and email: %v", err)
	}

	if foundFile.ID != savedFile.ID {
		t.Errorf("Expected file ID %d, got %d", savedFile.ID, foundFile.ID)
	}
}

func TestGetFile_NotFound(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	fr := file.NewFileRepository(db)

	_, err := fr.Get(context.Background(), 999999)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestGetFile_WrongUser(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")

	user1, err := ur.Create(context.Background(), "Test", "User1", "user1@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	_, err = ur.Create(context.Background(), "Test", "User2", "user2@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	fr := file.NewFileRepository(db)

	_, err = fr.Save(context.Background(), user1.ID, "private.txt", "text/plain", "test-bucket", "files/private.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Попытка получить файл user1 по email user2
	_, err = fr.GetByOriginalNameAndUserEmail(context.Background(), "user2@test.com", "private.txt")
	if err == nil {
		t.Error("Expected error when accessing other user's file, got nil")
	}
}

func TestDeleteFile_Success(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "delete@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	savedFile, err := fr.Save(context.Background(), testUser.ID, "to_delete.pdf", "application/pdf", "test-bucket", "files/to_delete.pdf", 512, false)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	deletedFile, err := fr.Delete(context.Background(), testUser.Email, savedFile.OriginalName)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	if deletedFile.ID != savedFile.ID {
		t.Errorf("Expected deleted file ID %d, got %d", savedFile.ID, deletedFile.ID)
	}

	_, err = fr.Get(context.Background(), savedFile.ID)
	if err == nil {
		t.Error("Expected error when getting deleted file, got nil")
	}
}

func TestDeleteFile_NotFound(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "deletenotfound@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	_, err = fr.Delete(context.Background(), testUser.Email, "nonexistent.txt")
	if err == nil {
		t.Error("Expected error when deleting non-existent file, got nil")
	}
}

func TestChangeVisibility_Success(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "visibility@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	savedFile, err := fr.Save(context.Background(), testUser.ID, "visibility_test.txt", "text/plain", "test-bucket", "files/visibility.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	updatedFile, err := fr.ChangeVisibility(context.Background(), testUser.Email, savedFile.OriginalName, true)
	if err != nil {
		t.Fatalf("Failed to change visibility: %v", err)
	}

	if !updatedFile.IsPublic {
		t.Error("Expected file to be public after visibility change")
	}

	foundFile, err := fr.GetByOriginalNameAndUserEmail(context.Background(), testUser.Email, "visibility_test.txt")
	if err != nil {
		t.Fatalf("Failed to get file after visibility change: %v", err)
	}

	if !foundFile.IsPublic {
		t.Error("Expected file to be public in database")
	}
}

func TestSetStatus_Success(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "status@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	savedFile, err := fr.Save(context.Background(), testUser.ID, "status_test.txt", "text/plain", "test-bucket", "files/status.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	newStatus := 1
	updatedFile, err := fr.SetStatus(context.Background(), testUser.Email, savedFile.OriginalName, newStatus)
	if err != nil {
		t.Fatalf("Failed to set status: %v", err)
	}

	if updatedFile.Status != newStatus {
		t.Errorf("Expected status %d, got %d", newStatus, updatedFile.Status)
	}

	foundFile, err := fr.GetByOriginalNameAndUserEmail(context.Background(), testUser.Email, "status_test.txt")
	if err != nil {
		t.Fatalf("Failed to get file after status change: %v", err)
	}

	if foundFile.Status != newStatus {
		t.Errorf("Expected status %d in database, got %d", newStatus, foundFile.Status)
	}
}

func TestMultipleUsersSameFileName(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")

	user1, err := ur.Create(context.Background(), "Test", "User1", "multi1@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := ur.Create(context.Background(), "Test", "User2", "multi2@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	fr := file.NewFileRepository(db)

	saved1, err := fr.Save(context.Background(), user1.ID, "same_name.txt", "text/plain", "test-bucket", "files/user1_same.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save file for user1: %v", err)
	}

	saved2, err := fr.Save(context.Background(), user2.ID, "same_name.txt", "text/plain", "test-bucket", "files/user2_same.txt", 200, true)
	if err != nil {
		t.Fatalf("Failed to save file for user2: %v", err)
	}

	if saved1.ID == saved2.ID {
		t.Error("Expected different file IDs for different users")
	}

	found1, err := fr.GetByOriginalNameAndUserEmail(context.Background(), user1.Email, "same_name.txt")
	if err != nil {
		t.Fatalf("Failed to get file for user1: %v", err)
	}
	if found1.ID != saved1.ID {
		t.Errorf("User1 got wrong file ID: expected %d, got %d", saved1.ID, found1.ID)
	}

	found2, err := fr.GetByOriginalNameAndUserEmail(context.Background(), user2.Email, "same_name.txt")
	if err != nil {
		t.Fatalf("Failed to get file for user2: %v", err)
	}
	if found2.ID != saved2.ID {
		t.Errorf("User2 got wrong file ID: expected %d, got %d", saved2.ID, found2.ID)
	}
}

func TestRenameFile_Success(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "rename@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	originalName := "old_name.txt"
	savedFile, err := fr.Save(context.Background(), testUser.ID, originalName, "text/plain", "test-bucket", "files/old_name.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	newName := "new_name.txt"
	renamedFile, err := fr.Rename(context.Background(), testUser.Email, originalName, newName)
	if err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	if renamedFile.OriginalName != newName {
		t.Errorf("Expected renamed file name %s, got %s", newName, renamedFile.OriginalName)
	}

	_, err = fr.GetByOriginalNameAndUserEmail(context.Background(), testUser.Email, originalName)
	if err == nil {
		t.Error("Expected error when fetching file by old name after rename, got nil")
	}

	foundFile, err := fr.GetByOriginalNameAndUserEmail(context.Background(), testUser.Email, newName)
	if err != nil {
		t.Fatalf("Failed to get renamed file: %v", err)
	}
	if foundFile.ID != savedFile.ID {
		t.Errorf("Expected file ID %d after rename, got %d", savedFile.ID, foundFile.ID)
	}
}

func TestRenameFile_WrongUser(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")

	user1, err := ur.Create(context.Background(), "Test", "Owner", "owner@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := ur.Create(context.Background(), "Test", "Attacker", "attacker@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	fr := file.NewFileRepository(db)

	savedFile, err := fr.Save(context.Background(), user1.ID, "private_file.txt", "text/plain", "test-bucket", "files/private.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save file for user1: %v", err)
	}

	_, err = fr.Rename(context.Background(), user2.Email, savedFile.OriginalName, "hacked.txt")
	if err == nil {
		t.Error("Expected error when user2 tries to rename user1's file, got nil")
	}

	foundFile, err := fr.GetByOriginalNameAndUserEmail(context.Background(), user1.Email, "private_file.txt")
	if err != nil {
		t.Fatalf("File of user1 disappeared or was renamed unexpectedly: %v", err)
	}
	if foundFile.ID != savedFile.ID {
		t.Error("File ID mismatch after unauthorized rename attempt")
	}
}

func TestRenameFile_NameConflict(t *testing.T) {
	db, teardown := initTestDB(t)
	defer teardown()

	ur := user.NewUserRepository(db)
	passwordHash := hashPassword(t, "password")
	testUser, err := ur.Create(context.Background(), "Test", "User", "rename_conflict@test.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	fr := file.NewFileRepository(db)

	saved1, err := fr.Save(context.Background(), testUser.ID, "file1.txt", "text/plain", "test-bucket", "files/file1.txt", 100, false)
	if err != nil {
		t.Fatalf("Failed to save file1: %v", err)
	}

	_, err = fr.Save(context.Background(), testUser.ID, "file2.txt", "text/plain", "test-bucket", "files/file2.txt", 200, false)
	if err != nil {
		t.Fatalf("Failed to save file2: %v", err)
	}

	_, err = fr.Rename(context.Background(), testUser.Email, "file1.txt", "file2.txt")
	if err == nil {
		t.Error("Expected error due to duplicate file name after rename, got nil")
	}

	found1, err := fr.GetByOriginalNameAndUserEmail(context.Background(), testUser.Email, "file1.txt")
	if err != nil {
		t.Fatalf("file1 disappeared after failed rename: %v", err)
	}
	if found1.ID != saved1.ID {
		t.Error("file1 ID changed unexpectedly")
	}
}
