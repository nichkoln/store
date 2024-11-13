// internal/config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	StrapiURL    string
	FrontendURL  string
	JWTSecret    string
	APIProxyPort string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		StrapiURL:    getEnv("STRAPI_URL", "http://strapi:1337"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
		JWTSecret:    getEnv("JWT_SECRET", "your_jwt_secret"),
		APIProxyPort: getEnv("API_PROXY_PORT", "8000"),
	}

	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	return config
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
