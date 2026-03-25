package config

import (
	"os"
)

type Config struct {
	Mode   string
	Server ServerConfig
	Database DatabaseConfig
	JWT    JWTConfig
	SMTP   SMTPConfig
}

type ServerConfig struct {
	Addr string
}

type DatabaseConfig struct {
	Path string
}

type JWTConfig struct {
	Secret string
	ExpireHours int
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func Load() *Config {
	cfg := &Config{
		Mode: getEnv("APP_MODE", "debug"),
		Server: ServerConfig{
			Addr: getEnv("SERVER_ADDR", ":8080"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/jumpfrp.db"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "jumpfrp-secret-change-in-production"),
			ExpireHours: 72,
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     587,
			Username: getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASS", ""),
			From:     getEnv("SMTP_FROM", "noreply@jumpfrp.top"),
		},
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
