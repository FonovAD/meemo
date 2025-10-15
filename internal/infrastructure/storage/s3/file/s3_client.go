package file

import (
	"context"
	"github.com/minio/minio-go/v7"
	"io"
	"meemo/internal/domain/entity"
)

type S3Client interface {
	SaveFile(ctx context.Context, user *entity.User, fileMetadata *entity.File, file io.Reader) error
	GetFileByOriginalName(ctx context.Context, user *entity.User, fileMetadata *entity.File, file io.Writer) error
	DeleteFile(ctx context.Context, user *entity.User, fileMetadata *entity.File) error
	RenameFile(ctx context.Context, user *entity.User, fileMetadata *entity.File, newName string) error
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

func (s3 *S3ClientImpl) SaveFile(ctx context.Context, user *entity.User, fileMetadata *entity.File, file io.Reader) error {
	_, err := s3.Client.PutObject(ctx, s3.BucketName, user.Email+fileMetadata.OriginalName, file, fileMetadata.SizeInBytes, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	return err
}

func (s3 *S3ClientImpl) GetFileByOriginalName(ctx context.Context, user *entity.User, fileMetadata *entity.File, file io.Writer) error {
	objectReader, err := s3.Client.GetObject(ctx, s3.BucketName, user.Email+fileMetadata.OriginalName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	defer objectReader.Close()

	_, err = io.Copy(file, objectReader)
	return err
}

func (s3 *S3ClientImpl) DeleteFile(ctx context.Context, user *entity.User, fileMetadata *entity.File) error {
	err := s3.Client.RemoveObject(ctx, s3.BucketName, user.Email+fileMetadata.OriginalName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s3 *S3ClientImpl) RenameFile(ctx context.Context, user *entity.User, fileMetadata *entity.File, newName string) error {
	src := minio.CopySrcOptions{
		Bucket: s3.BucketName,
		Object: user.Email + fileMetadata.OriginalName,
	}

	dst := minio.CopyDestOptions{
		Bucket: s3.BucketName,
		Object: user.Email + newName,
	}

	_, err := s3.Client.CopyObject(ctx, dst, src)
	if err != nil {
		return err
	}

	err = s3.Client.RemoveObject(ctx, s3.BucketName, user.Email+fileMetadata.OriginalName, minio.RemoveObjectOptions{})
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
