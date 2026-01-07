package s3

type Config struct {
	Region          string `yaml:"aws_region"`
	Endpoint        string `yaml:"aws_endpoint"`
	AccessKeyID     string `yaml:"aws_access_key_id"`
	SecretAccessKey string `yaml:"aws_secret_access_key"`
	BucketName      string `yaml:"aws_bucket_name"`
	ForcePathStyle  bool   `yaml:"force_path_style"`
}
