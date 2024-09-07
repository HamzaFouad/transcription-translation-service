package middleware

import (
	"net/http"
	"transcriptions-translation-service/data"
	"transcriptions-translation-service/utils"

	"github.com/gin-gonic/gin"
)

func TranscriptionValidation(c *gin.Context) {
	var transcriptions []data.Transcription

	if err := utils.FromJSON(&transcriptions, c.Request.Body); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "unable to unmarshal JSON")
		return
	}

	c.Set(data.KeyTranscriptionContext, transcriptions)
	c.Next()
}
