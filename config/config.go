package config

import (
	"os"
	"strconv"

	"meemo/internal/infrastructure/storage/pg"
	"meemo/internal/infrastructure/storage/s3"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	LogLevel string `yaml:"log_level"`

	RegistrationEnabled bool `yaml:"registration_enabled"`

	Postgres     pg.PGConfig `yaml:"postgres"`
	S3           s3.Config   `yaml:"s3"`
	S3BucketName string      `yaml:"s3_bucket_name"`
}

func (c *Config) LoadSecretsFromEnv() {
	if v := os.Getenv("POSTGRES_PASSWORD"); v != "" {
		c.Postgres.Password = v
	}
	if v := os.Getenv("POSTGRES_USER"); v != "" {
		c.Postgres.User = v
	}

	if v := os.Getenv("AWS_ACCESS_KEY_ID"); v != "" {
		c.S3.AccessKeyID = v
	}
	if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v != "" {
		c.S3.SecretAccessKey = v
	}

	if v := os.Getenv("APP_HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("APP_PORT"); v != "" {
		c.Port = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("REGISTRATION_ENABLED"); v != "" {
		c.RegistrationEnabled, _ = strconv.ParseBool(v)
	}
}
