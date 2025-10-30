package service

type FileService interface {
}

type fileService struct{}

func NewFileService() FileService {
	return &fileService{}
}

const (
	uploading = iota
	uploaded
	deleted
	error
)
