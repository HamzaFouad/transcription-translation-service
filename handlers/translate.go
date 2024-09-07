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
		transcriptions, ok := getTranscriptionsFromContext(c)
		if !ok {
			logger.Error("Failed to get transcriptions from context")
			return
		}

		textsToTranslate := extractTextProperties(transcriptions)
		batches := groupTranscriptions(textsToTranslate, openai.DefaultMaxCharSizePerRequest)

		logger.Info("Number of transcriptions: %d, Batches after grouping: %d", len(transcriptions), len(batches))

		var translatedTexts []string

		for _, batch := range batches {
			serializedBatch, _ := utils.SerializeToString(batch)

			translatedText, err := translator.Translate(serializedBatch, sourceLang, targetLang)

			if err != nil {
				logger.Error("translation error: %v", err)
				utils.HandleError(c, http.StatusInternalServerError, err.Error())
				return
			}

			var translatedBatch []string
			if err := utils.DeserializeFromString(translatedText, &translatedBatch); err != nil {
				logger.Error("unable to deserialize translated text: %v", err)
				utils.HandleError(c, http.StatusInternalServerError, "unable to deserialize translated text")
				return
			}

			translatedTexts = append(translatedTexts, translatedBatch...)
		}

		translatedTranscriptions := reintegrateTranslatedProperty(transcriptions, translatedTexts)

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

func groupTranscriptions(transcriptions []string, maxChars int) [][]string {
	var batches [][]string
	var currentBatch []string
	currentChars := 0

	for _, transcription := range transcriptions {
		transcriptionLength := len(transcription)

		if currentChars+transcriptionLength > maxChars {
			batches = append(batches, currentBatch)
			currentBatch = []string{} // Reset the current batch
			currentChars = 0          // Reset character count
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

func extractTextProperties(transcriptions []data.Transcription) []string {
	var texts []string
	for _, t := range transcriptions {
		texts = append(texts, t.Sentence)
	}
	return texts
}

func reintegrateTranslatedProperty(originalTranscriptions []data.Transcription, translatedValues []string) []data.Transcription {
	for i := range originalTranscriptions {
		originalTranscriptions[i].Sentence = translatedValues[i]
	}
	return originalTranscriptions
}
