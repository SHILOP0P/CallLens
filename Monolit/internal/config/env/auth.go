package env

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type authEnvConfig struct {
	PasswordPepper     string        `env:"PASSWORD_PEPPER,required"`
	JWTSecret          string        `env:"JWT_SECRET,required"`
	AccessTokenTTL     time.Duration `env:"JWT_ACCESS_TOKEN_TTL,required"`
	RefreshTokenSecret string        `env:"REFRESH_TOKEN_SECRET,required"`
	RefreshTokenTTL    time.Duration `env:"REFRESH_TOKEN_TTL,required"`
	SessionTrustAge    time.Duration `env:"AUTH_SESSION_TRUST_AGE" envDefault:"24h"`
}

type authConfig struct {
	raw authEnvConfig
}

func NewAuthConfig() (*authConfig, error) {
	var raw authEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &authConfig{raw: raw}, nil
}

func (cfg *authConfig) PasswordPepper() string {
	return cfg.raw.PasswordPepper
}

func (cfg *authConfig) JWTSecret() string {
	return cfg.raw.JWTSecret
}

func (cfg *authConfig) AccessTokenTTL() time.Duration {
	return cfg.raw.AccessTokenTTL
}

func (cfg *authConfig) RefreshTokenSecret() string {
	return cfg.raw.RefreshTokenSecret
}

func (cfg *authConfig) RefreshTokenTTL() time.Duration {
	return cfg.raw.RefreshTokenTTL
}

func (cfg *authConfig) SessionTrustAge() time.Duration {
	return cfg.raw.SessionTrustAge
}
