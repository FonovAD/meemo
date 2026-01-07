package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v3"
	"meemo/config"
	_ "meemo/docs"
	"meemo/internal/infrastructure/logger"
	"meemo/internal/infrastructure/storage/pg"
	"meemo/internal/infrastructure/storage/s3"
	"meemo/internal/interactor"
	"meemo/internal/presenter/http/router"
)

// @title Meemo File Storage API
// @version 1.0
// @description API для управления файловым хранилищем
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@meemo.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-User-ID
// @description User ID header

// @securityDefinitions.apikey ApiKeyAuth2
// @in header
// @name X-User-Email
// @description User Email header

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer {access_token}

var configPathFlag = flag.String("config", ".config.yaml", "path to config file")

func main() {
	flag.Parse()
	cfg := setupConfig(*configPathFlag)

	log, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx := context.Background()

	S3, err := s3.NewS3(ctx, &cfg.S3)
	if err != nil {
		log.Fatal("failed to connect to S3")
	}
	PG, err := pg.NewPGConnection(&cfg.Postgres)
	if err != nil {
		log.Fatal("failed to connect to PostgreSQL")
	}

	i := interactor.NewInteractor(PG, S3, cfg.S3BucketName, log)
	h := i.NewAppHandler()

	e := setupEcho()
	router.NewRouter(e, h)

	log.Info("starting server on port " + cfg.Port)

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("shutting down the server")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server shutdown failed")
	}

	log.Info("server stopped")
}

func setupEcho() *echo.Echo {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))

	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-User-ID", "X-User-Email", echo.HeaderXCSRFToken},
		AllowCredentials: true,
		ExposeHeaders:    []string{echo.HeaderXCSRFToken},
	}))

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:" + echo.HeaderXCSRFToken,
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSecure:   false,
		CookieHTTPOnly: false,
		CookieSameSite: http.SameSiteStrictMode,
		Skipper: func(c echo.Context) bool {
			path := c.Path()
			method := c.Request().Method

			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				return true
			}

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				return true
			}

			if path == "/api/v1/users/register" ||
				path == "/api/v1/users/login" ||
				path == "/api/v1/users/refresh" ||
				path == "/api/v1/users/logout" {
				return true
			}

			if path == "/ping" || len(path) >= 8 && path[:8] == "/swagger" {
				return true
			}

			return false
		},
	}))

	return e
}

func setupConfig(path string) *config.Config {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	c := config.Config{}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		panic(err)
	}
	return &c
}
