package file

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	tokenservice "meemo/internal/domain/token/service"
	"meemo/internal/infrastructure/logger"
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
	GetStorageInfo(c echo.Context) error
	FileMiddleware() echo.MiddlewareFunc
}

type fileHandler struct {
	fileUsecase fileusecase.Usecase
	jwtService  tokenservice.TokenService
	log         logger.Logger
}

func NewFileHandler(usecase fileusecase.Usecase, jwtService tokenservice.TokenService, log logger.Logger) FileHandler {
	return &fileHandler{
		fileUsecase: usecase,
		jwtService:  jwtService,
		log:         log,
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
		h.log.Warn("invalid JSON in SaveFileMetadata", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	userID := getUserID(c)
	userEmail := getUserEmail(c)

	if userID == 0 || userEmail == "" {
		h.log.Warn("user information not found in token", zap.Int64("userID", userID), zap.String("userEmail", userEmail))
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
		if errors.Is(err, fileusecase.ErrInsufficientStorage) {
			h.log.Warn("insufficient storage space", zap.Int64("userID", userID), zap.Int64("sizeInBytes", req.SizeInBytes))
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "insufficient storage space"})
		}
		h.log.Error("failed to create file metadata", zap.Int64("userID", userID), zap.String("originalName", req.OriginalName), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create file metadata"})
	}

	h.log.Info("file metadata created", zap.Int64("fileID", resp.ID), zap.String("originalName", resp.OriginalName))
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
		h.log.Warn("file is required in SaveFileContent", zap.String("fileID", fileID), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	src, err := file.Open()
	if err != nil {
		h.log.Error("failed to open uploaded file", zap.String("fileID", fileID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open uploaded file"})
	}
	defer src.Close()

	req := &fileusecase.SaveFileContentDtoIn{
		ID:          mustParseInt64(fileID),
		SizeInBytes: file.Size,
	}

	resp, err := h.fileUsecase.SaveFileContent(c.Request().Context(), req, src)
	if err != nil {
		h.log.Error("failed to upload file content", zap.Int64("fileID", req.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to upload file content"})
	}

	h.log.Info("file content uploaded", zap.Int64("fileID", req.ID), zap.Int64("sizeInBytes", req.SizeInBytes))
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
	originalName, err := url.PathUnescape(c.Param("name"))
	if err != nil || originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	req := &fileusecase.GetFileDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: originalName,
	}

	metadata, err := h.fileUsecase.GetFileMetadataByName(c.Request().Context(), req)
	if err != nil {
		h.log.Warn("file not found", zap.String("originalName", originalName), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	filename := ensureFileExtension(metadata.OriginalName, metadata.MimeType)

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Response().Header().Set("Content-Type", metadata.MimeType)

	_, err = h.fileUsecase.GetFile(c.Request().Context(), req, c.Response().Writer)
	if err != nil {
		h.log.Error("failed to download file", zap.Int64("fileID", metadata.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to download file"})
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

	metadata, err := h.fileUsecase.GetFileMetadataByID(c.Request().Context(), req)
	if err != nil {
		h.log.Warn("file not found by ID", zap.Int64("fileID", fileID), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	filename := ensureFileExtension(metadata.OriginalName, metadata.MimeType)

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Response().Header().Set("Content-Type", metadata.MimeType)

	_, err = h.fileUsecase.GetFileByID(c.Request().Context(), req, c.Response().Writer)
	if err != nil {
		h.log.Error("failed to download file by ID", zap.Int64("fileID", fileID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to download file"})
	}

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
	originalName, err := url.PathUnescape(c.Param("name"))
	if err != nil || originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	req := &fileusecase.GetFileInfoDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: originalName,
	}

	resp, err := h.fileUsecase.GetFileInfo(c.Request().Context(), req)
	if err != nil {
		h.log.Warn("file info not found", zap.String("originalName", originalName), zap.Error(err))
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
// @Param request body RenameFileRequest true "Запрос на переименование"
// @Success 200 {object} fileusecase.RenameFileDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/rename [put]
func (h *fileHandler) RenameFile(c echo.Context) error {
	var req RenameFileRequest
	if err := c.Bind(&req); err != nil {
		h.log.Warn("invalid JSON in RenameFile", zap.Error(err))
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
		h.log.Error("failed to rename file", zap.String("oldName", req.OldName), zap.String("newName", req.NewName), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to rename file"})
	}

	h.log.Info("file renamed", zap.Int64("fileID", resp.ID), zap.String("oldName", req.OldName), zap.String("newName", req.NewName))
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
	originalName, err := url.PathUnescape(c.Param("name"))
	if err != nil || originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	req := &fileusecase.DeleteFileDtoIn{
		UserID:       getUserID(c),
		UserEmail:    getUserEmail(c),
		OriginalName: originalName,
	}

	resp, err := h.fileUsecase.DeleteFile(c.Request().Context(), req)
	if err != nil {
		h.log.Error("failed to delete file", zap.String("originalName", originalName), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	h.log.Info("file deleted", zap.Int64("fileID", resp.ID), zap.String("originalName", originalName))
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
		h.log.Error("failed to list files", zap.Int64("userID", req.UserID), zap.Error(err))
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
// @Param request body ChangeVisibilityRequest true "Запрос на изменение приватности"
// @Success 200 {object} fileusecase.ChangeVisibilityDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/visibility [put]
func (h *fileHandler) ChangeVisibility(c echo.Context) error {
	var req ChangeVisibilityRequest
	if err := c.Bind(&req); err != nil {
		h.log.Warn("invalid JSON in ChangeVisibility", zap.Error(err))
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
		h.log.Error("failed to change visibility", zap.String("originalName", req.OriginalName), zap.Bool("isPublic", req.IsPublic), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to change visibility"})
	}

	h.log.Info("visibility changed", zap.Int64("fileID", resp.ID), zap.Bool("isPublic", resp.IsPublic))
	return c.JSON(http.StatusOK, resp)
}

// SetStatus изменяет статус файла
// @Summary Изменить статус файла
// @Description Изменяет статус файла (например, 0 - активен, 1 - удален)
// @Tags files
// @Accept json
// @Produce json
// @Param request body SetStatusRequest true "Запрос на изменение статуса"
// @Success 200 {object} fileusecase.SetStatusDtoOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/status [put]
func (h *fileHandler) SetStatus(c echo.Context) error {
	var req SetStatusRequest
	if err := c.Bind(&req); err != nil {
		h.log.Warn("invalid JSON in SetStatus", zap.Error(err))
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
		h.log.Error("failed to set status", zap.String("originalName", req.OriginalName), zap.Int("status", req.Status), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to set status"})
	}

	h.log.Info("status changed", zap.Int64("fileID", resp.ID), zap.Int("status", resp.Status))
	return c.JSON(http.StatusOK, resp)
}

// GetStorageInfo получает информацию о хранилище пользователя
// @Summary Получить информацию о хранилище
// @Description Возвращает информацию об использованном и доступном месте в хранилище
// @Tags files
// @Produce json
// @Success 200 {object} fileusecase.GetStorageInfoDtoOut
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/storage [get]
func (h *fileHandler) GetStorageInfo(c echo.Context) error {
	dto := &fileusecase.GetStorageInfoDtoIn{
		UserID:    getUserID(c),
		UserEmail: getUserEmail(c),
	}

	resp, err := h.fileUsecase.GetStorageInfo(c.Request().Context(), dto)
	if err != nil {
		h.log.Error("failed to get storage info", zap.Int64("userID", dto.UserID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get storage info"})
	}

	return c.JSON(http.StatusOK, resp)
}

func getExtensionFromMimeType(mimeType string) string {
	if mimeType == "" {
		return ""
	}
	
	lastSlash := -1
	for i := len(mimeType) - 1; i >= 0; i-- {
		if mimeType[i] == '/' {
			lastSlash = i
			break
		}
	}
	
	if lastSlash == -1 || lastSlash == len(mimeType)-1 {
		return ""
	}
	
	extension := mimeType[lastSlash+1:]
	
	for i, ch := range extension {
		if ch == ';' || ch == '+' {
			extension = extension[:i]
			break
		}
	}
	
	if extension == "" || extension == "octet-stream" {
		return ""
	}
	
	return "." + extension
}

func ensureFileExtension(filename, mimeType string) string {
	hasExtension := false
	for i := len(filename) - 1; i >= 0 && i > len(filename)-6; i-- {
		if filename[i] == '.' {
			hasExtension = true
			break
		}
	}
	
	if hasExtension {
		return filename
	}
	
	ext := getExtensionFromMimeType(mimeType)
	if ext != "" {
		return filename + ext
	}
	
	return filename
}

func mustParseInt64(s string) int64 {
	var id int64
	for _, r := range s {
		id = id*10 + int64(r-'0')
	}
	return id
}
