package env

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type postgresEnvConfig struct {
	Host         string `env:"POSTGRES_HOST,required"`
	Port         string `env:"POSTGRES_PORT,required"`
	Database     string `env:"POSTGRES_DB,required"`
	User         string `env:"POSTGRES_USER,required"`
	Password     string `env:"POSTGRES_PASSWORD,required"`
	SSLMode      string `env:"POSTGRES_SSL_MODE,required"`
	MigrationDir string `env:"MIGRATION_DIRECTORY,required"`
}

type postgresConfig struct {
	raw postgresEnvConfig
}

func NewPostgresConfig() (*postgresConfig, error) {
	var raw postgresEnvConfig
	if err := env.Parse(&raw); err != nil {
		return &postgresConfig{}, err
	}
	return &postgresConfig{raw: raw}, nil
}

func (cfg *postgresConfig) URI() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.raw.User,
		cfg.raw.Password,
		cfg.raw.Host,
		cfg.raw.Port,
		cfg.raw.Database,
		cfg.raw.SSLMode,
	)
}

func (cfg *postgresConfig) DatabaseName() string {
	return cfg.raw.Database
}

func (cfg *postgresConfig) MigrationDir() string {
	return cfg.raw.MigrationDir
}
