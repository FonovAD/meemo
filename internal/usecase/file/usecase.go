package file

import (
	"context"
	"errors"
	"io"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/file/repository"
	"meemo/internal/domain/file/service"
	"meemo/internal/infrastructure/storage/s3/file"
	"time"
)

type Usecase interface {
	GetUserFilesList(ctx context.Context, in *GetAllUserFilesDtoIn) (*GetAllUserFilesDtoOut, error)
	GetFileInfo(ctx context.Context, in *GetFileInfoDtoIn) (*GetFileInfoDtoOut, error)
	RenameFile(ctx context.Context, in *RenameFileDtoIn) (*RenameFileDtoOut, error)
	DeleteFile(ctx context.Context, in *DeleteFileDtoIn) (*DeleteFileDtoOut, error)
	SaveFileMetadata(ctx context.Context, in *SaveFileMetadataDtoIn) (*SaveFileMetadataDtoOut, error)
	SaveFileContent(ctx context.Context, in *SaveFileContentDtoIn, inReader io.Reader) (*SaveFileContentDtoOut, error)
	GetFile(ctx context.Context, in *GetFileDtoIn, inWriter io.Writer) (*GetFileDtoOut, error)
}

type fileUsecase struct {
	fileRepo    repository.FileRepository
	s3Client    file.S3Client
	fileService service.FileService
}

func NewFileUsecase(fileRepo repository.FileRepository, fileService service.FileService, s3Client file.S3Client) *fileUsecase {
	return &fileUsecase{
		fileRepo:    fileRepo,
		s3Client:    s3Client,
		fileService: fileService,
	}
}

func (u *fileUsecase) SaveFileMetadata(ctx context.Context, in *SaveFileMetadataDtoIn) (*SaveFileMetadataDtoOut, error) {
	user := &entity.User{
		ID:    in.UserID,
		Email: in.UserEmail,
	}

	fileEntity := &entity.File{
		UserID:       in.UserID,
		OriginalName: in.OriginalName,
		MimeType:     in.MimeType,
		SizeInBytes:  in.SizeInBytes,
		S3Bucket:     in.S3Bucket,
		S3Key:        in.S3Key,
		Status:       in.Status,
		IsPublic:     in.IsPublic,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := u.s3Client.SaveFile(ctx, user, fileEntity); err != nil {
		return nil, err
	}

	savedFile, err := u.fileRepo.Save(ctx, user, fileEntity)
	if err != nil {
		return nil, err
	}

	return &SaveFileMetadataDtoOut{
		ID:           savedFile.ID,
		OriginalName: savedFile.OriginalName,
		MimeType:     savedFile.MimeType,
		SizeInBytes:  savedFile.SizeInBytes,
		Status:       savedFile.Status,
		CreatedAt:    savedFile.CreatedAt,
		IsPublic:     savedFile.IsPublic,
	}, nil
}

func (u *fileUsecase) SaveFileContent(ctx context.Context, in *SaveFileContentDtoIn, inReader io.Reader) (*SaveFileContentDtoOut, error) {
	if inReader == nil {
		return nil, errors.New("input reader is nil")
	}

	user := &entity.User{
		Email: in.Email,
	}

	fileEntity := &entity.File{
		ID: in.ID,
		R:  inReader,
	}

	if err := u.s3Client.SaveFile(ctx, user, fileEntity); err != nil {
		return nil, err
	}

	return &SaveFileContentDtoOut{
		LoadingResult: true,
	}, nil
}

func (u *fileUsecase) GetFile(ctx context.Context, in *GetFileDtoIn, inWriter io.Writer) (*GetFileDtoOut, error) {
	if inWriter == nil {
		return nil, errors.New("output writer is nil")
	}

	user := &entity.User{
		ID:    in.UserID,
		Email: in.UserEmail,
	}

	fileEntity := &entity.File{
		OriginalName: in.OriginalName,
		W:            inWriter,
	}

	metaFile, err := u.fileRepo.Get(ctx, user, fileEntity)
	if err != nil {
		return nil, err
	}

	if err := u.s3Client.GetFileByOriginalName(ctx, user, metaFile); err != nil {
		return nil, err
	}

	return &GetFileDtoOut{
		ID:           metaFile.ID,
		OriginalName: metaFile.OriginalName,
		MimeType:     metaFile.MimeType,
		SizeInBytes:  metaFile.SizeInBytes,
	}, nil
}

func (u *fileUsecase) GetFileInfo(ctx context.Context, in *GetFileInfoDtoIn) (*GetFileInfoDtoOut, error) {
	user := &entity.User{
		ID:    in.UserID,
		Email: in.UserEmail,
	}

	fileEntity := &entity.File{
		OriginalName: in.OriginalName,
	}

	metaFile, err := u.fileRepo.Get(ctx, user, fileEntity)
	if err != nil {
		return nil, err
	}

	return &GetFileInfoDtoOut{
		ID:           metaFile.ID,
		UserID:       metaFile.UserID,
		OriginalName: metaFile.OriginalName,
		MimeType:     metaFile.MimeType,
		SizeInBytes:  metaFile.SizeInBytes,
		Status:       metaFile.Status,
		CreatedAt:    metaFile.CreatedAt,
		UpdatedAt:    metaFile.UpdatedAt,
		IsPublic:     metaFile.IsPublic,
	}, nil
}

func (u *fileUsecase) DeleteFile(ctx context.Context, in *DeleteFileDtoIn) (*DeleteFileDtoOut, error) {
	user := &entity.User{
		ID:    in.UserID,
		Email: in.UserEmail,
	}

	fileEntity := &entity.File{
		OriginalName: in.OriginalName,
	}

	if err := u.s3Client.DeleteFile(ctx, user, fileEntity); err != nil {
		return nil, err
	}

	deletedFile, err := u.fileRepo.Delete(ctx, user, fileEntity)
	if err != nil {
		return nil, err
	}

	return &DeleteFileDtoOut{
		ID: deletedFile.ID,
	}, nil
}

func (u *fileUsecase) RenameFile(ctx context.Context, in *RenameFileDtoIn) (*RenameFileDtoOut, error) {
	user := &entity.User{
		ID:    in.UserID,
		Email: in.UserEmail,
	}

	fileEntity := &entity.File{
		OriginalName: in.OldName,
	}

	if err := u.s3Client.RenameFile(ctx, user, fileEntity, in.NewName); err != nil {
		return nil, err
	}
	renamedFile, err := u.fileRepo.Rename(ctx, user, fileEntity, in.NewName)
	if err != nil {
		return nil, err
	}

	return &RenameFileDtoOut{
		ID:        renamedFile.ID,
		OldName:   in.OldName,
		NewName:   renamedFile.OriginalName,
		UpdatedAt: renamedFile.UpdatedAt,
	}, nil
}

func (u *fileUsecase) GetUserFilesList(ctx context.Context, in *GetAllUserFilesDtoIn) (*GetAllUserFilesDtoOut, error) {
	return nil, errors.New("not implemented: GetUserFilesList requires List method in FileRepository")
}
