package interactor

import (
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	handler "meemo/internal/presenter/http/handler"
	filehandler "meemo/internal/presenter/http/handler/file"
)

type Interactor interface {
	NewAppHandler() handler.AppHandler
}
type interactor struct {
	conn     *sqlx.DB
	s3client *minio.Client
	s3bucket string
}

func NewInteractor(conn *sqlx.DB, s3client *minio.Client, bucket string) Interactor {
	return &interactor{conn: conn, s3client: s3client, s3bucket: bucket}
}

type appHandler struct {
	filehandler.FileHandler
}

func (i *interactor) NewAppHandler() handler.AppHandler {
	appHandler := &appHandler{}
	appHandler.FileHandler = i.NewFileHandler()
	return appHandler
}
