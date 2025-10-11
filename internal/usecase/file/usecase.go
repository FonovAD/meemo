package file

import (
	"context"
	"io"
)

type FileUsecase interface {
	GetUserFilesList(ctx context.Context, in *GetAllUserFilesDtoIn) (*GetAllUserFilesDtoOut, error)
	GetFileInfo(ctx context.Context, in *GetFileInfoDtoIn) (*GetFileInfoDtoOut, error)
	RenameFile(ctx context.Context, in *RenameFileDtoIn) (*RenameFileDtoOut, error)
	DeleteFile(ctx context.Context, in *DeleteFileDtoIn) (*DeleteFileDtoOut, error)
	SaveFile(ctx context.Context, in *SaveFileDtoIn, inReader *io.Reader) (*SaveFileDtoOut, error)
	GetFile(ctx context.Context, in *GetFileDtoIn, inWriter *io.Writer) (*GetFileDtoOut, error)
}
