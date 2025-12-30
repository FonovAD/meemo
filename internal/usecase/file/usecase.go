package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"meemo/internal/domain/entity"
	"meemo/internal/domain/file/repository"
	"meemo/internal/domain/file/service"
	"meemo/internal/infrastructure/storage/s3/file"
	"strconv"
)

type Usecase interface {
	GetUserFilesList(ctx context.Context, in *GetAllUserFilesDtoIn) (*GetAllUserFilesDtoOut, error)
	GetFileInfo(ctx context.Context, in *GetFileInfoDtoIn) (*GetFileInfoDtoOut, error)
	RenameFile(ctx context.Context, in *RenameFileDtoIn) (*RenameFileDtoOut, error)
	DeleteFile(ctx context.Context, in *DeleteFileDtoIn) (*DeleteFileDtoOut, error)
	SaveFileMetadata(ctx context.Context, in *SaveFileMetadataDtoIn) (*SaveFileMetadataDtoOut, error)
	SaveFileContent(ctx context.Context, in *SaveFileContentDtoIn, inReader io.Reader) (*SaveFileContentDtoOut, error)
	GetFile(ctx context.Context, in *GetFileDtoIn, inWriter io.Writer) (*GetFileDtoOut, error)
	GetFileMetadataByName(ctx context.Context, in *GetFileDtoIn) (*GetFileDtoOut, error)
	GetFileByID(ctx context.Context, in *GetFileByIDDtoIn, inWriter io.Writer) (*GetFileByIDDtoOut, error)
	GetFileMetadataByID(ctx context.Context, in *GetFileByIDDtoIn) (*GetFileByIDDtoOut, error)
	ChangeVisibility(ctx context.Context, in *ChangeVisibilityDtoIn) (*ChangeVisibilityDtoOut, error)
	SetStatus(ctx context.Context, in *SetStatusDtoIn) (*SetStatusDtoOut, error)
}

type fileUsecase struct {
	fileRepo    repository.FileRepository
	s3Client    file.S3Client
	fileService service.FileService
}

func NewFileUsecase(fileRepo repository.FileRepository, fileService service.FileService, s3Client file.S3Client) Usecase {
	return &fileUsecase{
		fileRepo:    fileRepo,
		s3Client:    s3Client,
		fileService: fileService,
	}
}

func (u *fileUsecase) SaveFileMetadata(ctx context.Context, in *SaveFileMetadataDtoIn) (*SaveFileMetadataDtoOut, error) {
	fileEntity := &entity.File{
		UserID:       in.UserID,
		OriginalName: in.OriginalName,
		IsPublic:     in.IsPublic,
		SizeInBytes:  in.SizeInBytes,
		MimeType:     in.MimeType,
	}

	u.fileService.CreateFileMetadata(fileEntity)
	fileEntity.S3Key = strconv.FormatInt(in.UserID, 10) + "/" + fileEntity.OriginalName

	savedFile, err := u.fileRepo.Save(ctx, in.UserID, fileEntity.OriginalName, fileEntity.MimeType, fileEntity.S3Bucket, fileEntity.S3Key, fileEntity.SizeInBytes, fileEntity.IsPublic)
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

	fmt.Println("fileID: ", in.ID, "sizeInBytes:", in.SizeInBytes)

	if err := u.s3Client.SaveFile(ctx, in.ID, inReader, in.SizeInBytes); err != nil {
		return nil, err
	}

	return &SaveFileContentDtoOut{
		LoadingResult: true,
	}, nil
}

func (u *fileUsecase) GetFileMetadataByName(ctx context.Context, in *GetFileDtoIn) (*GetFileDtoOut, error) {
	metaFile, err := u.fileRepo.GetByOriginalNameAndUserEmail(ctx, in.UserEmail, in.OriginalName)
	if err != nil {
		return nil, err
	}

	return &GetFileDtoOut{
		ID:           metaFile.ID,
		OriginalName: metaFile.OriginalName,
		MimeType:     metaFile.MimeType,
		SizeInBytes:  metaFile.SizeInBytes,
	}, nil
}

func (u *fileUsecase) GetFile(ctx context.Context, in *GetFileDtoIn, inWriter io.Writer) (*GetFileDtoOut, error) {
	if inWriter == nil {
		return nil, errors.New("output writer is nil")
	}

	metaFile, err := u.fileRepo.GetByOriginalNameAndUserEmail(ctx, in.UserEmail, in.OriginalName)
	if err != nil {
		return nil, err
	}

	if err := u.s3Client.GetFileByID(ctx, metaFile.ID, inWriter); err != nil {
		return nil, err
	}

	return &GetFileDtoOut{
		ID:           metaFile.ID,
		OriginalName: metaFile.OriginalName,
		MimeType:     metaFile.MimeType,
		SizeInBytes:  metaFile.SizeInBytes,
	}, nil
}

