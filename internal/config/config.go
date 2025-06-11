// internal/config/config.go
package config

import (
	"os"
	"strings"
)

type Config struct {
	Port              string
	DatabaseURL       string
	KratosPublicURL   string
	KratosAdminURL    string
	AllowedOrigins    []string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "3000"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://userms:userms_password@localhost:5432/userms?sslmode=disable"),
		KratosPublicURL: getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
		KratosAdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"http://172.16.1.65:3000",
			"http://172.16.1.65:3001",
			"file://",
		},
	}
}

func (c *Config) GetMaskedDatabaseURL() string {
	return strings.ReplaceAll(c.DatabaseURL, "userms_password", "***")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
