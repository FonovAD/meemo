package file

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	tokenservice "meemo/internal/domain/token/service"
	fileusecase "meemo/internal/usecase/file"
)

type FileHandler interface {
	GetUserFilesList(c echo.Context) error
	GetFileInfo(c echo.Context) error
	RenameFile(c echo.Context) error
	DeleteFile(c echo.Context) error
	SaveFileMetadata(c echo.Context) error
	SaveFileContent(c echo.Context) error
	GetFile(c echo.Context) error
	GetFileByID(c echo.Context) error
	ChangeVisibility(c echo.Context) error
	SetStatus(c echo.Context) error
	FileMiddleware() echo.MiddlewareFunc
}

type fileHandler struct {
	fileUsecase fileusecase.Usecase
	jwtService  tokenservice.TokenService
}

func NewFileHandler(usecase fileusecase.Usecase, jwtService tokenservice.TokenService) FileHandler {
	return &fileHandler{
		fileUsecase: usecase,
		jwtService:  jwtService,
	}
}

// SaveFileMetadata создает метаданные файла
// @Summary Создать метаданные файла
// @Description Создает метаданные для нового файла
// @Tags files
// @Accept json
// @Produce json
// @Param file body SaveFileMetadata true "Метаданные файла"
// @Success 201 {object} fileusecase.SaveFileMetadataDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/metadata [post]
func (h *fileHandler) SaveFileMetadata(c echo.Context) error {
	var req SaveFileMetadata
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	userID := getUserID(c)
	userEmail := getUserEmail(c)

	if userID == 0 || userEmail == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user information not found in token"})
	}

	dto := fileusecase.SaveFileMetadataDtoIn{
		UserID:       userID,
		UserEmail:    userEmail,
		MimeType:     req.MimeType,
		SizeInBytes:  req.SizeInBytes,
		OriginalName: req.OriginalName,
		IsPublic:     req.IsPublic,
	}

	resp, err := h.fileUsecase.SaveFileMetadata(c.Request().Context(), &dto)
	if err != nil {
		if errors.Is(err, nil) { // TODO: описать ошибки
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create file metadata"})
	}

	return c.JSON(http.StatusCreated, resp)
}

// SaveFileContent загружает содержимое файла
// @Summary Загрузить содержимое файла
// @Description Загружает содержимое файла по его ID
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID файла"
// @Param file formData file true "Содержимое файла"
// @Success 200 {object} fileusecase.SaveFileContentDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/{id}/content [post]
func (h *fileHandler) SaveFileContent(c echo.Context) error {
	fileID := c.Param("id")
	if fileID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file ID is required"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open uploaded file"})
	}
	defer src.Close()

	req := &fileusecase.SaveFileContentDtoIn{
		ID:          mustParseInt64(fileID),
		SizeInBytes: file.Size,
	}

	resp, err := h.fileUsecase.SaveFileContent(c.Request().Context(), req, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to upload file content"})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetFile получает файл по имени
// @Summary Получить файл
// @Description Скачивает файл по его имени (включая расширение, например: file.txt)
// @Tags files
// @Produce application/octet-stream
// @Param name path string true "Имя файла с расширением"
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /files/{name} [get]
func (h *fileHandler) GetFile(c echo.Context) error {
	originalName := c.Param("name")
	if originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	req := &fileusecase.GetFileDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: originalName,
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+originalName)
	c.Response().Header().Set("Content-Type", "application/octet-stream") // будет переопределён в UseCase при необходимости

	_, err := h.fileUsecase.GetFile(c.Request().Context(), req, c.Response().Writer)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	c.Response().Status = http.StatusOK
	return nil
}

// GetFileByID получает файл по ID
// @Summary Получить файл по ID
// @Description Скачивает файл по его ID
// @Tags files
// @Produce application/octet-stream
// @Param id path int true "ID файла"
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /files/by-id/{id} [get]
func (h *fileHandler) GetFileByID(c echo.Context) error {
	fileIDStr := c.Param("id")
	if fileIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file ID is required"})
	}

	fileID := mustParseInt64(fileIDStr)

	req := &fileusecase.GetFileByIDDtoIn{
		FileID:    fileID,
		UserID:    getUserID(c),
		UserEmail: getUserEmail(c),
	}

	c.Response().Header().Set("Content-Disposition", "attachment")
	c.Response().Header().Set("Content-Type", "application/octet-stream")

	result, err := h.fileUsecase.GetFileByID(c.Request().Context(), req, c.Response().Writer)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+result.OriginalName)
	c.Response().Status = http.StatusOK
	return nil
}

