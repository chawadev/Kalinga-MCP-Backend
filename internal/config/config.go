package config

import (
	"os"
)

type Config struct {
	MongoURI      string
	DatabaseName  string
	JWTSecret     string
	Port          string
	GeminiAPIKey  string
	GoogleAPIKey  string
}

func Load() *Config {
	return &Config{
		MongoURI:     getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DatabaseName: getEnv("DATABASE_NAME", "kalinga"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:         getEnv("PORT", "8080"),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GoogleAPIKey: getEnv("GOOGLE_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
