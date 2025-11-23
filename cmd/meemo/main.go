package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
	"meemo/config"
	_ "meemo/docs" // swagger docs
	"meemo/internal/infrastructure/storage/pg"
	"meemo/internal/infrastructure/storage/s3"
	"meemo/internal/interactor"
	"meemo/internal/presenter/http/router"
	"net/http"
	"os"
	"os/signal"
	"time"
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

var configPathFlag = flag.String("config", ".config.yaml", "path to config file")

func main() {
	flag.Parse()
	cfg := setupConfig(*configPathFlag)

	S3, err := s3.NewS3(&cfg.S3)
	if err != nil {
		panic(err)
	}
	fmt.Println("1")
	PG, err := pg.NewPGConnection(&cfg.Postgres)
	if err != nil {
		panic(err)
	}
	fmt.Println("2")
	i := interactor.NewInteractor(PG, S3, cfg.S3BucketName)
	h := i.NewAppHandler()

	e := setupEcho()
	fmt.Println("3")
	router.NewRouter(e, h)
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	fmt.Println("4")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func setupEcho() *echo.Echo {
	e := echo.New()
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