// GetFileInfo получает информацию о файле
// @Summary Получить информацию о файле
// @Description Возвращает метаданные файла по его имени (включая расширение, например: file.txt)
// @Tags files
// @Produce json
// @Param name path string true "Имя файла с расширением"
// @Success 200 {object} fileusecase.GetFileInfoDtoOut
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /files/{name}/info [get]
func (h *fileHandler) GetFileInfo(c echo.Context) error {
	originalName := c.Param("name")
	if originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	req := &fileusecase.GetFileInfoDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: originalName,
	}

	resp, err := h.fileUsecase.GetFileInfo(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	return c.JSON(http.StatusOK, resp)
}

// RenameFile переименовывает файл
// @Summary Переименовать файл
// @Description Изменяет имя файла
// @Tags files
// @Accept json
// @Produce json
// @Param request body object true "Запрос на переименование" example({"old_name":"old.txt","new_name":"new.txt"})
// @Success 200 {object} fileusecase.RenameFileDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/rename [put]
func (h *fileHandler) RenameFile(c echo.Context) error {
	var req struct {
		OldName string `json:"old_name"`
		NewName string `json:"new_name"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	dto := &fileusecase.RenameFileDtoIn{
		UserID:    getUserID(c),
		UserEmail: getUserEmail(c),
		OldName:   req.OldName,
		NewName:   req.NewName,
	}

	resp, err := h.fileUsecase.RenameFile(c.Request().Context(), dto)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to rename file"})
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteFile удаляет файл
// @Summary Удалить файл
// @Description Удаляет файл по его имени (включая расширение, например: file.txt)
// @Tags files
// @Produce json
// @Param name path string true "Имя файла с расширением"
// @Success 200 {object} fileusecase.DeleteFileDtoOut
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /files/{name} [delete]
func (h *fileHandler) DeleteFile(c echo.Context) error {
	originalName := c.Param("name")
	if originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	req := &fileusecase.DeleteFileDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: originalName,
	}

	resp, err := h.fileUsecase.DeleteFile(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetUserFilesList получает список файлов пользователя
// @Summary Получить список файлов
// @Description Возвращает список всех файлов текущего пользователя
// @Tags files
// @Produce json
// @Success 200 {object} fileusecase.GetAllUserFilesDtoOut
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files [get]
func (h *fileHandler) GetUserFilesList(c echo.Context) error {
	req := &fileusecase.GetAllUserFilesDtoIn{
		UserID:    getUserID(c),
		UserEmail: getUserEmail(c),
	}

	resp, err := h.fileUsecase.GetUserFilesList(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list files"})
	}

	return c.JSON(http.StatusOK, resp)
}

// ChangeVisibility изменяет приватность файла
// @Summary Изменить приватность файла
// @Description Изменяет видимость файла (публичный/приватный)
// @Tags files
// @Accept json
// @Produce json
// @Param request body object true "Запрос на изменение приватности" example({"original_name":"file.txt","is_public":true})
// @Success 200 {object} fileusecase.ChangeVisibilityDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/visibility [put]
func (h *fileHandler) ChangeVisibility(c echo.Context) error {
	var req struct {
		OriginalName string `json:"original_name"`
		IsPublic     bool   `json:"is_public"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	dto := &fileusecase.ChangeVisibilityDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: req.OriginalName,
		IsPublic:     req.IsPublic,
	}

	resp, err := h.fileUsecase.ChangeVisibility(c.Request().Context(), dto)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to change visibility"})
	}

	return c.JSON(http.StatusOK, resp)
}

// SetStatus изменяет статус файла
// @Summary Изменить статус файла
// @Description Изменяет статус файла (например, 0 - активен, 1 - удален)
// @Tags files
// @Accept json
// @Produce json
// @Param request body object true "Запрос на изменение статуса" example({"original_name":"file.txt","status":1})
// @Success 200 {object} fileusecase.SetStatusDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/status [put]
func (h *fileHandler) SetStatus(c echo.Context) error {
	var req struct {
		OriginalName string `json:"original_name"`
		Status       int    `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	dto := &fileusecase.SetStatusDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: req.OriginalName,
		Status:       req.Status,
	}

	resp, err := h.fileUsecase.SetStatus(c.Request().Context(), dto)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to set status"})
	}

	return c.JSON(http.StatusOK, resp)
}

func mustParseInt64(s string) int64 {
	var id int64
	for _, r := range s {
		id = id*10 + int64(r-'0')
	}
	return id
}
