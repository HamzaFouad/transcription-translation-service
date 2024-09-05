package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	OpenAIKey string
}

func LoadConfig() Config {
	// Load environment variables from .env file, if available
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config := Config{
		Port:      getEnv("PORT", "9000"), // Default port is 9000
		OpenAIKey: getEnv("OPENAI_API_KEY", ""),
	}

	return config
}

// Helper function to get an environment variable or default value if not present
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
