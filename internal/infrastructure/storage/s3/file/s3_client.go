package file

import (
	"context"
	"io"
	"strconv"

	"meemo/internal/infrastructure/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
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

func NewS3Client(client *s3.Client, bucketName string, log logger.Logger) S3Client {
	return &S3ClientImpl{
		BucketName: bucketName,
		Client:     client,
		log:        log,
	}
}

type S3ClientImpl struct {
	BucketName string
	Client     *s3.Client
	log        logger.Logger
}

func (s3Client *S3ClientImpl) SaveFile(ctx context.Context, fileID int64, fileReader io.Reader, sizeInBytes int64) error {
	s3Client.log.Debug("saving file to S3", zap.Int64("fileID", fileID), zap.Int64("sizeInBytes", sizeInBytes))

	key := strconv.FormatInt(fileID, 10)
	contentType := "application/octet-stream"

	_, err := s3Client.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s3Client.BucketName),
		Key:           aws.String(key),
		Body:          fileReader,
		ContentLength: aws.Int64(sizeInBytes),
		ContentType:   aws.String(contentType),
	})

	if err != nil {
		s3Client.log.Error("failed to upload file to S3", zap.Int64("fileID", fileID), zap.Error(err))
		return err
	}

	s3Client.log.Info("file uploaded successfully", zap.String("key", key))
	return nil
}

func (s3Client *S3ClientImpl) GetFileByID(ctx context.Context, fileID int64, inWriter io.Writer) error {
	key := strconv.FormatInt(fileID, 10)

	result, err := s3Client.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3Client.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		s3Client.log.Error("failed to get file from S3", zap.Int64("fileID", fileID), zap.Error(err))
		return err
	}
	defer func() { _ = result.Body.Close() }()

	_, err = io.Copy(inWriter, result.Body)
	if err != nil {
		s3Client.log.Error("failed to copy file content", zap.Int64("fileID", fileID), zap.Error(err))
		return err
	}
	return nil
}

func (s3Client *S3ClientImpl) GetFileByOriginalName(ctx context.Context, userEmail, originalName string, inWriter io.Writer) error {
	key := userEmail + originalName

	result, err := s3Client.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3Client.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		s3Client.log.Error("failed to get file by name from S3", zap.String("key", key), zap.Error(err))
		return err
	}
	defer func() { _ = result.Body.Close() }()

	_, err = io.Copy(inWriter, result.Body)
	if err != nil {
		s3Client.log.Error("failed to copy file content", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

func (s3Client *S3ClientImpl) DeleteFile(ctx context.Context, fileID int64) error {
	key := strconv.FormatInt(fileID, 10)

	_, err := s3Client.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3Client.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		s3Client.log.Error("failed to delete file from S3", zap.Int64("fileID", fileID), zap.Error(err))
		return err
	}

	s3Client.log.Info("file deleted from S3", zap.Int64("fileID", fileID))
	return nil
}

func (s3Client *S3ClientImpl) RenameFile(ctx context.Context, userEmail, originalName, newName string) error {
	srcKey := userEmail + originalName
	dstKey := userEmail + newName

	copySource := s3Client.BucketName + "/" + srcKey
	_, err := s3Client.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s3Client.BucketName),
		CopySource: aws.String(copySource),
		Key:        aws.String(dstKey),
	})
	if err != nil {
		s3Client.log.Error("failed to copy file in S3", zap.String("srcKey", srcKey), zap.String("dstKey", dstKey), zap.Error(err))
		return err
	}

	_, err = s3Client.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3Client.BucketName),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		s3Client.log.Error("failed to delete source file in S3 after copy", zap.String("srcKey", srcKey), zap.Error(err))
		return err
	}

	s3Client.log.Info("file renamed in S3", zap.String("srcKey", srcKey), zap.String("dstKey", dstKey))
	return nil
}

func (s3Client *S3ClientImpl) CreateBucket(ctx context.Context, bucketName string) error {
	_, err := s3Client.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s3Client.Client.Options().Region),
		},
	})
	if err != nil {
		s3Client.log.Error("failed to create bucket", zap.String("bucketName", bucketName), zap.Error(err))
		return err
	}

	s3Client.log.Info("bucket created", zap.String("bucketName", bucketName))
	return nil
}

func (s3Client *S3ClientImpl) DeleteBucket(ctx context.Context, bucketName string) error {
	_, err := s3Client.Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		s3Client.log.Error("failed to delete bucket", zap.String("bucketName", bucketName), zap.Error(err))
		return err
	}

	s3Client.log.Info("bucket deleted", zap.String("bucketName", bucketName))
	return nil
}
