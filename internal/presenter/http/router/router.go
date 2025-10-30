package router

import (
	"meemo/internal/presenter/http/handler"
	"net/http"

	"github.com/labstack/echo/v4"
)

func NewRouter(e *echo.Echo, h handler.AppHandler) {
	e.GET("/ping", Ping)

	fileRouter := e.Group("/api/v1/file", h.FileMiddleware())
	fileRouter.GET("", h.GetFile)
	fileRouter.GET("/file", h.GetFile)

}

func Ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}
