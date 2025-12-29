package file

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
)

func (h *fileHandler) FileMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			authHeader := ctx.Request().Header.Get("Authorization")
			if authHeader == "" {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization header is required"})
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				token = authHeader
			}

			claims, err := h.jwtService.ParseAccessToken(token)
			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}

			exp, err := claims.GetExpirationTime()
			if err != nil || exp == nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "token expired"})
			}

			userID, err := strconv.ParseInt(claims.UserID, 10, 64)
			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token: user_id is not a valid number"})
			}

			ctx.Set(UserIDKey, userID)
			ctx.Set(UserEmailKey, claims.Email)

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
