package main

import (
	"fmt"
	"log"
	"transcriptions-translation-service/config"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/handlers"
	"transcriptions-translation-service/services/openai"
	"transcriptions-translation-service/utils"
)

func main() {
	cfg := config.LoadConfig()
	logger := utils.NewStandardLogger("transcription-translation-service")
	translator := openai.NewOpenAIService(&cfg.OpenAIConfig, logger)

	router := utils.SetupRouter()
	router.POST("/translate",
		data.TranscriptionValidation,
		handlers.TranslateHandler(translator, logger, data.Arabic, data.English))

	log.Println("Starting server on port", cfg.Port)
	err := router.Run(":" + cfg.Port)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
