package config

import "time"

type HTTPConfig interface {
	Address() string
	ReadTimeout() time.Duration
}

type PostgresConfig interface {
	URI() string
	DatabaseName() string
	MigrationDir() string
}

type UploadConfig interface {
	Path() string
}
