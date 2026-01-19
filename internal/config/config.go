package config

import "os"

// Config holds application configuration
type Config struct {
	Port      string
	RedisURL  string
	BaseURL   string
	StaticDir string
}

// Load returns configuration from environment variables
func Load() *Config {
	port := getEnv("PORT", "8080")
	return &Config{
		Port:      port,
		RedisURL:  getEnv("REDIS_URL", "redis://localhost:6379"),
		BaseURL:   getEnv("BASE_URL", "http://localhost:"+port),
		StaticDir: getEnv("STATIC_DIR", "./static"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
