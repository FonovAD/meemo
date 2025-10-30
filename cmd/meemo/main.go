package main

import (
	"errors"
	"flag"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
	"meemo/config"
	"meemo/internal/infrastructure/storage/pg"
	"meemo/internal/infrastructure/storage/s3"
	"meemo/internal/interactor"
	"meemo/internal/presenter/http/router"
	"net/http"
	"os"
)

var configPathFlag = flag.String("config", "", "path to config file")

func main() {
	flag.Parse()
	cfg := setupConfig(*configPathFlag)

	s3, err := s3.NewS3(&cfg.S3)
	if err != nil {
		panic(err)
	}
	pg, err := pg.NewPGConnection(&cfg.Postgres)
	if err != nil {
		panic(err)
	}
	i := interactor.NewInteractor(pg, s3, cfg.S3BucketName)
	h := i.NewAppHandler()

	e := setupEcho()

	router.NewRouter(e, h)
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
		}
	}()

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
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		panic(err)
	}
	return &c
}
