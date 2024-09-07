package handlers

import (
	"net/http"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/services"
	"transcriptions-translation-service/services/openai"
	"transcriptions-translation-service/utils"

	"github.com/gin-gonic/gin"
)

func TranslateHandler(translator services.Translator, logger utils.Logger, sourceLang, targetLang data.Language) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger = logger.WithPrefix("[TranslateHandler]")

		transcriptions, ok := getTranscriptionsFromContext(c)
		if !ok {
			logger.Error("Failed to get transcriptions from context")
			return
		}

		batches := groupTranscriptions(transcriptions, openai.DefaultMaxCharSizePerRequest)
		logger.Info("Batches after grouping: %d", len(batches))
		var translatedTranscriptions []data.Transcription

		for _, batch := range batches {
			serializedBatch, err := utils.SerializeToString(batch)
			if err != nil {
				logger.Error("unable to serialize batch: %v", err)
				utils.HandleError(c, http.StatusInternalServerError, "unable to serialize batch")
				return
			}

			translatedText, err := translator.Translate(serializedBatch, sourceLang, targetLang)
			if err != nil {
				logger.Error("translation error: %v", err)
				utils.HandleError(c, http.StatusInternalServerError, err.Error())
				return
			}

			var translatedBatch []data.Transcription
			if err := utils.DeserializeFromString(translatedText, &translatedBatch); err != nil {
				logger.Error("unable to deserialize translated text: %v", err)
				utils.HandleError(c, http.StatusInternalServerError, "unable to deserialize translated text")
				return
			}

			translatedTranscriptions = append(translatedTranscriptions, translatedBatch...)
		}

		if err := utils.ToJSON(translatedTranscriptions, c.Writer); err != nil {
			logger.Error("unable to marshal JSON: %v", err)
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
