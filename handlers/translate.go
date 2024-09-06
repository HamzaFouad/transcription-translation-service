package handlers

import (
	"fmt"
	"net/http"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/services/openai"
	"transcriptions-translation-service/utils"

	"github.com/gin-gonic/gin"
)

func TranslateHandler(translator *openai.OpenAIService, sourceLang, targetLang data.Language) gin.HandlerFunc {
	return func(c *gin.Context) {
		transcriptions, ok := getTranscriptionsFromContext(c)
		if !ok {
			return
		}

		batches := groupTranscriptions(transcriptions, openai.DefaultMaxCharSizePerRequest)
		fmt.Println("Batches after grouping: ", len(batches))
		var translatedTranscriptions []data.Transcription

		for _, batch := range batches {
			serializedBatch, err := utils.SerializeToString(batch)
			if err != nil {
				utils.HandleError(c, http.StatusInternalServerError, "unable to serialize batch")
				return
			}

			translatedText, err := translator.Translate(serializedBatch, sourceLang, targetLang)
			if err != nil {
				utils.HandleError(c, http.StatusInternalServerError, err.Error())
				return
			}

			var translatedBatch []data.Transcription
			if err := utils.DeserializeFromString(translatedText, &translatedBatch); err != nil {
				utils.HandleError(c, http.StatusInternalServerError, "unable to deserialize translated text")
				return
			}

			translatedTranscriptions = append(translatedTranscriptions, translatedBatch...)
		}

		if err := utils.ToJSON(translatedTranscriptions, c.Writer); err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "unable to marshal json")
		}
	}
}

func getTranscriptionsFromContext(c *gin.Context) ([]data.Transcription, bool) {
	transcriptions, exists := c.Get(data.KeyTranscription)
	if !exists {
		utils.HandleError(c, http.StatusBadRequest, "transcriptions not found in context")
		return nil, false
	}
	return transcriptions.([]data.Transcription), true
}

func groupTranscriptions(transcriptions []data.Transcription, maxChars int) [][]data.Transcription {
	var batches [][]data.Transcription
	var currentBatch []data.Transcription
	currentChars := 0

	for _, transcription := range transcriptions {
		transcriptionLength := len(transcription.Sentence)

		if currentChars+transcriptionLength > maxChars {
			batches = append(batches, currentBatch)
			currentBatch = []data.Transcription{}
			currentChars = 0
		}

		currentBatch = append(currentBatch, transcription)
		currentChars += transcriptionLength
	}

	// Add the final batch if it has any transcriptions
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches
}
