package s3

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"meemo/internal/infrastructure/logger"
	"meemo/internal/infrastructure/storage/s3/file"
)

func TestS3Client_SaveAndGetFile(t *testing.T) {
	s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
	defer cleanup()

	log, _ := logger.NewLogger("error")
	client := file.NewS3Client(s3Client, testBucket, log)

	testContent := "Hello, this is test file content!"
	fileID := int64(12345)

	t.Run("SaveFile_Success", func(t *testing.T) {
		reader := strings.NewReader(testContent)
		err := client.SaveFile(t.Context(), fileID, reader, int64(len(testContent)))
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}
	})

	t.Run("GetFile_Success", func(t *testing.T) {
		var buf bytes.Buffer
		err := client.GetFileByID(t.Context(), fileID, &buf)
		if err != nil {
			t.Fatalf("Failed to get file: %v", err)
		}

		if buf.String() != testContent {
			t.Errorf("File content mismatch. Expected: %s, Got: %s", testContent, buf.String())
		}
	})
}

func TestS3Client_GetFileByOriginalName(t *testing.T) {
	s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
	defer cleanup()

	log, _ := logger.NewLogger("error")
	client := file.NewS3Client(s3Client, testBucket, log)

	userEmail := "test@example.com"
	originalName := "testfile.txt"
	testContent := "File content by name"

	key := userEmail + originalName
	_, err := s3Client.PutObject(t.Context(), &s3.PutObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(testContent),
	})
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	t.Run("GetFileByOriginalName_Success", func(t *testing.T) {
		var buf bytes.Buffer
		err := client.GetFileByOriginalName(t.Context(), userEmail, originalName, &buf)
		if err != nil {
			t.Fatalf("Failed to get file by original name: %v", err)
		}

		if buf.String() != testContent {
			t.Errorf("File content mismatch. Expected: %s, Got: %s", testContent, buf.String())
		}
	})
}

func TestS3Client_DeleteFile(t *testing.T) {
	s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
	defer cleanup()

	log, _ := logger.NewLogger("error")
	client := file.NewS3Client(s3Client, testBucket, log)

	fileID := int64(99999)
	testContent := "content to delete"

	reader := strings.NewReader(testContent)
	err := client.SaveFile(t.Context(), fileID, reader, int64(len(testContent)))
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("DeleteFile_Success", func(t *testing.T) {
		err := client.DeleteFile(t.Context(), fileID)
		if err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}

		var buf bytes.Buffer
		err = client.GetFileByID(t.Context(), fileID, &buf)
		if err == nil {
			t.Error("Expected error when getting deleted file, got nil")
		}
	})
}

func TestS3Client_RenameFile(t *testing.T) {
	s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
	defer cleanup()

	log, _ := logger.NewLogger("error")
	client := file.NewS3Client(s3Client, testBucket, log)

	userEmail := "rename@example.com"
	originalName := "old-name.txt"
	newName := "new-name.txt"
	originalContent := "original content"

	key := userEmail + originalName
	_, err := s3Client.PutObject(t.Context(), &s3.PutObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(originalContent),
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("RenameFile_Success", func(t *testing.T) {
		err := client.RenameFile(t.Context(), userEmail, originalName, newName)
		if err != nil {
			t.Fatalf("Failed to rename file: %v", err)
		}

		var buf bytes.Buffer
		err = client.GetFileByOriginalName(t.Context(), userEmail, newName, &buf)
		if err != nil {
			t.Fatalf("Failed to get renamed file: %v", err)
		}

		if buf.String() != originalContent {
			t.Errorf("File content mismatch. Expected: %s, Got: %s", originalContent, buf.String())
		}

		var buf2 bytes.Buffer
		err = client.GetFileByOriginalName(t.Context(), userEmail, originalName, &buf2)
		if err == nil {
			t.Error("Expected error when getting file by old name, got nil")
		}
	})
}

func TestS3Client_MultipleFiles(t *testing.T) {
	s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
	defer cleanup()

	log, _ := logger.NewLogger("error")
	client := file.NewS3Client(s3Client, testBucket, log)

	fileIDs := []int64{1001, 1002, 1003}
	testContent := "test content"

	for _, fileID := range fileIDs {
		reader := strings.NewReader(testContent)
		err := client.SaveFile(t.Context(), fileID, reader, int64(len(testContent)))
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", fileID, err)
		}
	}

	for _, fileID := range fileIDs {
		var buf bytes.Buffer
		err := client.GetFileByID(t.Context(), fileID, &buf)
		if err != nil {
			t.Errorf("Failed to get file %d: %v", fileID, err)
		}
	}
}

func TestS3Client_Parallel(t *testing.T) {
	t.Run("ParallelOperations", func(t *testing.T) {
		t.Parallel()

		s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
		defer cleanup()

		log, _ := logger.NewLogger("error")
		client := file.NewS3Client(s3Client, testBucket, log)

		testContent := "parallel content"
		fileID := int64(2001)
		reader := strings.NewReader(testContent)

		err := client.SaveFile(t.Context(), fileID, reader, int64(len(testContent)))
		if err != nil {
			t.Errorf("Parallel save failed: %v", err)
		}
	})

	t.Run("AnotherParallel", func(t *testing.T) {
		t.Parallel()

		s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
		defer cleanup()

		log, _ := logger.NewLogger("error")
		client := file.NewS3Client(s3Client, testBucket, log)

		testContent := "another content"
		fileID := int64(2002)
		reader := strings.NewReader(testContent)

		err := client.SaveFile(t.Context(), fileID, reader, int64(len(testContent)))
		if err != nil {
			t.Errorf("Another parallel save failed: %v", err)
		}
	})
}

func TestS3Client_DifferentUsers(t *testing.T) {
	s3Client, testBucket, cleanup := SetupS3ClientForTest(t)
	defer cleanup()

	log, _ := logger.NewLogger("error")
	client := file.NewS3Client(s3Client, testBucket, log)

	user1Email := "user1@example.com"
	user2Email := "user2@example.com"
	fileName := "shared-name.txt"
	user1Content := "user1 content"
	user2Content := "user2 content"

	key1 := user1Email + fileName
	_, err := s3Client.PutObject(t.Context(), &s3.PutObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key1),
		Body:   strings.NewReader(user1Content),
	})
	if err != nil {
		t.Fatalf("Failed to save file for user1: %v", err)
	}

	key2 := user2Email + fileName
	_, err = s3Client.PutObject(t.Context(), &s3.PutObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(key2),
		Body:   strings.NewReader(user2Content),
	})
	if err != nil {
		t.Fatalf("Failed to save file for user2: %v", err)
	}

	var buf1 bytes.Buffer
	err = client.GetFileByOriginalName(t.Context(), user1Email, fileName, &buf1)
	if err != nil {
		t.Fatalf("Failed to get file for user1: %v", err)
	}
	if buf1.String() != user1Content {
		t.Errorf("User1 got wrong content: %s", buf1.String())
	}

	var buf2 bytes.Buffer
	err = client.GetFileByOriginalName(t.Context(), user2Email, fileName, &buf2)
	if err != nil {
		t.Fatalf("Failed to get file for user2: %v", err)
	}
	if buf2.String() != user2Content {
		t.Errorf("User2 got wrong content: %s", buf2.String())
	}
}
