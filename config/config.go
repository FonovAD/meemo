package config

import (
	"meemo/internal/infrastructure/storage/pg"
	"meemo/internal/infrastructure/storage/s3"
)

type Config struct {
	Host     string `config:"APP_HOST" yaml:"host"`
	Port     string `config:"APP_PORT" yaml:"port"`
	LogLevel string `config:"LOG_LEVEL" yaml:"log_level"`

	Postgres     pg.PGConfig `config:"postgres"  yaml:"postgres"`
	S3           s3.Config   `config:"s3"  yaml:"s3"`
	S3BucketName string      `config:"S3_BUCKET_NAME" yaml:"s3_bucket_name"`
}
