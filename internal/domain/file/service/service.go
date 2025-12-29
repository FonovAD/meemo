package service

import (
	"meemo/internal/domain/entity"
	"time"
)

type FileService interface {
	CreateFileMetadata(file *entity.File)
}

type fileService struct{}

func NewFileService() FileService {
	return &fileService{}
}

func (s *fileService) CreateFileMetadata(file *entity.File) {
	file.Status = entity.Pending
	file.SizeInBytes = 0
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()
}
