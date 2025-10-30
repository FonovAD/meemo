package interactor

import (
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	filehandler "meemo/internal/presenter/http/handler/file"
	handler "meemo/internal/presenter/http/handler/file"
)

type Interactor interface {
	NewFileHandler() filehandler.FileHandler
}
type interactor struct {
	conn     *sqlx.DB
	s3client *minio.Client
	s3bucket string
}

func NewInteractor(conn *sqlx.DB) Interactor {
	return &interactor{conn: conn}
}

type appHandler struct {
	filehandler.FileHandler
}

func (i *interactor) NewAppHandler() handler.FileHandler {
	appHandler := &appHandler{}
	appHandler.FileHandler = i.NewFileHandler()
	return appHandler
}
