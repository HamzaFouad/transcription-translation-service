package handlers

import (
	"fmt"
	"net/http"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/services"
	"transcriptions-translation-service/utils"

	"github.com/gin-gonic/gin"
)

// TranslateHandler handles the translation of transcriptions retrieved from the context.
// extracts the concerned texts that needs to be translated -> then batches the texts, translates them concurrently, and reintegrates the translated texts.
func TranslateHandler(translator services.Translator, logger utils.Logger, sourceLang, targetLang data.Language) gin.HandlerFunc {
	return func(c *gin.Context) {
		transcriptions, err := fetchTranscriptions(c, logger)
		if err != nil {
			utils.HandleError(c, http.StatusBadRequest, err.Error())
			return
		}

		textsToTranslate := extractTextProperties(transcriptions)

		batches := groupTranscriptionsIntoBatches(textsToTranslate, translator.GetMaxCharSizePerRequest())
		logger.Info("Number of transcriptions: %d, Batches after grouping: %d", len(transcriptions), len(batches))

		translatedTexts, err := processBatchesConcurrently(batches, translator, sourceLang, targetLang, logger)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, err.Error())
			return
		}

		translatedTranscriptions := reintegrateTranslations(transcriptions, translatedTexts)

		if err := sendJSONResponse(translatedTranscriptions, c, logger); err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "unable to marshal json")
		}
	}
}

func fetchTranscriptions(c *gin.Context, logger utils.Logger) ([]data.Transcription, error) {
	transcriptions, ok := getTranscriptionsFromContext(c)
	if !ok {
		logger.Error("Failed to get transcriptions from context")
		return nil, fmt.Errorf("failed to get transcriptions from context")
	}
	return transcriptions, nil
}

func processBatchesConcurrently(batches [][]string, translator services.Translator, sourceLang, targetLang data.Language, logger utils.Logger) ([]string, error) {
	translatedTexts := make([]string, 0)
	type result struct {
		index           int
		translatedBatch []string
		err             error
	}

	resultsChan := make(chan result)

	for index, batch := range batches {
		go func(index int, batch []string) {
			serializedBatch, _ := utils.SerializeToString(batch)

			translatedText, err := translator.TranslateAsync(serializedBatch, sourceLang, targetLang)
			if err != nil {
				resultsChan <- result{index: index, err: err}
				return
			}

			if !utils.IsValidJSON(translatedText) {
				resultsChan <- result{index: index, err: fmt.Errorf("invalid JSON format received: %s", translatedText)}
				return
			}

			var translatedBatch []string
			if err := utils.DeserializeFromString(translatedText, &translatedBatch); err != nil {
				resultsChan <- result{index: index, err: fmt.Errorf("unable to deserialize translated text: %v", err)}
				return
			}

			resultsChan <- result{index: index, translatedBatch: translatedBatch}
		}(index, batch)
	}

	// Ensure the results are collected in the same order they were dispatched
	// This is critical for reintegrating the translated texts back correctly.
	results := make([][]string, len(batches))

	for i := 0; i < len(batches); i++ {
		res := <-resultsChan
		if res.err != nil {
			logger.Error("There happen to be a translation error for batch: %v", res.err)
			return nil, res.err
		}
		results[res.index] = res.translatedBatch
	}

	// Flatten the results into a single slice
	for _, translatedBatch := range results {
		translatedTexts = append(translatedTexts, translatedBatch...)
	}

	return translatedTexts, nil
}

func getTranscriptionsFromContext(c *gin.Context) ([]data.Transcription, bool) {
	transcriptions, exists := c.Get(data.KeyTranscription)
	if !exists {
		utils.HandleError(c, http.StatusBadRequest, "transcriptions not found in context")
		return nil, false
	}
	return transcriptions.([]data.Transcription), true
}

func groupTranscriptionsIntoBatches(transcriptions []string, maxChars int) [][]string {
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

func reintegrateTranslations(transcriptions []data.Transcription, translatedTexts []string) []data.Transcription {
	for i := range transcriptions {
		transcriptions[i].Sentence = translatedTexts[i]
	}
	return transcriptions
}

func sendJSONResponse(data interface{}, c *gin.Context, logger utils.Logger) error {
	if err := utils.ToJSON(data, c.Writer); err != nil {
		logger.Error("unable to marshal JSON: %v", err)
		return err
	}
	return nil
}
