package config

import (
	"github.com/wb-go/wbf/config"
	"log"
	"time"
)

type Config struct {
	DB     DBConfig     `mapstructure:",squash"`
	Server ServerConfig `mapstructure:",squash"`
	Minio  MinioConfig  `mapstructure:",squash"`
}

type DBConfig struct {
	PgUser          string        `mapstructure:"PGUSER"`
	PgPassword      string        `mapstructure:"PGPASSWORD"`
	PgHost          string        `mapstructure:"PGHOST"`
	PgPort          int           `mapstructure:"PGPORT"`
	PgDatabase      string        `mapstructure:"PGDATABASE"`
	PgSSLMode       string        `mapstructure:"PGSSLMODE"`
	MaxOpenConns    int           `mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"CONN_MAX_LIFETIME"`
}

type ServerConfig struct {
	HTTPPort string `mapstructure:"HTTP_PORT"`
}

type MinioConfig struct {
	Endpoint     string `mapstructure:"MINIO_ENDPOINT"`    // Адрес конечной точки
	BucketName   string `mapstructure:"MINIO_BUCKET_NAME"` // Название конкретного бакета в
	RootUser     string `mapstructure:"MINIO_USER"`        // Имя пользователя для доступа к
	RootPassword string `mapstructure:"MINIO_PASSWORD"`    // Пароль для доступа к
	UseSSL       bool   `mapstructure:"MINIO_USE_SSL"`     // Переменная, отвечающая за

}

func MustLoad() *Config {
	c := config.New()
	if err := c.Load(".env", ".env", ""); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var cfg Config

	if err := c.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &cfg
}
