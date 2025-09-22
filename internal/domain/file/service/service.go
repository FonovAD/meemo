package service

type FileService interface {
}

const (
	uploading = iota
	uploaded
	deleted
	error
)
