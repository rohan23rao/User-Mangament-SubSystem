package config

import "os"

type Config struct {
	Port             string
	KratosPublicURL  string
	KratosAdminURL   string
	DatabaseURL      string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "3000"),
		KratosPublicURL:  getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
		KratosAdminURL:   getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://userms:userms_password@localhost:5432/userms?sslmode=disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}