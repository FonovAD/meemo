package handler

import (
	filehandler "meemo/internal/presenter/http/handler/file"
)

type AppHandler interface {
	filehandler.FileHandler
}
