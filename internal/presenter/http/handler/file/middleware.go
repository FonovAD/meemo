package file

import (
	"strconv"
	"github.com/labstack/echo/v4"
)

const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
)

func (h *fileHandler) FileMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// В реальном приложении здесь должна быть аутентификация через JWT
			// Для тестирования извлекаем user_id и user_email из заголовков
			userIDStr := ctx.Request().Header.Get("X-User-ID")
			userEmail := ctx.Request().Header.Get("X-User-Email")
			
			if userIDStr != "" {
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err == nil {
					ctx.Set(UserIDKey, userID)
				}
			}
			
			if userEmail != "" {
				ctx.Set(UserEmailKey, userEmail)
			}
			
			return next(ctx)
		}
	}
}

func getUserID(ctx echo.Context) int64 {
	if userID, ok := ctx.Get(UserIDKey).(int64); ok {
		return userID
	}
	return 0
}

func getUserEmail(ctx echo.Context) string {
	if email, ok := ctx.Get(UserEmailKey).(string); ok {
		return email
	}
	return ""
}