func (u *fileUsecase) getFileMetadataAndCheckAccess(ctx context.Context, fileID, userID int64) (*entity.File, error) {
	metaFile, err := u.fileRepo.Get(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if metaFile.UserID != userID && !metaFile.IsPublic {
		return nil, errors.New("access denied: file is private")
	}

	return metaFile, nil
}

// GetFileMetadataByID получает только метаданные файла без загрузки содержимого
func (u *fileUsecase) GetFileMetadataByID(ctx context.Context, in *GetFileByIDDtoIn) (*GetFileByIDDtoOut, error) {
	metaFile, err := u.getFileMetadataAndCheckAccess(ctx, in.FileID, in.UserID)
	if err != nil {
		return nil, err
	}

	return &GetFileByIDDtoOut{
		ID:           metaFile.ID,
		OriginalName: metaFile.OriginalName,
		MimeType:     metaFile.MimeType,
		SizeInBytes:  metaFile.SizeInBytes,
	}, nil
}

func (u *fileUsecase) GetFileByID(ctx context.Context, in *GetFileByIDDtoIn, inWriter io.Writer) (*GetFileByIDDtoOut, error) {
	if inWriter == nil {
		return nil, errors.New("output writer is nil")
	}

	metaFile, err := u.getFileMetadataAndCheckAccess(ctx, in.FileID, in.UserID)
	if err != nil {
		return nil, err
	}

	if err := u.s3Client.GetFileByID(ctx, in.FileID, inWriter); err != nil {
		return nil, err
	}

	return &GetFileByIDDtoOut{
		ID:           metaFile.ID,
		OriginalName: metaFile.OriginalName,
		MimeType:     metaFile.MimeType,
		SizeInBytes:  metaFile.SizeInBytes,
	}, nil
}

func (u *fileUsecase) GetFileInfo(ctx context.Context, in *GetFileInfoDtoIn) (*GetFileInfoDtoOut, error) {
	metaFile, err := u.fileRepo.GetByOriginalNameAndUserEmail(ctx, in.UserEmail, in.OriginalName)
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
	metaFile, err := u.fileRepo.GetByOriginalNameAndUserEmail(ctx, in.UserEmail, in.OriginalName)
	if err != nil {
		return nil, err
	}

	if err := u.s3Client.DeleteFile(ctx, metaFile.ID); err != nil {
		return nil, err
	}

	deletedFile, err := u.fileRepo.Delete(ctx, in.UserEmail, in.OriginalName)
	if err != nil {
		return nil, err
	}

	return &DeleteFileDtoOut{
		ID: deletedFile.ID,
	}, nil
}

func (u *fileUsecase) RenameFile(ctx context.Context, in *RenameFileDtoIn) (*RenameFileDtoOut, error) {
	if err := u.s3Client.RenameFile(ctx, in.UserEmail, in.OldName, in.NewName); err != nil {
		return nil, err
	}
	renamedFile, err := u.fileRepo.Rename(ctx, in.UserEmail, in.OldName, in.NewName)
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
	files, err := u.fileRepo.List(ctx, in.UserEmail)
	if err != nil {
		return nil, err
	}

	fileList := make([]FileListItemDto, 0, len(files))
	for _, file := range files {
		fileList = append(fileList, FileListItemDto{
			ID:           file.ID,
			OriginalName: file.OriginalName,
			MimeType:     file.MimeType,
			SizeInBytes:  file.SizeInBytes,
			Status:       file.Status,
			CreatedAt:    file.CreatedAt,
			UpdatedAt:    file.UpdatedAt,
			IsPublic:     file.IsPublic,
		})
	}

	return &GetAllUserFilesDtoOut{
		Files: fileList,
	}, nil
}

func (u *fileUsecase) ChangeVisibility(ctx context.Context, in *ChangeVisibilityDtoIn) (*ChangeVisibilityDtoOut, error) {
	updatedFile, err := u.fileRepo.ChangeVisibility(ctx, in.UserEmail, in.OriginalName, in.IsPublic)
	if err != nil {
		return nil, err
	}

	return &ChangeVisibilityDtoOut{
		ID:           updatedFile.ID,
		OriginalName: updatedFile.OriginalName,
		IsPublic:     updatedFile.IsPublic,
		UpdatedAt:    updatedFile.UpdatedAt,
	}, nil
}

func (u *fileUsecase) SetStatus(ctx context.Context, in *SetStatusDtoIn) (*SetStatusDtoOut, error) {
	updatedFile, err := u.fileRepo.SetStatus(ctx, in.UserEmail, in.OriginalName, in.Status)
	if err != nil {
		return nil, err
	}

	return &SetStatusDtoOut{
		ID:           updatedFile.ID,
		OriginalName: updatedFile.OriginalName,
		Status:       updatedFile.Status,
		UpdatedAt:    updatedFile.UpdatedAt,
	}, nil
}
