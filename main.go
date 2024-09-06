package main

import (
	"fmt"
	"log"
	"net/http"
	"transcriptions-translation-service/config"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/services/openai"
	"transcriptions-translation-service/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	router := gin.Default()

	translator := openai.NewOpenAIService(&cfg.OpenAIConfig)

	// POST route for translation
	router.POST("/translate", func(c *gin.Context) {
		// Define a request struct to bind the incoming JSON body
		var request struct {
			Text string `json:"text"`
		}

		// Use FromJSON utility to decode the request body
		if err := utils.FromJSON(&request, c.Request.Body); err != nil {
			utils.HandleError(c, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Call the Translate method with the extracted text
		translatedText, err := translator.Translate(request.Text, data.Arabic, data.English)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Failed to translate text")
			return
		}

		// Respond with the translated text
		c.JSON(http.StatusOK, gin.H{"translated_text": translatedText})
	})
	log.Println("Starting server on port", cfg.Port)
	err := router.Run(":" + cfg.Port)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
