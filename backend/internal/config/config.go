package config

import "os"

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DBDSN         string
	SessionSecret string
	PINReset      string
	Port          string
}

// Load reads required and optional environment variables into a Config.
// Panics if a required variable is missing.
func Load() Config {
	return Config{
		DBDSN:         mustEnv("DB_DSN"),
		SessionSecret: mustEnv("SESSION_SECRET"),
		PINReset:      os.Getenv("APP_PIN_RESET"),
		Port:          envOrDefault("PORT", "8080"),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
