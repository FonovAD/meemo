package router

import (
	"meemo/internal/presenter/http/handler"
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func NewRouter(e *echo.Echo, h handler.AppHandler) {
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/ping", Ping)

	userRouter := e.Group("/api/v1/users")
	userRouter.POST("/register", h.CreateUser)
	userRouter.POST("/login", h.AuthUser)
	userRouter.POST("/refresh", h.UpdateToken)
	userRouter.POST("/logout", h.Logout)

	userProtectedRouter := e.Group("/api/v1/users", h.AuthMiddleware())
	userProtectedRouter.GET("/me", h.GetUserInfo)

	fileRouter := e.Group("/api/v1/files", h.FileMiddleware())
	fileRouter.GET("", h.GetUserFilesList)
	fileRouter.POST("/metadata", h.SaveFileMetadata)
	fileRouter.POST("/:id/content", h.SaveFileContent)
	fileRouter.PUT("/rename", h.RenameFile)
	fileRouter.PUT("/visibility", h.ChangeVisibility)
	fileRouter.PUT("/status", h.SetStatus)
	
	fileRouter.GET("/by-id/:id", h.GetFileByID)
	fileRouter.GET("/:name/info", h.GetFileInfo)
	fileRouter.GET("/:name", h.GetFile)
	fileRouter.DELETE("/:name", h.DeleteFile)
}

// Ping проверяет доступность сервера
// @Summary Проверка доступности сервера
// @Description Возвращает pong для проверки работоспособности API
// @Tags health
// @Produce text/plain
// @Success 200 {string} string "pong"
// @Router /ping [get]
func Ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}
