package file

import (
	"context"
	"github.com/minio/minio-go/v7"
	"io"
	"log"
	"strconv"
)

type S3Client interface {
	SaveFile(ctx context.Context, fileID int64, fileReader io.Reader, sizeInBytes int64) error
	GetFileByID(ctx context.Context, fileID int64, inWriter io.Writer) error
	GetFileByOriginalName(ctx context.Context, userEmail, originalName string, inWriter io.Writer) error
	DeleteFile(ctx context.Context, fileID int64) error
	RenameFile(ctx context.Context, userEmail, originalName, newName string) error
	CreateBucket(ctx context.Context, bucketName string) error
	DeleteBucket(ctx context.Context, bucketName string) error
}

func NewS3Client(client *minio.Client, bucketName string) S3Client {
	return &S3ClientImpl{
		BucketName: bucketName,
		Client:     client,
	}
}

type S3ClientImpl struct {
	BucketName string
	Client     *minio.Client
}

func (s3 *S3ClientImpl) SaveFile(ctx context.Context, fileID int64, fileReader io.Reader, sizeInBytes int64) error {
	log.Println("fileID:", fileID, "sizeInBytes:", sizeInBytes)
	info, err := s3.Client.PutObject(ctx, s3.BucketName, strconv.FormatInt(fileID, 10), fileReader, sizeInBytes, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	log.Println(info)
	log.Println("err: ", err)
	return err
}

func (s3 *S3ClientImpl) GetFileByID(ctx context.Context, fileID int64, inWriter io.Writer) error {
	objectReader, err := s3.Client.GetObject(ctx, s3.BucketName, strconv.FormatInt(fileID, 10), minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer objectReader.Close()

	_, err = io.Copy(inWriter, objectReader)
	if err != nil {
		return err
	}
	return nil
}

func (s3 *S3ClientImpl) GetFileByOriginalName(ctx context.Context, userEmail, originalName string, inWriter io.Writer) error {
	objectReader, err := s3.Client.GetObject(ctx, s3.BucketName, userEmail+originalName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer objectReader.Close()

	_, err = io.Copy(inWriter, objectReader)
	if err != nil {
		return err
	}
	return nil
}

func (s3 *S3ClientImpl) DeleteFile(ctx context.Context, fileID int64) error {
	err := s3.Client.RemoveObject(ctx, s3.BucketName, strconv.FormatInt(fileID, 10), minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s3 *S3ClientImpl) RenameFile(ctx context.Context, userEmail, originalName, newName string) error {
	src := minio.CopySrcOptions{
		Bucket: s3.BucketName,
		Object: userEmail + originalName,
	}

	dst := minio.CopyDestOptions{
		Bucket: s3.BucketName,
		Object: userEmail + newName,
	}

	_, err := s3.Client.CopyObject(ctx, dst, src)
	if err != nil {
		return err
	}

	err = s3.Client.RemoveObject(ctx, s3.BucketName, userEmail+originalName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s3 *S3ClientImpl) CreateBucket(ctx context.Context, bucketName string) error {
	err := s3.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s3 *S3ClientImpl) DeleteBucket(ctx context.Context, bucketName string) error {
	err := s3.Client.RemoveBucket(ctx, bucketName)
	if err != nil {
		return err
	}
	return nil
}
