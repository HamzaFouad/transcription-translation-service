package handlers

import (
	"fmt"
	"net/http"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/services"
	"transcriptions-translation-service/utils"

	"github.com/gin-gonic/gin"
)

// handles the translation of transcriptions retrieved from the context.
// -> extracts the concerned texts that needs to be translated
// -> then batches the texts
// -> translates them concurrently
// -> and reintegrates the translated texts.
func TranslateHandler(translator services.Translator, logger utils.Logger, sourceLang, targetLang data.Language) gin.HandlerFunc {
	return func(c *gin.Context) {
		transcriptions, err := fetchTranscriptions(c, logger)
		if err != nil {
			utils.HandleError(c, http.StatusBadRequest, err.Error())
			return
		}

		textsToTranslate := utils.ExtractTextProperties(transcriptions)

		batches := utils.GroupTranscriptionsIntoBatches(textsToTranslate, translator.GetMaxCharSizePerRequest())
		logger.Info("Number of transcriptions: %d, Batches after grouping: %d", len(transcriptions), len(batches))

		translatedTexts, err := services.ProcessBatchesConcurrently(batches, translator, sourceLang, targetLang, logger)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, err.Error())
			return
		}

		translatedTranscriptions := utils.ReintegrateTranslations(transcriptions, translatedTexts)

		if err := utils.SendJSONResponse(translatedTranscriptions, c, logger); err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "unable to marshal json")
		}
	}
}

func fetchTranscriptions(c *gin.Context, logger utils.Logger) ([]data.Transcription, error) {
	transcriptions, ok := utils.GetTranscriptionsFromContext(c)
	if !ok {
		logger.Error("failed to get transcriptions from context")
		return nil, fmt.Errorf("failed to get transcriptions from context")
	}
	return transcriptions, nil
}
