package s3

type Config struct {
	Region          string `yaml:"aws_region" env:"AWS_REGION"`
	Endpoint        string `yaml:"aws_endpoint" env:"AWS_ENDPOINT"`
	AccessKeyID     string `yaml:"aws_access_key_id" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `yaml:"aws_secret_access_key" env:"AWS_SECRET_ACCESS_KEY"`
	BucketName      string `yaml:"aws_bucket_name" env:"AWS_BUCKET_NAME"`
	ForcePathStyle  bool   `yaml:"force_path_style" env:"AWS_FORCE_PATH_STYLE"`
}
