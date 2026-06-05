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

type LoggerConfig interface {
	Level() string
	AsJSON() bool
}

type AuthConfig interface {
	PasswordPepper() string
	JWTSecret() string
	AccessTokenTTL() time.Duration
	RefreshTokenSecret() string
	RefreshTokenTTL() time.Duration
}
