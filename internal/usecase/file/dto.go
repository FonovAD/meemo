package file

import "time"

// TODO: добавить пагинацию
type GetAllUserFilesDtoIn struct {
	Email string `json:"email"`
}

type FileDto struct {
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

type GetAllUserFilesDtoOut struct {
	Files []FileDto `json:"files"`
}

type GetFileInfoDtoIn struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GetFileInfoDtoOut struct {
	Name      string    `json:"name"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	Status    int       `json:"status"`
	S3Bucket  string    `json:"s3_bucket"`
	S3Key     string    `json:"s3_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsPublic  bool      `json:"is_public"`
}

type RenameFileDtoIn struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	NewName string `json:"new_name"`
}

type RenameFileDtoOut struct {
	NewName   string    `json:"new_name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeleteFileDtoIn struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type DeleteFileDtoOut struct{}

type SaveFileDtoIn struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Status   int    `json:"status"`
	IsPublic bool   `json:"is_public"`
}

type SaveFileDtoOut struct {
	Name      string    `json:"name"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsPublic  bool      `json:"is_public"`
}

type GetFileDtoIn struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GetFileDtoOut struct {
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}
