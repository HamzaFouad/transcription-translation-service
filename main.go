package main

import (
	"fmt"
	"log"
	"transcriptions-translation-service/config"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/handlers"
	"transcriptions-translation-service/services/openai"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	router := gin.Default()

	translator := openai.NewOpenAIService(&cfg.OpenAIConfig)

	router.POST("/translate",
		data.TranscriptionValidation,
		handlers.TranslateHandler(translator, data.Arabic, data.English))

	log.Println("Starting server on port", cfg.Port)
	err := router.Run(":" + cfg.Port)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
