package data

import (
	"net/http"
	"transcriptions-translation-service/utils"

	"github.com/gin-gonic/gin"
)

const KeyTranscription = "transcriptions"

func TranscriptionValidation(c *gin.Context) {
	var transcriptions []Transcription

	if err := utils.FromJSON(&transcriptions, c.Request.Body); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "unable to unmarshal JSON")
		return
	}

	c.Set(KeyTranscription, transcriptions)
	c.Next()
}
