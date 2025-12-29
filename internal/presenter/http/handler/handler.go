package handler

import (
	filehandler "meemo/internal/presenter/http/handler/file"
	userhandler "meemo/internal/presenter/http/handler/user"
)

type AppHandler interface {
	filehandler.FileHandler
	userhandler.UserHandler
}
