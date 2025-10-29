package file

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
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
	FileMiddleware() echo.MiddlewareFunc
}

type fileHandler struct {
	fileUsecase fileusecase.Usecase
}

func NewFileHandler(usecase fileusecase.Usecase) FileHandler {
	return &fileHandler{
		fileUsecase: usecase,
	}
}

func getUserIDAndEmail(c echo.Context) (int64, string, error) {
	userIDVal := c.Get("user_id")
	emailVal := c.Get("user_email")

	if userIDVal == nil {
		return 0, "", errors.New("user_id not found in context")
	}
	if emailVal == nil {
		return 0, "", errors.New("user_email not found in context")
	}

	userID, ok := userIDVal.(int64)
	if !ok {
		return 0, "", errors.New("invalid user_id type")
	}

	email, ok := emailVal.(string)
	if !ok {
		return 0, "", errors.New("invalid user_email type")
	}

	return userID, email, nil
}

func (h *fileHandler) SaveFileMetadata(c echo.Context) error {
	var req fileusecase.SaveFileMetadataDtoIn
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	userID, userEmail, err := getUserIDAndEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req.UserID = userID
	req.UserEmail = userEmail

	resp, err := h.fileUsecase.SaveFileMetadata(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, nil) { // TODO: описать ошибки
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create file metadata"})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *fileHandler) SaveFileContent(c echo.Context) error {
	fileID := c.Param("id")
	if fileID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file ID is required"})
	}

	// В реальности, вы можете получить email из метаданных по fileID,
	// но для упрощения предположим, что email передаётся в теле или в заголовке.
	// Лучше: добавить метод GetFileOwner в репозиторий.
	_, userEmail, err := getUserIDAndEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req := &fileusecase.SaveFileContentDtoIn{
		ID:    mustParseInt64(fileID), // реализуйте безопасный парсинг
		Email: userEmail,
	}

	resp, err := h.fileUsecase.SaveFileContent(c.Request().Context(), req, c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to upload file content"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *fileHandler) GetFile(c echo.Context) error {
	originalName := c.Param("name")
	if originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	userID, userEmail, err := getUserIDAndEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req := &fileusecase.GetFileDtoIn{
		UserID:       userID,
		UserEmail:    userEmail,
		OriginalName: originalName,
	}

	// Отправляем заголовки до копирования
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+originalName)
	c.Response().Header().Set("Content-Type", "application/octet-stream") // будет переопределён в UseCase при необходимости

	_, err = h.fileUsecase.GetFile(c.Request().Context(), req, c.Response().Writer)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	// Важно: не вызывать c.JSON после записи в Writer!
	// UseCase уже записал тело, поэтому просто завершаем
	c.Response().Status = http.StatusOK
	return nil
}

func (h *fileHandler) GetFileInfo(c echo.Context) error {
	originalName := c.Param("name")
	if originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	userID, userEmail, err := getUserIDAndEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req := &fileusecase.GetFileInfoDtoIn{
		UserID:       userID,
		UserEmail:    userEmail,
		OriginalName: originalName,
	}

	resp, err := h.fileUsecase.GetFileInfo(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *fileHandler) RenameFile(c echo.Context) error {
	var req fileusecase.RenameFileDtoIn
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	userID, userEmail, err := getUserIDAndEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req.UserID = userID
	req.UserEmail = userEmail

	resp, err := h.fileUsecase.RenameFile(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to rename file"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *fileHandler) DeleteFile(c echo.Context) error {
	originalName := c.Param("name")
	if originalName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file name is required"})
	}

	userID, userEmail, err := getUserIDAndEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req := &fileusecase.DeleteFileDtoIn{
		UserID:       userID,
		UserEmail:    userEmail,
		OriginalName: originalName,
	}

	resp, err := h.fileUsecase.DeleteFile(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *fileHandler) GetUserFilesList(c echo.Context) error {
	userID, userEmail, err := getUserIDAndEmail(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req := &fileusecase.GetAllUserFilesDtoIn{
		UserID:    userID,
		UserEmail: userEmail,
	}

	resp, err := h.fileUsecase.GetUserFilesList(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, errors.New("not implemented")) {
			return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list files"})
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
