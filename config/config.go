package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	OpenAIConfig OpenAIConfig
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port: getEnv("PORT", "9000"),
		OpenAIConfig: OpenAIConfig{
			APIKey:       getEnv("OPENAI_API_KEY", ""),
			ModelName:    getEnv("OPENAI_MODEL_NAME", "gpt-4o-mini"),
			OpenAIAPIURL: getEnv("OPENAI_API_URL", "https://api.openai.com/v1"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
