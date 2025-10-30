package config

import (
	"meemo/internal/infrastructure/storage/pg"
	"meemo/internal/infrastructure/storage/s3"
)

type Config struct {
	Host string `config:"APP_HOST" yaml:"host"`
	Port string `config:"APP_PORT" yaml:"port"`

	Postgres pg.PGConfig `config:"postgres"`
	S3       s3.Config   `config:"s3"`
}
