package interactor

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
	"meemo/internal/infrastructure/logger"
	handler "meemo/internal/presenter/http/handler"
	filehandler "meemo/internal/presenter/http/handler/file"
	userhandler "meemo/internal/presenter/http/handler/user"
)

type Interactor interface {
	NewAppHandler() handler.AppHandler
}
type interactor struct {
	conn                *sqlx.DB
	s3client            *s3.Client
	s3bucket            string
	log                 logger.Logger
	registrationEnabled bool
}

func NewInteractor(conn *sqlx.DB, s3client *s3.Client, bucket string, log logger.Logger, registrationEnabled bool) Interactor {
	return &interactor{conn: conn, s3client: s3client, s3bucket: bucket, log: log, registrationEnabled: registrationEnabled}
}

type appHandler struct {
	filehandler.FileHandler
	userhandler.UserHandler
}

func (i *interactor) NewAppHandler() handler.AppHandler {
	appHandler := &appHandler{}
	appHandler.FileHandler = i.NewFileHandler()
	appHandler.UserHandler = i.NewUserHandler()
	return appHandler
}
