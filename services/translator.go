package services

import "transcriptions-translation-service/data"

type Translator interface {
	TranslateAsync(text string, sourceLang, targetLang data.Language) (string, error)
	GetMaxCharSizePerRequest() int
}
