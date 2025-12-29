package user

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	userusecase "meemo/internal/usecase/user"
)

type UserHandler interface {
	CreateUser(c echo.Context) error
	GetUserInfo(c echo.Context) error
	AuthUser(c echo.Context) error
	UpdateToken(c echo.Context) error
	Logout(c echo.Context) error
	AuthMiddleware() echo.MiddlewareFunc
}

type userHandler struct {
	userUsecase userusecase.UseCase
}

func NewUserHandler(usecase userusecase.UseCase) UserHandler {
	return &userHandler{
		userUsecase: usecase,
	}
}

// CreateUser создает нового пользователя
// @Summary Создать пользователя
// @Description Регистрирует нового пользователя и возвращает токены доступа
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "Данные пользователя"
// @Success 201 {object} userusecase.CreateUserDtoOut
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/register [post]
func (h *userHandler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	// Валидация
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "all fields are required"})
	}

	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "password must be at least 6 characters"})
	}

	dto := &userusecase.CreateUserDtoIn{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}

	resp, err := h.userUsecase.CreateUser(c.Request().Context(), dto)
	if err != nil {
		// Проверка на дубликат email
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "user with this email already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	return c.JSON(http.StatusCreated, resp)
}

// AuthUser аутентифицирует пользователя
// @Summary Аутентификация пользователя
// @Description Авторизует пользователя и возвращает токены доступа
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body AuthUserRequest true "Учетные данные"
// @Success 200 {object} userusecase.UserDtoOut
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/login [post]
func (h *userHandler) AuthUser(c echo.Context) error {
	var req AuthUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
	}

	dto := &userusecase.UserDtoIn{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.userUsecase.AuthUser(c.Request().Context(), dto)
	if err != nil {
		if errors.Is(err, errors.New("wrong password")) || strings.Contains(err.Error(), "password") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		}
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to authenticate user"})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetUserInfo получает информацию о текущем пользователе
// @Summary Получить информацию о пользователе
// @Description Возвращает информацию о текущем авторизованном пользователе
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {access_token}"
// @Success 200 {object} userusecase.GetUserInfoOut
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /users/me [get]
func (h *userHandler) GetUserInfo(c echo.Context) error {
	// Извлекаем токен из заголовка Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization header is required"})
	}

	// Убираем префикс "Bearer "
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		// Если префикс не найден, пробуем использовать заголовок как есть
		token = authHeader
	}

	dto := &userusecase.GetUserInfoDtoIn{
		AccessToken: token,
	}

	resp, err := h.userUsecase.GetUserInfo(c.Request().Context(), dto)
	if err != nil {
		if strings.Contains(err.Error(), "token") || strings.Contains(err.Error(), "expired") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user info"})
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateToken обновляет токен доступа
// @Summary Обновить токен доступа
// @Description Обновляет access token используя refresh token
// @Tags users
// @Accept json
// @Produce json
// @Param token body UpdateTokenRequest true "Refresh token"
// @Success 200 {object} userusecase.UpdateTokenDtoOut
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/refresh [post]
func (h *userHandler) UpdateToken(c echo.Context) error {
	var req UpdateTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
	}

	dto := &userusecase.UpdateTokenDtoIn{
		RefreshToken: req.RefreshToken,
	}

	resp, err := h.userUsecase.UpdateToken(c.Request().Context(), dto)
	if err != nil {
		if strings.Contains(err.Error(), "token") || strings.Contains(err.Error(), "expired") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update token"})
	}

	return c.JSON(http.StatusOK, resp)
}

// Logout выходит из системы
// @Summary Выход из системы
// @Description Инвалидирует токен доступа пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param token body LogoutRequest true "Access token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/logout [post]
func (h *userHandler) Logout(c echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	if req.AccessToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "access_token is required"})
	}

	dto := &userusecase.LogoutDtoIn{
		AccessToken: req.AccessToken,
	}

	_, err := h.userUsecase.Logout(c.Request().Context(), dto)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to logout"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// AuthMiddleware middleware для проверки JWT токена
func (h *userHandler) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization header is required"})
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				token = authHeader
			}

			// Проверяем токен через usecase
			dto := &userusecase.GetUserInfoDtoIn{
				AccessToken: token,
			}

			userInfo, err := h.userUsecase.GetUserInfo(c.Request().Context(), dto)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}

			// Сохраняем информацию о пользователе в контексте
			c.Set("user_email", userInfo.Email)
			c.Set("user_info", userInfo)

			return next(c)
		}
	}
}


