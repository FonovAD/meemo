package s3

type Config struct {
	Endpoint        string `endpoint:"MINIO_ENDPOINT" yaml:"minio_endpoint"`
	AccessKeyID     string `config:"MINIO_ACCESS_KEY_ID" yaml:"minio_access_key_id"`
	SecretAccessKey string `config:"MINIO_SECRET_ACCESS_KEY" yaml:"minio_secret_key"`
	UseSSL          bool   `config:"MINIO_USE_SSL" yaml:"minio_use_ssl"`
	BucketName      string `config:"MINIO_BUCKET_NAME" yaml:"minio_bucket_name"`
}
