package file

import (
	"io"
	"time"
)

type SaveFileMetadataDtoIn struct {
	UserID       int64  `json:"user_id"`
	UserEmail    string `json:"user_email"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	SizeInBytes  int64  `json:"size_in_bytes"`
	S3Bucket     string `json:"s3_bucket"`
	S3Key        string `json:"s3_key"`
	Status       int    `json:"status"`
	IsPublic     bool   `json:"is_public"`
}

type SaveFileMetadataDtoOut struct {
	ID           int64     `json:"id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	SizeInBytes  int64     `json:"size_in_bytes"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	IsPublic     bool      `json:"is_public"`
}

type SaveFileContentDtoIn struct {
	Email string `json:"email"`
	ID    int64  `json:"id"`
	R     io.Reader
}

type SaveFileContentDtoOut struct {
	LoadingResult bool `json:"loading_result"`
}

type GetFileDtoIn struct {
	UserID       int64  `json:"user_id"`
	UserEmail    string `json:"user_email"`
	OriginalName string `json:"original_name"`
}

type GetFileDtoOut struct {
	ID           int64  `json:"id"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	SizeInBytes  int64  `json:"size_in_bytes"`
}

type GetFileInfoDtoIn struct {
	UserID       int64  `json:"user_id"`
	UserEmail    string `json:"user_email"`
	OriginalName string `json:"original_name"`
}

type GetFileInfoDtoOut struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	SizeInBytes  int64     `json:"size_in_bytes"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsPublic     bool      `json:"is_public"`
}

type RenameFileDtoIn struct {
	UserID    int64  `json:"user_id"`
	UserEmail string `json:"user_email"`
	OldName   string `json:"old_name"`
	NewName   string `json:"new_name"`
}

type RenameFileDtoOut struct {
	ID        int64     `json:"id"`
	OldName   string    `json:"old_name"`
	NewName   string    `json:"new_name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeleteFileDtoIn struct {
	UserID       int64  `json:"user_id"`
	UserEmail    string `json:"user_email"`
	OriginalName string `json:"original_name"`
}

type DeleteFileDtoOut struct {
	ID int64 `json:"id"`
}

type GetAllUserFilesDtoIn struct {
	UserID    int64  `json:"user_id"`
	UserEmail string `json:"user_email"`
}

type GetAllUserFilesDtoOut struct {
	Files []FileListItemDto `json:"files"`
}

type FileListItemDto struct {
	ID           int64     `json:"id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	SizeInBytes  int64     `json:"size_in_bytes"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsPublic     bool      `json:"is_public"`
}
