package s3

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gopkg.in/yaml.v3"
	"log"
	"meemo/internal/domain/entity"
	"meemo/internal/infrastructure/storage/s3"
	"meemo/internal/infrastructure/storage/s3/file"
	"os"
	"strings"
	"testing"
	"time"
)

type TestS3Client struct {
	testBucket string
	*file.S3ClientImpl
}

func setupTestS3Client(t *testing.T) *TestS3Client {
	configFile, err := os.ReadFile("s3_config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg s3.Config
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		t.Fatalf("Failed to create MinIO client: %v", err)
	}

	testBucket := fmt.Sprintf("test-bucket-%d", time.Now().UnixNano())

	err = minioClient.MakeBucket(t.Context(), testBucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(t.Context(), testBucket)
		if !(exists && errBucketExists == nil) {
			t.Fatalf("Failed to create test bucket: %v", err)
		}
	}

	return &TestS3Client{
		S3ClientImpl: &file.S3ClientImpl{
			BucketName: testBucket,
			Client:     minioClient,
		},
		testBucket: testBucket,
	}
}

func (ts3 *TestS3Client) cleanup(t *testing.T) {

	t.Logf("Cleaning up bucket: %s", ts3.testBucket)

	objectsCh := ts3.Client.ListObjects(t.Context(), ts3.testBucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	removeErrors := 0
	for object := range objectsCh {
		if object.Err != nil {
			t.Logf("Error listing object: %v", object.Err)
			continue
		}

		err := ts3.Client.RemoveObject(t.Context(), ts3.testBucket, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			t.Logf("Error removing object %s: %v", object.Key, err)
			removeErrors++
		} else {
			t.Logf("Removed object: %s", object.Key)
		}
	}

	if removeErrors > 0 {
		t.Logf("Failed to remove %d objects", removeErrors)
	}
	time.Sleep(100 * time.Millisecond)

	err := ts3.Client.RemoveBucket(t.Context(), ts3.testBucket)
	if err != nil {
		t.Logf("Failed to remove bucket %s: %v. Attempting force cleanup...", ts3.testBucket, err)

		ts3.forceCleanupBucket(t)

		err = ts3.Client.RemoveBucket(t.Context(), ts3.testBucket)
		if err != nil {
			t.Logf("Final bucket removal failed for %s: %v", ts3.testBucket, err)
		}
	} else {
		t.Logf("Successfully removed bucket: %s", ts3.testBucket)
	}
}

func (ts3 *TestS3Client) forceCleanupBucket(t *testing.T) {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)

		listCh := ts3.Client.ListObjects(t.Context(), ts3.testBucket, minio.ListObjectsOptions{
			Recursive: true,
		})

		for object := range listCh {
			if object.Err != nil {
				t.Logf("Error listing object: %v", object.Err)
				continue
			}
			objectsCh <- object
		}
	}()

	errorsCh := ts3.Client.RemoveObjects(t.Context(), ts3.testBucket, objectsCh, minio.RemoveObjectsOptions{})
	errorCount := 0

	for removeErr := range errorsCh {
		t.Logf("Force remove error: %v", removeErr.Err)
		errorCount++
	}

	if errorCount > 0 {
		t.Logf("Encountered %d errors during force cleanup", errorCount)
	}
}

func createTestUser(email string) *entity.User {
	return &entity.User{
		Email: email,
	}
}

func createTestFile(originalName string, size int64) *entity.File {
	return &entity.File{
		OriginalName: originalName,
		SizeInBytes:  size,
	}
}

func TestS3Client_SaveAndGetFile(t *testing.T) {
	ts3 := setupTestS3Client(t)
	defer ts3.cleanup(t)

	user := createTestUser("test@example.com")
	file := createTestFile("test-file.txt", 33)

	testContent := "Hello, this is test file content!"
	file.R = strings.NewReader(testContent)

	t.Run("SaveFile_Success", func(t *testing.T) {
		err := ts3.SaveFile(t.Context(), user, file)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}
	})

	t.Run("GetFile_Success", func(t *testing.T) {
		var buf bytes.Buffer
		file.W = &buf
		err := ts3.GetFileByOriginalName(t.Context(), user, file)
		if err != nil {
			t.Fatalf("Failed to get file: %v", err)
		}

		if buf.String() != testContent {
			t.Errorf("File content mismatch. Expected: %s, Got: %s", testContent, buf.String())
		}
	})
}

func TestS3Client_DeleteFile(t *testing.T) {
	ts3 := setupTestS3Client(t)
	defer ts3.cleanup(t)

	user := createTestUser("test@example.com")
	file := createTestFile("to-delete.txt", 17)

	file.R = strings.NewReader("content to delete")
	err := ts3.SaveFile(t.Context(), user, file)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("DeleteFile_Success", func(t *testing.T) {
		err := ts3.DeleteFile(t.Context(), user, file)
		if err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}
	})
}

func TestS3Client_RenameFile(t *testing.T) {
	ts3 := setupTestS3Client(t)
	defer ts3.cleanup(t)

	user := createTestUser("test@example.com")
	originalFile := createTestFile("old-name.txt", 16)
	newName := "new-name.txt"

	originalContent := "original content"
	originalFile.R = strings.NewReader(originalContent)
	err := ts3.SaveFile(t.Context(), user, originalFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("RenameFile_Success", func(t *testing.T) {
		err := ts3.RenameFile(t.Context(), user, originalFile, newName)
		if err != nil {
			t.Fatalf("Failed to rename file: %v", err)
		}

		var buf bytes.Buffer
		newfile := createTestFile(newName, 100)
		newfile.W = &buf

		err = ts3.GetFileByOriginalName(t.Context(), user, newfile)
		if err != nil {
			t.Fatalf("Failed to get renamed file: %v", err)
		}
	})
}

func TestS3Client_Cleanup_Effectiveness(t *testing.T) {
	ts3 := setupTestS3Client(t)
	testBucket := ts3.testBucket // сохраняем имя bucket'а

	user := createTestUser("cleanup-test@example.com")

	files := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, filename := range files {
		file := createTestFile(filename, 12)
		file.R = strings.NewReader("test content")
		err := ts3.SaveFile(t.Context(), user, file)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	ts3.cleanup(t)

	exists, err := ts3.Client.BucketExists(t.Context(), testBucket)
	if err == nil && exists {
		t.Error("Bucket still exists after cleanup - cleanup failed!")
	} else {
		t.Logf("Cleanup successful - bucket %s removed", testBucket)
	}
}

func TestS3Client_Parallel(t *testing.T) {
	t.Run("ParallelOperations", func(t *testing.T) {
		t.Parallel()

		ts3 := setupTestS3Client(t)
		defer ts3.cleanup(t)

		user := createTestUser("parallel@example.com")
		file := createTestFile("parallel-test.txt", 16)
		file.R = strings.NewReader("parallel content")

		err := ts3.SaveFile(t.Context(), user, file)
		if err != nil {
			t.Errorf("Parallel save failed: %v", err)
		}
	})

	t.Run("AnotherParallel", func(t *testing.T) {
		t.Parallel()

		ts3 := setupTestS3Client(t)
		defer ts3.cleanup(t)

		user := createTestUser("parallel2@example.com")
		file := createTestFile("another-test.txt", 14)
		file.R = strings.NewReader("another content")

		err := ts3.SaveFile(t.Context(), user, file)
		if err != nil {
			t.Errorf("Another parallel save failed: %v", err)
		}
	})
}
