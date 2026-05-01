package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type PostgresConfig struct {
	Host         string `env:"POSTGRES_HOST" env-required:"true"`
	Port         string `env:"POSTGRES_PORT" env-required:"true"`
	User         string `env:"POSTGRES_USER" env-required:"true"`
	Password     string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Database     string `env:"POSTGRES_DB" env-required:"true"`
	MigrationDir string `env:"MIGRATION_DIR" env-default:"./migrations"`
}

type HTTPConfig struct {
	Port         string        `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"15s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"15s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type MinIOConfig struct {
	Endpoint   string `env:"MINIO_ENDPOINT" env-required:"true"`
	AccessKey  string `env:"MINIO_ACCESS_KEY" env-required:"true"`
	SecretKey  string `env:"MINIO_SECRET_KEY" env-required:"true"`
	BucketName string `env:"MINIO_BUCKET_NAME" env-default:"hrmate-docs"`
	UseSSL     bool   `env:"MINIO_USE_SSL" env-default:"false"`
}

type Config struct {
	HTTP          HTTPConfig
	Postgres      PostgresConfig
	MinIO         MinIOConfig
	JWTSecret     string `env:"JWT_SECRET" env-required:"true"`
	AdminEmail    string `env:"ADMIN_EMAIL" env-required:"true"`
	AdminPassword string `env:"ADMIN_PASSWORD" env-required:"true"`
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewDB() (*PostgresConfig, error) {
	var cfg PostgresConfig

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}
