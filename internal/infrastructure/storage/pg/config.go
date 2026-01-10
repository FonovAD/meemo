package pg

type PGConfig struct {
	Database string `config:"POSTGRES_DB" yaml:"database"`
	User     string `config:"POSTGRES_USER" yaml:"user"`
	Password string `config:"POSTGRES_PASSWORD" yaml:"password"`
	Host     string `config:"POSTGRES_HOST" yaml:"host"`
	Port     string `config:"POSTGRES_PORT" yaml:"port"`
	Retries  int    `config:"DB_CONNECT_RETRY" yaml:"retries"`
	PoolSize int    `config:"DB_POOL_SIZE" yaml:"pool_size"`
	SSLMode  string `config:"SSL_MODE" yaml:"ssl_mode"`
}
