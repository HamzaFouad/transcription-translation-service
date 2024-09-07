package utils

import (
	"fmt"
	"net/http"
	"transcriptions-translation-service/data"

	"github.com/gin-gonic/gin"
)

func GetTranscriptionsFromContext(c *gin.Context) ([]data.Transcription, bool) {
	transcriptions, exists := c.Get(data.KeyTranscriptionContext)
	if !exists {
		HandleError(c, http.StatusBadRequest, "transcriptions not found in context")
		return nil, false
	}
	return transcriptions.([]data.Transcription), true
}

func ExtractTextProperties(transcriptions []data.Transcription) []string {
	var texts []string
	for _, t := range transcriptions {
		texts = append(texts, t.Sentence)
	}
	return texts
}

func GroupTranscriptionsIntoBatches(transcriptions []string, maxChars int) [][]string {
	var batches [][]string
	var currentBatch []string
	currentChars := 0

	for _, transcription := range transcriptions {
		transcriptionLength := len(transcription)

		if currentChars+transcriptionLength > maxChars {
			batches = append(batches, currentBatch)
			currentBatch = []string{}
			currentChars = 0
		}

		currentBatch = append(currentBatch, transcription)
		currentChars += transcriptionLength
	}

	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches
}

func ReintegrateTranslations(transcriptions []data.Transcription, translatedTexts []string) []data.Transcription {
	for i := range transcriptions {
		transcriptions[i].Sentence = translatedTexts[i]
	}
	return transcriptions
}

func SendJSONResponse(data interface{}, c *gin.Context, logger Logger) error {
	if err := ToJSON(data, c.Writer); err != nil {
		logger.Error("unable to marshal JSON: %v", err)
		return err
	}
	return nil
}

func DeserializeTranslatedText(translatedText string) ([]string, error) {
	var translatedBatch []string
	if err := DeserializeFromString(translatedText, &translatedBatch); err != nil {
		return nil, fmt.Errorf("unable to deserialize translated text: %v", err)
	}
	return translatedBatch, nil
}
