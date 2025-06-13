package config

import (
	"os"
)

type Config struct {
	Port             string
	DatabaseURL      string
	KratosPublicURL  string
	KratosAdminURL   string
	HydraPublicURL   string  // NEW: Hydra public URL
	HydraAdminURL    string  // NEW: Hydra admin URL
	GoogleClientID   string
	GoogleClientSecret string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "3000"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://userms:userms_password@localhost:5434/userms?sslmode=disable"),
		KratosPublicURL: getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
		KratosAdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		HydraPublicURL:  getEnv("HYDRA_PUBLIC_URL", "http://localhost:4444"),   // NEW
		HydraAdminURL:   getEnv("HYDRA_ADMIN_URL", "http://localhost:4445"),    // NEW
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}